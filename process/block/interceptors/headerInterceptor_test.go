package interceptors_test

import (
	"bytes"
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

	headers := &mock.CacherStub{}
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
		mock.NewOneShardCoordinatorMock())

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
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMessage, hi.ProcessReceivedMessage(nil))
}

func TestHeaderInterceptor_ProcessReceivedMessageValsOkShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	wasCalled := 0
	testedNonce := uint64(67)
	headers := &mock.CacherStub{}
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

	headers.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			wasCalled++
		}
		return false, false
	}

	assert.Nil(t, hi.ProcessReceivedMessage(msg))
	assert.Equal(t, 2, wasCalled)
}

func TestHeaderInterceptor_ProcessReceivedMessageIsInStorageShouldNotAdd(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	wasCalled := 0

	testedNonce := uint64(67)

	headers := &mock.CacherStub{}
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

	headers.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			wasCalled++
		}
		return false, false
	}

	assert.Nil(t, hi.ProcessReceivedMessage(msg))
	assert.Equal(t, 0, wasCalled)
}

func TestHeaderInterceptor_ProcessReceivedMessageNotForCurrentShardShouldNotAdd(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	wasCalled := 0

	testedNonce := uint64(67)

	headers := &mock.CacherStub{}
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
		shardCoordinator)

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

	headers.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			wasCalled++
		}
		return false, false
	}

	assert.Nil(t, hi.ProcessReceivedMessage(msg))
	assert.Equal(t, 0, wasCalled)
}
