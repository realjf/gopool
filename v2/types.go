package v2

import (
	"sync"
)

type workerMap struct {
	lock    sync.Mutex
	workers map[int64]bool
}

func newWorkerMap() *workerMap {
	return &workerMap{
		lock:    sync.Mutex{},
		workers: make(map[int64]bool),
	}
}

func (w *workerMap) Delete(gId int64) {
	w.lock.Lock()
	defer w.lock.Unlock()
	delete(w.workers, gId)
}

func (w *workerMap) Store(gId int64, val bool) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.workers[gId] = val
}

func (w *workerMap) LoadOrStore(gId int64, val bool) bool {
	w.lock.Lock()
	defer w.lock.Unlock()
	if v, ok := w.workers[gId]; ok {
		return v
	} else {
		w.workers[gId] = val
		return val
	}
}

func (w *workerMap) GetBusyNum() int {
	w.lock.Lock()
	defer w.lock.Unlock()
	var num int = 0
	for _, v := range w.workers {
		if v {
			num++
		}
	}
	return num
}

func (w *workerMap) GetIdleNum() int {
	w.lock.Lock()
	defer w.lock.Unlock()
	var num int = 0
	for _, v := range w.workers {
		if !v {
			num++
		}
	}
	return num
}

func (w *workerMap) GetLength() int {
	w.lock.Lock()
	defer w.lock.Unlock()
	return int(len(w.workers))
}
