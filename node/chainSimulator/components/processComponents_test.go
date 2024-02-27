package components

import (
	"math/big"
	"sync"
	"testing"

	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	"github.com/multiversx/mx-chain-core-go/marshal"
	commonFactory "github.com/multiversx/mx-chain-go/common/factory"
	disabledStatistics "github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	retriever "github.com/multiversx/mx-chain-go/dataRetriever"
	mockFactory "github.com/multiversx/mx-chain-go/factory/mock"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/sharding"
	chainStorage "github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks"
	"github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/guardianMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	updateMocks "github.com/multiversx/mx-chain-go/update/mock"
	"github.com/stretchr/testify/require"
)

const testingProtocolSustainabilityAddress = "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp"

var (
	addrPubKeyConv, _ = commonFactory.NewPubkeyConverter(config.PubkeyConfig{
		Length:          32,
		Type:            "bech32",
		SignatureLength: 0,
		Hrp:             "erd",
	})
	valPubKeyConv, _ = commonFactory.NewPubkeyConverter(config.PubkeyConfig{
		Length:          96,
		Type:            "hex",
		SignatureLength: 48,
	})
)

func createArgsProcessComponentsHolder() ArgsProcessComponentsHolder {
	nodesSetup, _ := sharding.NewNodesSetup("../../../integrationTests/factory/testdata/nodesSetup.json", addrPubKeyConv, valPubKeyConv, 3)

	args := ArgsProcessComponentsHolder{
		Config: testscommon.GetGeneralConfig(),
		EpochConfig: config.EpochConfig{
			GasSchedule: config.GasScheduleConfig{
				GasScheduleByEpochs: []config.GasScheduleByEpochs{
					{
						StartEpoch: 0,
						FileName:   "../../../cmd/node/config/gasSchedules/gasScheduleV7.toml",
					},
				},
			},
		},
		PrefsConfig:    config.Preferences{},
		ImportDBConfig: config.ImportDbConfig{},
		FlagsConfig: config.ContextFlagsConfig{
			Version: "v1.0.0",
		},
		NodesCoordinator: &shardingMocks.NodesCoordinatorStub{},
		SystemSCConfig: config.SystemSmartContractsConfig{
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
				NodeLimitPercentage:                  0.1,
				StakeLimitPercentage:                 1,
				UnBondPeriodInEpochs:                 10,
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
		DataComponents: &mock.DataComponentsStub{
			DataPool: dataRetriever.NewPoolsHolderMock(),
			BlockChain: &testscommon.ChainHandlerStub{
				GetGenesisHeaderHashCalled: func() []byte {
					return []byte("genesis hash")
				},
				GetGenesisHeaderCalled: func() coreData.HeaderHandler {
					return &testscommon.HeaderHandlerStub{}
				},
			},
			MbProvider: &mock.MiniBlocksProviderStub{},
			Store:      genericMocks.NewChainStorerMock(0),
		},
		CoreComponents: &mockFactory.CoreComponentsMock{
			IntMarsh:            &marshal.GogoProtoMarshalizer{},
			TxMarsh:             &marshal.JsonMarshalizer{},
			UInt64ByteSliceConv: &mock.Uint64ByteSliceConverterMock{},
			AddrPubKeyConv:      addrPubKeyConv,
			ValPubKeyConv:       valPubKeyConv,
			NodesConfig:         nodesSetup,
			EpochChangeNotifier: &epochNotifier.EpochNotifierStub{},
			EconomicsHandler: &economicsmocks.EconomicsHandlerStub{
				ProtocolSustainabilityAddressCalled: func() string {
					return testingProtocolSustainabilityAddress
				},
				GenesisTotalSupplyCalled: func() *big.Int {
					return big.NewInt(0).Mul(big.NewInt(1000000000000000000), big.NewInt(20000000))
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
		CryptoComponents: &mock.CryptoComponentsStub{
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
		NetworkComponents: &mock.NetworkComponentsStub{
			Messenger:                        &p2pmocks.MessengerStub{},
			FullArchiveNetworkMessengerField: &p2pmocks.MessengerStub{},
			InputAntiFlood:                   &mock.P2PAntifloodHandlerStub{},
			OutputAntiFlood:                  &mock.P2PAntifloodHandlerStub{},
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
		StatusComponents: &mock.StatusComponentsStub{
			Outport: &outport.OutportStub{},
		},
		StatusCoreComponents: &factory.StatusCoreComponentsStub{
			AppStatusHandlerField:  &statusHandler.AppStatusHandlerStub{},
			StateStatsHandlerField: disabledStatistics.NewStateStatistics(),
		},
		EconomicsConfig: config.EconomicsConfig{
			GlobalSettings: config.GlobalSettings{
				GenesisTotalSupply:          "20000000000000000000000000",
				MinimumInflation:            0,
				GenesisMintingSenderAddress: "erd17rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rcqqkhty3",
				YearSettings: []*config.YearSetting{
					{
						Year:             0,
						MaximumInflation: 0.01,
					},
				},
			},
		},
		ConfigurationPathsHolder: config.ConfigurationPathsHolder{
			Genesis:        "../../../integrationTests/factory/testdata/genesis.json",
			SmartContracts: "../../../integrationTests/factory/testdata/genesisSmartContracts.json",
			Nodes:          "../../../integrationTests/factory/testdata/genesis.json",
		},
	}

	args.StateComponents = components.GetStateComponents(args.CoreComponents, args.StatusCoreComponents)
	return args
}

func TestCreateProcessComponents(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		// TODO reinstate test after Wasm VM pointer fix
		if testing.Short() {
			t.Skip("cannot run with -race -short; requires Wasm VM fix")
		}

		t.Parallel()

		comp, err := CreateProcessComponents(createArgsProcessComponentsHolder())
		require.NoError(t, err)
		require.NotNil(t, comp)

		require.Nil(t, comp.Create())
		require.Nil(t, comp.Close())
	})
	t.Run("NewImportStartHandler failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsProcessComponentsHolder()
		args.FlagsConfig.Version = ""
		comp, err := CreateProcessComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("total supply conversion failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsProcessComponentsHolder()
		args.EconomicsConfig.GlobalSettings.GenesisTotalSupply = "invalid number"
		comp, err := CreateProcessComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("NewAccountsParser failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsProcessComponentsHolder()
		args.ConfigurationPathsHolder.Genesis = ""
		comp, err := CreateProcessComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("NewSmartContractsParser failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsProcessComponentsHolder()
		args.ConfigurationPathsHolder.SmartContracts = ""
		comp, err := CreateProcessComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("NewHistoryRepositoryFactory failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsProcessComponentsHolder()
		dataMock, ok := args.DataComponents.(*mock.DataComponentsStub)
		require.True(t, ok)
		dataMock.Store = nil
		comp, err := CreateProcessComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("historyRepositoryFactory.Create failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsProcessComponentsHolder()
		args.Config.DbLookupExtensions.Enabled = true
		dataMock, ok := args.DataComponents.(*mock.DataComponentsStub)
		require.True(t, ok)
		dataMock.Store = &storage.ChainStorerStub{
			GetStorerCalled: func(unitType retriever.UnitType) (chainStorage.Storer, error) {
				if unitType == retriever.ESDTSuppliesUnit {
					return nil, expectedErr
				}
				return &storage.StorerStub{}, nil
			},
		}
		comp, err := CreateProcessComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("NewGasScheduleNotifier failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsProcessComponentsHolder()
		args.EpochConfig.GasSchedule = config.GasScheduleConfig{}
		comp, err := CreateProcessComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("NewProcessComponentsFactory failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsProcessComponentsHolder()
		dataMock, ok := args.DataComponents.(*mock.DataComponentsStub)
		require.True(t, ok)
		dataMock.BlockChain = nil
		comp, err := CreateProcessComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("managedProcessComponents.Create failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsProcessComponentsHolder()
		args.NodesCoordinator = nil
		comp, err := CreateProcessComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
}

func TestProcessComponentsHolder_IsInterfaceNil(t *testing.T) {
	// TODO reinstate test after Wasm VM pointer fix
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Wasm VM fix")
	}

	t.Parallel()

	var comp *processComponentsHolder
	require.True(t, comp.IsInterfaceNil())

	comp, _ = CreateProcessComponents(createArgsProcessComponentsHolder())
	require.False(t, comp.IsInterfaceNil())
	require.Nil(t, comp.Close())
}

func TestProcessComponentsHolder_Getters(t *testing.T) {
	// TODO reinstate test after Wasm VM pointer fix
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Wasm VM fix")
	}

	t.Parallel()

	comp, err := CreateProcessComponents(createArgsProcessComponentsHolder())
	require.NoError(t, err)

	require.NotNil(t, comp.SentSignaturesTracker())
	require.NotNil(t, comp.NodesCoordinator())
	require.NotNil(t, comp.ShardCoordinator())
	require.NotNil(t, comp.InterceptorsContainer())
	require.NotNil(t, comp.FullArchiveInterceptorsContainer())
	require.NotNil(t, comp.ResolversContainer())
	require.NotNil(t, comp.RequestersFinder())
	require.NotNil(t, comp.RoundHandler())
	require.NotNil(t, comp.EpochStartTrigger())
	require.NotNil(t, comp.EpochStartNotifier())
	require.NotNil(t, comp.ForkDetector())
	require.NotNil(t, comp.BlockProcessor())
	require.NotNil(t, comp.BlackListHandler())
	require.NotNil(t, comp.BootStorer())
	require.NotNil(t, comp.HeaderSigVerifier())
	require.NotNil(t, comp.HeaderIntegrityVerifier())
	require.NotNil(t, comp.ValidatorsStatistics())
	require.NotNil(t, comp.ValidatorsProvider())
	require.NotNil(t, comp.BlockTracker())
	require.NotNil(t, comp.PendingMiniBlocksHandler())
	require.NotNil(t, comp.RequestHandler())
	require.NotNil(t, comp.TxLogsProcessor())
	require.NotNil(t, comp.HeaderConstructionValidator())
	require.NotNil(t, comp.PeerShardMapper())
	require.NotNil(t, comp.FullArchivePeerShardMapper())
	require.NotNil(t, comp.FallbackHeaderValidator())
	require.NotNil(t, comp.APITransactionEvaluator())
	require.NotNil(t, comp.WhiteListHandler())
	require.NotNil(t, comp.WhiteListerVerifiedTxs())
	require.NotNil(t, comp.HistoryRepository())
	require.NotNil(t, comp.ImportStartHandler())
	require.NotNil(t, comp.RequestedItemsHandler())
	require.NotNil(t, comp.NodeRedundancyHandler())
	require.NotNil(t, comp.CurrentEpochProvider())
	require.NotNil(t, comp.ScheduledTxsExecutionHandler())
	require.NotNil(t, comp.TxsSenderHandler())
	require.NotNil(t, comp.HardforkTrigger())
	require.NotNil(t, comp.ProcessedMiniBlocksTracker())
	require.NotNil(t, comp.ESDTDataStorageHandlerForAPI())
	require.NotNil(t, comp.AccountsParser())
	require.NotNil(t, comp.ReceiptsRepository())
	require.Nil(t, comp.CheckSubcomponents())
	require.Empty(t, comp.String())

	require.Nil(t, comp.Close())
}
