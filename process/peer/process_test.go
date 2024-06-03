package peer_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

const (
	validatorIncreaseRatingStep     = int32(1)
	validatorDecreaseRatingStep     = int32(-2)
	proposerIncreaseRatingStep      = int32(2)
	proposerDecreaseRatingStep      = int32(-4)
	metaValidatorIncreaseRatingStep = int32(3)
	metaValidatorDecreaseRatingStep = int32(-4)
	metaProposerIncreaseRatingStep  = int32(5)
	metaProposerDecreaseRatingStep  = int32(-10)
	minRating                       = uint32(1)
	maxRating                       = uint32(100)
	startRating                     = uint32(50)
	defaultChancesSelection         = uint32(1)
	consensusGroupFormat            = "%s_%v_%v_%v"
)

func createMockPubkeyConverter() *testscommon.PubkeyConverterMock {
	return testscommon.NewPubkeyConverterMock(32)
}

func createMockArguments() peer.ArgValidatorStatisticsProcessor {
	argsNewEconomicsData := economics.ArgsNewEconomicsData{
		Economics: &config.EconomicsConfig{
			GlobalSettings: config.GlobalSettings{
				GenesisTotalSupply: "2000000000000000000000",
				MinimumInflation:   0,
				YearSettings: []*config.YearSetting{
					{
						Year:             0,
						MaximumInflation: 0.01,
					},
				},
			},
			RewardsSettings: config.RewardsSettings{
				RewardsConfigByEpoch: []config.EpochRewardSettings{
					{
						LeaderPercentage:                 0.1,
						DeveloperPercentage:              0.1,
						ProtocolSustainabilityPercentage: 0.1,
						ProtocolSustainabilityAddress:    "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
						TopUpGradientPoint:               "300000000000000000000",
						TopUpFactor:                      0.25,
					},
				},
			},
			FeeSettings: config.FeeSettings{
				GasLimitSettings: []config.GasLimitSetting{
					{
						MaxGasLimitPerBlock:         "10000000",
						MaxGasLimitPerMiniBlock:     "10000000",
						MaxGasLimitPerMetaBlock:     "10000000",
						MaxGasLimitPerMetaMiniBlock: "10000000",
						MaxGasLimitPerTx:            "10000000",
						MinGasLimit:                 "10",
						ExtraGasLimitGuardedTx:      "50000",
					},
				},
				MinGasPrice:            "10",
				GasPerDataByte:         "1",
				GasPriceModifier:       1.0,
				MaxGasPriceSetGuardian: "100000",
			},
		},
		EpochNotifier:       &epochNotifier.EpochNotifierStub{},
		EnableEpochsHandler: enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
		TxVersionChecker:    &testscommon.TxVersionCheckerStub{},
	}
	economicsData, _ := economics.NewEconomicsData(argsNewEconomicsData)

	arguments := peer.ArgValidatorStatisticsProcessor{
		Marshalizer: &mock.MarshalizerMock{},
		DataPool: &dataRetrieverMock.PoolsHolderStub{
			HeadersCalled: func() dataRetriever.HeadersPool {
				return nil
			},
		},
		StorageService:                       &storageStubs.ChainStorerStub{},
		NodesCoordinator:                     &shardingMocks.NodesCoordinatorMock{},
		ShardCoordinator:                     mock.NewOneShardCoordinatorMock(),
		PubkeyConv:                           createMockPubkeyConverter(),
		PeerAdapter:                          getAccountsMock(),
		Rater:                                createMockRater(),
		RewardsHandler:                       economicsData,
		MaxComputableRounds:                  1000,
		MaxConsecutiveRoundsOfRatingDecrease: 2000,
		NodesSetup:                           &genesisMocks.NodesSetupStub{},
		EnableEpochsHandler:                  enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.SwitchJailWaitingFlag, common.BelowSignedThresholdFlag),
	}
	return arguments
}

func createMockRater() *mock.RaterMock {
	rater := mock.GetNewMockRater()
	rater.MinRating = minRating
	rater.MaxRating = maxRating
	rater.StartRating = startRating
	rater.IncreaseProposer = proposerIncreaseRatingStep
	rater.DecreaseProposer = proposerDecreaseRatingStep
	rater.IncreaseValidator = validatorIncreaseRatingStep
	rater.DecreaseValidator = validatorDecreaseRatingStep
	rater.MetaIncreaseProposer = metaProposerIncreaseRatingStep
	rater.MetaDecreaseProposer = metaProposerDecreaseRatingStep
	rater.MetaIncreaseValidator = metaValidatorIncreaseRatingStep
	rater.MetaDecreaseValidator = metaValidatorDecreaseRatingStep
	return rater
}

func createMockCache() map[string]data.HeaderHandler {
	return make(map[string]data.HeaderHandler)
}

func TestNewValidatorStatisticsProcessor_NilPeerAdaptersShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.PeerAdapter = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilPeerAccountsAdapter, err)
}

func TestNewValidatorStatisticsProcessor_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.PubkeyConv = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewValidatorStatisticsProcessor_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.NodesCoordinator = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilNodesCoordinator, err)
}

func TestNewValidatorStatisticsProcessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.ShardCoordinator = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewValidatorStatisticsProcessor_NilStorageShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.StorageService = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewValidatorStatisticsProcessor_ZeroMaxComputableRoundsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.MaxComputableRounds = 0
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrZeroMaxComputableRounds, err)
}

func TestNewValidatorStatisticsProcessor_ZeroMaxConsecutiveRoundsOfRatingDecreaseShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.MaxConsecutiveRoundsOfRatingDecrease = 0
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrZeroMaxConsecutiveRoundsOfRatingDecrease, err)
}

func TestNewValidatorStatisticsProcessor_NilRaterShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.Rater = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilRater, err)
}

func TestNewValidatorStatisticsProcessor_NilRewardsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.RewardsHandler = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilRewardsHandler, err)
}
func TestNewValidatorStatisticsProcessor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.Marshalizer = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewValidatorStatisticsProcessor_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.DataPool = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewValidatorStatisticsProcessor_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.EnableEpochsHandler = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestNewValidatorStatisticsProcessor_InvalidEnableEpochsHandlerhouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined()
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
}

func TestNewValidatorStatisticsProcessor(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.NotNil(t, validatorStatistics)
	assert.Nil(t, err)
}

func TestValidatorStatisticsProcessor_SaveInitialStateErrOnGetAccountFail(t *testing.T) {
	t.Parallel()

	adapterError := errors.New("account error")
	peerAdapters := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return nil, adapterError
		},
	}

	arguments := createMockArguments()
	arguments.PeerAdapter = peerAdapters
	arguments.NodesSetup = &genesisMocks.NodesSetupStub{InitialNodesInfoCalled: func() (m map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, m2 map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
		oneMap := make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
		oneMap[0] = append(oneMap[0], mock.NewNodeInfo([]byte("aaaa"), []byte("aaaa"), 0, 50))
		return oneMap, oneMap
	}}

	_, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Equal(t, adapterError, err)
}

func TestValidatorStatisticsProcessor_SaveInitialStateGetAccountReturnsInvalid(t *testing.T) {
	t.Parallel()

	peerAdapter := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return &stateMock.AccountWrapMock{}, nil
		},
	}

	arguments := createMockArguments()
	arguments.PeerAdapter = peerAdapter
	arguments.NodesSetup = &genesisMocks.NodesSetupStub{InitialNodesInfoCalled: func() (m map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, m2 map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
		oneMap := make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
		oneMap[0] = append(oneMap[0], mock.NewNodeInfo([]byte("aaaa"), []byte("aaaa"), 0, 50))
		return oneMap, oneMap
	}}
	_, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Equal(t, process.ErrInvalidPeerAccount, err)
}

func TestValidatorStatisticsProcessor_SaveInitialStateSetAddressErrors(t *testing.T) {
	t.Parallel()

	saveAccountError := errors.New("save account error")
	peerAccount, _ := accounts.NewPeerAccount([]byte("1234"))
	peerAdapter := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return peerAccount, nil
		},
		SaveAccountCalled: func(accountHandler vmcommon.AccountHandler) error {
			return saveAccountError
		},
	}

	arguments := createMockArguments()
	arguments.NodesSetup = &genesisMocks.NodesSetupStub{InitialNodesInfoCalled: func() (m map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, m2 map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
		oneMap := make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
		oneMap[0] = append(oneMap[0], mock.NewNodeInfo([]byte("aaaa"), []byte("aaaa"), 0, 50))
		return oneMap, oneMap
	}}
	arguments.PeerAdapter = peerAdapter
	_, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Equal(t, saveAccountError, err)
}

func TestValidatorStatisticsProcessor_SaveInitialStateCommitErrors(t *testing.T) {
	t.Parallel()

	commitError := errors.New("commit error")
	peerAccount, _ := accounts.NewPeerAccount([]byte("1234"))
	peerAdapter := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return peerAccount, nil
		},
		CommitCalled: func() (bytes []byte, e error) {
			return nil, commitError
		},
		SaveAccountCalled: func(accountHandler vmcommon.AccountHandler) error {
			return nil
		},
	}

	arguments := createMockArguments()
	arguments.PeerAdapter = peerAdapter
	_, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Equal(t, commitError, err)
}

func TestValidatorStatisticsProcessor_SaveInitialStateCommit(t *testing.T) {
	t.Parallel()

	peerAccount, _ := accounts.NewPeerAccount([]byte("1234"))
	peerAdapter := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return peerAccount, nil
		},
		CommitCalled: func() (bytes []byte, e error) {
			return nil, nil
		},
		SaveAccountCalled: func(accountHandler vmcommon.AccountHandler) error {
			return nil
		},
	}

	arguments := createMockArguments()
	arguments.PeerAdapter = peerAdapter
	_, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateReturnsRootHashForGenesis(t *testing.T) {
	t.Parallel()

	expectedRootHash := []byte("root hash")
	peerAdapter := getAccountsMock()
	peerAdapter.RootHashCalled = func() (bytes []byte, e error) {
		return expectedRootHash, nil
	}

	arguments := createMockArguments()
	arguments.PeerAdapter = peerAdapter
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	header.Nonce = 0
	rootHash, err := validatorStatistics.UpdatePeerState(header, make(map[string]data.CommonHeaderHandler))

	assert.Nil(t, err)
	assert.Equal(t, expectedRootHash, rootHash)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateReturnsErrForRootHashErr(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	peerAdapter := getAccountsMock()
	peerAdapter.RootHashCalled = func() (bytes []byte, e error) {
		return nil, expectedError
	}

	arguments := createMockArguments()
	arguments.PeerAdapter = peerAdapter
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	header.Nonce = 0
	_, err := validatorStatistics.UpdatePeerState(header, make(map[string]data.CommonHeaderHandler))

	assert.Equal(t, expectedError, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateComputeValidatorErrShouldError(t *testing.T) {
	t.Parallel()

	computeValidatorsErr := errors.New("compute validators error")

	arguments := createMockArguments()
	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			return nil, computeValidatorsErr
		},
	}
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	cache := make(map[string]data.CommonHeaderHandler)
	cache[string(header.GetPrevHash())] = getMetaHeaderHandler([]byte("header"))

	_, err := validatorStatistics.UpdatePeerState(header, cache)

	assert.Equal(t, computeValidatorsErr, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateGetExistingAccountErr(t *testing.T) {
	t.Parallel()

	existingAccountErr := errors.New("existing account err")
	adapter := getAccountsMock()
	adapter.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return nil, existingAccountErr
	}

	arguments := createMockArguments()
	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			return []nodesCoordinator.Validator{&shardingMocks.ValidatorMock{}}, nil
		},
	}
	arguments.PeerAdapter = adapter
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	cache := make(map[string]data.CommonHeaderHandler)
	cache[string(header.GetPrevHash())] = getMetaHeaderHandler([]byte("header"))
	_, err := validatorStatistics.UpdatePeerState(header, cache)

	assert.Equal(t, existingAccountErr, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateGetExistingAccountInvalidType(t *testing.T) {
	t.Parallel()

	adapter := getAccountsMock()
	adapter.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return &stateMock.AccountWrapMock{}, nil
	}

	arguments := createMockArguments()
	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			return []nodesCoordinator.Validator{&shardingMocks.ValidatorMock{}}, nil
		},
	}
	arguments.PeerAdapter = adapter
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	cache := make(map[string]data.CommonHeaderHandler)
	cache[string(header.GetPrevHash())] = getMetaHeaderHandler([]byte("header"))
	_, err := validatorStatistics.UpdatePeerState(header, cache)

	assert.Equal(t, process.ErrInvalidPeerAccount, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateGetHeaderError(t *testing.T) {
	t.Parallel()

	getHeaderError := errors.New("get header error")
	adapter := getAccountsMock()
	marshalizer := &mock.MarshalizerStub{}

	adapter.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return accounts.NewPeerAccount(address)
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := createMockArguments()
	arguments.Marshalizer = marshalizer
	arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{}
		},
	}
	arguments.StorageService = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, e error) {
					return nil, getHeaderError
				},
			}, nil
		},
	}
	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			return []nodesCoordinator.Validator{&shardingMocks.ValidatorMock{}, &shardingMocks.ValidatorMock{}}, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.PeerAdapter = adapter
	arguments.Rater = mock.GetNewMockRater()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	header.Nonce = 2
	_, err := validatorStatistics.UpdatePeerState(header, make(map[string]data.CommonHeaderHandler))

	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestValidatorStatisticsProcessor_UpdatePeerStateCallsIncrease(t *testing.T) {
	t.Parallel()

	adapter := getAccountsMock()
	increaseLeaderCalled := false
	increaseValidatorCalled := false
	marshalizer := &mock.MarshalizerStub{}

	adapter.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return &stateMock.PeerAccountHandlerMock{
			IncreaseLeaderSuccessRateCalled: func(value uint32) {
				increaseLeaderCalled = true
			},
			IncreaseValidatorSuccessRateCalled: func(value uint32) {
				increaseValidatorCalled = true
			},
		}, nil
	}
	adapter.RootHashCalled = func() (bytes []byte, e error) {
		return nil, nil
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := createMockArguments()
	arguments.Marshalizer = marshalizer
	arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{}
		},
	}
	arguments.StorageService = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, e error) {
					return nil, nil
				},
			}, nil
		},
	}
	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			return []nodesCoordinator.Validator{&shardingMocks.ValidatorMock{}, &shardingMocks.ValidatorMock{}}, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.PeerAdapter = adapter
	arguments.Rater = mock.GetNewMockRater()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	header.PubKeysBitmap = []byte{255, 0}

	marshalizer.UnmarshalCalled = func(obj interface{}, buff []byte) error {
		switch v := obj.(type) {
		case *block.MetaBlock:
			*v = block.MetaBlock{
				PubKeysBitmap:   []byte{255, 255},
				AccumulatedFees: big.NewInt(0),
				DeveloperFees:   big.NewInt(0),
			}
		case *block.Header:
			*v = block.Header{
				AccumulatedFees: big.NewInt(0),
				DeveloperFees:   big.NewInt(0),
			}
		default:
			fmt.Println(v)
		}

		return nil
	}
	cache := make(map[string]data.CommonHeaderHandler)
	cache[string(header.GetPrevHash())] = &block.MetaBlock{
		PubKeysBitmap:   []byte{255, 255},
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
	}
	_, err := validatorStatistics.UpdatePeerState(header, cache)

	assert.Nil(t, err)
	assert.True(t, increaseLeaderCalled)
	assert.True(t, increaseValidatorCalled)
}

func TestValidatorStatisticsProcessor_UpdatePeerState_IncreasesConsensusPreviousMetaBlock_SameEpoch(t *testing.T) {
	t.Parallel()

	consensusGroup := make(map[string][]nodesCoordinator.Validator)

	arguments := createUpdateTestArgs(consensusGroup)
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	cache := createMockCache()
	prevHeader, header := generateTestMetaBlockHeaders(cache)

	header.Round = prevHeader.Round + 1
	header.Epoch = 1

	v1 := shardingMocks.NewValidatorMock([]byte("pk1"), 1, 1)
	v2 := shardingMocks.NewValidatorMock([]byte("pk2"), 1, 1)
	v3 := shardingMocks.NewValidatorMock([]byte("pk3"), 1, 1)
	v4 := shardingMocks.NewValidatorMock([]byte("pk4"), 1, 1)

	prevHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.PrevRandSeed, prevHeader.Round, prevHeader.GetShardID(), prevHeader.Epoch)
	prevHeaderConsensus := []nodesCoordinator.Validator{v1, v2}
	consensusGroup[prevHeaderConsensusKey] = prevHeaderConsensus

	currentHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, header.PrevRandSeed, header.Round, header.GetShardID(), header.Epoch)
	currentHeaderConsensus := []nodesCoordinator.Validator{v3, v4}
	consensusGroup[currentHeaderConsensusKey] = currentHeaderConsensus

	_, err := validatorStatistics.UpdatePeerState(header, peer.CreateCommonHeaderCacheMap(cache))
	assert.Nil(t, err)

	pa1, _ := validatorStatistics.LoadPeerAccount(v1.PubKey())
	leader := pa1.(*stateMock.PeerAccountHandlerMock)
	pa2, _ := validatorStatistics.LoadPeerAccount(v2.PubKey())
	validator := pa2.(*stateMock.PeerAccountHandlerMock)

	assert.Equal(t, uint32(1), leader.IncreaseLeaderSuccessRateValue)
	assert.Equal(t, uint32(1), validator.IncreaseValidatorSuccessRateValue)
}

func TestValidatorStatisticsProcessor_UpdatePeerState_IncreasesIgnoredSignatures_SameEpoch(t *testing.T) {
	t.Parallel()

	consensusGroup := make(map[string][]nodesCoordinator.Validator)

	arguments := createUpdateTestArgs(consensusGroup)
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	cache := createMockCache()
	prevHeader, header := generateTestMetaBlockHeaders(cache)
	prevHeader.PubKeysBitmap = []byte{5}
	header.Round = prevHeader.Round + 1
	header.Epoch = 1

	v1 := shardingMocks.NewValidatorMock([]byte("pk1"), 1, 1)
	v2 := shardingMocks.NewValidatorMock([]byte("pk2"), 1, 1)
	v3 := shardingMocks.NewValidatorMock([]byte("pk3"), 1, 1)
	v4 := shardingMocks.NewValidatorMock([]byte("pk4"), 1, 1)

	prevHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.PrevRandSeed, prevHeader.Round, prevHeader.GetShardID(), prevHeader.Epoch)
	prevHeaderConsensus := []nodesCoordinator.Validator{v1, v2, v3}
	consensusGroup[prevHeaderConsensusKey] = prevHeaderConsensus

	currentHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, header.PrevRandSeed, header.Round, header.GetShardID(), header.Epoch)
	currentHeaderConsensus := []nodesCoordinator.Validator{v3, v4, v1}
	consensusGroup[currentHeaderConsensusKey] = currentHeaderConsensus

	_, err := validatorStatistics.UpdatePeerState(header, peer.CreateCommonHeaderCacheMap(cache))
	assert.Nil(t, err)

	pa1, _ := validatorStatistics.LoadPeerAccount(v1.PubKey())
	leader := pa1.(*stateMock.PeerAccountHandlerMock)
	pa2, _ := validatorStatistics.LoadPeerAccount(v2.PubKey())
	validatorIgnored := pa2.(*stateMock.PeerAccountHandlerMock)
	pa3, _ := validatorStatistics.LoadPeerAccount(v3.PubKey())
	validator := pa3.(*stateMock.PeerAccountHandlerMock)

	assert.Equal(t, uint32(1), leader.IncreaseLeaderSuccessRateValue)
	assert.Equal(t, uint32(1), validatorIgnored.IncreaseValidatorIgnoredSignaturesValue)
	assert.Equal(t, uint32(0), validatorIgnored.IncreaseValidatorSuccessRateValue)
	assert.Equal(t, uint32(0), validatorIgnored.DecreaseValidatorSuccessRateValue)
	assert.Equal(t, uint32(1), validator.IncreaseValidatorSuccessRateValue)
}

func generateTestMetaBlockHeaders(cache map[string]data.HeaderHandler) (*block.MetaBlock, *block.MetaBlock) {
	prevHeader := &block.MetaBlock{
		Round:           1,
		Epoch:           1,
		Nonce:           1,
		PubKeysBitmap:   []byte{255, 255},
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		PrevRandSeed:    []byte("prevRandSeed"),
	}

	header := getMetaHeaderHandler([]byte("header"))
	header.PubKeysBitmap = []byte{255, 0}
	header.RandSeed = []byte{1}

	cache[string(header.GetPrevHash())] = prevHeader
	return prevHeader, header
}

func generateTestShardBlockHeaders(cache map[string]data.HeaderHandler) (*block.Header, *block.Header) {
	prevHeader := &block.Header{
		Round:           1,
		Epoch:           1,
		Nonce:           1,
		PubKeysBitmap:   []byte{255, 255},
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		PrevRandSeed:    []byte("prevRandSeed"),
		RandSeed:        []byte("prevHeaderRandSeed"),
	}

	header := getShardHeaderHandler([]byte("header"))
	header.PubKeysBitmap = []byte{255, 0}
	header.RandSeed = []byte{1}

	cache[string(header.GetPrevHash())] = prevHeader
	return prevHeader, header
}

func TestValidatorStatisticsProcessor_UpdatePeerState_DecreasesMissedMetaBlock_SameEpoch(t *testing.T) {
	t.Parallel()

	consensusGroup := make(map[string][]nodesCoordinator.Validator)

	arguments := createUpdateTestArgs(consensusGroup)
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	cache := createMockCache()

	prevHeader, header := generateTestMetaBlockHeaders(cache)

	header.Round = prevHeader.Round + 2
	header.Epoch = 1

	v1 := shardingMocks.NewValidatorMock([]byte("pk1"), 1, 1)
	v2 := shardingMocks.NewValidatorMock([]byte("pk2"), 1, 1)
	v3 := shardingMocks.NewValidatorMock([]byte("pk3"), 1, 1)
	v4 := shardingMocks.NewValidatorMock([]byte("pk4"), 1, 1)

	prevHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.PrevRandSeed, prevHeader.Round, prevHeader.GetShardID(), prevHeader.Epoch)
	prevHeaderConsensus := []nodesCoordinator.Validator{v1, v2}
	consensusGroup[prevHeaderConsensusKey] = prevHeaderConsensus

	missedHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.RandSeed, prevHeader.Round+1, prevHeader.GetShardID(), prevHeader.Epoch)
	missedHeaderConsensus := []nodesCoordinator.Validator{v2, v3}
	consensusGroup[missedHeaderConsensusKey] = missedHeaderConsensus

	currentHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, header.PrevRandSeed, header.Round, header.GetShardID(), header.Epoch)
	currentHeaderConsensus := []nodesCoordinator.Validator{v3, v4}
	consensusGroup[currentHeaderConsensusKey] = currentHeaderConsensus

	_, err := validatorStatistics.UpdatePeerState(header, peer.CreateCommonHeaderCacheMap(cache))
	assert.Nil(t, err)

	pa1, _ := validatorStatistics.LoadPeerAccount(v2.PubKey())
	missedLeader := pa1.(*stateMock.PeerAccountHandlerMock)
	pa3, _ := validatorStatistics.LoadPeerAccount(v3.PubKey())
	missedValidator := pa3.(*stateMock.PeerAccountHandlerMock)

	assert.Equal(t, uint32(1), missedLeader.DecreaseLeaderSuccessRateValue)
	assert.Equal(t, uint32(1), missedValidator.DecreaseValidatorSuccessRateValue)
}

func TestValidatorStatisticsProcessor_UpdatePeerState_IncreasesConsensusPreviousMetaBlock_StartOfEpoch(t *testing.T) {
	t.Parallel()

	consensusGroup := make(map[string][]nodesCoordinator.Validator)

	arguments := createUpdateTestArgs(consensusGroup)
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	cache := createMockCache()
	prevHeader, header := generateTestMetaBlockHeaders(cache)

	header.Round = prevHeader.Round + 1
	header.Epoch = 1
	header.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0}}

	v1 := shardingMocks.NewValidatorMock([]byte("pk1"), 1, 1)
	v2 := shardingMocks.NewValidatorMock([]byte("pk2"), 1, 1)
	v3 := shardingMocks.NewValidatorMock([]byte("pk3"), 1, 1)
	v4 := shardingMocks.NewValidatorMock([]byte("pk4"), 1, 1)

	prevHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.PrevRandSeed, prevHeader.Round, prevHeader.GetShardID(), prevHeader.Epoch)
	prevHeaderConsensus := []nodesCoordinator.Validator{v1, v2}
	consensusGroup[prevHeaderConsensusKey] = prevHeaderConsensus

	currentHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, header.PrevRandSeed, header.Round, header.GetShardID(), header.Epoch-1)
	currentHeaderConsensus := []nodesCoordinator.Validator{v3, v4}
	consensusGroup[currentHeaderConsensusKey] = currentHeaderConsensus

	_, err := validatorStatistics.UpdatePeerState(header, peer.CreateCommonHeaderCacheMap(cache))
	assert.Nil(t, err)

	pa1, _ := validatorStatistics.LoadPeerAccount(v1.PubKey())
	leader := pa1.(*stateMock.PeerAccountHandlerMock)
	pa2, _ := validatorStatistics.LoadPeerAccount(v2.PubKey())
	validator := pa2.(*stateMock.PeerAccountHandlerMock)

	assert.Equal(t, uint32(1), leader.IncreaseLeaderSuccessRateValue)
	assert.Equal(t, uint32(1), validator.IncreaseValidatorSuccessRateValue)
}

func TestValidatorStatisticsProcessor_UpdatePeerState_DecreasesMissedMetaBlock_StartOfEpoch(t *testing.T) {
	t.Parallel()

	consensusGroup := make(map[string][]nodesCoordinator.Validator)

	arguments := createUpdateTestArgs(consensusGroup)
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	cache := createMockCache()

	prevHeader, header := generateTestMetaBlockHeaders(cache)

	header.Round = prevHeader.Round + 2
	header.Epoch = 2
	header.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0}}

	v1 := shardingMocks.NewValidatorMock([]byte("pk1"), 1, 1)
	v2 := shardingMocks.NewValidatorMock([]byte("pk2"), 1, 1)
	v3 := shardingMocks.NewValidatorMock([]byte("pk3"), 1, 1)
	v4 := shardingMocks.NewValidatorMock([]byte("pk4"), 1, 1)

	prevHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.PrevRandSeed, prevHeader.Round, prevHeader.GetShardID(), prevHeader.Epoch)
	prevHeaderConsensus := []nodesCoordinator.Validator{v1, v2}
	consensusGroup[prevHeaderConsensusKey] = prevHeaderConsensus

	missedHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.RandSeed, prevHeader.Round+1, prevHeader.GetShardID(), prevHeader.Epoch)
	missedHeaderConsensus := []nodesCoordinator.Validator{v2, v3}
	consensusGroup[missedHeaderConsensusKey] = missedHeaderConsensus

	currentHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, header.PrevRandSeed, header.Round, header.GetShardID(), header.Epoch-1)
	currentHeaderConsensus := []nodesCoordinator.Validator{v3, v4}
	consensusGroup[currentHeaderConsensusKey] = currentHeaderConsensus

	_, err := validatorStatistics.UpdatePeerState(header, peer.CreateCommonHeaderCacheMap(cache))
	assert.Nil(t, err)

	pa1, _ := validatorStatistics.LoadPeerAccount(v2.PubKey())
	missedLeader := pa1.(*stateMock.PeerAccountHandlerMock)
	pa3, _ := validatorStatistics.LoadPeerAccount(v3.PubKey())
	missedValidator := pa3.(*stateMock.PeerAccountHandlerMock)

	assert.Equal(t, uint32(1), missedLeader.DecreaseLeaderSuccessRateValue)
	assert.Equal(t, uint32(1), missedValidator.DecreaseValidatorSuccessRateValue)
}

func TestValidatorStatisticsProcessor_UpdatePeerState_IncreasesConsensusPreviousMetaBlock_PrevStartOfEpoch(t *testing.T) {
	t.Parallel()

	consensusGroup := make(map[string][]nodesCoordinator.Validator)

	arguments := createUpdateTestArgs(consensusGroup)
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	cache := createMockCache()

	prevHeader, header := generateTestMetaBlockHeaders(cache)

	prevHeader.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0}}

	header.Round = prevHeader.Round + 1
	header.Epoch = 1

	v1 := shardingMocks.NewValidatorMock([]byte("pk1"), 1, 1)
	v2 := shardingMocks.NewValidatorMock([]byte("pk2"), 1, 1)
	v3 := shardingMocks.NewValidatorMock([]byte("pk3"), 1, 1)
	v4 := shardingMocks.NewValidatorMock([]byte("pk4"), 1, 1)

	prevHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.PrevRandSeed, prevHeader.Round, prevHeader.GetShardID(), prevHeader.Epoch-1)
	prevHeaderConsensus := []nodesCoordinator.Validator{v1, v2}
	consensusGroup[prevHeaderConsensusKey] = prevHeaderConsensus

	currentHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, header.PrevRandSeed, header.Round, header.GetShardID(), header.Epoch)
	currentHeaderConsensus := []nodesCoordinator.Validator{v3, v4}
	consensusGroup[currentHeaderConsensusKey] = currentHeaderConsensus

	_, err := validatorStatistics.UpdatePeerState(header, peer.CreateCommonHeaderCacheMap(cache))
	assert.Nil(t, err)

	pa1, _ := validatorStatistics.LoadPeerAccount(v1.PubKey())
	leader := pa1.(*stateMock.PeerAccountHandlerMock)
	pa2, _ := validatorStatistics.LoadPeerAccount(v2.PubKey())
	validator := pa2.(*stateMock.PeerAccountHandlerMock)

	assert.Equal(t, uint32(1), leader.IncreaseLeaderSuccessRateValue)
	assert.Equal(t, uint32(1), validator.IncreaseValidatorSuccessRateValue)
}

func TestValidatorStatisticsProcessor_UpdatePeerState_DecreasesMissedMetaBlock_PrevStartOfEpoch(t *testing.T) {
	t.Parallel()

	consensusGroup := make(map[string][]nodesCoordinator.Validator)

	arguments := createUpdateTestArgs(consensusGroup)
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	cache := createMockCache()

	prevHeader, header := generateTestMetaBlockHeaders(cache)

	prevHeader.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0}}

	header.Round = prevHeader.Round + 2
	header.Epoch = 1

	v1 := shardingMocks.NewValidatorMock([]byte("pk1"), 1, 1)
	v2 := shardingMocks.NewValidatorMock([]byte("pk2"), 1, 1)
	v3 := shardingMocks.NewValidatorMock([]byte("pk3"), 1, 1)
	v4 := shardingMocks.NewValidatorMock([]byte("pk4"), 1, 1)

	prevHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.PrevRandSeed, prevHeader.Round, prevHeader.GetShardID(), prevHeader.Epoch-1)
	prevHeaderConsensus := []nodesCoordinator.Validator{v1, v2}
	consensusGroup[prevHeaderConsensusKey] = prevHeaderConsensus

	missedHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.RandSeed, prevHeader.Round+1, prevHeader.GetShardID(), prevHeader.Epoch)
	missedHeaderConsensus := []nodesCoordinator.Validator{v2, v3}
	consensusGroup[missedHeaderConsensusKey] = missedHeaderConsensus

	currentHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, header.PrevRandSeed, header.Round, header.GetShardID(), header.Epoch)
	currentHeaderConsensus := []nodesCoordinator.Validator{v3, v4}
	consensusGroup[currentHeaderConsensusKey] = currentHeaderConsensus

	_, err := validatorStatistics.UpdatePeerState(header, peer.CreateCommonHeaderCacheMap(cache))
	assert.Nil(t, err)

	pa1, _ := validatorStatistics.LoadPeerAccount(v2.PubKey())
	missedLeader := pa1.(*stateMock.PeerAccountHandlerMock)
	pa3, _ := validatorStatistics.LoadPeerAccount(v3.PubKey())
	missedValidator := pa3.(*stateMock.PeerAccountHandlerMock)

	assert.Equal(t, uint32(1), missedLeader.DecreaseLeaderSuccessRateValue)
	assert.Equal(t, uint32(1), missedValidator.DecreaseValidatorSuccessRateValue)
}

func TestValidatorStatisticsProcessor_UpdateShardDataPeerState_IncreasesConsensusCurrentShardBlock_SameEpoch(t *testing.T) {
	t.Parallel()

	consensusGroup := make(map[string][]nodesCoordinator.Validator)

	arguments := createUpdateTestArgs(consensusGroup)
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	cache := createMockCache()

	prevHeader, header := generateTestShardBlockHeaders(cache)

	header.Round = prevHeader.Round + 1
	header.Nonce = prevHeader.Nonce + 1
	header.Epoch = 1
	header.PrevRandSeed = prevHeader.RandSeed

	metaHeader := getMetaHeaderHandler([]byte("metaheader"))
	metaHeader.Round = prevHeader.Round + 1
	metaHeader.PubKeysBitmap = []byte{255, 0}
	metaHeader.Epoch = 1
	metaHeader.RandSeed = []byte{1}

	headerHash := []byte("headerHash")
	prevHeaderHash := []byte("prevHeaderHash")
	metaHeader.ShardInfo = []block.ShardData{
		shardDataFromHeader(headerHash, header),
	}

	cache[string(prevHeaderHash)] = prevHeader
	cache[string(headerHash)] = header

	v1 := shardingMocks.NewValidatorMock([]byte("pk1"), 1, 1)
	v2 := shardingMocks.NewValidatorMock([]byte("pk2"), 1, 1)
	v3 := shardingMocks.NewValidatorMock([]byte("pk3"), 1, 1)
	v4 := shardingMocks.NewValidatorMock([]byte("pk4"), 1, 1)

	prevHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.PrevRandSeed, prevHeader.Round, prevHeader.GetShardID(), prevHeader.Epoch)
	prevHeaderConsensus := []nodesCoordinator.Validator{v1, v2}
	consensusGroup[prevHeaderConsensusKey] = prevHeaderConsensus

	currentHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, header.PrevRandSeed, header.Round, header.GetShardID(), header.Epoch)
	currentHeaderConsensus := []nodesCoordinator.Validator{v3, v4}
	consensusGroup[currentHeaderConsensusKey] = currentHeaderConsensus

	err := validatorStatistics.UpdateShardDataPeerState(metaHeader, cache)
	assert.Nil(t, err)

	pa3, _ := validatorStatistics.LoadPeerAccount(v3.PubKey())
	leader := pa3.(*stateMock.PeerAccountHandlerMock)
	pa4, _ := validatorStatistics.LoadPeerAccount(v4.PubKey())
	validator := pa4.(*stateMock.PeerAccountHandlerMock)

	assert.Equal(t, uint32(1), leader.IncreaseLeaderSuccessRateValue)
	assert.Equal(t, uint32(1), validator.IncreaseValidatorSuccessRateValue)
}

func TestValidatorStatisticsProcessor_UpdateShardDataPeerState_DecreasesMissedShardBlock_SameEpoch(t *testing.T) {
	t.Parallel()

	consensusGroup := make(map[string][]nodesCoordinator.Validator)

	arguments := createUpdateTestArgs(consensusGroup)
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	cache := createMockCache()

	prevHeader, header := generateTestShardBlockHeaders(cache)

	header.Round = prevHeader.Round + 2
	header.Nonce = prevHeader.Nonce + 1
	header.Epoch = 1
	header.PrevRandSeed = prevHeader.RandSeed

	metaHeader := getMetaHeaderHandler([]byte("metaheader"))
	metaHeader.Round = prevHeader.Round + 1
	metaHeader.PubKeysBitmap = []byte{255, 0}
	metaHeader.Epoch = 1
	metaHeader.RandSeed = []byte{1}

	headerHash := []byte("headerHash")
	prevHeaderHash := []byte("prevHeaderHash")
	metaHeader.ShardInfo = []block.ShardData{
		shardDataFromHeader(headerHash, header),
	}

	cache[string(prevHeaderHash)] = prevHeader
	cache[string(headerHash)] = header

	v1 := shardingMocks.NewValidatorMock([]byte("pk1"), 1, 1)
	v2 := shardingMocks.NewValidatorMock([]byte("pk2"), 1, 1)
	v3 := shardingMocks.NewValidatorMock([]byte("pk3"), 1, 1)
	v4 := shardingMocks.NewValidatorMock([]byte("pk4"), 1, 1)

	prevHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.PrevRandSeed, prevHeader.Round, prevHeader.GetShardID(), prevHeader.Epoch)
	prevHeaderConsensus := []nodesCoordinator.Validator{v1, v2}
	consensusGroup[prevHeaderConsensusKey] = prevHeaderConsensus

	missedHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.RandSeed, prevHeader.Round+1, prevHeader.GetShardID(), prevHeader.Epoch)
	missedHeaderConsensus := []nodesCoordinator.Validator{v2, v3}
	consensusGroup[missedHeaderConsensusKey] = missedHeaderConsensus

	currentHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, header.PrevRandSeed, header.Round, header.GetShardID(), header.Epoch)
	currentHeaderConsensus := []nodesCoordinator.Validator{v3, v4}
	consensusGroup[currentHeaderConsensusKey] = currentHeaderConsensus

	err := validatorStatistics.UpdateShardDataPeerState(metaHeader, cache)
	assert.Nil(t, err)

	err = validatorStatistics.UpdateMissedBlocksCounters()
	assert.Nil(t, err)

	pa2, _ := validatorStatistics.LoadPeerAccount(v2.PubKey())
	missedLeader := pa2.(*stateMock.PeerAccountHandlerMock)
	pa3, _ := validatorStatistics.LoadPeerAccount(v3.PubKey())
	missedValidator := pa3.(*stateMock.PeerAccountHandlerMock)

	assert.Equal(t, uint32(1), missedLeader.DecreaseLeaderSuccessRateValue)
	assert.Equal(t, uint32(1), missedValidator.DecreaseValidatorSuccessRateValue)
}

func TestValidatorStatisticsProcessor_UpdateShardDataPeerState_IncreasesConsensusShardBlock_StartOfEpoch(t *testing.T) {
	t.Parallel()

	consensusGroup := make(map[string][]nodesCoordinator.Validator)

	arguments := createUpdateTestArgs(consensusGroup)
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	cache := createMockCache()

	prevHeader, header := generateTestShardBlockHeaders(cache)

	header.Round = prevHeader.Round + 1
	header.Nonce = prevHeader.Nonce + 1
	header.Epoch = 2
	header.PrevRandSeed = prevHeader.RandSeed
	header.EpochStartMetaHash = []byte("epochStartMetaHash")

	metaHeader := getMetaHeaderHandler([]byte("metaheader"))
	metaHeader.Round = prevHeader.Round + 1
	metaHeader.PubKeysBitmap = []byte{255, 0}
	metaHeader.Epoch = 1
	metaHeader.RandSeed = []byte{1}

	headerHash := []byte("headerHash")
	prevHeaderHash := []byte("prevHeaderHash")
	metaHeader.ShardInfo = []block.ShardData{
		shardDataFromHeader(headerHash, header),
	}

	cache[string(prevHeaderHash)] = prevHeader
	cache[string(headerHash)] = header

	v1 := shardingMocks.NewValidatorMock([]byte("pk1"), 1, 1)
	v2 := shardingMocks.NewValidatorMock([]byte("pk2"), 1, 1)
	v3 := shardingMocks.NewValidatorMock([]byte("pk3"), 1, 1)
	v4 := shardingMocks.NewValidatorMock([]byte("pk4"), 1, 1)

	prevHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.PrevRandSeed, prevHeader.Round, prevHeader.GetShardID(), prevHeader.Epoch)
	prevHeaderConsensus := []nodesCoordinator.Validator{v1, v2}
	consensusGroup[prevHeaderConsensusKey] = prevHeaderConsensus

	currentHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, header.PrevRandSeed, header.Round, header.GetShardID(), header.Epoch-1)
	currentHeaderConsensus := []nodesCoordinator.Validator{v3, v4}
	consensusGroup[currentHeaderConsensusKey] = currentHeaderConsensus

	err := validatorStatistics.UpdateShardDataPeerState(metaHeader, cache)
	assert.Nil(t, err)

	err = validatorStatistics.UpdateMissedBlocksCounters()
	assert.Nil(t, err)

	pa3, _ := validatorStatistics.LoadPeerAccount(v3.PubKey())
	leader := pa3.(*stateMock.PeerAccountHandlerMock)
	pa4, _ := validatorStatistics.LoadPeerAccount(v4.PubKey())
	validator := pa4.(*stateMock.PeerAccountHandlerMock)

	assert.Equal(t, uint32(1), leader.IncreaseLeaderSuccessRateValue)
	assert.Equal(t, uint32(1), validator.IncreaseValidatorSuccessRateValue)
}

func TestValidatorStatisticsProcessor_UpdateShardDataPeerState_DecreasesMissedShardBlock_StartOfEpoch(t *testing.T) {
	t.Parallel()

	consensusGroup := make(map[string][]nodesCoordinator.Validator)

	arguments := createUpdateTestArgs(consensusGroup)
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	cache := createMockCache()

	prevHeader, header := generateTestShardBlockHeaders(cache)

	header.Round = prevHeader.Round + 2
	header.Nonce = prevHeader.Nonce + 1
	header.Epoch = 1
	header.PrevRandSeed = prevHeader.RandSeed

	metaHeader := getMetaHeaderHandler([]byte("metaheader"))
	metaHeader.Round = prevHeader.Round + 1
	metaHeader.PubKeysBitmap = []byte{255, 0}
	metaHeader.Epoch = 1
	metaHeader.RandSeed = []byte{1}

	headerHash := []byte("headerHash")
	prevHeaderHash := []byte("prevHeaderHash")
	metaHeader.ShardInfo = []block.ShardData{
		shardDataFromHeader(headerHash, header),
	}

	cache[string(prevHeaderHash)] = prevHeader
	cache[string(headerHash)] = header

	v1 := shardingMocks.NewValidatorMock([]byte("pk1"), 1, 1)
	v2 := shardingMocks.NewValidatorMock([]byte("pk2"), 1, 1)
	v3 := shardingMocks.NewValidatorMock([]byte("pk3"), 1, 1)
	v4 := shardingMocks.NewValidatorMock([]byte("pk4"), 1, 1)

	prevHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.PrevRandSeed, prevHeader.Round, prevHeader.GetShardID(), prevHeader.Epoch)
	prevHeaderConsensus := []nodesCoordinator.Validator{v1, v2}
	consensusGroup[prevHeaderConsensusKey] = prevHeaderConsensus

	missedHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, prevHeader.RandSeed, prevHeader.Round+1, prevHeader.GetShardID(), prevHeader.Epoch)
	missedHeaderConsensus := []nodesCoordinator.Validator{v2, v3}
	consensusGroup[missedHeaderConsensusKey] = missedHeaderConsensus

	currentHeaderConsensusKey := fmt.Sprintf(consensusGroupFormat, header.PrevRandSeed, header.Round, header.GetShardID(), header.Epoch)
	currentHeaderConsensus := []nodesCoordinator.Validator{v3, v4}
	consensusGroup[currentHeaderConsensusKey] = currentHeaderConsensus

	err := validatorStatistics.UpdateShardDataPeerState(metaHeader, cache)
	assert.Nil(t, err)

	err = validatorStatistics.UpdateMissedBlocksCounters()
	assert.Nil(t, err)

	pa2, _ := validatorStatistics.LoadPeerAccount(v2.PubKey())
	missedLeader := pa2.(*stateMock.PeerAccountHandlerMock)
	pa3, _ := validatorStatistics.LoadPeerAccount(v3.PubKey())
	missedValidator := pa3.(*stateMock.PeerAccountHandlerMock)

	assert.Equal(t, uint32(1), missedLeader.DecreaseLeaderSuccessRateValue)
	assert.Equal(t, uint32(1), missedValidator.DecreaseValidatorSuccessRateValue)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateCheckForMissedBlocksErr(t *testing.T) {
	t.Parallel()

	adapter := getAccountsMock()
	missedBlocksErr := errors.New("missed blocks error")
	shouldErr := false
	marshalizer := &mock.MarshalizerStub{}

	adapter.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return &stateMock.PeerAccountHandlerMock{
			DecreaseLeaderSuccessRateCalled: func(value uint32) {
				shouldErr = true
			},
		}, nil
	}

	adapter.SaveAccountCalled = func(account vmcommon.AccountHandler) error {
		if shouldErr {
			return missedBlocksErr
		}
		return nil
	}
	adapter.RootHashCalled = func() (bytes []byte, e error) {
		return nil, nil
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := createMockArguments()
	arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{}
		},
	}
	arguments.StorageService = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, e error) {
					return nil, nil
				},
			}, nil
		},
	}
	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			return []nodesCoordinator.Validator{&shardingMocks.ValidatorMock{}, &shardingMocks.ValidatorMock{}}, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.PeerAdapter = adapter
	arguments.Marshalizer = marshalizer
	arguments.Rater = mock.GetNewMockRater()

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	header.Nonce = 2
	header.Round = 2

	marshalizer.UnmarshalCalled = func(obj interface{}, buff []byte) error {
		switch v := obj.(type) {
		case *block.MetaBlock:
			*v = block.MetaBlock{
				Nonce:         0,
				PubKeysBitmap: []byte{0, 0},
			}
		case *block.Header:
			*v = block.Header{}
		default:
			fmt.Println(v)
		}

		return nil
	}
	cache := createMockCache()
	cache[string(header.GetPrevHash())] = &block.MetaBlock{
		Nonce:         0,
		PubKeysBitmap: []byte{0, 0},
	}
	_, err := validatorStatistics.UpdatePeerState(header, peer.CreateCommonHeaderCacheMap(cache))

	assert.Equal(t, missedBlocksErr, err)
}

func shardDataFromHeader(headerHash []byte, prevHeader *block.Header) block.ShardData {
	sd := block.ShardData{HeaderHash: headerHash,
		PrevRandSeed:    prevHeader.PrevRandSeed,
		PubKeysBitmap:   prevHeader.PubKeysBitmap,
		Signature:       prevHeader.Signature,
		Round:           prevHeader.Round,
		PrevHash:        prevHeader.PrevHash,
		Nonce:           prevHeader.Nonce,
		ShardID:         prevHeader.ShardID,
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
	}

	return sd
}

func TestValidatorStatisticsProcessor_CheckForMissedBlocksNoMissedBlocks(t *testing.T) {
	t.Parallel()

	computeValidatorGroupCalled := false
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := createMockArguments()
	arguments.Marshalizer = &mock.MarshalizerMock{}
	arguments.DataPool = dataRetrieverMock.NewPoolsHolderStub()
	arguments.StorageService = &storageStubs.ChainStorerStub{}
	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			computeValidatorGroupCalled = true
			return nil, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.PeerAdapter = getAccountsMock()

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	err := validatorStatistics.CheckForMissedBlocks(1, 0, []byte("prev"), 0, 0)
	assert.Nil(t, err)
	assert.False(t, computeValidatorGroupCalled)

	err = validatorStatistics.CheckForMissedBlocks(1, 1, []byte("prev"), 0, 0)
	assert.Nil(t, err)
	assert.False(t, computeValidatorGroupCalled)

	err = validatorStatistics.CheckForMissedBlocks(2, 1, []byte("prev"), 0, 0)
	assert.Nil(t, err)
	assert.False(t, computeValidatorGroupCalled)
}

func TestValidatorStatisticsProcessor_CheckForMissedBlocksMissedRoundsGreaterThanMaxConsecutiveRoundsOfRatingDecrease(t *testing.T) {
	t.Parallel()

	validatorPublicKeys := make(map[uint32][][]byte)
	validatorPublicKeys[0] = make([][]byte, 1)
	validatorPublicKeys[0][0] = []byte("validator")
	validatorRating := 100

	nodesCoordinatorMock := &shardingMocks.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func(_ uint32) (map[uint32][][]byte, error) {
			return validatorPublicKeys, nil
		},
	}

	peerAdapter := getAccountsMock()
	peerAdapter.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return &stateMock.PeerAccountHandlerMock{
			SetTempRatingCalled: func(value uint32) {
				validatorRating--
			},
		}, nil
	}

	arguments := createMockArguments()
	arguments.PeerAdapter = peerAdapter
	arguments.NodesCoordinator = nodesCoordinatorMock
	arguments.MaxComputableRounds = 1
	enableEpochsHandler, _ := arguments.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.RemoveActiveFlags(common.StopDecreasingValidatorRatingWhenStuckFlag)
	arguments.MaxConsecutiveRoundsOfRatingDecrease = 4

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	// Flag to stop decreasing validator rating is NOT set => decrease validator rating
	err := validatorStatistics.CheckForMissedBlocks(5, 0, []byte("prev"), 0, 0)
	require.Nil(t, err)
	require.Equal(t, 99, validatorRating)

	// Flag to stop decreasing validator rating is set, but NOT enough missed rounds to stop decreasing ratings => decrease validator rating again
	enableEpochsHandler.AddActiveFlags(common.StopDecreasingValidatorRatingWhenStuckFlag)
	err = validatorStatistics.CheckForMissedBlocks(4, 0, []byte("prev"), 0, 0)
	require.Nil(t, err)
	require.Equal(t, 98, validatorRating)

	// Flag to stop decreasing validator rating is set AND missed rounds > max rounds of rating decrease => validator rating is NOT decreased
	err = validatorStatistics.CheckForMissedBlocks(5, 0, []byte("prev"), 0, 0)
	require.Nil(t, err)
	require.Equal(t, 98, validatorRating)
	err = validatorStatistics.CheckForMissedBlocks(6, 0, []byte("prev"), 0, 0)
	require.Nil(t, err)
	require.Equal(t, 98, validatorRating)
}

func TestValidatorStatisticsProcessor_CheckForMissedBlocksErrOnComputeValidatorList(t *testing.T) {
	t.Parallel()

	computeErr := errors.New("compute err")
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := createMockArguments()
	arguments.Marshalizer = &mock.MarshalizerMock{}
	arguments.DataPool = dataRetrieverMock.NewPoolsHolderStub()
	arguments.StorageService = &storageStubs.ChainStorerStub{}
	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			return nil, computeErr
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.PeerAdapter = getAccountsMock()
	arguments.Rater = mock.GetNewMockRater()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	err := validatorStatistics.CheckForMissedBlocks(2, 0, []byte("prev"), 0, 0)
	assert.Equal(t, computeErr, err)
}

func TestValidatorStatisticsProcessor_CheckForMissedBlocksErrOnDecrease(t *testing.T) {
	t.Parallel()

	decreaseErr := false
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerAdapter := getAccountsMock()
	peerAdapter.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return &stateMock.PeerAccountHandlerMock{
			DecreaseLeaderSuccessRateCalled: func(value uint32) {
				decreaseErr = true
			},
		}, nil
	}

	arguments := createMockArguments()
	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			return []nodesCoordinator.Validator{
				&shardingMocks.ValidatorMock{},
			}, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.PeerAdapter = peerAdapter
	arguments.Rater = mock.GetNewMockRater()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	_ = validatorStatistics.CheckForMissedBlocks(2, 0, []byte("prev"), 0, 0)
	_ = validatorStatistics.UpdateMissedBlocksCounters()
	assert.True(t, decreaseErr)
}

func TestValidatorStatisticsProcessor_CheckForMissedBlocksCallsDecrease(t *testing.T) {
	t.Parallel()

	currentHeaderRound := 10
	previousHeaderRound := 4
	decreaseCount := 0
	pubKey := []byte("pubKey")
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerAdapter := getAccountsMock()
	peerAdapter.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return &stateMock.PeerAccountHandlerMock{
			DecreaseLeaderSuccessRateCalled: func(value uint32) {
				decreaseCount += 5
			},
		}, nil
	}

	arguments := createMockArguments()
	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			return []nodesCoordinator.Validator{
				&shardingMocks.ValidatorMock{
					PubKeyCalled: func() []byte {
						return pubKey
					},
				},
			}, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.PeerAdapter = peerAdapter
	arguments.Rater = mock.GetNewMockRater()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	_ = validatorStatistics.CheckForMissedBlocks(uint64(currentHeaderRound), uint64(previousHeaderRound), []byte("prev"), 0, 0)
	counters := validatorStatistics.GetLeaderDecreaseCount(pubKey)
	_ = validatorStatistics.UpdateMissedBlocksCounters()
	assert.Equal(t, uint32(currentHeaderRound-previousHeaderRound-1), counters)
}

func TestValidatorStatisticsProcessor_CheckForMissedBlocksWithRoundDifferenceGreaterThanMaxComputableCallsDecreaseOnlyOnce(t *testing.T) {
	t.Parallel()

	currentHeaderRound := 20
	previousHeaderRound := 10
	decreaseValidatorCalls := 0
	decreaseLeaderCalls := 0
	setTempRatingCalls := 0

	validatorPublicKeys := make(map[uint32][][]byte)
	validatorPublicKeys[0] = make([][]byte, 1)
	validatorPublicKeys[0][0] = []byte("testpk")

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerAdapter := getAccountsMock()
	peerAdapter.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return &stateMock.PeerAccountHandlerMock{
			DecreaseLeaderSuccessRateCalled: func(value uint32) {
				decreaseLeaderCalls++
			},
			DecreaseValidatorSuccessRateCalled: func(value uint32) {
				decreaseValidatorCalls++
			},
			SetTempRatingCalled: func(value uint32) {
				setTempRatingCalls++
			},
		}, nil
	}

	arguments := createMockArguments()
	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, _ uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			return []nodesCoordinator.Validator{
				&shardingMocks.ValidatorMock{},
			}, nil
		},
		GetAllEligibleValidatorsPublicKeysCalled: func(_ uint32) (map[uint32][][]byte, error) {
			return validatorPublicKeys, nil
		},
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (nodesCoordinator.Validator, uint32, error) {
			validator, _ := nodesCoordinator.NewValidator(publicKey, defaultChancesSelection, 1)
			return validator, 0, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.PeerAdapter = peerAdapter
	arguments.Rater = mock.GetNewMockRater()
	arguments.MaxComputableRounds = 5

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	_ = validatorStatistics.CheckForMissedBlocks(uint64(currentHeaderRound), uint64(previousHeaderRound), []byte("prev"), 0, 0)
	assert.Equal(t, 1, decreaseLeaderCalls)
	assert.Equal(t, 1, decreaseValidatorCalls)
	assert.Equal(t, 1, setTempRatingCalls)
}

func TestValidatorStatisticsProcessor_CheckForMissedBlocksWithRoundDifferenceGreaterThanMaxComputableCallsOnlyOnce(t *testing.T) {
	t.Parallel()

	currentHeaderRound := 20
	previousHeaderRound := 10
	decreaseValidatorCalls := 0
	decreaseLeaderCalls := 0
	setTempRatingCalls := 0
	nrValidators := 1

	validatorPublicKeys := make(map[uint32][][]byte)
	validatorPublicKeys[0] = make([][]byte, nrValidators)
	for i := 0; i < nrValidators; i++ {
		validatorPublicKeys[0][i] = []byte(fmt.Sprintf("testpk_%v", i))
	}

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerAdapter := getAccountsMock()
	peerAdapter.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return &stateMock.PeerAccountHandlerMock{
			DecreaseLeaderSuccessRateCalled: func(value uint32) {
				decreaseLeaderCalls++
			},
			DecreaseValidatorSuccessRateCalled: func(value uint32) {
				decreaseValidatorCalls++
			},
			SetTempRatingCalled: func(value uint32) {
				setTempRatingCalls++
			},
		}, nil
	}

	arguments := createMockArguments()
	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, _ uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			return []nodesCoordinator.Validator{
				&shardingMocks.ValidatorMock{},
			}, nil
		},
		GetAllEligibleValidatorsPublicKeysCalled: func(_ uint32) (map[uint32][][]byte, error) {
			return validatorPublicKeys, nil
		},
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (nodesCoordinator.Validator, uint32, error) {
			validator, _ := nodesCoordinator.NewValidator(publicKey, defaultChancesSelection, 1)
			return validator, 0, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.PeerAdapter = peerAdapter
	arguments.Rater = mock.GetNewMockRater()
	arguments.MaxComputableRounds = 5

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	_ = validatorStatistics.CheckForMissedBlocks(uint64(currentHeaderRound), uint64(previousHeaderRound), []byte("prev"), 0, 0)
	assert.Equal(t, 1, decreaseLeaderCalls)
	assert.Equal(t, 1, decreaseValidatorCalls)
	assert.Equal(t, 1, setTempRatingCalls)
}

func TestValidatorStatisticsProcessor_CheckForMissedBlocksWithRoundDifferences(t *testing.T) {
	t.Parallel()

	currentHeaderRound := uint64(101)
	previousHeaderRound := uint64(1)
	maxComputableRounds := uint64(5)

	type args struct {
		currentHeaderRound  uint64
		previousHeaderRound uint64
		maxComputableRounds uint64
		nrValidators        int
		consensusGroupSize  int
	}

	type result struct {
		decreaseValidatorValue uint32
		decreaseLeaderValue    uint32
		tempRating             uint32
	}

	type testSuite struct {
		name string
		args args
		want result
	}

	rater := mock.GetNewMockRater()
	rater.StartRating = 500000
	rater.MinRating = 100000
	rater.MaxRating = 1000000
	rater.DecreaseProposer = -2000
	rater.DecreaseValidator = -10

	validators := []struct {
		validators    int
		consensusSize int
	}{
		{validators: 1, consensusSize: 1},
		{validators: 2, consensusSize: 1},
		{validators: 10, consensusSize: 1},
		{validators: 100, consensusSize: 1},
		{validators: 400, consensusSize: 1},
		{validators: 400, consensusSize: 2},
		{validators: 400, consensusSize: 10},
		{validators: 400, consensusSize: 63},
		{validators: 400, consensusSize: 400},
	}

	tests := make([]testSuite, len(validators))

	for i, nodes := range validators {
		{
			leaderProbability := computeLeaderProbability(currentHeaderRound, previousHeaderRound, nodes.validators)
			intValidatorProbability := uint32(leaderProbability*float64(nodes.consensusSize) + 1 - math.SmallestNonzeroFloat64)
			intLeaderProbability := uint32(leaderProbability + 1 - math.SmallestNonzeroFloat64)

			tests[i] = testSuite{
				args: args{
					currentHeaderRound:  currentHeaderRound,
					previousHeaderRound: previousHeaderRound,
					maxComputableRounds: maxComputableRounds,
					nrValidators:        nodes.validators,
					consensusGroupSize:  nodes.consensusSize,
				},
				want: result{
					decreaseValidatorValue: intValidatorProbability,
					decreaseLeaderValue:    intLeaderProbability,
					tempRating: uint32(int32(rater.StartRating) +
						int32(intLeaderProbability)*rater.DecreaseProposer +
						int32(intValidatorProbability)*rater.DecreaseValidator),
				},
			}
		}
	}

	for _, tt := range tests {
		ttCopy := tt
		t.Run(tt.name, func(t *testing.T) {
			decreaseLeader, decreaseValidator, rating := DoComputeMissingBlocks(
				rater,
				tt.args.nrValidators,
				tt.args.consensusGroupSize,
				tt.args.currentHeaderRound,
				tt.args.previousHeaderRound,
				tt.args.maxComputableRounds)

			res := result{
				decreaseValidatorValue: decreaseValidator,
				decreaseLeaderValue:    decreaseLeader,
				tempRating:             rating,
			}

			if res != ttCopy.want {
				t.Errorf("ComputeMissingBlocks = %v, want %v", res, ttCopy.want)
			}

			t.Logf("validators:%v, consensusSize:%v, missedRounds: %v, decreased leader: %v, decreased validator: %v, startRating: %v, endRating: %v",
				ttCopy.args.nrValidators,
				ttCopy.args.consensusGroupSize,
				ttCopy.args.currentHeaderRound-ttCopy.args.previousHeaderRound,
				ttCopy.want.decreaseLeaderValue,
				ttCopy.want.decreaseValidatorValue,
				rater.StartRating,
				ttCopy.want.tempRating,
			)

		})
	}

}

func computeLeaderProbability(
	currentHeaderRound uint64,
	previousHeaderRound uint64,
	validators int,
) float64 {
	return (float64(currentHeaderRound) - float64(previousHeaderRound) - 1) / float64(validators)
}

func DoComputeMissingBlocks(
	rater *mock.RaterMock,
	nrValidators int,
	consensusGroupSize int,
	currentHeaderRounds uint64,
	previousHeaderRound uint64,
	maxComputableRounds uint64,
) (uint32, uint32, uint32) {
	validatorPublicKeys := make(map[uint32][][]byte)
	validatorPublicKeys[0] = make([][]byte, nrValidators)
	for i := 0; i < nrValidators; i++ {
		validatorPublicKeys[0][i] = []byte(fmt.Sprintf("testpk_%v", i))
	}

	consensus := make([]nodesCoordinator.Validator, consensusGroupSize)
	for i := 0; i < consensusGroupSize; i++ {
		consensus[i] = &shardingMocks.ValidatorMock{}
	}

	accountsMap := make(map[string]*stateMock.PeerAccountHandlerMock)
	leaderSuccesRateMap := make(map[string]uint32)
	validatorSuccesRateMap := make(map[string]uint32)
	ratingMap := make(map[string]uint32)

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerAdapter := getAccountsMock()
	peerAdapter.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		key := string(address)
		account, found := accountsMap[key]

		if !found {
			account = &stateMock.PeerAccountHandlerMock{
				DecreaseLeaderSuccessRateCalled: func(value uint32) {
					leaderSuccesRateMap[key] += value
				},
				DecreaseValidatorSuccessRateCalled: func(value uint32) {
					validatorSuccesRateMap[key] += value
				},
				GetTempRatingCalled: func() uint32 {
					return ratingMap[key]
				},
				SetTempRatingCalled: func(value uint32) {
					ratingMap[key] = value
				},
			}
			accountsMap[key] = account
			leaderSuccesRateMap[key] = 0
			validatorSuccesRateMap[key] = 0
			ratingMap[key] = rater.StartRating
		}

		return account, nil
	}

	arguments := createMockArguments()
	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, _ uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			return consensus, nil
		},
		GetAllEligibleValidatorsPublicKeysCalled: func(_ uint32) (map[uint32][][]byte, error) {
			return validatorPublicKeys, nil
		},
		ConsensusGroupSizeCalled: func(uint32) int {
			return consensusGroupSize
		},
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (nodesCoordinator.Validator, uint32, error) {
			validator, _ := nodesCoordinator.NewValidator(publicKey, defaultChancesSelection, 1)
			return validator, 0, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.PeerAdapter = peerAdapter
	arguments.Rater = rater

	arguments.MaxComputableRounds = maxComputableRounds

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	_ = validatorStatistics.CheckForMissedBlocks(currentHeaderRounds, previousHeaderRound, []byte("prev"), 0, 0)

	firstKey := "testpk_0"

	return leaderSuccesRateMap[firstKey], validatorSuccesRateMap[firstKey], ratingMap[firstKey]
}

func TestValidatorStatisticsProcessor_GetMatchingPrevShardDataEmptySDReturnsNil(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()

	currentShardData := block.ShardData{}
	shardInfo := make([]block.ShardData, 0)

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	sd := validatorStatistics.GetMatchingPrevShardData(currentShardData, shardInfo)

	assert.Nil(t, sd)
}

func TestValidatorStatisticsProcessor_GetMatchingPrevShardDataNoMatch(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()

	currentShardData := block.ShardData{ShardID: 1, Nonce: 10}
	shardInfo := []block.ShardData{{ShardID: 1, Nonce: 8}, {ShardID: 2, Nonce: 9}}

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	sd := validatorStatistics.GetMatchingPrevShardData(currentShardData, shardInfo)

	assert.Nil(t, sd)
}

func TestValidatorStatisticsProcessor_GetMatchingPrevShardDataFindsMatch(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()

	currentShardData := block.ShardData{ShardID: 1, Nonce: 10}
	shardInfo := []block.ShardData{{ShardID: 1, Nonce: 9}, {ShardID: 2, Nonce: 9}}

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	sd := validatorStatistics.GetMatchingPrevShardData(currentShardData, shardInfo)

	assert.NotNil(t, sd)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateCallsPubKeyForValidator(t *testing.T) {
	t.Parallel()

	pubKeyCalled := false
	arguments := createMockArguments()
	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			return []nodesCoordinator.Validator{&shardingMocks.ValidatorMock{
				PubKeyCalled: func() []byte {
					pubKeyCalled = true
					return make([]byte, 0)
				},
			}, &shardingMocks.ValidatorMock{}}, nil
		},
	}
	arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{
				GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, e error) {
					return getMetaHeaderHandler([]byte("header")), nil
				},
			}
		},
	}

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	header := getMetaHeaderHandler([]byte("header"))

	cache := make(map[string]data.CommonHeaderHandler)
	cache[string(header.GetPrevHash())] = getMetaHeaderHandler([]byte("header"))
	_, _ = validatorStatistics.UpdatePeerState(header, cache)

	assert.True(t, pubKeyCalled)
}

func getMetaHeaderHandler(randSeed []byte) *block.MetaBlock {
	return &block.MetaBlock{
		Nonce:           2,
		PrevRandSeed:    randSeed,
		PrevHash:        randSeed,
		PubKeysBitmap:   randSeed,
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
	}
}

func getShardHeaderHandler(randSeed []byte) *block.Header {
	return &block.Header{
		Nonce:           2,
		PrevRandSeed:    randSeed,
		PrevHash:        randSeed,
		PubKeysBitmap:   randSeed,
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
	}
}

func getAccountsMock() *stateMock.AccountsStub {
	return &stateMock.AccountsStub{
		CommitCalled: func() (bytes []byte, e error) {
			return make([]byte, 0), nil
		},
		LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return &stateMock.PeerAccountHandlerMock{}, nil
		},
	}
}

func TestValidatorStatistics_RootHashWithErrShouldReturnNil(t *testing.T) {
	t.Parallel()

	hash := []byte("nonExistingRootHash")
	expectedErr := errors.New("invalid rootHash")

	arguments := createMockArguments()

	peerAdapter := getAccountsMock()
	peerAdapter.GetAllLeavesCalled = func(_ *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.TrieLeafParser) error {
		return expectedErr
	}
	arguments.PeerAdapter = peerAdapter

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	validatorInfos, err := validatorStatistics.GetValidatorInfoForRootHash(hash)
	assert.Nil(t, validatorInfos)
	assert.Equal(t, expectedErr, err)
}

func TestValidatorStatistics_ResetValidatorStatisticsAtNewEpoch(t *testing.T) {
	t.Parallel()

	hash := []byte("correctRootHash")
	expectedErr := errors.New("unknown peer")
	arguments := createMockArguments()

	addrBytes0 := []byte("addr1")
	addrBytesMeta := []byte("addrM")

	pa0, _ := createPeerAccounts(addrBytes0, addrBytesMeta)

	marshalizedPa0, _ := arguments.Marshalizer.Marshal(pa0)

	peerAdapter := getAccountsMock()
	peerAdapter.GetAllLeavesCalled = func(ch *common.TrieIteratorChannels, _ context.Context, rootHash []byte, _ common.TrieLeafParser) error {
		if bytes.Equal(rootHash, hash) {
			go func() {
				ch.LeavesChan <- keyValStorage.NewKeyValStorage(addrBytes0, marshalizedPa0)
				close(ch.LeavesChan)
				ch.ErrChan.Close()
			}()

			return nil
		}
		return expectedErr
	}
	peerAdapter.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, err error) {
		if bytes.Equal(pa0.AddressBytes(), address) {
			return pa0, nil
		}
		return nil, expectedErr
	}
	arguments.PeerAdapter = peerAdapter
	arguments.PubkeyConv = testscommon.NewPubkeyConverterMock(4)
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	validatorInfos, _ := validatorStatistics.GetValidatorInfoForRootHash(hash)

	assert.NotEqual(t, pa0.GetTempRating(), pa0.GetRating())

	err := validatorStatistics.ResetValidatorStatisticsAtNewEpoch(validatorInfos)

	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(0), pa0.GetAccumulatedFees())

	assert.Equal(t, uint32(11), pa0.GetTotalValidatorSuccessRate().GetNumSuccess())
	assert.Equal(t, uint32(22), pa0.GetTotalValidatorSuccessRate().GetNumFailure())
	assert.Equal(t, uint32(33), pa0.GetTotalLeaderSuccessRate().GetNumSuccess())
	assert.Equal(t, uint32(44), pa0.GetTotalLeaderSuccessRate().GetNumFailure())
	assert.Equal(t, uint32(55), pa0.GetTotalValidatorIgnoredSignaturesRate())

	assert.Equal(t, uint32(0), pa0.GetValidatorSuccessRate().GetNumSuccess())
	assert.Equal(t, uint32(0), pa0.GetValidatorSuccessRate().GetNumFailure())
	assert.Equal(t, uint32(0), pa0.GetLeaderSuccessRate().GetNumSuccess())
	assert.Equal(t, uint32(0), pa0.GetLeaderSuccessRate().GetNumFailure())
	assert.Equal(t, uint32(0), pa0.GetValidatorIgnoredSignaturesRate())

	assert.Equal(t, uint32(0), pa0.GetNumSelectedInSuccessBlocks())
	assert.Equal(t, pa0.GetTempRating(), pa0.GetRating())
}

func TestValidatorStatistics_Process(t *testing.T) {
	t.Parallel()

	hash := []byte("correctRootHash")
	expectedErr := errors.New("error rootHash")
	arguments := createMockArguments()

	addrBytes0 := []byte("addr1")
	addrBytesMeta := []byte("addrMeta")

	pa0, paMeta := createPeerAccounts(addrBytes0, addrBytesMeta)

	marshalizedPa0, _ := arguments.Marshalizer.Marshal(pa0)
	marshalizedPaMeta, _ := arguments.Marshalizer.Marshal(paMeta)

	peerAdapter := getAccountsMock()
	peerAdapter.GetAllLeavesCalled = func(ch *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.TrieLeafParser) error {
		if bytes.Equal(rootHash, hash) {
			go func() {
				ch.LeavesChan <- keyValStorage.NewKeyValStorage(addrBytes0, marshalizedPa0)
				ch.LeavesChan <- keyValStorage.NewKeyValStorage(addrBytesMeta, marshalizedPaMeta)
				close(ch.LeavesChan)
				ch.ErrChan.Close()
			}()

			return nil
		}
		return expectedErr
	}
	peerAdapter.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, err error) {
		if bytes.Equal(pa0.AddressBytes(), address) {
			return pa0, nil
		}
		return nil, expectedErr
	}
	arguments.PeerAdapter = peerAdapter

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	validatorInfos, _ := validatorStatistics.GetValidatorInfoForRootHash(hash)
	vi0 := validatorInfos.GetShardValidatorsInfoMap()[0][0]
	newTempRating := uint32(25)
	vi0.SetTempRating(newTempRating)

	assert.NotEqual(t, newTempRating, pa0.GetRating())

	err := validatorStatistics.Process(vi0)

	assert.Nil(t, err)
	assert.Equal(t, newTempRating, pa0.GetRating())
}

func TestValidatorStatistics_GetValidatorInfoForRootHash(t *testing.T) {
	t.Parallel()

	hash := []byte("correctRootHash")
	expectedErr := errors.New("error rootHash")
	arguments := createMockArguments()

	addrBytes0 := []byte("addr1")
	addrBytesMeta := []byte("addrMeta")

	t.Run("should fail on getting all leaves from trie", func(t *testing.T) {
		peerAdapter := getAccountsMock()

		peerAdapter.GetAllLeavesCalled = func(ch *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.TrieLeafParser) error {
			if bytes.Equal(rootHash, hash) {
				go func() {
					ch.ErrChan.WriteInChanNonBlocking(expectedErr)
					close(ch.LeavesChan)
				}()

				return nil
			}
			return expectedErr
		}
		arguments.PeerAdapter = peerAdapter

		validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

		validatorInfos, err := validatorStatistics.GetValidatorInfoForRootHash(hash)
		assert.Nil(t, validatorInfos)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		pa0, paMeta := createPeerAccounts(addrBytes0, addrBytesMeta)

		marshalizedPa0, _ := arguments.Marshalizer.Marshal(pa0)
		marshalizedPaMeta, _ := arguments.Marshalizer.Marshal(paMeta)

		peerAdapter := getAccountsMock()
		peerAdapter.GetAllLeavesCalled = func(ch *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.TrieLeafParser) error {
			if bytes.Equal(rootHash, hash) {
				go func() {
					ch.LeavesChan <- keyValStorage.NewKeyValStorage(addrBytes0, marshalizedPa0)
					ch.LeavesChan <- keyValStorage.NewKeyValStorage(addrBytesMeta, marshalizedPaMeta)
					close(ch.LeavesChan)
					ch.ErrChan.Close()
				}()

				return nil
			}
			return expectedErr
		}
		arguments.PeerAdapter = peerAdapter

		validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

		validatorInfos, err := validatorStatistics.GetValidatorInfoForRootHash(hash)
		assert.NotNil(t, validatorInfos)
		assert.Nil(t, err)
		assert.Equal(t, uint32(0), validatorInfos.GetShardValidatorsInfoMap()[0][0].GetShardId())
		compare(t, pa0, validatorInfos.GetShardValidatorsInfoMap()[0][0])
		assert.Equal(t, core.MetachainShardId, validatorInfos.GetShardValidatorsInfoMap()[core.MetachainShardId][0].GetShardId())
		compare(t, paMeta, validatorInfos.GetShardValidatorsInfoMap()[core.MetachainShardId][0])
	})
}

func TestValidatorStatistics_ProcessValidatorInfosEndOfEpochWithNilMapShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	err := validatorStatistics.ProcessRatingsEndOfEpoch(nil, 1)
	assert.Equal(t, process.ErrNilValidatorInfos, err)

	vi := state.NewShardValidatorsInfoMap()
	err = validatorStatistics.ProcessRatingsEndOfEpoch(vi, 1)
	assert.Equal(t, process.ErrNilValidatorInfos, err)
}

func TestValidatorStatistics_ProcessValidatorInfosEndOfEpochWithNoValidatorFailureShouldNotChangeTempRating(t *testing.T) {
	arguments := createMockArguments()
	rater := createMockRater()
	rater.GetSignedBlocksThresholdCalled = func() float32 {
		return 0.025
	}
	arguments.Rater = rater

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	tempRating1 := uint32(75)
	tempRating2 := uint32(80)

	vi := state.NewShardValidatorsInfoMap()
	_ = vi.Add(&state.ValidatorInfo{
		PublicKey:                  nil,
		ShardId:                    core.MetachainShardId,
		List:                       "",
		Index:                      0,
		TempRating:                 tempRating1,
		Rating:                     0,
		RewardAddress:              nil,
		LeaderSuccess:              10,
		LeaderFailure:              0,
		ValidatorSuccess:           10,
		ValidatorFailure:           0,
		NumSelectedInSuccessBlocks: 20,
		AccumulatedFees:            nil,
	})
	_ = vi.Add(&state.ValidatorInfo{
		PublicKey:                  nil,
		ShardId:                    0,
		List:                       "",
		Index:                      0,
		TempRating:                 tempRating2,
		Rating:                     0,
		RewardAddress:              nil,
		LeaderSuccess:              10,
		LeaderFailure:              0,
		ValidatorSuccess:           10,
		ValidatorFailure:           0,
		NumSelectedInSuccessBlocks: 20,
		AccumulatedFees:            nil,
	})

	err := validatorStatistics.ProcessRatingsEndOfEpoch(vi, 1)
	assert.Nil(t, err)
	assert.Equal(t, tempRating1, vi.GetShardValidatorsInfoMap()[core.MetachainShardId][0].GetTempRating())
	assert.Equal(t, tempRating2, vi.GetShardValidatorsInfoMap()[0][0].GetTempRating())
}

func TestValidatorStatistics_ProcessValidatorInfosEndOfEpochWithSmallValidatorFailureShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	rater := createMockRater()
	rater.GetSignedBlocksThresholdCalled = func() float32 {
		return 0.025
	}
	rater.MinRating = 1000
	rater.MaxRating = 10000
	arguments.Rater = rater

	updateArgumentsWithNeeded(arguments)

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	tempRating1 := uint32(5000)
	tempRating2 := uint32(8000)

	validatorSuccess1 := uint32(2)
	validatorIgnored1 := uint32(90)
	validatorFailure1 := uint32(8)
	validatorSuccess2 := uint32(1)
	validatorIgnored2 := uint32(90)
	validatorFailure2 := uint32(9)

	vi := state.NewShardValidatorsInfoMap()
	_ = vi.Add(createMockValidatorInfo(core.MetachainShardId, tempRating1, validatorSuccess1, validatorIgnored1, validatorFailure1))
	_ = vi.Add(createMockValidatorInfo(0, tempRating2, validatorSuccess2, validatorIgnored2, validatorFailure2))

	err := validatorStatistics.ProcessRatingsEndOfEpoch(vi, 1)
	assert.Nil(t, err)
	expectedTempRating1 := tempRating1 - uint32(rater.MetaIncreaseValidator)*(validatorSuccess1+validatorIgnored1)
	assert.Equal(t, expectedTempRating1, vi.GetShardValidatorsInfoMap()[core.MetachainShardId][0].GetTempRating())
	expectedTempRating2 := tempRating2 - uint32(rater.IncreaseValidator)*(validatorSuccess2+validatorIgnored2)
	assert.Equal(t, expectedTempRating2, vi.GetShardValidatorsInfoMap()[0][0].GetTempRating())
}

func TestValidatorStatistics_ProcessValidatorInfosEndOfEpochComputesJustEligible(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	rater := createMockRater()
	rater.GetSignedBlocksThresholdCalled = func() float32 {
		return 0.025
	}
	rater.MinRating = 1000
	rater.MaxRating = 10000

	arguments.Rater = rater

	updateArgumentsWithNeeded(arguments)

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	tempRating1 := uint32(5000)
	tempRating2 := uint32(8000)

	validatorSuccess1 := uint32(2)
	validatorIgnored1 := uint32(90)
	validatorFailure1 := uint32(8)
	validatorSuccess2 := uint32(1)
	validatorIgnored2 := uint32(90)
	validatorFailure2 := uint32(9)

	vi := state.NewShardValidatorsInfoMap()
	_ = vi.Add(createMockValidatorInfo(core.MetachainShardId, tempRating1, validatorSuccess1, validatorIgnored1, validatorFailure1))

	validatorWaiting := createMockValidatorInfo(0, tempRating2, validatorSuccess2, validatorIgnored2, validatorFailure2)
	validatorWaiting.SetList(string(common.WaitingList))
	_ = vi.Add(validatorWaiting)

	err := validatorStatistics.ProcessRatingsEndOfEpoch(vi, 1)
	assert.Nil(t, err)
	expectedTempRating1 := tempRating1 - uint32(rater.MetaIncreaseValidator)*(validatorSuccess1+validatorIgnored1)
	assert.Equal(t, expectedTempRating1, vi.GetShardValidatorsInfoMap()[core.MetachainShardId][0].GetTempRating())

	assert.Equal(t, tempRating2, vi.GetShardValidatorsInfoMap()[0][0].GetTempRating())
}

func TestValidatorStatistics_ProcessValidatorInfosEndOfEpochV2ComputesEligibleLeaving(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	rater := createMockRater()
	rater.GetSignedBlocksThresholdCalled = func() float32 {
		return 0.025
	}
	rater.MinRating = 1000
	rater.MaxRating = 10000

	arguments.Rater = rater

	updateArgumentsWithNeeded(arguments)
	enableEpochsHandler, _ := arguments.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	enableEpochsHandler.AddActiveFlags(common.StakingV2FlagAfterEpoch)

	tempRating1 := uint32(5000)
	tempRating2 := uint32(8000)

	validatorSuccess1 := uint32(2)
	validatorIgnored1 := uint32(90)
	validatorFailure1 := uint32(8)
	validatorSuccess2 := uint32(1)
	validatorIgnored2 := uint32(90)
	validatorFailure2 := uint32(9)

	vi := state.NewShardValidatorsInfoMap()
	validatorLeaving := createMockValidatorInfo(core.MetachainShardId, tempRating1, validatorSuccess1, validatorIgnored1, validatorFailure1)
	validatorLeaving.SetList(string(common.LeavingList))
	_ = vi.Add(validatorLeaving)

	validatorWaiting := createMockValidatorInfo(0, tempRating2, validatorSuccess2, validatorIgnored2, validatorFailure2)
	validatorWaiting.SetList(string(common.WaitingList))
	_ = vi.Add(validatorWaiting)

	err := validatorStatistics.ProcessRatingsEndOfEpoch(vi, 1)
	assert.Nil(t, err)
	expectedTempRating1 := tempRating1 - uint32(rater.MetaIncreaseValidator)*(validatorSuccess1+validatorIgnored1)
	assert.Equal(t, expectedTempRating1, vi.GetShardValidatorsInfoMap()[core.MetachainShardId][0].GetTempRating())

	assert.Equal(t, tempRating2, vi.GetShardValidatorsInfoMap()[0][0].GetTempRating())
}

func TestValidatorStatistics_ProcessValidatorInfosEndOfEpochWithLargeValidatorFailureBelowMinRatingShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	rater := createMockRater()
	rater.GetSignedBlocksThresholdCalled = func() float32 {
		return 0.025
	}
	rater.MinRating = 1000
	rater.MaxRating = 10000
	arguments.Rater = rater
	rater.MetaIncreaseValidator = 100
	rater.IncreaseValidator = 99
	updateArgumentsWithNeeded(arguments)

	tempRating1 := uint32(5000)
	tempRating2 := uint32(8000)

	validatorSuccess1 := uint32(2)
	validatorIgnored1 := uint32(90)
	validatorFailure1 := uint32(8)
	validatorSuccess2 := uint32(1)
	validatorIgnored2 := uint32(90)
	validatorFailure2 := uint32(9)

	vi := state.NewShardValidatorsInfoMap()
	_ = vi.Add(createMockValidatorInfo(core.MetachainShardId, tempRating1, validatorSuccess1, validatorIgnored1, validatorFailure1))
	_ = vi.Add(createMockValidatorInfo(0, tempRating2, validatorSuccess2, validatorIgnored2, validatorFailure2))

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	err := validatorStatistics.ProcessRatingsEndOfEpoch(vi, 1)

	assert.Nil(t, err)
	assert.Equal(t, rater.MinRating, vi.GetShardValidatorsInfoMap()[core.MetachainShardId][0].GetTempRating())
	assert.Equal(t, rater.MinRating, vi.GetShardValidatorsInfoMap()[0][0].GetTempRating())
}

func TestValidatorsProvider_PeerAccoutToValidatorInfo(t *testing.T) {
	t.Parallel()

	baseRating := uint32(50)
	rating := uint32(70)
	chancesForStartRating := uint32(20)
	chancesForRating := uint32(22)
	newRater := createMockRater()
	newRater.GetChancesCalled = func(val uint32) uint32 {
		if val == baseRating {
			return chancesForStartRating
		}
		if val == rating {
			return chancesForRating
		}
		return uint32(0)
	}

	arguments := createMockArguments()
	arguments.Rater = newRater

	pad := accounts.PeerAccountData{
		BLSPublicKey:  []byte("blsKey"),
		ShardId:       7,
		List:          "list",
		IndexInList:   2,
		TempRating:    51,
		Rating:        70,
		RewardAddress: []byte("rewardAddress"),
		LeaderSuccessRate: accounts.SignRate{
			NumSuccess: 1,
			NumFailure: 2,
		},
		ValidatorSuccessRate: accounts.SignRate{
			NumSuccess: 3,
			NumFailure: 4,
		},
		TotalLeaderSuccessRate: accounts.SignRate{
			NumSuccess: 5,
			NumFailure: 6,
		},
		TotalValidatorSuccessRate: accounts.SignRate{
			NumSuccess: 7,
			NumFailure: 8,
		},
		NumSelectedInSuccessBlocks: 3,
		AccumulatedFees:            big.NewInt(70),
		UnStakedEpoch:              common.DefaultUnstakedEpoch,
	}

	peerAccount, _ := accounts.NewPeerAccount([]byte("mock address"))
	peerAccount.PeerAccountData = pad

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	vs := validatorStatistics.PeerAccountToValidatorInfo(peerAccount)

	ratingModifier := float32(chancesForRating) / float32(chancesForStartRating)

	assert.Equal(t, peerAccount.AddressBytes(), vs.PublicKey)
	assert.Equal(t, peerAccount.GetShardId(), vs.ShardId)
	assert.Equal(t, peerAccount.GetList(), vs.List)
	assert.Equal(t, peerAccount.GetIndexInList(), vs.Index)
	assert.Equal(t, peerAccount.GetTempRating(), vs.TempRating)
	assert.Equal(t, peerAccount.GetRating(), vs.Rating)
	assert.Equal(t, ratingModifier, vs.RatingModifier)
	assert.Equal(t, peerAccount.GetRewardAddress(), vs.RewardAddress)
	assert.Equal(t, peerAccount.GetLeaderSuccessRate().GetNumSuccess(), vs.LeaderSuccess)
	assert.Equal(t, peerAccount.GetLeaderSuccessRate().GetNumFailure(), vs.LeaderFailure)
	assert.Equal(t, peerAccount.GetValidatorSuccessRate().GetNumSuccess(), vs.ValidatorSuccess)
	assert.Equal(t, peerAccount.GetValidatorSuccessRate().GetNumFailure(), vs.ValidatorFailure)
	assert.Equal(t, peerAccount.GetTotalLeaderSuccessRate().GetNumSuccess(), vs.TotalLeaderSuccess)
	assert.Equal(t, peerAccount.GetTotalLeaderSuccessRate().GetNumFailure(), vs.TotalLeaderFailure)
	assert.Equal(t, peerAccount.GetTotalValidatorSuccessRate().GetNumSuccess(), vs.TotalValidatorSuccess)
	assert.Equal(t, peerAccount.GetTotalValidatorSuccessRate().GetNumFailure(), vs.TotalValidatorFailure)
	assert.Equal(t, peerAccount.GetNumSelectedInSuccessBlocks(), vs.NumSelectedInSuccessBlocks)
	assert.Equal(t, big.NewInt(0).Set(peerAccount.GetAccumulatedFees()), vs.AccumulatedFees)
}

func createMockValidatorInfo(shardId uint32, tempRating uint32, validatorSuccess uint32, validatorIgnored uint32, validatorFailure uint32) *state.ValidatorInfo {
	return &state.ValidatorInfo{
		PublicKey:                  nil,
		ShardId:                    shardId,
		List:                       string(common.EligibleList),
		Index:                      0,
		TempRating:                 tempRating,
		Rating:                     0,
		RewardAddress:              nil,
		LeaderSuccess:              0,
		LeaderFailure:              0,
		ValidatorSuccess:           validatorSuccess,
		ValidatorIgnoredSignatures: validatorIgnored,
		ValidatorFailure:           validatorFailure,
		NumSelectedInSuccessBlocks: validatorSuccess + validatorFailure,
		AccumulatedFees:            nil,
	}
}

func compare(t *testing.T, peerAccount state.PeerAccountHandler, validatorInfo state.ValidatorInfoHandler) {
	assert.Equal(t, peerAccount.GetShardId(), validatorInfo.GetShardId())
	assert.Equal(t, peerAccount.GetRating(), validatorInfo.GetRating())
	assert.Equal(t, peerAccount.GetTempRating(), validatorInfo.GetTempRating())
	assert.Equal(t, peerAccount.AddressBytes(), validatorInfo.GetPublicKey())
	assert.Equal(t, peerAccount.GetValidatorSuccessRate().GetNumFailure(), validatorInfo.GetValidatorFailure())
	assert.Equal(t, peerAccount.GetValidatorSuccessRate().GetNumSuccess(), validatorInfo.GetValidatorSuccess())
	assert.Equal(t, peerAccount.GetValidatorIgnoredSignaturesRate(), validatorInfo.GetValidatorIgnoredSignatures())
	assert.Equal(t, peerAccount.GetLeaderSuccessRate().GetNumFailure(), validatorInfo.GetLeaderFailure())
	assert.Equal(t, peerAccount.GetLeaderSuccessRate().GetNumSuccess(), validatorInfo.GetLeaderSuccess())
	assert.Equal(t, peerAccount.GetTotalValidatorSuccessRate().GetNumFailure(), validatorInfo.GetTotalValidatorFailure())
	assert.Equal(t, peerAccount.GetTotalValidatorSuccessRate().GetNumSuccess(), validatorInfo.GetTotalValidatorSuccess())
	assert.Equal(t, peerAccount.GetTotalValidatorIgnoredSignaturesRate(), validatorInfo.GetTotalValidatorIgnoredSignatures())
	assert.Equal(t, peerAccount.GetTotalLeaderSuccessRate().GetNumFailure(), validatorInfo.GetTotalLeaderFailure())
	assert.Equal(t, peerAccount.GetTotalLeaderSuccessRate().GetNumSuccess(), validatorInfo.GetTotalLeaderSuccess())
	assert.Equal(t, peerAccount.GetList(), validatorInfo.GetList())
	assert.Equal(t, peerAccount.GetIndexInList(), validatorInfo.GetIndex())
	assert.Equal(t, peerAccount.GetRewardAddress(), validatorInfo.GetRewardAddress())
	assert.Equal(t, peerAccount.GetAccumulatedFees(), validatorInfo.GetAccumulatedFees())
	assert.Equal(t, peerAccount.GetNumSelectedInSuccessBlocks(), validatorInfo.GetNumSelectedInSuccessBlocks())
}

func createPeerAccounts(addrBytes0 []byte, addrBytesMeta []byte) (state.PeerAccountHandler, state.PeerAccountHandler) {
	addr := addrBytes0
	pa0, _ := accounts.NewPeerAccount(addr)
	pa0.PeerAccountData = accounts.PeerAccountData{
		BLSPublicKey:    []byte("bls0"),
		RewardAddress:   []byte("reward0"),
		AccumulatedFees: big.NewInt(11),
		ValidatorSuccessRate: accounts.SignRate{
			NumSuccess: 1,
			NumFailure: 2,
		},
		LeaderSuccessRate: accounts.SignRate{
			NumSuccess: 3,
			NumFailure: 4,
		},
		ValidatorIgnoredSignaturesRate: 5,
		TotalValidatorSuccessRate: accounts.SignRate{
			NumSuccess: 10,
			NumFailure: 20,
		},
		TotalLeaderSuccessRate: accounts.SignRate{
			NumSuccess: 30,
			NumFailure: 40,
		},
		TotalValidatorIgnoredSignaturesRate: 50,
		NumSelectedInSuccessBlocks:          5,
		Rating:                              51,
		TempRating:                          61,
		Nonce:                               7,
		UnStakedEpoch:                       common.DefaultUnstakedEpoch,
	}

	addr = addrBytesMeta
	paMeta, _ := accounts.NewPeerAccount(addr)
	paMeta.PeerAccountData = accounts.PeerAccountData{
		BLSPublicKey:    []byte("blsM"),
		RewardAddress:   []byte("rewardM"),
		AccumulatedFees: big.NewInt(111),
		ValidatorSuccessRate: accounts.SignRate{
			NumSuccess: 11,
			NumFailure: 21,
		},
		LeaderSuccessRate: accounts.SignRate{
			NumSuccess: 31,
			NumFailure: 41,
		},
		NumSelectedInSuccessBlocks: 3,
		Rating:                     511,
		TempRating:                 611,
		Nonce:                      8,
		ShardId:                    core.MetachainShardId,
		UnStakedEpoch:              common.DefaultUnstakedEpoch,
	}
	return pa0, paMeta
}

func updateArgumentsWithNeeded(arguments peer.ArgValidatorStatisticsProcessor) {
	addrBytes0 := []byte("addr1")
	addrBytesMeta := []byte("addrMeta")

	pa0, paMeta := createPeerAccounts(addrBytes0, addrBytesMeta)

	marshalizedPa0, _ := arguments.Marshalizer.Marshal(pa0)
	marshalizedPaMeta, _ := arguments.Marshalizer.Marshal(paMeta)

	peerAdapter := getAccountsMock()
	peerAdapter.GetAllLeavesCalled = func(ch *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.TrieLeafParser) error {
		go func() {
			ch.LeavesChan <- keyValStorage.NewKeyValStorage(addrBytes0, marshalizedPa0)
			ch.LeavesChan <- keyValStorage.NewKeyValStorage(addrBytesMeta, marshalizedPaMeta)
			close(ch.LeavesChan)
			ch.ErrChan.Close()
		}()

		return nil
	}
	peerAdapter.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, err error) {
		return pa0, nil
	}
	arguments.PeerAdapter = peerAdapter
}

func createUpdateTestArgs(consensusGroup map[string][]nodesCoordinator.Validator) peer.ArgValidatorStatisticsProcessor {
	peerAccountsMap := make(map[string]state.PeerAccountHandler)
	arguments := createMockArguments()

	arguments.Rater = mock.GetNewMockRater()
	adapter := getAccountsMock()
	adapter.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		pk := string(address)
		_, ok := peerAccountsMap[pk]
		if !ok {
			peerAccountsMap[pk] = &stateMock.PeerAccountHandlerMock{}
		}
		return peerAccountsMap[pk], nil
	}
	adapter.RootHashCalled = func() ([]byte, error) {
		return nil, nil
	}
	arguments.PeerAdapter = adapter

	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			key := fmt.Sprintf(consensusGroupFormat, string(randomness), round, shardId, epoch)
			validatorsArray, ok := consensusGroup[key]
			if !ok {
				return nil, process.ErrEmptyConsensusGroup
			}
			return validatorsArray, nil
		},
	}
	return arguments
}

func TestValidatorStatisticsProcessor_SaveNodesCoordinatorUpdates(t *testing.T) {
	t.Parallel()

	peerAdapter := getAccountsMock()
	arguments := createMockArguments()
	arguments.PeerAdapter = peerAdapter

	peerAdapter.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		peerAcc, _ := accounts.NewPeerAccount(address)
		peerAcc.List = string(common.LeavingList)
		return peerAcc, nil
	}

	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func(_ uint32) (map[uint32][][]byte, error) {
			mapNodes := make(map[uint32][][]byte)
			mapNodes[0] = [][]byte{[]byte("someAddress")}
			return mapNodes, nil
		},
	}

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	nodeForcedToRemain, err := validatorStatistics.SaveNodesCoordinatorUpdates(0)
	assert.Nil(t, err)
	assert.True(t, nodeForcedToRemain)

	peerAdapter.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		return accounts.NewPeerAccount(address)
	}
	nodeForcedToRemain, err = validatorStatistics.SaveNodesCoordinatorUpdates(0)
	assert.Nil(t, err)
	assert.False(t, nodeForcedToRemain)
}

func TestValidatorStatisticsProcessor_SaveNodesCoordinatorUpdatesWithStakingV4(t *testing.T) {
	t.Parallel()

	peerAdapter := getAccountsMock()
	arguments := createMockArguments()
	arguments.PeerAdapter = peerAdapter

	pk0 := []byte("pk0")
	pk1 := []byte("pk1")
	pk2 := []byte("pk2")

	account0, _ := accounts.NewPeerAccount(pk0)
	account1, _ := accounts.NewPeerAccount(pk1)
	account2, _ := accounts.NewPeerAccount(pk2)

	ctLoadAccount := &atomic.Counter{}
	ctSaveAccount := &atomic.Counter{}

	peerAdapter.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		ctLoadAccount.Increment()

		switch string(address) {
		case string(pk0):
			return account0, nil
		case string(pk1):
			return account1, nil
		case string(pk2):
			return account2, nil
		default:
			require.Fail(t, "should not have called this for other address")
			return nil, nil
		}
	}
	peerAdapter.SaveAccountCalled = func(account vmcommon.AccountHandler) error {
		ctSaveAccount.Increment()
		peerAccount := account.(state.PeerAccountHandler)
		require.Equal(t, uint32(0), peerAccount.GetIndexInList())

		switch string(account.AddressBytes()) {
		case string(pk0):
			require.Equal(t, string(common.EligibleList), peerAccount.GetList())
			require.Equal(t, uint32(0), peerAccount.GetShardId())
			return nil
		case string(pk1):
			require.Equal(t, string(common.AuctionList), peerAccount.GetList())
			require.Equal(t, uint32(0), peerAccount.GetShardId())
			return nil
		case string(pk2):
			require.Equal(t, string(common.AuctionList), peerAccount.GetList())
			require.Equal(t, core.MetachainShardId, peerAccount.GetShardId())
			return nil
		default:
			require.Fail(t, "should not have called this for other account")
			return nil
		}
	}

	arguments.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
			mapNodes := map[uint32][][]byte{
				0: {pk0},
			}
			return mapNodes, nil
		},
		GetAllShuffledOutValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
			mapNodes := map[uint32][][]byte{
				0:                     {pk1},
				core.MetachainShardId: {pk2},
			}
			return mapNodes, nil
		},
		GetShuffledOutToAuctionValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
			mapNodes := map[uint32][][]byte{
				0:                     {pk1},
				core.MetachainShardId: {pk2},
			}
			return mapNodes, nil
		},
	}
	stakingV4Step2EnableEpochCalledCt := 0
	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			if flag == common.StakingV4Step2Flag {
				stakingV4Step2EnableEpochCalledCt++
				switch stakingV4Step2EnableEpochCalledCt {
				case 1:
					return false
				case 2:
					return true
				default:
					require.Fail(t, "should only call this twice")
				}
			}

			return false
		},
	}

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	nodeForcedToRemain, err := validatorStatistics.SaveNodesCoordinatorUpdates(0)
	require.Nil(t, err)
	require.False(t, nodeForcedToRemain)
	require.Equal(t, int64(1), ctSaveAccount.Get())
	require.Equal(t, int64(1), ctLoadAccount.Get())

	ctSaveAccount.Reset()
	ctLoadAccount.Reset()

	nodeForcedToRemain, err = validatorStatistics.SaveNodesCoordinatorUpdates(0)
	require.Nil(t, err)
	require.False(t, nodeForcedToRemain)
	require.Equal(t, int64(3), ctSaveAccount.Get())
	require.Equal(t, int64(3), ctLoadAccount.Get())
}

func TestValidatorStatisticsProcessor_getActualList(t *testing.T) {
	t.Parallel()

	eligibleList := string(common.EligibleList)
	eligiblePeer := &stateMock.PeerAccountHandlerMock{
		GetListCalled: func() string {
			return eligibleList
		},
	}
	computedEligibleList := peer.GetActualList(eligiblePeer)
	assert.Equal(t, eligibleList, computedEligibleList)

	waitingList := string(common.WaitingList)
	waitingPeer := &stateMock.PeerAccountHandlerMock{
		GetListCalled: func() string {
			return waitingList
		},
	}
	computedWaiting := peer.GetActualList(waitingPeer)
	assert.Equal(t, waitingList, computedWaiting)

	leavingList := string(common.LeavingList)
	leavingPeer := &stateMock.PeerAccountHandlerMock{
		GetListCalled: func() string {
			return leavingList
		},
	}
	computedLeavingList := peer.GetActualList(leavingPeer)
	assert.Equal(t, leavingList, computedLeavingList)

	newList := string(common.NewList)
	newPeer := &stateMock.PeerAccountHandlerMock{
		GetListCalled: func() string {
			return newList
		},
	}
	computedNewList := peer.GetActualList(newPeer)
	assert.Equal(t, newList, computedNewList)

	inactiveList := string(common.InactiveList)
	inactivePeer := &stateMock.PeerAccountHandlerMock{
		GetListCalled: func() string {
			return inactiveList
		},
		GetUnStakedEpochCalled: func() uint32 {
			return 2
		},
	}
	computedInactiveList := peer.GetActualList(inactivePeer)
	assert.Equal(t, inactiveList, computedInactiveList)

	inactivePeer2 := &stateMock.PeerAccountHandlerMock{
		GetListCalled: func() string {
			return inactiveList
		},
		GetUnStakedEpochCalled: func() uint32 {
			return 0
		},
	}
	computedInactiveList = peer.GetActualList(inactivePeer2)
	assert.Equal(t, inactiveList, computedInactiveList)

	jailedList := string(common.JailedList)
	jailedPeer := &stateMock.PeerAccountHandlerMock{
		GetListCalled: func() string {
			return inactiveList
		},
		GetUnStakedEpochCalled: func() uint32 {
			return common.DefaultUnstakedEpoch
		},
	}
	computedJailedList := peer.GetActualList(jailedPeer)
	assert.Equal(t, jailedList, computedJailedList)
}
