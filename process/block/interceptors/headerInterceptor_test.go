package interceptors_test

import (
	"bytes"
	"errors"
	"testing"

	dataBlock "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
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
		mock.NewMultiSigner(),
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
		mock.NewMultiSigner(),
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
		mock.NewMultiSigner(),
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
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilHeadersStorage, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilMultiSignerShouldErr(t *testing.T) {
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
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Nil(t, hi)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
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
		mock.NewMultiSigner(),
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
		mock.NewMultiSigner(),
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
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Nil(t, err)
	assert.NotNil(t, hi)
}

//------- ProcessReceivedMessage

func TestHeaderInterceptor_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, _ := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMessage, hi.ProcessReceivedMessage(nil))
}

func TestHeaderInterceptor_ProcessReceivedMessageNilDataToProcessShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, _ := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{}

	assert.Equal(t, process.ErrNilDataToProcess, hi.ProcessReceivedMessage(msg))
}

func TestHeaderInterceptor_ProcessReceivedMessageMarshalizerErrorsAtUnmarshalingShouldErr(t *testing.T) {
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
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, errMarshalizer, hi.ProcessReceivedMessage(msg))
}

func TestHeaderInterceptor_ProcessReceivedMessageSanityCheckFailedShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}
	marshalizer := &mock.MarshalizerMock{}
	multisigner := mock.NewMultiSigner()

	hi, _ := interceptors.NewHeaderInterceptor(
		marshalizer,
		headers,
		headersNonces,
		storer,
		multisigner,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	hdr := block.NewInterceptedHeader(multisigner)
	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	assert.Equal(t, process.ErrNilPubKeysBitmap, hi.ProcessReceivedMessage(msg))
}

func TestHeaderInterceptor_ProcessReceivedMessageValsOkShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	wasCalled := 0

	testedNonce := uint64(67)

	headers := &mock.ShardedDataStub{}
	multisigner := mock.NewMultiSigner()

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
		multisigner,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	hdr := block.NewInterceptedHeader(multisigner)
	hdr.Nonce = testedNonce
	hdr.ShardId = 0
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = dataBlock.TxBlock
	hdr.Signature = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))
	hdr.RootHash = make([]byte, 0)
	hdr.MiniBlockHeaders = make([]dataBlock.MiniBlockHeader, 0)

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

	assert.Nil(t, hi.ProcessReceivedMessage(msg))
	assert.Equal(t, 2, wasCalled)
}

func TestHeaderInterceptor_ProcessReceivedMessageIsInStorageShouldNotAdd(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	wasCalled := 0

	testedNonce := uint64(67)

	headers := &mock.ShardedDataStub{}
	multisigner := mock.NewMultiSigner()

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
		multisigner,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	hdr := block.NewInterceptedHeader(multisigner)
	hdr.Nonce = testedNonce
	hdr.ShardId = 0
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = dataBlock.TxBlock
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))
	hdr.MiniBlockHeaders = make([]dataBlock.MiniBlockHeader, 0)

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

	assert.Nil(t, hi.ProcessReceivedMessage(msg))
	assert.Equal(t, 0, wasCalled)
}
