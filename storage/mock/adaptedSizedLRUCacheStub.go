package mock

// AdaptedSizedLruCacheStub --
type AdaptedSizedLruCacheStub struct {
	AddSizedCalled                 func(key, value interface{}, sizeInBytes int64) bool
	GetCalled                      func(key interface{}) (value interface{}, ok bool)
	ContainsCalled                 func(key interface{}) (ok bool)
	AddSizedIfMissingCalled        func(key, value interface{}, sizeInBytes int64) (ok, evicted bool)
	PeekCalled                     func(key interface{}) (value interface{}, ok bool)
	RemoveCalled                   func(key interface{}) bool
	KeysCalled                     func() []interface{}
	LenCalled                      func() int
	SizeInBytesContainedCalled     func() uint64
	PurgeCalled                    func()
	AddSizedAndReturnEvictedCalled func(key, value interface{}, sizeInBytes int64) map[interface{}]interface{}
}

// AddSized -
func (a *AdaptedSizedLruCacheStub) AddSized(key, value interface{}, sizeInBytes int64) bool {
	if a.AddSizedCalled != nil {
		return a.AddSizedCalled(key, value, sizeInBytes)
	}

	return false
}

// Get -
func (a *AdaptedSizedLruCacheStub) Get(key interface{}) (value interface{}, ok bool) {
	if a.GetCalled != nil {
		return a.GetCalled(key)
	}

	return nil, false
}

// Contains -
func (a *AdaptedSizedLruCacheStub) Contains(key interface{}) (ok bool) {
	if a.ContainsCalled != nil {
		return a.ContainsCalled(key)
	}

	return false
}

// AddSizedIfMissing -
func (a *AdaptedSizedLruCacheStub) AddSizedIfMissing(key, value interface{}, sizeInBytes int64) (ok, evicted bool) {
	if a.AddSizedIfMissingCalled != nil {
		return a.AddSizedIfMissingCalled(key, value, sizeInBytes)
	}

	return false, false
}

// Peek -
func (a *AdaptedSizedLruCacheStub) Peek(key interface{}) (value interface{}, ok bool) {
	if a.PeekCalled != nil {
		return a.PeekCalled(key)
	}

	return nil, false
}

// Remove -
func (a *AdaptedSizedLruCacheStub) Remove(key interface{}) bool {
	if a.RemoveCalled != nil {
		return a.RemoveCalled(key)
	}

	return false
}

// Keys -
func (a *AdaptedSizedLruCacheStub) Keys() []interface{} {
	if a.KeysCalled != nil {
		return a.KeysCalled()
	}

	return nil
}

// Len -
func (a *AdaptedSizedLruCacheStub) Len() int {
	if a.LenCalled != nil {
		return a.LenCalled()
	}

	return 0
}

// SizeInBytesContained -
func (a *AdaptedSizedLruCacheStub) SizeInBytesContained() uint64 {
	if a.SizeInBytesContainedCalled != nil {
		return a.SizeInBytesContainedCalled()
	}

	return 0
}

// Purge -
func (a *AdaptedSizedLruCacheStub) Purge() {
	if a.PurgeCalled != nil {
		a.PurgeCalled()
	}
}

// AddSizedAndReturnEvicted -
func (a *AdaptedSizedLruCacheStub) AddSizedAndReturnEvicted(key, value interface{}, sizeInBytes int64) map[interface{}]interface{} {
	if a.AddSizedAndReturnEvictedCalled != nil {
		return a.AddSizedAndReturnEvictedCalled(key, value, sizeInBytes)
	}

	return nil
}

// IsInterfaceNil -
func (a *AdaptedSizedLruCacheStub) IsInterfaceNil() bool {
	return a == nil
}
