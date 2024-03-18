package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type GitConfig struct {
	branchName     string
	branchNameSlug string
	branchNameRef  plumbing.ReferenceName
	remoteName     string
	repoClient     *git.Repository
	Username       string
	Password       string
	WrapperConf    Config
}

func gitClone(gitcfg GitConfig) *git.Repository {
	r, err := git.PlainClone(gitcfg.WrapperConf.Git.SourceFolder, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: gitcfg.Username, // yes, this can be anything except an empty string
			Password: gitcfg.Password,
		},
		URL:      gitcfg.WrapperConf.Git.DeployRepo,
		Progress: os.Stdout,
	})
	if err != nil {
		panic(err)
	}
	return r
}

func gitBranchCheckout(gitcfg GitConfig) error {
	br, _ := gitcfg.repoClient.Branch(gitcfg.branchName)

	if br == nil {
		fmt.Printf("Branch %s does not exist, creating.\n", gitcfg.branchName)
		headRef, _ := gitcfg.repoClient.Head()
		ref := plumbing.NewHashReference(gitcfg.branchNameRef, headRef.Hash())
		err := gitcfg.repoClient.Storer.SetReference(ref)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Switching on branch %s", gitcfg.branchName)
	w, _ := gitcfg.repoClient.Worktree()
	err := w.Checkout(&git.CheckoutOptions{
		Branch: gitcfg.branchNameRef,
	})

	if err != nil {
		return err
	}
	return nil
}

func gitAddCommitPush(gitcfg GitConfig) (plumbing.Hash, error) {
	w, _ := gitcfg.repoClient.Worktree()
	_, err := w.Add(".")
	if err != nil {
		return plumbing.ZeroHash, err
	}

	commit, err := w.Commit("Deploying Hemera", &git.CommitOptions{
		Author: &object.Signature{
			Name:  gitcfg.WrapperConf.Git.CommitOwnerName,
			Email: gitcfg.WrapperConf.Git.CommitOwnerEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return commit, err
	}

	remote, _ := gitcfg.repoClient.Remote(gitcfg.remoteName)

	err = remote.Push(&git.PushOptions{
		RemoteName: gitcfg.remoteName,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("%s:%s", gitcfg.branchNameRef, gitcfg.branchNameRef)),
		},
		Auth: &http.BasicAuth{
			Username: gitcfg.Username,
			Password: gitcfg.Password,
		},
	})

	return commit, err
}

func initGitConfig(vclusterName string) GitConfig {
	branchName := fmt.Sprintf("env/%s", vclusterName)
	cfg := getConfig()
	gitcfg := GitConfig{
		branchName:     branchName,
		branchNameSlug: strings.ReplaceAll(branchName, "/", "-"),
		branchNameRef:  plumbing.NewBranchReferenceName(branchName),
		remoteName:     "origin",
		Username:       "oauth2",
		Password:       os.Getenv("GITLAB_ACCESS_TOKEN"),
		WrapperConf:    cfg,
	}
	repoClient, err := git.PlainOpen(cfg.Git.SourceFolder)
	gitcfg.repoClient = repoClient
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			gitcfg.repoClient = gitClone(gitcfg)
		} else {
			panic(err)
		}
	}
	return gitcfg
}
