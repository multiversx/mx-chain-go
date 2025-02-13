package epochStartTrigger

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
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
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding"
	shardingMock "github.com/multiversx/mx-chain-go/sharding/mock"
	chainStorage "github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	testsFactory "github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	validatorInfoCacherStub "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
	updateMock "github.com/multiversx/mx-chain-go/update/mock"
)

func createArgs(shardID uint32) factory.ArgsEpochStartTrigger {
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
			GenesisNodesSetupCalled: func() sharding.GenesisNodesSetupHandler {
				return &genesisMocks.NodesSetupStub{}
			},
		},
		BootstrapComponents: createBootstrapComps(shardID),
		DataComps:           createDataCompsMock(),
		StatusCoreComponentsHolder: &testsFactory.StatusCoreComponentsStub{
			AppStatusHandlerField: &statusHandler.AppStatusHandlerStub{},
		},
		RunTypeComponentsHolder: mainFactoryMocks.NewRunTypeComponentsStub(),
		Config: config.Config{
			EpochStartConfig: config.EpochStartConfig{
				RoundsPerEpoch:         22,
				MinRoundsBetweenEpochs: 22,
			},
		},
	}
}

func createBootstrapComps(shardID uint32) *mainFactoryMocks.BootstrapComponentsStub {
	return &mainFactoryMocks.BootstrapComponentsStub{
		ShardCoordinatorCalled: func() sharding.Coordinator {
			return &testscommon.ShardsCoordinatorMock{
				NoShards: 1,
				SelfIDCalled: func() uint32 {
					return shardID
				},
			}
		},
		BootstrapParams:      &bootstrapMocks.BootstrapParamsHandlerMock{},
		HdrIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		Bootstrapper:         &bootstrapMocks.EpochStartBootstrapperStub{},
	}
}

func createDataCompsMock() *nodeFactoryMock.DataComponentsMock {
	return &nodeFactoryMock.DataComponentsMock{
		DataPool: createDataPoolMock(),
		Store: &storage.ChainStorerStub{
			GetStorerCalled: func(unitType retriever.UnitType) (chainStorage.Storer, error) {
				return &storage.StorerStub{}, nil
			},
		},
		BlockChain: &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.HeaderV2{}
			},
		},
	}
}

func createDataPoolMock() *dataRetriever.PoolsHolderStub {
	return &dataRetriever.PoolsHolderStub{
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
	}
}

func TestNewEpochStartTriggerFactory(t *testing.T) {
	t.Parallel()

	f := NewEpochStartTriggerFactory()
	require.False(t, f.IsInterfaceNil())

	t.Run("create for shard", func(t *testing.T) {
		args := createArgs(0)
		trigger, err := f.CreateEpochStartTrigger(args)
		require.Nil(t, err)
		require.Equal(t, "*shardchain.trigger", fmt.Sprintf("%T", trigger))
	})
	t.Run("create for meta", func(t *testing.T) {
		args := createArgs(core.MetachainShardId)
		trigger, err := f.CreateEpochStartTrigger(args)
		require.Nil(t, err)
		require.Equal(t, "*metachain.trigger", fmt.Sprintf("%T", trigger))
	})
	t.Run("invalid shard id", func(t *testing.T) {
		args := createArgs(444)
		trigger, err := f.CreateEpochStartTrigger(args)
		require.ErrorIs(t, err, process.ErrInvalidShardId)
		require.Nil(t, trigger)
	})
}
