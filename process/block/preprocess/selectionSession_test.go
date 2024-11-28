package preprocess

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewSelectionSession(t *testing.T) {
	t.Parallel()

	session, err := newSelectionSession(argsSelectionSession{
		accountsAdapter:       nil,
		transactionsProcessor: &testscommon.TxProcessorStub{},
		marshalizer:           &marshal.GogoProtoMarshalizer{},
	})
	require.Nil(t, session)
	require.ErrorIs(t, err, process.ErrNilAccountsAdapter)

	session, err = newSelectionSession(argsSelectionSession{
		accountsAdapter:       &stateMock.AccountsStub{},
		transactionsProcessor: nil,
		marshalizer:           &marshal.GogoProtoMarshalizer{},
	})
	require.Nil(t, session)
	require.ErrorIs(t, err, process.ErrNilTxProcessor)

	session, err = newSelectionSession(argsSelectionSession{
		accountsAdapter:       &stateMock.AccountsStub{},
		transactionsProcessor: &testscommon.TxProcessorStub{},
		marshalizer:           &marshal.GogoProtoMarshalizer{},
	})
	require.NoError(t, err)
	require.NotNil(t, session)
}

func TestSelectionSession_GetAccountState(t *testing.T) {
	t.Parallel()

	accounts := &stateMock.AccountsStub{}
	processor := &testscommon.TxProcessorStub{}

	accounts.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, []byte("alice")) {
			return &stateMock.UserAccountStub{
				Address: []byte("alice"),
				Nonce:   42,
			}, nil
		}

		if bytes.Equal(address, []byte("bob")) {
			return &stateMock.UserAccountStub{
				Address: []byte("bob"),
				Nonce:   7,
				IsGuardedCalled: func() bool {
					return true
				},
			}, nil
		}

		return nil, fmt.Errorf("account not found: %s", address)
	}

	session, err := newSelectionSession(argsSelectionSession{
		accountsAdapter:       accounts,
		transactionsProcessor: processor,
		marshalizer:           &marshal.GogoProtoMarshalizer{},
	})
	require.NoError(t, err)
	require.NotNil(t, session)

	state, err := session.GetAccountState([]byte("alice"))
	require.NoError(t, err)
	require.Equal(t, uint64(42), state.Nonce)

	state, err = session.GetAccountState([]byte("bob"))
	require.NoError(t, err)
	require.Equal(t, uint64(7), state.Nonce)

	state, err = session.GetAccountState([]byte("carol"))
	require.ErrorContains(t, err, "account not found: carol")
	require.Nil(t, state)
}

func TestSelectionSession_IsIncorrectlyGuarded(t *testing.T) {
	t.Parallel()

	accounts := &stateMock.AccountsStub{}
	processor := &testscommon.TxProcessorStub{}

	accounts.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, []byte("bob")) {
			return &stateMock.BaseAccountMock{}, nil
		}

		return &stateMock.UserAccountStub{}, nil
	}

	processor.VerifyGuardianCalled = func(tx *transaction.Transaction, account state.UserAccountHandler) error {
		if tx.Nonce == 43 {
			return process.ErrTransactionNotExecutable
		}
		if tx.Nonce == 44 {
			return fmt.Errorf("arbitrary processing error")
		}

		return nil
	}

	session, err := newSelectionSession(argsSelectionSession{
		accountsAdapter:       accounts,
		transactionsProcessor: processor,
		marshalizer:           &marshal.GogoProtoMarshalizer{},
	})
	require.NoError(t, err)
	require.NotNil(t, session)

	isIncorrectlyGuarded := session.IsIncorrectlyGuarded(&transaction.Transaction{Nonce: 42, SndAddr: []byte("alice")})
	require.False(t, isIncorrectlyGuarded)

	isIncorrectlyGuarded = session.IsIncorrectlyGuarded(&transaction.Transaction{Nonce: 43, SndAddr: []byte("alice")})
	require.True(t, isIncorrectlyGuarded)

	isIncorrectlyGuarded = session.IsIncorrectlyGuarded(&transaction.Transaction{Nonce: 44, SndAddr: []byte("alice")})
	require.False(t, isIncorrectlyGuarded)

	isIncorrectlyGuarded = session.IsIncorrectlyGuarded(&transaction.Transaction{Nonce: 45, SndAddr: []byte("bob")})
	require.True(t, isIncorrectlyGuarded)
}

func TestSelectionSession_GetTransferredValue(t *testing.T) {
	t.Parallel()

	session, err := newSelectionSession(argsSelectionSession{
		accountsAdapter:       &stateMock.AccountsStub{},
		transactionsProcessor: &testscommon.TxProcessorStub{},
		marshalizer:           &marshal.GogoProtoMarshalizer{},
	})
	require.NoError(t, err)
	require.NotNil(t, session)

	t.Run("with value", func(t *testing.T) {
		value := session.GetTransferredValue(&transaction.Transaction{
			Value: big.NewInt(1000000000000000000),
		})
		require.Equal(t, big.NewInt(1000000000000000000), value)
	})

	t.Run("with value and data", func(t *testing.T) {
		value := session.GetTransferredValue(&transaction.Transaction{
			Value: big.NewInt(1000000000000000000),
			Data:  []byte("data"),
		})
		require.Equal(t, big.NewInt(1000000000000000000), value)
	})

	t.Run("native transfer within MultiESDTNFTTransfer", func(t *testing.T) {
		value := session.GetTransferredValue(&transaction.Transaction{
			SndAddr: testscommon.TestPubKeyAlice,
			RcvAddr: testscommon.TestPubKeyAlice,
			Data:    []byte("MultiESDTNFTTransfer@8049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f8@03@4e46542d313233343536@0a@01@544553542d393837363534@01@01@45474c442d303030303030@@0de0b6b3a7640000"),
		})
		require.Equal(t, big.NewInt(1000000000000000000), value)
	})

	t.Run("native transfer within MultiESDTNFTTransfer; transfer & execute", func(t *testing.T) {
		value := session.GetTransferredValue(&transaction.Transaction{
			SndAddr: testscommon.TestPubKeyAlice,
			RcvAddr: testscommon.TestPubKeyAlice,
			Data:    []byte("MultiESDTNFTTransfer@00000000000000000500b9353fe8407f87310c87e12fa1ac807f0485da39d152@03@4e46542d313233343536@01@01@4e46542d313233343536@2a@01@45474c442d303030303030@@0de0b6b3a7640000@64756d6d79@07"),
		})
		require.Equal(t, big.NewInt(1000000000000000000), value)
	})
}

func TestBenchmarkSelectionSession_GetTransferredValue(t *testing.T) {
	session, err := newSelectionSession(argsSelectionSession{
		accountsAdapter:       &stateMock.AccountsStub{},
		transactionsProcessor: &testscommon.TxProcessorStub{},
		marshalizer:           &marshal.GogoProtoMarshalizer{},
	})
	require.NoError(t, err)
	require.NotNil(t, session)

	sw := core.NewStopWatch()

	valueMultiplier := int64(1_000_000_000_000)

	t.Run("numTransactions = 5_000", func(t *testing.T) {
		numTransactions := 5_000
		transactions := createMultiESDTNFTTransfersWithNativeTransfer(numTransactions, valueMultiplier)

		sw.Start(t.Name())

		for i := 0; i < numTransactions; i++ {
			tx := transactions[i]
			value := session.GetTransferredValue(tx)
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
			value := session.GetTransferredValue(tx)
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
			value := session.GetTransferredValue(tx)
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
	// 0.012993s (TestBenchmarkSelectionSession_GetTransferredValue/numTransactions_=_5_000)
	// 0.024580s (TestBenchmarkSelectionSession_GetTransferredValue/numTransactions_=_10_000)
	// 0.048808s (TestBenchmarkSelectionSession_GetTransferredValue/numTransactions_=_20_000)
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
