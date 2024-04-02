package db

import "sync"

type Data interface {
	*User | *Path
}

type Map[T Data] struct {
	lock *sync.RWMutex
	data map[string]T
}

func NewMap[T Data]() *Map[T] {
	return &Map[T]{data: make(map[string]T), lock: new(sync.RWMutex)}
}
func (m *Map[T]) Get(key string) (T, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if m.data == nil {
		return nil, false
	}
	t, ok := m.data[key]
	return t, ok
}
func (m *Map[T]) Save(key string, t T) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.data == nil {
		m.data = make(map[string]T)
	}
	m.data[key] = t
}
func (m *Map[T]) Clean() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.data = make(map[string]T)
}
