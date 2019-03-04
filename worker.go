package gopool

type Worker struct {
	WorkID int // 当前work id
	Task   *Task
}

func NewWorker(workID int, task *Task) *Worker {
	return &Worker{
		WorkID: workID,
		Task:   task,
	}
}

func (w *Worker) Run() error {
	err := w.Task.Execute()
	if err != nil {
		return err
	}
	return nil
}
