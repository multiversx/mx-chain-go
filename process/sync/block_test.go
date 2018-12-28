package sync_test

import (
	"bytes"
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
	"github.com/stretchr/testify/assert"
)

// WaitTime defines the time in milliseconds until node waits the requested info from the network
const WaitTime = time.Duration(100 * time.Millisecond)

//------- NewBootstrap

func TestNewBootstrap_NilTransientDataHolderShouldErr(t *testing.T) {
	t.Parallel()

	blkc := &blockchain.BlockChain{}
	round := &chronology.Round{}
	blkExec := &mock.BlockProcessorMock{}

	bs, err := sync.NewBootstrap(nil, blkc, round, blkExec, WaitTime)

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

	bs, err := sync.NewBootstrap(transient, blkc, round, blkExec, WaitTime)

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

	bs, err := sync.NewBootstrap(transient, blkc, round, blkExec, WaitTime)

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

	bs, err := sync.NewBootstrap(transient, blkc, round, blkExec, WaitTime)

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

	bs, err := sync.NewBootstrap(transient, nil, round, blkExec, WaitTime)

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

	bs, err := sync.NewBootstrap(transient, blkc, nil, blkExec, WaitTime)

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

	bs, err := sync.NewBootstrap(transient, blkc, round, nil, WaitTime)

	assert.Nil(t, bs)
	assert.Equal(t, process.ErrNilBlockExecutor, err)
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

	bs, err := sync.NewBootstrap(transient, blkc, round, blkExec, WaitTime)

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

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		WaitTime)

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

		sds.SearchDataCalled = func(key []byte) (shardValuesPairs map[uint32]interface{}) {
			m := make(map[uint32]interface{})

			if bytes.Equal([]byte("aaa"), key) {
				m[0] = &block.Header{Nonce: 2}
			}

			return m
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

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		WaitTime)

	bs.RequestHeader(2)

	r := bs.SyncBlock()

	assert.Equal(t, process.ErrMissingBody, r)
}

func TestBootstrap_ShouldNotNeedToSync(t *testing.T) {
	t.Parallel()

	ebm := mock.BlockProcessorMock{}
	ebm.ProcessBlockCalled = func(blk *blockchain.BlockChain, hdr *block.Header, bdy *block.TxBlockBody) error {
		blk.CurrentBlockHeader = hdr
		return nil
	}

	hdr := block.Header{Nonce: 1, Round: 0}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}

		sds.SearchDataCalled = func(key []byte) (shardValuesPairs map[uint32]interface{}) {
			return nil
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

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		chronology.NewRound(time.Now(), time.Now().Add(0*time.Millisecond), time.Duration(100*time.Millisecond)),
		&ebm,
		WaitTime)

	bs.StartSync()
	time.Sleep(200 * time.Millisecond)
	bs.StopSync()
}

func TestBootstrap_SyncShouldSyncOneBlock(t *testing.T) {
	t.Parallel()

	ebm := mock.BlockProcessorMock{}
	ebm.ProcessBlockCalled = func(blk *blockchain.BlockChain, hdr *block.Header, bdy *block.TxBlockBody) error {
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

		sds.SearchDataCalled = func(key []byte) (shardValuesPairs map[uint32]interface{}) {
			mutDataAvailable.RLock()
			defer mutDataAvailable.RUnlock()

			m := make(map[uint32]interface{})

			if bytes.Equal([]byte("aaa"), key) && dataAvailable {
				m[0] = &block.Header{
					Nonce:         2,
					Round:         1,
					BlockBodyType: block.TxBlock,
					BlockBodyHash: []byte("bbb")}
			}

			return m
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

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		chronology.NewRound(time.Now(), time.Now().Add(200*time.Millisecond), time.Duration(100*time.Millisecond)),
		&ebm,
		WaitTime)

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
	ebm.ProcessBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
		return nil
	}

	hdr := block.Header{Nonce: 1}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}

		sds.SearchDataCalled = func(key []byte) (shardValuesPairs map[uint32]interface{}) {
			m := make(map[uint32]interface{})

			if bytes.Equal([]byte("aaa"), key) {
				m[0] = &block.Header{
					Nonce:         2,
					Round:         1,
					BlockBodyType: block.TxBlock,
					BlockBodyHash: []byte("bbb")}
			}

			return m
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

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&ebm,
		WaitTime)

	r := bs.SyncBlock()

	assert.Nil(t, r)
}

func TestBootstrap_ShouldSyncShouldReturnFalseWhenCurrentBlockIsNilAndRoundIndexIsZero(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}
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

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		WaitTime)

	assert.False(t, bs.ShouldSync())
}

func TestBootstrap_ShouldReturnTrueWhenCurrentBlockIsNilAndRoundIndexIsGreaterThanZero(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}
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

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		WaitTime)

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

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		WaitTime)

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

	bs, _ := sync.NewBootstrap(
		transient,
		&blkc,
		chronology.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		WaitTime)

	assert.False(t, bs.ShouldSync())
}

func TestBootstrap_GetHeaderFromPoolShouldReturnNil(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}
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

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		WaitTime)

	assert.Nil(t, bs.GetHeaderFromPool(0))
}

func TestBootstrap_GetHeaderFromPoolShouldReturnHeader(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{Nonce: 0}

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}
		sds.SearchDataCalled = func(key []byte) (shardValuesPairs map[uint32]interface{}) {
			m := make(map[uint32]interface{})
			if bytes.Equal([]byte("aaa"), key) {
				m[0] = hdr
			}
			return m
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

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		WaitTime)

	assert.True(t, hdr == bs.GetHeaderFromPool(0))
}

func TestBootstrap_GetBlockFromPoolShouldReturnNil(t *testing.T) {
	t.Parallel()

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}
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

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		WaitTime)

	r := bs.GetTxBodyFromPool([]byte("aaa"))

	assert.Nil(t, r)
}

func TestGetBlockFromPoolShouldReturnBlock(t *testing.T) {
	blk := &block.TxBlockBody{}

	transient := &mock.TransientDataPoolMock{}
	transient.HeadersCalled = func() data.ShardedDataCacherNotifier {
		sds := &mock.ShardedDataStub{}
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

	bs, _ := sync.NewBootstrap(
		transient,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.BlockProcessorMock{},
		WaitTime)

	assert.True(t, blk == bs.GetTxBodyFromPool([]byte("aaa")))

}
