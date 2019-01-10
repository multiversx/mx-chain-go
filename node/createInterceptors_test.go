package node

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
)

func createRequiredObjects(marshalizer marshal.Marshalizer) (*p2p.Topic, *mock.TransientDataPoolMock) {
	topic := p2p.NewTopic("", &mock.StringCreatorMock{}, marshalizer)
	topic.RegisterTopicValidator = func(v pubsub.Validator) error {
		return nil
	}

	dataPool := &mock.TransientDataPoolMock{}

	return topic, dataPool
}

//------- createTxInterceptor

func TestCreateTxInterceptor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	err := node.createTxInterceptor()

	assert.Equal(t, "nil Messenger", err.Error())
	assert.Equal(t, 0, len(node.interceptors))
}

func TestCreateTxInterceptor_IncompleteSettingsShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == string(TransactionTopic) {
			return topic
		}

		return nil
	}

	dataPool.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}

	node.dataPool = dataPool
	node.addrConverter = mock.AddressConverterStub{}
	//nil hasher
	node.singleSignKeyGen = &mock.SingleSignKeyGenMock{}
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createTxInterceptor()

	assert.Equal(t, "nil Hasher", err.Error())
	assert.Equal(t, 0, len(node.interceptors))
}

func TestCreateTxInterceptor_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == string(TransactionTopic) {
			return topic
		}

		return nil
	}

	dataPool.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}

	node.dataPool = dataPool
	node.addrConverter = mock.AddressConverterStub{}
	node.hasher = mock.HasherMock{}
	node.singleSignKeyGen = &mock.SingleSignKeyGenMock{}
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createTxInterceptor()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(node.interceptors))
}

//------- createHdrInterceptor

func TestCreateHdrInterceptor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	err := node.createHdrInterceptor()

	assert.Equal(t, "nil Messenger", err.Error())
	assert.Equal(t, 0, len(node.interceptors))
}

func TestCreateHdrInterceptor_IncompleteSettingsShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == string(HeadersTopic) {
			return topic
		}

		return nil
	}

	dataPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	node.dataPool = dataPool
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createHdrInterceptor()

	assert.Equal(t, "nil Hasher", err.Error())
	assert.Equal(t, 0, len(node.interceptors))
}

func TestCreateHdrInterceptor_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == string(HeadersTopic) {
			return topic
		}

		return nil
	}

	dataPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	node.dataPool = dataPool
	node.hasher = mock.HasherMock{}
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createHdrInterceptor()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(node.interceptors))
}

//------- createTxBlockBodyInterceptor

func TestCreateTxBlockBodyInterceptor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	err := node.createTxBlockBodyInterceptor()

	assert.Equal(t, "nil Messenger", err.Error())
	assert.Equal(t, 0, len(node.interceptors))
}

func TestCreateTxBlockBodyInterceptor_IncompleteSettingsShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == string(TxBlockBodyTopic) {
			return topic
		}

		return nil
	}

	dataPool.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}

	node.dataPool = dataPool
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createTxBlockBodyInterceptor()
	assert.Equal(t, "nil Hasher", err.Error())
	assert.Equal(t, 0, len(node.interceptors))
}

func TestCreateTxBlockBodyInterceptor_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == string(TxBlockBodyTopic) {
			return topic
		}

		return nil
	}

	dataPool.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}

	node.dataPool = dataPool
	node.hasher = mock.HasherMock{}
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createTxBlockBodyInterceptor()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(node.interceptors))
}

//------- createPeerChBlockBodyInterceptor

func TestCreatePeerChBlockBodyInterceptor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	err := node.createPeerChBlockBodyInterceptor()

	assert.Equal(t, "nil Messenger", err.Error())
	assert.Equal(t, 0, len(node.interceptors))
}

func TestCreatePeerChBlockBodyInterceptor_IncompleteSettingsShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == string(PeerChBodyTopic) {
			return topic
		}

		return nil
	}

	dataPool.PeerChangesBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}

	node.dataPool = dataPool
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createPeerChBlockBodyInterceptor()
	assert.Equal(t, "nil Hasher", err.Error())
	assert.Equal(t, 0, len(node.interceptors))
}

func TestCreatePeerChBlockBodyInterceptor_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == string(PeerChBodyTopic) {
			return topic
		}

		return nil
	}

	dataPool.PeerChangesBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}

	node.dataPool = dataPool
	node.hasher = mock.HasherMock{}
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createPeerChBlockBodyInterceptor()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(node.interceptors))
}

//------- createStateBlockBodyInterceptor

func TestCreateStateBlockBodyInterceptor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	err := node.createStateBlockBodyInterceptor()

	assert.Equal(t, "nil Messenger", err.Error())
	assert.Equal(t, 0, len(node.interceptors))
}

func TestCreateStateBlockBodyInterceptor_IncompleteSettingsShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == string(StateBodyTopic) {
			return topic
		}

		return nil
	}

	dataPool.StateBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}

	node.dataPool = dataPool
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createStateBlockBodyInterceptor()
	assert.Equal(t, "nil Hasher", err.Error())
	assert.Equal(t, 0, len(node.interceptors))
}

func TestCreateStateBlockBodyInterceptor_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		if name == string(StateBodyTopic) {
			return topic
		}

		return nil
	}

	dataPool.StateBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}

	node.dataPool = dataPool
	node.hasher = mock.HasherMock{}
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createStateBlockBodyInterceptor()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(node.interceptors))
}

//------- createInterceptors

func TestCreateInterceptors_NilTransactionsShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	genericTopic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		switch name {
		case string(TransactionTopic):
			return genericTopic
		case string(HeadersTopic):
			return genericTopic
		case string(TxBlockBodyTopic):
			return genericTopic
		case string(PeerChBodyTopic):
			return genericTopic
		case string(StateBodyTopic):
			return genericTopic
		}

		return nil
	}

	dataPool.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return nil
	}
	dataPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	dataPool.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	dataPool.PeerChangesBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	dataPool.StateBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}

	node.dataPool = dataPool
	node.addrConverter = mock.AddressConverterStub{}
	node.hasher = mock.HasherMock{}
	node.singleSignKeyGen = &mock.SingleSignKeyGenMock{}
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createInterceptors()

	assert.Equal(t, "nil transaction data pool", err.Error())
}

func TestCreateInterceptors_NilHeadersShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	genericTopic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		switch name {
		case string(TransactionTopic):
			return genericTopic
		case string(HeadersTopic):
			return genericTopic
		case string(TxBlockBodyTopic):
			return genericTopic
		case string(PeerChBodyTopic):
			return genericTopic
		case string(StateBodyTopic):
			return genericTopic
		}

		return nil
	}

	dataPool.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return nil
	}
	dataPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	dataPool.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	dataPool.PeerChangesBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	dataPool.StateBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}

	node.dataPool = dataPool
	node.addrConverter = mock.AddressConverterStub{}
	node.hasher = mock.HasherMock{}
	node.singleSignKeyGen = &mock.SingleSignKeyGenMock{}
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createInterceptors()

	assert.Equal(t, "nil headers data pool", err.Error())
}

func TestCreateInterceptors_NilHeadersNoncesShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	genericTopic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		switch name {
		case string(TransactionTopic):
			return genericTopic
		case string(HeadersTopic):
			return genericTopic
		case string(TxBlockBodyTopic):
			return genericTopic
		case string(PeerChBodyTopic):
			return genericTopic
		case string(StateBodyTopic):
			return genericTopic
		}

		return nil
	}

	dataPool.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return nil
	}
	dataPool.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	dataPool.PeerChangesBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	dataPool.StateBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}

	node.dataPool = dataPool
	node.addrConverter = mock.AddressConverterStub{}
	node.hasher = mock.HasherMock{}
	node.singleSignKeyGen = &mock.SingleSignKeyGenMock{}
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createInterceptors()

	assert.Equal(t, "nil headers nonces cache", err.Error())
}

func TestCreateInterceptors_NilTxBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	genericTopic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		switch name {
		case string(TransactionTopic):
			return genericTopic
		case string(HeadersTopic):
			return genericTopic
		case string(TxBlockBodyTopic):
			return genericTopic
		case string(PeerChBodyTopic):
			return genericTopic
		case string(StateBodyTopic):
			return genericTopic
		}

		return nil
	}

	dataPool.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	dataPool.TxBlocksCalled = func() storage.Cacher {
		return nil
	}
	dataPool.PeerChangesBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	dataPool.StateBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}

	node.dataPool = dataPool
	node.addrConverter = mock.AddressConverterStub{}
	node.hasher = mock.HasherMock{}
	node.singleSignKeyGen = &mock.SingleSignKeyGenMock{}
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createInterceptors()

	assert.Equal(t, "nil cacher", err.Error())
}

func TestCreateInterceptors_NilPeerBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	genericTopic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		switch name {
		case string(TransactionTopic):
			return genericTopic
		case string(HeadersTopic):
			return genericTopic
		case string(TxBlockBodyTopic):
			return genericTopic
		case string(PeerChBodyTopic):
			return genericTopic
		case string(StateBodyTopic):
			return genericTopic
		}

		return nil
	}

	dataPool.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	dataPool.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	dataPool.PeerChangesBlocksCalled = func() storage.Cacher {
		return nil
	}
	dataPool.StateBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}

	node.dataPool = dataPool
	node.addrConverter = mock.AddressConverterStub{}
	node.hasher = mock.HasherMock{}
	node.singleSignKeyGen = &mock.SingleSignKeyGenMock{}
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createInterceptors()

	assert.Equal(t, "nil cacher", err.Error())
}

func TestCreateInterceptors_NilStateBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	genericTopic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		switch name {
		case string(TransactionTopic):
			return genericTopic
		case string(HeadersTopic):
			return genericTopic
		case string(TxBlockBodyTopic):
			return genericTopic
		case string(PeerChBodyTopic):
			return genericTopic
		case string(StateBodyTopic):
			return genericTopic
		}

		return nil
	}

	dataPool.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	dataPool.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	dataPool.PeerChangesBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	dataPool.StateBlocksCalled = func() storage.Cacher {
		return nil
	}

	node.dataPool = dataPool
	node.addrConverter = mock.AddressConverterStub{}
	node.hasher = mock.HasherMock{}
	node.singleSignKeyGen = &mock.SingleSignKeyGenMock{}
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createInterceptors()

	assert.Equal(t, "nil cacher", err.Error())
}

func TestCreateInterceptors_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	genericTopic, dataPool := createRequiredObjects(mes.Marshalizer())

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		switch name {
		case string(TransactionTopic):
			return genericTopic
		case string(HeadersTopic):
			return genericTopic
		case string(TxBlockBodyTopic):
			return genericTopic
		case string(PeerChBodyTopic):
			return genericTopic
		case string(StateBodyTopic):
			return genericTopic
		}

		return nil
	}

	dataPool.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dataPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	dataPool.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	dataPool.PeerChangesBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	dataPool.StateBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}

	node.dataPool = dataPool
	node.addrConverter = mock.AddressConverterStub{}
	node.hasher = mock.HasherMock{}
	node.singleSignKeyGen = &mock.SingleSignKeyGenMock{}
	node.shardCoordinator = mock.NewOneShardCoordinatorMock()

	err := node.createInterceptors()

	assert.Nil(t, err)
	assert.Equal(t, 5, len(node.interceptors))
}
