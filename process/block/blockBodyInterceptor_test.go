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

func createNewBlockBodyInterceptorForTests() (
	hi *GenericBlockBodyInterceptor,
	mes *mock.MessengerStub,
	pool *dataPool.DataPool,
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

	pool = dataPool.NewDataPool(&storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	})

	hi, err = NewGenericBlockBodyInterceptor("test", mes, pool, mock.HasherMock{}, NewInterceptedTxBlockBody())

	return
}

//------- NewBlockBodyInterceptor

func TestNewBlockBodyInterceptorNilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	blockBodyPool := dataPool.NewDataPool(&storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	})

	_, err := NewGenericBlockBodyInterceptor("test", nil, blockBodyPool, mock.HasherMock{}, NewInterceptedTxBlockBody())
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewBlockBodyInterceptorNilPoolShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}

	_, err := NewGenericBlockBodyInterceptor("test", mes, nil, mock.HasherMock{}, NewInterceptedTxBlockBody())
	assert.Equal(t, process.ErrNilBodyBlockDataPool, err)
}

func TestNewBlockBodyInterceptorNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}
	blockBodyPool := dataPool.NewDataPool(&storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	})

	_, err := NewGenericBlockBodyInterceptor("test", mes, blockBodyPool, nil, NewInterceptedTxBlockBody())
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewBlockBodyInterceptorNilTemplateObjectShouldErr(t *testing.T) {
	t.Parallel()

	mes := &mock.MessengerStub{}
	blockBodyPool := dataPool.NewDataPool(&storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	})

	_, err := NewGenericBlockBodyInterceptor("test", mes, blockBodyPool, mock.HasherMock{}, nil)
	assert.Equal(t, process.ErrNilTemplateObj, err)
}

func TestNewBlockBodyInterceptorOkValsShouldWork(t *testing.T) {
	_, _, _, err := createNewBlockBodyInterceptorForTests()
	assert.Nil(t, err)
}

//------- processBodyBlock

func TestBlockBodyInterceptorProcessNilHdrShouldRetFalse(t *testing.T) {
	hi, _, _, _ := createNewBlockBodyInterceptorForTests()

	assert.False(t, hi.ProcessBodyBlock(nil, make([]byte, 0)))
}

func TestBlockBodyInterceptorProcessHdrWrongTypeOfNewerShouldRetFalse(t *testing.T) {
	hi, _, _, _ := createNewBlockBodyInterceptorForTests()

	assert.False(t, hi.ProcessBodyBlock(&mock.StringNewer{}, make([]byte, 0)))
}

func TestBlockBodyInterceptorProcessHdrSanityCheckFailedShouldRetFalse(t *testing.T) {
	hi, _, _, _ := createNewBlockBodyInterceptorForTests()

	assert.False(t, hi.ProcessBodyBlock(NewInterceptedHeader(), make([]byte, 0)))
}

func TestBlockBodyInterceptorProcessOkValsShouldRetTrue(t *testing.T) {
	hi, _, headerPool, _ := createNewBlockBodyInterceptorForTests()

	miniBlock := block.MiniBlock{}
	miniBlock.TxHashes = append(miniBlock.TxHashes, []byte{65})

	txBody := NewInterceptedTxBlockBody()
	txBody.ShardId = 4
	txBody.MiniBlocks = make([]block.MiniBlock, 0)
	txBody.MiniBlocks = append(txBody.MiniBlocks, miniBlock)

	assert.True(t, hi.ProcessBodyBlock(txBody, []byte("aaa")))

	minipool := headerPool.MiniPoolDataStore(4)
	if minipool == nil {
		assert.Fail(t, "should have added tx body in minipool 4")
		return
	}

	assert.True(t, minipool.Has(mock.HasherMock{}.Compute("aaa")))
}
