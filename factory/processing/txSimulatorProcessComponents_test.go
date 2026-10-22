package processing_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/factory/processing"
	processFactory "github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/components"
	factoryMock "github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimulationMetaVM_OwnerSaveDoesNotChangeLiveAccounts(t *testing.T) {
	coordinator := mock.NewMultiShardsCoordinatorMock(3)
	coordinator.CurrentShard = core.MetachainShardId
	args := components.GetProcessComponentsFactoryArgs(coordinator)
	liveAccounts := args.State.AccountsAdapter()
	contractAddress := bytes.Clone(vm.FirstDelegationSCAddress)
	contractAddress[28]++
	oldOwner := bytes.Repeat([]byte{1}, 32)
	newOwner := bytes.Repeat([]byte{2}, 32)
	account, err := liveAccounts.LoadAccount(contractAddress)
	require.NoError(t, err)
	userAccount := account.(state.UserAccountHandler)
	userAccount.SetOwnerAddress(oldOwner)
	userAccount.SetCode(vm.FirstDelegationSCAddress)
	require.NoError(t, userAccount.SaveKeyValue([]byte("owner"), newOwner))
	require.NoError(t, liveAccounts.SaveAccount(userAccount))
	root, err := liveAccounts.Commit()
	require.NoError(t, err)

	stateComponents := factoryMock.NewStateComponentsMockFromRealComponent(args.State)
	stateComponents.AccountsAPI = liveAccounts
	args.State = stateComponents
	pcf, err := processing.NewProcessComponentsFactory(args)
	require.NoError(t, err)
	_, vmFactory, err := pcf.CreateAPITransactionEvaluator(&testscommon.EpochStartTriggerStub{})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, vmFactory.Close()) })
	container, err := vmFactory.Create()
	require.NoError(t, err)
	systemVM, err := container.Get(processFactory.SystemVirtualMachine)
	require.NoError(t, err)
	output, err := systemVM.RunSmartContractCall(&vmcommon.ContractCallInput{
		RecipientAddr: contractAddress,
		Function:      "synchronizeOwner",
		VMInput:       vmcommon.VMInput{CallerAddr: oldOwner, CallValue: big.NewInt(0), GasProvided: 100000000},
	})
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, output.ReturnCode, output.ReturnMessage)
	simulatedAccount, err := vmFactory.BlockChainHookImpl().GetUserAccount(contractAddress)
	require.NoError(t, err)
	require.Equal(t, newOwner, simulatedAccount.GetOwnerAddress())
	account, err = liveAccounts.GetExistingAccount(contractAddress)
	require.NoError(t, err)
	require.Equal(t, oldOwner, account.(state.UserAccountHandler).GetOwnerAddress())
	require.Zero(t, liveAccounts.JournalLen())
	currentRoot, err := liveAccounts.RootHash()
	require.NoError(t, err)
	require.Equal(t, root, currentRoot)
}

func TestManagedProcessComponents_createAPITransactionEvaluator(t *testing.T) {
	t.Parallel()

	shardCoordinatorForShardID2 := mock.NewMultiShardsCoordinatorMock(3)
	shardCoordinatorForShardID2.CurrentShard = 2

	shardCoordinatorForMetachain := mock.NewMultiShardsCoordinatorMock(3)
	shardCoordinatorForMetachain.CurrentShard = core.MetachainShardId

	// no further t.Parallel as these tests are quite heavy (they open netMessengers and other components that start a lot of goroutines)
	t.Run("invalid VMOutputCacher config should error", func(t *testing.T) {
		processArgs := components.GetProcessComponentsFactoryArgs(shardCoordinatorForShardID2)
		processArgs.Config.VMOutputCacher.Type = "invalid"
		pcf, err := processing.NewProcessComponentsFactory(processArgs)
		require.Nil(t, err)

		apiTransactionEvaluator, vmContainerFactory, err := pcf.CreateAPITransactionEvaluator(&testscommon.EpochStartTriggerStub{})
		assert.NotNil(t, err)
		assert.True(t, check.IfNil(apiTransactionEvaluator))
		assert.True(t, check.IfNil(vmContainerFactory))
		assert.Contains(t, err.Error(), "not supported cache type")
	})
	t.Run("should work for shard", func(t *testing.T) {
		processArgs := components.GetProcessComponentsFactoryArgs(shardCoordinatorForShardID2)
		pcf, err := processing.NewProcessComponentsFactory(processArgs)
		require.Nil(t, err)

		apiTransactionEvaluator, vmContainerFactory, err := pcf.CreateAPITransactionEvaluator(&testscommon.EpochStartTriggerStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(apiTransactionEvaluator))
		assert.False(t, check.IfNil(vmContainerFactory))
		require.NoError(t, vmContainerFactory.Close())
	})
	t.Run("should work for metachain", func(t *testing.T) {
		processArgs := components.GetProcessComponentsFactoryArgs(shardCoordinatorForMetachain)
		pcf, err := processing.NewProcessComponentsFactory(processArgs)
		require.Nil(t, err)

		apiTransactionEvaluator, vmContainerFactory, err := pcf.CreateAPITransactionEvaluator(&testscommon.EpochStartTriggerStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(apiTransactionEvaluator))
		assert.False(t, check.IfNil(vmContainerFactory))
		require.NoError(t, vmContainerFactory.Close())
	})
}
