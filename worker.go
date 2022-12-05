package gopool

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

type Worker struct {
	WorkID int // 当前work id
	task   ITask
	done   chan bool
}

func NewWorker(workID int, task ITask) *Worker {
	return &Worker{
		WorkID: workID,
		task:   task,
		done:   make(chan bool),
	}
}

func (w *Worker) GetTask() ITask {
	return w.task
}

func (w *Worker) Run(pctx context.Context) (err error) {
	debug := pctx.Value(Debug).(bool)
	timeout := pctx.Value(Timeout).(time.Duration)
	if timeout > 0 {
		ctx, cancel := context.WithTimeout(pctx, timeout)
		defer cancel()
		go func() {
			err = w.GetTask().Execute()
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
	} else {
		err = w.GetTask().Execute()
		if debug {
			log.Infof("worker[%d]: done", w.WorkID)
		}
		return err
	}
}
