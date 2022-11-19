package gopool

import "context"

type Worker struct {
	WorkID int // 当前work id
	Task   *Task
}

func NewWorker(workID int, task *Task) *Worker {
	return &Worker{
		WorkID: workID,
		Task:   task,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	return w.Task.Execute(ctx)
}
