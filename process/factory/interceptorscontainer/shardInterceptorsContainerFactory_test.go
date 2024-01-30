package interceptorscontainer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/versioning"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

var providedHardforkPubKey = []byte("provided hardfork pub key")

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

func createShardDataPools() dataRetriever.PoolsHolder {
	pools := dataRetrieverMock.NewPoolsHolderStub()
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
	pools.TrieNodesChunksCalled = func() storage.Cacher {
		return testscommon.NewCacherStub()
	}
	pools.CurrBlockTxsCalled = func() dataRetriever.TransactionCacher {
		return &mock.TxForCurrentBlockStub{}
	}
	return pools
}

func createShardStore() *storageStubs.ChainStorerStub {
	return &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{}, nil
		},
	}
}

// ------- NewInterceptorsContainerFactory
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

func TestNewShardInterceptorsContainerFactory_NilMainMessengerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.MainMessenger = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.True(t, errors.Is(err, process.ErrNilMessenger))
}

func TestNewShardInterceptorsContainerFactory_NilFullArchiveMessengerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.FullArchiveMessenger = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.True(t, errors.Is(err, process.ErrNilMessenger))
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
	cryptoComp.MultiSigContainer = cryptoMocks.NewMultiSignerContainerMock(nil)
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

func TestNewShardInterceptorsContainerFactory_NilMainPeerShardMapperShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.MainPeerShardMapper = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.True(t, errors.Is(err, process.ErrNilPeerShardMapper))
}

func TestNewShardInterceptorsContainerFactory_NilFullArchivePeerShardMapperShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.FullArchivePeerShardMapper = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.True(t, errors.Is(err, process.ErrNilPeerShardMapper))
}

func TestNewShardInterceptorsContainerFactory_NilHardforkTriggerShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.HardforkTrigger = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilHardforkTrigger, err)
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

// ------- Create

func TestShardInterceptorsContainerFactory_CreateTopicsAndRegisterFailure(t *testing.T) {
	t.Parallel()

	testCreateShardTopicShouldFailOnAllMessenger(t, "generateTxInterceptors_create", factory.TransactionTopic, "")
	testCreateShardTopicShouldFailOnAllMessenger(t, "generateTxInterceptors_register", "", factory.TransactionTopic)

	testCreateShardTopicShouldFailOnAllMessenger(t, "generateUnsignedTxsInterceptors", factory.UnsignedTransactionTopic, "")

	testCreateShardTopicShouldFailOnAllMessenger(t, "generateRewardTxInterceptor", factory.RewardsTransactionTopic, "")

	testCreateShardTopicShouldFailOnAllMessenger(t, "generateHeaderInterceptors", factory.ShardBlocksTopic, "")

	testCreateShardTopicShouldFailOnAllMessenger(t, "generateMiniBlocksInterceptors", factory.MiniBlocksTopic, "")

	testCreateShardTopicShouldFailOnAllMessenger(t, "generateMetachainHeaderInterceptors", factory.MetachainBlocksTopic, "")

	testCreateShardTopicShouldFailOnAllMessenger(t, "generateTrieNodesInterceptors", factory.AccountTrieNodesTopic, "")

	testCreateShardTopicShouldFailOnAllMessenger(t, "generateValidatorInfoInterceptor", common.ValidatorInfoTopic, "")

	testCreateShardTopicShouldFailOnAllMessenger(t, "generateHeartbeatInterceptor", common.HeartbeatV2Topic, "")

	testCreateShardTopicShouldFailOnAllMessenger(t, "generatePeerShardIntercepto", common.ConnectionTopic, "")

	t.Run("generatePeerAuthenticationInterceptor_main", testCreateShardTopicShouldFail(common.PeerAuthenticationTopic, ""))
}
func testCreateShardTopicShouldFailOnAllMessenger(t *testing.T, testNamePrefix string, matchStrToErrOnCreate string, matchStrToErrOnRegister string) {
	t.Run(testNamePrefix+"main messenger", testCreateShardTopicShouldFail(matchStrToErrOnCreate, matchStrToErrOnRegister))
	t.Run(testNamePrefix+"full archive messenger", testCreateShardTopicShouldFail(matchStrToErrOnCreate, matchStrToErrOnRegister))
}

func testCreateShardTopicShouldFail(matchStrToErrOnCreate string, matchStrToErrOnRegister string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createMockComponentHolders()
		args := getArgumentsShard(coreComp, cryptoComp)
		if strings.Contains(t.Name(), "full_archive") {
			args.NodeOperationMode = common.FullArchiveMode
			args.FullArchiveMessenger = createShardStubTopicHandler(matchStrToErrOnCreate, matchStrToErrOnRegister)
		} else {
			args.MainMessenger = createShardStubTopicHandler(matchStrToErrOnCreate, matchStrToErrOnRegister)
		}
		icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

		mainContainer, fullArchiveContainer, err := icf.Create()

		assert.Nil(t, mainContainer)
		assert.Nil(t, fullArchiveContainer)
		assert.Equal(t, errExpected, err)
	}
}

func TestShardInterceptorsContainerFactory_NilSignaturesHandler(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.SignaturesHandler = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilSignaturesHandler, err)
}

func TestShardInterceptorsContainerFactory_NilPeerSignatureHandler(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.PeerSignatureHandler = nil
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilPeerSignatureHandler, err)
}

func TestShardInterceptorsContainerFactory_InvalidExpiryTimespan(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.HeartbeatExpiryTimespanInSec = 0
	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrInvalidExpiryTimespan, err)
}

func TestShardInterceptorsContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	args.MainMessenger = &mock.TopicHandlerStub{
		CreateTopicCalled: func(name string, createChannelForTopic bool) error {
			return nil
		},
		RegisterMessageProcessorCalled: func(topic string, identifier string, handler p2p.MessageProcessor) error {
			return nil
		},
	}
	args.WhiteListerVerifiedTxs = &testscommon.WhiteListHandlerStub{}

	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

	mainContainer, fullArchiveContainer, err := icf.Create()

	assert.NotNil(t, mainContainer)
	assert.NotNil(t, fullArchiveContainer)
	assert.Nil(t, err)
}

func TestShardInterceptorsContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("normal mode", func(t *testing.T) {
		t.Parallel()

		noOfShards := 4

		shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
		shardCoordinator.SetNoShards(uint32(noOfShards))
		shardCoordinator.CurrentShard = 1

		nodesCoordinator := &shardingMocks.NodesCoordinatorMock{
			ShardId:            1,
			ShardConsensusSize: 1,
			MetaConsensusSize:  1,
			NbShards:           uint32(noOfShards),
		}

		coreComp, cryptoComp := createMockComponentHolders()
		coreComp.AddrPubKeyConv = testscommon.NewPubkeyConverterMock(32)
		args := getArgumentsShard(coreComp, cryptoComp)
		args.ShardCoordinator = shardCoordinator
		args.NodesCoordinator = nodesCoordinator
		args.PreferredPeersHolder = &p2pmocks.PeersHolderStub{}

		icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

		mainContainer, fullArchiveContainer, err := icf.Create()

		numInterceptorTxs := noOfShards + 1
		numInterceptorsUnsignedTxs := numInterceptorTxs
		numInterceptorsRewardTxs := 1
		numInterceptorHeaders := 1
		numInterceptorMiniBlocks := noOfShards + 2
		numInterceptorMetachainHeaders := 1
		numInterceptorTrieNodes := 1
		numInterceptorPeerAuth := 1
		numInterceptorHeartbeat := 1
		numInterceptorsShardValidatorInfo := 1
		numInterceptorValidatorInfo := 1
		totalInterceptors := numInterceptorTxs + numInterceptorsUnsignedTxs + numInterceptorsRewardTxs +
			numInterceptorHeaders + numInterceptorMiniBlocks + numInterceptorMetachainHeaders + numInterceptorTrieNodes +
			numInterceptorPeerAuth + numInterceptorHeartbeat + numInterceptorsShardValidatorInfo + numInterceptorValidatorInfo

		assert.Nil(t, err)
		assert.Equal(t, totalInterceptors, mainContainer.Len())
		assert.Equal(t, 0, fullArchiveContainer.Len())
	})

	t.Run("full archive mode", func(t *testing.T) {
		t.Parallel()

		noOfShards := 4

		shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
		shardCoordinator.SetNoShards(uint32(noOfShards))
		shardCoordinator.CurrentShard = 1

		nodesCoordinator := &shardingMocks.NodesCoordinatorMock{
			ShardId:            1,
			ShardConsensusSize: 1,
			MetaConsensusSize:  1,
			NbShards:           uint32(noOfShards),
		}

		coreComp, cryptoComp := createMockComponentHolders()
		coreComp.AddrPubKeyConv = testscommon.NewPubkeyConverterMock(32)
		args := getArgumentsShard(coreComp, cryptoComp)
		args.NodeOperationMode = common.FullArchiveMode
		args.ShardCoordinator = shardCoordinator
		args.NodesCoordinator = nodesCoordinator
		args.PreferredPeersHolder = &p2pmocks.PeersHolderStub{}

		icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(args)

		mainContainer, fullArchiveContainer, err := icf.Create()

		numInterceptorTxs := noOfShards + 1
		numInterceptorsUnsignedTxs := numInterceptorTxs
		numInterceptorsRewardTxs := 1
		numInterceptorHeaders := 1
		numInterceptorMiniBlocks := noOfShards + 2
		numInterceptorMetachainHeaders := 1
		numInterceptorTrieNodes := 1
		numInterceptorPeerAuth := 1
		numInterceptorHeartbeat := 1
		numInterceptorsShardValidatorInfo := 1
		numInterceptorValidatorInfo := 1
		totalInterceptors := numInterceptorTxs + numInterceptorsUnsignedTxs + numInterceptorsRewardTxs +
			numInterceptorHeaders + numInterceptorMiniBlocks + numInterceptorMetachainHeaders + numInterceptorTrieNodes +
			numInterceptorPeerAuth + numInterceptorHeartbeat + numInterceptorsShardValidatorInfo + numInterceptorValidatorInfo

		assert.Nil(t, err)
		assert.Equal(t, totalInterceptors, mainContainer.Len())
		assert.Equal(t, totalInterceptors-1, fullArchiveContainer.Len()) // no peerAuthentication needed
	})
}

func createMockComponentHolders() (*mock.CoreComponentsMock, *mock.CryptoComponentsMock) {
	coreComponents := &mock.CoreComponentsMock{
		IntMarsh:            &mock.MarshalizerMock{},
		TxMarsh:             &mock.MarshalizerMock{},
		TxSignHasherField:   &hashingMocks.HasherMock{},
		Hash:                &hashingMocks.HasherMock{},
		UInt64ByteSliceConv: mock.NewNonceHashConverterMock(),
		AddrPubKeyConv:      testscommon.NewPubkeyConverterMock(32),
		ChainIdCalled: func() string {
			return chainID
		},
		MinTransactionVersionCalled: func() uint32 {
			return 1
		},
		EpochNotifierField:         &epochNotifier.EpochNotifierStub{},
		TxVersionCheckField:        versioning.NewTxVersionChecker(1),
		HardforkTriggerPubKeyField: providedHardforkPubKey,
		EnableEpochsHandlerField:   &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	multiSigner := cryptoMocks.NewMultiSigner()
	cryptoComponents := &mock.CryptoComponentsMock{
		BlockSig:          &mock.SignerMock{},
		TxSig:             &mock.SignerMock{},
		MultiSigContainer: cryptoMocks.NewMultiSignerContainerMock(multiSigner),
		BlKeyGen:          &mock.SingleSignKeyGenMock{},
		TxKeyGen:          &mock.SingleSignKeyGenMock{},
	}

	return coreComponents, cryptoComponents
}

func getArgumentsShard(
	coreComp *mock.CoreComponentsMock,
	cryptoComp *mock.CryptoComponentsMock,
) interceptorscontainer.CommonInterceptorsContainerFactoryArgs {
	return interceptorscontainer.CommonInterceptorsContainerFactoryArgs{
		CoreComponents:               coreComp,
		CryptoComponents:             cryptoComp,
		Accounts:                     &stateMock.AccountsStub{},
		ShardCoordinator:             mock.NewOneShardCoordinatorMock(),
		NodesCoordinator:             shardingMocks.NewNodesCoordinatorMock(),
		MainMessenger:                &mock.TopicHandlerStub{},
		FullArchiveMessenger:         &mock.TopicHandlerStub{},
		Store:                        createShardStore(),
		DataPool:                     createShardDataPools(),
		MaxTxNonceDeltaAllowed:       maxTxNonceDeltaAllowed,
		TxFeeHandler:                 &economicsmocks.EconomicsHandlerStub{},
		BlockBlackList:               &testscommon.TimeCacheStub{},
		HeaderSigVerifier:            &consensus.HeaderSigVerifierMock{},
		HeaderIntegrityVerifier:      &mock.HeaderIntegrityVerifierStub{},
		SizeCheckDelta:               0,
		ValidityAttester:             &mock.ValidityAttesterStub{},
		EpochStartTrigger:            &mock.EpochStartTriggerStub{},
		AntifloodHandler:             &mock.P2PAntifloodHandlerStub{},
		WhiteListHandler:             &testscommon.WhiteListHandlerStub{},
		WhiteListerVerifiedTxs:       &testscommon.WhiteListHandlerStub{},
		ArgumentsParser:              &mock.ArgumentParserMock{},
		PreferredPeersHolder:         &p2pmocks.PeersHolderStub{},
		RequestHandler:               &testscommon.RequestHandlerStub{},
		PeerSignatureHandler:         &mock.PeerSignatureHandlerStub{},
		SignaturesHandler:            &mock.SignaturesHandlerStub{},
		HeartbeatExpiryTimespanInSec: 30,
		MainPeerShardMapper:          &p2pmocks.NetworkShardingCollectorStub{},
		FullArchivePeerShardMapper:   &p2pmocks.NetworkShardingCollectorStub{},
		HardforkTrigger:              &testscommon.HardforkTriggerStub{},
	}
}
