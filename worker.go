package gopool

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type IWorker interface {
	GetTask() ITask
	Run(debug bool) (err error)
}

type Worker struct {
	WorkID int // 当前work id
	task   ITask
	done   chan bool
	lock   sync.Mutex
}

func NewWorker(workID int, task ITask) IWorker {
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

func (w *Worker) Run(debug bool) (err error) {
	err = w.GetTask().Execute()
	if debug {
		log.Infof("worker[%d]: done from no timeout context", w.WorkID)
	}
	return
}
