package process

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestCache() storage.Cacher {
	cache, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 1000, Shards: 1, SizeInBytes: 0})
	return cache
}

func generateTestUnit() storage.Storer {
	storer, _ := storageUnit.NewStorageUnit(
		generateTestCache(),
		memorydb.New(),
	)

	return storer
}

func initStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.RewardTransactionUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, generateTestUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ReceiptsUnit, generateTestUnit())
	return store
}

func createMockArgsNewShardBlockCreatorAfterHardFork() ArgsNewShardBlockCreatorAfterHardFork {
	return ArgsNewShardBlockCreatorAfterHardFork{
		Hasher:             &mock.HasherMock{},
		ImportHandler:      &mock.ImportHandlerStub{},
		Marshalizer:        &mock.MarshalizerMock{},
		PendingTxProcessor: &mock.PendingTransactionProcessorStub{},
		ShardCoordinator:   mock.NewOneShardCoordinatorMock(),
		Storage:            initStore(),
		TxCoordinator:      &mock.TransactionCoordinatorMock{},
	}
}

func TestNewShardBlockCreatorAfterHardFork(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewShardBlockCreatorAfterHardFork()

	shardBlockCreator, err := NewShardBlockCreatorAfterHardFork(args)
	assert.NoError(t, err)
	assert.False(t, check.IfNil(shardBlockCreator))
}

func TestCreateBody(t *testing.T) {
	t.Parallel()

	mb1 := &block.MiniBlock{SenderShardID: 1}
	mb2, mb3 := &block.MiniBlock{SenderShardID: 2}, &block.MiniBlock{SenderShardID: 3}

	args := createMockArgsNewShardBlockCreatorAfterHardFork()
	args.PendingTxProcessor = &mock.PendingTransactionProcessorStub{
		ProcessTransactionsDstMeCalled: func(mbInfo *update.MbInfo) (*block.MiniBlock, error) {
			return mb1, nil
		},
	}
	args.TxCoordinator = &mock.TransactionCoordinatorMock{
		CreatePostProcessMiniBlocksCalled: func() block.MiniBlockSlice {
			return block.MiniBlockSlice{mb2, mb3}
		},
	}

	metaBlock := &block.MetaBlock{
		Round: 2,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					FirstPendingMetaBlock: []byte("metaBlock_hash"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash: []byte("miniBlock_hash"),
						},
					},
				},
			},
		},
	}

	unFinishedMetaBlocks := map[string]data.HeaderHandler{
		"metaBlock_hash": &block.MetaBlock{Round: 1},
	}
	args.ImportHandler = &mock.ImportHandlerStub{
		GetHardForkMetaBlockCalled: func() data.HeaderHandler {
			return metaBlock
		},
		GetUnFinishedMetaBlocksCalled: func() map[string]data.HeaderHandler {
			return unFinishedMetaBlocks
		},
		GetMiniBlocksCalled: func() map[string]*block.MiniBlock {
			return map[string]*block.MiniBlock{
				"miniBlock_hash": {},
			}
		},
	}

	shardBlockCreator, _ := NewShardBlockCreatorAfterHardFork(args)

	expectedBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{mb1, mb2, mb3},
	}
	body, _, err := shardBlockCreator.CreateBody()
	assert.NoError(t, err)
	assert.Equal(t, expectedBody, body)
}

func TestCreateMiniBlockHeader(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewShardBlockCreatorAfterHardFork()
	shardBlockCreator, _ := NewShardBlockCreatorAfterHardFork(args)

	hashTx1, hashTx2 := []byte("hash1"), []byte("hash2")
	mb1 := &block.MiniBlock{SenderShardID: 1, ReceiverShardID: 2, TxHashes: [][]byte{hashTx1}}
	mb2 := &block.MiniBlock{SenderShardID: 3, ReceiverShardID: 4, TxHashes: [][]byte{hashTx2}}
	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{mb1, mb2},
	}

	mb1Hash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, mb1)
	mb2Hash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, mb2)

	expectedMbHeaders := []block.MiniBlockHeader{
		{
			Hash:            mb1Hash,
			TxCount:         1,
			Type:            0,
			SenderShardID:   1,
			ReceiverShardID: 2,
		},
		{
			Hash:            mb2Hash,
			TxCount:         1,
			Type:            0,
			SenderShardID:   3,
			ReceiverShardID: 4,
		},
	}
	totalTxs, mbHeaders, err := shardBlockCreator.createMiniBlockHeaders(body)
	assert.NoError(t, err)
	assert.Equal(t, 2, totalTxs)
	assert.Equal(t, expectedMbHeaders, mbHeaders)
}

func TestCreateBlock(t *testing.T) {
	t.Parallel()

	rootHash := []byte("rotHash")
	hashTx1, hashTx2 := []byte("hash1"), []byte("hash2")
	mb1 := &block.MiniBlock{SenderShardID: 1, ReceiverShardID: 2, TxHashes: [][]byte{hashTx1}}
	mb2 := &block.MiniBlock{SenderShardID: 3, ReceiverShardID: 4, TxHashes: [][]byte{hashTx2}}
	args := createMockArgsNewShardBlockCreatorAfterHardFork()
	args.PendingTxProcessor = &mock.PendingTransactionProcessorStub{
		ProcessTransactionsDstMeCalled: func(mbInfo *update.MbInfo) (*block.MiniBlock, error) {
			return mb1, nil
		},
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
	}
	args.TxCoordinator = &mock.TransactionCoordinatorMock{
		CreatePostProcessMiniBlocksCalled: func() block.MiniBlockSlice {
			return block.MiniBlockSlice{mb2}
		},
		GetAllCurrentUsedTxsCalled: func(blockType block.Type) map[string]data.TransactionHandler {
			return map[string]data.TransactionHandler{
				string(hashTx2): &transaction.Transaction{},
			}
		},
	}

	metaBlock := &block.MetaBlock{
		Round: 2,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					FirstPendingMetaBlock: []byte("metaBlock_hash"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash: []byte("miniBlock_hash"),
						},
					},
				},
			},
		},
	}
	metaHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, metaBlock)
	unFinishedMetaBlocks := map[string]data.HeaderHandler{
		"metaBlock_hash": &block.MetaBlock{Round: 1},
	}
	args.ImportHandler = &mock.ImportHandlerStub{
		GetHardForkMetaBlockCalled: func() data.HeaderHandler {
			return metaBlock
		},
		GetUnFinishedMetaBlocksCalled: func() map[string]data.HeaderHandler {
			return unFinishedMetaBlocks
		},
		GetMiniBlocksCalled: func() map[string]*block.MiniBlock {
			return map[string]*block.MiniBlock{
				"miniBlock_hash": {},
			}
		},
	}

	shardBlockCreator, _ := NewShardBlockCreatorAfterHardFork(args)

	chainID, round, nonce, epoch := "chainId", uint64(100), uint64(90), uint32(2)

	expectedBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{mb1, mb2},
	}
	mb1Hash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, mb1)
	mb2Hash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, mb2)

	expectedMbHeaders := []block.MiniBlockHeader{
		{
			Hash:            mb1Hash,
			TxCount:         1,
			Type:            0,
			SenderShardID:   1,
			ReceiverShardID: 2,
		},
		{
			Hash:            mb2Hash,
			TxCount:         1,
			Type:            0,
			SenderShardID:   3,
			ReceiverShardID: 4,
		},
	}

	expectedHeader := &block.Header{
		ChainID: []byte(chainID), Round: round, Nonce: nonce, Epoch: epoch,
		RandSeed: rootHash, RootHash: rootHash, PrevHash: rootHash, PrevRandSeed: rootHash,
		ReceiptsHash: []byte("receiptHash"), TxCount: 2,
		MetaBlockHashes: [][]byte{metaHash}, MiniBlockHeaders: expectedMbHeaders,
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		PubKeysBitmap:   []byte{1},
		SoftwareVersion: []byte(""),
	}
	body, _, err := shardBlockCreator.CreateBody()
	assert.NoError(t, err)
	header, err := shardBlockCreator.CreateBlock(body, chainID, round, nonce, epoch)
	assert.NoError(t, err)
	assert.Equal(t, expectedBody, body)
	assert.Equal(t, expectedHeader, header)
}

func TestGetPendingMbsAndTxsInCorrectOrder_ShouldErrWhenGetPendingMiniBlocksFails(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewShardBlockCreatorAfterHardFork()
	shardBlockCreator, _ := NewShardBlockCreatorAfterHardFork(args)

	_, err := shardBlockCreator.getPendingMbsAndTxsInCorrectOrder()
	assert.Equal(t, update.ErrNilEpochStartMetaBlock, err)
}

func TestGetPendingMbsAndTxsInCorrectOrder_ShouldErrWrongImportedMiniBlocksMap(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewShardBlockCreatorAfterHardFork()

	metaBlock := &block.MetaBlock{
		Round: 2,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					FirstPendingMetaBlock: []byte("metaBlock_hash"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash: []byte("miniBlock_hash"),
						},
					},
				},
			},
		},
	}
	unFinishedMetaBlocks := map[string]data.HeaderHandler{
		"metaBlock_hash": &block.MetaBlock{Round: 1},
	}
	args.ImportHandler = &mock.ImportHandlerStub{
		GetHardForkMetaBlockCalled: func() data.HeaderHandler {
			return metaBlock
		},
		GetUnFinishedMetaBlocksCalled: func() map[string]data.HeaderHandler {
			return unFinishedMetaBlocks
		},
	}

	shardBlockCreator, _ := NewShardBlockCreatorAfterHardFork(args)

	_, err := shardBlockCreator.getPendingMbsAndTxsInCorrectOrder()
	assert.Equal(t, update.ErrWrongImportedMiniBlocksMap, err)
}

func TestGetPendingMbsAndTxsInCorrectOrder_ShouldErrMiniBlockNotFoundInImportedMap(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewShardBlockCreatorAfterHardFork()

	metaBlock := &block.MetaBlock{
		Round: 2,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					FirstPendingMetaBlock: []byte("metaBlock_hash"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash: []byte("miniBlock_hash1"),
						},
					},
				},
			},
		},
	}
	unFinishedMetaBlocks := map[string]data.HeaderHandler{
		"metaBlock_hash": &block.MetaBlock{Round: 1},
	}
	args.ImportHandler = &mock.ImportHandlerStub{
		GetHardForkMetaBlockCalled: func() data.HeaderHandler {
			return metaBlock
		},
		GetUnFinishedMetaBlocksCalled: func() map[string]data.HeaderHandler {
			return unFinishedMetaBlocks
		},
		GetMiniBlocksCalled: func() map[string]*block.MiniBlock {
			return map[string]*block.MiniBlock{
				"miniBlock_hash2": {},
			}
		},
	}

	shardBlockCreator, _ := NewShardBlockCreatorAfterHardFork(args)

	_, err := shardBlockCreator.getPendingMbsAndTxsInCorrectOrder()
	assert.Equal(t, update.ErrMiniBlockNotFoundInImportedMap, err)
}

func TestGetPendingMbsAndTxsInCorrectOrder_ShouldErrTransactionNotFoundInImportedMap(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewShardBlockCreatorAfterHardFork()

	metaBlock := &block.MetaBlock{
		Round: 2,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					FirstPendingMetaBlock: []byte("metaBlock_hash"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash: []byte("miniBlock_hash"),
						},
					},
				},
			},
		},
	}
	unFinishedMetaBlocks := map[string]data.HeaderHandler{
		"metaBlock_hash": &block.MetaBlock{Round: 1},
	}
	args.ImportHandler = &mock.ImportHandlerStub{
		GetHardForkMetaBlockCalled: func() data.HeaderHandler {
			return metaBlock
		},
		GetUnFinishedMetaBlocksCalled: func() map[string]data.HeaderHandler {
			return unFinishedMetaBlocks
		},
		GetMiniBlocksCalled: func() map[string]*block.MiniBlock {
			return map[string]*block.MiniBlock{
				"miniBlock_hash": {
					TxHashes: [][]byte{
						[]byte("tx_hash1"),
						[]byte("tx_hash2"),
					},
				},
			}
		},
	}

	shardBlockCreator, _ := NewShardBlockCreatorAfterHardFork(args)

	_, err := shardBlockCreator.getPendingMbsAndTxsInCorrectOrder()
	assert.Equal(t, update.ErrTransactionNotFoundInImportedMap, err)
}

func TestGetPendingMbsAndTxsInCorrectOrder_ShouldErrWrongImportedTransactionsMap(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewShardBlockCreatorAfterHardFork()

	metaBlock := &block.MetaBlock{
		Round: 2,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					FirstPendingMetaBlock: []byte("metaBlock_hash"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash: []byte("miniBlock_hash"),
						},
					},
				},
			},
		},
	}
	unFinishedMetaBlocks := map[string]data.HeaderHandler{
		"metaBlock_hash": &block.MetaBlock{Round: 1},
	}
	args.ImportHandler = &mock.ImportHandlerStub{
		GetHardForkMetaBlockCalled: func() data.HeaderHandler {
			return metaBlock
		},
		GetUnFinishedMetaBlocksCalled: func() map[string]data.HeaderHandler {
			return unFinishedMetaBlocks
		},
		GetMiniBlocksCalled: func() map[string]*block.MiniBlock {
			return map[string]*block.MiniBlock{
				"miniBlock_hash": {
					TxHashes: [][]byte{
						[]byte("tx_hash1"),
						[]byte("tx_hash2"),
					},
				},
			}
		},
		GetTransactionsCalled: func() map[string]data.TransactionHandler {
			return map[string]data.TransactionHandler{
				"tx_hash1": &transaction.Transaction{Nonce: 0},
				"tx_hash2": &transaction.Transaction{Nonce: 1},
				"tx_hash3": &transaction.Transaction{Nonce: 2},
			}
		},
	}

	shardBlockCreator, _ := NewShardBlockCreatorAfterHardFork(args)

	_, err := shardBlockCreator.getPendingMbsAndTxsInCorrectOrder()
	assert.Equal(t, update.ErrWrongImportedTransactionsMap, err)
}

func TestGetPendingMbsAndTxsInCorrectOrder_ShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewShardBlockCreatorAfterHardFork()

	metaBlock := &block.MetaBlock{
		Round: 2,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					FirstPendingMetaBlock: []byte("metaBlock_hash"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash: []byte("miniBlock_hash"),
						},
					},
				},
			},
		},
	}
	unFinishedMetaBlocks := map[string]data.HeaderHandler{
		"metaBlock_hash": &block.MetaBlock{Round: 1},
	}
	args.ImportHandler = &mock.ImportHandlerStub{
		GetHardForkMetaBlockCalled: func() data.HeaderHandler {
			return metaBlock
		},
		GetUnFinishedMetaBlocksCalled: func() map[string]data.HeaderHandler {
			return unFinishedMetaBlocks
		},
		GetMiniBlocksCalled: func() map[string]*block.MiniBlock {
			return map[string]*block.MiniBlock{
				"miniBlock_hash": {
					TxHashes: [][]byte{
						[]byte("tx_hash1"),
						[]byte("tx_hash2"),
					},
				},
			}
		},
		GetTransactionsCalled: func() map[string]data.TransactionHandler {
			return map[string]data.TransactionHandler{
				"tx_hash1": &transaction.Transaction{Nonce: 0},
				"tx_hash2": &transaction.Transaction{Nonce: 1},
			}
		},
	}

	shardBlockCreator, _ := NewShardBlockCreatorAfterHardFork(args)

	mbsInfo, err := shardBlockCreator.getPendingMbsAndTxsInCorrectOrder()
	assert.Nil(t, err)
	require.Equal(t, 1, len(mbsInfo))
	require.Equal(t, 2, len(mbsInfo[0].TxsInfo))
	assert.Equal(t, []byte("miniBlock_hash"), mbsInfo[0].MbHash)
	assert.Equal(t, []byte("tx_hash1"), mbsInfo[0].TxsInfo[0].TxHash)
	assert.Equal(t, []byte("tx_hash2"), mbsInfo[0].TxsInfo[1].TxHash)
}

func TestCreateMiniBlockInfoForPostProcessMiniBlock_ShouldErrPostProcessTransactionNotFound(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewShardBlockCreatorAfterHardFork()
	shardBlockCreator, _ := NewShardBlockCreatorAfterHardFork(args)

	mb := &block.MiniBlock{
		TxHashes: [][]byte{
			[]byte("tx_hash"),
		},
	}
	mbHash := []byte("mb_hash")
	_, err := shardBlockCreator.createMiniBlockInfoForPostProcessMiniBlock(mbHash, mb)
	assert.Equal(t, update.ErrPostProcessTransactionNotFound, err)
}

func TestCreateMiniBlockInfoForPostProcessMiniBlock_ShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgsNewShardBlockCreatorAfterHardFork()
	txHash := []byte("tx_hash")
	tx := &transaction.Transaction{Nonce: 1}
	args.TxCoordinator = &mock.TransactionCoordinatorMock{
		GetAllCurrentUsedTxsCalled: func(blockType block.Type) map[string]data.TransactionHandler {
			return map[string]data.TransactionHandler{
				string(txHash): tx,
			}
		},
	}
	shardBlockCreator, _ := NewShardBlockCreatorAfterHardFork(args)

	txHashes := [][]byte{txHash}
	mb := &block.MiniBlock{
		SenderShardID:   0,
		ReceiverShardID: 1,
		Type:            block.SmartContractResultBlock,
		TxHashes:        txHashes,
	}
	mbHash := []byte("mb_hash")
	mbInfo, err := shardBlockCreator.createMiniBlockInfoForPostProcessMiniBlock(mbHash, mb)
	assert.Nil(t, err)
	assert.Equal(t, mbHash, mbInfo.MbHash)
	require.Equal(t, 1, len(mbInfo.TxsInfo))
	assert.Equal(t, txHash, mbInfo.TxsInfo[0].TxHash)
	assert.Equal(t, tx, mbInfo.TxsInfo[0].Tx)
}
