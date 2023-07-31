package transactionAPI

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/node/external/timemachine/fee"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/stretchr/testify/require"
)

func createEconomicsData(enableEpochsHandler common.EnableEpochsHandler) fee.EconomicsDataWithComputeFee {
	economicsConfig := testscommon.GetEconomicsConfig()
	economicsData, _ := economics.NewEconomicsData(economics.ArgsNewEconomicsData{
		BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
		Economics:                   &economicsConfig,
		EnableEpochsHandler:         enableEpochsHandler,
		TxVersionChecker:            &testscommon.TxVersionCheckerStub{},
		EpochNotifier:               &epochNotifier.EpochNotifierStub{},
	})

	return economicsData
}

var pubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
var flagActiveTrueHandler = func(epoch uint32) bool { return true }

func TestComputeTransactionGasUsedAndFeeMoveBalance(t *testing.T) {
	t.Parallel()

	req := require.New(t)
	feeComp, _ := fee.NewFeeComputer(createEconomicsData(&enableEpochsHandlerMock.EnableEpochsHandlerStub{}))
	computer := fee.NewTestFeeComputer(feeComp)

	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(computer, pubKeyConverter)

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
		IsGasPriceModifierFlagEnabledInEpochCalled:    flagActiveTrueHandler,
		IsPenalizedTooMuchGasFlagEnabledInEpochCalled: flagActiveTrueHandler,
	}))
	computer := fee.NewTestFeeComputer(feeComp)

	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(computer, pubKeyConverter)

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
		IsGasPriceModifierFlagEnabledInEpochCalled:    flagActiveTrueHandler,
		IsPenalizedTooMuchGasFlagEnabledInEpochCalled: flagActiveTrueHandler,
	}))
	computer := fee.NewTestFeeComputer(feeComp)

	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(computer, pubKeyConverter)

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
		IsGasPriceModifierFlagEnabledInEpochCalled:    flagActiveTrueHandler,
		IsPenalizedTooMuchGasFlagEnabledInEpochCalled: flagActiveTrueHandler,
	}))
	computer := fee.NewTestFeeComputer(feeComp)

	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(computer, pubKeyConverter)

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
		IsGasPriceModifierFlagEnabledInEpochCalled:    flagActiveTrueHandler,
		IsPenalizedTooMuchGasFlagEnabledInEpochCalled: flagActiveTrueHandler,
	}))
	computer := fee.NewTestFeeComputer(feeComp)
	req.Nil(err)

	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(computer, pubKeyConverter)

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
