package block_test

import (
	"bytes"
	"errors"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	blproc "github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/stretchr/testify/assert"
)

var r *rand.Rand
var mutex sync.Mutex

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func haveTime() time.Duration {
	return time.Duration(2000 * time.Millisecond)
}

func createTestBlockchain() *mock.BlockChainMock {
	return &mock.BlockChainMock{}
}

func generateTestCache() storage.Cacher {
	cache, _ := storage.NewCache(storage.LRUCache, 1000, 1)
	return cache
}

func generateTestUnit() storage.Storer {
	memDB, _ := memorydb.New()

	storer, _ := storage.NewStorageUnit(
		generateTestCache(),
		memDB,
	)

	return storer
}

func initDataPool() *mock.PoolsHolderStub {
	sdp := &mock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{
				RegisterHandlerCalled: func(i func(key []byte)) {},
				ShardDataStoreCalled: func(id string) (c storage.Cacher) {
					return &mock.CacherStub{
						PeekCalled: func(key []byte) (value interface{}, ok bool) {
							if reflect.DeepEqual(key, []byte("tx1_hash")) {
								return &transaction.Transaction{Nonce: 10}, true
							}
							return nil, false
						},
						KeysCalled: func() [][]byte {
							return [][]byte{[]byte("key1"), []byte("key2")}
						},
						LenCalled: func() int {
							return 0
						},
					}
				},
				RemoveSetOfDataFromPoolCalled: func(keys [][]byte, id string) {},
				SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
			}
		},
		HeadersNoncesCalled: func() dataRetriever.Uint64Cacher {
			return &mock.Uint64CacherStub{
				PutCalled: func(u uint64, i []byte) bool {
					return true
				},
			}
		},
		MetaBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				KeysCalled: func() [][]byte {
					return nil
				},
				LenCalled: func() int {
					return 0
				},
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				RegisterHandlerCalled: func(i func(key []byte)) {},
			}
		},
		MiniBlocksCalled: func() storage.Cacher {
			cs := &mock.CacherStub{}
			cs.RegisterHandlerCalled = func(i func(key []byte)) {
			}
			cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal([]byte("bbb"), key) {
					return make(block.MiniBlockSlice, 0), true
				}

				return nil, false
			}
			cs.PeekCalled = func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal([]byte("bbb"), key) {
					return make(block.MiniBlockSlice, 0), true
				}

				return nil, false
			}
			cs.RegisterHandlerCalled = func(i func(key []byte)) {}
			cs.RemoveCalled = func(key []byte) {}
			return cs
		},
		HeadersCalled: func() storage.Cacher {
			cs := &mock.CacherStub{}
			cs.RegisterHandlerCalled = func(i func(key []byte)) {
			}
			return cs
		},
	}
	return sdp
}

func initMetaDataPool() *mock.MetaPoolsHolderStub {
	mdp := &mock.MetaPoolsHolderStub{
		MetaBlockNoncesCalled: func() dataRetriever.Uint64Cacher {
			return &mock.Uint64CacherStub{
				PutCalled: func(u uint64, i []byte) bool {
					return true
				},
			}
		},
		MetaChainBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				KeysCalled: func() [][]byte {
					return nil
				},
				LenCalled: func() int {
					return 0
				},
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				RegisterHandlerCalled: func(i func(key []byte)) {},
			}
		},
		MiniBlockHashesCalled: func() dataRetriever.ShardedDataCacherNotifier {
			sdc := &mock.ShardedDataStub{}
			sdc.RegisterHandlerCalled = func(i func(key []byte)) {
			}
			sdc.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal([]byte("bbb"), key) {
					return make(block.MiniBlockSlice, 0), true
				}

				return nil, false
			}
			sdc.RegisterHandlerCalled = func(i func(key []byte)) {}
			sdc.RemoveDataCalled = func(key []byte, cacheId string) {

			}
			return sdc
		},
		ShardHeadersCalled: func() storage.Cacher {
			cs := &mock.CacherStub{}
			cs.RegisterHandlerCalled = func(i func(key []byte)) {
			}
			cs.PeekCalled = func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal([]byte("hdr_hash1"), key) {
					return &block.Header{Nonce: 1}, true
				}
				return nil, false
			}
			cs.LenCalled = func() int {
				return 0
			}
			cs.RemoveCalled = func(key []byte) {
			}
			cs.KeysCalled = func() [][]byte {
				return nil
			}
			return cs
		},
	}
	return mdp
}

func initStore() *dataRetriever.ChainStorer {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, generateTestUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, generateTestUnit())
	store.AddStorer(dataRetriever.HdrNonceHashDataUnit, generateTestUnit())
	return store
}

func createDummyMetaBlock(destShardId uint32, senderShardId uint32, miniBlockHashes ...[]byte) data.HeaderHandler {
	metaBlock := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{
				ShardMiniBlockHeaders: make([]block.ShardMiniBlockHeader, len(miniBlockHashes)),
			},
		},
	}

	for idx, mbHash := range miniBlockHashes {
		metaBlock.ShardInfo[0].ShardMiniBlockHeaders[idx].ReceiverShardId = destShardId
		metaBlock.ShardInfo[0].ShardMiniBlockHeaders[idx].SenderShardId = senderShardId
		metaBlock.ShardInfo[0].ShardMiniBlockHeaders[idx].Hash = mbHash
	}

	return metaBlock
}

func createDummyMiniBlock(
	txHash string,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	destShardId uint32,
	senderShardId uint32) (*block.MiniBlock, []byte) {

	miniblock := &block.MiniBlock{
		TxHashes:        [][]byte{[]byte(txHash)},
		ReceiverShardID: destShardId,
		SenderShardID:   senderShardId,
	}

	buff, _ := marshalizer.Marshal(miniblock)
	hash := hasher.Compute(string(buff))

	return miniblock, hash
}

func isInTxHashes(searched []byte, list [][]byte) bool {
	for _, txHash := range list {
		if bytes.Equal(txHash, searched) {
			return true
		}
	}
	return false
}

func computeHash(data interface{}, marshalizer marshal.Marshalizer, hasher hashing.Hasher) []byte {
	buff, _ := marshalizer.Marshal(data)
	return hasher.Compute(string(buff))
}

type wrongBody struct {
}

func (wr wrongBody) IntegrityAndValidity() error {
	return nil
}

func TestBlockProcessor_CheckBlockValidity(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	bp, _ := blproc.NewShardProcessor(
		tdp,
		initStore(),
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		func(destShardID uint32, txHashes [][]byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	blkc := createTestBlockchain()
	body := &block.Body{}
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = 0
	hdr.PrevHash = []byte("X")
	err := bp.CheckBlockValidity(blkc, hdr, body)
	assert.Equal(t, process.ErrInvalidBlockHash, err)

	hdr.PrevHash = []byte("")
	err = bp.CheckBlockValidity(blkc, hdr, body)
	assert.Nil(t, err)

	hdr.Nonce = 2
	err = bp.CheckBlockValidity(blkc, hdr, body)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)

	hdr.Nonce = 1
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.Header{Nonce: 1}
	}
	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = 0
	err = bp.CheckBlockValidity(blkc, hdr, body)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)

	hdr.Nonce = 2
	hdr.PrevHash = []byte("X")
	err = bp.CheckBlockValidity(blkc, hdr, body)
	assert.Equal(t, process.ErrInvalidBlockHash, err)

	hdr.Nonce = 3
	hdr.PrevHash = []byte("")
	err = bp.CheckBlockValidity(blkc, hdr, body)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)

	hdr.Nonce = 2
	marshalizerMock := mock.MarshalizerMock{}
	hasherMock := mock.HasherMock{}
	prevHeader, _ := marshalizerMock.Marshal(blkc.GetCurrentBlockHeader())
	hdr.PrevHash = hasherMock.Compute(string(prevHeader))

	err = bp.CheckBlockValidity(blkc, hdr, body)
	assert.Nil(t, err)
}

func TestVerifyStateRoot_ShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	rootHash := []byte("root hash to be tested")
	accounts := &mock.AccountsStub{
		RootHashCalled: func() []byte {
			return rootHash
		},
	}
	store := initStore()

	bp, _ := blproc.NewShardProcessor(
		tdp,
		store,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		accounts,
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		func(destShardID uint32, txHashes [][]byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	assert.True(t, bp.VerifyStateRoot(rootHash))
}

//------- ComputeNewNoncePrevHash

func TestBlockProcessor_computeHeaderHashMarshalizerFail1ShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	marshalizer := &mock.MarshalizerStub{}
	bp, _ := blproc.NewShardProcessor(
		tdp,
		initStore(),
		&mock.HasherStub{},
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		func(destShardID uint32, txHashes [][]byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	hdr, txBlock := createTestHdrTxBlockBody()
	expectedError := errors.New("marshalizer fail")
	marshalizer.MarshalCalled = func(obj interface{}) (bytes []byte, e error) {
		if hdr == obj {
			return nil, expectedError
		}

		if reflect.DeepEqual(txBlock, obj) {
			return []byte("txBlockBodyMarshalized"), nil
		}
		return nil, nil
	}
	_, err := bp.ComputeHeaderHash(hdr)
	assert.Equal(t, expectedError, err)
}

func TestBlockPorcessor_ComputeNewNoncePrevHashShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	marshalizer := &mock.MarshalizerStub{}
	hasher := &mock.HasherStub{}
	bp, _ := blproc.NewShardProcessor(
		tdp,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		func(destShardID uint32, txHashes [][]byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	hdr, txBlock := createTestHdrTxBlockBody()
	marshalizer.MarshalCalled = func(obj interface{}) (bytes []byte, e error) {
		if hdr == obj {
			return []byte("hdrHeaderMarshalized"), nil
		}
		if reflect.DeepEqual(txBlock, obj) {
			return []byte("txBlockBodyMarshalized"), nil
		}
		return nil, nil
	}
	hasher.ComputeCalled = func(s string) []byte {
		if s == "hdrHeaderMarshalized" {
			return []byte("hdr hash")
		}
		if s == "txBlockBodyMarshalized" {
			return []byte("tx block body hash")
		}
		return nil
	}
	_, err := bp.ComputeHeaderHash(hdr)
	assert.Nil(t, err)
}

func TestBlockPorcessor_DisplayHeaderShouldWork(t *testing.T) {
	lines := blproc.DisplayHeader(&block.Header{})
	assert.Equal(t, 10, len(lines))
}
