package components

import (
	"testing"

	"github.com/stretchr/testify/require"

	mock2 "github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
)

func createArgsDataComponentsHolder() ArgsDataComponentsHolder {
	return ArgsDataComponentsHolder{
		CoreComponents:       &factory.CoreComponentsHolderMock{},
		StatusCoreComponents: &factory.StatusCoreComponentsStub{},
		BootstrapComponents:  &mainFactoryMocks.BootstrapComponentsStub{},
		CryptoComponents: &mock2.CryptoComponentsStub{
			PubKey:                  &mock2.PublicKeyMock{},
			BlockSig:                &cryptoMocks.SingleSignerStub{},
			BlKeyGen:                &cryptoMocks.KeyGenStub{},
			TxSig:                   &cryptoMocks.SingleSignerStub{},
			TxKeyGen:                &cryptoMocks.KeyGenStub{},
			ManagedPeersHolderField: &testscommon.ManagedPeersHolderStub{},
		},
		RunTypeComponents: &mainFactoryMocks.RunTypeComponentsStub{},
	}
}

func TestCreateDataComponents(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		comp, err := CreateDataComponents(createArgsDataComponentsHolder())
		require.NoError(t, err)
		require.NotNil(t, comp)

		require.Nil(t, comp.Create())
		require.Nil(t, comp.Close())
	})
	//t.Run("NewMiniBlockProvider failure should error", func(t *testing.T) {
	//	t.Parallel()
	//
	//	args := createArgsDataComponentsHolder()
	//	args. = &dataRetriever.PoolsHolderStub{
	//		MiniBlocksCalled: func() chainStorage.Cacher {
	//			return nil
	//		},
	//	}
	//	comp, err := CreateDataComponents(args)
	//	require.Error(t, err)
	//	require.Nil(t, comp)
	//})
	//t.Run("GetStorer failure should error", func(t *testing.T) {
	//	t.Parallel()
	//
	//	args := createArgsDataComponentsHolder()
	//	args.StorageService = &storage.ChainStorerStub{
	//		GetStorerCalled: func(unitType retriever.UnitType) (chainStorage.Storer, error) {
	//			return nil, expectedErr
	//		},
	//	}
	//	comp, err := CreateDataComponents(args)
	//	require.Equal(t, expectedErr, err)
	//	require.Nil(t, comp)
	//})
}

func TestDataComponentsHolder_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var comp *dataComponentsHolder
	require.True(t, comp.IsInterfaceNil())

	comp, _ = CreateDataComponents(createArgsDataComponentsHolder())
	require.False(t, comp.IsInterfaceNil())
	require.Nil(t, comp.Close())
}

func TestDataComponentsHolder_Getters(t *testing.T) {
	t.Parallel()

	comp, err := CreateDataComponents(createArgsDataComponentsHolder())
	require.NoError(t, err)

	require.NotNil(t, comp.Blockchain())
	require.Nil(t, comp.SetBlockchain(nil))
	require.Nil(t, comp.Blockchain())
	require.NotNil(t, comp.StorageService())
	require.NotNil(t, comp.Datapool())
	require.NotNil(t, comp.MiniBlocksProvider())
	require.Nil(t, comp.CheckSubcomponents())
	require.Empty(t, comp.String())
	require.Nil(t, comp.Close())
}

func TestDataComponentsHolder_Clone(t *testing.T) {
	t.Parallel()

	comp, err := CreateDataComponents(createArgsDataComponentsHolder())
	require.NoError(t, err)

	compClone := comp.Clone()
	require.Equal(t, comp, compClone)
	require.False(t, comp == compClone) // pointer testing
	require.Nil(t, comp.Close())
}
