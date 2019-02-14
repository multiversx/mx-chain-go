package transaction_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/pkg/errors"

	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/stretchr/testify/assert"
)

func createMarshalizerInterceptor() (marshal.Marshalizer, *mock.InterceptorStub) {
	marshalizer := &mock.MarshalizerMock{}

	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i p2p.TopicValidatorHandler) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return marshalizer
		},
	}

	return marshalizer, interceptor
}

//------- NewTxInterceptor

func TestNewTxInterceptor_NilInterceptorShouldErr(t *testing.T) {
	t.Parallel()

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, err := transaction.NewTxInterceptor(
		nil,
		txPool,
		storer,
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
	storer := &mock.StorerStub{}

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		nil,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Equal(t, process.ErrNilTxDataPool, err)
	assert.Nil(t, txi)
}

func TestNewTxInterceptor_NilStorerShouldErr(t *testing.T) {
	t.Parallel()

	txPool := &mock.ShardedDataStub{}
	interceptor := &mock.InterceptorStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		nil,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Equal(t, process.ErrNilTxStorage, err)
	assert.Nil(t, txi)
}

func TestNewTxInterceptor_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	txPool := &mock.ShardedDataStub{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
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
	storer := &mock.StorerStub{}

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
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
	storer := &mock.StorerStub{}

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
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
	storer := &mock.StorerStub{}

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		nil)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, txi)
}

func TestNewTxInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	_, interceptor := createMarshalizerInterceptor()

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, err := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Nil(t, err)
	assert.NotNil(t, txi)
}

//------- processTx

func TestTransactionInterceptor_ProcessTxNilMesssageShouldErr(t *testing.T) {
	t.Parallel()

	_, interceptor := createMarshalizerInterceptor()

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Equal(t, process.ErrNilMessage, txi.ProcessTx(nil))
}

func TestTransactionInterceptor_ProcessTxMilMessageDataShouldErr(t *testing.T) {
	t.Parallel()

	_, interceptor := createMarshalizerInterceptor()

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	msg := &mock.P2PMessageMock{}

	assert.Equal(t, process.ErrNilDataToProcess, txi.ProcessTx(msg))
}

func TestTransactionInterceptor_ProcessTxNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	_, interceptor := createMarshalizerInterceptor()
	interceptor.MarshalizerCalled = func() marshal.Marshalizer {
		return nil
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, process.ErrNilMarshalizer, txi.ProcessTx(msg))
}

func TestTransactionInterceptor_ProcessTxMarshalizerFailsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")

	ms := &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return errMarshalizer
		},
	}

	_, interceptor := createMarshalizerInterceptor()
	interceptor.MarshalizerCalled = func() marshal.Marshalizer {
		return ms
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, errMarshalizer, txi.ProcessTx(msg))
}

func TestTransactionInterceptor_ProcessTxMarshalizerFailsAtMarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")

	ms := &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return nil
		},
		MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
			return nil, errMarshalizer
		},
	}

	_, interceptor := createMarshalizerInterceptor()
	interceptor.MarshalizerCalled = func() marshal.Marshalizer {
		return ms
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, errMarshalizer, txi.ProcessTx(msg))
}

func TestTransactionInterceptor_ProcessTxIntegrityFailedShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer, interceptor := createMarshalizerInterceptor()

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = nil
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)

	buff, _ := marshalizer.Marshal(txNewer)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	assert.Equal(t, process.ErrNilSignature, txi.ProcessTx(msg))
}

func TestTransactionInterceptor_ProcessTxVerifySigFailsShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer, interceptor := createMarshalizerInterceptor()

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
	storer := &mock.StorerStub{}

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)
	txNewer.Value = big.NewInt(0)

	buff, _ := marshalizer.Marshal(txNewer)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	assert.Equal(t, "sig not valid", txi.ProcessTx(msg).Error())
}

func TestTransactionInterceptor_ProcessTxOkValsSameShardShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer, interceptor := createMarshalizerInterceptor()

	wasAdded := 0

	txPool := &mock.ShardedDataStub{}

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
	storer := &mock.StorerStub{}
	storer.HasCalled = func(key []byte) (bool, error) {
		return false, nil
	}

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)
	txNewer.Value = big.NewInt(0)

	buff, _ := marshalizer.Marshal(txNewer)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	txPool.AddDataCalled = func(key []byte, data interface{}, destShardID uint32) {
		if bytes.Equal(mock.HasherMock{}.Compute(string(buff)), key) {
			wasAdded++
		}
	}

	assert.Nil(t, txi.ProcessTx(msg))
	assert.Equal(t, 1, wasAdded)
}

func TestTransactionInterceptor_ProcessTxOkValsOtherShardsShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer, interceptor := createMarshalizerInterceptor()

	wasAdded := 0

	txPool := &mock.ShardedDataStub{}

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
	storer := &mock.StorerStub{}

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		multiSharder)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)
	txNewer.Value = big.NewInt(0)

	buff, _ := marshalizer.Marshal(txNewer)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	txPool.AddDataCalled = func(key []byte, data interface{}, destShardID uint32) {
		if bytes.Equal(mock.HasherMock{}.Compute(string(buff)), key) {
			wasAdded++
		}
	}

	assert.Nil(t, txi.ProcessTx(msg))
	assert.Equal(t, 0, wasAdded)
}

func TestTransactionInterceptor_ProcessTxPresentInStorerShouldNotAdd(t *testing.T) {
	t.Parallel()

	marshalizer, interceptor := createMarshalizerInterceptor()

	wasAdded := 0

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}

	pubKey := &mock.SingleSignPublicKey{}
	pubKey.VerifyCalled = func(data []byte, signature []byte) error {
		return nil
	}

	keyGen := &mock.SingleSignKeyGenMock{}
	keyGen.PublicKeyFromByteArrayCalled = func(b []byte) (key crypto.PublicKey, e error) {
		return pubKey, nil
	}
	storer := &mock.StorerStub{}
	storer.HasCalled = func(key []byte) (bool, error) {
		return true, nil
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
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		multiSharder)

	txNewer := transaction.NewInterceptedTransaction()
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)
	txNewer.Value = big.NewInt(0)

	buff, _ := marshalizer.Marshal(txNewer)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	txPool.AddDataCalled = func(key []byte, data interface{}, destShardID uint32) {
		if bytes.Equal(mock.HasherMock{}.Compute(string(buff)), key) {
			wasAdded++
		}
	}

	assert.Nil(t, txi.ProcessTx(msg))
	assert.Equal(t, 0, wasAdded)
}
