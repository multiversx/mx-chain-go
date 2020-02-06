package interceptorscontainer_test

import (
	"errors"
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

const maxTxNonceDeltaAllowed = 100

var chainID = []byte("chain ID")
var errExpected = errors.New("expected error")

func createMetaStubTopicHandler(matchStrToErrOnCreate string, matchStrToErrOnRegister string) process.TopicHandler {
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

func createMetaDataPools() dataRetriever.PoolsHolder {
	pools := &mock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{}
		},
		MiniBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{}
		},
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		TrieNodesCalled: func() storage.Cacher {
			return &mock.CacherStub{}
		},
	}

	return pools
}

func createMetaStore() *mock.ChainStorerMock {
	return &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{}
		},
	}
}

//------- NewInterceptorsContainerFactory

func TestNewMetaInterceptorsContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		nil,
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		nil,
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_NilTopicHandlerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		nil,
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		nil,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		nil,
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_NilMarshalizerAndSizeCheckShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		nil,
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		nil,
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_NilMultiSignerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		nil,
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		nil,
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		nil,
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_NilAddrConvShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		nil,
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_NilSingleSignerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		nil,
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		nil,
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_NilFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_NilBlackListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_EmptyCahinIDShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_NilValidityAttesterShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestNewMetaInterceptorsContainerFactory_EpochStartTriggerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
		maxTxNonceDeltaAllowed,
		&mock.FeeHandlerStub{},
		&mock.BlackListHandlerStub{},
		&mock.HeaderSigVerifierStub{},
		nil,
		0,
		&mock.ValidityAttesterStub{},
		nil,
	)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilEpochStartTrigger, err)
}
func TestNewMetaInterceptorsContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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
}

func TestNewMetaInterceptorsContainerFactory_ShouldWorkWithSizeCheck(t *testing.T) {
	t.Parallel()

	icf, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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
	assert.False(t, icf.IsInterfaceNil())
}

//------- Create

func TestMetaInterceptorsContainerFactory_CreateTopicMetablocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createMetaStubTopicHandler(factory.MetachainBlocksTopic, ""),
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestMetaInterceptorsContainerFactory_CreateTopicShardHeadersForMetachainFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createMetaStubTopicHandler(factory.ShardBlocksTopic, ""),
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestMetaInterceptorsContainerFactory_CreateRegisterForMetablocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createMetaStubTopicHandler("", factory.MetachainBlocksTopic),
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestMetaInterceptorsContainerFactory_CreateRegisterShardHeadersForMetachainFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createMetaStubTopicHandler("", factory.ShardBlocksTopic),
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestMetaInterceptorsContainerFactory_CreateRegisterTrieNodesFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewMetaInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
		createMetaStubTopicHandler("", factory.AccountTrieNodesTopic),
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestMetaInterceptorsContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	icf, _ := interceptorscontainer.NewMetaInterceptorsContainerFactory(
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
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

func TestMetaInterceptorsContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	nodesCoordinator := &mock.NodesCoordinatorMock{
		ShardConsensusSize: 1,
		MetaConsensusSize:  1,
		NbShards:           uint32(noOfShards),
		ShardId:            1,
	}

	icf, _ := interceptorscontainer.NewMetaInterceptorsContainerFactory(
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
		createMetaStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		mock.NewMultiSigner(),
		createMetaDataPools(),
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		&mock.SignerMock{},
		&mock.SignerMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SingleSignKeyGenMock{},
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

	numInterceptorsMetablock := 1
	numInterceptorsShardHeadersForMetachain := noOfShards
	numInterceptorsTransactionsForMetachain := noOfShards + 1
	numInterceptorsMiniBlocksForMetachain := noOfShards + 1
	numInterceptorsUnsignedTxsForMetachain := noOfShards
	numInterceptorsTrieNodes := (noOfShards + 1) * 2
	totalInterceptors := numInterceptorsMetablock + numInterceptorsShardHeadersForMetachain + numInterceptorsTrieNodes +
		numInterceptorsTransactionsForMetachain + numInterceptorsUnsignedTxsForMetachain + numInterceptorsMiniBlocksForMetachain

	assert.Nil(t, err)
	assert.Equal(t, totalInterceptors, container.Len())
}
