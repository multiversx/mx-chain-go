package block

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/lrucache"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
)

var defaultTestConfig = storage.CacheConfig{
	Size: 1000,
	Type: storage.LRUCache,
}

func createNewHeaderInterceptorForTests() (
	hi *HeaderInterceptor,
	mes *mock.MessengerStub,
	headers data.ShardedDataCacherNotifier,
	headersNonces *dataPool.NonceToHashCacher,
	err error) {

	mes = mock.NewMessengerStub()

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		t.RegisterTopicValidator = func(v pubsub.Validator) error {
			return nil
		}
		t.UnregisterTopicValidator = func() error {
			return nil
		}

		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	headers, err = shardedData.NewShardedData(defaultTestConfig)
	if err != nil {
		return
	}

	headersNonces, err = dataPool.NewNonceToHashCacher(defaultTestConfig)
	if err != nil {
		return
	}

	hi, err = NewHeaderInterceptor(mes, headers, headersNonces, mock.HasherMock{})

	return
}

func createNewBlockBodyInterceptorForTests() (
	hi *GenericBlockBodyInterceptor,
	mes *mock.MessengerStub,
	cache storage.Cacher,
	err error) {

	mes = mock.NewMessengerStub()

	mes.AddTopicCalled = func(t *p2p.Topic) error {
		t.RegisterTopicValidator = func(v pubsub.Validator) error {
			return nil
		}
		t.UnregisterTopicValidator = func() error {
			return nil
		}

		return nil
	}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		return nil
	}

	cache, err = lrucache.NewCache(1000)
	if err != nil {
		return
	}

	hi, err = NewGenericBlockBodyInterceptor("test", mes, cache, mock.HasherMock{}, NewInterceptedTxBlockBody())
	return
}

//------- HeaderInterceptor

//NewHeaderInterceptor

func TestNewHeaderInterceptor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	headers, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

	headersNonces, err := dataPool.NewNonceToHashCacher(defaultTestConfig)
	assert.Nil(t, err)

	_, err = NewHeaderInterceptor(nil, headers, headersNonces, mock.HasherMock{})
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewHeaderInterceptor_NilHeadersShouldErr(t *testing.T) {
	t.Parallel()

	headersNonces, err := dataPool.NewNonceToHashCacher(defaultTestConfig)
	assert.Nil(t, err)

	mes := &mock.MessengerStub{}

	_, err = NewHeaderInterceptor(mes, nil, headersNonces, mock.HasherMock{})
	assert.Equal(t, process.ErrNilHeadersDataPool, err)
}

func TestNewHeaderInterceptor_NilHeadersNoncesShouldErr(t *testing.T) {
	t.Parallel()

	headers, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

	mes := &mock.MessengerStub{}

	_, err = NewHeaderInterceptor(mes, headers, nil, mock.HasherMock{})
	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
}

func TestNewHeaderInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	headers, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

	headersNonces, err := dataPool.NewNonceToHashCacher(defaultTestConfig)
	assert.Nil(t, err)

	_, err = NewHeaderInterceptor(mes, headers, headersNonces, nil)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewHeaderInterceptor_OkValsShouldWork(t *testing.T) {
	_, _, _, _, err := createNewHeaderInterceptorForTests()
	assert.Nil(t, err)
}

//processHdr

func TestTransactionInterceptor_ProcessHdrNilHdrShouldRetFalse(t *testing.T) {
	hi, _, _, _, _ := createNewHeaderInterceptorForTests()

	assert.False(t, hi.ProcessHdr(nil, make([]byte, 0)))
}

func TestTransactionInterceptor_ProcessHdrWrongTypeOfNewerShouldRetFalse(t *testing.T) {
	hi, _, _, _, _ := createNewHeaderInterceptorForTests()

	assert.False(t, hi.ProcessHdr(&mock.StringNewer{}, make([]byte, 0)))
}

func TestTransactionInterceptor_ProcessHdrSanityCheckFailedShouldRetFalse(t *testing.T) {
	hi, _, _, _, _ := createNewHeaderInterceptorForTests()

	assert.False(t, hi.ProcessHdr(NewInterceptedHeader(), make([]byte, 0)))
}

func TestTransactionInterceptor_ProcessOkValsShouldRetTrue(t *testing.T) {
	hi, _, headers, headersNonces, _ := createNewHeaderInterceptorForTests()

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

	shardStore := headers.ShardDataStore(4)
	if shardStore == nil {
		assert.Fail(t, "should have added header in minipool 4")
		return
	}

	assert.True(t, shardStore.Has(mock.HasherMock{}.Compute("aaa")))
	assert.True(t, headersNonces.Has(67))
}

//------- BlockBodyInterceptor

//NewBlockBodyInterceptor

func TestNewBlockBodyInterceptor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	cache, err := lrucache.NewCache(1000)
	assert.Nil(t, err)

	_, err = NewGenericBlockBodyInterceptor("test", nil, cache, mock.HasherMock{}, NewInterceptedTxBlockBody())
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewBlockBodyInterceptor_NilPoolShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	_, err := NewGenericBlockBodyInterceptor("test", mes, nil, mock.HasherMock{}, NewInterceptedTxBlockBody())
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestNewBlockBodyInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}
	cache, err := lrucache.NewCache(1000)
	assert.Nil(t, err)

	_, err = NewGenericBlockBodyInterceptor("test", mes, cache, nil, NewInterceptedTxBlockBody())
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewBlockBodyInterceptor_NilTemplateObjectShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}
	cache, err := lrucache.NewCache(1000)
	assert.Nil(t, err)

	_, err = NewGenericBlockBodyInterceptor("test", mes, cache, mock.HasherMock{}, nil)
	assert.Equal(t, process.ErrNilTemplateObj, err)
}

func TestNewBlockBodyInterceptor_OkValsShouldWork(t *testing.T) {
	_, _, _, err := createNewBlockBodyInterceptorForTests()
	assert.Nil(t, err)
}

//processBodyBlock

func TestBlockBodyInterceptor_ProcessNilHdrShouldRetFalse(t *testing.T) {
	hi, _, _, _ := createNewBlockBodyInterceptorForTests()

	assert.False(t, hi.ProcessBodyBlock(nil, make([]byte, 0)))
}

func TestBlockBodyInterceptor_ProcessHdrWrongTypeOfNewerShouldRetFalse(t *testing.T) {
	hi, _, _, _ := createNewBlockBodyInterceptorForTests()

	assert.False(t, hi.ProcessBodyBlock(&mock.StringNewer{}, make([]byte, 0)))
}

func TestBlockBodyInterceptor_ProcessHdrSanityCheckFailedShouldRetFalse(t *testing.T) {
	hi, _, _, _ := createNewBlockBodyInterceptorForTests()

	assert.False(t, hi.ProcessBodyBlock(NewInterceptedHeader(), make([]byte, 0)))
}

func TestBlockBodyInterceptor_ProcessOkValsShouldRetTrue(t *testing.T) {
	hi, _, cache, _ := createNewBlockBodyInterceptorForTests()

	miniBlock := block2.MiniBlock{}
	miniBlock.TxHashes = append(miniBlock.TxHashes, []byte{65})

	txBody := NewInterceptedTxBlockBody()
	txBody.ShardId = 4
	txBody.MiniBlocks = make([]block2.MiniBlock, 0)
	txBody.MiniBlocks = append(txBody.MiniBlocks, miniBlock)

	assert.True(t, hi.ProcessBodyBlock(txBody, []byte("aaa")))

	assert.True(t, cache.Has(mock.HasherMock{}.Compute("aaa")))
}
