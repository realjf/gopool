# gopool


### cgroup安装
```sh
apt-get install cgroup-tools golang-github-containerd-cgroups-dev libcgroup-dev
```

### 具体运行方法
```golang
// 设置运行参数
pool := NewPool(10000) // 并发数
pool.SetTaskNum(1000000) // 设置任务总数

// 设置调试开关
func (p *Pool) SetDebug(debug bool) {
 p.debug = debug
}

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

// 设置任务执行超时时间，默认1分钟
p.SetTimeout(time.Minute * 1)

// 添加任务
go func() {
	for i := 0; i < 1000000; i++ {
		pool.AddTask(gopool.NewTask(taskFunc, callbackFunc, i))
	}
}()


// 开始运行
pool.Run()

// 获取运行结果（可选）
pool.GetResult()

// 获取总运行时间
pool.GetRunTime()


// 获取完成任务数
pool.GetDoneNum()

// 获取成功任务数
pool.GetSuccessNum()

// 获取失败任务数
pool.GetFailNum()

// 获取当前运行中的goroutine数量
pool.GetGoroutineNum()

// 获取当前忙碌worker数量
pool.GetBusyWorkerNum()

// 获取当前空闲worker数量
pool.GetIdleWorkerNum()
```
