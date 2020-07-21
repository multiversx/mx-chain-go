package testscommon

import (
	"sync"
)

// CacherMock -
type CacherMock struct {
	mut                  sync.RWMutex
	dataMap              map[string]interface{}
	mutAddedDataHandlers sync.RWMutex
	addedDataHandlers    []func(key []byte, val interface{})
}

// NewCacherMock -
func NewCacherMock() *CacherMock {
	return &CacherMock{
		dataMap:           make(map[string]interface{}),
		addedDataHandlers: make([]func(key []byte, val interface{}), 0),
	}
}

// Clear -
func (cacher *CacherMock) Clear() {
	cacher.mut.Lock()
	defer cacher.mut.Unlock()

	cacher.dataMap = make(map[string]interface{})
}

// Put -
func (cacher *CacherMock) Put(key []byte, value interface{}, _ int) (evicted bool) {
	cacher.mut.Lock()
	defer cacher.mut.Unlock()

	cacher.dataMap[string(key)] = value
	cacher.callAddedDataHandlers(key, value)

	return false
}

func (cacher *CacherMock) callAddedDataHandlers(key []byte, val interface{}) {
	cacher.mutAddedDataHandlers.RLock()
	for _, handler := range cacher.addedDataHandlers {
		go handler(key, val)
	}
	cacher.mutAddedDataHandlers.RUnlock()
}

// Get -
func (cacher *CacherMock) Get(key []byte) (value interface{}, ok bool) {
	cacher.mut.RLock()
	defer cacher.mut.RUnlock()

	val, ok := cacher.dataMap[string(key)]

	return val, ok
}

// Has -
func (cacher *CacherMock) Has(key []byte) bool {
	cacher.mut.RLock()
	defer cacher.mut.RUnlock()

	_, ok := cacher.dataMap[string(key)]

	return ok
}

// Peek -
func (cacher *CacherMock) Peek(key []byte) (value interface{}, ok bool) {
	cacher.mut.RLock()
	defer cacher.mut.RUnlock()

	val, ok := cacher.dataMap[string(key)]

	return val, ok
}

// HasOrAdd -
func (cacher *CacherMock) HasOrAdd(key []byte, value interface{}, _ int) (has, added bool) {
	cacher.mut.Lock()
	defer cacher.mut.Unlock()

	_, has = cacher.dataMap[string(key)]
	if has {
		return true, false
	}

	cacher.dataMap[string(key)] = value
	cacher.callAddedDataHandlers(key, value)
	return false, true
}

// Remove -
func (cacher *CacherMock) Remove(key []byte) {
	cacher.mut.Lock()
	defer cacher.mut.Unlock()

	delete(cacher.dataMap, string(key))
}

// Keys -
func (cacher *CacherMock) Keys() [][]byte {
	keys := make([][]byte, len(cacher.dataMap))
	idx := 0
	for k := range cacher.dataMap {
		keys[idx] = []byte(k)
		idx++
	}

	return keys
}

// Len -
func (cacher *CacherMock) Len() int {
	cacher.mut.RLock()
	defer cacher.mut.RUnlock()

	return len(cacher.dataMap)
}

// MaxSize -
func (cacher *CacherMock) MaxSize() int {
	return 10000
}

// RegisterHandler -
func (cacher *CacherMock) RegisterHandler(handler func(key []byte, value interface{}), _ string) {
	if handler == nil {
		return
	}

	cacher.mutAddedDataHandlers.Lock()
	cacher.addedDataHandlers = append(cacher.addedDataHandlers, handler)
	cacher.mutAddedDataHandlers.Unlock()
}

// UnRegisterHandler -
func (cacher *CacherMock) UnRegisterHandler(string) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (cacher *CacherMock) IsInterfaceNil() bool {
	return cacher == nil
}
