package firehose

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/firehose"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

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
		},
	}

	rewards := []*outportcore.TransactionHandlerWithGasAndFee{
		{
			TransactionHandler: &rewardTx.RewardTx{
				Epoch: 4,
				Value: big.NewInt(4),
			},
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
		Transactions: map[string]*firehose.TxWithFee{
			"txHash1": {
				Transaction: txs[0].TransactionHandler.(*transaction.Transaction),
				FeeInfo: &firehose.FeeInfo{
					GasUsed:        txs[0].GasUsed,
					Fee:            txs[0].Fee,
					InitialPaidFee: txs[0].InitialPaidFee,
				},
			},
			"txHash2": {
				Transaction: txs[1].TransactionHandler.(*transaction.Transaction),
				FeeInfo: &firehose.FeeInfo{
					GasUsed:        txs[1].GasUsed,
					Fee:            txs[1].Fee,
					InitialPaidFee: txs[1].InitialPaidFee,
				},
			},
		},
		SmartContractResult: map[string]*firehose.SCRWithFee{
			"txHash3": {
				SmartContractResult: scrs[0].TransactionHandler.(*smartContractResult.SmartContractResult),
				FeeInfo: &firehose.FeeInfo{
					GasUsed:        scrs[0].GasUsed,
					Fee:            scrs[0].Fee,
					InitialPaidFee: scrs[0].InitialPaidFee,
				},
			},
		},
		Rewards: map[string]*rewardTx.RewardTx{
			"txHash4": rewards[0].TransactionHandler.(*rewardTx.RewardTx),
		},
		InvalidTxs: map[string]*firehose.TxWithFee{
			"txHash5": {
				Transaction: invalidTxs[0].TransactionHandler.(*transaction.Transaction),
				FeeInfo: &firehose.FeeInfo{
					GasUsed:        invalidTxs[0].GasUsed,
					Fee:            invalidTxs[0].Fee,
					InitialPaidFee: invalidTxs[0].InitialPaidFee,
				},
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
