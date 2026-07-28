package sync_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	goSync "sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	headersCache "github.com/multiversx/mx-chain-go/process/asyncExecution/cache"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/round"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	proofscache "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/proofsCache"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

const testProcessWaitTime = time.Second

var errExpected = errors.New("expected error")

func setupStore(marshaller marshal.Marshalizer, prevHdr data.HeaderHandler, returnError error) dataRetriever.StorageService {
	return &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if returnError != nil {
						return nil, returnError
					}
					prevHdrBytes, _ := marshaller.Marshal(prevHdr)
					return prevHdrBytes, nil
				},
			}, nil
		},
	}
}

func setupForkDetector(highestNonce uint64) process.ForkDetector {
	return &mock.ForkDetectorMock{
		CheckForkCalled: func() *process.ForkInfo {
			return process.NewForkInfo()
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return highestNonce
		},
		GetHighestFinalBlockHashCalled: func() []byte {
			return []byte("hash")
		},
		ProbableHighestNonceCalled: func() uint64 {
			return highestNonce
		},
		RemoveHeaderCalled: func(nonce uint64, hash []byte) {},
		GetNotarizedHeaderHashCalled: func(nonce uint64) []byte {
			return nil
		},
	}
}

type headerAndHash struct {
	header data.HeaderHandler
	hash   []byte
}

func setupPools(headersAndHashes ...headerAndHash) dataRetriever.PoolsHolder {
	pools := dataRetrieverMock.NewPoolsHolderStub()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{
			GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
				for _, hh := range headersAndHashes {
					if hh.header.GetNonce() == hdrNonce {
						return []data.HeaderHandler{hh.header}, [][]byte{hh.hash}, nil
					}
				}

				return nil, nil, errors.New("err")
			},
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				for i, hh := range headersAndHashes {
					if string(hh.hash) != string(hash) {
						continue
					}

					if i > 0 {
						return headersAndHashes[i-1].header, nil
					}

					return &block.HeaderV3{
						Nonce:         1,
						BlockBodyType: block.TxBlock,
						LastExecutionResult: &block.ExecutionResultInfo{
							ExecutionResult: &block.BaseExecutionResult{
								HeaderNonce: 0,
								HeaderHash:  []byte("hash0"),
							},
						},
					}, nil
				}

				return nil, errors.New("err")
			},
		}
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := cache.NewCacherStub()
		cs.RegisterHandlerCalled = func(i func(key []byte, value interface{})) {}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
			return make(block.MiniBlockSlice, 0), true
		}
		return cs
	}
	pools.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &testscommon.ShardedDataStub{
			OnExecutedBlockCalled: func(header data.HeaderHandler, rootHash []byte) error {
				return nil
			},
		}
	}
	pools.ProofsCalled = func() dataRetriever.ProofsPool {
		return &dataRetrieverMock.ProofsPoolMock{}
	}
	return pools
}

// setupPoolsDirectHashMapping creates a pools holder where GetHeaderByHash returns the header
// that actually has the given hash (direct mapping), rather than the off-by-one mapping in setupPools.
// This is needed for tests that exercise the backward hash-walk in prepareForSyncIfNeeded.
func setupPoolsDirectHashMapping(headersAndHashes ...headerAndHash) dataRetriever.PoolsHolder {
	pools := dataRetrieverMock.NewPoolsHolderStub()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{
			GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
				for _, hh := range headersAndHashes {
					if hh.header.GetNonce() == hdrNonce {
						return []data.HeaderHandler{hh.header}, [][]byte{hh.hash}, nil
					}
				}

				return nil, nil, errors.New("err")
			},
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				for _, hh := range headersAndHashes {
					if string(hh.hash) == string(hash) {
						return hh.header, nil
					}
				}

				return nil, errors.New("err")
			},
		}
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := cache.NewCacherStub()
		cs.RegisterHandlerCalled = func(i func(key []byte, value interface{})) {}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
			return make(block.MiniBlockSlice, 0), true
		}
		return cs
	}
	pools.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &testscommon.ShardedDataStub{
			OnExecutedBlockCalled: func(header data.HeaderHandler, rootHash []byte) error {
				return nil
			},
		}
	}
	pools.ProofsCalled = func() dataRetriever.ProofsPool {
		return &dataRetrieverMock.ProofsPoolMock{}
	}
	return pools
}

type removedFlags struct {
	flagHdrRemovedFromHeaders      bool
	flagHdrRemovedFromStorage      bool
	flagHdrRemovedFromForkDetector bool
}

func createMockPools() *dataRetrieverMock.PoolsHolderStub {
	pools := dataRetrieverMock.NewPoolsHolderStub()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{}
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &cache.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
			RegisterHandlerCalled: func(i func(key []byte, value interface{})) {},
		}
		return cs
	}
	pools.ProofsCalled = func() dataRetriever.ProofsPool {
		return &dataRetrieverMock.ProofsPoolMock{}
	}

	return pools
}

func createStore() *storageStubs.ChainStorerStub {
	return &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return nil, process.ErrMissingHeader
				},
				RemoveCalled: func(key []byte) error {
					return nil
				},
			}, nil
		},
	}
}

func generateTestCache() storage.Cacher {
	c, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 1000, Shards: 1, SizeInBytes: 0})
	return c
}

func generateTestUnit() storage.Storer {
	storer, _ := storageunit.NewStorageUnit(
		generateTestCache(),
		database.NewMemDB(),
	)

	return storer
}

func createFullStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, generateTestUnit())
	store.AddStorer(dataRetriever.RewardTransactionUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, generateTestUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ReceiptsUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ScheduledSCRsUnit, generateTestUnit())
	store.AddStorer(dataRetriever.UserAccountsUnit, generateTestUnit())
	store.AddStorer(dataRetriever.PeerAccountsUnit, generateTestUnit())
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, generateTestUnit())
	return store
}

func createBlockProcessor(blk data.ChainHandler) *testscommon.BlockProcessorStub {
	blockProcessorMock := &testscommon.BlockProcessorStub{
		ProcessBlockCalled: func(hdr data.HeaderHandler, bdy data.BodyHandler, haveTime func() time.Duration) error {
			_ = blk.SetCurrentBlockHeaderAndRootHash(hdr.(*block.Header), hdr.GetRootHash())
			return nil
		},
		RevertCurrentBlockCalled: func() {
		},
		CommitBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
			return nil
		},
		ProcessScheduledBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
			return nil
		},
	}

	return blockProcessorMock
}

func createForkDetector(removedNonce uint64, removedHash []byte, remFlags *removedFlags) process.ForkDetector {
	return &mock.ForkDetectorMock{
		RemoveHeaderCalled: func(nonce uint64, hash []byte) {
			if nonce == removedNonce {
				remFlags.flagHdrRemovedFromForkDetector = true
			}
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return removedNonce
		},
		GetHighestFinalBlockHashCalled: func() []byte {
			return removedHash
		},
		ProbableHighestNonceCalled: func() uint64 {
			return uint64(0)
		},
		GetNotarizedHeaderHashCalled: func(nonce uint64) []byte {
			return nil
		},
	}
}

func initBlockchain() *testscommon.ChainHandlerStub {
	blkc := &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Nonce:     uint64(0),
				Signature: []byte("genesis signature"),
				RandSeed:  []byte{0},
			}
		},
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("genesis header hash")
		},
	}

	return blkc
}

func initNetworkWatcher() process.NetworkConnectionWatcher {
	return &mock.NetworkConnectionWatcherStub{
		IsConnectedToTheNetworkCalled: func() bool {
			return true
		},
	}
}

func initRoundHandler() consensus.RoundHandler {
	roundArgs := createDefaultRoundArgs()
	roundArgs.CurrentTimeStamp = time.Now()
	rnd, _ := round.NewRound(roundArgs)

	return rnd
}

func CreateShardBootstrapMockArguments() sync.ArgShardBootstrapper {
	argsBaseBootstrapper := sync.ArgBaseBootstrapper{
		PoolsHolder:                  createMockPools(),
		Store:                        createStore(),
		ChainHandler:                 initBlockchain(),
		RoundHandler:                 &mock.RoundHandlerMock{},
		BlockProcessor:               &testscommon.BlockProcessorStub{},
		ExecutionManager:             &processMocks.ExecutionManagerMock{},
		Hasher:                       &hashingMocks.HasherMock{},
		Marshalizer:                  &mock.MarshalizerMock{},
		ForkDetector:                 &mock.ForkDetectorMock{},
		RequestHandler:               &testscommon.RequestHandlerStub{},
		ShardCoordinator:             mock.NewOneShardCoordinatorMock(),
		Accounts:                     &stateMock.AccountsStub{},
		BlackListHandler:             &testscommon.TimeCacheStub{},
		NetworkWatcher:               initNetworkWatcher(),
		BootStorer:                   &mock.BoostrapStorerMock{},
		StorageBootstrapper:          &mock.StorageBootstrapperMock{},
		EpochHandler:                 &mock.EpochStartTriggerStub{},
		EpochStartTrigger:            &mock.EpochStartTriggerStub{},
		MiniblocksProvider:           &mock.MiniBlocksProviderStub{},
		Uint64Converter:              &mock.Uint64ByteSliceConverterMock{},
		AppStatusHandler:             &statusHandlerMock.AppStatusHandlerStub{},
		OutportHandler:               &outport.OutportStub{},
		AccountsDBSyncer:             &mock.AccountsDBSyncerStub{},
		CurrentEpochProvider:         &testscommon.CurrentEpochProviderStub{},
		HistoryRepo:                  &dblookupext.HistoryRepositoryStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessWaitTime:              testProcessWaitTime,
		ProcessWaitTimeSupernova:     testProcessWaitTime,
		RepopulateTokensSupplies:     false,
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		EnableRoundsHandler:          &testscommon.EnableRoundsHandlerStub{},
		ProcessConfigsHandler:        testscommon.GetDefaultProcessConfigsHandler(),
	}

	argsShardBootstrapper := sync.ArgShardBootstrapper{
		ArgBaseBootstrapper: argsBaseBootstrapper,
		MetaFinalityView:    &testscommon.MetaFinalityViewStub{},
		BlockTracker:        &mock.BlockTrackerMock{},
	}

	return argsShardBootstrapper
}

// ------- NewShardBootstrap

func TestNewShardBootstrap_NilPoolsHolderShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.PoolsHolder = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilPoolsHolder, err)
}

func TestNewShardBootstrap_NilAccountsDBSyncerShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.AccountsDBSyncer = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilAccountsDBSyncer, err)
}

func TestNewShardBootstrap_PoolsHolderRetNilOnHeadersShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	pools := createMockPools()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		return nil
	}
	args.PoolsHolder = pools

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilHeadersDataPool, err)
}

func TestNewShardBootstrap_NilProofsPool(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	pools := createMockPools()
	pools.ProofsCalled = func() dataRetriever.ProofsPool {
		return nil
	}
	args.PoolsHolder = pools

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilProofsPool, err)
}

func TestNewShardBootstrap_PoolsHolderRetNilOnTxBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	pools := createMockPools()
	pools.MiniBlocksCalled = func() storage.Cacher {
		return nil
	}
	args.PoolsHolder = pools

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilTxBlockBody, err)
}

func TestNewShardBootstrap_NilMetaFinalityViewShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.MetaFinalityView = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilMetaFinalityView, err)
}

func TestNewShardBootstrap_NilBlockTrackerShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.BlockTracker = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilBlockTracker, err)
}

func TestNewShardBootstrap_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.Store = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilStore, err)
}

func TestNewShardBootstrap_NilAppStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.AppStatusHandler = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilAppStatusHandler, err)
}

func TestNewShardBootstrap_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.ChainHandler = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestNewShardBootstrap_NilRoundHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.RoundHandler = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilRoundHandler, err)
}

func TestNewShardBootstrap_NilBlockProcessorShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.BlockProcessor = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilBlockProcessor, err)
}

func TestNewShardBootstrap_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.Hasher = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewShardBootstrap_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.Marshalizer = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewShardBootstrap_NilForkDetectorShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.ForkDetector = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilForkDetector, err)
}

func TestNewShardBootstrap_NilRequestHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.RequestHandler = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilRequestHandler, err)
}

func TestNewShardBootstrap_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.ShardCoordinator = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewShardBootstrap_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.Accounts = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewShardBootstrap_NilBlackListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.BlackListHandler = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilBlackListCacher, err)
}

func TestNewShardBootstrap_InvalidProcessTimeShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.ProcessWaitTime = time.Millisecond*100 - 1

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.True(t, errors.Is(err, process.ErrInvalidProcessWaitTime))
}

func TestNewShardBootstrap_InvalidProcessWaitTimeSupernovaShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.ProcessWaitTimeSupernova = time.Millisecond*100 - 1

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.True(t, errors.Is(err, process.ErrInvalidProcessWaitTime))
	assert.Contains(t, err.Error(), "Supernova")
}

func TestNewShardBootstrap_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.EnableEpochsHandler = nil

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.True(t, errors.Is(err, process.ErrNilEnableEpochsHandler))
}

func TestNewShardBootstrap_PoolsHolderRetNilOnProofsShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	pools := createMockPools()
	pools.ProofsCalled = func() dataRetriever.ProofsPool {
		return nil
	}
	args.PoolsHolder = pools

	bs, err := sync.NewShardBootstrap(args)

	assert.True(t, check.IfNil(bs))
	assert.Equal(t, process.ErrNilProofsPool, err)
}

func TestNewShardBootstrap_MissingStorer(t *testing.T) {
	t.Parallel()

	t.Run("missing BlockHeaderUnit", testShardWithMissingStorer(dataRetriever.BlockHeaderUnit))
	t.Run("missing ShardHdrNonceHashDataUnit", testShardWithMissingStorer(dataRetriever.ShardHdrNonceHashDataUnit))
}

func testShardWithMissingStorer(missingUnit dataRetriever.UnitType) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		args := CreateShardBootstrapMockArguments()
		args.Store = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				if unitType == missingUnit ||
					strings.Contains(unitType.String(), missingUnit.String()) {
					return nil, fmt.Errorf("%w for %s", storage.ErrKeyNotFound, missingUnit.String())
				}
				return &storageStubs.StorerStub{}, nil
			},
		}

		bs, err := sync.NewShardBootstrap(args)
		assert.True(t, check.IfNil(bs))
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), storage.ErrKeyNotFound.Error()))
	}
}

func TestNewShardBootstrap_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	wasCalled := 0

	pools := dataRetrieverMock.NewPoolsHolderStub()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		sds := &mock.HeadersCacherStub{}

		sds.AddCalled = func(headerHash []byte, header data.HeaderHandler) {
			assert.Fail(t, "should have not reached this point")
		}
		sds.RegisterHandlerCalled = func(func(header data.HeaderHandler, key []byte)) {
			wasCalled++
		}
		return sds
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := cache.NewCacherStub()
		cs.RegisterHandlerCalled = func(i func(key []byte, value interface{})) {
			wasCalled++
		}

		return cs
	}
	pools.ProofsCalled = func() dataRetriever.ProofsPool {
		return &dataRetrieverMock.ProofsPoolMock{}
	}

	args.PoolsHolder = pools
	args.IsInImportMode = true
	bs, err := sync.NewShardBootstrap(args)

	assert.False(t, check.IfNil(bs))
	assert.Nil(t, err)
	assert.Equal(t, 3, wasCalled)
	assert.False(t, bs.IsInterfaceNil())
	assert.True(t, bs.IsInImportMode())

	args.IsInImportMode = false
	bs, err = sync.NewShardBootstrap(args)

	assert.False(t, check.IfNil(bs))
	assert.Nil(t, err)
	assert.False(t, bs.IsInImportMode())
	assert.Equal(t, testProcessWaitTime, bs.ProcessWaitTime())
}

// ------- processing

func TestBootstrap_ShouldReturnTimeIsOutWhenMissingHeader(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	hdr := block.Header{Nonce: 1}
	blkc := &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &hdr
		},
	}
	args.ChainHandler = blkc

	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 100
	}
	forkDetector.GetNotarizedHeaderHashCalled = func(nonce uint64) []byte {
		return nil
	}
	args.ForkDetector = forkDetector
	args.RoundHandler, _ = round.NewRound(createDefaultRoundArgs())
	args.BlockProcessor = createBlockProcessor(args.ChainHandler)

	bs, _ := sync.NewShardBootstrap(args)
	r := bs.SyncBlock(context.Background())

	assert.Equal(t, process.ErrTimeIsOut, r)
}

func TestBootstrap_ShouldReturnTimeIsOutWhenMissingBody(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	hdr := block.Header{Nonce: 1, PubKeysBitmap: []byte("X")}
	blkc := &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &hdr
		},
	}
	args.ChainHandler = blkc

	hash := []byte("aaa")

	pools := createMockPools()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		sds := &mock.HeadersCacherStub{}
		sds.GetHeaderByHashCalled = func(key []byte) (handler data.HeaderHandler, e error) {
			if bytes.Equal(hash, key) {
				return &block.Header{Nonce: 2}, nil
			}

			return nil, errors.New("err")
		}
		return sds
	}
	args.PoolsHolder = pools

	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 2
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return 1
	}
	forkDetector.GetNotarizedHeaderHashCalled = func(nonce uint64) []byte {
		return nil
	}
	args.ForkDetector = forkDetector
	args.RoundHandler, _ = round.NewRound(createDefaultRoundArgs())

	bs, _ := sync.NewShardBootstrap(args)
	bs.RequestHeaderWithNonce(2)
	r := bs.SyncBlock(context.Background())

	assert.Equal(t, process.ErrTimeIsOut, r)
}

func TestBootstrap_ShouldNotNeedToSync(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	hdr := block.Header{Nonce: 1, Round: 0}
	blkc := &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &hdr
		},
	}
	args.ChainHandler = blkc
	args.BlockProcessor = createBlockProcessor(args.ChainHandler)

	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return hdr.Nonce
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 1
	}
	forkDetector.GetNotarizedHeaderHashCalled = func(nonce uint64) []byte {
		return nil
	}
	args.ForkDetector = forkDetector
	args.RoundHandler = initRoundHandler()

	bs, _ := sync.NewShardBootstrap(args)

	_ = bs.StartSyncingBlocks()
	time.Sleep(200 * time.Millisecond)
	_ = bs.Close()
}

func TestBootstrap_SyncShouldSyncOneBlock(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	hdr := block.Header{Nonce: 1, Round: 0}
	blkc := &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &hdr
		},
	}
	args.ChainHandler = blkc
	args.BlockProcessor = createBlockProcessor(args.ChainHandler)

	hash := []byte("aaa")

	mutDataAvailable := goSync.RWMutex{}
	dataAvailable := false

	pools := dataRetrieverMock.NewPoolsHolderStub()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		sds := &mock.HeadersCacherStub{}
		sds.GetHeaderByHashCalled = func(key []byte) (handler data.HeaderHandler, e error) {
			mutDataAvailable.RLock()
			defer mutDataAvailable.RUnlock()

			if bytes.Equal(hash, key) && dataAvailable {
				return &block.Header{
					Nonce:         2,
					Round:         1,
					BlockBodyType: block.TxBlock,
					RootHash:      []byte("bbb")}, nil
			}

			return nil, errors.New("err")
		}

		sds.RegisterHandlerCalled = func(func(header data.HeaderHandler, key []byte)) {
		}

		return sds
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := cache.NewCacherStub()
		cs.RegisterHandlerCalled = func(i func(key []byte, value interface{})) {
		}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("bbb"), key) && dataAvailable {
				return make(block.MiniBlockSlice, 0), true
			}

			return nil, false
		}

		return cs
	}
	pools.ProofsCalled = func() dataRetriever.ProofsPool {
		return &dataRetrieverMock.ProofsPoolMock{}
	}

	args.PoolsHolder = pools

	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return hdr.Nonce
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 2
	}
	forkDetector.GetNotarizedHeaderHashCalled = func(nonce uint64) []byte {
		return nil
	}
	args.ForkDetector = forkDetector

	account := &stateMock.AccountsStub{}
	account.RootHashCalled = func() ([]byte, error) {
		return nil, nil
	}
	args.Accounts = account
	args.RoundHandler, _ = round.NewRound(createDefaultRoundArgs())

	bs, _ := sync.NewShardBootstrap(args)
	_ = bs.StartSyncingBlocks()

	time.Sleep(200 * time.Millisecond)

	mutDataAvailable.Lock()
	dataAvailable = true
	mutDataAvailable.Unlock()

	time.Sleep(500 * time.Millisecond)

	_ = bs.Close()
}

func TestBootstrap_ShouldReturnNilErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	hdr := block.Header{Nonce: 1}
	blkc := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &hdr
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	args.ChainHandler = blkc
	args.BlockProcessor = createBlockProcessor(args.ChainHandler)

	hash := []byte("aaa")
	header := &block.Header{
		Nonce:         2,
		Round:         1,
		BlockBodyType: block.TxBlock,
		RootHash:      []byte("bbb")}

	pools := dataRetrieverMock.NewPoolsHolderStub()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		sds := &mock.HeadersCacherStub{}
		sds.GetHeaderByNonceAndShardIdCalled = func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
			if hdrNonce == 2 {
				return []data.HeaderHandler{header}, [][]byte{hash}, nil
			}

			return nil, nil, errors.New("err")
		}

		return sds
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := cache.NewCacherStub()
		cs.RegisterHandlerCalled = func(i func(key []byte, value interface{})) {
		}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("bbb"), key) {
				return make(block.MiniBlockSlice, 0), true
			}

			return nil, false
		}

		return cs
	}
	pools.ProofsCalled = func() dataRetriever.ProofsPool {
		return &dataRetrieverMock.ProofsPoolMock{}
	}
	args.PoolsHolder = pools

	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 2
	}
	forkDetector.GetNotarizedHeaderHashCalled = func(nonce uint64) []byte {
		return nil
	}
	args.ForkDetector = forkDetector
	args.RoundHandler, _ = round.NewRound(createDefaultRoundArgs())

	bs, _ := sync.NewShardBootstrap(args)
	r := bs.SyncBlock(context.Background())

	assert.Nil(t, r)
}

func TestBootstrap_SyncBlockShouldReturnErrorWhenProcessBlockFailed(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	hdr := block.Header{Nonce: 1, PubKeysBitmap: []byte("X")}
	blkc := &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &hdr
		},
	}
	args.ChainHandler = blkc

	blockProcessor := createBlockProcessor(args.ChainHandler)
	blockProcessor.ProcessBlockCalled = func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
		return process.ErrBlockHashDoesNotMatch
	}
	args.BlockProcessor = blockProcessor

	hash := []byte("aaa")
	header := &block.Header{
		Nonce:         2,
		Round:         1,
		BlockBodyType: block.TxBlock,
		RootHash:      []byte("bbb")}

	pools := dataRetrieverMock.NewPoolsHolderStub()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		sds := &mock.HeadersCacherStub{}
		sds.GetHeaderByNonceAndShardIdCalled = func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
			if hdrNonce == 2 {
				return []data.HeaderHandler{header}, [][]byte{hash}, nil
			}
			return nil, nil, errors.New("err")
		}

		return sds
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := cache.NewCacherStub()
		cs.RegisterHandlerCalled = func(i func(key []byte, value interface{})) {
		}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("bbb"), key) {
				return make(block.MiniBlockSlice, 0), true
			}

			return nil, false
		}

		return cs
	}
	pools.ProofsCalled = func() dataRetriever.ProofsPool {
		return &dataRetrieverMock.ProofsPoolMock{}
	}
	args.PoolsHolder = pools

	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return hdr.Nonce
	}
	forkDetector.GetHighestFinalBlockHashCalled = func() []byte {
		return []byte("hash")
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 2
	}
	forkDetector.RemoveHeaderCalled = func(nonce uint64, hash []byte) {}
	forkDetector.GetNotarizedHeaderHashCalled = func(nonce uint64) []byte {
		return nil
	}
	args.ForkDetector = forkDetector
	args.RoundHandler, _ = round.NewRound(createDefaultRoundArgs())

	bs, _ := sync.NewShardBootstrap(args)

	err := bs.SyncBlock(context.Background())
	assert.Equal(t, process.ErrBlockHashDoesNotMatch, err)
}

func TestBootstrap_GetNodeStateShouldReturnSynchronizedWhenCurrentBlockIsNilAndRoundIndexIsZero(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	forkDetector := &mock.ForkDetectorMock{
		CheckForkCalled: func() *process.ForkInfo {
			return process.NewForkInfo()
		},
		ProbableHighestNonceCalled: func() uint64 {
			return 0
		},
	}
	args.ForkDetector = forkDetector
	args.RoundHandler = initRoundHandler()

	bs, _ := sync.NewShardBootstrap(args)
	bs.ComputeNodeState()

	assert.Equal(t, common.NsSynchronized, bs.GetNodeState())
}

func TestBootstrap_GetNodeStateShouldReturnNotSynchronizedWhenCurrentBlockIsNilAndRoundIndexIsGreaterThanZero(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 1
	}
	args.ForkDetector = forkDetector

	roundArgs := createDefaultRoundArgs()
	roundArgs.CurrentTimeStamp = time.Now().Add(100 * time.Millisecond)
	args.RoundHandler, _ = round.NewRound(roundArgs)

	bs, _ := sync.NewShardBootstrap(args)
	bs.ComputeNodeState()

	assert.Equal(t, common.NsNotSynchronized, bs.GetNodeState())
}

func TestBootstrap_GetNodeStateShouldReturnSynchronizedWhenNodeIsSynced(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	hdr := block.Header{Nonce: 0}
	blkc := &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &hdr
		},
	}
	args.ChainHandler = blkc

	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 0
	}
	args.ForkDetector = forkDetector
	args.RoundHandler = initRoundHandler()

	bs, _ := sync.NewShardBootstrap(args)
	bs.ComputeNodeState()

	assert.Equal(t, common.NsSynchronized, bs.GetNodeState())
}

func TestBootstrap_GetNodeStateShouldReturnNotSynchronizedWhenNodeIsNotSynced(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	hdr := block.Header{Nonce: 0}
	blkc := &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &hdr
		},
	}
	args.ChainHandler = blkc

	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 1
	}
	args.ForkDetector = forkDetector

	roundArgs := createDefaultRoundArgs()
	roundArgs.CurrentTimeStamp = time.Now().Add(100 * time.Millisecond)
	args.RoundHandler, _ = round.NewRound(roundArgs)

	bs, _ := sync.NewShardBootstrap(args)
	bs.ComputeNodeState()

	assert.Equal(t, common.NsNotSynchronized, bs.GetNodeState())
}

func TestBootstrap_GetNodeStateShouldReturnNotSynchronizedWhenForkIsDetectedAndItReceivesTheSameWrongHeader(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	hdr1 := block.Header{Nonce: 1, Round: 2, PubKeysBitmap: []byte("A")}
	hash1 := []byte("hash1")

	hdr2 := block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("B")}
	hash2 := []byte("hash2")

	blkc := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &hdr1
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	args.ChainHandler = blkc

	selfNotarizedHeaders := []data.HeaderHandler{
		&hdr2,
	}
	selfNotarizedHeadersHashes := [][]byte{
		hash2,
	}

	pools := createMockPools()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		sds := &mock.HeadersCacherStub{
			RegisterHandlerCalled: func(func(header data.HeaderHandler, key []byte)) {},
			GetHeaderByHashCalled: func(key []byte) (handler data.HeaderHandler, e error) {
				if bytes.Equal(key, hash1) {
					return &hdr1, nil
				}
				if bytes.Equal(key, hash2) {
					return &hdr2, nil
				}

				return nil, errors.New("err")
			},
		}
		return sds
	}
	args.PoolsHolder = pools
	args.RoundHandler = &mock.RoundHandlerMock{RoundIndex: 2}
	args.ForkDetector, _ = sync.NewShardForkDetector(
		args.RoundHandler,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)

	bs, _ := sync.NewShardBootstrap(args)

	_ = args.ForkDetector.AddHeader(&hdr1, hash1, process.BHProcessed, nil, nil)
	_ = args.ForkDetector.AddHeader(&hdr2, hash2, process.BHNotarized, selfNotarizedHeaders, selfNotarizedHeadersHashes)

	bs.ComputeNodeState()
	assert.Equal(t, common.NsNotSynchronized, bs.GetNodeState())
	assert.True(t, bs.IsForkDetected())

	if bs.GetNodeState() == common.NsNotSynchronized && bs.IsForkDetected() {
		args.ForkDetector.RemoveHeader(hdr1.GetNonce(), hash1)
		bs.ReceivedHeaders(&hdr1, hash1)
		_ = args.ForkDetector.AddHeader(&hdr1, hash1, process.BHProcessed, nil, nil)
	}

	bs.ComputeNodeState()
	assert.Equal(t, common.NsNotSynchronized, bs.GetNodeState())
	assert.True(t, bs.IsForkDetected())
}

func TestBootstrap_GetNodeStateShouldReturnSynchronizedWhenForkIsDetectedAndItReceivesTheGoodHeader(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	hdr1 := block.Header{Nonce: 1, Round: 2, PubKeysBitmap: []byte("A")}
	hash1 := []byte("hash1")

	hdr2 := block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("B")}
	hash2 := []byte("hash2")

	blkc := &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &hdr2
		},
	}
	args.ChainHandler = blkc

	selfNotarizedHeaders := []data.HeaderHandler{
		&hdr2,
	}
	selfNotarizedHeadersHashes := [][]byte{
		hash2,
	}

	pools := createMockPools()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		sds := &mock.HeadersCacherStub{
			RegisterHandlerCalled: func(func(header data.HeaderHandler, key []byte)) {},
			GetHeaderByHashCalled: func(key []byte) (handler data.HeaderHandler, e error) {
				if bytes.Equal(key, hash1) {
					return &hdr1, nil
				}
				if bytes.Equal(key, hash2) {
					return &hdr2, nil
				}

				return nil, errors.New("err")
			},
		}
		return sds
	}
	args.PoolsHolder = pools

	args.RoundHandler = &mock.RoundHandlerMock{RoundIndex: 2}
	args.ForkDetector, _ = sync.NewShardForkDetector(
		args.RoundHandler,
		&testscommon.TimeCacheStub{},
		&mock.BlockTrackerMock{},
		0,
		0,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{},
		&dataRetrieverMock.ProofsPoolMock{},
		&chainParameters.ChainParametersHandlerStub{},
		testscommon.GetDefaultProcessConfigsHandler(),
		0,
	)

	bs, _ := sync.NewShardBootstrap(args)

	_ = args.ForkDetector.AddHeader(&hdr1, hash1, process.BHProcessed, nil, nil)
	_ = args.ForkDetector.AddHeader(&hdr2, hash2, process.BHNotarized, selfNotarizedHeaders, selfNotarizedHeadersHashes)

	bs.ComputeNodeState()
	assert.Equal(t, common.NsNotSynchronized, bs.GetNodeState())
	assert.True(t, bs.IsForkDetected())

	if bs.GetNodeState() == common.NsNotSynchronized && bs.IsForkDetected() {
		args.ForkDetector.RemoveHeader(hdr1.GetNonce(), hash1)
		bs.ReceivedHeaders(&hdr2, hash2)
		_ = args.ForkDetector.AddHeader(&hdr2, hash2, process.BHProcessed, selfNotarizedHeaders, selfNotarizedHeadersHashes)
		bs.SetNodeStateCalculated(false)
	}

	bs.ComputeNodeState()
	assert.Equal(t, common.NsSynchronized, bs.GetNodeState())
	assert.False(t, bs.IsForkDetected())
}

func TestBootstrap_GetHeaderFromPoolShouldReturnNil(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	args.ForkDetector = forkDetector
	args.RoundHandler = initRoundHandler()

	bs, _ := sync.NewShardBootstrap(args)
	hdr, _, _ := process.GetShardHeaderFromPoolWithNonce(0, 0, args.PoolsHolder.Headers())

	assert.False(t, check.IfNil(bs))
	assert.Nil(t, hdr)
}

func TestBootstrap_GetHeaderFromPoolShouldReturnHeader(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	hdr := &block.Header{Nonce: 0}
	hash := []byte("aaa")

	pools := createMockPools()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		sds := &mock.HeadersCacherStub{}

		sds.GetHeaderByNonceAndShardIdCalled = func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
			if hdrNonce == 0 {
				return []data.HeaderHandler{hdr}, [][]byte{hash}, nil
			}
			return nil, nil, errors.New("err")
		}

		sds.RegisterHandlerCalled = func(func(header data.HeaderHandler, key []byte)) {
		}

		return sds
	}
	args.PoolsHolder = pools
	args.RoundHandler = initRoundHandler()

	bs, _ := sync.NewShardBootstrap(args)
	hdr2, _, _ := process.GetShardHeaderFromPoolWithNonce(0, 0, pools.Headers())

	assert.False(t, check.IfNil(bs))
	assert.True(t, hdr == hdr2)
}

func TestShardGetBlockFromPoolShouldReturnBlock(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	mbsAndHashes := make([]*block.MiniblockAndHash, 0)
	args.RoundHandler = initRoundHandler()
	args.MiniblocksProvider = &mock.MiniBlocksProviderStub{
		GetMiniBlocksCalled: func(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
			return mbsAndHashes, nil
		},
	}

	bs, _ := sync.NewShardBootstrap(args)
	mbHashes := make([][]byte, 0)
	mbHashes = append(mbHashes, []byte("aaaa"))
	gotMbsAndHashes, _ := bs.GetMiniBlocks(mbHashes)

	assert.True(t, reflect.DeepEqual(mbsAndHashes, gotMbsAndHashes))
}

// ------- testing received headers

func TestBootstrap_ReceivedHeadersFoundInPoolShouldAddToForkDetector(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	addedHash := []byte("hash")
	addedHdr := &block.Header{}

	pools := createMockPools()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		sds := &mock.HeadersCacherStub{}
		sds.RegisterHandlerCalled = func(func(header data.HeaderHandler, key []byte)) {
		}
		sds.GetHeaderByHashCalled = func(key []byte) (handler data.HeaderHandler, e error) {
			if bytes.Equal(key, addedHash) {
				return addedHdr, nil
			}

			return nil, errors.New("err")
		}

		return sds
	}
	args.PoolsHolder = pools

	wasAdded := false

	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.AddHeaderCalled = func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error {
		if state == process.BHProcessed {
			return errors.New("processed")
		}

		if !bytes.Equal(hash, addedHash) {
			return errors.New("hash mismatch")
		}

		if !reflect.DeepEqual(header, addedHdr) {
			return errors.New("header mismatch")
		}

		wasAdded = true
		return nil
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 0
	}
	args.ForkDetector = forkDetector
	args.RoundHandler = initRoundHandler()

	bs, _ := sync.NewShardBootstrap(args)
	bs.ReceivedHeaders(addedHdr, addedHash)

	assert.True(t, wasAdded)
}

// ------- RollBack

func TestBootstrap_RollBackNilBlockchainHeaderShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	bs, _ := sync.NewShardBootstrap(args)
	err := bs.RollBack(false)

	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestBootstrap_RollBackNilParamHeaderShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	blkc := initBlockchain()
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return nil
	}
	args.ChainHandler = blkc

	bs, _ := sync.NewShardBootstrap(args)
	err := bs.RollBack(false)

	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestBootstrap_RollBackIsNotEmptyShouldErr(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	newHdrHash := []byte("new hdr hash")
	newHdrNonce := uint64(6)

	remFlags := &removedFlags{}

	pools := createMockPools()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		sds := &mock.HeadersCacherStub{
			RemoveHeaderByHashCalled: func(key []byte) {
				if bytes.Equal(key, newHdrHash) {
					remFlags.flagHdrRemovedFromHeaders = true
				}
			},
		}
		return sds
	}
	args.PoolsHolder = pools

	blkc := initBlockchain()
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.Header{
			PubKeysBitmap: []byte("X"),
			Nonce:         newHdrNonce,
		}
	}
	blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
		return newHdrHash
	}
	args.ChainHandler = blkc
	args.ForkDetector = createForkDetector(newHdrNonce, newHdrHash, remFlags)

	bs, _ := sync.NewShardBootstrap(args)
	err := bs.RollBack(false)

	assert.Equal(t, sync.ErrRollBackBehindFinalHeader, err)
}

func TestBootstrap_RollBackIsEmptyCallRollBackOneBlockOkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	// retain if the remove process from different storage locations has been called
	remFlags := &removedFlags{}
	currentHdrNonce := uint64(8)
	currentHdrHash := []byte("current header hash")

	// define prev tx block body "strings" as in this test there are a lot of stubs that
	// constantly need to check some defined symbols
	// prevTxBlockBodyHash := []byte("prev block body hash")
	prevTxBlockBodyBytes := []byte("prev block body bytes")
	prevTxBlockBody := &block.Body{}

	// define prev header "strings"
	prevHdrHash := []byte("prev header hash")
	prevHdrBytes := []byte("prev header bytes")
	prevHdrRootHash := []byte("prev header root hash")
	prevHdr := &block.Header{
		Signature: []byte("sig of the prev header as to be unique in this context"),
		RootHash:  prevHdrRootHash,
	}

	pools := createMockPools()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		sds := &mock.HeadersCacherStub{
			RemoveHeaderByHashCalled: func(key []byte) {
				if bytes.Equal(key, currentHdrHash) {
					remFlags.flagHdrRemovedFromHeaders = true
				}
			},
		}
		return sds
	}
	args.PoolsHolder = pools

	// a mock blockchain with special header and tx block bodies stubs (defined above)
	blkc := &testscommon.ChainHandlerStub{}
	hdr := &block.Header{
		Nonce: currentHdrNonce,
		// empty bitmap
		PrevHash: prevHdrHash,
	}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return hdr
	}
	var setRootHash []byte
	blkc.SetCurrentBlockHeaderAndRootHashCalled = func(handler data.HeaderHandler, rootHash []byte) error {
		hdr = prevHdr
		setRootHash = rootHash
		return nil
	}

	hdrHash := make([]byte, 0)
	blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
		return hdrHash
	}
	blkc.SetCurrentBlockHeaderHashCalled = func(i []byte) {
		hdrHash = i
	}
	args.ChainHandler = blkc
	args.Store = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return prevHdrBytes, nil
				},
				RemoveCalled: func(key []byte) error {
					remFlags.flagHdrRemovedFromStorage = true
					return nil
				},
			}, nil
		},
	}
	args.BlockProcessor = &testscommon.BlockProcessorStub{
		RestoreBlockIntoPoolsCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
			return nil
		},
	}
	args.Hasher = &mock.HasherStub{
		ComputeCalled: func(s string) []byte {
			return currentHdrHash
		},
	}
	args.Marshalizer = &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return []byte("X"), nil
		},
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			if bytes.Equal(buff, prevHdrBytes) {
				_, ok := obj.(*block.Header)
				if !ok {
					return nil
				}

				// bytes represent a header (strings are returns from hdrUnit.Get which is also a stub here)
				// copy only defined fields
				obj.(*block.Header).Signature = prevHdr.Signature
				obj.(*block.Header).RootHash = prevHdrRootHash
				return nil
			}
			if bytes.Equal(buff, prevTxBlockBodyBytes) {
				// bytes represent a tx block body (strings are returns from txBlockUnit.Get which is also a stub here)
				// copy only defined fields
				_, ok := obj.(*block.Body)
				if !ok {
					return nil
				}

				obj.(*block.Body).MiniBlocks = prevTxBlockBody.MiniBlocks
				return nil
			}

			return nil
		},
	}
	args.ForkDetector = createForkDetector(currentHdrNonce, currentHdrHash, remFlags)
	args.Accounts = &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
			return nil
		},
	}

	bs, _ := sync.NewShardBootstrap(args)
	bs.SetForkNonce(currentHdrNonce)
	err := bs.RollBack(true)

	assert.Nil(t, err)
	assert.True(t, remFlags.flagHdrRemovedFromHeaders)
	assert.True(t, remFlags.flagHdrRemovedFromStorage)
	assert.True(t, remFlags.flagHdrRemovedFromForkDetector)
	assert.Equal(t, blkc.GetCurrentBlockHeader(), prevHdr)
	assert.Equal(t, blkc.GetCurrentBlockHeaderHash(), prevHdrHash)
	assert.Equal(t, prevHdr.RootHash, setRootHash)
}

func TestBootstrap_RollbackIsEmptyCallRollBackOneBlockToGenesisShouldWork(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	// retain if the remove process from different storage locations has been called
	remFlags := &removedFlags{}

	currentHdrNonce := uint64(1)
	currentHdrHash := []byte("current header hash")

	// define prev tx block body "strings" as in this test there are a lot of stubs that
	// constantly need to check some defined symbols
	// prevTxBlockBodyHash := []byte("prev block body hash")
	prevTxBlockBodyBytes := []byte("prev block body bytes")
	prevTxBlockBody := &block.Body{}

	// define prev header "strings"
	prevHdrHash := []byte("prev header hash")
	prevHdrBytes := []byte("prev header bytes")
	prevHdrRootHash := []byte("prev header root hash")
	prevHdr := &block.Header{
		Signature: []byte("sig of the prev header as to be unique in this context"),
		RootHash:  prevHdrRootHash,
	}

	pools := createMockPools()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		sds := &mock.HeadersCacherStub{
			RemoveHeaderByHashCalled: func(key []byte) {
				if bytes.Equal(key, currentHdrHash) {
					remFlags.flagHdrRemovedFromHeaders = true
				}
			},
		}
		return sds
	}
	args.PoolsHolder = pools

	// a mock blockchain with special header and tx block bodies stubs (defined above)
	blkc := &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return prevHdr
		},
	}
	hdr := &block.Header{
		Nonce: currentHdrNonce,
		// empty bitmap
		PrevHash: prevHdrHash,
	}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return hdr
	}
	var setRootHash []byte
	blkc.SetCurrentBlockHeaderAndRootHashCalled = func(handler data.HeaderHandler, rootHash []byte) error {
		hdr = nil
		setRootHash = rootHash
		return nil
	}

	hdrHash := make([]byte, 0)
	blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
		return hdrHash
	}
	blkc.SetCurrentBlockHeaderHashCalled = func(i []byte) {
		hdrHash = nil
	}
	args.ChainHandler = blkc
	args.Store = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return prevHdrBytes, nil
				},
				RemoveCalled: func(key []byte) error {
					remFlags.flagHdrRemovedFromStorage = true
					return nil
				},
			}, nil
		},
	}
	args.BlockProcessor = &testscommon.BlockProcessorStub{
		RestoreBlockIntoPoolsCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
			return nil
		},
	}
	args.Hasher = &mock.HasherStub{
		ComputeCalled: func(s string) []byte {
			return currentHdrHash
		},
	}
	args.Marshalizer = &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return []byte("X"), nil
		},
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			if bytes.Equal(buff, prevHdrBytes) {
				_, ok := obj.(*block.Header)
				if !ok {
					return nil
				}

				// bytes represent a header (strings are returns from hdrUnit.Get which is also a stub here)
				// copy only defined fields
				obj.(*block.Header).Signature = prevHdr.Signature
				obj.(*block.Header).RootHash = prevHdrRootHash
				return nil
			}
			if bytes.Equal(buff, prevTxBlockBodyBytes) {
				// bytes represent a tx block body (strings are returns from txBlockUnit.Get which is also a stub here)
				// copy only defined fields
				_, ok := obj.(*block.Body)
				if !ok {
					return nil
				}

				obj.(*block.Body).MiniBlocks = prevTxBlockBody.MiniBlocks
				return nil
			}

			return nil
		},
	}
	args.ForkDetector = createForkDetector(currentHdrNonce, currentHdrHash, remFlags)
	args.Accounts = &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
			return nil
		},
	}

	bs, _ := sync.NewShardBootstrap(args)
	bs.SetForkNonce(currentHdrNonce)
	err := bs.RollBack(true)

	assert.Nil(t, err)
	assert.True(t, remFlags.flagHdrRemovedFromHeaders)
	assert.True(t, remFlags.flagHdrRemovedFromStorage)
	assert.True(t, remFlags.flagHdrRemovedFromForkDetector)
	assert.Nil(t, blkc.GetCurrentBlockHeader())
	assert.Nil(t, blkc.GetCurrentBlockHeaderHash())
	assert.Nil(t, setRootHash)
}

// ------- GetTxBodyHavingHash

func TestBootstrap_GetTxBodyHavingHashReturnsFromCacherShouldWork(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	mbh := []byte("requested hash")
	requestedHash := make([][]byte, 0)
	requestedHash = append(requestedHash, mbh)
	mbsAndHashes := make([]*block.MiniblockAndHash, 0)

	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {},
	})
	args.ChainHandler = blkc
	args.MiniblocksProvider = &mock.MiniBlocksProviderStub{
		GetMiniBlocksCalled: func(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
			for _, hash := range hashes {
				if bytes.Equal(hash, mbh) {
					return mbsAndHashes, nil
				}
			}

			return nil, nil
		},
	}

	bs, _ := sync.NewShardBootstrap(args)
	gotMbsAndHashes, _ := bs.GetMiniBlocks(requestedHash)

	assert.True(t, reflect.DeepEqual(gotMbsAndHashes, mbsAndHashes))
}

func TestBootstrap_GetTxBodyHavingHashNotFoundInCacherOrStorageShouldRetEmptySlice(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	mbh := []byte("requested hash")
	requestedHash := make([][]byte, 0)
	requestedHash = append(requestedHash, mbh)

	txBlockUnit := &storageStubs.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return nil, errors.New("not found")
		},
	}

	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {},
	})
	args.ChainHandler = blkc
	args.Store = createFullStore()
	args.Store.AddStorer(dataRetriever.TransactionUnit, txBlockUnit)

	bs, err := sync.NewShardBootstrap(args)
	require.Nil(t, err)
	gotMbsAndHashes, _ := bs.GetMiniBlocks(requestedHash)

	assert.Equal(t, 0, len(gotMbsAndHashes))
}

func TestBootstrap_GetTxBodyHavingHashFoundInStorageShouldWork(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	mbh := []byte("requested hash")
	requestedHash := make([][]byte, 0)
	requestedHash = append(requestedHash, mbh)
	mbsAndHashes := make([]*block.MiniblockAndHash, 0)

	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {},
	})

	args.ChainHandler = blkc
	args.Store = createFullStore()
	args.MiniblocksProvider = &mock.MiniBlocksProviderStub{
		GetMiniBlocksCalled: func(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
			for _, hash := range hashes {
				if bytes.Equal(hash, mbh) {
					return mbsAndHashes, nil
				}
			}

			return nil, nil
		},
	}

	bs, err := sync.NewShardBootstrap(args)
	require.Nil(t, err)
	gotMbsAndHashes, _ := bs.GetMiniBlocks(requestedHash)

	assert.Equal(t, mbsAndHashes, gotMbsAndHashes)
}

func TestBootstrap_AddSyncStateListenerShouldAppendAnotherListener(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.BlockProcessor = createBlockProcessor(args.ChainHandler)

	bs, _ := sync.NewShardBootstrap(args)
	f1 := func(bool) {}
	f2 := func(bool) {}
	f3 := func(bool) {}
	bs.AddSyncStateListener(f1)
	bs.AddSyncStateListener(f2)
	bs.AddSyncStateListener(f3)

	assert.Equal(t, 3, len(bs.SyncStateListeners()))
}

func TestBootstrap_NotifySyncStateListenersShouldNotify(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	args.BlockProcessor = createBlockProcessor(args.ChainHandler)

	bs, _ := sync.NewShardBootstrap(args)

	mutex := goSync.RWMutex{}
	mutex.RLock()
	calls := 0
	mutex.RUnlock()
	var wg goSync.WaitGroup

	f1 := func(bool) {
		mutex.Lock()
		calls++
		mutex.Unlock()
		wg.Done()
	}

	f2 := func(bool) {
		mutex.Lock()
		calls++
		mutex.Unlock()
		wg.Done()
	}

	f3 := func(bool) {
		mutex.Lock()
		calls++
		mutex.Unlock()
		wg.Done()
	}

	wg.Add(3)

	bs.AddSyncStateListener(f1)
	bs.AddSyncStateListener(f2)
	bs.AddSyncStateListener(f3)

	bs.NotifySyncStateListeners()

	wg.Wait()

	assert.Equal(t, 3, calls)
}

func TestShardBootstrap_RequestMiniBlocksFromHeaderWithNonceIfMissing(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	requestDataWasCalled := false
	hdrHash := []byte("hash")
	hdr := &block.Header{Round: 5, Nonce: 1}
	pools := dataRetrieverMock.NewPoolsHolderStub()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		sds := &mock.HeadersCacherStub{}
		sds.RegisterHandlerCalled = func(func(header data.HeaderHandler, key []byte)) {
		}
		sds.GetHeaderByNonceAndShardIdCalled = func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
			return []data.HeaderHandler{hdr}, [][]byte{[]byte("hash")}, nil
		}
		sds.GetHeaderByHashCalled = func(hash []byte) (handler data.HeaderHandler, err error) {
			if bytes.Equal(hash, hdrHash) {
				return hdr, nil
			}
			return nil, nil
		}

		return sds
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := cache.NewCacherStub()
		cs.RegisterHandlerCalled = func(i func(key []byte, value interface{})) {
		}

		return cs
	}
	pools.ProofsCalled = func() dataRetriever.ProofsPool {
		return &dataRetrieverMock.ProofsPoolMock{}
	}
	args.PoolsHolder = pools

	blkc := initBlockchain()
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.Header{Round: 10}
	}
	args.ChainHandler = blkc
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return uint64(5)
	}
	args.ForkDetector = forkDetector

	store := createStore()
	store.GetCalled = func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
		nonceToBytes := mock.NewNonceHashConverterMock().ToByteSlice(uint64(1))
		if bytes.Equal(key, nonceToBytes) {
			return []byte("hdr"), nil
		}
		if bytes.Equal(key, []byte("hdr")) {
			newHdr := block.Header{}
			mshlzdHdr, _ := json.Marshal(newHdr)
			return mshlzdHdr, nil
		}

		return nil, nil
	}
	store.GetAllCalled = func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
		mapToRet := make(map[string][]byte)
		mb := block.MiniBlock{ReceiverShardID: 1, SenderShardID: 0}
		mshlzdMb, _ := json.Marshal(mb)
		mapToRet["mb1"] = mshlzdMb
		return mapToRet, nil
	}
	args.Store = store
	args.RequestHandler = &testscommon.RequestHandlerStub{
		RequestMiniBlocksHandlerCalled: func(destShardID uint32, miniblocksHashes [][]byte) {
			requestDataWasCalled = true
		},
	}
	args.MiniblocksProvider = &mock.MiniBlocksProviderStub{
		GetMiniBlocksFromPoolCalled: func(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
			return make([]*block.MiniblockAndHash, 0), [][]byte{[]byte("hash")}
		},
	}

	bs, _ := sync.NewShardBootstrap(args)
	bs.RequestMiniBlocksFromHeaderWithNonceIfMissing(hdr)

	assert.True(t, requestDataWasCalled)
}

func TestShardBootstrap_DoJobOnSyncBlockFailShouldSkipWhenBlockProcessorBusy(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	forkDetectorMock := &mock.ForkDetectorMock{
		ResetProbableHighestNonceCalled: func() {
			require.Fail(t, "should not have called ResetProbableHighestNonce")
		},
	}
	args.ForkDetector = forkDetectorMock
	args.ChainHandler = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 1}
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}

	bs, _ := sync.NewShardBootstrap(args)

	initialSyncErrors := bs.GetNumSyncedWithErrorsForNonce(2)

	bs.DoJobOnSyncBlockFail(&block.Body{}, &block.Header{Nonce: 2}, process.ErrBlockProcessorBusy)

	afterSyncErrors := bs.GetNumSyncedWithErrorsForNonce(2)
	assert.Equal(t, initialSyncErrors, afterSyncErrors, "sync error counter should not be incremented for busy processor")
}

func TestShardBootstrap_DoJobOnSyncBlockFailShouldNotResetProbableHighestNonceWhenAreNotEnoughErrorsPerNonce(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	wasCalled := false
	forkDetectorMock := &mock.ForkDetectorMock{
		ResetProbableHighestNonceCalled: func() {
			wasCalled = true
		},
	}
	args.ForkDetector = forkDetectorMock
	args.ChainHandler = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 1}
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}

	bs, _ := sync.NewShardBootstrap(args)
	bs.SetNumSyncedWithErrorsForNonce(2, 8)
	bs.DoJobOnSyncBlockFail(nil, nil, errors.New("error"))

	assert.False(t, wasCalled)
}

func TestShardBootstrap_DoJobOnSyncBlockFailShouldNotResetProbableHighestNonceWhenIsNotInProperRound(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	wasCalled := false
	forkDetectorMock := &mock.ForkDetectorMock{
		ResetProbableHighestNonceCalled: func() {
			wasCalled = true
		},
	}
	args.ForkDetector = forkDetectorMock
	args.ChainHandler = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 1}
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}

	roundHandlerMock := &mock.RoundHandlerMock{}
	roundHandlerMock.RoundIndex = 1
	args.RoundHandler = roundHandlerMock

	bs, _ := sync.NewShardBootstrap(args)
	bs.SetNumSyncedWithErrorsForNonce(2, 9)
	bs.DoJobOnSyncBlockFail(nil, nil, errors.New("error"))

	assert.False(t, wasCalled)
}

func TestShardBootstrap_DoJobOnSyncBlockFailRemovesBlockingUnprovenHeader(t *testing.T) {
	t.Parallel()

	nextNonce := uint64(2)
	hashY := []byte("hashY")
	headerY := &block.Header{Nonce: nextNonce, Round: 5}

	type capture struct {
		removedFromPool         []byte
		removedFromForkDetector bool
		removedNonce            uint64
	}

	buildArgs := func(headerHasProof bool) (sync.ArgShardBootstrapper, *capture) {
		c := &capture{}
		args := CreateShardBootstrapMockArguments()

		pools := createMockPools()
		pools.HeadersCalled = func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{
				GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
					if hdrNonce == nextNonce {
						return []data.HeaderHandler{headerY}, [][]byte{hashY}, nil
					}
					return nil, nil, errors.New("missing header")
				},
				RemoveHeaderByHashCalled: func(headerHash []byte) {
					c.removedFromPool = headerHash
				},
			}
		}
		pools.ProofsCalled = func() dataRetriever.ProofsPool {
			return &dataRetrieverMock.ProofsPoolMock{
				HasProofCalled: func(shardID uint32, headerHash []byte) bool {
					return headerHasProof
				},
			}
		}
		args.PoolsHolder = pools

		args.ChainHandler = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler { return &block.Header{Nonce: 1} },
			GetGenesisHeaderCalled:      func() data.HeaderHandler { return &block.Header{} },
		}
		args.ForkDetector = &mock.ForkDetectorMock{
			RemoveHeaderCalled: func(nonce uint64, hash []byte) {
				c.removedFromForkDetector = true
				c.removedNonce = nonce
			},
		}
		// not a proper round -> isolate the targeted removal from the rollback path
		roundHandlerMock := &mock.RoundHandlerMock{}
		roundHandlerMock.RoundIndex = 1
		args.RoundHandler = roundHandlerMock
		// proofs flag active -> the cached header genuinely needs a proof
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return true
			},
		}

		return args, c
	}

	t.Run("removes the unproven blocking header and resets the error counter when the limit is reached", func(t *testing.T) {
		t.Parallel()

		args, c := buildArgs(false)
		bs, _ := sync.NewShardBootstrap(args)
		bs.SetNumSyncedWithErrorsForNonce(nextNonce, 100)

		bs.DoJobOnSyncBlockFail(nil, nil, process.ErrTimeIsOut)

		assert.Equal(t, hashY, c.removedFromPool)
		assert.True(t, c.removedFromForkDetector)
		assert.Equal(t, nextNonce, c.removedNonce)
		assert.Equal(t, uint32(0), bs.GetNumSyncedWithErrorsForNonce(nextNonce))
	})

	t.Run("does not remove a header that has a proof", func(t *testing.T) {
		t.Parallel()

		args, c := buildArgs(true)
		bs, _ := sync.NewShardBootstrap(args)
		bs.SetNumSyncedWithErrorsForNonce(nextNonce, 100)

		bs.DoJobOnSyncBlockFail(nil, nil, process.ErrTimeIsOut)

		assert.Nil(t, c.removedFromPool)
		assert.False(t, c.removedFromForkDetector)
	})

	t.Run("does not remove before the error limit is reached", func(t *testing.T) {
		t.Parallel()

		args, c := buildArgs(false)
		bs, _ := sync.NewShardBootstrap(args)
		bs.SetNumSyncedWithErrorsForNonce(nextNonce, 0)

		bs.DoJobOnSyncBlockFail(nil, nil, process.ErrTimeIsOut)

		assert.Nil(t, c.removedFromPool)
		assert.False(t, c.removedFromForkDetector)
	})
}

func TestShardBootstrap_DoJobOnSyncBlockFailExecutionResultsMismatchRecovery(t *testing.T) {
	t.Parallel()

	createExecResult := func(nonce uint64, rootHash []byte) *block.ExecutionResult {
		return &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte(fmt.Sprintf("hash%d", nonce)),
				HeaderNonce: nonce,
				RootHash:    rootHash,
			},
		}
	}

	currentBlockNonce := uint64(5)
	syncedHeaderNonce := currentBlockNonce + 1

	createChainHandler := func() data.ChainHandler {
		return &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.HeaderV3{Nonce: currentBlockNonce}
			},
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{}
			},
		}
	}

	createSyncedHeader := func(execResults ...*block.ExecutionResult) *block.HeaderV3 {
		return &block.HeaderV3{
			Nonce:            syncedHeaderNonce,
			ExecutionResults: execResults,
		}
	}

	t.Run("mismatch at first result triggers recovery with the diverging nonce", func(t *testing.T) {
		t.Parallel()

		args := CreateShardBootstrapMockArguments()
		args.ChainHandler = createChainHandler()

		removedNonces := make([]uint64, 0)
		args.ExecutionManager = &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return []data.BaseExecutionResultHandler{createExecResult(5, []byte("localRoot"))}, nil
			},
			RemoveAtNonceAndHigherCalled: func(nonce uint64) error {
				removedNonces = append(removedNonces, nonce)
				return nil
			},
		}

		bs, _ := sync.NewShardBootstrap(args)
		bs.SetPreparedForSync(true)
		bs.SetNumSyncedWithErrorsForNonce(syncedHeaderNonce, 3)

		header := createSyncedHeader(createExecResult(5, []byte("canonicalRoot")))
		bs.DoJobOnSyncBlockFail(nil, header, process.ErrExecutionResultDoesNotMatch)

		require.Equal(t, []uint64{5}, removedNonces)
		assert.False(t, bs.GetPreparedForSync())
		assert.Equal(t, uint32(0), bs.GetNumSyncedWithErrorsForNonce(syncedHeaderNonce))
		assert.Equal(t, uint32(1), bs.GetRecoveryAttemptsForNonce(5))
	})

	t.Run("mismatch mid-list rewinds to the first diverging nonce only", func(t *testing.T) {
		t.Parallel()

		args := CreateShardBootstrapMockArguments()
		args.ChainHandler = createChainHandler()

		removedNonces := make([]uint64, 0)
		args.ExecutionManager = &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return []data.BaseExecutionResultHandler{
					createExecResult(3, []byte("rootA")),
					createExecResult(4, []byte("rootB")),
					createExecResult(5, []byte("localRoot")),
				}, nil
			},
			RemoveAtNonceAndHigherCalled: func(nonce uint64) error {
				removedNonces = append(removedNonces, nonce)
				return nil
			},
		}

		bs, _ := sync.NewShardBootstrap(args)
		bs.SetPreparedForSync(true)

		header := createSyncedHeader(
			createExecResult(3, []byte("rootA")),
			createExecResult(4, []byte("rootB")),
			createExecResult(5, []byte("canonicalRoot")),
		)
		bs.DoJobOnSyncBlockFail(nil, header, process.ErrExecutionResultDoesNotMatch)

		require.Equal(t, []uint64{5}, removedNonces)
		assert.False(t, bs.GetPreparedForSync())
	})

	t.Run("number mismatch error does not trigger recovery", func(t *testing.T) {
		t.Parallel()

		args := CreateShardBootstrapMockArguments()
		args.ChainHandler = createChainHandler()
		args.ExecutionManager = &processMocks.ExecutionManagerMock{
			RemoveAtNonceAndHigherCalled: func(nonce uint64) error {
				require.Fail(t, "should not have called RemoveAtNonceAndHigher")
				return nil
			},
		}

		bs, _ := sync.NewShardBootstrap(args)
		bs.SetPreparedForSync(true)

		header := createSyncedHeader(createExecResult(5, []byte("canonicalRoot")))
		bs.DoJobOnSyncBlockFail(nil, header, process.ErrExecutionResultsNumberMismatch)

		assert.True(t, bs.GetPreparedForSync())
		assert.Equal(t, uint32(1), bs.GetNumSyncedWithErrorsForNonce(syncedHeaderNonce))
	})

	t.Run("non-V3 header does not trigger recovery", func(t *testing.T) {
		t.Parallel()

		args := CreateShardBootstrapMockArguments()
		args.ChainHandler = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.Header{Nonce: currentBlockNonce}
			},
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{}
			},
		}
		args.ExecutionManager = &processMocks.ExecutionManagerMock{
			RemoveAtNonceAndHigherCalled: func(nonce uint64) error {
				require.Fail(t, "should not have called RemoveAtNonceAndHigher")
				return nil
			},
		}

		bs, _ := sync.NewShardBootstrap(args)
		bs.SetPreparedForSync(true)

		bs.DoJobOnSyncBlockFail(nil, &block.Header{Nonce: syncedHeaderNonce}, process.ErrExecutionResultDoesNotMatch)

		assert.True(t, bs.GetPreparedForSync())
		assert.Equal(t, uint32(1), bs.GetNumSyncedWithErrorsForNonce(syncedHeaderNonce))
	})

	t.Run("no nonce-matched divergence falls back to last notarized + 1", func(t *testing.T) {
		t.Parallel()

		args := CreateShardBootstrapMockArguments()
		args.ChainHandler = createChainHandler()

		removedNonces := make([]uint64, 0)
		args.ExecutionManager = &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				// pending results match the header ones, the mismatch is not pending-related
				return []data.BaseExecutionResultHandler{createExecResult(5, []byte("canonicalRoot"))}, nil
			},
			GetLastNotarizedExecutionResultCalled: func() (data.BaseExecutionResultHandler, error) {
				return createExecResult(2, []byte("notarizedRoot")), nil
			},
			RemoveAtNonceAndHigherCalled: func(nonce uint64) error {
				removedNonces = append(removedNonces, nonce)
				return nil
			},
		}

		bs, _ := sync.NewShardBootstrap(args)
		bs.SetPreparedForSync(true)

		header := createSyncedHeader(createExecResult(5, []byte("canonicalRoot")))
		bs.DoJobOnSyncBlockFail(nil, header, process.ErrExecutionResultDoesNotMatch)

		require.Equal(t, []uint64{3}, removedNonces)
		assert.False(t, bs.GetPreparedForSync())
	})

	t.Run("pending fetch error falls back to last notarized + 1", func(t *testing.T) {
		t.Parallel()

		args := CreateShardBootstrapMockArguments()
		args.ChainHandler = createChainHandler()

		removedNonces := make([]uint64, 0)
		args.ExecutionManager = &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return nil, errors.New("pending fetch error")
			},
			GetLastNotarizedExecutionResultCalled: func() (data.BaseExecutionResultHandler, error) {
				return createExecResult(2, []byte("notarizedRoot")), nil
			},
			RemoveAtNonceAndHigherCalled: func(nonce uint64) error {
				removedNonces = append(removedNonces, nonce)
				return nil
			},
		}

		bs, _ := sync.NewShardBootstrap(args)
		bs.SetPreparedForSync(true)

		header := createSyncedHeader(createExecResult(5, []byte("canonicalRoot")))
		bs.DoJobOnSyncBlockFail(nil, header, process.ErrExecutionResultDoesNotMatch)

		require.Equal(t, []uint64{3}, removedNonces)
		assert.False(t, bs.GetPreparedForSync())
	})

	t.Run("last notarized fetch error skips recovery", func(t *testing.T) {
		t.Parallel()

		args := CreateShardBootstrapMockArguments()
		args.ChainHandler = createChainHandler()
		args.ExecutionManager = &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return nil, errors.New("pending fetch error")
			},
			GetLastNotarizedExecutionResultCalled: func() (data.BaseExecutionResultHandler, error) {
				return nil, errors.New("notarized fetch error")
			},
			RemoveAtNonceAndHigherCalled: func(nonce uint64) error {
				require.Fail(t, "should not have called RemoveAtNonceAndHigher")
				return nil
			},
		}

		bs, _ := sync.NewShardBootstrap(args)
		bs.SetPreparedForSync(true)

		header := createSyncedHeader(createExecResult(5, []byte("canonicalRoot")))
		bs.DoJobOnSyncBlockFail(nil, header, process.ErrExecutionResultDoesNotMatch)

		assert.True(t, bs.GetPreparedForSync())
		assert.Equal(t, uint32(1), bs.GetNumSyncedWithErrorsForNonce(syncedHeaderNonce))
	})

	t.Run("cooldown limits recovery to one attempt until it expires", func(t *testing.T) {
		t.Parallel()

		args := CreateShardBootstrapMockArguments()
		args.ChainHandler = createChainHandler()

		removedNonces := make([]uint64, 0)
		args.ExecutionManager = &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return []data.BaseExecutionResultHandler{createExecResult(5, []byte("localRoot"))}, nil
			},
			RemoveAtNonceAndHigherCalled: func(nonce uint64) error {
				removedNonces = append(removedNonces, nonce)
				return nil
			},
		}

		bs, _ := sync.NewShardBootstrap(args)

		header := createSyncedHeader(createExecResult(5, []byte("canonicalRoot")))

		bs.SetPreparedForSync(true)
		bs.DoJobOnSyncBlockFail(nil, header, process.ErrExecutionResultDoesNotMatch)
		require.Equal(t, []uint64{5}, removedNonces)
		assert.False(t, bs.GetPreparedForSync())

		// second occurrence within the cooldown window does not recover again
		bs.SetPreparedForSync(true)
		bs.DoJobOnSyncBlockFail(nil, header, process.ErrExecutionResultDoesNotMatch)
		require.Equal(t, []uint64{5}, removedNonces)
		assert.True(t, bs.GetPreparedForSync())
		assert.Equal(t, uint32(1), bs.GetRecoveryAttemptsForNonce(5))

		// once the cooldown expires, recovery is re-attempted
		bs.SetExecutionResultsRecoveryCooldown(0)
		bs.SetPreparedForSync(true)
		bs.DoJobOnSyncBlockFail(nil, header, process.ErrExecutionResultDoesNotMatch)
		require.Equal(t, []uint64{5, 5}, removedNonces)
		assert.False(t, bs.GetPreparedForSync())
		assert.Equal(t, uint32(2), bs.GetRecoveryAttemptsForNonce(5))
	})

	t.Run("RemoveAtNonceAndHigher failure leaves preparedForSync untouched", func(t *testing.T) {
		t.Parallel()

		args := CreateShardBootstrapMockArguments()
		args.ChainHandler = createChainHandler()
		args.ExecutionManager = &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return []data.BaseExecutionResultHandler{createExecResult(5, []byte("localRoot"))}, nil
			},
			RemoveAtNonceAndHigherCalled: func(nonce uint64) error {
				return errors.New("remove error")
			},
		}

		bs, _ := sync.NewShardBootstrap(args)
		bs.SetPreparedForSync(true)

		header := createSyncedHeader(createExecResult(5, []byte("canonicalRoot")))
		bs.DoJobOnSyncBlockFail(nil, header, process.ErrExecutionResultDoesNotMatch)

		assert.True(t, bs.GetPreparedForSync())
		assert.Equal(t, uint32(1), bs.GetNumSyncedWithErrorsForNonce(syncedHeaderNonce))
	})
}

func TestShardBootstrap_DoJobOnSyncBlockFailShouldResetProbableHighestNonce(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	wasCalled := false
	forkDetectorMock := &mock.ForkDetectorMock{
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 1
		},
		GetHighestFinalBlockHashCalled: func() []byte {
			return []byte("hash")
		},
		ResetProbableHighestNonceCalled: func() {
			wasCalled = true
		},
	}
	args.ForkDetector = forkDetectorMock
	args.ChainHandler = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 2}
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}

	// revert final block should not reset probable highest nonce
	bs, _ := sync.NewShardBootstrap(args)
	bs.SetNumSyncedWithErrorsForNonce(2, 9)
	bs.DoJobOnSyncBlockFail(nil, nil, errors.New("error"))
	assert.False(t, wasCalled)

	// revert non final block should reset probable highest nonce
	bs.SetNumSyncedWithErrorsForNonce(3, 9)
	bs.DoJobOnSyncBlockFail(nil, nil, errors.New("error"))

	assert.True(t, wasCalled)
}

func TestShardBootstrap_CleanNoncesSyncedWithErrorsBehindFinalShouldWork(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	forkDetectorMock := &mock.ForkDetectorMock{
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 3
		},
	}
	args.ForkDetector = forkDetectorMock

	bs, _ := sync.NewShardBootstrap(args)
	bs.SetNumSyncedWithErrorsForNonce(1, 7)
	bs.SetNumSyncedWithErrorsForNonce(2, 8)
	bs.SetNumSyncedWithErrorsForNonce(3, 9)

	assert.Equal(t, 3, bs.GetMapNonceSyncedWithErrorsLen())

	bs.CleanNoncesSyncedWithErrorsBehindFinal()

	assert.Equal(t, 1, bs.GetMapNonceSyncedWithErrorsLen())
	assert.Equal(t, uint32(9), bs.GetNumSyncedWithErrorsForNonce(3))
}

func TestShardBootstrap_SyncBlockGetNodeDBErrorShouldSync(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	hdr := block.Header{Nonce: 1, PubKeysBitmap: []byte("X")}
	blkc := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &hdr
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	args.ChainHandler = blkc

	errGetNodeFromDB := core.NewGetNodeFromDBErrWithKey([]byte("key"), errors.New("get error"), "")
	blockProcessor := createBlockProcessor(args.ChainHandler)
	blockProcessor.ProcessBlockCalled = func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
		return errGetNodeFromDB
	}
	args.BlockProcessor = blockProcessor

	hash := []byte("aaa")
	header := &block.Header{
		Nonce:         2,
		Round:         1,
		BlockBodyType: block.TxBlock,
		RootHash:      []byte("bbb")}

	pools := dataRetrieverMock.NewPoolsHolderStub()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		sds := &mock.HeadersCacherStub{}
		sds.GetHeaderByNonceAndShardIdCalled = func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
			if hdrNonce == 2 {
				return []data.HeaderHandler{header}, [][]byte{hash}, nil
			}
			return nil, nil, errors.New("err")
		}

		return sds
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := cache.NewCacherStub()
		cs.RegisterHandlerCalled = func(i func(key []byte, value interface{})) {
		}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("bbb"), key) {
				return make(block.MiniBlockSlice, 0), true
			}

			return nil, false
		}

		return cs
	}
	pools.ProofsCalled = func() dataRetriever.ProofsPool {
		return &dataRetrieverMock.ProofsPoolMock{}
	}
	args.PoolsHolder = pools

	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return hdr.Nonce
	}
	forkDetector.GetHighestFinalBlockHashCalled = func() []byte {
		return []byte("hash")
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 2
	}
	forkDetector.RemoveHeaderCalled = func(nonce uint64, hash []byte) {}
	forkDetector.GetNotarizedHeaderHashCalled = func(nonce uint64) []byte {
		return nil
	}
	args.ForkDetector = forkDetector
	args.RoundHandler, _ = round.NewRound(createDefaultRoundArgs())

	syncCalled := false
	args.AccountsDBSyncer = &mock.AccountsDBSyncerStub{
		SyncAccountsCalled: func(rootHash []byte, _ common.StorageMarker) error {
			syncCalled = true
			return nil
		}}
	args.Accounts = &stateMock.AccountsStub{RootHashCalled: func() ([]byte, error) {
		return []byte("roothash"), nil
	}}

	bs, err := sync.NewShardBootstrap(args)
	require.Nil(t, err)

	err = bs.SyncBlock(context.Background())
	assert.Equal(t, errGetNodeFromDB, err)
	assert.True(t, syncCalled)
}

func TestShardBootstrap_SyncBlock_WithEquivalentProofs(t *testing.T) {
	t.Parallel()

	t.Run("time is out when existing header and missing proof", func(t *testing.T) {
		t.Parallel()

		args := CreateShardBootstrapMockArguments()

		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AndromedaFlag
			},
		}

		hdr := block.Header{Nonce: 1}
		blkc := &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{}
			},
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &hdr
			},
		}
		args.ChainHandler = blkc

		forkDetector := &mock.ForkDetectorMock{}
		forkDetector.CheckForkCalled = func() *process.ForkInfo {
			return process.NewForkInfo()
		}
		forkDetector.ProbableHighestNonceCalled = func() uint64 {
			return 100
		}
		forkDetector.GetNotarizedHeaderHashCalled = func(nonce uint64) []byte {
			return nil
		}
		args.ForkDetector = forkDetector
		args.RoundHandler, _ = round.NewRound(createDefaultRoundArgs())
		args.BlockProcessor = createBlockProcessor(args.ChainHandler)

		pools := createMockPools()
		pools.ProofsCalled = func() dataRetriever.ProofsPool {
			return &dataRetrieverMock.ProofsPoolMock{
				GetProofByNonceCalled: func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
					return nil, errors.New("missing proof")
				},
			}
		}

		args.PoolsHolder = pools

		bs, _ := sync.NewShardBootstrap(args)
		r := bs.SyncBlock(context.Background())

		assert.Equal(t, process.ErrTimeIsOut, r)
	})

	t.Run("should receive header and proof if missing, requesting by nonce", func(t *testing.T) {
		t.Parallel()

		args := CreateShardBootstrapMockArguments()

		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}

		hdr := block.Header{Nonce: 1}
		blkc := &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{}
			},
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &hdr
			},
		}
		args.ChainHandler = blkc

		forkDetector := &mock.ForkDetectorMock{}
		forkDetector.CheckForkCalled = func() *process.ForkInfo {
			return process.NewForkInfo()
		}
		forkDetector.ProbableHighestNonceCalled = func() uint64 {
			return 100
		}
		forkDetector.GetNotarizedHeaderHashCalled = func(nonce uint64) []byte {
			return nil
		}
		args.ForkDetector = forkDetector
		args.RoundHandler, _ = round.NewRound(createDefaultRoundArgs())
		args.BlockProcessor = createBlockProcessor(args.ChainHandler)

		pools := createMockPools()
		pools.ProofsCalled = func() dataRetriever.ProofsPool {
			return &dataRetrieverMock.ProofsPoolMock{
				GetProofByNonceCalled: func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
					return nil, errors.New("missing proof")
				},
				HasProofCalled: func(shardID uint32, headerHash []byte) bool {
					return true // second check after wait is done by hash
				},
			}
		}

		var numHeaderCalls atomic.Uint64
		pools.HeadersCalled = func() dataRetriever.HeadersPool {
			sds := &mock.HeadersCacherStub{}
			sds.GetHeaderByNonceAndShardIdCalled = func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
				if numHeaderCalls.Load() == 0 {
					numHeaderCalls.Add(1)
					return nil, nil, errors.New("err")
				}

				return []data.HeaderHandler{
					&block.Header{
						Nonce:    1,
						Round:    1,
						RootHash: []byte("bbb")},
				}, [][]byte{[]byte("aaa")}, nil
			}

			return sds
		}
		args.PoolsHolder = pools

		receive := make(chan bool, 2)

		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestShardHeaderByNonceCalled: func(shardID uint32, nonce uint64) {
				receive <- true
			},
			RequestEquivalentProofByNonceCalled: func(headerShard uint32, headerNonce uint64) {
				receive <- true
			},
		}

		bs, _ := sync.NewShardBootstrap(args)

		go func() {
			// wait for both header and proof requests
			<-receive
			<-receive

			bs.SetRcvHdrNonce()
		}()

		err := bs.SyncBlock(context.Background())

		assert.Nil(t, err)
	})

	t.Run("should receive header and proof if missing, requesting by hash", func(t *testing.T) {
		t.Parallel()

		args := CreateShardBootstrapMockArguments()

		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AndromedaFlag
			},
		}

		hdr := block.Header{Nonce: 1}
		blkc := &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{}
			},
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &hdr
			},
		}
		args.ChainHandler = blkc

		forkDetector := &mock.ForkDetectorMock{}
		forkDetector.CheckForkCalled = func() *process.ForkInfo {
			return process.NewForkInfo()
		}
		forkDetector.ProbableHighestNonceCalled = func() uint64 {
			return 100
		}

		hash := []byte("hash1")
		forkDetector.GetNotarizedHeaderHashCalled = func(nonce uint64) []byte {
			return hash
		}
		args.ForkDetector = forkDetector
		args.RoundHandler, _ = round.NewRound(createDefaultRoundArgs())
		args.BlockProcessor = createBlockProcessor(args.ChainHandler)

		pools := createMockPools()

		numProofCalls := 0
		pools.ProofsCalled = func() dataRetriever.ProofsPool {
			return &dataRetrieverMock.ProofsPoolMock{
				GetProofCalled: func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
					return nil, errors.New("missing proof")
				},
				GetProofByNonceCalled: func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
					return nil, errors.New("missing proof")
				},
				HasProofCalled: func(shardID uint32, headerHash []byte) bool {
					if numProofCalls == 0 {
						numProofCalls++
						return false
					}

					return true // second check after wait is done by hash
				},
			}
		}

		numHeaderCalls := 0
		pools.HeadersCalled = func() dataRetriever.HeadersPool {
			sds := &mock.HeadersCacherStub{}

			sds.GetHeaderByHashCalled = func(hash []byte) (data.HeaderHandler, error) {
				if numHeaderCalls == 0 {
					numHeaderCalls++
					return nil, errors.New("err")
				}

				return &block.Header{}, nil
			}

			return sds
		}
		args.PoolsHolder = pools

		receive := make(chan bool, 2)

		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestShardHeaderCalled: func(shardID uint32, hash []byte) {
				receive <- true
			},
			RequestEquivalentProofByHashCalled: func(headerShard uint32, headerHash []byte) {
				receive <- true
			},
		}

		bs, _ := sync.NewShardBootstrap(args)

		go func() {
			// wait for both header and proof requests
			<-receive
			<-receive

			bs.SetRcvHdrHash()
		}()

		err := bs.SyncBlock(context.Background())

		assert.Nil(t, err)
	})
}

func TestShardBootstrap_NilInnerBootstrapperClose(t *testing.T) {
	t.Parallel()

	bootstrapper := &sync.ShardBootstrap{}
	assert.Nil(t, bootstrapper.Close())
}

func TestShardBootstrap_SyncBlockV3(t *testing.T) {
	t.Parallel()

	createSyncBlockV3Args := func() sync.ArgShardBootstrapper {
		args := CreateShardBootstrapMockArguments()

		args.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, round uint64) bool {
				return flag == common.SupernovaRoundFlag
			},
		}

		hdr := &block.HeaderV3{
			Nonce: 2,
			LastExecutionResult: &block.ExecutionResultInfo{
				NotarizedInRound: 1,
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 1,
				},
			}}
		args.ChainHandler = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return hdr
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash")
			},
		}

		header := &block.HeaderV3{
			Nonce:         3,
			Round:         1,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				NotarizedInRound: 2,
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 2,
				},
			},
		}
		args.PoolsHolder = setupPools(headerAndHash{
			header: header,
			hash:   []byte("aaa"),
		})

		prevHdr := &block.HeaderV3{Nonce: 1}
		args.Store = setupStore(args.Marshalizer, prevHdr, nil)
		args.ForkDetector = setupForkDetector(3)

		return args
	}

	t.Run("should work on the ideal case(subsequent nonces)", func(t *testing.T) {
		t.Parallel()

		args := createSyncBlockV3Args()

		verifyBlockProposalCalled := false
		commitBlockCalled := false
		args.BlockProcessor = &testscommon.BlockProcessorStub{
			VerifyBlockProposalCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				verifyBlockProposalCalled = true
				return nil
			},
			CommitBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
				commitBlockCalled = true
				return nil
			},
		}

		addToQueueCalled := false
		args.ExecutionManager = &processMocks.ExecutionManagerMock{
			AddPairForExecutionCalled: func(pair headersCache.HeaderBodyPair) error {
				addToQueueCalled = true
				return nil
			},
		}

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Nil(t, err)
		assert.True(t, verifyBlockProposalCalled)
		assert.True(t, commitBlockCalled)
		assert.True(t, addToQueueCalled)
	})

	t.Run("should work and prepare the tx pool with multiple blocks", func(t *testing.T) {
		t.Parallel()

		// test details:
		// current header nonce: 4
		// fork detector highest nonce: 5 -> will sync nonce 5
		// current header holds last execution result with nonce 2
		// so when syncing header 5, we expect to prepare the pool
		// with nonces 3, 4 (backfill via backward hash-walk) + nonce 5 (synced block)
		args := createSyncBlockV3Args()

		hash1 := []byte("hash1")
		hash2 := []byte("hash2")
		hash3 := []byte("hash3")
		hash4 := []byte("hash4")
		hash5 := []byte("hash5")

		header2 := &block.HeaderV3{
			Nonce:         2,
			PrevHash:      hash1,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 1,
					HeaderHash:  hash1,
				},
			},
		}
		header3 := &block.HeaderV3{
			Nonce:         3,
			PrevHash:      hash2,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 2,
					HeaderHash:  hash2,
				},
			},
		}
		header4 := &block.HeaderV3{
			Nonce:         4,
			PrevHash:      hash3,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 2,
					HeaderHash:  hash2,
				},
			},
		}
		header5 := &block.HeaderV3{
			Nonce:         5,
			PrevHash:      hash4,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 4,
					HeaderHash:  hash4,
				},
			},
		}
		args.PoolsHolder = setupPoolsDirectHashMapping(
			headerAndHash{header: header2, hash: hash2},
			headerAndHash{header: header3, hash: hash3},
			headerAndHash{header: header4, hash: hash4},
			headerAndHash{header: header5, hash: hash5},
		)
		args.ChainHandler = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return header4 // forcing to sync nonce 5
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return hash4
			},
		}

		args.Store = setupStore(args.Marshalizer, header2, nil)
		args.ForkDetector = setupForkDetector(5)

		verifyBlockProposalCalled := false
		commitBlockCalled := false
		args.BlockProcessor = &testscommon.BlockProcessorStub{
			VerifyBlockProposalCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				verifyBlockProposalCalled = true
				return nil
			},
			CommitBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
				commitBlockCalled = true
				return nil
			},
		}

		cntAddToQueue := 0
		args.ExecutionManager = &processMocks.ExecutionManagerMock{
			AddPairForExecutionCalled: func(pair headersCache.HeaderBodyPair) error {
				cntAddToQueue++
				return nil
			},
		}

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Nil(t, err)
		assert.True(t, verifyBlockProposalCalled)
		assert.True(t, commitBlockCalled)
		assert.Equal(t, 3, cntAddToQueue) // two backfilled (nonces 3,4) + one synced (nonce 5)
	})

	t.Run("should error when GetPrevBlockLastExecutionResult fails", func(t *testing.T) {
		t.Parallel()

		args := createSyncBlockV3Args()
		chainHandler, ok := args.ChainHandler.(*testscommon.ChainHandlerStub)
		require.True(t, ok)
		chainHandler.GetCurrentBlockHeaderHashCalled = func() []byte {
			return nil
		}

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Error(t, err)
	})

	t.Run("should error when VerifyBlockProposal fails", func(t *testing.T) {
		t.Parallel()

		args := createSyncBlockV3Args()
		blockProcessor := &testscommon.BlockProcessorStub{
			VerifyBlockProposalCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				return errExpected
			},
		}
		args.BlockProcessor = blockProcessor

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Equal(t, errExpected, err)
	})

	t.Run("should error when OnExecutedBlock fails on the ideal case", func(t *testing.T) {
		t.Parallel()

		args := createSyncBlockV3Args()
		poolsStub, ok := args.PoolsHolder.(*dataRetrieverMock.PoolsHolderStub)
		require.True(t, ok)
		resetTrackerCalled := false
		poolsStub.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				OnExecutedBlockCalled: func(blockHeader data.HeaderHandler, rootHash []byte) error {
					return errExpected
				},
				ResetTrackerCalled: func() {
					resetTrackerCalled = true
				},
			}
		}
		args.PoolsHolder = poolsStub

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Equal(t, errExpected, err)
		assert.True(t, resetTrackerCalled)
	})

	t.Run("should error when OnProposedBlock fails on the ideal case", func(t *testing.T) {
		t.Parallel()

		args := createSyncBlockV3Args()
		blockProcessor := &testscommon.BlockProcessorStub{
			OnBackfilledBlockCalled: func(proposedBody data.BodyHandler, proposedHeader data.HeaderHandler, proposedHash []byte) error {
				return errExpected
			},
		}
		args.BlockProcessor = blockProcessor

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Equal(t, errExpected, err)
	})

	t.Run("should error when OnExecutedBlock fails on the bigger gap case", func(t *testing.T) {
		t.Parallel()

		// test details:
		// current header nonce: 4
		// fork detector highest nonce: 5 -> will sync nonce 5
		// current header holds last execution result with nonce 2
		args := createSyncBlockV3Args()
		header2 := &block.HeaderV3{
			Nonce:         2,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 1,
					HeaderHash:  []byte("hash1"),
				},
			},
		}
		header4 := &block.HeaderV3{
			Nonce:         4,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 2,
					HeaderHash:  []byte("hash3"),
				},
			},
		}
		header5 := &block.HeaderV3{
			Nonce:         5,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 4,
					HeaderHash:  []byte("hash4"),
				},
			},
		}
		args.PoolsHolder = setupPools(
			headerAndHash{
				header: header2,
				hash:   []byte("hash2"),
			},
			headerAndHash{
				header: header4,
				hash:   []byte("hash4"),
			},
			headerAndHash{
				header: header5,
				hash:   []byte("hash5"),
			},
		)
		poolsStub, ok := args.PoolsHolder.(*dataRetrieverMock.PoolsHolderStub)
		require.True(t, ok)
		resetTrackerCalled := false
		poolsStub.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				OnExecutedBlockCalled: func(blockHeader data.HeaderHandler, rootHash []byte) error {
					return errExpected
				},
				ResetTrackerCalled: func() {
					resetTrackerCalled = true
				},
			}
		}
		args.PoolsHolder = poolsStub
		args.ChainHandler = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return header4 // forcing to sync nonce 5
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash4")
			},
		}

		args.Store = setupStore(args.Marshalizer, header2, nil)
		args.ForkDetector = setupForkDetector(5)

		args.ExecutionManager = &processMocks.ExecutionManagerMock{
			AddPairForExecutionCalled: func(pair headersCache.HeaderBodyPair) error {
				return errExpected
			},
		}

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Equal(t, errExpected, err)
		assert.True(t, resetTrackerCalled)
	})

	t.Run("should error when AddOrReplace fails on the bigger gap case", func(t *testing.T) {
		t.Parallel()

		// test details:
		// current header nonce: 4
		// fork detector highest nonce: 5 -> will sync nonce 5
		// current header holds last execution result with nonce 2
		args := createSyncBlockV3Args()

		hash1 := []byte("hash1")
		hash2 := []byte("hash2")
		hash3 := []byte("hash3")
		hash4 := []byte("hash4")
		hash5 := []byte("hash5")

		header2 := &block.HeaderV3{
			Nonce:         2,
			PrevHash:      hash1,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 1,
					HeaderHash:  hash1,
				},
			},
		}
		header3 := &block.HeaderV3{
			Nonce:         3,
			PrevHash:      hash2,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 2,
					HeaderHash:  hash2,
				},
			},
		}
		header4 := &block.HeaderV3{
			Nonce:         4,
			PrevHash:      hash3,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 2,
					HeaderHash:  hash2,
				},
			},
		}
		header5 := &block.HeaderV3{
			Nonce:         5,
			PrevHash:      hash4,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 4,
					HeaderHash:  hash4,
				},
			},
		}
		args.PoolsHolder = setupPoolsDirectHashMapping(
			headerAndHash{header: header2, hash: hash2},
			headerAndHash{header: header3, hash: hash3},
			headerAndHash{header: header4, hash: hash4},
			headerAndHash{header: header5, hash: hash5},
		)
		args.ChainHandler = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return header4 // forcing to sync nonce 5
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return hash4
			},
		}

		args.Store = setupStore(args.Marshalizer, header2, nil)
		args.ForkDetector = setupForkDetector(5)

		args.ExecutionManager = &processMocks.ExecutionManagerMock{
			AddPairForExecutionCalled: func(pair headersCache.HeaderBodyPair) error {
				return errExpected
			},
		}

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Equal(t, errExpected, err)
	})

	t.Run("should error when OnProposedBlock fails on the bigger gap case", func(t *testing.T) {
		t.Parallel()

		// test details:
		// current header nonce: 4
		// fork detector highest nonce: 5 -> will sync nonce 5
		// current header holds last execution result with nonce 2
		args := createSyncBlockV3Args()

		hash1 := []byte("hash1")
		hash2 := []byte("hash2")
		hash3 := []byte("hash3")
		hash4 := []byte("hash4")
		hash5 := []byte("hash5")

		header2 := &block.HeaderV3{
			Nonce:         2,
			PrevHash:      hash1,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 1,
					HeaderHash:  hash1,
				},
			},
		}
		header3 := &block.HeaderV3{
			Nonce:         3,
			PrevHash:      hash2,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 2,
					HeaderHash:  hash2,
				},
			},
		}
		header4 := &block.HeaderV3{
			Nonce:         4,
			PrevHash:      hash3,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 2,
					HeaderHash:  hash2,
				},
			},
		}
		header5 := &block.HeaderV3{
			Nonce:         5,
			PrevHash:      hash4,
			BlockBodyType: block.TxBlock,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 4,
					HeaderHash:  hash4,
				},
			},
		}
		args.PoolsHolder = setupPoolsDirectHashMapping(
			headerAndHash{header: header2, hash: hash2},
			headerAndHash{header: header3, hash: hash3},
			headerAndHash{header: header4, hash: hash4},
			headerAndHash{header: header5, hash: hash5},
		)
		blockProcessor := &testscommon.BlockProcessorStub{
			OnBackfilledBlockCalled: func(proposedBody data.BodyHandler, proposedHeader data.HeaderHandler, proposedHash []byte) error {
				return errExpected
			},
		}
		args.BlockProcessor = blockProcessor
		args.ChainHandler = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return header4 // forcing to sync nonce 5
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return hash4
			},
		}

		args.Store = setupStore(args.Marshalizer, header2, nil)
		args.ForkDetector = setupForkDetector(5)

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Equal(t, errExpected, err)
	})

	t.Run("should error when CommitBlock fails", func(t *testing.T) {
		t.Parallel()

		args := createSyncBlockV3Args()
		blockProcessor := &testscommon.BlockProcessorStub{
			VerifyBlockProposalCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				return nil
			},
			CommitBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
				return errExpected
			},
		}
		args.BlockProcessor = blockProcessor

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Equal(t, errExpected, err)
	})

	t.Run("should return early when node is synchronized", func(t *testing.T) {
		t.Parallel()

		args := createSyncBlockV3Args()

		hdr := block.HeaderV3{
			Nonce: 2,
			LastExecutionResult: &block.ExecutionResultInfo{
				NotarizedInRound: 1,
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: 1,
				},
			}}
		blkc := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &hdr
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash")
			},
		}
		args.ChainHandler = blkc

		args.ForkDetector = setupForkDetector(hdr.Nonce) // synced

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Nil(t, err)
	})

	t.Run("should error when last execution result is nil", func(t *testing.T) {
		t.Parallel()

		args := createSyncBlockV3Args()

		hdr := testscommon.HeaderHandlerStub{
			IsHeaderV3Called: func() bool {
				return true
			},
			GetLastExecutionResultHandlerCalled: func() data.LastExecutionResultHandler {
				return nil
			},
		}
		blkc := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &hdr
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash")
			},
		}
		args.ChainHandler = blkc

		hash := []byte("aaa")
		header := &block.HeaderV3{
			Nonce: 2,
		}

		pools := dataRetrieverMock.NewPoolsHolderStub()
		pools.HeadersCalled = func() dataRetriever.HeadersPool {
			sds := &mock.HeadersCacherStub{}
			sds.GetHeaderByNonceAndShardIdCalled = func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
				if hdrNonce == header.Nonce {
					return []data.HeaderHandler{header}, [][]byte{hash}, nil
				}
				return nil, nil, errors.New("err")
			}

			return sds
		}
		pools.MiniBlocksCalled = func() storage.Cacher {
			cs := cache.NewCacherStub()
			cs.RegisterHandlerCalled = func(i func(key []byte, value interface{})) {}
			cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
				return make(block.MiniBlockSlice, 0), true
			}

			return cs
		}
		pools.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				OnExecutedBlockCalled: func(header data.HeaderHandler, rootHash []byte) error {
					return nil
				},
			}
		}
		pools.ProofsCalled = func() dataRetriever.ProofsPool {
			return &dataRetrieverMock.ProofsPoolMock{}
		}
		args.PoolsHolder = pools

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Equal(t, process.ErrNilLastExecutionResultHandler, err)
	})

	t.Run("should error when AddOrReplace to queue fails", func(t *testing.T) {
		t.Parallel()

		args := createSyncBlockV3Args()
		args.ExecutionManager = &processMocks.ExecutionManagerMock{
			AddPairForExecutionCalled: func(pair headersCache.HeaderBodyPair) error {
				return errExpected
			},
		}

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Equal(t, errExpected, err)
	})

	t.Run("should error when getNextHeaderRequestingIfMissing fails", func(t *testing.T) {
		t.Parallel()

		args := createSyncBlockV3Args()

		prevHdr := &block.HeaderV3{Nonce: 1}
		args.Store = setupStore(args.Marshalizer, prevHdr, nil)

		// Setup pools that don't have the next header
		pools := dataRetrieverMock.NewPoolsHolderStub()
		pools.HeadersCalled = func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{
				GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
					return nil, nil, errors.New("header not found")
				},
			}
		}
		pools.MiniBlocksCalled = func() storage.Cacher {
			cs := cache.NewCacherStub()
			cs.RegisterHandlerCalled = func(i func(key []byte, value interface{})) {}
			return cs
		}
		pools.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{}
		}
		pools.ProofsCalled = func() dataRetriever.ProofsPool {
			return &dataRetrieverMock.ProofsPoolMock{}
		}
		args.PoolsHolder = pools

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Equal(t, process.ErrTimeIsOut, err)
	})
}

func TestShardBootstrap_SyncBlockLegacy(t *testing.T) {
	t.Parallel()

	createSyncBlockLegacyArgs := func() sync.ArgShardBootstrapper {
		args := CreateShardBootstrapMockArguments()

		currentHdr := &block.Header{
			Nonce:    2,
			RootHash: []byte("currentRootHash"),
		}
		args.ChainHandler = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return currentHdr
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("currentHash")
			},
			GetCurrentBlockRootHashCalled: func() []byte {
				return []byte("currentRootHash")
			},
		}

		// Header to sync (non-V3)
		header := &block.Header{
			Nonce:    3,
			Round:    1,
			RootHash: []byte("rootHash"),
		}
		args.PoolsHolder = setupPools(headerAndHash{
			header: header,
			hash:   []byte("aaa"),
		})

		args.ForkDetector = setupForkDetector(3)

		return args
	}

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createSyncBlockLegacyArgs()

		processBlockCalled := false
		processScheduledBlockCalled := false
		commitBlockCalled := false
		args.BlockProcessor = &testscommon.BlockProcessorStub{
			ProcessBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				processBlockCalled = true
				return nil
			},
			ProcessScheduledBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				processScheduledBlockCalled = true
				return nil
			},
			CommitBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
				commitBlockCalled = true
				return nil
			},
		}
		poolsStub, ok := args.PoolsHolder.(*dataRetrieverMock.PoolsHolderStub)
		require.True(t, ok)
		onExecutedBlockCalled := false
		poolsStub.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				OnExecutedBlockCalled: func(header data.HeaderHandler, rootHash []byte) error {
					onExecutedBlockCalled = true
					return nil
				},
			}
		}
		args.PoolsHolder = poolsStub

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Nil(t, err)
		assert.True(t, processBlockCalled)
		assert.True(t, processScheduledBlockCalled)
		assert.True(t, commitBlockCalled)
		assert.True(t, onExecutedBlockCalled)

		// coverage only. should not prepare again
		err = bs.SyncBlock(context.Background())
		assert.Nil(t, err)
	})

	t.Run("should error when OnExecutedBlock fails", func(t *testing.T) {
		t.Parallel()

		args := createSyncBlockLegacyArgs()

		poolsStub, ok := args.PoolsHolder.(*dataRetrieverMock.PoolsHolderStub)
		require.True(t, ok)
		poolsStub.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				OnExecutedBlockCalled: func(header data.HeaderHandler, rootHash []byte) error {
					return errExpected
				},
			}
		}
		args.PoolsHolder = poolsStub

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)

		err = bs.SyncBlock(context.Background())
		assert.Equal(t, errExpected, err)
	})
}

func TestShardBootstrap_GetNextHeaderWithCompetingProofsUsesLowestRound(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()

	currentHeader := &block.Header{Nonce: 1}
	args.ChainHandler = &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return currentHeader
		},
	}

	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.GetNotarizedHeaderHashCalled = func(nonce uint64) []byte {
		return nil
	}
	args.ForkDetector = forkDetector

	hashLowRound := []byte("hashA-low-round")
	hashHighRound := []byte("hashB-high-round")
	headerLowRound := &block.Header{Nonce: 2, Round: 2}

	// real proofs pool: the higher-round proof arrives first, yet fork-choice must fetch the
	// lowest-round header (canonical selection by round, not arrival order)
	proofsPool := proofscache.NewProofsPool(3, 100)
	require.True(t, proofsPool.AddProof(&block.HeaderProof{HeaderHash: hashHighRound, HeaderNonce: 2, HeaderRound: 3}))
	require.True(t, proofsPool.AddProof(&block.HeaderProof{HeaderHash: hashLowRound, HeaderNonce: 2, HeaderRound: 2}))

	pools := createMockPools()
	pools.ProofsCalled = func() dataRetriever.ProofsPool {
		return proofsPool
	}
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, hashLowRound) {
					return headerLowRound, nil
				}
				return nil, errors.New("requested header for the wrong proof")
			},
		}
	}
	args.PoolsHolder = pools

	bs, err := sync.NewShardBootstrap(args)
	require.Nil(t, err)

	header, hash, err := bs.GetNextHeaderRequestingIfMissing()
	require.Nil(t, err)
	require.Equal(t, hashLowRound, hash)
	require.Equal(t, headerLowRound, header)
}

type revertedBlocksCapture struct {
	*outport.OutportStub
	revertedHashes [][]byte
}

func (capture *revertedBlocksCapture) RevertIndexedBlock(headerData *outportcore.HeaderDataWithBody) error {
	capture.revertedHashes = append(capture.revertedHashes, headerData.HeaderHash)
	return nil
}

func TestBootstrap_RollBackV3(t *testing.T) {
	t.Parallel()

	marshaller := &mock.MarshalizerMock{}
	prevHdrHash := []byte("prev header hash")
	currHdrHash := []byte("curr header hash")

	newV3Header := func(nonce uint64, round uint64, prevHash []byte) *block.HeaderV3 {
		return &block.HeaderV3{
			Nonce:    nonce,
			Round:    round,
			PrevHash: prevHash,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{HeaderNonce: nonce - 2, HeaderHash: []byte("execHash")},
			},
		}
	}
	prevHdr := newV3Header(7, 9, []byte("older hash"))
	// committed contended head at nonce 8, not final: the switch candidate
	currHdr := newV3Header(8, 12, prevHdrHash)
	prevHdrBytes, _ := marshaller.Marshal(prevHdr)

	buildBootstrapper := func(finalNonce uint64, tweaks ...func(*sync.ArgShardBootstrapper)) (
		*sync.ShardBootstrap,
		*testscommon.ChainHandlerStub,
		*revertedBlocksCapture,
		map[string]uint64,
		map[string]bool,
		*processMocks.ExecutionManagerMock,
	) {
		removedAtNonce := make(map[string]uint64)
		calledFlags := make(map[string]bool)

		args := CreateShardBootstrapMockArguments()
		args.Marshalizer = marshaller
		args.Hasher = &mock.HasherStub{
			ComputeCalled: func(s string) []byte {
				return currHdrHash
			},
		}

		blkc := &testscommon.ChainHandlerStub{}
		var currentHeader data.HeaderHandler = currHdr
		currentHeaderHash := currHdrHash
		blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
			return currentHeader
		}
		blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
			return currentHeaderHash
		}
		blkc.SetCurrentBlockHeaderAndHashCalled = func(headerHash []byte, header data.HeaderHandler) error {
			currentHeader = header
			currentHeaderHash = headerHash
			return nil
		}
		args.ChainHandler = blkc

		args.Store = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						return prevHdrBytes, nil
					},
					RemoveCalled: func(key []byte) error {
						calledFlags["nonceHashStoreRemove"] = true
						return nil
					},
				}, nil
			},
		}
		args.ForkDetector = &mock.ForkDetectorMock{
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return finalNonce
			},
			RemoveCommittedHeaderCalled: func(nonce uint64, hash []byte) {
				if nonce == currHdr.GetNonce() && bytes.Equal(hash, currHdrHash) {
					calledFlags["removeCommittedHeader"] = true
				}
			},
		}
		executionManagerMock := &processMocks.ExecutionManagerMock{
			RemoveAtNonceAndHigherCalled: func(nonce uint64) error {
				removedAtNonce["executionManager"] = nonce
				return nil
			},
			RewindExecutionStateToTipCalled: func(newTip data.HeaderHandler) error {
				removedAtNonce["rewindTip"] = newTip.GetNonce()
				return nil
			},
		}
		args.ExecutionManager = executionManagerMock
		args.BlockProcessor = &testscommon.BlockProcessorStub{
			RestoreBlockIntoPoolsCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
				removedAtNonce["restoredIntoPools"] = header.GetNonce()
				return nil
			},
		}
		args.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			RollBackToBlockCalled: func(headerHash []byte) error {
				calledFlags["scheduledRollBack"] = true
				return nil
			},
		}
		outportCapture := &revertedBlocksCapture{OutportStub: &outport.OutportStub{}}
		args.OutportHandler = outportCapture

		for _, tweak := range tweaks {
			tweak(&args)
		}

		bs, err := sync.NewShardBootstrap(args)
		require.Nil(t, err)
		bs.SetForkNonce(currHdr.GetNonce())

		return bs, blkc, outportCapture, removedAtNonce, calledFlags, executionManagerMock
	}

	t.Run("reverts the committed non-final head without touching tries or scheduled state", func(t *testing.T) {
		t.Parallel()

		bs, blkc, outportCapture, removedAtNonce, calledFlags, _ := buildBootstrapper(5)
		bs.SetPreparedForSync(true)

		err := bs.RollBack(true)
		require.Nil(t, err)

		require.Equal(t, prevHdr.GetNonce(), blkc.GetCurrentBlockHeader().GetNonce())
		require.Equal(t, prevHdrHash, blkc.GetCurrentBlockHeaderHash())
		require.Equal(t, currHdr.GetNonce(), removedAtNonce["executionManager"])
		require.Equal(t, currHdr.GetNonce(), removedAtNonce["restoredIntoPools"])
		require.True(t, calledFlags["removeCommittedHeader"])
		require.True(t, calledFlags["nonceHashStoreRemove"])
		require.False(t, calledFlags["scheduledRollBack"])
		require.Equal(t, [][]byte{currHdrHash}, outportCapture.revertedHashes)

		// the execution state was rewound to the new tip and the sync prepare step was re-armed
		require.Equal(t, prevHdr.GetNonce(), removedAtNonce["rewindTip"])
		require.False(t, bs.GetPreparedForSync())
	})

	t.Run("never crosses the final checkpoint, fork-driven included", func(t *testing.T) {
		t.Parallel()

		bs, blkc, outportCapture, removedAtNonce, calledFlags, _ := buildBootstrapper(currHdr.GetNonce())
		bs.SetPreparedForSync(true)

		err := bs.RollBack(true)
		require.Equal(t, sync.ErrRollBackBehindFinalHeader, err)

		require.Equal(t, currHdr.GetNonce(), blkc.GetCurrentBlockHeader().GetNonce())
		require.Empty(t, removedAtNonce)
		require.Empty(t, calledFlags)
		require.Empty(t, outportCapture.revertedHashes)
		require.True(t, bs.GetPreparedForSync())
	})

	t.Run("rewind failure does not re-arm the sync prepare step and surfaces in the result", func(t *testing.T) {
		t.Parallel()

		bs, _, _, _, _, executionManagerMock := buildBootstrapper(5)
		bs.SetPreparedForSync(true)
		executionManagerMock.RewindExecutionStateToTipCalled = func(newTip data.HeaderHandler) error {
			return errors.New("expected error")
		}

		// the roll back itself succeeded, but no caller may continue on unrewound execution state
		err := bs.RollBack(true)
		require.Equal(t, sync.ErrExecutionRealignPending, err)
		require.True(t, bs.GetPreparedForSync())
	})

	t.Run("realigns even when a step after the block revert fails", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		bs, _, _, removedAtNonce, _, _ := buildBootstrapper(5, func(args *sync.ArgShardBootstrapper) {
			args.HistoryRepo = &dblookupext.HistoryRepositoryStub{
				RevertBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
					return expectedErr
				},
			}
		})
		bs.SetPreparedForSync(true)

		err := bs.RollBack(true)
		require.Equal(t, expectedErr, err)

		// the tip moved back before the failure, so the execution state must still be realigned
		require.Equal(t, prevHdr.GetNonce(), removedAtNonce["rewindTip"])
		require.False(t, bs.GetPreparedForSync())
	})

	t.Run("resets the tx selection tracker after rewinding the execution state", func(t *testing.T) {
		t.Parallel()

		resetTrackerCalled := false
		bs, _, _, removedAtNonce, _, _ := buildBootstrapper(5, func(args *sync.ArgShardBootstrapper) {
			pools := createMockPools()
			pools.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
				return &testscommon.ShardedDataStub{
					ResetTrackerCalled: func() {
						resetTrackerCalled = true
					},
				}
			}
			args.PoolsHolder = pools
		})
		bs.SetPreparedForSync(true)

		err := bs.RollBack(true)
		require.Nil(t, err)

		require.True(t, resetTrackerCalled)
		require.Equal(t, prevHdr.GetNonce(), removedAtNonce["rewindTip"])
		require.False(t, bs.GetPreparedForSync())
	})

	t.Run("does not reset the tx selection tracker when the rewind fails", func(t *testing.T) {
		t.Parallel()

		resetTrackerCalled := false
		bs, _, _, _, _, executionManagerMock := buildBootstrapper(5, func(args *sync.ArgShardBootstrapper) {
			pools := createMockPools()
			pools.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
				return &testscommon.ShardedDataStub{
					ResetTrackerCalled: func() {
						resetTrackerCalled = true
					},
				}
			}
			args.PoolsHolder = pools
		})
		bs.SetPreparedForSync(true)
		executionManagerMock.RewindExecutionStateToTipCalled = func(newTip data.HeaderHandler) error {
			return errors.New("expected error")
		}

		err := bs.RollBack(true)
		require.Equal(t, sync.ErrExecutionRealignPending, err)

		require.False(t, resetTrackerCalled)
		require.True(t, bs.GetPreparedForSync())
	})

	t.Run("a pool restoration failure leaves the epoch trigger untouched", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		triggerReverted := false
		bs, blkc, _, _, _, _ := buildBootstrapper(5, func(args *sync.ArgShardBootstrapper) {
			args.BlockProcessor = &testscommon.BlockProcessorStub{
				RestoreBlockIntoPoolsCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
					return expectedErr
				},
			}
			args.EpochStartTrigger = &mock.EpochStartTriggerStub{
				RevertStateToBlockCalled: func(header data.HeaderHandler) error {
					triggerReverted = true
					return nil
				},
			}
		})
		bs.SetPreparedForSync(true)

		err := bs.RollBack(true)
		require.Equal(t, expectedErr, err)
		require.False(t, triggerReverted)
		require.Equal(t, currHdr.GetNonce(), blkc.GetCurrentBlockHeader().GetNonce())
	})

	t.Run("a trigger revert failure retries without restoring the block twice", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		restoreCalls := 0
		revertCalls := 0
		bs, blkc, _, _, calledFlags, _ := buildBootstrapper(5, func(args *sync.ArgShardBootstrapper) {
			args.BlockProcessor = &testscommon.BlockProcessorStub{
				RestoreBlockIntoPoolsCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
					restoreCalls++
					return nil
				},
			}
			args.EpochStartTrigger = &mock.EpochStartTriggerStub{
				RevertStateToBlockCalled: func(header data.HeaderHandler) error {
					revertCalls++
					if revertCalls == 1 {
						return expectedErr
					}
					return nil
				},
			}
		})
		bs.SetPreparedForSync(true)

		err := bs.RollBack(true)
		require.Equal(t, expectedErr, err)
		require.Equal(t, currHdr.GetNonce(), blkc.GetCurrentBlockHeader().GetNonce())

		bs.SetForkNonce(currHdr.GetNonce())
		err = bs.RollBack(true)
		require.Nil(t, err)
		require.Equal(t, 1, restoreCalls)
		require.Equal(t, prevHdr.GetNonce(), blkc.GetCurrentBlockHeader().GetNonce())
		require.True(t, calledFlags["removeCommittedHeader"])
	})

	t.Run("a failed rewind arms a mandatory realign that blocks syncing until it succeeds", func(t *testing.T) {
		t.Parallel()

		resetTrackerCalled := false
		bs, _, _, _, _, executionManagerMock := buildBootstrapper(5, func(args *sync.ArgShardBootstrapper) {
			pools := createMockPools()
			pools.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
				return &testscommon.ShardedDataStub{
					ResetTrackerCalled: func() {
						resetTrackerCalled = true
					},
				}
			}
			args.PoolsHolder = pools
		})
		bs.SetPreparedForSync(true)
		executionManagerMock.RewindExecutionStateToTipCalled = func(newTip data.HeaderHandler) error {
			return errors.New("expected error")
		}

		err := bs.RollBack(true)
		require.Equal(t, sync.ErrExecutionRealignPending, err)
		require.True(t, bs.GetPendingV3Realign())

		// still failing: the sync loop stays blocked on the mandatory retry
		err = bs.SyncBlockBase()
		require.Equal(t, sync.ErrExecutionRealignPending, err)
		require.True(t, bs.GetPendingV3Realign())
		require.False(t, resetTrackerCalled)

		// recovered: the retry completes the compensation and unblocks the loop
		executionManagerMock.RewindExecutionStateToTipCalled = func(newTip data.HeaderHandler) error {
			return nil
		}
		err = bs.SyncBlockBase()
		require.Nil(t, err)
		require.False(t, bs.GetPendingV3Realign())
		require.True(t, resetTrackerCalled)
		require.False(t, bs.GetPreparedForSync())
	})

	t.Run("an interrupted roll back is completed before any other sync work", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		restoreCalls := 0
		revertCalls := 0
		bs, blkc, _, _, calledFlags, _ := buildBootstrapper(5, func(args *sync.ArgShardBootstrapper) {
			args.BlockProcessor = &testscommon.BlockProcessorStub{
				RestoreBlockIntoPoolsCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
					restoreCalls++
					return nil
				},
			}
			args.EpochStartTrigger = &mock.EpochStartTriggerStub{
				RevertStateToBlockCalled: func(header data.HeaderHandler) error {
					revertCalls++
					if revertCalls == 1 {
						return expectedErr
					}
					return nil
				},
			}
		})
		bs.SetPreparedForSync(true)

		err := bs.RollBack(true)
		require.Equal(t, expectedErr, err)
		require.Equal(t, currHdr.GetNonce(), blkc.GetCurrentBlockHeader().GetNonce())

		// the next sync round must complete the roll back instead of doing any other work
		err = bs.SyncBlockBase()
		require.Nil(t, err)
		require.Equal(t, 1, restoreCalls)
		require.Equal(t, prevHdr.GetNonce(), blkc.GetCurrentBlockHeader().GetNonce())
		require.True(t, calledFlags["removeCommittedHeader"])
	})

	t.Run("a still-failing interrupted roll back keeps sync blocked", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		restoreCalls := 0
		bs, blkc, _, _, _, _ := buildBootstrapper(5, func(args *sync.ArgShardBootstrapper) {
			args.BlockProcessor = &testscommon.BlockProcessorStub{
				RestoreBlockIntoPoolsCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
					restoreCalls++
					return nil
				},
			}
			args.EpochStartTrigger = &mock.EpochStartTriggerStub{
				RevertStateToBlockCalled: func(header data.HeaderHandler) error {
					return expectedErr
				},
			}
		})
		bs.SetPreparedForSync(true)

		err := bs.RollBack(true)
		require.Equal(t, expectedErr, err)

		err = bs.SyncBlockBase()
		require.Equal(t, expectedErr, err)
		err = bs.SyncBlockBase()
		require.Equal(t, expectedErr, err)

		require.Equal(t, 1, restoreCalls)
		require.Equal(t, currHdr.GetNonce(), blkc.GetCurrentBlockHeader().GetNonce())
	})

	t.Run("a superseded interrupted roll back stands down instead of reverting the new tip", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		restoreCalls := 0
		bs, blkc, _, _, _, _ := buildBootstrapper(5, func(args *sync.ArgShardBootstrapper) {
			args.BlockProcessor = &testscommon.BlockProcessorStub{
				RestoreBlockIntoPoolsCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
					restoreCalls++
					return nil
				},
			}
			args.EpochStartTrigger = &mock.EpochStartTriggerStub{
				RevertStateToBlockCalled: func(header data.HeaderHandler) error {
					return expectedErr
				},
			}
		})
		bs.SetPreparedForSync(true)

		err := bs.RollBack(true)
		require.Equal(t, expectedErr, err)

		// consensus committed a new block on top before the next sync round
		newTip := newV3Header(9, 13, currHdrHash)
		_ = blkc.SetCurrentBlockHeaderAndHash([]byte("newTipHash"), newTip)

		err = bs.SyncBlockBase()
		require.Nil(t, err)
		require.Equal(t, 1, restoreCalls)
		require.Equal(t, newTip.GetNonce(), blkc.GetCurrentBlockHeader().GetNonce())
		// the marker is cleared: the next round is not a completion round anymore
		require.Empty(t, bs.GetLastRestoredHeaderHash())
	})

	t.Run("remove succeeds then restoration and rewind fail: header restored, trigger untouched, recovery armed", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		triggerReverted := false
		bs, blkc, _, removedAtNonce, _, executionManagerMock := buildBootstrapper(5, func(args *sync.ArgShardBootstrapper) {
			args.BlockProcessor = &testscommon.BlockProcessorStub{
				RestoreBlockIntoPoolsCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
					return expectedErr
				},
			}
			args.EpochStartTrigger = &mock.EpochStartTriggerStub{
				RevertStateToBlockCalled: func(header data.HeaderHandler) error {
					triggerReverted = true
					return nil
				},
			}
		})
		bs.SetPreparedForSync(true)
		executionManagerMock.RewindExecutionStateToTipCalled = func(newTip data.HeaderHandler) error {
			return errors.New("rewind error")
		}

		err := bs.RollBack(true)
		require.Equal(t, expectedErr, err)

		require.Equal(t, currHdr.GetNonce(), removedAtNonce["executionManager"])
		require.Equal(t, currHdr.GetNonce(), blkc.GetCurrentBlockHeader().GetNonce())
		require.False(t, triggerReverted)
		require.True(t, bs.GetPendingV3Realign())
		require.True(t, bs.GetPreparedForSync())
	})
}
