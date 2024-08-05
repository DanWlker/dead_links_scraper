package pkg

import "sync"

type AtomicSet[T comparable] struct {
	sync.Mutex
	data map[T]struct{}
}

func (a *AtomicSet[T]) Insert(val T) bool {
	a.Lock()
	_, ok := a.data[val]
	a.data[val] = struct{}{}
	a.Unlock()
	return !ok // will return false if already inside
}

func NewAtomicSet[T comparable]() *AtomicSet[T] {
	return &AtomicSet[T]{
		data: make(map[T]struct{}),
	}
}
