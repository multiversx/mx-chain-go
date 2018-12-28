package block

import (
	"bytes"
	"testing"

	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

//------- HeaderInterceptor

//NewHeaderInterceptor

func TestNewHeaderInterceptor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}

	hi, err := NewHeaderInterceptor(
		nil,
		headers,
		headersNonces,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilInterceptor, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilHeadersShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	headersNonces := &mock.Uint64CacherStub{}

	hi, err := NewHeaderInterceptor(
		interceptor,
		nil,
		headersNonces,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilHeadersDataPool, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilHeadersNoncesShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	headers := &mock.ShardedDataStub{}

	hi, err := NewHeaderInterceptor(
		interceptor,
		headers,
		nil,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}

	hi, err := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		nil,
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}

	hi, err := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		mock.HasherMock{},
		nil)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}

	hi, err := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Nil(t, err)
	assert.NotNil(t, hi)
}

//processHdr

func TestHeaderInterceptor_ProcessHdrNilHdrShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}

	hi, _ := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilBlockHeader, hi.ProcessHdr(nil, make([]byte, 0)))
}

func TestHeaderInterceptor_ProcessHdrNilDataToProcessShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}

	hi, _ := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilDataToProcess, hi.ProcessHdr(NewInterceptedHeader(), nil))
}

func TestHeaderInterceptor_ProcessHdrWrongTypeOfCreatorShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}

	hi, _ := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrBadInterceptorTopicImplementation,
		hi.ProcessHdr(&mock.StringCreator{}, make([]byte, 0)))
}

func TestHeaderInterceptor_ProcessHdrSanityCheckFailedShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}

	hi, _ := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilBlockBodyHash, hi.ProcessHdr(NewInterceptedHeader(), make([]byte, 0)))
}

func TestHeaderInterceptor_ProcessOkValsShouldWork(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	wasCalled := 0

	testedNonce := uint64(67)

	headers := &mock.ShardedDataStub{}
	headers.AddDataCalled = func(key []byte, data interface{}, destShardID uint32) {
		aaaHash := mock.HasherMock{}.Compute("aaa")
		if bytes.Equal(aaaHash, key) {
			wasCalled++
		}
	}

	headersNonces := &mock.Uint64CacherStub{}
	headersNonces.HasOrAddCalled = func(u uint64, i []byte) (b bool, b2 bool) {
		if u == testedNonce {
			wasCalled++
		}

		return
	}

	hi, _ := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	hdr := NewInterceptedHeader()
	hdr.Nonce = testedNonce
	hdr.ShardId = 0
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.TxBlock
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))

	assert.Nil(t, hi.ProcessHdr(hdr, []byte("aaa")))
	assert.Equal(t, 2, wasCalled)
}

//------- BlockBodyInterceptor

//NewBlockBodyInterceptor

func TestNewBlockBodyInterceptor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}

	gbbi, err := NewGenericBlockBodyInterceptor(
		nil,
		cache,
		mock.HasherMock{},
		NewInterceptedTxBlockBody(),
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilInterceptor, err)
	assert.Nil(t, gbbi)
}

func TestNewBlockBodyInterceptor_NilPoolShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}

	gbbi, err := NewGenericBlockBodyInterceptor(
		interceptor,
		nil,
		mock.HasherMock{},
		NewInterceptedTxBlockBody(),
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilCacher, err)
	assert.Nil(t, gbbi)
}

func TestNewBlockBodyInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}

	gbbi, err := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		nil,
		NewInterceptedTxBlockBody(),
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, gbbi)
}

func TestNewBlockBodyInterceptor_NilTemplateObjectShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}

	gbbi, err := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		mock.HasherMock{},
		nil,
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilTemplateObj, err)
	assert.Nil(t, gbbi)
}

func TestNewBlockBodyInterceptor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}

	gbbi, err := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		mock.HasherMock{},
		NewInterceptedTxBlockBody(),
		nil)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, gbbi)
}

func TestNewBlockBodyInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	gbbi, err := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		mock.HasherMock{},
		NewInterceptedTxBlockBody(),
		mock.NewOneShardCoordinatorMock())

	assert.Nil(t, err)
	assert.NotNil(t, gbbi)
}

//processBodyBlock

func TestBlockBodyInterceptor_ProcessNilHdrShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		mock.HasherMock{},
		NewInterceptedTxBlockBody(),
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilBlockBody, gbbi.ProcessBodyBlock(nil, make([]byte, 0)))
}

func TestBlockBodyInterceptor_ProcessNilDataToProcessShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		mock.HasherMock{},
		NewInterceptedTxBlockBody(),
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilDataToProcess,
		gbbi.ProcessBodyBlock(NewInterceptedTxBlockBody(), nil))
}

func TestBlockBodyInterceptor_ProcessHdrWrongTypeOfNewerShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		mock.HasherMock{},
		NewInterceptedTxBlockBody(),
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrBadInterceptorTopicImplementation,
		gbbi.ProcessBodyBlock(&mock.StringCreator{}, make([]byte, 0)))
}

func TestBlockBodyInterceptor_ProcessHdrSanityCheckFailedShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		mock.HasherMock{},
		NewInterceptedTxBlockBody(),
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilRootHash,
		gbbi.ProcessBodyBlock(NewInterceptedTxBlockBody(), make([]byte, 0)))
}

func TestBlockBodyInterceptor_ProcessOkValsShouldRetTrue(t *testing.T) {
	t.Parallel()

	wasCalled := 0

	cache := &mock.CacherStub{}
	cache.PutCalled = func(key []byte, value interface{}) (evicted bool) {
		if bytes.Equal(mock.HasherMock{}.Compute("aaa"), key) {
			wasCalled++
		}

		return
	}

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Creator, rawData []byte) error) {
	}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		mock.HasherMock{},
		NewInterceptedTxBlockBody(),
		mock.NewOneShardCoordinatorMock())

	miniBlock := block2.MiniBlock{}
	miniBlock.TxHashes = append(miniBlock.TxHashes, []byte{65})

	txBody := NewInterceptedTxBlockBody()
	txBody.ShardID = 0
	txBody.MiniBlocks = make([]block2.MiniBlock, 0)
	txBody.MiniBlocks = append(txBody.MiniBlocks, miniBlock)
	txBody.RootHash = make([]byte, 0)

	assert.Nil(t, gbbi.ProcessBodyBlock(txBody, []byte("aaa")))
	assert.Equal(t, 1, wasCalled)
}
