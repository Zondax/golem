package main

import (
	"github.com/zondax/golem/pkg/constants"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/runner"
)

func main() {
	println("[Demo] Panic handler")

	logger.InitLogger(logger.Config{Level: constants.DebugLevel})
	r := runner.NewRunner()

	// This will panic
	r.AddTask(metrics.NewTaskMetrics("BADURL", "8080", "demo"))

	// Now start all the tasks
	r.StartAndWait()
}
