package poolscleaner_test

import (
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/poolscleaner"
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

func initDataPoolTransactions() *mock.PoolsHolderStub {
	return &mock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{
				RegisterHandlerCalled: func(i func(key []byte)) {},
				ShardDataStoreCalled: func(id string) (c storage.Cacher) {
					return &mock.CacherStub{
						PeekCalled: func(key []byte) (value interface{}, ok bool) {
							switch string(key) {
							case "key1":
								time.Sleep(time.Second)
								return &transaction.Transaction{Nonce: 10}, true
							case "key2":
								return &transaction.Transaction{
									Nonce:   10,
									SndAddr: []byte("address_address_address_address_"),
								}, true
							case "key3":
								return &smartContractResult.SmartContractResult{}, true
							default:
								return nil, false
							}
						},
						KeysCalled: func() [][]byte {
							return [][]byte{[]byte("key1"), []byte("key2"), []byte("key3"), []byte("key4")}
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
							if reflect.DeepEqual(key, testHash) {
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
	txsPoolsCleaner, err := poolscleaner.NewTxsPoolsCleaner(nil, shardCoordinator, tdp, addrConverter)

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
	txsPoolsCleaner, err := poolscleaner.NewTxsPoolsCleaner(accounts, nil, tdp, addrConverter)

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
	txsPoolsCleaner, err := poolscleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, nil, addrConverter)

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
	txsPoolsCleaner, err := poolscleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, tdp, addrConverter)

	assert.Nil(t, txsPoolsCleaner)
	assert.Equal(t, process.ErrNilTransactionPool, err)
}

func TestNewTxsPoolsCleaner_ShouldWork(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	balance := big.NewInt(1)
	accounts := getAccAdapter(nonce, balance)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	tdp := initDataPool([]byte("test"))
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	txsPoolsCleaner, err := poolscleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, tdp, addrConverter)

	assert.NotNil(t, txsPoolsCleaner)
	assert.Nil(t, err)
}

func TestTxPoolsCleaner_CleanNilSenderAddrShouldRemoveTx(t *testing.T) {
	t.Parallel()

	cleanDurationSeconds := 1.0
	nonce := uint64(1)
	balance := big.NewInt(1)
	accounts := getAccAdapter(nonce, balance)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	tdp := initDataPoolTransactions()
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	txsPoolsCleaner, _ := poolscleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, tdp, addrConverter)

	startTime := time.Now()
	haveTime := func() bool {
		return time.Now().Sub(startTime).Seconds() < cleanDurationSeconds
	}

	err := txsPoolsCleaner.Clean(haveTime)
	assert.Nil(t, err)

	numRemovedTxs := txsPoolsCleaner.NumRemovedTxs()
	assert.Equal(t, uint64(1), numRemovedTxs)
}

func TestTxPoolsCleaner_CleanAccountNotExistsShouldRemoveTx(t *testing.T) {
	t.Parallel()

	cleanDurationSeconds := 2.0
	accounts := &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, state.ErrAccNotFound
		},
	}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	tdp := initDataPoolTransactions()
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	txsPoolsCleaner, _ := poolscleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, tdp, addrConverter)

	startTime := time.Now()
	haveTime := func() bool {
		return time.Now().Sub(startTime).Seconds() < cleanDurationSeconds
	}

	err := txsPoolsCleaner.Clean(haveTime)
	assert.Nil(t, err)

	numRemovedTxs := txsPoolsCleaner.NumRemovedTxs()
	assert.Equal(t, uint64(2), numRemovedTxs)
}

func TestTxPoolsCleaner_CleanLowerAccountNonceShouldRemoveTx(t *testing.T) {
	t.Parallel()

	cleanDurationSeconds := 2.0
	nonce := uint64(11)
	balance := big.NewInt(1)
	accounts := getAccAdapter(nonce, balance)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	tdp := initDataPoolTransactions()
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	txsPoolsCleaner, _ := poolscleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, tdp, addrConverter)

	startTime := time.Now()
	haveTime := func() bool {
		return time.Now().Sub(startTime).Seconds() < cleanDurationSeconds
	}

	err := txsPoolsCleaner.Clean(haveTime)
	assert.Nil(t, err)

	numRemovedTxs := txsPoolsCleaner.NumRemovedTxs()
	assert.Equal(t, uint64(2), numRemovedTxs)
}

func TestTxPoolsCleaner_CleanNilHaveTimeShouldErr(t *testing.T) {
	t.Parallel()

	nonce := uint64(11)
	balance := big.NewInt(1)
	accounts := getAccAdapter(nonce, balance)
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	tdp := initDataPoolTransactions()
	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	txsPoolsCleaner, _ := poolscleaner.NewTxsPoolsCleaner(accounts, shardCoordinator, tdp, addrConverter)

	err := txsPoolsCleaner.Clean(nil)
	assert.Equal(t, process.ErrNilHaveTimeHandler, err)
}
