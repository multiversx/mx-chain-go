package chainSimulator

import "github.com/multiversx/mx-chain-go/node/chainSimulator/process"

// ChainSimulatorMock -
type ChainSimulatorMock struct {
	GenerateBlocksCalled func(numOfBlocks int) error
	GetNodeHandlerCalled func(shardID uint32) process.NodeHandler
}

// GenerateBlocks -
func (mock *ChainSimulatorMock) GenerateBlocks(numOfBlocks int) error {
	if mock.GenerateBlocksCalled != nil {
		return mock.GenerateBlocksCalled(numOfBlocks)
	}

	return nil
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
