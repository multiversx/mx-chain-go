package mock

import (
	"github.com/multiversx/mx-chain-go/consensus"
)

// EquivalentMessagesDebuggerStub -
type EquivalentMessagesDebuggerStub struct {
	DisplayEquivalentMessagesStatisticsCalled func(getDataHandler func() map[string]*consensus.EquivalentMessageInfo)
}

// DisplayEquivalentMessagesStatistics -
func (stub *EquivalentMessagesDebuggerStub) DisplayEquivalentMessagesStatistics(getDataHandler func() map[string]*consensus.EquivalentMessageInfo) {
	if stub.DisplayEquivalentMessagesStatisticsCalled != nil {
		stub.DisplayEquivalentMessagesStatisticsCalled(getDataHandler)
	}
}

// IsInterfaceNil -
func (stub *EquivalentMessagesDebuggerStub) IsInterfaceNil() bool {
	return stub == nil
}
