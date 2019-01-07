package node

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/stretchr/testify/assert"
)

func createBlockchain() *blockchain.BlockChain {
	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{})

	return blkc
}

//------- createTxResolver

func TestCreateTxResolver_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	err := node.createTxResolver()

	assert.Equal(t, "nil Messenger", err.Error())
	assert.Equal(t, 0, len(node.resolvers))
}

func TestCreateTxResolver_NilHeadersDataPoolShouldErr(t *testing.T) {
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
		return nil
	}

	node.dataPool = dataPool
	node.blkc = createBlockchain()
	node.marshalizer = mock.MarshalizerMock{}

	err := node.createTxResolver()

	assert.Equal(t, "nil transaction data pool", err.Error())
	assert.Equal(t, 0, len(node.resolvers))
}

func TestCreateTxResolver_ShouldWork(t *testing.T) {
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
	node.marshalizer = mock.MarshalizerMock{}
	node.blkc = createBlockchain()

	err := node.createTxResolver()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(node.resolvers))
}

//------- createHdrResolver

func TestCreateHdrResolver_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	err := node.createHdrResolver()

	assert.Equal(t, "nil Messenger", err.Error())
	assert.Equal(t, 0, len(node.resolvers))
}

func TestCreateHdrResolver_NilTransactionDataPoolShouldErr(t *testing.T) {
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
		return nil
	}
	dataPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}

	node.dataPool = dataPool
	node.blkc = createBlockchain()
	node.marshalizer = mock.MarshalizerMock{}
	node.uint64ByteSliceConverter = mock.NewNonceHashConverterMock()

	err := node.createHdrResolver()

	assert.Equal(t, "nil headers data pool", err.Error())
	assert.Equal(t, 0, len(node.resolvers))
}

func TestCreateHdrResolver_ShouldWork(t *testing.T) {
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
	node.blkc = createBlockchain()
	node.marshalizer = mock.MarshalizerMock{}
	node.uint64ByteSliceConverter = mock.NewNonceHashConverterMock()

	err := node.createHdrResolver()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(node.resolvers))
}

//------- createTxBlockBodyResolver

func TestCreateTxBlockBodyResolver_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	err := node.createTxBlockBodyResolver()

	assert.Equal(t, "nil Messenger", err.Error())
	assert.Equal(t, 0, len(node.resolvers))
}

func TestCreateTxBlockBodyResolver_NilDataPoolCacherDataPoolShouldErr(t *testing.T) {
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
		return nil
	}

	node.dataPool = dataPool
	node.blkc = createBlockchain()
	node.marshalizer = mock.MarshalizerMock{}

	err := node.createTxBlockBodyResolver()

	assert.Equal(t, "nil block body pool", err.Error())
	assert.Equal(t, 0, len(node.resolvers))
}

func TestCreateTxBlockBodyResolver_ShouldWork(t *testing.T) {
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
	node.blkc = createBlockchain()
	node.marshalizer = mock.MarshalizerMock{}

	err := node.createTxBlockBodyResolver()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(node.resolvers))
}

//------- createPeerChBlockBodyResolver

func TestCreatePeerChBlockBodyResolver_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	err := node.createPeerChBlockBodyResolver()

	assert.Equal(t, "nil Messenger", err.Error())
	assert.Equal(t, 0, len(node.resolvers))
}

func TestCreatePeerChBlockBodyResolver_NilDataPoolCacherDataPoolShouldErr(t *testing.T) {
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
		return nil
	}

	node.dataPool = dataPool
	node.blkc = createBlockchain()
	node.marshalizer = mock.MarshalizerMock{}

	err := node.createPeerChBlockBodyResolver()

	assert.Equal(t, "nil block body pool", err.Error())
	assert.Equal(t, 0, len(node.resolvers))
}

func TestCreatePeerChBlockBodyResolver_ShouldWork(t *testing.T) {
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
	node.blkc = createBlockchain()
	node.marshalizer = mock.MarshalizerMock{}

	err := node.createPeerChBlockBodyResolver()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(node.resolvers))
}

//------- createStateBlockBodyResolver

func TestCreateStateBlockBodyResolver_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	err := node.createStateBlockBodyResolver()

	assert.Equal(t, "nil Messenger", err.Error())
	assert.Equal(t, 0, len(node.resolvers))
}

func TestCreateStateBlockBodyResolver_NilDataPoolCacherDataPoolShouldErr(t *testing.T) {
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
		return nil
	}

	node.dataPool = dataPool
	node.blkc = createBlockchain()
	node.marshalizer = mock.MarshalizerMock{}

	err := node.createStateBlockBodyResolver()

	assert.Equal(t, "nil block body pool", err.Error())
	assert.Equal(t, 0, len(node.resolvers))
}

func TestCreateStateBlockBodyResolver_ShouldWork(t *testing.T) {
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
	node.blkc = createBlockchain()
	node.marshalizer = mock.MarshalizerMock{}

	err := node.createStateBlockBodyResolver()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(node.resolvers))
}

//------- createResolvers

func TestCreateResolvers_NilTransactionsShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topicHdr := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicTxBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicPeerBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicStateBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})

	dataPool := &mock.TransientDataPoolMock{}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		switch name {
		case string(TransactionTopic):
			return nil
		case string(HeadersTopic):
			return topicHdr
		case string(TxBlockBodyTopic):
			return topicTxBlk
		case string(PeerChBodyTopic):
			return topicPeerBlk
		case string(StateBodyTopic):
			return topicStateBlk
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
	node.marshalizer = mock.MarshalizerMock{}
	node.blkc = createBlockchain()
	node.uint64ByteSliceConverter = mock.NewNonceHashConverterMock()

	err := node.createResolvers()

	assert.Equal(t, "nil topic", err.Error())
}

func TestCreateResolvers_NilHeadersShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topicTx := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicTxBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicPeerBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicStateBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})

	dataPool := &mock.TransientDataPoolMock{}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		switch name {
		case string(TransactionTopic):
			return topicTx
		case string(HeadersTopic):
			return nil
		case string(TxBlockBodyTopic):
			return topicTxBlk
		case string(PeerChBodyTopic):
			return topicPeerBlk
		case string(StateBodyTopic):
			return topicStateBlk
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
	node.marshalizer = mock.MarshalizerMock{}
	node.blkc = createBlockchain()
	node.uint64ByteSliceConverter = mock.NewNonceHashConverterMock()

	err := node.createResolvers()

	assert.Equal(t, "nil topic", err.Error())
}

func TestCreateResolvers_NilTxBlocksShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topicTx := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicHdr := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicPeerBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicStateBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})

	dataPool := &mock.TransientDataPoolMock{}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		switch name {
		case string(TransactionTopic):
			return topicTx
		case string(HeadersTopic):
			return topicHdr
		case string(TxBlockBodyTopic):
			return nil
		case string(PeerChBodyTopic):
			return topicPeerBlk
		case string(StateBodyTopic):
			return topicStateBlk
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
	node.marshalizer = mock.MarshalizerMock{}
	node.blkc = createBlockchain()
	node.uint64ByteSliceConverter = mock.NewNonceHashConverterMock()

	err := node.createResolvers()

	assert.Equal(t, "nil topic", err.Error())
}

func TestCreateResolvers_NilPeerBlocksShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topicTx := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicHdr := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicTxBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicStateBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})

	dataPool := &mock.TransientDataPoolMock{}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		switch name {
		case string(TransactionTopic):
			return topicTx
		case string(HeadersTopic):
			return topicHdr
		case string(TxBlockBodyTopic):
			return topicTxBlk
		case string(PeerChBodyTopic):
			return nil
		case string(StateBodyTopic):
			return topicStateBlk
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
	node.marshalizer = mock.MarshalizerMock{}
	node.blkc = createBlockchain()
	node.uint64ByteSliceConverter = mock.NewNonceHashConverterMock()

	err := node.createResolvers()

	assert.Equal(t, "nil topic", err.Error())
}

func TestCreateResolvers_NilStateBlocksShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topicTx := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicHdr := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicTxBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicPeerBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})

	dataPool := &mock.TransientDataPoolMock{}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		switch name {
		case string(TransactionTopic):
			return topicTx
		case string(HeadersTopic):
			return topicHdr
		case string(TxBlockBodyTopic):
			return topicTxBlk
		case string(PeerChBodyTopic):
			return topicPeerBlk
		case string(StateBodyTopic):
			return nil
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
	node.marshalizer = mock.MarshalizerMock{}
	node.blkc = createBlockchain()
	node.uint64ByteSliceConverter = mock.NewNonceHashConverterMock()

	err := node.createResolvers()

	assert.Equal(t, "nil topic", err.Error())
}

func TestCreateResolvers_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	mes := mock.NewMessengerStub()
	node.messenger = mes

	topicTx := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicHdr := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicTxBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicPeerBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})
	topicStateBlk := p2p.NewTopic("", &mock.StringCreatorMock{}, mock.MarshalizerMock{})

	dataPool := &mock.TransientDataPoolMock{}

	mes.GetTopicCalled = func(name string) *p2p.Topic {
		switch name {
		case string(TransactionTopic):
			return topicTx
		case string(HeadersTopic):
			return topicHdr
		case string(TxBlockBodyTopic):
			return topicTxBlk
		case string(PeerChBodyTopic):
			return topicPeerBlk
		case string(StateBodyTopic):
			return topicStateBlk
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
	node.marshalizer = mock.MarshalizerMock{}
	node.blkc = createBlockchain()
	node.uint64ByteSliceConverter = mock.NewNonceHashConverterMock()

	err := node.createResolvers()

	assert.Nil(t, err)
	assert.Equal(t, 5, len(node.resolvers))
}
