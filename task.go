package gopool

import (
	"errors"
)

type TaskFunc func(args interface{}) (interface{}, error)
type CallbackFunc func(result interface{}) (interface{}, error)

// 任务队列
type Task struct {
	TaskId   int
	taskFunc TaskFunc
	Callback CallbackFunc // 执行完成回到函数
	Result   interface{}  // 运行结果
	Args     interface{}  // 参数
}

func NewTask(taskFunc TaskFunc, callback CallbackFunc, args interface{}) *Task {
	return &Task{
		taskFunc: taskFunc,
		Callback: callback,
		Args:     args,
	}
}

// 执行任务函数
func (t *Task) Execute() error {
	if t.taskFunc == nil {
		return errors.New("task func is nil")
	}
	result, err := t.taskFunc(t.Args)
	if err != nil {
		return err
	}
	res, err := t.Callback(result)
	if err != nil {
		return err
	}
	t.Result = res
	return nil
}
