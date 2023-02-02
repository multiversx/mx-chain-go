package economics

import (
	"strconv"

	"github.com/multiversx/mx-chain-go/config"
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

// GetGasLimitSetting -
func (ed *economicsData) GetGasLimitSetting() *config.GasLimitSetting {
	gasLimitSetting := &config.GasLimitSetting{}

	ed.mutGasLimitSettings.RLock()
	defer ed.mutGasLimitSettings.RUnlock()

	gasLimitSetting.EnableEpoch = ed.gasLimitSettingEpoch
	gasLimitSetting.MaxGasLimitPerBlock = strconv.FormatUint(ed.maxGasLimitPerBlock, 10)
	gasLimitSetting.MaxGasLimitPerMiniBlock = strconv.FormatUint(ed.maxGasLimitPerMiniBlock, 10)
	gasLimitSetting.MaxGasLimitPerMetaBlock = strconv.FormatUint(ed.maxGasLimitPerMetaBlock, 10)
	gasLimitSetting.MaxGasLimitPerMetaMiniBlock = strconv.FormatUint(ed.maxGasLimitPerMetaMiniBlock, 10)
	gasLimitSetting.MaxGasLimitPerTx = strconv.FormatUint(ed.maxGasLimitPerTx, 10)
	gasLimitSetting.MinGasLimit = strconv.FormatUint(ed.minGasLimit, 10)
	gasLimitSetting.ExtraGasLimitGuardedTx = strconv.FormatUint(ed.extraGasLimitGuardedTx, 10)

	return gasLimitSetting
}
