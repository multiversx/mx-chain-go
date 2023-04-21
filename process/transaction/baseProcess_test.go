package transaction

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/guardianMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_checkOperationAllowedToBypassGuardian(t *testing.T) {
	t.Run("operations not allowed to bypass", func(t *testing.T) {
		txData := []byte("#@!")
		require.True(t, errors.Is(checkOperationAllowedToBypassGuardian(txData), process.ErrTransactionNotExecutable))
		txData = []byte(nil)
		require.True(t, errors.Is(checkOperationAllowedToBypassGuardian(txData), process.ErrTransactionNotExecutable))
		txData = []byte("SomeOtherFunction@")
		require.True(t, errors.Is(checkOperationAllowedToBypassGuardian(txData), process.ErrTransactionNotExecutable))
	})
	t.Run("setGuardian data field (non builtin call) not allowed", func(t *testing.T) {
		txData := []byte("setGuardian")
		require.True(t, errors.Is(checkOperationAllowedToBypassGuardian(txData), process.ErrTransactionNotExecutable))
	})
	t.Run("set guardian builtin call allowed to bypass", func(t *testing.T) {
		txData := []byte("SetGuardian@")
		require.Nil(t, checkOperationAllowedToBypassGuardian(txData))
	})
}

func Test_checkGuardedAccountUnguardedTxPermission(t *testing.T) {

	baseProc := baseTxProcessor{
		accounts:         &stateMock.AccountsStub{},
		shardCoordinator: mock.NewOneShardCoordinatorMock(),
		pubkeyConv:       testscommon.NewPubkeyConverterMock(32),
		economicsFee: &economicsmocks.EconomicsHandlerStub{
			CheckValidityTxValuesCalled: func(tx data.TransactionWithFeeHandler) error {
				return nil
			},
			ComputeMoveBalanceFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
				return big.NewInt(0)
			},
		},
		hasher:      &hashingMocks.HasherMock{},
		marshalizer: &testscommon.MarshalizerMock{},
		scProcessor: &testscommon.SCProcessorMock{},
		enableEpochsHandler: &testscommon.EnableEpochsHandlerStub{
			IsPenalizedTooMuchGasFlagEnabledField: true,
		},
		txVersionChecker: &testscommon.TxVersionCheckerStub{},
		guardianChecker:  &guardianMocks.GuardedAccountHandlerStub{},
	}

	account := &stateMock.UserAccountStub{}

	t.Run("nil txData", func(t *testing.T) {
		require.True(t, errors.Is(baseProc.checkGuardedAccountUnguardedTxPermission(nil, account), process.ErrTransactionNotExecutable))
	})
	t.Run("empty txData", func(t *testing.T) {
		require.True(t, errors.Is(baseProc.checkGuardedAccountUnguardedTxPermission([]byte(""), account), process.ErrTransactionNotExecutable))
	})
	t.Run("nil account", func(t *testing.T) {
		txData := []byte("SetGuardian@")
		require.Nil(t, baseProc.checkGuardedAccountUnguardedTxPermission(txData, nil))
	})
	t.Run("setGuardian data field (non builtin call) not allowed", func(t *testing.T) {
		txData := []byte("setGuardian")
		require.True(t, errors.Is(baseProc.checkGuardedAccountUnguardedTxPermission(txData, account), process.ErrTransactionNotExecutable))
	})
	t.Run("set guardian builtin call allowed to bypass", func(t *testing.T) {
		txData := []byte("SetGuardian@")
		require.Nil(t, baseProc.checkGuardedAccountUnguardedTxPermission(txData, account))
	})
	t.Run("set guardian builtin call with pending guardian not allowed", func(t *testing.T) {
		txData := []byte("SetGuardian@")
		baseProcLocal := baseProc
		baseProcLocal.guardianChecker = &guardianMocks.GuardedAccountHandlerStub{
			HasPendingGuardianCalled: func(uah state.UserAccountHandler) bool {
				return true
			},
		}

		err := baseProcLocal.checkGuardedAccountUnguardedTxPermission(txData, account)
		require.True(t, errors.Is(err, process.ErrTransactionNotExecutable))
		require.True(t, strings.Contains(err.Error(), process.ErrCannotReplaceGuardedAccountPendingGuardian.Error()))
	})
}

func TestBaseTxProcessor_VerifyGuardian(t *testing.T) {
	t.Parallel()

	baseProc := baseTxProcessor{
		accounts:         &stateMock.AccountsStub{},
		shardCoordinator: mock.NewOneShardCoordinatorMock(),
		pubkeyConv:       testscommon.NewPubkeyConverterMock(32),
		economicsFee: &economicsmocks.EconomicsHandlerStub{
			CheckValidityTxValuesCalled: func(tx data.TransactionWithFeeHandler) error {
				return nil
			},
			ComputeMoveBalanceFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
				return big.NewInt(0)
			},
		},
		hasher:      &hashingMocks.HasherMock{},
		marshalizer: &testscommon.MarshalizerMock{},
		scProcessor: &testscommon.SCProcessorMock{},
		enableEpochsHandler: &testscommon.EnableEpochsHandlerStub{
			IsPenalizedTooMuchGasFlagEnabledField: true,
		},
		txVersionChecker: &testscommon.TxVersionCheckerStub{},
		guardianChecker:  &guardianMocks.GuardedAccountHandlerStub{},
	}

	notGuardedAccount := &stateMock.UserAccountStub{}
	guardedAccount := &stateMock.UserAccountStub{
		IsGuardedCalled: func() bool {
			return true
		},
	}
	expectedErr := errors.New("expected error")
	tx := &transaction.Transaction{
		GuardianAddr: []byte("guardian"),
	}

	t.Run("nil account should not error", func(t *testing.T) {
		t.Parallel()

		localBaseProc := baseProc
		err := localBaseProc.verifyGuardian(&transaction.Transaction{}, nil)
		assert.Nil(t, err)
	})
	t.Run("guarded account with a not guarded transaction should error", func(t *testing.T) {
		t.Parallel()

		localBaseProc := baseProc
		localBaseProc.txVersionChecker = &testscommon.TxVersionCheckerStub{
			IsGuardedTransactionCalled: func(tx *transaction.Transaction) bool {
				return false
			},
		}

		err := localBaseProc.verifyGuardian(&transaction.Transaction{}, guardedAccount)
		assert.ErrorIs(t, err, process.ErrTransactionNotExecutable)
		assert.Contains(t, err.Error(), "not allowed to bypass guardian")
	})
	t.Run("not guarded account with guarded tx should error", func(t *testing.T) {
		t.Parallel()

		localBaseProc := baseProc
		localBaseProc.txVersionChecker = &testscommon.TxVersionCheckerStub{
			IsGuardedTransactionCalled: func(tx *transaction.Transaction) bool {
				return true
			},
		}

		err := localBaseProc.verifyGuardian(&transaction.Transaction{}, notGuardedAccount)
		assert.ErrorIs(t, err, process.ErrTransactionNotExecutable)
		assert.Contains(t, err.Error(), process.ErrGuardedTransactionNotExpected.Error())
	})
	t.Run("not guarded account with not guarded tx should work", func(t *testing.T) {
		t.Parallel()

		localBaseProc := baseProc
		localBaseProc.txVersionChecker = &testscommon.TxVersionCheckerStub{
			IsGuardedTransactionCalled: func(tx *transaction.Transaction) bool {
				return false
			},
		}

		err := localBaseProc.verifyGuardian(&transaction.Transaction{}, notGuardedAccount)
		assert.Nil(t, err)
	})
	t.Run("get active guardian fails should error", func(t *testing.T) {
		t.Parallel()

		localBaseProc := baseProc
		localBaseProc.txVersionChecker = &testscommon.TxVersionCheckerStub{
			IsGuardedTransactionCalled: func(tx *transaction.Transaction) bool {
				return true
			},
		}
		localBaseProc.guardianChecker = &guardianMocks.GuardedAccountHandlerStub{
			GetActiveGuardianCalled: func(uah vmcommon.UserAccountHandler) ([]byte, error) {
				return nil, expectedErr
			},
		}

		err := localBaseProc.verifyGuardian(&transaction.Transaction{}, guardedAccount)
		assert.ErrorIs(t, err, process.ErrTransactionNotExecutable)
		assert.Contains(t, err.Error(), expectedErr.Error())
	})
	t.Run("guardian address mismatch should error", func(t *testing.T) {
		t.Parallel()

		localBaseProc := baseProc
		localBaseProc.txVersionChecker = &testscommon.TxVersionCheckerStub{
			IsGuardedTransactionCalled: func(tx *transaction.Transaction) bool {
				return true
			},
		}
		localBaseProc.guardianChecker = &guardianMocks.GuardedAccountHandlerStub{
			GetActiveGuardianCalled: func(uah vmcommon.UserAccountHandler) ([]byte, error) {
				return []byte("account guardian"), nil
			},
		}

		err := localBaseProc.verifyGuardian(tx, guardedAccount)
		assert.ErrorIs(t, err, process.ErrTransactionNotExecutable)
		assert.Contains(t, err.Error(), process.ErrTransactionAndAccountGuardianMismatch.Error())
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		localBaseProc := baseProc
		localBaseProc.txVersionChecker = &testscommon.TxVersionCheckerStub{
			IsGuardedTransactionCalled: func(tx *transaction.Transaction) bool {
				return true
			},
		}
		localBaseProc.guardianChecker = &guardianMocks.GuardedAccountHandlerStub{
			GetActiveGuardianCalled: func(uah vmcommon.UserAccountHandler) ([]byte, error) {
				return []byte("guardian"), nil
			},
		}

		err := localBaseProc.verifyGuardian(tx, guardedAccount)
		assert.Nil(t, err)
	})
}
