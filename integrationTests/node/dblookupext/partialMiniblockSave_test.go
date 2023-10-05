package dblookupext

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon"
	dblookupextMock "github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/integrationtests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockHistoryRepositoryArguments() dblookupext.HistoryRepositoryArguments {
	return dblookupext.HistoryRepositoryArguments{
		SelfShardID:                 0,
		MiniblocksMetadataStorer:    testscommon.CreateMemUnit(),
		MiniblockHashByTxHashStorer: testscommon.CreateMemUnit(),
		BlockHashByRound:            testscommon.CreateMemUnit(),
		Uint64ByteSliceConverter:    uint64ByteSlice.NewBigEndianConverter(),
		EpochByHashStorer:           testscommon.CreateMemUnit(),
		EventsHashesByTxHashStorer:  testscommon.CreateMemUnit(),
		Marshalizer:                 integrationtests.TestMarshalizer,
		Hasher:                      integrationtests.TestHasher,
		ESDTSuppliesHandler:         &dblookupextMock.SuppliesHandlerStub{},
	}
}

func TestSavePartialMiniblocksShouldKeepAllData(t *testing.T) {
	// TODO(this feat) fix this test
	t.Skip("this test will be fixed in upcoming PRs")

	args := createMockHistoryRepositoryArguments()

	historyRepository, err := dblookupext.NewHistoryRepository(args)
	require.Nil(t, err)

	txHashPattern := "tx hash %d"
	numTxsInMb := 10
	miniblock := &block.MiniBlock{
		TxHashes:        make([][]byte, 0),
		ReceiverShardID: 0,
		SenderShardID:   0,
		Type:            block.TxBlock,
	}
	for i := 0; i < numTxsInMb; i++ {
		miniblock.TxHashes = append(miniblock.TxHashes, []byte(fmt.Sprintf(txHashPattern, i)))
	}
	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{miniblock},
	}
	miniblockHash, _ := core.CalculateHash(integrationtests.TestMarshalizer, integrationtests.TestHasher, miniblock)

	mbHeader1 := block.MiniBlockHeader{
		Hash:            miniblockHash,
		SenderShardID:   0,
		ReceiverShardID: 0,
		TxCount:         uint32(numTxsInMb),
		Type:            block.TxBlock,
	}
	_ = mbHeader1.SetConstructionState(int32(block.PartialExecuted))
	// transactions with indexes 0, 1 and 2 were executed in this mb header
	_ = mbHeader1.SetIndexOfLastTxProcessed(0)
	_ = mbHeader1.SetIndexOfLastTxProcessed(2)

	header1 := &block.Header{
		MiniBlockHeaders: []block.MiniBlockHeader{mbHeader1},
		TxCount:          mbHeader1.TxCount,
	}
	headerHash1, _ := core.CalculateHash(integrationtests.TestMarshalizer, integrationtests.TestHasher, header1)

	// record the first block header
	err = historyRepository.RecordBlock(
		headerHash1,
		header1,
		body,
		make(map[string]data.TransactionHandler),
		make(map[string]data.TransactionHandler),
		make([]*block.MiniBlock, 0),
		make([]*data.LogData, 0),
	)
	require.Nil(t, err)

	mbHeader2 := block.MiniBlockHeader{
		Hash:            miniblockHash,
		SenderShardID:   0,
		ReceiverShardID: 0,
		TxCount:         uint32(numTxsInMb),
		Type:            block.TxBlock,
	}
	_ = mbHeader2.SetConstructionState(int32(block.PartialExecuted))
	// transactions with indexes 3 and 4 were executed in this mb header
	_ = mbHeader2.SetIndexOfLastTxProcessed(3)
	_ = mbHeader2.SetIndexOfLastTxProcessed(4)

	header2 := &block.Header{
		MiniBlockHeaders: []block.MiniBlockHeader{mbHeader2},
		TxCount:          mbHeader2.TxCount,
	}
	headerHash2, _ := core.CalculateHash(integrationtests.TestMarshalizer, integrationtests.TestHasher, header2)

	// record the second block header
	err = historyRepository.RecordBlock(
		headerHash2,
		header2,
		body,
		make(map[string]data.TransactionHandler),
		make(map[string]data.TransactionHandler),
		make([]*block.MiniBlock, 0),
		make([]*data.LogData, 0),
	)
	require.Nil(t, err)

	txHash0 := []byte(fmt.Sprintf(txHashPattern, 0)) // hash 0 should be linked to header 1
	mbData, err := historyRepository.GetMiniblockMetadataByTxHash(txHash0)
	require.Nil(t, err)
	assert.Equal(t, headerHash1, mbData.HeaderHash)

	txHash3 := []byte(fmt.Sprintf(txHashPattern, 3)) // hash 3 should be linked to header 2
	mbData, err = historyRepository.GetMiniblockMetadataByTxHash(txHash3)
	require.Nil(t, err)
	assert.Equal(t, headerHash2, mbData.HeaderHash)

	txHash5 := []byte(fmt.Sprintf(txHashPattern, 5)) // hash 5 should not be linked, it was not processed yet
	mbData, err = historyRepository.GetMiniblockMetadataByTxHash(txHash5)
	require.Nil(t, mbData)
	assert.Equal(t, dblookupext.ErrNotFoundInStorage, err)
}
