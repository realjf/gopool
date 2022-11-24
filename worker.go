package gopool

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

type Worker struct {
	WorkID int // 当前work id
	Task   *Task
	done   chan bool
}

func NewWorker(workID int, task *Task) *Worker {
	return &Worker{
		WorkID: workID,
		Task:   task,
		done:   make(chan bool),
	}
}

func (w *Worker) Run(pctx context.Context) (err error) {
	debug := pctx.Value(Debug).(bool)
	timeout := pctx.Value(Timeout).(time.Duration)
	ctx, cancel := context.WithTimeout(pctx, timeout)
	defer cancel()
	go func() {
		err = w.Task.Execute()
		w.done <- true
	}()
	select {
	case <-ctx.Done():
		if debug {
			log.Infof("worker[%d]: done", w.WorkID)
		}
		return ctx.Err()
	case <-time.After(timeout + time.Second*1):
		if debug {
			log.Infof("worker[%d]: timeout", w.WorkID)
		}
		return TimecoutError
	case <-w.done:
		return
	}
}
