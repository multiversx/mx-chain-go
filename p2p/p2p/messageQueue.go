package p2p

import (
	"container/list"
	"sync"
)

type MessageQueue struct {
	queue map[string]bool
	list  *list.List

	mut         sync.Mutex
	maxCapacity int
}

func NewMessageQueue(maxCapacity int) *MessageQueue {
	return &MessageQueue{
		queue:       make(map[string]bool),
		list:        list.New(),
		mut:         sync.Mutex{},
		maxCapacity: maxCapacity,
	}
}

func (mq *MessageQueue) Contains(hash string) bool {
	_, ok := mq.queue[hash]
	return ok
}

func (mq *MessageQueue) Add(hash string) {
	mq.mut.Lock()
	defer mq.mut.Unlock()

	if mq.Contains(hash) {
		return
	}

	mq.clean()

	mq.list.PushFront(hash)
	mq.queue[hash] = true
}

func (mq *MessageQueue) Len() int {
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

	if !mq.Contains(elem) {
		panic("Inconsistent queue state. Element" + elem + "was not found!")
	}

	mq.list.Remove(mq.list.Back())
	delete(mq.queue, elem)
}
