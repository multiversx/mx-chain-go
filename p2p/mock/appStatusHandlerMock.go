package mock

import "sync"

// AppStatusHandlerMock -
type AppStatusHandlerMock struct {
	mut  sync.Mutex
	data map[string]interface{}
}

// NewAppStatusHandlerMock -
func NewAppStatusHandlerMock() *AppStatusHandlerMock {
	return &AppStatusHandlerMock{
		data: make(map[string]interface{}),
	}
}

// Increment -
func (ashm *AppStatusHandlerMock) Increment(key string) {
	ashm.mut.Lock()
	defer ashm.mut.Unlock()

	v, ok := ashm.data[key]
	if !ok {
		return
	}

	vUint64, ok := v.(uint64)
	if !ok {
		return
	}

	ashm.data[key] = vUint64 + 1
}

// AddUint64 -
func (ashm *AppStatusHandlerMock) AddUint64(key string, val uint64) {
	ashm.mut.Lock()
	defer ashm.mut.Unlock()

	v, ok := ashm.data[key]
	if !ok {
		return
	}

	vUint64, ok := v.(uint64)
	if !ok {
		return
	}

	ashm.data[key] = vUint64 + val
}

// Decrement -
func (ashm *AppStatusHandlerMock) Decrement(key string) {
	ashm.mut.Lock()
	defer ashm.mut.Unlock()

	v, ok := ashm.data[key]
	if !ok {
		return
	}

	vUint64, ok := v.(uint64)
	if !ok {
		return
	}

	if vUint64 != 0 {
		ashm.data[key] = vUint64 - 1
	}
}

// SetInt64Value -
func (ashm *AppStatusHandlerMock) SetInt64Value(key string, value int64) {
	ashm.mut.Lock()
	defer ashm.mut.Unlock()

	ashm.data[key] = value
}

// SetUInt64Value -
func (ashm *AppStatusHandlerMock) SetUInt64Value(key string, value uint64) {
	ashm.mut.Lock()
	defer ashm.mut.Unlock()

	ashm.data[key] = value
}

// SetStringValue -
func (ashm *AppStatusHandlerMock) SetStringValue(key string, value string) {
	ashm.mut.Lock()
	defer ashm.mut.Unlock()

	ashm.data[key] = value
}

// GetUint64 -
func (ashm *AppStatusHandlerMock) GetUint64(key string) uint64 {
	ashm.mut.Lock()
	defer ashm.mut.Unlock()

	return ashm.data[key].(uint64)
}

// Close -
func (ashm *AppStatusHandlerMock) Close() {}

// IsInterfaceNil returns true if there is no value under the interface
func (ashm *AppStatusHandlerMock) IsInterfaceNil() bool {
	return ashm == nil
}
