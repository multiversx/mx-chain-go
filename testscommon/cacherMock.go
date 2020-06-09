package testscommon

import (
	"sync"
)

// CacherMock -
type CacherMock struct {
	mut                  sync.Mutex
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
func (cm *CacherMock) Clear() {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	cm.dataMap = make(map[string]interface{})
}

// Put -
func (cm *CacherMock) Put(key []byte, value interface{}, _ int) (evicted bool) {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	cm.dataMap[string(key)] = value
	cm.callAddedDataHandlers(key, value)

	return false
}

func (cm *CacherMock) callAddedDataHandlers(key []byte, val interface{}) {
	cm.mutAddedDataHandlers.RLock()
	for _, handler := range cm.addedDataHandlers {
		go handler(key, val)
	}
	cm.mutAddedDataHandlers.RUnlock()
}

// Get -
func (cm *CacherMock) Get(key []byte) (value interface{}, ok bool) {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	val, ok := cm.dataMap[string(key)]

	return val, ok
}

// Has -
func (cm *CacherMock) Has(key []byte) bool {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	_, ok := cm.dataMap[string(key)]

	return ok
}

// Peek -
func (cm *CacherMock) Peek(key []byte) (value interface{}, ok bool) {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	val, ok := cm.dataMap[string(key)]

	return val, ok
}

// HasOrAdd -
func (cm *CacherMock) HasOrAdd(key []byte, value interface{}, _ int) (added bool) {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	_, ok := cm.dataMap[string(key)]
	if ok {
		return false
	}

	cm.dataMap[string(key)] = value
	cm.callAddedDataHandlers(key, value)
	return true
}

// Remove -
func (cm *CacherMock) Remove(key []byte) {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	delete(cm.dataMap, string(key))
}

// Keys -
func (cm *CacherMock) Keys() [][]byte {
	keys := make([][]byte, len(cm.dataMap))
	idx := 0
	for k := range cm.dataMap {
		keys[idx] = []byte(k)
		idx++
	}

	return keys
}

// Len -
func (cm *CacherMock) Len() int {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	return len(cm.dataMap)
}

// MaxSize -
func (cm *CacherMock) MaxSize() int {
	return 10000
}

// RegisterHandler -
func (cm *CacherMock) RegisterHandler(handler func(key []byte, value interface{}), _ string) {
	if handler == nil {
		return
	}

	cm.mutAddedDataHandlers.Lock()
	cm.addedDataHandlers = append(cm.addedDataHandlers, handler)
	cm.mutAddedDataHandlers.Unlock()
}

// UnRegisterHandler -
func (cm *CacherMock) UnRegisterHandler(string) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (cm *CacherMock) IsInterfaceNil() bool {
	return cm == nil
}
