# gopool

Go coordinated process pool(协程池)

### Usage


#### Steps needed(必要步骤)

```go
// Set operating parameters(设置运行参数)
pool := NewPool(10) // Concurrency number(并发数)
pool.SetTaskNum(1000) // Task number(设置任务总数)


// Set Task Function Template(设置任务函数)
func taskFunc(args any) (any, error) {
	// do something

	// !!! you should return the result so that the callback function can use it(应该返回结果给回调函数使用) !!!
	return args, nil
}

// Set Task Callback Function Template(设置任务结果回调函数)
// The result parameter is passed by the task function output(result参数是任务函数返回的结果)
func callbackFunc(result any) (any, error) {
	// do something
	return result, nil
}


// Add task to queue(添加任务)
go func() {
	for i := 0; i < 1000000; i++ {
		pool.AddTask(gopool.NewTask(taskFunc, callbackFunc, i))
	}
}()


// Start running(开始运行)
pool.Run()
// or(或者)
go pool.Run()

```

#### Optional operation(可选操作)
```go
// Set debugging flag(设置调试开关)
func (p *Pool) SetDebug(debug bool) {
 p.debug = debug
}

// Set a per-task run timeout; the default is one minute(设置任务执行超时时间，默认1分钟)
p.SetTimeout(time.Minute * 1)

// Get all result(获取运行结果)
pool.GetResult()

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

type TaskFunc func(args any) (any, error)
type CallbackFunc func(result any) (any, error)

task := NewEmptyTask()
task.SetTaskFunc(taskFunc TaskFunc)
task.SetCallbackFunc(callbackFunc CallbackFunc)
task.SetArgs(args any)
task.SetId(id int)
```
#### design your task by implement ITask interface
```go
type ITask interface {
 Execute() error
 GetResult() any
}

// for example
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

mytask := &MyTask{}
pool.AddTask(mytask)
...
```
