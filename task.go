package gopool

// 任务队列
type Task struct {
	taskFunc func(args interface{}) (error, interface{})
	Callback func(result interface{}) (error, interface{}) // 执行完成回到函数
	Result   interface{}                                   // 运行结果
	Args     interface{}                                   // 参数
}

func NewTask(taskFunc func(interface{}) (error, interface{}), callback func(interface{}) (error, interface{}), args interface{}) *Task {
	return &Task{
		taskFunc: taskFunc,
		Callback: callback,
		Args:     args,
	}
}

// 执行任务函数
func (t *Task) Execute() error {
	err, result := t.taskFunc(t.Args)
	if err != nil {
		return err
	}
	err, res := t.Callback(result)
	if err != nil {
		return err
	}
	t.Result = res
	return nil
}
