package main

import (
	"fmt"
	"time"
)

type TaskBar struct {
	count int
}

func NewTaskBar() *TaskBar {
	return &TaskBar{
		count: 0,
	}
}

func (t *TaskBar) Name() string {
	return "taskBar: ..."
}

func (t *TaskBar) Start() error {
	println(fmt.Sprintf("taskBar: %d", t.count))
	t.count += 1
	time.Sleep(300 * time.Millisecond)
	return nil
}

func (t *TaskBar) Stop() error {
	println("taskBar: stop")
	return nil
}
