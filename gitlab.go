package main

import (
	"strings"

	gitlab "github.com/xanzy/go-gitlab"
)

type GitlabVar struct {
	Name  string
	Value string
}

type GitlabClient struct {
	client *gitlab.Client
	conf   GitConfig
	projid int
}

func gitlabClient(gitcfg GitConfig) GitlabClient {
	git, _ := gitlab.NewClient(gitcfg.WrapperConf.Git.Password)

	// Extract repo name from Full URL > first split on "/", then on ".git" and take first element.
	URLslice := strings.Split(gitcfg.WrapperConf.Git.DeployRepo, "/")
	p := strings.Split(URLslice[len(URLslice)-1], ".git")[0]
	proj, _, err := git.Projects.ListProjects(&gitlab.ListProjectsOptions{
		Search: &p,
	})
	if err != nil {
		panic(err)
	}

	return GitlabClient{
		client: git,
		conf:   gitcfg,
		projid: proj[0].ID,
	}
}

func (g *GitlabClient) getGitlabVar(v GitlabVar) *gitlab.ProjectVariable {
	varContent, resp, err := g.client.ProjectVariables.GetVariable(g.projid, v.Name, &gitlab.GetProjectVariableOptions{
		Filter: &gitlab.VariableFilter{
			EnvironmentScope: g.conf.branchNameSlug,
		},
	})

	if err != nil && resp.StatusCode != 404 {
		panic(err)
	}

	return varContent
}

func (g *GitlabClient) gitlabVarExists(v GitlabVar) bool {
	return g.getGitlabVar(v) != nil
}

func (g *GitlabClient) updateVariable(v GitlabVar) error {
	raw := true
	protected := true
	_, _, err := g.client.ProjectVariables.UpdateVariable(g.projid, v.Name, &gitlab.UpdateProjectVariableOptions{
		Value:            &v.Value,
		EnvironmentScope: &g.conf.branchNameSlug,
		Filter: &gitlab.VariableFilter{
			EnvironmentScope: g.conf.branchNameSlug,
		},
		Raw:       &raw,
		Protected: &protected,
	})
	return err
}

func (g *GitlabClient) createVariable(v GitlabVar) error {
	raw := true
	protected := true
	options := &gitlab.CreateProjectVariableOptions{
		Key:              &v.Name,
		Value:            &v.Value,
		EnvironmentScope: &g.conf.branchNameSlug,
		Raw:              &raw,
		Protected:        &protected,
	}
	_, _, err := g.client.ProjectVariables.CreateVariable(g.projid, options)
	return err
}

func (g *GitlabClient) gitlabAddOrUpdateVariable(v GitlabVar) error {
	var err error
	if g.gitlabVarExists(v) {
		err = g.updateVariable(v)
	} else {
		err = g.createVariable(v)
	}
	return err
}
