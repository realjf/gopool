package gopool

import (
	"testing"
	"fmt"
)


func taskFuncv1(args interface{}) (error, interface{}) {
	fmt.Println("task ", args, "completed")
	return nil, args
}

func callbackFuncv1(result interface{}) (error, interface{}) {
	// 处理
	fmt.Println("callback completed [", result, "]")
	return nil, result
}

func TestNewTask(t *testing.T) {
	for i := 1; i<= 100000; i++{
		task := NewTask(taskFuncv1, callbackFuncv1, i)
		err := task.Execute()
		if err != nil {
			t.Logf(err.Error())
		}
	}
	t.Fatalf("com")
}


