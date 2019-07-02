package sardines

import (
	"errors"
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

type worker interface {
	Close()
	Run()
}

type loopWork struct {
	reqChain   chan<- interface{}
	closeChain chan<- struct{}
}

func (w *loopWork) Close() {
	w.closeChain <-
}

func (*loopWork) Run() {
	panic("implement me")
}
