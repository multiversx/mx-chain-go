package sync_test

import (
	"bytes"
	"errors"
	"math"
	"reflect"
	"strings"
	goSync "sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/sync"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/storageUnit"
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
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{
			GetCalled: func(u uint64) (bytes interface{}, b bool) {
				return nil, false
			},
			RegisterHandlerCalled: func(handler func(nonce uint64)) {},
			RemoveCalled: func(u uint64) {
				return
			},
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
			blk.SetCurrentBlockHeader(hdr.(*block.Header))
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
	remFlags *removedFlags) dataRetriever.Uint64Cacher {

	hnc := &mock.Uint64CacherStub{
		RegisterHandlerCalled: func(handler func(nonce uint64)) {},
		GetCalled: func(u uint64) (i interface{}, b bool) {
			if u == getNonceCompare {
				return getRetHash, true
			}

			return nil, false
		},
		RemoveCalled: func(u uint64) {
			if u == removedNonce {
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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilHeadersDataPool, err)
}

func TestNewShardBootstrap_PoolsHolderRetNilOnHeadersNoncesShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewShardBootstrap_NilHeaderResolverShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()

	errExpected := errors.New("expected error")

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
	)

	assert.Nil(t, bs)
	assert.Equal(t, errExpected, err)
}

func TestNewShardBootstrap_NilTxBlockBodyResolverShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()

	errExpected := errors.New("expected error")

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
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
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

	blkc.CurrentBlockHeader = &hdr

	pools := createMockPools()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return true, math.MaxUint64
	}
	forkDetector.RemoveHeadersCalled = func(nonce uint64, hash []byte) {
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return uint64(hdr.Nonce)
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 100
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	)

	r := bs.SyncBlock()

	assert.Equal(t, &sync.ErrSignedBlock{CurrentNonce: hdr.Nonce}, r)
}

func TestBootstrap_ShouldReturnMissingHeader(t *testing.T) {
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
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return false, math.MaxUint64
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 100
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(),
		time.Now().Add(2*time.Duration(100*time.Millisecond)),
		time.Duration(100*time.Millisecond),
		mock.SyncTimerMock{})

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
	)

	r := bs.SyncBlock()

	assert.Equal(t, process.ErrMissingHashForHeaderNonce, r)
}

func TestBootstrap_ShouldReturnMissingBody(t *testing.T) {
	t.Parallel()

	hdr := block.Header{Nonce: 1}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	pools := createMockPools()
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("aaa"), key) {
				return &block.Header{Nonce: 2}, true
			}

			return nil, false
		}

		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}

		hnc.GetCalled = func(u uint64) (bytes interface{}, b bool) {
			if u == 2 {
				return []byte("aaa"), false
			}

			return nil, false
		}
		return hnc
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return false, math.MaxUint64
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 2
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(),
		time.Now().Add(2*time.Duration(100*time.Millisecond)),
		time.Duration(100*time.Millisecond),
		mock.SyncTimerMock{})

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
	)

	bs.RequestHeader(2)
	r := bs.SyncBlock()
	assert.Equal(t, process.ErrMissingBody, r)
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
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return false, math.MaxUint64
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return uint64(hdr.Nonce)
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 1
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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

	mutDataAvailable := goSync.RWMutex{}
	dataAvailable := false

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			mutDataAvailable.RLock()
			defer mutDataAvailable.RUnlock()

			if bytes.Equal([]byte("aaa"), key) && dataAvailable {
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
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}

		hnc.GetCalled = func(u uint64) (bytes interface{}, b bool) {
			mutDataAvailable.RLock()
			defer mutDataAvailable.RUnlock()

			if u == 2 && dataAvailable {
				return []byte("aaa"), false
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
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return false, math.MaxUint64
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return uint64(hdr.Nonce)
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 2
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	account.RootHashCalled = func() []byte {
		return nil
	}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(200*time.Millisecond), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("aaa"), key) {
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
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}

		hnc.GetCalled = func(u uint64) (bytes interface{}, b bool) {
			if u == 2 {
				return []byte("aaa"), false
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
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return false, math.MaxUint64
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 2
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(),
		time.Now().Add(2*time.Duration(100*time.Millisecond)),
		time.Duration(100*time.Millisecond),
		mock.SyncTimerMock{})

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

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("aaa"), key) {
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
		sds.RemoveCalled = func(key []byte) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}
		hnc.RemoveCalled = func(u uint64) {
		}
		hnc.GetCalled = func(u uint64) (bytes interface{}, b bool) {
			if u == 2 {
				return []byte("aaa"), false
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
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return false, math.MaxUint64
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return uint64(hdr.Nonce)
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 2
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(),
		time.Now().Add(2*time.Duration(100*time.Millisecond)),
		time.Duration(100*time.Millisecond),
		mock.SyncTimerMock{})

	ebm.ProcessBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
		return process.ErrInvalidBlockHash
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
	)

	err := bs.SyncBlock()
	assert.Equal(t, &sync.ErrSignedBlock{CurrentNonce: hdr.Nonce}, err)
}

func TestBootstrap_ShouldSyncShouldReturnFalseWhenCurrentBlockIsNilAndRoundIndexIsZero(t *testing.T) {
	t.Parallel()

	pools := createMockPools()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{
		CheckForkCalled: func() (bool, uint64) {
			return false, math.MaxUint64
		},
		ProbableHighestNonceCalled: func() uint64 {
			return 0
		}}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	)

	assert.False(t, bs.ShouldSync())
}

func TestBootstrap_ShouldReturnTrueWhenCurrentBlockIsNilAndRoundIndexIsGreaterThanZero(t *testing.T) {
	t.Parallel()

	pools := createMockPools()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return false, math.MaxUint64
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 1
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return false, math.MaxUint64
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 0
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return false, math.MaxUint64
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 1
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	)

	assert.True(t, bs.ShouldSync())
}

func TestBootstrap_GetHeaderFromPoolShouldReturnNil(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return false, math.MaxUint64
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	)

	hdr, _ := bs.GetHeaderFromPoolWithNonce(0)
	assert.Nil(t, hdr)
}

func TestBootstrap_GetHeaderFromPoolShouldReturnHeader(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{Nonce: 0}

	pools := createMockPools()
	pools.HeadersCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("aaa"), key) {
				return hdr, true
			}
			return nil, false
		}

		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}
		hnc.GetCalled = func(u uint64) (i interface{}, b bool) {
			if u == 0 {
				return []byte("aaa"), true
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

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	)

	hdr2, _ := bs.GetHeaderFromPoolWithNonce(0)
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

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	forkDetector.AddHeaderCalled = func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
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

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	)

	bs.ReceivedHeaders(addedHash)

	assert.True(t, wasAdded)
}

func TestBootstrap_ReceivedHeadersNotFoundInPoolButFoundInStorageShouldAddToForkDetector(t *testing.T) {
	t.Parallel()

	addedHash := []byte("hash")
	addedHdr := &block.Header{}

	pools := createMockPools()

	wasAdded := false
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.AddHeaderCalled = func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
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

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	)

	bs.ReceivedHeaders(addedHash)

	assert.True(t, wasAdded)
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

	pools := createMockPools()
	pools.HeadersCalled = func() storage.Cacher {
		return createHeadersDataPool(newHdrHash, remFlags)
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
		return createHeadersNoncesDataPool(newHdrNonce, newHdrHash, newHdrNonce, remFlags)
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
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
		return createHeadersNoncesDataPool(
			currentHdrNonce,
			currentHdrHash,
			currentHdrNonce,
			remFlags,
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

	hasher := &mock.HasherMock{}

	//a marshalizer stub
	marshalizer := &mock.MarshalizerStub{
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
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
		return createHeadersNoncesDataPool(
			currentHdrNonce,
			currentHdrHash,
			currentHdrNonce,
			remFlags,
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

	hasher := &mock.HasherMock{}

	//a marshalizer stub
	marshalizer := &mock.MarshalizerStub{
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
