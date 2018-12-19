package transaction_test

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"

	transaction2 "github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/stretchr/testify/assert"
)

//------- NewTxInterceptor

func TestNewTxInterceptor_NilInterceptorShouldErr(t *testing.T) {
	t.Parallel()

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}

	_, err := transaction.NewTxInterceptor(nil, txPool, addrConv, mock.HasherMock{})
	assert.Equal(t, process.ErrNilInterceptor, err)
}

func TestNewTxInterceptor_NilTransactionPoolShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	addrConv := &mock.AddressConverterMock{}

	_, err := transaction.NewTxInterceptor(interceptor, nil, addrConv, mock.HasherMock{})
	assert.Equal(t, process.ErrNilTxDataPool, err)
}

func TestNewTxInterceptor_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	txPool := &mock.ShardedDataStub{}

	_, err := transaction.NewTxInterceptor(interceptor, txPool, nil, mock.HasherMock{})
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewTxInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}

	_, err := transaction.NewTxInterceptor(interceptor, txPool, addrConv, nil)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewTxInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}

	txi, err := transaction.NewTxInterceptor(interceptor, txPool, addrConv, mock.HasherMock{})
	assert.Nil(t, err)
	assert.NotNil(t, txi)
}

//------- processTx

func TestTransactionInterceptor_ProcessTxNilTxShouldRetFalse(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}

	txi, err := transaction.NewTxInterceptor(interceptor, txPool, addrConv, mock.HasherMock{})
	assert.Nil(t, err)
	assert.NotNil(t, txi)

	assert.False(t, txi.ProcessTx(nil, make([]byte, 0)))
}

func TestTransactionInterceptor_ProcessTxWrongTypeOfNewerShouldRetFalse(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}

	txi, err := transaction.NewTxInterceptor(interceptor, txPool, addrConv, mock.HasherMock{})
	assert.Nil(t, err)
	assert.NotNil(t, txi)

	sn := mock.StringNewer{}

	assert.False(t, txi.ProcessTx(&sn, make([]byte, 0)))
}

func TestTransactionInterceptor_ProcessTxSanityCheckFailedShouldRetFalse(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}

	txi, err := transaction.NewTxInterceptor(interceptor, txPool, addrConv, mock.HasherMock{})
	assert.Nil(t, err)
	assert.NotNil(t, txi)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = nil
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.False(t, txi.ProcessTx(txNewer, make([]byte, 0)))
}

func TestTransactionInterceptor_ProcessTxNotValidShouldRetFalse(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}

	txi, err := transaction.NewTxInterceptor(interceptor, txPool, addrConv, mock.HasherMock{})
	assert.Nil(t, err)
	assert.NotNil(t, txi)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = []byte("please fail, addrConverter!")
	txNewer.SndAddr = make([]byte, 0)

	addrConv.CreateAddressFromPublicKeyBytesRetErrForValue = []byte("please fail, addrConverter!")

	assert.False(t, txi.ProcessTx(txNewer, nil))
}

func TestTransactionInterceptor_ProcessValidValsShouldRetTrue(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	wasAdded := 0

	txPool := &mock.ShardedDataStub{}
	txPool.AddDataCalled = func(key []byte, data interface{}, destShardID uint32) {
		if bytes.Equal(mock.HasherMock{}.Compute("txHash"), key) {
			wasAdded++
		}
	}
	addrConv := &mock.AddressConverterMock{}

	txi, err := transaction.NewTxInterceptor(interceptor, txPool, addrConv, mock.HasherMock{})
	assert.Nil(t, err)
	assert.NotNil(t, txi)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.True(t, txi.ProcessTx(txNewer, []byte("txHash")))
	assert.Equal(t, 1, wasAdded)
}

func TestTransactionInterceptor_ProcessValidValsOtherShardsShouldRetTrue(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	wasAdded := 0

	txPool := &mock.ShardedDataStub{}
	txPool.AddDataCalled = func(key []byte, data interface{}, destShardID uint32) {
		if bytes.Equal(mock.HasherMock{}.Compute("txHash"), key) {
			wasAdded++
		}
	}
	addrConv := &mock.AddressConverterMock{}

	txi, err := transaction.NewTxInterceptor(interceptor, txPool, addrConv, mock.HasherMock{})
	assert.Nil(t, err)
	assert.NotNil(t, txi)

	tim := &mock.TransactionInterceptorMock{}
	tim.IsAddressedToOtherShardsVal = true
	tim.Tx = &transaction2.Transaction{}
	tim.IsChecked = true
	tim.IsVerified = true

	assert.True(t, txi.ProcessTx(tim, []byte("txHash")))
	assert.Equal(t, 0, wasAdded)
}

func TestTransactionInterceptor_ProcessValidVals2ShardsShouldRetTrue(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	wasAdded := 0

	txPool := &mock.ShardedDataStub{}
	txPool.AddDataCalled = func(key []byte, data interface{}, destShardID uint32) {
		if bytes.Equal(mock.HasherMock{}.Compute("txHash"), key) {
			wasAdded++
		}
	}
	addrConv := &mock.AddressConverterMock{}

	txi, err := transaction.NewTxInterceptor(interceptor, txPool, addrConv, mock.HasherMock{})
	assert.Nil(t, err)
	assert.NotNil(t, txi)

	tim := &mock.TransactionInterceptorMock{}
	tim.IsAddressedToOtherShardsVal = false
	tim.RcvShardVal = 2
	tim.SndShardVal = 3
	tim.Tx = &transaction2.Transaction{}
	tim.IsChecked = true
	tim.IsVerified = true

	assert.True(t, txi.ProcessTx(tim, []byte("txHash")))
	assert.Equal(t, 2, wasAdded)
}
