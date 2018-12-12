package interceptorSuite

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
)

func TestInterceptorSuiteMakeDefaultInterceptorsCreatesTheNoOfInterceptors(t *testing.T) {
	mes := mock.NewMessengerStub()

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

	suite, err := NewInterceptorSuite(mes, mock.HasherMock{})
	assert.Nil(t, err)

	cfg := &storage.CacheConfig{
		Size: 10000,
		Type: storage.LRUCache,
	}

	err = suite.MakeDefaultInterceptors(
		shardedData.NewShardedData(cfg),
		shardedData.NewShardedData(cfg),
		shardedData.NewShardedData(cfg),
		shardedData.NewShardedData(cfg),
		shardedData.NewShardedData(cfg),
		&mock.AddressConverterMock{},
	)
	assert.Nil(t, err)

	assert.NotNil(t, suite.txInterceptor)
	assert.NotNil(t, suite.blockbodyInterceptors)
	assert.Equal(t, 3, len(suite.blockbodyInterceptors))

}
