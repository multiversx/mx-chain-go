package interceptors_test

import (
	"bytes"
	"errors"
	"sync"
	"testing"
	"time"

	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

var durTimeout = time.Duration(time.Second)

//------- NewHeaderInterceptor

func TestNewHeaderInterceptor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		nil,
		headers,
		headersNonces,
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

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
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

	assert.Equal(t, process.ErrNilHeadersDataPool, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilHeadersNoncesShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		nil,
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, hi)
}

//------- ProcessReceivedMessage

func TestHeaderInterceptor_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, _ := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

	assert.Equal(t, process.ErrNilMessage, hi.ProcessReceivedMessage(nil))
}

func TestHeaderInterceptor_ProcessReceivedMessageValsOkShouldWork(t *testing.T) {
	t.Parallel()

	chanDone := make(chan struct{}, 1)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	testedNonce := uint64(67)
	marshalizer := &mock.MarshalizerMock{}
	headers := &mock.CacherStub{}

	multisigner := mock.NewMultiSigner()
	chronologyValidator := &mock.ChronologyValidatorStub{
		ValidateReceivedBlockCalled: func(shardID uint32, epoch uint32, nonce uint64, round uint32) error {
			return nil
		},
	}
	headersNonces := &mock.Uint64CacherStub{}
	headersNonces.HasOrAddCalled = func(u uint64, i interface{}) (b bool, b2 bool) {
		if u == testedNonce {
			wg.Done()
		}

		return
	}
	storer := &mock.StorerStub{}
	storer.HasCalled = func(key []byte) error {
		return errors.New("Key not found")
	}

	hi, _ := interceptors.NewHeaderInterceptor(
		marshalizer,
		headers,
		headersNonces,
		storer,
		multisigner,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		chronologyValidator,
	)

	hdr := block.NewInterceptedHeader(multisigner, chronologyValidator)
	hdr.Nonce = testedNonce
	hdr.ShardId = 0
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = dataBlock.TxBlock
	hdr.Signature = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)
	hdr.MiniBlockHeaders = make([]dataBlock.MiniBlockHeader, 0)

	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	headers.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			wg.Done()
		}
		return false, false
	}

	go func() {
		wg.Wait()
		chanDone <- struct{}{}
	}()

	assert.Nil(t, hi.ProcessReceivedMessage(msg))
	select {
	case <-chanDone:
	case <-time.After(durTimeout):
		assert.Fail(t, "timeout while waiting for block to be inserted in the pool")
	}
}

func TestHeaderInterceptor_ProcessReceivedMessageIsInStorageShouldNotAdd(t *testing.T) {
	t.Parallel()

	chanDone := make(chan struct{}, 10)
	testedNonce := uint64(67)
	marshalizer := &mock.MarshalizerMock{}
	headers := &mock.CacherStub{}

	multisigner := mock.NewMultiSigner()
	chronologyValidator := &mock.ChronologyValidatorStub{
		ValidateReceivedBlockCalled: func(shardID uint32, epoch uint32, nonce uint64, round uint32) error {
			return nil
		},
	}
	headersNonces := &mock.Uint64CacherStub{}
	headersNonces.HasOrAddCalled = func(u uint64, i interface{}) (b bool, b2 bool) {
		if u == testedNonce {
			chanDone <- struct{}{}
		}

		return
	}

	storer := &mock.StorerStub{}
	storer.HasCalled = func(key []byte) error {
		return nil
	}

	hi, _ := interceptors.NewHeaderInterceptor(
		marshalizer,
		headers,
		headersNonces,
		storer,
		multisigner,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		chronologyValidator,
	)

	hdr := block.NewInterceptedHeader(multisigner, chronologyValidator)
	hdr.Nonce = testedNonce
	hdr.ShardId = 0
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = dataBlock.TxBlock
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)
	hdr.MiniBlockHeaders = make([]dataBlock.MiniBlockHeader, 0)

	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	headers.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			chanDone <- struct{}{}
		}
		return false, false
	}

	assert.Nil(t, hi.ProcessReceivedMessage(msg))
	select {
	case <-chanDone:
		assert.Fail(t, "should have not add block in pool")
	case <-time.After(durTimeout):
	}
}

func TestHeaderInterceptor_ProcessReceivedMessageNotForCurrentShardShouldNotAdd(t *testing.T) {
	t.Parallel()

	chanDone := make(chan struct{}, 10)
	testedNonce := uint64(67)
	marshalizer := &mock.MarshalizerMock{}
	headers := &mock.CacherStub{}

	multisigner := mock.NewMultiSigner()
	chronologyValidator := &mock.ChronologyValidatorStub{
		ValidateReceivedBlockCalled: func(shardID uint32, epoch uint32, nonce uint64, round uint32) error {
			return nil
		},
	}
	headersNonces := &mock.Uint64CacherStub{}
	headersNonces.HasOrAddCalled = func(u uint64, i interface{}) (b bool, b2 bool) {
		if u == testedNonce {
			chanDone <- struct{}{}
		}

		return
	}
	storer := &mock.StorerStub{}
	storer.HasCalled = func(key []byte) error {
		return errors.New("Key not found")
	}
	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = 2
	shardCoordinator.SetNoShards(5)

	hi, _ := interceptors.NewHeaderInterceptor(
		marshalizer,
		headers,
		headersNonces,
		storer,
		multisigner,
		mock.HasherMock{},
		shardCoordinator,
		chronologyValidator,
	)

	hdr := block.NewInterceptedHeader(multisigner, chronologyValidator)
	hdr.Nonce = testedNonce
	hdr.ShardId = 0
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = dataBlock.TxBlock
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)
	hdr.MiniBlockHeaders = make([]dataBlock.MiniBlockHeader, 0)

	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	headers.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			chanDone <- struct{}{}
		}
		return false, false
	}

	assert.Nil(t, hi.ProcessReceivedMessage(msg))

}
