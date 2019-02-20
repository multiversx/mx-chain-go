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

//------- NewHeaderInterceptor

func TestNewHeaderInterceptor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		nil,
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilHeadersShouldErr(t *testing.T) {
	t.Parallel()

	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		nil,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilHeadersDataPool, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilHeadersNoncesShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.ShardedDataStub{}
	storer := &mock.StorerStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		nil,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilStorerShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		nil,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilHeadersStorage, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		storer,
		nil,
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		nil)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Nil(t, err)
	assert.NotNil(t, hi)
}

//------- Validate

func TestHeaderInterceptor_ValidateNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, _ := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMessage, hi.Validate(nil))
}

func TestHeaderInterceptor_ValidateNilDataToProcessShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, _ := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{}

	assert.Equal(t, process.ErrNilDataToProcess, hi.Validate(msg))
}

func TestHeaderInterceptor_ValidateMarshalizerErrorsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, _ := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errMarshalizer
			},
		},
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, errMarshalizer, hi.Validate(msg))
}

func TestHeaderInterceptor_ValidateSanityCheckFailedShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}
	marshalizer := &mock.MarshalizerMock{}

	hi, _ := interceptors.NewHeaderInterceptor(
		marshalizer,
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	hdr := block.NewInterceptedHeader()
	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	assert.Equal(t, process.ErrNilBlockBodyHash, hi.Validate(msg))
}

func TestHeaderInterceptor_ValidateValsOkShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	wasCalled := 0

	testedNonce := uint64(67)

	headers := &mock.ShardedDataStub{}

	headersNonces := &mock.Uint64CacherStub{}
	headersNonces.HasOrAddCalled = func(u uint64, i []byte) (b bool, b2 bool) {
		if u == testedNonce {
			wasCalled++
		}

		return
	}

	storer := &mock.StorerStub{}
	storer.HasCalled = func(key []byte) (bool, error) {
		return false, nil
	}

	hi, _ := interceptors.NewHeaderInterceptor(
		marshalizer,
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	hdr := block.NewInterceptedHeader()
	hdr.Nonce = testedNonce
	hdr.ShardId = 0
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.TxBlock
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))

	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	headers.AddDataCalled = func(key []byte, data interface{}, destShardID uint32) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			wasCalled++
		}
	}

	assert.Nil(t, hi.Validate(msg))
	assert.Equal(t, 2, wasCalled)
}

func TestHeaderInterceptor_ValidateIsInStorageShouldNotAdd(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	wasCalled := 0

	testedNonce := uint64(67)

	headers := &mock.ShardedDataStub{}

	headersNonces := &mock.Uint64CacherStub{}
	headersNonces.HasOrAddCalled = func(u uint64, i []byte) (b bool, b2 bool) {
		if u == testedNonce {
			wasCalled++
		}

		return
	}

	storer := &mock.StorerStub{}
	storer.HasCalled = func(key []byte) (bool, error) {
		return true, nil
	}

	hi, _ := interceptors.NewHeaderInterceptor(
		marshalizer,
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	hdr := block.NewInterceptedHeader()
	hdr.Nonce = testedNonce
	hdr.ShardId = 0
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.TxBlock
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))

	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	headers.AddDataCalled = func(key []byte, data interface{}, destShardID uint32) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			wasCalled++
		}
	}

	assert.Nil(t, hi.Validate(msg))
	assert.Equal(t, 0, wasCalled)
}
