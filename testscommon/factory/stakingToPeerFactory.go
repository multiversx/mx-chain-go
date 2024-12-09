package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	processMock "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/scToProtocol"
)

// StakingToPeerFactoryMock -
type StakingToPeerFactoryMock struct {
	CreateStakingToPeerCalled func(args scToProtocol.ArgStakingToPeer) (process.SmartContractToProtocolHandler, error)
}

// CreateStakingToPeer -
func (mock *StakingToPeerFactoryMock) CreateStakingToPeer(args scToProtocol.ArgStakingToPeer) (process.SmartContractToProtocolHandler, error) {
	if mock.CreateStakingToPeerCalled != nil {
		return mock.CreateStakingToPeerCalled(args)
	}

	return &processMock.SCToProtocolStub{}, nil
}

// IsInterfaceNil -
func (mock *StakingToPeerFactoryMock) IsInterfaceNil() bool {
	return mock == nil
}
