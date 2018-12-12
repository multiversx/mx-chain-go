package transaction_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	transaction2 "github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
)

func createNewTransactionInterceptorForTests() (
	ti *transaction.TxInterceptor,
	mes *mock.MessengerStub,
	pool data.ShardedDataCacherNotifier,
	addrConv *mock.AddressConverterMock,
	err error) {

	mes = mock.NewMessengerStub()
	pool, err = shardedData.NewShardedData(storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	})
	if err != nil {
		return
	}

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

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	ti, err = transaction.NewTxInterceptor(mes, pool, addrConv, mock.HasherMock{})

	return
}

//------- NewTransactionInterceptor

func TestNewTransactionInterceptor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	tPool, err := shardedData.NewShardedData(storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	})
	assert.Nil(t, err)

	addrConv := &mock.AddressConverterMock{}

	_, err = transaction.NewTxInterceptor(nil, tPool, addrConv, mock.HasherMock{})
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewTransactionInterceptor_NilTransactionPoolShouldErr(t *testing.T) {
	t.Parallel()

	mes := mock.NewMessengerStub()
	addrConv := &mock.AddressConverterMock{}

	_, err := transaction.NewTxInterceptor(mes, nil, addrConv, mock.HasherMock{})
	assert.Equal(t, process.ErrNilTxDataPool, err)
}

func TestNewTransactionInterceptor_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	mes := mock.NewMessengerStub()
	tPool, err := shardedData.NewShardedData(storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	})
	assert.Nil(t, err)

	_, err = transaction.NewTxInterceptor(mes, tPool, nil, mock.HasherMock{})
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewTransactionInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	mes := mock.NewMessengerStub()
	tPool, err := shardedData.NewShardedData(storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	})
	assert.Nil(t, err)
	addrConv := &mock.AddressConverterMock{}

	_, err = transaction.NewTxInterceptor(mes, tPool, addrConv, nil)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewTransactionInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	_, _, _, _, err := createNewTransactionInterceptorForTests()
	assert.Nil(t, err)
}

//------- processTx

func TestTransactionInterceptor_ProcessTxNilTxShouldRetFalse(t *testing.T) {
	t.Parallel()

	ti, _, _, _, _ := createNewTransactionInterceptorForTests()

	assert.False(t, ti.ProcessTx(nil, make([]byte, 0), mock.HasherMock{}))
}

func TestTransactionInterceptor_ProcessTxWrongTypeOfNewerShouldRetFalse(t *testing.T) {
	t.Parallel()

	ti, _, _, _, _ := createNewTransactionInterceptorForTests()

	sn := mock.StringNewer{}

	assert.False(t, ti.ProcessTx(&sn, make([]byte, 0), mock.HasherMock{}))
}

func TestTransactionInterceptor_ProcessTxSanityCheckFailedShouldRetFalse(t *testing.T) {
	t.Parallel()

	ti, _, _, _, _ := createNewTransactionInterceptorForTests()

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = nil
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.False(t, ti.ProcessTx(txNewer, make([]byte, 0), mock.HasherMock{}))
}

func TestTransactionInterceptor_ProcessTxNotValidShouldRetFalse(t *testing.T) {
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

func TestTransactionInterceptor_ProcessValidValsShouldRetTrue(t *testing.T) {
	t.Parallel()

	ti, _, pool, _, _ := createNewTransactionInterceptorForTests()

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.True(t, ti.ProcessTx(txNewer, []byte("txHash"), mock.HasherMock{}))

	if pool.ShardDataStore(0) == nil {
		assert.Fail(t, "failed to add tx")
		return
	}

	txStored, ok := pool.ShardDataStore(0).Get(mock.HasherMock{}.Compute("txHash"))
	assert.True(t, ok)
	assert.Equal(t, txNewer.Transaction, txStored)
}

func TestTransactionInterceptor_ProcessValidValsOtherShardsShouldRetTrue(t *testing.T) {
	t.Parallel()

	ti, _, pool, _, _ := createNewTransactionInterceptorForTests()

	tx := &transaction2.Transaction{}

	tim := &mock.TransactionInterceptorMock{}
	tim.IsAddressedToOtherShardsVal = true
	tim.Tx = tx
	tim.IsChecked = true
	tim.IsVerified = true

	assert.True(t, ti.ProcessTx(tim, []byte("txHash"), mock.HasherMock{}))

	if pool.ShardDataStore(0) != nil {
		assert.Fail(t, "failed and added tx")
		return
	}
}

func TestTransactionInterceptor_ProcessValidVals2ShardsShouldRetTrue(t *testing.T) {
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

	if pool.ShardDataStore(2) == nil || pool.ShardDataStore(3) == nil {
		assert.Fail(t, "failed to add tx")
		return
	}

	txStored, ok := pool.ShardDataStore(2).Get(mock.HasherMock{}.Compute("txHash"))
	assert.True(t, ok)
	assert.Equal(t, tx, txStored)

	txStored, ok = pool.ShardDataStore(3).Get(mock.HasherMock{}.Compute("txHash"))
	assert.True(t, ok)
	assert.Equal(t, tx, txStored)
}
