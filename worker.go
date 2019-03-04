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

func (w *Worker) Run(ctx context.Context, ch chan bool) error {
	defer func() {
		ch <- true
	}()

	err := make(chan error)
	go func(err chan error) {
		err <- w.Task.Execute()
	}(err)

	var err1 error

	select {
		case <-ctx.Done():
		case err1 = <- err:
			return err1
	}
	return nil
}
