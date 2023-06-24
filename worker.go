package gopool

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Worker struct {
	WorkID int // 当前work id
	task   ITask
	done   chan bool
	lock   sync.Mutex
}

func NewWorker(workID int, task ITask) *Worker {
	return &Worker{
		WorkID: workID,
		task:   task,
		done:   make(chan bool),
		lock:   sync.Mutex{},
	}
}

func (w *Worker) GetTask() ITask {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.task
}

func (w *Worker) Run(debug bool, timeout time.Duration) (err error) {
	if timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		go func() {
			err = w.GetTask().Execute()
			w.done <- true
		}()
		select {
		case <-ctx.Done():
			if debug {
				log.Infof("worker[%d]: done from timeout context", w.WorkID)
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
		go func() {
			err = w.GetTask().Execute()
			w.done <- true
		}()
		<-w.done
		if debug {
			log.Infof("worker[%d]: done from no timeout context", w.WorkID)
		}
		return

	}
}
