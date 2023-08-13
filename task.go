package gopool

import (
	"errors"
)

var _ ITask = (*Task)(nil)

type ITask interface {
	Execute() error
	GetResult() any
	ExecChan() <-chan error
}

type TaskFunc func(args any) (any, error)
type CallbackFunc func(result any) (any, error)

// 任务队列
type Task struct {
	id       int
	taskFunc TaskFunc
	callback CallbackFunc // 执行完成回到函数
	result   any          // 运行结果
	args     any          // 参数
	ch       chan error
}

func NewTask(taskFunc TaskFunc, callback CallbackFunc, args any) ITask {
	return &Task{
		taskFunc: taskFunc,
		callback: callback,
		args:     args,
		ch:       make(chan error, 1),
	}
}

func NewEmptyTask() *Task {
	return &Task{}
}

func (t *Task) SetTaskFunc(f TaskFunc) {
	t.taskFunc = f
}

func (t *Task) SetCallbackFunc(f CallbackFunc) {
	t.callback = f
}

func (t *Task) SetArgs(args any) {
	t.args = args
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
func (t *Task) Execute() (err error) {
	defer func(e error) {
		t.ch <- e
		close(t.ch)
	}(err)
	if t.taskFunc == nil {
		err = errors.New("task func is nil")
		return
	}
	result, err := t.taskFunc(t.args)
	if err != nil {
		return
	}
	if t.callback != nil {
		var res any
		res, err = t.callback(result)
		if err != nil {
			return
		}
		t.result = res
		return
	}
	t.result = result
	return
}

func (t *Task) ExecChan() <-chan error {
	return t.ch
}
