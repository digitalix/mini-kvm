package concurrents

import (
	"slices"
	"sync"
	"sync/atomic"
)

type Slice[T comparable] struct {
	size  atomic.Int64
	mutex sync.RWMutex
	data  []T
}

func NewSlice[T comparable](initial ...T) *Slice[T] {
	return &Slice[T]{
		data: initial,
	}
}

func (s *Slice[T]) Add(t T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data = append(s.data, t)
	s.size.Add(1)
}
func (s *Slice[T]) Remove(index int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data = slices.Delete(s.data, index, index)
	s.size.Add(-1)
}

func (s *Slice[T]) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data = s.data[:0]
	s.size.Store(0)
}

func (s *Slice[T]) Iterate(yield func(T) bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for _, t := range s.data {
		if !yield(t) {
			return
		}
	}
}

func (s *Slice[T]) Values() []T {
	values := make([]T, s.size.Load())
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	copy(values, s.data)
	return values
}

func (s *Slice[T]) Size(t T) int {
	return int(s.size.Load())
}

func (s *Slice[T]) Index(t T) int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for i, v := range s.data {
		if t == v {
			return i
		}
	}

	return -1
}
