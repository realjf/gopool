// #############################################################################
// # File: work_with_timeout.go                                                #
// # Project: gopool                                                           #
// # Created Date: 2023/08/13 01:44:01                                         #
// # Author: realjf                                                            #
// # -----                                                                     #
// # Last Modified: 2023/09/11 11:01:33                                        #
// # Modified By: realjf                                                       #
// # -----                                                                     #
// # Copyright (c) 2023                                                        #
// #############################################################################
package v2

import (
	"context"
	"time"
)

type workerWithTimeout struct {
	timeout time.Duration
	ch      chan bool
}

func NewWorkerWithTimeout(timeout time.Duration) IWorker {
	return &workerWithTimeout{
		timeout: timeout,
		ch:      make(chan bool),
	}
}

func (w *workerWithTimeout) Run(f func()) {
	ctx, cancel := context.WithTimeout(context.Background(), w.timeout)
	defer cancel()
	go func() {
		f()
		w.ch <- true
	}()
	select {
	case <-ctx.Done():
		return
	case <-w.ch:
		return
	}
}
