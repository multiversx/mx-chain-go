package runType_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/runType"
	stateComp "github.com/multiversx/mx-chain-go/factory/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/stretchr/testify/require"
)

func TestNewRunTypeComponentsFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil BlockChainHookHandlerCreator should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetRunTypeFactoryArgs()
		args.BlockChainHookHandlerCreator = nil

		scf, err := runType.NewRunTypeComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errors.ErrNilBlockChainHookHandlerCreator, err)
	})
	t.Run("nil BlockProcessorCreator should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetRunTypeFactoryArgs()
		args.BlockProcessorCreator = nil

		scf, err := runType.NewRunTypeComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errors.ErrNilBlockProcessorCreator, err)
	})
	t.Run("nil BlockTrackerCreator should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetRunTypeFactoryArgs()
		args.BlockTrackerCreator = nil

		scf, err := runType.NewRunTypeComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errors.ErrNilBlockTrackerCreator, err)
	})
	t.Run("nil BootstrapperFromStorageCreator should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetRunTypeFactoryArgs()
		args.BootstrapperFromStorageCreator = nil

		scf, err := runType.NewRunTypeComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errors.ErrNilBootstrapperFromStorageCreator, err)
	})
	t.Run("nil EpochStartBootstrapperCreator should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetRunTypeFactoryArgs()
		args.EpochStartBootstrapperCreator = nil

		scf, err := runType.NewRunTypeComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errors.ErrNilEpochStartBootstrapperCreator, err)
	})
	t.Run("nil ForkDetectorCreator should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetRunTypeFactoryArgs()
		args.ForkDetectorCreator = nil

		scf, err := runType.NewRunTypeComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errors.ErrNilForkDetectorCreator, err)
	})
	t.Run("nil HeaderValidatorCreator should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetRunTypeFactoryArgs()
		args.HeaderValidatorCreator = nil

		scf, err := runType.NewRunTypeComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errors.ErrNilHeaderValidatorCreator, err)
	})
	t.Run("nil RequestHandlerCreator should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetRunTypeFactoryArgs()
		args.RequestHandlerCreator = nil

		scf, err := runType.NewRunTypeComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errors.ErrNilRequestHandlerCreator, err)
	})
	t.Run("nil ScheduledTxsExecutionCreator should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetRunTypeFactoryArgs()
		args.ScheduledTxsExecutionCreator = nil

		scf, err := runType.NewRunTypeComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errors.ErrNilScheduledTxsExecutionCreator, err)
	})
	t.Run("nil TransactionCoordinatorCreator should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetRunTypeFactoryArgs()
		args.TransactionCoordinatorCreator = nil

		scf, err := runType.NewRunTypeComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errors.ErrNilTransactionCoordinatorCreator, err)
	})
	t.Run("nil ValidatorStatisticsProcessorCreator should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetRunTypeFactoryArgs()
		args.ValidatorStatisticsProcessorCreator = nil

		scf, err := runType.NewRunTypeComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errors.ErrNilValidatorStatisticsProcessorCreator, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetRunTypeFactoryArgs()
		rcf, err := runType.NewRunTypeComponentsFactory(args)
		require.NoError(t, err)
		require.NotNil(t, rcf)
	})
}

func TestStateComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	t.Run("CreateTriesComponentsForShardId fails should error", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetStateFactoryArgs(coreComponents)
		coreCompStub := factory.NewCoreComponentsHolderStubFromRealComponent(args.Core)
		coreCompStub.InternalMarshalizerCalled = func() marshal.Marshalizer {
			return nil
		}
		args.Core = coreCompStub
		scf, _ := stateComp.NewStateComponentsFactory(args)

		sc, err := scf.Create()
		require.Error(t, err)
		require.Nil(t, sc)
	})
	t.Run("NewMemoryEvictionWaitingList fails should error", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetStateFactoryArgs(coreComponents)
		args.Config.EvictionWaitingList.RootHashesSize = 0
		scf, _ := stateComp.NewStateComponentsFactory(args)

		sc, err := scf.Create()
		require.Error(t, err)
		require.Nil(t, sc)
	})
	t.Run("NewAccountsDB fails should error", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetStateFactoryArgs(coreComponents)

		coreCompStub := factory.NewCoreComponentsHolderStubFromRealComponent(args.Core)
		cnt := 0
		coreCompStub.HasherCalled = func() hashing.Hasher {
			cnt++
			if cnt > 1 {
				return nil
			}
			return &testscommon.HasherStub{}
		}
		args.Core = coreCompStub
		scf, _ := stateComp.NewStateComponentsFactory(args)

		sc, err := scf.Create()
		require.Error(t, err)
		require.Nil(t, sc)
	})
	t.Run("CreateAccountsAdapterAPIOnFinal fails should error", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetStateFactoryArgs(coreComponents)

		coreCompStub := factory.NewCoreComponentsHolderStubFromRealComponent(args.Core)
		cnt := 0
		coreCompStub.HasherCalled = func() hashing.Hasher {
			cnt++
			if cnt > 2 {
				return nil
			}
			return &testscommon.HasherStub{}
		}
		args.Core = coreCompStub
		scf, _ := stateComp.NewStateComponentsFactory(args)

		sc, err := scf.Create()
		require.Error(t, err)
		require.Nil(t, sc)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetStateFactoryArgs(coreComponents)
		scf, _ := stateComp.NewStateComponentsFactory(args)

		sc, err := scf.Create()
		require.NoError(t, err)
		require.NotNil(t, sc)
		require.NoError(t, sc.Close())
	})
}

func TestStateComponents_Close(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetStateFactoryArgs(coreComponents)
	scf, _ := stateComp.NewStateComponentsFactory(args)

	sc, err := scf.Create()
	require.NoError(t, err)
	require.NotNil(t, sc)

	require.NoError(t, sc.Close())
}
