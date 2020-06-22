package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewTriesComponentsFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := getTriesArgs()
	args.Marshalizer = nil
	tcf, err := factory.NewTriesComponentsFactory(args)
	require.Nil(t, tcf)
	require.Equal(t, factory.ErrNilMarshalizer, err)
}

func TestNewTriesComponentsFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := getTriesArgs()
	args.Hasher = nil
	tcf, err := factory.NewTriesComponentsFactory(args)
	require.Nil(t, tcf)
	require.Equal(t, factory.ErrNilHasher, err)
}

func TestNewTriesComponentsFactory_NilPathManagerShouldErr(t *testing.T) {
	t.Parallel()

	args := getTriesArgs()
	args.PathManager = nil
	tcf, err := factory.NewTriesComponentsFactory(args)
	require.Nil(t, tcf)
	require.Equal(t, factory.ErrNilPathManager, err)
}

func TestNewTriesComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getTriesArgs()
	args.ShardCoordinator = nil
	tcf, err := factory.NewTriesComponentsFactory(args)
	require.Nil(t, tcf)
	require.Equal(t, factory.ErrNilShardCoordinator, err)
}

func TestNewTriesComponentsFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := getTriesArgs()
	tcf, err := factory.NewTriesComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, tcf)
}

func TestTriesComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	args := getTriesArgs()
	tcf, _ := factory.NewTriesComponentsFactory(args)

	tc, err := tcf.Create()
	require.NoError(t, err)
	require.NotNil(t, tc)
}

func getTriesArgs() factory.TriesComponentsFactoryArgs {
	return factory.TriesComponentsFactoryArgs{
		Marshalizer:      &mock.MarshalizerMock{},
		Hasher:           &mock.HasherMock{},
		PathManager:      &mock.PathManagerStub{},
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(2),
		Config:           testscommon.GetGeneralConfig(),
	}
}
