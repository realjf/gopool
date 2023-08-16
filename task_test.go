package gopool_test

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/realjf/gopool"
)

func TestNewTask(t *testing.T) {
	cases := map[string]struct {
		args      any
		callback  gopool.CallbackFunc
		taskFunc  gopool.TaskFunc
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
			task := gopool.NewTask(tc.taskFunc, tc.callback, tc.args)
			go task.Execute()
			err := <-task.ExecChan()
			if err != nil {
				if errors.Is(err, gopool.ErrTimeout) {
					t.Log(gopool.ErrTimeout.Error())
				} else if os.IsTimeout(err) {
					t.Log("IsErrTimeout:" + err.Error())
				} else {
					assert.NoError(t, err)
				}
			} else {
				assert.Equal(t, tc.expectval, task.GetResult())
			}
			t.Log("done")
		})
	}
}
