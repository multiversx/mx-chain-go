package metachain

import "github.com/multiversx/mx-chain-go/epochStart"

type rewardsCreatorFactory struct {
}

// NewRewardsCreatorFactory creates a rewards creator factory for normal run type chain
func NewRewardsCreatorFactory() *rewardsCreatorFactory {
	return &rewardsCreatorFactory{}
}

// CreateRewardsCreator creates a rewards creator proxy for normal run type chain
func (f *rewardsCreatorFactory) CreateRewardsCreator(args RewardsCreatorProxyArgs) (epochStart.RewardsCreator, error) {
	return NewRewardsCreatorProxy(args)
}

// IsInterfaceNil returns nil if the underlying object is nil
func (f *rewardsCreatorFactory) IsInterfaceNil() bool {
	return f == nil
}
