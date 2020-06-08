package mock

type CacherStub struct {
	HasOrAddCalled func(key []byte, value interface{}, sizeInBytes int) (added bool)
}

// HasOrAdd -
func (cs *CacherStub) HasOrAdd(key []byte, value interface{}, sizeInBytes int) (added bool) {
	if cs.HasOrAddCalled != nil {
		return cs.HasOrAddCalled(key, value, sizeInBytes)
	}

	return false
}

// IsInterfaceNil -
func (cs *CacherStub) IsInterfaceNil() bool {
	return cs == nil
}
