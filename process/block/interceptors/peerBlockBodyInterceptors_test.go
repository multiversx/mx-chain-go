package interceptors_test

import (
	"errors"
	"testing"
	"time"

	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewPeerBlockBodyInterceptor

func TestNewPeerBlockBodyInterceptor_WithNilParameterShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	pbbi, err := interceptors.NewPeerBlockBodyInterceptor(
		nil,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, pbbi)
}

func TestNewPeerBlockBodyInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	pbbi, err := interceptors.NewPeerBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Nil(t, err)
	assert.NotNil(t, pbbi)
}

//------- ProcessReceivedMessage

func TestPeerBlockBodyInterceptor_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	pbbi, _ := interceptors.NewPeerBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMessage, pbbi.ProcessReceivedMessage(nil))
}

func TestPeerBlockBodyInterceptor_ProcessReceivedMessageNilMessageDataShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	pbbi, _ := interceptors.NewPeerBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{}

	assert.Equal(t, process.ErrNilDataToProcess, pbbi.ProcessReceivedMessage(msg))
}

func TestPeerBlockBodyInterceptor_ValidateMarshalizerErrorsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	pbbi, _ := interceptors.NewPeerBlockBodyInterceptor(
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

	assert.Equal(t, errMarshalizer, pbbi.ProcessReceivedMessage(msg))
}

func TestPeerBlockBodyInterceptor_ProcessReceivedMessageBlockShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{
		HasCalled: func(key []byte) error {
			return errors.New("Key not found")
		},
	}

	pbbi, _ := interceptors.NewPeerBlockBodyInterceptor(
		marshalizer,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	peerChangeBlock := block.NewInterceptedPeerBlockBody()
	peerChangeBlock.PeerBlockBody = []*block2.PeerChange{
		{PubKey: []byte("pub key"), ShardIdDest: uint32(0)},
	}

	buff, _ := marshalizer.Marshal(peerChangeBlock)

	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	chanDone := make(chan struct{}, 10)
	cache.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		chanDone <- struct{}{}
		return true, false
	}

	assert.Nil(t, pbbi.ProcessReceivedMessage(msg))
	select {
	case <-chanDone:
	case <-time.After(durTimeout):
		assert.Fail(t, "timeout while waiting for block to be inserted in the pool")
	}
}

func TestPeerBlockBodyInterceptor_ProcessReceivedMessageIsInStorageShouldNotAdd(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{
		HasCalled: func(key []byte) error {
			return nil
		},
	}

	pbbi, _ := interceptors.NewPeerBlockBodyInterceptor(
		marshalizer,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	peerChangeBlock := block.NewInterceptedPeerBlockBody()
	peerChangeBlock.PeerBlockBody = []*block2.PeerChange{
		{PubKey: []byte("pub key"), ShardIdDest: uint32(0)},
	}

	buff, _ := marshalizer.Marshal(peerChangeBlock)

	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	chanDone := make(chan struct{}, 10)
	cache.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		chanDone <- struct{}{}
		return true, false
	}

	assert.Nil(t, pbbi.ProcessReceivedMessage(msg))
	select {
	case <-chanDone:
		assert.Fail(t, "should have not add block in pool")
	case <-time.After(durTimeout):
	}
}
