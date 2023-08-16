package process

import (
	"crypto/rand"
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/outport/mock"
	"github.com/multiversx/mx-chain-go/outport/process/transactionsfee"
	"github.com/multiversx/mx-chain-go/testscommon"
	commonMocks "github.com/multiversx/mx-chain-go/testscommon/common"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/require"
)

func createArgOutportDataProvider() ArgOutportDataProvider {
	txsFeeProc, _ := transactionsfee.NewTransactionsFeeProcessor(transactionsfee.ArgTransactionsFeeProcessor{
		Marshaller:         &marshallerMock.MarshalizerMock{},
		TransactionsStorer: &genericMocks.StorerMock{},
		ShardCoordinator:   &testscommon.ShardsCoordinatorMock{},
		TxFeeCalculator:    &mock.EconomicsHandlerMock{},
	})

	return ArgOutportDataProvider{
		AlteredAccountsProvider:  &testscommon.AlteredAccountsProviderStub{},
		TransactionsFeeProcessor: txsFeeProc,
		TxCoordinator:            &testscommon.TransactionCoordinatorMock{},
		NodesCoordinator:         &shardingMocks.NodesCoordinatorMock{},
		GasConsumedProvider:      &testscommon.GasHandlerStub{},
		EconomicsData:            &mock.EconomicsHandlerMock{},
		ShardCoordinator:         &testscommon.ShardsCoordinatorMock{},
		ExecutionOrderHandler:    &commonMocks.TxExecutionOrderHandlerStub{},
		Marshaller:               &marshallerMock.MarshalizerMock{},
		Hasher:                   &hashingMocks.HasherMock{},
	}
}

func TestNewOutportDataProvider(t *testing.T) {
	t.Parallel()

	arg := createArgOutportDataProvider()
	outportDataP, err := NewOutportDataProvider(arg)
	require.Nil(t, err)
	require.False(t, outportDataP.IsInterfaceNil())
}

func TestPrepareOutportSaveBlockDataNilHeader(t *testing.T) {
	t.Parallel()

	arg := createArgOutportDataProvider()
	outportDataP, _ := NewOutportDataProvider(arg)

	_, err := outportDataP.PrepareOutportSaveBlockData(ArgPrepareOutportSaveBlockData{})
	require.Equal(t, ErrNilHeaderHandler, err)
}

func TestPrepareOutportSaveBlockDataNilBody(t *testing.T) {
	t.Parallel()

	arg := createArgOutportDataProvider()
	outportDataP, _ := NewOutportDataProvider(arg)

	_, err := outportDataP.PrepareOutportSaveBlockData(ArgPrepareOutportSaveBlockData{
		Header: &block.Header{},
	})
	require.Equal(t, ErrNilBodyHandler, err)
}

func TestPrepareOutportSaveBlockData(t *testing.T) {
	t.Parallel()

	arg := createArgOutportDataProvider()
	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		GetValidatorsPublicKeysCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error) {
			return nil, nil
		},
		GetValidatorsIndexesCalled: func(publicKeys []string, epoch uint32) ([]uint64, error) {
			return []uint64{0, 1}, nil
		},
	}
	outportDataP, _ := NewOutportDataProvider(arg)

	res, err := outportDataP.PrepareOutportSaveBlockData(ArgPrepareOutportSaveBlockData{
		Header:     &block.Header{},
		Body:       &block.Body{},
		HeaderHash: []byte("something"),
	})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.HeaderDataWithBody.HeaderHash)
	require.NotNil(t, res.HeaderDataWithBody.Body)
	require.NotNil(t, res.HeaderDataWithBody.Header)
	require.NotNil(t, res.SignersIndexes)
	require.NotNil(t, res.HeaderGasConsumption)
	require.NotNil(t, res.TransactionPool)
}

func TestOutportDataProvider_GetIntraShardMiniBlocks(t *testing.T) {
	t.Parallel()

	mb1 := &block.MiniBlock{
		Type:     block.SmartContractResultBlock,
		TxHashes: [][]byte{[]byte("scr1")},
	}
	mb2 := &block.MiniBlock{
		SenderShardID:   0,
		ReceiverShardID: 1,
		Type:            block.SmartContractResultBlock,
		TxHashes:        [][]byte{[]byte("scr2"), []byte("scr3")},
	}
	mb3 := &block.MiniBlock{
		Type:     block.SmartContractResultBlock,
		TxHashes: [][]byte{[]byte("scr4"), []byte("scr5")},
	}

	arg := createArgOutportDataProvider()
	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		GetValidatorsPublicKeysCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error) {
			return nil, nil
		},
		GetValidatorsIndexesCalled: func(publicKeys []string, epoch uint32) ([]uint64, error) {
			return []uint64{0, 1}, nil
		},
	}
	arg.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		GetCreatedInShardMiniBlocksCalled: func() []*block.MiniBlock {
			return []*block.MiniBlock{mb1, mb3}
		},
	}
	outportDataP, _ := NewOutportDataProvider(arg)

	res, err := outportDataP.PrepareOutportSaveBlockData(ArgPrepareOutportSaveBlockData{
		Header: &block.Header{},
		Body: &block.Body{
			MiniBlocks: []*block.MiniBlock{mb1, mb2},
		},
		HeaderHash: []byte("something"),
	})
	require.Nil(t, err)
	require.Equal(t, []*block.MiniBlock{mb3}, res.HeaderDataWithBody.IntraShardMiniBlocks)
}

func Test_extractExecutedTxsFromMb(t *testing.T) {
	t.Parallel()

	t.Run("nil mini block header", func(t *testing.T) {
		executedTxHashes := make(map[string]struct{})
		err := extractExecutedTxsFromMb(nil, &block.MiniBlock{}, executedTxHashes)
		require.Equal(t, ErrNilMiniBlockHeaderHandler, err)
		require.Equal(t, 0, len(executedTxHashes))
	})
	t.Run("nil mini block", func(t *testing.T) {
		executedTxHashes := make(map[string]struct{})
		err := extractExecutedTxsFromMb(&block.MiniBlockHeader{}, nil, executedTxHashes)
		require.Equal(t, ErrNilMiniBlock, err)
		require.Equal(t, 0, len(executedTxHashes))
	})
	t.Run("nil executed tx hashes", func(t *testing.T) {
		err := extractExecutedTxsFromMb(&block.MiniBlockHeader{}, &block.MiniBlock{}, nil)
		require.Equal(t, ErrNilExecutedTxHashes, err)
	})
	t.Run("should skip peer miniBlocks", func(t *testing.T) {
		executedTxHashes := make(map[string]struct{})
		mbHeader := &block.MiniBlockHeader{
			Type: block.PeerBlock,
		}
		err := extractExecutedTxsFromMb(mbHeader, &block.MiniBlock{}, executedTxHashes)
		require.Nil(t, err)
		require.Equal(t, 0, len(executedTxHashes))
	})
	t.Run("should skip processed miniBlocks", func(t *testing.T) {
		executedTxHashes := make(map[string]struct{})
		mbs, mbHeaders := createMbsAndMbHeaders(1, 10)
		mbHeaders[0].Type = block.TxBlock
		err := mbHeaders[0].SetProcessingType(int32(block.Processed))
		require.Nil(t, err)

		err = extractExecutedTxsFromMb(&mbHeaders[0], mbs[0], executedTxHashes)
		require.Nil(t, err)
		require.Equal(t, 0, len(executedTxHashes))
	})

	t.Run("should only return processed range (first txs in range)", func(t *testing.T) {
		executedTxHashes := make(map[string]struct{})
		mbs, mbHeaders := createMbsAndMbHeaders(1, 10)
		mbHeaders[0].Type = block.TxBlock
		_ = mbHeaders[0].SetIndexOfFirstTxProcessed(0)
		_ = mbHeaders[0].SetIndexOfLastTxProcessed(5)

		err := extractExecutedTxsFromMb(&mbHeaders[0], mbs[0], executedTxHashes)
		require.Nil(t, err)
		require.Equal(t, 6, len(executedTxHashes))
		for i := 0; i < 6; i++ {
			_, ok := executedTxHashes[string(mbs[0].TxHashes[i])]
			require.True(t, ok)
		}
	})
	t.Run("should only return processed range (last txs in range)", func(t *testing.T) {
		executedTxHashes := make(map[string]struct{})
		mbs, mbHeaders := createMbsAndMbHeaders(1, 10)
		mbHeaders[0].Type = block.TxBlock
		_ = mbHeaders[0].SetIndexOfFirstTxProcessed(5)
		_ = mbHeaders[0].SetIndexOfLastTxProcessed(9)

		err := extractExecutedTxsFromMb(&mbHeaders[0], mbs[0], executedTxHashes)
		require.Nil(t, err)
		require.Equal(t, 5, len(executedTxHashes))
		for i := 5; i < 10; i++ {
			_, ok := executedTxHashes[string(mbs[0].TxHashes[i])]
			require.True(t, ok)
		}
	})
	t.Run("should only return processed range (middle txs in range)", func(t *testing.T) {
		executedTxHashes := make(map[string]struct{})
		mbs, mbHeaders := createMbsAndMbHeaders(1, 10)
		mbHeaders[0].Type = block.TxBlock
		_ = mbHeaders[0].SetIndexOfFirstTxProcessed(3)
		_ = mbHeaders[0].SetIndexOfLastTxProcessed(7)

		err := extractExecutedTxsFromMb(&mbHeaders[0], mbs[0], executedTxHashes)
		require.Nil(t, err)
		require.Equal(t, 5, len(executedTxHashes))
		for i := 3; i < 8; i++ {
			_, ok := executedTxHashes[string(mbs[0].TxHashes[i])]
			require.True(t, ok)
		}
	})
	t.Run("should only return processed range (all txs in range)", func(t *testing.T) {
		executedTxHashes := make(map[string]struct{})
		mbs, mbHeaders := createMbsAndMbHeaders(1, 10)
		mbHeaders[0].Type = block.TxBlock
		_ = mbHeaders[0].SetIndexOfFirstTxProcessed(0)
		_ = mbHeaders[0].SetIndexOfLastTxProcessed(9)

		err := extractExecutedTxsFromMb(&mbHeaders[0], mbs[0], executedTxHashes)
		require.Nil(t, err)
		require.Equal(t, 10, len(executedTxHashes))
		for i := 0; i < 10; i++ {
			_, ok := executedTxHashes[string(mbs[0].TxHashes[i])]
			require.True(t, ok)
		}
	})
}

func Test_setExecutionOrderInTransactionPool(t *testing.T) {
	t.Parallel()

	t.Run("nil pool", func(t *testing.T) {
		args := createArgOutportDataProvider()
		odp, _ := NewOutportDataProvider(args)
		txHashes := createRandTxHashes(10)
		odp.executionOrderHandler = &commonMocks.TxExecutionOrderHandlerStub{
			GetItemsCalled: func() [][]byte {
				return txHashes
			},
		}

		orderedHashes, numProcessed := odp.setExecutionOrderInTransactionPool(nil)
		require.Equal(t, 0, numProcessed)
		require.Equal(t, len(txHashes), len(orderedHashes))
		for i := 0; i < len(txHashes); i++ {
			require.Equal(t, txHashes[i], orderedHashes[i])
		}
	})
	t.Run("nil pool txs, scrs, rewards", func(t *testing.T) {
		args := createArgOutportDataProvider()
		odp, _ := NewOutportDataProvider(args)
		pool := &outportcore.Pool{
			Txs:     nil,
			Scrs:    nil,
			Rewards: nil,
		}

		txHashes := createRandTxHashes(10)
		odp.executionOrderHandler = &commonMocks.TxExecutionOrderHandlerStub{
			GetItemsCalled: func() [][]byte {
				return txHashes
			},
		}

		orderedHashes, numProcessed := odp.setExecutionOrderInTransactionPool(pool)
		require.Equal(t, 0, numProcessed)
		require.Equal(t, len(txHashes), len(orderedHashes))
	})
	t.Run("transactions not found in pool txs, scrs or rewards", func(t *testing.T) {
		args := createArgOutportDataProvider()
		odp, _ := NewOutportDataProvider(args)
		pool := &outportcore.Pool{
			Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
				"tx1": &outportcore.TransactionHandlerWithGasAndFee{
					TransactionHandler: &transaction.Transaction{
						Nonce: 0,
					},
				},
			},
			Scrs: map[string]data.TransactionHandlerWithGasUsedAndFee{
				"scr1": &outportcore.TransactionHandlerWithGasAndFee{
					TransactionHandler: &smartContractResult.SmartContractResult{
						Nonce: 0,
					},
				},
			},
			Rewards: map[string]data.TransactionHandlerWithGasUsedAndFee{
				"reward1": &outportcore.TransactionHandlerWithGasAndFee{
					TransactionHandler: &rewardTx.RewardTx{
						Epoch: 0,
					},
				},
			},
		}

		txHashes := createRandTxHashes(10)
		odp.executionOrderHandler = &commonMocks.TxExecutionOrderHandlerStub{
			GetItemsCalled: func() [][]byte {
				return txHashes
			},
		}

		orderedHashes, numProcessed := odp.setExecutionOrderInTransactionPool(pool)
		require.Equal(t, 0, numProcessed)
		require.Equal(t, len(txHashes), len(orderedHashes))
	})

	t.Run("transactions partially found (first txs in list) in pool txs, scrs or rewards", func(t *testing.T) {
		args := createArgOutportDataProvider()
		odp, _ := NewOutportDataProvider(args)
		txHashes := createRandTxHashes(10)
		pool := &outportcore.Pool{
			Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
				string(txHashes[0]): &outportcore.TransactionHandlerWithGasAndFee{
					TransactionHandler: &transaction.Transaction{
						Nonce: 0,
					},
				},
			},
			Scrs: map[string]data.TransactionHandlerWithGasUsedAndFee{
				string(txHashes[1]): &outportcore.TransactionHandlerWithGasAndFee{
					TransactionHandler: &smartContractResult.SmartContractResult{
						Nonce: 0,
					},
				},
			},
			Rewards: map[string]data.TransactionHandlerWithGasUsedAndFee{
				string(txHashes[2]): &outportcore.TransactionHandlerWithGasAndFee{
					TransactionHandler: &rewardTx.RewardTx{
						Epoch: 0,
					},
				},
			},
		}

		odp.executionOrderHandler = &commonMocks.TxExecutionOrderHandlerStub{
			GetItemsCalled: func() [][]byte {
				return txHashes
			},
		}

		orderedHashes, numProcessed := odp.setExecutionOrderInTransactionPool(pool)
		require.Equal(t, 3, numProcessed)
		require.Equal(t, pool.Txs[string(txHashes[0])].GetExecutionOrder(), 0)
		require.Equal(t, pool.Scrs[string(txHashes[1])].GetExecutionOrder(), 1)
		require.Equal(t, pool.Rewards[string(txHashes[2])].GetExecutionOrder(), 2)
		require.Equal(t, len(txHashes), len(orderedHashes))
	})
	t.Run("transactions partially found (last txs in list) in pool txs, scrs or rewards", func(t *testing.T) {
		args := createArgOutportDataProvider()
		odp, _ := NewOutportDataProvider(args)
		txHashes := createRandTxHashes(10)
		pool := &outportcore.Pool{
			Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
				string(txHashes[7]): &outportcore.TransactionHandlerWithGasAndFee{
					TransactionHandler: &transaction.Transaction{
						Nonce: 0,
					},
				},
			},
			Scrs: map[string]data.TransactionHandlerWithGasUsedAndFee{
				string(txHashes[8]): &outportcore.TransactionHandlerWithGasAndFee{
					TransactionHandler: &smartContractResult.SmartContractResult{
						Nonce: 0,
					},
				},
			},
			Rewards: map[string]data.TransactionHandlerWithGasUsedAndFee{
				string(txHashes[9]): &outportcore.TransactionHandlerWithGasAndFee{
					TransactionHandler: &rewardTx.RewardTx{
						Epoch: 0,
					},
				},
			},
		}

		odp.executionOrderHandler = &commonMocks.TxExecutionOrderHandlerStub{
			GetItemsCalled: func() [][]byte {
				return txHashes
			},
		}

		orderedHashes, numProcessed := odp.setExecutionOrderInTransactionPool(pool)
		require.Equal(t, 3, numProcessed)
		require.Equal(t, pool.Txs[string(txHashes[7])].GetExecutionOrder(), 7)
		require.Equal(t, pool.Scrs[string(txHashes[8])].GetExecutionOrder(), 8)
		require.Equal(t, pool.Rewards[string(txHashes[9])].GetExecutionOrder(), 9)
		require.Equal(t, len(txHashes), len(orderedHashes))
	})
}

func Test_checkTxOrder(t *testing.T) {
	t.Parallel()

	t.Run("nil ordered txHashes with empty executed txs", func(t *testing.T) {
		err := checkTxOrder(nil, make(map[string]struct{}), 0)
		require.Nil(t, err)
	})
	t.Run("nil ordered txHashes with non-empty executed txs", func(t *testing.T) {
		err := checkTxOrder(nil, map[string]struct{}{"txHash": {}}, 0)
		require.True(t, errors.Is(err, ErrExecutedTxNotFoundInOrderedTxs))
	})
	t.Run("nil executed txHashes with empty ordered txHashes", func(t *testing.T) {
		err := checkTxOrder(make([][]byte, 0), nil, 0)
		require.Nil(t, err)
	})
	t.Run("nil executed txHashes with non-empty ordered txHashes", func(t *testing.T) {
		err := checkTxOrder([][]byte{{'a'}}, nil, 0)
		require.True(t, errors.Is(err, ErrOrderedTxNotFound))
	})
	t.Run("foundTxHashes < len(orderedTxHashes)", func(t *testing.T) {
		err := checkTxOrder(make([][]byte, 1), make(map[string]struct{}), 0)
		require.True(t, errors.Is(err, ErrOrderedTxNotFound))
	})
	t.Run("foundTxHashes == len(orderedTxHashes) and all txHashes found", func(t *testing.T) {
		orderedTxHashes := createRandTxHashes(10)
		executedTxHashes := make(map[string]struct{})
		for _, txHash := range orderedTxHashes {
			executedTxHashes[string(txHash)] = struct{}{}
		}

		err := checkTxOrder(orderedTxHashes, executedTxHashes, len(orderedTxHashes))
		require.Nil(t, err)
	})
	t.Run("foundTxHashes == len(orderedTxHashes) and not all executed txHashes found", func(t *testing.T) {
		orderedTxHashes := createRandTxHashes(10)
		executedTxHashes := make(map[string]struct{})
		for _, txHash := range orderedTxHashes {
			executedTxHashes[string(txHash)] = struct{}{}
		}
		executedTxHashes["newTxHash"] = struct{}{}
		err := checkTxOrder(orderedTxHashes, executedTxHashes, len(orderedTxHashes))
		require.True(t, errors.Is(err, ErrExecutedTxNotFoundInOrderedTxs))
	})
}

func Test_checkBodyTransactionsHaveOrder(t *testing.T) {
	t.Parallel()

	t.Run("nil orderedTxHashes", func(t *testing.T) {
		err := checkBodyTransactionsHaveOrder(nil, make(map[string]struct{}))
		require.Nil(t, err)
	})
	t.Run("empty orderedTxHashes", func(t *testing.T) {
		err := checkBodyTransactionsHaveOrder(make([][]byte, 0), make(map[string]struct{}))
		require.Nil(t, err)
	})

	t.Run("nil executedTxHashes", func(t *testing.T) {
		err := checkBodyTransactionsHaveOrder(createRandTxHashes(10), nil)
		require.Equal(t, ErrNilExecutedTxHashes, err)
	})
	t.Run("empty executedTxHashes", func(t *testing.T) {
		err := checkBodyTransactionsHaveOrder(createRandTxHashes(10), make(map[string]struct{}))
		require.Nil(t, err)
	})
	t.Run("executedTxHashes not found in orderedTxHashes", func(t *testing.T) {
		orderedTxHashes := createRandTxHashes(10)
		executedTxHashes := make(map[string]struct{})
		for _, txHash := range orderedTxHashes {
			executedTxHashes[string(txHash)] = struct{}{}
		}

		executedTxHashes["newTxHash"] = struct{}{}
		err := checkBodyTransactionsHaveOrder(orderedTxHashes, executedTxHashes)
		require.True(t, errors.Is(err, ErrExecutedTxNotFoundInOrderedTxs))
	})
}

func Test_setExecutionOrderIfFound(t *testing.T) {
	t.Parallel()

	t.Run("transaction not found", func(t *testing.T) {
		txHash := []byte("txHash")
		transactionHandlers := make(map[string]data.TransactionHandlerWithGasUsedAndFee)
		order := 0

		found := setExecutionOrderIfFound(txHash, transactionHandlers, order)
		require.False(t, found)
		require.Len(t, transactionHandlers, 0)
	})
	t.Run("transaction found", func(t *testing.T) {
		txHash := []byte("txHash")
		transactionHandlers := map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(txHash): &outportcore.TransactionHandlerWithGasAndFee{
				TransactionHandler: &transaction.Transaction{},
			},
		}

		res := setExecutionOrderIfFound(txHash, transactionHandlers, 0)
		require.True(t, res)
		require.Equal(t, 0, transactionHandlers[string(txHash)].GetExecutionOrder())
	})
}

func Test_collectExecutedTxHashes(t *testing.T) {
	t.Parallel()

	t.Run("mismatch between miniblocks and miniblock headers", func(t *testing.T) {
		mbs, mbHeaders := createMbsAndMbHeaders(1, 10)
		mbHeaders = mbHeaders[:len(mbHeaders)-1]
		body := &block.Body{
			MiniBlocks: mbs,
		}
		header := &block.Header{
			MiniBlockHeaders: mbHeaders,
		}

		_, err := collectExecutedTxHashes(body, header)
		require.Equal(t, ErrMiniBlocksHeadersMismatch, err)

		// more mbs
		mbs, mbHeaders = createMbsAndMbHeaders(100, 10)
		mbHeaders = mbHeaders[:len(mbHeaders)-1]
		body = &block.Body{
			MiniBlocks: mbs,
		}
		header = &block.Header{
			MiniBlockHeaders: mbHeaders,
		}

		_, err = collectExecutedTxHashes(body, header)
		require.Equal(t, ErrMiniBlocksHeadersMismatch, err)
	})
	t.Run("should work with empty miniBlocks", func(t *testing.T) {
		mbs, mbHeaders := createMbsAndMbHeaders(10, 0)
		body := &block.Body{
			MiniBlocks: mbs,
		}
		header := &block.Header{
			MiniBlockHeaders: mbHeaders,
		}

		collectedTxs, err := collectExecutedTxHashes(body, header)
		require.Nil(t, err)
		require.Equal(t, 0, len(collectedTxs))
	})
	t.Run("should work with nil miniBlocks", func(t *testing.T) {
		mbs := []*block.MiniBlock(nil)
		mbHeaders := []block.MiniBlockHeader(nil)
		body := &block.Body{
			MiniBlocks: mbs,
		}
		header := &block.Header{
			MiniBlockHeaders: mbHeaders,
		}

		collectedTxs, err := collectExecutedTxHashes(body, header)
		require.Nil(t, err)
		require.Equal(t, 0, len(collectedTxs))
	})
	t.Run("should work", func(t *testing.T) {
		mbs, mbHeaders := createMbsAndMbHeaders(10, 10)
		body := &block.Body{
			MiniBlocks: mbs,
		}
		header := &block.Header{
			MiniBlockHeaders: mbHeaders,
		}

		collectedTxs, err := collectExecutedTxHashes(body, header)
		require.Nil(t, err)
		require.Equal(t, 100, len(collectedTxs))
	})
}

func createMbsAndMbHeaders(numPairs int, numTxsPerMb int) ([]*block.MiniBlock, []block.MiniBlockHeader) {
	mbs := make([]*block.MiniBlock, numPairs)
	mbHeaders := make([]block.MiniBlockHeader, numPairs)

	for i := 0; i < numPairs; i++ {
		txHashes := createRandTxHashes(numTxsPerMb)
		mb, mbHeader := createMbAndMbHeader(txHashes)
		mbs[i] = mb
		mbHeaders[i] = mbHeader
	}

	return mbs, mbHeaders
}

func createMbAndMbHeader(txHashes [][]byte) (*block.MiniBlock, block.MiniBlockHeader) {
	mb := &block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 0,
		SenderShardID:   1,
	}

	mbHeader := block.MiniBlockHeader{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxCount:         uint32(len(txHashes)),
	}

	return mb, mbHeader
}

func createRandomByteArray(size uint32) []byte {
	buff := make([]byte, size)
	_, _ = rand.Read(buff)

	return buff
}

func createRandTxHashes(numTxHashes int) [][]byte {
	txHashes := make([][]byte, numTxHashes)
	for i := 0; i < numTxHashes; i++ {
		txHashes[i] = []byte(fmt.Sprintf("txHash_%s", createRandomByteArray(10)))
	}

	return txHashes
}
