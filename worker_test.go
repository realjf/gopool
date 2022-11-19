package gopool

import (
	"context"
	"testing"
	"time"

	"github.com/TwiN/go-color"
	log "github.com/sirupsen/logrus"
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
				return nil, result
			},
			callback: func(args interface{}) (err error, result interface{}) {
				return nil, result
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			go func(args interface{}) {
				task := NewTask(tc.taskFunc, tc.callback, args)
				worker := NewWorker(1, task)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				err := worker.Run(ctx)
				select {
				case <-ctx.Done():
					log.Println(color.InGreen("job done"))
				case <-time.After(time.Second * 6): // 比设置的超时时间延后5秒结束
					log.Println(color.InYellow("job execute timeout"))
				}
				assert.NoError(t, err)

			}(tc.args)
		})
	}
}
