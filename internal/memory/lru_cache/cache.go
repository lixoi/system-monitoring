package lrucache

import (
	"errors"
	"sync"
)

// Key ...
type Key int64

// Cache ...
type Cache interface {
	ResizeCacheOfTime(timeInterval Key) error
	ResizeCacheOfCap(c int) error
	GetCapacity() int
	Set(key Key, value interface{}) bool
	Get(interval Key) ([]interface{}, bool)
	Clear()
}

type lruCache struct {
	timeInterval Key
	mu           sync.Mutex
	capacity     int
	queue        List
	items        map[Key]*ListItem
}

type valueItem struct {
	val interface{}
	k   Key
}

// NewCache ...
func NewCache(capacity int, timeInterval Key) Cache {
	return &lruCache{
		timeInterval: timeInterval,
		capacity:     capacity,
		queue:        NewList(),
		items:        make(map[Key]*ListItem, capacity),
	}
}

func (lc *lruCache) ResizeCacheOfTime(timeInterval Key) error {
	if timeInterval == 0 {
		return errors.New("time intervar is empty")
	}
	lc.timeInterval = timeInterval

	return nil
}

func (lc *lruCache) ResizeCacheOfCap(c int) error {
	if c == 0 {
		return errors.New("capacity is empty")
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()
	for i := 0; i < c-lc.capacity; i++ {
		pntr := lc.queue.Back()
		delete(lc.items, pntr.Value.(valueItem).k)
		lc.queue.Remove(pntr)
	}
	lc.capacity = c

	return nil
}

func (lc *lruCache) GetCapacity() int {
	return lc.capacity
}

func (lc *lruCache) Set(key Key, value interface{}) bool {
	if key == 0 || value == nil {
		return false
	}
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.items[key] = lc.queue.PushFront(valueItem{
		val: value,
		k:   key,
	})

	if lc.timeInterval > 0 {
		minKey := key - lc.timeInterval
		for pntr := lc.queue.Back(); pntr.Value != nil && pntr.Value.(valueItem).k < minKey; {
			delete(lc.items, pntr.Value.(valueItem).k)
			lc.queue.Remove(pntr)
		}
		return true
	}

	if lc.queue.Len() > lc.capacity {
		pntr := lc.queue.Back()
		delete(lc.items, pntr.Value.(valueItem).k)
		lc.queue.Remove(pntr)
	}

	return true
}

func (lc *lruCache) Get(interval Key) ([]interface{}, bool) {
	res := make([]interface{}, 0, 100)

	lc.mu.Lock()
	defer lc.mu.Unlock()

	v := lc.queue.Front()
	if v == nil || v.Value == nil {
		return nil, false
	}
	res = append(res, v.Value.(valueItem).val)

	if interval == 0 {
		return res, true
	}

	frontKey := v.Value.(valueItem).k
	minTime := frontKey - interval
	for k, v := range lc.items {
		if k >= minTime && frontKey != k {
			res = append(res, v.Value.(valueItem).val)
		}
	}

	return res, true
}

func (lc *lruCache) Clear() {
	lc.mu.Lock()
	lc.queue = NewList()
	lc.items = make(map[Key]*ListItem, lc.capacity)
	lc.mu.Unlock()
}
