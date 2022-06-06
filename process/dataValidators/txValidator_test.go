package dataValidators_test

import (
	"bytes"
	"errors"
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.AddressBlacklistCheckerStub{},
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
		&testscommon.AddressBlacklistCheckerStub{},
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
		&testscommon.AddressBlacklistCheckerStub{},
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
		&testscommon.AddressBlacklistCheckerStub{},
		maxNonceDeltaAllowed,
	)

	assert.Nil(t, txValidator)
	assert.True(t, errors.Is(err, process.ErrNilPubkeyConverter))
}

func TestTxValidator_CheckTxValidityNilAddressBlackListCheckerShouldErr(t *testing.T) {
	t.Parallel()

	adb := getAccAdapter(0, big.NewInt(0))
	maxNonceDeltaAllowed := 100
	shardCoordinator := createMockCoordinator("_", 0)
	txValidator, err := dataValidators.NewTxValidator(
		adb,
		shardCoordinator,
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		nil,
		maxNonceDeltaAllowed,
	)

	assert.Nil(t, txValidator)
	assert.True(t, errors.Is(err, process.ErrNilAddrBlacklistChecker))
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
		&testscommon.AddressBlacklistCheckerStub{},
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
		&testscommon.AddressBlacklistCheckerStub{},
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
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.AddressBlacklistCheckerStub{},
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
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.AddressBlacklistCheckerStub{},
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
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.AddressBlacklistCheckerStub{},
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
		&testscommon.AddressBlacklistCheckerStub{},
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
		&testscommon.AddressBlacklistCheckerStub{},
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
		&testscommon.AddressBlacklistCheckerStub{},
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
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.AddressBlacklistCheckerStub{},
		maxNonceDeltaAllowed,
	)

	addressMock := []byte("address")
	currentShard := uint32(0)
	txValidatorHandler := getTxValidatorHandler(currentShard, currentShard, 1, addressMock, big.NewInt(0))

	result := txValidator.CheckTxValidity(txValidatorHandler)
	assert.Nil(t, result)
}

func TestTxValidator_checkBlacklist(t *testing.T) {
	adb := getAccAdapter(0, big.NewInt(0))
	shardCoordinator := createMockCoordinator("_", 0)
	maxNonceDeltaAllowed := 100
	blacklisted := []byte("blacklisted")
	notBlacklisted := []byte("not Blacklisted")

	txValidator, err := dataValidators.NewTxValidator(
		adb,
		shardCoordinator,
		&testscommon.WhiteListHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		&testscommon.AddressBlacklistCheckerStub{
			IsBlacklistedCalled: func(addrBytes []byte) bool {
				if bytes.Equal(addrBytes, blacklisted) {
					return true
				}
				return false
			},
		},
		maxNonceDeltaAllowed,
	)
	require.Nil(t, err)


	t.Run("normal tx with blacklist, destination should allow", func(t *testing.T) {
		inTx := createInterceptedTransactionHandler(&transaction.Transaction{SndAddr: blacklisted})
		require.Nil(t, err)

		inTx.SenderShardIdCalled = func() uint32 {
			return 1
		}

		errCheck := txValidator.CheckBlacklist(inTx)
		require.Nil(t, errCheck)
	})
	t.Run("normal tx with blacklist, source should block", func(t *testing.T) {
		inTx := createInterceptedTransactionHandler(&transaction.Transaction{SndAddr: blacklisted})
		inTx.SenderShardIdCalled = func() uint32 {
			return 0
		}

		errCheck := txValidator.CheckBlacklist(inTx)
		require.Equal(t, process.ErrBlacklistedAddress, errCheck)
	})
	t.Run("normal tx with no blacklist, should allow", func(t *testing.T) {
		inTx := createInterceptedTransactionHandler(&transaction.Transaction{SndAddr: notBlacklisted})
		inTx.SenderShardIdCalled = func() uint32 {
			return 0
		}

		errCheck := txValidator.CheckBlacklist(inTx)
		require.Nil(t, errCheck)
	})
	t.Run("relayed with blacklist relayer, on destination should allow", func(t *testing.T) {
		inTx := createInterceptedTransactionHandler(&transaction.Transaction{SndAddr: blacklisted, Data: []byte("relayed")})
		inTx.SenderShardIdCalled = func() uint32 {
			return 1
		}

		errCheck := txValidator.CheckBlacklist(inTx)
		require.Nil(t, errCheck)
	})
	t.Run("relayed with blacklist relayer, on source should block", func(t *testing.T) {
		inTx := createInterceptedTransactionHandler(&transaction.Transaction{SndAddr: blacklisted, Data: []byte("relayed")})
		inTx.SenderShardIdCalled = func() uint32 {
			return 0
		}

		errCheck := txValidator.CheckBlacklist(inTx)
		require.Equal(t, process.ErrBlacklistedAddress, errCheck)
	})
	t.Run("relayed with blacklist user, on destination should allow", func(t *testing.T) {
		inTx := createInterceptedTransactionHandler(&transaction.Transaction{SndAddr: notBlacklisted, RcvAddr: blacklisted, Data: []byte("relayed")})
		inTx.SenderShardIdCalled = func() uint32 {
			return 1
		}

		errCheck := txValidator.CheckBlacklist(inTx)
		require.Nil(t, errCheck)
	})
	t.Run("relayed with blacklist user, on source should block", func(t *testing.T) {
		inTx := createInterceptedTransactionHandler(&transaction.Transaction{SndAddr: notBlacklisted, RcvAddr: blacklisted, Data: []byte("relayed")})
		inTx.SenderShardIdCalled = func() uint32 {
			return 0
		}

		errCheck := txValidator.CheckBlacklist(inTx)
		require.Equal(t,process.ErrBlacklistedAddress, errCheck)
	})
	t.Run("relayed with no blacklist, on source should allow", func(t *testing.T) {
		inTx := createInterceptedTransactionHandler(&transaction.Transaction{SndAddr: notBlacklisted, RcvAddr: notBlacklisted, Data: []byte("relayed")})
		inTx.SenderShardIdCalled = func() uint32 {
			return 0
		}

		errCheck := txValidator.CheckBlacklist(inTx)
		require.Nil(t, errCheck)
	})
	t.Run("relayed with no blacklist, on destination should allow", func(t *testing.T) {
		inTx := createInterceptedTransactionHandler(&transaction.Transaction{SndAddr: notBlacklisted, RcvAddr: notBlacklisted, Data: []byte("relayed")})
		inTx.SenderShardIdCalled = func() uint32 {
			return 1
		}

		errCheck := txValidator.CheckBlacklist(inTx)
		require.Nil(t, errCheck)
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
		&testscommon.AddressBlacklistCheckerStub{},
		100,
	)
	_ = txValidator
	txValidator = nil

	assert.True(t, check.IfNil(txValidator))
}

func createInterceptedTransactionHandler(transaction *transaction.Transaction) *mock.InterceptedTxHandlerStub {
	return &mock.InterceptedTxHandlerStub{
		SenderShardIdCalled: func() uint32 {
			return 0
		},
		ReceiverShardIdCalled: func() uint32 {
			return 0
		},
		NonceCalled: func() uint64 {
			return 0
		},
		SenderAddressCalled: func() []byte {
			return transaction.SndAddr
		},
		FeeCalled: func() *big.Int {
			return big.NewInt(0)
		},
		TransactionCalled: func() data.TransactionHandler {
			return &*transaction
		},
		GetUserTxSenderInRelayedTxCalled: func() ([]byte, error) {
			if bytes.Equal(transaction.Data, []byte("relayed")) {
				return transaction.RcvAddr, nil
			}
			return nil, errors.New("error")
		},
	}
}
