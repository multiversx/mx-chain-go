package transaction_test

import (
	"bytes"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/stretchr/testify/assert"
)

var durTimeout = time.Duration(time.Second)

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

//------- ProcessReceivedMessage

func TestTransactionInterceptor_ProcessReceivedMessageNilMesssageShouldErr(t *testing.T) {
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

	err := txi.ProcessReceivedMessage(nil)

	assert.Equal(t, process.ErrNilMessage, err)
}

func TestTransactionInterceptor_ProcessReceivedMessageMilMessageDataShouldErr(t *testing.T) {
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

	err := txi.ProcessReceivedMessage(msg)

	assert.Equal(t, process.ErrNilDataToProcess, err)
}

func TestTransactionInterceptor_ProcessReceivedMessageMarshalizerFailsAtUnmarshalingShouldErr(t *testing.T) {
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

	err := txi.ProcessReceivedMessage(msg)

	assert.Equal(t, errMarshalizer, err)
}

func TestTransactionInterceptor_ProcessReceivedMessageNoTransactionInMessageShouldErr(t *testing.T) {
	t.Parallel()

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
				return nil, nil
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

	err := txi.ProcessReceivedMessage(msg)

	assert.Equal(t, process.ErrNoTransactionInMessage, err)
}

func TestTransactionInterceptor_ProcessReceivedMessageIntegrityFailedShouldErr(t *testing.T) {
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

	txNewer := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: nil,
	}
	txNewerBuff, _ := marshalizer.Marshal(txNewer)

	buff, _ := marshalizer.Marshal([][]byte{txNewerBuff})
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	err := txi.ProcessReceivedMessage(msg)

	assert.Equal(t, process.ErrNilSignature, err)
}

func TestTransactionInterceptor_ProcessReceivedMessageIntegrityFailedWithTwoTxsShouldErrAndFilter(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	txPool := &mock.ShardedDataStub{
		AddDataCalled: func(key []byte, data interface{}, cacheId string) {},
	}
	addrConv := &mock.AddressConverterMock{}
	keyGen := &mock.SingleSignKeyGenMock{
		PublicKeyFromByteArrayCalled: func(b []byte) (key crypto.PublicKey, e error) {
			return nil, nil
		},
	}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{
		HasCalled: func(key []byte) error {
			return errors.New("Key not found")
		},
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
		oneSharder)

	tx1 := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: nil,
	}
	tx1Buff, _ := marshalizer.Marshal(tx1)

	tx2 := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}
	tx2Buff, _ := marshalizer.Marshal(tx2)

	buff, _ := marshalizer.Marshal([][]byte{tx1Buff, tx2Buff})
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	txi.SetBroadcastCallback(func(buffToSend []byte) {
		buff = buffToSend
	})
	err := txi.ProcessReceivedMessage(msg)

	assert.Equal(t, process.ErrNilSignature, err)
	//unmarshal data and check there is only tx2 inside
	txBuffRecovered := make([][]byte, 0)
	_ = marshalizer.Unmarshal(&txBuffRecovered, buff)
	assert.Equal(t, 1, len(txBuffRecovered))
	txRecovered := &dataTransaction.Transaction{}
	_ = marshalizer.Unmarshal(txRecovered, txBuffRecovered[0])
	assert.Equal(t, tx2, txRecovered)
}

func TestTransactionInterceptor_ProcessReceivedMessageVerifySigFailsShouldErr(t *testing.T) {
	t.Parallel()

	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}

	marshalizer := &mock.MarshalizerMock{}
	pubKey := &mock.SingleSignPublicKey{}
	keyGen := &mock.SingleSignKeyGenMock{}
	keyGen.PublicKeyFromByteArrayCalled = func(b []byte) (key crypto.PublicKey, e error) {
		return pubKey, nil
	}

	errExpected := errors.New("sig not valid")

	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}
	signer := &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			return errExpected
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

	txNewer := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}
	txNewerBuff, _ := marshalizer.Marshal(txNewer)

	buff, _ := marshalizer.Marshal([][]byte{txNewerBuff})
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	err := txi.ProcessReceivedMessage(msg)

	assert.Equal(t, errExpected, err)
}

func TestTransactionInterceptor_ProcessReceivedMessageOkValsSameShardShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	chanDone := make(chan struct{}, 10)
	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}

	pubKey := &mock.SingleSignPublicKey{}
	keyGen := &mock.SingleSignKeyGenMock{}
	keyGen.PublicKeyFromByteArrayCalled = func(b []byte) (key crypto.PublicKey, e error) {
		return pubKey, nil
	}

	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}
	storer.HasCalled = func(key []byte) error {
		return errors.New("Key not found")
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
		oneSharder)

	txNewer := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}
	txNewerBuff, _ := marshalizer.Marshal(txNewer)

	buff, _ := marshalizer.Marshal([][]byte{txNewerBuff})
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}
	txBuff, _ := marshalizer.Marshal(txNewer)

	txPool.AddDataCalled = func(key []byte, data interface{}, cacheId string) {
		if bytes.Equal(mock.HasherMock{}.Compute(string(txBuff)), key) {
			chanDone <- struct{}{}
		}
	}

	err := txi.ProcessReceivedMessage(msg)

	assert.Nil(t, err)
	select {
	case <-chanDone:
	case <-time.After(durTimeout):
		assert.Fail(t, "timeout while waiting for tx to be inserted in the pool")
	}
}

func TestTransactionInterceptor_ProcessReceivedMessageOkValsOtherShardsShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	chanDone := make(chan struct{}, 10)
	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	pubKey := &mock.SingleSignPublicKey{}
	keyGen := &mock.SingleSignKeyGenMock{}
	keyGen.PublicKeyFromByteArrayCalled = func(b []byte) (key crypto.PublicKey, e error) {
		return pubKey, nil
	}

	multiSharder := mock.NewMultipleShardsCoordinatorMock()
	multiSharder.CurrentShard = 7
	multiSharder.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		return 0
	}
	storer := &mock.StorerStub{}
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

	txNewer := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}
	txNewerBuff, _ := marshalizer.Marshal(txNewer)

	buff, _ := marshalizer.Marshal([][]byte{txNewerBuff})
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	txPool.AddDataCalled = func(key []byte, data interface{}, cacheId string) {
		if bytes.Equal(mock.HasherMock{}.Compute(string(buff)), key) {
			chanDone <- struct{}{}
		}
	}

	err := txi.ProcessReceivedMessage(msg)

	assert.Nil(t, err)
	select {
	case <-chanDone:
		assert.Fail(t, "should have not add tx in pool")
	case <-time.After(durTimeout):
	}
}

func TestTransactionInterceptor_ProcessReceivedMessagePresentInStorerShouldNotAdd(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	chanDone := make(chan struct{}, 10)
	txPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	pubKey := &mock.SingleSignPublicKey{}
	keyGen := &mock.SingleSignKeyGenMock{}
	keyGen.PublicKeyFromByteArrayCalled = func(b []byte) (key crypto.PublicKey, e error) {
		return pubKey, nil
	}
	storer := &mock.StorerStub{}
	storer.HasCalled = func(key []byte) error {
		return nil
	}

	multiSharder := mock.NewMultipleShardsCoordinatorMock()
	multiSharder.CurrentShard = 0
	called := uint32(0)
	multiSharder.ComputeIdCalled = func(address state.AddressContainer) uint32 {
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

	txNewer := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}
	txNewerBuff, _ := marshalizer.Marshal(txNewer)

	buff, _ := marshalizer.Marshal([][]byte{txNewerBuff})
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	txPool.AddDataCalled = func(key []byte, data interface{}, cacheId string) {
		if bytes.Equal(mock.HasherMock{}.Compute(string(buff)), key) {
			chanDone <- struct{}{}
		}
	}

	err := txi.ProcessReceivedMessage(msg)

	assert.Nil(t, err)
	select {
	case <-chanDone:
		assert.Fail(t, "should have not add tx in pool")
	case <-time.After(durTimeout):
	}
}
