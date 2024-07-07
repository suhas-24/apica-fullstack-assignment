package cache

import (
    "container/list"
    "sync"
    "time"
)

type CacheItem struct {
    Key        string    `json:"key"`
    Value      string    `json:"value"`
    Expiration time.Time `json:"expiration"`
}

type LRUCache struct {
    capacity       int
    items          map[string]*list.Element
    list           *list.List
    mutex          sync.RWMutex
    maxMemory      int64
    currentSize    int64
    expirationQueue *list.List
}

func NewLRUCache(capacity int, maxMemory int64) *LRUCache {
    cache := &LRUCache{
        capacity:        capacity,
        items:           make(map[string]*list.Element),
        list:            list.New(),
        expirationQueue: list.New(),
        maxMemory:       maxMemory,
    }
    go cache.cleanupLoop()
    return cache
}

func (c *LRUCache) cleanupLoop() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        c.cleanup()
    }
}

func (c *LRUCache) cleanup() {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    now := time.Now()
    for e := c.expirationQueue.Front(); e != nil; {
        item := e.Value.(*CacheItem)
        if now.After(item.Expiration) {
            next := e.Next()
            c.removeElement(c.items[item.Key])
            c.expirationQueue.Remove(e)
            e = next
        } else {
            break // Items are sorted by expiration, so we can stop here
        }
    }
}

func (c *LRUCache) Get(key string) (string, bool) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    if elem, ok := c.items[key]; ok {
        item := elem.Value.(*CacheItem)
        if time.Now().Before(item.Expiration) {
            c.list.MoveToFront(elem)
            return item.Value, true
        }
        // Item has expired, remove it
        c.removeElement(elem)
    }
    return "", false
}

func (c *LRUCache) Set(key, value string, expiration time.Duration) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    newSize := int64(len(key) + len(value))
    if c.currentSize+newSize > c.maxMemory {
        c.evict(newSize)
    }

    expirationTime := time.Now().Add(expiration)
    if elem, ok := c.items[key]; ok {
        c.list.MoveToFront(elem)
        item := elem.Value.(*CacheItem)
        c.currentSize += int64(len(value) - len(item.Value))
        item.Value = value
        item.Expiration = expirationTime
        c.updateExpirationQueue(item)
    } else {
        if c.list.Len() >= c.capacity {
            c.removeOldest()
        }
        item := &CacheItem{Key: key, Value: value, Expiration: expirationTime}
        elem := c.list.PushFront(item)
        c.items[key] = elem
        c.currentSize += newSize
        c.insertIntoExpirationQueue(item)
    }
}

func (c *LRUCache) updateExpirationQueue(item *CacheItem) {
    for e := c.expirationQueue.Front(); e != nil; e = e.Next() {
        if e.Value.(*CacheItem) == item {
            c.expirationQueue.Remove(e)
            break
        }
    }
    c.insertIntoExpirationQueue(item)
}

func (c *LRUCache) insertIntoExpirationQueue(item *CacheItem) {
    for e := c.expirationQueue.Back(); e != nil; e = e.Prev() {
        if e.Value.(*CacheItem).Expiration.After(item.Expiration) {
            c.expirationQueue.InsertBefore(item, e)
            return
        }
    }
    c.expirationQueue.PushBack(item)
}

func (c *LRUCache) Delete(key string) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    if elem, ok := c.items[key]; ok {
        c.removeElement(elem)
    }
}

func (c *LRUCache) removeElement(e *list.Element) {
    c.list.Remove(e)
    item := e.Value.(*CacheItem)
    delete(c.items, item.Key)
    c.currentSize -= int64(len(item.Key) + len(item.Value))
}

func (c *LRUCache) removeOldest() {
    elem := c.list.Back()
    if elem != nil {
        c.removeElement(elem)
    }
}

func (c *LRUCache) evict(required int64) {
    for c.currentSize+required > c.maxMemory && c.list.Len() > 0 {
        c.removeOldest()
    }
}

func (c *LRUCache) GetAll() []CacheItem {
    c.mutex.RLock()
    defer c.mutex.RUnlock()

    items := make([]CacheItem, 0, c.list.Len())
    now := time.Now()
    for e := c.list.Front(); e != nil; e = e.Next() {
        item := e.Value.(*CacheItem)
        if now.Before(item.Expiration) {
            items = append(items, *item)
        }
    }
    return items
}
