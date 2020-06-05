package mock

import "sync"

// CacherMock -
type CacherMock struct {
	mut     sync.Mutex
	dataMap map[string]interface{}
}

// NewCacherMock -
func NewCacherMock() *CacherMock {
	return &CacherMock{
		dataMap: make(map[string]interface{}),
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

	return false
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
func (cm *CacherMock) HasOrAdd(key []byte, value interface{}, _ int) (bool, bool) {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	_, ok := cm.dataMap[string(key)]
	if !ok {
		cm.dataMap[string(key)] = value
	}

	return ok, false
}

// Remove -
func (cm *CacherMock) Remove(key []byte) {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	delete(cm.dataMap, string(key))
}

// Keys -
func (cm *CacherMock) Keys() [][]byte {
	panic("implement me")
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
func (cm *CacherMock) RegisterHandler(func(key []byte, value interface{}), string) {
}

// UnRegisterHandler -
func (cm *CacherMock) UnRegisterHandler(string) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (cm *CacherMock) IsInterfaceNil() bool {
	return cm == nil
}
