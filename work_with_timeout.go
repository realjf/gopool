// #############################################################################
// # File: work_with_timeout.go                                                #
// # Project: gopool                                                           #
// # Created Date: 2023/08/13 01:44:01                                         #
// # Author: realjf                                                            #
// # -----                                                                     #
// # Last Modified: 2023/08/13 11:07:38                                        #
// # Modified By: realjf                                                       #
// # -----                                                                     #
// # Copyright (c) 2023                                                        #
// #############################################################################
package gopool

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

type WorkerWithTimeout struct {
	WorkID  int // 当前work id
	task    ITask
	timeout time.Duration
}

func NewWorkerWithTimeout(workID int, task ITask, timeout time.Duration) IWorker {
	return &WorkerWithTimeout{
		WorkID:  workID,
		task:    task,
		timeout: timeout,
	}
}

func (w *WorkerWithTimeout) GetTask() ITask {
	return w.task
}

func (w *WorkerWithTimeout) Run(debug bool) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), w.timeout)
	go w.GetTask().Execute()
	select {
	case <-ctx.Done():
		if debug {
			log.Infof("worker[%d]: done from timeout context", w.WorkID)
		}
		return TimeoutError
	case <-w.GetTask().ExecChan():
		if debug {
			log.Infof("worker[%d]: done from exec", w.WorkID)
		}
		return
	}
}
