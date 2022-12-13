package testscommon

import vmcommon "github.com/ElrondNetwork/elrond-vm-common"

// BlockChainHookCounterStub -
type BlockChainHookCounterStub struct {
	ProcessCrtNumberOfTrieReadsCounterCalled func() error
	ProcessMaxBuiltInCountersCalled          func(input *vmcommon.ContractCallInput) error
	ResetCountersCalled                      func()
	SetMaximumValuesCalled                   func(mapsOfValues map[string]uint64)
}

// ProcessCrtNumberOfTrieReadsCounter -
func (stub *BlockChainHookCounterStub) ProcessCrtNumberOfTrieReadsCounter() error {
	if stub.ProcessCrtNumberOfTrieReadsCounterCalled != nil {
		return stub.ProcessCrtNumberOfTrieReadsCounterCalled()
	}

	return nil
}

// ProcessMaxBuiltInCounters -
func (stub *BlockChainHookCounterStub) ProcessMaxBuiltInCounters(input *vmcommon.ContractCallInput) error {
	if stub.ProcessMaxBuiltInCountersCalled != nil {
		return stub.ProcessMaxBuiltInCountersCalled(input)
	}

	return nil
}

// ResetCounters -
func (stub *BlockChainHookCounterStub) ResetCounters() {
	if stub.ResetCountersCalled != nil {
		stub.ResetCountersCalled()
	}
}

// SetMaximumValues -
func (stub *BlockChainHookCounterStub) SetMaximumValues(mapsOfValues map[string]uint64) {
	if stub.SetMaximumValuesCalled != nil {
		stub.SetMaximumValuesCalled(mapsOfValues)
	}
}

// IsInterfaceNil -
func (stub *BlockChainHookCounterStub) IsInterfaceNil() bool {
	return stub == nil
}
