package hw04lrucache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type cacheItem struct {
	key   Key
	value interface{}
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
	mutex    sync.Mutex
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (l *lruCache) Set(key Key, value interface{}) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if item, exists := l.items[key]; exists {
		item.Value.(*cacheItem).value = value
		l.queue.MoveToFront(item)
		return true
	}

	cacheElem := &cacheItem{
		key:   key,
		value: value,
	}

	listItem := l.queue.PushFront(cacheElem)
	l.items[key] = listItem

	if l.queue.Len() > l.capacity {
		backItem := l.queue.Back()
		if backItem != nil {
			cacheElem := backItem.Value.(*cacheItem)
			key := cacheElem.key
			delete(l.items, key)
			l.queue.Remove(backItem)
		}
	}
	return false
}

func (l *lruCache) Get(key Key) (interface{}, bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if item, exists := l.items[key]; exists {
		l.queue.MoveToFront(item)
		cacheElem := item.Value.(*cacheItem)
		return cacheElem.value, true
	}
	return nil, false
}

func (l *lruCache) Clear() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.queue = NewList()
	l.items = make(map[Key]*ListItem, l.capacity)
}
