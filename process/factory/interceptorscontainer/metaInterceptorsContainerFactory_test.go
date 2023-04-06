package interceptorscontainer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const maxTxNonceDeltaAllowed = 100

var chainID = "chain ID"
var errExpected = errors.New("expected error")

func createMetaStubTopicHandler(matchStrToErrOnCreate string, matchStrToErrOnRegister string) process.TopicHandler {
	return &mock.TopicHandlerStub{
		CreateTopicCalled: func(name string, createChannelForTopic bool) error {
			if matchStrToErrOnCreate == "" {
				return nil
			}

			if strings.Contains(name, matchStrToErrOnCreate) {
				return errExpected
			}

			return nil
		},
		RegisterMessageProcessorCalled: func(topic string, identifier string, handler p2p.MessageProcessor) error {
			if matchStrToErrOnRegister == "" {
				return nil
			}

			if strings.Contains(topic, matchStrToErrOnRegister) {
				return errExpected
			}

			return nil
		},
	}
}

func createMetaDataPools() dataRetriever.PoolsHolder {
	pools := &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{}
		},
		MiniBlocksCalled: func() storage.Cacher {
			return testscommon.NewCacherStub()
		},
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		TrieNodesCalled: func() storage.Cacher {
			return testscommon.NewCacherStub()
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
	}

	return pools
}

func createMetaStore() *storageStubs.ChainStorerStub {
	return &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{}, nil
		},
	}
}

// ------- NewInterceptorsContainerFactory

func TestNewMetaInterceptorsContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.ShardCoordinator = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewMetaInterceptorsContainerFactory_InvalidChainIDShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.ChainIdCalled = func() string {
		return ""
	}
	args := getArgumentsMeta(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrInvalidChainID, err)
}

func TestNewMetaInterceptorsContainerFactory_InvalidMinTransactionVersionShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.MinTransactionVersionCalled = func() uint32 {
		return 0
	}
	args := getArgumentsMeta(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrInvalidTransactionVersion, err)
}

func TestNewMetaInterceptorsContainerFactory_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.NodesCoordinator = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilNodesCoordinator, err)
}

func TestNewMetaInterceptorsContainerFactory_NilTopicHandlerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.Messenger = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewMetaInterceptorsContainerFactory_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.Store = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilStore, err)
}

func TestNewMetaInterceptorsContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.IntMarsh = nil
	args := getArgumentsMeta(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewMetaInterceptorsContainerFactory_NilMarshalizerAndSizeCheckShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.IntMarsh = nil
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.SizeCheckDelta = 1
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewMetaInterceptorsContainerFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.Hash = nil
	args := getArgumentsMeta(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewMetaInterceptorsContainerFactory_NilHeaderSigVerifierShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.HeaderSigVerifier = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilHeaderSigVerifier, err)
}

func TestNewMetaInterceptorsContainerFactory_NilHeaderIntegrityVerifierShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.HeaderIntegrityVerifier = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilHeaderIntegrityVerifier, err)
}

func TestNewMetaInterceptorsContainerFactory_NilMultiSignerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	cryptoComp.MultiSigContainer = cryptoMocks.NewMultiSignerContainerMock(nil)
	args := getArgumentsMeta(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewMetaInterceptorsContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.DataPool = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewMetaInterceptorsContainerFactory_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.Accounts = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewMetaInterceptorsContainerFactory_NilAddrConvShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.AddrPubKeyConv = nil
	args := getArgumentsMeta(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewMetaInterceptorsContainerFactory_NilSingleSignerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	cryptoComp.TxSig = nil
	args := getArgumentsMeta(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilSingleSigner, err)
}

func TestNewMetaInterceptorsContainerFactory_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	cryptoComp.TxKeyGen = nil
	args := getArgumentsMeta(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilKeyGen, err)
}

func TestNewMetaInterceptorsContainerFactory_NilTxSignHasherShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.TxSignHasherField = nil
	args := getArgumentsMeta(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewMetaInterceptorsContainerFactory_NilEpochNotifierShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.EpochNotifierField = nil
	args := getArgumentsMeta(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilEpochNotifier, err)
}

func TestNewMetaInterceptorsContainerFactory_NilFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.TxFeeHandler = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewMetaInterceptorsContainerFactory_NilBlackListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.BlockBlackList = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilBlackListCacher, err)
}

func TestNewMetaInterceptorsContainerFactory_NilValidityAttesterShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.ValidityAttester = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilValidityAttester, err)
}

func TestNewMetaInterceptorsContainerFactory_NilSignaturesHandler(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.SignaturesHandler = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilSignaturesHandler, err)
}

func TestNewMetaInterceptorsContainerFactory_NilPeerSignatureHandler(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.PeerSignatureHandler = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilPeerSignatureHandler, err)
}

func TestNewMetaInterceptorsContainerFactory_InvalidExpiryTimespan(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.HeartbeatExpiryTimespanInSec = 0
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrInvalidExpiryTimespan, err)
}

func TestNewMetaInterceptorsContainerFactory_EpochStartTriggerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.EpochStartTrigger = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilEpochStartTrigger, err)
}

func TestNewMetaInterceptorsContainerFactory_NilRequestHandlerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.RequestHandler = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilRequestHandler, err)
}

func TestNewMetaInterceptorsContainerFactory_NilPeerShardMapperShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.PeerShardMapper = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilPeerShardMapper, err)
}

func TestNewMetaInterceptorsContainerFactory_NilHardforkTriggerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.HardforkTrigger = nil
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilHardforkTrigger, err)
}

func TestNewMetaInterceptorsContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.NotNil(t, icf)
	assert.Nil(t, err)
}

func TestNewMetaInterceptorsContainerFactory_ShouldWorkWithSizeCheck(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.SizeCheckDelta = 1
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	assert.NotNil(t, icf)
	assert.Nil(t, err)
	assert.False(t, icf.IsInterfaceNil())
}

// ------- Create

func TestMetaInterceptorsContainerFactory_CreateTopicMetablocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.Messenger = createMetaStubTopicHandler(factory.MetachainBlocksTopic, "")
	icf, _ := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestMetaInterceptorsContainerFactory_CreateTopicShardHeadersForMetachainFailsShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.Messenger = createMetaStubTopicHandler(factory.ShardBlocksTopic, "")
	icf, _ := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestMetaInterceptorsContainerFactory_CreateRegisterForMetablocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.Messenger = createMetaStubTopicHandler("", factory.MetachainBlocksTopic)
	icf, _ := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestMetaInterceptorsContainerFactory_CreateRegisterShardHeadersForMetachainFailsShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.Messenger = createMetaStubTopicHandler("", factory.MetachainBlocksTopic)
	icf, _ := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestMetaInterceptorsContainerFactory_CreateRegisterTrieNodesFailsShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.Messenger = createMetaStubTopicHandler("", factory.AccountTrieNodesTopic)
	icf, _ := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestMetaInterceptorsContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.Messenger = &mock.TopicHandlerStub{
		CreateTopicCalled: func(name string, createChannelForTopic bool) error {
			return nil
		},
		RegisterMessageProcessorCalled: func(topic string, identifier string, handler p2p.MessageProcessor) error {
			return nil
		},
	}
	icf, _ := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestMetaInterceptorsContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	nodesCoordinator := &shardingMocks.NodesCoordinatorMock{
		ShardConsensusSize: 1,
		MetaConsensusSize:  1,
		NbShards:           uint32(noOfShards),
		ShardId:            1,
	}

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsMeta(coreComp, cryptoComp)
	args.ShardCoordinator = shardCoordinator
	args.NodesCoordinator = nodesCoordinator
	args.Messenger = &mock.TopicHandlerStub{
		CreateTopicCalled: func(name string, createChannelForTopic bool) error {
			return nil
		},
		RegisterMessageProcessorCalled: func(topic string, identifier string, handler p2p.MessageProcessor) error {
			return nil
		},
	}
	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(args)
	require.Nil(t, err)

	container, err := icf.Create()

	numInterceptorsMetablock := 1
	numInterceptorsShardHeadersForMetachain := noOfShards
	numInterceptorsTransactionsForMetachain := noOfShards + 1
	numInterceptorsMiniBlocksForMetachain := noOfShards + 1 + 1
	numInterceptorsUnsignedTxsForMetachain := noOfShards + 1
	numInterceptorsRewardsTxsForMetachain := noOfShards
	numInterceptorsTrieNodes := 2
	numInterceptorsPeerAuthForMetachain := 1
	numInterceptorsHeartbeatForMetachain := 1
	numInterceptorsShardValidatorInfoForMetachain := 1
	numInterceptorValidatorInfo := 1
	totalInterceptors := numInterceptorsMetablock + numInterceptorsShardHeadersForMetachain + numInterceptorsTrieNodes +
		numInterceptorsTransactionsForMetachain + numInterceptorsUnsignedTxsForMetachain + numInterceptorsMiniBlocksForMetachain +
		numInterceptorsRewardsTxsForMetachain + numInterceptorsPeerAuthForMetachain + numInterceptorsHeartbeatForMetachain +
		numInterceptorsShardValidatorInfoForMetachain + numInterceptorValidatorInfo

	assert.Nil(t, err)
	assert.Equal(t, totalInterceptors, container.Len())

	err = icf.AddShardTrieNodeInterceptors(container)
	assert.Nil(t, err)
	assert.Equal(t, totalInterceptors+noOfShards, container.Len())
}

func getArgumentsMeta(
	coreComp *mock.CoreComponentsMock,
	cryptoComp *mock.CryptoComponentsMock,
) interceptorscontainer.CommonInterceptorsContainerFactoryArgs {
	return interceptorscontainer.CommonInterceptorsContainerFactoryArgs{
		CoreComponents:               coreComp,
		CryptoComponents:             cryptoComp,
		Accounts:                     &stateMock.AccountsStub{},
		ShardCoordinator:             mock.NewOneShardCoordinatorMock(),
		NodesCoordinator:             shardingMocks.NewNodesCoordinatorMock(),
		Messenger:                    &mock.TopicHandlerStub{},
		Store:                        createMetaStore(),
		DataPool:                     createMetaDataPools(),
		MaxTxNonceDeltaAllowed:       maxTxNonceDeltaAllowed,
		TxFeeHandler:                 &economicsmocks.EconomicsHandlerStub{},
		BlockBlackList:               &testscommon.TimeCacheStub{},
		HeaderSigVerifier:            &mock.HeaderSigVerifierStub{},
		HeaderIntegrityVerifier:      &mock.HeaderIntegrityVerifierStub{},
		ValidityAttester:             &mock.ValidityAttesterStub{},
		EpochStartTrigger:            &mock.EpochStartTriggerStub{},
		WhiteListHandler:             &testscommon.WhiteListHandlerStub{},
		WhiteListerVerifiedTxs:       &testscommon.WhiteListHandlerStub{},
		AntifloodHandler:             &mock.P2PAntifloodHandlerStub{},
		ArgumentsParser:              &mock.ArgumentParserMock{},
		PreferredPeersHolder:         &p2pmocks.PeersHolderStub{},
		RequestHandler:               &testscommon.RequestHandlerStub{},
		PeerSignatureHandler:         &mock.PeerSignatureHandlerStub{},
		SignaturesHandler:            &mock.SignaturesHandlerStub{},
		HeartbeatExpiryTimespanInSec: 30,
		PeerShardMapper:              &p2pmocks.NetworkShardingCollectorStub{},
		HardforkTrigger:              &testscommon.HardforkTriggerStub{},
	}
}
