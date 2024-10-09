package factory

import (
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/metachain"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// RewardsCreatorFactoryMock -
type RewardsCreatorFactoryMock struct {
	CreateRewardsCreatorCalled func(args metachain.RewardsCreatorProxyArgs) (epochStart.RewardsCreator, error)
}

// CreateRewardsCreator creates a rewards creator proxy for normal run type chain
func (mock *RewardsCreatorFactoryMock) CreateRewardsCreator(args metachain.RewardsCreatorProxyArgs) (epochStart.RewardsCreator, error) {
	if mock.CreateRewardsCreatorCalled != nil {
		return mock.CreateRewardsCreatorCalled(args)
	}

	return &testscommon.RewardsCreatorStub{}, nil
}

// IsInterfaceNil returns nil if the underlying object is nil
func (mock *RewardsCreatorFactoryMock) IsInterfaceNil() bool {
	return mock == nil
}
