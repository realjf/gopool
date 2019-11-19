package gopool

import (
	"testing"
)

func BenchmarkPoolRun(b *testing.B) {
	pool := NewPool(100)
	pool.SetTaskNum(b.N)
	go func() {
		for i := 0; i < b.N; i++ {
			pool.AddTask(NewTask(taskFunc, callbackFunc, i))
		}
	}()

	pool.Run()

	//b.Logf("%v", pool.GetResult())
	//b.Errorf("program total run time is %f seconds", pool.GetRunTime())
}

//go:skip
func TestNewPool(t *testing.T) {

	//go func() {
	//	http.HandleFunc("/goroutines", func(w http.ResponseWriter, r *http.Request) {
	//		num := strconv.FormatInt(int64(runtime.NumGoroutine()), 10)
	//		w.Write([]byte(num))
	//	})
	//	http.ListenAndServe("localhost:6060", nil)
	//	glog.Info("goroutine stats and pprof listen on 6060")
	//}()

	pool := NewPool(100)
	pool.SetTaskNum(1000000)
	go func() {
		for i := 0; i < 1000000; i++ {
			pool.AddTask(NewTask(taskFunc, callbackFunc, i))
		}
	}()

	pool.Run()

	t.Logf("%v", pool.GetResult())
	t.Errorf("program total run time is %f seconds", pool.GetRunTime())

}

func taskFunc(args interface{}) (error, interface{}) {
	//fmt.Println("task ", args, "completed")
	_ = 1+1
	return nil, args
}

func callbackFunc(result interface{}) (error, interface{}) {
	// 处理
	//fmt.Println("callback completed [", result, "]")
	return nil, result
}
