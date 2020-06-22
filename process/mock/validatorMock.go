package mock

// ValidatorMock -
type ValidatorMock struct {
	pubKey  []byte
	chances uint32
	index   uint32

	PubKeyCalled func() []byte
}

// NewValidatorMock -
func NewValidatorMock(pubkey []byte) *ValidatorMock {
	return &ValidatorMock{
		pubKey: pubkey,
	}
}

// PubKey -
func (vm *ValidatorMock) PubKey() []byte {
	if vm.PubKeyCalled != nil {
		return vm.PubKeyCalled()
	}
	return vm.pubKey
}

// Chances -
func (vm *ValidatorMock) Chances() uint32 {
	return vm.chances
}

// Size -
func (vm *ValidatorMock) Size() int {
	return len(vm.pubKey) + 8
}

// Index -
func (vm *ValidatorMock) Index() uint32 {
	return vm.index
}
