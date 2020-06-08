package mock

type CacherStub struct {
	HasOrAddCalled func(key []byte, value interface{}, sizeInBytes int) (ok, evicted bool)
}

// HasOrAdd -
func (cs *CacherStub) HasOrAdd(key []byte, value interface{}, sizeInBytes int) (ok, evicted bool) {
	if cs.HasOrAddCalled != nil {
		return cs.HasOrAddCalled(key, value, sizeInBytes)
	}

	return false, false
}

// IsInterfaceNil -
func (cs *CacherStub) IsInterfaceNil() bool {
	return cs == nil
}
