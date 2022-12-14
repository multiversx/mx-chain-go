package executionOrder

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	processOut "github.com/ElrondNetwork/elrond-go/outport/process"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func newArgStorer() ArgSorter {
	return ArgSorter{
		Hasher:     testscommon.KeccakMock{},
		Marshaller: testscommon.MarshalizerMock{},
		MbsStorer:  testscommon.CreateMemUnit(),
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

	err := s.PutExecutionOrderInTransactionPool(pool, header, blockBody, &block.Header{})
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

	marshalizer := &marshal.GogoProtoMarshalizer{}

	mbhr := &block.MiniBlockHeaderReserved{
		ExecutionType: block.ProcessingType(1),
	}

	mbhrBytes, _ := marshalizer.Marshal(mbhr)
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

	err := s.PutExecutionOrderInTransactionPool(pool, header, blockBody, &block.Header{})
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
