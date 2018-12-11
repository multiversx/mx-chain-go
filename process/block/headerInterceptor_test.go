package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createNewHeaderInterceptorForTests() (
	hi *HeaderInterceptor,
	mes *mock.MessengerStub,
	headerPool *dataPool.DataPool,
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

	headerPool = dataPool.NewDataPool(&storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	})

	hi, err = NewHeaderInterceptor(mes, headerPool, mock.HasherMock{})

	return
}

//------- NewHeaderInterceptor

func TestNewHeaderInterceptorNilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	headerPool := dataPool.NewDataPool(&storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	})

	_, err := NewHeaderInterceptor(nil, headerPool, mock.HasherMock{})
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewHeaderInterceptorNilHeaderPoolShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	_, err := NewHeaderInterceptor(mes, nil, mock.HasherMock{})
	assert.Equal(t, process.ErrNilHeaderDataPool, err)
}

func TestNewHeaderInterceptorNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}
	headerPool := dataPool.NewDataPool(&storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	})

	_, err := NewHeaderInterceptor(mes, headerPool, nil)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewHeaderInterceptorOkValsShouldWork(t *testing.T) {
	_, _, _, err := createNewHeaderInterceptorForTests()
	assert.Nil(t, err)
}

//------- processHdr

func TestTransactionInterceptorProcessHdrNilHdrShouldRetFalse(t *testing.T) {
	hi, _, _, _ := createNewHeaderInterceptorForTests()

	assert.False(t, hi.ProcessHdr(nil, make([]byte, 0)))
}

func TestTransactionInterceptorProcessHdrWrongTypeOfNewerShouldRetFalse(t *testing.T) {
	hi, _, _, _ := createNewHeaderInterceptorForTests()

	assert.False(t, hi.ProcessHdr(&mock.StringNewer{}, make([]byte, 0)))
}

func TestTransactionInterceptorProcessHdrSanityCheckFailedShouldRetFalse(t *testing.T) {
	hi, _, _, _ := createNewHeaderInterceptorForTests()

	assert.False(t, hi.ProcessHdr(NewInterceptedHeader(), make([]byte, 0)))
}

func TestTransactionInterceptorProcessOkValsShouldRetTrue(t *testing.T) {
	hi, _, headerPool, _ := createNewHeaderInterceptorForTests()

	hdr := NewInterceptedHeader()
	hdr.ShardId = 4
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))

	assert.True(t, hi.ProcessHdr(hdr, []byte("aaa")))

	minipool := headerPool.MiniPoolDataStore(4)
	if minipool == nil {
		assert.Fail(t, "should have added header in minipool 4")
		return
	}

	assert.True(t, minipool.Has(mock.HasherMock{}.Compute("aaa")))
}
