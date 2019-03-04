package gopool

import "testing"

func TestNewWorker(t *testing.T) {
	for i := 1; i<= 10000; i++ {
		go func(taskId int) {
			task := NewTask(taskFuncv1, callbackFuncv1, taskId)
			worker := NewWorker(taskId, task)
			worker.Task.Execute()
		}(i)
	}
}

