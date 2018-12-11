package transaction_test

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	transaction2 "github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
)

func createNewTransactionInterceptorForTests() (
	ti *transaction.TransactionInterceptor,
	mes *mock.MessengerStub,
	pool *dataPool.DataPool,
	addrConv *mock.AddressConverterMock,
	err error) {

	mes = mock.NewMessengerStub()
	pool = dataPool.NewDataPool(&storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	})
	addrConv = &mock.AddressConverterMock{}

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		t.RegisterTopicValidator = func(v pubsub.Validator) error {
			return nil
		}
		t.UnregisterTopicValidator = func() error {
			return nil
		}

		return nil
	}

	ti, err = transaction.NewTransactionInterceptor(mes, pool, addrConv, mock.HasherMock{})

	return
}

//------- NewTransactionInterceptor

func TestNewTransactionInterceptorNilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	tPool := dataPool.NewDataPool(&storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	})
	addrConv := &mock.AddressConverterMock{}

	_, err := transaction.NewTransactionInterceptor(nil, tPool, addrConv, mock.HasherMock{})
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewTransactionInterceptorNilTransactionPoolShouldErr(t *testing.T) {
	t.Parallel()

	mes := mock.NewMessengerStub()
	addrConv := &mock.AddressConverterMock{}

	_, err := transaction.NewTransactionInterceptor(mes, nil, addrConv, mock.HasherMock{})
	assert.Equal(t, process.ErrNilTxDataPool, err)
}

func TestNewTransactionInterceptorNilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	mes := mock.NewMessengerStub()
	tPool := dataPool.NewDataPool(&storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	})

	_, err := transaction.NewTransactionInterceptor(mes, tPool, nil, mock.HasherMock{})
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewTransactionInterceptorNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	mes := mock.NewMessengerStub()
	tPool := dataPool.NewDataPool(&storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	})
	addrConv := &mock.AddressConverterMock{}

	_, err := transaction.NewTransactionInterceptor(mes, tPool, addrConv, nil)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewTransactionInterceptorOkValsShouldWork(t *testing.T) {
	t.Parallel()

	_, _, _, _, err := createNewTransactionInterceptorForTests()
	assert.Nil(t, err)
}

//------- processTx

func TestTransactionInterceptorProcessTxNilTxShouldRetFalse(t *testing.T) {
	t.Parallel()

	ti, _, _, _, _ := createNewTransactionInterceptorForTests()

	assert.False(t, ti.ProcessTx(nil, make([]byte, 0), mock.HasherMock{}))
}

func TestTransactionInterceptorProcessTxWrongTypeOfNewerShouldRetFalse(t *testing.T) {
	t.Parallel()

	ti, _, _, _, _ := createNewTransactionInterceptorForTests()

	sn := mock.StringNewer{}

	assert.False(t, ti.ProcessTx(&sn, make([]byte, 0), mock.HasherMock{}))
}

func TestTransactionInterceptorProcessTxSanityCheckFailedShouldRetFalse(t *testing.T) {
	t.Parallel()

	ti, _, _, _, _ := createNewTransactionInterceptorForTests()

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = nil
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.False(t, ti.ProcessTx(txNewer, make([]byte, 0), mock.HasherMock{}))
}

func TestTransactionInterceptorProcessTxNotValidShouldRetFalse(t *testing.T) {
	t.Parallel()

	ti, _, _, addrConv, _ := createNewTransactionInterceptorForTests()

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = []byte("please fail, addrConverter!")
	txNewer.SndAddr = make([]byte, 0)

	addrConv.CreateAddressFromPublicKeyBytesRetErrForValue = []byte("please fail, addrConverter!")

	assert.False(t, ti.ProcessTx(txNewer, nil, mock.HasherMock{}))
}

func TestTransactionInterceptorProcessValidValsShouldRetTrue(t *testing.T) {
	t.Parallel()

	ti, _, pool, _, _ := createNewTransactionInterceptorForTests()

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.True(t, ti.ProcessTx(txNewer, []byte("txHash"), mock.HasherMock{}))

	if pool.MiniPool(0) == nil {
		assert.Fail(t, "failed to add tx")
		return
	}

	txStored, ok := pool.MiniPool(0).DataStore.Get(mock.HasherMock{}.Compute("txHash"))
	assert.True(t, ok)
	assert.Equal(t, txNewer.Transaction, txStored)
}

func TestTransactionInterceptorProcessValidValsOtherShardsShouldRetTrue(t *testing.T) {
	t.Parallel()

	ti, _, pool, _, _ := createNewTransactionInterceptorForTests()

	tx := &transaction2.Transaction{}

	tim := &mock.TransactionInterceptorMock{}
	tim.IsAddressedToOtherShardsVal = true
	tim.Tx = tx
	tim.IsChecked = true
	tim.IsVerified = true

	assert.True(t, ti.ProcessTx(tim, []byte("txHash"), mock.HasherMock{}))

	if pool.MiniPool(0) != nil {
		assert.Fail(t, "failed and added tx")
		return
	}
}

func TestTransactionInterceptorProcessValidVals2ShardsShouldRetTrue(t *testing.T) {
	t.Parallel()

	ti, _, pool, _, _ := createNewTransactionInterceptorForTests()

	tx := &transaction2.Transaction{}

	tim := &mock.TransactionInterceptorMock{}
	tim.IsAddressedToOtherShardsVal = false
	tim.RcvShardVal = 2
	tim.SndShardVal = 3
	tim.Tx = tx
	tim.IsChecked = true
	tim.IsVerified = true

	assert.True(t, ti.ProcessTx(tim, []byte("txHash"), mock.HasherMock{}))

	if pool.MiniPool(2) == nil || pool.MiniPool(3) == nil {
		assert.Fail(t, "failed to add tx")
		return
	}

	txStored, ok := pool.MiniPool(2).DataStore.Get(mock.HasherMock{}.Compute("txHash"))
	assert.True(t, ok)
	assert.Equal(t, tx, txStored)

	txStored, ok = pool.MiniPool(3).DataStore.Get(mock.HasherMock{}.Compute("txHash"))
	assert.True(t, ok)
	assert.Equal(t, tx, txStored)
}
