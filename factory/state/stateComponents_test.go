package state_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/errors"
	stateComp "github.com/multiversx/mx-chain-go/factory/state"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/stretchr/testify/require"
)

func TestNewStateComponentsFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil Core should error", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetStateFactoryArgs(coreComponents, componentsMock.GetStatusCoreComponents())
		args.Core = nil

		scf, err := stateComp.NewStateComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errors.ErrNilCoreComponents, err)
	})
	t.Run("nil StatusCore should error", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetStateFactoryArgs(coreComponents, componentsMock.GetStatusCoreComponents())
		args.StatusCore = nil

		scf, err := stateComp.NewStateComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errors.ErrNilStatusCoreComponents, err)
	})
	t.Run("nil accounts creator, should error", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetStateFactoryArgs(coreComponents, componentsMock.GetStatusCoreComponents())
		args.AccountsCreator = nil

		scf, err := stateComp.NewStateComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, state.ErrNilAccountFactory, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetStateFactoryArgs(coreComponents, componentsMock.GetStatusCoreComponents())

		scf, err := stateComp.NewStateComponentsFactory(args)
		require.NoError(t, err)
		require.NotNil(t, scf)
	})
}

func TestStateComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	t.Run("CreateTriesComponentsForShardId fails should error", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetStateFactoryArgs(coreComponents, componentsMock.GetStatusCoreComponents())
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
		args := componentsMock.GetStateFactoryArgs(coreComponents, componentsMock.GetStatusCoreComponents())
		args.Config.EvictionWaitingList.RootHashesSize = 0
		scf, _ := stateComp.NewStateComponentsFactory(args)

		sc, err := scf.Create()
		require.Error(t, err)
		require.Nil(t, sc)
	})
	t.Run("NewAccountsDB fails should error", func(t *testing.T) {
		t.Parallel()

		coreComponents := componentsMock.GetCoreComponents()
		args := componentsMock.GetStateFactoryArgs(coreComponents, componentsMock.GetStatusCoreComponents())

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
		args := componentsMock.GetStateFactoryArgs(coreComponents, componentsMock.GetStatusCoreComponents())

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
		args := componentsMock.GetStateFactoryArgs(coreComponents, componentsMock.GetStatusCoreComponents())
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
	args := componentsMock.GetStateFactoryArgs(coreComponents, componentsMock.GetStatusCoreComponents())
	scf, _ := stateComp.NewStateComponentsFactory(args)

	sc, err := scf.Create()
	require.NoError(t, err)
	require.NotNil(t, sc)

	require.NoError(t, sc.Close())
}
