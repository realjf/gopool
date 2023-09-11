package gopool

type IWorker interface {
	Run(f func())
}

type worker struct {
	f func()
}

func NewWorker() IWorker {
	return &worker{}
}

func (w *worker) Run(f func()) {
	f()
}
