package gopool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TwiN/go-color"
	"github.com/stretchr/testify/assert"
)

func TestNewWorker(t *testing.T) {
	cases := map[string]struct {
		args     interface{}
		callback func(interface{}) (error, interface{})
		taskFunc func(interface{}) (error, interface{})
	}{
		"success": {
			args: 1,
			taskFunc: func(args interface{}) (err error, result interface{}) {

				return nil, result
			},
			callback: func(args interface{}) (err error, result interface{}) {
				return nil, result
			},
		},
		"timeout": {
			args: 2,
			taskFunc: func(args interface{}) (err error, result interface{}) {
				time.Sleep(10 * time.Second)
				return errors.New("timeout"), result
			},
			callback: func(args interface{}) (err error, result interface{}) {
				return nil, result
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			task := NewTask(tc.taskFunc, tc.callback, tc.args)
			worker := NewWorker(1, task)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			ctx2 := context.WithValue(ctx, "debug", true)
			err := worker.Run(ctx2, time.Second*5)
			select {
			case <-ctx.Done():
				t.Log(color.InGreen("job done"))
			case <-time.After(time.Second * 6): // 比设置的超时时间延后5秒结束
				t.Log(color.InYellow("job execute timeout"))
			}
			assert.NoError(t, err)
		})
	}
}
