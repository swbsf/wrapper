package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Vcluster struct {
		ImageName       string `yaml:"imageName"`
		Namespace       string `yaml:"namespace"`
		HostFqdn        string `yaml:"baseFqdn"`
		ChildBaseFqdn   string `yaml:"childBaseFqdn"`
		HostContextName string `yaml:"hostContextName"`
	} `yaml:"vcluster"`
	Git struct {
		SourceFolder     string `yaml:"sourceFolder"`
		SkeletonRepo     string `yaml:"skeletonRepo"`
		DeployRepo       string `yaml:"deployRepo"`
		CommitOwnerName  string `yaml:"commitOwnerName"`
		CommitOwnerEmail string `yaml:"commitOwnerEmail"`
		Username         string `yaml:"username"`
		Password         string `yaml:"password"`
	} `yaml:"git"`
	Runtime struct {
		VclusterName		 string `yaml:"vclusterName"`
		SourceFolder     string `yaml:"sourceFolder"`
		KubeconfPath     string `yaml:"kubeconfPath"`
	} `yaml:"runtime"`
}

func setRuntimeConfig(cfg *Config, vclusterName string) {
	cfg.Runtime.VclusterName = vclusterName
	cfg.Runtime.SourceFolder = cfg.Git.SourceFolder + "/" + vclusterName
	cfg.Runtime.KubeconfPath = cfg.Git.SourceFolder + "/kube-" + vclusterName + ".yaml"
}

func getConfig(vclusterName string) Config {
	config := Config{
		Vcluster: struct{ImageName string "yaml:\"imageName\""; Namespace string "yaml:\"namespace\""; HostFqdn string "yaml:\"baseFqdn\""; ChildBaseFqdn string "yaml:\"childBaseFqdn\""; HostContextName string "yaml:\"hostContextName\""}{
			ImageName:       "vcluster",
			Namespace:       "vclusters",
			HostFqdn:        "steven.env.devops.cleyrop.tech",
			ChildBaseFqdn:   "changeme.env.devops.cleyrop.tech",
			HostContextName: "",
		},
		Git: struct {
			SourceFolder     string "yaml:\"sourceFolder\""
			SkeletonRepo     string "yaml:\"skeletonRepo\""
			DeployRepo       string "yaml:\"deployRepo\""
			CommitOwnerName  string "yaml:\"commitOwnerName\""
			CommitOwnerEmail string "yaml:\"commitOwnerEmail\""
			Username         string "yaml:\"username\""
			Password         string "yaml:\"password\""
		}{
			SourceFolder:     fmt.Sprintf("%s/.wrapper", os.Getenv("HOME")),
			SkeletonRepo:     "https://gitlab.com/cleyrop-org/cleyrop-infra/skel/hemera/hemera.git",
			DeployRepo:       "https://gitlab.com/cleyrop-org/cleyrop-infra/internal/team-environments/developers/hemera-developers.git",
			CommitOwnerName:  "vclusterWrapper",
			CommitOwnerEmail: "bot@cleyrop.com",
			Username:         "oauth2",
			Password:         os.Getenv("GITLAB_ACCESS_TOKEN"),
		},
	}
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return config
	}

	// Mettre à jour la configuration avec les valeurs du fichier YAML
	err = viper.Unmarshal(&config)
	if err != nil {
		fmt.Printf("Failed updating config : %v", err)
		return config
	}
	if config.Git.Password == "" {
		fmt.Println("No credentials found. Please set GITLAB_ACCESS_TOKEN env var or git.password in config.yaml to continue...")
		os.Exit(1)
	}
	setRuntimeConfig(&config,vclusterName)
	return config
}
