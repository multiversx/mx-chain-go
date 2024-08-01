package metachain

import "github.com/multiversx/mx-chain-go/epochStart"

type sovereignRewardsCreatorFactory struct {
}

// NewSovereignRewardsCreatorFactory creates a rewards creator factory for sovereign run type chain
func NewSovereignRewardsCreatorFactory() *sovereignRewardsCreatorFactory {
	return &sovereignRewardsCreatorFactory{}
}

// CreateRewardsCreator creates a rewards creator proxy for sovereign run type chain
func (f *sovereignRewardsCreatorFactory) CreateRewardsCreator(args RewardsCreatorProxyArgs) (epochStart.RewardsCreator, error) {
	argsV2 := RewardsCreatorArgsV2{
		BaseRewardsCreatorArgs: args.BaseRewardsCreatorArgs,
		StakingDataProvider:    args.StakingDataProvider,
		EconomicsDataProvider:  args.EconomicsDataProvider,
		RewardsHandler:         args.RewardsHandler,
	}

	rc, err := NewRewardsCreatorV2(argsV2)
	if err != nil {
		return nil, err
	}

	return NewSovereignRewards(rc)
}

// IsInterfaceNil returns nil if the underlying object is nil
func (f *sovereignRewardsCreatorFactory) IsInterfaceNil() bool {
	return f == nil
}
