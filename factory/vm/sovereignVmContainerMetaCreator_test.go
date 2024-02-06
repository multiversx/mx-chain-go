package vm_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignVmContainerMetaCreatorFactory(t *testing.T) {
	t.Parallel()

	bhhc := &factory.BlockChainHookHandlerFactoryMock{}
	vmContainerMetaFactory, err := vm.NewSovereignVmContainerMetaFactory(bhhc)
	require.Nil(t, err)
	require.False(t, vmContainerMetaFactory.IsInterfaceNil())
}

func TestSovereignVmContainerMetaFactory_CreateVmContainerFactoryMeta(t *testing.T) {
	t.Parallel()

	bhhc := &factory.BlockChainHookHandlerFactoryMock{}
	vmContainerMetaFactory, err := vm.NewSovereignVmContainerMetaFactory(bhhc)
	require.Nil(t, err)
	require.False(t, vmContainerMetaFactory.IsInterfaceNil())

	argsBlockchain := createMockBlockChainHookArgs()
	gasSchedule := makeGasSchedule()
	argsMeta := createVmContainerMockArgument(gasSchedule)
	args := vm.ArgsVmContainerFactory{
		Economics:           argsMeta.Economics,
		MessageSignVerifier: argsMeta.MessageSignVerifier,
		GasSchedule:         argsMeta.GasSchedule,
		NodesConfigProvider: argsMeta.NodesConfigProvider,
		Hasher:              argsMeta.Hasher,
		Marshalizer:         argsMeta.Marshalizer,
		SystemSCConfig:      argsMeta.SystemSCConfig,
		ValidatorAccountsDB: argsMeta.ValidatorAccountsDB,
		UserAccountsDB:      argsMeta.UserAccountsDB,
		ChanceComputer:      argsMeta.ChanceComputer,
		ShardCoordinator:    argsMeta.ShardCoordinator,
		PubkeyConv:          argsMeta.PubkeyConv,
		EnableEpochsHandler: argsMeta.EnableEpochsHandler,
	}

	vmContainer, vmFactory, err := vmContainerMetaFactory.CreateVmContainerFactoryMeta(argsBlockchain, args)
	require.Nil(t, err)
	require.Equal(t, "*containers.virtualMachinesContainer", fmt.Sprintf("%T", vmContainer))
	require.Equal(t, "*metachain.vmContainerFactory", fmt.Sprintf("%T", vmFactory))
}
