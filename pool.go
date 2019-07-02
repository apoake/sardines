package sardines

import (
	"errors"
	"sync"
)

var (
	ErrInvalidPoolSize = errors.New("invalid pool size")
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
			worker.close()
		}
	}
	p.poolSize = 0
	close(p.reqChan)
}

func (p *pool) initWorkers() error {
	p.mut.Lock()
	defer p.mut.Unlock()
	if p.poolSize <= 0 {
		return ErrInvalidPoolSize
	}
	workers := make([]Worker, 0)
	for i := 0; i < p.poolSize; i++ {
		workers = append(workers, NewLoopWork(p.reqChan))
	}
	p.workers = workers
	return nil
}

func NewFixSizePools(size int) (*pool, error) {
	p := &pool{poolSize: size, reqChan: make(chan interface{})}
	err := p.initWorkers()
	return p, err
}
