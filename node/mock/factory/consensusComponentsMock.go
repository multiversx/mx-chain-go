package factory

import (
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
)

type ConsensusComponentsStub struct {
	ChronologyHandler         consensus.ChronologyHandler
	ConsensusWorkerHandler    factory.ConsensusWorker
	BroadcastMessengerHandler consensus.BroadcastMessenger
	GroupSize                 int
	BootstrapperHandler       process.Bootstrapper
}

// Create -
func (ccs *ConsensusComponentsStub) Create() error {
	return nil
}

// Close -
func (ccs *ConsensusComponentsStub) Close() error {
	return nil
}

// CheckSubcomponents -
func (ccs *ConsensusComponentsStub) CheckSubcomponents() error {
	return nil
}

// String -
func (ccs *ConsensusComponentsStub) String() string {
	return "ConsensusComponentsMock"
}

// Chronology -
func (ccs *ConsensusComponentsStub) Chronology() consensus.ChronologyHandler {
	return ccs.ChronologyHandler
}

// ConsensusWorker -
func (ccs *ConsensusComponentsStub) ConsensusWorker() factory.ConsensusWorker {
	return ccs.ConsensusWorkerHandler
}

// BroadcastMessenger -
func (ccs *ConsensusComponentsStub) BroadcastMessenger() consensus.BroadcastMessenger {
	return ccs.BroadcastMessengerHandler
}

// Bootstrapper -
func (ccs *ConsensusComponentsStub) Bootstrapper() process.Bootstrapper {
	return ccs.BootstrapperHandler
}

// ConsensusGroupSize -
func (ccs *ConsensusComponentsStub) ConsensusGroupSize() (int, error) {
	return ccs.GroupSize, nil
}

// IsInterfaceNil -
func (ccs *ConsensusComponentsStub) IsInterfaceNil() bool {
	return ccs == nil
}
