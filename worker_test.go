package gopool

import (
	"fmt"
	"testing"
)

func taskFuncv2(args interface{}) (error, interface{}) {
	fmt.Println("task ", args, "completed")
	return nil, args
}

func callbackFuncv2(result interface{}) (error, interface{}) {
	// 处理
	fmt.Println("callback completed [", result, "]")
	return nil, result
}

func TestNewWorker(t *testing.T) {
	for i := 1; i <= 10000; i++ {
		go func(taskId int) {
			task := NewTask(taskFuncv2, callbackFuncv2, taskId)
			worker := NewWorker(taskId, task)
			worker.Run()
			t.Logf("%v", worker.Task.Result)
		}(i)
	}

	t.Fatalf("com")
}
