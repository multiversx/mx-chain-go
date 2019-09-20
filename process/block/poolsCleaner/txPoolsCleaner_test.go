package poolsCleaner_test

import (
	"bytes"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/poolsCleaner"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

func getAccAdapter(nonce uint64, balance *big.Int) *mock.AccountsStub {
	accDB := &mock.AccountsStub{}
	accDB.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &state.Account{Nonce: nonce, Balance: balance}, nil
	}

	return accDB
}

func initDataPoolWithFourTransactions() *mock.PoolsHolderStub {
	delayedFetchingKey := "key1"
	validTxKey := "key2"
	invalidTxKey := "key3"

	return &mock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{
				RegisterHandlerCalled: func(i func(key []byte)) {},
				ShardDataStoreCalled: func(id string) (c storage.Cacher) {
					return &mock.CacherStub{
						PeekCalled: func(key []byte) (value interface{}, ok bool) {
							switch string(key) {
							case delayedFetchingKey:
								time.Sleep(time.Second)
								return &transaction.Transaction{Nonce: 10}, true
							case validTxKey:
								return &transaction.Transaction{
									Nonce:   10,
									SndAddr: []byte("address_address_address_address_"),
								}, true
							case invalidTxKey:
								return &smartContractResult.SmartContractResult{}, true
							default:
								return nil, false
							}
						},
						KeysCalled: func() [][]byte {
							return [][]byte{[]byte(delayedFetchingKey), []byte(validTxKey), []byte(invalidTxKey), []byte("key4")}
						},
						LenCalled: func() int {
							return 0
						},
						RemoveCalled: func(key []byte) {
							return
						},
					}
				},
			}
		},
	}
}

func initDataPool(testHash []byte) *mock.PoolsHolderStub {
	sdp := &mock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{
				ShardDataStoreCalled: func(id string) (c storage.Cacher) {
					return &mock.CacherStub{
						PeekCalled: func(key []byte) (value interface{}, ok bool) {
							if bytes.Equal(key, testHash) {
								return &transaction.Transaction{Nonce: 10}, true
							}
							return nil, false
						},
						KeysCalled: func() [][]byte {
							return [][]byte{[]byte("key1"), []byte("key2")}
						},
						LenCalled: func() int {
							return 0
						},
					}
				},
			}
		},
	}
	return sdp
}

func TestNewTxsPoolsCleaner_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	tdp := initDataPool([]byte("test"))
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	txsPoolsCleaner, err := poolsCleaner.NewTxsPoolsCleaner(nil, shardCoordinator, tdp, addrConverter)

	assert.Nil(t, txsPoolsCleaner)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewTxsPoolsCleaner_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	balance := big.NewInt(1)
	accounts := getAccAdapter(nonce, balance)
	tdp := initDataPool([]byte("test"))
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	txsPoolsCleaner, err := poolsCleaner.NewTxsPoolsCleaner(accounts, nil, tdp, addrConverter)

	assert.Nil(t, txsPoolsCleaner)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewTxsPoolsCleaner_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	balance := big.NewInt(1)
	accounts := getAccAdapter(nonce, balance)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	txsPoolsCleaner, err := poolsCleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, nil, addrConverter)

	assert.Nil(t, txsPoolsCleaner)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewTxsPoolsCleaner_NilTransactionPoolShouldErr(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	balance := big.NewInt(1)
	accounts := getAccAdapter(nonce, balance)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	tdp := &mock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return nil
		},
	}
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	txsPoolsCleaner, err := poolsCleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, tdp, addrConverter)

	assert.Nil(t, txsPoolsCleaner)
	assert.Equal(t, process.ErrNilTransactionPool, err)
}

func TestNewTxsPoolsCleaner_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	balance := big.NewInt(1)
	accounts := getAccAdapter(nonce, balance)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	tdp := initDataPool([]byte("test"))
	txsPoolsCleaner, err := poolsCleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, tdp, nil)

	assert.Nil(t, txsPoolsCleaner)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewTxsPoolsCleaner_ShouldWork(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	balance := big.NewInt(1)
	accounts := getAccAdapter(nonce, balance)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	tdp := initDataPool([]byte("test"))
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	txsPoolsCleaner, err := poolsCleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, tdp, addrConverter)

	assert.NotNil(t, txsPoolsCleaner)
	assert.Nil(t, err)
}

func TestTxPoolsCleaner_CleanNilSenderAddrShouldRemoveTx(t *testing.T) {
	t.Parallel()

	maxCleanTime := time.Second
	nonce := uint64(1)
	balance := big.NewInt(1)
	accounts := getAccAdapter(nonce, balance)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	tdp := initDataPoolWithFourTransactions()
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	txsPoolsCleaner, _ := poolsCleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, tdp, addrConverter)

	itRan, err := txsPoolsCleaner.Clean(maxCleanTime)
	assert.Nil(t, err)
	assert.Equal(t, true, itRan)

	numRemovedTxs := txsPoolsCleaner.NumRemovedTxs()
	assert.Equal(t, uint64(1), numRemovedTxs)
}

func TestTxPoolsCleaner_CleanAccountNotExistsShouldRemoveTx(t *testing.T) {
	t.Parallel()

	numRemovedTxsExpected := uint64(3)
	cleanDuration := 2 * time.Second
	accounts := &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, state.ErrAccNotFound
		},
	}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	tdp := initDataPoolWithFourTransactions()
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	txsPoolsCleaner, _ := poolsCleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, tdp, addrConverter)

	itRan, err := txsPoolsCleaner.Clean(cleanDuration)
	assert.Nil(t, err)
	assert.Equal(t, true, itRan)

	numRemovedTxs := txsPoolsCleaner.NumRemovedTxs()
	assert.Equal(t, numRemovedTxsExpected, numRemovedTxs)
}

func TestTxPoolsCleaner_CleanLowerAccountNonceShouldRemoveTx(t *testing.T) {
	t.Parallel()

	numRemovedTxsExpected := uint64(3)
	cleanDuration := 2 * time.Second
	nonce := uint64(11)
	balance := big.NewInt(1)
	accounts := getAccAdapter(nonce, balance)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	tdp := initDataPoolWithFourTransactions()
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	txsPoolsCleaner, _ := poolsCleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, tdp, addrConverter)

	itRan, err := txsPoolsCleaner.Clean(cleanDuration)
	assert.Nil(t, err)
	assert.Equal(t, true, itRan)

	numRemovedTxs := txsPoolsCleaner.NumRemovedTxs()
	assert.Equal(t, numRemovedTxsExpected, numRemovedTxs)
}

func TestTxPoolsCleaner_CleanNilHaveTimeShouldErr(t *testing.T) {
	t.Parallel()

	nonce := uint64(11)
	balance := big.NewInt(1)
	accounts := getAccAdapter(nonce, balance)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	tdp := initDataPoolWithFourTransactions()
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	txsPoolsCleaner, _ := poolsCleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, tdp, addrConverter)

	itRan, err := txsPoolsCleaner.Clean(0)
	assert.Equal(t, process.ErrZeroMaxCleanTime, err)
	assert.Equal(t, false, itRan)
}

func TestTxPoolsCleaner_CleanWillDoNothingIfIsCalledMultipleTime(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	balance := big.NewInt(1)
	accounts := getAccAdapter(nonce, balance)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	tdp := initDataPoolWithFourTransactions()
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	txsPoolsCleaner, _ := poolsCleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, tdp, addrConverter)

	go func() {
		_, _ = txsPoolsCleaner.Clean(time.Second)
	}()
	time.Sleep(time.Millisecond)
	go func() {
		itRan, _ := txsPoolsCleaner.Clean(time.Second)
		assert.Equal(t, false, itRan)
	}()

	time.Sleep(2 * time.Second)
}
