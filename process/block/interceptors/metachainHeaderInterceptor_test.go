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

//------- NewMetachainHeaderInterceptor

func TestNewMetachainHeaderInterceptor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	metachainStorer := &mock.StorerStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		nil,
		metachainHeaders,
		metachainStorer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_NilMetachainHeadersShouldErr(t *testing.T) {
	t.Parallel()

	metachainStorer := &mock.StorerStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		nil,
		metachainStorer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMetachainHeadersDataPool, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_NilMetachainStorerShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		nil,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMetachainHeadersStorage, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_NilMultiSignerShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	metachainStorer := &mock.StorerStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		metachainStorer,
		nil,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Nil(t, mhi)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewMetachainHeaderInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	metachainStorer := &mock.StorerStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		metachainStorer,
		mock.NewMultiSigner(),
		nil,
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	metachainStorer := &mock.StorerStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		metachainStorer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		nil)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	metachainStorer := &mock.StorerStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		metachainStorer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Nil(t, err)
	assert.NotNil(t, mhi)
}

//------- ProcessReceivedMessage

func TestMetachainHeaderInterceptor_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	metachainStorer := &mock.StorerStub{}

	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		metachainStorer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMessage, mhi.ProcessReceivedMessage(nil))
}

func TestMetachainHeaderInterceptor_ProcessReceivedMessageNilDataToProcessShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	metachainStorer := &mock.StorerStub{}

	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		metachainStorer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{}

	assert.Equal(t, process.ErrNilDataToProcess, mhi.ProcessReceivedMessage(msg))
}

func TestMetachainHeaderInterceptor_ProcessReceivedMessageMarshalizerErrorsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")
	metachainHeaders := &mock.CacherStub{}
	metachainStorer := &mock.StorerStub{}

	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errMarshalizer
			},
		},
		metachainHeaders,
		metachainStorer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, errMarshalizer, mhi.ProcessReceivedMessage(msg))
}

func TestMetachainHeaderInterceptor_ProcessReceivedMessageSanityCheckFailedShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	metachainStorer := &mock.StorerStub{}
	marshalizer := &mock.MarshalizerMock{}
	multisigner := mock.NewMultiSigner()

	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		marshalizer,
		metachainHeaders,
		metachainStorer,
		multisigner,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	hdr := block.NewInterceptedHeader(multisigner)
	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	assert.Equal(t, process.ErrNilPubKeysBitmap, mhi.ProcessReceivedMessage(msg))
}

func TestMetachainHeaderInterceptor_ProcessReceivedMessageValsOkShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	wasCalled := 0
	testedNonce := uint64(67)
	metachainHeaders := &mock.CacherStub{}
	metachainStorer := &mock.StorerStub{
		HasCalled: func(key []byte) (b bool, e error) {
			return false, nil
		},
	}
	multisigner := mock.NewMultiSigner()

	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		marshalizer,
		metachainHeaders,
		metachainStorer,
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

	metachainHeaders.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			wasCalled++
		}
		return
	}

	assert.Nil(t, mhi.ProcessReceivedMessage(msg))
	assert.Equal(t, 1, wasCalled)
}

func TestMetachainHeaderInterceptor_ProcessReceivedMessageIsInStorageShouldNotAdd(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	wasCalled := 0
	testedNonce := uint64(67)
	multisigner := mock.NewMultiSigner()
	metachainHeaders := &mock.CacherStub{}
	metachainStorer := &mock.StorerStub{
		HasCalled: func(key []byte) (b bool, e error) {
			return true, nil
		},
	}
	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		marshalizer,
		metachainHeaders,
		metachainStorer,
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

	metachainHeaders.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			wasCalled++
		}
		return
	}

	assert.Nil(t, mhi.ProcessReceivedMessage(msg))
	assert.Equal(t, 0, wasCalled)
}
