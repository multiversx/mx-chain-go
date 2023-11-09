package mock

import "github.com/multiversx/mx-chain-go/common"

// KeyBuilderStub -
type KeyBuilderStub struct {
	BuildKeyCalled func(keyPart []byte)
	GetKeyCalled   func() ([]byte, error)
	CloneCalled    func() common.KeyBuilder
}

// BuildKey -
func (stub *KeyBuilderStub) BuildKey(keyPart []byte) {
	if stub.BuildKeyCalled != nil {
		stub.BuildKeyCalled(keyPart)
	}
}

// GetKey -
func (stub *KeyBuilderStub) GetKey() ([]byte, error) {
	if stub.GetKeyCalled != nil {
		return stub.GetKeyCalled()
	}

	return []byte{}, nil
}

// Clone -
func (stub *KeyBuilderStub) Clone() common.KeyBuilder {
	if stub.CloneCalled != nil {
		return stub.CloneCalled()
	}

	return &KeyBuilderStub{}
}

// IsInterfaceNil -
func (stub *KeyBuilderStub) IsInterfaceNil() bool {
	return stub == nil
}
