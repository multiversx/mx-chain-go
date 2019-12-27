package mock

import "sync"

type AppStatusHandlerMock struct {
	mut  sync.Mutex
	data map[string]interface{}
}

func NewAppStatusHandlerMock() *AppStatusHandlerMock {
	return &AppStatusHandlerMock{
		data: make(map[string]interface{}),
	}
}

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

	ashm.data[key] = vUint64 - 1
}

func (ashm *AppStatusHandlerMock) SetInt64Value(key string, value int64) {
	ashm.mut.Lock()
	defer ashm.mut.Unlock()

	ashm.data[key] = value
}

func (ashm *AppStatusHandlerMock) SetUInt64Value(key string, value uint64) {
	ashm.mut.Lock()
	defer ashm.mut.Unlock()

	ashm.data[key] = value
}

func (ashm *AppStatusHandlerMock) SetStringValue(key string, value string) {
	ashm.mut.Lock()
	defer ashm.mut.Unlock()

	ashm.data[key] = value
}

func (ashm *AppStatusHandlerMock) GetUint64(key string) uint64 {
	ashm.mut.Lock()
	defer ashm.mut.Unlock()

	return ashm.data[key].(uint64)
}

func (ashm *AppStatusHandlerMock) Close() {}

func (ashm *AppStatusHandlerMock) IsInterfaceNil() bool {
	return ashm == nil
}
