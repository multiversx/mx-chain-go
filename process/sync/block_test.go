package sync_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	goSync "sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
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

//------- NewBootstrap

func TestNewBootstrap_NilTransientDataHolderShouldErr(t *testing.T) {
	t.Parallel()

	blkc := &blockchain.BlockChain{}
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, err := sync.NewBootstrap(nil, blkc, round, blkExec, waitTime, marshalizer, forkDetector)

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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, err := sync.NewBootstrap(transient, blkc, round, blkExec, waitTime, marshalizer, forkDetector)

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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, err := sync.NewBootstrap(transient, blkc, round, blkExec, waitTime, marshalizer, forkDetector)

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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, err := sync.NewBootstrap(transient, blkc, round, blkExec, waitTime, marshalizer, forkDetector)

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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, err := sync.NewBootstrap(transient, nil, round, blkExec, waitTime, marshalizer, forkDetector)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestNewBootstrap_NilRoundShouldErr(t *testing.T) {
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
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, err := sync.NewBootstrap(transient, blkc, nil, blkExec, waitTime, marshalizer, forkDetector)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilRound, err)
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
	round := &chronology.Round{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, err := sync.NewBootstrap(transient, blkc, round, nil, waitTime, marshalizer, forkDetector)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilBlockExecutor, err)
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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}

	bs, err := sync.NewBootstrap(transient, blkc, round, blkExec, waitTime, marshalizer, nil)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilForkDetector, err)
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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, err := sync.NewBootstrap(transient, blkc, round, blkExec, waitTime, nil, forkDetector)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilMarshalizer, err)
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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, err := sync.NewBootstrap(transient, blkc, round, blkExec, waitTime, marshalizer, forkDetector)

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
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		waitTime,
		marshalizer,
		forkDetector)

	bs.RequestHeaderHandler = func(nonce uint64) {}
	bs.RequestTxBodyHandler = func(hash []byte) {}

	r := bs.SyncBlock()

	assert.Equal(t, process.ErrMissingHeader, r)
}

func TestBootstrap_ShouldReturnMissingBody(t *testing.T) {
	t.Parallel()

	hdr := block.Header{Nonce: 1}
	blkc := blockchain.BlockChain{}
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
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		waitTime,
		marshalizer,
		forkDetector)

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
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	forkDetector.CheckForkCalled = func() bool {
		return false
	}

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		chronology.NewRound(time.Now(), time.Now().Add(0*time.Millisecond), time.Duration(100*time.Millisecond)),
		&ebm,
		waitTime,
		marshalizer,
		forkDetector)

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
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	forkDetector.CheckForkCalled = func() bool {
		return false
	}

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		chronology.NewRound(time.Now(), time.Now().Add(200*time.Millisecond), time.Duration(100*time.Millisecond)),
		&ebm,
		waitTime,
		marshalizer,
		forkDetector)

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

	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&ebm,
		waitTime,
		marshalizer,
		forkDetector)

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
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		waitTime,
		marshalizer,
		forkDetector)

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
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	forkDetector.CheckForkCalled = func() bool {
		return false
	}

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		waitTime,
		marshalizer,
		forkDetector)

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

	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	forkDetector.CheckForkCalled = func() bool {
		return false
	}

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		waitTime,
		marshalizer,
		forkDetector)

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

	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}
	forkDetector.CheckForkCalled = func() bool {
		return false
	}

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		chronology.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		waitTime,
		marshalizer,
		forkDetector)

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

	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	forkDetector.CheckForkCalled = func() bool {
		return false
	}

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		waitTime,
		marshalizer,
		forkDetector)

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
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		waitTime,
		marshalizer,
		forkDetector)

	assert.True(t, hdr == bs.GetHeaderFromPool(0))
}

//func TestBootstrap_GetBlockFromPoolShouldReturnNil(t *testing.T) {
//	t.Parallel()
//
//	transient := &mock.TransientDataPoolMock{}
//	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
//		sds := &mock.ShardedDataStub{}
//		sds.RegisterHandlerCalled = func(func(key []byte)) {
//		}
//		return sds
//	}
//	transient.HeadersNoncesCalled = func() data.Uint64Cacher {
//		hnc := &mock.Uint64CacherStub{}
//		hnc.RegisterHandlerCalled = func(handler func(nonce uint64)) {
//		}
//		return hnc
//	}
//	transient.TxBlocksCalled = func() storage.Cacher {
//		cs := &mock.CacherStub{}
//		cs.RegisterHandlerCalled = func(i func(key []byte)) {
//		}
//		cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
//			return nil, false
//		}
//		return cs
//	}
//
//	marshalizer := &mock.MarshalizerMock{}
//	forkDetector := &mock.ForkDetectorMock{}
//
//	bs, _ := sync.NewBootstrap(
//		transient,
//		&blockchain.BlockChain{},
//		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
//		&mock.BlockProcessorMock{},
//		waitTime,
//		marshalizer,
//		forkDetector)
//
//	r := bs.GetTxBodyHavingHash([]byte("aaa"))
//
//	assert.Nil(t, r)
//}

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
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		waitTime,
		marshalizer,
		forkDetector)

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

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		waitTime,
		marshalizer,
		forkDetector)

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

	bs, _ := sync.NewBootstrap(
		transient,
		blkc,
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		waitTime,
		marshalizer,
		forkDetector)

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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, _ := sync.NewBootstrap(transient, blkc, round, blkExec, waitTime, marshalizer, forkDetector)

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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, _ := sync.NewBootstrap(transient, blkc, round, blkExec, waitTime, marshalizer, forkDetector)

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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := createForkDetector(newHdrNonce, remFlags)

	bs, _ := sync.NewBootstrap(transient, blkc, round, blkExec, waitTime, marshalizer, forkDetector)

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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}

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
	bs, _ := sync.NewBootstrap(transient, blkc, round, blkExec, waitTime, marshalizer, forkDetector)

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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}

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
	bs, _ := sync.NewBootstrap(transient, blkc, round, blkExec, waitTime, marshalizer, forkDetector)

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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, _ := sync.NewBootstrap(transient, blkc, round, blkExec, waitTime, marshalizer, forkDetector)
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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}
	marshalizer := &mock.MarshalizerMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, _ := sync.NewBootstrap(transient, blkc, round, blkExec, waitTime, marshalizer, forkDetector)
	txBlockRecovered := bs.GetTxBody(requestedHash)

	assert.Nil(t, txBlockRecovered)
}

func TestBootstrap_GetTxBodyHavingHashFoundInStorageShouldWork(t *testing.T) {
	t.Parallel()

	requestedHash := []byte("requested hash")
	txBlock := &block.TxBlockBody{
		StateBlockBody: block.StateBlockBody{RootHash: []byte("root hash")},
	}

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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, _ := sync.NewBootstrap(transient, blkc, round, blkExec, waitTime, marshalizer, forkDetector)
	txBlockRecovered := bs.GetTxBody(requestedHash)

	assert.Equal(t, txBlock, txBlockRecovered)
}

func TestBootstrap_GetTxBodyHavingHashMarshalizerFailShouldRemoveAndRetNil(t *testing.T) {
	t.Parallel()

	removedCalled := false
	requestedHash := []byte("requested hash")

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
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}
	forkDetector := &mock.ForkDetectorMock{}

	bs, _ := sync.NewBootstrap(transient, blkc, round, blkExec, waitTime, marshalizer, forkDetector)
	txBlockRecovered := bs.GetTxBody(requestedHash)

	assert.Nil(t, txBlockRecovered)
	assert.True(t, removedCalled)
}
