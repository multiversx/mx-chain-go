package transaction_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/pkg/errors"

	//transaction2 "github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/stretchr/testify/assert"
)

//------- NewTxInterceptor

func TestNewTxInterceptor_NilInterceptorShouldErr(t *testing.T) {
	t.Parallel()

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	_, err := transaction.NewTxInterceptor(
		nil,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Equal(t, process.ErrNilInterceptor, err)
}

func TestNewTxInterceptor_NilTransactionPoolShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	_, err := transaction.NewTxInterceptor(
		interceptor,
		nil,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Equal(t, process.ErrNilTxDataPool, err)
}

func TestNewTxInterceptor_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	txPool := &mock.ShardedDataStub{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	_, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		nil,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewTxInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	_, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		nil,
		keyGen,
		oneSharder)

	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewTxInterceptor_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	_, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		nil,
		oneSharder)

	assert.Equal(t, process.ErrNilSingleSignKeyGen, err)
}

func TestNewTxInterceptor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}

	_, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		nil)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewTxInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

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
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

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
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Nil(t, err)
	assert.NotNil(t, txi)

	sn := mock.StringNewer{}

	assert.False(t, txi.ProcessTx(&sn, make([]byte, 0)))
}

func TestTransactionInterceptor_ProcessTxIntegrityFailedShouldRetFalse(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Nil(t, err)
	assert.NotNil(t, txi)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = nil
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.False(t, txi.ProcessTx(txNewer, make([]byte, 0)))
}

func TestTransactionInterceptor_ProcessTxIntegrityAndValidityShouldRetFalse(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

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

func TestTransactionInterceptor_ProcessTxVerifySigFailsShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}

	pubKey := &mock.SingleSignPublicKey{}
	pubKey.VerifyCalled = func(data []byte, signature []byte) (b bool, e error) {
		return false, errors.New("sig not valid")
	}

	keyGen := &mock.SingleSignKeyGenMock{}
	keyGen.PublicKeyFromByteArrayCalled = func(b []byte) (key crypto.PublicKey, e error) {
		return pubKey, nil
	}

	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Nil(t, err)
	assert.NotNil(t, txi)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)
	txNewer.Value = *big.NewInt(0)

	assert.False(t, txi.ProcessTx(txNewer, []byte("txHash")))
}

func TestTransactionInterceptor_ProcessTxOkValsSameShardShouldWork(t *testing.T) {
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

	pubKey := &mock.SingleSignPublicKey{}
	pubKey.VerifyCalled = func(data []byte, signature []byte) (b bool, e error) {
		return true, nil
	}

	keyGen := &mock.SingleSignKeyGenMock{}
	keyGen.PublicKeyFromByteArrayCalled = func(b []byte) (key crypto.PublicKey, e error) {
		return pubKey, nil
	}

	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

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

func TestTransactionInterceptor_ProcessTxOkValsOtherShardsShouldWork(t *testing.T) {
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

	pubKey := &mock.SingleSignPublicKey{}
	pubKey.VerifyCalled = func(data []byte, signature []byte) (b bool, e error) {
		return true, nil
	}

	keyGen := &mock.SingleSignKeyGenMock{}
	keyGen.PublicKeyFromByteArrayCalled = func(b []byte) (key crypto.PublicKey, e error) {
		return pubKey, nil
	}

	multiSharder := mock.NewMultipleShardsCoordinatorMock()
	multiSharder.CurrentShard = 7
	multiSharder.ComputeShardForAddressCalled = func(address state.AddressContainer, addressConverter state.AddressConverter) uint32 {
		return 0
	}

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		multiSharder)

	assert.Nil(t, err)
	assert.NotNil(t, txi)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.True(t, txi.ProcessTx(txNewer, []byte("txHash")))
	assert.Equal(t, 0, wasAdded)
}

func TestTransactionInterceptor_ProcessTxOkVals2ShardsShouldWork(t *testing.T) {
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

	pubKey := &mock.SingleSignPublicKey{}
	pubKey.VerifyCalled = func(data []byte, signature []byte) (b bool, e error) {
		return true, nil
	}

	keyGen := &mock.SingleSignKeyGenMock{}
	keyGen.PublicKeyFromByteArrayCalled = func(b []byte) (key crypto.PublicKey, e error) {
		return pubKey, nil
	}

	multiSharder := mock.NewMultipleShardsCoordinatorMock()
	multiSharder.CurrentShard = 0
	called := uint32(0)
	multiSharder.ComputeShardForAddressCalled = func(address state.AddressContainer, addressConverter state.AddressConverter) uint32 {
		defer func() {
			called++
		}()

		return called
	}

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		multiSharder)

	assert.Nil(t, err)
	assert.NotNil(t, txi)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.True(t, txi.ProcessTx(txNewer, []byte("txHash")))
	assert.Equal(t, 2, wasAdded)
}

//
//func TestTransactionInterceptor_ProcessValidValsOtherShardsShouldRetTrue(t *testing.T) {
//	t.Parallel()
//
//	interceptor := &mock.InterceptorStub{}
//	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
//	}
//
//	wasAdded := 0
//
//	txPool := &mock.ShardedDataStub{}
//	txPool.AddDataCalled = func(key []byte, data interface{}, destShardID uint32) {
//		if bytes.Equal(mock.HasherMock{}.Compute("txHash"), key) {
//			wasAdded++
//		}
//	}
//	addrConv := &mock.AddressConverterMock{}
//
//	txi, err := transaction.NewTxInterceptor(interceptor, txPool, addrConv, mock.HasherMock{})
//	assert.Nil(t, err)
//	assert.NotNil(t, txi)
//
//	tim := &mock.TransactionInterceptorMock{}
//	tim.IsAddressedToOtherShardsVal = true
//	tim.Tx = &transaction2.Transaction{}
//	tim.IsChecked = true
//	tim.IsVerified = true
//
//	assert.True(t, txi.ProcessTx(tim, []byte("txHash")))
//	assert.Equal(t, 0, wasAdded)
//}
//
//func TestTransactionInterceptor_ProcessValidVals2ShardsShouldRetTrue(t *testing.T) {
//	t.Parallel()
//
//	interceptor := &mock.InterceptorStub{}
//	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
//	}
//
//	wasAdded := 0
//
//	txPool := &mock.ShardedDataStub{}
//	txPool.AddDataCalled = func(key []byte, data interface{}, destShardID uint32) {
//		if bytes.Equal(mock.HasherMock{}.Compute("txHash"), key) {
//			wasAdded++
//		}
//	}
//	addrConv := &mock.AddressConverterMock{}
//
//	txi, err := transaction.NewTxInterceptor(interceptor, txPool, addrConv, mock.HasherMock{})
//	assert.Nil(t, err)
//	assert.NotNil(t, txi)
//
//	tim := &mock.TransactionInterceptorMock{}
//	tim.IsAddressedToOtherShardsVal = false
//	tim.RcvShardVal = 2
//	tim.SndShardVal = 3
//	tim.Tx = &transaction2.Transaction{}
//	tim.IsChecked = true
//	tim.IsVerified = true
//
//	assert.True(t, txi.ProcessTx(tim, []byte("txHash")))
//	assert.Equal(t, 2, wasAdded)
//}
