package integrationTests

import (
	"sync"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/config"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/forking"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	p2pRating "github.com/ElrondNetwork/elrond-go/p2p/rating"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/transactionLog"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
)

// NewTestProcessorNodeWithStateCheckpointModulus creates a new testNodeProcessor with custom state checkpoint modulus
func NewTestProcessorNodeWithStateCheckpointModulus(
	maxShards uint32,
	nodeShardId uint32,
	txSignPrivKeyShardId uint32,
	stateCheckpointModulus uint,
) *TestProcessorNode {

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	kg := &mock.KeyGenMock{}
	sk, pk := kg.GeneratePair()

	pkBytes := []byte("afafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafaf")
	address := []byte("afafafafafafafafafafafafafafafaf")

	nodesSetup := &mock.NodesSetupStub{
		InitialNodesInfoCalled: func() (m map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, m2 map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			oneMap := make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
			oneMap[0] = append(oneMap[0], mock.NewNodeInfo(address, pkBytes, 0, InitialRating))
			return oneMap, nil
		},
		InitialNodesInfoForShardCalled: func(shardId uint32) (handlers []nodesCoordinator.GenesisNodeInfoHandler, handlers2 []nodesCoordinator.GenesisNodeInfoHandler, err error) {
			list := make([]nodesCoordinator.GenesisNodeInfoHandler, 0)
			list = append(list, mock.NewNodeInfo(address, pkBytes, 0, InitialRating))
			return list, nil, nil
		},
		GetMinTransactionVersionCalled: func() uint32 {
			return MinTransactionVersion
		},
	}

	nodesCoordinatorInstance := &shardingMocks.NodesCoordinatorStub{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkBytes, defaultChancesSelection, 1)
			return []nodesCoordinator.Validator{v}, nil
		},
		GetAllValidatorsPublicKeysCalled: func() (map[uint32][][]byte, error) {
			keys := make(map[uint32][][]byte)
			keys[0] = make([][]byte, 0)
			keys[0] = append(keys[0], pkBytes)
			return keys, nil
		},
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (nodesCoordinator.Validator, uint32, error) {
			validator, _ := nodesCoordinator.NewValidator(publicKey, defaultChancesSelection, 1)
			return validator, 0, nil
		},
	}

	logsProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{Marshalizer: TestMarshalizer})
	peersRatingHandler, _ := p2pRating.NewPeersRatingHandler(
		p2pRating.ArgPeersRatingHandler{
			TopRatedCache: testscommon.NewCacherMock(),
			BadRatedCache: testscommon.NewCacherMock(),
		})

	messenger := CreateMessengerWithNoDiscoveryAndPeersRatingHandler(peersRatingHandler)
	tpn := &TestProcessorNode{
		ShardCoordinator:        shardCoordinator,
		Messenger:               messenger,
		NodesCoordinator:        nodesCoordinatorInstance,
		HeaderSigVerifier:       &mock.HeaderSigVerifierStub{},
		HeaderIntegrityVerifier: CreateHeaderIntegrityVerifier(),
		ChainID:                 ChainID,
		MinTransactionVersion:   MinTransactionVersion,
		HistoryRepository:       &dblookupext.HistoryRepositoryStub{},
		EpochNotifier:           forking.NewGenericEpochNotifier(),
		ArwenChangeLocker:       &sync.RWMutex{},
		TransactionLogProcessor: logsProcessor,
		PeersRatingHandler:      peersRatingHandler,
		PeerShardMapper:         disabled.NewPeerShardMapper(),
	}
	tpn.NodesSetup = nodesSetup

	tpn.NodeKeys = &TestKeyPair{
		Sk: sk,
		Pk: pk,
	}
	tpn.MultiSigner = TestMultiSig
	tpn.OwnAccount = CreateTestWalletAccount(shardCoordinator, txSignPrivKeyShardId)
	tpn.initDataPools()
	tpn.initHeaderValidator()
	tpn.initRoundHandler()
	tpn.NetworkShardingCollector = mock.NewNetworkShardingCollectorMock()
	tpn.initStorage()
	if tpn.EpochStartNotifier == nil {
		tpn.EpochStartNotifier = notifier.NewEpochStartSubscriptionHandler()
	}
	tpn.initAccountDBsWithPruningStorer()
	tpn.initChainHandler()
	tpn.initEconomicsData(tpn.createDefaultEconomicsConfig())
	tpn.initRatingsData()
	tpn.initRequestedItemsHandler()
	tpn.initResolvers()
	tpn.initValidatorStatistics()
	tpn.GenesisBlocks = CreateGenesisBlocks(
		tpn.AccntState,
		tpn.PeerState,
		tpn.TrieStorageManagers,
		TestAddressPubkeyConverter,
		tpn.NodesSetup,
		tpn.ShardCoordinator,
		tpn.Storage,
		tpn.BlockChain,
		TestMarshalizer,
		TestHasher,
		TestUint64Converter,
		tpn.DataPool,
		tpn.EconomicsData,
	)
	tpn.initBlockTracker()
	tpn.initInterceptors("")
	tpn.initInnerProcessors(arwenConfig.MakeGasMapForTests())
	argsNewScQueryService := smartContract.ArgsNewSCQueryService{
		VmContainer:              tpn.VMContainer,
		EconomicsFee:             tpn.EconomicsData,
		BlockChainHook:           tpn.BlockchainHook,
		BlockChain:               tpn.BlockChain,
		ArwenChangeLocker:        tpn.ArwenChangeLocker,
		Bootstrapper:             tpn.Bootstrapper,
		AllowExternalQueriesChan: common.GetClosedUnbufferedChannel(),
	}
	tpn.SCQueryService, _ = smartContract.NewSCQueryService(argsNewScQueryService)
	tpn.initBlockProcessor(stateCheckpointModulus)
	tpn.BroadcastMessenger, _ = sposFactory.GetBroadcastMessenger(
		TestMarshalizer,
		TestHasher,
		tpn.Messenger,
		tpn.ShardCoordinator,
		tpn.OwnAccount.SkTxSign,
		tpn.OwnAccount.PeerSigHandler,
		tpn.DataPool.Headers(),
		tpn.InterceptorsContainer,
		&testscommon.AlarmSchedulerStub{},
	)
	tpn.setGenesisBlock()
	tpn.initNode()
	tpn.addHandlersForCounters()
	tpn.addGenesisBlocksIntoStorage()

	return tpn
}
