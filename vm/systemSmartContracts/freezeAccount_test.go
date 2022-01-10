package systemSmartContracts

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	mockvm "github.com/ElrondNetwork/elrond-vm-common/mock"
	"github.com/stretchr/testify/require"
)

func TestFreezeAccount_ExecuteSetGuardianCase1(t *testing.T) {
	userAccount := mockvm.NewUserAccount([]byte("user address"))
	blockChainHook := &mock.BlockChainHookStub{
		GetUserAccountCalled: func(address []byte) (vmcommon.UserAccountHandler, error) {
			return userAccount, nil
		},
	}
	systemEI := &mock.SystemEIStub{
		BlockChainHookCalled: func() vm.BlockchainHook {
			return blockChainHook
		},
	}

	args := createAccountFreezerMockArgs()
	args.SystemEI = systemEI
	accountFreezer, _ := NewFreezeAccountSmartContract(args)

	guardianAddress := generateRandomByteArray(32)
	vmInput := getDefaultVmInputForFunc(setGuardian, [][]byte{guardianAddress})

	output := accountFreezer.Execute(vmInput)
	require.Empty(t, systemEI.ReturnMessage)
	require.Equal(t, vmcommon.Ok, output)

	key := append([]byte(core.ElrondProtectedKeyPrefix), []byte(GuardiansKey)...)
	marshalledData, _ := userAccount.AccountDataHandler().RetrieveValue(key)
	storedGuardian := &Guardians{}
	_ = args.Marshaller.Unmarshal(storedGuardian, marshalledData)

	expectedStoredGuardian := &Guardian{
		Address:         guardianAddress,
		ActivationEpoch: blockChainHook.CurrentEpoch() + args.AccountFreezerConfig.GuardianEnableEpochs,
	}
	require.Len(t, storedGuardian.Data, 1)
	require.Equal(t, storedGuardian.Data[0], expectedStoredGuardian)
}

func TestFreezeAccount_ExecuteSetGuardianCase2(t *testing.T) {
	userAccount := mockvm.NewUserAccount([]byte("user address"))
	key := append([]byte(core.ElrondProtectedKeyPrefix), []byte(GuardiansKey)...)
	guardianAddress := generateRandomByteArray(32)

	args := createAccountFreezerMockArgs()

	pendingGuardian := Guardians{
		Data: []*Guardian{
			{
				Address:         guardianAddress,
				ActivationEpoch: args.SystemEI.BlockChainHook().CurrentEpoch() + args.AccountFreezerConfig.GuardianEnableEpochs - 1,
			},
		},
	}
	marshalledPendingGuardian, _ := args.Marshaller.Marshal(pendingGuardian)
	_ = userAccount.SaveKeyValue(key, marshalledPendingGuardian)

	blockChainHook := &mock.BlockChainHookStub{
		GetUserAccountCalled: func(address []byte) (vmcommon.UserAccountHandler, error) {
			return userAccount, nil
		},
	}

	systemEI := &mock.SystemEIStub{
		BlockChainHookCalled: func() vm.BlockchainHook {
			return blockChainHook
		},
	}

	args.SystemEI = systemEI
	accountFreezer, _ := NewFreezeAccountSmartContract(args)

	vmInput := getDefaultVmInputForFunc(setGuardian, [][]byte{guardianAddress})

	output := accountFreezer.Execute(vmInput)
	require.True(t, strings.Contains(systemEI.ReturnMessage, "owner already has one guardian"))
	require.Equal(t, vmcommon.UserError, output)

	marshalledData, _ := userAccount.AccountDataHandler().RetrieveValue(key)
	require.Equal(t, marshalledPendingGuardian, marshalledData)
}

func TestFreezeAccount_ExecuteSetGuardianCase3(t *testing.T) {
	userAccount := mockvm.NewUserAccount([]byte("user address"))
	key := append([]byte(core.ElrondProtectedKeyPrefix), []byte(GuardiansKey)...)
	guardianAddress := generateRandomByteArray(32)
	enabledGuardianAddress := generateRandomByteArray(32)

	args := createAccountFreezerMockArgs()

	enabledGuardian := Guardians{
		Data: []*Guardian{
			{
				Address:         enabledGuardianAddress,
				ActivationEpoch: args.SystemEI.BlockChainHook().CurrentEpoch() + args.AccountFreezerConfig.GuardianEnableEpochs + 1,
			},
		},
	}
	marshalledEnabledGuardian, _ := args.Marshaller.Marshal(enabledGuardian)
	_ = userAccount.SaveKeyValue(key, marshalledEnabledGuardian)

	blockChainHook := &mock.BlockChainHookStub{
		GetUserAccountCalled: func(address []byte) (vmcommon.UserAccountHandler, error) {
			return userAccount, nil
		},
	}

	systemEI := &mock.SystemEIStub{
		BlockChainHookCalled: func() vm.BlockchainHook {
			return blockChainHook
		},
	}

	args.SystemEI = systemEI
	accountFreezer, _ := NewFreezeAccountSmartContract(args)

	vmInput := getDefaultVmInputForFunc(setGuardian, [][]byte{guardianAddress})

	output := accountFreezer.Execute(vmInput)
	require.Empty(t, systemEI.ReturnMessage)
	require.Equal(t, vmcommon.Ok, output)

	marshalledData, _ := userAccount.AccountDataHandler().RetrieveValue(key)
	guardians := Guardians{}
	_ = args.Marshaller.Unmarshal(&guardians, marshalledData)
	expectedStoredGuardians := Guardians{
		Data: []*Guardian{
			{
				Address:         enabledGuardianAddress,
				ActivationEpoch: args.SystemEI.BlockChainHook().CurrentEpoch() + args.AccountFreezerConfig.GuardianEnableEpochs + 1,
			},
			{
				Address:         guardianAddress,
				ActivationEpoch: args.SystemEI.BlockChainHook().CurrentEpoch() + args.AccountFreezerConfig.GuardianEnableEpochs + args.EpochConfig.EnableEpochs.AccountFreezerEnableEpoch,
			},
		},
	}

	require.Equal(t, expectedStoredGuardians, guardians)
}

func TestFreezeAccount_ExecuteSetGuardianCase4(t *testing.T) {
	userAccount := mockvm.NewUserAccount([]byte("user address"))
	key := append([]byte(core.ElrondProtectedKeyPrefix), []byte(GuardiansKey)...)
	guardianAddress := generateRandomByteArray(32)
	enabledGuardianAddress1 := generateRandomByteArray(32)
	enabledGuardianAddress2 := generateRandomByteArray(32)

	args := createAccountFreezerMockArgs()

	storedGuardians := Guardians{
		Data: []*Guardian{
			{
				Address:         enabledGuardianAddress1,
				ActivationEpoch: args.SystemEI.BlockChainHook().CurrentEpoch() + args.AccountFreezerConfig.GuardianEnableEpochs + 1,
			},
			{
				Address:         enabledGuardianAddress2,
				ActivationEpoch: args.SystemEI.BlockChainHook().CurrentEpoch() + args.AccountFreezerConfig.GuardianEnableEpochs - 1,
			},
		},
	}
	marshalledStoredGuardians, _ := args.Marshaller.Marshal(storedGuardians)
	_ = userAccount.SaveKeyValue(key, marshalledStoredGuardians)

	blockChainHook := &mock.BlockChainHookStub{
		GetUserAccountCalled: func(address []byte) (vmcommon.UserAccountHandler, error) {
			return userAccount, nil
		},
	}

	systemEI := &mock.SystemEIStub{
		BlockChainHookCalled: func() vm.BlockchainHook {
			return blockChainHook
		},
	}

	args.SystemEI = systemEI
	accountFreezer, _ := NewFreezeAccountSmartContract(args)

	vmInput := getDefaultVmInputForFunc(setGuardian, [][]byte{guardianAddress})

	output := accountFreezer.Execute(vmInput)
	require.True(t, strings.Contains(systemEI.ReturnMessage, "owner already has one guardian"))
	require.Equal(t, vmcommon.UserError, output)

	marshalledData, _ := userAccount.AccountDataHandler().RetrieveValue(key)
	require.Equal(t, marshalledStoredGuardians, marshalledData)
}

func TestFreezeAccount_ExecuteSetGuardianCase5(t *testing.T) {
	userAccount := mockvm.NewUserAccount([]byte("user address"))
	key := append([]byte(core.ElrondProtectedKeyPrefix), []byte(GuardiansKey)...)
	guardianAddress := generateRandomByteArray(32)
	enabledGuardianAddress1 := generateRandomByteArray(32)
	enabledGuardianAddress2 := generateRandomByteArray(32)

	args := createAccountFreezerMockArgs()

	storedGuardians := Guardians{
		Data: []*Guardian{
			{
				Address:         enabledGuardianAddress1,
				ActivationEpoch: args.SystemEI.BlockChainHook().CurrentEpoch() + args.AccountFreezerConfig.GuardianEnableEpochs + 1,
			},
			{
				Address:         enabledGuardianAddress2,
				ActivationEpoch: args.SystemEI.BlockChainHook().CurrentEpoch() + args.AccountFreezerConfig.GuardianEnableEpochs + 1,
			},
		},
	}
	marshalledStoredGuardians, _ := args.Marshaller.Marshal(storedGuardians)
	_ = userAccount.SaveKeyValue(key, marshalledStoredGuardians)

	blockChainHook := &mock.BlockChainHookStub{
		GetUserAccountCalled: func(address []byte) (vmcommon.UserAccountHandler, error) {
			return userAccount, nil
		},
	}

	systemEI := &mock.SystemEIStub{
		BlockChainHookCalled: func() vm.BlockchainHook {
			return blockChainHook
		},
	}

	args.SystemEI = systemEI
	accountFreezer, _ := NewFreezeAccountSmartContract(args)

	vmInput := getDefaultVmInputForFunc(setGuardian, [][]byte{guardianAddress})

	output := accountFreezer.Execute(vmInput)
	require.Empty(t, systemEI.ReturnMessage)
	require.Equal(t, vmcommon.Ok, output)

	marshalledData, _ := userAccount.AccountDataHandler().RetrieveValue(key)
	guardians := Guardians{}
	_ = args.Marshaller.Unmarshal(&guardians, marshalledData)
	expectedStoredGuardians := Guardians{
		Data: []*Guardian{
			{
				Address:         enabledGuardianAddress2,
				ActivationEpoch: args.SystemEI.BlockChainHook().CurrentEpoch() + args.AccountFreezerConfig.GuardianEnableEpochs + 1,
			},
			{
				Address:         guardianAddress,
				ActivationEpoch: args.SystemEI.BlockChainHook().CurrentEpoch() + args.AccountFreezerConfig.GuardianEnableEpochs + args.EpochConfig.EnableEpochs.AccountFreezerEnableEpoch,
			},
		},
	}

	require.Equal(t, expectedStoredGuardians, guardians)
}

// TODO: Remove this from all duplicate places
func generateRandomByteArray(size int) []byte {
	r := make([]byte, size)
	_, _ = rand.Read(r)
	return r
}

func createAccountFreezerMockArgs() ArgsFreezeAccountSC {

	return ArgsFreezeAccountSC{
		GasCost:         vm.GasCost{},
		Marshaller:      testscommon.MarshalizerMock{},
		SystemEI:        &mock.SystemEIStub{},
		PubKeyConverter: testscommon.NewPubkeyConverterMock(32),
		AccountFreezerConfig: config.AccountFreezerSCConfig{
			GuardianEnableEpochs: 100,
		},
		EpochConfig: config.EpochConfig{
			EnableEpochs: config.EnableEpochs{
				AccountFreezerEnableEpoch: 0,
			},
		},
		EpochNotifier: &mock.EpochNotifierStub{},
	}
}
