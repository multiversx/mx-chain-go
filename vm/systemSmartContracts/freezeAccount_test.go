package systemSmartContracts

import (
	"math/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	mock2 "github.com/ElrondNetwork/elrond-vm-common/mock"
	"github.com/stretchr/testify/require"
)

func TestFreezeAccount_ExecuteSetGuardianCase1(t *testing.T) {
	args := createAccountFreezerMockArgs()
	userAccount := mock2.NewUserAccount([]byte("d"))

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
	guardianAddress := generateRandomByteArray(32)

	vmInput := getDefaultVmInputForFunc(setGuardian, [][]byte{guardianAddress})

	output := accountFreezer.Execute(vmInput)
	require.Empty(t, systemEI.ReturnMessage)
	require.Equal(t, vmcommon.Ok, output)
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
