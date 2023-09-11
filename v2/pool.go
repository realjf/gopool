package gopool

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	color "github.com/TwiN/go-color"
	log "github.com/sirupsen/logrus"
)

const (
	POOL_MAX_CAP int = 1000000
)

type Pool interface {
	AddTask(f func()) error
	Run()
	DefaultInit()
	GetRunTime() float64
	SetTaskNum(int) // 设置任务总数
	GetDoneNum() int
	GetFailNum() int
	GetSuccessNum() int
	SetDebug(bool)
	SetTimeout(time.Duration)
	GetGoroutineNum() int // 获取goroutine数量
	GetBusyWorkerNum() int
	GetIdleWorkerNum() int
	Close() error
}

// goroutine协程池
type pool struct {
	cap           *atomic.Value // 协程池work容量
	taskQueue     chan func()   // 接收任务队列
	jobQueue      chan func()   // 工作队列
	startTime     time.Time     // 开始时间
	endTime       time.Time     // 结束时间
	taskNum       *atomic.Value // 接收任务总数量
	doneNum       *atomic.Value // 完成任务总数量
	successNum    *atomic.Value // 成功任务数量
	failNum       *atomic.Value // 失败任务数量
	debug         *atomic.Value // 是否开启调试
	timeout       time.Duration //每个任务执行的超时时间， 默认是没有超时限制
	workerMap     *workerMap    // key-协程id, value-是否忙碌中
	idleWorkerNum *atomic.Value // 空闲work数量
	busyWorkerNum *atomic.Value // 工作中的work数量
	nextWorkId    int           // 用于补充worker时的id分配
	lock          sync.Mutex
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewPool(cap int) Pool {
	ctx, cancel := context.WithCancel(context.Background())
	p := &pool{
		cap:           &atomic.Value{},
		taskNum:       &atomic.Value{},
		doneNum:       &atomic.Value{},
		successNum:    &atomic.Value{},
		failNum:       &atomic.Value{},
		idleWorkerNum: &atomic.Value{},
		busyWorkerNum: &atomic.Value{},
		debug:         &atomic.Value{},
		timeout:       0,
		taskQueue:     make(chan func(), 100),
		jobQueue:      make(chan func(), 100),
		workerMap:     newWorkerMap(),
		lock:          sync.Mutex{},
		nextWorkId:    cap + 1,
		ctx:           ctx,
		cancel:        cancel,
	}

	p.cap.Store(cap)
	p.taskNum.Store(0)
	p.doneNum.Store(0)
	p.successNum.Store(0)
	p.failNum.Store(0)
	p.idleWorkerNum.Store(0)
	p.busyWorkerNum.Store(0)
	p.debug.Store(false)

	return p
}

// 手动初始化
func (p *pool) init() {
	p.startTime = time.Now()
	if p.cap.Load().(int) <= 0 {
		p.DefaultInit()
	}
	if p.cap.Load().(int) >= POOL_MAX_CAP {
		p.cap.Store(POOL_MAX_CAP)
	}

	// 初始化协程池
	for i := int(0); i < p.cap.Load().(int); i++ {
		go p.runWorker(i + 1)
	}
}

// 默认初始化
func (p *pool) DefaultInit() {
	p.cap.Store(100)
	p.taskNum.Store(0)
}

// 设置调试开关
func (p *pool) SetDebug(debug bool) {
	p.debug.Store(debug)
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)
}

// 添加任务到任务队列
func (p *pool) AddTask(f func()) error {
	if f == nil {
		return errors.New("add task error: task is nil")
	}
	p.taskQueue <- f
	return nil
}

func (p *pool) SetTaskNum(total int) {
	p.taskNum.Store(total)
}

func (p *pool) GetTaskNum() int {
	return p.taskNum.Load().(int)
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
	p.workerMap.Store(gId, false)
	defer p.workerMap.Delete(gId)
	workerCtx, cancel := context.WithCancel(p.ctx)
	defer cancel()

stop:
	for {
		select {
		case task := <-p.jobQueue:
			if task == nil {
				continue
			}
			p.workerMap.Store(gId, true)
			var err error
			if p.timeout > 0 {
				NewWorkerWithTimeout(p.timeout).Run(task)
			} else {
				NewWorker().Run(task)
			}
			// 返回处理结果
			p.lock.Lock()
			oldDone := p.doneNum.Load().(int)
			p.doneNum.CompareAndSwap(oldDone, oldDone+1)
			if err != nil {
				oldFail := p.failNum.Load().(int)
				p.failNum.CompareAndSwap(oldFail, oldFail+1)
				log.Error(color.InRed("job execute error:" + err.Error()))
			} else {
				oldSucc := p.successNum.Load().(int)
				p.successNum.CompareAndSwap(oldSucc, oldSucc+1)
			}
			log.Info(color.InBlue("the number of jobs completed :" + strconv.FormatInt(int64(p.doneNum.Load().(int)), 10)))
			log.Printf("job completed: %d\n", p.doneNum.Load().(int))
			p.lock.Unlock()
			p.workerMap.Store(gId, false)
		case <-workerCtx.Done():
			break stop
		}
	}
}

// 获取运行总时间
func (p *pool) GetRunTime() float64 {
	t := p.endTime.Sub(p.startTime)
	return t.Seconds()
}

// 获取完成任务数
func (p *pool) GetDoneNum() int {
	return p.doneNum.Load().(int)
}

// 获取成功任务数
func (p *pool) GetSuccessNum() int {
	return p.successNum.Load().(int)
}

// 获取失败任务数
func (p *pool) GetFailNum() int {
	return p.failNum.Load().(int)
}

// 设置任务执行超时时间
func (p *pool) SetTimeout(t time.Duration) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.timeout = t
}

func (p *pool) GetGoroutineNum() int {
	return p.workerMap.GetLength()
}

func (p *pool) GetBusyWorkerNum() int {
	return p.workerMap.GetBusyNum()
}

func (p *pool) GetIdleWorkerNum() int {
	return p.workerMap.GetIdleNum()
}

func (p *pool) healthCheck() {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()
	heartBeatCtx, cancel := context.WithCancel(p.ctx)
	defer cancel()
	for {
		select {
		case <-ticker.C:
			log.Info(color.InBlue("total goroutines: " + strconv.FormatInt(int64(p.GetGoroutineNum()), 10)))
			log.Info(color.InGreen("idle goroutines: " + strconv.FormatInt(int64(p.GetIdleWorkerNum()), 10)))
			log.Info(color.InRed("busy goroutines: " + strconv.FormatInt(int64(p.GetBusyWorkerNum()), 10)))
			log.Info(color.InRed("job done num: " + strconv.FormatInt(int64(p.GetDoneNum()), 10)))
			if p.GetGoroutineNum() < p.cap.Load().(int) {
				for i := int(0); i < p.cap.Load().(int)-p.GetGoroutineNum(); i++ {
					go p.runWorker(p.nextWorkId)
					p.nextWorkId++
				}
			}
		case <-heartBeatCtx.Done():
			return
		}
	}
}

// 启动协程池
func (p *pool) Run() {
	// 初始化
	p.init()
	go p.healthCheck()

	now := time.Now()
	for p.GetGoroutineNum() != p.cap.Load().(int) {
		if time.Now().After(now.Add(2*time.Second)) && p.debug.Load().(bool) {
			log.Info(color.InRed("goroutines: " + strconv.FormatInt(int64(p.GetGoroutineNum()), 10)))
			now = time.Now()
		}
		runtime.Gosched()
	}

	go p.run()
}

func (p *pool) run() {
stop:
	for {
		select {
		case task := <-p.taskQueue:
			p.jobQueue <- task
		case <-p.ctx.Done():
			break stop
		}
	}
}

func (p *pool) Close() error {
	// 运行结束时间
	if p.debug.Load().(bool) {
		log.Println("waiting for stop...")
	}

	if p.taskNum.Load().(int) == 0 {
		log.Println("waiting for worker1")
	exit1:
		for {
			select {
			case <-time.After(time.Second):
				log.Println("waiting for worker")
				if p.GetBusyWorkerNum() > 0 {
					if p.debug.Load().(bool) {
						log.Info(color.InRed(strconv.FormatInt(int64(p.GetBusyWorkerNum()), 10) + " worker is runing"))
					}
				} else {
					break exit1
				}
			}
		}
		p.cancel()
	} else {
	exit2:
		for {
			select {
			case <-time.After(time.Second):
				log.Println("waiting for task...")
				p.lock.Lock()
				log.Printf("total task %d\n", p.taskNum.Load().(int))
				if p.doneNum.Load().(int) == p.taskNum.Load().(int) {
					p.cancel()
					p.lock.Unlock()
					break exit2
				}
				p.lock.Unlock()
			}
		}
	}
	debug.FreeOSMemory()

	stopTime, ok := p.ctx.Deadline()
	if !ok {
		p.endTime = time.Now()
	} else {
		p.endTime = stopTime
	}
	p.lock.Lock()
	close(p.taskQueue)
	close(p.jobQueue)
	p.lock.Unlock()
	if p.debug.Load().(bool) {
		log.Println("done")
	}
	return nil
}
