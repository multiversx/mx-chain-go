package factory_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/data/state"
	errorsErd "github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/require"
)

const testHasher = "blake2b"
const testMarshalizer = "json"
const signedBlocksThreshold = 0.025
const consecutiveMissedBlocksPenalty = 1.1

func TestNewCoreComponentsFactory_OkValuesShouldWork(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	ccf, _ := factory.NewCoreComponentsFactory(args)

	require.NotNil(t, ccf)
}

func TestCoreComponentsFactory_CreateCoreComponents_NoHasherConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           testMarshalizer,
			SizeCheckDelta: 0,
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsErd.ErrHasherCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_InvalidHasherConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           testMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: "invalid_type",
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsErd.ErrHasherCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_NoInternalMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Hasher: config.TypeConfig{
			Type: testHasher,
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsErd.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_InvalidInternalMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           "invalid_marshalizer_type",
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: testHasher,
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsErd.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_NoVmMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Hasher: config.TypeConfig{
			Type: testHasher,
		},
		Marshalizer: config.MarshalizerConfig{
			Type:           testMarshalizer,
			SizeCheckDelta: 0,
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsErd.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_InvalidVmMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           testMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: testHasher,
		},
		VmMarshalizer: config.TypeConfig{
			Type: "invalid",
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsErd.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_NoTxSignMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Hasher: config.TypeConfig{
			Type: testHasher,
		},
		Marshalizer: config.MarshalizerConfig{
			Type:           testMarshalizer,
			SizeCheckDelta: 0,
		},
		VmMarshalizer: config.TypeConfig{
			Type: testMarshalizer,
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsErd.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_InvalidTxSignMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           testMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: testHasher,
		},
		VmMarshalizer: config.TypeConfig{
			Type: testMarshalizer,
		},
		TxSignMarshalizer: config.TypeConfig{
			Type: "invalid",
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsErd.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidValPubKeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config.ValidatorPubkeyConverter.Type = "invalid"
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, state.ErrInvalidPubkeyConverterType))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidAddrPubKeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config.AddressPubkeyConverter.Type = "invalid"
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, state.ErrInvalidPubkeyConverterType))
}

func TestCoreComponentsFactory_CreateCoreComponents_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

// ------------ Test CoreComponents --------------------
func TestCoreComponents_Close_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	ccf, _ := factory.NewCoreComponentsFactory(args)
	cc, _ := ccf.Create()
	err := cc.Close()

	require.NoError(t, err)
}

func getEpochStartConfig() config.EpochStartConfig {
	return config.EpochStartConfig{
		MinRoundsBetweenEpochs:            20,
		RoundsPerEpoch:                    20,
		MaxShuffledOutRestartThreshold:    0.2,
		MinShuffledOutRestartThreshold:    0.1,
		MinNumConnectedPeersToStart:       2,
		MinNumOfPeersToConsiderBlockValid: 2,
	}
}

func getCoreArgs() factory.CoreComponentsFactoryArgs {
	return factory.CoreComponentsFactoryArgs{
		Config: config.Config{
			EpochStartConfig: getEpochStartConfig(),
			PublicKeyPeerId: config.CacheConfig{
				Type:     "LRU",
				Capacity: 5000,
				Shards:   16,
			},
			PublicKeyShardId: config.CacheConfig{
				Type:     "LRU",
				Capacity: 5000,
				Shards:   16,
			},
			PeerIdShardId: config.CacheConfig{
				Type:     "LRU",
				Capacity: 5000,
				Shards:   16,
			},
			PeerHonesty: config.CacheConfig{
				Type:     "LRU",
				Capacity: 5000,
				Shards:   16,
			},
			GeneralSettings: config.GeneralSettingsConfig{
				ChainID:                  "undefined",
				MinTransactionVersion:    1,
				GenesisMaxNumberOfShards: 3,
			},
			Marshalizer: config.MarshalizerConfig{
				Type:           testMarshalizer,
				SizeCheckDelta: 0,
			},
			Hasher: config.TypeConfig{
				Type: testHasher,
			},
			VmMarshalizer: config.TypeConfig{
				Type: testMarshalizer,
			},
			TxSignMarshalizer: config.TypeConfig{
				Type: testMarshalizer,
			},
			TxSignHasher: config.TypeConfig{
				Type: testHasher,
			},
			AddressPubkeyConverter: config.PubkeyConfig{
				Length:          32,
				Type:            "bech32",
				SignatureLength: 0,
			},
			ValidatorPubkeyConverter: config.PubkeyConfig{
				Length:          96,
				Type:            "hex",
				SignatureLength: 48,
			},
			Consensus: config.ConsensusConfig{
				Type: "bls",
			},
			ValidatorStatistics: config.ValidatorStatisticsConfig{
				CacheRefreshIntervalInSec: uint32(100),
			},
			SoftwareVersionConfig: config.SoftwareVersionConfig{
				PollingIntervalInMinutes: 30,
			},
			Versions: config.VersionsConfig{
				DefaultVersion:   "1",
				VersionsByEpochs: nil,
				Cache: config.CacheConfig{
					Type:     "LRU",
					Capacity: 1000,
					Shards:   1,
				},
			},
		},
		RatingsConfig:         createDummyRatingsConfig(),
		EconomicsConfig:       createDummyEconomicsConfig(),
		NodesFilename:         "mock/testdata/nodesSetupMock.json",
		WorkingDirectory:      "home",
		ChanStopNodeProcess:   make(chan endProcess.ArgEndProcess),
		StatusHandlersFactory: &mock.StatusHandlersFactoryMock{},
	}
}

func createDummyEconomicsConfig() config.EconomicsConfig {
	return config.EconomicsConfig{
		GlobalSettings: config.GlobalSettings{
			GenesisTotalSupply: "20000000000000000000000000",
			MinimumInflation:   0,
			YearSettings: []*config.YearSetting{
				{
					Year:             0,
					MaximumInflation: 0.01,
				},
			},
		},
		RewardsSettings: config.RewardsSettings{
			LeaderPercentage:                 0.1,
			ProtocolSustainabilityPercentage: 0.1,
			ProtocolSustainabilityAddress:    "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
			TopUpFactor:                      0.25,
			TopUpGradientPoint:               "3000000000000000000000000",
		},
		FeeSettings: config.FeeSettings{
			MaxGasLimitPerBlock:     "1500000000",
			MaxGasLimitPerMetaBlock: "15000000000",
			MinGasPrice:             "1000000000",
			MinGasLimit:             "50000",
			GasPerDataByte:          "1500",
			GasPriceModifier:        1,
		},
	}
}

func createDummyRatingsConfig() config.RatingsConfig {
	return config.RatingsConfig{
		General: config.General{
			StartRating:           5000001,
			MaxRating:             10000000,
			MinRating:             1,
			SignedBlocksThreshold: signedBlocksThreshold,
			SelectionChances: []*config.SelectionChance{
				{MaxThreshold: 0, ChancePercent: 5},
				{MaxThreshold: 2500000, ChancePercent: 19},
				{MaxThreshold: 7500000, ChancePercent: 20},
				{MaxThreshold: 10000000, ChancePercent: 21},
			},
		},
		ShardChain: config.ShardChain{
			RatingSteps: config.RatingSteps{
				HoursToMaxRatingFromStartRating: 2,
				ProposerValidatorImportance:     1,
				ProposerDecreaseFactor:          -4,
				ValidatorDecreaseFactor:         -4,
				ConsecutiveMissedBlocksPenalty:  consecutiveMissedBlocksPenalty,
			},
		},
		MetaChain: config.MetaChain{
			RatingSteps: config.RatingSteps{
				HoursToMaxRatingFromStartRating: 2,
				ProposerValidatorImportance:     1,
				ProposerDecreaseFactor:          -4,
				ValidatorDecreaseFactor:         -4,
				ConsecutiveMissedBlocksPenalty:  consecutiveMissedBlocksPenalty,
			},
		},
	}
}
