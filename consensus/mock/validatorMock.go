package mock

const uint32Size = 4
const numUint32 = 2

type validator struct {
	pubKey  []byte
	chances uint32
	index   uint32
}

// NewValidator creates a new instance of a validator
func NewValidator(pubKey []byte, chances uint32, index uint32) *validator {
	return &validator{
		pubKey:  pubKey,
		chances: chances,
		index:   index,
	}
}

// PubKey returns the validator's public key
func (v *validator) PubKey() []byte {
	return v.pubKey
}

// Chances returns the validator's chances
func (v *validator) Chances() uint32 {
	return v.chances
}

// Index returns the validators index
func (v *validator) Index() uint32 {
	return v.index
}

// Size -
func (v *validator) Size() int {
	return len(v.pubKey) + uint32Size*numUint32
}

// IsInterfaceNil returns true if there is no value under the interface
func (v *validator) IsInterfaceNil() bool {
	return v == nil
}
