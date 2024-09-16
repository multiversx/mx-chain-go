package bootstrap

import (
	"context"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	epochStartMocks "github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks/epochStart"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/require"
)

func createSovBootStrapProc() *sovereignBootStrapShardProcessor {
	args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
	args.RunTypeComponents = mock.NewSovereignRunTypeComponentsStub()
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	return &sovereignBootStrapShardProcessor{
		&sovereignChainEpochStartBootstrap{
			epochStartProvider,
		},
	}
}

func TestBootStrapSovereignShardProcessor_requestAndProcessForShard(t *testing.T) {
	t.Parallel()

	notarizedShardHeaderHash := []byte("notarizedShardHeaderHash")
	prevShardHeaderHash := []byte("prevShardHeaderHash")
	notarizedMetaHeaderHash := []byte("notarizedMetaHeaderHash")
	prevMetaHeaderHash := []byte("prevMetaHeaderHash")

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.RunTypeComponents = mock.NewSovereignRunTypeComponentsStub()

	prevShardHeader := &block.Header{}
	notarizedShardHeader := &block.Header{
		PrevHash: prevShardHeaderHash,
	}
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
				string(notarizedShardHeaderHash): notarizedShardHeader,
				string(notarizedMetaHeaderHash):  notarizedMetaHeader,
				string(prevShardHeaderHash):      prevShardHeader,
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

	// This will not work properly until we have a sovereign shard storage handler and
	// overwrite SaveDataToStorage, tbd in next PRs
	err := sovProc.requestAndProcessForShard(make([]*block.MiniBlock, 0))
	require.NotNil(t, err)
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
	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{
			Epoch: 4,
		},
		EpochStart: block.EpochStartSovereign{
			Economics: block.Economics{
				PrevEpochStartHash: prevEpochStartHash,
			},
		},
	}

	syncedHeaders := map[string]data.HeaderHandler{
		"hash": &block.SovereignChainHeader{},
	}
	headersSyncedCt := 0
	sovProc.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
		SyncMissingHeadersByHashCalled: func(shardIDs []uint32, headersHashes [][]byte, ctx context.Context) error {
			require.Equal(t, []uint32{core.SovereignChainShardId}, shardIDs)
			require.Equal(t, [][]byte{prevEpochStartHash}, headersHashes)
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

// TODO HERE THIS TEST NEEDS REFACTOR
func TestBootStrapSovereignShardProcessor_processNodesConfigFromStorage(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	hdrHash1 := []byte("hdrHash1")
	hdrHash2 := []byte("hdrHash2")
	sovBlock := &block.SovereignChainHeader{
		Header: &block.Header{
			Epoch: 0,
			MiniBlockHeaders: []block.MiniBlockHeader{
				{
					Hash:          hdrHash1,
					SenderShardID: 1,
				},
				{
					Hash:          hdrHash2,
					SenderShardID: 2,
				},
			},
		},
	}

	pksBytes := createPkBytes(1)
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
					"4294967295": {
						&nodesCoordinator.SerializableValidator{
							PubKey:  pksBytes[core.MetachainShardId],
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

	args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GeneralConfig = testscommon.GetGeneralConfig()
	args.GenesisNodesConfig = getNodesConfigMock(1)

	sesb := initializeStorageEpochStartBootstrap(args)
	sesb.dataPool = dataRetrieverMock.NewPoolsHolderMock()
	sesb.requestHandler = &testscommon.RequestHandlerStub{}
	sesb.epochStartMeta = sovBlock
	sesb.prevEpochStartMeta = sovBlock

	sovProc := &sovereignBootStrapShardProcessor{
		&sovereignChainEpochStartBootstrap{
			sesb.epochStartBootstrap,
		},
	}
	nodesConfig, shardId, err := sovProc.processNodesConfigFromStorage([]byte("pubkey"), sesb.importDbConfig.ImportDBTargetShardID)
	require.Nil(t, err)
	require.Equal(t, core.SovereignChainShardId, shardId)
	require.Equal(t, nodesConfig, expectedNodesConfig)
}
