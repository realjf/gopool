package gopool

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewWorker(t *testing.T) {
	cases := map[string]struct {
		args      any
		callback  CallbackFunc
		taskFunc  TaskFunc
		expectval any
		timeout   time.Duration
	}{
		"success": {
			args: 1,
			taskFunc: func(args any) (any, error) {
				a := args.(int) + args.(int)
				return a, nil
			},
			callback: func(result any) (any, error) {
				return result, nil
			},
			expectval: 2,
			timeout:   5 * time.Second,
		},
		"timeout": {
			args: 2,
			taskFunc: func(args any) (any, error) {
				time.Sleep(10 * time.Second)
				return nil, TimecoutError
			},
			callback: func(result any) (any, error) {
				return nil, nil
			},
			timeout: 5 * time.Second,
		},
		"no_timeout": {
			args: 3,
			taskFunc: func(args any) (any, error) {
				time.Sleep(12 * time.Second)
				return nil, TimecoutError
			},
			callback: func(result any) (any, error) {
				return nil, nil
			},
		},
	}
	var workId int = 1
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			task := NewTask(tc.taskFunc, tc.callback, tc.args)
			worker := NewWorker(workId, task)
			workId++
			err := worker.Run(true, tc.timeout)
			if err != nil {
				if errors.Is(err, TimecoutError) {
					t.Log(TimecoutError.Error())
				} else if errors.Is(err, context.DeadlineExceeded) {
					t.Log(context.DeadlineExceeded.Error())
				} else if os.IsTimeout(err) {
					t.Log("IsTimeoutError:" + err.Error())
				} else {
					assert.NoError(t, err)
				}
			} else {
				assert.Equal(t, tc.expectval, task.GetResult())
			}
		})
	}
}
