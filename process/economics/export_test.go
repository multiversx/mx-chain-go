package economics

import (
	"github.com/ElrondNetwork/elrond-go/config"
)

// GetRewardsActiveConfig -
func (ed *economicsData) GetRewardsActiveConfig() *config.EpochRewardSettings {
	rewardsParams := &config.EpochRewardSettings{}

	ed.mutRewardsSettings.RLock()
	defer ed.mutRewardsSettings.RUnlock()

	rewardsParams.EpochEnable = ed.rewardsSettingEpoch
	rewardsParams.LeaderPercentage = ed.leaderPercentage
	rewardsParams.DeveloperPercentage = ed.developerPercentage
	rewardsParams.ProtocolSustainabilityAddress = ed.protocolSustainabilityAddress
	rewardsParams.ProtocolSustainabilityPercentage = ed.protocolSustainabilityPercentage
	rewardsParams.TopUpFactor = ed.topUpFactor
	rewardsParams.TopUpGradientPoint = ed.topUpGradientPoint.String()

	return rewardsParams
}
