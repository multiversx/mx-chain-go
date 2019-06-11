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
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/storageUnit"
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
				PutCalled: func(u uint64, i interface{}) bool {
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
				PutCalled: func(u uint64, i interface{}) bool {
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
	store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit, generateTestUnit())
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
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.ServiceContainerMock{},
		tdp,
		store,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		accounts,
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		true,
		&mock.RequestHandlerMock{},
	)
	assert.True(t, bp.VerifyStateRoot(rootHash))
}

//------- ComputeNewNoncePrevHash

func TestBlockProcessor_computeHeaderHashMarshalizerFail1ShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	marshalizer := &mock.MarshalizerStub{}
	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		&mock.HasherStub{},
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		true,
		&mock.RequestHandlerMock{},
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

func TestBaseProcessor_SetLastNotarizedHeadersSliceNil(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewOneShardCoordinatorMock())

	err := base.SetLastNotarizedHeadersSlice(nil, false)

	assert.Equal(t, process.ErrNotarizedHdrsSliceIsNil, err)
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceNotEnoughHeaders(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewOneShardCoordinatorMock())

	err := base.SetLastNotarizedHeadersSlice(make(map[uint32]data.HeaderHandler, 0), false)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceOneShardWrongType(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewOneShardCoordinatorMock())

	lastNotHdrs := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	lastNotHdrs[0] = &block.MetaBlock{}
	err := base.SetLastNotarizedHeadersSlice(lastNotHdrs, true)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceOneShardGood(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewOneShardCoordinatorMock())

	lastNotHdrs := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	err := base.SetLastNotarizedHeadersSlice(lastNotHdrs, true)

	assert.Nil(t, err)
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceOneShardMetaMissing(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewOneShardCoordinatorMock())

	lastNotHdrs := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	lastNotHdrs[sharding.MetachainShardId] = nil
	err := base.SetLastNotarizedHeadersSlice(lastNotHdrs, true)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceOneShardMetaWrongType(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewOneShardCoordinatorMock())

	lastNotHdrs := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	lastNotHdrs[sharding.MetachainShardId] = &block.Header{}
	err := base.SetLastNotarizedHeadersSlice(lastNotHdrs, true)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceMultiShardGood(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewMultiShardsCoordinatorMock(5))

	lastNotHdrs := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(5))
	err := base.SetLastNotarizedHeadersSlice(lastNotHdrs, true)

	assert.Nil(t, err)
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceMultiShardWithoutMeta(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewMultiShardsCoordinatorMock(5))

	lastNotHdrs := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(5))
	lastNotHdrs[sharding.MetachainShardId] = nil
	err := base.SetLastNotarizedHeadersSlice(lastNotHdrs, false)

	assert.Nil(t, err)
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceMultiShardNotEnough(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewMultiShardsCoordinatorMock(5))

	lastNotHdrs := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(4))
	lastNotHdrs[sharding.MetachainShardId] = nil
	err := base.SetLastNotarizedHeadersSlice(lastNotHdrs, false)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func createShardProcessHeadersToSaveLastNoterized(
	highestNonce uint64,
	genesisHdr data.HeaderHandler,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) []data.HeaderHandler {
	rootHash := []byte("roothash")
	processedHdrs := make([]data.HeaderHandler, 0)

	headerMarsh, _ := marshalizer.Marshal(genesisHdr)
	headerHash := hasher.Compute(string(headerMarsh))

	for i := uint64(1); i <= highestNonce; i++ {
		hdr := &block.Header{
			Nonce:         i,
			Round:         uint32(i),
			Signature:     rootHash,
			RandSeed:      rootHash,
			PrevRandSeed:  rootHash,
			PubKeysBitmap: rootHash,
			RootHash:      rootHash,
			PrevHash:      headerHash}
		processedHdrs = append(processedHdrs, hdr)

		headerMarsh, _ = marshalizer.Marshal(hdr)
		headerHash = hasher.Compute(string(headerMarsh))
	}

	return processedHdrs
}

func createMetaProcessHeadersToSaveLastNoterized(
	highestNonce uint64,
	genesisHdr data.HeaderHandler,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) []data.HeaderHandler {
	rootHash := []byte("roothash")
	processedHdrs := make([]data.HeaderHandler, 0)

	headerMarsh, _ := marshalizer.Marshal(genesisHdr)
	headerHash := hasher.Compute(string(headerMarsh))

	for i := uint64(1); i <= highestNonce; i++ {
		hdr := &block.MetaBlock{
			Nonce:         i,
			Round:         uint32(i),
			Signature:     rootHash,
			RandSeed:      rootHash,
			PrevRandSeed:  rootHash,
			PubKeysBitmap: rootHash,
			RootHash:      rootHash,
			PrevHash:      headerHash}
		processedHdrs = append(processedHdrs, hdr)

		headerMarsh, _ = marshalizer.Marshal(hdr)
		headerHash = hasher.Compute(string(headerMarsh))
	}

	return processedHdrs
}

func TestBaseProcessor_SaveLastNoterizedHdrLastNotSliceNotSet(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewMultiShardsCoordinatorMock(5))
	base.SetHasher(mock.HasherMock{})
	base.SetMarshalizer(&mock.MarshalizerMock{})
	prHdrs := createShardProcessHeadersToSaveLastNoterized(10, &block.Header{}, mock.HasherMock{}, &mock.MarshalizerMock{})

	err := base.SaveLastNotarizedHeader(2, prHdrs)

	assert.Equal(t, process.ErrNotarizedHdrsSliceIsNil, err)
}

func TestBaseProcessor_SaveLastNoterizedHdrLastNotShardIdMissmatch(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	base.SetHasher(mock.HasherMock{})
	base.SetMarshalizer(&mock.MarshalizerMock{})
	_ = base.SetLastNotarizedHeadersSlice(createGenesisBlocks(shardCoordinator), true)
	prHdrs := createShardProcessHeadersToSaveLastNoterized(10, &block.Header{}, mock.HasherMock{}, &mock.MarshalizerMock{})

	err := base.SaveLastNotarizedHeader(6, prHdrs)

	assert.Equal(t, process.ErrShardIdMissmatch, err)
}

func TestBaseProcessor_SaveLastNoterizedHdrLastNotHdrNil(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	base.SetHasher(mock.HasherMock{})
	base.SetMarshalizer(&mock.MarshalizerMock{})

	// make it wrong
	shardId := uint32(2)
	genesisBlock := createGenesisBlocks(shardCoordinator)
	genesisBlock[shardId] = nil

	_ = base.SetLastNotarizedHeadersSlice(genesisBlock, true)
	prHdrs := createShardProcessHeadersToSaveLastNoterized(10, &block.Header{}, mock.HasherMock{}, &mock.MarshalizerMock{})

	err := base.SaveLastNotarizedHeader(shardId, prHdrs)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestBaseProcessor_SaveLastNoterizedHdrLastNotWrongTypeShard(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	base.SetHasher(mock.HasherMock{})
	base.SetMarshalizer(&mock.MarshalizerMock{})

	// make it wrong
	shardId := uint32(2)
	genesisBlock := createGenesisBlocks(shardCoordinator)
	genesisBlock[shardId] = &block.MetaBlock{Nonce: 0}

	_ = base.SetLastNotarizedHeadersSlice(genesisBlock, true)
	prHdrs := createShardProcessHeadersToSaveLastNoterized(10, &block.Header{}, mock.HasherMock{}, &mock.MarshalizerMock{})

	err := base.SaveLastNotarizedHeader(shardId, prHdrs)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestBaseProcessor_SaveLastNoterizedHdrLastNotWrongTypeMeta(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	base.SetHasher(mock.HasherMock{})
	base.SetMarshalizer(&mock.MarshalizerMock{})

	// make it wrong
	genesisBlock := createGenesisBlocks(shardCoordinator)
	genesisBlock[sharding.MetachainShardId] = &block.Header{Nonce: 0}

	_ = base.SetLastNotarizedHeadersSlice(genesisBlock, true)
	prHdrs := createMetaProcessHeadersToSaveLastNoterized(10, &block.Header{}, mock.HasherMock{}, &mock.MarshalizerMock{})

	err := base.SaveLastNotarizedHeader(sharding.MetachainShardId, prHdrs)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestBaseProcessor_SaveLastNoterizedHdrShardWrongProcessed(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	base.SetHasher(mock.HasherMock{})
	base.SetMarshalizer(&mock.MarshalizerMock{})
	_ = base.SetLastNotarizedHeadersSlice(createGenesisBlocks(shardCoordinator), true)
	highestNonce := uint64(10)
	prHdrs := createMetaProcessHeadersToSaveLastNoterized(highestNonce, &block.Header{}, mock.HasherMock{}, &mock.MarshalizerMock{})

	shardId := uint32(0)
	err := base.SaveLastNotarizedHeader(shardId, prHdrs)
	assert.Nil(t, err)

	lastNodesHdrs := base.LastNotarizedHdrs()
	assert.Equal(t, uint64(0), lastNodesHdrs[shardId].GetNonce())
}

func TestBaseProcessor_SaveLastNoterizedHdrMetaWrongProcessed(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	base.SetHasher(mock.HasherMock{})
	base.SetMarshalizer(&mock.MarshalizerMock{})
	_ = base.SetLastNotarizedHeadersSlice(createGenesisBlocks(shardCoordinator), true)
	highestNonce := uint64(10)
	prHdrs := createShardProcessHeadersToSaveLastNoterized(highestNonce, &block.Header{}, mock.HasherMock{}, &mock.MarshalizerMock{})

	err := base.SaveLastNotarizedHeader(sharding.MetachainShardId, prHdrs)
	assert.Nil(t, err)

	lastNodesHdrs := base.LastNotarizedHdrs()
	assert.Equal(t, uint64(0), lastNodesHdrs[sharding.MetachainShardId].GetNonce())
}

func TestBaseProcessor_SaveLastNoterizedHdrShardGood(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	hasher := mock.HasherMock{}
	base.SetHasher(hasher)
	marshalizer := &mock.MarshalizerMock{}
	base.SetMarshalizer(marshalizer)
	genesisBlcks := createGenesisBlocks(shardCoordinator)
	_ = base.SetLastNotarizedHeadersSlice(genesisBlcks, true)

	highestNonce := uint64(10)
	shardId := uint32(0)
	prHdrs := createShardProcessHeadersToSaveLastNoterized(highestNonce, genesisBlcks[shardId], hasher, marshalizer)

	err := base.SaveLastNotarizedHeader(shardId, prHdrs)
	assert.Nil(t, err)

	lastNodesHdrs := base.LastNotarizedHdrs()
	assert.Equal(t, highestNonce, lastNodesHdrs[shardId].GetNonce())
}

func TestBaseProcessor_SaveLastNoterizedHdrMetaGood(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	hasher := mock.HasherMock{}
	base.SetHasher(hasher)
	marshalizer := &mock.MarshalizerMock{}
	base.SetMarshalizer(marshalizer)
	genesisBlcks := createGenesisBlocks(shardCoordinator)
	_ = base.SetLastNotarizedHeadersSlice(genesisBlcks, true)

	highestNonce := uint64(10)
	prHdrs := createMetaProcessHeadersToSaveLastNoterized(highestNonce, genesisBlcks[sharding.MetachainShardId], hasher, marshalizer)

	err := base.SaveLastNotarizedHeader(sharding.MetachainShardId, prHdrs)
	assert.Nil(t, err)

	lastNodesHdrs := base.LastNotarizedHdrs()
	assert.Equal(t, highestNonce, lastNodesHdrs[sharding.MetachainShardId].GetNonce())
}
