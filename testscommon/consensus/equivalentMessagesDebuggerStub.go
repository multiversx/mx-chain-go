package consensus

import "github.com/multiversx/mx-chain-core-go/data"

// EquivalentMessagesDebuggerStub -
type EquivalentMessagesDebuggerStub struct {
	DisplayEquivalentMessagesStatisticsCalled func()
}

// DisplayEquivalentMessagesStatistics -
func (stub *EquivalentMessagesDebuggerStub) DisplayEquivalentMessagesStatistics() {
	if stub.DisplayEquivalentMessagesStatisticsCalled != nil {
		stub.DisplayEquivalentMessagesStatisticsCalled()
	}
}

// SetValidEquivalentProof -
func (stub *EquivalentMessagesDebuggerStub) SetValidEquivalentProof(
	headerHash []byte,
	proof data.HeaderProofHandler,
) {
}

// UpsertEquivalentMessage -
func (stub *EquivalentMessagesDebuggerStub) UpsertEquivalentMessage(headerHash []byte) {}

func (stub *EquivalentMessagesDebuggerStub) ResetEquivalentMessages() {}

func (stub *EquivalentMessagesDebuggerStub) DeleteEquivalentMessage(headerHash []byte) {}

// IsInterfaceNil -
func (stub *EquivalentMessagesDebuggerStub) IsInterfaceNil() bool {
	return stub == nil
}
