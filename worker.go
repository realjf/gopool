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
	debug := ctx.Value("debug").(bool)
	for {
		select {
		case <-ctx.Done():
			if debug {
				log.Infof("worker[%d]: done", w.WorkID)
			}
			return err
		default:
			if !w.running {
				if debug {
					log.Infof("worker:[%d]: working...", w.WorkID)
				}
				w.running = true
				err = w.Task.Execute(ctx)
			}
		}

	}
}
