package executionOrder

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	processOut "github.com/multiversx/mx-chain-go/outport/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func newArgStorer() ArgSorter {
	return ArgSorter{
		Hasher:     testscommon.KeccakMock{},
		Marshaller: testscommon.MarshalizerMock{},
		MbsStorer:  testscommon.CreateMemUnit(),
		EnableEpochsHandler: &testscommon.EnableEpochsHandlerStub{
			IsFrontRunningProtectionFlagEnabledField: true,
		},
	}
}

func TestNewSorter(t *testing.T) {
	t.Parallel()

	arg := newArgStorer()
	arg.Hasher = nil
	s, err := NewSorter(arg)
	require.Equal(t, process.ErrNilHasher, err)
	require.Nil(t, s)

	arg = newArgStorer()
	arg.Marshaller = nil
	s, err = NewSorter(arg)
	require.Equal(t, process.ErrNilMarshalizer, err)
	require.Nil(t, s)

	arg = newArgStorer()
	arg.MbsStorer = nil
	s, err = NewSorter(arg)
	require.Equal(t, processOut.ErrNilStorer, err)
	require.Nil(t, s)

	arg = newArgStorer()
	arg.EnableEpochsHandler = nil
	s, err = NewSorter(arg)
	require.Equal(t, processOut.ErrNilEnableEpochsHandler, err)
	require.Nil(t, s)

	arg = newArgStorer()
	s, err = NewSorter(arg)
	require.Nil(t, err)
	require.NotNil(t, s)
}

func TestAddExecutionOrderInTransactionPool(t *testing.T) {
	t.Parallel()

	randomness := "randomness"
	hasher := &testscommon.HasherStub{
		ComputeCalled: func(s string) []byte {
			if s == randomness {
				return []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
			}

			return []byte(s)
		},
	}

	arg := newArgStorer()
	arg.Hasher = hasher
	s, _ := NewSorter(arg)

	header := &block.Header{
		PrevRandSeed: []byte(randomness),
		ShardID:      1,
	}

	txHashToMe, txHashFromMe, invalidTxHash, rewardTxHash := []byte("toMe"), []byte("fromMe"), []byte("invalid"), []byte("reward")
	scrHashToMe, scrHashFromMe, scrHashIntra := []byte("scrHashToMe"), []byte("scrHashFromMe"), []byte("scrHashIntra")
	blockBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				SenderShardID:   0,
				ReceiverShardID: 1,
				Type:            block.TxBlock,
				TxHashes:        [][]byte{txHashToMe},
			},
			{
				SenderShardID:   1,
				ReceiverShardID: 0,
				Type:            block.TxBlock,
				TxHashes:        [][]byte{txHashFromMe},
			},
			{
				SenderShardID:   1,
				ReceiverShardID: 1,
				Type:            block.InvalidBlock,
				TxHashes:        [][]byte{invalidTxHash},
			},
			{
				SenderShardID:   1,
				ReceiverShardID: 2,
				Type:            block.SmartContractResultBlock,
				TxHashes:        [][]byte{scrHashFromMe},
			},
			{
				SenderShardID:   2,
				ReceiverShardID: 1,
				Type:            block.SmartContractResultBlock,
				TxHashes:        [][]byte{scrHashToMe},
			},
			{
				SenderShardID:   core.MetachainShardId,
				ReceiverShardID: 1,
				Type:            block.RewardsBlock,
				TxHashes:        [][]byte{rewardTxHash},
			},
		},
	}

	pool := &outport.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(txHashToMe):   &outport.TransactionHandlerWithGasAndFee{TransactionHandler: &transaction.Transaction{Nonce: 1}},
			string(txHashFromMe): &outport.TransactionHandlerWithGasAndFee{TransactionHandler: &transaction.Transaction{Nonce: 2}},
		},
		Scrs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(scrHashToMe): &outport.TransactionHandlerWithGasAndFee{TransactionHandler: &smartContractResult.SmartContractResult{Nonce: 3}},
			string(scrHashFromMe): &outport.TransactionHandlerWithGasAndFee{TransactionHandler: &smartContractResult.SmartContractResult{
				Nonce:          4,
				OriginalTxHash: txHashToMe,
			}},
			string(scrHashIntra): &outport.TransactionHandlerWithGasAndFee{TransactionHandler: &smartContractResult.SmartContractResult{
				Nonce:          0,
				OriginalTxHash: txHashToMe,
			}},
		},
		Rewards: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(rewardTxHash): &outport.TransactionHandlerWithGasAndFee{TransactionHandler: &rewardTx.RewardTx{}},
		},
		Invalid: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(invalidTxHash): &outport.TransactionHandlerWithGasAndFee{TransactionHandler: &transaction.Transaction{Nonce: 5}},
		},
		Receipts: map[string]data.TransactionHandlerWithGasUsedAndFee{},
		Logs:     nil,
	}

	_, _, err := s.PutExecutionOrderInTransactionPool(pool, header, blockBody, &block.Header{})
	require.Nil(t, err)

	require.Equal(t, &outport.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(txHashToMe): &outport.TransactionHandlerWithGasAndFee{
				TransactionHandler: &transaction.Transaction{Nonce: 1},
				ExecutionOrder:     0,
			},
			string(txHashFromMe): &outport.TransactionHandlerWithGasAndFee{
				TransactionHandler: &transaction.Transaction{Nonce: 2},
				ExecutionOrder:     3,
			},
		},
		Scrs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(scrHashToMe): &outport.TransactionHandlerWithGasAndFee{
				TransactionHandler: &smartContractResult.SmartContractResult{Nonce: 3},
				ExecutionOrder:     1,
			},
			string(scrHashFromMe): &outport.TransactionHandlerWithGasAndFee{
				TransactionHandler: &smartContractResult.SmartContractResult{
					Nonce:          4,
					OriginalTxHash: txHashToMe,
				},
				ExecutionOrder: 0,
			},
			string(scrHashIntra): &outport.TransactionHandlerWithGasAndFee{
				TransactionHandler: &smartContractResult.SmartContractResult{
					Nonce:          0,
					OriginalTxHash: txHashToMe,
				},
				ExecutionOrder: 0,
			},
		},
		Rewards: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(rewardTxHash): &outport.TransactionHandlerWithGasAndFee{
				TransactionHandler: &rewardTx.RewardTx{},
				ExecutionOrder:     2,
			},
		},
		Invalid: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(invalidTxHash): &outport.TransactionHandlerWithGasAndFee{
				TransactionHandler: &transaction.Transaction{Nonce: 5},
				ExecutionOrder:     4,
			},
		},
		Receipts: map[string]data.TransactionHandlerWithGasUsedAndFee{},
		Logs:     nil,
	}, pool)
}

func TestAddExecutionOrderInTransactionPoolFromMeTransactionAndScheduled(t *testing.T) {
	t.Parallel()

	randomness := "randomness"
	hasher := &testscommon.HasherStub{
		ComputeCalled: func(s string) []byte {
			if s == randomness {
				return []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
			}

			return []byte(s)
		},
	}

	arg := newArgStorer()
	arg.Hasher = hasher
	s, _ := NewSorter(arg)

	marshaller := &marshal.GogoProtoMarshalizer{}

	mbhr := &block.MiniBlockHeaderReserved{
		ExecutionType: block.ProcessingType(1),
	}

	mbhrBytes, _ := marshaller.Marshal(mbhr)
	header := &block.Header{
		PrevRandSeed: []byte(randomness),
		ShardID:      1,
	}

	firstTxHash, secondTxHash := []byte("t1"), []byte("t2")
	blockBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				SenderShardID:   1,
				ReceiverShardID: 1,
				Type:            block.TxBlock,
				Reserved:        mbhrBytes,
				TxHashes:        [][]byte{secondTxHash},
			},
			{
				SenderShardID:   1,
				ReceiverShardID: 1,
				Type:            block.TxBlock,
				TxHashes:        [][]byte{firstTxHash},
			},
		},
	}

	pool := &outport.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(firstTxHash):  &outport.TransactionHandlerWithGasAndFee{TransactionHandler: &transaction.Transaction{Nonce: 1}},
			string(secondTxHash): &outport.TransactionHandlerWithGasAndFee{TransactionHandler: &transaction.Transaction{Nonce: 2}},
		},
	}

	_, _, err := s.PutExecutionOrderInTransactionPool(pool, header, blockBody, &block.Header{})
	require.Nil(t, err)

	require.Equal(t, &outport.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(firstTxHash): &outport.TransactionHandlerWithGasAndFee{
				TransactionHandler: &transaction.Transaction{Nonce: 1},
				ExecutionOrder:     0,
			},
			string(secondTxHash): &outport.TransactionHandlerWithGasAndFee{
				TransactionHandler: &transaction.Transaction{Nonce: 2},
				ExecutionOrder:     1,
			},
		},
	}, pool)
}

func TestAddExecutionOrderInTransactionPoolFromMeTransactionAndScheduledInvalid(t *testing.T) {
	t.Parallel()

	randomness := "randomness"
	hasher := &testscommon.HasherStub{
		ComputeCalled: func(s string) []byte {
			if s == randomness {
				return []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
			}

			return []byte(s)
		},
	}

	marshaller := &marshal.GogoProtoMarshalizer{}

	arg := newArgStorer()
	arg.Hasher = hasher
	arg.Marshaller = marshaller
	scheduledInvalidTxHash := []byte("scheduledInvalidTx")
	scheduledTx := []byte("scheduledTx")
	arg.MbsStorer = &storage.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			mb := &block.MiniBlock{
				TxHashes: [][]byte{scheduledInvalidTxHash, scheduledTx},
			}
			return marshaller.Marshal(mb)
		},
	}
	s, _ := NewSorter(arg)

	mbhr := &block.MiniBlockHeaderReserved{
		ExecutionType: block.ProcessingType(1),
	}

	scheduledMBHash := []byte("scheduled")
	scrHash := []byte("scrHash")
	mbhrBytes, _ := marshaller.Marshal(mbhr)
	header := &block.Header{
		PrevRandSeed: []byte(randomness),
		ShardID:      1,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Type: block.InvalidBlock,
			},
		},
	}
	prevHeader := &block.Header{
		PrevRandSeed: []byte(randomness),
		ShardID:      1,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Type:            block.TxBlock,
				SenderShardID:   1,
				ReceiverShardID: 1,
			},
			{
				Hash:            scheduledMBHash,
				Reserved:        mbhrBytes,
				SenderShardID:   1,
				ReceiverShardID: 1,
			},
		},
	}

	firstTxHash, secondTxHash := []byte("t1"), []byte("t2")
	blockBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				SenderShardID:   1,
				ReceiverShardID: 1,
				Type:            block.TxBlock,
				TxHashes:        [][]byte{secondTxHash},
			},
			{
				SenderShardID:   1,
				ReceiverShardID: 1,
				Type:            block.InvalidBlock,
				TxHashes:        [][]byte{firstTxHash, scheduledInvalidTxHash},
			},
			{
				SenderShardID:   1,
				ReceiverShardID: 1,
				Type:            block.SmartContractResultBlock,
				TxHashes:        [][]byte{scrHash},
			},
		},
	}

	pool := &outport.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(secondTxHash): &outport.TransactionHandlerWithGasAndFee{TransactionHandler: &transaction.Transaction{Nonce: 2}},
		},
		Invalid: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(firstTxHash): &outport.TransactionHandlerWithGasAndFee{TransactionHandler: &transaction.Transaction{Nonce: 1}},
		},
		Scrs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(scrHash): &outport.TransactionHandlerWithGasAndFee{TransactionHandler: &smartContractResult.SmartContractResult{
				Nonce:          3,
				OriginalTxHash: scheduledTx,
			}},
		},
	}

	scrsHashes, invalidTxsHashes, err := s.PutExecutionOrderInTransactionPool(pool, header, blockBody, prevHeader)
	require.Nil(t, err)
	require.Equal(t, &outport.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(secondTxHash): &outport.TransactionHandlerWithGasAndFee{
				TransactionHandler: &transaction.Transaction{Nonce: 2},
				ExecutionOrder:     1,
			},
		},
		Invalid: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(firstTxHash): &outport.TransactionHandlerWithGasAndFee{
				TransactionHandler: &transaction.Transaction{Nonce: 1},
				ExecutionOrder:     0,
			},
		},
		Scrs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			string(scrHash): &outport.TransactionHandlerWithGasAndFee{
				TransactionHandler: &smartContractResult.SmartContractResult{
					Nonce:          3,
					OriginalTxHash: scheduledTx,
				},
				ExecutionOrder: 0,
			},
		},
	}, pool)

	require.Equal(t, []string{string(scrHash)}, scrsHashes)
	require.Equal(t, []string{string(scheduledInvalidTxHash)}, invalidTxsHashes)
}
