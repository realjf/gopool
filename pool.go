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
	runWorker(int, chan bool)
	AddTask(ITask) error
	Run()
	DefaultInit()
	GetRunTime() float64
	GetResult() []any
	SetTaskNum(int) // 设置任务总数
	GetDoneNum() int
	GetFailNum() int
	GetSuccessNum() int
	SetDebug(bool)
	SetTimeout(time.Duration)
	GetGoroutineNum() int // 获取goroutine数量
	GetBusyWorkerNum() int
	GetIdleWorkerNum() int
	healthCheck()
}

// goroutine协程池
type pool struct {
	cap           int             // 协程池work容量
	taskNum       int             // 接收任务总数量
	TaskQueue     chan ITask      // 接收任务队列
	JobQueue      chan ITask      // 工作队列
	startTime     time.Time       // 开始时间
	endTime       time.Time       // 结束时间
	wg            *sync.WaitGroup // 同步所有goroutine
	result        []any           // 所有的运行结果
	doneNum       int             // 完成任务总数量
	successNum    int             // 成功任务数量
	failNum       int             // 失败任务数量
	debug         bool            // 是否开启调试
	timeout       time.Duration   //每个任务执行的超时时间， 默认是没有超时限制
	ticker        *time.Ticker
	heartBeat     chan int64 // 心跳通道,用于接收goroutines的心跳通知
	workerMap     *workerMap // key-协程id, value-是否忙碌中
	idleWorkerNum int        // 空闲work数量
	busyWorkerNum int        // 工作中的work数量
	nextWorkId    int        // 用于补充worker时的id分配
	killSignal    chan int64 // 强制杀死worker,传递goroutine的id
	ch            chan bool
	lock          sync.RWMutex
}

func NewPool(cap int) Pool {
	return &pool{
		cap:           cap,
		taskNum:       0,
		TaskQueue:     make(chan ITask, 100),
		JobQueue:      make(chan ITask, 100),
		wg:            &sync.WaitGroup{},
		ch:            make(chan bool),
		lock:          sync.RWMutex{},
		doneNum:       0,
		successNum:    0,
		failNum:       0,
		debug:         false,
		timeout:       0,
		ticker:        time.NewTicker(time.Second),
		heartBeat:     make(chan int64, 10),
		workerMap:     newWorkerMap(),
		idleWorkerNum: 0,
		busyWorkerNum: 0,
		nextWorkId:    cap + 1,
		killSignal:    make(chan int64),
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
		go p.runWorker(i+1, p.ch)
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
func (p *pool) AddTask(task ITask) error {
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
func (p *pool) runWorker(workId int, ch chan bool) {
	gId := getGoroutineId()
	p.workerMap.Store(gId, false)
	defer func() {
		p.workerMap.Delete(gId)
	}()
	go func(gId int64) {
		for {
			select {
			case <-p.ticker.C:
				p.heartBeat <- gId
			case <-ch:
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
			ctx = context.WithValue(ctx, Debug, p.debug)
			ctx = context.WithValue(ctx, Timeout, p.timeout)
			err := worker.Run(ctx)
			// 返回处理结果
			p.lock.Lock()
			p.doneNum++
			if err != nil {
				p.failNum++
				if errors.Is(err, TimecoutError) {
					log.Error(TimecoutError.Error())
				} else if errors.Is(err, context.DeadlineExceeded) {
					log.Error(context.DeadlineExceeded.Error())
				} else if os.IsTimeout(err) {
					log.Error("IsTimeoutError:" + err.Error())
				} else {
					log.Error(color.InRed("job execute error:" + err.Error()))
				}
			} else {
				p.successNum++
			}
			p.result = append(p.result, worker.GetTask().GetResult())
			log.Info(color.InBlue("the number of jobs completed :" + strconv.Itoa(p.doneNum)))
			p.lock.Unlock()
			p.workerMap.Store(gId, false)
		case kgId := <-p.killSignal:
			if kgId == gId {
				log.Warn(color.InYellow("worker[" + strconv.FormatInt(kgId, 10) + "] is killed"))
				runtime.Goexit()
				// break stop
			}
		case <-ch:
			break stop
		}
	}
}

// 返回所有运行结果
func (p *pool) GetResult() []any {
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

	close(p.heartBeat)

	close(p.killSignal)

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

	for {
		select {
		case gId := <-p.heartBeat:
			// check health
			p.workerMap.LoadOrStore(gId, false)
		case <-ticker.C:
			log.Info(color.InBlue("total goroutines: " + strconv.Itoa(p.GetGoroutineNum())))
			log.Info(color.InGreen("idle goroutines: " + strconv.Itoa(p.GetIdleWorkerNum())))
			log.Info(color.InRed("busy goroutines: " + strconv.Itoa(p.GetBusyWorkerNum())))
			log.Info(color.InRed("job queue length: " + strconv.Itoa(len(p.JobQueue))))
			if p.GetGoroutineNum() < p.cap {
				for i := 0; i < p.cap-p.GetGoroutineNum(); i++ {
					go p.runWorker(p.nextWorkId, p.ch)
					p.nextWorkId++
				}
			}
		case <-p.ch:
			return
		}
	}
}

// 启动协程池
func (p *pool) Run() {
	// 初始化
	p.init()
	go p.healthCheck()

	var now = time.Now()
	for p.GetGoroutineNum() != p.cap {
		if time.Now().After(now.Add(2*time.Second)) && p.debug {
			log.Info(color.InRed("goroutines: " + strconv.Itoa(p.GetGoroutineNum())))
			now = time.Now()
		}
		runtime.Gosched()
	}

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
