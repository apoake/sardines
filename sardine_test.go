package sardines

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	p, err := NewFixSizePools(10)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 100; i++ {
		p.Summit(func() {
			fmt.Println("summit Func: ", time.Now())
		})
	}

	f1 := func() (interface{}, error) {
		//Something todo
		return 1, nil
	}
	result, _ := p.SummitTask(f1)
	data, err := result.Get()
	fmt.Printf("%v, %v\n", data, err)
	startTime := time.Now()
	result, _ = p.SummitTask(func() (interface{}, error) {
		time.Sleep(5 * time.Second)
		return nil, errors.New("for test")
	})
	data, err = result.Get()
	endTime := time.Now()
	fmt.Printf("data: %v, \nerr: %v, \nstartTime: %v, \nendTime: %v, \ninterval:%v\n", data, err, startTime, endTime, endTime.Sub(startTime))

	startTime = time.Now()
	result, _ = p.SummitTask(func() (interface{}, error) {
		time.Sleep(5 * time.Second)
		return nil, errors.New("for test")
	})
	data, err = result.GetTimed(2 * time.Second)
	endTime = time.Now()
	fmt.Printf("data: %v, \nerr: %v, \nstartTime: %v, \nendTime: %v, \ninterval:%v\n", data, err, startTime, endTime, endTime.Sub(startTime))
	time.Sleep(1 * time.Minute)
}

func TestPoolClose(t *testing.T) {
	p, err := NewFixSizePools(50)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 100; i++ {
		p.Summit(func() {
			time.Sleep(1000 *time.Millisecond)
			fmt.Println("summit Func: ", time.Now())
		})
	}
	p.Close()
}

func TestOnePool(t *testing.T) {
	p, _ := NewOneFixSizePools(80)
	var index int64 = 0
	for i := 0; i < 100; i++ {
		p.Summit(func() {
			time.Sleep(2 * time.Second)
			fmt.Println(atomic.AddInt64(&index, 1))
		})
	}
	p.Wait()
}
