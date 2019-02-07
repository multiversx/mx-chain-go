package block

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/pkg/errors"

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
	storer := &mock.StorerStub{}

	hi, err := NewHeaderInterceptor(
		nil,
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilInterceptor, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilHeadersShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, err := NewHeaderInterceptor(
		interceptor,
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

	interceptor := &mock.InterceptorStub{}
	headers := &mock.ShardedDataStub{}
	storer := &mock.StorerStub{}

	hi, err := NewHeaderInterceptor(
		interceptor,
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

	interceptor := &mock.InterceptorStub{}
	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}

	hi, err := NewHeaderInterceptor(
		interceptor,
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

	interceptor := &mock.InterceptorStub{}
	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, err := NewHeaderInterceptor(
		interceptor,
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

	interceptor := &mock.InterceptorStub{}
	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, err := NewHeaderInterceptor(
		interceptor,
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

	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
	}

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, err := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Nil(t, err)
	assert.NotNil(t, hi)
}

//processHdr

func TestHeaderInterceptor_ProcessHdrNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
	}

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, _ := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMessage, hi.ProcessHdr(nil))
}

func TestHeaderInterceptor_ProcessHdrNilDataToProcessShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
	}

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, _ := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{}

	assert.Equal(t, process.ErrNilDataToProcess, hi.ProcessHdr(msg))
}

func TestHeaderInterceptor_ProcessHdrNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return nil
		},
	}

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, _ := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, process.ErrNilMarshalizer, hi.ProcessHdr(msg))
}

func TestHeaderInterceptor_ProcessHdrMarshalizerErrorsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")

	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return &mock.MarshalizerStub{
				UnmarshalCalled: func(obj interface{}, buff []byte) error {
					return errMarshalizer
				},
			}
		},
	}

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, _ := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, errMarshalizer, hi.ProcessHdr(msg))
}

func TestHeaderInterceptor_ProcessHdrSanityCheckFailedShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return marshalizer
		},
	}

	headers := &mock.ShardedDataStub{}
	headersNonces := &mock.Uint64CacherStub{}
	storer := &mock.StorerStub{}

	hi, _ := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	hdr := NewInterceptedHeader()
	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	assert.Equal(t, process.ErrNilBlockBodyHash, hi.ProcessHdr(msg))
}

func TestHeaderInterceptor_ProcessOkValsShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return marshalizer
		},
	}

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

	hi, _ := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		storer,
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

	assert.Nil(t, hi.ProcessHdr(msg))
	assert.Equal(t, 2, wasCalled)
}

func TestHeaderInterceptor_ProcessIsInStorageShouldNotAdd(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return marshalizer
		},
	}

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

	hi, _ := NewHeaderInterceptor(
		interceptor,
		headers,
		headersNonces,
		storer,
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

	assert.Nil(t, hi.ProcessHdr(msg))
	assert.Equal(t, 0, wasCalled)
}

//------- BlockBodyInterceptor

//NewBlockBodyInterceptor

func TestNewBlockBodyInterceptor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	gbbi, err := NewGenericBlockBodyInterceptor(
		nil,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilInterceptor, err)
	assert.Nil(t, gbbi)
}

func TestNewBlockBodyInterceptor_NilPoolShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	storer := &mock.StorerStub{}

	gbbi, err := NewGenericBlockBodyInterceptor(
		interceptor,
		nil,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilCacher, err)
	assert.Nil(t, gbbi)
}

func TestNewBlockBodyInterceptor_NilStorerShouldErr(t *testing.T) {
	t.Parallel()

	interceptor := &mock.InterceptorStub{}
	cache := &mock.CacherStub{}

	gbbi, err := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		nil,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilBlockBodyStorage, err)
	assert.Nil(t, gbbi)
}

func TestNewBlockBodyInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}
	storer := &mock.StorerStub{}

	gbbi, err := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		nil,
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, gbbi)
}

func TestNewBlockBodyInterceptor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{}
	storer := &mock.StorerStub{}

	gbbi, err := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		nil)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, gbbi)
}

func TestNewBlockBodyInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
	}
	storer := &mock.StorerStub{}

	gbbi, err := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Nil(t, err)
	assert.NotNil(t, gbbi)
}

//ProcessTxBodyBlock

func TestBlockBodyInterceptor_ProcessTxBodyBlockNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
	}
	storer := &mock.StorerStub{}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMessage, gbbi.ProcessTxBodyBlock(nil))
}

func TestBlockBodyInterceptor_ProcessTxBodyBlockNilMessageDataShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
	}
	storer := &mock.StorerStub{}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{}

	assert.Equal(t, process.ErrNilDataToProcess, gbbi.ProcessTxBodyBlock(msg))
}

func TestBlockBodyInterceptor_ProcessTxBodyBlockNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return nil
		},
	}
	storer := &mock.StorerStub{}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, process.ErrNilMarshalizer, gbbi.ProcessTxBodyBlock(msg))
}

func TestBlockBodyInterceptor_ProcessTxBodyBlockMarshalizerErrorsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return &mock.MarshalizerStub{
				UnmarshalCalled: func(obj interface{}, buff []byte) error {
					return errMarshalizer
				},
			}
		},
	}
	storer := &mock.StorerStub{}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, errMarshalizer, gbbi.ProcessTxBodyBlock(msg))
}

func TestBlockBodyInterceptor_ProcessTxBodyBlockIntegrityFailsShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return marshalizer
		},
	}
	storer := &mock.StorerStub{}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	txBlock := NewInterceptedTxBlockBody()
	txBlock.RootHash = []byte("root hash")
	txBlock.ShardID = uint32(0)
	txBlock.MiniBlocks = nil

	buff, _ := marshalizer.Marshal(txBlock)

	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	assert.Equal(t, process.ErrNilMiniBlocks, gbbi.ProcessTxBodyBlock(msg))
}

func TestBlockBodyInterceptor_ProcessTxBodyBlockShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return marshalizer
		},
	}
	storer := &mock.StorerStub{
		HasCalled: func(key []byte) (b bool, e error) {
			return false, nil
		},
	}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	txBlock := NewInterceptedTxBlockBody()
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

	assert.Nil(t, gbbi.ProcessTxBodyBlock(msg))
	assert.True(t, putInCacheWasCalled)
}

//ProcessStateBodyBlock

func TestBlockBodyInterceptor_ProcessStateBodyBlockNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
	}
	storer := &mock.StorerStub{}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMessage, gbbi.ProcessStateBodyBlock(nil))
}

func TestBlockBodyInterceptor_ProcessStateBodyBlockNilMessageDataShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
	}
	storer := &mock.StorerStub{}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{}

	assert.Equal(t, process.ErrNilDataToProcess, gbbi.ProcessStateBodyBlock(msg))
}

func TestBlockBodyInterceptor_ProcessStateBodyBlockNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return nil
		},
	}
	storer := &mock.StorerStub{}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, process.ErrNilMarshalizer, gbbi.ProcessStateBodyBlock(msg))
}

func TestBlockBodyInterceptor_ProcessStateBodyBlockMarshalizerErrorsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return &mock.MarshalizerStub{
				UnmarshalCalled: func(obj interface{}, buff []byte) error {
					return errMarshalizer
				},
			}
		},
	}
	storer := &mock.StorerStub{}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, errMarshalizer, gbbi.ProcessStateBodyBlock(msg))
}

func TestBlockBodyInterceptor_ProcessStateBodyBlockIntegrityFailsShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return marshalizer
		},
	}
	storer := &mock.StorerStub{}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	stateBlock := NewInterceptedStateBlockBody()
	stateBlock.ShardID = uint32(0)
	stateBlock.RootHash = nil

	buff, _ := marshalizer.Marshal(stateBlock)

	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	assert.Equal(t, process.ErrNilRootHash, gbbi.ProcessStateBodyBlock(msg))
}

//ProcessPeerChangeBodyBlock

func TestBlockBodyInterceptor_ProcessPeerChangeBodyBlockNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
	}
	storer := &mock.StorerStub{}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMessage, gbbi.ProcessPeerChangeBodyBlock(nil))
}

func TestBlockBodyInterceptor_ProcessPeerChangeBodyBlockNilMessageDataShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
	}
	storer := &mock.StorerStub{}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{}

	assert.Equal(t, process.ErrNilDataToProcess, gbbi.ProcessPeerChangeBodyBlock(msg))
}

func TestBlockBodyInterceptor_ProcessPeerChangeBodyBlockNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return nil
		},
	}
	storer := &mock.StorerStub{}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, process.ErrNilMarshalizer, gbbi.ProcessPeerChangeBodyBlock(msg))
}

func TestBlockBodyInterceptor_ProcessPeerChangeBodyBlockMarshalizerErrorsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return &mock.MarshalizerStub{
				UnmarshalCalled: func(obj interface{}, buff []byte) error {
					return errMarshalizer
				},
			}
		},
	}
	storer := &mock.StorerStub{}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, errMarshalizer, gbbi.ProcessPeerChangeBodyBlock(msg))
}

func TestBlockBodyInterceptor_ProcessPeerChangeBodyBlockIntegrityFailsShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	cache := &mock.CacherStub{}
	interceptor := &mock.InterceptorStub{
		SetReceivedMessageHandlerCalled: func(i func(message p2p.MessageP2P) error) {},
		MarshalizerCalled: func() marshal.Marshalizer {
			return marshalizer
		},
	}
	storer := &mock.StorerStub{}

	gbbi, _ := NewGenericBlockBodyInterceptor(
		interceptor,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	peerChangeBlock := NewInterceptedPeerBlockBody()
	peerChangeBlock.ShardID = uint32(0)
	peerChangeBlock.RootHash = []byte("root hash")
	peerChangeBlock.Changes = nil

	buff, _ := marshalizer.Marshal(peerChangeBlock)

	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	assert.Equal(t, process.ErrNilPeerChanges, gbbi.ProcessPeerChangeBodyBlock(msg))
}
