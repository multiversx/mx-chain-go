package mock

// InterceptedDataStub -
type InterceptedDataStub struct {
	CheckValidityCalled     func() error
	IsForCurrentShardCalled func() bool
	HashCalled              func() []byte
}

// Hash -
func (ids *InterceptedDataStub) Hash() []byte {
	if ids.HashCalled != nil {
		return ids.HashCalled()
	}

	return []byte("mock hash")
}

// CheckValidity -
func (ids *InterceptedDataStub) CheckValidity() error {
	return ids.CheckValidityCalled()
}

// IsForCurrentShard -
func (ids *InterceptedDataStub) IsForCurrentShard() bool {
	return ids.IsForCurrentShardCalled()
}

// Type -
func (ids *InterceptedDataStub) Type() string {
	return "intercepted data stub"
}

// IsInterfaceNil -
func (ids *InterceptedDataStub) IsInterfaceNil() bool {
	return ids == nil
}
