package mock

import "github.com/multiversx/mx-chain-go/common"

// KeyBuilderStub -
type KeyBuilderStub struct {
	BuildKeyCalled     func(keyPart []byte)
	GetKeyCalled       func() ([]byte, error)
	GetRawKeyCalled    func() []byte
	ShallowCloneCalled func() common.KeyBuilder
	DeepCloneCalled    func() common.KeyBuilder
	SizeCalled         func() uint
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

// GetRawKey -
func (stub *KeyBuilderStub) GetRawKey() []byte {
	if stub.GetRawKeyCalled != nil {
		return stub.GetRawKeyCalled()
	}

	return []byte{}
}

// ShallowClone -
func (stub *KeyBuilderStub) ShallowClone() common.KeyBuilder {
	if stub.ShallowCloneCalled != nil {
		return stub.ShallowCloneCalled()
	}

	return &KeyBuilderStub{}
}

// DeepClone -
func (stub *KeyBuilderStub) DeepClone() common.KeyBuilder {
	if stub.DeepCloneCalled != nil {
		return stub.DeepCloneCalled()
	}

	return &KeyBuilderStub{}
}

// Size -
func (stub *KeyBuilderStub) Size() uint {
	if stub.SizeCalled != nil {
		return stub.SizeCalled()
	}

	return 0
}

// IsInterfaceNil -
func (stub *KeyBuilderStub) IsInterfaceNil() bool {
	return stub == nil
}
