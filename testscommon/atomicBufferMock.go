package testscommon

import "sync"

type atomicBufferMock struct {
	size int
	buff [][]byte
	mut  sync.RWMutex
}

// NewAtomicBufferMock -
func NewAtomicBufferMock(size uint32) *atomicBufferMock {
	return &atomicBufferMock{
		size: int(size),
		buff: make([][]byte, 0),
	}
}

// Add -
func (abm *atomicBufferMock) Add(rootHash []byte) {
	if len(abm.buff) > abm.size {
		panic("atomicBufferMock.Add: can not add more elements")
	}

	abm.mut.Lock()
	abm.buff = append(abm.buff, rootHash)
	abm.mut.Unlock()
}

// RemoveAll -
func (abm *atomicBufferMock) RemoveAll() [][]byte {
	abm.mut.Lock()
	newBuff := make([][]byte, len(abm.buff))
	copy(newBuff, abm.buff)
	abm.buff = make([][]byte, 0)
	abm.mut.Unlock()

	return newBuff
}

// Len -
func (abm *atomicBufferMock) Len() int {
	abm.mut.RLock()
	defer abm.mut.RUnlock()

	return len(abm.buff)
}

// MaximumSize -
func (abm *atomicBufferMock) MaximumSize() int {
	return abm.size
}

// IsInterfaceNil -
func (abm *atomicBufferMock) IsInterfaceNil() bool {
	return abm == nil
}
