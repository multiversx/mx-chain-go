package transactionAPI

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/node/external/timemachine/fee"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createEconomicsData(enableEpochsHandler common.EnableEpochsHandler) process.EconomicsDataHandler {
	economicsConfig := testscommon.GetEconomicsConfig()
	economicsData, _ := economics.NewEconomicsData(economics.ArgsNewEconomicsData{
		Economics:           &economicsConfig,
		EnableEpochsHandler: enableEpochsHandler,
		TxVersionChecker:    &testscommon.TxVersionCheckerStub{},
		EpochNotifier:       &epochNotifier.EpochNotifierStub{},
	})

	dataFieldParser, _ := datafield.NewOperationDataFieldParser(&datafield.ArgsOperationDataFieldParser{
		AddressLength: 32,
		Marshalizer:   &mock.MarshalizerMock{},
	})

	_ = economicsData.SetTxTypeHandler(&testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			if core.IsSmartContractAddress(tx.GetRcvAddr()) {
				return process.SCInvoking, process.SCInvoking
			}

			res := dataFieldParser.Parse(tx.GetData(), tx.GetSndAddr(), tx.GetRcvAddr(), 3)
			if len(res.Tokens) > 0 {
				return process.BuiltInFunctionCall, process.BuiltInFunctionCall
			}

			return process.MoveBalance, process.MoveBalance
		},
	})

	return economicsData
}

var pubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, "erd")

func TestComputeTransactionGasUsedAndFeeMoveBalance(t *testing.T) {
	t.Parallel()

	req := require.New(t)
	feeComp, _ := fee.NewFeeComputer(createEconomicsData(&enableEpochsHandlerMock.EnableEpochsHandlerStub{}))
	computer := fee.NewTestFeeComputer(feeComp)

	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(
		computer,
		pubKeyConverter,
		&testscommon.ArgumentParserMock{},
		&testscommon.MarshallerStub{},
		enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
	)

	sender := "erd1wc3uh22g2aved3qeehkz9kzgrjwxhg9mkkxp2ee7jj7ph34p2csq0n2y5x"
	receiver := "erd1wc3uh22g2aved3qeehkz9kzgrjwxhg9mkkxp2ee7jj7ph34p2csq0n2y5x"

	moveBalanceTx := &transaction.ApiTransactionResult{
		Tx: &transaction.Transaction{
			GasLimit: 80_000,
			GasPrice: 1000000000,
			SndAddr:  silentDecodeAddress(sender),
			RcvAddr:  silentDecodeAddress(receiver),
		},
	}

	gasUsedAndFeeProc.computeAndAttachGasUsedAndFee(moveBalanceTx)
	req.Equal(uint64(50_000), moveBalanceTx.GasUsed)
	req.Equal("50000000000000", moveBalanceTx.Fee)
}

func TestComputeTransactionGasUsedAndFeeLogWithError(t *testing.T) {
	t.Parallel()

	req := require.New(t)
	feeComp, _ := fee.NewFeeComputer(createEconomicsData(&enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.GasPriceModifierFlag || flag == common.PenalizedTooMuchGasFlag
		},
	}))
	computer := fee.NewTestFeeComputer(feeComp)

	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(
		computer,
		pubKeyConverter,
		&testscommon.ArgumentParserMock{},
		&testscommon.MarshallerStub{},
		enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
	)

	sender := "erd1wc3uh22g2aved3qeehkz9kzgrjwxhg9mkkxp2ee7jj7ph34p2csq0n2y5x"
	receiver := "erd1wc3uh22g2aved3qeehkz9kzgrjwxhg9mkkxp2ee7jj7ph34p2csq0n2y5x"

	txWithSignalErrorLog := &transaction.ApiTransactionResult{
		Tx: &transaction.Transaction{
			GasLimit: 80_000,
			GasPrice: 1000000000,
			SndAddr:  silentDecodeAddress(sender),
			RcvAddr:  silentDecodeAddress(receiver),
		},
		GasLimit: 80_000,
		Logs: &transaction.ApiLogs{
			Events: []*transaction.Events{
				{
					Identifier: core.SignalErrorOperation,
				},
			},
		},
	}

	gasUsedAndFeeProc.computeAndAttachGasUsedAndFee(txWithSignalErrorLog)
	req.Equal(uint64(80_000), txWithSignalErrorLog.GasUsed)
	req.Equal("50300000000000", txWithSignalErrorLog.Fee)
}

func silentDecodeAddress(address string) []byte {
	decoded, _ := pubKeyConverter.Decode(address)
	return decoded
}

func TestComputeTransactionGasUsedAndFeeRelayedTxWithWriteLog(t *testing.T) {
	t.Parallel()

	req := require.New(t)
	feeComp, _ := fee.NewFeeComputer(createEconomicsData(&enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.GasPriceModifierFlag || flag == common.PenalizedTooMuchGasFlag
		},
	}))
	computer := fee.NewTestFeeComputer(feeComp)

	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(
		computer,
		pubKeyConverter,
		&testscommon.ArgumentParserMock{},
		&testscommon.MarshallerStub{},
		enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
	)

	sender := "erd1wc3uh22g2aved3qeehkz9kzgrjwxhg9mkkxp2ee7jj7ph34p2csq0n2y5x"
	receiver := "erd1wc3uh22g2aved3qeehkz9kzgrjwxhg9mkkxp2ee7jj7ph34p2csq0n2y5x"

	relayedTxWithWriteLog := &transaction.ApiTransactionResult{
		Tx: &transaction.Transaction{
			GasLimit: 200_000,
			GasPrice: 1000000000,
			SndAddr:  silentDecodeAddress(sender),
			RcvAddr:  silentDecodeAddress(receiver),
			Data:     []byte("relayedTx@"),
		},
		GasLimit: 200_000,
		Logs: &transaction.ApiLogs{
			Events: []*transaction.Events{
				{
					Identifier: core.WriteLogIdentifier,
				},
			},
		},
		IsRelayed: true,
	}

	gasUsedAndFeeProc.computeAndAttachGasUsedAndFee(relayedTxWithWriteLog)
	req.Equal(uint64(200_000), relayedTxWithWriteLog.GasUsed)
	req.Equal("66350000000000", relayedTxWithWriteLog.Fee)
}

func TestComputeTransactionGasUsedAndFeeTransactionWithScrWithRefund(t *testing.T) {
	req := require.New(t)
	feeComp, _ := fee.NewFeeComputer(createEconomicsData(&enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.GasPriceModifierFlag || flag == common.PenalizedTooMuchGasFlag
		},
	}))
	computer := fee.NewTestFeeComputer(feeComp)

	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(
		computer,
		pubKeyConverter,
		&testscommon.ArgumentParserMock{},
		&testscommon.MarshallerStub{},
		enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
	)

	sender := "erd1wc3uh22g2aved3qeehkz9kzgrjwxhg9mkkxp2ee7jj7ph34p2csq0n2y5x"
	receiver := "erd1wc3uh22g2aved3qeehkz9kzgrjwxhg9mkkxp2ee7jj7ph34p2csq0n2y5x"

	txWithSRefundSCR := &transaction.ApiTransactionResult{
		Tx: &transaction.Transaction{
			GasLimit: 10_000_000,
			GasPrice: 1000000000,
			SndAddr:  silentDecodeAddress(sender),
			RcvAddr:  silentDecodeAddress(receiver),
			Data:     []byte("relayedTx@"),
		},
		Sender:   sender,
		Receiver: receiver,
		GasLimit: 10_000_000,
		SmartContractResults: []*transaction.ApiSmartContractResult{
			{
				Value:    big.NewInt(66350000000000),
				IsRefund: true,
				RcvAddr:  sender,
			},
		},
		Logs: &transaction.ApiLogs{
			Events: []*transaction.Events{
				{
					Identifier: core.WriteLogIdentifier,
				},
			},
		},
		IsRelayed: true,
	}

	gasUsedAndFeeProc.computeAndAttachGasUsedAndFee(txWithSRefundSCR)
	req.Equal(uint64(3_365_000), txWithSRefundSCR.GasUsed)
	req.Equal("98000000000000", txWithSRefundSCR.Fee)
}

func TestNFTTransferWithScCall(t *testing.T) {
	req := require.New(t)
	feeComp, err := fee.NewFeeComputer(createEconomicsData(&enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.GasPriceModifierFlag || flag == common.PenalizedTooMuchGasFlag
		},
	}))
	computer := fee.NewTestFeeComputer(feeComp)
	req.Nil(err)

	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(
		computer,
		pubKeyConverter,
		&testscommon.ArgumentParserMock{},
		&testscommon.MarshallerStub{},
		enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
	)

	sender := "erd1wc3uh22g2aved3qeehkz9kzgrjwxhg9mkkxp2ee7jj7ph34p2csq0n2y5x"
	receiver := "erd1wc3uh22g2aved3qeehkz9kzgrjwxhg9mkkxp2ee7jj7ph34p2csq0n2y5x"

	tx := &transaction.ApiTransactionResult{
		Tx: &transaction.Transaction{
			GasLimit: 55_000_000,
			GasPrice: 1000000000,
			SndAddr:  silentDecodeAddress(sender),
			RcvAddr:  silentDecodeAddress(receiver),
			Data:     []byte("ESDTNFTTransfer@434f572d636434363364@080c@01@00000000000000000500d3b28828d62052124f07dcd50ed31b0825f60eee1526@616363657074476c6f62616c4f66666572@c3e5q"),
		},
		GasLimit:  55_000_000,
		Receivers: []string{"erd1qqqqqqqqqqqqqpgq6wegs2xkypfpync8mn2sa5cmpqjlvrhwz5nqgepyg8"},
		Function:  "acceptGlobalOffer",
		Operation: "ESDTNFTTransfer",
	}
	tx.InitiallyPaidFee = feeComp.ComputeTransactionFee(tx).String()

	gasUsedAndFeeProc.computeAndAttachGasUsedAndFee(tx)
	req.Equal(uint64(55_000_000), tx.GasUsed)
	req.Equal("822250000000000", tx.Fee)
}

func TestComputeAndAttachGasUsedAndFeeTransactionWithMultipleScrWithRefund(t *testing.T) {
	t.Parallel()

	feeComp, _ := fee.NewFeeComputer(createEconomicsData(&enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.GasPriceModifierFlag ||
				flag == common.PenalizedTooMuchGasFlag ||
				flag == common.RelayedTransactionsV3Flag
		},
	}))
	computer := fee.NewTestFeeComputer(feeComp)

	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(
		computer,
		pubKeyConverter,
		&testscommon.ArgumentParserMock{},
		&testscommon.MarshallerStub{},
		enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
	)

	txWithSRefundSCR := &transaction.ApiTransactionResult{}
	err := core.LoadJsonFile(txWithSRefundSCR, "testData/scInvokingWithMultipleRefunds.json")
	require.NoError(t, err)

	txWithSRefundSCR.Fee = ""
	txWithSRefundSCR.GasUsed = 0

	snd, _ := pubKeyConverter.Decode(txWithSRefundSCR.Sender)
	rcv, _ := pubKeyConverter.Decode(txWithSRefundSCR.Receiver)
	val, _ := big.NewInt(0).SetString(txWithSRefundSCR.Value, 10)
	txWithSRefundSCR.Tx = &transaction.Transaction{
		Nonce:    txWithSRefundSCR.Nonce,
		Value:    val,
		RcvAddr:  rcv,
		SndAddr:  snd,
		GasPrice: txWithSRefundSCR.GasPrice,
		GasLimit: txWithSRefundSCR.GasLimit,
		Data:     txWithSRefundSCR.Data,
	}

	gasUsedAndFeeProc.computeAndAttachGasUsedAndFee(txWithSRefundSCR)
	require.Equal(t, uint64(20313408), txWithSRefundSCR.GasUsed)
	require.Equal(t, "319459080000000", txWithSRefundSCR.Fee)
}

func TestComputeAndAttachGasUsedAndFeeRelayedV3WithRefund(t *testing.T) {
	t.Parallel()

	feeComp, _ := fee.NewFeeComputer(createEconomicsData(&enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.GasPriceModifierFlag ||
				flag == common.PenalizedTooMuchGasFlag ||
				flag == common.RelayedTransactionsV3Flag
		},
	}))
	computer := fee.NewTestFeeComputer(feeComp)

	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(
		computer,
		pubKeyConverter,
		&testscommon.ArgumentParserMock{},
		&testscommon.MarshallerStub{},
		enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
	)

	txWithSRefundSCR := &transaction.ApiTransactionResult{}
	err := core.LoadJsonFile(txWithSRefundSCR, "testData/relayedV3WithOneRefund.json")
	require.NoError(t, err)

	txWithSRefundSCR.Fee = ""
	txWithSRefundSCR.GasUsed = 0

	innerTxs := make([]*transaction.Transaction, 0, len(txWithSRefundSCR.InnerTransactions))
	for _, innerTx := range txWithSRefundSCR.InnerTransactions {
		snd, _ := pubKeyConverter.Decode(innerTx.Sender)
		rcv, _ := pubKeyConverter.Decode(innerTx.Receiver)
		val, _ := big.NewInt(0).SetString(innerTx.Value, 10)

		innerTxs = append(innerTxs, &transaction.Transaction{
			Nonce:    innerTx.Nonce,
			Value:    val,
			RcvAddr:  rcv,
			SndAddr:  snd,
			GasPrice: innerTx.GasPrice,
			GasLimit: innerTx.GasLimit,
			Data:     innerTx.Data,
		})
	}

	snd, _ := pubKeyConverter.Decode(txWithSRefundSCR.Sender)
	rcv, _ := pubKeyConverter.Decode(txWithSRefundSCR.Receiver)
	val, _ := big.NewInt(0).SetString(txWithSRefundSCR.Value, 10)
	txWithSRefundSCR.Tx = &transaction.Transaction{
		Nonce:             txWithSRefundSCR.Nonce,
		Value:             val,
		RcvAddr:           rcv,
		SndAddr:           snd,
		GasPrice:          txWithSRefundSCR.GasPrice,
		GasLimit:          txWithSRefundSCR.GasLimit,
		Data:              txWithSRefundSCR.Data,
		InnerTransactions: innerTxs,
	}

	gasUsedAndFeeProc.computeAndAttachGasUsedAndFee(txWithSRefundSCR)
	require.Equal(t, uint64(55149500), txWithSRefundSCR.GasUsed)
	require.Equal(t, "699500000000000", txWithSRefundSCR.Fee)
}

func TestComputeAndAttachGasUsedAndFeeFailedRelayedV1(t *testing.T) {
	t.Parallel()

	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.GasPriceModifierFlag ||
				flag == common.PenalizedTooMuchGasFlag ||
				flag == common.RelayedTransactionsV3Flag
		},
	}
	feeComp, _ := fee.NewFeeComputer(createEconomicsData(enableEpochsHandler))
	computer := fee.NewTestFeeComputer(feeComp)

	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(
		computer,
		pubKeyConverter,
		smartContract.NewArgumentParser(),
		&marshal.JsonMarshalizer{},
		enableEpochsHandler,
	)

	txWithSRefundSCR := &transaction.ApiTransactionResult{}
	err := core.LoadJsonFile(txWithSRefundSCR, "testData/failedRelayedV1.json")
	require.NoError(t, err)

	snd, _ := pubKeyConverter.Decode(txWithSRefundSCR.Sender)
	rcv, _ := pubKeyConverter.Decode(txWithSRefundSCR.Receiver)
	val, _ := big.NewInt(0).SetString(txWithSRefundSCR.Value, 10)
	txWithSRefundSCR.Tx = &transaction.Transaction{
		Nonce:    txWithSRefundSCR.Nonce,
		Value:    val,
		RcvAddr:  rcv,
		SndAddr:  snd,
		GasPrice: txWithSRefundSCR.GasPrice,
		GasLimit: txWithSRefundSCR.GasLimit,
		Data:     txWithSRefundSCR.Data,
	}

	txWithSRefundSCR.InitiallyPaidFee = ""
	txWithSRefundSCR.Fee = ""
	txWithSRefundSCR.GasUsed = 0

	gasUsedAndFeeProc.computeAndAttachGasUsedAndFee(txWithSRefundSCR)
	require.Equal(t, uint64(1274230), txWithSRefundSCR.GasUsed)
	require.Equal(t, "1274230000000000", txWithSRefundSCR.Fee)
	require.Equal(t, "1274230000000000", txWithSRefundSCR.InitiallyPaidFee)
}

func TestComputeAndAttachGasUsedAndFeeRelayedV3WithAllInnerTxFailed(t *testing.T) {
	t.Parallel()

	t.Run("all inner txs are failed", testComputeAndAttachGasUsedAndFeeRelayedV3WithInnerTxFailed(
		"testData/relayedV3WithAllInnerTxFailed.json",
		uint64(60150000),
		"2226090000000000",
		"2226090000000000",
	))
	t.Run("one inner tx is failed, other have refunds", testComputeAndAttachGasUsedAndFeeRelayedV3WithInnerTxFailed(
		"testData/relayedV3WithOneInnerFailedAndTwoRefunds.json",
		uint64(160766000),
		"2670920000000000",
		"2864760000000000",
	))

}

func testComputeAndAttachGasUsedAndFeeRelayedV3WithInnerTxFailed(
	inputFile string,
	expectedGasUsed uint64,
	expectedFee string,
	expectedInitiallyPaidFee string,
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.GasPriceModifierFlag ||
					flag == common.PenalizedTooMuchGasFlag ||
					flag == common.RelayedTransactionsV3Flag
			},
		}
		feeComp, _ := fee.NewFeeComputer(createEconomicsData(enableEpochsHandler))
		computer := fee.NewTestFeeComputer(feeComp)

		gasUsedAndFeeProc := newGasUsedAndFeeProcessor(
			computer,
			pubKeyConverter,
			smartContract.NewArgumentParser(),
			&marshal.JsonMarshalizer{},
			enableEpochsHandler,
		)

		txWithFailedInners := &transaction.ApiTransactionResult{}
		err := core.LoadJsonFile(txWithFailedInners, inputFile)
		require.NoError(t, err)

		innerTxs := make([]*transaction.Transaction, 0, len(txWithFailedInners.InnerTransactions))
		for _, innerTx := range txWithFailedInners.InnerTransactions {
			snd, _ := pubKeyConverter.Decode(innerTx.Sender)
			rcv, _ := pubKeyConverter.Decode(innerTx.Receiver)
			val, _ := big.NewInt(0).SetString(innerTx.Value, 10)
			innerTxs = append(innerTxs, &transaction.Transaction{
				Nonce:    innerTx.Nonce,
				Value:    val,
				RcvAddr:  rcv,
				SndAddr:  snd,
				GasPrice: innerTx.GasPrice,
				GasLimit: innerTx.GasLimit,
				Data:     innerTx.Data,
			})
		}

		snd, _ := pubKeyConverter.Decode(txWithFailedInners.Sender)
		rcv, _ := pubKeyConverter.Decode(txWithFailedInners.Receiver)
		val, _ := big.NewInt(0).SetString(txWithFailedInners.Value, 10)
		txWithFailedInners.Tx = &transaction.Transaction{
			Nonce:             txWithFailedInners.Nonce,
			Value:             val,
			RcvAddr:           rcv,
			SndAddr:           snd,
			GasPrice:          txWithFailedInners.GasPrice,
			GasLimit:          txWithFailedInners.GasLimit,
			Data:              txWithFailedInners.Data,
			InnerTransactions: innerTxs,
		}

		txWithFailedInners.InitiallyPaidFee = ""
		txWithFailedInners.Fee = ""
		txWithFailedInners.GasUsed = 0

		gasUsedAndFeeProc.computeAndAttachGasUsedAndFee(txWithFailedInners)
		assert.Equal(t, expectedGasUsed, txWithFailedInners.GasUsed)
		assert.Equal(t, expectedFee, txWithFailedInners.Fee)
		assert.Equal(t, expectedInitiallyPaidFee, txWithFailedInners.InitiallyPaidFee)
	}
}
