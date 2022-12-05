//go:build !race

package gopool

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

// func BenchmarkPoolRun(b *testing.B) {
// 	pool := NewPool(10000)
// 	pool.SetTaskNum(b.N)
// 	go func() {
// 		for i := 0; i < b.N; i++ {
// 			pool.AddTask(NewTask(taskFunc, callbackFunc, i))
// 		}
// 	}()

// 	pool.Run()
// }

type MyTask struct {
	ITask
}

func (m *MyTask) Execute() error {
	fmt.Println("my task running...")
	return nil
}

func (m *MyTask) GetResult() any {
	return 1
}

func TestNewPool(t *testing.T) {
	cases := map[string]struct {
		cap        int
		taskNum    int
		customTask ITask
	}{
		"5/10": {
			cap:     5,
			taskNum: 10,
		},
		"custom task": {
			cap:        5,
			taskNum:    10,
			customTask: &MyTask{},
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
			pool := NewPool(tc.cap)
			pool.SetDebug(true)
			pool.SetTaskNum(tc.taskNum)
			pool.SetTimeout(5 * time.Second)
			go func() {
				for i := 0; i < tc.taskNum; i++ {
					if tc.customTask != nil {
						pool.AddTask(tc.customTask)
					} else {
						pool.AddTask(NewTask(taskFunc, callbackFunc, i))
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
