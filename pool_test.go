//go:build !race

package gopool_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/realjf/gopool"
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
				pool := gopool.NewPool(item.poolSize)
				pool.SetTaskNum(item.taskTimes)

				go func() {
					for j := 0; j < item.taskTimes; j++ {
						pool.AddTask(gopool.NewTask(taskFunc, callbackFunc, j))
					}
				}()
				pool.Run()
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
				pool := gopool.NewPool(item.poolSize)
				pool.SetTaskNum(item.taskTimes)

				go func() {
					for j := 0; j < item.taskTimes; j++ {
						pool.AddTask(gopool.NewTask(taskFunc, callbackFunc, j))
					}
				}()
				pool.Run()
			}

		})
	}
}

type MyTask struct {
	gopool.ITask
	ch <-chan error
}

func (m *MyTask) Execute() error {
	fmt.Println("my task running...")
	return nil
}

func (m *MyTask) GetResult() any {
	return 1
}

func (m *MyTask) ExecChan() <-chan error {
	return m.ch
}

func TestPoolRun(t *testing.T) {
	cases := map[string]struct {
		cap        int
		taskNum    int
		customTask gopool.ITask
	}{
		"5/10": {
			cap:     5,
			taskNum: 10,
		},
		"custom task": {
			cap:     5,
			taskNum: 10,
			customTask: &MyTask{
				ch: make(chan error),
			},
		},
		// "5/100": {
		// 	cap:     5,
		// 	taskNum: 100,
		// },
		// "10/1000": {
		// 	cap:     10,
		// 	taskNum: 1000,
		// },
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			pool := gopool.NewPool(tc.cap)
			pool.SetDebug(true)
			pool.SetTaskNum(tc.taskNum)
			pool.SetTimeout(5 * time.Second)
			go func() {
				for i := 0; i < tc.taskNum; i++ {
					if tc.customTask != nil {
						pool.AddTask(tc.customTask)
					} else {
						pool.AddTask(gopool.NewTask(taskFunc, callbackFunc, i))
					}

					t.Log("task:" + strconv.Itoa(i))
				}
			}()

			pool.Run()
			for _, v := range pool.GetResult() {
				t.Logf("result: %v", v)
			}

		})
	}
}

func taskFunc(args any) (any, error) {
	a := args.(int) + args.(int)
	return a, nil
}

func callbackFunc(result any) (any, error) {
	// 处理
	//fmt.Println("callback completed [", result, "]")
	return result, nil
}
