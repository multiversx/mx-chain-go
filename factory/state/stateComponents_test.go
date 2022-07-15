package state_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	componentsMock "github.com/ElrondNetwork/elrond-go/factory/mock/components"
	stateComp "github.com/ElrondNetwork/elrond-go/factory/state"
	"github.com/stretchr/testify/require"
)

func TestNewStateComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetStateArgs(coreComponents, shardCoordinator)
	args.ShardCoordinator = nil

	scf, err := stateComp.NewStateComponentsFactory(args)
	require.Nil(t, scf)
	require.Equal(t, errors.ErrNilShardCoordinator, err)
}

func TestNewStateComponentsFactory_NilCoreComponents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetStateArgs(coreComponents, shardCoordinator)
	args.Core = nil

	scf, err := stateComp.NewStateComponentsFactory(args)
	require.Nil(t, scf)
	require.Equal(t, errors.ErrNilCoreComponents, err)
}

func TestNewStateComponentsFactory_ShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetStateArgs(coreComponents, shardCoordinator)

	scf, err := stateComp.NewStateComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, scf)
}

func TestStateComponentsFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetStateArgs(coreComponents, shardCoordinator)

	scf, _ := stateComp.NewStateComponentsFactory(args)

	res, err := scf.Create()
	require.NoError(t, err)
	require.NotNil(t, res)
}

// ------------ Test StateComponents --------------------
func TestStateComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetStateArgs(coreComponents, shardCoordinator)
	scf, _ := stateComp.NewStateComponentsFactory(args)

	sc, _ := scf.Create()

	err := sc.Close()
	require.NoError(t, err)
}
