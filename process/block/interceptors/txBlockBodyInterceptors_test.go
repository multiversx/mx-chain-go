package interceptors_test

import (
	"bytes"
	"errors"
	"testing"

	dataBlock "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewMiniBlocksInterceptor

func TestNewTxBlockBodyInterceptor_WithNilParameterShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	tbbi, err := interceptors.NewMiniBlocksInterceptor(
		nil,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, tbbi)
}

func TestNewTxBlockBodyInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	tbbi, err := interceptors.NewMiniBlocksInterceptor(
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

	tbbi, _ := interceptors.NewMiniBlocksInterceptor(
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

	tbbi, _ := interceptors.NewMiniBlocksInterceptor(
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

	tbbi, _ := interceptors.NewMiniBlocksInterceptor(
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

	tbbi, _ := interceptors.NewMiniBlocksInterceptor(
		marshalizer,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	mb := dataBlock.MiniBlock{
		ShardID:  uint32(0),
		TxHashes: [][]byte{[]byte("tx hash 1")},
	}
	miniBlocks := dataBlock.Body{&mb}

	buff, _ := marshalizer.Marshal(miniBlocks)
	miniBlockBuff, _ := marshalizer.Marshal(mb)

	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	putInCacheWasCalled := false
	cache.PutCalled = func(key []byte, value interface{}) (evicted bool) {
		if bytes.Equal(key, mock.HasherMock{}.Compute(string(miniBlockBuff))) {
			putInCacheWasCalled = true
		}
		return false
	}

	assert.Nil(t, tbbi.ProcessReceivedMessage(msg))
	assert.True(t, putInCacheWasCalled)
}
