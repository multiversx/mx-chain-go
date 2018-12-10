package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/config"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createNewHeaderInterceptorForTests() (
	hi *headerInterceptor,
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

	headerPool = dataPool.NewDataPool(&config.CacheConfig{
		Size: 10000,
		Type: config.LRUCache,
	})

	hi, err = NewHeaderInterceptor(mes, headerPool)

	return
}

//------- NewHeaderInterceptor

func TestNewHeaderInterceptorNilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	headerPool := dataPool.NewDataPool(&config.CacheConfig{
		Size: 10000,
		Type: config.LRUCache,
	})

	_, err := NewHeaderInterceptor(nil, headerPool)
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewHeaderInterceptorNilHeaderPoolShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	_, err := NewHeaderInterceptor(mes, nil)
	assert.Equal(t, process.ErrNilHeaderDataPool, err)
}

func TestNewHeaderInterceptorOkValsShouldWork(t *testing.T) {
	_, _, _, err := createNewHeaderInterceptorForTests()
	assert.Nil(t, err)
}

//------- processHdr

func TestTransactionInterceptorProcessHdrNilHdrShouldRetFalse(t *testing.T) {
	hi, _, _, _ := createNewHeaderInterceptorForTests()

	assert.False(t, hi.ProcessHdr(nil, make([]byte, 0), mock.HasherMock{}))
}

func TestTransactionInterceptorProcessHdrNilHasherShouldRetFalse(t *testing.T) {
	hi, _, _, _ := createNewHeaderInterceptorForTests()

	assert.False(t, hi.ProcessHdr(NewInterceptedHeader(), make([]byte, 0), nil))
}

func TestTransactionInterceptorProcessHdrWrongTypeOfNewerShouldRetFalse(t *testing.T) {
	hi, _, _, _ := createNewHeaderInterceptorForTests()

	assert.False(t, hi.ProcessHdr(&mock.StringNewer{}, make([]byte, 0), mock.HasherMock{}))
}

func TestTransactionInterceptorProcessHdrSanityCheckFailedShouldRetFalse(t *testing.T) {
	hi, _, _, _ := createNewHeaderInterceptorForTests()

	assert.False(t, hi.ProcessHdr(NewInterceptedHeader(), make([]byte, 0), mock.HasherMock{}))
}

func TestTransactionInterceptorProcessAddressedToOtherShardsShouldRetTrue(t *testing.T) {
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

	assert.True(t, hi.ProcessHdr(hdr, []byte("aaa"), mock.HasherMock{}))

	minipool := headerPool.MiniPoolDataStore(4)
	if minipool == nil {
		assert.Fail(t, "should have added header in minipool 4")
		return
	}

	assert.True(t, minipool.Has(mock.HasherMock{}.Compute("aaa")))
}
