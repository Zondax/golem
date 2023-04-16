package main

import "github.com/zondax/golem/pkg/runner"

func main() {
	println("Task runner demo")

	r := runner.NewRunner()

	// First we define some very simple error handler
	// we will shut down the runner after two errors
	errCount := 0
	errorHandler := func(te *runner.TaskError) {
		println("[Handler] ", te.Err.Error())
		errCount += 1
		if errCount == 2 {
			println("[Handler] Shutting down...")
			r.Shutdown()
		}
	}

	r.AddTask(NewTaskFoo())
	r.AddTask(NewTaskBar())
	r.AddErrorHandler(errorHandler)

	// Now start all the tasks
	r.Start()

	// Wait for all tasks to finish
	err := r.Wait()
	if err != nil {
		println("Error: ", err.Error())
		return
	}
}
