package main

import (
	"github.com/zondax/golem/pkg/cli"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/runner"
)

func main() {
	println("[Demo] Panic handler")

	_, _ = cli.InitGlobalLogger(cli.DebugLevel)

	r := runner.NewRunner()

	// This will panic
	r.AddTask(metrics.NewTaskMetrics("BADURL", "8080"))

	// Now start all the tasks
	r.StartAndWait()
}
