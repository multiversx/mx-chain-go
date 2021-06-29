package txsimulator

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
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
			exError: ErrNilShardCoordinator,
		},
		{
			name: "NilTransactionProcessor",
			argsFunc: func() ArgsTxSimulator {
				args := getTxSimulatorArgs()
				args.TransactionProcessor = nil
				return args
			},
			exError: ErrNilTxSimulatorProcessor,
		},
		{
			name: "NilIntermProcessorContainer",
			argsFunc: func() ArgsTxSimulator {
				args := getTxSimulatorArgs()
				args.IntermediateProcContainer = nil
				return args
			},
			exError: ErrNilIntermediateProcessorContainer,
		},
		{
			name: "NilPubkeyConverter",
			argsFunc: func() ArgsTxSimulator {
				args := getTxSimulatorArgs()
				args.AddressPubKeyConverter = nil
				return args
			},
			exError: ErrNilPubkeyConverter,
		},
		{
			name: "NilHasher",
			argsFunc: func() ArgsTxSimulator {
				args := getTxSimulatorArgs()
				args.Hasher = nil
				return args
			},
			exError: ErrNilHasher,
		},
		{
			name: "NilMarshalizer",
			argsFunc: func() ArgsTxSimulator {
				args := getTxSimulatorArgs()
				args.Marshalizer = nil
				return args
			},
			exError: ErrNilMarshalizer,
		},
		{
			name: "NilVMOutputCacher",
			argsFunc: func() ArgsTxSimulator {
				args := getTxSimulatorArgs()
				args.VMOutputCacher = nil
				return args
			},
			exError: ErrNilCacher,
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

func TestTransactionSimulator_getVMOutputComputeHashFails(t *testing.T) {
	t.Parallel()

	args := getTxSimulatorArgs()
	args.Marshalizer = &mock.MarshalizerMock{
		Fail: true,
	}
	ts, _ := NewTransactionSimulator(args)
	require.False(t, ts.IsInterfaceNil())

	_, ok := ts.getVMOutputOfTx(nil)
	require.False(t, ok)
}

func TestTransactionSimulator_getVMOutput(t *testing.T) {
	t.Parallel()

	args := getTxSimulatorArgs()
	args.VMOutputCacher, _ = storageUnit.NewCache(storageUnit.CacheConfig{
		Type:     storageUnit.LRUCache,
		Capacity: 100,
	})

	ts, _ := NewTransactionSimulator(args)
	require.False(t, ts.IsInterfaceNil())

	// cannot find vm output in cacher
	_, ok := ts.getVMOutputOfTx(&transaction.Transaction{})
	require.False(t, ok)

	// wrong data in cacher
	tx := &transaction.Transaction{}
	txHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, tx)
	args.VMOutputCacher.Put(txHash, 1, 0)
	_, ok = ts.getVMOutputOfTx(tx)
	require.False(t, ok)
}

func TestTransactionSimulator_ProcessTxShouldIncludeScrsAndReceipts(t *testing.T) {
	t.Parallel()

	expectedSCr := map[string]data.TransactionHandler{
		"keySCr": &smartContractResult.SmartContractResult{
			RcvAddr:        []byte("rcvr"),
			RelayerAddr:    []byte("relayer"),
			OriginalSender: []byte("original"),
		},
	}
	expectedReceipts := map[string]data.TransactionHandler{
		"keyReceipt": &receipt.Receipt{SndAddr: []byte("sndr")},
	}

	args := getTxSimulatorArgs()
	args.VMOutputCacher, _ = storageUnit.NewCache(storageUnit.CacheConfig{
		Type:     storageUnit.LRUCache,
		Capacity: 100,
	})

	args.IntermediateProcContainer = &mock.IntermProcessorContainerStub{
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

	tx := &transaction.Transaction{Nonce: 37}
	txHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, tx)
	args.VMOutputCacher.Put(txHash, &vmcommon.VMOutput{}, 0)

	results, err := ts.ProcessTx(tx)
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
		TransactionProcessor:      &mock.TxProcessorStub{},
		IntermediateProcContainer: &mock.IntermProcessorContainerStub{},
		AddressPubKeyConverter:    &mock.PubkeyConverterMock{},
		ShardCoordinator:          mock.NewMultiShardsCoordinatorMock(2),
		VMOutputCacher:            txcache.NewDisabledCache(),
		Marshalizer:               &mock.MarshalizerMock{},
		Hasher:                    &mock.HasherMock{},
	}
}
