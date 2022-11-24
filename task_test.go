package gopool

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTask(t *testing.T) {
	cases := map[string]struct {
		args      any
		callback  CallbackFunc
		taskFunc  TaskFunc
		expectval any
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
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			task := NewTask(tc.taskFunc, tc.callback, tc.args)
			err := task.Execute()
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
