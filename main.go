package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

type ApiResponse struct {
	Data string `json:"data"`
}

func main() {
	var noInit bool
	var switchContext bool
	var skelBranch string
	var kubeVersion string

	var rootCmd = &cobra.Command{
		Use:   "wrapper",
		Short: "A CLI for making API calls",
		Long:  `A CLI for making API calls to various services`,
	}

	var vclusterListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all deployed vcluster",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			mainConf := getConfig("")
			_ = getAndValidateContext(mainConf)
			cli, err := client.NewClientWithOpts(client.FromEnv)
			if err != nil {
				panic(err)
			}
			ctx := context.Background()

			cfg := containerConfig(mainConf, Vcluster, strslice.StrSlice{"list"})
			hostCfg := hostConfig()
			if err != nil {
				panic(err)
			}
			runContainer(ctx, cli, &cfg, &hostCfg)
		},
	}
	var vclusterConnectCmd = &cobra.Command{
		Use:   "connect",
		Short: "Grab vcluster Kubeconfig",
		Long:  `Grab vcluster Kubeconfig and use it.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			mainConf := getConfig(args[0])
			// Only validating context here
			_ = getAndValidateContext(mainConf)
			cli, err := client.NewClientWithOpts(client.FromEnv)
			if err != nil {
				panic(err)
			}
			ctx := context.Background()
			config := getConfig(args[0])

			cfg := containerConfig(mainConf, Vcluster, strslice.StrSlice{
				"connect",
				args[0],
				"-n",
				config.Vcluster.Namespace,
				fmt.Sprintf("--server=https://%s.%s", args[0], config.Vcluster.HostFqdn),
				fmt.Sprintf("--kube-config=%s", mainConf.Runtime.KubeconfPath),
			})
			hostCfg := hostConfig()

			// Getting the new context
			kubeconf := getAndValidateContext(mainConf)
			runContainer(ctx, cli, &cfg, &hostCfg)
			// hacky way to force new kube context, I'm pretty sure we can find something better
			kubeconf.CurrentContext = fmt.Sprintf("vcluster_%s_vclusters_%s", mainConf.Runtime.VclusterName, mainConf.Vcluster.HostContextName)
			dumpKubeconf(mainConf, kubeconf)
		},
	}
	var vclusterDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Remove a vcluster",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			mainConf := getConfig(args[0])
			_ = getAndValidateContext(mainConf)
			cli, err := client.NewClientWithOpts(client.FromEnv)
			if err != nil {
				panic(err)
			}
			ctx := context.Background()

			cfg := containerConfig(mainConf, Vcluster, strslice.StrSlice{
				"delete",
				args[0],
				"-n",
				mainConf.Vcluster.Namespace,
			})
			hostCfg := hostConfig()

			runContainer(ctx, cli, &cfg, &hostCfg)
		},
	}
	var vclusterCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a vcluster",
		Long:  `Create a vcluster, choose a name.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			mainConf := getConfig(args[0])
			_ = getAndValidateContext(mainConf)
			cli, err := client.NewClientWithOpts(client.FromEnv)
			if err != nil {
				panic(err)
			}
			ctx := context.Background()

			cliArgs := strslice.StrSlice{}
			// Allow local values.yaml to be used to deploy cluster
			if _, err := os.Stat("vc-values.yaml"); errors.Is(err, os.ErrNotExist) {
				cliArgs = strslice.StrSlice{
					"--set",
					"ingress.enabled=true",
					"--set",
					"ingress.ingressClassName=nginx",
					"--set",
					fmt.Sprintf("ingress.host=%s.%s", args[0], mainConf.Vcluster.HostFqdn),
					"--set",
					fmt.Sprintf("syncer.extraArgs.0=--tls-san=%s.%s", args[0], mainConf.Vcluster.HostFqdn),
				}
			} else {
				cliArgs = strslice.StrSlice{
					"--values",
					"vc-values.yaml",
				}
			}

			baseArgs := strslice.StrSlice{
				"create",
				args[0],
				"-n",
				mainConf.Vcluster.Namespace,
				"--connect=false",
				"--upgrade",
				"--kubernetes-version",
				kubeVersion,
			}

			cfg := containerConfig(mainConf, Vcluster, append(baseArgs, cliArgs...))
			hostCfg := hostConfig()

			runContainer(ctx, cli, &cfg, &hostCfg)
		},
	}
	var repoCmd = &cobra.Command{
		Use:   "repo",
		Short: "Create/Switch to git branch to deploy hemera.",
		Long: `Create/Switch to git branch to deploy hemera,
      if not overriden in config.yaml it will try to work in local folder ".".
      Please provide vcluster name as first argument.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			mainConf := getConfig(args[0])
			gitcfg := initGitConfig(mainConf)
			err := gitBranchCheckout(gitcfg)
			if err != nil {
				panic(err)
			}

			// Init repo by templating it with copier
			if !noInit {
				config := containerConfig(mainConf, Copier, strslice.StrSlice{
					"copy",
					"-f",
					"-r",
					skelBranch,
					mainConf.Git.SkeletonRepo,
					mainConf.Runtime.SourceFolder,
					"--trust",
					"--data",
					fmt.Sprintf("cluster_name=%s", gitcfg.branchNameSlug),
				})
				hostCfg := hostConfig()

				cli, err := client.NewClientWithOpts(client.FromEnv)
				if err != nil {
					panic(err)
				}
				runContainer(cmd.Context(), cli, &config, &hostCfg)
			}
			fmt.Printf("Branch has been templated, please review/commit/push in folder %s to trigger Hemera deployment.\n", mainConf.Runtime.SourceFolder)
		},
	}
	var hemeraDeployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Basically does a git push. Please provide vcluster name you want to deploy to.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := getConfig(args[0])
			repo := initGitConfig(cfg)
			gitlab := gitlabClient(repo)

			// Push env variables for CI
			gitlab.gitlabAddOrUpdateVariable(GitlabVar{
				Name:  "CLEYROP_DOMAIN_NAME",
				Value: cfg.Runtime.VclusterFqdn,
			})
			gitlab.gitlabAddOrUpdateVariable(GitlabVar{
				Name:  "KUBE_CONFIG",
				Value: getKubeConfAsString(cfg),
			})

			commit, err := gitAddCommitPush(repo)

			if err != nil {
				panic(err)
			}

			fmt.Printf("Branch %s, was pushed to %s with commit %s\n", repo.branchName, cfg.Git.DeployRepo, commit)
		},
	}

	rootCmd.AddCommand(vclusterListCmd)
	//rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(vclusterConnectCmd)
	rootCmd.AddCommand(vclusterCreateCmd)
	vclusterCreateCmd.PersistentFlags().StringVar(&kubeVersion, "kube-version", "v1.26", "Set kubernete version (only including minor, not patch).")
	rootCmd.AddCommand(vclusterDeleteCmd)
	rootCmd.AddCommand(repoCmd)
	rootCmd.AddCommand(hemeraDeployCmd)
	rootCmd.Flags().BoolVarP(&switchContext, "force-switch", "f", false, "Experimental: Force kube context switch when needed.")
	repoCmd.Flags().BoolVarP(&noInit, "no-init", "n", false, "Should we disable update branch from skeleton?")
	repoCmd.PersistentFlags().StringVarP(&skelBranch, "skeleton-branch", "b", "develop", "hemera-skel branch or tag to source from.")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
