package sync_test

import (
	"bytes"
	"encoding/json"
	"errors"
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
	"github.com/stretchr/testify/assert"
)

func createMockMetaPools() *mock.MetaPoolsHolderStub {
	pools := &mock.MetaPoolsHolderStub{}
	pools.MetaBlocksCalled = func() storage.Cacher {
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
			RemoveCalled:          func(nonce uint64, shardId uint32) {},
			MergeCalled:           func(u uint64, src dataRetriever.ShardIdHashMap) {},
		}
		return hnc
	}

	return pools
}

func createMockResolversFinderMeta() *mock.ResolversFinderStub {
	return &mock.ResolversFinderStub{
		MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
			if strings.Contains(baseTopic, factory.MetachainBlocksTopic) {
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

func createMetaBlockProcessor() *mock.BlockProcessorMock {
	blockProcessorMock := &mock.BlockProcessorMock{
		ProcessBlockCalled: func(blk data.ChainHandler, hdr data.HeaderHandler, bdy data.BodyHandler, haveTime func() time.Duration) error {
			_ = blk.SetCurrentBlockHeader(hdr.(*block.MetaBlock))
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

func createMetaStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit, generateTestUnit())
	return store
}

//------- NewMetaBootstrap

func TestNewMetaBootstrap_NilPoolsHolderShouldErr(t *testing.T) {
	t.Parallel()

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewMetaBootstrap(
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilPoolsHolder, err)
}

func TestNewMetaBootstrap_PoolsHolderRetNilOnHeadersShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	pools.MetaBlocksCalled = func() storage.Cacher {
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

	bs, err := sync.NewMetaBootstrap(
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilMetaBlocksPool, err)
}

func TestNewMetaBootstrap_NilStoreShouldErr(t *testing.T) {
	t.Parallel()
	blkc := initBlockchain()
	pools := createMockMetaPools()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewMetaBootstrap(
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilStore, err)
}

func TestNewMetaBootstrap_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewMetaBootstrap(
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestNewMetaBootstrap_NilRounderShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	blkc := initBlockchain()
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewMetaBootstrap(
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilRounder, err)
}

func TestNewMetaBootstrap_NilBlockProcessorShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewMetaBootstrap(
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilBlockExecutor, err)
}

func TestNewMetaBootstrap_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewMetaBootstrap(
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewMetaBootstrap_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewMetaBootstrap(
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewMetaBootstrap_NilForkDetectorShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewMetaBootstrap(
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilForkDetector, err)
}

func TestNewMetaBootstrap_NilResolversContainerShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewMetaBootstrap(
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilResolverContainer, err)
}

func TestNewMetaBootstrap_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	account := &mock.AccountsStub{}

	bs, err := sync.NewMetaBootstrap(
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewMetaBootstrap_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()

	bs, err := sync.NewMetaBootstrap(
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewMetaBootstrap_NilBlackListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewMetaBootstrap(
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
		nil,
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilBlackListHandler, err)
}

func TestNewMetaBootstrap_NilHeaderResolverShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	pools := createMockMetaPools()
	resFinder := &mock.ResolversFinderStub{
		MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
			if strings.Contains(baseTopic, factory.MetachainBlocksTopic) {
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

	bs, err := sync.NewMetaBootstrap(
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, errExpected, err)
}

func TestNewMetaBootstrap_NilTxBlockBodyResolverShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	pools := createMockMetaPools()
	resFinder := &mock.ResolversFinderStub{
		MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
			if strings.Contains(baseTopic, factory.MetachainBlocksTopic) {
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

	bs, err := sync.NewMetaBootstrap(
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
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.Nil(t, bs)
	assert.Equal(t, errExpected, err)
}

func TestNewMetaBootstrap_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := 0

	pools := &mock.MetaPoolsHolderStub{}
	pools.MetaBlocksCalled = func() storage.Cacher {
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

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewMetaBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	assert.NotNil(t, bs)
	assert.Nil(t, err)
	assert.Equal(t, 1, wasCalled)
	assert.False(t, bs.IsInterfaceNil())
}

//------- processing

func TestMetaBootstrap_SyncBlockShouldCallRollBack(t *testing.T) {
	t.Parallel()

	hdr := block.MetaBlock{Nonce: 1, PubKeysBitmap: []byte("X")}

	store := createMetaStore()

	blkc, _ := blockchain.NewMetaChain(
		&mock.CacherStub{},
	)

	blkc.CurrentBlock = &hdr

	pools := createMockMetaPools()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return &process.ForkInfo{true, 90, 90, []byte("hash")}
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

	blockProcessorMock := createMetaBlockProcessor()

	bs, _ := sync.NewMetaBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blockProcessorMock,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
	)

	r := bs.SyncBlock()

	assert.Equal(t, process.ErrNilHeadersNonceHashStorage, r)
}

func TestMetaBootstrap_ShouldReturnTimeIsOutWhenMissingHeader(t *testing.T) {
	t.Parallel()

	hdr := block.MetaBlock{Nonce: 1}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	pools := createMockMetaPools()

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
		&mock.SyncTimerMock{})

	blockProcessorMock := createMetaBlockProcessor()

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		&blkc,
		rnd,
		blockProcessorMock,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
	)

	r := bs.SyncBlock()

	assert.Equal(t, process.ErrTimeIsOut, r)
}

func TestMetaBootstrap_ShouldNotNeedToSync(t *testing.T) {
	t.Parallel()

	ebm := createMetaBlockProcessor()

	hdr := block.MetaBlock{Nonce: 1, Round: 0}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	pools := createMockMetaPools()

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
		return 0
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), 100*time.Millisecond, &mock.SyncTimerMock{})

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		&blkc,
		rnd,
		ebm,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
	)

	bs.StartSync()
	time.Sleep(200 * time.Millisecond)
	bs.StopSync()
}

func TestMetaBootstrap_SyncShouldSyncOneBlock(t *testing.T) {
	t.Parallel()

	ebm := createMetaBlockProcessor()

	hdr := block.MetaBlock{Nonce: 1, Round: 0}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	mutDataAvailable := goSync.RWMutex{}
	dataAvailable := false
	hash := []byte("aaa")

	pools := &mock.MetaPoolsHolderStub{}
	pools.MetaBlocksCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			mutDataAvailable.RLock()
			defer mutDataAvailable.RUnlock()

			if bytes.Equal(hash, key) && dataAvailable {
				return &block.MetaBlock{
					Nonce:    2,
					Round:    1,
					RootHash: []byte("bbb")}, true
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
				syncMap.Store(sharding.MetachainShardId, hash)

				return syncMap, true
			}

			return nil, false
		}
		return hnc
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
		return 0
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	account.RootHashCalled = func() ([]byte, error) {
		return nil, nil
	}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(200*time.Millisecond), 100*time.Millisecond, &mock.SyncTimerMock{})

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		&blkc,
		rnd,
		ebm,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
	)

	bs.StartSync()

	time.Sleep(200 * time.Millisecond)

	mutDataAvailable.Lock()
	dataAvailable = true
	mutDataAvailable.Unlock()

	time.Sleep(200 * time.Millisecond)

	bs.StopSync()
}

func TestMetaBootstrap_ShouldReturnNilErr(t *testing.T) {
	t.Parallel()

	ebm := createMetaBlockProcessor()

	hdr := block.MetaBlock{Nonce: 1}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	hash := []byte("aaa")
	pools := &mock.MetaPoolsHolderStub{}
	pools.MetaBlocksCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(hash, key) {
				return &block.MetaBlock{
					Nonce:    2,
					Round:    1,
					RootHash: []byte("bbb")}, true
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
				syncMap.Store(sharding.MetachainShardId, hash)

				return syncMap, true
			}

			return nil, false
		}
		return hnc
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
		&mock.SyncTimerMock{})

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		&blkc,
		rnd,
		ebm,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
	)

	r := bs.SyncBlock()

	assert.Nil(t, r)
}

func TestMetaBootstrap_SyncBlockShouldReturnErrorWhenProcessBlockFailed(t *testing.T) {
	t.Parallel()

	ebm := createMetaBlockProcessor()

	hdr := block.MetaBlock{Nonce: 1, PubKeysBitmap: []byte("X")}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	hash := []byte("aaa")
	pools := &mock.MetaPoolsHolderStub{}
	pools.MetaBlocksCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(hash, key) {
				return &block.MetaBlock{
					Nonce:    2,
					Round:    1,
					RootHash: []byte("bbb")}, true
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
				syncMap.Store(sharding.MetachainShardId, hash)

				return syncMap, true
			}

			return nil, false
		}
		hnc.RemoveCalled = func(nonce uint64, shardId uint32) {}
		return hnc
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

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		&blkc,
		rnd,
		ebm,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
	)

	err := bs.SyncBlock()

	assert.Equal(t, process.ErrBlockHashDoesNotMatch, err)
}

func TestMetaBootstrap_ShouldSyncShouldReturnFalseWhenCurrentBlockIsNilAndRoundIndexIsZero(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{CheckForkCalled: func() *process.ForkInfo {
		return process.NewForkInfo()
	}}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 0
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), 100*time.Millisecond, &mock.SyncTimerMock{})

	bs, err := sync.NewMetaBootstrap(
		pools,
		createStore(),
		initBlockchain(),
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
	)

	assert.Nil(t, err)
	assert.False(t, bs.ShouldSync())
}

func TestMetaBootstrap_ShouldReturnTrueWhenCurrentBlockIsNilAndRoundIndexIsGreaterThanZero(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()

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

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		initBlockchain(),
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
	)

	assert.True(t, bs.ShouldSync())
}

func TestMetaBootstrap_ShouldReturnFalseWhenNodeIsSynced(t *testing.T) {
	t.Parallel()

	hdr := block.MetaBlock{Nonce: 0}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	pools := createMockMetaPools()
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

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		&blkc,
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
	)

	assert.False(t, bs.ShouldSync())
}

func TestMetaBootstrap_ShouldReturnTrueWhenNodeIsNotSynced(t *testing.T) {
	t.Parallel()

	hdr := block.MetaBlock{Nonce: 0}
	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr
	}

	pools := createMockMetaPools()
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

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		&blkc,
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
	)

	assert.True(t, bs.ShouldSync())
}

func TestMetaBootstrap_ShouldSyncShouldReturnFalseWhenForkIsDetectedAndItReceivesTheSameWrongHeader(t *testing.T) {
	t.Parallel()

	hdr1 := block.MetaBlock{Nonce: 1, Round: 2, PubKeysBitmap: []byte("A")}
	hash1 := []byte("hash1")

	hdr2 := block.MetaBlock{Nonce: 1, Round: 1, PubKeysBitmap: []byte("B")}
	hash2 := []byte("hash2")

	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr1
	}

	pools := createMockMetaPools()
	pools.MetaBlocksCalled = func() storage.Cacher {
		return sync.GetCacherWithHeaders(&hdr1, &hdr2, hash1, hash2)
	}

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	rounder := &mock.RounderMock{}
	rounder.RoundIndex = 2
	forkDetector, _ := sync.NewMetaForkDetector(rounder, &mock.BlackListHandlerStub{
		AddCalled: func(key string) error {
			return nil
		},
		HasCalled: func(key string) bool {
			return false
		},
	})
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		&blkc,
		rounder,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
	)

	_ = forkDetector.AddHeader(&hdr1, hash1, process.BHProcessed, nil, nil, false)
	_ = forkDetector.AddHeader(&hdr2, hash2, process.BHReceived, nil, nil, false)

	shouldSync := bs.ShouldSync()
	assert.True(t, shouldSync)
	assert.True(t, bs.IsForkDetected())

	if shouldSync && bs.IsForkDetected() {
		forkDetector.RemoveHeaders(hdr1.GetNonce(), hash1)
		bs.ReceivedHeaders(hash1)
		_ = forkDetector.AddHeader(&hdr1, hash1, process.BHProcessed, nil, nil, false)
	}

	shouldSync = bs.ShouldSync()
	assert.False(t, shouldSync)
	assert.False(t, bs.IsForkDetected())
}

func TestMetaBootstrap_ShouldSyncShouldReturnFalseWhenForkIsDetectedAndItReceivesTheGoodHeader(t *testing.T) {
	t.Parallel()

	hdr1 := block.MetaBlock{Nonce: 1, Round: 2, PubKeysBitmap: []byte("A")}
	hash1 := []byte("hash1")

	hdr2 := block.MetaBlock{Nonce: 1, Round: 1, PubKeysBitmap: []byte("B")}
	hash2 := []byte("hash2")

	blkc := mock.BlockChainMock{}
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &hdr2
	}

	pools := createMockMetaPools()
	pools.MetaBlocksCalled = func() storage.Cacher {
		return sync.GetCacherWithHeaders(&hdr1, &hdr2, hash1, hash2)
	}

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	rounder := &mock.RounderMock{}
	rounder.RoundIndex = 2
	forkDetector, _ := sync.NewMetaForkDetector(rounder, &mock.BlackListHandlerStub{
		AddCalled: func(key string) error {
			return nil
		},
		HasCalled: func(key string) bool {
			return false
		},
	})
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		&blkc,
		rounder,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
	)

	_ = forkDetector.AddHeader(&hdr1, hash1, process.BHProcessed, nil, nil, false)
	_ = forkDetector.AddHeader(&hdr2, hash2, process.BHReceived, nil, nil, false)

	shouldSync := bs.ShouldSync()
	assert.True(t, shouldSync)
	assert.True(t, bs.IsForkDetected())

	if shouldSync && bs.IsForkDetected() {
		forkDetector.RemoveHeaders(hdr1.GetNonce(), hash1)
		bs.ReceivedHeaders(hash2)
		_ = forkDetector.AddHeader(&hdr2, hash2, process.BHProcessed, nil, nil, false)
	}

	shouldSync = bs.ShouldSync()
	assert.False(t, shouldSync)
	assert.False(t, bs.IsForkDetected())
}

func TestMetaBootstrap_GetHeaderFromPoolShouldReturnNil(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() *process.ForkInfo {
		return process.NewForkInfo()
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), 100*time.Millisecond, &mock.SyncTimerMock{})

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		initBlockchain(),
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	hdr, _, _ := process.GetMetaHeaderFromPoolWithNonce(0, pools.MetaBlocks(), pools.HeadersNonces())
	assert.NotNil(t, bs)
	assert.Nil(t, hdr)
}

func TestMetaBootstrap_GetHeaderFromPoolShouldReturnHeader(t *testing.T) {
	t.Parallel()

	hdr := &block.MetaBlock{Nonce: 0}

	hash := []byte("aaa")
	pools := createMockMetaPools()
	pools.MetaBlocksCalled = func() storage.Cacher {
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
				syncMap.Store(sharding.MetachainShardId, hash)

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

	rnd, _ := round.NewRound(time.Now(), time.Now(), 100*time.Millisecond, &mock.SyncTimerMock{})

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		initBlockchain(),
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	hdr2, _, _ := process.GetMetaHeaderFromPoolWithNonce(0, pools.MetaBlocks(), pools.HeadersNonces())
	assert.NotNil(t, bs)
	assert.True(t, hdr == hdr2)
}

//------- testing received headers

func TestMetaBootstrap_ReceivedHeadersFoundInPoolShouldAddToForkDetector(t *testing.T) {
	t.Parallel()

	addedHash := []byte("hash")
	addedHdr := &block.MetaBlock{}

	pools := createMockMetaPools()
	pools.MetaBlocksCalled = func() storage.Cacher {
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

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), 100*time.Millisecond, &mock.SyncTimerMock{})

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		initBlockchain(),
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	bs.ReceivedHeaders(addedHash)

	assert.True(t, wasAdded)
}

func TestMetaBootstrap_ReceivedHeadersNotFoundInPoolShouldNotAddToForkDetector(t *testing.T) {
	t.Parallel()

	addedHash := []byte("hash")
	addedHdr := &block.MetaBlock{Nonce: 1}

	pools := createMockMetaPools()

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

	store := createMetaStore()
	store.AddStorer(dataRetriever.MetaBlockUnit, headerStorage)

	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
	)

	rnd, _ := round.NewRound(time.Now(), time.Now(), 100*time.Millisecond, &mock.SyncTimerMock{})

	bs, _ := sync.NewMetaBootstrap(
		pools,
		store,
		blkc,
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	bs.ReceivedHeaders(addedHash)

	assert.False(t, wasAdded)
}

//------- RollBack

func TestMetaBootstrap_RollBackNilBlockchainHeaderShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	err := bs.RollBack(false)
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestMetaBootstrap_RollBackNilParamHeaderShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return nil
	}

	err := bs.RollBack(false)
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestMetaBootstrap_RollBackIsNotEmptyShouldErr(t *testing.T) {
	t.Parallel()

	newHdrHash := []byte("new hdr hash")
	newHdrNonce := uint64(6)

	remFlags := &removedFlags{}
	shardId := sharding.MetachainShardId

	pools := createMockMetaPools()
	pools.MetaBlocksCalled = func() storage.Cacher {
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

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
	)

	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.MetaBlock{
			PubKeysBitmap: []byte("X"),
			Nonce:         newHdrNonce,
		}
	}

	err := bs.RollBack(false)
	assert.Equal(t, sync.ErrRollBackBehindFinalHeader, err)
}

func TestMetaBootstrap_RollBackIsEmptyCallRollBackOneBlockOkValsShouldWork(t *testing.T) {
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
	prevHdr := &block.MetaBlock{
		Signature: []byte("sig of the prev header as to be unique in this context"),
		RootHash:  prevHdrRootHash,
	}

	pools := createMockMetaPools()
	shardId := sharding.MetachainShardId

	//data pool headers
	pools.MetaBlocksCalled = func() storage.Cacher {
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
				obj.(*block.MetaBlock).Signature = prevHdr.Signature
				obj.(*block.MetaBlock).RootHash = prevHdrRootHash
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

	bs, _ := sync.NewMetaBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{
			AddCalled: func(key string) error {
				return nil
			},
		},
		&mock.NetworkConnectionWatcherStub{},
	)

	bs.SetForkNonce(currentHdrNonce)

	hdr := &block.MetaBlock{
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
	assert.True(t, remFlags.flagHdrRemovedFromNonces)
	assert.False(t, remFlags.flagHdrRemovedFromHeaders)
	assert.True(t, remFlags.flagHdrRemovedFromStorage)
	assert.True(t, remFlags.flagHdrRemovedFromForkDetector)
	assert.Equal(t, blkc.GetCurrentBlockHeader(), prevHdr)
	assert.Equal(t, blkc.GetCurrentBlockBody(), prevTxBlockBody)
	assert.Equal(t, blkc.GetCurrentBlockHeaderHash(), prevHdrHash)
}

func TestMetaBootstrap_RollBackIsEmptyCallRollBackOneBlockToGenesisShouldWork(t *testing.T) {
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
	prevHdr := &block.MetaBlock{
		Signature: []byte("sig of the prev header as to be unique in this context"),
		RootHash:  prevHdrRootHash,
	}

	pools := createMockMetaPools()
	shardId := sharding.MetachainShardId

	//data pool headers
	pools.MetaBlocksCalled = func() storage.Cacher {
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
				obj.(*block.MetaBlock).Signature = prevHdr.Signature
				obj.(*block.MetaBlock).RootHash = prevHdrRootHash
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

	bs, _ := sync.NewMetaBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{
			AddCalled: func(key string) error {
				return nil
			},
		},
		&mock.NetworkConnectionWatcherStub{},
	)

	bs.SetForkNonce(currentHdrNonce)

	hdr := &block.MetaBlock{
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

	hdrHash := make([]byte, 0)
	blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
		return hdrHash
	}
	blkc.SetCurrentBlockHeaderHashCalled = func(i []byte) {
		hdrHash = nil
	}

	err := bs.RollBack(true)
	assert.Nil(t, err)
	assert.True(t, remFlags.flagHdrRemovedFromNonces)
	assert.False(t, remFlags.flagHdrRemovedFromHeaders)
	assert.True(t, remFlags.flagHdrRemovedFromStorage)
	assert.True(t, remFlags.flagHdrRemovedFromForkDetector)
	assert.Nil(t, blkc.GetCurrentBlockHeader())
	assert.Nil(t, blkc.GetCurrentBlockHeaderHash())
}

func TestMetaBootstrap_AddSyncStateListenerShouldAppendAnotherListener(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := createMetaBlockProcessor()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	f1 := func(bool) {}
	f2 := func(bool) {}
	f3 := func(bool) {}

	bs.AddSyncStateListener(f1)
	bs.AddSyncStateListener(f2)
	bs.AddSyncStateListener(f3)

	assert.Equal(t, 3, len(bs.SyncStateListeners()))
}

func TestMetaBootstrap_NotifySyncStateListenersShouldNotify(t *testing.T) {
	t.Parallel()

	mutex := goSync.RWMutex{}
	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := createMetaBlockProcessor()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
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

func TestMetaBootstrap_SyncFromStorerShouldErrWhenLoadBlocksFails(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
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

	bs, _ := sync.NewMetaBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	err := bs.SyncFromStorer(process.MetaBlockFinality,
		dataRetriever.MetaBlockUnit,
		dataRetriever.MetaHdrNonceHashDataUnit,
		process.ShardBlockFinality,
		dataRetriever.ShardHdrNonceHashDataUnit)

	assert.Equal(t, process.ErrNotEnoughValidBlocksInStorage, err)
}

func TestMetaBootstrap_SyncFromStorerShouldErrNilNotarized(t *testing.T) {
	t.Parallel()

	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	pools := createMockMetaPools()
	blkc := initBlockchain()
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.Header{}
	}

	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{
		AddLastNotarizedHdrCalled: func(shardId uint32, processedHdr data.HeaderHandler) {},
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte, isNotarizedShardStuck bool) error {
			return nil
		},
		GetNotarizedHeaderHashCalled: func(nonce uint64) []byte {
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
	store.HasCalled = func(unitType dataRetriever.UnitType, key []byte) error {
		if bytes.Equal(key, uint64Converter.ToByteSlice(1)) ||
			bytes.Equal(key, uint64Converter.ToByteSlice(2)) ||
			bytes.Equal(key, uint64Converter.ToByteSlice(3)) {
			return nil
		}

		return errors.New("key not found")
	}

	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		if unitType == dataRetriever.BlockHeaderUnit {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					buff, _ := marshalizer.Marshal(&block.Header{})
					return buff, nil
				},
			}
		}

		if unitType == dataRetriever.MetaBlockUnit {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					buff, _ := marshalizer.Marshal(&block.MetaBlock{})
					return buff, nil
				},
			}
		}

		return nil
	}

	bs, _ := sync.NewMetaBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	err := bs.SyncFromStorer(process.MetaBlockFinality,
		dataRetriever.MetaBlockUnit,
		dataRetriever.MetaHdrNonceHashDataUnit,
		process.ShardBlockFinality,
		dataRetriever.ShardHdrNonceHashDataUnit)

	assert.Equal(t, sync.ErrNilNotarizedHeader, err)
}

func TestMetaBootstrap_ApplyNotarizedBlockShouldErrWhenGetFinalNotarizedShardHeaderFromStorageFails(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
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
		if unitType == dataRetriever.BlockHeaderUnit {
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
		if unitType == dataRetriever.ShardHdrNonceHashDataUnit {
			if bytes.Equal(key, nonceToByteSlice) {
				return nil, errKeyNotFound
			}
			return []byte("hash"), nil
		}

		return nil, errKeyNotFound
	}

	bs, _ := sync.NewMetaBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	lastNotarized := bs.InitNotarizedMap()
	finalNotarized := bs.InitNotarizedMap()

	bs.SetNotarizedMap(lastNotarized, 0, 1, []byte("A"))
	bs.SetNotarizedMap(finalNotarized, 0, 1, []byte("A"))

	err := bs.ApplyNotarizedBlocks(finalNotarized, lastNotarized)

	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestMetaBootstrap_ApplyNotarizedBlockShouldErrWhenGetLastNotarizedShardHeaderFromStorageFails(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{
		AddLastNotarizedHdrCalled: func(shardId uint32, processedHdr data.HeaderHandler) {},
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
		if unitType == dataRetriever.BlockHeaderUnit {
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
		if unitType == dataRetriever.ShardHdrNonceHashDataUnit {
			if bytes.Equal(key, nonceToByteSlice) {
				return nil, errKeyNotFound
			}
			return []byte("hash"), nil
		}

		return nil, errKeyNotFound
	}

	bs, _ := sync.NewMetaBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	lastNotarized := bs.InitNotarizedMap()
	finalNotarized := bs.InitNotarizedMap()

	bs.SetNotarizedMap(lastNotarized, 0, 1, []byte("A"))
	bs.SetNotarizedMap(finalNotarized, 0, 2, []byte("B"))

	err := bs.ApplyNotarizedBlocks(finalNotarized, lastNotarized)

	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestMetaBootstrap_ApplyNotarizedBlockShouldWork(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{
		AddLastNotarizedHdrCalled: func(shardId uint32, processedHdr data.HeaderHandler) {},
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
		if unitType == dataRetriever.BlockHeaderUnit {
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
		if unitType == dataRetriever.ShardHdrNonceHashDataUnit {
			if bytes.Equal(key, nonceToByteSlice) {
				return nil, errKeyNotFound
			}
			return []byte("hash"), nil
		}

		return nil, errKeyNotFound
	}

	bs, _ := sync.NewMetaBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	lastNotarized := bs.InitNotarizedMap()
	finalNotarized := bs.InitNotarizedMap()

	bs.SetNotarizedMap(lastNotarized, 0, 1, hash)
	bs.SetNotarizedMap(finalNotarized, 0, 1, hash)

	err := bs.ApplyNotarizedBlocks(finalNotarized, lastNotarized)

	assert.Nil(t, err)
}

func TestMetaBootstrap_GetMaxNotarizedHeadersNoncesInMetaBlock(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")
	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{
		AddLastNotarizedHdrCalled: func(shardId uint32, processedHdr data.HeaderHandler) {},
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	nonceToByteSlice := uint64Converter.ToByteSlice(uint64(0))

	store := createStore()
	errKeyNotFound := errors.New("key not found")
	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		if unitType == dataRetriever.BlockHeaderUnit {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if !bytes.Equal(key, hash) {
						return nil, errKeyNotFound
					}

					hdr := block.Header{Nonce: 3}
					mshldzHdr, _ := json.Marshal(hdr)
					return mshldzHdr, nil
				},
			}
		}

		return nil
	}

	store.GetCalled = func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
		if unitType == dataRetriever.ShardHdrNonceHashDataUnit {
			if bytes.Equal(key, nonceToByteSlice) {
				return nil, errKeyNotFound
			}
			return hash, nil
		}

		return nil, errKeyNotFound
	}

	bs, _ := sync.NewMetaBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.NetworkConnectionWatcherStub{},
	)

	lastNotarized := make(map[uint32]uint64, 0)
	finalNotarized := make(map[uint32]uint64, 0)
	blockWithLastNotarized := make(map[uint32]uint64, 0)
	blockWithFinalNotarized := make(map[uint32]uint64, 0)

	lastNotarized[0] = 1
	finalNotarized[0] = 1
	blockWithLastNotarized[0] = 1
	blockWithFinalNotarized[0] = 1
	startNonce := uint64(0)

	shardData := block.ShardData{TxCount: 10, HeaderHash: hash}
	metaBlock := &block.MetaBlock{
		ShardInfo: []block.ShardData{shardData},
	}
	notarizedInfo := bs.GetNotarizedInfo(lastNotarized, finalNotarized, blockWithLastNotarized, blockWithFinalNotarized, startNonce)
	resMap, err := bs.GetMaxNotarizedHeadersNoncesInMetaBlock(metaBlock, notarizedInfo)

	assert.Nil(t, err)
	assert.NotNil(t, resMap)
}

func TestMetaBootstrap_AreNotarizedShardHeadersFoundShouldReturnFalse(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")
	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{
		AddLastNotarizedHdrCalled: func(shardId uint32, processedHdr data.HeaderHandler) {},
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return json.Unmarshal(buff, obj)
		},
	}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	nonceToByteSlice := uint64Converter.ToByteSlice(uint64(0))

	store := createStore()
	errKeyNotFound := errors.New("key not found")
	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		if unitType == dataRetriever.BlockHeaderUnit {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if !bytes.Equal(key, hash) {
						return nil, errKeyNotFound
					}

					hdr := block.Header{Nonce: 3}
					mshldzHdr, _ := json.Marshal(hdr)
					return mshldzHdr, nil
				},
			}
		}

		return nil
	}

	store.GetCalled = func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
		if unitType == dataRetriever.ShardHdrNonceHashDataUnit {
			if bytes.Equal(key, nonceToByteSlice) {
				return nil, errKeyNotFound
			}
			return hash, nil
		}

		return nil, errKeyNotFound
	}

	bs, _ := sync.NewMetaBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.NetworkConnectionWatcherStub{},
	)

	lastNotarized := make(map[uint32]uint64, 0)
	finalNotarized := make(map[uint32]uint64, 0)
	blockWithLastNotarized := make(map[uint32]uint64, 0)
	blockWithFinalNotarized := make(map[uint32]uint64, 0)

	lastNotarized[0] = 1
	lastNotarized[1] = 0

	finalNotarized[0] = 0
	finalNotarized[1] = 0
	blockWithLastNotarized[0] = 0
	blockWithFinalNotarized[0] = 0
	startNonce := uint64(0)

	notarizedInfo := bs.GetNotarizedInfo(lastNotarized, finalNotarized, blockWithLastNotarized, blockWithFinalNotarized, startNonce)

	notarizedNonces := make(map[uint32]uint64, 0)
	notarizedNonces[0] = 0
	res := bs.AreNotarizedShardHeadersFound(notarizedInfo, notarizedNonces, 0)

	assert.False(t, res)
}

func TestMetaBootstrap_AreNotarizedShardHeadersFoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")
	pools := createMockMetaPools()
	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{
		AddLastNotarizedHdrCalled: func(shardId uint32, processedHdr data.HeaderHandler) {},
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return json.Unmarshal(buff, obj)
		},
	}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	nonceToByteSlice := uint64Converter.ToByteSlice(uint64(0))

	store := createStore()
	errKeyNotFound := errors.New("key not found")
	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		if unitType == dataRetriever.BlockHeaderUnit {
			return &mock.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if !bytes.Equal(key, hash) {
						return nil, errKeyNotFound
					}

					hdr := block.Header{Nonce: 3}
					mshldzHdr, _ := json.Marshal(hdr)
					return mshldzHdr, nil
				},
			}
		}

		return nil
	}

	store.GetCalled = func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
		if unitType == dataRetriever.ShardHdrNonceHashDataUnit {
			if bytes.Equal(key, nonceToByteSlice) {
				return nil, errKeyNotFound
			}
			return hash, nil
		}

		return nil, errKeyNotFound
	}

	bs, _ := sync.NewMetaBootstrap(
		pools,
		store,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.NetworkConnectionWatcherStub{},
	)

	lastNotarized := make(map[uint32]uint64, 0)
	finalNotarized := make(map[uint32]uint64, 0)
	blockWithLastNotarized := make(map[uint32]uint64, 0)
	blockWithFinalNotarized := make(map[uint32]uint64, 0)

	lastNotarized[0] = 0
	finalNotarized[0] = 1

	blockWithLastNotarized[0] = 0
	blockWithFinalNotarized[0] = 0
	startNonce := uint64(0)

	notarizedInfo := bs.GetNotarizedInfo(lastNotarized, finalNotarized, blockWithLastNotarized, blockWithFinalNotarized, startNonce)

	notarizedNonces := make(map[uint32]uint64, 0)
	notarizedNonces[0] = 5

	res := bs.AreNotarizedShardHeadersFound(notarizedInfo, notarizedNonces, 0)

	assert.True(t, res)
}

func TestMetaBootstrap_SetStatusHandlerNilHandlerShouldErr(t *testing.T) {
	t.Parallel()

	pools := &mock.MetaPoolsHolderStub{}
	pools.MetaBlocksCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
			return false, false
		}

		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		hnc := &mock.Uint64SyncMapCacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64, shardId uint32, hash []byte)) {}

		return hnc
	}

	blkc := initBlockchain()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewMetaBootstrap(
		pools,
		createStore(),
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversFinderMeta(),
		shardCoordinator,
		account,
		math.MaxUint32,
		&mock.BlackListHandlerStub{},
		&mock.NetworkConnectionWatcherStub{},
	)

	err := bs.SetStatusHandler(nil)
	assert.Equal(t, process.ErrNilAppStatusHandler, err)
}
