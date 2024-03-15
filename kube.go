package main

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

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

func validateContext(forceContext bool) {
	cfg := getConfig()
	ctx, _ := getCurrentContext()
	if ! strings.Contains(ctx, cfg.Vcluster.HostContextName){
		fmt.Printf("Please use correct kubeconfig: %s instead of %s\n", cfg.Vcluster.HostContextName, ctx)
		os.Exit(1)
	}
}
