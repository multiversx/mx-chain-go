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
			if key == string(factory.HeadersTopic) {
				return &mock.HeaderResolverMock{
					RequestDataFromNonceCalled: func(nonce uint64) error {
						return nil
					},
					RequestDataFromHashCalled: func(hash []byte) error {
						return nil
					},
				}, nil
			}

			if key == string(factory.TxBlockBodyTopic) {
				return &mock.ResolverStub{
					RequestDataFromHashCalled: func(hash []byte) error {
						return nil
					},
				}, nil
			}

			return nil, nil
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

	bs, err := sync.NewBootstrap(
		nil,
		blkc,
		rnd,
		blkExec,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		&mock.ResolversContainerStub{})

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilTransientDataHolder, err)
}

func TestNewBootstrap_TransientDataHolderRetNilOnHeadersShouldErr(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return nil
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilHeadersDataPool, err)
}

func TestNewBootstrap_TransientDataHolderRetNilOnHeadersNoncesShouldErr(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return nil
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
}

func TestNewBootstrap_TransientDataHolderRetNilOnTxBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return nil
	}
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilTxBlockBody, err)
}

func TestNewBootstrap_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestNewBootstrap_NilRounderShouldErr(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	blkc := &blockchain.BlockChain{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilRounder, err)
}

func TestNewBootstrap_NilBlockProcessorShouldErr(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilBlockExecutor, err)
}

func TestNewBootstrap_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewBootstrap_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewBootstrap_NilForkDetectorShouldErr(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}

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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilForkDetector, err)
}

func TestNewBootstrap_NilResolversContainerShouldErr(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}

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
	)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilResolverContainer, err)
}

func TestNewBootstrap_NilHeaderResolverShouldErr(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}

	errExpected := errors.New("expected error")

	resContainer := &mock.ResolversContainerStub{
		GetCalled: func(key string) (resolver process.Resolver, e error) {
			if key == string(factory.HeadersTopic) {
				return nil, errExpected
			}

			if key == string(factory.TxBlockBodyTopic) {
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
	)

	assert.Nil(t, bs)
	assert.Equal(t, errExpected, err)
}

func TestNewBootstrap_NilTxBlockBodyResolverShouldErr(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}

	errExpected := errors.New("expected error")

	resContainer := &mock.ResolversContainerStub{
		GetCalled: func(key string) (resolver process.Resolver, e error) {
			if key == string(factory.HeadersTopic) {
				return &mock.HeaderResolverMock{}, errExpected
			}

			if key == string(factory.TxBlockBodyTopic) {
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
	transient.TxBlocksCalled = func() storage.Cacher {
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

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}
		return sds
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}

		hnc.GetCalled = func(u uint64) (bytes []byte, b bool) {
			return nil, false
		}
		return hnc
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}

		return cs
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() bool {
		return true
	}

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
	)

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
		&mock.StorerStub{}, &mock.StorerStub{}, &mock.StorerStub{})
	blkc.CurrentBlockHeader = &hdr

	transient := &mock.TransientDataPoolMock{}
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
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
			return nil, false
		}

		return cs
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() bool {
		return true
	}

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
	)

	bs.RequestHeader(2)

	r := bs.SyncBlock()

	assert.Equal(t, process.ErrMissingBody, r)
}

func TestBootstrap_ShouldNotNeedToSync(t *testing.T) {
	t.Parallel()

	ebm := mock.BlockProcessorMock{}
	ebm.ProcessAndCommitCalled = func(blk *blockchain.BlockChain, hdr *block.Header, bdy *block.TxBlockBody, haveTime func() time.Duration) error {
		blk.CurrentBlockHeader = hdr
		return nil
	}

	hdr := block.Header{Nonce: 1, Round: 0}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}

		sds.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
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
			return nil, false
		}
		return hnc
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
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

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		rnd,
		&ebm,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
	)

	bs.StartSync()
	time.Sleep(200 * time.Millisecond)
	bs.StopSync()
}

func TestBootstrap_SyncShouldSyncOneBlock(t *testing.T) {
	t.Parallel()

	ebm := mock.BlockProcessorMock{}
	ebm.ProcessAndCommitCalled = func(blk *blockchain.BlockChain, hdr *block.Header, bdy *block.TxBlockBody, haveTime func() time.Duration) error {
		blk.CurrentBlockHeader = hdr
		return nil
	}

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
					BlockBodyHash: []byte("bbb")}, true
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
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("bbb"), key) && dataAvailable {
				return &block.TxBlockBody{}, true
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

	rnd, _ := round.NewRound(time.Now(), time.Now().Add(200*time.Millisecond), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		rnd,
		&ebm,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
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

	ebm := mock.BlockProcessorMock{}
	ebm.ProcessAndCommitCalled = func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error {
		return nil
	}

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
					BlockBodyHash: []byte("bbb")}, true
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
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal([]byte("bbb"), key) {
				return &block.TxBlockBody{}, true
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

	rnd, _ := round.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond), mock.SyncTimerMock{})

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		rnd,
		&ebm,
		waitTime,
		hasher,
		marshalizer,
		forkDetector,
		createMockResolversContainer(),
	)

	r := bs.SyncBlock()

	assert.Nil(t, r)
}

func TestBootstrap_ShouldSyncShouldReturnFalseWhenCurrentBlockIsNilAndRoundIndexIsZero(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}
		return sds
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}
		return hnc
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		return cs
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)

	assert.False(t, bs.ShouldSync())
}

func TestBootstrap_ShouldReturnTrueWhenCurrentBlockIsNilAndRoundIndexIsGreaterThanZero(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}
		return sds
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}
		return hnc
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		return cs
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	forkDetector.CheckForkCalled = func() bool {
		return false
	}

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
	)

	assert.True(t, bs.ShouldSync())
}

func TestBootstrap_ShouldReturnFalseWhenNodeIsSynced(t *testing.T) {
	t.Parallel()

	hdr := block.Header{Nonce: 0}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}
		return sds
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}
		return hnc
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		return cs
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	forkDetector.CheckForkCalled = func() bool {
		return false
	}

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
	)

	assert.False(t, bs.ShouldSync())
}

func TestBootstrap_ShouldReturnTrueWhenNodeIsNotSynced(t *testing.T) {
	t.Parallel()

	hdr := block.Header{Nonce: 0}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}
		return sds
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}
		return hnc
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		return cs
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() bool {
		return false
	}

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
	)

	assert.False(t, bs.ShouldSync())
}

func TestBootstrap_GetHeaderFromPoolShouldReturnNil(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}
		return sds
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}
		hnc.GetCalled = func(u uint64) (i []byte, b bool) {
			return nil, false
		}

		return hnc
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		return cs
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	forkDetector.CheckForkCalled = func() bool {
		return false
	}

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
	)

	assert.Nil(t, bs.GetHeaderFromPool(0))
}

func TestBootstrap_GetHeaderFromPoolShouldReturnHeader(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{Nonce: 0}

	transient := &mock.TransientDataPoolMock{}
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
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		return cs
	}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)

	assert.True(t, hdr == bs.GetHeaderFromPool(0))
}

func TestGetBlockFromPoolShouldReturnBlock(t *testing.T) {
	blk := &block.TxBlockBody{}

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}
		return sds
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}
		return hnc
	}
	transient.TxBlocksCalled = func() storage.Cacher {
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
	)

	assert.True(t, blk == bs.GetTxBody([]byte("aaa")))

}

//------- testing received headers

func TestBootstrap_ReceivedHeadersFoundInPoolShouldAddToForkDetector(t *testing.T) {
	t.Parallel()

	addedHash := []byte("hash")
	addedHdr := &block.Header{}

	transient := &mock.TransientDataPoolMock{}
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
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}
		return hnc
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
			return nil, false
		}
		return cs
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
	)

	bs.ReceivedHeaders(addedHash)

	assert.True(t, wasAdded)
}

func TestBootstrap_ReceivedHeadersNotFoundInPoolButFoundInStorageShouldAddToForkDetector(t *testing.T) {
	t.Parallel()

	addedHash := []byte("hash")
	addedHdr := &block.Header{}

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}
		sds.RegisterHandlerCalled = func(func(key []byte)) {
		}
		sds.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
			//not found in data pool as it was already moved out to storage unit
			//should not happen normally, but this test takes this situation into account

			return nil, false
		}
		return sds
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{}
		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
		}
		return hnc
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{}
		cs.RegisterHandlerCalled = func(i func(key []byte)) {
		}
		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
			return nil, false
		}
		return cs
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
	)

	bs.ReceivedHeaders(addedHash)

	assert.True(t, wasAdded)
}

//------- ForkChoice

func TestBootstrap_ForkChoiceNilBlockchainHeaderShouldErr(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{
			AddDataCalled:         func(key []byte, data interface{}, destShardID uint32) {},
			RegisterHandlerCalled: func(func(key []byte)) {},
		}
		return sds
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{
			RegisterHandlerCalled: func(handler func(nonce uint64)) {},
		}
		return hnc
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{
			RegisterHandlerCalled: func(i func(key []byte)) {},
		}
		return cs
	}
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)

	err := bs.ForkChoice(&block.Header{})
	assert.Equal(t, sync.ErrNilCurrentHeader, err)
}

func TestBootstrap_ForkChoiceNilParamHeaderShouldErr(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{
			AddDataCalled:         func(key []byte, data interface{}, destShardID uint32) {},
			RegisterHandlerCalled: func(func(key []byte)) {},
		}
		return sds
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		hnc := &mock.Uint64CacherStub{
			RegisterHandlerCalled: func(handler func(nonce uint64)) {},
		}
		return hnc
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{
			RegisterHandlerCalled: func(i func(key []byte)) {},
		}
		return cs
	}
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)

	blkc.CurrentBlockHeader = &block.Header{}

	err := bs.ForkChoice(nil)
	assert.Equal(t, sync.ErrNilHeader, err)
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

func TestBootstrap_ForkChoiceIsNotEmptyShouldRemove(t *testing.T) {
	t.Parallel()

	newHdrHash := []byte("new hdr hash")
	newHdrNonce := uint64(6)

	remFlags := &removedFlags{}

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return createHeadersDataPool(newHdrHash, remFlags)
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return createHeadersNoncesDataPool(newHdrNonce, newHdrHash, newHdrNonce, remFlags)
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{
			RegisterHandlerCalled: func(i func(key []byte)) {},
		}
		return cs
	}
	blkc := &blockchain.BlockChain{}
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := createForkDetector(newHdrNonce, remFlags)

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
	prevTxBlockBody := &block.TxBlockBody{
		StateBlockBody: block.StateBlockBody{RootHash: []byte("state root hash")},
	}

	//define prev header "strings"
	prevHdrHash := []byte("prev header hash")
	prevHdrBytes := []byte("prev header bytes")
	prevHdr := &block.Header{
		Signature:     []byte("sig of the prev header as to be unique in this context"),
		BlockBodyHash: prevTxBlockBodyHash,
	}

	transient := &mock.TransientDataPoolMock{}
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
	//data pool tx block bodies
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{
			RegisterHandlerCalled: func(i func(key []byte)) {},
		}
		return cs
	}

	hdrUnit := createHeadersStorage(prevHdrHash, prevHdrBytes, currentHdrHash, remFlags)
	txBlockUnit := createHeadersStorage(prevTxBlockBodyHash, prevTxBlockBodyBytes, nil, remFlags)

	//a mock blockchain with special header and tx block bodies stubs (defined above)
	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
		&mock.StorerStub{},
		txBlockUnit,
		&mock.StorerStub{},
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
				obj.(*block.Header).BlockBodyHash = prevTxBlockBodyHash
				return nil
			}
			if bytes.Equal(buff, prevTxBlockBodyBytes) {
				//bytes represent a tx block body (strings are returns from txBlockUnit.Get which is also a stub here)
				//copy only defined fields
				obj.(*block.TxBlockBody).RootHash = prevTxBlockBody.RootHash
				return nil
			}

			return nil
		},
	}
	forkDetector := createForkDetector(currentHdrNonce, remFlags)

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
	prevTxBlockBody := &block.TxBlockBody{
		StateBlockBody: block.StateBlockBody{RootHash: []byte("state root hash")},
	}

	//define prev header "strings"
	prevHdrHash := []byte("prev header hash")
	prevHdrBytes := []byte("prev header bytes")
	prevHdr := &block.Header{
		Signature:     []byte("sig of the prev header as to be unique in this context"),
		BlockBodyHash: prevTxBlockBodyHash,
	}

	transient := &mock.TransientDataPoolMock{}
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
	//data pool tx block bodies
	transient.TxBlocksCalled = func() storage.Cacher {
		cs := &mock.CacherStub{
			RegisterHandlerCalled: func(i func(key []byte)) {},
		}
		return cs
	}

	hdrUnit := createHeadersStorage(prevHdrHash, prevHdrBytes, currentHdrHash, remFlags)
	txBlockUnit := createHeadersStorage(prevTxBlockBodyHash, prevTxBlockBodyBytes, nil, remFlags)

	//a mock blockchain with special header and tx block bodies stubs (defined above)
	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
		&mock.StorerStub{},
		txBlockUnit,
		&mock.StorerStub{},
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
				obj.(*block.Header).BlockBodyHash = prevTxBlockBodyHash
				return nil
			}
			if bytes.Equal(buff, prevTxBlockBodyBytes) {
				//bytes represent a tx block body (strings are returns from txBlockUnit.Get which is also a stub here)
				//copy only defined fields
				obj.(*block.TxBlockBody).RootHash = prevTxBlockBody.RootHash
				return nil
			}

			return nil
		},
	}
	forkDetector := createForkDetector(currentHdrNonce, remFlags)
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

	requestedHash := []byte("requested hash")
	txBlock := &block.TxBlockBody{}

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{
			AddDataCalled:         func(key []byte, data interface{}, destShardID uint32) {},
			RegisterHandlerCalled: func(func(key []byte)) {},
		}
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{
			RegisterHandlerCalled: func(handler func(nonce uint64)) {},
		}
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{
			RegisterHandlerCalled: func(i func(key []byte)) {},
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal(key, requestedHash) {
					return txBlock, true
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
		&mock.StorerStub{},
	)
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)
	txBlockRecovered := bs.GetTxBody(requestedHash)

	assert.True(t, txBlockRecovered == txBlock)
}

func TestBootstrap_GetTxBodyHavingHashNotFoundInCacherOrStorageShouldRetNil(t *testing.T) {
	t.Parallel()

	requestedHash := []byte("requested hash")

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{
			AddDataCalled:         func(key []byte, data interface{}, destShardID uint32) {},
			RegisterHandlerCalled: func(func(key []byte)) {},
		}
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{
			RegisterHandlerCalled: func(handler func(nonce uint64)) {},
		}
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{
			RegisterHandlerCalled: func(i func(key []byte)) {},
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}
	}

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
		&mock.StorerStub{},
	)
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)
	txBlockRecovered := bs.GetTxBody(requestedHash)

	assert.Nil(t, txBlockRecovered)
}

func TestBootstrap_GetTxBodyHavingHashFoundInStorageShouldWork(t *testing.T) {
	t.Parallel()

	requestedHash := []byte("requested hash")
	txBlock := &block.TxBlockBody{
		StateBlockBody: block.StateBlockBody{RootHash: []byte("root hash")},
	}

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{
			AddDataCalled:         func(key []byte, data interface{}, destShardID uint32) {},
			RegisterHandlerCalled: func(func(key []byte)) {},
		}
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{
			RegisterHandlerCalled: func(handler func(nonce uint64)) {},
		}
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{
			RegisterHandlerCalled: func(i func(key []byte)) {},
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}
	}

	txBlockUnit := &mock.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			if bytes.Equal(key, requestedHash) {
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
		&mock.StorerStub{},
	)
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)
	txBlockRecovered := bs.GetTxBody(requestedHash)

	assert.Equal(t, txBlock, txBlockRecovered)
}

func TestBootstrap_GetTxBodyHavingHashMarshalizerFailShouldRemoveAndRetNil(t *testing.T) {
	t.Parallel()

	removedCalled := false
	requestedHash := []byte("requested hash")

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return errors.New("marshalizer failure")
		},
	}

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{
			AddDataCalled:         func(key []byte, data interface{}, destShardID uint32) {},
			RegisterHandlerCalled: func(func(key []byte)) {},
		}
	}
	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{
			RegisterHandlerCalled: func(handler func(nonce uint64)) {},
		}
	}
	transient.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{
			RegisterHandlerCalled: func(i func(key []byte)) {},
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}
	}

	txBlockUnit := &mock.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			if bytes.Equal(key, requestedHash) {
				return make([]byte, 0), nil
			}

			return nil, errors.New("not found")
		},
		RemoveCalled: func(key []byte) error {
			if bytes.Equal(key, requestedHash) {
				removedCalled = true
			}
			return nil
		},
	}

	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
		&mock.StorerStub{},
		txBlockUnit,
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
	)
	rnd := &mock.RounderMock{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}

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
	)
	txBlockRecovered := bs.GetTxBody(requestedHash)

	assert.Nil(t, txBlockRecovered)
	assert.True(t, removedCalled)
}

func TestBootstrap_CreateEmptyBlockShouldReturnNilWhenMarshalErr(t *testing.T) {
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
	transient.TxBlocksCalled = func() storage.Cacher {
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

	blkExec.RevertAccountStateCalled = func() {
	}

	blkExec.CreateEmptyBlockBodyCalled = func(shardId uint32, round int32) *block.TxBlockBody {
		return &block.TxBlockBody{}
	}

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
	)

	blk, hdr := bs.CreateAndCommitEmptyBlock(0)

	assert.Nil(t, blk)
	assert.Nil(t, hdr)
}

func TestBootstrap_CreateEmptyBlockShouldReturnNilWhenCommitBlockErr(t *testing.T) {
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
	transient.TxBlocksCalled = func() storage.Cacher {
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

	blkc.CurrentBlockHeader = &block.Header{Nonce: 1}

	blkExec.RevertAccountStateCalled = func() {
	}

	blkExec.CreateEmptyBlockBodyCalled = func(shardId uint32, round int32) *block.TxBlockBody {
		return &block.TxBlockBody{}
	}

	err := errors.New("error")
	blkExec.CommitBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error {
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
	)

	blk, hdr := bs.CreateAndCommitEmptyBlock(0)

	assert.Nil(t, blk)
	assert.Nil(t, hdr)
}

func TestBootstrap_CreateEmptyBlockShouldWork(t *testing.T) {
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
	transient.TxBlocksCalled = func() storage.Cacher {
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

	blkExec.RevertAccountStateCalled = func() {
	}

	blkExec.CreateEmptyBlockBodyCalled = func(shardId uint32, round int32) *block.TxBlockBody {
		return &block.TxBlockBody{}
	}

	blkExec.CommitBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error {
		return nil
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
	)

	blk, hdr := bs.CreateAndCommitEmptyBlock(0)

	assert.NotNil(t, blk)
	assert.NotNil(t, hdr)
}

func TestBootstrap_AddSyncStateListenerShouldAppendAnotherListener(t *testing.T) {
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
	transient.TxBlocksCalled = func() storage.Cacher {
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

	blkExec.RevertAccountStateCalled = func() {
	}

	blkExec.CreateEmptyBlockBodyCalled = func(shardId uint32, round int32) *block.TxBlockBody {
		return &block.TxBlockBody{}
	}

	blkExec.CommitBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error {
		return nil
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
	transient.TxBlocksCalled = func() storage.Cacher {
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

	blkExec.RevertAccountStateCalled = func() {
	}

	blkExec.CreateEmptyBlockBodyCalled = func(shardId uint32, round int32) *block.TxBlockBody {
		return &block.TxBlockBody{}
	}

	blkExec.CommitBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error {
		return nil
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
