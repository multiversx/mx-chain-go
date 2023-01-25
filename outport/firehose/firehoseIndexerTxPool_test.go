package firehose

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/firehose"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestFirehoseIndexer_SaveBlockNilTxPool(t *testing.T) {
	t.Parallel()

	fi, _ := NewFirehoseIndexer(&testscommon.IoWriterStub{})

	headerHash := []byte("hash")
	argsSaveBlock := &outportcore.ArgsSaveBlockData{
		HeaderHash:       headerHash,
		Header:           &block.Header{},
		TransactionsPool: nil,
	}
	err := fi.SaveBlock(argsSaveBlock)
	require.True(t, strings.Contains(err.Error(), errNilTxPool.Error()))
	require.True(t, strings.Contains(err.Error(), hex.EncodeToString(headerHash)))
}

func TestFirehoseIndexer_SaveBlockTxPool(t *testing.T) {
	t.Parallel()

	protoMarshaller := &marshal.GogoProtoMarshalizer{}

	shardHeaderV1 := &block.Header{
		Nonce:     2,
		PrevHash:  []byte("prevHashV1"),
		TimeStamp: 200,
	}
	marshalledHeader, err := protoMarshaller.Marshal(shardHeaderV1)
	require.Nil(t, err)

	headerHashShardV1 := []byte("headerHashShardV1")

	txs := []*outportcore.TransactionHandlerWithGasAndFee{
		{
			FeeInfo: outportcore.FeeInfo{
				GasUsed:        111,
				Fee:            big.NewInt(111),
				InitialPaidFee: big.NewInt(111),
			},
			TransactionHandler: &transaction.Transaction{
				Nonce: 1,
				Data:  []byte("data1"),
			},
			ExecutionOrder: 1,
		},
		{
			FeeInfo: outportcore.FeeInfo{
				GasUsed:        222,
				Fee:            big.NewInt(222),
				InitialPaidFee: big.NewInt(222),
			},
			TransactionHandler: &transaction.Transaction{
				Nonce: 2,
				Data:  []byte("data2"),
			},
			ExecutionOrder: 2,
		},
	}

	scrs := []*outportcore.TransactionHandlerWithGasAndFee{
		{
			FeeInfo: outportcore.FeeInfo{
				GasUsed:        333,
				Fee:            big.NewInt(333),
				InitialPaidFee: big.NewInt(333),
			},
			TransactionHandler: &smartContractResult.SmartContractResult{
				Nonce: 3,
				Data:  []byte("data3"),
			},
			ExecutionOrder: 3,
		},
	}

	rewards := []*outportcore.TransactionHandlerWithGasAndFee{
		{
			TransactionHandler: &rewardTx.RewardTx{
				Epoch: 4,
				Value: big.NewInt(4),
			},
			ExecutionOrder: 4,
		},
	}

	invalidTxs := []*outportcore.TransactionHandlerWithGasAndFee{
		{
			FeeInfo: outportcore.FeeInfo{
				GasUsed:        444,
				Fee:            big.NewInt(444),
				InitialPaidFee: big.NewInt(444),
			},
			TransactionHandler: &transaction.Transaction{
				Nonce: 4,
				Data:  []byte("data4"),
			},
			ExecutionOrder: 5,
		},
	}

	receipts := []*outportcore.TransactionHandlerWithGasAndFee{
		{
			TransactionHandler: &receipt.Receipt{
				Value: big.NewInt(44),
				Data:  []byte("data5"),
			},
		},
		{
			TransactionHandler: &receipt.Receipt{
				SndAddr: []byte("sndr"),
			},
		},
	}

	logs := []*data.LogData{
		{
			LogHandler: &transaction.Log{
				Address: []byte("addr"),
				Events: []*transaction.Event{
					{
						Identifier: []byte("id1"),
					},
					{
						Identifier: []byte("id2"),
					},
				},
			},
			TxHash: "txHash8",
		},
	}

	argsSaveBlock := &outportcore.ArgsSaveBlockData{
		HeaderHash: headerHashShardV1,
		Header:     shardHeaderV1,
		TransactionsPool: &outportcore.Pool{
			Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
				"txHash1": txs[0],
				"txHash2": txs[1],
			},
			Scrs: map[string]data.TransactionHandlerWithGasUsedAndFee{
				"txHash3": scrs[0],
			},
			Rewards: map[string]data.TransactionHandlerWithGasUsedAndFee{
				"txHash4": rewards[0],
			},

			Invalid: map[string]data.TransactionHandlerWithGasUsedAndFee{
				"txHash5": invalidTxs[0],
			},
			Receipts: map[string]data.TransactionHandlerWithGasUsedAndFee{
				"txHash6": receipts[0],
				"txHash7": receipts[1],
			},
			Logs: logs,
		},
	}

	firehoseBlock := &firehose.FirehoseBlock{
		HeaderHash:  headerHashShardV1,
		HeaderType:  string(core.ShardHeaderV1),
		HeaderBytes: marshalledHeader,
		Transactions: map[string]*firehose.TxInfo{
			"txHash1": {
				Transaction: txs[0].TransactionHandler.(*transaction.Transaction),
				FeeInfo: &firehose.FeeInfo{
					GasUsed:        txs[0].GasUsed,
					Fee:            txs[0].Fee,
					InitialPaidFee: txs[0].InitialPaidFee,
				},
				ExecutionOrder: 1,
			},
			"txHash2": {
				Transaction: txs[1].TransactionHandler.(*transaction.Transaction),
				FeeInfo: &firehose.FeeInfo{
					GasUsed:        txs[1].GasUsed,
					Fee:            txs[1].Fee,
					InitialPaidFee: txs[1].InitialPaidFee,
				},
				ExecutionOrder: 2,
			},
		},
		SmartContractResults: map[string]*firehose.SCRInfo{
			"txHash3": {
				SmartContractResult: scrs[0].TransactionHandler.(*smartContractResult.SmartContractResult),
				FeeInfo: &firehose.FeeInfo{
					GasUsed:        scrs[0].GasUsed,
					Fee:            scrs[0].Fee,
					InitialPaidFee: scrs[0].InitialPaidFee,
				},
				ExecutionOrder: 3,
			},
		},
		Rewards: map[string]*firehose.RewardInfo{
			"txHash4": {
				Reward:         rewards[0].TransactionHandler.(*rewardTx.RewardTx),
				ExecutionOrder: 4,
			},
		},
		InvalidTxs: map[string]*firehose.TxInfo{
			"txHash5": {
				Transaction: invalidTxs[0].TransactionHandler.(*transaction.Transaction),
				FeeInfo: &firehose.FeeInfo{
					GasUsed:        invalidTxs[0].GasUsed,
					Fee:            invalidTxs[0].Fee,
					InitialPaidFee: invalidTxs[0].InitialPaidFee,
				},
				ExecutionOrder: 5,
			},
		},
		Receipts: map[string]*receipt.Receipt{
			"txHash6": receipts[0].TransactionHandler.(*receipt.Receipt),
			"txHash7": receipts[1].TransactionHandler.(*receipt.Receipt),
		},
		Logs: map[string]*transaction.Log{
			"txHash8": logs[0].LogHandler.(*transaction.Log),
		},
	}
	marshalledFirehoseBlock, err := protoMarshaller.Marshal(firehoseBlock)
	require.Nil(t, err)

	ioWriterCalledCt := 0
	ioWriter := &testscommon.IoWriterStub{
		WriteCalled: func(p []byte) (n int, err error) {
			ioWriterCalledCt++
			switch ioWriterCalledCt {
			case 1:
				require.Equal(t, []byte("FIRE BLOCK_BEGIN 2\n"), p)
			case 2:

				require.Equal(t, []byte(fmt.Sprintf("FIRE BLOCK_END 2 %s 200 %x\n",
					hex.EncodeToString(shardHeaderV1.PrevHash),
					marshalledFirehoseBlock)), p)
			default:
				require.Fail(t, "should not write again")
			}
			return 0, nil
		},
	}

	fi, _ := NewFirehoseIndexer(ioWriter)
	err = fi.SaveBlock(argsSaveBlock)
	require.Nil(t, err)
}
