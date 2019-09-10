package dataValidators_test

import (
	"math/big"
	"strconv"
	"testing"

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
		GetSenderShardIdCalled: func() uint32 {
			return sndShardId
		},
		GetNonceCalled: func() uint64 {
			return nonce
		},
		GetSenderAddressCalled: func() state.AddressContainer {
			return sndAddr
		},
		GetTotalValueCalled: func() *big.Int {
			return totalValue
		},
	}
}

func TestTxValidator_NewValidator_ShouldErrNilAccounts(t *testing.T) {
	t.Parallel()

	txValidator, err := dataValidators.NewTxValidator(nil, nil)

	assert.Nil(t, txValidator)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestTxValidator_NewValidator_ShouldErrNilShardCoordinator(t *testing.T) {
	t.Parallel()

	accounts := getAccAdapter(0, big.NewInt(0))
	txValidator, err := dataValidators.NewTxValidator(accounts, nil)

	assert.Nil(t, txValidator)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestTxValidator_NewValidator_ShouldWork(t *testing.T) {
	t.Parallel()

	accounts := getAccAdapter(0, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	txValidator, err := dataValidators.NewTxValidator(accounts, shardCoordinator)

	assert.Nil(t, err)
	assert.NotNil(t, txValidator)

	result := txValidator.IsInterfaceNil()
	assert.Equal(t, false, result)
}

func TestTxValidator_IsTxValidForProcessing_ShouldReturnTrue_TxIsCrossShard(t *testing.T) {
	t.Parallel()

	accounts := getAccAdapter(1, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	txValidator, err := dataValidators.NewTxValidator(accounts, shardCoordinator)
	assert.Nil(t, err)

	addressMock := mock.NewAddressMock([]byte("address"))
	txValidatorHandler := getTxValidatorHandler(1, 1, addressMock, big.NewInt(0))

	result := txValidator.IsTxValidForProcessing(txValidatorHandler)
	assert.Equal(t, true, result)
}

func TestTxValidator_IsTxValidForProcessing_ShouldReturnFalse_AccountNonceIsGreaterThanTxNonce(t *testing.T) {
	t.Parallel()

	accountNonce := uint64(100)
	txNonce := uint64(0)

	accounts := getAccAdapter(accountNonce, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	txValidator, err := dataValidators.NewTxValidator(accounts, shardCoordinator)
	assert.Nil(t, err)

	addressMock := mock.NewAddressMock([]byte("address"))
	txValidatorHandler := getTxValidatorHandler(0, txNonce, addressMock, big.NewInt(0))

	result := txValidator.IsTxValidForProcessing(txValidatorHandler)
	assert.Equal(t, false, result)
}

func TestTxValidator_IsTxValidForProcessing_ShouldReturnFalse_AccountBalanceIsLessThanTxTotalValue(t *testing.T) {
	t.Parallel()

	accountNonce := uint64(0)
	txNonce := uint64(1)
	totalCost := big.NewInt(1000)
	accountBalance := big.NewInt(10)

	accounts := getAccAdapter(accountNonce, accountBalance)
	shardCoordinator := createMockCoordinator("_", 0)
	txValidator, err := dataValidators.NewTxValidator(accounts, shardCoordinator)
	assert.Nil(t, err)

	addressMock := mock.NewAddressMock([]byte("address"))
	txValidatorHandler := getTxValidatorHandler(0, txNonce, addressMock, totalCost)

	result := txValidator.IsTxValidForProcessing(txValidatorHandler)
	assert.Equal(t, false, result)
}

func TestTxValidator_IsTxValidForProcessing_ShouldReturnFalse_NumOfRejectedTxShouldIncreased(t *testing.T) {
	t.Parallel()

	accountNonce := uint64(0)
	txNonce := uint64(1)
	totalCost := big.NewInt(1000)
	accountBalance := big.NewInt(10)

	accounts := getAccAdapter(accountNonce, accountBalance)
	shardCoordinator := createMockCoordinator("_", 0)
	txValidator, err := dataValidators.NewTxValidator(accounts, shardCoordinator)
	assert.Nil(t, err)

	addressMock := mock.NewAddressMock([]byte("address"))
	txValidatorHandler := getTxValidatorHandler(0, txNonce, addressMock, totalCost)

	result := txValidator.IsTxValidForProcessing(txValidatorHandler)
	assert.Equal(t, false, result)

	numRejectedTx := txValidator.GetNumRejectedTxs()
	assert.Equal(t, uint64(1), numRejectedTx)
}
