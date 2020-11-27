package txsimulator

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/require"
)

func TestNewTransactionSimulator(t *testing.T) {
	tests := []struct {
		name     string
		argsFunc func() ArgsTxSimulator
		exError  error
	}{
		{
			name: "NilShardCoordinator",
			argsFunc: func() ArgsTxSimulator {
				args := getTxSimulatorArgs()
				args.ShardCoordinator = nil
				return args
			},
			exError: node.ErrNilShardCoordinator,
		},
		{
			name: "NilTransactionProcessor",
			argsFunc: func() ArgsTxSimulator {
				args := getTxSimulatorArgs()
				args.TransactionProcessor = nil
				return args
			},
			exError: node.ErrNilTxSimulatorProcessor,
		},
		{
			name: "NilIntermProcessorContainer",
			argsFunc: func() ArgsTxSimulator {
				args := getTxSimulatorArgs()
				args.IntermmediateProcContainer = nil
				return args
			},
			exError: node.ErrNilIntermediateProcessorContainer,
		},
		{
			name: "Ok",
			argsFunc: func() ArgsTxSimulator {
				args := getTxSimulatorArgs()
				return args
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTransactionSimulator(tt.argsFunc())
			require.Equal(t, err, tt.exError)
		})
	}
}

func TestTransactionSimulator_ProcessTxProcessingErrShouldSignal(t *testing.T) {
	t.Parallel()

	expRetCode := vmcommon.Ok
	expErr := errors.New("transaction failed")
	args := getTxSimulatorArgs()
	args.TransactionProcessor = &mock.TxProcessorStub{
		ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
			return expRetCode, expErr
		},
	}
	ts, _ := NewTransactionSimulator(args)

	results, err := ts.ProcessTx(&transaction.Transaction{Nonce: 37})
	require.NoError(t, err)
	require.Equal(t, expErr.Error(), results.FailReason)
}

func TestTransactionSimulator_ProcessTxShouldIncludeScrsAndReceipts(t *testing.T) {
	t.Parallel()

	expectedSCr := map[string]data.TransactionHandler{
		"keySCr": &smartContractResult.SmartContractResult{RcvAddr: []byte("rcvr")},
	}
	expectedReceipts := map[string]data.TransactionHandler{
		"keyReceipt": &receipt.Receipt{SndAddr: []byte("sndr")},
	}

	args := getTxSimulatorArgs()
	args.IntermmediateProcContainer = &mock.IntermProcessorContainerStub{
		GetCalled: func(key block.Type) (process.IntermediateTransactionHandler, error) {
			return &mock.IntermediateTransactionHandlerStub{
				GetAllCurrentFinishedTxsCalled: func() map[string]data.TransactionHandler {
					if key == block.SmartContractResultBlock {
						return expectedSCr
					}
					return expectedReceipts
				},
			}, nil
		},
		KeysCalled: nil,
	}
	ts, _ := NewTransactionSimulator(args)

	results, err := ts.ProcessTx(&transaction.Transaction{Nonce: 37})
	require.NoError(t, err)
	require.Equal(
		t,
		hex.EncodeToString(expectedSCr["keySCr"].GetRcvAddr()),
		results.ScResults[hex.EncodeToString([]byte("keySCr"))].RcvAddr,
	)
	require.Equal(
		t,
		hex.EncodeToString(expectedReceipts["keyReceipt"].GetSndAddr()),
		results.Receipts[hex.EncodeToString([]byte("keyReceipt"))].SndAddr,
	)
}

func getTxSimulatorArgs() ArgsTxSimulator {
	return ArgsTxSimulator{
		TransactionProcessor:       &mock.TxProcessorStub{},
		IntermmediateProcContainer: &mock.IntermProcessorContainerStub{},
		AddressPubKeyConverter:     &mock.PubkeyConverterMock{},
		ShardCoordinator:           mock.NewMultiShardsCoordinatorMock(2),
	}
}
