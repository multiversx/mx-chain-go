package transaction_test

import (
	"bytes"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	coreTransaction "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/transaction"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/stretchr/testify/require"
)

const minGasLimit = uint64(1)

func getDefaultTx() *coreTransaction.Transaction {
	return &coreTransaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  []byte("rel"),
		SndAddr:  []byte("rel"),
		GasPrice: 1,
		GasLimit: minGasLimit * 4,
		InnerTransactions: []*coreTransaction.Transaction{
			{
				Nonce:       0,
				Value:       big.NewInt(1),
				RcvAddr:     []byte("rcv1"),
				SndAddr:     []byte("snd1"),
				GasPrice:    1,
				GasLimit:    minGasLimit,
				RelayerAddr: []byte("rel"),
			},
			{
				Nonce:       0,
				Value:       big.NewInt(1),
				RcvAddr:     []byte("rcv1"),
				SndAddr:     []byte("snd2"),
				GasPrice:    1,
				GasLimit:    minGasLimit,
				RelayerAddr: []byte("rel"),
			},
		},
	}
}

func createMockArgRelayedTxV3Processor() transaction.ArgRelayedTxV3Processor {
	return transaction.ArgRelayedTxV3Processor{
		EconomicsFee:           &economicsmocks.EconomicsHandlerStub{},
		ShardCoordinator:       &testscommon.ShardsCoordinatorMock{},
		MaxTransactionsAllowed: 10,
	}
}

func TestNewRelayedTxV3Processor(t *testing.T) {
	t.Parallel()

	t.Run("nil economics fee should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgRelayedTxV3Processor()
		args.EconomicsFee = nil
		proc, err := transaction.NewRelayedTxV3Processor(args)
		require.Nil(t, proc)
		require.Equal(t, process.ErrNilEconomicsFeeHandler, err)
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgRelayedTxV3Processor()
		args.ShardCoordinator = nil
		proc, err := transaction.NewRelayedTxV3Processor(args)
		require.Nil(t, proc)
		require.Equal(t, process.ErrNilShardCoordinator, err)
	})
	t.Run("invalid max transactions allowed should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgRelayedTxV3Processor()
		args.MaxTransactionsAllowed = 0
		proc, err := transaction.NewRelayedTxV3Processor(args)
		require.Nil(t, proc)
		require.True(t, errors.Is(err, process.ErrInvalidValue))
		require.True(t, strings.Contains(err.Error(), "MaxTransactionsAllowed"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(createMockArgRelayedTxV3Processor())
		require.NoError(t, err)
		require.NotNil(t, proc)
	})
}

func TestRelayedTxV3Processor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	args := createMockArgRelayedTxV3Processor()
	args.EconomicsFee = nil
	proc, _ := transaction.NewRelayedTxV3Processor(args)
	require.True(t, proc.IsInterfaceNil())

	proc, _ = transaction.NewRelayedTxV3Processor(createMockArgRelayedTxV3Processor())
	require.False(t, proc.IsInterfaceNil())
}

func TestRelayedTxV3Processor_CheckRelayedTx(t *testing.T) {
	t.Parallel()

	t.Run("invalid num of inner txs should error", func(t *testing.T) {
		t.Parallel()

		tx := getDefaultTx()
		args := createMockArgRelayedTxV3Processor()
		args.MaxTransactionsAllowed = len(tx.InnerTransactions) - 1
		proc, err := transaction.NewRelayedTxV3Processor(args)
		require.NoError(t, err)

		tx.Value = big.NewInt(1)

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRelayedTxV3TooManyInnerTransactions, err)
	})
	t.Run("value on relayed tx should error", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(createMockArgRelayedTxV3Processor())
		require.NoError(t, err)

		tx := getDefaultTx()
		tx.Value = big.NewInt(1)

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRelayedTxV3ZeroVal, err)
	})
	t.Run("relayed tx not to self should error", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(createMockArgRelayedTxV3Processor())
		require.NoError(t, err)

		tx := getDefaultTx()
		tx.RcvAddr = []byte("another rcv")

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRelayedTxV3SenderDoesNotMatchReceiver, err)
	})
	t.Run("invalid gas limit should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgRelayedTxV3Processor()
		args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return minGasLimit
			},
		}
		proc, err := transaction.NewRelayedTxV3Processor(args)
		require.NoError(t, err)

		tx := getDefaultTx()
		tx.GasLimit = minGasLimit

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRelayedTxV3GasLimitMismatch, err)
	})
	t.Run("data field not empty should error", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(createMockArgRelayedTxV3Processor())
		require.NoError(t, err)

		tx := getDefaultTx()
		tx.Data = []byte("dummy")

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRelayedTxV3InvalidDataField, err)
	})
	t.Run("inner txs on inner should error", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(createMockArgRelayedTxV3Processor())
		require.NoError(t, err)

		tx := getDefaultTx()
		tx.InnerTransactions[0].InnerTransactions = []*coreTransaction.Transaction{{}}

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRecursiveRelayedTxIsNotAllowed, err)
	})
	t.Run("relayer mismatch on inner should error", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(createMockArgRelayedTxV3Processor())
		require.NoError(t, err)

		tx := getDefaultTx()
		tx.InnerTransactions[0].RelayerAddr = []byte("another relayer")

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRelayedTxV3RelayerMismatch, err)
	})
	t.Run("gas price mismatch on inner should error", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(createMockArgRelayedTxV3Processor())
		require.NoError(t, err)

		tx := getDefaultTx()
		tx.InnerTransactions[0].GasPrice = tx.GasPrice + 1

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRelayedV3GasPriceMismatch, err)
	})
	t.Run("shard mismatch on inner should error", func(t *testing.T) {
		t.Parallel()

		tx := getDefaultTx()
		args := createMockArgRelayedTxV3Processor()
		args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
			ComputeIdCalled: func(address []byte) uint32 {
				if bytes.Equal(address, tx.SndAddr) {
					return 0
				}

				return 1
			},
		}
		proc, err := transaction.NewRelayedTxV3Processor(args)
		require.NoError(t, err)

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRelayedTxV3SenderShardMismatch, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(createMockArgRelayedTxV3Processor())
		require.NoError(t, err)

		tx := getDefaultTx()
		err = proc.CheckRelayedTx(tx)
		require.NoError(t, err)
	})
}
