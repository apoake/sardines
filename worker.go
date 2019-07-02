package sardines

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrWorkerTimeOut = errors.New("worker timeout")
)

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
}

func NewLoopWork(reqChan <-chan interface{}) *loopWork {
	lw := &loopWork{reqChain: reqChan, closeChain: make(chan struct{})}
	go lw.run()
	return lw
}

type loopWork struct {
	reqChain   <-chan interface{}
	closeChain chan struct{}
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
