package mock

// InterceptedDataStub -
type InterceptedDataStub struct {
	CheckValidityCalled     func() error
	IsForCurrentShardCalled func() bool
	HashCalled              func() []byte
	TypeCalled              func() string
	IdentifiersCalled       func() [][]byte
	StringCalled            func() string
}

// CheckValidity -
func (ids *InterceptedDataStub) CheckValidity() error {
	if ids.CheckValidityCalled != nil {
		return ids.CheckValidityCalled()
	}

	return nil
}

// IsForCurrentShard -
func (ids *InterceptedDataStub) IsForCurrentShard() bool {
	if ids.IsForCurrentShardCalled != nil {
		return ids.IsForCurrentShardCalled()
	}

	return false
}

// IsInterfaceNil -
func (ids *InterceptedDataStub) IsInterfaceNil() bool {
	return ids == nil
}

// Hash -
func (ids *InterceptedDataStub) Hash() []byte {
	if ids.HashCalled != nil {
		return ids.HashCalled()
	}

	return nil
}

// Type -
func (ids *InterceptedDataStub) Type() string {
	if ids.TypeCalled != nil {
		return ids.TypeCalled()
	}

	return ""
}

// Identifiers -
func (ids *InterceptedDataStub) Identifiers() [][]byte {
	if ids.IdentifiersCalled != nil {
		return ids.IdentifiersCalled()
	}

	return nil
}

// String -
func (ids *InterceptedDataStub) String() string {
	if ids.StringCalled != nil {
		return ids.StringCalled()
	}

	return ""
}
