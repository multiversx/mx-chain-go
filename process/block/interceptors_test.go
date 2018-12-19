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

	_, err := NewHeaderInterceptor(nil, headers, headersNonces, mock.HasherMock{})
	assert.Equal(t, process.ErrNilInterceptor, err)
}

func TestNewHeaderInterceptor_NilHeadersShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	headersNonces := &mock.Uint64CacherStub{}

	_, err := NewHeaderInterceptor(interceptor, nil, headersNonces, mock.HasherMock{})
	assert.Equal(t, process.ErrNilHeadersDataPool, err)
}

func TestNewHeaderInterceptor_NilHeadersNoncesShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	headers := &mock.ShardedDataStub{}

	_, err := NewHeaderInterceptor(interceptor, headers, nil, mock.HasherMock{})
	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
}

func TestNewHeaderInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}

	_, err := NewHeaderInterceptor(interceptor, headers, headersNonces, nil)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewHeaderInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}

	hi, err := NewHeaderInterceptor(interceptor, headers, headersNonces, mock.HasherMock{})
	assert.Nil(t, err)
	assert.NotNil(t, hi)
}

//processHdr

func TestHeaderInterceptor_ProcessHdrNilHdrShouldRetFalse(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}

	hi, err := NewHeaderInterceptor(interceptor, headers, headersNonces, mock.HasherMock{})
	assert.Nil(t, err)
	assert.NotNil(t, hi)

	assert.False(t, hi.ProcessHdr(nil, make([]byte, 0)))
}

func TestHeaderInterceptor_ProcessHdrWrongTypeOfNewerShouldRetFalse(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}

	hi, err := NewHeaderInterceptor(interceptor, headers, headersNonces, mock.HasherMock{})
	assert.Nil(t, err)
	assert.NotNil(t, hi)

	assert.False(t, hi.ProcessHdr(&mock.StringNewer{}, make([]byte, 0)))
}

func TestHeaderInterceptor_ProcessHdrSanityCheckFailedShouldRetFalse(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}

	hi, err := NewHeaderInterceptor(interceptor, headers, headersNonces, mock.HasherMock{})
	assert.Nil(t, err)
	assert.NotNil(t, hi)

	assert.False(t, hi.ProcessHdr(NewInterceptedHeader(), make([]byte, 0)))
}

func TestHeaderInterceptor_ProcessOkValsShouldRetTrue(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	wasCalled := 0

	headers := &mock.ShardedDataStub{}
	headers.AddDataCalled = func(key []byte, data interface{}, destShardID uint32) {
		if bytes.Equal(mock.HasherMock{}.Compute("aaa"), key) {
			wasCalled++
		}
	}

	headersNonces := &mock.Uint64CacherStub{}
	headersNonces.HasOrAddCalled = func(u uint64, i []byte) (b bool, b2 bool) {
		if u == 67 {
			wasCalled++
		}

		return
	}

	hi, err := NewHeaderInterceptor(interceptor, headers, headersNonces, mock.HasherMock{})
	assert.Nil(t, err)
	assert.NotNil(t, hi)

	hdr := NewInterceptedHeader()
	hdr.Nonce = 67
	hdr.ShardId = 4
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))

	assert.True(t, hi.ProcessHdr(hdr, []byte("aaa")))
	assert.Equal(t, 2, wasCalled)
}

//------- BlockBodyInterceptor

//NewBlockBodyInterceptor

func TestNewBlockBodyInterceptor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}

	_, err := NewGenericBlockBodyInterceptor(nil, cache, mock.HasherMock{}, NewInterceptedTxBlockBody())
	assert.Equal(t, process.ErrNilInterceptor, err)
}

func TestNewBlockBodyInterceptor_NilPoolShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}

	_, err := NewGenericBlockBodyInterceptor(interceptor, nil, mock.HasherMock{}, NewInterceptedTxBlockBody())
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestNewBlockBodyInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}

	_, err := NewGenericBlockBodyInterceptor(interceptor, cache, nil, NewInterceptedTxBlockBody())
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewBlockBodyInterceptor_NilTemplateObjectShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}

	_, err := NewGenericBlockBodyInterceptor(interceptor, cache, mock.HasherMock{}, nil)
	assert.Equal(t, process.ErrNilTemplateObj, err)
}

func TestNewBlockBodyInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	gbbi, err := NewGenericBlockBodyInterceptor(interceptor, cache, mock.HasherMock{}, NewInterceptedTxBlockBody())
	assert.Nil(t, err)
	assert.NotNil(t, gbbi)
}

//processBodyBlock

func TestBlockBodyInterceptor_ProcessNilHdrShouldRetFalse(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	gbbi, err := NewGenericBlockBodyInterceptor(interceptor, cache, mock.HasherMock{}, NewInterceptedTxBlockBody())
	assert.Nil(t, err)
	assert.NotNil(t, gbbi)

	assert.False(t, gbbi.ProcessBodyBlock(nil, make([]byte, 0)))
}

func TestBlockBodyInterceptor_ProcessHdrWrongTypeOfNewerShouldRetFalse(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	gbbi, err := NewGenericBlockBodyInterceptor(interceptor, cache, mock.HasherMock{}, NewInterceptedTxBlockBody())
	assert.Nil(t, err)
	assert.NotNil(t, gbbi)

	assert.False(t, gbbi.ProcessBodyBlock(&mock.StringNewer{}, make([]byte, 0)))
}

func TestBlockBodyInterceptor_ProcessHdrSanityCheckFailedShouldRetFalse(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	gbbi, err := NewGenericBlockBodyInterceptor(interceptor, cache, mock.HasherMock{}, NewInterceptedTxBlockBody())
	assert.Nil(t, err)
	assert.NotNil(t, gbbi)

	assert.False(t, gbbi.ProcessBodyBlock(NewInterceptedHeader(), make([]byte, 0)))
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
	interceptor.SetCheckReceivedObjectHandlerCalled = func(i func(newer p2p.Newer, rawData []byte) bool) {
	}

	gbbi, err := NewGenericBlockBodyInterceptor(interceptor, cache, mock.HasherMock{}, NewInterceptedTxBlockBody())
	assert.Nil(t, err)
	assert.NotNil(t, gbbi)

	miniBlock := block2.MiniBlock{}
	miniBlock.TxHashes = append(miniBlock.TxHashes, []byte{65})

	txBody := NewInterceptedTxBlockBody()
	txBody.ShardId = 4
	txBody.MiniBlocks = make([]block2.MiniBlock, 0)
	txBody.MiniBlocks = append(txBody.MiniBlocks, miniBlock)

	assert.True(t, gbbi.ProcessBodyBlock(txBody, []byte("aaa")))
	assert.Equal(t, 1, wasCalled)
}
