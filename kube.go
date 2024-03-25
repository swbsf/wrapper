package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func getKubeConfAsString(cfg Config) string {
	if _, err := os.Stat(cfg.Runtime.KubeconfPath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Cannot find kubeconfig file, please run 'connect' cmd")
		os.Exit(1)
	} else {
		b, err := os.ReadFile(cfg.Runtime.KubeconfPath)
		if err != nil {
			panic(err)
		}
		return string(b)
	}
	return ""
}

func getCurrentContext() (string, api.Config) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{Precedence: strings.Split(os.Getenv("KUBECONFIG"), ":")},
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		}).RawConfig()
	if err != nil {
		panic(err)
	}
	return config.CurrentContext, config
}

func switchContext(kubectx string, config api.Config) error {
	//override := &clientcmd.ConfigOverrides{CurrentContext: kubectx}
	//return clientcmd.NewNonInteractiveClientConfig(config, override.CurrentContext, override, &clientcmd.ClientConfigLoadingRules{})
	config.CurrentContext = kubectx
	return clientcmd.ModifyConfig(clientcmd.NewDefaultPathOptions(), config, true)
}

func validateContext(cfg Config) api.Config {
	ctx, kubeconf := getCurrentContext()
	if !strings.Contains(ctx, cfg.Vcluster.HostContextName) {
		fmt.Printf("Please use correct kubeconfig: %s instead of %s\n", cfg.Vcluster.HostContextName, ctx)
		os.Exit(1)
	}
	return kubeconf
}

func dumpKubeconf(mainConf Config, cfg api.Config) {
	// write down vcluster host kubeconfig
	ctx := cfg.Contexts[cfg.CurrentContext]
	newConfig := api.NewConfig()
	newConfig.CurrentContext = cfg.CurrentContext
	newConfig.Contexts[cfg.CurrentContext] = ctx
	newConfig.Clusters[cfg.CurrentContext] = cfg.Clusters[ctx.Cluster]
	newConfig.AuthInfos[cfg.CurrentContext] = cfg.AuthInfos[ctx.AuthInfo]
	dump, _ := clientcmd.Write(*newConfig)
	os.WriteFile(mainConf.Runtime.KubeconfPath, dump, 0o644)
}
