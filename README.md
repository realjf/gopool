# gopool

Go coordinated process pool(协程池)

### v2

### Usage

```bash
go get github.com/realjf/gopool/v2

```

#### Steps needed(必要步骤)

```go
// Set operating parameters(设置运行参数)
pool := gopool.NewPool(10) // Concurrency number(并发数)
pool.SetTaskNum(1000) // Task number(设置任务总数)
// Start running(开始运行)
pool.Run()
// Add task to queue(添加任务)
 for i := 0; i < 1000000; i++ {
  pool.AddTask(func(){
   fmt.Println("mytask running...")
  })
 }

// Waiting for task to finish(等待结束关闭)
pool.Close()

```

#### Optional operation(可选操作)

```go
// Set debugging flag(设置调试开关)
pool.SetDebug(debug bool)

// Set a per-task run timeout; the default is  no limit(设置任务执行超时时间，默认没限制)
pool.SetTimeout(time.Minute * 1)

// Get the total run time(获取总运行时间)
pool.GetRunTime()


// Get the number of tasks done(获取完成任务数)
pool.GetDoneNum()

// Get the number of tasks that were successful(获取成功任务数)
pool.GetSuccessNum()

// Get the number of tasks that were failure(获取失败任务数)
pool.GetFailNum()

// Get the current goroutine number(获取当前运行中的goroutine数量)
pool.GetGoroutineNum()

// Get the current busy worker number(获取当前忙碌worker数量)
pool.GetBusyWorkerNum()

// Get the current idle worker number(获取当前空闲worker数量)
pool.GetIdleWorkerNum()
```

### [v1](./README_v1.md)
