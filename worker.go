package gopool

import (
	"context"

	log "github.com/sirupsen/logrus"
)

type Worker struct {
	WorkID  int // 当前work id
	Task    *Task
	running bool
}

func NewWorker(workID int, task *Task) *Worker {
	return &Worker{
		WorkID:  workID,
		Task:    task,
		running: false,
	}
}

func (w *Worker) Run(ctx context.Context) (err error) {
	for {
		select {
		case <-ctx.Done():
			log.Infof("worker[%d]: done", w.WorkID)
			return err
		default:
			log.Infof("worker:[%d]: working...", w.WorkID)
			if !w.running {
				w.running = true
				err = w.Task.Execute(ctx)
			}
		}

	}
}
