package gopool

import (
	"errors"
)

type TaskFunc func(args any) (any, error)
type CallbackFunc func(result any) (any, error)

// 任务队列
type Task struct {
	id       int
	taskFunc TaskFunc
	callback CallbackFunc // 执行完成回到函数
	result   any          // 运行结果
	args     any          // 参数
}

func NewTask(taskFunc TaskFunc, callback CallbackFunc, args any) *Task {
	return &Task{
		taskFunc: taskFunc,
		callback: callback,
		args:     args,
	}
}

func (t *Task) GetResult() any {
	return t.result
}

func (t *Task) SetId(id int) {
	t.id = id
}

func (t *Task) GetId() int {
	return t.id
}

// 执行任务函数
func (t *Task) Execute() error {
	if t.taskFunc == nil {
		return errors.New("task func is nil")
	}
	result, err := t.taskFunc(t.args)
	if err != nil {
		return err
	}
	if t.callback != nil {
		res, err := t.callback(result)
		if err != nil {
			return err
		}
		t.result = res
		return nil
	}
	t.result = result
	return nil
}
