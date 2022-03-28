package leveldb

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/storage"
)

type batchWrapper struct {
	mut   sync.RWMutex
	batch storage.Batcher
	size  int
}

func newBatchWrapper() *batchWrapper {
	return &batchWrapper{
		batch: NewBatch(),
	}
}

func (bw *batchWrapper) updateBatchSizeReturningCurrent() int {
	bw.mut.Lock()
	defer bw.mut.Unlock()

	bw.size++

	return bw.size
}

func (bw *batchWrapper) resetBatchReturningLast() storage.Batcher {
	bw.mut.Lock()
	defer bw.mut.Unlock()

	lastBatch := bw.batch
	bw.batch = NewBatch()
	bw.size = 0

	return lastBatch
}

func (bw *batchWrapper) put(key []byte, val []byte) error {
	bw.mut.RLock()
	defer bw.mut.RUnlock()

	return bw.batch.Put(key, val)
}

func (bw *batchWrapper) delete(key []byte) {
	bw.mut.Lock() // TODO evaluate if we can use RLock here, the pointer does not change, same as in Put operation
	defer bw.mut.Unlock()

	_ = bw.batch.Delete(key)
}

func (bw *batchWrapper) get(key []byte) []byte {
	bw.mut.RLock()
	defer bw.mut.RUnlock()

	return bw.batch.Get(key)
}
