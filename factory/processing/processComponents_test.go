package processing_test

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"strings"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	outportCore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/factory"
	disabledStatistics "github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/mock"
	processComp "github.com/multiversx/mx-chain-go/factory/processing"
	"github.com/multiversx/mx-chain-go/genesis"
	genesisMocks "github.com/multiversx/mx-chain-go/genesis/mock"
	testsMocks "github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks"
	txExecOrderStub "github.com/multiversx/mx-chain-go/testscommon/common"
	"github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	factoryMocks "github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	nodesSetupMock "github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/guardianMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	testState "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	updateMocks "github.com/multiversx/mx-chain-go/update/mock"
	"github.com/stretchr/testify/require"
)

const (
	testingProtocolSustainabilityAddress = "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp"
)

var (
	gasSchedule, _    = common.LoadGasScheduleConfig("../../cmd/node/config/gasSchedules/" + common.LatestGasScheduleFileName)
	addrPubKeyConv, _ = factory.NewPubkeyConverter(config.PubkeyConfig{
		Length:          32,
		Type:            "bech32",
		SignatureLength: 0,
		Hrp:             "erd",
	})
	valPubKeyConv, _ = factory.NewPubkeyConverter(config.PubkeyConfig{
		Length:          96,
		Type:            "hex",
		SignatureLength: 48,
	})
)

func createMockProcessComponentsFactoryArgs() processComp.ProcessComponentsFactoryArgs {

	args := processComp.ProcessComponentsFactoryArgs{
		Config: testscommon.GetGeneralConfig(),
		EpochConfig: config.EpochConfig{
			EnableEpochs: config.EnableEpochs{
				MaxNodesChangeEnableEpoch: []config.MaxNodesChangeConfig{
					{
						EpochEnable:            0,
						MaxNumNodes:            100,
						NodesToShufflePerShard: 2,
					},
				},
			},
		},
		RoundConfig:    testscommon.GetDefaultRoundsConfig(),
		PrefConfigs:    config.Preferences{},
		ImportDBConfig: config.ImportDbConfig{},
		FlagsConfig: config.ContextFlagsConfig{
			Version: "v1.0.0",
		},
		AccountsParser: &mock.AccountsParserStub{
			GenerateInitialTransactionsCalled: func(shardCoordinator sharding.Coordinator, initialIndexingData map[uint32]*genesis.IndexingData) ([]*dataBlock.MiniBlock, map[uint32]*outportCore.TransactionPool, error) {
				return []*dataBlock.MiniBlock{
						{},
					},
					map[uint32]*outportCore.TransactionPool{
						0: {},
					}, nil
			},
		},
		SmartContractParser: &mock.SmartContractParserStub{},
		GasSchedule: &testscommon.GasScheduleNotifierMock{
			GasSchedule: gasSchedule,
		},
		NodesCoordinator:       &shardingMocks.NodesCoordinatorStub{},
		RequestedItemsHandler:  &testscommon.RequestedItemsHandlerStub{},
		WhiteListHandler:       &testscommon.WhiteListHandlerStub{},
		WhiteListerVerifiedTxs: &testscommon.WhiteListHandlerStub{},
		MaxRating:              100,
		SystemSCConfig: &config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "1000",
				OwnerAddress:    "erd1fpkcgel4gcmh8zqqdt043yfcn5tyx8373kg6q2qmkxzu4dqamc0swts65c",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				V1: config.GovernanceSystemSCConfigV1{
					ProposalCost:     "500",
					NumNodes:         100,
					MinQuorum:        50,
					MinPassThreshold: 50,
					MinVetoThreshold: 50,
				},
				Active: config.GovernanceSystemSCConfigActive{
					ProposalCost:     "500",
					MinQuorum:        0.5,
					MinPassThreshold: 0.5,
					MinVetoThreshold: 0.5,
				},
				OwnerAddress: "erd1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awqgqky80",
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "2500",
				MinStakeValue:                        "1",
				UnJailValue:                          "1",
				MinStepValue:                         "1",
				UnBondPeriod:                         0,
				NumRoundsWithoutBleed:                0,
				MaximumPercentageToBleed:             0,
				BleedPercentagePerRound:              0,
				MaxNumberOfNodesForStake:             10,
				ActivateBLSPubKeyMessageVerification: false,
				MinUnstakeTokensValue:                "1",
				NodeLimitPercentage:                  100.0,
				StakeLimitPercentage:                 100.0,
			},
			DelegationManagerSystemSCConfig: config.DelegationManagerSystemSCConfig{
				MinCreationDeposit:  "100",
				MinStakeAmount:      "100",
				ConfigChangeAddress: "erd1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awqgqky80",
			},
			DelegationSystemSCConfig: config.DelegationSystemSCConfig{
				MinServiceFee: 0,
				MaxServiceFee: 100,
			},
			SoftAuctionConfig: config.SoftAuctionConfig{
				TopUpStep:             "10",
				MinTopUp:              "1",
				MaxTopUp:              "32000000",
				MaxNumberOfIterations: 100000,
			},
		},
		ImportStartHandler: &testscommon.ImportStartHandlerStub{},
		HistoryRepo:        &dblookupext.HistoryRepositoryStub{},
		Data: &testsMocks.DataComponentsStub{
			DataPool: dataRetriever.NewPoolsHolderMock(),
			BlockChain: &testscommon.ChainHandlerStub{
				GetGenesisHeaderHashCalled: func() []byte {
					return []byte("genesis hash")
				},
				GetGenesisHeaderCalled: func() coreData.HeaderHandler {
					return &testscommon.HeaderHandlerStub{}
				},
			},
			MbProvider: &testsMocks.MiniBlocksProviderStub{},
			Store:      genericMocks.NewChainStorerMock(0),
		},
		CoreData: &mock.CoreComponentsMock{
			IntMarsh:            &marshal.GogoProtoMarshalizer{},
			TxMarsh:             &marshal.JsonMarshalizer{},
			UInt64ByteSliceConv: &testsMocks.Uint64ByteSliceConverterMock{},
			AddrPubKeyConv:      addrPubKeyConv,
			ValPubKeyConv:       valPubKeyConv,
			NodesConfig: &nodesSetupMock.NodesSetupStub{
				GetShardConsensusGroupSizeCalled: func() uint32 {
					return 2
				},
				GetMetaConsensusGroupSizeCalled: func() uint32 {
					return 2
				},
			},
			EpochChangeNotifier: &epochNotifier.EpochNotifierStub{},
			EconomicsHandler: &economicsmocks.EconomicsHandlerStub{
				ProtocolSustainabilityAddressCalled: func() string {
					return testingProtocolSustainabilityAddress
				},
			},
			Hash:                         blake2b.NewBlake2b(),
			TxVersionCheckHandler:        &testscommon.TxVersionCheckerStub{},
			RatingHandler:                &testscommon.RaterMock{},
			EnableEpochsHandlerField:     &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			EnableRoundsHandlerField:     &testscommon.EnableRoundsHandlerStub{},
			EpochNotifierWithConfirm:     &updateMocks.EpochStartNotifierStub{},
			RoundHandlerField:            &testscommon.RoundHandlerMock{},
			RoundChangeNotifier:          &epochNotifier.RoundNotifierStub{},
			ChanStopProcess:              make(chan endProcess.ArgEndProcess, 1),
			TxSignHasherField:            keccak.NewKeccak(),
			HardforkTriggerPubKeyField:   []byte("hardfork pub key"),
			WasmVMChangeLockerInternal:   &sync.RWMutex{},
			NodeTypeProviderField:        &nodeTypeProviderMock.NodeTypeProviderStub{},
			RatingsConfig:                &testscommon.RatingsInfoMock{},
			PathHdl:                      &testscommon.PathManagerStub{},
			ProcessStatusHandlerInternal: &testscommon.ProcessStatusHandlerStub{},
		},
		Crypto: &testsMocks.CryptoComponentsStub{
			BlKeyGen: &cryptoMocks.KeyGenStub{},
			BlockSig: &cryptoMocks.SingleSignerStub{},
			MultiSigContainer: &cryptoMocks.MultiSignerContainerMock{
				MultiSigner: &cryptoMocks.MultisignerMock{},
			},
			PrivKey:                 &cryptoMocks.PrivateKeyStub{},
			PubKey:                  &cryptoMocks.PublicKeyStub{},
			PubKeyString:            "pub key string",
			PubKeyBytes:             []byte("pub key bytes"),
			TxKeyGen:                &cryptoMocks.KeyGenStub{},
			TxSig:                   &cryptoMocks.SingleSignerStub{},
			PeerSignHandler:         &cryptoMocks.PeerSignatureHandlerStub{},
			MsgSigVerifier:          &testscommon.MessageSignVerifierMock{},
			ManagedPeersHolderField: &testscommon.ManagedPeersHolderStub{},
			KeysHandlerField:        &testscommon.KeysHandlerStub{},
		},
		Network: &testsMocks.NetworkComponentsStub{
			Messenger:                        &p2pmocks.MessengerStub{},
			FullArchiveNetworkMessengerField: &p2pmocks.MessengerStub{},
			InputAntiFlood:                   &testsMocks.P2PAntifloodHandlerStub{},
			OutputAntiFlood:                  &testsMocks.P2PAntifloodHandlerStub{},
			PreferredPeersHolder:             &p2pmocks.PeersHolderStub{},
			PeersRatingHandlerField:          &p2pmocks.PeersRatingHandlerStub{},
			FullArchivePreferredPeersHolder:  &p2pmocks.PeersHolderStub{},
		},
		BootstrapComponents: &mainFactoryMocks.BootstrapComponentsStub{
			ShCoordinator:              mock.NewMultiShardsCoordinatorMock(2),
			BootstrapParams:            &bootstrapMocks.BootstrapParamsHandlerMock{},
			HdrIntegrityVerifier:       &mock.HeaderIntegrityVerifierStub{},
			GuardedAccountHandlerField: &guardianMocks.GuardedAccountHandlerStub{},
			VersionedHdrFactory:        &testscommon.VersionedHeaderFactoryStub{},
		},
		StatusComponents: &testsMocks.StatusComponentsStub{
			Outport: &outport.OutportStub{},
		},
		StatusCoreComponents: &factoryMocks.StatusCoreComponentsStub{
			AppStatusHandlerField:  &statusHandler.AppStatusHandlerStub{},
			StateStatsHandlerField: disabledStatistics.NewStateStatistics(),
		},
		TxExecutionOrderHandler: &txExecOrderStub.TxExecutionOrderHandlerStub{},
	}

	args.State = components.GetStateComponents(args.CoreData, args.StatusCoreComponents)

	return args
}

func TestNewProcessComponentsFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil AccountsParser should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.AccountsParser = nil
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilAccountsParser))
		require.Nil(t, pcf)
	})
	t.Run("nil GasSchedule should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.GasSchedule = nil
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilGasSchedule))
		require.Nil(t, pcf)
	})
	t.Run("nil Data should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Data = nil
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilDataComponentsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil BlockChain should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Data = &testsMocks.DataComponentsStub{
			BlockChain: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilBlockChainHandler))
		require.Nil(t, pcf)
	})
	t.Run("nil DataPool should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Data = &testsMocks.DataComponentsStub{
			BlockChain: &testscommon.ChainHandlerStub{},
			DataPool:   nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilDataPoolsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil StorageService should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Data = &testsMocks.DataComponentsStub{
			BlockChain: &testscommon.ChainHandlerStub{},
			DataPool:   &dataRetriever.PoolsHolderStub{},
			Store:      nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilStorageService))
		require.Nil(t, pcf)
	})
	t.Run("nil CoreData should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = nil
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilCoreComponentsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil EconomicsData should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = &mock.CoreComponentsMock{
			EconomicsHandler: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilEconomicsData))
		require.Nil(t, pcf)
	})
	t.Run("nil GenesisNodesSetup should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = &mock.CoreComponentsMock{
			EconomicsHandler: &economicsmocks.EconomicsHandlerStub{},
			NodesConfig:      nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilGenesisNodesSetupHandler))
		require.Nil(t, pcf)
	})
	t.Run("nil AddressPubKeyConverter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = &mock.CoreComponentsMock{
			EconomicsHandler: &economicsmocks.EconomicsHandlerStub{},
			NodesConfig:      &nodesSetupMock.NodesSetupStub{},
			AddrPubKeyConv:   nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilAddressPublicKeyConverter))
		require.Nil(t, pcf)
	})
	t.Run("nil EpochNotifier should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = &mock.CoreComponentsMock{
			EconomicsHandler:    &economicsmocks.EconomicsHandlerStub{},
			NodesConfig:         &nodesSetupMock.NodesSetupStub{},
			AddrPubKeyConv:      &testscommon.PubkeyConverterStub{},
			EpochChangeNotifier: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilEpochNotifier))
		require.Nil(t, pcf)
	})
	t.Run("nil ValidatorPubKeyConverter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = &mock.CoreComponentsMock{
			EconomicsHandler:    &economicsmocks.EconomicsHandlerStub{},
			NodesConfig:         &nodesSetupMock.NodesSetupStub{},
			AddrPubKeyConv:      &testscommon.PubkeyConverterStub{},
			EpochChangeNotifier: &epochNotifier.EpochNotifierStub{},
			ValPubKeyConv:       nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilPubKeyConverter))
		require.Nil(t, pcf)
	})
	t.Run("nil InternalMarshalizer should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = &mock.CoreComponentsMock{
			EconomicsHandler:    &economicsmocks.EconomicsHandlerStub{},
			NodesConfig:         &nodesSetupMock.NodesSetupStub{},
			AddrPubKeyConv:      &testscommon.PubkeyConverterStub{},
			EpochChangeNotifier: &epochNotifier.EpochNotifierStub{},
			ValPubKeyConv:       &testscommon.PubkeyConverterStub{},
			IntMarsh:            nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilInternalMarshalizer))
		require.Nil(t, pcf)
	})
	t.Run("nil Uint64ByteSliceConverter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = &mock.CoreComponentsMock{
			EconomicsHandler:    &economicsmocks.EconomicsHandlerStub{},
			NodesConfig:         &nodesSetupMock.NodesSetupStub{},
			AddrPubKeyConv:      &testscommon.PubkeyConverterStub{},
			EpochChangeNotifier: &epochNotifier.EpochNotifierStub{},
			ValPubKeyConv:       &testscommon.PubkeyConverterStub{},
			IntMarsh:            &marshallerMock.MarshalizerStub{},
			UInt64ByteSliceConv: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilUint64ByteSliceConverter))
		require.Nil(t, pcf)
	})
	t.Run("nil Crypto should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Crypto = nil
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilCryptoComponentsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil BlockSignKeyGen should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Crypto = &testsMocks.CryptoComponentsStub{
			BlKeyGen: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilBlockSignKeyGen))
		require.Nil(t, pcf)
	})
	t.Run("nil State should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.State = nil
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilStateComponentsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil AccountsAdapter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.State = &factoryMocks.StateComponentsMock{
			Accounts: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilAccountsAdapter))
		require.Nil(t, pcf)
	})
	t.Run("nil Network should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Network = nil
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilNetworkComponentsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil NetworkMessenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Network = &testsMocks.NetworkComponentsStub{
			Messenger: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilMessenger))
		require.Nil(t, pcf)
	})
	t.Run("nil InputAntiFloodHandler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Network = &testsMocks.NetworkComponentsStub{
			Messenger:      &p2pmocks.MessengerStub{},
			InputAntiFlood: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilInputAntiFloodHandler))
		require.Nil(t, pcf)
	})
	t.Run("nil SystemSCConfig should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.SystemSCConfig = nil
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilSystemSCConfig))
		require.Nil(t, pcf)
	})
	t.Run("nil BootstrapComponents should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.BootstrapComponents = nil
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilBootstrapComponentsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil ShardCoordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.BootstrapComponents = &mainFactoryMocks.BootstrapComponentsStub{
			ShCoordinator: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilShardCoordinator))
		require.Nil(t, pcf)
	})
	t.Run("nil EpochBootstrapParams should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.BootstrapComponents = &mainFactoryMocks.BootstrapComponentsStub{
			ShCoordinator:   &testscommon.ShardsCoordinatorMock{},
			BootstrapParams: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilBootstrapParamsHandler))
		require.Nil(t, pcf)
	})
	t.Run("nil StatusComponents should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.StatusComponents = nil
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilStatusComponentsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil OutportHandler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.StatusComponents = &testsMocks.StatusComponentsStub{
			Outport: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilOutportHandler))
		require.Nil(t, pcf)
	})
	t.Run("nil HistoryRepo should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.HistoryRepo = nil
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilHistoryRepository))
		require.Nil(t, pcf)
	})
	t.Run("nil StatusCoreComponents should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.StatusCoreComponents = nil
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilStatusCoreComponents))
		require.Nil(t, pcf)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pcf, err := processComp.NewProcessComponentsFactory(createMockProcessComponentsFactoryArgs())
		require.NoError(t, err)
		require.NotNil(t, pcf)
	})
}

func TestProcessComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("CreateCurrentEpochProvider fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.EpochStartConfig.RoundsPerEpoch = 0
		args.PrefConfigs.Preferences.FullArchive = true
		testCreateWithArgs(t, args, "rounds per epoch")
	})
	t.Run("createNetworkShardingCollector fails due to invalid PublicKeyPeerId config should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.PublicKeyPeerId.Type = "invalid"
		testCreateWithArgs(t, args, "cache type")
	})
	t.Run("createNetworkShardingCollector fails due to invalid PublicKeyShardId config should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.PublicKeyShardId.Type = "invalid"
		testCreateWithArgs(t, args, "cache type")
	})
	t.Run("createNetworkShardingCollector fails due to invalid PeerIdShardId config should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.PeerIdShardId.Type = "invalid"
		testCreateWithArgs(t, args, "cache type")
	})
	t.Run("prepareNetworkShardingCollector fails due to SetPeerShardResolver failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		netwCompStub, ok := args.Network.(*testsMocks.NetworkComponentsStub)
		require.True(t, ok)
		netwCompStub.Messenger = &p2pmocks.MessengerStub{
			SetPeerShardResolverCalled: func(peerShardResolver p2p.PeerShardResolver) error {
				return expectedErr
			},
		}
		testCreateWithArgs(t, args, expectedErr.Error())
	})
	t.Run("prepareNetworkShardingCollector fails due to SetPeerValidatorMapper failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		netwCompStub, ok := args.Network.(*testsMocks.NetworkComponentsStub)
		require.True(t, ok)
		netwCompStub.InputAntiFlood = &testsMocks.P2PAntifloodHandlerStub{
			SetPeerValidatorMapperCalled: func(validatorMapper process.PeerValidatorMapper) error {
				return expectedErr
			},
		}
		testCreateWithArgs(t, args, expectedErr.Error())
	})
	t.Run("newStorageRequester fails due to NewStorageServiceFactory failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.ImportDBConfig.IsImportDBMode = true
		args.Config.StoragePruning.NumActivePersisters = 0
		testCreateWithArgs(t, args, "active persisters")
	})
	t.Run("newResolverContainerFactory fails due to NewPeerAuthenticationPayloadValidator failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.HeartbeatV2.HeartbeatExpiryTimespanInSec = 0
		testCreateWithArgs(t, args, "expiry timespan")
	})
	t.Run("generateGenesisHeadersAndApplyInitialBalances fails due to invalid GenesisNodePrice should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.LogsAndEvents.SaveInStorageEnabled = false // coverage
		args.Config.DbLookupExtensions.Enabled = true          // coverage
		args.SystemSCConfig.StakingSystemSCConfig.GenesisNodePrice = "invalid"
		testCreateWithArgs(t, args, "invalid genesis node price")
	})
	t.Run("newValidatorStatisticsProcessor fails due to nil genesis header should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.ImportDBConfig.IsImportDBMode = true // coverage
		dataCompStub, ok := args.Data.(*testsMocks.DataComponentsStub)
		require.True(t, ok)
		blockChainStub, ok := dataCompStub.BlockChain.(*testscommon.ChainHandlerStub)
		require.True(t, ok)
		blockChainStub.GetGenesisHeaderCalled = func() coreData.HeaderHandler {
			return nil
		}
		testCreateWithArgs(t, args, errorsMx.ErrGenesisBlockNotInitialized.Error())
	})
	t.Run("indexGenesisBlocks fails due to GenerateInitialTransactions failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.AccountsParser = &mock.AccountsParserStub{
			GenerateInitialTransactionsCalled: func(shardCoordinator sharding.Coordinator, initialIndexingData map[uint32]*genesis.IndexingData) ([]*dataBlock.MiniBlock, map[uint32]*outportCore.TransactionPool, error) {
				return nil, nil, expectedErr
			},
		}
		testCreateWithArgs(t, args, expectedErr.Error())
	})
	t.Run("NewMiniBlocksPoolsCleaner fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.PoolsCleanersConfig.MaxRoundsToKeepUnprocessedMiniBlocks = 0
		testCreateWithArgs(t, args, "MaxRoundsToKeepUnprocessedData")
	})
	t.Run("NewTxsPoolsCleaner fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.PoolsCleanersConfig.MaxRoundsToKeepUnprocessedTransactions = 0
		testCreateWithArgs(t, args, "MaxRoundsToKeepUnprocessedData")
	})
	t.Run("createHardforkTrigger fails due to Decode failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.Hardfork.PublicKeyToListenFrom = "invalid key"
		testCreateWithArgs(t, args, "PublicKeyToListenFrom")
	})
	t.Run("NewCache fails for vmOutput should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.VMOutputCacher.Type = "invalid"
		testCreateWithArgs(t, args, "cache type")
	})
	t.Run("newShardBlockProcessor: attachProcessDebugger fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.Debug.Process.Enabled = true
		args.Config.Debug.Process.PollingTimeInSeconds = 0
		testCreateWithArgs(t, args, "PollingTimeInSeconds")
	})
	t.Run("nodesSetupChecker.Check fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		coreCompStub := factoryMocks.NewCoreComponentsHolderStubFromRealComponent(args.CoreData)
		coreCompStub.GenesisNodesSetupCalled = func() sharding.GenesisNodesSetupHandler {
			return &nodesSetupMock.NodesSetupStub{
				AllInitialNodesCalled: func() []nodesCoordinator.GenesisNodeInfoHandler {
					return []nodesCoordinator.GenesisNodeInfoHandler{
						&genesisMocks.GenesisNodeInfoHandlerMock{
							PubKeyBytesValue: []byte("no stake"),
						},
					}
				},
				GetShardConsensusGroupSizeCalled: func() uint32 {
					return 2
				},
				GetMetaConsensusGroupSizeCalled: func() uint32 {
					return 2
				},
			}
		}
		args.CoreData = coreCompStub
		testCreateWithArgs(t, args, "no one staked")
	})
	t.Run("should work with indexAndReturnGenesisAccounts failing due to RootHash failure", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		statusCompStub, ok := args.StatusComponents.(*testsMocks.StatusComponentsStub)
		require.True(t, ok)
		statusCompStub.Outport = &outport.OutportStub{
			HasDriversCalled: func() bool {
				return true
			},
		}
		stateCompMock := factoryMocks.NewStateComponentsMockFromRealComponent(args.State)
		realAccounts := stateCompMock.AccountsAdapter()
		stateCompMock.Accounts = &testState.AccountsStub{
			GetAllLeavesCalled: realAccounts.GetAllLeaves,
			RootHashCalled: func() ([]byte, error) {
				return nil, expectedErr
			},
			CommitCalled: realAccounts.Commit,
		}
		args.State = stateCompMock

		pcf, _ := processComp.NewProcessComponentsFactory(args)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.Nil(t, err)
		require.NotNil(t, instance)

		err = instance.Close()
		require.NoError(t, err)
		_ = args.State.Close()
	})
	t.Run("should work with indexAndReturnGenesisAccounts failing due to GetAllLeaves failure", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		statusCompStub, ok := args.StatusComponents.(*testsMocks.StatusComponentsStub)
		require.True(t, ok)
		statusCompStub.Outport = &outport.OutportStub{
			HasDriversCalled: func() bool {
				return true
			},
		}
		stateCompMock := factoryMocks.NewStateComponentsMockFromRealComponent(args.State)
		realAccounts := stateCompMock.AccountsAdapter()
		stateCompMock.Accounts = &testState.AccountsStub{
			GetAllLeavesCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeavesParser common.TrieLeafParser) error {
				close(leavesChannels.LeavesChan)
				leavesChannels.ErrChan.Close()
				return expectedErr
			},
			RootHashCalled: realAccounts.RootHash,
			CommitCalled:   realAccounts.Commit,
		}
		args.State = stateCompMock

		pcf, _ := processComp.NewProcessComponentsFactory(args)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.Nil(t, err)
		require.NotNil(t, instance)

		err = instance.Close()
		require.NoError(t, err)
		_ = args.State.Close()
	})
	t.Run("should work with indexAndReturnGenesisAccounts failing due to Unmarshal failure", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		statusCompStub, ok := args.StatusComponents.(*testsMocks.StatusComponentsStub)
		require.True(t, ok)
		statusCompStub.Outport = &outport.OutportStub{
			HasDriversCalled: func() bool {
				return true
			},
		}
		stateCompMock := factoryMocks.NewStateComponentsMockFromRealComponent(args.State)
		realAccounts := stateCompMock.AccountsAdapter()
		stateCompMock.Accounts = &testState.AccountsStub{
			GetAllLeavesCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeavesParser common.TrieLeafParser) error {
				addrOk, _ := addrPubKeyConv.Decode("erd17c4fs6mz2aa2hcvva2jfxdsrdknu4220496jmswer9njznt22eds0rxlr4")
				addrNOK, _ := addrPubKeyConv.Decode("erd1ulhw20j7jvgfgak5p05kv667k5k9f320sgef5ayxkt9784ql0zssrzyhjp")
				leavesChannels.LeavesChan <- keyValStorage.NewKeyValStorage(addrOk, []byte("value")) // coverage
				leavesChannels.LeavesChan <- keyValStorage.NewKeyValStorage(addrNOK, []byte("value"))
				close(leavesChannels.LeavesChan)
				leavesChannels.ErrChan.Close()
				return nil
			},
			RootHashCalled: realAccounts.RootHash,
			CommitCalled:   realAccounts.Commit,
		}
		args.State = stateCompMock

		coreCompStub := factoryMocks.NewCoreComponentsHolderStubFromRealComponent(args.CoreData)
		cnt := 0
		coreCompStub.InternalMarshalizerCalled = func() marshal.Marshalizer {
			return &marshallerMock.MarshalizerStub{
				UnmarshalCalled: func(obj interface{}, buff []byte) error {
					cnt++
					if cnt == 1 {
						return nil // coverage, key_ok
					}
					return expectedErr
				},
			}
		}
		args.CoreData = coreCompStub
		pcf, _ := processComp.NewProcessComponentsFactory(args)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.Nil(t, err)
		require.NotNil(t, instance)

		err = instance.Close()
		require.NoError(t, err)
		_ = args.State.Close()
	})
	t.Run("should work with indexAndReturnGenesisAccounts failing due to error on GetAllLeaves", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		statusCompStub, ok := args.StatusComponents.(*testsMocks.StatusComponentsStub)
		require.True(t, ok)
		statusCompStub.Outport = &outport.OutportStub{
			HasDriversCalled: func() bool {
				return true
			},
		}
		realStateComp := args.State
		args.State = &factoryMocks.StateComponentsMock{
			Accounts: &testState.AccountsStub{
				GetAllLeavesCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeavesParser common.TrieLeafParser) error {
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.WriteInChanNonBlocking(expectedErr)
					leavesChannels.ErrChan.Close()
					return nil
				},
				CommitCalled:   realStateComp.AccountsAdapter().Commit,
				RootHashCalled: realStateComp.AccountsAdapter().RootHash,
			},
			PeersAcc:             realStateComp.PeerAccounts(),
			Tries:                realStateComp.TriesContainer(),
			AccountsAPI:          realStateComp.AccountsAdapterAPI(),
			StorageManagers:      realStateComp.TrieStorageManagers(),
			MissingNodesNotifier: realStateComp.MissingTrieNodesNotifier(),
		}

		pcf, _ := processComp.NewProcessComponentsFactory(args)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.Nil(t, err)
		require.NotNil(t, instance)

		err = instance.Close()
		require.NoError(t, err)
		_ = args.State.Close()
	})
	t.Run("should work with indexAndReturnGenesisAccounts failing due to error on Encode", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		statusCompStub, ok := args.StatusComponents.(*testsMocks.StatusComponentsStub)
		require.True(t, ok)
		statusCompStub.Outport = &outport.OutportStub{
			HasDriversCalled: func() bool {
				return true
			},
		}
		realStateComp := args.State
		args.State = &factoryMocks.StateComponentsMock{
			Accounts: &testState.AccountsStub{
				GetAllLeavesCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeavesParser common.TrieLeafParser) error {
					leavesChannels.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("invalid addr"), []byte("value"))
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
					return nil
				},
				CommitCalled:   realStateComp.AccountsAdapter().Commit,
				RootHashCalled: realStateComp.AccountsAdapter().RootHash,
			},
			PeersAcc:             realStateComp.PeerAccounts(),
			Tries:                realStateComp.TriesContainer(),
			AccountsAPI:          realStateComp.AccountsAdapterAPI(),
			StorageManagers:      realStateComp.TrieStorageManagers(),
			MissingNodesNotifier: realStateComp.MissingTrieNodesNotifier(),
		}
		coreCompStub := factoryMocks.NewCoreComponentsHolderStubFromRealComponent(args.CoreData)
		coreCompStub.InternalMarshalizerCalled = func() marshal.Marshalizer {
			return &marshallerMock.MarshalizerStub{
				UnmarshalCalled: func(obj interface{}, buff []byte) error {
					return nil
				},
			}
		}
		args.CoreData = coreCompStub

		pcf, _ := processComp.NewProcessComponentsFactory(args)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.Nil(t, err)
		require.NotNil(t, instance)

		err = instance.Close()
		require.NoError(t, err)
		_ = args.State.Close()
	})
	t.Run("should work - shard", func(t *testing.T) {
		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		processArgs := components.GetProcessComponentsFactoryArgs(shardCoordinator)
		pcf, _ := processComp.NewProcessComponentsFactory(processArgs)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.NoError(t, err)
		require.NotNil(t, instance)

		err = instance.Close()
		require.NoError(t, err)
		_ = processArgs.State.Close()
	})
	t.Run("should work - meta", func(t *testing.T) {
		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		shardCoordinator.CurrentShard = common.MetachainShardId
		processArgs := components.GetProcessComponentsFactoryArgs(shardCoordinator)

		shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
			protocolSustainabilityAddr, err := processArgs.CoreData.AddressPubKeyConverter().Decode(testingProtocolSustainabilityAddress)
			require.NoError(t, err)
			if bytes.Equal(protocolSustainabilityAddr, address) {
				return 0
			}
			return shardCoordinator.CurrentShard
		}
		fundGenesisWallets(t, processArgs)

		pcf, _ := processComp.NewProcessComponentsFactory(processArgs)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.NoError(t, err)
		require.NotNil(t, instance)

		err = instance.Close()
		require.NoError(t, err)
		_ = processArgs.State.Close()
	})
}

func fundGenesisWallets(t *testing.T, args processComp.ProcessComponentsFactoryArgs) {
	accounts := args.State.AccountsAdapter()
	initialNodes := args.CoreData.GenesisNodesSetup().AllInitialNodes()
	nodePrice, ok := big.NewInt(0).SetString(args.SystemSCConfig.StakingSystemSCConfig.GenesisNodePrice, 10)
	require.True(t, ok)
	for _, node := range initialNodes {
		account, err := accounts.LoadAccount(node.AddressBytes())
		require.NoError(t, err)

		userAccount := account.(state.UserAccountHandler)
		err = userAccount.AddToBalance(nodePrice)
		require.NoError(t, err)

		require.NoError(t, accounts.SaveAccount(userAccount))
		_, err = accounts.Commit()
		require.NoError(t, err)
	}
}

func testCreateWithArgs(t *testing.T, args processComp.ProcessComponentsFactoryArgs, expectedErrSubstr string) {
	pcf, _ := processComp.NewProcessComponentsFactory(args)
	require.NotNil(t, pcf)

	instance, err := pcf.Create()
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), expectedErrSubstr))
	require.Nil(t, instance)

	_ = args.State.Close()
}
