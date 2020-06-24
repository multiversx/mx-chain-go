package factory_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatusComponentsFactory_NilParamsShouldErr(t *testing.T) {
	t.Parallel()

	okArgs := getStatusComponentsFactoryArgs()

	testInput := make(map[factory.StatusComponentsFactoryArgs]error)

	nilCoreDataArgs := okArgs
	nilCoreDataArgs.CoreData = nil
	testInput[nilCoreDataArgs] = factory.ErrNilCoreComponentsHolder

	nilShardCoordinatorArgs := okArgs
	nilShardCoordinatorArgs.ShardCoordinator = nil
	testInput[nilShardCoordinatorArgs] = factory.ErrNilShardCoordinator

	nilEpochStartNotifierArgs := okArgs
	nilEpochStartNotifierArgs.EpochNotifier = nil
	testInput[nilEpochStartNotifierArgs] = factory.ErrNilEpochStartNotifier

	nilNodesCoordinatorArgs := okArgs
	nilNodesCoordinatorArgs.NodesCoordinator = nil
	testInput[nilNodesCoordinatorArgs] = factory.ErrNilNodesCoordinator

	invalidRoundDurationArgs := okArgs
	invalidRoundDurationArgs.RoundDurationSec = 0
	testInput[invalidRoundDurationArgs] = factory.ErrInvalidRoundDuration

	nilAddressPubKeyConverterArgs := okArgs
	nilAddressPubKeyConverterArgs.AddressPubkeyConverter = nil
	testInput[nilAddressPubKeyConverterArgs] = factory.ErrNilPubKeyConverter

	nilValidatorPubKeyConverterArgs := okArgs
	nilValidatorPubKeyConverterArgs.ValidatorPubkeyConverter = nil
	testInput[nilValidatorPubKeyConverterArgs] = factory.ErrNilPubKeyConverter

	nilElasticOptionsArgs := okArgs
	nilElasticOptionsArgs.ElasticOptions = nil
	testInput[nilElasticOptionsArgs] = factory.ErrNilElasticOptions

	for args, expectedErr := range testInput {
		scf, err := factory.NewStatusComponentsFactory(args)
		assert.True(t, check.IfNil(scf))
		assert.True(t, errors.Is(err, expectedErr), "expected %s in test, but got %s", expectedErr.Error(), err.Error())
	}
}

func TestNewStatusComponentsFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	scf, err := factory.NewStatusComponentsFactory(getStatusComponentsFactoryArgs())
	require.NoError(t, err)
	require.False(t, check.IfNil(scf))
}

func TestStatusComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	args := getStatusComponentsFactoryArgs()
	scf, _ := factory.NewStatusComponentsFactory(args)

	res, err := scf.Create()
	require.NoError(t, err)
	require.NotNil(t, res)
}

func getStatusComponentsFactoryArgs() factory.StatusComponentsFactoryArgs {
	return factory.StatusComponentsFactoryArgs{
		CoreData:                 getCoreComponents(),
		ShardCoordinator:         mock.NewMultiShardsCoordinatorMock(2),
		ElasticConfig:            config.ElasticSearchConfig{},
		EpochNotifier:            &mock.EpochStartNotifierStub{},
		NodesCoordinator:         &mock.NodesCoordinatorMock{},
		AddressPubkeyConverter:   &mock.PubkeyConverterStub{},
		ValidatorPubkeyConverter: &mock.PubkeyConverterStub{},
		ElasticOptions:           &indexer.Options{},
		RoundDurationSec:         4,
		SoftwareVersionConfig: config.SoftwareVersionConfig{
			StableTagLocation:        "tag",
			PollingIntervalInMinutes: 1,
		},
	}
}
