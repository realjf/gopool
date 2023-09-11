package v2_test

import (
	"fmt"
	"testing"
	"time"

	v2 "github.com/realjf/gopool/v2"
)

func BenchmarkPoolRun(b *testing.B) {

	cases := map[string]struct {
		taskTimes int
		poolSize  int
	}{
		"task-10": {
			taskTimes: 10,
			poolSize:  10,
		},
		"task-100": {
			taskTimes: 100,
			poolSize:  100,
		},
		"task-1000": {
			taskTimes: 1000,
			poolSize:  1000,
		},
		// "task-1e4": {
		// 	taskTimes: 1e4,
		// 	poolSize:  1e4,
		// },
		// "task-1e5": {
		// 	taskTimes: 1e5,
		// 	poolSize:  1e4,
		// },
		// "task-1e6": {
		// 	taskTimes: 1e6,
		// 	poolSize:  1e4,
		// },
	}

	for name, item := range cases {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				pool := v2.NewPool(item.poolSize)
				pool.SetTaskNum(item.taskTimes)
				pool.Run()

				for j := int(0); j < item.taskTimes; j++ {
					pool.AddTask(func() {
						fmt.Println("run test...")
					})
				}
				pool.Close()
			}
		})
	}

}

func BenchmarkPoolRunParallel(b *testing.B) {

	cases := map[string]struct {
		taskTimes int
		poolSize  int
	}{
		"task-10": {
			taskTimes: 10,
			poolSize:  10,
		},
		"task-100": {
			taskTimes: 100,
			poolSize:  100,
		},
		"task-1000": {
			taskTimes: 1000,
			poolSize:  1000,
		},
		// "task-1e4": {
		// 	taskTimes: 1e4,
		// 	poolSize:  1e4,
		// },
		// "task-1e5": {
		// 	taskTimes: 1e5,
		// 	poolSize:  1e4,
		// },
		// "task-1e6": {
		// 	taskTimes: 1e6,
		// 	poolSize:  1e4,
		// },
	}

	for _, item := range cases {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				pool := v2.NewPool(item.poolSize)
				pool.SetTaskNum(item.taskTimes)
				pool.Run()

				for j := int(0); j < item.taskTimes; j++ {
					pool.AddTask(func() {
						fmt.Println("run test...")
					})
				}
				pool.Close()
			}

		})
	}
}

func TestPoolRun(t *testing.T) {
	cases := []struct {
		cap     int
		taskNum int
		name    string
	}{

		{
			name:    "5/10",
			cap:     5,
			taskNum: 10,
		},
		{
			name:    "5/100",
			cap:     5,
			taskNum: 100,
		},
		{
			name:    "10/1000",
			cap:     10,
			taskNum: 1000,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := v2.NewPool(tc.cap)
			pool.SetDebug(true)
			pool.SetTaskNum(tc.taskNum)
			pool.SetTimeout(5 * time.Second)
			pool.Run()

			for i := int(0); i < tc.taskNum; i++ {
				pool.AddTask(func() {
					fmt.Println("run test...")
				})
			}
			pool.Close()
		})
	}
}
