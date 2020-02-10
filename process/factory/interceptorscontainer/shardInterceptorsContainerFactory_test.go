package interceptorscontainer_test

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/interceptorscontainer"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

func createShardStubTopicHandler(matchStrToErrOnCreate string, matchStrToErrOnRegister string) process.TopicHandler {
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

func createShardDataPools() dataRetriever.PoolsHolder {
	pools := &mock.PoolsHolderStub{}
	pools.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{}
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
	pools.UnsignedTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	pools.RewardTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	pools.TrieNodesCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	pools.CurrBlockTxsCalled = func() dataRetriever.TransactionCacher {
		return &mock.TxForCurrentBlockStub{}
	}
	return pools
}

func createShardStore() *mock.ChainStorerMock {
	return &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{}
		},
	}
}

//------- NewInterceptorsContainerFactory
func TestNewShardInterceptorsContainerFactory_NilAccountsAdapter(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		nil,
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewShardInterceptorsContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		nil,
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewShardInterceptorsContainerFactory_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		nil,
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilNodesCoordinator, err)
}

func TestNewShardInterceptorsContainerFactory_NilTopicHandlerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		nil,
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewShardInterceptorsContainerFactory_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		nil,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilStore, err)
}

func TestNewShardInterceptorsContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		nil,
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewShardInterceptorsContainerFactory_NilMarshalizerAndSizeCheckShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		nil,
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		1,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewShardInterceptorsContainerFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		nil,
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewShardInterceptorsContainerFactory_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		nil,
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilKeyGen, err)
}

func TestNewShardInterceptorsContainerFactory_NilSingleSignerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		nil,
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilSingleSigner, err)
}

func TestNewShardInterceptorsContainerFactory_NilMultiSignerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		nil,
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewShardInterceptorsContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		nil,
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewShardInterceptorsContainerFactory_NilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		nil,
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewShardInterceptorsContainerFactory_NilTxFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		nil,
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewShardInterceptorsContainerFactory_NilBlackListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		nil,
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilBlackListHandler, err)
}

func TestNewShardInterceptorsContainerFactory_EmptyChainIDShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		nil,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrInvalidChainID, err)
}

func TestNewShardInterceptorsContainerFactory_NilValidityAttesterShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		nil,
		&mock.EpochStartTriggerStub{},
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilValidityAttester, err)
}

func TestNewShardInterceptorsContainerFactory_EmptyEpochStartTriggerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		nil,
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilEpochStartTrigger, err)
}

func TestNewShardInterceptorsContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.NotNil(t, icf)
	assert.Nil(t, err)
	assert.False(t, icf.IsInterfaceNil())
}
func TestNewShardInterceptorsContainerFactory_ShouldWorkWithSizeCheck(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		1,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	assert.NotNil(t, icf)
	assert.Nil(t, err)
}

//------- Create

func TestShardInterceptorsContainerFactory_CreateTopicCreationTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createShardStubTopicHandler(factory.TransactionTopic, ""),
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateTopicCreationHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createShardStubTopicHandler(factory.ShardBlocksTopic, ""),
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateTopicCreationMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createShardStubTopicHandler(factory.MiniBlocksTopic, ""),
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateTopicCreationMetachainHeadersFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createShardStubTopicHandler(factory.MetachainBlocksTopic, ""),
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateRegisterTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createShardStubTopicHandler("", factory.TransactionTopic),
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateRegisterHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createShardStubTopicHandler("", factory.ShardBlocksTopic),
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateRegisterMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createShardStubTopicHandler("", factory.MiniBlocksTopic),
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateRegisterMetachainHeadersShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createShardStubTopicHandler("", factory.MetachainBlocksTopic),
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateRegisterTrieNodesShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createShardStubTopicHandler("", factory.AccountTrieNodesTopic),
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardInterceptorsContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
		},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestShardInterceptorsContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	nodesCoordinator := &mock.NodesCoordinatorMock{
		ShardId:            1,
		ShardConsensusSize: 1,
		MetaConsensusSize:  1,
		NbShards:           uint32(noOfShards),
	}

	icf, _ := interceptorscontainer.NewShardInterceptorsContainerFactory(
		&mock.AccountsStub{},
		shardCoordinator,
		nodesCoordinator,
		&mock.TopicHandlerStub{
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
		},
		createShardStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createShardDataPools(),
		&mock.AddressConverterMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		chainID,
		0,
		&mock.ValidityAttesterStub{},
		&mock.EpochStartTriggerStub{},
	)

	container, err := icf.Create()

	numInterceptorTxs := noOfShards + 1
	numInterceptorsUnsignedTxs := numInterceptorTxs
	numInterceptorsRewardTxs := numInterceptorTxs
	numInterceptorHeaders := 1
	numInterceptorMiniBlocks := noOfShards + 1
	numInterceptorMetachainHeaders := 1
	numInterceptorTrieNodes := 2
	totalInterceptors := numInterceptorTxs + numInterceptorsUnsignedTxs + numInterceptorsRewardTxs +
		numInterceptorHeaders + numInterceptorMiniBlocks + numInterceptorMetachainHeaders + numInterceptorTrieNodes

	assert.Nil(t, err)
	assert.Equal(t, totalInterceptors, container.Len())
}
