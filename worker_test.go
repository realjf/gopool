package gopool

import (
	"context"
	"testing"
	"time"

	"github.com/TwiN/go-color"
	"github.com/stretchr/testify/assert"
)

func TestNewWorker(t *testing.T) {
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
			worker := NewWorker(1, task)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ctx2 := context.WithValue(ctx, Debug, true)
			ctx2 = context.WithValue(ctx2, Timeout, time.Second*5)
			err := worker.Run(ctx2)
			select {
			case <-ctx.Done():
				t.Log(color.InGreen("job done"))
			case <-time.After(time.Second * 6): // 比设置的超时时间延后5秒结束
				t.Log(color.InYellow("job execute timeout"))
			}
			if err != nil {
				assert.ErrorIs(t, err, TimecoutError)
			}
		})
	}
}
