package coordinator

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

func initDataPool(testHash []byte) *mock.PoolsHolderStub {
	sdp := &mock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{
				RegisterHandlerCalled: func(i func(key []byte)) {},
				ShardDataStoreCalled: func(id string) (c storage.Cacher) {
					return &mock.CacherStub{
						PeekCalled: func(key []byte) (value interface{}, ok bool) {
							if reflect.DeepEqual(key, testHash) {
								return &transaction.Transaction{Nonce: 10}, true
							}
							return nil, false
						},
						KeysCalled: func() [][]byte {
							return [][]byte{[]byte("key1"), []byte("key2"), testHash}
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
				AddDataCalled: func(key []byte, data interface{}, cacheId string) {
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
		MetaHeadersNoncesCalled: func() dataRetriever.Uint64Cacher {
			cs := &mock.Uint64CacherStub{}
			cs.HasCalled = func(u uint64) bool {
				return true
			}
			return cs
		},
	}
	return sdp
}

func TestNewTransactionCoordinator_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		nil,
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewTransactionCoordinator_NilAccountsStub(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		nil,
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewTransactionCoordinator_NilDataPool(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		nil,
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewTransactionCoordinator_NilRequestHandler(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		nil,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilRequestHandler, err)
}

func TestNewTransactionCoordinator_NilHasher(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		nil,
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewTransactionCoordinator_NilMarshalizer(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		nil,
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewTransactionCoordinator_NilTxProc(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		nil,
		&mock.ChainStorerMock{},
	)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilTxProcessor, err)
}

func TestNewTransactionCoordinator_NilStorer(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		nil,
	)

	assert.Nil(t, tc)
	assert.Equal(t, process.ErrNilTxStorage, err)
}

func TestNewTransactionCoordinator_OK(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, tc)
}

func TestTransactionCoordinator_SeparateBodyNil(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	separated := tc.separateBodyByType(nil)
	assert.Equal(t, 0, len(separated))
}

func TestTransactionCoordinator_SeparateBody(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	body := block.Body{}
	body = append(body, &block.MiniBlock{Type: block.TxBlock})
	body = append(body, &block.MiniBlock{Type: block.TxBlock})
	body = append(body, &block.MiniBlock{Type: block.TxBlock})
	body = append(body, &block.MiniBlock{Type: block.SmartContractResultBlock})
	body = append(body, &block.MiniBlock{Type: block.SmartContractResultBlock})
	body = append(body, &block.MiniBlock{Type: block.SmartContractResultBlock})
	body = append(body, &block.MiniBlock{Type: block.SmartContractResultBlock})

	separated := tc.separateBodyByType(body)
	assert.Equal(t, 2, len(separated))
	assert.Equal(t, 3, len(separated[block.TxBlock]))
	assert.Equal(t, 4, len(separated[block.SmartContractResultBlock]))
}

func TestTransactionCoordinator_CreateBlockStarted(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	tc.CreateBlockStarted()

	tc.mutPreprocessor.Lock()
	for _, value := range tc.txPreprocessors {
		txs := value.GetAllCurrentUsedTxs()
		assert.Equal(t, 0, len(txs))
	}
	tc.mutPreprocessor.Unlock()
}

func TestTransactionCoordinator_CreateMarshalizedDataNilBody(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	mrBody, mrTxs := tc.CreateMarshalizedData(nil)
	assert.Equal(t, 0, len(mrTxs))
	assert.Equal(t, 0, len(mrBody))
}

func createMiniBlockWithOneTx(sndId, dstId uint32, blockType block.Type, txHash []byte) *block.MiniBlock {
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)

	return &block.MiniBlock{Type: blockType, SenderShardID: sndId, ReceiverShardID: dstId, TxHashes: txHashes}
}

func createTestBody() block.Body {
	body := block.Body{}
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, []byte("tx_hash1"))
	body = append(body, createMiniBlockWithOneTx(0, 1, block.TxBlock, []byte("tx_hash1")))
	body = append(body, createMiniBlockWithOneTx(0, 1, block.TxBlock, []byte("tx_hash2")))
	body = append(body, createMiniBlockWithOneTx(0, 1, block.TxBlock, []byte("tx_hash3")))
	body = append(body, createMiniBlockWithOneTx(0, 1, block.SmartContractResultBlock, []byte("tx_hash1")))
	body = append(body, createMiniBlockWithOneTx(0, 1, block.SmartContractResultBlock, []byte("tx_hash2")))
	body = append(body, createMiniBlockWithOneTx(0, 1, block.SmartContractResultBlock, []byte("tx_hash3")))

	return body
}

func TestTransactionCoordinator_CreateMarshalizedData(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	mrBody, mrTxs := tc.CreateMarshalizedData(createTestBody())
	assert.Equal(t, 0, len(mrTxs))
	assert.Equal(t, 1, len(mrBody))
}

func TestTransactionCoordinator_CreateMbsAndProcessCrossShardTransactionsDstMeNilHeader(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	maxTxRemaining := uint32(15000)
	haveTime := func() bool {
		return true
	}
	mbs, txs, finalized := tc.CreateMbsAndProcessCrossShardTransactionsDstMe(nil, maxTxRemaining, 10, haveTime)

	assert.Equal(t, 0, len(mbs))
	assert.Equal(t, uint32(0), txs)
	assert.True(t, finalized)
}

func createTestMetablock() *block.MetaBlock {
	meta := &block.MetaBlock{}

	meta.ShardInfo = make([]block.ShardData, 0)

	shardMbs := make([]block.ShardMiniBlockHeader, 0)
	shardMbs = append(shardMbs, block.ShardMiniBlockHeader{Hash: []byte("mb0"), SenderShardId: 0, ReceiverShardId: 0, TxCount: 1})
	shardMbs = append(shardMbs, block.ShardMiniBlockHeader{Hash: []byte("mb1"), SenderShardId: 0, ReceiverShardId: 1, TxCount: 1})
	shardData := block.ShardData{ShardId: 0, HeaderHash: []byte("header0"), TxCount: 2, ShardMiniBlockHeaders: shardMbs}

	meta.ShardInfo = append(meta.ShardInfo, shardData)

	shardMbs = make([]block.ShardMiniBlockHeader, 0)
	shardMbs = append(shardMbs, block.ShardMiniBlockHeader{Hash: []byte("mb2"), SenderShardId: 1, ReceiverShardId: 0, TxCount: 1})
	shardMbs = append(shardMbs, block.ShardMiniBlockHeader{Hash: []byte("mb3"), SenderShardId: 1, ReceiverShardId: 1, TxCount: 1})
	shardData = block.ShardData{ShardId: 1, HeaderHash: []byte("header0"), TxCount: 2, ShardMiniBlockHeaders: shardMbs}

	meta.ShardInfo = append(meta.ShardInfo, shardData)

	return meta
}

func TestTransactionCoordinator_CreateMbsAndProcessCrossShardTransactionsDstMeNoTime(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	maxTxRemaining := uint32(15000)
	haveTime := func() bool {
		return false
	}
	mbs, txs, finalized := tc.CreateMbsAndProcessCrossShardTransactionsDstMe(createTestMetablock(), maxTxRemaining, 10, haveTime)

	assert.Equal(t, 0, len(mbs))
	assert.Equal(t, uint32(0), txs)
	assert.False(t, finalized)
}

func TestTransactionCoordinator_CreateMbsAndProcessCrossShardTransactionsNothingInPool(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	maxTxRemaining := uint32(15000)
	haveTime := func() bool {
		return true
	}
	mbs, txs, finalized := tc.CreateMbsAndProcessCrossShardTransactionsDstMe(createTestMetablock(), maxTxRemaining, 10, haveTime)

	assert.Equal(t, 0, len(mbs))
	assert.Equal(t, uint32(0), txs)
	assert.False(t, finalized)
}

func TestTransactionCoordinator_CreateMbsAndProcessTransactionsFromMeNothingToProcess(t *testing.T) {
	t.Parallel()

	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		mock.NewPoolsHolderFake(),
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	maxTxRemaining := uint32(15000)
	haveTime := func() bool {
		return true
	}
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(maxTxRemaining, 10, haveTime)

	assert.Equal(t, 0, len(mbs))
}

func TestTransactionCoordinator_CreateMbsAndProcessTransactionsFromMeNoTime(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		tdp,
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	maxTxRemaining := uint32(15000)
	haveTime := func() bool {
		return false
	}
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(maxTxRemaining, 10, haveTime)

	assert.Equal(t, 0, len(mbs))
}

func TestTransactionCoordinator_CreateMbsAndProcessTransactionsFromMeNoSpace(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AccountsStub{},
		tdp,
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.ChainStorerMock{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	maxTxRemaining := uint32(0)
	haveTime := func() bool {
		return true
	}
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(maxTxRemaining, 10, haveTime)

	assert.Equal(t, 0, len(mbs))
}

func TestTransactionCoordinator_CreateMbsAndProcessTransactionsFromMe(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	nrShards := uint32(5)
	tc, err := NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(nrShards),
		&mock.AccountsStub{},
		tdp,
		&mock.RequestHandlerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction, round uint32) error {
				return nil
			},
		},
		&mock.ChainStorerMock{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, tc)

	maxTxRemaining := uint32(15000)
	haveTime := func() bool {
		return true
	}
	mbs := tc.CreateMbsAndProcessTransactionsFromMe(maxTxRemaining, 10, haveTime)

	assert.Equal(t, int(nrShards), len(mbs))
}

/*
//------- processMiniBlockComplete

func TestShardProcessor_ProcessMiniBlockCompleteWithOkTxsShouldExecuteThemAndNotRevertAccntState(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewPoolsHolderFake()

	//we will have a miniblock that will have 3 tx hashes
	//all txs will be in datapool and none of them will return err when processed
	//so, tx processor will return nil on processing tx

	txHash1 := []byte("tx hash 1")
	txHash2 := []byte("tx hash 2")
	txHash3 := []byte("tx hash 3")

	senderShardId := uint32(1)
	receiverShardId := uint32(2)

	miniBlock := block.MiniBlock{
		SenderShardID:   senderShardId,
		ReceiverShardID: receiverShardId,
		TxHashes:        [][]byte{txHash1, txHash2, txHash3},
	}

	tx1Nonce := uint64(45)
	tx2Nonce := uint64(46)
	tx3Nonce := uint64(47)

	//put the existing tx inside datapool
	cacheId := process.ShardCacherIdentifier(senderShardId, receiverShardId)
	dataPool.Transactions().AddData(txHash1, &transaction.Transaction{
		Nonce: tx1Nonce,
		Data:  txHash1,
	}, cacheId)
	dataPool.Transactions().AddData(txHash2, &transaction.Transaction{
		Nonce: tx2Nonce,
		Data:  txHash2,
	}, cacheId)
	dataPool.Transactions().AddData(txHash3, &transaction.Transaction{
		Nonce: tx3Nonce,
		Data:  txHash3,
	}, cacheId)

	tx1ExecutionResult := uint64(0)
	tx2ExecutionResult := uint64(0)
	tx3ExecutionResult := uint64(0)

	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		&mock.AccountsStub{
			RevertToSnapshotCalled: func(snapshot int) error {
				assert.Fail(t, "revert should have not been called")
				return nil
			},
			JournalLenCalled: func() int {
				return 0
			},
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	err := bp.ProcessMiniBlockComplete(&miniBlock, 0, func() bool {
		return true
	})

	assert.Nil(t, err)
	assert.Equal(t, tx1Nonce, tx1ExecutionResult)
	assert.Equal(t, tx2Nonce, tx2ExecutionResult)
	assert.Equal(t, tx3Nonce, tx3ExecutionResult)
}

func TestShardProcessor_ProcessMiniBlockCompleteWithErrorWhileProcessShouldCallRevertAccntState(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewPoolsHolderFake()

	//we will have a miniblock that will have 3 tx hashes
	//all txs will be in datapool and none of them will return err when processed
	//so, tx processor will return nil on processing tx

	txHash1 := []byte("tx hash 1")
	txHash2 := []byte("tx hash 2 - this will cause the tx processor to err")
	txHash3 := []byte("tx hash 3")

	senderShardId := uint32(1)
	receiverShardId := uint32(2)

	miniBlock := block.MiniBlock{
		SenderShardID:   senderShardId,
		ReceiverShardID: receiverShardId,
		TxHashes:        [][]byte{txHash1, txHash2, txHash3},
	}

	tx1Nonce := uint64(45)
	tx2Nonce := uint64(46)
	tx3Nonce := uint64(47)

	errTxProcessor := errors.New("tx processor failing")

	//put the existing tx inside datapool
	cacheId := process.ShardCacherIdentifier(senderShardId, receiverShardId)
	dataPool.Transactions().AddData(txHash1, &transaction.Transaction{
		Nonce: tx1Nonce,
		Data:  txHash1,
	}, cacheId)
	dataPool.Transactions().AddData(txHash2, &transaction.Transaction{
		Nonce: tx2Nonce,
		Data:  txHash2,
	}, cacheId)
	dataPool.Transactions().AddData(txHash3, &transaction.Transaction{
		Nonce: tx3Nonce,
		Data:  txHash3,
	}, cacheId)

	currentJournalLen := 445
	revertAccntStateCalled := false

	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction, round uint32) error {
				if bytes.Equal(transaction.Data, txHash2) {
					return errTxProcessor
				}

				return nil
			},
		},
		&mock.AccountsStub{
			RevertToSnapshotCalled: func(snapshot int) error {
				if snapshot == currentJournalLen {
					revertAccntStateCalled = true
				}

				return nil
			},
			JournalLenCalled: func() int {
				return currentJournalLen
			},
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
	)

	err := bp.ProcessMiniBlockComplete(&miniBlock, 0, func() bool { return true })

	assert.Equal(t, errTxProcessor, err)
	assert.True(t, revertAccntStateCalled)
}

//------- receivedMiniBlock

func TestShardProcessor_ReceivedMiniBlockShouldRequestMissingTransactions(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewPoolsHolderFake()

	//we will have a miniblock that will have 3 tx hashes
	//1 tx hash will be in cache
	//2 will be requested on network

	txHash1 := []byte("tx hash 1 found in cache")
	txHash2 := []byte("tx hash 2")
	txHash3 := []byte("tx hash 3")

	senderShardId := uint32(1)
	receiverShardId := uint32(2)

	miniBlock := block.MiniBlock{
		SenderShardID:   senderShardId,
		ReceiverShardID: receiverShardId,
		TxHashes:        [][]byte{txHash1, txHash2, txHash3},
	}

	//put this miniblock inside datapool
	miniBlockHash := []byte("miniblock hash")
	dataPool.MiniBlocks().Put(miniBlockHash, miniBlock)

	//put the existing tx inside datapool
	cacheId := process.ShardCacherIdentifier(senderShardId, receiverShardId)
	dataPool.Transactions().AddData(txHash1, &transaction.Transaction{}, cacheId)

	txHash1Requested := int32(0)
	txHash2Requested := int32(0)
	txHash3Requested := int32(0)

	requestHandler := &mock.RequestHandlerMock{}
	requestHandler.RequestTransactionHandlerCalled = func(destShardID uint32, txHashes [][]byte) {
		if containsHash(txHashes, txHash1) {
			atomic.AddInt32(&txHash1Requested, 1)
		}
		if containsHash(txHashes, txHash2) {
			atomic.AddInt32(&txHash2Requested, 1)
		}
		if containsHash(txHashes, txHash3) {
			atomic.AddInt32(&txHash3Requested, 1)
		}
	}

	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		requestHandler,
	)

	bp.ReceivedMiniBlock(miniBlockHash)

	//we have to wait to be sure txHash1Requested is not incremented by a late call
	time.Sleep(time.Second)

	assert.Equal(t, int32(0), atomic.LoadInt32(&txHash1Requested))
	assert.Equal(t, int32(1), atomic.LoadInt32(&txHash2Requested))
	assert.Equal(t, int32(1), atomic.LoadInt32(&txHash2Requested))
}
*/
