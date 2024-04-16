package sovereign

import (
	sovereignCore "github.com/multiversx/mx-chain-core-go/data/sovereign"
)

// OutGoingOperationsPool defines the behavior of a timed cache for outgoing operations
type OutGoingOperationsPool interface {
	Add(data *sovereignCore.BridgeOutGoingData)
	Get(hash []byte) *sovereignCore.BridgeOutGoingData
	Delete(hash []byte)
	GetUnconfirmedOperations() []*sovereignCore.BridgeOutGoingData
	ConfirmOperation(hashOfHashes []byte, hash []byte) error
	IsInterfaceNil() bool
}
