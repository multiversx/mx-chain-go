package economics

import (
	"strconv"

	"github.com/multiversx/mx-chain-go/config"
)

// GetRewardsActiveConfig -
func (ed *economicsData) GetRewardsActiveConfig(epoch uint32) *config.EpochRewardSettings {
	rewardsParams := &config.EpochRewardSettings{}

	cfg := ed.getRewardsConfigForEpoch(epoch)

	rewardsParams.EpochEnable = cfg.rewardsSettingEpoch
	rewardsParams.LeaderPercentage = cfg.leaderPercentage
	rewardsParams.DeveloperPercentage = cfg.developerPercentage
	rewardsParams.ProtocolSustainabilityAddress = cfg.protocolSustainabilityAddress
	rewardsParams.ProtocolSustainabilityPercentage = cfg.protocolSustainabilityPercentage
	rewardsParams.TopUpFactor = cfg.topUpFactor
	rewardsParams.TopUpGradientPoint = cfg.topUpGradientPoint.String()

	return rewardsParams
}

// GetGasLimitSetting -
func (ed *economicsData) GetGasLimitSetting(epoch uint32) *config.GasLimitSetting {
	gasLimitSetting := &config.GasLimitSetting{}

	cfg := ed.getGasConfigForEpoch(epoch)

	gasLimitSetting.EnableEpoch = cfg.gasLimitSettingEpoch
	gasLimitSetting.MaxGasLimitPerBlock = strconv.FormatUint(cfg.maxGasLimitPerBlock, 10)
	gasLimitSetting.MaxGasLimitPerMiniBlock = strconv.FormatUint(cfg.maxGasLimitPerMiniBlock, 10)
	gasLimitSetting.MaxGasLimitPerMetaBlock = strconv.FormatUint(cfg.maxGasLimitPerMetaBlock, 10)
	gasLimitSetting.MaxGasLimitPerMetaMiniBlock = strconv.FormatUint(cfg.maxGasLimitPerMetaMiniBlock, 10)
	gasLimitSetting.MaxGasLimitPerTx = strconv.FormatUint(cfg.maxGasLimitPerTx, 10)
	gasLimitSetting.MinGasLimit = strconv.FormatUint(cfg.minGasLimit, 10)
	gasLimitSetting.ExtraGasLimitGuardedTx = strconv.FormatUint(cfg.extraGasLimitGuardedTx, 10)

	return gasLimitSetting
}
