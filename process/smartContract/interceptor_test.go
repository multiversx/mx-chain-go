package smartContract_test

import (
	"bytes"
	"errors"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

var durTimeout = time.Duration(time.Second)

//------- NewScrInterceptor

func TestNewScrInterceptor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	scrPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	scri, err := smartContract.NewScrInterceptor(
		nil,
		scrPool,
		storer,
		addrConv,
		mock.HasherMock{},
		oneSharder)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, scri)
}

func TestNewScrInterceptor_NilTransactionPoolShouldErr(t *testing.T) {
	t.Parallel()

	addrConv := &mock.AddressConverterMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	scri, err := smartContract.NewScrInterceptor(
		&mock.MarshalizerMock{},
		nil,
		storer,
		addrConv,
		mock.HasherMock{},
		oneSharder)

	assert.Equal(t, process.ErrNilScrDataPool, err)
	assert.Nil(t, scri)
}

func TestNewScrInterceptor_NilStorerShouldErr(t *testing.T) {
	t.Parallel()

	scrPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()

	scri, err := smartContract.NewScrInterceptor(
		&mock.MarshalizerMock{},
		scrPool,
		nil,
		addrConv,
		mock.HasherMock{},
		oneSharder)

	assert.Equal(t, process.ErrNilScrStorage, err)
	assert.Nil(t, scri)
}

func TestNewScrInterceptor_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	scrPool := &mock.ShardedDataStub{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	scri, err := smartContract.NewScrInterceptor(
		&mock.MarshalizerMock{},
		scrPool,
		storer,
		nil,
		mock.HasherMock{},
		oneSharder)

	assert.Equal(t, process.ErrNilAddressConverter, err)
	assert.Nil(t, scri)
}

func TestNewScrInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	scrPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	scri, err := smartContract.NewScrInterceptor(
		&mock.MarshalizerMock{},
		scrPool,
		storer,
		addrConv,
		nil,
		oneSharder)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, scri)
}

func TestNewScrInterceptor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	scrPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	storer := &mock.StorerStub{}

	scri, err := smartContract.NewScrInterceptor(
		&mock.MarshalizerMock{},
		scrPool,
		storer,
		addrConv,
		mock.HasherMock{},
		nil)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, scri)
}

func TestNewScrInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	scrPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	scri, err := smartContract.NewScrInterceptor(
		&mock.MarshalizerMock{},
		scrPool,
		storer,
		addrConv,
		mock.HasherMock{},
		oneSharder)

	assert.Nil(t, err)
	assert.NotNil(t, scri)
}

//------- ProcessReceivedMessage

func TestTransactionInterceptor_ProcessReceivedMessageNilMesssageShouldErr(t *testing.T) {
	t.Parallel()

	scrPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	scri, _ := smartContract.NewScrInterceptor(
		&mock.MarshalizerMock{},
		scrPool,
		storer,
		addrConv,
		mock.HasherMock{},
		oneSharder)

	err := scri.ProcessReceivedMessage(nil)

	assert.Equal(t, process.ErrNilMessage, err)
}

func TestTransactionInterceptor_ProcessReceivedMessageMilMessageDataShouldErr(t *testing.T) {
	t.Parallel()

	scrPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	scri, _ := smartContract.NewScrInterceptor(
		&mock.MarshalizerMock{},
		scrPool,
		storer,
		addrConv,
		mock.HasherMock{},
		oneSharder)

	msg := &mock.P2PMessageMock{}

	err := scri.ProcessReceivedMessage(msg)

	assert.Equal(t, process.ErrNilDataToProcess, err)
}

func TestTransactionInterceptor_ProcessReceivedMessageMarshalizerFailsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")

	scrPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	scri, _ := smartContract.NewScrInterceptor(
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errMarshalizer
			},
		},
		scrPool,
		storer,
		addrConv,
		mock.HasherMock{},
		oneSharder)

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	err := scri.ProcessReceivedMessage(msg)

	assert.Equal(t, errMarshalizer, err)
}

func TestTransactionInterceptor_ProcessReceivedMessageNoTransactionInMessageShouldErr(t *testing.T) {
	t.Parallel()

	scrPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}

	scri, _ := smartContract.NewScrInterceptor(
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return nil
			},
			MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
				return nil, nil
			},
		},
		scrPool,
		storer,
		addrConv,
		mock.HasherMock{},
		oneSharder)

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	err := scri.ProcessReceivedMessage(msg)

	assert.Equal(t, process.ErrNoSmartContractResultInMessage, err)
}

func TestTransactionInterceptor_ProcessReceivedMessageOkValsSameShardShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	chanDone := make(chan struct{}, 10)
	scrPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
	oneSharder := mock.NewOneShardCoordinatorMock()
	storer := &mock.StorerStub{}
	storer.HasCalled = func(key []byte) error {
		return errors.New("Key not found")
	}

	scri, _ := smartContract.NewScrInterceptor(
		marshalizer,
		scrPool,
		storer,
		addrConv,
		mock.HasherMock{},
		oneSharder)

	scrNewer := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(2),
		Data:    []byte("data"),
		RcvAddr: recvAddress,
		SndAddr: senderAddress,
		TxHash:  []byte("txHash"),
	}
	scrNewerBuff, _ := marshalizer.Marshal(scrNewer)

	buff, _ := marshalizer.Marshal([][]byte{scrNewerBuff})
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}
	scrBuff, _ := marshalizer.Marshal(scrNewer)

	scrPool.AddDataCalled = func(key []byte, data interface{}, cacheId string) {
		if bytes.Equal(mock.HasherMock{}.Compute(string(scrBuff)), key) {
			chanDone <- struct{}{}
		}
	}

	err := scri.ProcessReceivedMessage(msg)

	assert.Nil(t, err)
	select {
	case <-chanDone:
	case <-time.After(durTimeout):
		assert.Fail(t, "timeout while waiting for scr to be inserted in the pool")
	}
}

func TestTransactionInterceptor_ProcessReceivedMessageOkValsOtherShardsShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	chanDone := make(chan struct{}, 10)
	scrPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}

	multiSharder := mock.NewMultipleShardsCoordinatorMock()
	multiSharder.CurrentShard = 7
	multiSharder.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		return 0
	}
	storer := &mock.StorerStub{}

	scri, _ := smartContract.NewScrInterceptor(
		marshalizer,
		scrPool,
		storer,
		addrConv,
		mock.HasherMock{},
		multiSharder)

	scrNewer := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(2),
		Data:    []byte("data"),
		RcvAddr: recvAddress,
		SndAddr: senderAddress,
		TxHash:  []byte("txHash"),
	}
	scrNewerBuff, _ := marshalizer.Marshal(scrNewer)

	buff, _ := marshalizer.Marshal([][]byte{scrNewerBuff})
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	scrPool.AddDataCalled = func(key []byte, data interface{}, cacheId string) {
		if bytes.Equal(mock.HasherMock{}.Compute(string(buff)), key) {
			chanDone <- struct{}{}
		}
	}

	err := scri.ProcessReceivedMessage(msg)

	assert.Nil(t, err)
	select {
	case <-chanDone:
		assert.Fail(t, "should have not add scr in pool")
	case <-time.After(durTimeout):
	}
}

func TestTransactionInterceptor_ProcessReceivedMessagePresentInStorerShouldNotAdd(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	chanDone := make(chan struct{}, 10)
	scrPool := &mock.ShardedDataStub{}
	addrConv := &mock.AddressConverterMock{}
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

	scri, _ := smartContract.NewScrInterceptor(
		marshalizer,
		scrPool,
		storer,
		addrConv,
		mock.HasherMock{},
		multiSharder)

	scrNewer := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(2),
		Data:    []byte("data"),
		RcvAddr: recvAddress,
		SndAddr: senderAddress,
		TxHash:  []byte("txHash"),
	}
	scrNewerBuff, _ := marshalizer.Marshal(scrNewer)

	buff, _ := marshalizer.Marshal([][]byte{scrNewerBuff})
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	scrPool.AddDataCalled = func(key []byte, data interface{}, cacheId string) {
		if bytes.Equal(mock.HasherMock{}.Compute(string(buff)), key) {
			chanDone <- struct{}{}
		}
	}

	err := scri.ProcessReceivedMessage(msg)

	assert.Nil(t, err)
	select {
	case <-chanDone:
		assert.Fail(t, "should have not add scr in pool")
	case <-time.After(durTimeout):
	}
}
