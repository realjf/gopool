package gopool

import (
	"errors"
	"sync"
	"time"
)

const (
	POOL_MAX_CAP = 1000000
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
	SetTaskNum(int) // 设置任务总数
}

// goroutine协程池
type Pool struct {
	cap       int             // 协程池work容量
	taskNum   int             // 接收任务总数量
	TaskQueue chan *Task      // 接收任务队列
	JobQueue  chan *Task      // 工作队列
	startTime time.Time       // 开始时间
	endTime   time.Time       // 结束时间
	wg        *sync.WaitGroup // 同步所有goroutine
	result    []interface{}   // 所有的运行结果
	ch        chan bool
	lock      sync.RWMutex
}

func NewPool(cap int) *Pool {
	return &Pool{
		cap:       cap,
		taskNum:   0,
		TaskQueue: make(chan *Task, 100),
		JobQueue:  make(chan *Task, 100),
		wg:        &sync.WaitGroup{},
		ch:        make(chan bool),
		lock:      sync.RWMutex{},
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

	// 初始化协程池
	for i := 1; i <= p.cap; i++ {
		go p.runWorker(i, p.ch)
		// log.Printf("worker [%v]", i)
	}
}

// 默认初始化
func (p *Pool) DefaultInit() {
	p.cap = 100
	p.taskNum = 0
}

// 添加任务到任务队列
func (p *Pool) AddTask(task *Task) error {
	if task == nil {
		return errors.New("add task error: task is nil")
	}
	p.TaskQueue <- task
	return nil
}

func (p *Pool) SetTaskNum(total int) {
	p.taskNum = total
}

// 运行一个goroutine协程
func (p *Pool) runWorker(workId int, ch chan bool) {
stop:
	for {
		select {
		case task := <-p.JobQueue:
			if task == nil {
				continue
			}
			worker := NewWorker(workId, task)
			_ = worker.Run()
			// 返回处理结果
			p.lock.Lock()
			p.result = append(p.result, worker.Task.Result)
			p.lock.Unlock()
		case <-ch:
			break stop
		}
	}
}

// 返回所有运行结果
func (p *Pool) GetResult() []interface{} {
	return p.result
}

func (p *Pool) stop() {
	// 关闭接收任务队列
	// log.Println("close task channel")
	close(p.TaskQueue)

	// 关闭处理任务队列
	// log.Println("close job channel")
	close(p.JobQueue)

	close(p.ch)

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
			p.JobQueue <- task
		default:
			p.lock.RLock()
			if len(p.result) == p.taskNum {
				p.ch <- true // 通知回收goroutine
				break stop
			}
			p.lock.RUnlock()
		}
	}

	// 结束
	p.stop()
}
