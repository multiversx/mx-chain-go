package runType_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/runType"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
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
	t.Run("nil AdditionalStorageServiceCreator should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetRunTypeFactoryArgs()
		args.AdditionalStorageServiceCreator = nil

		scf, err := runType.NewRunTypeComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errors.ErrNilAdditionalStorageServiceCreator, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetRunTypeFactoryArgs()
		rcf, err := runType.NewRunTypeComponentsFactory(args)
		require.NoError(t, err)
		require.NotNil(t, rcf)
	})
}

func TestRunTypeComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	t.Run("Create should work", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetRunTypeFactoryArgs()
		rcf, err := runType.NewRunTypeComponentsFactory(args)
		require.NoError(t, err)
		require.NotNil(t, rcf)

		rc, err := rcf.Create()

		require.NoError(t, err)
		require.NotNil(t, rc)
	})
}

func TestRunTypeComponentsFactory_Close(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetRunTypeFactoryArgs()
	rcf, err := runType.NewRunTypeComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, rcf)

	rc, err := rcf.Create()
	require.NoError(t, err)
	require.NotNil(t, rc)

	require.NoError(t, rc.Close())
}
