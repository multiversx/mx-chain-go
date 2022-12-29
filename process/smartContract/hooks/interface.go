package hooks

import vmcommon "github.com/ElrondNetwork/elrond-vm-common"

// BlockChainHookCounter defines the operations of a blockchain hook counter handler
type BlockChainHookCounter interface {
	ProcessCrtNumberOfTrieReadsCounter() error
	ProcessMaxBuiltInCounters(input *vmcommon.ContractCallInput) error
	ResetCounters()
	SetMaximumValues(mapsOfValues map[string]uint64)
	GetCounterValues() map[string]uint64
	IsInterfaceNil() bool
}
