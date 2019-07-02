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
	initWorkers() error
}

type pool struct {
	mut      sync.Mutex
	workers  []*worker
	poolSize int
}

func (*pool) PoolSize() int {
	panic("implement me")
}

func (*pool) Close() {
	panic("implement me")
}

func (p *pool) initWorkers() error {
	p.mut.Lock()
	defer p.mut.Unlock()
	if p.poolSize <= 0 {
		return ErrInvalidPoolSize
	}
	workers := make([]*worker, 0)
	// TODO
	p.workers = workers
	return nil
}

func NewFixSizePools(size int) (*pool, error) {
	p := &pool{poolSize: size}
	err := p.initWorkers()
	return p, err
}
