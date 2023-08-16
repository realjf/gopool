package gopool_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/realjf/gopool"
)

func TestNewWorker(t *testing.T) {
	cases := map[string]struct {
		args      any
		callback  gopool.CallbackFunc
		taskFunc  gopool.TaskFunc
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
			timeout:   3 * time.Second,
		},
		"timeout": {
			args: 2,
			taskFunc: func(args any) (any, error) {
				time.Sleep(5 * time.Second)
				return nil, gopool.ErrTimeout
			},
			callback: func(result any) (any, error) {
				return nil, nil
			},
			timeout: 3 * time.Second,
		},
		"no_timeout": {
			args: 3,
			taskFunc: func(args any) (any, error) {
				time.Sleep(5 * time.Second)
				return nil, gopool.ErrTimeout
			},
			callback: func(result any) (any, error) {
				return nil, nil
			},
		},
	}
	var workId int = 1
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			task := gopool.NewTask(tc.taskFunc, tc.callback, tc.args)
			var err error
			if tc.timeout > 0 {
				worker := gopool.NewWorkerWithTimeout(workId, task, tc.timeout)
				workId++
				err = worker.Run(true)
			} else {
				worker := gopool.NewWorker(workId, task)
				workId++
				err = worker.Run(true)
			}
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
		})
	}
}

func BenchmarkWorkRun(b *testing.B) {
	cases := map[string]struct {
		args      any
		callback  gopool.CallbackFunc
		taskFunc  gopool.TaskFunc
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
			timeout:   3 * time.Second,
		},
		"timeout": {
			args: 2,
			taskFunc: func(args any) (any, error) {
				time.Sleep(5 * time.Second)
				return nil, gopool.ErrTimeout
			},
			callback: func(result any) (any, error) {
				return nil, nil
			},
			timeout: 3 * time.Second,
		},
		"no_timeout": {
			args: 3,
			taskFunc: func(args any) (any, error) {
				time.Sleep(5 * time.Second)
				return nil, gopool.ErrTimeout
			},
			callback: func(result any) (any, error) {
				return nil, nil
			},
		},
	}
	var workId int = 1
	for name, tc := range cases {
		b.Run(name, func(b *testing.B) {
			task := gopool.NewTask(tc.taskFunc, tc.callback, tc.args)
			var err error
			if tc.timeout > 0 {
				worker := gopool.NewWorkerWithTimeout(workId, task, tc.timeout)
				workId++
				err = worker.Run(true)
			} else {
				worker := gopool.NewWorker(workId, task)
				workId++
				err = worker.Run(true)
			}
			if err != nil {
				if errors.Is(err, gopool.ErrTimeout) {
					b.Log(gopool.ErrTimeout.Error())
				} else if os.IsTimeout(err) {
					b.Log("IsErrTimeout:" + err.Error())
				} else {
					assert.NoError(b, err)
				}
			} else {
				assert.Equal(b, tc.expectval, task.GetResult())
			}
		})
	}
}

func TestWorkerRace(t *testing.T) {
	cases := map[string]struct {
		args      any
		callback  gopool.CallbackFunc
		taskFunc  gopool.TaskFunc
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
			timeout:   3 * time.Second,
		},
		"timeout": {
			args: 2,
			taskFunc: func(args any) (any, error) {
				time.Sleep(5 * time.Second)
				return nil, gopool.ErrTimeout
			},
			callback: func(result any) (any, error) {
				return nil, nil
			},
			timeout: 3 * time.Second,
		},
		"no_timeout": {
			args: 3,
			taskFunc: func(args any) (any, error) {
				time.Sleep(3 * time.Second)
				return nil, gopool.ErrTimeout
			},
			callback: func(result any) (any, error) {
				return nil, nil
			},
		},
	}
	var workId int = 1
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			task := gopool.NewTask(tc.taskFunc, tc.callback, tc.args)
			var err error
			if tc.timeout > 0 {
				worker := gopool.NewWorkerWithTimeout(workId, task, tc.timeout)
				workId++
				err = worker.Run(true)
			} else {
				worker := gopool.NewWorker(workId, task)
				workId++
				err = worker.Run(true)
			}
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
		})
	}
}
