package gopool

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TwiN/go-color"
	log "github.com/sirupsen/logrus"
)

const (
	POOL_MAX_CAP = 1000000
)

type Pool interface {
	init()
	stop()
	runWorker(int)
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
	SetTimeout(time.Duration)
	GetGoroutineNum() int // 获取goroutine数量
	GetBusyWorkerNum() int
	GetIdleWorkerNum() int
}

// goroutine协程池
type pool struct {
	cap           int             // 协程池work容量
	taskNum       int             // 接收任务总数量
	TaskQueue     chan *Task      // 接收任务队列
	JobQueue      chan *Task      // 工作队列
	startTime     time.Time       // 开始时间
	endTime       time.Time       // 结束时间
	wg            *sync.WaitGroup // 同步所有goroutine
	result        []interface{}   // 所有的运行结果
	doneNum       int             // 完成任务总数量
	successNum    int             // 成功任务数量
	failNum       int             // 失败任务数量
	debug         bool            // 是否开启调试
	timeout       time.Duration   //每个任务执行的超时时间， 默认是1分钟
	ticker        *time.Ticker
	heartBeat     chan int64 // 心跳通道,用于接收goroutines的心跳通知
	workerMap     *sync.Map  // key-协程id, value-是否忙碌中
	idleWorkerNum int        // 空闲work数量
	busyWorkerNum int        // 工作中的work数量
	ch            chan bool
	lock          sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewPool(cap int) Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &pool{
		cap:           cap,
		taskNum:       0,
		TaskQueue:     make(chan *Task, 100),
		JobQueue:      make(chan *Task, 100),
		wg:            &sync.WaitGroup{},
		ch:            make(chan bool),
		lock:          sync.RWMutex{},
		doneNum:       0,
		successNum:    0,
		failNum:       0,
		debug:         false,
		timeout:       time.Minute * 1,
		ticker:        time.NewTicker(time.Second),
		heartBeat:     make(chan int64, 10),
		workerMap:     &sync.Map{},
		idleWorkerNum: 0,
		busyWorkerNum: 0,
		ctx:           ctx,
		cancel:        cancel,
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
	for i := 0; i < p.cap; i++ {
		go p.runWorker(i + 1)
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
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)
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

// 获取当前goroutine的id
func getGoroutineId() int64 {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("panic recover:panic info:%v\n", err)
		}
	}()

	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.ParseInt(idField, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

// 运行一个goroutine协程
func (p *pool) runWorker(workId int) {
	gId := getGoroutineId()
	defer func() {
		p.workerMap.Delete(gId)
	}()
	go func(gId int64) {
		for {
			select {
			case t := <-p.ticker.C:
				if p.debug {
					log.Info("ticker:" + t.Local().String())
				}

				p.heartBeat <- gId
			case <-p.ctx.Done():
				return
			}
		}
	}(gId)
stop:
	for {
		select {
		case task := <-p.JobQueue:
			if task == nil {
				continue
			}
			p.workerMap.Store(gId, true)
			worker := NewWorker(workId, task)
			ctx := context.Background()
			ctx2 := context.WithValue(ctx, "debug", p.debug)
			err := worker.Run(ctx2, p.timeout)
			log.Infof("job run over with: %v", err)
			select {
			case <-ctx.Done():
				if p.debug {
					log.Info(color.InGreen("job done with err:" + ctx.Err().Error()))
				}
			case <-time.After(p.timeout + time.Second*1):
				if p.debug {
					log.Info(color.InYellow("job execute timeout"))
				}
			}
			// 返回处理结果
			p.lock.Lock()
			p.doneNum++
			if err != nil {
				p.failNum++
				log.Error(color.InRed("job execute error:" + err.Error()))
			} else {
				p.successNum++
			}
			p.result = append(p.result, worker.Task.Result)
			log.Info(color.InBlue("the number of jobs completed :" + strconv.Itoa(p.doneNum)))
			p.lock.Unlock()
			p.workerMap.Store(gId, false)
			p.lock.RLock()
			if p.doneNum == p.taskNum {
				p.ch <- true // 通知回收goroutine
			}
			p.lock.RUnlock()
		case <-p.ctx.Done():
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
		log.Info("close task channel")
	}

	close(p.TaskQueue)

	// 关闭处理任务队列
	if p.debug {
		log.Info("close job channel")
	}

	close(p.JobQueue)

	close(p.ch)

	p.cancel()

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

// 设置任务执行超时时间
func (p *pool) SetTimeout(t time.Duration) {
	p.timeout = t
}

func (p *pool) GetGoroutineNum() int {
	go func() {
		gId := <-p.heartBeat
		p.workerMap.LoadOrStore(gId, false)
	}()
	var num int = 0
	p.workerMap.Range(func(key, value any) bool {
		num++
		return true
	})

	return num
}

func (p *pool) GetBusyWorkerNum() int {
	var num int = 0
	p.workerMap.Range(func(key, value any) bool {
		if value.(bool) {
			num++
		}
		return true
	})

	return num
}

func (p *pool) GetIdleWorkerNum() int {
	var num int = 0
	p.workerMap.Range(func(key, value any) bool {
		if !value.(bool) {
			num++
		}
		return true
	})

	return num
}

// 启动协程池
func (p *pool) Run() {
	// 初始化
	p.init()

	for p.GetGoroutineNum() != p.cap {
		if p.debug {
			log.Info(color.InRed("goroutines: " + strconv.Itoa(p.GetGoroutineNum())))
		}

		runtime.Gosched()
	}

	ticker := time.NewTicker(time.Second * 3)
	go func() {
		for {
			select {
			case <-ticker.C:
				if p.debug {
					p.lock.RLock()
					log.Info(color.InGreen("job done: " + strconv.Itoa(p.doneNum)))
					p.lock.RUnlock()
				}
				log.Info(color.InBlue("total goroutines: " + strconv.Itoa(p.GetGoroutineNum())))
				log.Info(color.InGreen("idle goroutines: " + strconv.Itoa(p.GetIdleWorkerNum())))
				log.Info(color.InRed("busy goroutines: " + strconv.Itoa(p.GetBusyWorkerNum())))
				log.Info(color.InRed("job queue length: " + strconv.Itoa(len(p.JobQueue))))
			case <-p.ctx.Done():
				return
			}
		}

	}()
stop:
	for {
		select {
		case task := <-p.TaskQueue:
			p.JobQueue <- task
		case <-p.ch:
			p.cancel()
			break stop
		}
	}

	// 结束
	p.stop()

}
