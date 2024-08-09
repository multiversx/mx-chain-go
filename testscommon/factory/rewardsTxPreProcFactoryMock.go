package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	processMock "github.com/multiversx/mx-chain-go/process/mock"
)

// RewardsTxPreProcFactoryMock -
type RewardsTxPreProcFactoryMock struct {
	CreateRewardsTxPreProcessorCalled func(args preprocess.ArgsRewardTxPreProcessor) (process.PreProcessor, error)
}

// CreateRewardsTxPreProcessor -
func (mock *RewardsTxPreProcFactoryMock) CreateRewardsTxPreProcessor(args preprocess.ArgsRewardTxPreProcessor) (process.PreProcessor, error) {
	if mock.CreateRewardsTxPreProcessorCalled != nil {
		return mock.CreateRewardsTxPreProcessorCalled(args)
	}

	return &processMock.PreProcessorMock{}, nil
}

// IsInterfaceNil -
func (mock *RewardsTxPreProcFactoryMock) IsInterfaceNil() bool {
	return mock == nil
}
