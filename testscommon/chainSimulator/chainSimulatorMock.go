package chainSimulator

import (
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"math/big"
)

// ChainSimulatorMock -
type ChainSimulatorMock struct {
	GenerateBlocksCalled               func(numOfBlocks int) error
	GetNodeHandlerCalled               func(shardID uint32) process.NodeHandler
	GenerateAddressInShardCalled       func(providedShardID uint32) dtos.WalletAddress
	GenerateAndMintWalletAddressCalled func(targetShardID uint32, value *big.Int) (dtos.WalletAddress, error)
}

// GenerateAddressInShard -
func (mock *ChainSimulatorMock) GenerateAddressInShard(providedShardID uint32) dtos.WalletAddress {
	if mock.GenerateAddressInShardCalled != nil {
		return mock.GenerateAddressInShardCalled(providedShardID)
	}

	return dtos.WalletAddress{}
}

// GenerateAndMintWalletAddress -
func (mock *ChainSimulatorMock) GenerateAndMintWalletAddress(targetShardID uint32, value *big.Int) (dtos.WalletAddress, error) {
	if mock.GenerateAndMintWalletAddressCalled != nil {
		return mock.GenerateAndMintWalletAddressCalled(targetShardID, value)
	}
	return dtos.WalletAddress{}, nil
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
