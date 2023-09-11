package gopool_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/realjf/gopool/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewWorker(t *testing.T) {
	cases := map[string]struct {
		timeout time.Duration
	}{
		"timeout": {

			timeout: 3 * time.Second,
		},
		"no_timeout": {},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			task := func() {
				fmt.Println("run test...")
			}
			var err error
			if tc.timeout > 0 {
				gopool.NewWorkerWithTimeout(tc.timeout).Run(task)

			} else {
				gopool.NewWorker().Run(task)
			}
			if err != nil {
				assert.NoError(t, err)
			}
		})
	}
}

func BenchmarkWorkRun(b *testing.B) {
	cases := map[string]struct {
		timeout time.Duration
	}{
		"timeout": {

			timeout: 3 * time.Second,
		},
		"no_timeout": {},
	}
	for name, tc := range cases {
		b.Run(name, func(b *testing.B) {
			task := func() {
				fmt.Println("run test...")
			}
			var err error
			if tc.timeout > 0 {
				gopool.NewWorkerWithTimeout(tc.timeout).Run(task)

			} else {
				gopool.NewWorker().Run(task)
			}
			if err != nil {
				assert.NoError(b, err)
			}
		})
	}
}
