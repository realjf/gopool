package gopool

import (
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

//go:skip
func TestNewPool(t *testing.T) {
	cases := map[string]struct {
		cap     int
		taskNum int
	}{
		"5/10": {
			cap:     5,
			taskNum: 10,
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
			pool.SetTimeout(2 * time.Second)
			go func() {
				for i := 0; i < tc.taskNum; i++ {
					pool.AddTask(NewTask(taskFunc, callbackFunc, i))
					t.Log("task:" + strconv.Itoa(i))
				}
			}()

			pool.Run()
		})
	}
}

func taskFunc(args interface{}) (interface{}, error) {
	//fmt.Println("task ", args, "completed")
	_ = 1 + 1
	return args, nil
}

func callbackFunc(result interface{}) (interface{}, error) {
	// 处理
	//fmt.Println("callback completed [", result, "]")
	return result, nil
}
