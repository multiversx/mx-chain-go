package triesHolder

import (
	"container/list"

	"github.com/multiversx/mx-chain-go/common"
)

type dataTrieCacheEntry struct {
	key  string
	trie common.Trie
	size uint64
}

// The holder serializes access to dataTrieCache.
type dataTrieCache struct {
	entries map[string]*list.Element
	order   list.List
	size    uint64
	budget  uint64
}

func newDataTrieCache(budget uint64) dataTrieCache {
	return dataTrieCache{entries: make(map[string]*list.Element), budget: budget}
}

func (c *dataTrieCache) put(key string, trie common.Trie, size uint64) {
	element := c.entries[key]
	if element != nil {
		entry := element.Value.(*dataTrieCacheEntry)
		entry.trie = trie
		c.setSize(entry, size)
		c.order.MoveToFront(element)
		return
	}

	entry := &dataTrieCacheEntry{key: key, trie: trie, size: size}
	c.entries[key] = c.order.PushFront(entry)
	c.size += size
}

func (c *dataTrieCache) get(key string) common.Trie {
	element := c.entries[key]
	if element == nil {
		return nil
	}
	c.order.MoveToFront(element)
	return element.Value.(*dataTrieCacheEntry).trie
}

func (c *dataTrieCache) peek(key string) *dataTrieCacheEntry {
	element := c.entries[key]
	if element == nil {
		return nil
	}
	return element.Value.(*dataTrieCacheEntry)
}

func (c *dataTrieCache) setSize(entry *dataTrieCacheEntry, size uint64) {
	c.size = c.size - entry.size + size
	entry.size = size
}

func (c *dataTrieCache) evictOldestOverBudget() *dataTrieCacheEntry {
	if c.size <= c.budget || c.order.Len() <= 1 {
		return nil
	}
	return c.removeElement(c.order.Back())
}

func (c *dataTrieCache) remove(key string) {
	if element := c.entries[key]; element != nil {
		c.removeElement(element)
	}
}

func (c *dataTrieCache) removeElement(element *list.Element) *dataTrieCacheEntry {
	entry := c.order.Remove(element).(*dataTrieCacheEntry)
	delete(c.entries, entry.key)
	c.size -= entry.size
	return entry
}

func (c *dataTrieCache) reset() {
	c.entries = make(map[string]*list.Element)
	c.order.Init()
	c.size = 0
}
