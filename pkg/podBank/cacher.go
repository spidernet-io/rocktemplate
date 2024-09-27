package podBank

import (
	"container/list"
	"github.com/spidernet-io/rocktemplate/pkg/lock"
)

// Key 结构体，包含 Cgroup 和 Pid，都是 uint32 类型
type Key struct {
	Cgroup, Pid uint32
}

// Value 结构体，包含 Namespace 和 Podname
type Value struct {
	Namespace, Podname string
}

// LimitedStore 是一个有容量限制的键值存储结构
type LimitedStore struct {
	capacity int
	items    map[Key]*list.Element
	order    *list.List
	mutex    lock.RWMutex
}

type entry struct {
	key   Key
	value Value
}

// NewLimitedStore 创建一个新的 LimitedStore
func NewLimitedStore(capacity int) *LimitedStore {
	return &LimitedStore{
		capacity: capacity,
		items:    make(map[Key]*list.Element),
		order:    list.New(),
	}
}

// Set 添加或更新一个键值对
func (s *LimitedStore) Set(key Key, value Value) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if elem, exists := s.items[key]; exists {
		s.order.MoveToBack(elem)
		elem.Value.(*entry).value = value
		return
	}

	if s.order.Len() >= s.capacity {
		oldest := s.order.Front()
		if oldest != nil {
			s.order.Remove(oldest)
			delete(s.items, oldest.Value.(*entry).key)
		}
	}

	elem := s.order.PushBack(&entry{key, value})
	s.items[key] = elem
}

// Get 获取一个键对应的值
func (s *LimitedStore) Get(key Key) (Value, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if elem, exists := s.items[key]; exists {
		return elem.Value.(*entry).value, true
	}
	return Value{}, false
}

// Delete 删除一个键值对
func (s *LimitedStore) Delete(key Key) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if elem, exists := s.items[key]; exists {
		s.order.Remove(elem)
		delete(s.items, key)
	}
}

// Len 返回当前存储的项目数量
func (s *LimitedStore) Len() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.items)
}

// GetAll 返回所有存储的键值对
func (s *LimitedStore) GetAll() map[Key]Value {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make(map[Key]Value, len(s.items))
	for key, elem := range s.items {
		result[key] = elem.Value.(*entry).value
	}
	return result
}
