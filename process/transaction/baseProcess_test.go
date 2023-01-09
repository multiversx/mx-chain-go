package transaction

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/guardianMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	"github.com/stretchr/testify/require"
)

func Test_checkOperationAllowedToBypassGuardian(t *testing.T) {
	t.Run("operations not allowed to bypass", func(t *testing.T) {
		txData := []byte("#@!")
		require.Equal(t, process.ErrOperationNotPermitted, checkOperationAllowedToBypassGuardian(txData))
		txData = []byte(nil)
		require.Equal(t, process.ErrOperationNotPermitted, checkOperationAllowedToBypassGuardian(txData))
		txData = []byte("SomeOtherFunction@")
		require.Equal(t, process.ErrOperationNotPermitted, checkOperationAllowedToBypassGuardian(txData))
	})
	t.Run("setGuardian data field (non builtin call) not allowed", func(t *testing.T) {
		txData := []byte("setGuardian")
		require.Equal(t, process.ErrOperationNotPermitted, checkOperationAllowedToBypassGuardian(txData))
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

	account:= &stateMock.UserAccountStub{}

	t.Run("nil txData", func(t *testing.T) {
		require.Equal(t, process.ErrOperationNotPermitted, baseProc.checkGuardedAccountUnguardedTxPermission(nil, account))
	})
	t.Run("empty txData", func(t *testing.T) {
		require.Equal(t, process.ErrOperationNotPermitted, baseProc.checkGuardedAccountUnguardedTxPermission([]byte(""), account))
	})
	t.Run("nil account", func(t *testing.T) {
		txData := []byte("SetGuardian@")
		require.Nil(t, baseProc.checkGuardedAccountUnguardedTxPermission(txData, nil))
	})
	t.Run("setGuardian data field (non builtin call) not allowed", func(t *testing.T) {
		txData := []byte("setGuardian")
		require.Equal(t, process.ErrOperationNotPermitted, baseProc.checkGuardedAccountUnguardedTxPermission(txData, account))
	})
	t.Run("set guardian builtin call allowed to bypass", func(t *testing.T) {
		txData := []byte("SetGuardian@")
		require.Nil(t, baseProc.checkGuardedAccountUnguardedTxPermission(txData, account))
	})
	t.Run("set guardian builtin call with pending guardian not allowed", func(t *testing.T) {
		txData := []byte("SetGuardian@")
		baseProc :=	baseProc
		baseProc.guardianChecker = &guardianMocks.GuardedAccountHandlerStub{
			HasPendingGuardianCalled: func(uah state.UserAccountHandler) bool {
				return true
			},
		}
		require.Equal(t, process.ErrCannotReplaceGuardedAccountPendingGuardian, baseProc.checkGuardedAccountUnguardedTxPermission(txData, account))
	})
}
