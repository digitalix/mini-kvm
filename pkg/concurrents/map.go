package concurrents

import (
	"sync"
	"sync/atomic"
)

type Map[K comparable, V any] struct {
	data sync.Map
	size atomic.Int64
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{}
}

func (m *Map[K, V]) Load(key K) (v V, _ bool) {
	value, exists := m.data.Load(key)
	if !exists {
		return v, false
	}

	return value.(V), true
}

func (m *Map[K, V]) LoadOrStore(key K, val V) (v V, loaded bool) {
	defer func() {
		if !loaded {
			m.size.Add(1)
		}
	}()
	value, loaded := m.data.LoadOrStore(key, val)
	if !loaded {
		return v, loaded
	}

	return value.(V), loaded
}

func (m *Map[K, V]) Set(key K, value V) {
	if _, loaded := m.data.Swap(key, value); !loaded {
		m.size.Add(1)
	}
}

func (m *Map[K, V]) Delete(key K) {
	if _, loaded := m.data.LoadAndDelete(key); loaded {
		m.size.Add(-1)
	}
}

func (m *Map[K, V]) Clear() {
	m.data.Clear()
	m.size.Store(0)
}

func (m *Map[K, V]) Size() int {
	return int(m.size.Load())
}

func (m *Map[K, V]) Iterate(yield func(K, V) bool) {
	m.data.Range(func(key, value interface{}) bool {
		return yield(key.(K), value.(V))
	})
}

func (m *Map[K, V]) Values() []V {
	result := make([]V, 0, m.size.Load())
	for _, v := range m.Iterate {
		result = append(result, v)
	}

	return result
}

func (m *Map[K, V]) Keys() []K {
	result := make([]K, 0, m.size.Load())
	for k, _ := range m.Iterate {
		result = append(result, k)
	}

	return result
}
