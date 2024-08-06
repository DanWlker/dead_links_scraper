package pkg

import (
	"sync"
)

type AtomicMap[T comparable, U any] struct {
	sync.RWMutex
	data map[T]U
}

func (a *AtomicMap[T, U]) Set(key T, val U) {
	a.Lock()
	// fmt.Println("Locking")
	defer a.Unlock()
	// func() {
	// fmt.Println("Unlocking")
	// }()
	a.data[key] = val
}

func (a *AtomicMap[T, U]) Get(key T) (U, bool) {
	res, ok := a.data[key]
	return res, ok
}

func (a *AtomicMap[T, U]) Range(rangeFunc func(T, U)) {
	// fmt.Println("Ranging")
	for key, val := range a.data {
		rangeFunc(key, val)
	}
	// fmt.Println("Ranging done")
}

func NewAtomicMap[T comparable, U any]() *AtomicMap[T, U] {
	return &AtomicMap[T, U]{
		data: make(map[T]U),
	}
}
