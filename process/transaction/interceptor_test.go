package transaction_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/pkg/errors"
)

//------- NewTxInterceptor

func TestNewTxInterceptor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}
	signer := &mock.SignerMock{}

	txi, err := transaction.NewTxInterceptor(
		nil,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		signer,
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
	signer := &mock.SignerMock{}

	txi, err := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
		nil,
		storer,
		addrConv,
		mock.HasherMock{},
		signer,
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
	signer := &mock.SignerMock{}

	txi, err := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
		txPool,
		nil,
		addrConv,
		mock.HasherMock{},
		signer,
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
	signer := &mock.SignerMock{}

	txi, err := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
		txPool,
		storer,
		nil,
		mock.HasherMock{},
		signer,
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
	signer := &mock.SignerMock{}

	txi, err := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
		txPool,
		storer,
		addrConv,
		nil,
		signer,
		keyGen,
		oneSharder)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, txi)
}

func TestNewTxInterceptor_NilSignerShouldErr(t *testing.T) {
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
		nil,
		keyGen,
		oneSharder)

	assert.Equal(t, process.ErrNilSingleSigner, err)
	assert.Nil(t, txi)
}

func TestNewTxInterceptor_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}
	signer := &mock.SignerMock{}

	txi, err := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		signer,
		nil,
		oneSharder)

	assert.Equal(t, process.ErrNilKeyGen, err)
	assert.Nil(t, txi)
}

func TestNewTxInterceptor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	storer := &mock.StorerStub{}
	signer := &mock.SignerMock{}

	txi, err := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		signer,
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
	signer := &mock.SignerMock{}

	txi, err := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		signer,
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
	signer := &mock.SignerMock{}

	txi, _ := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		signer,
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
	signer := &mock.SignerMock{}

	txi, _ := transaction.NewTxInterceptor(
		&mock.MarshalizerMock{},
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		signer,
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
	signer := &mock.SignerMock{}

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
		signer,
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
	signer := &mock.SignerMock{}

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
		signer,
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
	signer := &mock.SignerMock{}

	txi, _ := transaction.NewTxInterceptor(
		marshalizer,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		signer,
		keyGen,
		oneSharder)

	txNewer := transaction.NewInterceptedTransaction(signer)
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
	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}
	interceptor.MarshalizerCalled = func() marshal.Marshalizer {
		return &mock.MarshalizerMock{}
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}
	signer := &mock.SignerMock{}

	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		signer,
		keyGen,
		oneSharder)

	txNewer := transaction.NewInterceptedTransaction(signer)
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = []byte("please fail, addrConverter!")
	txNewer.SndAddr = make([]byte, 0)
	txNewer.Value = big.NewInt(0)

	addrConv.CreateAddressFromPublicKeyBytesRetErrForValue = []byte("please fail, addrConverter!")

	assert.Equal(t, process.ErrInvalidRcvAddr, txi.ProcessTx(txNewer, make([]byte, 0)))
}

func TestTransactionInterceptor_ProcessTxVerifySigFailsShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}
	interceptor.MarshalizerCalled = func() marshal.Marshalizer {
		return &mock.MarshalizerMock{}
	}

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}

	pubKey := &mock.SingleSignPublicKey{}
	keyGen := &mock.SingleSignKeyGenMock{}
	keyGen.PublicKeyFromByteArrayCalled = func(b []byte) (key crypto.PublicKey, e error) {
		return pubKey, nil
	}

	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}
	signer := &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			return errors.New("sig not valid")
		},
	}

	txi, _ := transaction.NewTxInterceptor(
		marshalizer,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		signer,
		keyGen,
		oneSharder)

	txNewer := transaction.NewInterceptedTransaction(signer)
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
	keyGen := &mock.SingleSignKeyGenMock{}
	keyGen.PublicKeyFromByteArrayCalled = func(b []byte) (key crypto.PublicKey, e error) {
		return pubKey, nil
	}

	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}
	storer.HasCalled = func(key []byte) (bool, error) {
		return false, nil
	}
	signer := &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			return  nil
		},
	}

	txi, _ := transaction.NewTxInterceptor(
		marshalizer,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		signer,
		keyGen,
		oneSharder)

	txNewer := transaction.NewInterceptedTransaction(signer)
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
	signer := &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			return  nil
		},
	}

	txi, _ := transaction.NewTxInterceptor(
		marshalizer,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		signer,
		keyGen,
		multiSharder)

	txNewer := transaction.NewInterceptedTransaction(signer)
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
	signer := &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			return nil
		},
	}

	txi, _ := transaction.NewTxInterceptor(
		marshalizer,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		signer,
		keyGen,
		multiSharder)

	txNewer := transaction.NewInterceptedTransaction(signer)
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)
	txNewer.Value = big.NewInt(0)

	buff, _ := marshalizer.Marshal(txNewer)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

func TestTransactionInterceptor_ProcessTxPresentInStorerShouldNotAdd(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}
	interceptor.MarshalizerCalled = func() marshal.Marshalizer {
		return &mock.MarshalizerMock{}
	}

	wasAdded := 0

	txPool := &mock.ShardedDataStub{}
	txPool.AddDataCalled = func(key []byte, data interface{}, destShardID uint32) {
		if bytes.Equal(mock.HasherMock{}.Compute(string(buff)), key) {
			wasAdded++
		}
	}
	addrConv := &mock.AddressConverterMock{}

	pubKey := &mock.SingleSignPublicKey{}
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
	signer := &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			return nil
		},
	}

	assert.Nil(t, txi.Validate(msg))
	txi, _ := transaction.NewTxInterceptor(
		interceptor,
		txPool,
		storer,
		addrConv,
		mock.HasherMock{},
		signer,
		keyGen,
		multiSharder)

	txNewer := transaction.NewInterceptedTransaction(signer)
	txNewer.Signature = make([]byte, 0)
	txNewer.Challenge = make([]byte, 0)
	txNewer.RcvAddr = make([]byte, 0)
	txNewer.SndAddr = make([]byte, 0)
	txNewer.Value = big.NewInt(0)

	assert.Nil(t, txi.ProcessTx(txNewer, []byte("txHash")))
	assert.Equal(t, 0, wasAdded)
}
