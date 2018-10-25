package p2p

import (
	"container/list"
	"sync"
)

// MessageQueue implements a queue where first string in will be the first string out
// It is concurent safe and has a maxCapacity set
type MessageQueue struct {
	queue map[string]bool
	list  *list.List

	mut         sync.RWMutex
	maxCapacity int
}

// NewMessageQueue creates a new instance of the MessageQueue struct
func NewMessageQueue(maxCapacity int) *MessageQueue {
	return &MessageQueue{
		queue:       make(map[string]bool),
		list:        list.New(),
		mut:         sync.RWMutex{},
		maxCapacity: maxCapacity,
	}
}

// ContainsAndAdd atomically adds the hash if the string was not found. It returns the existence of this string
// before it was added
func (mq *MessageQueue) ContainsAndAdd(hash string) bool {
	mq.mut.Lock()
	defer mq.mut.Unlock()

	_, found := mq.queue[hash]

	if !found {
		//will add
		mq.clean()

		mq.list.PushFront(hash)
		mq.queue[hash] = true
	}

	return found
}

// Contains returns true if the hash is present in the map
func (mq *MessageQueue) Contains(hash string) bool {
	mq.mut.RLock()
	defer mq.mut.RUnlock()

	_, found := mq.queue[hash]

	return found
}

// Len returns the size of this MessageQueue
func (mq *MessageQueue) Len() int {
	mq.mut.RLock()
	defer mq.mut.RUnlock()

	return len(mq.queue)
}

func (mq *MessageQueue) clean() {
	if len(mq.queue) < mq.maxCapacity {
		return
	}

	if mq.list.Back() == nil {
		return
	}

	elem := mq.list.Back().Value.(string)

	mq.list.Remove(mq.list.Back())
	delete(mq.queue, elem)
}
