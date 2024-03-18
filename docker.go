package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type Program string

const (
	Vcluster Program = "vcluster"
	Copier           = "copier"
	Kubectl          = "kubectl"
)

func containerConfig(entrypoint Program, cmd strslice.StrSlice) container.Config {
	config := getConfig()
	return container.Config{
		Image: config.Vcluster.ImageName,
		Env: []string{
			fmt.Sprintf("KUBECONFIG=%s", os.Getenv("KUBECONFIG")),
			fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
		},
		Entrypoint: strslice.StrSlice{
			string(entrypoint),
		},
		Cmd:        cmd,
		Tty:        false,
		WorkingDir: os.Getenv("PWD"),
	}
}

func hostConfig() container.HostConfig {
	pwd := os.Getenv("PWD")
	home := os.Getenv("HOME")
	return container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s/:%s/", home, home),
			fmt.Sprintf("%s/:%s/", pwd, pwd),
		},
	}
}

func waitForContainer(ctx context.Context, ID string, cli *client.Client) {
	// wait for container to finish
	statusCh, errCh := cli.ContainerWait(ctx, ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}
}

func runContainer(ctx context.Context, cli *client.Client, containerCfg *container.Config, hostCfg *container.HostConfig) {
	resp, err := cli.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, "")
	if err != nil {
		panic(err)
	}
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		panic(err)
	}

	waitForContainer(ctx, resp.ID, cli)
	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		panic(err)
	}
	// print container logs to stdout
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}
