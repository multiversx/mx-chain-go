package memorydb

import (
	"sync"
)

type Batch struct {
	batch map[string][]byte
	mutx  sync.RWMutex
}

func NewBatch() *Batch {
	return &Batch{
		batch: make(map[string][]byte),
		mutx:  sync.RWMutex{},
	}
}

func (b *Batch) Put(key []byte, val []byte) error {
	b.mutx.Lock()
	defer b.mutx.Unlock()

	b.batch[string(key)] = val

	return nil
}

func (b *Batch) Delete(key []byte) error {
	b.mutx.Lock()
	defer b.mutx.Unlock()

	delete(b.batch, string(key))

	return nil
}

func (b *Batch) Reset() {
	b.mutx.Lock()
	defer b.mutx.Unlock()

	b.batch = make(map[string][]byte)
}
