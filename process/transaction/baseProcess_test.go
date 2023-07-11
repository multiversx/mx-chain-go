package transaction

import (
	"errors"
	"fmt"
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
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/guardianMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var flagActiveTrueHandler = func(epoch uint32) bool { return true }

func createMockBaseTxProcessor() *baseTxProcessor {
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
		marshalizer: &marshallerMock.MarshalizerMock{},
		scProcessor: &testscommon.SCProcessorMock{},
		enableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsPenalizedTooMuchGasFlagEnabledInEpochCalled: flagActiveTrueHandler,
		},
		txVersionChecker: &testscommon.TxVersionCheckerStub{},
		guardianChecker:  &guardianMocks.GuardedAccountHandlerStub{},
	}

	return &baseProc
}

func Test_checkOperationAllowedToBypassGuardian(t *testing.T) {
	tx := &transaction.Transaction{}

	t.Run("operations not allowed to bypass", func(t *testing.T) {
		txCopy := *tx
		baseProc := createMockBaseTxProcessor()
		baseProc.scProcessor = &testscommon.SCProcessorMock{
			CheckBuiltinFunctionIsExecutableCalled: func(expectedBuiltinFunction string, tx data.TransactionHandler) error {
				return nil
			},
		}

		txCopy.Data = []byte("#@!")
		require.True(t, errors.Is(baseProc.checkOperationAllowedToBypassGuardian(&txCopy), process.ErrTransactionNotExecutable))
		txCopy.Data = []byte(nil)
		require.True(t, errors.Is(baseProc.checkOperationAllowedToBypassGuardian(&txCopy), process.ErrTransactionNotExecutable))
		txCopy.Data = []byte("SomeOtherFunction@")
		require.True(t, errors.Is(baseProc.checkOperationAllowedToBypassGuardian(&txCopy), process.ErrTransactionNotExecutable))
	})
	t.Run("setGuardian data field (non builtin call) not allowed", func(t *testing.T) {
		txCopy := *tx
		baseProc := createMockBaseTxProcessor()
		baseProc.scProcessor = &testscommon.SCProcessorMock{
			CheckBuiltinFunctionIsExecutableCalled: func(expectedBuiltinFunction string, tx data.TransactionHandler) error {
				return nil
			},
		}
		txCopy.Data = []byte("setGuardian")
		require.True(t, errors.Is(baseProc.checkOperationAllowedToBypassGuardian(&txCopy), process.ErrTransactionNotExecutable))

	})
	t.Run("set guardian builtin call not allowed to bypass if not executable check fails", func(t *testing.T) {
		txCopy := *tx
		baseProc := createMockBaseTxProcessor()
		expectedError := fmt.Errorf("expected error")
		baseProc.scProcessor = &testscommon.SCProcessorMock{
			CheckBuiltinFunctionIsExecutableCalled: func(expectedBuiltinFunction string, tx data.TransactionHandler) error {
				return expectedError
			},
		}

		txCopy.Data = []byte("SetGuardian@")
		err := baseProc.checkOperationAllowedToBypassGuardian(&txCopy)
		require.True(t, errors.Is(err, process.ErrTransactionNotExecutable))
		require.True(t, strings.Contains(err.Error(), expectedError.Error()))
	})
	t.Run("set guardian builtin call not allowed to bypass if transaction has sender username", func(t *testing.T) {
		txCopy := *tx
		baseProc := createMockBaseTxProcessor()
		baseProc.scProcessor = &testscommon.SCProcessorMock{
			CheckBuiltinFunctionIsExecutableCalled: func(expectedBuiltinFunction string, tx data.TransactionHandler) error {
				return nil
			},
		}
		txCopy.Data = []byte("SetGuardian@")
		txCopy.SndUserName = []byte("someUsername")
		err := baseProc.checkOperationAllowedToBypassGuardian(&txCopy)
		require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
		require.True(t, strings.Contains(err.Error(), "username"))
	})
	t.Run("set guardian builtin call not allowed to bypass if transaction has receiver username address", func(t *testing.T) {
		txCopy := *tx
		baseProc := createMockBaseTxProcessor()
		baseProc.scProcessor = &testscommon.SCProcessorMock{
			CheckBuiltinFunctionIsExecutableCalled: func(expectedBuiltinFunction string, tx data.TransactionHandler) error {
				return nil
			},
		}
		txCopy.Data = []byte("SetGuardian@")
		txCopy.RcvUserName = []byte("someUsername")
		err := baseProc.checkOperationAllowedToBypassGuardian(&txCopy)
		require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
		require.True(t, strings.Contains(err.Error(), "username"))
	})
	t.Run("set guardian builtin call ok", func(t *testing.T) {
		txCopy := *tx
		baseProc := createMockBaseTxProcessor()
		baseProc.scProcessor = &testscommon.SCProcessorMock{
			CheckBuiltinFunctionIsExecutableCalled: func(expectedBuiltinFunction string, tx data.TransactionHandler) error {
				return nil
			},
		}
		txCopy.Data = []byte("SetGuardian@")
		require.Nil(t, baseProc.checkOperationAllowedToBypassGuardian(&txCopy))
	})
}

func Test_checkGuardedAccountUnguardedTxPermission(t *testing.T) {
	account := &stateMock.UserAccountStub{}
	baseProc := createMockBaseTxProcessor()

	t.Run("nil txData", func(t *testing.T) {
		tx := &transaction.Transaction{
			Data: nil,
		}

		require.True(t, errors.Is(baseProc.checkGuardedAccountUnguardedTxPermission(tx, account), process.ErrTransactionNotExecutable))
	})
	t.Run("empty txData", func(t *testing.T) {
		tx := &transaction.Transaction{
			Data: []byte(""),
		}

		require.True(t, errors.Is(baseProc.checkGuardedAccountUnguardedTxPermission(tx, account), process.ErrTransactionNotExecutable))
	})
	t.Run("nil account", func(t *testing.T) {
		tx := &transaction.Transaction{
			Data: []byte("SetGuardian@"),
		}
		require.Nil(t, baseProc.checkGuardedAccountUnguardedTxPermission(tx, nil))
	})
	t.Run("setGuardian data field (non builtin call) not allowed", func(t *testing.T) {
		tx := &transaction.Transaction{
			Data: []byte("setGuardian"),
		}

		require.True(t, errors.Is(baseProc.checkGuardedAccountUnguardedTxPermission(tx, account), process.ErrTransactionNotExecutable))
	})
	t.Run("set guardian builtin call allowed to bypass", func(t *testing.T) {
		tx := &transaction.Transaction{
			Data: []byte("SetGuardian@"),
		}
		require.Nil(t, baseProc.checkGuardedAccountUnguardedTxPermission(tx, account))
	})
	t.Run("set guardian builtin call with pending guardian not allowed", func(t *testing.T) {
		tx := &transaction.Transaction{
			Data: []byte("SetGuardian@"),
		}

		baseProcLocal := baseProc
		baseProcLocal.guardianChecker = &guardianMocks.GuardedAccountHandlerStub{
			HasPendingGuardianCalled: func(uah state.UserAccountHandler) bool {
				return true
			},
		}

		err := baseProcLocal.checkGuardedAccountUnguardedTxPermission(tx, account)
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
		marshalizer: &marshallerMock.MarshalizerMock{},
		scProcessor: &testscommon.SCProcessorMock{},
		enableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsPenalizedTooMuchGasFlagEnabledInEpochCalled: flagActiveTrueHandler,
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

func Test_CheckSetGuardianExecutable(t *testing.T) {
	t.Run("sc processor CheckBuiltinFunctionIsExecutable with error should error with ErrTransactionNotExecutable", func(t *testing.T) {
		t.Parallel()

		baseProc := createMockBaseTxProcessor()
		expectedErr := errors.New("expected error")

		baseProc.scProcessor = &testscommon.SCProcessorMock{
			CheckBuiltinFunctionIsExecutableCalled: func(expectedBuiltinFunction string, tx data.TransactionHandler) error {
				return expectedErr
			},
		}
		tx := &transaction.Transaction{}
		err := baseProc.CheckSetGuardianExecutable(tx)
		require.NotNil(t, err)
		require.True(t, errors.Is(err, process.ErrTransactionNotExecutable))
		require.True(t, strings.Contains(err.Error(), expectedErr.Error()))
	})
	t.Run("sc processor CheckBuiltinFunctionIsExecutable with no error OK", func(t *testing.T) {
		t.Parallel()

		baseProc := createMockBaseTxProcessor()
		baseProc.scProcessor = &testscommon.SCProcessorMock{
			CheckBuiltinFunctionIsExecutableCalled: func(expectedBuiltinFunction string, tx data.TransactionHandler) error {
				return nil
			},
		}
		tx := &transaction.Transaction{}
		err := baseProc.CheckSetGuardianExecutable(tx)
		require.Nil(t, err)
	})
}
