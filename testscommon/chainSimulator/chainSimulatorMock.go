package chainSimulator

import "github.com/multiversx/mx-chain-go/node/chainSimulator/process"

// ChainSimulatorMock -
type ChainSimulatorMock struct {
	GetNodeHandlerCalled func(shardID uint32) process.NodeHandler
}

// GetNodeHandler -
func (mock *ChainSimulatorMock) GetNodeHandler(shardID uint32) process.NodeHandler {
	if mock.GetNodeHandlerCalled != nil {
		return mock.GetNodeHandlerCalled(shardID)
	}
	return nil
}

// IsInterfaceNil -
func (mock *ChainSimulatorMock) IsInterfaceNil() bool {
	return mock == nil
}
