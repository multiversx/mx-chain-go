package interceptorscontainer_test

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/versioning"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/interceptorscontainer"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func createShardStubTopicHandler(matchStrToErrOnCreate string, matchStrToErrOnRegister string) process.TopicHandler {
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
		RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
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

func createShardDataPools() dataRetriever.PoolsHolder {
	pools := testscommon.NewPoolsHolderStub()
	pools.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{}
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		return testscommon.NewCacherStub()
	}
	pools.PeerChangesBlocksCalled = func() storage.Cacher {
		return testscommon.NewCacherStub()
	}
	pools.MetaBlocksCalled = func() storage.Cacher {
		return testscommon.NewCacherStub()
	}
	pools.UnsignedTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	pools.RewardTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	pools.TrieNodesCalled = func() storage.Cacher {
		return testscommon.NewCacherStub()
	}
	pools.CurrBlockTxsCalled = func() dataRetriever.TransactionCacher {
		return &mock.TxForCurrentBlockStub{}
	}
	return pools
}

func createShardStore() *mock.ChainStorerMock {
	return &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{}
		},
	}
}

//------- NewInterceptorsContainerFactory
func TestNewShardInterceptorsContainerFactory_NilAccountsAdapter(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.Accounts = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewShardInterceptorsContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.ShardCoordinator = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewShardInterceptorsContainerFactory_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.NodesCoordinator = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilNodesCoordinator, err)
}

func TestNewShardInterceptorsContainerFactory_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.Messenger = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewShardInterceptorsContainerFactory_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.Store = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilStore, err)
}

func TestNewShardInterceptorsContainerFactory_NilEpochNotifierShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.EpochNotifierField = nil
	args := getArgumentsShard(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilEpochNotifier, err)
}

func TestNewShardInterceptorsContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.IntMarsh = nil
	args := getArgumentsShard(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewShardInterceptorsContainerFactory_NilMarshalizerAndSizeCheckShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.TxMarsh = nil
	args := getArgumentsShard(coreComp, cryptoComp)

	args.SizeCheckDelta = 1
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewShardInterceptorsContainerFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.Hash = nil
	args := getArgumentsShard(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewShardInterceptorsContainerFactory_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	cryptoComp.TxKeyGen = nil
	args := getArgumentsShard(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilKeyGen, err)
}

func TestNewShardInterceptorsContainerFactory_NilHeaderSigVerifierShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.HeaderSigVerifier = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilHeaderSigVerifier, err)
}

func TestNewShardInterceptorsContainerFactory_NilHeaderIntegrityVerifierShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.HeaderIntegrityVerifier = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilHeaderIntegrityVerifier, err)
}

func TestNewShardInterceptorsContainerFactory_NilTxSignHasherShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.TxSignHasherField = nil
	args := getArgumentsShard(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewShardInterceptorsContainerFactory_NilSingleSignerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	cryptoComp.TxSig = nil
	args := getArgumentsShard(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilSingleSigner, err)
}

func TestNewShardInterceptorsContainerFactory_NilMultiSignerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	cryptoComp.MultiSig = nil
	args := getArgumentsShard(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewShardInterceptorsContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.DataPool = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewShardInterceptorsContainerFactory_NilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.AddrPubKeyConv = nil
	args := getArgumentsShard(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewShardInterceptorsContainerFactory_NilTxFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.TxFeeHandler = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewShardInterceptorsContainerFactory_NilBlackListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.BlockBlackList = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilBlackListCacher, err)
}

func TestNewShardInterceptorsContainerFactory_NilValidityAttesterShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.ValidityAttester = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilValidityAttester, err)
}

func TestNewShardInterceptorsContainerFactory_InvalidChainIDShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.ChainIdCalled = func() string {
		return ""
	}
	args := getArgumentsShard(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrInvalidChainID, err)
}

func TestNewShardInterceptorsContainerFactory_InvalidMinTransactionVersionShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.MinTransactionVersionCalled = func() uint32 {
		return 0
	}
	args := getArgumentsShard(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrInvalidTransactionVersion, err)
}

func TestNewShardInterceptorsContainerFactory_EmptyEpochStartTriggerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.EpochStartTrigger = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilEpochStartTrigger, err)
}

func TestNewShardInterceptorsContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.NotNil(t, icf)
	assert.Nil(t, err)
	assert.False(t, icf.IsInterfaceNil())
}

func TestNewShardInterceptorsContainerFactory_ShouldWorkWithSizeCheck(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.SizeCheckDelta = 1
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.NotNil(t, icf)
	assert.Nil(t, err)
}

//------- Create

func TestShardInterceptorsContainerFactory_CreateTopicCreationTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.Messenger = createShardStubTopicHandler(factory.TransactionTopic, "")
	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateTopicCreationHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.Messenger = createShardStubTopicHandler(factory.ShardBlocksTopic, "")
	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateTopicCreationMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.Messenger = createShardStubTopicHandler(factory.MiniBlocksTopic, "")
	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateTopicCreationMetachainHeadersFailsShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.Messenger = createShardStubTopicHandler(factory.MetachainBlocksTopic, "")
	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateRegisterTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.Messenger = createShardStubTopicHandler("", factory.TransactionTopic)
	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateRegisterHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.Messenger = createShardStubTopicHandler("", factory.ShardBlocksTopic)
	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateRegisterMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.Messenger = createShardStubTopicHandler("", factory.MiniBlocksTopic)
	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateRegisterMetachainHeadersShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.Messenger = createShardStubTopicHandler("", factory.MetachainBlocksTopic)
	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateRegisterTrieNodesShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.Messenger = createShardStubTopicHandler("", factory.AccountTrieNodesTopic)
	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.Messenger = &mock.TopicHandlerStub{
		CreateTopicCalled: func(name string, createChannelForTopic bool) error {
			return nil
		},
		RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
			return nil
		},
	}
	args.WhiteListerVerifiedTxs = &mock.WhiteListHandlerStub{}

	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	container, err := icf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestShardInterceptorsContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	nodesCoordinator := &mock.NodesCoordinatorMock{
		ShardId:            1,
		ShardConsensusSize: 1,
		MetaConsensusSize:  1,
		NbShards:           uint32(noOfShards),
	}

	mesenger := &mock.TopicHandlerStub{
		CreateTopicCalled: func(name string, createChannelForTopic bool) error {
			return nil
		},
		RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
			return nil
		},
	}

	coreComp, cryptoComp := createMockComponentHolders()
	coreComp.AddrPubKeyConv = mock.NewPubkeyConverterMock(32)
	args := getArgumentsShard(coreComp, cryptoComp)
	args.ShardCoordinator = shardCoordinator
	args.NodesCoordinator = nodesCoordinator
	args.Messenger = mesenger

	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	container, err := icf.Create()

	numInterceptorTxs := noOfShards + 1
	numInterceptorsUnsignedTxs := numInterceptorTxs
	numInterceptorsRewardTxs := 1
	numInterceptorHeaders := 1
	numInterceptorMiniBlocks := noOfShards + 2
	numInterceptorMetachainHeaders := 1
	numInterceptorTrieNodes := 1
	totalInterceptors := numInterceptorTxs + numInterceptorsUnsignedTxs + numInterceptorsRewardTxs +
		numInterceptorHeaders + numInterceptorMiniBlocks + numInterceptorMetachainHeaders + numInterceptorTrieNodes

	assert.Nil(t, err)
	assert.Equal(t, totalInterceptors, container.Len())
}

func createMockComponentHolders() (*mock.CoreComponentsMock, *mock.CryptoComponentsMock) {
	coreComponents := &mock.CoreComponentsMock{
		IntMarsh:            &mock.MarshalizerMock{},
		TxMarsh:             &mock.MarshalizerMock{},
		TxSignHasherField:   &mock.HasherMock{},
		Hash:                &mock.HasherMock{},
		UInt64ByteSliceConv: mock.NewNonceHashConverterMock(),
		AddrPubKeyConv:      mock.NewPubkeyConverterMock(32),
		ChainIdCalled: func() string {
			return chainID
		},
		MinTransactionVersionCalled: func() uint32 {
			return 1
		},
		EpochNotifierField: &mock.EpochNotifierStub{},
		TxVersionCheckField: versioning.NewTxVersionChecker(1),
	}
	cryptoComponents := &mock.CryptoComponentsMock{
		BlockSig: &mock.SignerMock{},
		TxSig:    &mock.SignerMock{},
		MultiSig: mock.NewMultiSigner(),
		BlKeyGen: &mock.SingleSignKeyGenMock{},
		TxKeyGen: &mock.SingleSignKeyGenMock{},
	}

	return coreComponents, cryptoComponents
}

func getArgumentsShard(
	coreComp *mock.CoreComponentsMock,
	cryptoComp *mock.CryptoComponentsMock,
) interceptorscontainer.ShardInterceptorsContainerFactoryArgs {
	return interceptorscontainer.ShardInterceptorsContainerFactoryArgs{
		CoreComponents:          coreComp,
		CryptoComponents:        cryptoComp,
		Accounts:                &mock.AccountsStub{},
		ShardCoordinator:        mock.NewOneShardCoordinatorMock(),
		NodesCoordinator:        mock.NewNodesCoordinatorMock(),
		Messenger:               &mock.TopicHandlerStub{},
		Store:                   createShardStore(),
		DataPool:                createShardDataPools(),
		MaxTxNonceDeltaAllowed:  maxTxNonceDeltaAllowed,
		TxFeeHandler:            &mock.FeeHandlerStub{},
		BlockBlackList:          &mock.BlackListHandlerStub{},
		HeaderSigVerifier:       &mock.HeaderSigVerifierStub{},
		HeaderIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		SizeCheckDelta:          0,
		ValidityAttester:        &mock.ValidityAttesterStub{},
		EpochStartTrigger:       &mock.EpochStartTriggerStub{},
		AntifloodHandler:        &mock.P2PAntifloodHandlerStub{},
		WhiteListHandler:        &mock.WhiteListHandlerStub{},
		WhiteListerVerifiedTxs:  &mock.WhiteListHandlerStub{},
		ArgumentsParser:         &mock.ArgumentParserMock{},
	}
}
