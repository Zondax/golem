package main

import (
	"github.com/zondax/golem/pkg/cli"
	"github.com/zondax/golem/pkg/metrics"
)

func main() {
	appSettings := cli.AppSettings{
		Name:        "golem test",
		Description: "some fake tool",
		ConfigPath:  "$HOME/.golem/",
		EnvPrefix:   "golem",
		GitVersion:  GitVersion,
		GitRevision: GitRevision,
	}

	// Define application level features
	cli := cli.New[cli.ConfigMock](appSettings)
	defer cli.Close()

	metrics.StartMetricsServer("metrics", "8080")

	cli.Run()
}
