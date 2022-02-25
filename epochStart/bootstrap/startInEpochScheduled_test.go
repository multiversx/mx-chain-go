package bootstrap

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/scheduled"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	epochStartMocks "github.com/ElrondNetwork/elrond-go/testscommon/bootstrapMocks/epochStart"
	"github.com/ElrondNetwork/elrond-go/testscommon/syncer"
	"github.com/stretchr/testify/require"
)

func Test_newStartInEpochShardHeaderDataSyncerWithScheduledNilScheduledTxHandler(t *testing.T) {
	t.Parallel()

	args := createDefaultDataSyncerFactoryArgs()
	ds, err := newStartInEpochShardHeaderDataSyncerWithScheduled(nil, args.HeadersSyncer, args.MiniBlocksSyncer, args.TxSyncer, args.ScheduledEnableEpoch)
	require.Nil(t, ds)
	require.Equal(t, epochStart.ErrNilScheduledTxsHandler, err)
}

func Test_newStartInEpochShardHeaderDataSyncerWithScheduledNilHeadersSyncer(t *testing.T) {
	t.Parallel()

	args := createDefaultDataSyncerFactoryArgs()
	ds, err := newStartInEpochShardHeaderDataSyncerWithScheduled(args.ScheduledTxsHandler, nil, args.MiniBlocksSyncer, args.TxSyncer, args.ScheduledEnableEpoch)
	require.Nil(t, ds)
	require.Equal(t, epochStart.ErrNilHeadersSyncer, err)
}

func Test_newStartInEpochShardHeaderDataSyncerWithScheduledNilMiniBlocksSyncer(t *testing.T) {
	t.Parallel()

	args := createDefaultDataSyncerFactoryArgs()
	ds, err := newStartInEpochShardHeaderDataSyncerWithScheduled(args.ScheduledTxsHandler, args.HeadersSyncer, nil, args.TxSyncer, args.ScheduledEnableEpoch)
	require.Nil(t, ds)
	require.Equal(t, epochStart.ErrNilMiniBlocksSyncer, err)
}

func Test_newStartInEpochShardHeaderDataSyncerWithScheduledNilTxSyncer(t *testing.T) {
	t.Parallel()

	args := createDefaultDataSyncerFactoryArgs()
	ds, err := newStartInEpochShardHeaderDataSyncerWithScheduled(args.ScheduledTxsHandler, args.HeadersSyncer, args.MiniBlocksSyncer, nil, args.ScheduledEnableEpoch)
	require.Nil(t, ds)
	require.Equal(t, epochStart.ErrNilTransactionsSyncer, err)
}

func Test_newStartInEpochShardHeaderDataSyncerWithScheduled(t *testing.T) {
	t.Parallel()

	args := createDefaultDataSyncerFactoryArgs()
	ds, err := newStartInEpochShardHeaderDataSyncerWithScheduled(args.ScheduledTxsHandler, args.HeadersSyncer, args.MiniBlocksSyncer, args.TxSyncer, args.ScheduledEnableEpoch)
	require.Nil(t, err)
	require.NotNil(t, ds)
}

func TestStartInEpochWithScheduledDataSyncer_UpdateSyncDataIfNeededScheduledNotEnabled(t *testing.T) {
	t.Parallel()

	args := createDefaultDataSyncerFactoryArgs()
	ds, _ := newStartInEpochShardHeaderDataSyncerWithScheduled(args.ScheduledTxsHandler, args.HeadersSyncer, args.MiniBlocksSyncer, args.TxSyncer, 10)

	notarizedShardHeader := createTestHeader()
	notarizedShardHeader.Epoch = 2

	header, headersMap, err := ds.UpdateSyncDataIfNeeded(notarizedShardHeader)
	require.Nil(t, err)
	require.Nil(t, headersMap)
	require.Equal(t, notarizedShardHeader, header)
}

func TestStartInEpochWithScheduledDataSyncer_UpdateSyncDataIfNeededGetRequiredHeadersWithError(t *testing.T) {
	t.Parallel()

	args := createDefaultDataSyncerFactoryArgs()
	expectedErr := fmt.Errorf("expected error")
	args.HeadersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
		SyncMissingHeadersByHashCalled: func(shardIDs []uint32, headersHashes [][]byte, ctx context.Context) error {
			return expectedErr
		},
	}
	ds, _ := newStartInEpochShardHeaderDataSyncerWithScheduled(args.ScheduledTxsHandler, args.HeadersSyncer, args.MiniBlocksSyncer, args.TxSyncer, 10)

	notarizedShardHeader := createTestHeader()
	notarizedShardHeader.Epoch = 2

	header, headersMap, err := ds.UpdateSyncDataIfNeeded(notarizedShardHeader)
	require.Nil(t, err)
	require.Nil(t, headersMap)
	require.Equal(t, notarizedShardHeader, header)
}

func TestStartInEpochWithScheduledDataSyncer_UpdateSyncDataIfNeededGetMiniBlocksWithError(t *testing.T) {
	t.Parallel()

	args := createDefaultDataSyncerFactoryArgs()
	expectedErr := fmt.Errorf("expected error")
	args.MiniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{
		SyncPendingMiniBlocksCalled: func(miniBlockHeaders []data.MiniBlockHeaderHandler, ctx context.Context) error {
			return expectedErr
		},
	}

	ds, _ := newStartInEpochShardHeaderDataSyncerWithScheduled(args.ScheduledTxsHandler, args.HeadersSyncer, args.MiniBlocksSyncer, args.TxSyncer, 10)

	notarizedShardHeader := createTestHeader()
	notarizedShardHeader.Epoch = 2

	header, headersMap, err := ds.UpdateSyncDataIfNeeded(notarizedShardHeader)
	require.Nil(t, err)
	require.Nil(t, headersMap)
	require.Equal(t, notarizedShardHeader, header)
}

func TestStartInEpochWithScheduledDataSyncer_UpdateSyncDataIfNeededScheduledEnabled(t *testing.T) {
	t.Parallel()

	args := createDefaultDataSyncerFactoryArgs()
	notarizedShardHeader := createTestHeader()
	notarizedShardHeader.Epoch = 2
	prevHeader := &block.Header{Nonce: 2}
	expectedHeadersMap := map[string]data.HeaderHandler{
		"hash1": notarizedShardHeader,
		string(notarizedShardHeader.GetPrevHash()): prevHeader,
	}
	args.HeadersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
		GetHeadersCalled: func() (map[string]data.HeaderHandler, error) {
			return expectedHeadersMap, nil
		},
	}

	ds, _ := newStartInEpochShardHeaderDataSyncerWithScheduled(args.ScheduledTxsHandler, args.HeadersSyncer, args.MiniBlocksSyncer, args.TxSyncer, 0)

	header, headersMap, err := ds.UpdateSyncDataIfNeeded(notarizedShardHeader)
	require.Nil(t, err)
	require.Equal(t, expectedHeadersMap, headersMap)
	require.Equal(t, prevHeader, header)
}

func TestStartInEpochWithScheduledDataSyncer_getRequiredHeaderByHash(t *testing.T) {
	// TODO: add test
}

func TestStartInEpochWithScheduledDataSyncer_syncHeadersShouldErrOnFailureToSync(t *testing.T) {
	t.Parallel()

	args := createDefaultDataSyncerFactoryArgs()
	expectedErr := fmt.Errorf("expected error")
	args.HeadersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
		SyncMissingHeadersByHashCalled: func(shardIDs []uint32, headersHashes [][]byte, ctx context.Context) error {
			return expectedErr
		},
	}
	ds, _ := newStartInEpochShardHeaderDataSyncerWithScheduled(args.ScheduledTxsHandler, args.HeadersSyncer, args.MiniBlocksSyncer, args.TxSyncer, 10)
	hashesToRequest := [][]byte{[]byte("hash1"), []byte("hash2")}
	shardIDs := []uint32{0, 1}

	mapHeaders, err := ds.syncHeaders(shardIDs, hashesToRequest)
	require.Equal(t, expectedErr, err)
	require.Nil(t, mapHeaders)
}

func TestStartInEpochWithScheduledDataSyncer_syncHeaders(t *testing.T) {
	t.Parallel()

	args := createDefaultDataSyncerFactoryArgs()
	expectedHeadersMap := map[string]data.HeaderHandler{
		"hash1": &block.Header{Nonce: 1},
		"hash2": &block.Header{Nonce: 2},
	}
	args.HeadersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
		GetHeadersCalled: func() (map[string]data.HeaderHandler, error) {
			return expectedHeadersMap, nil
		},
	}
	ds, _ := newStartInEpochShardHeaderDataSyncerWithScheduled(args.ScheduledTxsHandler, args.HeadersSyncer, args.MiniBlocksSyncer, args.TxSyncer, 10)
	hashesToRequest := [][]byte{[]byte("hash1"), []byte("hash2")}
	shardIDs := []uint32{0, 1}

	mapHeaders, err := ds.syncHeaders(shardIDs, hashesToRequest)
	require.Equal(t, expectedHeadersMap, mapHeaders)
	require.Nil(t, err)
}

func TestStartInEpochWithScheduledDataSyncer_getMiniBlocks(t *testing.T) {
	// TODO: add test
}

func TestStartInEpochWithScheduledDataSyncer_getRequiredMiniBlocksByMbHeaderWithSyncPendingMbsError(t *testing.T) {
	t.Parallel()

	args := createDefaultDataSyncerFactoryArgs()
	expectedErr := fmt.Errorf("expected error")
	args.MiniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{
		SyncPendingMiniBlocksCalled: func(miniBlockHeaders []data.MiniBlockHeaderHandler, ctx context.Context) error {
			return expectedErr
		},
	}

	ds, _ := newStartInEpochShardHeaderDataSyncerWithScheduled(args.ScheduledTxsHandler, args.HeadersSyncer, args.MiniBlocksSyncer, args.TxSyncer, 10)
	mbHeader := &block.MiniBlockHeader{
		Hash:            []byte("some miniblock"),
		SenderShardID:   0,
		ReceiverShardID: 0,
	}
	mbHeaderHandlers := []data.MiniBlockHeaderHandler{
		mbHeader,
	}

	mapMbs, err := ds.getRequiredMiniBlocksByMbHeader(mbHeaderHandlers)
	require.Equal(t, expectedErr, err)
	require.Nil(t, mapMbs)
}

func TestStartInEpochWithScheduledDataSyncer_getRequiredMiniBlocksByMbHeader(t *testing.T) {
	t.Parallel()

	args := createDefaultDataSyncerFactoryArgs()
	hashMb := []byte("some miniblock")
	mbHeader := &block.MiniBlockHeader{
		Hash:            hashMb,
		SenderShardID:   0,
		ReceiverShardID: 0,
	}
	mb := block.MiniBlock{
		TxHashes:        [][]byte{[]byte("txHash1"), []byte("txHash2")},
		ReceiverShardID: 0,
		SenderShardID:   0,
	}
	mbs := map[string]*block.MiniBlock{
		string(hashMb): &mb,
	}

	args.MiniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{
		SyncPendingMiniBlocksCalled: func(miniBlockHeaders []data.MiniBlockHeaderHandler, ctx context.Context) error {
			return nil
		},
		GetMiniBlocksCalled: func() (map[string]*block.MiniBlock, error) {
			return mbs, nil
		},
	}

	ds, _ := newStartInEpochShardHeaderDataSyncerWithScheduled(args.ScheduledTxsHandler, args.HeadersSyncer, args.MiniBlocksSyncer, args.TxSyncer, 10)
	mbHeaderHandlers := []data.MiniBlockHeaderHandler{
		mbHeader,
	}

	mapMbs, err := ds.getRequiredMiniBlocksByMbHeader(mbHeaderHandlers)
	require.Nil(t, err)
	require.Equal(t, mbs, mapMbs)
}

func TestStartInEpochWithScheduledDataSyncer_GetRootHashToSyncNoScheduled(t *testing.T) {
	t.Parallel()

	args := createDefaultDataSyncerFactoryArgs()
	ds, _ := newStartInEpochShardHeaderDataSyncerWithScheduled(args.ScheduledTxsHandler, args.HeadersSyncer, args.MiniBlocksSyncer, args.TxSyncer, 10)
	expectedRootHash := []byte("root hash")
	notarizedHeader := &block.Header{
		Nonce:    1,
		RootHash: expectedRootHash,
		Reserved: nil,
	}

	rootHash := ds.GetRootHashToSync(notarizedHeader)
	require.Equal(t, expectedRootHash, rootHash)
}

func TestStartInEpochWithScheduledDataSyncer_GetRootHashToSyncWithScheduled(t *testing.T) {
	t.Parallel()

	args := createDefaultDataSyncerFactoryArgs()
	ds, _ := newStartInEpochShardHeaderDataSyncerWithScheduled(args.ScheduledTxsHandler, args.HeadersSyncer, args.MiniBlocksSyncer, args.TxSyncer, 10)
	blockRootHash := []byte("root hash")
	expectedRootHash := []byte("scheduled root hash")
	notarizedHeader := &block.HeaderV2{
		Header: &block.Header{
			Nonce:    1,
			Epoch:    22,
			RootHash: blockRootHash,
			Reserved: nil,
		},
		ScheduledRootHash:        expectedRootHash,
		ScheduledAccumulatedFees: big.NewInt(0),
		ScheduledDeveloperFees:   big.NewInt(0),
	}

	rootHash := ds.GetRootHashToSync(notarizedHeader)
	require.Equal(t, expectedRootHash, rootHash)
}

func TestStartInEpochWithScheduledDataSyncer_prepareScheduledSCRs(t *testing.T) {
	// TODO: add test
}

func TestStartInEpochWithScheduledDataSyncer_filterScheduledIntermediateTxs(t *testing.T) {
	t.Parallel()

	sds := &startInEpochWithScheduledDataSyncer{}
	scheduledTxHashes := []string{
		"txHash1", "txHash2",
	}

	scheduledTxHashesMap := map[string]uint32{
		scheduledTxHashes[0]: 0,
		scheduledTxHashes[1]: 0,
	}

	scheduledSCR1Hash := "scheduled SCR 1"
	scheduledSCR2Hash := "scheduled SCR 2"
	allTxsMap := map[string]data.TransactionHandler{
		scheduledSCR1Hash: &smartContractResult.SmartContractResult{Nonce: 0, PrevTxHash: []byte(scheduledTxHashes[0])},
		scheduledSCR2Hash: &smartContractResult.SmartContractResult{Nonce: 1, PrevTxHash: []byte(scheduledTxHashes[1])},
		"regularTxHash1":  &transaction.Transaction{Nonce: 2},
		"regularTxHash2":  &smartContractResult.SmartContractResult{Nonce: 3},
	}
	expectedScheduledTxsMap := map[string]data.TransactionHandler{
		scheduledSCR1Hash: allTxsMap[scheduledSCR1Hash],
		scheduledSCR2Hash: allTxsMap[scheduledSCR2Hash],
	}

	miniBlocks := make(map[string]*block.MiniBlock)
	miniBlocks["1"] = &block.MiniBlock{
		Type:     block.SmartContractResultBlock,
		TxHashes: [][]byte{[]byte("scheduledSCR1Hash")},
	}
	miniBlocks["2"] = &block.MiniBlock{
		Type:     block.SmartContractResultBlock,
		TxHashes: [][]byte{[]byte("scheduledSCR2Hash")},
	}
	miniBlocks["3"] = &block.MiniBlock{
		Type:     block.TxBlock,
		TxHashes: [][]byte{[]byte("regularTxHash1")},
	}
	miniBlocks["1"] = &block.MiniBlock{
		Type:     block.SmartContractResultBlock,
		TxHashes: [][]byte{[]byte("regularTxHash2")},
	}

	scheduledTxs, err := sds.filterScheduledIntermediateTxs(miniBlocks, scheduledTxHashesMap, allTxsMap, 0)
	require.Nil(t, err)
	require.Equal(t, expectedScheduledTxsMap, scheduledTxs)
}

func TestStartInEpochWithScheduledDataSyncer_getScheduledIntermediateTxsMap(t *testing.T) {
	tx1 := &smartContractResult.SmartContractResult{Nonce: 0}
	tx2 := &smartContractResult.SmartContractResult{Nonce: 1}
	tx3 := &transaction.Transaction{Nonce: 2}
	tx4 := &smartContractResult.SmartContractResult{Nonce: 3}

	intermediateTxs := map[string]data.TransactionHandler{
		"hash1": tx1,
		"hash2": tx2,
		"hash3": tx3,
		"hash4": tx4,
	}

	miniBlocks := make(map[string]*block.MiniBlock)
	miniBlocks["1"] = &block.MiniBlock{
		Type:     block.SmartContractResultBlock,
		TxHashes: [][]byte{[]byte("hash1")},
	}
	miniBlocks["2"] = &block.MiniBlock{
		Type:     block.SmartContractResultBlock,
		TxHashes: [][]byte{[]byte("hash2")},
	}
	miniBlocks["3"] = &block.MiniBlock{
		Type:     block.InvalidBlock,
		TxHashes: [][]byte{[]byte("hash3")},
	}
	miniBlocks["4"] = &block.MiniBlock{
		Type:     block.SmartContractResultBlock,
		TxHashes: [][]byte{[]byte("hash4")},
	}

	scheduledIntermediateTxsMap := getScheduledIntermediateTxsMap(miniBlocks, intermediateTxs)
	require.Equal(t, 2, len(scheduledIntermediateTxsMap))
	require.Equal(t, 3, len(scheduledIntermediateTxsMap[block.SmartContractResultBlock]))
	require.Equal(t, 1, len(scheduledIntermediateTxsMap[block.InvalidBlock]))
	assert.Equal(t, tx3, scheduledIntermediateTxsMap[block.InvalidBlock][0])
}

func TestStartInEpochWithScheduledDataSyncer_saveScheduledInfoNoScheduledRootHash(t *testing.T) {
	t.Parallel()

	scheduledIntermediateTxs := map[string]data.TransactionHandler{}
	headerHash := []byte("header hash")
	gasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
	}

	sds := &startInEpochWithScheduledDataSyncer{
		scheduledTxsHandler: &testscommon.ScheduledTxsExecutionStub{
			SaveStateCalled: func(headerHash []byte, scheduledInfo *process.ScheduledInfo) {
				t.Error("should not be called")
			},
		},
	}

	scheduledIntermediateTxsMap := getScheduledIntermediateTxsMap(make(map[string]*block.MiniBlock), scheduledIntermediateTxs)
	scheduledInfo := &process.ScheduledInfo{
		RootHash:        nil,
		IntermediateTxs: scheduledIntermediateTxsMap,
		GasAndFees:      gasAndFees,
		MiniBlocks:      make(block.MiniBlockSlice, 0),
	}
	sds.saveScheduledInfo(headerHash, scheduledInfo)
}

func TestStartInEpochWithScheduledDataSyncer_saveScheduledInfo(t *testing.T) {
	t.Parallel()

	scr1 := &smartContractResult.SmartContractResult{
		Nonce: 0,
	}
	scr2 := &smartContractResult.SmartContractResult{
		Nonce: 1,
	}

	scheduledIntermediateTxs := map[string]data.TransactionHandler{
		"txHash1": scr1,
		"txHash2": scr2,
	}
	scheduledRootHash := []byte("scheduled root hash")
	headerHash := []byte("header hash")

	expectedHeaderHash := headerHash
	expectedScheduledRootHash := scheduledRootHash
	expectedScheduledIntermediateTxs := map[block.Type][]data.TransactionHandler{
		block.SmartContractResultBlock: {scr1, scr2},
	}
	gasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(100),
		DeveloperFees:   big.NewInt(30),
		GasProvided:     100,
		GasPenalized:    0,
		GasRefunded:     10,
	}

	expectedGasAndFees := gasAndFees

	sds := &startInEpochWithScheduledDataSyncer{
		scheduledTxsHandler: &testscommon.ScheduledTxsExecutionStub{
			SaveStateCalled: func(headerHash []byte, scheduledInfo *process.ScheduledInfo) {
				require.Equal(t, expectedHeaderHash, headerHash)
				require.Equal(t, expectedScheduledRootHash, scheduledRootHash)
				require.Equal(t, expectedGasAndFees, gasAndFees)
				for i, v := range scheduledInfo.IntermediateTxs {
					require.Equal(t, len(expectedScheduledIntermediateTxs[i]), len(v))
					for j := range v {
						require.Contains(t, expectedScheduledIntermediateTxs[i], v[j])
					}
				}
			},
		},
	}

	scheduledIntermediateTxsMap := getScheduledIntermediateTxsMap(make(map[string]*block.MiniBlock), scheduledIntermediateTxs)
	scheduledInfo := &process.ScheduledInfo{
		RootHash:        scheduledRootHash,
		IntermediateTxs: scheduledIntermediateTxsMap,
		GasAndFees:      gasAndFees,
		MiniBlocks:      make(block.MiniBlockSlice, 0),
	}
	sds.saveScheduledInfo(headerHash, scheduledInfo)
}

func TestStartInEpochWithScheduledDataSyncer_getAllTransactionsForMiniBlocksWithSyncErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("expected error")
	sds := &startInEpochWithScheduledDataSyncer{
		txSyncer: &syncer.TransactionsSyncHandlerMock{
			SyncTransactionsForCalled: func(miniBlocks map[string]*block.MiniBlock, epoch uint32, ctx context.Context) error {
				return expectedErr
			},
		},
	}

	miniBlocks := map[string]*block.MiniBlock{}
	epoch := uint32(0)

	txsMap, err := sds.getAllTransactionsForMiniBlocks(miniBlocks, epoch)
	require.Nil(t, txsMap)
	require.Equal(t, expectedErr, err)
}

func TestStartInEpochWithScheduledDataSyncer_getAllTransactionsForMiniBlocks(t *testing.T) {
	t.Parallel()

	expectedTxs := map[string]data.TransactionHandler{
		"txHash": &transaction.Transaction{Nonce: 1},
	}
	sds := &startInEpochWithScheduledDataSyncer{
		txSyncer: &syncer.TransactionsSyncHandlerMock{
			SyncTransactionsForCalled: func(miniBlocks map[string]*block.MiniBlock, epoch uint32, ctx context.Context) error {
				return nil
			},
			GetTransactionsCalled: func() (map[string]data.TransactionHandler, error) {
				return expectedTxs, nil
			},
		},
	}

	miniBlocks := map[string]*block.MiniBlock{}
	epoch := uint32(0)

	txsMap, err := sds.getAllTransactionsForMiniBlocks(miniBlocks, epoch)
	require.Nil(t, err)
	require.Equal(t, expectedTxs, txsMap)
}

func TestStartInEpochWithScheduledDataSyncer_getScheduledMiniBlockHeaders(t *testing.T) {
	sds := &startInEpochWithScheduledDataSyncer{}
	mbHeaderScheduled1 := block.MiniBlockHeader{
		Hash: []byte("hash1"),
	}
	_ = mbHeaderScheduled1.SetProcessingType(int32(block.Scheduled))
	mbHeaderScheduled2 := block.MiniBlockHeader{
		Hash: []byte("hash2"),
	}
	_ = mbHeaderScheduled2.SetProcessingType(int32(block.Scheduled))
	mbHeader := block.MiniBlockHeader{
		Hash: []byte("hash3"),
	}
	expectedScheduledMbs := []data.MiniBlockHeaderHandler{&mbHeaderScheduled1, &mbHeaderScheduled2}

	header := &block.Header{
		MiniBlockHeaders: []block.MiniBlockHeader{mbHeaderScheduled1, mbHeader, mbHeaderScheduled2},
	}

	mbHeaders := sds.getScheduledMiniBlockHeaders(header)
	require.Equal(t, expectedScheduledMbs, mbHeaders)
}

func TestStartInEpochWithScheduledDataSyncer_getScheduledTransactionHashesWithDestMe(t *testing.T) {
	hashMb1 := []byte("hashMb1")
	hashMb2 := []byte("hashMb2")
	hashMb3 := []byte("hashMb3")
	hashMb4 := []byte("hashMb4")
	txHashes := [][]byte{[]byte("txHash1"), []byte("txHash2"), []byte("txHash3"), []byte("txHash4")}

	mb1 := block.MiniBlock{TxHashes: txHashes[:2]}
	mb2 := block.MiniBlock{TxHashes: txHashes[2:]}
	mbHeaderScheduled1 := block.MiniBlockHeader{
		Hash: hashMb1,
	}
	_ = mbHeaderScheduled1.SetProcessingType(int32(block.Scheduled))
	mbHeaderScheduled2 := block.MiniBlockHeader{
		Hash: hashMb2,
	}
	_ = mbHeaderScheduled2.SetProcessingType(int32(block.Scheduled))
	mbHeaderScheduled3 := block.MiniBlockHeader{
		Hash: hashMb3,
	}
	_ = mbHeaderScheduled3.SetProcessingType(int32(block.Scheduled))
	mbHeader := block.MiniBlockHeader{
		Hash: hashMb4,
	}
	header := &block.Header{
		MiniBlockHeaders: []block.MiniBlockHeader{mbHeaderScheduled1, mbHeader, mbHeaderScheduled2, mbHeaderScheduled3},
	}

	expectedScheduledTxHashes := map[string]uint32{
		string(txHashes[0]): 0,
		string(txHashes[1]): 0,
		string(txHashes[2]): 0,
		string(txHashes[3]): 0,
	}
	sds := &startInEpochWithScheduledDataSyncer{
		scheduledMiniBlocksSyncer: &epochStartMocks.PendingMiniBlockSyncHandlerStub{
			SyncPendingMiniBlocksCalled: func(miniBlockHeaders []data.MiniBlockHeaderHandler, ctx context.Context) error {
				for i := range miniBlockHeaders {
					require.Len(t, miniBlockHeaders[i].GetReserved(), 2)
				}
				return nil
			},
			GetMiniBlocksCalled: func() (map[string]*block.MiniBlock, error) {
				return map[string]*block.MiniBlock{string(hashMb1): &mb1, string(hashMb2): &mb2}, nil
			},
		},
	}

	scheduledTxHashes, err := sds.getScheduledTransactionHashes(header)
	require.Nil(t, err)
	require.Equal(t, expectedScheduledTxHashes, scheduledTxHashes)
}

func Test_getShardIDAndHashesForIncludedMetaBlocks(t *testing.T) {
	metaHash1 := []byte("metaHash1")
	metaHash2 := []byte("metaHash2")
	shardHeader := &block.Header{
		Nonce:           100,
		MetaBlockHashes: [][]byte{metaHash1, metaHash2},
	}

	expectedShardIDs := []uint32{core.MetachainShardId, core.MetachainShardId}
	expectedMetaHashes := [][]byte{metaHash1, metaHash2}
	shardIDs, metaHashes := getShardIDAndHashesForIncludedMetaBlocks(shardHeader)
	require.Equal(t, expectedShardIDs, shardIDs)
	require.Equal(t, expectedMetaHashes, metaHashes)
}

func Test_getPreviousToFirstReferencedMetaHeaderHashNoMetaBlocksReturnsNil(t *testing.T) {
	shardHeader := &block.Header{
		MetaBlockHashes: nil,
	}

	headers := map[string]data.HeaderHandler{}
	hash := getPreviousToFirstReferencedMetaHeaderHash(shardHeader, headers)
	require.Nil(t, hash)
}

func Test_getPreviousToFirstReferencedMetaHeaderHashFirstReferencedMetaNotInMap(t *testing.T) {
	metaHash := []byte("metaHash")
	shardHeader := &block.Header{
		MetaBlockHashes: [][]byte{metaHash},
	}

	headers := map[string]data.HeaderHandler{}
	hash := getPreviousToFirstReferencedMetaHeaderHash(shardHeader, headers)
	require.Nil(t, hash)
}

func Test_getPreviousToFirstReferencedMetaHeaderHashFirstReferencedMetaInvalid(t *testing.T) {
	metaHash := []byte("metaHash")
	shardHeader := &block.Header{
		MetaBlockHashes: [][]byte{metaHash},
	}

	headers := map[string]data.HeaderHandler{
		string(metaHash): &block.Header{},
	}
	hash := getPreviousToFirstReferencedMetaHeaderHash(shardHeader, headers)
	require.Nil(t, hash)
}

func Test_getPreviousToFirstReferencedMetaHeaderHashOK(t *testing.T) {
	metaHash := []byte("metaHash")
	prevMetaHash := []byte("prevMetaHash")
	shardHeader := &block.Header{
		MetaBlockHashes: [][]byte{metaHash},
	}

	headers := map[string]data.HeaderHandler{
		string(metaHash): &block.MetaBlock{
			PrevHash: prevMetaHash,
		},
	}
	hash := getPreviousToFirstReferencedMetaHeaderHash(shardHeader, headers)
	require.Equal(t, prevMetaHash, hash)
}

func createTestHeader() *block.Header {
	return &block.Header{
		Nonce:            100,
		Epoch:            2,
		PrevHash:         []byte("prevShardHeaderHash"),
		ShardID:          0,
		MiniBlockHeaders: nil,
		MetaBlockHashes:  [][]byte{[]byte("metaBlockHash")},
	}
}

func TestGetBlockTypeOfTx(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")

	t.Run("tx not found in miniBlocks, get default block type", func(t *testing.T) {
		t.Parallel()

		miniBlocks := map[string]*block.MiniBlock{
			"dummyhash1": {
				TxHashes: [][]byte{[]byte("hash1")},
				Type:     block.TxBlock,
			},
		}

		blockType := getBlockTypeOfTx(hash, miniBlocks)
		assert.Equal(t, block.SmartContractResultBlock, blockType)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedBlockType := block.PeerBlock
		miniBlocks := map[string]*block.MiniBlock{
			"dummyhash1": {
				TxHashes: [][]byte{[]byte("hash1")},
				Type:     block.TxBlock,
			},
			"dummyhash2": {
				TxHashes: [][]byte{hash},
				Type:     expectedBlockType,
			},
		}

		blockType := getBlockTypeOfTx(hash, miniBlocks)
		assert.Equal(t, expectedBlockType, blockType)
	})
}

func TestGetScheduledMiniBlocks(t *testing.T) {
	t.Parallel()

	hash1, hash2, hash3 := "hash1", "hash2", "hash3"
	txHash1, txHash2, txHash3 := "txHash1", "txHash2", "txHash3"
	mb1 := &block.MiniBlock{
		TxHashes: [][]byte{[]byte(txHash1)},
		Type:     block.TxBlock,
	}
	mb2 := &block.MiniBlock{
		TxHashes: [][]byte{[]byte(txHash2)},
		Type:     block.TxBlock,
	}
	mb3 := &block.MiniBlock{
		TxHashes: [][]byte{[]byte(txHash3)},
		Type:     block.InvalidBlock,
	}
	miniBlocks := map[string]*block.MiniBlock{
		hash1: mb1,
		hash2: mb2,
		hash3: mb3,
	}

	header := &block.Header{
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: []byte(hash1)},
			{Hash: []byte(hash2)},
			{Hash: []byte(hash3)},
		},
	}

	schedulesTxHashes := map[string]uint32{
		txHash1: 1,
		txHash2: 2,
	}

	expectedMiniBlocks := block.MiniBlockSlice{
		mb1,
		mb2,
	}

	mbs := getScheduledMiniBlocks(header, miniBlocks, schedulesTxHashes)
	assert.Equal(t, expectedMiniBlocks, mbs)
}

func TestNumScheduledIntermediateTxs(t *testing.T) {
	t.Parallel()

	mapTxs := map[block.Type][]data.TransactionHandler{
		block.TxBlock: {
			&transaction.Transaction{
				Nonce: 1,
			},
			&transaction.Transaction{
				Nonce: 2,
			},
		},
		block.PeerBlock: {
			&transaction.Transaction{
				Nonce: 3,
			},
		},
	}

	num := getNumScheduledIntermediateTxs(mapTxs)
	assert.Equal(t, 3, num)
}
