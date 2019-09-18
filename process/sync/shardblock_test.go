package sync_test

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	goSync "sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus/round"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

// waitTime defines the time in milliseconds until node waits the requested info from the network
const waitTime = time.Duration(100 * time.Millisecond)

type removedFlags struct {
	flagHdrRemovedFromNonces       bool
	flagHdrRemovedFromHeaders      bool
	flagHdrRemovedFromStorage      bool
	flagHdrRemovedFromForkDetector bool
}

func createMockResolversFinder() *mock.ResolversFinderStub {
	return &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
			if strings.Contains(baseTopic, factory.HeadersTopic) {
				return &mock.HeaderResolverMock{
					RequestDataFromNonceCalled: func(nonce uint64) error {
						return nil
					},
					RequestDataFromHashCalled: func(hash []byte) error {
						return nil
					},
				}, nil
			}

			if strings.Contains(baseTopic, factory.MiniBlocksTopic) {
				return &mock.MiniBlocksResolverMock{
					GetMiniBlocksCalled: func(hashes [][]byte) block.MiniBlockSlice {
						return make(block.MiniBlockSlice, 0)
					},
				}, nil
			}

			return nil, nil
		},
	}
}

func createMockResolversFinderNilMiniBlocks() *mock.ResolversFinderStub {
	return &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
			if strings.Contains(baseTopic, factory.HeadersTopic) {
				return &mock.HeaderResolverMock{
					RequestDataFromNonceCalled: func(nonce uint64) error {
						return nil
					},
					RequestDataFromHashCalled: func(hash []byte) error {
						return nil
					},
				}, nil
			}

			if strings.Contains(baseTopic, factory.MiniBlocksTopic) {
				return &mock.MiniBlocksResolverMock{
					RequestDataFromHashCalled: func(hash []byte) error {
						return nil
					},
					RequestDataFromHashArrayCalled: func(hash [][]byte) error {
						return nil
					},
					GetMiniBlocksCalled: func(hashes [][]byte) block.MiniBlockSlice {
						return nil
					},
				}, nil
			}

			return nil, nil
		},
	}
}

func createMockPools() *mock.PoolsHolderStub {
	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{
			HasOrAddCalled: func(key []byte, value interface{}) (ok, evicted bool) {
				return false, false
			},
			RegisterHandlerCalled: func(func(key []byte)) {},
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
			RemoveCalled: func(key []byte) {
				return
			},
		}
		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{
			GetCalled: func(u uint64) (dataRetriever.ShardIdHashMap, bool) {
				return nil, false
			},
			RegisterHandlerCalled: func(handler func(nonce uint64, shardId uint32, hash []byte)) {},
		}
		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
			RegisterHandlerCalled: func(i func(key []byte)) {},
		}
		return cs
	}

	return pools
}

func createStore() *mock.ChainStorerMock {
	return &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return nil, process.ErrMissingHeader
				},
				RemoveCalled: func(key []byte) error {
					return nil
				},
			}
		},
	}
}

func generateTestCache() storage.Cacher {
	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, 1000, 1)
	return cache
}

func generateTestUnit() storage.Storer {
	memDB, _ := memorydb.New()
	storer, _ := storageUnit.NewStorageUnit(
		generateTestCache(),
		memDB,
	)
	return storer
}

func createFullStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, generateTestUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit, generateTestUnit())

	return store
}

func createBlockProcessor() *mock.BlockProcessorMock {
	blockProcessorMock := &mock.BlockProcessorMock{
		ProcessBlockCalled: func(blk data.ChainHandler, hdr data.HeaderHandler, bdy data.BodyHandler, haveTime func() time.Duration) error {
			_ = blk.SetCurrentBlockHeader(hdr.(*block.Header))
			return nil
		},
		RevertAccountStateCalled: func() {
			return
		},
		CommitBlockCalled: func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
			return nil
		},
	}

	return blockProcessorMock
}

func createHeadersDataPool(removedHashCompare []byte, remFlags *removedFlags) storage.Cacher {
	sds := &mock.CacherStub{
		HasOrAddCalled: func(key []byte, value interface{}) (ok, evicted bool) {
			return false, false
		},
		RegisterHandlerCalled: func(func(key []byte)) {},
		RemoveCalled: func(key []byte) {
			if bytes.Equal(key, removedHashCompare) {
				remFlags.flagHdrRemovedFromHeaders = true
			}
		},
	}
	return sds
}

func createHeadersNoncesDataPool(
	getNonceCompare uint64,
	getRetHash []byte,
	removedNonce uint64,
	remFlags *removedFlags,
	shardId uint32,
) dataRetriever.Uint64SyncMapCacher {

	hnc := &mock.Uint64SyncMapCacherStub{
		RegisterHandlerCalled: func(handler func(nonce uint64, shardId uint32, hash []byte)) {},
		GetCalled: func(u uint64) (dataRetriever.ShardIdHashMap, bool) {
			if u == getNonceCompare {
				syncMap := &dataPool.ShardIdHashSyncMap{}
				syncMap.Store(shardId, getRetHash)

				return syncMap, true
			}

			return nil, false
		},
		RemoveCalled: func(nonce uint64, providedShardId uint32) {
			if nonce == removedNonce && shardId == providedShardId {
				remFlags.flagHdrRemovedFromNonces = true
			}
		},
	}
	return hnc
}

func createForkDetector(removedNonce uint64, remFlags *removedFlags) process.ForkDetector {
	return &mock.ForkDetectorMock{
		RemoveHeadersCalled: func(nonce uint64, hash []byte) {
			if nonce == removedNonce {
				remFlags.flagHdrRemovedFromForkDetector = true
			}
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return uint64(removedNonce)
		},
		ProbableHighestNonceCalled: func() uint64 {
			return uint64(0)
		},
	}
}

func initBlockchain() *mock.BlockChainMock {
	blkc := &mock.BlockChainMock{
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

//------- NewShardBootstrap

func TestNewShardBootstrap_NilPoolsHolderShouldErr(t *testing.T) {
	t.Parallel()

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		nil,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversFinderStub{},
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilPoolsHolder, err)
}

func TestNewShardBootstrap_PoolsHolderRetNilOnHeadersShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	pools.HeadersCalled = func() storage.Cacher {
		return nil
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversFinderStub{},
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilHeadersDataPool, err)
}

func TestNewShardBootstrap_PoolsHolderRetNilOnHeadersNoncesShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		return nil
	}
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversFinderStub{},
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
}

func TestNewShardBootstrap_PoolsHolderRetNilOnTxBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	pools.MiniBlocksCalled = func() storage.Cacher {
		return nil
	}
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversFinderStub{},
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilTxBlockBody, err)
}

func TestNewShardBootstrap_NilStoreShouldErr(t *testing.T) {
	t.Parallel()
	blkc := initBlockchain()
	pools := createMockPools()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		pools,
		nil,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversFinderStub{},
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilStore, err)
}

func TestNewShardBootstrap_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		pools,
		createStore(),
		nil,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversFinderStub{},
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestNewShardBootstrap_NilRounderShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	blkc := initBlockchain()
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		nil,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversFinderStub{},
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilRounder, err)
}

func TestNewShardBootstrap_NilBlockProcessorShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		nil,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversFinderStub{},
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilBlockExecutor, err)
}

func TestNewShardBootstrap_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		nil,
		marshalizer,
		forkDetector,
		&mock.ResolversFinderStub{},
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewShardBootstrap_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		nil,
		forkDetector,
		&mock.ResolversFinderStub{},
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewShardBootstrap_NilForkDetectorShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		nil,
		&mock.ResolversFinderStub{},
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilForkDetector, err)
}

func TestNewShardBootstrap_NilResolversContainerShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		nil,
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilResolverContainer, err)
}

func TestNewShardBootstrap_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversFinderStub{},
		nil,
		account,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewShardBootstrap_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()

	bs, err := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversFinderStub{},
		shardCoordinator,
		nil,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewShardBootstrap_NilHeaderResolverShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	pools := createMockPools()
	resFinder := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
			if strings.Contains(baseTopic, factory.HeadersTopic) {
				return nil, errExpected
			}

			if strings.Contains(baseTopic, factory.MiniBlocksTopic) {
				return &mock.ResolverStub{}, nil
			}

			return nil, nil
		},
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		resFinder,
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, errExpected, err)
}

func TestNewShardBootstrap_NilTxBlockBodyResolverShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	pools := createMockPools()
	resFinder := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
			if strings.Contains(baseTopic, factory.HeadersTopic) {
				return &mock.HeaderResolverMock{}, errExpected
			}

			if strings.Contains(baseTopic, factory.MiniBlocksTopic) {
				return nil, errExpected
			}

			return nil, nil
		},
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		resFinder,
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.Nil(t, bs)
	assert.Equal(t, errExpected, err)
}

func TestNewShardBootstrap_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := 0

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
			assert.Fail(t, "should have not reached this point")
			return false, false
		}

		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {
			wasCalled++
		}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
			wasCalled++
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.NotNil(t, bs)
	assert.Nil(t, err)
	assert.Equal(t, 2, wasCalled)
}

//------- processing

func TestBootstrap_SyncBlockShouldCallForkChoice(t *testing.T) {
	t.Parallel()

	hdr := block.Header{Nonce: 1, PubKeysBitmap: []byte("X")}
	blockBodyUnit := &mock.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return nil, nil
		},
	}

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MiniBlockUnit, blockBodyUnit)

	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
	)

	_ = blkc.SetAppStatusHandler(&mock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {},
	})

	blkc.CurrentBlockHeader = &hdr

	pools := createMockPools()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64, []byte) {
		return true, math.MaxUint64, nil
	}
	forkDetector.RemoveHeadersCalled = func(nonce uint64, hash []byte) {
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return uint64(hdr.Nonce)
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 100
	}
	forkDetector.ResetProbableHighestNonceIfNeededCalled = func() {
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), &mock.SyncTimerMock{})

	blockProcessorMock := createBlockProcessor()

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blockProcessorMock,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	r := bs.SyncBlock()

	assert.Equal(t, process.ErrTimeIsOut, r)
}

func TestBootstrap_ShouldReturnTimeIsOutWhenMissingHeader(t *testing.T) {
	t.Parallel()

	hdr := block.Header{Nonce: 1}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	pools := createMockPools()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64, []byte) {
		return false, math.MaxUint64, nil
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 100
	}
	forkDetector.ResetProbableHighestNonceIfNeededCalled = func() {
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(),
		time.Now().Add(2*time.Duration(100*time.Millisecond)),
		time.Duration(100*time.Millisecond),
		&mock.SyncTimerMock{})

	blockProcessorMock := createBlockProcessor()

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		&blkc,
		rnd,
		blockProcessorMock,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	r := bs.SyncBlock()

	assert.Equal(t, process.ErrTimeIsOut, r)
}

func TestBootstrap_ShouldReturnTimeIsOutWhenMissingBody(t *testing.T) {
	t.Parallel()

	hdr := block.Header{Nonce: 1, PubKeysBitmap: []byte("X")}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	shardId := uint32(0)
	hash := []byte("aaa")

	pools := createMockPools()
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(hash, key) {
				return &block.Header{Nonce: 2}, true
			}

			return nil, false
		}

		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		sds.RemoveCalled = func(key []byte) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}
		hnc.GetCalled = func(u uint64) (dataRetriever.ShardIdHashMap, bool) {
			if u == 2 {
				syncMap := &dataPool.ShardIdHashSyncMap{}
				syncMap.Store(shardId, hash)

				return syncMap, true
			}

			return nil, false
		}

		return hnc
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64, []byte) {
		return false, math.MaxUint64, nil
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 2
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return 1
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(),
		time.Now().Add(2*time.Duration(100*time.Millisecond)),
		time.Duration(100*time.Millisecond),
		&mock.SyncTimerMock{})

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		&blkc,
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderNilMiniBlocks(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.RequestHeaderWithNonce(2)
	r := bs.SyncBlock()
	assert.Equal(t, process.ErrTimeIsOut, r)
}

func TestBootstrap_ShouldNotNeedToSync(t *testing.T) {
	t.Parallel()

	ebm := createBlockProcessor()

	hdr := block.Header{Nonce: 1, Round: 0}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	pools := createMockPools()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64, []byte) {
		return false, math.MaxUint64, nil
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return uint64(hdr.Nonce)
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 1
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), &mock.SyncTimerMock{})

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		&blkc,
		rnd,
		ebm,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.StartSync()
	time.Sleep(200 * time.Millisecond)
	bs.StopSync()
}

func TestBootstrap_SyncShouldSyncOneBlock(t *testing.T) {
	t.Parallel()

	ebm := createBlockProcessor()

	hdr := block.Header{Nonce: 1, Round: 0}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	shardId := uint32(0)
	hash := []byte("aaa")

	mutDataAvailable := goSync.RWMutex{}
	dataAvailable := false

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			mutDataAvailable.RLock()
			defer mutDataAvailable.RUnlock()

			if bytes.Equal(hash, key) && dataAvailable {
				return &block.Header{
					Nonce:         2,
					Round:         1,
					BlockBodyType: block.TxBlock,
					RootHash:      []byte("bbb")}, true
			}

			return nil, false
		}

		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}
		hnc.GetCalled = func(u uint64) (dataRetriever.ShardIdHashMap, bool) {
			mutDataAvailable.RLock()
			defer mutDataAvailable.RUnlock()

			if u == 2 && dataAvailable {
				syncMap := &dataPool.ShardIdHashSyncMap{}
				syncMap.Store(shardId, hash)

				return syncMap, true
			}

			return nil, false
		}
		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("bbb"), key) && dataAvailable {
				return make(block.MiniBlockSlice, 0), true
			}

			return nil, false
		}

		return cs
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64, []byte) {
		return false, math.MaxUint64, nil
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return uint64(hdr.Nonce)
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 2
	}
	forkDetector.ResetProbableHighestNonceIfNeededCalled = func() {
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	account.RootHashCalled = func() ([]byte, error) {
		return nil, nil
	}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(200*time.Millisecond), time.Duration(100*time.Millisecond), &mock.SyncTimerMock{})

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		&blkc,
		rnd,
		ebm,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.StartSync()

	time.Sleep(200 * time.Millisecond)

	mutDataAvailable.Lock()
	dataAvailable = true
	mutDataAvailable.Unlock()

	time.Sleep(200 * time.Millisecond)

	bs.StopSync()
}

func TestBootstrap_ShouldReturnNilErr(t *testing.T) {
	t.Parallel()

	ebm := createBlockProcessor()

	hdr := block.Header{Nonce: 1}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	shardId := uint32(0)
	hash := []byte("aaa")

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(hash, key) {
				return &block.Header{
					Nonce:         2,
					Round:         1,
					BlockBodyType: block.TxBlock,
					RootHash:      []byte("bbb")}, true
			}

			return nil, false
		}

		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}
		hnc.GetCalled = func(u uint64) (dataRetriever.ShardIdHashMap, bool) {
			if u == 2 {
				syncMap := &dataPool.ShardIdHashSyncMap{}
				syncMap.Store(shardId, hash)

				return syncMap, true
			}

			return nil, false
		}
		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("bbb"), key) {
				return make(block.MiniBlockSlice, 0), true
			}

			return nil, false
		}

		return cs
	}

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64, []byte) {
		return false, math.MaxUint64, nil
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 2
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(),
		time.Now().Add(2*time.Duration(100*time.Millisecond)),
		time.Duration(100*time.Millisecond),
		&mock.SyncTimerMock{})

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		&blkc,
		rnd,
		ebm,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	r := bs.SyncBlock()

	assert.Nil(t, r)
}

func TestBootstrap_SyncBlockShouldReturnErrorWhenProcessBlockFailed(t *testing.T) {
	t.Parallel()

	ebm := createBlockProcessor()

	hdr := block.Header{Nonce: 1, PubKeysBitmap: []byte("X")}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	shardId := uint32(0)
	hash := []byte("aaa")

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(hash, key) {
				return &block.Header{
					Nonce:         2,
					Round:         1,
					BlockBodyType: block.TxBlock,
					RootHash:      []byte("bbb")}, true
			}

			return nil, false
		}
		sds.RegisterHandlerCalled = func(func(key []byte)) {}
		sds.RemoveCalled = func(key []byte) {}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}
		hnc.GetCalled = func(u uint64) (dataRetriever.ShardIdHashMap, bool) {
			if u == 2 {
				syncMap := &dataPool.ShardIdHashSyncMap{}
				syncMap.Store(shardId, hash)

				return syncMap, true
			}

			return nil, false
		}
		hnc.RemoveCalled = func(nonce uint64, shardId uint32) {}
		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("bbb"), key) {
				return make(block.MiniBlockSlice, 0), true
			}

			return nil, false
		}

		return cs
	}

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64, []byte) {
		return false, math.MaxUint64, nil
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return uint64(hdr.Nonce)
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 2
	}
	forkDetector.RemoveHeadersCalled = func(nonce uint64, hash []byte) {}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(),
		time.Now().Add(2*time.Duration(100*time.Millisecond)),
		time.Duration(100*time.Millisecond),
		&mock.SyncTimerMock{})

	ebm.ProcessBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
		return process.ErrBlockHashDoesNotMatch
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		&blkc,
		rnd,
		ebm,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	err := bs.SyncBlock()
	assert.Equal(t, process.ErrBlockHashDoesNotMatch, err)
}

func TestBootstrap_ShouldSyncShouldReturnFalseWhenCurrentBlockIsNilAndRoundIndexIsZero(t *testing.T) {
	t.Parallel()

	pools := createMockPools()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{
		CheckForkCalled: func() (bool, uint64, []byte) {
			return false, math.MaxUint64, nil
		},
		ProbableHighestNonceCalled: func() uint64 {
			return 0
		},
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), &mock.SyncTimerMock{})

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		initBlockchain(),
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.False(t, bs.ShouldSync())
}

func TestBootstrap_ShouldReturnTrueWhenCurrentBlockIsNilAndRoundIndexIsGreaterThanZero(t *testing.T) {
	t.Parallel()

	pools := createMockPools()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64, []byte) {
		return false, math.MaxUint64, nil
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 1
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond), &mock.SyncTimerMock{})

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		initBlockchain(),
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.True(t, bs.ShouldSync())
}

func TestBootstrap_ShouldReturnFalseWhenNodeIsSynced(t *testing.T) {
	t.Parallel()

	hdr := block.Header{Nonce: 0}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	pools := createMockPools()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64, []byte) {
		return false, math.MaxUint64, nil
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 0
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), &mock.SyncTimerMock{})

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		&blkc,
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.False(t, bs.ShouldSync())
}

func TestBootstrap_ShouldReturnTrueWhenNodeIsNotSynced(t *testing.T) {
	t.Parallel()

	hdr := block.Header{Nonce: 0}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	pools := createMockPools()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64, []byte) {
		return false, math.MaxUint64, nil
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 1
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond), &mock.SyncTimerMock{})

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		&blkc,
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	assert.True(t, bs.ShouldSync())
}

func TestBootstrap_ShouldSyncShouldReturnTrueWhenForkIsDetectedAndItReceivesTheSameWrongHeader(t *testing.T) {
	t.Parallel()

	hdr1 := block.Header{Nonce: 1, Round: 2, PubKeysBitmap: []byte("A")}
	hash1 := []byte("hash1")

	hdr2 := block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("B")}
	hash2 := []byte("hash2")

	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr1
	}

	finalHeaders := []data.HeaderHandler{
		&hdr2,
	}
	finalHeadersHashes := [][]byte{
		hash2,
	}

	pools := createMockPools()
	pools.HeadersCalled = func() storage.Cacher {
		return GetCacherWithHeaders(&hdr1, &hdr2, hash1, hash2)
	}

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	rounder := &mock.RounderMock{}
	rounder.RoundIndex = 2
	forkDetector, _ := sync.NewShardForkDetector(rounder)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		&blkc,
		rounder,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	_ = forkDetector.AddHeader(&hdr1, hash1, process.BHProcessed, nil, nil)
	_ = forkDetector.AddHeader(&hdr2, hash2, process.BHNotarized, finalHeaders, finalHeadersHashes)

	shouldSync := bs.ShouldSync()
	assert.True(t, shouldSync)
	assert.True(t, bs.IsForkDetected())

	if shouldSync && bs.IsForkDetected() {
		forkDetector.RemoveHeaders(hdr1.GetNonce(), hash1)
		bs.ReceivedHeaders(hash1)
		_ = forkDetector.AddHeader(&hdr1, hash1, process.BHProcessed, nil, nil)
	}

	shouldSync = bs.ShouldSync()
	assert.True(t, shouldSync)
	assert.True(t, bs.IsForkDetected())
}

func TestBootstrap_ShouldSyncShouldReturnFalseWhenForkIsDetectedAndItReceivesTheGoodHeader(t *testing.T) {
	t.Parallel()

	hdr1 := block.Header{Nonce: 1, Round: 2, PubKeysBitmap: []byte("A")}
	hash1 := []byte("hash1")

	hdr2 := block.Header{Nonce: 1, Round: 1, PubKeysBitmap: []byte("B")}
	hash2 := []byte("hash2")

	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr2
	}

	finalHeaders := []data.HeaderHandler{
		&hdr2,
	}
	finalHeadersHashes := [][]byte{
		hash2,
	}

	pools := createMockPools()
	pools.HeadersCalled = func() storage.Cacher {
		return GetCacherWithHeaders(&hdr1, &hdr2, hash1, hash2)
	}

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	rounder := &mock.RounderMock{}
	rounder.RoundIndex = 2
	forkDetector, _ := sync.NewShardForkDetector(rounder)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		&blkc,
		rounder,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	_ = forkDetector.AddHeader(&hdr1, hash1, process.BHProcessed, nil, nil)
	_ = forkDetector.AddHeader(&hdr2, hash2, process.BHNotarized, finalHeaders, finalHeadersHashes)

	shouldSync := bs.ShouldSync()
	assert.True(t, shouldSync)
	assert.True(t, bs.IsForkDetected())

	if shouldSync && bs.IsForkDetected() {
		forkDetector.RemoveHeaders(hdr1.GetNonce(), hash1)
		bs.ReceivedHeaders(hash2)
		_ = forkDetector.AddHeader(&hdr2, hash2, process.BHProcessed, finalHeaders, finalHeadersHashes)
	}

	shouldSync = bs.ShouldSync()
	assert.False(t, shouldSync)
	assert.False(t, bs.IsForkDetected())
}

func TestBootstrap_GetHeaderFromPoolShouldReturnNil(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64, []byte) {
		return false, math.MaxUint64, nil
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), &mock.SyncTimerMock{})

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		initBlockchain(),
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	hdr, _, _ := process.GetShardHeaderFromPoolWithNonce(0, 0, pools.Headers(), pools.HeadersNonces())
	assert.NotNil(t, bs)
	assert.Nil(t, hdr)
}

func TestBootstrap_GetHeaderFromPoolShouldReturnHeader(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{Nonce: 0}

	shardId := uint32(0)
	hash := []byte("aaa")

	pools := createMockPools()
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(hash, key) {
				return hdr, true
			}
			return nil, false
		}

		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}
		hnc.GetCalled = func(u uint64) (dataRetriever.ShardIdHashMap, bool) {
			if u == 0 {
				syncMap := &dataPool.ShardIdHashSyncMap{}
				syncMap.Store(shardId, hash)

				return syncMap, true
			}

			return nil, false
		}

		return hnc
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	account := &mock.AccountsStub{}

	shardCoordinator := mock.NewOneShardCoordinatorMock()

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), &mock.SyncTimerMock{})

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		initBlockchain(),
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	hdr2, _, _ := process.GetShardHeaderFromPoolWithNonce(0, 0, pools.Headers(), pools.HeadersNonces())
	assert.NotNil(t, bs)
	assert.True(t, hdr == hdr2)
}

func TestShardGetBlockFromPoolShouldReturnBlock(t *testing.T) {
	blk := make(block.MiniBlockSlice, 0)

	pools := createMockPools()

	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(key, []byte("aaa")) {
				return blk, true
			}

			return nil, false
		}
		return cs
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), &mock.SyncTimerMock{})

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		initBlockchain(),
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	mbHashes := make([][]byte, 0)
	mbHashes = append(mbHashes, []byte("aaaa"))

	mb := bs.GetMiniBlocks(mbHashes)
	assert.True(t, reflect.DeepEqual(blk, mb))

}

//------- testing received headers

func TestBootstrap_ReceivedHeadersFoundInPoolShouldAddToForkDetector(t *testing.T) {
	t.Parallel()

	addedHash := []byte("hash")
	addedHdr := &block.Header{}

	pools := createMockPools()
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}
		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(key, addedHash) {
				return addedHdr, true
			}

			return nil, false
		}
		return sds
	}

	wasAdded := false
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.AddHeaderCalled = func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) error {
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

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), &mock.SyncTimerMock{})

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		initBlockchain(),
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.ReceivedHeaders(addedHash)

	assert.True(t, wasAdded)
}

func TestBootstrap_ReceivedHeadersNotFoundInPoolShouldNotAddToForkDetector(t *testing.T) {
	t.Parallel()

	addedHash := []byte("hash")
	addedHdr := &block.Header{}

	pools := createMockPools()

	wasAdded := false
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.AddHeaderCalled = func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) error {
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

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	headerStorage := &mock.StorerStub{}
	headerStorage.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal(key, addedHash) {
			buff, _ := marshalizer.Marshal(addedHdr)

			return buff, nil
		}

		return nil, nil
	}

	store := createFullStore()
	store.AddStorer(dataRetriever.BlockHeaderUnit, headerStorage)

	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
	)

	_ = blkc.SetAppStatusHandler(&mock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {},
	})

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), &mock.SyncTimerMock{})

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.ReceivedHeaders(addedHash)

	assert.False(t, wasAdded)
}

//------- ForkChoice

func TestBootstrap_ForkChoiceNilBlockchainHeaderShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	err := bs.ForkChoice()
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestBootstrap_ForkChoiceNilParamHeaderShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return nil
	}

	err := bs.ForkChoice()
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestBootstrap_ForkChoiceIsNotEmptyShouldErr(t *testing.T) {
	t.Parallel()

	newHdrHash := []byte("new hdr hash")
	newHdrNonce := uint64(6)

	remFlags := &removedFlags{}
	shardId := uint32(0)

	pools := createMockPools()
	pools.HeadersCalled = func() storage.Cacher {
		return createHeadersDataPool(newHdrHash, remFlags)
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		return createHeadersNoncesDataPool(
			newHdrNonce,
			newHdrHash,
			newHdrNonce,
			remFlags,
			shardId,
		)
	}
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := createForkDetector(newHdrNonce, remFlags)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.Header{
			PubKeysBitmap: []byte("X"),
			Nonce:         newHdrNonce,
		}
	}

	err := bs.ForkChoice()
	assert.Equal(t, reflect.TypeOf(&sync.ErrSignedBlock{}), reflect.TypeOf(err))
}

func TestBootstrap_ForkChoiceIsEmptyCallRollBackOkValsShouldWork(t *testing.T) {
	t.Parallel()

	//retain if the remove process from different storage locations has been called
	remFlags := &removedFlags{}
	shardId := uint32(0)

	currentHdrNonce := uint64(8)
	currentHdrHash := []byte("current header hash")

	//define prev tx block body "strings" as in this test there are a lot of stubs that
	//constantly need to check some defined symbols
	//prevTxBlockBodyHash := []byte("prev block body hash")
	prevTxBlockBodyBytes := []byte("prev block body bytes")
	prevTxBlockBody := make(block.Body, 0)

	//define prev header "strings"
	prevHdrHash := []byte("prev header hash")
	prevHdrBytes := []byte("prev header bytes")
	prevHdrRootHash := []byte("prev header root hash")
	prevHdr := &block.Header{
		Signature: []byte("sig of the prev header as to be unique in this context"),
		RootHash:  prevHdrRootHash,
	}

	pools := createMockPools()

	//data pool headers
	pools.HeadersCalled = func() storage.Cacher {
		return createHeadersDataPool(currentHdrHash, remFlags)
	}
	//data pool headers-nonces
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		return createHeadersNoncesDataPool(
			currentHdrNonce,
			currentHdrHash,
			currentHdrNonce,
			remFlags,
			shardId,
		)
	}

	//a mock blockchain with special header and tx block bodies stubs (defined above)
	blkc := &mock.BlockChainMock{}

	store := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return prevHdrBytes, nil
				},
				RemoveCalled: func(key []byte) error {
					remFlags.flagHdrRemovedFromStorage = true
					return nil
				},
			}
		},
	}

	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{
		RestoreBlockIntoPoolsCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
			return nil
		},
	}

	hasher := &mock.HasherStub{
		ComputeCalled: func(s string) []byte {
			return currentHdrHash
		},
	}

	//a marshalizer stub
	marshalizer := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return []byte("X"), nil
		},
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			if bytes.Equal(buff, prevHdrBytes) {
				//bytes represent a header (strings are returns from hdrUnit.Get which is also a stub here)
				//copy only defined fields
				obj.(*block.Header).Signature = prevHdr.Signature
				obj.(*block.Header).RootHash = prevHdrRootHash
				return nil
			}
			if bytes.Equal(buff, prevTxBlockBodyBytes) {
				//bytes represent a tx block body (strings are returns from txBlockUnit.Get which is also a stub here)
				//copy only defined fields
				obj = prevTxBlockBody
				return nil
			}

			return nil
		},
	}

	forkDetector := createForkDetector(currentHdrNonce, remFlags)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.SetForkNonce(currentHdrNonce)

	hdr := &block.Header{
		Nonce: currentHdrNonce,
		//empty bitmap
		PrevHash: prevHdrHash,
	}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return hdr
	}
	blkc.SetCurrentBlockHeaderCalled = func(handler data.HeaderHandler) error {
		hdr = prevHdr
		return nil
	}

	body := make(block.Body, 0)
	blkc.GetCurrentBlockBodyCalled = func() data.BodyHandler {
		return body
	}
	blkc.SetCurrentBlockBodyCalled = func(handler data.BodyHandler) error {
		body = prevTxBlockBody
		return nil
	}

	hdrHash := make([]byte, 0)
	blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
		return hdrHash
	}
	blkc.SetCurrentBlockHeaderHashCalled = func(i []byte) {
		hdrHash = i
	}

	err := bs.ForkChoice()
	assert.Nil(t, err)
	assert.True(t, remFlags.flagHdrRemovedFromNonces)
	assert.True(t, remFlags.flagHdrRemovedFromHeaders)
	assert.True(t, remFlags.flagHdrRemovedFromStorage)
	assert.True(t, remFlags.flagHdrRemovedFromForkDetector)
	assert.Equal(t, blkc.GetCurrentBlockHeader(), prevHdr)
	assert.Equal(t, blkc.GetCurrentBlockBody(), prevTxBlockBody)
	assert.Equal(t, blkc.GetCurrentBlockHeaderHash(), prevHdrHash)
}

func TestBootstrap_ForkChoiceIsEmptyCallRollBackToGenesisShouldWork(t *testing.T) {
	t.Parallel()

	//retain if the remove process from different storage locations has been called
	remFlags := &removedFlags{}
	shardId := uint32(0)

	currentHdrNonce := uint64(1)
	currentHdrHash := []byte("current header hash")

	//define prev tx block body "strings" as in this test there are a lot of stubs that
	//constantly need to check some defined symbols
	//prevTxBlockBodyHash := []byte("prev block body hash")
	prevTxBlockBodyBytes := []byte("prev block body bytes")
	prevTxBlockBody := make(block.Body, 0)

	//define prev header "strings"
	prevHdrHash := []byte("prev header hash")
	prevHdrBytes := []byte("prev header bytes")
	prevHdrRootHash := []byte("prev header root hash")
	prevHdr := &block.Header{
		Signature: []byte("sig of the prev header as to be unique in this context"),
		RootHash:  prevHdrRootHash,
	}

	pools := createMockPools()

	//data pool headers
	pools.HeadersCalled = func() storage.Cacher {
		return createHeadersDataPool(currentHdrHash, remFlags)
	}
	//data pool headers-nonces
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		return createHeadersNoncesDataPool(
			currentHdrNonce,
			currentHdrHash,
			currentHdrNonce,
			remFlags,
			shardId,
		)
	}

	//a mock blockchain with special header and tx block bodies stubs (defined above)
	blkc := &mock.BlockChainMock{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return prevHdr
		},
	}
	store := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return prevHdrBytes, nil
				},
				RemoveCalled: func(key []byte) error {
					remFlags.flagHdrRemovedFromStorage = true
					return nil
				},
			}
		},
	}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{
		RestoreBlockIntoPoolsCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
			return nil
		},
	}

	hasher := &mock.HasherStub{
		ComputeCalled: func(s string) []byte {
			return currentHdrHash
		},
	}

	//a marshalizer stub
	marshalizer := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return []byte("X"), nil
		},
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			if bytes.Equal(buff, prevHdrBytes) {
				//bytes represent a header (strings are returns from hdrUnit.Get which is also a stub here)
				//copy only defined fields
				obj.(*block.Header).Signature = prevHdr.Signature
				obj.(*block.Header).RootHash = prevHdrRootHash
				return nil
			}
			if bytes.Equal(buff, prevTxBlockBodyBytes) {
				//bytes represent a tx block body (strings are returns from txBlockUnit.Get which is also a stub here)
				//copy only defined fields
				obj = prevTxBlockBody
				return nil
			}

			return nil
		},
	}

	forkDetector := createForkDetector(currentHdrNonce, remFlags)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.SetForkNonce(currentHdrNonce)

	hdr := &block.Header{
		Nonce: currentHdrNonce,
		//empty bitmap
		PrevHash: prevHdrHash,
	}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return hdr
	}
	blkc.SetCurrentBlockHeaderCalled = func(handler data.HeaderHandler) error {
		hdr = nil
		return nil
	}

	body := make(block.Body, 0)
	blkc.GetCurrentBlockBodyCalled = func() data.BodyHandler {
		return body
	}
	blkc.SetCurrentBlockBodyCalled = func(handler data.BodyHandler) error {
		body = nil
		return nil
	}

	hdrHash := make([]byte, 0)
	blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
		return hdrHash
	}
	blkc.SetCurrentBlockHeaderHashCalled = func(i []byte) {
		hdrHash = nil
	}

	err := bs.ForkChoice()
	assert.Nil(t, err)
	assert.True(t, remFlags.flagHdrRemovedFromNonces)
	assert.True(t, remFlags.flagHdrRemovedFromHeaders)
	assert.True(t, remFlags.flagHdrRemovedFromStorage)
	assert.True(t, remFlags.flagHdrRemovedFromForkDetector)
	assert.Nil(t, blkc.GetCurrentBlockHeader())
	assert.Nil(t, blkc.GetCurrentBlockBody())
	assert.Nil(t, blkc.GetCurrentBlockHeaderHash())
}

//------- GetTxBodyHavingHash

func TestBootstrap_GetTxBodyHavingHashReturnsFromCacherShouldWork(t *testing.T) {
	t.Parallel()

	mbh := []byte("requested hash")
	requestedHash := make([][]byte, 0)
	requestedHash = append(requestedHash, mbh)
	mb := &block.MiniBlock{}
	txBlock := make(block.MiniBlockSlice, 0)

	pools := createMockPools()
	pools.MiniBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{
			RegisterHandlerCalled: func(i func(key []byte)) {},
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal(key, mbh) {
					return mb, true
				}
				return nil, false
			},
		}
	}
	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
	)
	_ = blkc.SetAppStatusHandler(&mock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {},
	})
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)
	txBlockRecovered := bs.GetMiniBlocks(requestedHash)

	assert.True(t, reflect.DeepEqual(txBlockRecovered, txBlock))
}

func TestBootstrap_GetTxBodyHavingHashNotFoundInCacherOrStorageShouldRetNil(t *testing.T) {
	t.Parallel()

	mbh := []byte("requested hash")
	requestedHash := make([][]byte, 0)
	requestedHash = append(requestedHash, mbh)

	pools := createMockPools()

	txBlockUnit := &mock.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return nil, errors.New("not found")
		},
	}

	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
	)

	_ = blkc.SetAppStatusHandler(&mock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {},
	})

	store := createFullStore()
	store.AddStorer(dataRetriever.TransactionUnit, txBlockUnit)

	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderNilMiniBlocks(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)
	txBlockRecovered := bs.GetMiniBlocks(requestedHash)

	assert.Nil(t, txBlockRecovered)
}

func TestBootstrap_GetTxBodyHavingHashFoundInStorageShouldWork(t *testing.T) {
	t.Parallel()

	mbh := []byte("requested hash")
	requestedHash := make([][]byte, 0)
	requestedHash = append(requestedHash, mbh)
	txBlock := make(block.MiniBlockSlice, 0)

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}

	pools := createMockPools()

	txBlockUnit := &mock.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			if bytes.Equal(key, mbh) {
				buff, _ := marshalizer.Marshal(txBlock)
				return buff, nil
			}

			return nil, errors.New("not found")
		},
	}

	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
	)

	_ = blkc.SetAppStatusHandler(&mock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {},
	})
	store := createFullStore()
	store.AddStorer(dataRetriever.TransactionUnit, txBlockUnit)

	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)
	txBlockRecovered := bs.GetMiniBlocks(requestedHash)

	assert.Equal(t, txBlock, txBlockRecovered)
}

func TestBootstrap_AddSyncStateListenerShouldAppendAnotherListener(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := createBlockProcessor()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

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

	mutex := goSync.RWMutex{}
	pools := createMockPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := createBlockProcessor()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewShardBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

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

func TestBootstrap_LoadBlocksShouldErrBootstrapFromStorage(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}
	storageBootstrapper := NewStorageBootstrapperMock()

	store := createStore()
	store.HasCalled = func(unitType dataRetriever.UnitType, key []byte) error {
		return errors.New("key not found")
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.SetStorageBootstrapper(storageBootstrapper)

	err := bs.LoadBlocks(
		process.ShardBlockFinality,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()),
	)

	assert.Equal(t, process.ErrNotEnoughValidBlocksInStorage, err)
}

func TestBootstrap_LoadBlocksShouldErrBootstrapFromStorageWhenBlocksAreNotValid(t *testing.T) {
	t.Parallel()

	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}
	storageBootstrapper := NewStorageBootstrapperMock()
	storageBootstrapper.GetHeaderCalled = func(shardId uint32, nonce uint64) (data.HeaderHandler, []byte, error) {
		return nil, nil, errors.New("header not found")
	}

	store := createStore()
	store.HasCalled = func(unitType dataRetriever.UnitType, key []byte) error {
		if bytes.Equal(key, uint64Converter.ToByteSlice(1)) ||
			bytes.Equal(key, uint64Converter.ToByteSlice(2)) {
			return nil
		}

		return errors.New("key not found")
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.SetStorageBootstrapper(storageBootstrapper)

	err := bs.LoadBlocks(
		process.ShardBlockFinality,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()),
	)

	assert.Equal(t, process.ErrNotEnoughValidBlocksInStorage, err)
}

func TestBootstrap_LoadBlocksShouldErrWhenRecreateTrieFail(t *testing.T) {
	t.Parallel()

	wasCalled := false
	errExpected := errors.New("error to recreate trie")
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.Header{}
	}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) error {
			return nil
		},
	}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			wasCalled = true
			return errExpected
		},
	}
	storageBootstrapper := NewStorageBootstrapperMock()

	store := createStore()
	store.HasCalled = func(unitType dataRetriever.UnitType, key []byte) error {
		if bytes.Equal(key, uint64Converter.ToByteSlice(1)) ||
			bytes.Equal(key, uint64Converter.ToByteSlice(2)) {
			return nil
		}

		return errors.New("key not found")
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.SetStorageBootstrapper(storageBootstrapper)

	err := bs.LoadBlocks(
		process.ShardBlockFinality,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()),
	)

	assert.True(t, wasCalled)
	assert.Equal(t, process.ErrNotEnoughValidBlocksInStorage, err)
}

func TestBootstrap_LoadBlocksShouldWorkAfterRemoveInvalidBlocks(t *testing.T) {
	t.Parallel()

	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.Header{}
	}

	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) error {
			return nil
		},
	}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	storageBootstrapper := NewStorageBootstrapperMock()
	storageBootstrapper.GetHeaderCalled = func(shardId uint32, nonce uint64) (data.HeaderHandler, []byte, error) {
		if nonce == 3 {
			return nil, nil, errors.New("header not found")
		}
		return &block.Header{Nonce: 1}, []byte("hash"), nil
	}

	store := createStore()
	store.HasCalled = func(unitType dataRetriever.UnitType, key []byte) error {
		if bytes.Equal(key, uint64Converter.ToByteSlice(1)) ||
			bytes.Equal(key, uint64Converter.ToByteSlice(2)) ||
			bytes.Equal(key, uint64Converter.ToByteSlice(3)) {
			return nil
		}

		return errors.New("key not found")
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.SetStorageBootstrapper(storageBootstrapper)

	err := bs.LoadBlocks(
		process.ShardBlockFinality,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()),
	)

	assert.Nil(t, err)
}

func TestBootstrap_ApplyBlockShouldErrWhenHeaderIsNotFound(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("header not found")

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}
	storageBootstrapper := NewStorageBootstrapperMock()
	storageBootstrapper.GetHeaderCalled = func(shardId uint32, nonce uint64) (data.HeaderHandler, []byte, error) {
		return nil, nil, errExpected
	}

	store := createStore()

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.SetStorageBootstrapper(storageBootstrapper)

	err := bs.ApplyBlock(
		0,
		1)

	assert.Equal(t, errExpected, err)
}

func TestBootstrap_ApplyBlockShouldErrWhenBodyIsNotFound(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("body not found")

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) error {
			return nil
		},
	}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}
	storageBootstrapper := NewStorageBootstrapperMock()
	storageBootstrapper.GetBlockBodyCalled = func(handler data.HeaderHandler) (data.BodyHandler, error) {
		return nil, errExpected
	}

	store := createStore()

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.SetStorageBootstrapper(storageBootstrapper)

	err := bs.ApplyBlock(
		0,
		1)

	assert.Equal(t, errExpected, err)
}

func TestBootstrap_ApplyBlockShouldErrWhenSetCurrentBlockBodyFails(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("set block body failed")
	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	blkc.SetCurrentBlockBodyCalled = func(handler data.BodyHandler) error {
		return errExpected
	}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) error {
			return nil
		},
	}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}
	storageBootstrapper := NewStorageBootstrapperMock()

	store := createStore()

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.SetStorageBootstrapper(storageBootstrapper)

	err := bs.ApplyBlock(
		0,
		1)

	assert.Equal(t, errExpected, err)
}

func TestBootstrap_ApplyBlockShouldErrWhenSetCurrentBlockHeaderFails(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("set block header failed")
	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	blkc.SetCurrentBlockHeaderCalled = func(handler data.HeaderHandler) error {
		return errExpected
	}

	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) error {
			return nil
		},
	}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}
	storageBootstrapper := NewStorageBootstrapperMock()

	store := createStore()

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.SetStorageBootstrapper(storageBootstrapper)

	err := bs.ApplyBlock(
		0,
		1)

	assert.Equal(t, errExpected, err)
}

func TestBootstrap_ApplyBlockShouldWork(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) error {
			return nil
		},
	}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}
	storageBootstrapper := NewStorageBootstrapperMock()

	store := createStore()

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	bs.SetStorageBootstrapper(storageBootstrapper)

	err := bs.ApplyBlock(
		0,
		1)

	assert.Nil(t, err)
}

func TestBootstrap_RemoveBlockHeaderShouldErrNilHeadersStorage(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	store := createStore()
	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		if unitType == dataRetriever.BlockHeaderUnit {
			return nil
		}
		if unitType == dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()) {
			return &mock.StorerStub{}
		}

		return nil
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	err := bs.RemoveBlockHeader(1,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()))

	assert.Equal(t, process.ErrNilHeadersStorage, err)
}

func TestBootstrap_RemoveBlockHeaderShouldErrNilHeadersNonceHashStorage(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	store := createStore()
	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		if unitType == dataRetriever.BlockHeaderUnit {
			return &mock.StorerStub{}
		}
		if unitType == dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()) {
			return nil
		}

		return nil
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	err := bs.RemoveBlockHeader(1,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()))

	assert.Equal(t, process.ErrNilHeadersNonceHashStorage, err)
}

func TestBootstrap_RemoveBlockHeaderShouldErrWhenHeaderIsNotFound(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("key not found")
	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	store := createStore()
	store.GetCalled = func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
		return nil, errExpected
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	err := bs.RemoveBlockHeader(1,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()))

	assert.Equal(t, errExpected, err)
}

func TestBootstrap_RemoveBlockHeaderShouldErrWhenRemoveHeaderHashFails(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("remove header hash failed")
	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	store := createStore()
	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		if unitType == dataRetriever.BlockHeaderUnit {
			return &mock.StorerStub{
				RemoveCalled: func(key []byte) error {
					return errExpected
				},
			}
		}
		if unitType == dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()) {
			return &mock.StorerStub{
				RemoveCalled: func(key []byte) error {
					return nil
				},
			}
		}

		return nil
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	err := bs.RemoveBlockHeader(1,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()))

	assert.Equal(t, errExpected, err)
}

func TestBootstrap_RemoveBlockHeaderShouldErrWhenRemoveHeaderNonceFails(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("remove header nonce failed")
	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	store := createStore()
	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		if unitType == dataRetriever.BlockHeaderUnit {
			return &mock.StorerStub{
				RemoveCalled: func(key []byte) error {
					return nil
				},
			}
		}
		if unitType == dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()) {
			return &mock.StorerStub{
				RemoveCalled: func(key []byte) error {
					return errExpected
				},
			}
		}

		return nil
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	err := bs.RemoveBlockHeader(1,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()))

	assert.Equal(t, errExpected, err)
}

func TestBootstrap_RemoveBlockHeaderShouldWork(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	store := createStore()

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	err := bs.RemoveBlockHeader(1,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()))

	assert.Nil(t, err)
}

func TestBootstrap_SyncFromStorerShouldErrWhenLoadBlocksFails(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	store := createStore()
	store.HasCalled = func(unitType dataRetriever.UnitType, key []byte) error {
		return errors.New("key not found")
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	err := bs.SyncFromStorer(process.ShardBlockFinality,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()),
		process.MetaBlockFinality,
		dataRetriever.MetaHdrNonceHashDataUnit)

	assert.Equal(t, process.ErrNotEnoughValidBlocksInStorage, err)
}

func TestBootstrap_SyncFromStorerShouldWork(t *testing.T) {
	t.Parallel()

	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.Header{}
	}

	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{
		AddLastNotarizedHdrCalled: func(shardId uint32, processedHdr data.HeaderHandler) {
		},
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) error {
			return nil
		},
	}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}

	store := createStore()

	errKeyNotFound := errors.New("key not found")
	store.HasCalled = func(unitType dataRetriever.UnitType, key []byte) error {
		if bytes.Equal(key, uint64Converter.ToByteSlice(1)) ||
			bytes.Equal(key, uint64Converter.ToByteSlice(2)) ||
			bytes.Equal(key, uint64Converter.ToByteSlice(3)) ||
			bytes.Equal(key, uint64Converter.ToByteSlice(4)) ||
			bytes.Equal(key, uint64Converter.ToByteSlice(5)) ||
			bytes.Equal(key, uint64Converter.ToByteSlice(6)) {
			return nil
		}

		return errKeyNotFound
	}

	shardBlockHash3 := []byte("shardhash_3")
	shardBlockHash4 := []byte("shardhash_4")
	shardBlockHash5 := []byte("shardhash_5")
	shardBlockHash6 := []byte("shardhash_6")

	metaBlockHashes1 := make([][]byte, 0)
	metaBlockHash1 := []byte("metahash_1")
	metaBlockHashes1 = append(metaBlockHashes1, metaBlockHash1)

	metaBlockHashes2 := make([][]byte, 0)
	metaBlockHash2 := []byte("metahash_2")
	metaBlockHashes2 = append(metaBlockHashes2, metaBlockHash2)

	metaBlockHashes3 := make([][]byte, 0)
	metaBlockHash3 := []byte("metahash_3")
	metaBlockHashes3 = append(metaBlockHashes3, metaBlockHash3)

	metaBlockHashes4 := make([][]byte, 0)
	metaBlockHash4 := []byte("metahash_4")
	metaBlockHashes4 = append(metaBlockHashes4, metaBlockHash4)

	shardInfo := make([]block.ShardData, 0)
	shardInfo = append(shardInfo, block.ShardData{HeaderHash: shardBlockHash3})
	shardInfo = append(shardInfo, block.ShardData{HeaderHash: shardBlockHash4})
	shardInfo = append(shardInfo, block.ShardData{HeaderHash: shardBlockHash5})

	store.GetCalled = func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
		if unitType == dataRetriever.ShardHdrNonceHashDataUnit {
			if bytes.Equal(key, uint64Converter.ToByteSlice(3)) {
				return shardBlockHash3, nil
			}
			if bytes.Equal(key, uint64Converter.ToByteSlice(4)) {
				return shardBlockHash4, nil
			}
			if bytes.Equal(key, uint64Converter.ToByteSlice(5)) {
				return shardBlockHash5, nil
			}
			if bytes.Equal(key, uint64Converter.ToByteSlice(6)) {
				return shardBlockHash6, nil
			}
		}

		if unitType == dataRetriever.MetaHdrNonceHashDataUnit {
			if bytes.Equal(key, uint64Converter.ToByteSlice(1)) {
				return metaBlockHash1, nil
			}
			if bytes.Equal(key, uint64Converter.ToByteSlice(2)) {
				return metaBlockHash2, nil
			}
			if bytes.Equal(key, uint64Converter.ToByteSlice(3)) {
				return metaBlockHash3, nil
			}
			if bytes.Equal(key, uint64Converter.ToByteSlice(4)) {
				return metaBlockHash4, nil
			}
		}

		return nil, errKeyNotFound
	}

	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		if unitType == dataRetriever.BlockHeaderUnit {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, shardBlockHash3) {
						buff, _ := marshalizer.Marshal(&block.Header{Nonce: 3, MetaBlockHashes: metaBlockHashes1})
						return buff, nil
					}
					if bytes.Equal(key, shardBlockHash4) {
						buff, _ := marshalizer.Marshal(&block.Header{Nonce: 4, MetaBlockHashes: metaBlockHashes2})
						return buff, nil
					}
					if bytes.Equal(key, shardBlockHash5) {
						buff, _ := marshalizer.Marshal(&block.Header{Nonce: 5, MetaBlockHashes: metaBlockHashes3})
						return buff, nil
					}
					if bytes.Equal(key, shardBlockHash6) {
						buff, _ := marshalizer.Marshal(&block.Header{Nonce: 6, MetaBlockHashes: metaBlockHashes4})
						return buff, nil
					}

					return nil, errKeyNotFound
				},
			}
		}

		if unitType == dataRetriever.MetaBlockUnit {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, metaBlockHash1) {
						buff, _ := marshalizer.Marshal(&block.MetaBlock{Nonce: 1})
						return buff, nil
					}
					if bytes.Equal(key, metaBlockHash2) {
						buff, _ := marshalizer.Marshal(&block.MetaBlock{Nonce: 2})
						return buff, nil
					}
					if bytes.Equal(key, metaBlockHash3) {
						buff, _ := marshalizer.Marshal(&block.MetaBlock{Nonce: 3, ShardInfo: shardInfo})
						return buff, nil
					}
					if bytes.Equal(key, metaBlockHash4) {
						buff, _ := marshalizer.Marshal(&block.MetaBlock{Nonce: 4})
						return buff, nil
					}

					return nil, errKeyNotFound
				},
			}
		}

		return nil
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	err := bs.SyncFromStorer(process.ShardBlockFinality,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()),
		process.MetaBlockFinality,
		dataRetriever.MetaHdrNonceHashDataUnit)

	assert.Nil(t, err)
}

func TestBootstrap_ApplyNotarizedBlockShouldErrWhenGetFinalNotarizedMetaHeaderFromStorageFails(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	nonceToByteSlice := uint64Converter.ToByteSlice(uint64(1))

	store := createStore()
	errKeyNotFound := errors.New("key not found")
	hash := []byte("hash")
	buff := []byte("buff")
	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		if unitType == dataRetriever.MetaBlockUnit {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if !bytes.Equal(key, hash) {
						return nil, errKeyNotFound
					}

					return buff, nil
				},
			}
		}

		return nil
	}

	store.GetCalled = func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
		if unitType == dataRetriever.MetaHdrNonceHashDataUnit {
			if bytes.Equal(key, nonceToByteSlice) {
				return nil, errKeyNotFound
			}
			return []byte("hash"), nil
		}

		return nil, errKeyNotFound
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	lastNotarized := make(map[uint32]uint64, 0)
	finalNotarized := make(map[uint32]uint64, 0)

	lastNotarized[sharding.MetachainShardId] = 1
	finalNotarized[sharding.MetachainShardId] = 1

	err := bs.ApplyNotarizedBlocks(finalNotarized, lastNotarized)

	assert.Equal(t, errKeyNotFound, err)
}

func TestBootstrap_ApplyNotarizedBlockShouldErrWhenGetLastNotarizedMetaHeaderFromStorageFails(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{
		AddLastNotarizedHdrCalled: func(shardId uint32, processedHdr data.HeaderHandler) {
		},
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return nil
		},
	}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	nonceToByteSlice := uint64Converter.ToByteSlice(uint64(1))

	store := createStore()
	errKeyNotFound := errors.New("key not found")
	hash := []byte("hash")
	buff := []byte("buff")
	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		if unitType == dataRetriever.MetaBlockUnit {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if !bytes.Equal(key, hash) {
						return nil, errKeyNotFound
					}

					return buff, nil
				},
			}
		}

		return nil
	}

	store.GetCalled = func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
		if unitType == dataRetriever.MetaHdrNonceHashDataUnit {
			if bytes.Equal(key, nonceToByteSlice) {
				return nil, errKeyNotFound
			}
			return []byte("hash"), nil
		}

		return nil, errKeyNotFound
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	lastNotarized := make(map[uint32]uint64, 0)
	finalNotarized := make(map[uint32]uint64, 0)

	lastNotarized[sharding.MetachainShardId] = 1
	finalNotarized[sharding.MetachainShardId] = 2

	err := bs.ApplyNotarizedBlocks(finalNotarized, lastNotarized)

	assert.Equal(t, errKeyNotFound, err)
}

func TestBootstrap_ApplyNotarizedBlockShouldWork(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{
		AddLastNotarizedHdrCalled: func(shardId uint32, processedHdr data.HeaderHandler) {
		},
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return nil
		},
	}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	nonceToByteSlice := uint64Converter.ToByteSlice(uint64(0))

	store := createStore()
	errKeyNotFound := errors.New("key not found")
	hash := []byte("hash")
	buff := []byte("buff")
	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		if unitType == dataRetriever.MetaBlockUnit {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if !bytes.Equal(key, hash) {
						return nil, errKeyNotFound
					}

					return buff, nil
				},
			}
		}

		return nil
	}

	store.GetCalled = func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
		if unitType == dataRetriever.MetaHdrNonceHashDataUnit {
			if bytes.Equal(key, nonceToByteSlice) {
				return nil, errKeyNotFound
			}
			return []byte("hash"), nil
		}

		return nil, errKeyNotFound
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	lastNotarized := make(map[uint32]uint64, 0)
	finalNotarized := make(map[uint32]uint64, 0)

	finalNotarized[sharding.MetachainShardId] = 1
	lastNotarized[sharding.MetachainShardId] = 1

	err := bs.ApplyNotarizedBlocks(finalNotarized, lastNotarized)

	assert.Nil(t, err)
}

func TestBootstrap_RemoveNotarizedBlockShouldErrNilHeadersStorage(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	store := createStore()
	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		return nil
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	err := bs.RemoveBlockHeader(
		1,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()))

	assert.Equal(t, process.ErrNilHeadersStorage, err)
}

func TestBootstrap_RemoveNotarizedBlockShouldErrNilHeadersNonceHashStorage(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	store := createStore()
	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		if unitType == dataRetriever.BlockHeaderUnit {
			return &mock.StorerStub{}
		}

		return nil
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	err := bs.RemoveBlockHeader(
		1,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()))

	assert.Equal(t, process.ErrNilHeadersNonceHashStorage, err)
}

func TestBootstrap_RemoveNotarizedBlockShouldErrWhenRemoveHeaderHashFails(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("remove header hash failed")
	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	store := createStore()
	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		if unitType == dataRetriever.BlockHeaderUnit {
			return &mock.StorerStub{
				RemoveCalled: func(key []byte) error {
					return errExpected
				},
			}
		}

		if unitType == dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()) {
			return &mock.StorerStub{
				RemoveCalled: func(key []byte) error {
					return nil
				},
			}
		}

		return nil
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	err := bs.RemoveBlockHeader(
		1,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()))

	assert.Equal(t, errExpected, err)
}

func TestBootstrap_RemoveNotarizedBlockShouldErrWhenRemoveHeaderNonceFails(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("remove header nonce failed")
	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	store := createStore()
	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		if unitType == dataRetriever.BlockHeaderUnit {
			return &mock.StorerStub{
				RemoveCalled: func(key []byte) error {
					return nil
				},
			}
		}

		if unitType == dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()) {
			return &mock.StorerStub{
				RemoveCalled: func(key []byte) error {
					return errExpected
				},
			}
		}

		return nil
	}

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	err := bs.RemoveBlockHeader(
		1,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()))

	assert.Equal(t, errExpected, err)
}

func TestBootstrap_RemoveNotarizedBlockShouldWork(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	store := createStore()

	bs, _ := sync.NewShardBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinder(),
		shardCoordinator,
		account,
		math.MaxUint32,
	)

	err := bs.RemoveBlockHeader(
		1,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()))

	assert.Nil(t, err)
}

func NewStorageBootstrapperMock() *sync.StorageBootstrapperMock {
	sbm := sync.StorageBootstrapperMock{
		GetHeaderCalled: func(shardId uint32, nonce uint64) (data.HeaderHandler, []byte, error) {
			return &block.Header{ShardId: shardId, Nonce: nonce, Round: 2}, []byte("hash"), nil
		},
		GetBlockBodyCalled: func(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
			if headerHandler != nil {
				return &block.Body{}, nil
			}

			return nil, errors.New("get block body failed")
		},
		RemoveBlockBodyCalled: func(nonce uint64, blockUnit dataRetriever.UnitType, hdrNonceHashDataUnit dataRetriever.UnitType) error {
			fmt.Printf("remove block body with nonce %d with hdr nonce hash data unit type %d and block unit type %d\n",
				nonce,
				hdrNonceHashDataUnit,
				blockUnit)

			return nil
		},
		GetNonceWithLastNotarizedCalled: func(currentNonce uint64) (uint64, map[uint32]uint64, map[uint32]uint64) {
			lastNotarized := make(map[uint32]uint64, 0)
			finalNotarized := make(map[uint32]uint64, 0)
			nonceWithLastNotarized := currentNonce

			fmt.Printf("get last notarized at nonce %d\n", currentNonce)

			return nonceWithLastNotarized, finalNotarized, lastNotarized
		},
		ApplyNotarizedBlocksCalled: func(finalNotarized map[uint32]uint64, lastNotarized map[uint32]uint64) error {
			fmt.Printf("final notarized items: %d and last notarized items: %d\n",
				len(finalNotarized),
				len(lastNotarized))

			return nil
		},
		CleanupNotarizedStorageCalled: func(lastNotarized map[uint32]uint64) {
			fmt.Printf("last notarized items: %d\n", len(lastNotarized))
		},
		AddHeaderToForkDetectorCalled: func(shardId uint32, nonce uint64, lastNotarizedMeta uint64) {
			fmt.Printf("add header to fork detector called")
		},
	}

	return &sbm
}
