package sync_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	goSync "sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus/round"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

// waitTime defines the time in milliseconds until node waits the requested info from the network
const waitTime = 100 * time.Millisecond

type removedFlags struct {
	flagHdrRemovedFromNonces       bool
	flagHdrRemovedFromHeaders      bool
	flagHdrRemovedFromStorage      bool
	flagHdrRemovedFromForkDetector bool
}

func createMockResolversFinder() *mock.ResolversFinderStub {
	return &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
			if strings.Contains(baseTopic, factory.MiniBlocksTopic) {
				return &mock.MiniBlocksResolverMock{
					GetMiniBlocksCalled: func(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
						return make(block.MiniBlockSlice, 0), make([][]byte, 0)
					},
					GetMiniBlocksFromPoolCalled: func(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
						return make(block.MiniBlockSlice, 0), make([][]byte, 0)
					},
				}, nil
			}

			return nil, nil
		},
		CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
			if strings.Contains(baseTopic, factory.ShardBlocksTopic) {
				return &mock.HeaderResolverMock{
					RequestDataFromNonceCalled: func(nonce uint64) error {
						return nil
					},
					RequestDataFromHashCalled: func(hash []byte) error {
						return nil
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
			if strings.Contains(baseTopic, factory.MiniBlocksTopic) {
				return &mock.MiniBlocksResolverMock{
					RequestDataFromHashCalled: func(hash []byte) error {
						return nil
					},
					RequestDataFromHashArrayCalled: func(hash [][]byte) error {
						return nil
					},
					GetMiniBlocksCalled: func(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
						return make(block.MiniBlockSlice, 0), [][]byte{[]byte("hash")}
					},
					GetMiniBlocksFromPoolCalled: func(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
						return make(block.MiniBlockSlice, 0), [][]byte{[]byte("hash")}
					},
				}, nil
			}

			return nil, nil
		},
		CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
			if strings.Contains(baseTopic, factory.ShardBlocksTopic) {
				return &mock.HeaderResolverMock{
					RequestDataFromNonceCalled: func(nonce uint64) error {
						return nil
					},
					RequestDataFromHashCalled: func(hash []byte) error {
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
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{}
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

func createForkDetector(removedNonce uint64, remFlags *removedFlags) process.ForkDetector {
	return &mock.ForkDetectorMock{
		RemoveHeadersCalled: func(nonce uint64, hash []byte) {
			if nonce == removedNonce {
				remFlags.flagHdrRemovedFromForkDetector = true
			}
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return removedNonce
		},
		ProbableHighestNonceCalled: func() uint64 {
			return uint64(0)
		},
		GetNotarizedHeaderHashCalled: func(nonce uint64) []byte {
			return nil
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilPoolsHolder, err)
}

func TestNewShardBootstrap_PoolsHolderRetNilOnHeadersShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockPools()
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilHeadersDataPool, err)
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewShardBootstrap_NilBlackListHandlerShouldErr(t *testing.T) {
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
		&mock.ResolversFinderStub{},
		shardCoordinator,
		account,
		nil,
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilBlackListHandler, err)
}

func TestNewShardBootstrap_NilHeaderResolverShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	pools := createMockPools()
	resFinder := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
			if strings.Contains(baseTopic, factory.MiniBlocksTopic) {
				return &mock.ResolverStub{}, nil
			}

			return nil, nil
		},
		CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
			if strings.Contains(baseTopic, factory.ShardBlocksTopic) {
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
			if strings.Contains(baseTopic, factory.MiniBlocksTopic) {
				return nil, errExpected
			}

			return nil, nil
		},
		CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
			if strings.Contains(baseTopic, factory.ShardBlocksTopic) {
				return &mock.HeaderResolverMock{}, errExpected
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, errExpected, err)
}

func TestNewShardBootstrap_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := 0

	pools := &mock.PoolsHolderStub{}
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	assert.NotNil(t, bs)
	assert.Nil(t, err)
	assert.Equal(t, 2, wasCalled)
	assert.False(t, bs.IsInterfaceNil())
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
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return &process.ForkInfo{
			IsDetected: true,
			Nonce:      90,
			Round:      90,
			Hash:       []byte("hash"),
		}
	}
	forkDetector.RemoveHeadersCalled = func(nonce uint64, hash []byte) {
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return hdr.Nonce
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 100
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), 100*time.Millisecond, &mock.SyncTimerMock{})

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	r := bs.SyncBlock()

	assert.Equal(t, process.ErrNilHeadersStorage, r)
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
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 100
	}
	forkDetector.GetNotarizedHeaderHashCalled = func(nonce uint64) []byte {
		return nil
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(),
		time.Now().Add(2*100*time.Millisecond),
		100*time.Millisecond,
		&mock.SyncTimerMock{},
	)

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
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

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(),
		time.Now().Add(2*100*time.Millisecond),
		100*time.Millisecond,
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), 100*time.Millisecond, &mock.SyncTimerMock{})

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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

	hash := []byte("aaa")

	mutDataAvailable := goSync.RWMutex{}
	dataAvailable := false

	pools := &mock.PoolsHolderStub{}
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

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	account.RootHashCalled = func() ([]byte, error) {
		return nil, nil
	}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(200*time.Millisecond), 100*time.Millisecond, &mock.SyncTimerMock{})

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	bs.StartSync()

	time.Sleep(200 * time.Millisecond)

	mutDataAvailable.Lock()
	dataAvailable = true
	mutDataAvailable.Unlock()

	time.Sleep(500 * time.Millisecond)

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

	hash := []byte("aaa")
	header := &block.Header{
		Nonce:         2,
		Round:         1,
		BlockBodyType: block.TxBlock,
		RootHash:      []byte("bbb")}

	pools := &mock.PoolsHolderStub{}
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
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 2
	}
	forkDetector.GetNotarizedHeaderHashCalled = func(nonce uint64) []byte {
		return nil
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(),
		time.Now().Add(2*100*time.Millisecond),
		100*time.Millisecond,
		&mock.SyncTimerMock{},
	)

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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

	hash := []byte("aaa")
	header := &block.Header{
		Nonce:         2,
		Round:         1,
		BlockBodyType: block.TxBlock,
		RootHash:      []byte("bbb")}

	pools := &mock.PoolsHolderStub{}
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
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return hdr.Nonce
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 2
	}
	forkDetector.RemoveHeadersCalled = func(nonce uint64, hash []byte) {}
	forkDetector.ResetProbableHighestNonceCalled = func() {}
	forkDetector.GetNotarizedHeaderHashCalled = func(nonce uint64) []byte {
		return nil
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(),
		time.Now().Add(2*100*time.Millisecond),
		100*time.Millisecond,
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
		CheckForkCalled: func() *process.ForkInfo {
			return process.NewForkInfo()
		},
		ProbableHighestNonceCalled: func() uint64 {
			return 0
		},
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), 100*time.Millisecond, &mock.SyncTimerMock{})

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	assert.False(t, bs.ShouldSync())
}

func TestBootstrap_ShouldReturnTrueWhenCurrentBlockIsNilAndRoundIndexIsGreaterThanZero(t *testing.T) {
	t.Parallel()

	pools := createMockPools()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 1
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), 100*time.Millisecond, &mock.SyncTimerMock{})

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 0
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), 100*time.Millisecond, &mock.SyncTimerMock{})

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 1
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), 100*time.Millisecond, &mock.SyncTimerMock{})

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	rounder := &mock.RounderMock{}
	rounder.RoundIndex = 2
	forkDetector, _ := sync.NewShardForkDetector(rounder, &mock.BlackListHandlerStub{}, 0)
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	_ = forkDetector.AddHeader(&hdr1, hash1, process.BHProcessed, nil, nil, false)
	_ = forkDetector.AddHeader(&hdr2, hash2, process.BHNotarized, finalHeaders, finalHeadersHashes, false)

	shouldSync := bs.ShouldSync()
	assert.True(t, shouldSync)
	assert.True(t, bs.IsForkDetected())

	if shouldSync && bs.IsForkDetected() {
		forkDetector.RemoveHeaders(hdr1.GetNonce(), hash1)
		bs.ReceivedHeaders(&hdr1, hash1)
		_ = forkDetector.AddHeader(&hdr1, hash1, process.BHProcessed, nil, nil, false)
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

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	rounder := &mock.RounderMock{}
	rounder.RoundIndex = 2
	forkDetector, _ := sync.NewShardForkDetector(rounder, &mock.BlackListHandlerStub{}, 0)
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	_ = forkDetector.AddHeader(&hdr1, hash1, process.BHProcessed, nil, nil, false)
	_ = forkDetector.AddHeader(&hdr2, hash2, process.BHNotarized, finalHeaders, finalHeadersHashes, false)

	shouldSync := bs.ShouldSync()
	assert.True(t, shouldSync)
	assert.True(t, bs.IsForkDetected())

	if shouldSync && bs.IsForkDetected() {
		forkDetector.RemoveHeaders(hdr1.GetNonce(), hash1)
		bs.ReceivedHeaders(&hdr2, hash2)
		_ = forkDetector.AddHeader(&hdr2, hash2, process.BHProcessed, finalHeaders, finalHeadersHashes, false)
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
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), 100*time.Millisecond, &mock.SyncTimerMock{})

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	hdr, _, _ := process.GetShardHeaderFromPoolWithNonce(0, 0, pools.Headers())
	assert.NotNil(t, bs)
	assert.Nil(t, hdr)
}

func TestBootstrap_GetHeaderFromPoolShouldReturnHeader(t *testing.T) {
	t.Parallel()

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

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	account := &mock.AccountsStub{}

	shardCoordinator := mock.NewOneShardCoordinatorMock()

	rnd, _ := round.NewRound(time.Now(), time.Now(), 100*time.Millisecond, &mock.SyncTimerMock{})

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	hdr2, _, _ := process.GetShardHeaderFromPoolWithNonce(0, 0, pools.Headers())
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

	rnd, _ := round.NewRound(time.Now(), time.Now(), 100*time.Millisecond, &mock.SyncTimerMock{})

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	mbHashes := make([][]byte, 0)
	mbHashes = append(mbHashes, []byte("aaaa"))

	mb, _ := bs.GetMiniBlocks(mbHashes)
	assert.True(t, reflect.DeepEqual(blk, mb))

}

//------- testing received headers

func TestBootstrap_ReceivedHeadersFoundInPoolShouldAddToForkDetector(t *testing.T) {
	t.Parallel()

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

	wasAdded := false
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.AddHeaderCalled = func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte, isNotarizedShardStuck bool) error {
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

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), 100*time.Millisecond, &mock.SyncTimerMock{})

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	bs.ReceivedHeaders(addedHdr, addedHash)

	assert.True(t, wasAdded)
}

//------- RollBack

func TestBootstrap_RollBackNilBlockchainHeaderShouldErr(t *testing.T) {
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	err := bs.RollBack(false)
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestBootstrap_RollBackNilParamHeaderShouldErr(t *testing.T) {
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return nil
	}

	err := bs.RollBack(false)
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestBootstrap_RollBackIsNotEmptyShouldErr(t *testing.T) {
	t.Parallel()

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.Header{
			PubKeysBitmap: []byte("X"),
			Nonce:         newHdrNonce,
		}
	}

	err := bs.RollBack(false)
	assert.Equal(t, sync.ErrRollBackBehindFinalHeader, err)
}

func TestBootstrap_RollBackIsEmptyCallRollBackOneBlockOkValsShouldWork(t *testing.T) {
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
				_, ok := obj.(*block.Header)
				if !ok {
					return nil
				}

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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

	err := bs.RollBack(true)
	assert.Nil(t, err)
	assert.True(t, remFlags.flagHdrRemovedFromHeaders)
	assert.True(t, remFlags.flagHdrRemovedFromStorage)
	assert.True(t, remFlags.flagHdrRemovedFromForkDetector)
	assert.Equal(t, blkc.GetCurrentBlockHeader(), prevHdr)
	assert.Equal(t, blkc.GetCurrentBlockBody(), prevTxBlockBody)
	assert.Equal(t, blkc.GetCurrentBlockHeaderHash(), prevHdrHash)
}

func TestBootstrap_RollbackIsEmptyCallRollBackOneBlockToGenesisShouldWork(t *testing.T) {
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
				_, ok := obj.(*block.Header)
				if !ok {
					return nil
				}

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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

	err := bs.RollBack(true)
	assert.Nil(t, err)
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)
	txBlockRecovered, _ := bs.GetMiniBlocks(requestedHash)

	assert.True(t, reflect.DeepEqual(txBlockRecovered, txBlock))
}

func TestBootstrap_GetTxBodyHavingHashNotFoundInCacherOrStorageShouldRetEmptySlice(t *testing.T) {
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)
	txBlockRecovered, _ := bs.GetMiniBlocks(requestedHash)

	assert.Equal(t, 0, len(txBlockRecovered))
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)
	txBlockRecovered, _ := bs.GetMiniBlocks(requestedHash)

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
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

func TestShardBootstrap_SetStatusHandlerNilHandlerShouldErr(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		sds := &mock.HeadersCacherStub{}

		sds.AddCalled = func(headerHash []byte, header data.HeaderHandler) {
			assert.Fail(t, "should have not reached this point")
		}

		sds.RegisterHandlerCalled = func(func(header data.HeaderHandler, key []byte)) {
		}

		return sds
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {}

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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	err := bs.SetStatusHandler(nil)
	assert.Equal(t, process.ErrNilAppStatusHandler, err)

}

func TestShardBootstrap_RequestMiniBlocksFromHeaderWithNonceIfMissing(t *testing.T) {
	t.Parallel()

	requestDataWasCalled := false
	hdrHash := []byte("hash")
	hdr := &block.Header{Round: 5, Nonce: 1}
	pools := &mock.PoolsHolderStub{}
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
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}

	blkc := initBlockchain()
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.Header{Round: 10}
	}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return uint64(5)
	}
	resFinder := createMockResolversFinderNilMiniBlocks()
	resFinder.IntraShardResolverCalled = func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
		if strings.Contains(baseTopic, factory.ShardBlocksTopic) {
			return &mock.HeaderResolverMock{
				RequestDataFromHashCalled: func(hash []byte) error {
					return nil
				},
			}, nil
		}

		if strings.Contains(baseTopic, factory.MiniBlocksTopic) {
			return &mock.MiniBlocksResolverMock{
				RequestDataFromHashArrayCalled: func(hash [][]byte) error {
					requestDataWasCalled = true
					return nil
				},
				GetMiniBlocksFromPoolCalled: func(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
					return make(block.MiniBlockSlice, 0), [][]byte{[]byte("hash")}
				},
			}, nil
		}

		return nil, nil
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	store := createStore()
	store.GetCalled = func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
		nonceToBytes := mock.NewNonceHashConverterMock().ToByteSlice(uint64(1))
		if bytes.Equal(key, nonceToBytes) {
			return []byte("hdr"), nil
		}
		if bytes.Equal(key, []byte("hdr")) {
			hdr := block.Header{}
			mshlzdHdr, _ := json.Marshal(hdr)
			return mshlzdHdr, nil
		}

		return nil, nil
	}

	store.GetAllCalled = func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
		mapToRet := make(map[string][]byte, 0)
		mb := block.MiniBlock{ReceiverShardID: 1, SenderShardID: 0}
		mshlzdMb, _ := json.Marshal(mb)
		mapToRet["mb1"] = mshlzdMb
		return mapToRet, nil
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
		resFinder,
		shardCoordinator,
		account,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
		&mock.BoostrapStorerMock{},
		&mock.StorageBootstrapperMock{},
		&mock.RequestedItemsHandlerStub{},
	)

	bs.RequestMiniBlocksFromHeaderWithNonceIfMissing(hdr)
	assert.True(t, requestDataWasCalled)
}
