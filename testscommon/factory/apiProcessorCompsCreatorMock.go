package factory

import (
	"github.com/multiversx/mx-chain-go/factory/processing/api"
	"github.com/multiversx/mx-chain-go/testscommon/stakingcommon"
)

// APIProcessorCompsCreatorMock -
type APIProcessorCompsCreatorMock struct {
	CreateAPICompsCalled func(args api.ArgsCreateAPIProcessComps) (*api.APIProcessComps, error)
}

// CreateAPIComps -
func (c *APIProcessorCompsCreatorMock) CreateAPIComps(args api.ArgsCreateAPIProcessComps) (*api.APIProcessComps, error) {
	if c.CreateAPICompsCalled != nil {
		return c.CreateAPICompsCalled(args)
	}

	return &api.APIProcessComps{
		StakingDataProviderAPI: &stakingcommon.StakingDataProviderStub{},
		AuctionListSelector:    &stakingcommon.AuctionListSelectorStub{},
	}, nil
}

// IsInterfaceNil -
func (c *APIProcessorCompsCreatorMock) IsInterfaceNil() bool {
	return c == nil
}
