package main

import (
	"github.com/zondax/golem/pkg/constants"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/runner"
)

func main() {
	println("[Demo] Microservice example")

	logger.InitLogger(logger.Config{Level: constants.DebugLevel})
	r := runner.NewRunner()

	r.AddTask(metrics.NewTaskMetrics("/metrics", "9090", "demo"))

	// Now start all the tasks
	r.StartAndWait()
}
