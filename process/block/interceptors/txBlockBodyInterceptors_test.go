package interceptors_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	dataBlock "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewTxBlockBodyInterceptor

func TestNewTxBlockBodyInterceptor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	tbbi, err := interceptors.NewTxBlockBodyInterceptor(
		nil,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, tbbi)
}

func TestNewTxBlockBodyInterceptor_NilPoolShouldErr(t *testing.T) {
	t.Parallel()

	storer := &mock.StorerStub{}

	tbbi, err := interceptors.NewTxBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		nil,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilCacher, err)
	assert.Nil(t, tbbi)
}

func TestNewTxBlockBodyInterceptor_NilStorerShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}

	tbbi, err := interceptors.NewTxBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		nil,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilBlockBodyStorage, err)
	assert.Nil(t, tbbi)
}

func TestNewTxBlockBodyInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	tbbi, err := interceptors.NewTxBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		storer,
		nil,
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, tbbi)
}

func TestNewTxBlockBodyInterceptor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	tbbi, err := interceptors.NewTxBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		storer,
		mock.HasherMock{},
		nil)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, tbbi)
}

func TestNewTxBlockBodyInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	tbbi, err := interceptors.NewTxBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Nil(t, err)
	assert.NotNil(t, tbbi)
}

//------- ProcessReceivedMessage

func TestTxBlockBodyInterceptor_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	tbbi, _ := interceptors.NewTxBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMessage, tbbi.ProcessReceivedMessage(nil))
}

func TestTxBlockBodyInterceptor_ProcessReceivedMessageNilMessageDataShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	tbbi, _ := interceptors.NewTxBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{}

	assert.Equal(t, process.ErrNilDataToProcess, tbbi.ProcessReceivedMessage(msg))
}

func TestTxBlockBodyInterceptor_ProcessReceivedMessageMarshalizerErrorsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	tbbi, _ := interceptors.NewTxBlockBodyInterceptor(
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errMarshalizer
			},
		},
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, errMarshalizer, tbbi.ProcessReceivedMessage(msg))
}

func TestTxBlockBodyInterceptor_ProcessReceivedMessageBlockShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{
		HasCalled: func(key []byte) (b bool, e error) {
			return false, nil
		},
	}

	tbbi, _ := interceptors.NewTxBlockBodyInterceptor(
		marshalizer,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	mb := dataBlock.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        [][]byte{[]byte("tx hash 1")},
	}
	miniBlocks := dataBlock.Body{&mb}

	buff, _ := marshalizer.Marshal(miniBlocks)
	miniBlockBuff, _ := marshalizer.Marshal(mb)

	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	chanDone := make(chan struct{}, 10)
	cache.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		if bytes.Equal(key, mock.HasherMock{}.Compute(string(miniBlockBuff))) {
			chanDone <- struct{}{}
		}

		return false, false
	}

	assert.Nil(t, tbbi.ProcessReceivedMessage(msg))
	select {
	case <-chanDone:
	case <-time.After(durTimeout):
		assert.Fail(t, "timeout while waiting for block to be inserted in the pool")
	}
}
