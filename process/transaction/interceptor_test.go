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

	txi, err := transaction.NewTxInterceptor(
		nil,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Equal(t, process.ErrNilInterceptor, err)
	assert.Nil(t, txi)
}

func TestNewTxInterceptor_NilTransactionPoolShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		nil,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Equal(t, process.ErrNilTxDataPool, err)
	assert.Nil(t, txi)
}

func TestNewTxInterceptor_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	txPool := &mock.ShardedDataStub{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		nil,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Equal(t, process.ErrNilAddressConverter, err)
	assert.Nil(t, txi)
}

func TestNewTxInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		nil,
		keyGen,
		oneSharder)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, txi)
}

func TestNewTxInterceptor_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		nil,
		oneSharder)

	assert.Equal(t, process.ErrNilSingleSignKeyGen, err)
	assert.Nil(t, txi)
}

func TestNewTxInterceptor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		nil)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, txi)
}

func TestNewTxInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
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

func TestTransactionInterceptor_ProcessTxNilTxShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Equal(t, process.ErrNilTransaction, txi.ProcessTx(nil, make([]byte, 0)))
}

func TestTransactionInterceptor_ProcessTxWrongTypeOfNewerShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	sn := mock.StringCreator{}

	assert.Equal(t, process.ErrBadInterceptorTopicImplementation, txi.ProcessTx(&sn, make([]byte, 0)))
}

func TestTransactionInterceptor_ProcessTxIntegrityFailedShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = nil
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.Equal(t, process.ErrNilSignature, txi.ProcessTx(txNewer, make([]byte, 0)))
}

func TestTransactionInterceptor_ProcessNilDataToProcessShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.Equal(t, process.ErrNilDataToProcess, txi.ProcessTx(txNewer, nil))
}

func TestTransactionInterceptor_ProcessTxIntegrityAndValidityShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = []byte("please fail, addrConverter!")
	txNewer.SndAddr = make([]byte, 0)

	addrConv.CreateAddressFromPublicKeyBytesRetErrForValue = []byte("please fail, addrConverter!")

	assert.Equal(t, process.ErrInvalidRcvAddr, txi.ProcessTx(txNewer, make([]byte, 0)))
}

func TestTransactionInterceptor_ProcessTxVerifySigFailsShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}

	pubKey := &mock.SingleSignPublicKey{}
	pubKey.VerifyCalled = func(data []byte, signature []byte) error {
		return errors.New("sig not valid")
	}

	keyGen := &mock.SingleSignKeyGenMock{}
	keyGen.PublicKeyFromByteArrayCalled = func(b []byte) (key crypto.PublicKey, e error) {
		return pubKey, nil
	}

	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)
	txNewer.Value = *big.NewInt(0)

	assert.Equal(t, "sig not valid", txi.ProcessTx(txNewer, []byte("txHash")).Error())
}

func TestTransactionInterceptor_ProcessTxOkValsSameShardShouldWork(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
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
	pubKey.VerifyCalled = func(data []byte, signature []byte) error {
		return nil
	}

	keyGen := &mock.SingleSignKeyGenMock{}
	keyGen.PublicKeyFromByteArrayCalled = func(b []byte) (key crypto.PublicKey, e error) {
		return pubKey, nil
	}

	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.Nil(t, txi.ProcessTx(txNewer, []byte("txHash")))
	assert.Equal(t, 1, wasAdded)
}

func TestTransactionInterceptor_ProcessTxOkValsOtherShardsShouldWork(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
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
	pubKey.VerifyCalled = func(data []byte, signature []byte) error {
		return nil
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

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		multiSharder)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.Nil(t, txi.ProcessTx(txNewer, []byte("txHash")))
	assert.Equal(t, 0, wasAdded)
}

func TestTransactionInterceptor_ProcessTxOkVals2ShardsShouldWork(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
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
	pubKey.VerifyCalled = func(data []byte, signature []byte) error {
		return nil
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

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		addrConv,
		mock.HasherMock{},
		keyGen,
		multiSharder)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	assert.Nil(t, txi.ProcessTx(txNewer, []byte("txHash")))
	assert.Equal(t, 1, wasAdded)
}
