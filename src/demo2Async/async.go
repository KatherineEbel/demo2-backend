package demo2Async

import (
	"fmt"
)

type Task interface {
	Perform() error
}

type TaskWorker struct {
	Id              int
	TaskChannel     chan Task
	TaskWorkerQueue chan chan Task
}

var (
	TaskQueue       chan Task
	TaskWorkerQueue chan chan Task
)

func init() {
	TaskQueue = make(chan Task, 108)
}

func NewTaskWorker(id int, queue chan chan Task) TaskWorker {
	return TaskWorker{
		Id:              id,
		TaskChannel:     make(chan Task),
		TaskWorkerQueue: queue,
	}
}

func (w *TaskWorker) Start() {
	go func() {
		for {
			w.TaskWorkerQueue <- w.TaskChannel
			select {
			case task := <-w.TaskChannel:
				_ = fmt.Sprintf("Async task worker %d performing task. \n", w.Id)
				if err := task.Perform(); err != nil {
					fmt.Println("Worker perform task error for worker", w.Id)
				}
			}
		}
	}()
}

func StartTaskDispatcher(taskWorkerSize int) {
	TaskWorkerQueue = make(chan chan Task, taskWorkerSize)
	for i := 0; i < taskWorkerSize; i++ {
		fmt.Println("Starting task worker #", i+1)
		worker := NewTaskWorker(i+1, TaskWorkerQueue)
		worker.Start()
	}
	go func() {
		for {
			select {
			case task := <-TaskQueue:
				go func() {
					channel := <-TaskWorkerQueue
					channel <- task
				}()

			}
		}
	}()
}
