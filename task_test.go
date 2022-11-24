package gopool

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func taskFuncv1(args interface{}) (error, interface{}) {
	fmt.Println("task ", args, "completed")
	return nil, args
}

func callbackFuncv1(result interface{}) (error, interface{}) {
	// 处理结果
	fmt.Println("callback completed [", result, "]")
	return nil, result
}

func TestNewTask(t *testing.T) {
	cases := map[string]struct {
		args     interface{}
		callback CallbackFunc
		taskFunc TaskFunc
	}{
		"success": {
			args: 1,
			taskFunc: func(args interface{}) (interface{}, error) {
				_ = 1 + 1
				return nil, nil
			},
			callback: func(result interface{}) (interface{}, error) {
				return nil, nil
			},
		},
		"timeout": {
			args: 2,
			taskFunc: func(args interface{}) (interface{}, error) {
				time.Sleep(10 * time.Second)
				return nil, TimecoutError
			},
			callback: func(result interface{}) (interface{}, error) {
				return nil, nil
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			task := NewTask(tc.taskFunc, tc.callback, tc.args)
			err := task.Execute()
			if err != nil {
				assert.ErrorIs(t, err, TimecoutError)
			}

		})
	}
}
