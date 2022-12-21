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
				GasUsed:        100,
				Fee:            big.NewInt(100),
				InitialPaidFee: big.NewInt(200),
			},
			TransactionHandler: &transaction.Transaction{
				Nonce: 1,
				Data:  []byte("data1"),
			},
		},
		{
			FeeInfo: outportcore.FeeInfo{
				GasUsed:        220,
				Fee:            big.NewInt(340),
				InitialPaidFee: big.NewInt(500),
			},
			TransactionHandler: &transaction.Transaction{
				Nonce: 2,
				Data:  []byte("data2"),
			},
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
			Scrs:     nil,
			Rewards:  nil,
			Invalid:  nil,
			Receipts: nil,
			Logs:     nil,
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
