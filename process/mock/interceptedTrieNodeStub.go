package mock

// InterceptedTrieNodeStub -
type InterceptedTrieNodeStub struct {
	CheckValidityCalled     func() error
	IsForCurrentShardCalled func() bool
	SizeInBytesCalled       func() int
	HashField               []byte
	StringField             string
	SerializedNode          []byte
	IdentifiersField        [][]byte
	TypeField               string
}

// CheckValidity -
func (ins *InterceptedTrieNodeStub) CheckValidity() error {
	if ins.CheckValidityCalled != nil {
		return ins.CheckValidityCalled()
	}

	return nil
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

// EncodedNode -
func (ins *InterceptedTrieNodeStub) EncodedNode() []byte {
	return ins.SerializedNode
}

// IsInterfaceNil -
func (ins *InterceptedTrieNodeStub) IsInterfaceNil() bool {
	return ins == nil
}
