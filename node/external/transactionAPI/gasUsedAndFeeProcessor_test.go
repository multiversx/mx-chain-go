package transactionAPI

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/node/external/timemachine/fee"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/stretchr/testify/require"
)

func createEconomicsData(enableEpochsHandler common.EnableEpochsHandler) process.EconomicsDataHandler {
	economicsConfig := testscommon.GetEconomicsConfig()
	economicsData, _ := economics.NewEconomicsData(economics.ArgsNewEconomicsData{
		Economics:           &economicsConfig,
		EnableEpochsHandler: enableEpochsHandler,
		TxVersionChecker:    &testscommon.TxVersionCheckerStub{},
		EpochNotifier:       &epochNotifier.EpochNotifierStub{},
		PubkeyConverter:     &testscommon.PubkeyConverterStub{},
		ShardCoordinator:    &testscommon.ShardsCoordinatorMock{},
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

	eeh := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.GasPriceModifierFlag ||
				flag == common.PenalizedTooMuchGasFlag ||
				flag == common.FixRelayedBaseCostFlag
		},
	}
	feeComp, _ := fee.NewFeeComputer(createEconomicsData(eeh))
	computer := fee.NewTestFeeComputer(feeComp)

	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(
		computer,
		pubKeyConverter,
		&testscommon.ArgumentParserMock{},
		&testscommon.MarshallerStub{},
		eeh,
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

func TestComputeAndAttachGasUsedAndFeeFailedRelayedV1(t *testing.T) {
	t.Parallel()

	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.GasPriceModifierFlag ||
				flag == common.PenalizedTooMuchGasFlag ||
				flag == common.FixRelayedBaseCostFlag
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
	require.Equal(t, uint64(6148000), txWithSRefundSCR.GasUsed)
	require.Equal(t, "1198000000000000", txWithSRefundSCR.Fee)
	require.Equal(t, "1274230000000000", txWithSRefundSCR.InitiallyPaidFee)
}

func TestComputeAndAttachGasUsedAndFeeRelayedV1CreateNewDelegationContractWithRefund(t *testing.T) {
	t.Parallel()

	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.GasPriceModifierFlag ||
				flag == common.PenalizedTooMuchGasFlag ||
				flag == common.FixRelayedBaseCostFlag
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
	err := core.LoadJsonFile(txWithSRefundSCR, "testData/relayedV1CreateNewDelegationContract.json")
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
	require.Equal(t, uint64(56328500), txWithSRefundSCR.GasUsed)
	require.Equal(t, "1878500000000000", txWithSRefundSCR.Fee)
	require.Equal(t, "2177505000000000", txWithSRefundSCR.InitiallyPaidFee)
}

func TestComputeAndAttachGasUsedAndFeeRelayedV3WithMultipleRefunds(t *testing.T) {
	t.Parallel()

	eeh := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.GasPriceModifierFlag ||
				flag == common.PenalizedTooMuchGasFlag ||
				flag == common.FixRelayedBaseCostFlag ||
				flag == common.RelayedTransactionsV3Flag
		},
	}
	feeComp, _ := fee.NewFeeComputer(createEconomicsData(eeh))
	computer := fee.NewTestFeeComputer(feeComp)

	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(
		computer,
		pubKeyConverter,
		&testscommon.ArgumentParserMock{},
		&testscommon.MarshallerStub{},
		eeh,
	)

	txWithRefunds := &transaction.ApiTransactionResult{}
	err := core.LoadJsonFile(txWithRefunds, "testData/relayedV3WithMultipleRefunds.json")
	require.NoError(t, err)

	txWithRefunds.Fee = ""
	txWithRefunds.GasUsed = 0

	snd, _ := pubKeyConverter.Decode(txWithRefunds.Sender)
	rcv, _ := pubKeyConverter.Decode(txWithRefunds.Receiver)
	rel, _ := pubKeyConverter.Decode(txWithRefunds.RelayerAddress)
	val, _ := big.NewInt(0).SetString(txWithRefunds.Value, 10)
	sig, _ := hex.DecodeString(txWithRefunds.Signature)
	relayerSig, _ := hex.DecodeString(txWithRefunds.RelayerSignature)
	txWithRefunds.Tx = &transaction.Transaction{
		Nonce:            txWithRefunds.Nonce,
		Value:            val,
		RcvAddr:          rcv,
		SndAddr:          snd,
		RelayerAddr:      rel,
		GasPrice:         txWithRefunds.GasPrice,
		GasLimit:         txWithRefunds.GasLimit,
		Data:             txWithRefunds.Data,
		Signature:        sig,
		RelayerSignature: relayerSig,
	}

	gasUsedAndFeeProc.computeAndAttachGasUsedAndFee(txWithRefunds)
	require.Equal(t, uint64(4220447), txWithRefunds.GasUsed)
	require.Equal(t, "289704470000000", txWithRefunds.Fee)
}
