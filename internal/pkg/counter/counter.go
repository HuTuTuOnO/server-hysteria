package counter

import (
	"sync/atomic"
)

type Counter struct {
	num int64
}

func NewCounter(num uint64) *Counter {
	return &Counter{0}
}

func (c *Counter) Add(value int64) {
	atomic.AddInt64(&c.num, value)
}

func (c *Counter) Reset() {
	atomic.StoreInt64(&c.num, 0)
}

func (c *Counter) Value() int64 {
	return atomic.LoadInt64(&c.num)
}
