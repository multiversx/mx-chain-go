package interceptTransaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/interceptors"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/interceptors/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transactionPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createNewTransactionInterceptorForTests() (
	ti *transactionInterceptor,
	mes *mock.MessengerStub,
	pool *transactionPool.TransactionPool,
	addrConv *mock.AddressConverterMock,
	err error) {

	mes = mock.NewMessengerStub()
	pool = transactionPool.NewTransactionPool()
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

	ti, err = NewTransactionInterceptor(mes, pool, addrConv)

	return
}

//------- NewTransactionInterceptor

func TestNewTransactionInterceptorNilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	tPool := transactionPool.NewTransactionPool()
	addrConv := &mock.AddressConverterMock{}

	_, err := NewTransactionInterceptor(nil, tPool, addrConv)
	assert.Equal(t, interceptors.ErrNilMessenger, err)
}

func TestNewTransactionInterceptorNilTransactionPoolShouldErr(t *testing.T) {
	t.Parallel()

	mes := mock.NewMessengerStub()
	addrConv := &mock.AddressConverterMock{}

	_, err := NewTransactionInterceptor(mes, nil, addrConv)
	assert.Equal(t, interceptors.ErrNilTxPool, err)
}

func TestNewTransactionInterceptorNilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	mes := mock.NewMessengerStub()
	tPool := transactionPool.NewTransactionPool()

	_, err := NewTransactionInterceptor(mes, tPool, nil)
	assert.Equal(t, interceptors.ErrNilAddressConverter, err)
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

	assert.False(t, ti.ProcessTx(nil, make([]byte, 0)))
}

func TestTransactionInterceptorProcessTxWrongTypeOfNewerShouldRetFalse(t *testing.T) {
	t.Parallel()

	ti, _, _, _, _ := createNewTransactionInterceptorForTests()

	sn := mock.StringNewer{}

	assert.False(t, ti.ProcessTx(&sn, make([]byte, 0)))
}

func TestTransactionInterceptorProcessTxSanityCheckFailedShouldRetFalse(t *testing.T) {
	t.Parallel()

	ti, _, _, _, _ := createNewTransactionInterceptorForTests()

	txNewer := NewInterceptedTransaction()
	txNewer.Signature = nil
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.False(t, ti.ProcessTx(txNewer, make([]byte, 0)))
}

func TestTransactionInterceptorProcessTxNilHasherShouldRetFalse(t *testing.T) {
	t.Parallel()

	ti, mes, _, _, _ := createNewTransactionInterceptorForTests()

	mes.HasherObj = nil

	txNewer := NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = []byte("please fail, addrConverter!")
	txNewer.SndAddr = make([]byte, 0)

	assert.False(t, ti.ProcessTx(txNewer, make([]byte, 0)))
}

func TestTransactionInterceptorProcessTxNotValidShouldRetFalse(t *testing.T) {
	t.Parallel()

	ti, _, _, addrConv, _ := createNewTransactionInterceptorForTests()

	txNewer := NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = []byte("please fail, addrConverter!")
	txNewer.SndAddr = make([]byte, 0)

	addrConv.CreateAddressFromPublicKeyBytesRetErrForValue = []byte("please fail, addrConverter!")

	assert.False(t, ti.ProcessTx(txNewer, nil))
}

func TestTransactionInterceptorProcessValidValsShouldRetTrue(t *testing.T) {
	t.Parallel()

	ti, _, pool, _, _ := createNewTransactionInterceptorForTests()

	txNewer := NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.True(t, ti.ProcessTx(txNewer, []byte("txHash")))

	if pool.MiniPool(0) == nil {
		assert.Fail(t, "failed to add tx")
		return
	}

	txStored, ok := pool.MiniPool(0).TxStore.Get(mock.HasherMock{}.Compute("txHash"))
	assert.True(t, ok)
	assert.Equal(t, txNewer.Transaction, txStored)
}

func TestTransactionInterceptorProcessValidValsOtherShardsShouldRetTrue(t *testing.T) {
	t.Parallel()

	ti, _, pool, _, _ := createNewTransactionInterceptorForTests()

	tx := &transaction.Transaction{}

	tim := &mock.TransactionInterceptorMock{}
	tim.IsAddressedToOtherShardsVal = true
	tim.Tx = tx
	tim.IsChecked = true
	tim.IsVerified = true

	assert.True(t, ti.ProcessTx(tim, []byte("txHash")))

	if pool.MiniPool(0) != nil {
		assert.Fail(t, "failed and added tx")
		return
	}
}

func TestTransactionInterceptorProcessValidVals2ShardsShouldRetTrue(t *testing.T) {
	t.Parallel()

	ti, _, pool, _, _ := createNewTransactionInterceptorForTests()

	tx := &transaction.Transaction{}

	tim := &mock.TransactionInterceptorMock{}
	tim.IsAddressedToOtherShardsVal = false
	tim.RcvShardVal = 2
	tim.SndShardVal = 3
	tim.Tx = tx
	tim.IsChecked = true
	tim.IsVerified = true

	assert.True(t, ti.ProcessTx(tim, []byte("txHash")))

	if pool.MiniPool(2) == nil || pool.MiniPool(3) == nil {
		assert.Fail(t, "failed to add tx")
		return
	}

	txStored, ok := pool.MiniPool(2).TxStore.Get(mock.HasherMock{}.Compute("txHash"))
	assert.True(t, ok)
	assert.Equal(t, tx, txStored)

	txStored, ok = pool.MiniPool(3).TxStore.Get(mock.HasherMock{}.Compute("txHash"))
	assert.True(t, ok)
	assert.Equal(t, tx, txStored)
}
