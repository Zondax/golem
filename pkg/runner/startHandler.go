package runner

type StartTask struct {
	runner *TaskRunner
}

func newStartTask(runner *TaskRunner) *StartTask {
	return &StartTask{
		runner: runner,
	}
}

func (e *StartTask) Name() string {
	return "ErrorHandler"
}

func (e *StartTask) Start() error {
	for {
		select {
		case <-e.runner.ctx.Done():
			return nil
		case t := <-e.runner.tasksCh:
			e.runner.runTask(t)
		}
	}
}

func (e *StartTask) Stop() error {
	return nil
}
