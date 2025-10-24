package dblookupext

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/require"
)

func TestGetIntermediateTxs(t *testing.T) {
	t.Parallel()

	t.Run("getIntermediateTxs cannot find in cache", func(t *testing.T) {
		t.Parallel()

		cacher := cache.NewCacherMock()

		headerHash := []byte("h")

		_, _, err := getIntermediateTxs(cacher, headerHash)
		require.True(t, errors.Is(err, process.ErrMissingHeader))
	})

	t.Run("getIntermediateTxs wrong type in cache should return empty maps", func(t *testing.T) {
		t.Parallel()

		cacher := cache.NewCacherMock()

		headerHash := []byte("h")
		cacher.Put(headerHash, []byte("wrong"), 0)

		_, _, err := getIntermediateTxs(cacher, headerHash)
		require.True(t, errors.Is(err, process.ErrWrongTypeAssertion))
	})

	t.Run("getIntermediateTxs should work", func(t *testing.T) {
		t.Parallel()

		cachedIntermediateTxsMap := map[block.Type]map[string]data.TransactionHandler{}
		cachedIntermediateTxsMap[block.SmartContractResultBlock] = map[string]data.TransactionHandler{
			"h1": &smartContractResult.SmartContractResult{},
		}
		cachedIntermediateTxsMap[block.ReceiptBlock] = map[string]data.TransactionHandler{
			"r1": &receipt.Receipt{},
		}

		cacher := cache.NewCacherMock()

		headerHash := []byte("h")
		cacher.Put(headerHash, cachedIntermediateTxsMap, 0)

		scrs, receipts, err := getIntermediateTxs(cacher, headerHash)
		require.NoError(t, err)
		require.Len(t, scrs, 1)
		require.Len(t, receipts, 1)
		require.Equal(t, cachedIntermediateTxsMap[block.SmartContractResultBlock], scrs)
		require.Equal(t, cachedIntermediateTxsMap[block.ReceiptBlock], receipts)
	})

}

func TestGetLogs(t *testing.T) {
	t.Parallel()

	t.Run("getLogs cannot find in cache", func(t *testing.T) {
		t.Parallel()

		cacher := cache.NewCacherMock()

		headerHash := []byte("h")

		_, err := getLogs(cacher, headerHash)
		require.True(t, errors.Is(err, process.ErrMissingHeader))
	})

	t.Run("getLogs wrong type in cache should return empty slice", func(t *testing.T) {
		t.Parallel()

		cacher := cache.NewCacherMock()

		headerHash := []byte("h")
		logsKey := common.PrepareLogEventsKey(headerHash)
		cacher.Put(logsKey, "wrong type", 0)

		_, err := getLogs(cacher, headerHash)
		require.True(t, errors.Is(err, process.ErrWrongTypeAssertion))
	})

	t.Run("getLogs should work", func(t *testing.T) {
		t.Parallel()

		cacher := cache.NewCacherMock()

		headerHash := []byte("h")
		expectedLogs := []*data.LogData{
			{
				LogHandler: &transaction.Log{},
				TxHash:     "t1",
			},
			{
				LogHandler: &transaction.Log{},
				TxHash:     "t2",
			},
		}
		logsKey := common.PrepareLogEventsKey(headerHash)

		cacher.Put(logsKey, expectedLogs, 0)

		logs, err := getLogs(cacher, headerHash)
		require.Nil(t, err)
		require.Len(t, logs, 2)
		require.Equal(t, expectedLogs, logs)
	})

}

func TestGetIntraMbs(t *testing.T) {
	t.Parallel()

	t.Run("getIntraMbs cannot find in cache", func(t *testing.T) {
		t.Parallel()

		cacher := cache.NewCacherMock()
		marshaller := &marshallerMock.MarshalizerMock{}

		headerHash := []byte("h")

		_, err := getIntraMbs(cacher, marshaller, headerHash)
		require.True(t, errors.Is(err, process.ErrMissingHeader))
	})

	t.Run("getIntraMbs wrong type should error", func(t *testing.T) {
		t.Parallel()

		cacher := cache.NewCacherMock()
		marshaller := &marshallerMock.MarshalizerMock{}

		headerHash := []byte("h")
		cacher.Put(headerHash, []byte("wrong type"), 0)

		intraMBs, err := getIntraMbs(cacher, marshaller, headerHash)
		require.Nil(t, intraMBs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), "getIntraMbs: cannot unmarshall"))
	})

	t.Run("getIntraMbs should work", func(t *testing.T) {
		t.Parallel()

		cacher := cache.NewCacherMock()
		marshaller := &marshallerMock.MarshalizerMock{}

		headerHash := []byte("h")
		expectedMbs := []*block.MiniBlock{
			{SenderShardID: 0},
			{SenderShardID: 0},
		}
		intraMbsBytes, _ := marshaller.Marshal(expectedMbs)

		cacher.Put(headerHash, intraMbsBytes, 0)

		intraMBs, err := getIntraMbs(cacher, marshaller, headerHash)
		require.Nil(t, err)
		require.Equal(t, expectedMbs, intraMBs)
	})

}

func TestExtractMiniBlocksHeaderHandlersFromExecResult(t *testing.T) {
	t.Run("wrong type shard", func(t *testing.T) {
		t.Parallel()

		executionResult := &block.BaseExecutionResult{}

		_, err := extractMiniBlocksHeaderHandlersFromExecResult(executionResult, 0)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})
	t.Run("should work shard", func(t *testing.T) {
		executionResult := &block.ExecutionResult{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{SenderShardID: 1},
				{SenderShardID: 0},
			},
		}

		res, err := extractMiniBlocksHeaderHandlersFromExecResult(executionResult, 0)
		require.Nil(t, err)
		require.Len(t, res, 2)
	})
	t.Run("wrong type meta", func(t *testing.T) {
		t.Parallel()

		executionResult := &block.BaseExecutionResult{}
		_, err := extractMiniBlocksHeaderHandlersFromExecResult(executionResult, core.MetachainShardId)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})
	t.Run("should work meta", func(t *testing.T) {
		t.Parallel()

		executionResult := &block.MetaExecutionResult{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{SenderShardID: 0},
				{SenderShardID: 1},
			},
		}

		res, err := extractMiniBlocksHeaderHandlersFromExecResult(executionResult, core.MetachainShardId)
		require.Nil(t, err)
		require.Len(t, res, 2)
	})
}

func TestGetBody(t *testing.T) {
	t.Parallel()

	t.Run("cannot get mb headers should error", func(t *testing.T) {
		t.Parallel()

		executionResult := &block.BaseExecutionResult{}

		marshaller := &marshallerMock.MarshalizerMock{}
		cacher := cache.NewCacherMock()

		_, err := getBody(cacher, marshaller, executionResult, 0)
		require.NotNil(t, err)
	})

	t.Run("missing miniblock should error", func(t *testing.T) {
		t.Parallel()

		mb1 := &block.MiniBlock{
			SenderShardID: 1,
		}
		h1 := []byte("h1")
		h2 := []byte("h2")

		executionResult := &block.ExecutionResult{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{
					Hash: h1,
				},
				{
					Hash: h2,
				},
			},
		}

		marshaller := &marshallerMock.MarshalizerMock{}
		mb1Bytes, _ := marshaller.Marshal(mb1)

		cacher := cache.NewCacherMock()
		cacher.Put(h1, mb1Bytes, 0)

		_, err := getBody(cacher, marshaller, executionResult, 0)
		require.Equal(t, process.ErrMissingMiniBlock, err)
	})

	t.Run("marshaller error", func(t *testing.T) {
		t.Parallel()

		h1 := []byte("h1")
		h2 := []byte("h2")

		executionResult := &block.ExecutionResult{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{
					Hash: h1,
				},
				{
					Hash: h2,
				},
			},
		}

		marshaller := &marshallerMock.MarshalizerMock{}

		cacher := cache.NewCacherMock()
		cacher.Put(h1, []byte("wrong"), 0)

		_, err := getBody(cacher, marshaller, executionResult, 0)
		require.NotNil(t, err)
	})

	t.Run("getBody should work", func(t *testing.T) {
		t.Parallel()

		mb1 := &block.MiniBlock{
			SenderShardID: 1,
		}
		mb2 := &block.MiniBlock{
			SenderShardID: 2,
		}
		h1 := []byte("h1")
		h2 := []byte("h2")

		executionResult := &block.ExecutionResult{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{
					Hash: h1,
				},
				{
					Hash: h2,
				},
			},
		}

		marshaller := &marshallerMock.MarshalizerMock{}
		mb1Bytes, _ := marshaller.Marshal(mb1)
		mb2Bytes, _ := marshaller.Marshal(mb2)

		cacher := cache.NewCacherMock()
		cacher.Put(h1, mb1Bytes, 0)
		cacher.Put(h2, mb2Bytes, 0)

		res, err := getBody(cacher, marshaller, executionResult, 0)
		require.Nil(t, err)
		require.Equal(t, &block.Body{MiniBlocks: []*block.MiniBlock{mb1, mb2}}, res)
	})
}
