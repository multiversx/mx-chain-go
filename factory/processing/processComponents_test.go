package processing_test

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"strings"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	outportCore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/factory"
	"github.com/multiversx/mx-chain-go/config"
	retriever "github.com/multiversx/mx-chain-go/dataRetriever"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/mock"
	processComp "github.com/multiversx/mx-chain-go/factory/processing"
	"github.com/multiversx/mx-chain-go/genesis"
	genesisMocks "github.com/multiversx/mx-chain-go/genesis/mock"
	testsMocks "github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	mxState "github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	factoryMocks "github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/guardianMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/testscommon/trie"
	trieFactory "github.com/multiversx/mx-chain-go/trie/factory"
	updateMocks "github.com/multiversx/mx-chain-go/update/mock"
	"github.com/stretchr/testify/require"
)

func createMockProcessComponentsFactoryArgs() processComp.ProcessComponentsFactoryArgs {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../cmd/node/config/gasSchedules/gasScheduleV1.toml")
	addrPubKeyConv, _ := factory.NewPubkeyConverter(config.PubkeyConfig{
		Length:          32,
		Type:            "bech32",
		SignatureLength: 0,
	})
	valPubKeyConv, _ := factory.NewPubkeyConverter(config.PubkeyConfig{
		Length:          96,
		Type:            "hex",
		SignatureLength: 48,
	})
	return processComp.ProcessComponentsFactoryArgs{
		Config:         testscommon.GetGeneralConfig(),
		EpochConfig:    config.EpochConfig{},
		PrefConfigs:    config.PreferencesConfig{},
		ImportDBConfig: config.ImportDbConfig{},
		AccountsParser: &mock.AccountsParserStub{
			GenerateInitialTransactionsCalled: func(shardCoordinator sharding.Coordinator, initialIndexingData map[uint32]*genesis.IndexingData) ([]*dataBlock.MiniBlock, map[uint32]*outportCore.Pool, error) {
				return []*dataBlock.MiniBlock{
						{},
					},
					map[uint32]*outportCore.Pool{
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
					MinQuorum:        "50",
					MinPassThreshold: "50",
					MinVetoThreshold: "50",
				},
				FirstWhitelistedAddress: "erd1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awqgqky80",
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "2500000000000000000000",
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
		},
		Version:            "v1.0.0",
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
			IntMarsh:            &testscommon.MarshalizerStub{},
			TxMarsh:             &testscommon.MarshalizerStub{},
			UInt64ByteSliceConv: &testsMocks.Uint64ByteSliceConverterMock{},
			AddrPubKeyConv:      addrPubKeyConv,
			ValPubKeyConv:       valPubKeyConv,
			NodesConfig: &testscommon.NodesSetupStub{
				GetShardConsensusGroupSizeCalled: func() uint32 {
					return 2
				},
				GetMetaConsensusGroupSizeCalled: func() uint32 {
					return 2
				},
			},
			EpochChangeNotifier:        &epochNotifier.EpochNotifierStub{},
			EconomicsHandler:           &economicsmocks.EconomicsHandlerStub{},
			Hash:                       &testscommon.HasherStub{},
			TxVersionCheckHandler:      &testscommon.TxVersionCheckerStub{},
			RatingHandler:              &testscommon.RaterMock{},
			EnableEpochsHandlerField:   &testscommon.EnableEpochsHandlerStub{},
			EnableRoundsHandlerField:   &testscommon.EnableRoundsHandlerStub{},
			EpochNotifierWithConfirm:   &updateMocks.EpochStartNotifierStub{},
			RoundHandlerField:          &testscommon.RoundHandlerMock{},
			ChanStopProcess:            make(chan endProcess.ArgEndProcess, 1),
			TxSignHasherField:          &testscommon.HasherStub{},
			HardforkTriggerPubKeyField: []byte("hardfork pub key"),
			WasmVMChangeLockerInternal: &sync.RWMutex{},
			NodeTypeProviderField:      &nodeTypeProviderMock.NodeTypeProviderStub{},
		},
		Crypto: &testsMocks.CryptoComponentsStub{
			BlKeyGen: &cryptoMocks.KeyGenStub{},
			BlockSig: &cryptoMocks.SingleSignerStub{},
			MultiSigContainer: &cryptoMocks.MultiSignerContainerMock{
				MultiSigner: &cryptoMocks.MultisignerMock{},
			},
			PrivKey:         &cryptoMocks.PrivateKeyStub{},
			PubKey:          &cryptoMocks.PublicKeyStub{},
			PubKeyString:    "pub key string",
			PubKeyBytes:     []byte("pub key bytes"),
			TxKeyGen:        &cryptoMocks.KeyGenStub{},
			TxSig:           &cryptoMocks.SingleSignerStub{},
			PeerSignHandler: &cryptoMocks.PeerSignatureHandlerStub{},
		},
		State: &testscommon.StateComponentsMock{
			Accounts: &state.AccountsStub{
				CommitCalled: func() ([]byte, error) {
					return []byte(""), nil
				},
				RootHashCalled: func() ([]byte, error) {
					return []byte("root hash"), nil
				},
			},
			PeersAcc: &state.AccountsStub{
				CommitCalled: func() ([]byte, error) {
					return []byte("hash"), nil
				},
				RootHashCalled: func() ([]byte, error) {
					return []byte("root hash"), nil
				},
			},
			Tries: &trie.TriesHolderStub{
				GetCalled: func(bytes []byte) common.Trie {
					return &trie.TrieStub{}
				},
			},
			AccountsAPI: &state.AccountsStub{},
			StorageManagers: map[string]common.StorageManager{
				trieFactory.UserAccountTrie: &testscommon.StorageManagerStub{},
				trieFactory.PeerAccountTrie: &testscommon.StorageManagerStub{},
			},
		},
		Network: &testsMocks.NetworkComponentsStub{
			Messenger:               &p2pmocks.MessengerStub{},
			InputAntiFlood:          &testsMocks.P2PAntifloodHandlerStub{},
			OutputAntiFlood:         &testsMocks.P2PAntifloodHandlerStub{},
			PreferredPeersHolder:    &p2pmocks.PeersHolderStub{},
			PeersRatingHandlerField: &p2pmocks.PeersRatingHandlerStub{},
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
			AppStatusHandlerField: &statusHandler.AppStatusHandlerStub{},
		},
	}
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
	t.Run("nil Blockchain should error", func(t *testing.T) {
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
			NodesConfig:      &testscommon.NodesSetupStub{},
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
			NodesConfig:         &testscommon.NodesSetupStub{},
			AddrPubKeyConv:      &mock.PubkeyConverterStub{},
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
			NodesConfig:         &testscommon.NodesSetupStub{},
			AddrPubKeyConv:      &mock.PubkeyConverterStub{},
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
			NodesConfig:         &testscommon.NodesSetupStub{},
			AddrPubKeyConv:      &mock.PubkeyConverterStub{},
			EpochChangeNotifier: &epochNotifier.EpochNotifierStub{},
			ValPubKeyConv:       &mock.PubkeyConverterStub{},
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
			NodesConfig:         &testscommon.NodesSetupStub{},
			AddrPubKeyConv:      &mock.PubkeyConverterStub{},
			EpochChangeNotifier: &epochNotifier.EpochNotifierStub{},
			ValPubKeyConv:       &mock.PubkeyConverterStub{},
			IntMarsh:            &testscommon.MarshalizerStub{},
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
		args.State = &testscommon.StateComponentsMock{
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
		args.PrefConfigs.FullArchive = true
		testCreateWithArgs(t, args, "rounds per epoch")
	})
	t.Run("NewFallbackHeaderValidator fails should error", testWithNilMarshaller(1, "Marshalizer"))
	t.Run("NewHeaderSigVerifier fails should error", testWithNilMarshaller(2, "Marshalizer"))
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
	t.Run("newStorageResolver fails due to NewStorageServiceFactory failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.ImportDBConfig.IsImportDBMode = true
		args.Config.StoragePruning.NumActivePersisters = 0
		testCreateWithArgs(t, args, "active persisters")
	})
	t.Run("newStorageResolver fails due to CreateForMeta failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.ImportDBConfig.IsImportDBMode = true
		args.Config.ShardHdrNonceHashStorage.Cache.Type = "invalid"
		bootstrapCompStub, ok := args.BootstrapComponents.(*mainFactoryMocks.BootstrapComponentsStub)
		require.True(t, ok)
		bootstrapCompStub.ShCoordinator = &testscommon.ShardsCoordinatorMock{
			NoShards:     2,
			CurrentShard: common.MetachainShardId,
		}
		testCreateWithArgs(t, args, "ShardHdrNonceHashStorage")
	})
	t.Run("newStorageResolver fails due to CreateForMeta failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.ImportDBConfig.IsImportDBMode = true
		args.Config.ShardHdrNonceHashStorage.Cache.Type = "invalid"
		testCreateWithArgs(t, args, "ShardHdrNonceHashStorage")
	})
	t.Run("newResolverContainerFactory fails due to NewPeerAuthenticationPayloadValidator failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.HeartbeatV2.HeartbeatExpiryTimespanInSec = 0
		testCreateWithArgs(t, args, "expiry timespan")
	})
	t.Run("newResolverContainerFactory fails due to invalid shard should error",
		testWithInvalidShard(0, "could not create interceptor and resolver container factory"))
	t.Run("newMetaResolverContainerFactory fails due to NewSimpleDataPacker failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		bootstrapCompStub, ok := args.BootstrapComponents.(*mainFactoryMocks.BootstrapComponentsStub)
		require.True(t, ok)
		bootstrapCompStub.ShCoordinator = &testscommon.ShardsCoordinatorMock{
			NoShards:     2,
			CurrentShard: common.MetachainShardId,
		}
		coreCompStub := factoryMocks.NewCoreComponentsHolderStubFromRealComponent(args.CoreData)
		cnt := 0
		coreCompStub.InternalMarshalizerCalled = func() marshal.Marshalizer {
			cnt++
			if cnt > 3 {
				return nil
			}
			return &testscommon.MarshalizerStub{}
		}
		args.CoreData = coreCompStub
		testCreateWithArgs(t, args, "marshalizer")
	})
	t.Run("newShardResolverContainerFactory fails due to NewSimpleDataPacker failure should error", testWithNilMarshaller(3, "marshalizer"))
	t.Run("NewResolversFinder fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.ImportDBConfig.IsImportDBMode = true // coverage
		bootstrapCompStub, ok := args.BootstrapComponents.(*mainFactoryMocks.BootstrapComponentsStub)
		require.True(t, ok)
		cnt := 0
		bootstrapCompStub.ShardCoordinatorCalled = func() sharding.Coordinator {
			cnt++
			if cnt > 5 {
				return nil
			}
			return &testscommon.ShardsCoordinatorMock{
				NoShards:     2,
				CurrentShard: common.MetachainShardId, // coverage
			}
		}
		testCreateWithArgs(t, args, "shard coordinator")
	})
	t.Run("GetStorer TxLogsUnit fails should error", testWithMissingStorer(0, retriever.TxLogsUnit))
	t.Run("NewResolversFinder fails should error", testWithNilMarshaller(5, "Marshalizer"))
	t.Run("generateGenesisHeadersAndApplyInitialBalances fails due to invalid GenesisNodePrice should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.LogsAndEvents.SaveInStorageEnabled = false // coverage
		args.Config.DbLookupExtensions.Enabled = true          // coverage
		args.SystemSCConfig.StakingSystemSCConfig.GenesisNodePrice = "invalid"
		testCreateWithArgs(t, args, "invalid genesis node price")
	})
	t.Run("generateGenesisHeadersAndApplyInitialBalances fails due to NewGenesisBlockCreator failure should error",
		testWithNilMarshaller(6, "Marshalizer"))
	t.Run("setGenesisHeader fails due to invalid shard should error",
		testWithInvalidShard(8, "genesis block does not exist"))
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
	t.Run("indexGenesisBlocks fails due to CalculateHash failure should error", testWithNilMarshaller(41, "marshalizer"))
	t.Run("indexGenesisBlocks fails due to GenerateInitialTransactions failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.AccountsParser = &mock.AccountsParserStub{
			GenerateInitialTransactionsCalled: func(shardCoordinator sharding.Coordinator, initialIndexingData map[uint32]*genesis.IndexingData) ([]*dataBlock.MiniBlock, map[uint32]*outportCore.Pool, error) {
				return nil, nil, expectedErr
			},
		}
		testCreateWithArgs(t, args, expectedErr.Error())
	})
	t.Run("NewValidatorsProvider fails should error", testWithNilPubKeyConv(2, "pubkey converter"))
	t.Run("newEpochStartTrigger fails due to invalid shard should error",
		testWithInvalidShard(16, "error creating new start of epoch trigger because of invalid shard id"))
	t.Run("newEpochStartTrigger fails due to NewHeaderValidator failure should error", testWithNilMarshaller(46, "Marshalizer"))
	t.Run("newEpochStartTrigger fails due to NewPeerMiniBlockSyncer failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		dataCompStub, ok := args.Data.(*testsMocks.DataComponentsStub)
		require.True(t, ok)
		dataPool := dataCompStub.DataPool
		cnt := 0
		dataCompStub.DataPool = &dataRetriever.PoolsHolderStub{
			HeadersCalled:      dataPool.Headers,
			TransactionsCalled: dataPool.Transactions,
			MiniBlocksCalled:   dataPool.MiniBlocks,
			CurrBlockTxsCalled: dataPool.CurrentBlockTxs,
			TrieNodesCalled:    dataPool.TrieNodes,
			ValidatorsInfoCalled: func() retriever.ShardedDataCacherNotifier {
				cnt++
				if cnt > 3 {
					return nil
				}
				return dataPool.ValidatorsInfo()
			},
			CloseCalled: nil,
		}
		testCreateWithArgs(t, args, "validators info pool")
	})
	t.Run("newEpochStartTrigger fails due to NewPeerMiniBlockSyncer failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		bootstrapCompStub, ok := args.BootstrapComponents.(*mainFactoryMocks.BootstrapComponentsStub)
		require.True(t, ok)
		cntShardC := 0
		bootstrapCompStub.ShardCoordinatorCalled = func() sharding.Coordinator {
			cntShardC++
			shardC := testscommon.NewMultiShardsCoordinatorMock(2)
			if cntShardC > 16 {
				shardC.CurrentShard = common.MetachainShardId
			}
			return shardC
		}
		dataCompStub, ok := args.Data.(*testsMocks.DataComponentsStub)
		require.True(t, ok)
		blockChainStub, ok := dataCompStub.BlockChain.(*testscommon.ChainHandlerStub)
		require.True(t, ok)
		cnt := 0
		blockChainStub.GetGenesisHeaderCalled = func() coreData.HeaderHandler {
			cnt++
			if cnt > 1 {
				return nil
			}
			return &testscommon.HeaderHandlerStub{}
		}
		testCreateWithArgs(t, args, errorsMx.ErrGenesisBlockNotInitialized.Error())
	})
	t.Run("newEpochStartTrigger fails due to invalid shard should error",
		testWithInvalidShard(17, "genesis block does not exist"))
	t.Run("NewHeaderValidator fails should error", testWithNilMarshaller(48, "marshalizer"))
	t.Run("prepareGenesisBlock fails due to CalculateHash failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		dataCompStub, ok := args.Data.(*testsMocks.DataComponentsStub)
		require.True(t, ok)
		blockChainStub, ok := dataCompStub.BlockChain.(*testscommon.ChainHandlerStub)
		require.True(t, ok)
		cnt := 0
		blockChainStub.SetGenesisHeaderCalled = func(handler coreData.HeaderHandler) error {
			cnt++
			if cnt > 1 {
				return expectedErr
			}
			return nil
		}
		testCreateWithArgs(t, args, expectedErr.Error())
	})
	t.Run("saveGenesisHeaderToStorage fails due to Marshal failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		coreCompStub := factoryMocks.NewCoreComponentsHolderStubFromRealComponent(args.CoreData)
		cnt := 0
		coreCompStub.InternalMarshalizerCalled = func() marshal.Marshalizer {
			return &testscommon.MarshalizerStub{
				MarshalCalled: func(obj interface{}) ([]byte, error) {
					cnt++
					if cnt > 38 {
						return nil, expectedErr
					}
					return []byte(""), nil
				},
			}
		}
		args.CoreData = coreCompStub
		testCreateWithArgs(t, args, expectedErr.Error())
	})
	t.Run("GetStorer TxLogsUnit fails should error", testWithMissingStorer(2, retriever.BootstrapUnit))
	t.Run("NewBootstrapStorer fails should error", testWithNilMarshaller(50, "Marshalizer"))
	t.Run("NewHeaderValidator fails should error", testWithNilMarshaller(51, "Marshalizer"))
	t.Run("newBlockTracker fails due to invalid shard should error",
		testWithInvalidShard(19, "could not create block tracker"))
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
	t.Run("NewMiniBlockTrack fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		bootstrapCompStub, ok := args.BootstrapComponents.(*mainFactoryMocks.BootstrapComponentsStub)
		require.True(t, ok)
		cnt := 0
		bootstrapCompStub.ShardCoordinatorCalled = func() sharding.Coordinator {
			cnt++
			if cnt > 22 {
				return nil
			}
			return testscommon.NewMultiShardsCoordinatorMock(2)
		}
		testCreateWithArgs(t, args, "shard coordinator")
	})
	t.Run("createHardforkTrigger fails due to Decode failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.Hardfork.PublicKeyToListenFrom = "invalid key"
		testCreateWithArgs(t, args, "PublicKeyToListenFrom")
	})
	t.Run("newInterceptorContainerFactory fails due to invalid shard should error",
		testWithInvalidShard(23, "could not create interceptor container factory"))
	t.Run("createExportFactoryHandler fails", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		bootstrapCompStub, ok := args.BootstrapComponents.(*mainFactoryMocks.BootstrapComponentsStub)
		require.True(t, ok)
		cnt := 0
		bootstrapCompStub.ShardCoordinatorCalled = func() sharding.Coordinator {
			cnt++
			if cnt > 25 {
				return nil
			}
			return testscommon.NewMultiShardsCoordinatorMock(2)
		}
		testCreateWithArgs(t, args, "shard coordinator")
	})
	t.Run("newForkDetector fails due to invalid shard should error",
		testWithInvalidShard(27, "could not create fork detector"))
	t.Run("NewCache fails for vmOutput should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.VMOutputCacher.Type = "invalid"
		testCreateWithArgs(t, args, "cache type")
	})
	t.Run("GetStorer TxLogsUnit fails should error", testWithMissingStorer(0, retriever.ScheduledSCRsUnit))
	t.Run("NewScheduledTxsExecution fails should error", testWithNilMarshaller(104, "Marshalizer"))
	t.Run("NewESDTDataStorage fails should error", testWithNilMarshaller(105, "Marshalizer"))
	t.Run("NewReceiptsRepository fails should error", testWithNilMarshaller(106, "marshalizer"))
	t.Run("newBlockProcessor fails due to invalid shard should error",
		testWithInvalidShard(31, "could not create block processor"))
	t.Run("newShardBlockProcessor: NewESDTTransferParser fails should error",
		testWithNilMarshaller(107, "marshaller"))
	t.Run("newShardBlockProcessor: createBuiltInFunctionContainer fails should error",
		testWithNilAddressPubKeyConv(46, "public key converter"))
	t.Run("newShardBlockProcessor: createVMFactoryShard fails due to NewBlockChainHookImpl failure should error",
		testWithNilAddressPubKeyConv(47, "pubkey converter"))
	t.Run("newShardBlockProcessor: NewIntermediateProcessorsContainerFactory fails should error",
		testWithNilMarshaller(110, "Marshalizer"))
	t.Run("newShardBlockProcessor: NewTxTypeHandler fails should error",
		testWithNilAddressPubKeyConv(49, "pubkey converter"))
	t.Run("newShardBlockProcessor: NewGasComputation fails should error",
		testWithNilEnableEpochsHandler(13, "enable epochs handler"))
	t.Run("newShardBlockProcessor: NewSmartContractProcessor fails should error",
		testWithNilAddressPubKeyConv(50, "pubkey converter"))
	t.Run("newShardBlockProcessor: NewRewardTxProcessor fails should error",
		testWithNilAddressPubKeyConv(51, "pubkey converter"))
	t.Run("newShardBlockProcessor: NewTxProcessor fails should error",
		testWithNilAddressPubKeyConv(52, "pubkey converter"))
	t.Run("newShardBlockProcessor: createShardTxSimulatorProcessor fails due to NewIntermediateProcessorsContainerFactory failure should error",
		testWithNilAddressPubKeyConv(53, "pubkey converter"))
	t.Run("newShardBlockProcessor: createShardTxSimulatorProcessor fails due to createBuiltInFunctionContainer failure should error",
		testWithNilAddressPubKeyConv(54, "public key converter"))
	t.Run("newShardBlockProcessor: createOutportDataProvider fails due to missing TransactionUnit should error",
		testWithMissingStorer(3, retriever.TransactionUnit))
	t.Run("newShardBlockProcessor: createOutportDataProvider fails due to missing MiniBlockUnit should error",
		testWithMissingStorer(4, retriever.MiniBlockUnit))
	t.Run("newShardBlockProcessor: NewShardProcessor fails should error",
		testWithNilEnableEpochsHandler(23, "enable epochs handler"))
	t.Run("newShardBlockProcessor: attachProcessDebugger fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.Debug.Process.Enabled = true
		args.Config.Debug.Process.PollingTimeInSeconds = 0
		testCreateWithArgs(t, args, "PollingTimeInSeconds")
	})
	t.Run("newShardBlockProcessor: NewBlockSizeComputation fails should error",
		testWithNilMarshaller(116, "Marshalizer"))
	t.Run("newShardBlockProcessor: NewPreProcessorsContainerFactory fails should error",
		testWithNilMarshaller(117, "Marshalizer"))
	t.Run("newShardBlockProcessor: NewPrintDoubleTransactionsDetector fails should error",
		testWithNilMarshaller(118, "Marshalizer"))
	t.Run("newShardBlockProcessor: NewTransactionCoordinator fails should error",
		testWithNilMarshaller(119, "Marshalizer"))

	t.Run("NewNodesSetupChecker fails should error", testWithNilPubKeyConv(5, "pubkey converter"))
	t.Run("nodesSetupChecker.Check fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		coreCompStub := factoryMocks.NewCoreComponentsHolderStubFromRealComponent(args.CoreData)
		coreCompStub.GenesisNodesSetupCalled = func() sharding.GenesisNodesSetupHandler {
			return &testscommon.NodesSetupStub{
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
	t.Run("NewNodeRedundancy fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		netwCompStub, ok := args.Network.(*testsMocks.NetworkComponentsStub)
		require.True(t, ok)
		cnt := 0
		netwCompStub.MessengerCalled = func() p2p.Messenger {
			cnt++
			if cnt > 7 {
				return nil
			}
			return &p2pmocks.MessengerStub{}
		}
		testCreateWithArgs(t, args, "messenger")
	})
	t.Run("NewReceiptsRepository fails should error", testWithNilMarshaller(123, "marshalizer"))
	t.Run("NewTxsSenderWithAccumulator fails should error", testWithNilMarshaller(124, "Marshalizer"))
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
		stateCompStub, ok := args.State.(*testscommon.StateComponentsMock)
		require.True(t, ok)
		accountsStub, ok := stateCompStub.Accounts.(*state.AccountsStub)
		require.True(t, ok)
		accountsStub.RootHashCalled = func() ([]byte, error) {
			return nil, expectedErr
		}
		pcf, _ := processComp.NewProcessComponentsFactory(args)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.Nil(t, err)
		require.NotNil(t, instance)
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
		stateCompStub, ok := args.State.(*testscommon.StateComponentsMock)
		require.True(t, ok)
		accountsStub, ok := stateCompStub.Accounts.(*state.AccountsStub)
		require.True(t, ok)
		accountsStub.GetAllLeavesCalled = func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte) error {
			close(leavesChannels.LeavesChan)
			leavesChannels.ErrChan.Close()
			return expectedErr
		}

		pcf, _ := processComp.NewProcessComponentsFactory(args)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.Nil(t, err)
		require.NotNil(t, instance)
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
		stateCompStub, ok := args.State.(*testscommon.StateComponentsMock)
		require.True(t, ok)
		accountsStub, ok := stateCompStub.Accounts.(*state.AccountsStub)
		require.True(t, ok)
		accountsStub.GetAllLeavesCalled = func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte) error {
			leavesChannels.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("key_ok"), []byte("value")) // coverage
			leavesChannels.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("key_invalid"), []byte("value"))
			close(leavesChannels.LeavesChan)
			leavesChannels.ErrChan.Close()
			return nil
		}
		coreCompStub := factoryMocks.NewCoreComponentsHolderStubFromRealComponent(args.CoreData)
		cnt := 0
		coreCompStub.InternalMarshalizerCalled = func() marshal.Marshalizer {
			return &testscommon.MarshalizerStub{
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
		stateCompStub, ok := args.State.(*testscommon.StateComponentsMock)
		require.True(t, ok)
		accountsStub, ok := stateCompStub.Accounts.(*state.AccountsStub)
		require.True(t, ok)
		accountsStub.GetAllLeavesCalled = func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte) error {
			close(leavesChannels.LeavesChan)
			leavesChannels.ErrChan.WriteInChanNonBlocking(expectedErr)
			leavesChannels.ErrChan.Close()
			return nil
		}
		pcf, _ := processComp.NewProcessComponentsFactory(args)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.Nil(t, err)
		require.NotNil(t, instance)
	})
	t.Run("should work - shard", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
		pcf, _ := processComp.NewProcessComponentsFactory(processArgs)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.NoError(t, err)
		require.NotNil(t, instance)

		err = instance.Close()
		require.NoError(t, err)
	})
	t.Run("should work - meta", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		shardCoordinator.CurrentShard = common.MetachainShardId
		processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)

		protocolSustainabilityAddress := "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp"
		shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
			protocolSustainabilityAddr, err := processArgs.CoreData.AddressPubKeyConverter().Decode(protocolSustainabilityAddress)
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

		userAccount := account.(mxState.UserAccountHandler)
		err = userAccount.AddToBalance(nodePrice)
		require.NoError(t, err)

		require.NoError(t, accounts.SaveAccount(userAccount))
		_, err = accounts.Commit()
		require.NoError(t, err)
	}
}

func testWithNilMarshaller(nilStep int, expectedErrSubstr string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		coreCompStub := factoryMocks.NewCoreComponentsHolderStubFromRealComponent(args.CoreData)
		step := 0
		coreCompStub.InternalMarshalizerCalled = func() marshal.Marshalizer {
			step++
			println(step)
			if step > nilStep {
				return nil
			}
			return &testscommon.MarshalizerStub{}
		}
		args.CoreData = coreCompStub
		testCreateWithArgs(t, args, expectedErrSubstr)
	}
}

func testWithNilPubKeyConv(nilStep int, expectedErrSubstr string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		coreCompStub := factoryMocks.NewCoreComponentsHolderStubFromRealComponent(args.CoreData)
		pubKeyConv := args.CoreData.ValidatorPubKeyConverter()
		step := 0
		coreCompStub.ValidatorPubKeyConverterCalled = func() core.PubkeyConverter {
			step++
			if step > nilStep {
				return nil
			}
			return pubKeyConv
		}
		args.CoreData = coreCompStub
		testCreateWithArgs(t, args, expectedErrSubstr)
	}
}

func testWithNilAddressPubKeyConv(nilStep int, expectedErrSubstr string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		coreCompStub := factoryMocks.NewCoreComponentsHolderStubFromRealComponent(args.CoreData)
		pubKeyConv := args.CoreData.AddressPubKeyConverter()
		step := 0
		coreCompStub.AddressPubKeyConverterCalled = func() core.PubkeyConverter {
			step++
			println(step)
			if step > nilStep {
				return nil
			}
			return pubKeyConv
		}
		args.CoreData = coreCompStub
		testCreateWithArgs(t, args, expectedErrSubstr)
	}
}

func testWithNilEnableEpochsHandler(nilStep int, expectedErrSubstr string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		coreCompStub := factoryMocks.NewCoreComponentsHolderStubFromRealComponent(args.CoreData)
		enableEpochsHandler := coreCompStub.EnableEpochsHandler()
		step := 0
		coreCompStub.EnableEpochsHandlerCalled = func() common.EnableEpochsHandler {
			step++
			println(step)
			if step > nilStep {
				return nil
			}
			return enableEpochsHandler
		}
		args.CoreData = coreCompStub
		testCreateWithArgs(t, args, expectedErrSubstr)
	}
}

func testWithMissingStorer(failStep int, missingUnitType retriever.UnitType) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockProcessComponentsFactoryArgs()
		dataCompStub, ok := args.Data.(*testsMocks.DataComponentsStub)
		require.True(t, ok)
		store := args.Data.StorageService()
		cnt := 0
		dataCompStub.Store = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType retriever.UnitType) (storage.Storer, error) {
				if unitType == missingUnitType {
					cnt++
					println(cnt)
					if cnt > failStep {
						return nil, expectedErr
					}
				}
				return store.GetStorer(unitType)
			},
		}
		testCreateWithArgs(t, args, expectedErr.Error())
	}
}

func testWithInvalidShard(failingStep int, expectedErrSubstr string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		bootstrapCompStub, ok := args.BootstrapComponents.(*mainFactoryMocks.BootstrapComponentsStub)
		require.True(t, ok)
		cnt := 0
		bootstrapCompStub.ShardCoordinatorCalled = func() sharding.Coordinator {
			cnt++
			if cnt > failingStep {
				return &testscommon.ShardsCoordinatorMock{
					NoShards:     2,
					CurrentShard: 3,
				}
			}
			return testscommon.NewMultiShardsCoordinatorMock(2)
		}
		testCreateWithArgs(t, args, expectedErrSubstr)
	}
}

func testCreateWithArgs(t *testing.T, args processComp.ProcessComponentsFactoryArgs, expectedErrSubstr string) {
	pcf, _ := processComp.NewProcessComponentsFactory(args)
	require.NotNil(t, pcf)

	instance, err := pcf.Create()
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), expectedErrSubstr))
	require.Nil(t, instance)
}
