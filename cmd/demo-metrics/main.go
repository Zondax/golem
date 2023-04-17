package main

import (
	"github.com/zondax/golem/pkg/cli"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/runner"
)

func main() {
	println("[Demo] Microservice example")

	cli.InitGlobalLogger()

	r := runner.NewRunner()

	r.AddTask(metrics.NewTaskMetrics("/metrics", "9090"))

	// Now start all the tasks
	r.StartAndWait()
}
