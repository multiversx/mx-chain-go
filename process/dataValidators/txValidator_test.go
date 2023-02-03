package dataValidators_test

import (
	"errors"
	"math/big"
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/dataValidators"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getAccAdapter(nonce uint64, balance *big.Int) *stateMock.AccountsStub {
	accDB := &stateMock.AccountsStub{}
	accDB.GetExistingAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		acc, _ := state.NewUserAccount(address)
		acc.Nonce = nonce
		acc.Balance = balance

		return acc, nil
	}

	return accDB
}

func createMockCoordinator(identifierPrefix string, currentShardID uint32) *mock.CoordinatorStub {
	return &mock.CoordinatorStub{
		CommunicationIdentifierCalled: func(destShardID uint32) string {
			return identifierPrefix + strconv.Itoa(int(destShardID))
		},
		SelfIdCalled: func() uint32 {
			return currentShardID
		},
	}
}

func getInterceptedTxHandler(
	sndShardId uint32,
	rcvShardId uint32,
	nonce uint64,
	sndAddr []byte,
	fee *big.Int,
) process.InterceptedTransactionHandler {
	return &mock.InterceptedTxHandlerStub{
		SenderShardIdCalled: func() uint32 {
			return sndShardId
		},
		ReceiverShardIdCalled: func() uint32 {
			return rcvShardId
		},
		NonceCalled: func() uint64 {
			return nonce
		},
		SenderAddressCalled: func() []byte {
			return sndAddr
		},
		FeeCalled: func() *big.Int {
			return fee
		},
		TransactionCalled: func() data.TransactionHandler {
			return &transaction.Transaction{}
		},
	}
}

func TestNewTxValidator_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(
		nil,
		shardCoordinator,
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.TxVersionCheckerStub{},
		maxNonceDeltaAllowed,
	)

	assert.Nil(t, txValidator)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewTxValidator_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	adb := getAccAdapter(0, big.NewInt(0))
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(
		adb,
		nil,
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.TxVersionCheckerStub{},
		maxNonceDeltaAllowed,
	)

	assert.Nil(t, txValidator)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestTxValidator_NewValidatorNilWhiteListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	adb := getAccAdapter(0, big.NewInt(0))
	maxNonceDeltaAllowed := 100
	shardCoordinator := createMockCoordinator("_", 0)
	txValidator, err := dataValidators.NewTxValidator(
		adb,
		shardCoordinator,
		nil,
		mock.NewPubkeyConverterMock(32),
		&testscommon.TxVersionCheckerStub{},
		maxNonceDeltaAllowed,
	)

	assert.Nil(t, txValidator)
	assert.Equal(t, process.ErrNilWhiteListHandler, err)
}

func TestNewTxValidator_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	adb := getAccAdapter(0, big.NewInt(0))
	maxNonceDeltaAllowed := 100
	shardCoordinator := createMockCoordinator("_", 0)
	txValidator, err := dataValidators.NewTxValidator(
		adb,
		shardCoordinator,
		&testscommon.WhiteListHandlerStub{},
		nil,
		&testscommon.TxVersionCheckerStub{},
		maxNonceDeltaAllowed,
	)

	assert.Nil(t, txValidator)
	assert.True(t, errors.Is(err, process.ErrNilPubkeyConverter))
}

func TestNewTxValidator_NilTxVersionCheckerShouldErr(t *testing.T) {
	t.Parallel()

	adb := getAccAdapter(0, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(
		adb,
		shardCoordinator,
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		nil,
		maxNonceDeltaAllowed,
	)
	assert.Nil(t, txValidator)
	assert.True(t, errors.Is(err, process.ErrNilTransactionVersionChecker))
}

func TestNewTxValidator_ShouldWork(t *testing.T) {
	t.Parallel()

	adb := getAccAdapter(0, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(
		adb,
		shardCoordinator,
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.TxVersionCheckerStub{},
		maxNonceDeltaAllowed,
	)

	assert.Nil(t, err)
	assert.NotNil(t, txValidator)

	result := txValidator.IsInterfaceNil()
	assert.Equal(t, false, result)
}

func TestTxValidator_CheckTxValidityTxCrossShardShouldWork(t *testing.T) {
	t.Parallel()

	adb := getAccAdapter(1, big.NewInt(0))
	currentShard := uint32(0)
	shardCoordinator := createMockCoordinator("_", currentShard)
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(
		adb,
		shardCoordinator,
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.TxVersionCheckerStub{},
		maxNonceDeltaAllowed,
	)
	assert.Nil(t, err)

	addressMock := []byte("address")
	txValidatorHandler := getInterceptedTxHandler(currentShard+1, currentShard, 1, addressMock, big.NewInt(0))

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.Nil(t, result)
}

func TestTxValidator_CheckTxValidityAccountNonceIsGreaterThanTxNonceShouldReturnFalse(t *testing.T) {
	t.Parallel()

	accountNonce := uint64(100)
	txNonce := uint64(0)

	adb := getAccAdapter(accountNonce, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(
		adb,
		shardCoordinator,
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.TxVersionCheckerStub{},
		maxNonceDeltaAllowed,
	)
	assert.Nil(t, err)

	addressMock := []byte("address")
	currentShard := uint32(0)
	txValidatorHandler := getInterceptedTxHandler(currentShard, currentShard, txNonce, addressMock, big.NewInt(0))

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.True(t, errors.Is(result, process.ErrWrongTransaction))
}

func TestTxValidator_CheckTxValidityTxNonceIsTooHigh(t *testing.T) {
	t.Parallel()

	accountNonce := uint64(100)
	maxNonceDeltaAllowed := 100
	txNonce := accountNonce + uint64(maxNonceDeltaAllowed) + 1

	adb := getAccAdapter(accountNonce, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	txValidator, err := dataValidators.NewTxValidator(
		adb,
		shardCoordinator,
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.TxVersionCheckerStub{},
		maxNonceDeltaAllowed,
	)
	assert.Nil(t, err)

	addressMock := []byte("address")
	currentShard := uint32(0)
	txValidatorHandler := getInterceptedTxHandler(currentShard, currentShard, txNonce, addressMock, big.NewInt(0))

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.True(t, errors.Is(result, process.ErrWrongTransaction))
}

func TestTxValidator_CheckTxValidityAccountBalanceIsLessThanTxTotalValueShouldReturnFalse(t *testing.T) {
	t.Parallel()

	accountNonce := uint64(0)
	txNonce := uint64(1)
	fee := big.NewInt(1000)
	accountBalance := big.NewInt(10)

	adb := getAccAdapter(accountNonce, accountBalance)
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(
		adb,
		shardCoordinator,
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.TxVersionCheckerStub{},
		maxNonceDeltaAllowed,
	)
	assert.Nil(t, err)

	addressMock := []byte("address")
	currentShard := uint32(0)
	txValidatorHandler := getInterceptedTxHandler(currentShard, currentShard, txNonce, addressMock, fee)

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.NotNil(t, result)
	assert.True(t, errors.Is(result, process.ErrInsufficientFunds))
}

func TestTxValidator_CheckTxValidityAccountNotExitsShouldReturnFalse(t *testing.T) {
	t.Parallel()

	accDB := &stateMock.AccountsStub{}
	accDB.GetExistingAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return nil, errors.New("cannot find account")
	}
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, _ := dataValidators.NewTxValidator(
		accDB,
		shardCoordinator,
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.TxVersionCheckerStub{},
		maxNonceDeltaAllowed,
	)

	addressMock := []byte("address")
	currentShard := uint32(0)
	txValidatorHandler := getInterceptedTxHandler(currentShard, currentShard, 1, addressMock, big.NewInt(0))

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.True(t, errors.Is(result, process.ErrAccountNotFound))
}

func TestTxValidator_CheckTxValidityAccountNotExitsButWhiteListedShouldReturnTrue(t *testing.T) {
	t.Parallel()

	accDB := &stateMock.AccountsStub{}
	accDB.GetExistingAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return nil, errors.New("cannot find account")
	}
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, _ := dataValidators.NewTxValidator(
		accDB,
		shardCoordinator,
		&testscommon.WhiteListHandlerStub{
			IsWhiteListedCalled: func(interceptedData process.InterceptedData) bool {
				return true
			},
		},
		mock.NewPubkeyConverterMock(32),
		&testscommon.TxVersionCheckerStub{},
		maxNonceDeltaAllowed,
	)

	addressMock := []byte("address")
	currentShard := uint32(0)
	txValidatorHandler := getInterceptedTxHandler(currentShard, currentShard, 1, addressMock, big.NewInt(0))

	interceptedTx := struct {
		process.InterceptedData
		process.InterceptedTransactionHandler
	}{
		InterceptedData:               nil,
		InterceptedTransactionHandler: txValidatorHandler,
	}

	// interceptedTx needs to be of type InterceptedData & TxValidatorHandler
	result := txValidator.CheckTxValidity(interceptedTx)
	assert.Nil(t, result)
}

func TestTxValidator_CheckTxValidityWrongAccountTypeShouldReturnFalse(t *testing.T) {
	t.Parallel()

	accDB := &stateMock.AccountsStub{}
	accDB.GetExistingAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return state.NewPeerAccount(address)
	}
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, _ := dataValidators.NewTxValidator(
		accDB,
		shardCoordinator,
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.TxVersionCheckerStub{},
		maxNonceDeltaAllowed,
	)

	addressMock := []byte("address")
	currentShard := uint32(0)
	txValidatorHandler := getInterceptedTxHandler(currentShard, currentShard, 1, addressMock, big.NewInt(0))

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.True(t, errors.Is(result, process.ErrWrongTypeAssertion))
}

func TestTxValidator_CheckTxValidityTxIsOkShouldReturnTrue(t *testing.T) {
	t.Parallel()

	accountNonce := uint64(0)
	accountBalance := big.NewInt(10)
	adb := getAccAdapter(accountNonce, accountBalance)
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, _ := dataValidators.NewTxValidator(
		adb,
		shardCoordinator,
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.TxVersionCheckerStub{},
		maxNonceDeltaAllowed,
	)

	addressMock := []byte("address")
	currentShard := uint32(0)
	txValidatorHandler := getInterceptedTxHandler(currentShard, currentShard, 1, addressMock, big.NewInt(0))

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.Nil(t, result)
}

func Test_getTxData(t *testing.T) {
	t.Run("nil tx in intercepted tx returns error", func(t *testing.T) {
		interceptedTx := getDefaultInterceptedTx()
		interceptedTx.TransactionCalled = func() data.TransactionHandler { return nil }
		txData, err := dataValidators.GetTxData(interceptedTx)
		require.Equal(t, process.ErrNilTransaction, err)
		require.Nil(t, txData)
	})
	t.Run("non nil intercepted tx without data", func(t *testing.T) {
		expectedData := []byte(nil)
		interceptedTx := getDefaultInterceptedTx()
		interceptedTx.TransactionCalled = func() data.TransactionHandler {
			return &transaction.Transaction{
				Data: expectedData,
			}
		}
		txData, err := dataValidators.GetTxData(interceptedTx)
		require.Nil(t, err)
		require.Equal(t, expectedData, txData)
	})
	t.Run("non nil intercepted tx with data", func(t *testing.T) {
		expectedData := []byte("expected data")
		interceptedTx := getDefaultInterceptedTx()
		interceptedTx.TransactionCalled = func() data.TransactionHandler {
			return &transaction.Transaction{
				Data: expectedData,
			}
		}
		txData, err := dataValidators.GetTxData(interceptedTx)
		require.Nil(t, err)
		require.Equal(t, expectedData, txData)
	})
}

//------- IsInterfaceNil

func TestTxValidator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	adb := getAccAdapter(0, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	txValidator, _ := dataValidators.NewTxValidator(
		adb,
		shardCoordinator,
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.TxVersionCheckerStub{},
		100,
	)
	_ = txValidator
	txValidator = nil

	assert.True(t, check.IfNil(txValidator))
}

func getDefaultInterceptedTx() *mock.InterceptedTxHandlerStub {
	return &mock.InterceptedTxHandlerStub{
		SenderShardIdCalled: func() uint32 {
			return 0
		},
		ReceiverShardIdCalled: func() uint32 {
			return 1
		},
		NonceCalled: func() uint64 {
			return 0
		},
		SenderAddressCalled: func() []byte {
			return []byte("sender address")
		},
		FeeCalled: func() *big.Int {
			return big.NewInt(100000)
		},
		TransactionCalled: func() data.TransactionHandler {
			return &transaction.Transaction{}
		},
	}
}
