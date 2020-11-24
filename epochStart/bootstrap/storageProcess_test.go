package bootstrap

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func createMockStorageEpochStartBootstrapArgs(
	coreMock *mock.CoreComponentsMock,
	cryptoMock *mock.CryptoComponentsMock,
) ArgsStorageEpochStartBootstrap {
	return ArgsStorageEpochStartBootstrap{
		ArgsEpochStartBootstrap: createMockEpochStartBootstrapArgs(coreMock, cryptoMock),
		ImportDbConfig:          config.ImportDbConfig{},
		ChanGracefullyClose:     make(chan endProcess.ArgEndProcess, 1),
		TimeToWait:              time.Second,
	}
}

func TestNewStorageEpochStartBootstrap_InvalidArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	coreComp.Hash = nil
	args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
	sesb, err := NewStorageEpochStartBootstrap(args)
	assert.True(t, check.IfNil(sesb))
	assert.True(t, errors.Is(err, epochStart.ErrNilHasher))

	coreComp, cryptoComp = createComponentsForEpochStart()
	args = createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.ChanGracefullyClose = nil
	sesb, err = NewStorageEpochStartBootstrap(args)
	assert.True(t, check.IfNil(sesb))
	assert.True(t, errors.Is(err, dataRetriever.ErrNilGracefullyCloseChannel))
}

func TestNewStorageEpochStartBootstrap_ShouldWork(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
	sesb, err := NewStorageEpochStartBootstrap(args)
	assert.False(t, check.IfNil(sesb))
	assert.Nil(t, err)
}

func TestStorageEpochStartBootstrap_BootstrapStartInEpochNotEnabled(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)

	err := errors.New("localErr")
	args.LatestStorageDataProvider = &mock.LatestStorageDataProviderStub{
		GetCalled: func() (storage.LatestDataFromStorage, error) {
			return storage.LatestDataFromStorage{}, err
		},
	}
	sesb, _ := NewStorageEpochStartBootstrap(args)

	params, err := sesb.Bootstrap()
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), params.Epoch)
}

func TestStorageEpochStartBootstrap_BootstrapFromGenesis(t *testing.T) {
	roundsPerEpoch := int64(100)
	roundDuration := uint64(60000)
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.EconomicsData = &mock.EconomicsHandlerStub{
		MinGasPriceCalled: func() uint64 {
			return 1
		},
	}
	args.GenesisNodesConfig = &mock.NodesSetupStub{
		GetRoundDurationCalled: func() uint64 {
			return roundDuration
		},
	}
	args.GeneralConfig = testscommon.GetGeneralConfig()
	args.GeneralConfig.EpochStartConfig.RoundsPerEpoch = roundsPerEpoch
	sesb, _ := NewStorageEpochStartBootstrap(args)

	params, err := sesb.Bootstrap()
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), params.Epoch)
}

func TestStorageEpochStartBootstrap_BootstrapMetablockNotFound(t *testing.T) {
	roundsPerEpoch := int64(100)
	roundDuration := uint64(6000)
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.EconomicsData = &mock.EconomicsHandlerStub{
		MinGasPriceCalled: func() uint64 {
			return 1
		},
	}
	args.GenesisNodesConfig = &mock.NodesSetupStub{
		GetRoundDurationCalled: func() uint64 {
			return roundDuration
		},
	}
	args.Rounder = &mock.RounderStub{
		RoundIndex: 2*roundsPerEpoch + 1,
	}
	args.GeneralConfig = testscommon.GetGeneralConfig()
	args.GeneralConfig.EpochStartConfig.RoundsPerEpoch = roundsPerEpoch
	sesb, _ := NewStorageEpochStartBootstrap(args)

	params, err := sesb.Bootstrap()
	assert.Equal(t, process.ErrNilMetaBlockHeader, err)
	assert.Equal(t, uint32(0), params.Epoch)
}
