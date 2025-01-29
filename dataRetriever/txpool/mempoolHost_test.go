package txpool

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func TestNewMempoolHost(t *testing.T) {
	t.Parallel()

	host, err := newMempoolHost(argsMempoolHost{
		txGasHandler: nil,
		marshalizer:  &marshal.GogoProtoMarshalizer{},
	})
	require.Nil(t, host)
	require.ErrorIs(t, err, dataRetriever.ErrNilTxGasHandler)

	host, err = newMempoolHost(argsMempoolHost{
		txGasHandler: txcachemocks.NewTxGasHandlerMock(),
		marshalizer:  nil,
	})
	require.Nil(t, host)
	require.ErrorIs(t, err, dataRetriever.ErrNilMarshalizer)

	host, err = newMempoolHost(argsMempoolHost{
		txGasHandler: txcachemocks.NewTxGasHandlerMock(),
		marshalizer:  &marshal.GogoProtoMarshalizer{},
	})
	require.NoError(t, err)
	require.NotNil(t, host)
}

func TestMempoolHost_GetTransferredValue(t *testing.T) {
	t.Parallel()

	host, err := newMempoolHost(argsMempoolHost{
		txGasHandler: txcachemocks.NewTxGasHandlerMock(),
		marshalizer:  &marshal.GogoProtoMarshalizer{},
	})
	require.NoError(t, err)
	require.NotNil(t, host)

	t.Run("with value", func(t *testing.T) {
		value := host.GetTransferredValue(&transaction.Transaction{
			Value: big.NewInt(1000000000000000000),
		})
		require.Equal(t, big.NewInt(1000000000000000000), value)
	})

	t.Run("with value and data", func(t *testing.T) {
		value := host.GetTransferredValue(&transaction.Transaction{
			Value: big.NewInt(1000000000000000000),
			Data:  []byte("data"),
		})
		require.Equal(t, big.NewInt(1000000000000000000), value)
	})

	t.Run("native transfer within MultiESDTNFTTransfer", func(t *testing.T) {
		value := host.GetTransferredValue(&transaction.Transaction{
			SndAddr: testscommon.TestPubKeyAlice,
			RcvAddr: testscommon.TestPubKeyAlice,
			Data:    []byte("MultiESDTNFTTransfer@8049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f8@03@4e46542d313233343536@0a@01@544553542d393837363534@01@01@45474c442d303030303030@@0de0b6b3a7640000"),
		})
		require.Equal(t, big.NewInt(1000000000000000000), value)
	})

	t.Run("native transfer within MultiESDTNFTTransfer; transfer & execute", func(t *testing.T) {
		value := host.GetTransferredValue(&transaction.Transaction{
			SndAddr: testscommon.TestPubKeyAlice,
			RcvAddr: testscommon.TestPubKeyAlice,
			Data:    []byte("MultiESDTNFTTransfer@00000000000000000500b9353fe8407f87310c87e12fa1ac807f0485da39d152@03@4e46542d313233343536@01@01@4e46542d313233343536@2a@01@45474c442d303030303030@@0de0b6b3a7640000@64756d6d79@07"),
		})
		require.Equal(t, big.NewInt(1000000000000000000), value)
	})
}

func TestBenchmarkMempoolHost_GetTransferredValue(t *testing.T) {
	host, err := newMempoolHost(argsMempoolHost{
		txGasHandler: txcachemocks.NewTxGasHandlerMock(),
		marshalizer:  &marshal.GogoProtoMarshalizer{},
	})
	require.NoError(t, err)
	require.NotNil(t, host)

	sw := core.NewStopWatch()

	valueMultiplier := int64(1_000_000_000_000)

	t.Run("numTransactions = 5_000", func(t *testing.T) {
		numTransactions := 5_000
		transactions := createMultiESDTNFTTransfersWithNativeTransfer(numTransactions, valueMultiplier)

		sw.Start(t.Name())

		for i := 0; i < numTransactions; i++ {
			tx := transactions[i]
			value := host.GetTransferredValue(tx)
			require.Equal(t, big.NewInt(int64(i)*valueMultiplier), value)
		}

		sw.Stop(t.Name())
	})

	t.Run("numTransactions = 10_000", func(t *testing.T) {
		numTransactions := 10_000
		transactions := createMultiESDTNFTTransfersWithNativeTransfer(numTransactions, valueMultiplier)

		sw.Start(t.Name())

		for i := 0; i < numTransactions; i++ {
			tx := transactions[i]
			value := host.GetTransferredValue(tx)
			require.Equal(t, big.NewInt(int64(i)*valueMultiplier), value)
		}

		sw.Stop(t.Name())
	})

	t.Run("numTransactions = 20_000", func(t *testing.T) {
		numTransactions := 20_000
		transactions := createMultiESDTNFTTransfersWithNativeTransfer(numTransactions, valueMultiplier)

		sw.Start(t.Name())

		for i := 0; i < numTransactions; i++ {
			tx := transactions[i]
			value := host.GetTransferredValue(tx)
			require.Equal(t, big.NewInt(int64(i)*valueMultiplier), value)
		}

		sw.Stop(t.Name())
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}

	// (1)
	// Vendor ID:                GenuineIntel
	//   Model name:             11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
	//     CPU family:           6
	//     Model:                140
	//     Thread(s) per core:   2
	//     Core(s) per socket:   4
	//
	// NOTE: 20% is also due to the require() / assert() calls.
	// 0.012993s (TestBenchmarkMempoolHost_GetTransferredValue/numTransactions_=_5_000)
	// 0.024580s (TestBenchmarkMempoolHost_GetTransferredValue/numTransactions_=_10_000)
	// 0.048808s (TestBenchmarkMempoolHost_GetTransferredValue/numTransactions_=_20_000)
}

func createMultiESDTNFTTransfersWithNativeTransfer(numTransactions int, valueMultiplier int64) []*transaction.Transaction {
	transactions := make([]*transaction.Transaction, 0, numTransactions)

	for i := 0; i < numTransactions; i++ {
		nativeValue := big.NewInt(int64(i) * valueMultiplier)
		data := fmt.Sprintf(
			"MultiESDTNFTTransfer@8049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f8@03@4e46542d313233343536@0a@01@544553542d393837363534@01@01@45474c442d303030303030@@%s",
			hex.EncodeToString(nativeValue.Bytes()),
		)

		tx := &transaction.Transaction{
			SndAddr: testscommon.TestPubKeyAlice,
			RcvAddr: testscommon.TestPubKeyAlice,
			Data:    []byte(data),
		}

		transactions = append(transactions, tx)
	}

	return transactions
}
