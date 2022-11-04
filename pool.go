package gopool

import (
	"errors"
	"log"
	"sync"
	"time"
)

const (
	POOL_MAX_CAP = 1000000
)

type Pool interface {
	init()
	stop()
	runWorker(int, chan bool)
	AddTask(*Task) error
	Run()
	DefaultInit()
	GetRunTime() float64
	GetResult() []interface{}
	SetTaskNum(int) // 设置任务总数
	GetDoneNum() int
	GetFailNum() int
	GetSuccessNum() int
	SetDebug(bool)
}

// goroutine协程池
type pool struct {
	cap        int             // 协程池work容量
	taskNum    int             // 接收任务总数量
	TaskQueue  chan *Task      // 接收任务队列
	JobQueue   chan *Task      // 工作队列
	startTime  time.Time       // 开始时间
	endTime    time.Time       // 结束时间
	wg         *sync.WaitGroup // 同步所有goroutine
	result     []interface{}   // 所有的运行结果
	doneNum    int             // 完成任务总数量
	successNum int             // 成功任务数量
	failNum    int             // 失败任务数量
	debug      bool            // 是否开启调试
	ch         chan bool
	lock       sync.RWMutex
}

func NewPool(cap int) Pool {
	return &pool{
		cap:        cap,
		taskNum:    0,
		TaskQueue:  make(chan *Task, 100),
		JobQueue:   make(chan *Task, 100),
		wg:         &sync.WaitGroup{},
		ch:         make(chan bool),
		lock:       sync.RWMutex{},
		doneNum:    0,
		successNum: 0,
		failNum:    0,
		debug:      false,
	}
}

// 手动初始化
func (p *pool) init() {
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
func (p *pool) DefaultInit() {
	p.cap = 100
	p.taskNum = 0
}

// 设置调试开关
func (p *pool) SetDebug(debug bool) {
	p.debug = debug
}

// 添加任务到任务队列
func (p *pool) AddTask(task *Task) error {
	if task == nil {
		return errors.New("add task error: task is nil")
	}
	p.TaskQueue <- task
	return nil
}

func (p *pool) SetTaskNum(total int) {
	p.taskNum = total
}

// 运行一个goroutine协程
func (p *pool) runWorker(workId int, ch chan bool) {
stop:
	for {
		select {
		case task := <-p.JobQueue:
			if task == nil {
				continue
			}
			worker := NewWorker(workId, task)
			err := worker.Run()
			// 返回处理结果
			p.lock.Lock()
			p.doneNum++
			if err != nil {
				p.failNum++
			} else {
				p.successNum++
			}
			p.result = append(p.result, worker.Task.Result)
			p.lock.Unlock()
		case <-ch:
			break stop
		}
	}
}

// 返回所有运行结果
func (p *pool) GetResult() []interface{} {
	return p.result
}

func (p *pool) stop() {
	// 关闭接收任务队列
	if p.debug {
		log.Println("close task channel")
	}

	close(p.TaskQueue)

	// 关闭处理任务队列
	if p.debug {
		log.Println("close job channel")
	}

	close(p.JobQueue)

	close(p.ch)

	// 运行结束时间
	p.endTime = time.Now()
}

// 获取运行总时间
func (p *pool) GetRunTime() float64 {
	t := p.endTime.Sub(p.startTime)
	return t.Seconds()
}

// 获取完成任务数
func (p *pool) GetDoneNum() int {
	return p.doneNum
}

// 获取成功任务数
func (p *pool) GetSuccessNum() int {
	return p.successNum
}

// 获取失败任务数
func (p *pool) GetFailNum() int {
	return p.failNum
}

// 启动协程池
func (p *pool) Run() {
	// 初始化
	p.init()

stop:
	for {
		select {
		case task := <-p.TaskQueue:
			p.JobQueue <- task
		default:
			p.lock.RLock()
			if p.doneNum == p.taskNum {
				p.ch <- true // 通知回收goroutine
				break stop
			}
			p.lock.RUnlock()
		}
	}

	// 结束
	p.stop()
}
