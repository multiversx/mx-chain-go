package transaction_test

import (
	"bytes"
	"math/big"
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

func TestNewRelayedTxV3Processor(t *testing.T) {
	t.Parallel()

	t.Run("nil economics fee should error", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(nil, nil)
		require.Nil(t, proc)
		require.Equal(t, process.ErrNilEconomicsFeeHandler, err)
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(&economicsmocks.EconomicsHandlerStub{}, nil)
		require.Nil(t, proc)
		require.Equal(t, process.ErrNilShardCoordinator, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(&economicsmocks.EconomicsHandlerStub{}, &testscommon.ShardsCoordinatorMock{})
		require.NoError(t, err)
		require.NotNil(t, proc)
	})
}

func TestRelayedTxV3Processor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	proc, _ := transaction.NewRelayedTxV3Processor(nil, nil)
	require.True(t, proc.IsInterfaceNil())

	proc, _ = transaction.NewRelayedTxV3Processor(&economicsmocks.EconomicsHandlerStub{}, &testscommon.ShardsCoordinatorMock{})
	require.False(t, proc.IsInterfaceNil())
}

func TestRelayedTxV3Processor_CheckRelayedTx(t *testing.T) {
	t.Parallel()

	t.Run("value on relayed tx should error", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(&economicsmocks.EconomicsHandlerStub{}, &testscommon.ShardsCoordinatorMock{})
		require.NoError(t, err)

		tx := getDefaultTx()
		tx.Value = big.NewInt(1)

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRelayedTxV3ZeroVal, err)
	})
	t.Run("relayed tx not to self should error", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(&economicsmocks.EconomicsHandlerStub{}, &testscommon.ShardsCoordinatorMock{})
		require.NoError(t, err)

		tx := getDefaultTx()
		tx.RcvAddr = []byte("another rcv")

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRelayedTxV3SenderDoesNotMatchReceiver, err)
	})
	t.Run("invalid gas limit should error", func(t *testing.T) {
		t.Parallel()

		economicsFeeHandler := &economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return minGasLimit
			},
		}
		proc, err := transaction.NewRelayedTxV3Processor(economicsFeeHandler, &testscommon.ShardsCoordinatorMock{})
		require.NoError(t, err)

		tx := getDefaultTx()
		tx.GasLimit = minGasLimit

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRelayedTxV3GasLimitMismatch, err)
	})
	t.Run("empty relayer on inner should error", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(&economicsmocks.EconomicsHandlerStub{}, &testscommon.ShardsCoordinatorMock{})
		require.NoError(t, err)

		tx := getDefaultTx()
		tx.InnerTransactions[0].RelayerAddr = []byte("")

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRelayedTxV3EmptyRelayer, err)
	})
	t.Run("relayer mismatch on inner should error", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(&economicsmocks.EconomicsHandlerStub{}, &testscommon.ShardsCoordinatorMock{})
		require.NoError(t, err)

		tx := getDefaultTx()
		tx.InnerTransactions[0].RelayerAddr = []byte("another relayer")

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRelayedTxV3RelayerMismatch, err)
	})
	t.Run("gas price mismatch on inner should error", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(&economicsmocks.EconomicsHandlerStub{}, &testscommon.ShardsCoordinatorMock{})
		require.NoError(t, err)

		tx := getDefaultTx()
		tx.InnerTransactions[0].GasPrice = tx.GasPrice + 1

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRelayedV3GasPriceMismatch, err)
	})
	t.Run("shard mismatch on inner should error", func(t *testing.T) {
		t.Parallel()

		tx := getDefaultTx()
		shardC := &testscommon.ShardsCoordinatorMock{
			ComputeIdCalled: func(address []byte) uint32 {
				if bytes.Equal(address, tx.SndAddr) {
					return 0
				}

				return 1
			},
		}
		proc, err := transaction.NewRelayedTxV3Processor(&economicsmocks.EconomicsHandlerStub{}, shardC)
		require.NoError(t, err)

		err = proc.CheckRelayedTx(tx)
		require.Equal(t, process.ErrRelayedTxV3SenderShardMismatch, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		proc, err := transaction.NewRelayedTxV3Processor(&economicsmocks.EconomicsHandlerStub{}, &testscommon.ShardsCoordinatorMock{})
		require.NoError(t, err)

		tx := getDefaultTx()
		err = proc.CheckRelayedTx(tx)
		require.NoError(t, err)
	})
}

func TestRelayedTxV3Processor_ComputeRelayedTxFees(t *testing.T) {
	t.Parallel()

	economicsFeeHandler := &economicsmocks.EconomicsHandlerStub{
		ComputeMoveBalanceFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return big.NewInt(int64(minGasLimit * tx.GetGasPrice()))
		},
	}
	proc, err := transaction.NewRelayedTxV3Processor(economicsFeeHandler, &testscommon.ShardsCoordinatorMock{})
	require.NoError(t, err)

	tx := getDefaultTx()
	relayerFee, totalFee := proc.ComputeRelayedTxFees(tx)
	expectedRelayerFee := big.NewInt(int64(2 * minGasLimit * tx.GetGasPrice())) // 2 move balance
	require.Equal(t, expectedRelayerFee, relayerFee)
	require.Equal(t, big.NewInt(int64(tx.GetGasLimit()*tx.GetGasPrice())), totalFee)
}
