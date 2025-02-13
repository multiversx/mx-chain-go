package hooks

import (
	"fmt"
	"testing"

	vmcommonBuiltInFunctions "github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	storageMock "github.com/multiversx/mx-chain-go/testscommon/storage"
)

func TestNewBlockChainHookFactory(t *testing.T) {
	t.Parallel()

	factory := NewBlockChainHookFactory()
	require.NotNil(t, factory)
}

func TestBlockChainHookFactory_CreateBlockChainHook(t *testing.T) {
	t.Parallel()

	factory := NewBlockChainHookFactory()
	blockChainHook, err := factory.CreateBlockChainHookHandler(getDefaultArgs())
	require.Equal(t, "*hooks.BlockChainHookImpl", fmt.Sprintf("%T", blockChainHook))
	require.Nil(t, err)
}

func TestBlockChainHookFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	factory := NewBlockChainHookFactory()
	require.False(t, factory.IsInterfaceNil())

	factory = (*blockChainHookFactory)(nil)
	require.True(t, factory.IsInterfaceNil())
}

func getDefaultArgs() ArgBlockChainHook {
	return ArgBlockChainHook{
		Accounts:              &state.AccountsStub{},
		PubkeyConv:            &testscommon.PubkeyConverterMock{},
		StorageService:        &storageMock.ChainStorerStub{},
		BlockChain:            &testscommon.ChainHandlerStub{},
		ShardCoordinator:      &testscommon.ShardsCoordinatorMock{},
		Marshalizer:           &testscommon.ProtoMarshalizerMock{},
		Uint64Converter:       &testscommon.Uint64ByteSliceConverterStub{},
		BuiltInFunctions:      vmcommonBuiltInFunctions.NewBuiltInFunctionContainer(),
		NFTStorageHandler:     &testscommon.SimpleNFTStorageHandlerStub{},
		GlobalSettingsHandler: &testscommon.ESDTGlobalSettingsHandlerStub{},
		DataPool:              &dataRetriever.PoolsHolderMock{},
		CompiledSCPool:        &testscommon.CacherStub{},
		EpochNotifier:         &epochNotifier.EpochNotifierStub{},
		EnableEpochsHandler:   &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		NilCompiledSCStore:    true,
		GasSchedule: &testscommon.GasScheduleNotifierMock{
			LatestGasScheduleCalled: func() map[string]map[string]uint64 {
				return make(map[string]map[string]uint64)
			},
		},
		Counter:                  &testscommon.BlockChainHookCounterStub{},
		MissingTrieNodesNotifier: &testscommon.MissingTrieNodesNotifierStub{},
	}
}
