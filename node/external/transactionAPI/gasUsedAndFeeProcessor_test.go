package transactionAPI

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/node/external/timemachine/fee"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

var pubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, log)

func TestComputeTransactionGasUsedAndFeeMoveBalance(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	feeComp, _ := fee.NewFeeComputer(fee.ArgsNewFeeComputer{
		BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
		EconomicsConfig:             testscommon.GetEconomicsConfig(),
	})
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
	require.Equal(uint64(50_000), moveBalanceTx.GasUsed)
	require.Equal("50000000000000", moveBalanceTx.Fee)
}

func TestComputeTransactionGasUsedAndFeeLogWithError(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	feeComp, _ := fee.NewFeeComputer(fee.ArgsNewFeeComputer{
		BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
		EconomicsConfig:             testscommon.GetEconomicsConfig(),
	})
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
	require.Equal(uint64(80_000), txWithSignalErrorLog.GasUsed)
	require.Equal("50300000000000", txWithSignalErrorLog.Fee)
}

func silentDecodeAddress(address string) []byte {
	decoded, _ := pubKeyConverter.Decode(address)
	return decoded
}

func TestComputeTransactionGasUsedAndFeeRelayedTxWithWriteLog(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	feeComp, _ := fee.NewFeeComputer(fee.ArgsNewFeeComputer{
		BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
		EconomicsConfig:             testscommon.GetEconomicsConfig(),
	})
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
	require.Equal(uint64(200_000), relayedTxWithWriteLog.GasUsed)
	require.Equal("66350000000000", relayedTxWithWriteLog.Fee)
}

func TestComputeTransactionGasUsedAndFeeTransactionWithScrWithRefund(t *testing.T) {
	require := require.New(t)
	feeComp, _ := fee.NewFeeComputer(fee.ArgsNewFeeComputer{
		BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
		EconomicsConfig:             testscommon.GetEconomicsConfig(),
	})
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
		GasLimit: 10_000_000,
		SmartContractResults: []*transaction.ApiSmartContractResult{
			{
				Value:    big.NewInt(66350000000000),
				IsRefund: true,
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
	require.Equal(uint64(3_365_000), txWithSRefundSCR.GasUsed)
	require.Equal("98000000000000", txWithSRefundSCR.Fee)
}

func TestNFTTransferWithScCall(t *testing.T) {
	require := require.New(t)
	feeComp, _ := fee.NewFeeComputer(fee.ArgsNewFeeComputer{
		BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
		EconomicsConfig:             testscommon.GetEconomicsConfig(),
	})
	computer := fee.NewTestFeeComputer(feeComp)

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
	require.Equal(uint64(55_000_000), tx.GasUsed)
	require.Equal("822250000000000", tx.Fee)
}
