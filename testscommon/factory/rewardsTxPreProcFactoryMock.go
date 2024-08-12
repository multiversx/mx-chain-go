package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
)

// RewardsTxPreProcFactoryMock -
type RewardsTxPreProcFactoryMock struct {
	CreateRewardsTxPreProcessorAndAddToContainerCalled func(args preprocess.ArgsRewardTxPreProcessor, container process.PreProcessorsContainer) error
}

// CreateRewardsTxPreProcessorAndAddToContainer -
func (mock *RewardsTxPreProcFactoryMock) CreateRewardsTxPreProcessorAndAddToContainer(args preprocess.ArgsRewardTxPreProcessor, container process.PreProcessorsContainer) error {
	if mock.CreateRewardsTxPreProcessorAndAddToContainerCalled != nil {
		return mock.CreateRewardsTxPreProcessorAndAddToContainerCalled(args, container)
	}

	return nil
}

// IsInterfaceNil -
func (mock *RewardsTxPreProcFactoryMock) IsInterfaceNil() bool {
	return mock == nil
}
