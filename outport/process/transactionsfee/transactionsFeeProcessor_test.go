package transactionsfee

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/outport/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

var pubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, "erd")

func prepareMockArg() ArgTransactionsFeeProcessor {
	return ArgTransactionsFeeProcessor{
		Marshaller:         testscommon.MarshalizerMock{},
		TransactionsStorer: genericMocks.NewStorerMock(),
		ShardCoordinator:   &testscommon.ShardsCoordinatorMock{},
		TxFeeCalculator:    &mock.EconomicsHandlerMock{},
		PubKeyConverter:    pubKeyConverter,
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
	initialTx := &outportcore.TxInfo{
		Transaction: &transaction.Transaction{
			Nonce:    1196667,
			SndAddr:  []byte("erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"),
			RcvAddr:  []byte("erd14eyayfrvlrhzfrwg5zwleua25mkzgncggn35nvc6xhv5yxwml2es0f3dht"),
			GasLimit: 16610000,
			GasPrice: 1000000000,
			Data:     []byte("relayedTx@7b226e6f6e6365223a322c2276616c7565223a302c227265636569766572223a22414141414141414141414146414974673738352f736c73554148686b57334569624c6e47524b76496f4e4d3d222c2273656e646572223a22726b6e534a477a343769534e794b43642f504f717075776b5477684534306d7a476a585a51686e622b724d3d222c226761735072696365223a313030303030303030302c226761734c696d6974223a31353030303030302c2264617461223a22633246325a5546306447567a644746306157397551444668597a49314d6a5935596d51335a44497759324a6959544d31596d566c4f4459314d4464684f574e6a4e7a677a5a4755774f445a694e4445334e546b345a54517a59544e6b5a6a566a593245795a5468684d6a6c414d6a51344e54677a4d574e6d4d5445304d54566d596a41354d6a63774e4451324e5755324e7a597a59574d314f4445345a5467314e4751345957526d4e54417a596a63354d6a6c6b4f54526c4e6d49794e6a49775a673d3d222c22636861696e4944223a224d513d3d222c2276657273696f6e223a312c227369676e6174757265223a225239462b34546352415a386d7771324559303163596c337662716c46657176387a76474a775a6833594d4f556b4234643451574e66376744626c484832576b71614a76614845744356617049713365356562384e41773d3d227d"),
			Value:    big.NewInt(0),
		},
		FeeInfo: &outportcore.FeeInfo{Fee: big.NewInt(0)},
	}

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

	pool := &outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			hex.EncodeToString(txHash): initialTx,
		},
		SmartContractResults: map[string]*outportcore.SCRInfo{
			hex.EncodeToString(scrHash1): {
				SmartContractResult: scr1,
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
			hex.EncodeToString(scrWithRefund): {
				SmartContractResult: &smartContractResult.SmartContractResult{
					Nonce:          3,
					SndAddr:        []byte("erd1qqqqqqqqqqqqqpgq3dswlnnlkfd3gqrcv3dhzgnvh8ryf27g5rfsecnn2s"),
					RcvAddr:        []byte("erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"),
					PrevTxHash:     []byte("f639cb7a0231191e04ec19dcb1359bd93a03fe8dc4a28a80d00835c5d1c988f8"),
					OriginalTxHash: txHash,
					Value:          refundValueBig,
					Data:           []byte(""),
					ReturnMessage:  []byte("gas refund for relayer"),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
	}

	arg := prepareMockArg()
	txsFeeProc, err := NewTransactionsFeeProcessor(arg)
	require.NotNil(t, txsFeeProc)
	require.Nil(t, err)

	err = txsFeeProc.PutFeeAndGasUsed(pool)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(1673728170000000), initialTx.GetFeeInfo().GetFee())
	require.Equal(t, uint64(7982817), initialTx.GetFeeInfo().GetGasUsed())
	require.Equal(t, "1760000000000000", initialTx.GetFeeInfo().GetInitialPaidFee().String())
}

func TestPutFeeAndGasUsedScrNoTx(t *testing.T) {
	t.Parallel()

	txHash := []byte("relayedTx")
	scrWithRefund := []byte("scrWithRefund")

	refundValueBig, _ := big.NewInt(0).SetString("226498540000000", 10)

	scr := &outportcore.SCRInfo{
		SmartContractResult: &smartContractResult.SmartContractResult{
			Nonce:          3,
			SndAddr:        []byte("erd1qqqqqqqqqqqqqpgq3dswlnnlkfd3gqrcv3dhzgnvh8ryf27g5rfsecnn2s"),
			RcvAddr:        []byte("erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"),
			PrevTxHash:     []byte("f639cb7a0231191e04ec19dcb1359bd93a03fe8dc4a28a80d00835c5d1c988f8"),
			OriginalTxHash: txHash,
			Value:          refundValueBig,
			Data:           []byte(""),
			ReturnMessage:  []byte("gas refund for relayer"),
		},
		FeeInfo: &outportcore.FeeInfo{Fee: big.NewInt(0)},
	}

	pool := &outportcore.TransactionPool{
		SmartContractResults: map[string]*outportcore.SCRInfo{
			string(scrWithRefund): scr,
		},
	}

	arg := prepareMockArg()

	initialTx := &transaction.Transaction{
		GasLimit: 30000000,
		GasPrice: 1000000000,
		SndAddr:  []byte("erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"),
	}
	txBytes, _ := arg.Marshaller.Marshal(initialTx)

	_ = arg.TransactionsStorer.Put(txHash, txBytes)

	txsFeeProc, err := NewTransactionsFeeProcessor(arg)
	require.NotNil(t, txsFeeProc)
	require.Nil(t, err)

	err = txsFeeProc.PutFeeAndGasUsed(pool)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(123001460000000), scr.GetFeeInfo().GetFee())
	require.Equal(t, uint64(7350146), scr.GetFeeInfo().GetGasUsed())
}

func TestPutFeeAndGasUsedInvalidTxs(t *testing.T) {
	t.Parallel()

	tx := &outportcore.TxInfo{
		Transaction: &transaction.Transaction{
			GasLimit: 30000000,
			GasPrice: 1000000000,
		},
		FeeInfo: &outportcore.FeeInfo{Fee: big.NewInt(0)},
	}

	pool := &outportcore.TransactionPool{
		InvalidTxs: map[string]*outportcore.TxInfo{
			"tx": tx,
		},
	}

	arg := prepareMockArg()
	txsFeeProc, err := NewTransactionsFeeProcessor(arg)
	require.NotNil(t, txsFeeProc)
	require.Nil(t, err)

	err = txsFeeProc.PutFeeAndGasUsed(pool)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(349500000000000), tx.GetFeeInfo().GetFee())
	require.Equal(t, tx.GetTxHandler().GetGasLimit(), tx.GetFeeInfo().GetGasUsed())
}

func TestPutFeeAndGasUsedLogWithErrorAndInformative(t *testing.T) {
	t.Parallel()

	tx1Hash := "h1"
	tx1 := &outportcore.TxInfo{
		Transaction: &transaction.Transaction{
			GasLimit: 30000000,
			GasPrice: 1000000000,
		},
		FeeInfo: &outportcore.FeeInfo{Fee: big.NewInt(0)},
	}

	tx2Hash := "h2"
	tx2 := &outportcore.TxInfo{
		Transaction: &transaction.Transaction{
			GasLimit: 50000000,
			GasPrice: 1000000000,
		},
		FeeInfo: &outportcore.FeeInfo{Fee: big.NewInt(0)},
	}

	pool := &outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			tx1Hash: tx1,
			tx2Hash: tx2,
			"t3":    {Transaction: &transaction.Transaction{}, FeeInfo: &outportcore.FeeInfo{Fee: big.NewInt(0)}}},
		Logs: []*outportcore.LogData{
			{
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Identifier: []byte("ignore"),
						},
					},
				},
				TxHash: "hhh",
			},
			{
				Log: &transaction.Log{
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
				Log: &transaction.Log{
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

	require.Equal(t, tx1.GetTxHandler().GetGasLimit(), tx1.GetFeeInfo().GetGasUsed())
	require.Equal(t, tx2.GetTxHandler().GetGasLimit(), tx2.GetFeeInfo().GetGasUsed())
}

func TestPutFeeAndGasUsedWrongRelayedTx(t *testing.T) {
	t.Parallel()

	txHash := []byte("relayedTx")
	scrHash1 := []byte("scrHash1")
	initialTx := &outportcore.TxInfo{
		Transaction: &transaction.Transaction{
			Nonce:    1011,
			SndAddr:  []byte("erd1dglncxk6sl9a3xumj78n6z2xux4ghp5c92cstv5zsn56tjgtdwpsk46qrs"),
			RcvAddr:  []byte("erd1xlrw5j482m3fwl72fsu9saj984rxqdrjd860e02tcz0qakvqrp6q2pjqgg"),
			GasLimit: 550000000,
			GasPrice: 1000000000,
			Data:     []byte("relayedTxV2@000000000000000005005eaf5311cedc6fa17f08f33e156926f8f3816d8ed8dc@06e2@7472616e73666572546f6b656e4064633132346163313733323937623836623936316362636663363339326231643130303533326533336530663933313838373634396336613935636236633931403031@ba26daf1353b8fa62d183b7d7df8db48846ea982a0cb26450b703e16720c77b9d7d4e47b652d270b160ae6866ca7b04aae38ca83a58ce508bf660db07d5b6401"),
			Value:    big.NewInt(0),
		},
		FeeInfo: &outportcore.FeeInfo{Fee: big.NewInt(0)},
	}

	scr1 := &outportcore.SCRInfo{
		SmartContractResult: &smartContractResult.SmartContractResult{
			Nonce:          1011,
			SndAddr:        []byte("erd1xlrw5j482m3fwl72fsu9saj984rxqdrjd860e02tcz0qakvqrp6q2pjqgg"),
			RcvAddr:        []byte("erd1dglncxk6sl9a3xumj78n6z2xux4ghp5c92cstv5zsn56tjgtdwpsk46qrs"),
			PrevTxHash:     txHash,
			OriginalTxHash: txHash,
			ReturnMessage:  []byte("higher nonce in transaction"),
		},
		FeeInfo: &outportcore.FeeInfo{Fee: big.NewInt(0)},
	}

	pool := &outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			hex.EncodeToString(txHash): initialTx,
		},
		SmartContractResults: map[string]*outportcore.SCRInfo{
			hex.EncodeToString(scrHash1): scr1,
		},
	}

	arg := prepareMockArg()
	txsFeeProc, err := NewTransactionsFeeProcessor(arg)
	require.NotNil(t, txsFeeProc)
	require.Nil(t, err)

	err = txsFeeProc.PutFeeAndGasUsed(pool)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(6103405000000000), initialTx.GetFeeInfo().GetFee())
	require.Equal(t, uint64(550000000), initialTx.GetFeeInfo().GetGasUsed())
	require.Equal(t, "6103405000000000", initialTx.GetFeeInfo().GetInitialPaidFee().String())
}

func TestPutFeeAndGasUsedESDTWithScCall(t *testing.T) {
	t.Parallel()

	txHash := []byte("tx")
	tx := &outportcore.TxInfo{
		Transaction: &transaction.Transaction{
			Nonce:    1011,
			SndAddr:  silentDecodeAddress("erd1dglncxk6sl9a3xumj78n6z2xux4ghp5c92cstv5zsn56tjgtdwpsk46qrs"),
			RcvAddr:  silentDecodeAddress("erd1dglncxk6sl9a3xumj78n6z2xux4ghp5c92cstv5zsn56tjgtdwpsk46qrs"),
			GasLimit: 55_000_000,
			GasPrice: 1000000000,
			Data:     []byte("ESDTNFTTransfer@434f572d636434363364@080c@01@00000000000000000500d3b28828d62052124f07dcd50ed31b0825f60eee1526@616363657074476c6f62616c4f66666572@c3e5"),
			Value:    big.NewInt(0),
		},
		FeeInfo: &outportcore.FeeInfo{Fee: big.NewInt(0)},
	}

	pool := &outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			string(txHash): tx,
		},
	}

	arg := prepareMockArg()
	txsFeeProc, err := NewTransactionsFeeProcessor(arg)
	require.NotNil(t, txsFeeProc)
	require.Nil(t, err)

	err = txsFeeProc.PutFeeAndGasUsed(pool)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(820765000000000), tx.GetFeeInfo().GetFee())
	require.Equal(t, uint64(55_000_000), tx.GetFeeInfo().GetGasUsed())
	require.Equal(t, "820765000000000", tx.GetFeeInfo().GetInitialPaidFee().String())
}

func silentDecodeAddress(address string) []byte {
	decoded, _ := pubKeyConverter.Decode(address)
	return decoded
}

func TestPutFeeAndGasUsedScrWithRefundNoTx(t *testing.T) {
	t.Parallel()

	_ = logger.SetLogLevel("*:TRACE")

	txHash := []byte("relayedTx")
	scrWithRefund := []byte("scrWithRefund")

	refundValueBig, _ := big.NewInt(0).SetString("226498540000000", 10)

	scr := &outportcore.SCRInfo{
		SmartContractResult: &smartContractResult.SmartContractResult{
			Nonce:          3,
			SndAddr:        []byte("erd1qqqqqqqqqqqqqpgq3dswlnnlkfd3gqrcv3dhzgnvh8ryf27g5rfsecnn2s"),
			RcvAddr:        []byte("erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"),
			PrevTxHash:     []byte("f639cb7a0231191e04ec19dcb1359bd93a03fe8dc4a28a80d00835c5d1c988f8"),
			OriginalTxHash: txHash,
			Value:          refundValueBig,
			Data:           []byte("@ok"),
		},
		FeeInfo: &outportcore.FeeInfo{Fee: big.NewInt(0)},
	}

	pool := &outportcore.TransactionPool{
		SmartContractResults: map[string]*outportcore.SCRInfo{
			string(scrWithRefund): scr,
		},
	}

	arg := prepareMockArg()

	wasCalled := false
	txsFeeProc, err := NewTransactionsFeeProcessor(arg)
	txsFeeProc.log = &testscommon.LoggerStub{
		TraceCalled: func(message string, args ...interface{}) {
			wasCalled = true
			require.Equal(t, "transactionsFeeProcessor.prepareScrsNoTx: cannot find transaction in storage", message)
		},
	}

	require.NotNil(t, txsFeeProc)
	require.Nil(t, err)

	err = txsFeeProc.PutFeeAndGasUsed(pool)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(0), scr.GetFeeInfo().GetFee())
	require.Equal(t, uint64(0), scr.GetFeeInfo().GetGasUsed())
	require.True(t, wasCalled)
}

func TestPutFeeAndGasUsedScrWithRefundNotForInitialSender(t *testing.T) {
	t.Parallel()

	txHash := []byte("tx")
	scrWithRefund := []byte("scrWithRefund")

	refundValueBig, _ := big.NewInt(0).SetString("226498540000000", 10)

	scr := &outportcore.SCRInfo{
		SmartContractResult: &smartContractResult.SmartContractResult{
			Nonce:          3,
			SndAddr:        []byte("erd1qqqqqqqqqqqqqpgq3dswlnnlkfd3gqrcv3dhzgnvh8ryf27g5rfsecnn2s"),
			RcvAddr:        []byte("erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"),
			PrevTxHash:     []byte("f639cb7a0231191e04ec19dcb1359bd93a03fe8dc4a28a80d00835c5d1c988f8"),
			OriginalTxHash: txHash,
			Value:          refundValueBig,
			Data:           []byte(""),
		},
		FeeInfo: &outportcore.FeeInfo{Fee: big.NewInt(0)},
	}

	pool := &outportcore.TransactionPool{
		SmartContractResults: map[string]*outportcore.SCRInfo{
			string(scrWithRefund): scr,
		},
	}

	arg := prepareMockArg()

	initialTx := &transaction.Transaction{
		GasLimit: 30_000_000,
		GasPrice: 1000000000,
		SndAddr:  []byte("erd1dglncxk6sl9a3xumj78n6z2xux4ghp5c92cstv5zsn56tjgtdwpsk46qrs"),
	}
	txBytes, _ := arg.Marshaller.Marshal(initialTx)

	_ = arg.TransactionsStorer.Put(txHash, txBytes)

	txsFeeProc, err := NewTransactionsFeeProcessor(arg)
	require.NotNil(t, txsFeeProc)
	require.Nil(t, err)

	err = txsFeeProc.PutFeeAndGasUsed(pool)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(0), scr.GetFeeInfo().GetFee())
	require.Equal(t, uint64(0), scr.GetFeeInfo().GetGasUsed())
}

func TestPutFeeAndGasUsedScrWithRefund(t *testing.T) {
	t.Parallel()

	txHash := []byte("e3cdb8b4936fdbee2d3b1244b4c49959df5f90ada683d650019d244e5a64afaf")
	scrWithRefund := []byte("scrWithRefund")

	initialTx := &outportcore.TxInfo{Transaction: &transaction.Transaction{
		Nonce:    1004,
		GasLimit: 60_000_000,
		GasPrice: 1000000000,
		SndAddr:  []byte("erd1s8jr8e8hsvv7c9ehmshcjlpzf9ua5l50qeswa8feshrp6xlz9c7quacmtx"),
	}, FeeInfo: &outportcore.FeeInfo{Fee: big.NewInt(0)}}

	refundValueBig, _ := big.NewInt(0).SetString("96635000000000", 10)

	scr := &outportcore.SCRInfo{
		SmartContractResult: &smartContractResult.SmartContractResult{
			Nonce:          1005,
			SndAddr:        []byte("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"),
			RcvAddr:        []byte("erd1s8jr8e8hsvv7c9ehmshcjlpzf9ua5l50qeswa8feshrp6xlz9c7quacmtx"),
			PrevTxHash:     txHash,
			OriginalTxHash: txHash,
			Value:          refundValueBig,
			Data:           []byte("@6f6b"),
		},
		FeeInfo: &outportcore.FeeInfo{Fee: big.NewInt(0)},
	}

	pool := &outportcore.TransactionPool{
		SmartContractResults: map[string]*outportcore.SCRInfo{
			hex.EncodeToString(scrWithRefund): scr,
		},
		Transactions: map[string]*outportcore.TxInfo{
			hex.EncodeToString(txHash): initialTx,
		},
	}

	arg := prepareMockArg()
	txsFeeProc, err := NewTransactionsFeeProcessor(arg)
	require.NotNil(t, txsFeeProc)
	require.Nil(t, err)

	err = txsFeeProc.PutFeeAndGasUsed(pool)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(552865000000000), initialTx.GetFeeInfo().GetFee())
	require.Equal(t, uint64(50_336_500), initialTx.GetFeeInfo().GetGasUsed())
}
