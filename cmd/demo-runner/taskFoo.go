package main

import (
	"fmt"
	"time"
)

type TaskFoo struct {
	count int
}

func NewTaskFoo() *TaskFoo {
	return &TaskFoo{
		count: 0,
	}
}

func (t *TaskFoo) Name() string {
	return "taskFoo: ..."
}

func (t *TaskFoo) Start() error {
	t.count += 1
	println(fmt.Sprintf("taskFoo: %d", t.count))
	time.Sleep(500 * time.Millisecond)
	if t.count%3 == 0 {
		println("[taskFoo] throw error")
		return fmt.Errorf("taskFoo: error")
	}
	return nil
}

func (t *TaskFoo) Stop() error {
	println("taskFoo: stop")
	return nil
}
