package economics

import "math/big"

// SetRewardsParameter -
type SetRewardsParameter struct {
	RewardsSettingEpoch              uint32
	LeaderPercentage                 float64
	ProtocolSustainabilityPercentage float64
	ProtocolSustainabilityAddress    string
	DeveloperPercentage              float64
	TopUpGradientPoint               *big.Int
	TopUpFactor                      float64
}

// GetRewardsSetting -
func (ed *economicsData) GetRewardsSetting() *SetRewardsParameter {
	rewardsParams := &SetRewardsParameter{}

	ed.mutRewardsSettings.RLock()
	defer ed.mutRewardsSettings.RUnlock()

	rewardsParams.RewardsSettingEpoch = ed.rewardsSettingEpoch
	rewardsParams.LeaderPercentage = ed.leaderPercentage
	rewardsParams.DeveloperPercentage = ed.developerPercentage
	rewardsParams.ProtocolSustainabilityAddress = ed.protocolSustainabilityAddress
	rewardsParams.ProtocolSustainabilityPercentage = ed.protocolSustainabilityPercentage
	rewardsParams.TopUpFactor = ed.topUpFactor
	rewardsParams.TopUpGradientPoint = big.NewInt(0).Set(ed.topUpGradientPoint)

	return rewardsParams
}
