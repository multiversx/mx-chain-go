package bootstrap

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	factoryInterceptors "github.com/multiversx/mx-chain-go/epochStart/bootstrap/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	epochStartMocks "github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks/epochStart"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
)

func createSovBootStrapProc() *sovereignBootStrapShardProcessor {
	args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
	args.RunTypeComponents = mock.NewSovereignRunTypeComponentsStub()
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}
	return &sovereignBootStrapShardProcessor{
		&sovereignChainEpochStartBootstrap{
			epochStartProvider,
		},
	}
}

func TestBootStrapSovereignShardProcessor_requestAndProcessForShard(t *testing.T) {
	t.Parallel()

	prevShardHeaderHash := []byte("prevShardHeaderHash")
	notarizedMetaHeaderHash := []byte("notarizedMetaHeaderHash")
	prevMetaHeaderHash := []byte("prevMetaHeaderHash")

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.RunTypeComponents = mock.NewSovereignRunTypeComponentsStub()

	prevShardHeader := &block.Header{}
	notarizedMetaHeader := &block.SovereignChainHeader{
		Header: &block.Header{
			PrevHash: prevMetaHeaderHash,
		},
	}
	metaBlockInstance := &block.SovereignChainHeader{
		Header: &block.Header{},
	}
	prevMetaBlock := &block.SovereignChainHeader{
		Header: &block.Header{},
	}

	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.syncedHeaders = make(map[string]data.HeaderHandler)
	epochStartProvider.epochStartMeta = metaBlockInstance
	epochStartProvider.prevEpochStartMeta = prevMetaBlock
	epochStartProvider.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
		GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
			return map[string]data.HeaderHandler{
				string(notarizedMetaHeaderHash): notarizedMetaHeader,
				string(prevShardHeaderHash):     prevShardHeader,
			}, nil
		},
	}
	epochStartProvider.dataPool = &dataRetrieverMock.PoolsHolderStub{
		TrieNodesCalled: func() storage.Cacher {
			return &testscommon.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, true
				},
			}
		},
	}

	epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{}
	epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}
	epochStartProvider.nodesConfig = &nodesCoordinator.NodesCoordinatorRegistry{}
	sovProc := &sovereignBootStrapShardProcessor{
		&sovereignChainEpochStartBootstrap{
			epochStartProvider,
		},
	}

	err := sovProc.requestAndProcessForShard(make([]*block.MiniBlock, 0))
	require.Nil(t, err)
}

func TestBootStrapSovereignShardProcessor_computeNumShards(t *testing.T) {
	t.Parallel()

	sovProc := createSovBootStrapProc()
	require.Equal(t, uint32(0x1), sovProc.computeNumShards(nil))
}

func TestBootStrapSovereignShardProcessor_createRequestHandler(t *testing.T) {
	t.Parallel()

	sovProc := createSovBootStrapProc()
	reqHandler, err := sovProc.createRequestHandler()
	require.Nil(t, err)
	require.Equal(t, "*requestHandlers.sovereignResolverRequestHandler", fmt.Sprintf("%T", reqHandler))
}

func TestBootStrapSovereignShardProcessor_createResolversContainer(t *testing.T) {
	t.Parallel()

	sovProc := createSovBootStrapProc()
	require.Nil(t, sovProc.createResolversContainer())
}

func TestBootStrapSovereignShardProcessor_syncHeadersFrom(t *testing.T) {
	t.Parallel()

	sovProc := createSovBootStrapProc()

	prevEpochStartHash := []byte("prevEpochStartHash")
	lastCrossChainHeaderHash := []byte("lastCrossChainHeaderHash")
	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{
			Epoch: 4,
		},
		EpochStart: block.EpochStartSovereign{
			Economics: block.Economics{
				PrevEpochStartHash: prevEpochStartHash,
			},
			LastFinalizedCrossChainHeader: block.EpochStartCrossChainData{
				ShardID:    core.MainChainShardId,
				HeaderHash: lastCrossChainHeaderHash,
			},
		},
	}

	syncedHeaders := map[string]data.HeaderHandler{
		"hash": &block.SovereignChainHeader{},
	}
	headersSyncedCt := 0
	sovProc.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
		SyncMissingHeadersByHashCalled: func(shardIDs []uint32, headersHashes [][]byte, ctx context.Context) error {
			require.Equal(t, []uint32{core.MainChainShardId, core.SovereignChainShardId}, shardIDs)
			require.Equal(t, [][]byte{lastCrossChainHeaderHash, prevEpochStartHash}, headersHashes)
			headersSyncedCt++
			return nil
		},
		GetHeadersCalled: func() (map[string]data.HeaderHandler, error) {
			return syncedHeaders, nil
		},
	}

	res, err := sovProc.syncHeadersFrom(sovHdr)
	require.Nil(t, err)
	require.Equal(t, res, syncedHeaders)
	require.Equal(t, 1, headersSyncedCt)

	res, err = sovProc.syncHeadersFromStorage(sovHdr, 0, 0, DefaultTimeToWaitForRequestedData)
	require.Nil(t, err)
	require.Equal(t, res, syncedHeaders)
	require.Equal(t, 2, headersSyncedCt)
}

func TestBootStrapSovereignShardProcessor_processNodesConfigFromStorage(t *testing.T) {
	t.Parallel()

	sovBlock := &block.SovereignChainHeader{
		Header: &block.Header{
			Epoch: 0,
		},
	}

	pksBytes := createPkBytes(1)
	delete(pksBytes, core.MetachainShardId)

	expectedNodesConfig := &nodesCoordinator.NodesCoordinatorRegistry{
		EpochsConfig: map[string]*nodesCoordinator.EpochValidators{
			"0": {
				EligibleValidators: map[string][]*nodesCoordinator.SerializableValidator{
					"0": {
						&nodesCoordinator.SerializableValidator{
							PubKey:  pksBytes[0],
							Chances: 1,
							Index:   0,
						},
					},
				},
				WaitingValidators: map[string][]*nodesCoordinator.SerializableValidator{},
				LeavingValidators: map[string][]*nodesCoordinator.SerializableValidator{},
			},
		},
		CurrentEpoch: 0,
	}

	sovProc := createSovBootStrapProc()
	sovProc.runTypeComponents = &mock.RunTypeComponentsStub{
		NodesCoordinatorWithRaterFactory: &testscommon.NodesCoordinatorFactoryMock{
			CreateNodesCoordinatorWithRaterCalled: func(args *nodesCoordinator.NodesCoordinatorWithRaterArgs) (nodesCoordinator.NodesCoordinator, error) {
				return &shardingMocks.NodesCoordinatorMock{
					NodesCoordinatorToRegistryCalled: func(epoch uint32) nodesCoordinator.NodesCoordinatorRegistryHandler {
						return expectedNodesConfig
					},
				}, nil
			},
		},
	}

	sovProc.dataPool = dataRetrieverMock.NewPoolsHolderMock()
	sovProc.epochStartMeta = sovBlock
	sovProc.prevEpochStartMeta = sovBlock

	nodesConfig, shardId, err := sovProc.processNodesConfigFromStorage([]byte("pubkey"), 0)
	require.Nil(t, err)
	require.Equal(t, core.SovereignChainShardId, shardId)
	require.Equal(t, nodesConfig, expectedNodesConfig)
}

func TestBootStrapSovereignShardProcessor_createEpochStartMetaSyncer(t *testing.T) {
	t.Parallel()

	sovProc := createSovBootStrapProc()
	epochStartBlockSyncer, err := sovProc.createEpochStartMetaSyncer()
	require.Nil(t, err)
	require.Equal(t, "*bootstrap.epochStartSovereignSyncer", fmt.Sprintf("%T", epochStartBlockSyncer))

	epochStartSyncer := epochStartBlockSyncer.(*epochStartSovereignSyncer)
	require.Equal(t, "*bootstrap.epochStartSovereignBlockProcessor", fmt.Sprintf("%T", epochStartSyncer.metaBlockProcessor))
}

func TestBootStrapSovereignShardProcessor_createStorageEpochStartMetaSyncer(t *testing.T) {
	t.Parallel()

	sovProc := createSovBootStrapProc()

	args := getEpochStartSyncerArgs()
	epochStartBlockSyncer, err := sovProc.createStorageEpochStartMetaSyncer(args)
	require.Nil(t, err)
	require.Equal(t, "*bootstrap.epochStartSovereignSyncer", fmt.Sprintf("%T", epochStartBlockSyncer))
}

func TestBootStrapSovereignShardProcessor_createEpochStartInterceptorsContainers(t *testing.T) {
	t.Parallel()

	sovProc := createSovBootStrapProc()
	sovProc.dataPool = dataRetrieverMock.NewPoolsHolderMock()

	args := factoryInterceptors.ArgsEpochStartInterceptorContainer{
		CoreComponents:          sovProc.coreComponentsHolder,
		CryptoComponents:        sovProc.cryptoComponentsHolder,
		Config:                  sovProc.generalConfig,
		ShardCoordinator:        sovProc.shardCoordinator,
		MainMessenger:           sovProc.mainMessenger,
		FullArchiveMessenger:    sovProc.fullArchiveMessenger,
		DataPool:                dataRetrieverMock.NewPoolsHolderMock(),
		WhiteListHandler:        sovProc.whiteListHandler,
		WhiteListerVerifiedTxs:  sovProc.whiteListerVerifiedTxs,
		ArgumentsParser:         sovProc.argumentsParser,
		HeaderIntegrityVerifier: sovProc.headerIntegrityVerifier,
		RequestHandler:          sovProc.requestHandler,
		SignaturesHandler:       sovProc.mainMessenger,
		NodeOperationMode:       sovProc.nodeOperationMode,
		AccountFactory:          sovProc.runTypeComponents.AccountsCreator(),
	}
	mainContainer, fullContainer, err := sovProc.createEpochStartInterceptorsContainers(args)
	require.Nil(t, err)

	sovShardIDStr := fmt.Sprintf("_%d", core.SovereignChainShardId)
	allKeys := map[string]struct{}{
		factory.TransactionTopic + sovShardIDStr:         {},
		factory.UnsignedTransactionTopic + sovShardIDStr: {},
		factory.ShardBlocksTopic + sovShardIDStr:         {},
		factory.MiniBlocksTopic + sovShardIDStr:          {},
		factory.ValidatorTrieNodesTopic + sovShardIDStr:  {},
		factory.AccountTrieNodesTopic + sovShardIDStr:    {},
		common.PeerAuthenticationTopic:                   {},
		common.HeartbeatV2Topic + sovShardIDStr:          {},
		common.ConnectionTopic:                           {},
		common.ValidatorInfoTopic + sovShardIDStr:        {},
		factory.ExtendedHeaderProofTopic + sovShardIDStr: {},
	}

	iterateFunc := func(key string, interceptor process.Interceptor) bool {
		require.False(t, strings.Contains(strings.ToLower(key), "meta"))
		delete(allKeys, key)
		return true
	}

	mainContainer.Iterate(iterateFunc)
	require.Empty(t, allKeys)
	require.Zero(t, fullContainer.Len())
}

func TestBootStrapSovereignShardProcessor_createCrossHeaderRequester(t *testing.T) {
	t.Parallel()

	sovProc := createSovBootStrapProc()
	sovProc.dataPool = dataRetrieverMock.NewPoolsHolderMock()

	requester, err := sovProc.createCrossHeaderRequester()
	require.Nil(t, requester)
	require.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	require.ErrorContains(t, err, "extendedHeaderRequester")

	sovProc.requestHandler = &testscommon.ExtendedShardHeaderRequestHandlerStub{}
	requester, err = sovProc.createCrossHeaderRequester()
	require.Nil(t, err)
	require.Equal(t, "*sync.extendedHeaderRequester", fmt.Sprintf("%T", requester))
}
