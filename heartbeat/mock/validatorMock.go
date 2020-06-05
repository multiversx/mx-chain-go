package mock

const uint32Size = 4
const numUint32 = 2

// ValidatorMock -
type ValidatorMock struct {
	pubKey  []byte
	chances uint32
	index   uint32
}

// NewValidatorMock -
func NewValidatorMock(pubKey []byte, chances uint32, index uint32) *ValidatorMock {
	return &ValidatorMock{pubKey: pubKey, index: index, chances: chances}
}

// PubKey -
func (vm *ValidatorMock) PubKey() []byte {
	return vm.pubKey
}

// Chances -
func (vm *ValidatorMock) Chances() uint32 {
	return vm.chances
}

// Size -
func (vm *ValidatorMock) Size() int {
	return len(vm.pubKey) + uint32Size*numUint32
}

// Index -
func (vm *ValidatorMock) Index() uint32 {
	return vm.index
}
