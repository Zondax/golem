package runner

import (
	"errors"
)

type ErrorHandlerTask struct {
	runner       *TaskRunner
	errorHandler func(te *TaskError)
}

func NewErrorHandlerTask(runner *TaskRunner, errorHandler func(te *TaskError)) *ErrorHandlerTask {
	return &ErrorHandlerTask{
		runner:       runner,
		errorHandler: errorHandler,
	}
}

func (e *ErrorHandlerTask) Name() string {
	return "ErrorHandler"
}

func (e *ErrorHandlerTask) Start() error {
	for {
		select {
		case <-e.runner.ctx.Done():
			return nil
		case terr := <-e.runner.errCh:
			if e.runner.ctx.Err() != nil && errors.Is(terr.Err, e.runner.ctx.Err()) {
				// Do not process the error if the context is canceled
				continue
			}
			e.errorHandler(&terr)
		}
	}
}

func (e *ErrorHandlerTask) Stop() error {
	return nil
}
