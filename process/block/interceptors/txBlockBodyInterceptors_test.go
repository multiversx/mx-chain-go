package interceptors_test

import (
	"bytes"
	"errors"
	"testing"

	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewTxBlockBodyInterceptor

func TestNewTxBlockBodyInterceptor_WithNilParameterShouldErr(t *testing.T) {
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

//------- Validate

func TestTxBlockBodyInterceptor_ValidateNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	tbbi, _ := interceptors.NewTxBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMessage, tbbi.Validate(nil))
}

func TestTxBlockBodyInterceptor_ValidateNilMessageDataShouldErr(t *testing.T) {
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

	assert.Equal(t, process.ErrNilDataToProcess, tbbi.Validate(msg))
}

func TestTxBlockBodyInterceptor_ValidateMarshalizerErrorsAtUnmarshalingShouldErr(t *testing.T) {
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

	assert.Equal(t, errMarshalizer, tbbi.Validate(msg))
}

func TestTxBlockBodyInterceptor_ValidateIntegrityFailsShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	tbbi, _ := interceptors.NewTxBlockBodyInterceptor(
		marshalizer,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	txBlock := block.NewInterceptedTxBlockBody()
	txBlock.RootHash = []byte("root hash")
	txBlock.ShardID = uint32(0)
	txBlock.MiniBlocks = nil

	buff, _ := marshalizer.Marshal(txBlock)

	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	assert.Equal(t, process.ErrNilMiniBlocks, tbbi.Validate(msg))
}

func TestTxBlockBodyInterceptor_ValidateBlockShouldWork(t *testing.T) {
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

	txBlock := block.NewInterceptedTxBlockBody()
	txBlock.RootHash = []byte("root hash")
	txBlock.ShardID = uint32(0)
	txBlock.MiniBlocks = []block2.MiniBlock{
		{
			ShardID:  uint32(0),
			TxHashes: [][]byte{[]byte("tx hash 1")},
		},
	}

	buff, _ := marshalizer.Marshal(txBlock)

	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	putInCacheWasCalled := false
	cache.PutCalled = func(key []byte, value interface{}) (evicted bool) {
		if bytes.Equal(key, mock.HasherMock{}.Compute(string(buff))) {
			putInCacheWasCalled = true
		}
		return false
	}

	assert.Nil(t, tbbi.Validate(msg))
	assert.True(t, putInCacheWasCalled)
}
