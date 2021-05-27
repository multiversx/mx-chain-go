package dataValidators_test

import (
	"errors"
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func getAccAdapter(nonce uint64, balance *big.Int) *testscommon.AccountsStub {
	accDB := &testscommon.AccountsStub{}
	accDB.GetExistingAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
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

func getTxValidatorHandler(
	sndShardId uint32,
	rcvShardId uint32,
	nonce uint64,
	sndAddr []byte,
	fee *big.Int,
) process.TxValidatorHandler {
	return &mock.TxValidatorHandlerStub{
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
	}
}

func TestNewTxValidator_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(
		nil,
		shardCoordinator,
		&mock.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
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
		&mock.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
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
		&mock.WhiteListHandlerStub{},
		nil,
		maxNonceDeltaAllowed,
	)

	assert.Nil(t, txValidator)
	assert.True(t, errors.Is(err, process.ErrNilPubkeyConverter))
}

func TestNewTxValidator_ShouldWork(t *testing.T) {
	t.Parallel()

	adb := getAccAdapter(0, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(
		adb,
		shardCoordinator,
		&mock.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
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
		&mock.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		maxNonceDeltaAllowed,
	)
	assert.Nil(t, err)

	addressMock := []byte("address")
	txValidatorHandler := getTxValidatorHandler(currentShard+1, currentShard, 1, addressMock, big.NewInt(0))

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
		&mock.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		maxNonceDeltaAllowed,
	)
	assert.Nil(t, err)

	addressMock := []byte("address")
	currentShard := uint32(0)
	txValidatorHandler := getTxValidatorHandler(currentShard, currentShard, txNonce, addressMock, big.NewInt(0))

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
		&mock.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		maxNonceDeltaAllowed,
	)
	assert.Nil(t, err)

	addressMock := []byte("address")
	currentShard := uint32(0)
	txValidatorHandler := getTxValidatorHandler(currentShard, currentShard, txNonce, addressMock, big.NewInt(0))

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
		&mock.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		maxNonceDeltaAllowed,
	)
	assert.Nil(t, err)

	addressMock := []byte("address")
	currentShard := uint32(0)
	txValidatorHandler := getTxValidatorHandler(currentShard, currentShard, txNonce, addressMock, fee)

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.NotNil(t, result)
	assert.True(t, errors.Is(result, process.ErrInsufficientFunds))
}

func TestTxValidator_CheckTxValidityAccountNotExitsShouldReturnFalse(t *testing.T) {
	t.Parallel()

	accDB := &testscommon.AccountsStub{}
	accDB.GetExistingAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return nil, errors.New("cannot find account")
	}
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, _ := dataValidators.NewTxValidator(
		accDB,
		shardCoordinator,
		&mock.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		maxNonceDeltaAllowed,
	)

	addressMock := []byte("address")
	currentShard := uint32(0)
	txValidatorHandler := getTxValidatorHandler(currentShard, currentShard, 1, addressMock, big.NewInt(0))

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.True(t, errors.Is(result, process.ErrAccountNotFound))
}

func TestTxValidator_CheckTxValidityAccountNotExitsButWhiteListedShouldReturnTrue(t *testing.T) {
	t.Parallel()

	accDB := &testscommon.AccountsStub{}
	accDB.GetExistingAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return nil, errors.New("cannot find account")
	}
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, _ := dataValidators.NewTxValidator(
		accDB,
		shardCoordinator,
		&mock.WhiteListHandlerStub{
			IsWhiteListedCalled: func(interceptedData process.InterceptedData) bool {
				return true
			},
		},
		mock.NewPubkeyConverterMock(32),
		maxNonceDeltaAllowed,
	)

	addressMock := []byte("address")
	currentShard := uint32(0)
	txValidatorHandler := getTxValidatorHandler(currentShard, currentShard, 1, addressMock, big.NewInt(0))

	interceptedTx := struct {
		process.InterceptedData
		process.TxValidatorHandler
	}{
		InterceptedData:    nil,
		TxValidatorHandler: txValidatorHandler,
	}

	// interceptedTx needs to be of type InterceptedData & TxValidatorHandler
	result := txValidator.CheckTxValidity(interceptedTx)
	assert.Nil(t, result)
}

func TestTxValidator_CheckTxValidityWrongAccountTypeShouldReturnFalse(t *testing.T) {
	t.Parallel()

	accDB := &testscommon.AccountsStub{}
	accDB.GetExistingAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return state.NewPeerAccount(address)
	}
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, _ := dataValidators.NewTxValidator(
		accDB,
		shardCoordinator,
		&mock.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		maxNonceDeltaAllowed,
	)

	addressMock := []byte("address")
	currentShard := uint32(0)
	txValidatorHandler := getTxValidatorHandler(currentShard, currentShard, 1, addressMock, big.NewInt(0))

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
		&mock.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		maxNonceDeltaAllowed,
	)

	addressMock := []byte("address")
	currentShard := uint32(0)
	txValidatorHandler := getTxValidatorHandler(currentShard, currentShard, 1, addressMock, big.NewInt(0))

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.Nil(t, result)
}

//------- IsInterfaceNil

func TestTxValidator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	adb := getAccAdapter(0, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	txValidator, _ := dataValidators.NewTxValidator(
		adb,
		shardCoordinator,
		&mock.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		100,
	)
	_ = txValidator
	txValidator = nil

	assert.True(t, check.IfNil(txValidator))
}
