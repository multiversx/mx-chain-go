package systemSmartContracts

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func createMockArgumentsForLiquidStaking() ArgsNewLiquidStaking {
	return ArgsNewLiquidStaking{
		EpochConfig:            config.EpochConfig{},
		Eei:                    &mock.SystemEIStub{},
		LiquidStakingSCAddress: vm.LiquidStakingSCAddress,
		GasCost:                vm.GasCost{MetaChainSystemSCsCost: vm.MetaChainSystemSCsCost{LiquidStakingOps: 10}},
		Marshalizer:            &mock.MarshalizerMock{},
		Hasher:                 &mock.HasherMock{},
		EpochNotifier:          &mock.EpochNotifierStub{},
	}
}

func createLiquidStakingContractAndEEI() (*liquidStaking, *vmContext) {
	args := createMockArgumentsForLiquidStaking()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{
			CurrentEpochCalled: func() uint32 {
				return 2
			},
		},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&stateMock.AccountsStub{},
		&mock.RaterMock{},
	)
	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}}

	_ = eei.SetSystemSCContainer(systemSCContainerStub)

	args.Eei = eei
	l, _ := NewLiquidStakingSystemSC(args)

	return l, eei
}

func TestLiquidStaking_NilEEI(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForLiquidStaking()
	args.Eei = nil
	_, err := NewLiquidStakingSystemSC(args)
	assert.Equal(t, err, vm.ErrNilSystemEnvironmentInterface)
}

func TestLiquidStaking_NilAddress(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForLiquidStaking()
	args.LiquidStakingSCAddress = nil
	_, err := NewLiquidStakingSystemSC(args)
	assert.True(t, errors.Is(err, vm.ErrInvalidAddress))
}

func TestLiquidStaking_NilMarshalizer(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForLiquidStaking()
	args.Marshalizer = nil
	_, err := NewLiquidStakingSystemSC(args)
	assert.True(t, errors.Is(err, vm.ErrNilMarshalizer))
}

func TestLiquidStaking_NilHasher(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForLiquidStaking()
	args.Hasher = nil
	_, err := NewLiquidStakingSystemSC(args)
	assert.True(t, errors.Is(err, vm.ErrNilHasher))
}

func TestLiquidStaking_NilEpochNotifier(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForLiquidStaking()
	args.EpochNotifier = nil
	l, err := NewLiquidStakingSystemSC(args)
	assert.True(t, errors.Is(err, vm.ErrNilEpochNotifier))
	assert.True(t, l.IsInterfaceNil())
}

func TestLiquidStaking_New(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForLiquidStaking()
	l, err := NewLiquidStakingSystemSC(args)
	assert.Nil(t, err)
	assert.NotNil(t, l)
	assert.False(t, l.IsInterfaceNil())
}

func TestLiquidStaking_CanUseContract(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForLiquidStaking()
	args.EpochConfig.EnableEpochs.BuiltInFunctionOnMetaEnableEpoch = 10
	l, _ := NewLiquidStakingSystemSC(args)
	assert.False(t, l.CanUseContract())

	args.EpochConfig.EnableEpochs.BuiltInFunctionOnMetaEnableEpoch = 0
	l, _ = NewLiquidStakingSystemSC(args)
	assert.True(t, l.CanUseContract())
}

func TestLiquidStaking_SetNewGasConfig(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForLiquidStaking()
	l, _ := NewLiquidStakingSystemSC(args)

	assert.Equal(t, l.gasCost.MetaChainSystemSCsCost.LiquidStakingOps, uint64(10))
	gasCost := vm.GasCost{MetaChainSystemSCsCost: vm.MetaChainSystemSCsCost{LiquidStakingOps: 100}}
	l.SetNewGasCost(gasCost)
	assert.Equal(t, l.gasCost.MetaChainSystemSCsCost.LiquidStakingOps, uint64(100))
}

func TestLiquidStaking_NotActiveWrongCalls(t *testing.T) {
	t.Parallel()

	l, eei := createLiquidStakingContractAndEEI()

	returnCode := l.Execute(nil)
	assert.Equal(t, returnCode, vmcommon.UserError)
	assert.Equal(t, eei.returnMessage, vm.ErrInputArgsIsNil.Error())

	l.flagLiquidStaking.Unset()
	eei.returnMessage = ""
	vmInput := getDefaultVmInputForFunc("returnViaLiquidStaking", make([][]byte, 0))
	returnCode = l.Execute(vmInput)
	assert.Equal(t, returnCode, vmcommon.UserError)
	assert.Equal(t, eei.returnMessage, "liquid staking contract is not enabled")

	l.flagLiquidStaking.Set()
	eei.returnMessage = ""
	returnCode = l.Execute(vmInput)
	assert.Equal(t, returnCode, vmcommon.UserError)
	assert.Equal(t, eei.returnMessage, vmInput.Function+" is an unknown function")
}

func TestLiquidStaking_init(t *testing.T) {
	t.Parallel()

	l, eei := createLiquidStakingContractAndEEI()
	vmInput := getDefaultVmInputForFunc(core.SCDeployInitFunctionName, make([][]byte, 0))

	eei.returnMessage = ""
	returnCode := l.Execute(vmInput)
	assert.Equal(t, returnCode, vmcommon.UserError)
	assert.Equal(t, eei.returnMessage, "invalid caller")

	eei.returnMessage = ""
	vmInput.CallerAddr = vm.LiquidStakingSCAddress
	vmInput.CallValue = big.NewInt(10)
	returnCode = l.Execute(vmInput)
	assert.Equal(t, returnCode, vmcommon.UserError)
	assert.Equal(t, eei.returnMessage, "function is not payable in eGLD")

	eei.returnMessage = ""
	vmInput.CallValue = big.NewInt(0)
	returnCode = l.Execute(vmInput)
	assert.Equal(t, returnCode, vmcommon.UserError)
	assert.Equal(t, eei.returnMessage, "invalid number of arguments")

	vmInput.Arguments = append(vmInput.Arguments, []byte("tokenID"))
	eei.returnMessage = ""
	returnCode = l.Execute(vmInput)
	assert.Equal(t, returnCode, vmcommon.Ok)
	assert.Equal(t, l.getTokenID(), []byte("tokenID"))
}
