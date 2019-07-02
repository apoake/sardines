package sardines

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	ONE               int    = 1
	onePoolStopSignal Signal = -1
)

var (
	ErrWorkerTimeOut = errors.New("worker timeout")
)

type Signal int

type innerResult struct {
	res interface{}
	err error
}

type Result struct {
	resp    chan innerResult
	iResult *innerResult
}

func (r *Result) Get() (interface{}, error) {
	if r.iResult == nil {
		iResult := <-r.resp
		r.iResult = &iResult
	}
	return r.iResult.res, r.iResult.err
}

func (r *Result) GetTimed(d time.Duration) (interface{}, error) {
	if r.iResult == nil {
	loop:
		after := time.After(d)
		for {
			select {
			case iResult := <-r.resp:
				r.iResult = &iResult
				goto loop
			case <-after:
				return nil, ErrWorkerTimeOut
			}
		}
	}
	return r.iResult.res, r.iResult.err
}

type Run func()

type Task func() (interface{}, error)

type taskContext struct {
	task     Task
	resqChan chan<- innerResult
}

type Worker interface {
	close()
	run()
	workNo() string
}

func NewLoopWork(reqChan <-chan interface{}, workerNo string) *loopWork {
	lw := &loopWork{reqChain: reqChan, closeChain: make(chan struct{}), workerNo: workerNo}
	go lw.run()
	return lw
}

func NewOneLoopWork(reqChan <-chan interface{}, wg sync.WaitGroup, workerNo string) *oneLoopWork {
	lw := &oneLoopWork{loopWork: loopWork{reqChain: reqChan, closeChain: make(chan struct{}), workerNo: workerNo}, wg: wg}
	go lw.run()
	return lw
}

type loopWork struct {
	reqChain   <-chan interface{}
	closeChain chan struct{}
	workerNo   string
}

func (l *loopWork) workNo() string {
	return l.workerNo
}

type oneLoopWork struct {
	loopWork
	wg sync.WaitGroup
}

func (l *loopWork) close() {
	close(l.closeChain)
}

func (l *loopWork) run() {
	defer func() {
		l.close()
	}()
	for {
		select {
		case context := <-l.reqChain:
			switch sContext := context.(type) {
			case Run:
				sContext()
			case taskContext:
				result, err := sContext.task()
				sContext.resqChan <- innerResult{res: result, err: err}
				close(sContext.resqChan)
			default:
				fmt.Println("unknow type: ", sContext)
			}
		case <-l.closeChain:
			return
		}
	}
}

func (o *oneLoopWork) close() {}

func (o *oneLoopWork) run() {
	o.wg.Add(ONE)
	defer o.close()
	for {
		select {
		case context := <-o.reqChain:
			switch sContext := context.(type) {
			case Run:
				sContext()
			case Signal:
				if sContext == onePoolStopSignal {
					o.wg.Done()
					return
				} else {
					fmt.Printf("unknow signal: %v\n", sContext)
				}
			default:
				fmt.Printf("unknow type[%T]: %v\n", sContext, sContext)
			}
		}
	}
}
