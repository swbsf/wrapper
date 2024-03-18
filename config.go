package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

//const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

type Config struct {
	Vcluster struct {
		ImageName       string `yaml:"imageName"`
		Namespace       string `yaml:"namespace"`
		BaseHost        string `yaml:"baseHost"`
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
}

//func RandStringBytes(n int) string {
//	b := make([]byte, n)
//	for i := range b {
//			b[i] = letters[rand.Intn(len(letters))]
//	}
//	return string(b)
//}

func getConfig() Config {
	config := Config{
		Vcluster: struct {
			ImageName       string "yaml:\"imageName\""
			Namespace       string "yaml:\"namespace\""
			BaseHost        string "yaml:\"baseHost\""
			HostContextName string "yaml:\"hostContextName\""
		}{
			ImageName:       "vcluster",
			Namespace:       "vclusters",
			BaseHost:        "steven.env.devops.cleyrop.tech",
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
	return config
}
