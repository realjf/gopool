# gopool

### 具体运行方法
```golang
// 设置运行参数
pool := NewPool(10000) // 并发数
pool.SetTaskNum(1000000) // 设置任务总数

// 设置任务函数
func taskFunc(args interface{}) (error, interface{}) {
	//fmt.Println("task ", args, "completed")
	_ = 1 + 1
	return nil, args
}

// 设置任务结果回调函数
func callbackFunc(result interface{}) (error, interface{}) {
	// 处理
	//fmt.Println("callback completed [", result, "]")
	return nil, result
}

// 添加任务
go func() {
	for i := 0; i < 1000000; i++ {
		pool.AddTask(gopool.NewTask(taskFunc, callbackFunc, i))
	}
}()


// 开始运行
pool.Run()

// 获取运行结果
pool.GetResult()

// 获取总运行时间
pool.GetRunTime()
```
