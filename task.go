package gopool

import (
	"context"
	"errors"
)

type TaskFunc func(args interface{}) (error, interface{})
type CallbackFunc func(result interface{}) (error, interface{})

// 任务队列
type Task struct {
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
func (t *Task) Execute(ctx context.Context) error {
	if t.taskFunc == nil {
		return errors.New("task func is nil")
	}
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
