package block

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

func createGenesisBlock(shardId uint32) *block.Header {
	rootHash := []byte("roothash")
	return &block.Header{
		Nonce:           0,
		Round:           0,
		Signature:       rootHash,
		RandSeed:        rootHash,
		PrevRandSeed:    rootHash,
		ShardID:         shardId,
		PubKeysBitmap:   rootHash,
		RootHash:        rootHash,
		PrevHash:        rootHash,
		MetaBlockHashes: [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
	}
}

func TestDisplayBlock_NewTransactionCounterShouldErrWhenHasherIsNil(t *testing.T) {
	t.Parallel()

	txCounter, err := NewTransactionCounter(nil, &testscommon.MarshalizerMock{})

	assert.Nil(t, txCounter)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestDisplayBlock_NewTransactionCounterShouldErrWhenMarshalizerIsNil(t *testing.T) {
	t.Parallel()

	txCounter, err := NewTransactionCounter(&testscommon.HasherStub{}, nil)

	assert.Nil(t, txCounter)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestDisplayBlock_NewTransactionCounterShouldWork(t *testing.T) {
	t.Parallel()

	txCounter, err := NewTransactionCounter(&testscommon.HasherStub{}, &testscommon.MarshalizerMock{})

	assert.NotNil(t, txCounter)
	assert.Nil(t, err)
}

func TestDisplayBlock_DisplayMetaHashesIncluded(t *testing.T) {
	t.Parallel()

	shardLines := make([]*display.LineData, 0)
	header := createGenesisBlock(0)
	txCounter, _ := NewTransactionCounter(&testscommon.HasherStub{}, &testscommon.MarshalizerMock{})
	lines := txCounter.displayMetaHashesIncluded(
		shardLines,
		header,
	)

	assert.NotNil(t, lines)
	assert.Equal(t, len(header.MetaBlockHashes), len(lines))
}

func TestDisplayBlock_DisplayTxBlockBody(t *testing.T) {
	t.Parallel()

	shardLines := make([]*display.LineData, 0)
	body := &block.Body{}
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)
	txCounter, _ := NewTransactionCounter(&testscommon.HasherStub{}, &testscommon.MarshalizerMock{})
	lines := txCounter.displayTxBlockBody(
		shardLines,
		&block.Header{},
		body,
	)

	assert.NotNil(t, lines)
	assert.Equal(t, len(miniblock.TxHashes), len(lines))
}

func TestDisplayBlock_GetConstructionStateAsString(t *testing.T) {
	miniBlockHeader := &block.MiniBlockHeader{}

	_ = miniBlockHeader.SetConstructionState(int32(block.Proposed))
	str := getConstructionStateAsString(miniBlockHeader)
	assert.Equal(t, "Proposed_", str)

	_ = miniBlockHeader.SetConstructionState(int32(block.PartialExecuted))
	str = getConstructionStateAsString(miniBlockHeader)
	assert.Equal(t, "Partial_", str)

	_ = miniBlockHeader.SetConstructionState(int32(block.Final))
	str = getConstructionStateAsString(miniBlockHeader)
	assert.Equal(t, "", str)
}

func TestDisplayBlock_setNumProcessedTxsMetric(t *testing.T) {
	t.Parallel()

	setValue := uint64(0)
	appStatusHandler := &statusHandler.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			if key == common.MetricNumProcessedTxs {
				setValue = value
			}
		},
	}
	txCounter, _ := NewTransactionCounter(&testscommon.HasherStub{}, &testscommon.MarshalizerMock{})
	txCounter.totalTxs = 37
	txCounter.setNumProcessedTxsMetric(appStatusHandler)

	assert.Equal(t, uint64(37), setValue)
}

func TestDisplayBlock_ConcurrencyTestForTotalTxs(t *testing.T) {
	t.Parallel()

	numCalls := 100
	wg := sync.WaitGroup{}
	wg.Add(numCalls)

	txCounter, _ := NewTransactionCounter(&testscommon.HasherStub{}, &testscommon.MarshalizerMock{})
	txCounter.totalTxs = math.MaxUint64
	appStatusHandler := &statusHandler.AppStatusHandlerStub{}

	for i := 0; i < numCalls; i++ {
		go func(idx int) {
			time.Sleep(time.Millisecond * 10)
			defer wg.Done()

			switch idx % 2 {
			case 0:
				txCounter.subtractRestoredTxs(1)
			case 1:
				txCounter.setNumProcessedTxsMetric(appStatusHandler)
			}
		}(i)
	}

	wg.Wait()
}
