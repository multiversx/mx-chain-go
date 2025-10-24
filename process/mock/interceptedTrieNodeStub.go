package mock

// InterceptedTrieNodeStub -
type InterceptedTrieNodeStub struct {
	CheckValidityCalled         func() error
	IsForCurrentShardCalled     func() bool
	SizeInBytesCalled           func() int
	ShouldAllowDuplicatesCalled func() bool
	HashField                   []byte
	StringField                 string
	IdentifiersField            [][]byte
	TypeField                   string
}

// CheckValidity -
func (ins *InterceptedTrieNodeStub) CheckValidity() error {
	if ins.CheckValidityCalled != nil {
		return ins.CheckValidityCalled()
	}

	return nil
}

// ShouldAllowDuplicates -
func (ins *InterceptedTrieNodeStub) ShouldAllowDuplicates() bool {
	if ins.ShouldAllowDuplicatesCalled != nil {
		return ins.ShouldAllowDuplicatesCalled()
	}

	return true
}

// IsForCurrentShard -
func (ins *InterceptedTrieNodeStub) IsForCurrentShard() bool {
	if ins.IsForCurrentShardCalled != nil {
		return ins.IsForCurrentShardCalled()
	}

	return true
}

// SizeInBytes -
func (ins *InterceptedTrieNodeStub) SizeInBytes() int {
	if ins.SizeInBytesCalled != nil {
		return ins.SizeInBytesCalled()
	}

	return 0
}

// Hash -
func (ins *InterceptedTrieNodeStub) Hash() []byte {
	return ins.HashField
}

// Type -
func (ins *InterceptedTrieNodeStub) Type() string {
	panic("implement me")
}

// Identifiers -
func (ins *InterceptedTrieNodeStub) Identifiers() [][]byte {
	return ins.IdentifiersField
}

// String -
func (ins *InterceptedTrieNodeStub) String() string {
	return ins.StringField
}

// IsInterfaceNil -
func (ins *InterceptedTrieNodeStub) IsInterfaceNil() bool {
	return ins == nil
}
