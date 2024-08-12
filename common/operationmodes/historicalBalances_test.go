package operationmodes

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestProcessHistoricalBalancesMode(t *testing.T) {
	t.Parallel()

	cfg := &config.Configs{
		GeneralConfig:     &config.Config{},
		PreferencesConfig: &config.Preferences{},
	}
	ProcessHistoricalBalancesMode(&testscommon.LoggerStub{}, cfg)

	assert.True(t, cfg.GeneralConfig.StoragePruning.Enabled)
	assert.False(t, cfg.GeneralConfig.StoragePruning.ValidatorCleanOldEpochsData)
	assert.False(t, cfg.GeneralConfig.StoragePruning.ObserverCleanOldEpochsData)
	assert.False(t, cfg.GeneralConfig.GeneralSettings.StartInEpochEnabled)
	assert.False(t, cfg.GeneralConfig.StoragePruning.AccountsTrieCleanOldEpochsData)
	assert.False(t, cfg.GeneralConfig.StateTriesConfig.AccountsStatePruningEnabled)
	assert.True(t, cfg.GeneralConfig.DbLookupExtensions.Enabled)
	assert.True(t, cfg.PreferencesConfig.Preferences.FullArchive)
}

func TestIsInHistoricalBalancesMode(t *testing.T) {
	t.Parallel()

	t.Run("empty configs should return false", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Configs{
			GeneralConfig:     &config.Config{},
			PreferencesConfig: &config.Preferences{},
		}
		assert.False(t, IsInHistoricalBalancesMode(cfg))
	})
	t.Run("storage pruning disabled should return false", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Configs{
			GeneralConfig:     &config.Config{},
			PreferencesConfig: &config.Preferences{},
		}
		ProcessHistoricalBalancesMode(&testscommon.LoggerStub{}, cfg)
		cfg.GeneralConfig.StoragePruning.Enabled = false
		assert.False(t, IsInHistoricalBalancesMode(cfg))
	})
	t.Run("validator clean old epoch data enabled should return false", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Configs{
			GeneralConfig:     &config.Config{},
			PreferencesConfig: &config.Preferences{},
		}
		ProcessHistoricalBalancesMode(&testscommon.LoggerStub{}, cfg)
		cfg.GeneralConfig.StoragePruning.ValidatorCleanOldEpochsData = true
		assert.False(t, IsInHistoricalBalancesMode(cfg))
	})
	t.Run("observer clean old epoch data enabled should return false", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Configs{
			GeneralConfig:     &config.Config{},
			PreferencesConfig: &config.Preferences{},
		}
		ProcessHistoricalBalancesMode(&testscommon.LoggerStub{}, cfg)
		cfg.GeneralConfig.StoragePruning.ObserverCleanOldEpochsData = true
		assert.False(t, IsInHistoricalBalancesMode(cfg))
	})
	t.Run("start in epoch enabled should return false", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Configs{
			GeneralConfig:     &config.Config{},
			PreferencesConfig: &config.Preferences{},
		}
		ProcessHistoricalBalancesMode(&testscommon.LoggerStub{}, cfg)
		cfg.GeneralConfig.GeneralSettings.StartInEpochEnabled = true
		assert.False(t, IsInHistoricalBalancesMode(cfg))
	})
	t.Run("accounts trie clean old epoch data enabled should return false", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Configs{
			GeneralConfig:     &config.Config{},
			PreferencesConfig: &config.Preferences{},
		}
		ProcessHistoricalBalancesMode(&testscommon.LoggerStub{}, cfg)
		cfg.GeneralConfig.StoragePruning.AccountsTrieCleanOldEpochsData = true
		assert.False(t, IsInHistoricalBalancesMode(cfg))
	})
	t.Run("accounts state pruning enabled should return false", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Configs{
			GeneralConfig:     &config.Config{},
			PreferencesConfig: &config.Preferences{},
		}
		ProcessHistoricalBalancesMode(&testscommon.LoggerStub{}, cfg)
		cfg.GeneralConfig.StateTriesConfig.AccountsStatePruningEnabled = true
		assert.False(t, IsInHistoricalBalancesMode(cfg))
	})
	t.Run("db lookup extension disabled should return false", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Configs{
			GeneralConfig:     &config.Config{},
			PreferencesConfig: &config.Preferences{},
		}
		ProcessHistoricalBalancesMode(&testscommon.LoggerStub{}, cfg)
		cfg.GeneralConfig.DbLookupExtensions.Enabled = false
		assert.False(t, IsInHistoricalBalancesMode(cfg))
	})
	t.Run("not a full archive node should return false", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Configs{
			GeneralConfig:     &config.Config{},
			PreferencesConfig: &config.Preferences{},
		}
		ProcessHistoricalBalancesMode(&testscommon.LoggerStub{}, cfg)
		cfg.PreferencesConfig.Preferences.FullArchive = false
		assert.False(t, IsInHistoricalBalancesMode(cfg))
	})
	t.Run("with historical balances config should return true", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Configs{
			GeneralConfig:     &config.Config{},
			PreferencesConfig: &config.Preferences{},
		}
		ProcessHistoricalBalancesMode(&testscommon.LoggerStub{}, cfg)
		assert.True(t, IsInHistoricalBalancesMode(cfg))
	})

}
