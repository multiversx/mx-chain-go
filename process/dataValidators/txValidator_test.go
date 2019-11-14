package dataValidators_test

import (
	"errors"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func getAccAdapter(nonce uint64, balance *big.Int) *mock.AccountsStub {
	accDB := &mock.AccountsStub{}
	accDB.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &state.Account{Nonce: nonce, Balance: balance}, nil
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
	nonce uint64,
	sndAddr state.AddressContainer,
	totalValue *big.Int,
) process.TxValidatorHandler {
	return &mock.TxValidatorHandlerStub{
		SenderShardIdCalled: func() uint32 {
			return sndShardId
		},
		NonceCalled: func() uint64 {
			return nonce
		},
		SenderAddressCalled: func() state.AddressContainer {
			return sndAddr
		},
		TotalValueCalled: func() *big.Int {
			return totalValue
		},
	}
}

func TestTxValidator_NewValidatorNilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(nil, shardCoordinator, maxNonceDeltaAllowed)

	assert.Nil(t, txValidator)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestTxValidator_NewValidatorNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	accounts := getAccAdapter(0, big.NewInt(0))
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(accounts, nil, maxNonceDeltaAllowed)

	assert.Nil(t, txValidator)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestTxValidator_NewValidatorShouldWork(t *testing.T) {
	t.Parallel()

	accounts := getAccAdapter(0, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(accounts, shardCoordinator, maxNonceDeltaAllowed)

	assert.Nil(t, err)
	assert.NotNil(t, txValidator)

	result := txValidator.IsInterfaceNil()
	assert.Equal(t, false, result)
}

func TestTxValidator_CheckTxValidityTxCrossShardShouldWork(t *testing.T) {
	t.Parallel()

	accounts := getAccAdapter(1, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(accounts, shardCoordinator, maxNonceDeltaAllowed)
	assert.Nil(t, err)

	addressMock := mock.NewAddressMock([]byte("address"))
	txValidatorHandler := getTxValidatorHandler(1, 1, addressMock, big.NewInt(0))

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.Nil(t, result)
}

func TestTxValidator_CheckTxValidityAccountNonceIsGreaterThanTxNonceShouldReturnFalse(t *testing.T) {
	t.Parallel()

	accountNonce := uint64(100)
	txNonce := uint64(0)

	accounts := getAccAdapter(accountNonce, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(accounts, shardCoordinator, maxNonceDeltaAllowed)
	assert.Nil(t, err)

	addressMock := mock.NewAddressMock([]byte("address"))
	txValidatorHandler := getTxValidatorHandler(0, txNonce, addressMock, big.NewInt(0))

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.NotNil(t, result)
	assert.True(t, strings.Contains(result.Error(), "nonce"))
}

func TestTxValidator_CheckTxValidityTxNonceIsTooHigh(t *testing.T) {
	t.Parallel()

	accountNonce := uint64(100)
	maxNonceDeltaAllowed := 100
	txNonce := accountNonce + uint64(maxNonceDeltaAllowed) + 1

	accounts := getAccAdapter(accountNonce, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	txValidator, err := dataValidators.NewTxValidator(accounts, shardCoordinator, maxNonceDeltaAllowed)
	assert.Nil(t, err)

	addressMock := mock.NewAddressMock([]byte("address"))
	txValidatorHandler := getTxValidatorHandler(0, txNonce, addressMock, big.NewInt(0))

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.NotNil(t, result)
	assert.True(t, strings.Contains(result.Error(), "nonce"))
}

func TestTxValidator_CheckTxValidityAccountBalanceIsLessThanTxTotalValueShouldReturnFalse(t *testing.T) {
	t.Parallel()

	accountNonce := uint64(0)
	txNonce := uint64(1)
	totalCost := big.NewInt(1000)
	accountBalance := big.NewInt(10)

	accounts := getAccAdapter(accountNonce, accountBalance)
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(accounts, shardCoordinator, maxNonceDeltaAllowed)
	assert.Nil(t, err)

	addressMock := mock.NewAddressMock([]byte("address"))
	txValidatorHandler := getTxValidatorHandler(0, txNonce, addressMock, totalCost)

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.NotNil(t, result)
	assert.True(t, strings.Contains(result.Error(), "balance"))
}

func TestTxValidator_CheckTxValidityNumOfRejectedTxShouldIncreaseShouldReturnFalse(t *testing.T) {
	t.Parallel()

	accountNonce := uint64(0)
	txNonce := uint64(1)
	totalCost := big.NewInt(1000)
	accountBalance := big.NewInt(10)

	accounts := getAccAdapter(accountNonce, accountBalance)
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, err := dataValidators.NewTxValidator(accounts, shardCoordinator, maxNonceDeltaAllowed)
	assert.Nil(t, err)

	addressMock := mock.NewAddressMock([]byte("address"))
	txValidatorHandler := getTxValidatorHandler(0, txNonce, addressMock, totalCost)

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.NotNil(t, result)

	numRejectedTx := txValidator.NumRejectedTxs()
	assert.Equal(t, uint64(1), numRejectedTx)
}

func TestTxValidator_CheckTxValidityAccountNotExitsShouldReturnFalse(t *testing.T) {
	t.Parallel()

	accDB := &mock.AccountsStub{}
	accDB.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return nil, errors.New("cannot find account")
	}
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, _ := dataValidators.NewTxValidator(accDB, shardCoordinator, maxNonceDeltaAllowed)

	addressMock := mock.NewAddressMock([]byte("address"))
	txValidatorHandler := getTxValidatorHandler(0, 1, addressMock, big.NewInt(0))

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.NotNil(t, result)
}

func TestTxValidator_CheckTxValidityWrongAccountTypeShouldReturnFalse(t *testing.T) {
	t.Parallel()

	accDB := &mock.AccountsStub{}
	accDB.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &state.MetaAccount{}, nil
	}
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, _ := dataValidators.NewTxValidator(accDB, shardCoordinator, maxNonceDeltaAllowed)

	addressMock := mock.NewAddressMock([]byte("address"))
	txValidatorHandler := getTxValidatorHandler(0, 1, addressMock, big.NewInt(0))

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.NotNil(t, result)
}

func TestTxValidator_CheckTxValidityTxIsOkShouldReturnTrue(t *testing.T) {
	t.Parallel()

	accountNonce := uint64(0)
	accountBalance := big.NewInt(10)
	accounts := getAccAdapter(accountNonce, accountBalance)
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	txValidator, _ := dataValidators.NewTxValidator(accounts, shardCoordinator, maxNonceDeltaAllowed)

	addressMock := mock.NewAddressMock([]byte("address"))
	txValidatorHandler := getTxValidatorHandler(0, 1, addressMock, big.NewInt(0))

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.Nil(t, result)
}

//------- IsInterfaceNil

func TestTxValidator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	accounts := getAccAdapter(0, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	txValidator, _ := dataValidators.NewTxValidator(accounts, shardCoordinator, 100)
	txValidator = nil

	assert.True(t, check.IfNil(txValidator))
}
