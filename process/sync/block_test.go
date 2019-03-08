package sync_test

import (
	"bytes"
	"fmt"
	"reflect"
	goSync "sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/sync"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/pkg/errors"
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

func createMockResolversContainer() *mock.ResolversContainerStub {
	return &mock.ResolversContainerStub{
		GetCalled: func(key string) (resolver process.Resolver, e error) {
			if key == factory.HeadersTopic {
				return &mock.HeaderResolverMock{
					RequestDataFromNonceCalled: func(nonce uint64) error {
						return nil
					},
					RequestDataFromHashCalled: func(hash []byte) error {
						return nil
					},
				}, nil
			}

			if key == factory.MiniBlocksTopic {
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

func createMockResolversContainerNilMiniBlocks() *mock.ResolversContainerStub {
	return &mock.ResolversContainerStub{
		GetCalled: func(key string) (resolver process.Resolver, e error) {
			if key == factory.HeadersTopic {
				return &mock.HeaderResolverMock{
					RequestDataFromNonceCalled: func(nonce uint64) error {
						return nil
					},
					RequestDataFromHashCalled: func(hash []byte) error {
						return nil
					},
				}, nil
			}

			if key == factory.MiniBlocksTopic {
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

func createMockTransient() *mock.TransientDataPoolMock {
	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{
			AddDataCalled:         func(key []byte, data interface{}, destShardID uint32) {},
			RegisterHandlerCalled: func(func(key []byte)) {},
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}
		return sds
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{
			GetCalled: func(u uint64) (bytes []byte, b bool) {
				return nil, false
			},
			RegisterHandlerCalled: func(handler func(nonce uint64)) {},
		}
		return hnc
	}
	transient.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
			RegisterHandlerCalled: func(i func(key []byte)) {},
		}
		return cs
	}

	return transient
}

func createBlockProcessor() *mock.BlockProcessorMock {
	blockProcessorMock := &mock.BlockProcessorMock{
		ProcessAndCommitCalled: func(blk *blockchain.BlockChain, hdr *block.Header, bdy block.Body, haveTime func() time.Duration) error {
			blk.CurrentBlockHeader = hdr
			return nil
		},
		RevertAccountStateCalled: func() {
			return
		},
		CommitBlockCalled: func(blockChain *blockchain.BlockChain, header *block.Header, block block.Body) error {
			return nil
		},
	}

	return blockProcessorMock
}

func createHeadersDataPool(removedHashCompare []byte, remFlags *removedFlags) data.ShardedDataCacherNotifier {
	sds := &mock.ShardedDataStub{
		AddDataCalled:         func(key []byte, data interface{}, destShardID uint32) {},
		RegisterHandlerCalled: func(func(key []byte)) {},
		RemoveDataCalled: func(key []byte, destShardID uint32) {
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
	remFlags *removedFlags) data.Uint64Cacher {

	hnc := &mock.Uint64CacherStub{
		RegisterHandlerCalled: func(handler func(nonce uint64)) {},
		GetCalled: func(u uint64) (i []byte, b bool) {
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
		RemoveHeadersCalled: func(nonce uint64) {
			if nonce == removedNonce {
				remFlags.flagHdrRemovedFromForkDetector = true
			}
		},
	}
}

func createHeadersStorage(
	getHashCompare []byte,
	getRetBytes []byte,
	removedHash []byte,
	remFlags *removedFlags,
) storage.Storer {

	return &mock.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			if bytes.Equal(key, getHashCompare) {
				return getRetBytes, nil
			}

			return nil, errors.New("not found")
		},
		RemoveCalled: func(key []byte) error {
			if bytes.Equal(key, removedHash) {
				remFlags.flagHdrRemovedFromStorage = true
			}
			return nil
		},
	}
}

//------- NewBootstrap

func TestNewBootstrap_NilTransientDataHolderShouldErr(t *testing.T) {
	t.Parallel()

	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewBootstrap(
		nil,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversContainerStub{},
		shardCoordinator,
		account,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilTransientDataHolder, err)
}

func TestNewBootstrap_TransientDataHolderRetNilOnHeadersShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return nil
	}

	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversContainerStub{},
		shardCoordinator,
		account,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilHeadersDataPool, err)
}

func TestNewBootstrap_TransientDataHolderRetNilOnHeadersNoncesShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return nil
	}
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversContainerStub{},
		shardCoordinator,
		account,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
}

func TestNewBootstrap_TransientDataHolderRetNilOnTxBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	transient.MiniBlocksCalled = func() storage.Cacher {
		return nil
	}
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversContainerStub{},
		shardCoordinator,
		account,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilTxBlockBody, err)
}

func TestNewBootstrap_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewBootstrap(
		transient,
		nil,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversContainerStub{},
		shardCoordinator,
		account,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestNewBootstrap_NilRounderShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewBootstrap(
		transient,
		blkc,
		nil,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversContainerStub{},
		shardCoordinator,
		account,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilRounder, err)
}

func TestNewBootstrap_NilBlockProcessorShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		nil,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversContainerStub{},
		shardCoordinator,
		account,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilBlockExecutor, err)
}

func TestNewBootstrap_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		nil,
		marshalizer,
		forkDetector,
		&mock.ResolversContainerStub{},
		shardCoordinator,
		account,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewBootstrap_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		nil,
		forkDetector,
		&mock.ResolversContainerStub{},
		shardCoordinator,
		account,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewBootstrap_NilForkDetectorShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		nil,
		&mock.ResolversContainerStub{},
		shardCoordinator,
		account,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilForkDetector, err)
}

func TestNewBootstrap_NilResolversContainerShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewBootstrap(
		transient,
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

func TestNewBootstrap_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	account := &mock.AccountsStub{}

	bs, err := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversContainerStub{},
		nil,
		account,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewBootstrap_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()

	bs, err := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversContainerStub{},
		shardCoordinator,
		nil,
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewBootstrap_NilHeaderResolverShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()

	errExpected := errors.New("expected error")

	resContainer := &mock.ResolversContainerStub{
		GetCalled: func(key string) (resolver process.Resolver, e error) {
			if key == factory.HeadersTopic {
				return nil, errExpected
			}

			if key == factory.MiniBlocksTopic {
				return &mock.ResolverStub{}, nil
			}

			return nil, nil
		},
	}

	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		resContainer,
		shardCoordinator,
		account,
	)

	assert.Nil(t, bs)
	assert.Equal(t, errExpected, err)
}

func TestNewBootstrap_NilTxBlockBodyResolverShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()

	errExpected := errors.New("expected error")

	resContainer := &mock.ResolversContainerStub{
		GetCalled: func(key string) (resolver process.Resolver, e error) {
			if key == factory.HeadersTopic {
				return &mock.HeaderResolverMock{}, errExpected
			}

			if key == factory.MiniBlocksTopic {
				return nil, errExpected
			}

			return nil, nil
		},
	}

	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		resContainer,
		shardCoordinator,
		account,
	)

	assert.Nil(t, bs)
	assert.Equal(t, errExpected, err)
}

func TestNewBootstrap_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := 0

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}

		sds.AddDataCalled = func(key []byte, data interface{}, destShardID uint32) {
			assert.Fail(t, "should have not reached this point")
		}

		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
			wasCalled++
		}

		return hnc
	}
	transient.MiniBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
			wasCalled++
		}

		return cs
	}

	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, err := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	assert.NotNil(t, bs)
	assert.Nil(t, err)
	assert.Equal(t, 2, wasCalled)
}

//------- processing

func TestBootstrap_ShouldReturnMissingHeader(t *testing.T) {
	t.Parallel()

	hdr := block.Header{Nonce: 1}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	transient := createMockTransient()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() bool {
		return true
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

	blockProcessorMock := createBlockProcessor()

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		rnd,
		blockProcessorMock,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	bs.BroadcastBlock = func(body block.Body, header *block.Header) error {
		return nil
	}

	bs.SetHighestNonceReceived(100)

	r := bs.SyncBlock()

	assert.Equal(t, process.ErrMissingHeader, r)
}

func TestBootstrap_ShouldReturnMissingBody(t *testing.T) {
	t.Parallel()

	hdr := block.Header{Nonce: 1}
	blockBodyUnit := &mock.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return nil, nil
		},
	}
	blkc, _ := blockchain.NewBlockChain(&mock.CacherStub{}, &mock.StorerStub{}, blockBodyUnit,
		&mock.StorerStub{}, &mock.StorerStub{})
	blkc.CurrentBlockHeader = &hdr

	transient := createMockTransient()
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}

		sds.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("aaa"), key) {
				return &block.Header{Nonce: 2}, true
			}

			return nil, false
		}

		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
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
	forkDetector.CheckForkCalled = func() bool {
		return true
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainerNilMiniBlocks(),
		shardCoordinator,
		account,
	)

	bs.BroadcastBlock = func(body block.Body, header *block.Header) error {
		return nil
	}

	bs.RequestHeader(2)

	r := bs.SyncBlock()

	assert.Equal(t, process.ErrMissingBody, r)
}

func TestBootstrap_ShouldNotNeedToSync(t *testing.T) {
	t.Parallel()

	ebm := createBlockProcessor()

	hdr := block.Header{Nonce: 1, Round: 0}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	transient := createMockTransient()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	forkDetector.CheckForkCalled = func() bool {
		return false
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		rnd,
		ebm,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	bs.BroadcastBlock = func(body block.Body, header *block.Header) error {
		return nil
	}

	bs.StartSync()
	time.Sleep(200 * time.Millisecond)
	bs.StopSync()
}

func TestBootstrap_SyncShouldSyncOneBlock(t *testing.T) {
	t.Parallel()

	ebm := createBlockProcessor()

	hdr := block.Header{Nonce: 1, Round: 0}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	mutDataAvailable := goSync.RWMutex{}
	dataAvailable := false

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}

		sds.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
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
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
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
	transient.MiniBlocksCalled = func() storage.Cacher {
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

	forkDetector.CheckForkCalled = func() bool {
		return false
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	account.RootHashCalled = func() []byte {
		return nil
	}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(200*time.Millisecond), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		rnd,
		ebm,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	bs.BroadcastBlock = func(body block.Body, header *block.Header) error {
		return nil
	}

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
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}

		sds.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
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
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
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
	transient.MiniBlocksCalled = func() storage.Cacher {
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
	forkDetector.CheckForkCalled = func() bool {
		return false
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		rnd,
		ebm,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	r := bs.SyncBlock()

	assert.Nil(t, r)
}

func TestBootstrap_ShouldSyncShouldReturnFalseWhenCurrentBlockIsNilAndRoundIndexIsZero(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{CheckForkCalled: func() bool {
		return false
	}}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	assert.False(t, bs.ShouldSync())
}

func TestBootstrap_ShouldReturnTrueWhenCurrentBlockIsNilAndRoundIndexIsGreaterThanZero(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	forkDetector.CheckForkCalled = func() bool {
		return false
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	assert.True(t, bs.ShouldSync())
}

func TestBootstrap_ShouldReturnFalseWhenNodeIsSynced(t *testing.T) {
	t.Parallel()

	hdr := block.Header{Nonce: 0}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	transient := createMockTransient()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	forkDetector.CheckForkCalled = func() bool {
		return false
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	assert.False(t, bs.ShouldSync())
}

func TestBootstrap_ShouldReturnTrueWhenNodeIsNotSynced(t *testing.T) {
	t.Parallel()

	hdr := block.Header{Nonce: 0}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	transient := createMockTransient()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() bool {
		return false
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	assert.False(t, bs.ShouldSync())
}

func TestBootstrap_GetHeaderFromPoolShouldReturnNil(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	forkDetector.CheckForkCalled = func() bool {
		return false
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	assert.Nil(t, bs.GetHeaderFromPool(0))
}

func TestBootstrap_GetHeaderFromPoolShouldReturnHeader(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{Nonce: 0}

	transient := createMockTransient()
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}

		sds.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("aaa"), key) {
				return hdr, true
			}
			return nil, false
		}

		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}

		return sds
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
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

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	assert.True(t, hdr == bs.GetHeaderFromPool(0))
}

func TestGetBlockFromPoolShouldReturnBlock(t *testing.T) {
	blk := make(block.MiniBlockSlice, 0)

	transient := createMockTransient()

	transient.MiniBlocksCalled = func() storage.Cacher {
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

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
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

	transient := createMockTransient()
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}
		sds.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
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
	forkDetector.AddHeaderCalled = func(header *block.Header, hash []byte, isReceived bool) error {
		if !isReceived {
			return errors.New("not received")
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

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
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

	transient := createMockTransient()

	wasAdded := false
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.AddHeaderCalled = func(header *block.Header, hash []byte, isReceived bool) error {
		if !isReceived {
			return errors.New("not received")
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

	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		headerStorage)

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		&mock.BlockProcessorMock{},
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	bs.ReceivedHeaders(addedHash)

	assert.True(t, wasAdded)
}

//------- ForkChoice

func TestBootstrap_ForkChoiceNilBlockchainHeaderShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	err := bs.ForkChoice(&block.Header{})
	assert.Equal(t, sync.ErrNilCurrentHeader, err)
}

func TestBootstrap_ForkChoiceNilParamHeaderShouldErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	blkc.CurrentBlockHeader = &block.Header{}

	err := bs.ForkChoice(nil)
	assert.Equal(t, sync.ErrNilHeader, err)
}

func TestBootstrap_ForkChoiceIsNotEmptyShouldRemove(t *testing.T) {
	t.Parallel()

	newHdrHash := []byte("new hdr hash")
	newHdrNonce := uint64(6)

	remFlags := &removedFlags{}

	transient := createMockTransient()
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return createHeadersDataPool(newHdrHash, remFlags)
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return createHeadersNoncesDataPool(newHdrNonce, newHdrHash, newHdrNonce, remFlags)
	}
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := createForkDetector(newHdrNonce, remFlags)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	blkc.CurrentBlockHeader = &block.Header{
		PubKeysBitmap: []byte{1},
	}

	newHdr := &block.Header{Nonce: newHdrNonce}

	err := bs.ForkChoice(newHdr)
	assert.Equal(t, reflect.TypeOf(&sync.ErrNotEmptyHeader{}), reflect.TypeOf(err))
	fmt.Printf(err.Error())
	assert.True(t, remFlags.flagHdrRemovedFromNonces)
	assert.True(t, remFlags.flagHdrRemovedFromHeaders)
	assert.True(t, remFlags.flagHdrRemovedFromForkDetector)

}

func TestBootstrap_ForkChoiceIsEmptyCallRollBackOkValsShouldWork(t *testing.T) {
	t.Skip("unskip this test after the fix is applied on rollback, storer not erasing header")

	t.Parallel()

	//retain if the remove process from different storage locations has been called
	remFlags := &removedFlags{}

	currentHdrNonce := uint64(8)
	currentHdrHash := []byte("current header hash")

	//define prev tx block body "strings" as in this test there are a lot of stubs that
	//constantly need to check some defined symbols
	prevTxBlockBodyHash := []byte("prev block body hash")
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

	transient := createMockTransient()

	//data pool headers
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return createHeadersDataPool(currentHdrHash, remFlags)
	}
	//data pool headers-nonces
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return createHeadersNoncesDataPool(
			currentHdrNonce,
			currentHdrHash,
			currentHdrNonce,
			remFlags,
		)
	}

	hdrUnit := createHeadersStorage(prevHdrHash, prevHdrBytes, currentHdrHash, remFlags)
	txBlockUnit := createHeadersStorage(prevTxBlockBodyHash, prevTxBlockBodyBytes, nil, remFlags)

	//a mock blockchain with special header and tx block bodies stubs (defined above)
	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
		&mock.StorerStub{},
		txBlockUnit,
		&mock.StorerStub{},
		hdrUnit,
	)
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}

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
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	//this is the block we want to revert
	blkc.CurrentBlockHeader = &block.Header{
		Nonce: currentHdrNonce,
		//empty bitmap
		PrevHash: prevHdrHash,
	}

	err := bs.ForkChoice(&block.Header{})
	assert.Nil(t, err)
	assert.True(t, remFlags.flagHdrRemovedFromNonces)
	assert.True(t, remFlags.flagHdrRemovedFromHeaders)
	assert.True(t, remFlags.flagHdrRemovedFromStorage)
	assert.True(t, remFlags.flagHdrRemovedFromForkDetector)
	assert.Equal(t, blkc.CurrentBlockHeader, prevHdr)
	assert.Equal(t, blkc.CurrentTxBlockBody, prevTxBlockBody)
	assert.Equal(t, blkc.CurrentBlockHeaderHash, prevHdrHash)
}

func TestBootstrap_ForkChoiceIsEmptyCallRollBackToGenesisShouldWork(t *testing.T) {
	t.Skip("unskip this test after the fix is applied on rollback, storer not erasing header")

	t.Parallel()

	//retain if the remove process from different storage locations has been called
	remFlags := &removedFlags{}

	currentHdrNonce := uint64(1)
	currentHdrHash := []byte("current header hash")

	//define prev tx block body "strings" as in this test there are a lot of stubs that
	//constantly need to check some defined symbols
	prevTxBlockBodyHash := []byte("prev block body hash")
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

	transient := createMockTransient()

	//data pool headers
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return createHeadersDataPool(currentHdrHash, remFlags)
	}
	//data pool headers-nonces
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return createHeadersNoncesDataPool(
			currentHdrNonce,
			currentHdrHash,
			currentHdrNonce,
			remFlags,
		)
	}

	hdrUnit := createHeadersStorage(prevHdrHash, prevHdrBytes, currentHdrHash, remFlags)
	txBlockUnit := createHeadersStorage(prevTxBlockBodyHash, prevTxBlockBodyBytes, nil, remFlags)

	//a mock blockchain with special header and tx block bodies stubs (defined above)
	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
		&mock.StorerStub{},
		txBlockUnit,
		&mock.StorerStub{},
		hdrUnit,
	)
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}

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
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	//this is the block we want to revert
	blkc.CurrentBlockHeader = &block.Header{
		Nonce: currentHdrNonce,
		//empty bitmap
		PrevHash: prevHdrHash,
	}

	err := bs.ForkChoice(&block.Header{})
	assert.Nil(t, err)
	assert.True(t, remFlags.flagHdrRemovedFromNonces)
	assert.True(t, remFlags.flagHdrRemovedFromHeaders)
	assert.True(t, remFlags.flagHdrRemovedFromStorage)
	assert.True(t, remFlags.flagHdrRemovedFromForkDetector)
	assert.Nil(t, blkc.CurrentBlockHeader)
	assert.Nil(t, blkc.CurrentTxBlockBody)
	assert.Nil(t, blkc.CurrentBlockHeaderHash)
}

//------- GetTxBodyHavingHash

func TestBootstrap_GetTxBodyHavingHashReturnsFromCacherShouldWork(t *testing.T) {
	t.Parallel()

	mbh := []byte("requested hash")
	requestedHash := make([][]byte, 0)
	requestedHash = append(requestedHash, mbh)
	mb := &block.MiniBlock{}
	txBlock := make(block.MiniBlockSlice, 0)

	transient := createMockTransient()
	transient.MiniBlocksCalled = func() storage.Cacher {
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
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
	)
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
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

	transient := createMockTransient()

	txBlockUnit := &mock.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return nil, errors.New("not found")
		},
	}

	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
		&mock.StorerStub{},
		txBlockUnit,
		&mock.StorerStub{},
		&mock.StorerStub{},
	)
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainerNilMiniBlocks(),
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

	transient := createMockTransient()

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
		&mock.StorerStub{},
		txBlockUnit,
		&mock.StorerStub{},
		&mock.StorerStub{},
	)

	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)
	txBlockRecovered := bs.GetMiniBlocks(requestedHash)

	assert.Equal(t, txBlock, txBlockRecovered)
}

func TestBootstrap_CreateEmptyBlockShouldReturnNilWhenMarshalErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := createBlockProcessor()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	marshalizer.Fail = true

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	blk, hdr, _ := bs.CreateAndCommitEmptyBlock(0)

	assert.Nil(t, blk)
	assert.Nil(t, hdr)
}

func TestBootstrap_CreateEmptyBlockShouldReturnNilWhenCommitBlockErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := createBlockProcessor()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	blkc.CurrentBlockHeader = &block.Header{Nonce: 1}

	err := errors.New("error")
	blkExec.CommitBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, block block.Body) error {
		return err
	}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	blk, hdr, _ := bs.CreateAndCommitEmptyBlock(0)

	assert.Nil(t, blk)
	assert.Nil(t, hdr)
}

func TestBootstrap_CreateEmptyBlockShouldWork(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := createBlockProcessor()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	blk, hdr, _ := bs.CreateAndCommitEmptyBlock(0)

	assert.NotNil(t, blk)
	assert.NotNil(t, hdr)
}

func TestBootstrap_AddSyncStateListenerShouldAppendAnotherListener(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := createBlockProcessor()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
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

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := createBlockProcessor()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	calls := 0
	var wg goSync.WaitGroup

	f1 := func(bool) {
		calls++
		wg.Done()
	}

	f2 := func(bool) {
		calls++
		wg.Done()
	}

	f3 := func(bool) {
		calls++
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

func TestNewBootstrap_GetTimeStampForRoundShouldWork(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	rnd.RoundIndex = 0
	rnd.RoundTimeStamp = time.Unix(0, 0)
	rnd.RoundTimeDuration = 4000 * time.Millisecond

	timeStamp := bs.GetTimeStampForRound(2)
	assert.Equal(t, time.Unix(8, 0), timeStamp)

	rnd.RoundIndex = 5
	rnd.RoundTimeStamp = time.Unix(20, 0)
	rnd.RoundTimeDuration = 4000 * time.Millisecond

	timeStamp = bs.GetTimeStampForRound(1)
	assert.Equal(t, time.Unix(4, 0), timeStamp)
}

func TestNewBootstrap_ShouldCreateEmptyBlockShouldReturnFalseWhenForkIsDetected(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	bs.SetIsForkDetected(true)

	assert.False(t, bs.ShouldCreateEmptyBlock(0))
}

func TestNewBootstrap_ShouldCreateEmptyBlockShouldReturnFalseWhenNonceIsSmallerOrEqualThanMaxHeaderNonceReceived(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	bs.SetHighestNonceReceived(1)

	assert.False(t, bs.ShouldCreateEmptyBlock(0))
}

func TestNewBootstrap_ShouldCreateEmptyBlockShouldReturnTrue(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	assert.True(t, bs.ShouldCreateEmptyBlock(1))
}

func TestNewBootstrap_CreateAndBroadcastEmptyBlockShouldReturnErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := createBlockProcessor()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	err := errors.New("error")
	blkExec.CommitBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, block block.Body) error {
		return err
	}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	err2 := bs.CreateAndBroadcastEmptyBlock()

	assert.Equal(t, err2, err)
}

func TestNewBootstrap_CreateAndBroadcastEmptyBlockShouldReturnNil(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := createBlockProcessor()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	bs.BroadcastBlock = func(body block.Body, header *block.Header) error {
		return nil
	}

	err := bs.CreateAndBroadcastEmptyBlock()

	assert.Nil(t, err)
}

func TestNewBootstrap_BroadcastEmptyBlockShouldErrWhenBroadcastBlockErr(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	err := errors.New("error")
	bs.BroadcastBlock = func(body block.Body, header *block.Header) error {
		return err
	}

	err2 := bs.BroadcastEmptyBlock(nil, nil)

	assert.Equal(t, err, err2)
}

func TestNewBootstrap_BroadcastEmptyBlockShouldReturnNil(t *testing.T) {
	t.Parallel()

	transient := createMockTransient()
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	account := &mock.AccountsStub{}

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
		shardCoordinator,
		account,
	)

	bs.BroadcastBlock = func(body block.Body, header *block.Header) error {
		return nil
	}

	err := bs.BroadcastEmptyBlock(nil, nil)

	assert.Nil(t, err)
}
