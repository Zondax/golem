package runner

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Task interface {
	Name() string
	Start() error
	Stop() error
}

const MaximumPendingTasks = 1000

type TaskError struct {
	Task *Task
	Err  error
}

type TaskRunner struct {
	ctx     context.Context
	cancel  context.CancelFunc
	tasks   errgroup.Group
	errCh   chan TaskError
	tasksCh chan Task
}

func NewRunner() *TaskRunner {
	ctx, cancel := context.WithCancel(context.Background())

	return &TaskRunner{
		ctx:     ctx,
		cancel:  cancel,
		tasks:   errgroup.Group{},
		errCh:   make(chan TaskError),
		tasksCh: make(chan Task, MaximumPendingTasks),
	}
}

func (tr *TaskRunner) AddTask(task Task) {
	tr.tasksCh <- task
}

func (tr *TaskRunner) AddErrorHandler(errorHandler func(te *TaskError)) {
	tr.AddTask(NewErrorHandlerTask(tr, errorHandler))
}

func (tr *TaskRunner) Start() {
	tr.runTask(newStartTask(tr))
}

func (tr *TaskRunner) sendError(te TaskError) {
	select {
	case tr.errCh <- te:
		break
	default:
		// FIXME: log error at least? Nobody is registered to receive errors
		fmt.Println("No receiver ready, error not sent!")
	}
}

func (tr *TaskRunner) runTask(task Task) {
	tr.tasks.Go(func() error {
		for {
			select {
			case <-tr.ctx.Done():
				_ = task.Stop()
				return tr.ctx.Err()

			case <-time.After(1 * time.Second):
				// FIXME: are we sure we want to loop and keep retrying?
				// Send some event so we are aware of this?
				err := task.Start()
				if err != nil {
					tr.sendError(TaskError{Task: &task, Err: err})
				}
			}
		}
	})
}

func (tr *TaskRunner) Wait() error {
	// FIXME: maybe rename Wait to Start? or StartAndWait?
	tr.waitForShutdownSignals()
	err := tr.tasks.Wait()
	return err
}

func (tr *TaskRunner) StartAndWait() {
	tr.Start()
	err := tr.Wait()
	if err != nil {
		fmt.Println("Error waiting for tasks to finish: ", err)
	}
}

func (tr *TaskRunner) waitForShutdownSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		tr.Shutdown()
		break
	case <-tr.ctx.Done():
		break
	}
}

func (tr *TaskRunner) Shutdown() {
	tr.cancel()
}
