package transactionsfee

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/outport/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/stretchr/testify/require"
)

func prepareMockArg() ArgTransactionsFeeProcessor {
	return ArgTransactionsFeeProcessor{
		Marshaller:         testscommon.MarshalizerMock{},
		TransactionsStorer: genericMocks.NewStorerMock(),
		ShardCoordinator:   &testscommon.ShardsCoordinatorMock{},
		TxFeeCalculator:    &mock.EconomicsHandlerMock{},
	}
}

func TestNewTransactionFeeProcessor(t *testing.T) {
	t.Parallel()

	arg := prepareMockArg()
	arg.Marshaller = nil
	_, err := NewTransactionsFeeProcessor(arg)
	require.Equal(t, ErrNilMarshaller, err)

	arg = prepareMockArg()
	arg.TransactionsStorer = nil
	_, err = NewTransactionsFeeProcessor(arg)
	require.Equal(t, ErrNilStorage, err)

	arg = prepareMockArg()
	arg.ShardCoordinator = nil
	_, err = NewTransactionsFeeProcessor(arg)
	require.Equal(t, ErrNilShardCoordinator, err)

	arg = prepareMockArg()
	arg.TxFeeCalculator = nil
	_, err = NewTransactionsFeeProcessor(arg)
	require.Equal(t, ErrNilTransactionFeeCalculator, err)

	arg = prepareMockArg()
	txsFeeProc, err := NewTransactionsFeeProcessor(arg)
	require.NotNil(t, txsFeeProc)
	require.Nil(t, err)
	require.False(t, txsFeeProc.IsInterfaceNil())
}

func TestPutFeeAndGasUsedTx1(t *testing.T) {
	t.Parallel()

	txHash := []byte("relayedTx")
	scrHash1 := []byte("scrHash1")
	scrWithRefund := []byte("scrWithRefund")
	refundValueBig, _ := big.NewInt(0).SetString("86271830000000", 10)
	initialTx := outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
		Nonce:    1196667,
		SndAddr:  []byte("erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"),
		RcvAddr:  []byte("erd14eyayfrvlrhzfrwg5zwleua25mkzgncggn35nvc6xhv5yxwml2es0f3dht"),
		GasLimit: 16610000,
		GasPrice: 1000000000,
		Data:     []byte("relayedTx@7b226e6f6e6365223a322c2276616c7565223a302c227265636569766572223a22414141414141414141414146414974673738352f736c73554148686b57334569624c6e47524b76496f4e4d3d222c2273656e646572223a22726b6e534a477a343769534e794b43642f504f717075776b5477684534306d7a476a585a51686e622b724d3d222c226761735072696365223a313030303030303030302c226761734c696d6974223a31353030303030302c2264617461223a22633246325a5546306447567a644746306157397551444668597a49314d6a5935596d51335a44497759324a6959544d31596d566c4f4459314d4464684f574e6a4e7a677a5a4755774f445a694e4445334e546b345a54517a59544e6b5a6a566a593245795a5468684d6a6c414d6a51344e54677a4d574e6d4d5445304d54566d596a41354d6a63774e4451324e5755324e7a597a59574d314f4445345a5467314e4751345957526d4e54417a596a63354d6a6c6b4f54526c4e6d49794e6a49775a673d3d222c22636861696e4944223a224d513d3d222c2276657273696f6e223a312c227369676e6174757265223a225239462b34546352415a386d7771324559303163596c337662716c46657176387a76474a775a6833594d4f556b4234643451574e66376744626c484832576b71614a76614845744356617049713365356562384e41773d3d227d"),
		Value:    big.NewInt(0),
	}, 0, big.NewInt(0))

	scr1 := &smartContractResult.SmartContractResult{
		Nonce:          2,
		GasPrice:       1000000000,
		GasLimit:       14732500,
		SndAddr:        []byte("erd14eyayfrvlrhzfrwg5zwleua25mkzgncggn35nvc6xhv5yxwml2es0f3dht"),
		RcvAddr:        []byte("erd1qqqqqqqqqqqqqpgq3dswlnnlkfd3gqrcv3dhzgnvh8ryf27g5rfsecnn2s"),
		Data:           []byte("saveAttestation@1ac25269bd7d20cbba35bee86507a9cc783de086b417598e43a3df5cca2e8a29@2485831cf11415fb092704465e6763ac5818e854d8adf503b7929d94e6b2620f"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	pool := &outportcore.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(txHash): initialTx,
		},
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			"wrong":          outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{}, 0, big.NewInt(0)),
			string(scrHash1): outportcore.NewTransactionHandlerWithGasAndFee(scr1, 0, big.NewInt(0)),
			string(scrWithRefund): outportcore.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
				Nonce:          3,
				SndAddr:        []byte("erd1qqqqqqqqqqqqqpgq3dswlnnlkfd3gqrcv3dhzgnvh8ryf27g5rfsecnn2s"),
				RcvAddr:        []byte("erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"),
				PrevTxHash:     []byte("f639cb7a0231191e04ec19dcb1359bd93a03fe8dc4a28a80d00835c5d1c988f8"),
				OriginalTxHash: txHash,
				Value:          refundValueBig,
				Data:           []byte(""),
				ReturnMessage:  []byte("gas refund for relayer"),
			}, 0, big.NewInt(0)),
		},
	}

	arg := prepareMockArg()
	txsFeeProc, err := NewTransactionsFeeProcessor(arg)
	require.NotNil(t, txsFeeProc)
	require.Nil(t, err)

	err = txsFeeProc.PutFeeAndGasUsed(pool)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(1673728170000000), initialTx.GetFee())
	require.Equal(t, uint64(7982817), initialTx.GetGasUsed())
	require.Equal(t, "1760000000000000", initialTx.GetInitialPaidFee().String())
}

func TestPutFeeAndGasUsedScrNoTx(t *testing.T) {
	t.Parallel()

	txHash := []byte("relayedTx")
	scrWithRefund := []byte("scrWithRefund")

	refundValueBig, _ := big.NewInt(0).SetString("226498540000000", 10)

	scr := outportcore.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
		Nonce:          3,
		SndAddr:        []byte("erd1qqqqqqqqqqqqqpgq3dswlnnlkfd3gqrcv3dhzgnvh8ryf27g5rfsecnn2s"),
		RcvAddr:        []byte("erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"),
		PrevTxHash:     []byte("f639cb7a0231191e04ec19dcb1359bd93a03fe8dc4a28a80d00835c5d1c988f8"),
		OriginalTxHash: txHash,
		Value:          refundValueBig,
		Data:           []byte(""),
		ReturnMessage:  []byte("gas refund for relayer"),
	}, 0, big.NewInt(0))

	pool := &outportcore.Pool{
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			"wrong":               outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{}, 0, big.NewInt(0)),
			string(scrWithRefund): scr,
		},
	}

	arg := prepareMockArg()

	initialTx := &transaction.Transaction{
		GasLimit: 30000000,
		GasPrice: 1000000000,
	}
	txBytes, _ := arg.Marshaller.Marshal(initialTx)

	_ = arg.TransactionsStorer.Put(txHash, txBytes)

	txsFeeProc, err := NewTransactionsFeeProcessor(arg)
	require.NotNil(t, txsFeeProc)
	require.Nil(t, err)

	err = txsFeeProc.PutFeeAndGasUsed(pool)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(123001460000000), scr.GetFee())
	require.Equal(t, uint64(7350146), scr.GetGasUsed())
}

func TestPutFeeAndGasUsedInvalidTxs(t *testing.T) {
	t.Parallel()

	tx := outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
		GasLimit: 30000000,
		GasPrice: 1000000000,
	}, 0, big.NewInt(0))

	pool := &outportcore.Pool{
		Invalid: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			"tx": tx,
		},
	}

	arg := prepareMockArg()
	txsFeeProc, err := NewTransactionsFeeProcessor(arg)
	require.NotNil(t, txsFeeProc)
	require.Nil(t, err)

	err = txsFeeProc.PutFeeAndGasUsed(pool)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(349500000000000), tx.GetFee())
	require.Equal(t, tx.GetGasLimit(), tx.GetGasUsed())
}

func TestPutFeeAndGasUsedLogWithErrorAndInformative(t *testing.T) {
	t.Parallel()

	tx1Hash := "h1"
	tx1 := outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
		GasLimit: 30000000,
		GasPrice: 1000000000,
	}, 0, big.NewInt(0))

	tx2Hash := "h2"
	tx2 := outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
		GasLimit: 50000000,
		GasPrice: 1000000000,
	}, 0, big.NewInt(0))

	pool := &outportcore.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			tx1Hash: tx1,
			tx2Hash: tx2,
			"t3":    outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{}, 0, big.NewInt(0)),
		},
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Identifier: []byte("ignore"),
						},
					},
				},
			},
			{
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Identifier: []byte("ignore"),
						},
						{
							Identifier: []byte(core.SignalErrorOperation),
						},
					},
				},
				TxHash: tx1Hash,
			},
			{
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Identifier: []byte(core.WriteLogIdentifier),
						},
					},
				},
				TxHash: tx2Hash,
			},
		},
	}

	arg := prepareMockArg()
	txsFeeProc, err := NewTransactionsFeeProcessor(arg)
	require.NotNil(t, txsFeeProc)
	require.Nil(t, err)

	err = txsFeeProc.PutFeeAndGasUsed(pool)
	require.Nil(t, err)

	require.Equal(t, tx1.GetGasLimit(), tx1.GetGasUsed())
	require.Equal(t, tx2.GetGasLimit(), tx2.GetGasUsed())
}
