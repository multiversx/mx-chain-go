package epochStartTrigger

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	retriever "github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/factory"
	nodeFactoryMock "github.com/multiversx/mx-chain-go/node/mock/factory"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding"
	shardingMock "github.com/multiversx/mx-chain-go/sharding/mock"
	chainStorage "github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	testsFactory "github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	validatorInfoCacherStub "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
	updateMock "github.com/multiversx/mx-chain-go/update/mock"
)

func createArgs() factory.ArgsEpochStartTrigger {
	return factory.ArgsEpochStartTrigger{
		RequestHandler: &testscommon.RequestHandlerStub{},
		CoreData: &testsFactory.CoreComponentsHolderMock{
			HasherCalled: func() hashing.Hasher {
				return &testscommon.HasherStub{}
			},
			InternalMarshalizerCalled: func() marshal.Marshalizer {
				return &testscommon.MarshallerStub{}
			},
			Uint64ByteSliceConverterCalled: func() typeConverters.Uint64ByteSliceConverter {
				return &testscommon.Uint64ByteSliceConverterStub{}
			},
			EpochStartNotifierWithConfirmCalled: func() factory.EpochStartNotifierWithConfirm {
				return &updateMock.EpochStartNotifierStub{}
			},
			RoundHandlerCalled: func() consensus.RoundHandler {
				return &testscommon.RoundHandlerMock{}
			},
			EnableEpochsHandlerCalled: func() common.EnableEpochsHandler {
				return &shardingMock.EnableEpochsHandlerMock{}
			},
		},
		BootstrapComponents: &mainFactoryMocks.BootstrapComponentsStub{
			ShardCoordinatorCalled: func() sharding.Coordinator {
				return &testscommon.ShardsCoordinatorMock{
					NoShards: 1,
				}
			},
			BootstrapParams:      &bootstrapMocks.BootstrapParamsHandlerMock{},
			HdrIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		},
		DataComps: &nodeFactoryMock.DataComponentsMock{
			DataPool: &dataRetriever.PoolsHolderStub{
				MetaBlocksCalled: func() chainStorage.Cacher {
					return &testscommon.CacherStub{}
				},
				HeadersCalled: func() retriever.HeadersPool {
					return &pool.HeadersPoolStub{}
				},
				ValidatorsInfoCalled: func() retriever.ShardedDataCacherNotifier {
					return &testscommon.ShardedDataCacheNotifierMock{}
				},
				CurrEpochValidatorInfoCalled: func() retriever.ValidatorInfoCacher {
					return &validatorInfoCacherStub.ValidatorInfoCacherStub{}
				},
			},
			Store: &storage.ChainStorerStub{
				GetStorerCalled: func(unitType retriever.UnitType) (chainStorage.Storer, error) {
					return &storage.StorerStub{}, nil
				},
			},
		},
		StatusCoreComponentsHolder: &testsFactory.StatusCoreComponentsStub{
			AppStatusHandlerField: &statusHandler.AppStatusHandlerStub{},
		},
		RunTypeComponentsHolder: mainFactoryMocks.NewRunTypeComponentsStub(),
		Config:                  config.Config{},
	}
}

func TestNewEpochStartTriggerFactory(t *testing.T) {
	t.Parallel()

	f := NewEpochStartTriggerFactory()

	args := createArgs()
	trigger, err := f.CreateEpochStartTrigger(args)
	require.Nil(t, err)
	require.NotNil(t, trigger)
}
