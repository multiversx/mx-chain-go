package interceptors_test

import (
	"bytes"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/mock"
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
		&mock.Uint64SyncMapCacherStub{},
		metachainStorer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_NilMetachainHeadersShouldErr(t *testing.T) {
	t.Parallel()

	metachainStorer := &mock.StorerStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		nil,
		&mock.Uint64SyncMapCacherStub{},
		metachainStorer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

	assert.Equal(t, process.ErrNilMetaHeadersDataPool, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_NilMetachainHeadersNoncesShouldErr(t *testing.T) {
	t.Parallel()

	metachainStorer := &mock.StorerStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		&mock.CacherStub{},
		nil,
		metachainStorer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

	assert.Equal(t, process.ErrNilMetaHeadersNoncesDataPool, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_NilMetachainStorerShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		&mock.Uint64SyncMapCacherStub{},
		nil,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

	assert.Equal(t, process.ErrNilMetaHeadersStorage, err)
	assert.Nil(t, mhi)
}

func TestNewMetachainHeaderInterceptor_NilMultiSignerShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	metachainStorer := &mock.StorerStub{}

	mhi, err := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		&mock.Uint64SyncMapCacherStub{},
		metachainStorer,
		nil,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

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
		&mock.Uint64SyncMapCacherStub{},
		metachainStorer,
		mock.NewMultiSigner(),
		nil,
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

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
		&mock.Uint64SyncMapCacherStub{},
		metachainStorer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		nil,
		&mock.ChronologyValidatorStub{},
	)

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
		&mock.Uint64SyncMapCacherStub{},
		metachainStorer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

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
		&mock.Uint64SyncMapCacherStub{},
		metachainStorer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

	assert.Equal(t, process.ErrNilMessage, mhi.ProcessReceivedMessage(nil))
}

func TestMetachainHeaderInterceptor_ProcessReceivedMessageNilDataToProcessShouldErr(t *testing.T) {
	t.Parallel()

	metachainHeaders := &mock.CacherStub{}
	metachainStorer := &mock.StorerStub{}

	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		&mock.MarshalizerMock{},
		metachainHeaders,
		&mock.Uint64SyncMapCacherStub{},
		metachainStorer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

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
		&mock.Uint64SyncMapCacherStub{},
		metachainStorer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

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
	chronologyValidator := &mock.ChronologyValidatorStub{
		ValidateReceivedBlockCalled: func(shardID uint32, epoch uint32, nonce uint64, round uint64) error {
			return nil
		},
	}
	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		marshalizer,
		metachainHeaders,
		&mock.Uint64SyncMapCacherStub{},
		metachainStorer,
		multisigner,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		chronologyValidator,
	)

	hdr := block.NewInterceptedMetaHeader(multisigner, chronologyValidator)
	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	assert.Equal(t, process.ErrNilPubKeysBitmap, mhi.ProcessReceivedMessage(msg))
}

func TestMetachainHeaderInterceptor_ProcessReceivedMessageValsOkShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	chanDone := make(chan struct{}, 1)
	testedNonce := uint64(67)
	metachainHeaders := &mock.CacherStub{}
	metachainHeadersNonces := &mock.Uint64SyncMapCacherStub{}
	metachainStorer := &mock.StorerStub{
		HasCalled: func(key []byte) error {
			return errors.New("key not found")
		},
	}
	multisigner := mock.NewMultiSigner()
	chronologyValidator := &mock.ChronologyValidatorStub{
		ValidateReceivedBlockCalled: func(shardID uint32, epoch uint32, nonce uint64, round uint64) error {
			return nil
		},
	}
	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		marshalizer,
		metachainHeaders,
		metachainHeadersNonces,
		metachainStorer,
		multisigner,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		chronologyValidator,
	)

	hdr := block.NewInterceptedMetaHeader(multisigner, chronologyValidator)
	hdr.Nonce = testedNonce
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	metachainHeaders.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			wg.Done()
		}
		return
	}
	metachainHeadersNonces.MergeCalled = func(nonce uint64, src dataRetriever.ShardIdHashMap) {
		if nonce != testedNonce {
			return
		}

		aaaHash := mock.HasherMock{}.Compute(string(buff))
		src.Range(func(sharId uint32, hash []byte) bool {
			if bytes.Equal(aaaHash, hash) {
				wg.Done()

				return false
			}

			return true
		})
	}

	go func() {
		wg.Wait()
		chanDone <- struct{}{}
	}()

	assert.Nil(t, mhi.ProcessReceivedMessage(msg))
	select {
	case <-chanDone:
	case <-time.After(durTimeout):
		assert.Fail(t, "timeout while waiting for block to be inserted in the pool")
	}
}

func TestMetachainHeaderInterceptor_ProcessReceivedMessageIsInStorageShouldNotAdd(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	chanDone := make(chan struct{}, 1)
	testedNonce := uint64(67)
	multisigner := mock.NewMultiSigner()
	chronologyValidator := &mock.ChronologyValidatorStub{
		ValidateReceivedBlockCalled: func(shardID uint32, epoch uint32, nonce uint64, round uint64) error {
			return nil
		},
	}
	metachainHeaders := &mock.CacherStub{}
	metachainHeadersNonces := &mock.Uint64SyncMapCacherStub{}
	metachainStorer := &mock.StorerStub{
		HasCalled: func(key []byte) error {
			return nil
		},
	}
	mhi, _ := interceptors.NewMetachainHeaderInterceptor(
		marshalizer,
		metachainHeaders,
		metachainHeadersNonces,
		metachainStorer,
		multisigner,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		chronologyValidator,
	)

	hdr := block.NewInterceptedMetaHeader(multisigner, chronologyValidator)
	hdr.Nonce = testedNonce
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	metachainHeaders.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			chanDone <- struct{}{}
		}
		return
	}
	metachainHeadersNonces.MergeCalled = func(nonce uint64, src dataRetriever.ShardIdHashMap) {
		if nonce != testedNonce {
			return
		}

		aaaHash := mock.HasherMock{}.Compute(string(buff))
		src.Range(func(shardId uint32, hash []byte) bool {
			if bytes.Equal(aaaHash, hash) {
				chanDone <- struct{}{}

				return false
			}

			return true
		})
	}

	assert.Nil(t, mhi.ProcessReceivedMessage(msg))
	select {
	case <-chanDone:
		assert.Fail(t, "should have not add block in pool")
	case <-time.After(durTimeout):
	}
}
