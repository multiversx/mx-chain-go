package hooks

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// BlockChainHookCounter defines the operations of a blockchain hook counter handler
type BlockChainHookCounter interface {
	ProcessCrtNumberOfTrieReadsCounter() error
	ProcessMaxBuiltInCounters(input *vmcommon.ContractCallInput) error
	ResetCounters()
	SetMaximumValues(mapsOfValues map[string]uint64)
	GetCounterValues() map[string]uint64
	IsInterfaceNil() bool
}

// EpochStartTriggerHandler defines the operations of an epoch start trigger handler needed by the blockchain hook
type EpochStartTriggerHandler interface {
	LastCommitedEpochStartHdr() (data.HeaderHandler, error)
	GetEpochStartHdrFromStorage(epoch uint32) (data.HeaderHandler, error)
	IsInterfaceNil() bool
}

// RoundHandler defines the operations of a round handler needed by the blockchain hook
type RoundHandler interface {
	TimeDuration() time.Duration
	IsInterfaceNil() bool
}
