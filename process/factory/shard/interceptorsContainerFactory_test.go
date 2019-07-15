package shard_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

var errExpected = errors.New("expected error")

func createStubTopicHandler(matchStrToErrOnCreate string, matchStrToErrOnRegister string) process.TopicHandler {
	return &mock.TopicHandlerStub{
		CreateTopicCalled: func(name string, createChannelForTopic bool) error {
			if matchStrToErrOnCreate == "" {
				return nil
			}
			if strings.Contains(name, matchStrToErrOnCreate) {
				return errExpected
			}

			return nil
		},
		RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
			if matchStrToErrOnRegister == "" {
				return nil
			}
			if strings.Contains(topic, matchStrToErrOnRegister) {
				return errExpected
			}

			return nil
		},
	}
}

func createDataPools() dataRetriever.PoolsHolder {
	pools := &mock.PoolsHolderStub{}
	pools.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	pools.HeadersCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	pools.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	pools.PeerChangesBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	pools.MetaBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	pools.MetaHeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	pools.UnsignedTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	return pools
}

func createStore() *mock.ChainStorerMock {
	return &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{}
		},
	}
}

//------- NewInterceptorsContainerFactory

func TestNewInterceptorsContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := shard.NewInterceptorsContainerFactory(
		nil,
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptorsContainerFactory_NilTopicHandlerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		nil,
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewInterceptorsContainerFactory_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		nil,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestNewInterceptorsContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		nil,
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptorsContainerFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		nil,
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptorsContainerFactory_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		nil,
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilKeyGen, err)
}

func TestNewInterceptorsContainerFactory_NilSingleSignerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		nil,
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilSingleSigner, err)
}

func TestNewInterceptorsContainerFactory_NilMultiSignerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		nil,
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewInterceptorsContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		nil,
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewInterceptorsContainerFactory_NilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		nil,
		&mock.ChronologyValidatorStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewInterceptorsContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	icf, err := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	assert.NotNil(t, icf)
	assert.Nil(t, err)
}

//------- Create

func TestInterceptorsContainerFactory_CreateTopicCreationTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicHandler(factory.TransactionTopic, ""),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateTopicCreationHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicHandler(factory.HeadersTopic, ""),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateTopicCreationMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicHandler(factory.MiniBlocksTopic, ""),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateTopicCreationPeerChBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicHandler(factory.PeerChBodyTopic, ""),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateTopicCreationMetachainHeadersFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicHandler(factory.MetachainBlocksTopic, ""),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateRegisterTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicHandler("", factory.TransactionTopic),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateRegisterHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicHandler("", factory.HeadersTopic),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateRegisterMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicHandler("", factory.MiniBlocksTopic),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateRegisterPeerChBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicHandler("", factory.PeerChBodyTopic),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateRegisterMetachainHeadersShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubTopicHandler("", factory.MetachainBlocksTopic),
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	icf, _ := shard.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
		},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	container, err := icf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestInterceptorsContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	icf, _ := shard.NewInterceptorsContainerFactory(
		shardCoordinator,
		&mock.TopicHandlerStub{
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
		},
		createStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPools(),
		&mock.AddressConverterMock{},
		&mock.ChronologyValidatorStub{},
	)

	container, _ := icf.Create()

	numInterceptorTxs := noOfShards + 1
	numInterceptorHeaders := 1
	numInterceptorMiniBlocks := noOfShards
	numInterceptorPeerChanges := 1
	numInterceptorMetachainHeaders := 1
	totalInterceptors := numInterceptorTxs + numInterceptorHeaders + numInterceptorMiniBlocks +
		numInterceptorPeerChanges + numInterceptorMetachainHeaders + numInterceptorTxs

	assert.Equal(t, totalInterceptors, container.Len())
}
