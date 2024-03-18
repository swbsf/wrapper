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

const ErrBranchNotFound = "branch not found"

func main() {
	var init bool
	var switchContext bool
	var skelBranch string
	var kubeVersion string

	var rootCmd = &cobra.Command{
		Use:   "wrapper",
		Short: "A CLI for making API calls",
		Long:  `A CLI for making API calls to various services`,
	}
	//var testCmd = &cobra.Command{
	//  Use:   "d",
	//  Short: "List all deployed vcluster",
	//  Args:  cobra.ExactArgs(0),
	//  Run: func(cmd *cobra.Command, args []string) {
	//    config, _ := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
	//      &clientcmd.ClientConfigLoadingRules{Precedence: strings.Split(os.Getenv("KUBECONFIG"), ":")},
	//      &clientcmd.ConfigOverrides{
	//          CurrentContext: "",
	//      }).RawConfig()
	//    fmt.Println(config.CurrentContext)
	//  },
	//}
	//var getCmd = &cobra.Command{
	//  Use:   "get",
	//  Short: "Make a GET request to an API endpoint",
	//  Long:  `Make a GET request to an API endpoint and print the response`,
	//  Args:  cobra.ExactArgs(1),
	//  Run: func(cmd *cobra.Command, args []string) {
	//    url := args[0]
	//    fmt.Println(url)
	//    resp, err := http.Get(url)
	//    if err != nil {
	//      fmt.Println("Error making GET request:", err)
	//      os.Exit(1)
	//    }
	//    defer resp.Body.Close()
	//
	//    body, err := io.ReadAll(resp.Body)
	//    fmt.Println(string(body))
	//    if err != nil {
	//      fmt.Println("Error reading response body:", err)
	//      os.Exit(1)
	//    }
	//
	//    var apiResponse ApiResponse
	//    err = json.Unmarshal(body, &apiResponse)
	//    if err != nil {
	//      fmt.Println("Error unmarshalling response:", err)
	//      os.Exit(1)
	//    }
	//
	//    fmt.Println("API response:", apiResponse)
	//  },
	//}
	//
	//var postCmd = &cobra.Command{
	//  Use:   "post",
	//  Short: "Make a POST request to an API endpoint",
	//  Long:  `Make a POST request to an API endpoint and print the response`,
	//  Args:  cobra.ExactArgs(1),
	//  Run: func(cmd *cobra.Command, args []string) {
	//    url := args[0]
	//    data := []byte(`{"key": "value"}`)
	//    resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	//    if err != nil {
	//      fmt.Println("Error making POST request:", err)
	//      os.Exit(1)
	//    }
	//    defer resp.Body.Close()
	//
	//    body, err := io.ReadAll(resp.Body)
	//    if err != nil {
	//      fmt.Println("Error reading response body:", err)
	//      os.Exit(1)
	//    }
	//
	//    var apiResponse ApiResponse
	//    err = json.Unmarshal(body, &apiResponse)
	//    if err != nil {
	//      fmt.Println("Error unmarshalling response:", err)
	//      os.Exit(1)
	//    }
	//
	//    fmt.Println("API response:", apiResponse.Data)
	//  },
	//}
	var vclusterListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all deployed vcluster",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			validateContext(switchContext)
			cli, err := client.NewClientWithOpts(client.FromEnv)
			if err != nil {
				panic(err)
			}
			ctx := context.Background()

			cfg := containerConfig(Vcluster, strslice.StrSlice{"list"})
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
			validateContext(switchContext)
			cli, err := client.NewClientWithOpts(client.FromEnv)
			if err != nil {
				panic(err)
			}
			ctx := context.Background()
			config := getConfig()

			cfg := containerConfig(Vcluster, strslice.StrSlice{
				"connect",
				args[0],
				"-n",
				config.Vcluster.Namespace,
				fmt.Sprintf("--server=https://%s.%s", args[0], config.Vcluster.BaseHost),
				fmt.Sprintf("--kube-config=%s.yaml", args[0]),
			})
			hostCfg := hostConfig()

			runContainer(ctx, cli, &cfg, &hostCfg)
		},
	}
	var vclusterDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Remove a vcluster",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			validateContext(switchContext)
			cli, err := client.NewClientWithOpts(client.FromEnv)
			if err != nil {
				panic(err)
			}
			ctx := context.Background()
			config := getConfig()

			cfg := containerConfig(Vcluster, strslice.StrSlice{
				"delete",
				args[0],
				"-n",
				config.Vcluster.Namespace,
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
			validateContext(switchContext)
			cli, err := client.NewClientWithOpts(client.FromEnv)
			if err != nil {
				panic(err)
			}
			ctx := context.Background()
			config := getConfig()

			cliArgs := strslice.StrSlice{}
			// Allow local values.yaml to be used to deploy cluster
			if _, err := os.Stat("vc-values.yaml"); errors.Is(err, os.ErrNotExist) {
				cliArgs = strslice.StrSlice{
					"--set",
					"ingress.enabled=true",
					"--set",
					"ingress.ingressClassName=nginx",
					"--set",
					fmt.Sprintf("ingress.host=%s.%s", args[0], config.Vcluster.BaseHost),
					"--set",
					fmt.Sprintf("syncer.extraArgs.0=--tls-san=%s.%s", args[0], config.Vcluster.BaseHost),
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
				config.Vcluster.Namespace,
				"--connect=false",
				"--upgrade",
				"--kubernetes-version",
				kubeVersion,
			}

			cfg := containerConfig(Vcluster, append(baseArgs, cliArgs...))
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
			cfg := getConfig()
			gitcfg := initGitConfig(args[0])
			err := gitBranchCheckout(gitcfg)
			if err != nil {
				panic(err)
			}

			// Init repo by templating it with copier
			if init {
				config := containerConfig(Copier, strslice.StrSlice{
					"copy",
					"-f",
					"-r",
					skelBranch,
					cfg.Git.SkeletonRepo,
					cfg.Git.SourceFolder,
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
			fmt.Printf("Branch has been templated, please review/commit/push in folder %s to trigger Hemera deployment.\n", cfg.Git.SourceFolder)
		},
	}
	var hemeraDeployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Basically does a git push. Please provide vcluster name you want to deploy to.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			repo := initGitConfig(args[0])
			commit, err := gitAddCommitPush(repo)

			if err != nil {
				panic(err)
			}
			fmt.Printf("Branch %s, was pushed to %s with commit %s\n", repo.branchName, repo.WrapperConf.Git.DeployRepo, commit)
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
	repoCmd.Flags().BoolVarP(&init, "init", "i", false, "Should we update branch from skeleton?")
	repoCmd.PersistentFlags().StringVarP(&skelBranch, "skeleton-branch", "b", "develop", "hemera-skel branch or tag to source from.")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
