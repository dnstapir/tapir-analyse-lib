package ringbuf

import (
	"sync"

	"github.com/dnstapir/tapir-analyse-lib/common"
)

type Conf struct {
	Size int
}

type Ringbuf[T any] struct {
	sync.RWMutex
	b     []T
	size  int
	write int
	count int
}

func Create[T any](conf Conf) (*Ringbuf[T], error) {
	r := new(Ringbuf[T])

	if conf.Size <= 0 {
		return nil, common.ErrBadParam
	}

	r.b = make([]T, conf.Size)
	r.size = conf.Size

	return r, nil
}

func (r *Ringbuf[T]) Add(t T) {
	r.Lock()
	defer r.Unlock()

	r.b[r.write] = t
	r.write = (r.write + 1) % r.size

	if r.count < r.size {
		r.count++
	}
}

func (r *Ringbuf[T]) Contents() []T {
	r.RLock()
	defer r.RUnlock()

	result := make([]T, r.count, r.count)

	for i := range r.count {
		idx := (i + r.write - r.count + r.size) % r.size
		result[i] = r.b[idx]
	}

	return result
}
