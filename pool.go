package sardines

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

var (
	ErrInvalidPoolSize   = errors.New("invalid pool size")
	ErrOnePoolNotSupport = errors.New("one pool not support")

	PoolIndex int64 = 0
)

type Pool interface {
	PoolSize() int
	Close()
	Summit(run Run)
	SummitTask(task Task) (*Result, error)
	initWorkers() error
}

type pool struct {
	mut      sync.Mutex
	workers  []Worker
	poolSize int
	reqChan  chan interface{}
	poolNo   int64
}

type onePool struct {
	pool
	wg sync.WaitGroup
}


func (p *pool) Summit(run Run) {
	p.reqChan <- run
}

func (p *pool) SummitTask(task Task) (*Result, error) {
	resqChain := make(chan innerResult)
	p.reqChan <- taskContext{task: task, resqChan: resqChain}
	return &Result{resp: resqChain}, nil
}

func (p *pool) PoolSize() int {
	return p.poolSize
}

func (p *pool) Close() {
	p.mut.Lock()
	defer p.mut.Unlock()
	if p.poolSize != 0 {
		for _, worker := range p.workers {
			fmt.Printf("Pool close worker: %s\n", worker.workNo())
			worker.close()
		}
	}
	p.poolSize = 0
	close(p.reqChan)
}

func (p *pool) initWorkers(f func(string) Worker) error {
	p.mut.Lock()
	defer p.mut.Unlock()
	if p.poolSize <= 0 {
		return ErrInvalidPoolSize
	}
	workers := make([]Worker, 0)
	for i := 0; i < p.poolSize; i++ {
		workers = append(workers, f(fmt.Sprintf("Pool%d-%d", p.poolNo, i)))
	}
	p.workers = workers
	return nil
}

func (p *onePool) SummitTask(task Task) (*Result, error) {
	return nil, ErrOnePoolNotSupport
}

func (p *onePool) Wait() {
	p.wg.Done()
	for i := 0; i < len(p.workers); i++ {
		p.reqChan <- onePoolStopSignal
	}
	p.wg.Wait()
}

func (p *onePool) Close() {
	p.Wait()
	p.mut.Lock()
	defer p.mut.Unlock()
	p.workers = nil
	p.poolSize = 0
	close(p.reqChan)
}

func NewFixSizePools(size int) (*pool, error) {
	p := &pool{poolSize: size, reqChan: make(chan interface{}), poolNo: getPoolNo()}
	err := p.initWorkers(func(workerNo string) Worker {
		return NewLoopWork(p.reqChan, workerNo)
	})
	return p, err
}

func NewOneFixSizePools(size int) (*onePool, error) {
	p := &onePool{pool: pool{poolSize: size, reqChan: make(chan interface{}), poolNo: getPoolNo()}}
	err := p.initWorkers(func(workerNo string) Worker {
		return NewOneLoopWork(p.reqChan, p.wg, workerNo)
	})
	p.wg.Add(ONE)
	return p, err
}

func getPoolNo() int64 {
	return atomic.AddInt64(&PoolIndex, 1)
}
