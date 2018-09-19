package p2p

import (
	"container/list"
	"sync"
)

type MessageQueue struct {
	queue map[string]bool
	list  *list.List

	mut         sync.RWMutex
	maxCapacity int
}

func NewMessageQueue(maxCapacity int) *MessageQueue {
	return &MessageQueue{
		queue:       make(map[string]bool),
		list:        list.New(),
		mut:         sync.RWMutex{},
		maxCapacity: maxCapacity,
	}
}

func (mq *MessageQueue) ContainsAndAdd(hash string) bool {
	mq.mut.Lock()
	defer mq.mut.Unlock()

	if mq.contains(hash) {
		return true
	}

	mq.add(hash)

	return false
}

func (mq *MessageQueue) Contains(hash string) bool {
	mq.mut.RLock()
	defer mq.mut.RUnlock()

	return mq.contains(hash)
}

func (mq *MessageQueue) contains(hash string) bool {

	_, ok := mq.queue[hash]
	return ok
}

func (mq *MessageQueue) add(hash string) {
	mq.clean()

	mq.list.PushFront(hash)
	mq.queue[hash] = true
}

func (mq *MessageQueue) Add(hash string) {
	mq.mut.Lock()
	defer mq.mut.Unlock()

	if mq.contains(hash) {
		return
	}

	mq.add(hash)
}

func (mq *MessageQueue) Len() int {
	mq.mut.RLock()
	defer mq.mut.RUnlock()

	return mq.list.Len()
}

func (mq *MessageQueue) clean() {
	if mq.list.Len() < mq.maxCapacity {
		return
	}

	if mq.list.Back() == nil {
		return
	}

	elem := mq.list.Back().Value.(string)

	if !mq.contains(elem) {
		panic("Inconsistent queue state. Element" + elem + "was not found!")
	}

	mq.list.Remove(mq.list.Back())
	delete(mq.queue, elem)
}
