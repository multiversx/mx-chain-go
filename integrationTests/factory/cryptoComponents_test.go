package factory

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/stretchr/testify/require"
)

// ------------ Test CryptoComponents --------------------
func TestCryptoComponents_Create_ShouldWork(t *testing.T) {
	t.Skip()

	generalConfig, _ := loadMainConfig(configPath)
	systemSCConfig, _ := loadSystemSmartContractsConfig(systemSCConfigPath)
	ratingsConfig, _ := loadRatingsConfig(ratingsPath)
	economicsConfig, _ := loadEconomicsConfig(economicsPath)

	ccf, _ := createCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)

	cryptoComponents, err := createCryptoComponents(*generalConfig, *systemSCConfig, ccf)

	require.Nil(t, err)
	require.NotNil(t, cryptoComponents)
}

func TestCryptoComponents_Create_Close_ShouldWork(t *testing.T) {
	t.Skip()

	generalConfig, _ := loadMainConfig(configPath)
	systemSCConfig, _ := loadSystemSmartContractsConfig(systemSCConfigPath)
	ratingsConfig, _ := loadRatingsConfig(ratingsPath)
	economicsConfig, _ := loadEconomicsConfig(economicsPath)

	ccf, _ := createCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	time.Sleep(2 * time.Second)

	nrBefore := runtime.NumGoroutine()
	cryptoComponents, _ := createCryptoComponents(*generalConfig, *systemSCConfig, ccf)
	err := cryptoComponents.Close()
	time.Sleep(2 * time.Second)
	nrAfter := runtime.NumGoroutine()

	require.Nil(t, err)
	if nrBefore != nrAfter {
		printStack()
	}

	require.Equal(t, nrBefore, nrAfter)
}

func createCryptoComponents(
	generalConfig config.Config,
	systemSCConfig config.SystemSmartContractsConfig,
	managedCoreComponents factory.CoreComponentsHandler) (factory.CryptoComponentsHandler, error) {
	cryptoComponentsHandlerArgs := factory.CryptoComponentsFactoryArgs{
		ValidatorKeyPemFileName:              "validatorKey.pem",
		SkIndex:                              0,
		Config:                               generalConfig,
		CoreComponentsHolder:                 managedCoreComponents,
		ActivateBLSPubKeyMessageVerification: systemSCConfig.StakingSystemSCConfig.ActivateBLSPubKeyMessageVerification,
		KeyLoader:                            &core.KeyLoader{},
	}

	cryptoComponentsFactory, err := factory.NewCryptoComponentsFactory(cryptoComponentsHandlerArgs)
	if err != nil {
		return nil, fmt.Errorf("NewCryptoComponentsFactory failed: %w", err)
	}

	managedCryptoComponents, err := factory.NewManagedCryptoComponents(cryptoComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedCryptoComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedCryptoComponents, nil
}
