package bootstrap

import (
	"context"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
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

func TestStartInEpochWithScheduledDataSyncer_getRequiredHeaderByHash(t *testing.T) {

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
		ScheduledRootHash: expectedRootHash,
	}

	rootHash := ds.GetRootHashToSync(notarizedHeader)
	require.Equal(t, expectedRootHash, rootHash)
}

func TestStartInEpochWithScheduledDataSyncer_prepareScheduledSCRs(t *testing.T) {

}

func TestStartInEpochWithScheduledDataSyncer_filterScheduledSCRs(t *testing.T) {
	t.Parallel()

	sds := &startInEpochWithScheduledDataSyncer{}
	scheduledTxHashes := []string{
		"txHash1", "txHash2",
	}

	scheduledTxHashesMap := map[string]struct{}{
		scheduledTxHashes[0]: {},
		scheduledTxHashes[1]: {},
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

	scheduledTxs, err := sds.filterScheduledSCRs(scheduledTxHashesMap, allTxsMap)
	require.Nil(t, err)
	require.Equal(t, expectedScheduledTxsMap, scheduledTxs)
}

func TestStartInEpochWithScheduledDataSyncer_saveScheduledSCRsNoScheduledRootHash(t *testing.T) {
	t.Parallel()

	scheduledSCRs := map[string]data.TransactionHandler{}
	headerHash := []byte("header hash")

	sds := &startInEpochWithScheduledDataSyncer{
		scheduledTxsHandler: &testscommon.ScheduledTxsExecutionStub{
			SaveStateCalled: func(headerHash []byte, scheduledRootHash []byte, mapScheduledSCRs map[block.Type][]data.TransactionHandler) {
				t.Error("should not be called")
			},
		},
	}

	sds.saveScheduledSCRs(scheduledSCRs, nil, headerHash)
}

func TestStartInEpochWithScheduledDataSyncer_saveScheduledSCRs(t *testing.T) {
	t.Parallel()

	scr1 := &smartContractResult.SmartContractResult{
		Nonce: 0,
	}
	scr2 := &smartContractResult.SmartContractResult{
		Nonce: 1,
	}

	scheduledSCRs := map[string]data.TransactionHandler{
		"txhash1": scr1,
		"txHash2": scr2,
	}
	scheduledRootHash := []byte("scheduled root hash")
	headerHash := []byte("header hash")

	expectedHeaderHash := headerHash
	expectedScheduledRootHash := scheduledRootHash
	expectedScheduledSCRsMap := map[block.Type][]data.TransactionHandler{
		block.TxBlock: {scr1, scr2},
	}

	sds := &startInEpochWithScheduledDataSyncer{
		scheduledTxsHandler: &testscommon.ScheduledTxsExecutionStub{
			SaveStateCalled: func(headerHash []byte, scheduledRootHash []byte, mapScheduledSCRs map[block.Type][]data.TransactionHandler) {
				require.Equal(t, expectedHeaderHash, headerHash)
				require.Equal(t, expectedScheduledRootHash, scheduledRootHash)
				require.Equal(t, expectedScheduledSCRsMap, mapScheduledSCRs)
			},
		},
	}

	sds.saveScheduledSCRs(scheduledSCRs, scheduledRootHash, headerHash)
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
		Hash:     []byte("hash1"),
		Reserved: []byte{byte(block.Scheduled)},
	}
	mbHeaderScheduled2 := block.MiniBlockHeader{
		Hash:     []byte("hash2"),
		Reserved: []byte{byte(block.Scheduled)},
	}
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

func TestStartInEpochWithScheduledDataSyncer_getScheduledTransactionHashes(t *testing.T) {
	hashMb1 := []byte("hashMb1")
	hashMb2 := []byte("hashMb2")
	hashMb3 := []byte("hashMb3")
	txHashes := [][]byte{[]byte("txHash1"), []byte("txHash2"), []byte("txHash3"), []byte("txHash4")}

	mb1 := block.MiniBlock{TxHashes: txHashes[:2]}
	mb2 := block.MiniBlock{TxHashes: txHashes[2:]}
	mbHeaderScheduled1 := block.MiniBlockHeader{
		Hash:     hashMb1,
		Reserved: []byte{byte(block.Scheduled)},
	}
	mbHeaderScheduled2 := block.MiniBlockHeader{
		Hash:     hashMb2,
		Reserved: []byte{byte(block.Scheduled)},
	}
	mbHeader := block.MiniBlockHeader{
		Hash: hashMb3,
	}
	header := &block.Header{
		MiniBlockHeaders: []block.MiniBlockHeader{mbHeaderScheduled1, mbHeader, mbHeaderScheduled2},
	}

	expectedScheduledTxHashes := map[string]struct{}{
		string(txHashes[0]): {},
		string(txHashes[1]): {},
		string(txHashes[2]): {},
		string(txHashes[3]): {},
	}
	sds := &startInEpochWithScheduledDataSyncer{
		scheduledMiniBlocksSyncer: &epochStartMocks.PendingMiniBlockSyncHandlerStub{
			SyncPendingMiniBlocksCalled: func(miniBlockHeaders []data.MiniBlockHeaderHandler, ctx context.Context) error {
				for i := range miniBlockHeaders {
					require.Len(t, miniBlockHeaders[i].GetReserved(), 1)
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
