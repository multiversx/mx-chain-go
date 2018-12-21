package mock

import (
	"errors"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
)

var errMemoryStorerMock = errors.New("MemoryStorerMock generic error")

type MemoryStorerMock struct {
	db   map[string][]byte
	lock sync.RWMutex
	Fail bool
}

func NewMemoryStorerMock() *MemoryStorerMock {
	return &MemoryStorerMock{
		db: make(map[string][]byte),
	}
}

func (msm *MemoryStorerMock) Put(key, data []byte) error {
	if msm.Fail {
		return errMemoryStorerMock
	}

	msm.lock.Lock()
	defer msm.lock.Unlock()

	msm.db[string(key)] = encoding.CopyBytes(data)
	return nil
}

func (msm *MemoryStorerMock) Get(key []byte) ([]byte, error) {
	if msm.Fail {
		return nil, errMemoryStorerMock
	}

	msm.lock.RLock()
	defer msm.lock.RUnlock()

	if entry, ok := msm.db[string(key)]; ok {
		return encoding.CopyBytes(entry), nil
	}
	return nil, errors.New("not found")
}

func (msm *MemoryStorerMock) Has(key []byte) (bool, error) {
	if msm.Fail {
		return false, errMemoryStorerMock
	}

	msm.lock.RLock()
	defer msm.lock.RUnlock()

	_, ok := msm.db[string(key)]
	return ok, nil
}

func (msm *MemoryStorerMock) HasOrAdd(key []byte, value []byte) (bool, error) {
	if msm.Fail {
		return false, errMemoryStorerMock
	}

	msm.lock.RLock()
	defer msm.lock.RUnlock()

	_, ok := msm.db[string(key)]
	if ok {
		return true, nil
	}

	msm.db[string(key)] = encoding.CopyBytes(value)
	return false, nil
}

func (msm *MemoryStorerMock) Remove(key []byte) error {
	if msm.Fail {
		return errMemoryStorerMock
	}

	msm.lock.RLock()
	defer msm.lock.RUnlock()

	_, ok := msm.db[string(key)]
	if ok {
		delete(msm.db, string(key))
	}

	return nil
}

func (msm *MemoryStorerMock) ClearCache() {
}

func (msm *MemoryStorerMock) DestroyUnit() error {
	if msm.Fail {
		return errMemoryStorerMock
	}

	msm.lock.Lock()
	defer msm.lock.Unlock()

	msm.db = make(map[string][]byte, 0)
	return nil
}
