package gopool

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	POOL_MAX_CAP = 10000
)

type pool interface {
	init()
	stop()
	runWorker(int)
	AddTask(*Task) error
	Run()
	DefaultInit()
	GetRunTime() float64
	GetResult() []interface{}
}

// goroutine协程池
type Pool struct {
	cap           int             // 协程池容量
	taskNum       int             // 接收任务总数量
	TaskQueue     chan *Task      // 接收任务队列
	JobQueue      chan *Task      // 工作队列
	startTime     time.Time       // 开始时间
	endTime       time.Time       // 结束时间
	wg            *sync.WaitGroup // 同步所有goroutine
	ctx           context.Context
	ctxCancelFunc context.CancelFunc
	ctxWithCancel context.Context
	result        []interface{} // 所有的运行结果
}

func NewPool(cap int) *Pool {
	return &Pool{
		cap:       cap,
		taskNum:   0,
		TaskQueue: make(chan *Task, 100),
		JobQueue:  make(chan *Task, 100),
		wg:        &sync.WaitGroup{},
		ctx:       context.Background(),
	}
}

// 手动初始化
func (p *Pool) init() {
	p.startTime = time.Now()
	if p.cap <= 0 {
		p.DefaultInit()
	}
	if p.cap >= POOL_MAX_CAP {
		p.cap = POOL_MAX_CAP
	}
	// 从ctx中派生一个context
	p.ctxWithCancel, p.ctxCancelFunc = context.WithCancel(p.ctx)

	// 初始化协程池
	for i := 0; i < p.cap; i++ {
		go p.runWorker(i, p.ctxWithCancel)
	}
}

// 默认初始化
func (p *Pool) DefaultInit() {
	p.cap = 100
	p.cap = 0
}

// 添加任务到任务队列
func (p *Pool) AddTask(task *Task) error {
	if task == nil {
		return errors.New("add task error: task is nil")
	}
	p.TaskQueue <- task
	p.taskNum++
	return nil
}

// 运行一个goroutine协程
func (p *Pool) runWorker(workId int, ctx context.Context) {
	// 超时限制 600s
	ctxWithTimeout, ctxTimeoutFunc := context.WithTimeout(ctx, time.Duration(600) * time.Second)
	defer func() {
		ctxTimeoutFunc()
	}()

	for task := range p.JobQueue {
		// @todo 死锁问题
		worker := NewWorker(workId, task)
		_ = worker.Run(ctxWithTimeout)
		// 返回处理结果
		p.result = append(p.result, worker.Task.Result)
	}
}

// 返回所有运行结果
func (p *Pool) GetResult() []interface{} {
	return p.result
}

func (p *Pool) stop() {
	// 关闭接收任务队列
	log.Println("close task channel")
	close(p.TaskQueue)

	// 关闭处理任务队列
	log.Println("close job channel")
	close(p.JobQueue)

	// 运行结束时间
	p.endTime = time.Now()
}

// 获取运行总时间
func (p *Pool) GetRunTime() float64 {
	t := p.endTime.Sub(p.startTime)
	return t.Seconds()
}

// 启动协程池
func (p *Pool) Run() {
	// 初始化
	p.init()

	stop:
	for {
		select {
		case task := <-p.TaskQueue:
			fmt.Println("add task")
			p.JobQueue <- task
		case <-p.ctx.Done():
			fmt.Println("main done")
			break stop
		}
	}

	//for task := range p.TaskQueue {
	//	// @todo 死锁问题
	//	p.JobQueue <- task
	//}

	// 结束
	p.stop()
}
