package gopool_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	v2 "github.com/realjf/gopool/v2"
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
				v2.NewWorkerWithTimeout(tc.timeout).Run(task)

			} else {
				v2.NewWorker().Run(task)
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
				v2.NewWorkerWithTimeout(tc.timeout).Run(task)

			} else {
				v2.NewWorker().Run(task)
			}
			if err != nil {
				assert.NoError(b, err)
			}
		})
	}
}
