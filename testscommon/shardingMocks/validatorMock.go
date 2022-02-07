package shardingMocks

const uint32Size = 4
const numUint32 = 2

// ValidatorMock defines a mocked validator
type ValidatorMock struct {
	pubKey  []byte
	chances uint32
	index   uint32

	PubKeyCalled func() []byte
}

// NewValidatorMock creates a new instance of a validator
func NewValidatorMock(pubKey []byte, chances uint32, index uint32) *ValidatorMock {
	return &ValidatorMock{
		pubKey:  pubKey,
		chances: chances,
		index:   index,
	}
}

// PubKey returns the validator's public key
func (v *ValidatorMock) PubKey() []byte {
	if v.PubKeyCalled != nil {
		return v.PubKeyCalled()
	}
	return v.pubKey
}

// Chances returns the validator's chances
func (v *ValidatorMock) Chances() uint32 {
	return v.chances
}

// Index returns the validators index
func (v *ValidatorMock) Index() uint32 {
	return v.index
}

// Size -
func (v *ValidatorMock) Size() int {
	return len(v.pubKey) + uint32Size*numUint32
}

// IsInterfaceNil returns true if there is no value under the interface
func (v *ValidatorMock) IsInterfaceNil() bool {
	return v == nil
}
