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
	"github.com/stretchr/testify/assert"
)

func createMockMetaPools() *mock.MetaPoolsHolderStub {
	pools := &mock.MetaPoolsHolderStub{}
	pools.MetaChainBlocksCalled = func() storage.Cacher {
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
	pools.MetaBlockNoncesCalled = func() dataRetriever.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{
			GetCalled: func(u uint64) (bytes []byte, b bool) {
				return nil, false
			},
			RegisterHandlerCalled: func(handler func(nonce uint64)) {},
			RemoveCalled: func(u uint64) {
				return
			},
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
			blk.SetCurrentBlockHeader(hdr.(*block.MetaBlock))
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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilPoolsHolder, err)
}

func TestNewMetaBootstrap_PoolsHolderRetNilOnHeadersShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	pools.MetaChainBlocksCalled = func() storage.Cacher {
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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilMetaBlockPool, err)
}

func TestNewMetaBootstrap_PoolsHolderRetNilOnHeadersNoncesShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	pools.MetaBlockNoncesCalled = func() dataRetriever.Uint64Cacher {
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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewMetaBootstrap_NilHeaderResolverShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()

	errExpected := errors.New("expected error")

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
	)

	assert.Nil(t, bs)
	assert.Equal(t, errExpected, err)
}

func TestNewMetaBootstrap_NilTxBlockBodyResolverShouldErr(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()

	errExpected := errors.New("expected error")

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
	)

	assert.Nil(t, bs)
	assert.Equal(t, errExpected, err)
}

func TestNewMetaBootstrap_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := 0

	pools := &mock.MetaPoolsHolderStub{}
	pools.MetaChainBlocksCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
			assert.Fail(t, "should have not reached this point")
			return false, false
		}

		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	pools.MetaBlockNoncesCalled = func() dataRetriever.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
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
	)

	assert.NotNil(t, bs)
	assert.Nil(t, err)
	assert.Equal(t, 1, wasCalled)
}

//------- processing

func TestMetaBootstrap_SyncBlockShouldCallForkChoice(t *testing.T) {
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
	)

	r := bs.SyncBlock()

	assert.Equal(t, &sync.ErrSignedBlock{CurrentNonce: hdr.Nonce}, r)
}

func TestMetaBootstrap_ShouldReturnMissingHeader(t *testing.T) {
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
	)

	r := bs.SyncBlock()

	assert.Equal(t, process.ErrMissingHeader, r)
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
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return false, math.MaxUint64
	}
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return uint64(hdr.Nonce)
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 0
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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

	pools := &mock.MetaPoolsHolderStub{}
	pools.MetaChainBlocksCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			mutDataAvailable.RLock()
			defer mutDataAvailable.RUnlock()

			if bytes.Equal([]byte("aaa"), key) && dataAvailable {
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
	pools.MetaBlockNoncesCalled = func() dataRetriever.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}

		hnc.GetCalled = func(u uint64) (bytes []byte, b bool) {
			mutDataAvailable.RLock()
			defer mutDataAvailable.RUnlock()

			if u == 2 && dataAvailable {
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
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return uint64(hdr.Nonce)
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 0
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	account.RootHashCalled = func() []byte {
		return nil
	}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(200*time.Millisecond), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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

	pools := &mock.MetaPoolsHolderStub{}
	pools.MetaChainBlocksCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("aaa"), key) {
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
	pools.MetaBlockNoncesCalled = func() dataRetriever.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}

		hnc.GetCalled = func(u uint64) (bytes []byte, b bool) {
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

	pools := &mock.MetaPoolsHolderStub{}
	pools.MetaChainBlocksCalled = func() storage.Cacher {
		sds := &mock.CacherStub{}

		sds.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("aaa"), key) {
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
	pools.MetaBlockNoncesCalled = func() dataRetriever.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}
		hnc.RemoveCalled = func(u uint64) {
		}
		hnc.GetCalled = func(u uint64) (bytes []byte, b bool) {
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
	)

	err := bs.SyncBlock()

	assert.Equal(t, &sync.ErrSignedBlock{CurrentNonce: hdr.Nonce}, err)
}

func TestMetaBootstrap_ShouldSyncShouldReturnFalseWhenCurrentBlockIsNilAndRoundIndexIsZero(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{CheckForkCalled: func() (bool, uint64) {
		return false, math.MaxUint64
	}}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 0
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return false, math.MaxUint64
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 1
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return false, math.MaxUint64
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 0
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return false, math.MaxUint64
	}
	forkDetector.ProbableHighestNonceCalled = func() uint64 {
		return 1
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	)

	assert.True(t, bs.ShouldSync())
}

func TestMetaBootstrap_GetHeaderFromPoolShouldReturnNil(t *testing.T) {
	t.Parallel()

	pools := createMockMetaPools()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() (bool, uint64) {
		return false, math.MaxUint64
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	)

	assert.Nil(t, bs.GetHeaderFromPoolWithNonce(0))
}

func TestMetaBootstrap_GetHeaderFromPoolShouldReturnHeader(t *testing.T) {
	t.Parallel()

	hdr := &block.MetaBlock{Nonce: 0}

	pools := createMockMetaPools()
	pools.MetaChainBlocksCalled = func() storage.Cacher {
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
	pools.MetaBlockNoncesCalled = func() dataRetriever.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}
		hnc.GetCalled = func(u uint64) (i []byte, b bool) {
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
	)

	assert.True(t, hdr == bs.GetHeaderFromPoolWithNonce(0))
}

//------- testing received headers

func TestMetaBootstrap_ReceivedHeadersFoundInPoolShouldAddToForkDetector(t *testing.T) {
	t.Parallel()

	addedHash := []byte("hash")
	addedHdr := &block.MetaBlock{}

	pools := createMockMetaPools()
	pools.MetaChainBlocksCalled = func() storage.Cacher {
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
	)

	bs.ReceivedHeaders(addedHash)

	assert.True(t, wasAdded)
}

func TestMetaBootstrap_ReceivedHeadersNotFoundInPoolButFoundInStorageShouldAddToForkDetector(t *testing.T) {
	t.Parallel()

	addedHash := []byte("hash")
	addedHdr := &block.MetaBlock{Nonce: 1}

	pools := createMockMetaPools()

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

	store := createMetaStore()
	store.AddStorer(dataRetriever.MetaBlockUnit, headerStorage)

	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
	)

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

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
	)

	bs.ReceivedHeaders(addedHash)

	assert.True(t, wasAdded)
}

//------- ForkChoice

func TestMetaBootstrap_ForkChoiceNilBlockchainHeaderShouldErr(t *testing.T) {
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
	)

	err := bs.ForkChoice()
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestMetaBootstrap_ForkChoiceNilParamHeaderShouldErr(t *testing.T) {
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
	)

	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return nil
	}

	err := bs.ForkChoice()
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestMetaBootstrap_ForkChoiceIsNotEmptyShouldErr(t *testing.T) {
	t.Parallel()

	newHdrHash := []byte("new hdr hash")
	newHdrNonce := uint64(6)

	remFlags := &removedFlags{}

	pools := createMockMetaPools()
	pools.MetaChainBlocksCalled = func() storage.Cacher {
		return createHeadersDataPool(newHdrHash, remFlags)
	}
	pools.MetaBlockNoncesCalled = func() dataRetriever.Uint64Cacher {
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
	)

	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.MetaBlock{
			PubKeysBitmap: []byte("X"),
			Nonce:         newHdrNonce,
		}
	}

	err := bs.ForkChoice()
	assert.Equal(t, reflect.TypeOf(&sync.ErrSignedBlock{}), reflect.TypeOf(err))
}

func TestMetaBootstrap_ForkChoiceIsEmptyCallRollBackOkValsShouldWork(t *testing.T) {
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

	//data pool headers
	pools.MetaChainBlocksCalled = func() storage.Cacher {
		return createHeadersDataPool(currentHdrHash, remFlags)
	}
	//data pool headers-nonces
	pools.MetaBlockNoncesCalled = func() dataRetriever.Uint64Cacher {
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

func TestMetaBootstrap_ForkChoiceIsEmptyCallRollBackToGenesisShouldWork(t *testing.T) {
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

	//data pool headers
	pools.MetaChainBlocksCalled = func() storage.Cacher {
		return createHeadersDataPool(currentHdrHash, remFlags)
	}
	//data pool headers-nonces
	pools.MetaBlockNoncesCalled = func() dataRetriever.Uint64Cacher {
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

	err := bs.ForkChoice()
	assert.Nil(t, err)
	assert.True(t, remFlags.flagHdrRemovedFromNonces)
	assert.True(t, remFlags.flagHdrRemovedFromHeaders)
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
