package operationmodes

import (
	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// ProcessHistoricalBalancesMode will process the provided flags for the historical balances
func ProcessHistoricalBalancesMode(log logger.Logger, configs *config.Configs) {
	configs.GeneralConfig.StoragePruning.Enabled = true
	configs.GeneralConfig.StoragePruning.ValidatorCleanOldEpochsData = false
	configs.GeneralConfig.StoragePruning.ObserverCleanOldEpochsData = false
	configs.GeneralConfig.GeneralSettings.StartInEpochEnabled = false
	configs.GeneralConfig.StoragePruning.AccountsTrieCleanOldEpochsData = false
	configs.GeneralConfig.StateTriesConfig.AccountsStatePruningEnabled = false
	configs.GeneralConfig.DbLookupExtensions.Enabled = true
	configs.PreferencesConfig.Preferences.FullArchive = true

	log.Warn("the node is in historical balances mode! Will auto-set some config values",
		"StoragePruning.Enabled", configs.GeneralConfig.StoragePruning.Enabled,
		"StoragePruning.ValidatorCleanOldEpochsData", configs.GeneralConfig.StoragePruning.ValidatorCleanOldEpochsData,
		"StoragePruning.ObserverCleanOldEpochsData", configs.GeneralConfig.StoragePruning.ObserverCleanOldEpochsData,
		"StoragePruning.AccountsTrieCleanOldEpochsData", configs.GeneralConfig.StoragePruning.AccountsTrieCleanOldEpochsData,
		"GeneralSettings.StartInEpochEnabled", configs.GeneralConfig.GeneralSettings.StartInEpochEnabled,
		"StateTriesConfig.AccountsStatePruningEnabled", configs.GeneralConfig.StateTriesConfig.AccountsStatePruningEnabled,
		"DbLookupExtensions.Enabled", configs.GeneralConfig.DbLookupExtensions.Enabled,
		"Preferences.FullArchive", configs.PreferencesConfig.Preferences.FullArchive,
	)
}

// IsInHistoricalBalancesMode returns true if the configuration provided denotes a historical balances mode
func IsInHistoricalBalancesMode(configs *config.Configs) bool {
	return configs.GeneralConfig.StoragePruning.Enabled &&
		!configs.GeneralConfig.StoragePruning.ValidatorCleanOldEpochsData &&
		!configs.GeneralConfig.StoragePruning.ObserverCleanOldEpochsData &&
		!configs.GeneralConfig.GeneralSettings.StartInEpochEnabled &&
		!configs.GeneralConfig.StoragePruning.AccountsTrieCleanOldEpochsData &&
		!configs.GeneralConfig.StateTriesConfig.AccountsStatePruningEnabled &&
		configs.GeneralConfig.DbLookupExtensions.Enabled &&
		configs.PreferencesConfig.Preferences.FullArchive
}
