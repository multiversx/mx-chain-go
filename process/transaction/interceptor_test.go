package transaction_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/pkg/errors"

	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/stretchr/testify/assert"
)

//------- NewTxInterceptor

func TestNewTxInterceptor_NilMarshalizerShouldErr(t *testing.T) {
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

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, txi)
}

func TestNewTxInterceptor_NilTransactionPoolShouldErr(t *testing.T) {
	t.Parallel()

	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, err := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
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
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	txi, err := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
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

	txPool := &mock.ShardedDataStub{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, err := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
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

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, err := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
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

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, err := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
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

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	storer := &mock.StorerStub{}

	txi, err := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
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

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, err := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Nil(t, err)
	assert.NotNil(t, txi)
}

//------- Validate

func TestTransactionInterceptor_ValidateNilMesssageShouldErr(t *testing.T) {
	t.Parallel()

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, _ := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	assert.Equal(t, process.ErrNilMessage, txi.Validate(nil))
}

func TestTransactionInterceptor_ValidateMilMessageDataShouldErr(t *testing.T) {
	t.Parallel()

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, _ := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	msg := &mock.P2PMessageMock{}

	assert.Equal(t, process.ErrNilDataToProcess, txi.Validate(msg))
}

func TestTransactionInterceptor_ValidateMarshalizerFailsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, _ := transaction.NewTxInterceptor(
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errMarshalizer
			},
		},
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, errMarshalizer, txi.Validate(msg))
}

func TestTransactionInterceptor_ValidateMarshalizerFailsAtMarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, _ := transaction.NewTxInterceptor(
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return nil
			},
			MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
				return nil, errMarshalizer
			},
		},
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		keyGen,
		oneSharder)

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, errMarshalizer, txi.Validate(msg))
}

func TestTransactionInterceptor_ValidateIntegrityFailedShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	txi, _ := transaction.NewTxInterceptor(
		marshalizer,
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

	assert.Equal(t, process.ErrNilSignature, txi.Validate(msg))
}

func TestTransactionInterceptor_ValidateVerifySigFailsShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

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
		marshalizer,
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

	assert.Equal(t, "sig not valid", txi.Validate(msg).Error())
}

func TestTransactionInterceptor_ValidateOkValsSameShardShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

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
		marshalizer,
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

	assert.Nil(t, txi.Validate(msg))
	assert.Equal(t, 1, wasAdded)
}

func TestTransactionInterceptor_ValidateOkValsOtherShardsShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

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
		marshalizer,
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

	assert.Nil(t, txi.Validate(msg))
	assert.Equal(t, 0, wasAdded)
}

func TestTransactionInterceptor_ValidatePresentInStorerShouldNotAdd(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

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
		marshalizer,
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

	assert.Nil(t, txi.Validate(msg))
	assert.Equal(t, 0, wasAdded)
}
