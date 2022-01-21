package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	vmData "github.com/ElrondNetwork/elrond-go-core/data/vm"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/state"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockStakingScArgumentsWithSystemScAddresses(
	validatorScAddress []byte,
	jailScAddress []byte,
	endOfEpochAddress []byte,
) ArgsNewStakingSmartContract {
	return ArgsNewStakingSmartContract{
		Eei:                  &mock.SystemEIStub{},
		StakingAccessAddr:    validatorScAddress,
		JailAccessAddr:       jailScAddress,
		EndOfEpochAccessAddr: endOfEpochAddress,
		MinNumNodes:          1,
		Marshalizer:          &mock.MarshalizerMock{},
		StakingSCConfig: config.StakingSystemSCConfig{
			GenesisNodePrice:                     "100",
			MinStakeValue:                        "1",
			UnJailValue:                          "1",
			MinStepValue:                         "1",
			UnBondPeriod:                         0,
			NumRoundsWithoutBleed:                0,
			MaximumPercentageToBleed:             0,
			BleedPercentagePerRound:              0,
			MaxNumberOfNodesForStake:             10,
			ActivateBLSPubKeyMessageVerification: false,
			MinUnstakeTokensValue:                "1",
		},
		EpochNotifier: &mock.EpochNotifierStub{},
		EpochConfig: config.EpochConfig{
			EnableEpochs: config.EnableEpochs{
				StakingV2EnableEpoch: 10,
				StakeEnableEpoch:     0,
			},
		},
	}
}

func createMockStakingScArguments() ArgsNewStakingSmartContract {
	return createMockStakingScArgumentsWithSystemScAddresses(
		[]byte("validator"),
		[]byte("jail"),
		[]byte("endOfEpoch"),
	)
}

func createMockStakingScArgumentsWithRealSystemScAddresses() ArgsNewStakingSmartContract {
	return createMockStakingScArgumentsWithSystemScAddresses(
		vm.ValidatorSCAddress,
		vm.JailingAddress,
		vm.EndOfEpochAddress,
	)
}

func CreateVmContractCallInput() *vmcommon.ContractCallInput {
	return &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("validator"),
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    0,
			GasProvided: 0,
			CallType:    vmData.DirectCall,
		},
		RecipientAddr: []byte("rcpntaddr"),
		Function:      "something",
	}
}

func TestNewStakingSmartContract_NilSystemEIShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockStakingScArguments()
	args.Eei = nil
	stakingSmartContract, err := NewStakingSmartContract(args)

	assert.Nil(t, stakingSmartContract)
	assert.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
}

func TestNewStakingSmartContract_NilStakingAccessAddrEIShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockStakingScArguments()
	args.StakingAccessAddr = nil
	stakingSmartContract, err := NewStakingSmartContract(args)

	assert.Nil(t, stakingSmartContract)
	assert.Equal(t, vm.ErrInvalidStakingAccessAddress, err)
}

func TestNewStakingSmartContract_NilJailAccessAddrEIShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockStakingScArguments()
	args.JailAccessAddr = nil
	stakingSmartContract, err := NewStakingSmartContract(args)

	assert.Nil(t, stakingSmartContract)
	assert.Equal(t, vm.ErrInvalidJailAccessAddress, err)
}

func TestNewStakingSmartContract(t *testing.T) {
	t.Parallel()

	args := createMockStakingScArguments()
	stakingSmartContract, err := NewStakingSmartContract(args)

	assert.False(t, check.IfNil(stakingSmartContract))
	assert.Nil(t, err)
}

func TestStakingSC_ExecuteInit(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&stateMock.AccountsStub{},
		&mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = core.SCDeployInitFunctionName

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	ownerAddr := stakingSmartContract.eei.GetStorage([]byte(ownerKey))
	assert.Equal(t, arguments.CallerAddr, ownerAddr)

	ownerBalanceBytes := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	ownerBalance := big.NewInt(0).SetBytes(ownerBalanceBytes)
	assert.Equal(t, big.NewInt(0), ownerBalance)

}

func TestStakingSC_ExecuteInitTwoTimeShouldReturnUserError(t *testing.T) {
	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = core.SCDeployInitFunctionName

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteStakeWrongStakeValueShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{
		GetUserAccountCalled: func(address []byte) (vmcommon.UserAccountHandler, error) {
			return nil, state.ErrAccNotFound
		},
	}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)

	balance := eei.GetBalance(arguments.CallerAddr)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestStakingSC_ExecuteStakeWrongUnmarshalDataShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteStakeRegistrationDataStakedShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		registrationDataMarshalized, _ := json.Marshal(
			&StakedDataV2_0{
				Staked: true,
			})
		return registrationDataMarshalized
	}
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteStakeNotEnoughArgsShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		registrationDataMarshalized, _ := json.Marshal(&StakedDataV2_0{})
		return registrationDataMarshalized
	}
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteStake(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)
	expectedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: []byte{100},
		StakeValue:    big.NewInt(100),
		JailedRound:   math.MaxUint64,
		UnStakedEpoch: common.DefaultUnstakedEpoch,
		SlashValue:    big.NewInt(0),
	}

	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = []byte("validator")
	arguments.Arguments = [][]byte{stakerPubKey.Bytes(), stakerAddress.Bytes(), stakerAddress.Bytes()}
	arguments.CallValue = big.NewInt(100)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData StakedDataV2_0
	data := stakingSmartContract.eei.GetStorage(stakerPubKey.Bytes())
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)
	assert.Equal(t, expectedRegistrationData, registrationData)
}

func TestStakingSC_ExecuteUnStakeAddressNotStakedShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake@abc"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteUnStakeUnmarshalErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake@abc"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteUnStakeAlreadyUnStakedAddrShouldErr(t *testing.T) {
	t.Parallel()

	stakedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		RewardAddress: nil,
		StakeValue:    nil,
	}

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes(), big.NewInt(200).Bytes()}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.CallerAddr, marshalizedExpectedRegData)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteUnStakeFailsWithWrongCaller(t *testing.T) {
	t.Parallel()

	expectedCallerAddress := []byte("caller")
	wrongCallerAddress := []byte("wrongCaller")

	stakedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: expectedCallerAddress,
		StakeValue:    nil,
	}

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{wrongCallerAddress}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteUnStakeShouldErrorNotEnoughNodes(t *testing.T) {
	t.Parallel()

	callerAddress := []byte("caller")

	expectedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		RewardAddress: callerAddress,
		StakeValue:    nil,
		JailedRound:   math.MaxUint64,
		SlashValue:    big.NewInt(0),
	}

	stakedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: callerAddress,
		StakeValue:    nil,
		JailedRound:   math.MaxUint64,
	}

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	args.MinNumNodes = 1
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{[]byte("abc"), callerAddress}
	arguments.CallerAddr = []byte("validator")
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)
	stakingSmartContract.setConfig(&StakingNodesConfig{MinNumNodes: 5, StakedNodes: 10})

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData StakedDataV2_0
	data := stakingSmartContract.eei.GetStorage(arguments.Arguments[0])
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)
	assert.Equal(t, expectedRegistrationData, registrationData)
}

func TestStakingSC_ExecuteUnStake(t *testing.T) {
	t.Parallel()

	callerAddress := []byte("caller")

	expectedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		RewardAddress: callerAddress,
		StakeValue:    nil,
		JailedRound:   math.MaxUint64,
		SlashValue:    big.NewInt(0),
	}

	stakedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: callerAddress,
		StakeValue:    nil,
		JailedRound:   math.MaxUint64,
	}

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{[]byte("abc"), callerAddress}
	arguments.CallerAddr = []byte("validator")
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)
	stakingSmartContract.setConfig(&StakingNodesConfig{MinNumNodes: 5, StakedNodes: 10})

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData StakedDataV2_0
	data := stakingSmartContract.eei.GetStorage(arguments.Arguments[0])
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)
	assert.Equal(t, expectedRegistrationData, registrationData)
}

func TestStakingSC_ExecuteUnBoundUnmarshalErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes(), big.NewInt(200).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteUnBoundValidatorNotUnStakeShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		switch {
		case bytes.Equal(key, []byte(ownerKey)):
			return []byte("data")
		default:
			registrationDataMarshalized, _ := json.Marshal(
				&StakedDataV2_0{
					UnStakedNonce: 0,
				})
			return registrationDataMarshalized
		}
	}
	eei.BlockChainHookCalled = func() vm.BlockchainHook {
		return &mock.BlockChainHookStub{CurrentNonceCalled: func() uint64 {
			return 10000
		}}
	}
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteFinalizeUnBoundBeforePeriodEnds(t *testing.T) {
	t.Parallel()

	unstakedNonce := uint64(10)
	registrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: unstakedNonce,
		RewardAddress: nil,
		StakeValue:    big.NewInt(100),
	}
	blsPubKey := big.NewInt(100)
	stakeValue := big.NewInt(100)
	marshalizedRegData, _ := json.Marshal(&registrationData)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return unstakedNonce + 1
		},
	}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))
	eei.SetStorage([]byte(ownerKey), []byte("data"))
	eei.SetStorage(blsPubKey.Bytes(), marshalizedRegData)
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "finalizeUnStake"
	arguments.Arguments = [][]byte{blsPubKey.Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteUnBoundStillValidator(t *testing.T) {
	t.Parallel()

	unBondPeriod := uint64(100)
	unstakedNonce := uint64(10)
	registrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: unstakedNonce,
		RewardAddress: []byte("validator"),
		StakeValue:    big.NewInt(100),
		JailedRound:   math.MaxUint64,
	}

	peerAccount := state.NewEmptyPeerAccount()
	peerAccount.List = string(common.EligibleList)
	stakeValue := big.NewInt(100)
	marshalizedRegData, _ := json.Marshal(&registrationData)
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{
			CurrentNonceCalled: func() uint64 {
				return unstakedNonce + unBondPeriod + 1
			},
		},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&stateMock.AccountsStub{
			GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return peerAccount, nil
			}},
		&mock.RaterMock{})
	scAddress := []byte("owner")
	eei.SetSCAddress(scAddress)
	eei.SetStorage([]byte(ownerKey), scAddress)

	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("validator")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{[]byte("abc")}

	stakingSmartContract.eei.SetStorage(arguments.Arguments[0], marshalizedRegData)
	stakingSmartContract.setConfig(&StakingNodesConfig{MinNumNodes: 5, StakedNodes: 10})

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteUnBound(t *testing.T) {
	t.Parallel()

	unBondPeriod := uint64(100)
	unstakedNonce := uint64(10)
	registrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: unstakedNonce,
		RewardAddress: []byte("validator"),
		StakeValue:    big.NewInt(100),
		JailedRound:   math.MaxUint64,
	}

	stakeValue := big.NewInt(100)
	marshalizedRegData, _ := json.Marshal(&registrationData)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return unstakedNonce + unBondPeriod + 1
		},
	}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	scAddress := []byte("owner")
	eei.SetSCAddress(scAddress)
	eei.SetStorage([]byte(ownerKey), scAddress)

	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("validator")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{[]byte("abc")}

	stakingSmartContract.eei.SetStorage(arguments.Arguments[0], marshalizedRegData)
	stakingSmartContract.setConfig(&StakingNodesConfig{MinNumNodes: 5, StakedNodes: 10})

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	data := stakingSmartContract.eei.GetStorage(arguments.Arguments[0])
	assert.Equal(t, 0, len(data))
}

func TestStakingSC_ExecuteSlashOwnerAddrNotOkShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteSlashArgumentsNotOkShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"
	arguments.CallerAddr = []byte("data")

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteSlashUnmarhsalErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"
	arguments.CallerAddr = []byte("data")
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes(), big.NewInt(100).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteSlashNotStake(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		switch {
		case bytes.Equal(key, []byte(ownerKey)):
			return []byte("data")
		default:
			registrationDataMarshalized, _ := json.Marshal(
				&StakedDataV2_0{
					StakeValue: big.NewInt(100),
				})
			return registrationDataMarshalized
		}
	}

	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"
	arguments.CallerAddr = []byte("data")
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes(), big.NewInt(100).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteUnStakeAndUnBoundStake(t *testing.T) {
	t.Parallel()

	// Preparation
	unBondPeriod := uint64(100)
	stakeValue := big.NewInt(100)
	valueStakedByTheCaller := big.NewInt(100)
	stakerAddress := []byte("address")
	stakerPubKey := []byte("pubKey")
	blockChainHook := &mock.BlockChainHookStub{}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})

	smartcontractAddress := "smartcontractAddress"
	eei.SetSCAddress([]byte(smartcontractAddress))

	ownerAddress := "ownerAddress"
	eei.SetStorage([]byte(ownerKey), []byte(ownerAddress))

	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)
	stakingSmartContract.setConfig(&StakingNodesConfig{MinNumNodes: 5, StakedNodes: 10})

	arguments := CreateVmContractCallInput()
	arguments.Arguments = [][]byte{stakerPubKey, stakerAddress}
	arguments.CallerAddr = []byte("validator")

	stakedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: stakerAddress,
		StakeValue:    valueStakedByTheCaller,
		JailedRound:   math.MaxUint64,
		SlashValue:    big.NewInt(0),
	}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)

	arguments.Function = "unStake"

	unStakeNonce := uint64(10)
	blockChainHook.CurrentNonceCalled = func() uint64 {
		return unStakeNonce
	}
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData StakedDataV2_0
	data := stakingSmartContract.eei.GetStorage(arguments.Arguments[0])
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)

	expectedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: unStakeNonce,
		RewardAddress: stakerAddress,
		StakeValue:    valueStakedByTheCaller,
		JailedRound:   math.MaxUint64,
		SlashValue:    big.NewInt(0),
	}
	assert.Equal(t, expectedRegistrationData, registrationData)

	arguments.Function = "unBond"

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return unStakeNonce + unBondPeriod + 1
	}
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestStakingSC_ExecuteGetShouldReturnUserErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	arguments := CreateVmContractCallInput()
	arguments.Function = "get"
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	err := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.UserError, err)
}

func TestStakingSC_ExecuteGetShouldOk(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	arguments := CreateVmContractCallInput()
	arguments.Function = "get"
	arguments.Arguments = [][]byte{arguments.CallerAddr}
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	err := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.Ok, err)
}

func TestStakingSc_ExecuteNilArgs(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})

	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	retCode := stakingSmartContract.Execute(nil)
	assert.Equal(t, vmcommon.UserError, retCode)
}

// TestStakingSc_ExecuteIsStaked
// Will test next behaviour
// 1 - will execute function isStaked should return UserError
// 2 - will execute function stake and after that will call function isStaked and will return Ok
// 3 - will execute function unStake and after that will cal function isStaked and will return UserError
func TestStakingSc_ExecuteIsStaked(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.StakingAccessAddr = stakingAccessAddress
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	stakerAddress := []byte("stakerAddr")
	stakerPubKey := []byte("stakerPublicKey")
	callerAddress := []byte("data")

	// check if account is staked should return error code
	checkIsStaked(t, stakingSmartContract, callerAddress, nil, vmcommon.UserError)
	// check if account is staked should return error code
	checkIsStaked(t, stakingSmartContract, callerAddress, stakerPubKey, vmcommon.UserError)
	// do stake should work
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("anotherKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, stakerPubKey)
	// check again isStaked should return vmcommon.Ok
	checkIsStaked(t, stakingSmartContract, callerAddress, stakerPubKey, vmcommon.Ok)
	//do unStake
	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, stakerPubKey, vmcommon.Ok)
	// check if account is staked should return error code
	checkIsStaked(t, stakingSmartContract, callerAddress, stakerPubKey, vmcommon.UserError)
}

func TestStakingSc_StakeWithV1ShouldWork(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	jailAccessAddr := []byte("jailAccessAddr")
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.EpochConfig.EnableEpochs.StakeEnableEpoch = 10
	args.StakingAccessAddr = stakingAccessAddress
	args.Eei = eei
	args.StakingSCConfig.NumRoundsWithoutBleed = 100
	args.StakingSCConfig.MaximumPercentageToBleed = 0.5
	args.StakingSCConfig.BleedPercentagePerRound = 0.00001
	args.JailAccessAddr = jailAccessAddr
	args.MinNumNodes = 0
	stakingSmartContract, _ := NewStakingSmartContract(args)

	stakerAddress := []byte("stakerAddr")
	stakerPubKey := []byte("stakerPublicKey")

	//do stake should work
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, stakerPubKey)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 11
	}

	//do unStake with V2 should work
	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, stakerPubKey, vmcommon.Ok)
}

// Test scenario
// 1 -- will set stake value for current epoch should work
// 2 -- will try to do jail before stake should return user error
// 3 -- will stake and stake should work
// 4 -- will jail user that stake and should work
// 5 -- will try do to unStake and should not work because cannot do unStake if validator is jail
// 6 -- will try to do unJail with wrong access address should not work
// 7 -- will do unJail with correct parameters and should work and after that stakeValue should be 999
func TestStakingSc_StakeJailAndUnJail(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	jailAccessAddr := []byte("jailAccessAddr")
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.StakingAccessAddr = stakingAccessAddress
	args.Eei = eei
	args.StakingSCConfig.NumRoundsWithoutBleed = 100
	args.StakingSCConfig.MaximumPercentageToBleed = 0.5
	args.StakingSCConfig.BleedPercentagePerRound = 0.00001
	args.JailAccessAddr = jailAccessAddr
	stakingSmartContract, _ := NewStakingSmartContract(args)

	stakerAddress := []byte("stakerAddr")
	stakerPubKey := []byte("stakerPublicKey")

	// cannot do jail if access addr is wrong should return userError
	doJail(t, stakingSmartContract, []byte("addr"), stakerPubKey, vmcommon.UserError)
	// cannot do jail if no stake should return userError
	doJail(t, stakingSmartContract, jailAccessAddr, stakerPubKey, vmcommon.UserError)
	//do stake should work
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, stakerPubKey)
	// jail should work
	blockChainHook.CurrentRoundCalled = func() uint64 {
		return 1000
	}
	doJail(t, stakingSmartContract, jailAccessAddr, stakerPubKey, vmcommon.Ok)

	//do unStake should return error because validator is jail
	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, stakerPubKey, vmcommon.UserError)

	// unJail wrong access address should not work
	doUnJail(t, stakingSmartContract, []byte("addr"), stakerPubKey, vmcommon.UserError)
	// cannot do unJail on a address that not stake
	doUnJail(t, stakingSmartContract, stakingAccessAddress, []byte("addr"), vmcommon.UserError)
	// unJail should work
	blockChainHook.CurrentRoundCalled = func() uint64 {
		return 1200
	}
	doUnJail(t, stakingSmartContract, stakingAccessAddress, stakerPubKey, vmcommon.Ok)
}

func TestStakingSc_ExecuteStakeStakeJailAndSwitch(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.StakingAccessAddr = stakingAccessAddress
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.StakingSCConfig.MaxNumberOfNodesForStake = 2
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	stakingSmartContract.flagCorrectJailedNotUnstakedEmptyQueue.Reset()

	stakerAddress := []byte("stakerAddr")
	stakerPubKey := []byte("stakerPublicKey")
	callerAddress := []byte("data")

	// do stake should work
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("firstKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("secondKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, stakerPubKey)

	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("firstKey"), vmcommon.Ok)
	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("secondKey"), vmcommon.Ok)
	checkIsStaked(t, stakingSmartContract, callerAddress, stakerPubKey, vmcommon.UserError)

	arguments := CreateVmContractCallInput()
	arguments.Function = "switchJailedWithWaiting"
	arguments.CallerAddr = args.EndOfEpochAccessAddr
	arguments.Arguments = [][]byte{[]byte("firstKey")}
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)
	// check if account is staked should return error code
	checkIsStaked(t, stakingSmartContract, callerAddress, stakerPubKey, vmcommon.Ok)
	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("firstKey"), vmcommon.UserError)

	arguments = CreateVmContractCallInput()
	arguments.Function = "switchJailedWithWaiting"
	arguments.CallerAddr = args.EndOfEpochAccessAddr
	arguments.Arguments = [][]byte{[]byte("secondKey")}
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)

	marshaledData := args.Eei.GetStorage([]byte("secondKey"))
	stakedData := &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Jailed)
	assert.True(t, stakedData.Staked)

	arguments.Function = "getTotalNumberOfRegisteredNodes"
	arguments.Arguments = [][]byte{}
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)

	lastOutput := eei.output[len(eei.output)-1]
	assert.Equal(t, lastOutput, []byte{2})
}

func TestStakingSc_ExecuteStakeStakeJailAndSwitchWithBoundaries(t *testing.T) {
	t.Parallel()

	maxStakedNodesNumber := 3
	minStakedNodesNumber := 1
	stakingAccessAddress := []byte("stakingAccessAddress")
	stakerAddress := []byte("stakerAddr")
	callerAddress := []byte("data")
	stakeValue := big.NewInt(100)

	didNotSwitchNoWaitingMessage := "did not switch as nobody in waiting, but jailed"
	didNotSwitchNotEnoughValidatorsMessage := "did not switch as not enough validators remaining"

	tests := []struct {
		name                       string
		stakedNodesNumber          int
		flagJailedRemoveEnabled    bool
		shouldBeJailed             bool
		shouldBeStaked             bool
		remainingStakedNodesNumber int
		returnMessage              string
	}{
		{
			name:                       "no queue, before fix, max nodes",
			stakedNodesNumber:          maxStakedNodesNumber,
			flagJailedRemoveEnabled:    false,
			shouldBeJailed:             true,
			shouldBeStaked:             true,
			remainingStakedNodesNumber: maxStakedNodesNumber,
			returnMessage:              didNotSwitchNoWaitingMessage,
		},
		{
			name:                       "no queue, before fix, min nodes",
			stakedNodesNumber:          minStakedNodesNumber,
			flagJailedRemoveEnabled:    false,
			shouldBeJailed:             true,
			shouldBeStaked:             true,
			remainingStakedNodesNumber: minStakedNodesNumber,
			returnMessage:              didNotSwitchNoWaitingMessage,
		},
		{
			name:                       "no queue, after fix, max nodes",
			stakedNodesNumber:          maxStakedNodesNumber,
			flagJailedRemoveEnabled:    true,
			shouldBeJailed:             true,
			shouldBeStaked:             false,
			remainingStakedNodesNumber: maxStakedNodesNumber - 1,
			returnMessage:              "",
		},
		{
			name:                       "no queue, after fix, min nodes ",
			stakedNodesNumber:          minStakedNodesNumber,
			flagJailedRemoveEnabled:    true,
			shouldBeJailed:             true,
			shouldBeStaked:             true,
			remainingStakedNodesNumber: minStakedNodesNumber,
			returnMessage:              didNotSwitchNotEnoughValidatorsMessage,
		},
		{
			name:                       "with 1 queue, before fix, max nodes",
			stakedNodesNumber:          maxStakedNodesNumber + 1,
			flagJailedRemoveEnabled:    false,
			shouldBeJailed:             true,
			shouldBeStaked:             false,
			remainingStakedNodesNumber: maxStakedNodesNumber,
			returnMessage:              "",
		},
		{
			name:                       "with 1 queue, after fix, max nodes",
			stakedNodesNumber:          maxStakedNodesNumber + 1,
			flagJailedRemoveEnabled:    true,
			shouldBeJailed:             true,
			shouldBeStaked:             false,
			remainingStakedNodesNumber: maxStakedNodesNumber,
			returnMessage:              "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jailedKey := []byte(fmt.Sprintf("staked_%v", 0))

			var stakedResult vmcommon.ReturnCode
			blockChainHook := &mock.BlockChainHookStub{}
			blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
				return nil, nil
			}

			eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
			args := createStakingSCArgs(eei, stakingAccessAddress, stakeValue, maxStakedNodesNumber)
			stakingSmartContract, _ := NewStakingSmartContract(args)

			stakingSmartContract.flagCorrectJailedNotUnstakedEmptyQueue.SetValue(tt.flagJailedRemoveEnabled)

			for i := 0; i < tt.stakedNodesNumber; i++ {
				doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte(fmt.Sprintf("staked_%v", i)))
			}

			for i := 0; i < tt.stakedNodesNumber; i++ {
				stakedResult = vmcommon.Ok
				shouldBeOnQueue := i >= maxStakedNodesNumber
				if shouldBeOnQueue {
					stakedResult = vmcommon.UserError
				}
				checkIsStaked(t, stakingSmartContract, callerAddress, []byte(fmt.Sprintf("staked_%v", i)), stakedResult)
			}

			eei.returnMessage = ""

			arguments := CreateVmContractCallInput()
			arguments.Function = "switchJailedWithWaiting"
			arguments.CallerAddr = args.EndOfEpochAccessAddr
			arguments.Arguments = [][]byte{jailedKey}
			retCode := stakingSmartContract.Execute(arguments)
			assert.Equal(t, vmcommon.Ok, retCode)

			assert.Equal(t, tt.returnMessage, eei.returnMessage)

			stakedResult = vmcommon.Ok
			if !tt.shouldBeStaked {
				stakedResult = vmcommon.UserError
			}
			checkIsStaked(t, stakingSmartContract, callerAddress, jailedKey, stakedResult)

			marshaledData := args.Eei.GetStorage(jailedKey)
			stakedData := &StakedDataV2_0{}
			_ = json.Unmarshal(marshaledData, stakedData)
			assert.Equal(t, tt.shouldBeJailed, stakedData.Jailed)
			assert.Equal(t, tt.shouldBeStaked, stakedData.Staked)

			arguments.Function = "getTotalNumberOfRegisteredNodes"
			arguments.Arguments = [][]byte{}
			retCode = stakingSmartContract.Execute(arguments)
			assert.Equal(t, vmcommon.Ok, retCode)

			lastOutput := eei.output[len(eei.output)-1]
			assert.Equal(t, []byte{byte(tt.remainingStakedNodesNumber)}, lastOutput)
		})
	}
}

func createStakingSCArgs(eei *vmContext, stakingAccessAddress []byte, stakeValue *big.Int, maxStakedNodesNumber int) ArgsNewStakingSmartContract {
	eei.SetSCAddress([]byte("addr"))

	args := createMockStakingScArguments()
	args.StakingAccessAddr = stakingAccessAddress
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.StakingSCConfig.MaxNumberOfNodesForStake = uint64(maxStakedNodesNumber)
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
	args.Eei = eei
	return args
}

func TestStakingSc_ExecuteJailNoQueueActivation(t *testing.T) {
	maxStakedNodesNumber := 3
	stakingAccessAddress := []byte("stakingAccessAddress")
	stakeValue := big.NewInt(100)

	correctJailedNoQueueEnableEpoch := uint32(5)

	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	args := createStakingSCArgs(eei, stakingAccessAddress, stakeValue, maxStakedNodesNumber)
	args.EpochConfig.EnableEpochs.CorrectJailedNotUnstakedEmptyQueueEpoch = correctJailedNoQueueEnableEpoch
	stakingSmartContract, _ := NewStakingSmartContract(args)

	assert.False(t, stakingSmartContract.flagCorrectJailedNotUnstakedEmptyQueue.IsSet())

	stakingSmartContract.EpochConfirmed(correctJailedNoQueueEnableEpoch-1, 0)
	assert.False(t, stakingSmartContract.flagCorrectJailedNotUnstakedEmptyQueue.IsSet())

	stakingSmartContract.EpochConfirmed(correctJailedNoQueueEnableEpoch, 0)
	assert.True(t, stakingSmartContract.flagCorrectJailedNotUnstakedEmptyQueue.IsSet())

	stakingSmartContract.EpochConfirmed(correctJailedNoQueueEnableEpoch+1, 0)
	assert.True(t, stakingSmartContract.flagCorrectJailedNotUnstakedEmptyQueue.IsSet())
}

func TestStakingSc_ExecuteStakeStakeStakeJailJailUnJailTwice(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.StakingAccessAddr = stakingAccessAddress
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.StakingSCConfig.MaxNumberOfNodesForStake = 2
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	stakerAddress := []byte("stakerAddr")
	stakerPubKey := []byte("stakerPublicKey")
	callerAddress := []byte("data")

	// do stake should work
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("firsstKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("secondKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, stakerPubKey)
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("fourthKey"))

	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("firsstKey"), vmcommon.Ok)
	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("secondKey"), vmcommon.Ok)
	checkIsStaked(t, stakingSmartContract, callerAddress, stakerPubKey, vmcommon.UserError)
	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("fourthKey"), vmcommon.UserError)

	arguments := CreateVmContractCallInput()
	arguments.Function = "switchJailedWithWaiting"
	arguments.CallerAddr = args.EndOfEpochAccessAddr
	arguments.Arguments = [][]byte{[]byte("firsstKey")}
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)
	// check if account is staked should return error code
	checkIsStaked(t, stakingSmartContract, callerAddress, stakerPubKey, vmcommon.Ok)
	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("firsstKey"), vmcommon.UserError)

	arguments = CreateVmContractCallInput()
	arguments.Function = "switchJailedWithWaiting"
	arguments.CallerAddr = args.EndOfEpochAccessAddr
	arguments.Arguments = [][]byte{[]byte("secondKey")}
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)
	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("fourthKey"), vmcommon.Ok)
	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("secondKey"), vmcommon.UserError)

	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("fifthhKey"))
	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("fifthhKey"), vmcommon.UserError)

	doGetStatus(t, stakingSmartContract, eei, []byte("firsstKey"), "jailed")
	doUnJail(t, stakingSmartContract, stakingAccessAddress, []byte("firsstKey"), vmcommon.Ok)
	doGetStatus(t, stakingSmartContract, eei, []byte("firsstKey"), "queued")
	doUnJail(t, stakingSmartContract, stakingAccessAddress, []byte("secondKey"), vmcommon.Ok)

	waitingList, _ := stakingSmartContract.getWaitingListHead()
	assert.Equal(t, uint32(3), waitingList.Length)
	assert.Equal(t, []byte("w_secondKey"), waitingList.LastJailedKey)
	assert.Equal(t, []byte("w_firsstKey"), waitingList.FirstKey)
	assert.Equal(t, []byte("w_fifthhKey"), waitingList.LastKey)

	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("sixthhKey"))
	doGetWaitingListIndex(t, stakingSmartContract, eei, []byte("firsstKey"), vmcommon.Ok, 1)
	doGetWaitingListIndex(t, stakingSmartContract, eei, []byte("secondKey"), vmcommon.Ok, 2)
	doGetWaitingListIndex(t, stakingSmartContract, eei, []byte("fifthhKey"), vmcommon.Ok, 3)
	doGetWaitingListIndex(t, stakingSmartContract, eei, []byte("sixthhKey"), vmcommon.Ok, 4)

	outPut := doGetWaitingListRegisterNonceAndRewardAddress(t, stakingSmartContract, eei)
	assert.Equal(t, 12, len(outPut))

	stakingSmartContract.unBondPeriod = 0
	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("secondKey"), vmcommon.Ok)
	doUnBond(t, stakingSmartContract, stakingAccessAddress, []byte("secondKey"), vmcommon.Ok)
	waitingList, _ = stakingSmartContract.getWaitingListHead()
	assert.Equal(t, []byte("w_firsstKey"), waitingList.LastJailedKey)

	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("firsstKey"), vmcommon.Ok)
	doUnBond(t, stakingSmartContract, stakingAccessAddress, []byte("firsstKey"), vmcommon.Ok)
	waitingList, _ = stakingSmartContract.getWaitingListHead()
	assert.Equal(t, 0, len(waitingList.LastJailedKey))

	doGetWaitingListSize(t, stakingSmartContract, eei, 2)
	doGetRewardAddress(t, stakingSmartContract, eei, []byte("fifthhKey"), string(stakerAddress))
	doGetStatus(t, stakingSmartContract, eei, []byte("fifthhKey"), "queued")
	doGetStatus(t, stakingSmartContract, eei, []byte("fourthKey"), "staked")

	stakingSmartContract.unBondPeriod = 100
	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 1
	}
	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("fourthKey"), vmcommon.Ok)
	doGetRemainingUnbondPeriod(t, stakingSmartContract, eei, []byte("fourthKey"), 100)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 50
	}
	doGetRemainingUnbondPeriod(t, stakingSmartContract, eei, []byte("fourthKey"), 51)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 101
	}
	doGetRemainingUnbondPeriod(t, stakingSmartContract, eei, []byte("fourthKey"), 0)

	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("seventKey"))
	doGetWaitingListSize(t, stakingSmartContract, eei, 2)
	outPut = doGetWaitingListRegisterNonceAndRewardAddress(t, stakingSmartContract, eei)
	assert.Equal(t, 6, len(outPut))

	arguments.Function = "getTotalNumberOfRegisteredNodes"
	arguments.Arguments = [][]byte{}
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)

	lastOutput := eei.output[len(eei.output)-1]
	assert.Equal(t, lastOutput, []byte{4})
}

func TestStakingSc_ExecuteStakeUnStakeJailCombinations(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.StakingAccessAddr = stakingAccessAddress
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.StakingSCConfig.MaxNumberOfNodesForStake = 2
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	stakerAddress := []byte("stakerAddr")
	stakerPubKey := []byte("stakerPun")
	callerAddress := []byte("data")

	// do stake should work
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("firsstKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("secondKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, stakerPubKey)
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("fourthKey"))

	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("firsstKey"), vmcommon.Ok)
	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("secondKey"), vmcommon.Ok)
	checkIsStaked(t, stakingSmartContract, callerAddress, stakerPubKey, vmcommon.UserError)
	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("fourthKey"), vmcommon.UserError)

	doSwitchJailedWithWaiting(t, stakingSmartContract, []byte("firsstKey"))
	// check if account is staked should return error code
	checkIsStaked(t, stakingSmartContract, callerAddress, stakerPubKey, vmcommon.Ok)
	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("firsstKey"), vmcommon.UserError)

	doSwitchJailedWithWaiting(t, stakingSmartContract, []byte("secondKey"))
	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("fourthKey"), vmcommon.Ok)
	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("secondKey"), vmcommon.UserError)

	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("fifthhKey"))
	checkIsStaked(t, stakingSmartContract, callerAddress, []byte("fifthhKey"), vmcommon.UserError)

	doGetStatus(t, stakingSmartContract, eei, []byte("firsstKey"), "jailed")
	doUnJail(t, stakingSmartContract, stakingAccessAddress, []byte("firsstKey"), vmcommon.Ok)
	doGetStatus(t, stakingSmartContract, eei, []byte("firsstKey"), "queued")
	doUnJail(t, stakingSmartContract, stakingAccessAddress, []byte("secondKey"), vmcommon.Ok)

	waitingList, _ := stakingSmartContract.getWaitingListHead()
	assert.Equal(t, uint32(3), waitingList.Length)
	assert.Equal(t, []byte("w_secondKey"), waitingList.LastJailedKey)
	assert.Equal(t, []byte("w_firsstKey"), waitingList.FirstKey)
	assert.Equal(t, []byte("w_fifthhKey"), waitingList.LastKey)

	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("sixthhKey"))
	doGetWaitingListIndex(t, stakingSmartContract, eei, []byte("firsstKey"), vmcommon.Ok, 1)
	doGetWaitingListIndex(t, stakingSmartContract, eei, []byte("secondKey"), vmcommon.Ok, 2)
	doGetWaitingListIndex(t, stakingSmartContract, eei, []byte("fifthhKey"), vmcommon.Ok, 3)
	doGetWaitingListIndex(t, stakingSmartContract, eei, []byte("sixthhKey"), vmcommon.Ok, 4)
	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("firsstKey"), vmcommon.Ok)
	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("secondKey"), vmcommon.Ok)

	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("sixthhKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("seventKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("eigthhKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("ninethKey"))

	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, stakerPubKey, vmcommon.Ok)
	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("fourthKey"), vmcommon.Ok)
	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("fifthhKey"), vmcommon.Ok)
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("tenthhKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("elventKey"))

	// unstake from additional queue
	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("tenthhKey"), vmcommon.Ok)

	// jail and unjail the node
	doSwitchJailedWithWaiting(t, stakingSmartContract, []byte("sixthhKey"))
	doUnJail(t, stakingSmartContract, stakingAccessAddress, []byte("sixthhKey"), vmcommon.Ok)

	doGetWaitingListIndex(t, stakingSmartContract, eei, []byte("sixthhKey"), vmcommon.Ok, 1)
}

func TestStakingSc_UnBondFromWaitingNotPossible(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.StakingAccessAddr = stakingAccessAddress
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.StakingSCConfig.MaxNumberOfNodesForStake = 2
	args.Eei = eei
	args.StakingSCConfig.UnBondPeriod = 100
	stakingSmartContract, _ := NewStakingSmartContract(args)

	stakerAddress := []byte("stakerAddr")

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 1
	}

	// do stake should work
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("firsstKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("secondKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("thirdKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("fourthKey"))

	doUnBond(t, stakingSmartContract, stakingAccessAddress, []byte("secondKey"), vmcommon.UserError)
	doUnBond(t, stakingSmartContract, stakingAccessAddress, []byte("thirdKey"), vmcommon.UserError)

	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("thirdKey"), vmcommon.Ok)
	doUnBond(t, stakingSmartContract, stakingAccessAddress, []byte("thirdKey"), vmcommon.Ok)

	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("secondKey"), vmcommon.Ok)
	doUnBond(t, stakingSmartContract, stakingAccessAddress, []byte("secondKey"), vmcommon.UserError)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 200
	}
	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("secondKey"), vmcommon.UserError)
	doUnBond(t, stakingSmartContract, stakingAccessAddress, []byte("secondKey"), vmcommon.Ok)

	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("fourthKey"), vmcommon.Ok)
	doUnBond(t, stakingSmartContract, stakingAccessAddress, []byte("fourthKey"), vmcommon.UserError)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 400
	}

	doUnBond(t, stakingSmartContract, stakingAccessAddress, []byte("fourthKey"), vmcommon.Ok)
}

func Test_NoActionAllowedForBadRatingOrJailed(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	accountsStub := &stateMock.AccountsStub{}
	raterStub := &mock.RaterMock{}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, accountsStub, raterStub)
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.StakingAccessAddr = stakingAccessAddress
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.StakingSCConfig.MaxNumberOfNodesForStake = 1
	args.Eei = eei
	args.StakingSCConfig.UnBondPeriod = 100
	stakingSmartContract, _ := NewStakingSmartContract(args)

	stakerAddress := []byte("stakerAddr")

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 1
	}

	// do stake should work
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("firsstKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("secondKey"))

	peerAccount := state.NewEmptyPeerAccount()
	accountsStub.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		return peerAccount, nil
	}
	peerAccount.List = string(common.JailedList)
	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("secondKey"), vmcommon.UserError)
	doUnBond(t, stakingSmartContract, stakingAccessAddress, []byte("secondKey"), vmcommon.UserError)

	peerAccount.List = string(common.EligibleList)
	peerAccount.TempRating = 9
	raterStub.GetChancesCalled = func(u uint32) uint32 {
		if u == 0 {
			return 10
		}
		return 5
	}
	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("firsstKey"), vmcommon.UserError)
	doUnBond(t, stakingSmartContract, stakingAccessAddress, []byte("firsstKey"), vmcommon.UserError)
}

func Test_UnJailNotAllowedIfJailed(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	accountsStub := &stateMock.AccountsStub{}
	raterStub := &mock.RaterMock{}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, accountsStub, raterStub)
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.StakingAccessAddr = stakingAccessAddress
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.StakingSCConfig.MaxNumberOfNodesForStake = 1
	args.Eei = eei
	args.StakingSCConfig.UnBondPeriod = 100
	stakingSmartContract, _ := NewStakingSmartContract(args)

	stakerAddress := []byte("stakerAddr")

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 1
	}

	// do stake should work
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("firsstKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("secondKey"))

	peerAccount := state.NewEmptyPeerAccount()
	accountsStub.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		return peerAccount, nil
	}
	peerAccount.List = string(common.EligibleList)
	doUnJail(t, stakingSmartContract, stakingAccessAddress, []byte("firsstKey"), vmcommon.UserError)
	doUnJail(t, stakingSmartContract, stakingAccessAddress, []byte("secondKey"), vmcommon.UserError)

	peerAccount.List = string(common.JailedList)
	doUnJail(t, stakingSmartContract, stakingAccessAddress, []byte("firsstKey"), vmcommon.Ok)
	doUnJail(t, stakingSmartContract, stakingAccessAddress, []byte("secondKey"), vmcommon.Ok)
}

func TestStakingSc_updateConfigMinNodesOK(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.StakingAccessAddr = stakingAccessAddress
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.StakingSCConfig.MaxNumberOfNodesForStake = 40
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)
	stakingConfig := &StakingNodesConfig{
		MinNumNodes: 5,
		MaxNumNodes: 40,
		StakedNodes: 10,
		JailedNodes: 2,
	}
	stakingSmartContract.setConfig(stakingConfig)

	originalStakeConfigMarshalled := args.Eei.GetStorage([]byte(nodesConfigKey))
	require.NotEqual(t, 0, originalStakeConfigMarshalled)

	originalStakeConfig := &StakingNodesConfig{}
	err := json.Unmarshal(originalStakeConfigMarshalled, originalStakeConfig)
	require.Nil(t, err)
	require.Equal(t, stakingConfig, originalStakeConfig)

	newMinNodes := int64(30)
	arguments := CreateVmContractCallInput()
	arguments.Function = "updateConfigMinNodes"
	arguments.CallerAddr = args.EndOfEpochAccessAddr
	arguments.Arguments = [][]byte{big.NewInt(0).SetInt64(newMinNodes).Bytes()}
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)

	// check storage is updated
	updatedStakeConfigMarshalled := args.Eei.GetStorage([]byte(nodesConfigKey))
	require.NotEqual(t, 0, updatedStakeConfigMarshalled)

	updatedStakeConfig := &StakingNodesConfig{}
	err = json.Unmarshal(updatedStakeConfigMarshalled, updatedStakeConfig)
	require.Nil(t, err)

	require.Equal(t, originalStakeConfig.JailedNodes, updatedStakeConfig.JailedNodes)
	require.Equal(t, originalStakeConfig.MaxNumNodes, updatedStakeConfig.MaxNumNodes)
	require.Equal(t, originalStakeConfig.StakedNodes, updatedStakeConfig.StakedNodes)
	require.NotEqual(t, newMinNodes, originalStakeConfig.MinNumNodes)
	require.Equal(t, newMinNodes, updatedStakeConfig.MinNumNodes)
}

func TestStakingSc_updateConfigMaxNodesOK(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
	args.StakingAccessAddr = stakingAccessAddress
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.StakingSCConfig.MaxNumberOfNodesForStake = 40
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)
	stakingConfig := &StakingNodesConfig{
		MinNumNodes: 5,
		MaxNumNodes: 40,
		StakedNodes: 10,
		JailedNodes: 2,
	}
	stakingSmartContract.setConfig(stakingConfig)

	originalStakeConfigMarshalled := args.Eei.GetStorage([]byte(nodesConfigKey))
	require.NotEqual(t, 0, originalStakeConfigMarshalled)

	originalStakeConfig := &StakingNodesConfig{}
	err := json.Unmarshal(originalStakeConfigMarshalled, originalStakeConfig)
	require.Nil(t, err)
	require.Equal(t, stakingConfig, originalStakeConfig)

	newMaxNodes := int64(100)
	arguments := CreateVmContractCallInput()
	arguments.Function = "updateConfigMaxNodes"
	arguments.CallerAddr = args.EndOfEpochAccessAddr
	arguments.Arguments = [][]byte{big.NewInt(0).SetInt64(newMaxNodes).Bytes()}
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)

	// check storage is updated
	updatedStakeConfigMarshalled := args.Eei.GetStorage([]byte(nodesConfigKey))
	require.NotEqual(t, 0, updatedStakeConfigMarshalled)

	updatedStakeConfig := &StakingNodesConfig{}
	err = json.Unmarshal(updatedStakeConfigMarshalled, updatedStakeConfig)
	require.Nil(t, err)

	require.Equal(t, originalStakeConfig.JailedNodes, updatedStakeConfig.JailedNodes)
	require.Equal(t, originalStakeConfig.MinNumNodes, updatedStakeConfig.MinNumNodes)
	require.Equal(t, originalStakeConfig.StakedNodes, updatedStakeConfig.StakedNodes)
	require.NotEqual(t, newMaxNodes, originalStakeConfig.MaxNumNodes)
	require.Equal(t, newMaxNodes, updatedStakeConfig.MaxNumNodes)
}

func TestStakingSC_SetOwnersOnAddressesNotEnabledShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockStakingScArguments()
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 100
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	args.Eei = eei

	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "setOwnersOnAddresses"
	arguments.CallerAddr = []byte("owner")
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.UserError)
	assert.Equal(t, "invalid method to call", eei.returnMessage)
}

func TestStakingSC_SetOwnersOnAddressesWrongCallerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockStakingScArguments()
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	args.Eei = eei

	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "setOwnersOnAddresses"
	arguments.CallerAddr = []byte("owner")
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.UserError)
	assert.True(t, strings.Contains(eei.returnMessage, "setOwnersOnAddresses function not allowed to be called by address"))
}

func TestStakingSC_SetOwnersOnAddressesWrongArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockStakingScArguments()
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	args.Eei = eei

	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "setOwnersOnAddresses"
	arguments.CallerAddr = args.EndOfEpochAccessAddr
	arguments.Arguments = [][]byte{[]byte("bls key")}
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.UserError)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments: expected an even number of arguments"))
}

func TestStakingSC_SetOwnersOnAddressesShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockStakingScArguments()
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	args.Eei = eei

	stakingSmartContract, _ := NewStakingSmartContract(args)
	blsKey1 := []byte("blsKey1")
	owner1 := []byte("owner1")
	blsKey2 := []byte("blsKey2")
	owner2 := []byte("owner2")

	doStake(t, stakingSmartContract, args.StakingAccessAddr, owner1, blsKey1)
	doStake(t, stakingSmartContract, args.StakingAccessAddr, owner2, blsKey2)

	arguments := CreateVmContractCallInput()
	arguments.Function = "setOwnersOnAddresses"
	arguments.CallerAddr = args.EndOfEpochAccessAddr
	arguments.Arguments = [][]byte{blsKey1, owner1, blsKey2, owner2}
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)

	registrationData, err := stakingSmartContract.getOrCreateRegisteredData(blsKey1)
	require.Nil(t, err)
	assert.Equal(t, owner1, registrationData.OwnerAddress)

	registrationData, err = stakingSmartContract.getOrCreateRegisteredData(blsKey2)
	require.Nil(t, err)
	assert.Equal(t, owner2, registrationData.OwnerAddress)
}

func TestStakingSC_SetOwnersOnAddressesEmptyArgsShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockStakingScArguments()
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	args.Eei = eei

	stakingSmartContract, _ := NewStakingSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "setOwnersOnAddresses"
	arguments.CallerAddr = args.EndOfEpochAccessAddr
	arguments.Arguments = make([][]byte, 0)
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)
}

func TestStakingSC_GetOwnerStakingV2NotEnabledShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockStakingScArguments()
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 100
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	args.Eei = eei

	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "getOwner"
	arguments.CallerAddr = []byte("owner")
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.UserError)
	assert.Equal(t, "invalid method to call", eei.returnMessage)
}

func TestStakingSC_GetOwnerWrongCallerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockStakingScArguments()
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	args.Eei = eei

	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "getOwner"
	arguments.CallerAddr = []byte("owner")
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.UserError)
	assert.True(t, strings.Contains(eei.returnMessage, "this is only a view function"))
}

func TestStakingSC_GetOwnerWrongArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockStakingScArguments()
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	args.Eei = eei

	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "getOwner"
	arguments.CallerAddr = args.StakingAccessAddr
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.UserError)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments: expected min"))
}

func TestStakingSC_GetOwnerShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockStakingScArguments()
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	args.Eei = eei

	stakingSmartContract, _ := NewStakingSmartContract(args)
	blsKey := []byte("blsKey")
	owner := []byte("owner")

	doStake(t, stakingSmartContract, args.StakingAccessAddr, owner, blsKey)

	arguments := CreateVmContractCallInput()
	arguments.Function = "setOwnersOnAddresses"
	arguments.CallerAddr = args.EndOfEpochAccessAddr
	arguments.Arguments = [][]byte{blsKey, owner}
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)

	arguments = CreateVmContractCallInput()
	arguments.Function = "getOwner"
	arguments.CallerAddr = args.StakingAccessAddr
	arguments.Arguments = [][]byte{blsKey}
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)

	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, owner, vmOutput.ReturnData[0])
}

func TestStakingSc_StakeFromQueue(t *testing.T) {
	t.Parallel()

	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := vm.ValidatorSCAddress
	args := createMockStakingScArguments()
	args.StakingAccessAddr = stakingAccessAddress
	args.StakingSCConfig.MaxNumberOfNodesForStake = 1
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
	args.Eei = eei
	args.StakingSCConfig.UnBondPeriod = 100
	stakingSmartContract, _ := NewStakingSmartContract(args)

	stakerAddress := []byte("stakerAddr")

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 1
	}

	// do stake should work
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("firsstKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("secondKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("thirdKeyy"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("fourthKey"))

	waitingReturn := doGetWaitingListRegisterNonceAndRewardAddress(t, stakingSmartContract, eei)
	assert.Equal(t, len(waitingReturn), 9)

	newMaxNodes := int64(100)
	arguments := CreateVmContractCallInput()
	arguments.Function = "updateConfigMaxNodes"
	arguments.CallerAddr = args.EndOfEpochAccessAddr
	arguments.Arguments = [][]byte{big.NewInt(0).SetInt64(newMaxNodes).Bytes()}
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)

	validatorData := &ValidatorDataV2{
		TotalStakeValue: big.NewInt(200),
		TotalUnstaked:   big.NewInt(0),
		RewardAddress:   stakerAddress,
		BlsPubKeys:      [][]byte{[]byte("firsstKey"), []byte("secondKey"), []byte("thirdKeyy"), []byte("fourthKey")},
	}
	marshaledData, _ := stakingSmartContract.marshalizer.Marshal(validatorData)
	eei.SetStorageForAddress(vm.ValidatorSCAddress, stakerAddress, marshaledData)

	currentOutPutIndex := len(eei.output)
	arguments.Function = "stakeNodesFromQueue"
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)

	// nothing to stake - as not enough funds - one remains in waiting queue
	assert.Equal(t, currentOutPutIndex, len(eei.output))

	cleanAdditionalInput := CreateVmContractCallInput()
	cleanAdditionalInput.Function = "cleanAdditionalQueue"
	cleanAdditionalInput.CallerAddr = args.EndOfEpochAccessAddr
	retCode = stakingSmartContract.Execute(cleanAdditionalInput)
	assert.Equal(t, retCode, vmcommon.Ok)

	newHead, _ := stakingSmartContract.getWaitingListHead()
	assert.Equal(t, uint32(1), newHead.Length)

	doGetStatus(t, stakingSmartContract, eei, []byte("secondKey"), "queued")

	newMaxNodes = int64(1)
	arguments = CreateVmContractCallInput()
	arguments.Function = "updateConfigMaxNodes"
	arguments.CallerAddr = args.EndOfEpochAccessAddr
	arguments.Arguments = [][]byte{big.NewInt(0).SetInt64(newMaxNodes).Bytes()}
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)

	// stake them again - as they were deleted from waiting list
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("thirdKeyy"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("fourthKey"))

	validatorData = &ValidatorDataV2{
		TotalStakeValue: big.NewInt(400),
	}
	marshaledData, _ = stakingSmartContract.marshalizer.Marshal(validatorData)
	eei.SetStorageForAddress(vm.ValidatorSCAddress, stakerAddress, marshaledData)

	newMaxNodes = int64(100)
	arguments.Arguments = [][]byte{big.NewInt(0).SetInt64(newMaxNodes).Bytes()}
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)

	currentOutPutIndex = len(eei.output)
	arguments.Function = "stakeNodesFromQueue"
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)

	for i := currentOutPutIndex; i < len(eei.output); i += 2 {
		checkIsStaked(t, stakingSmartContract, arguments.CallerAddr, eei.output[i], vmcommon.Ok)
	}
	assert.Equal(t, 6, len(eei.output)-currentOutPutIndex)
	stakingConfig := stakingSmartContract.getConfig()
	assert.Equal(t, stakingConfig.StakedNodes, int64(4))

	retCode = stakingSmartContract.Execute(cleanAdditionalInput)
	assert.Equal(t, retCode, vmcommon.Ok)
	newHead, _ = stakingSmartContract.getWaitingListHead()
	assert.Equal(t, uint32(0), newHead.Length)
}

func TestStakingSC_UnstakeAtEndOfEpoch(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.StakingAccessAddr = stakingAccessAddress
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	stakerAddress := []byte("stakerAddr")
	stakerPubKey := []byte("stakerPublicKey")
	callerAddress := []byte("data")

	// do stake should work
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, stakerPubKey)
	checkIsStaked(t, stakingSmartContract, callerAddress, stakerPubKey, vmcommon.Ok)

	doUnStakeAtEndOfEpoch(t, stakingSmartContract, stakerPubKey, vmcommon.Ok)
	checkIsStaked(t, stakingSmartContract, callerAddress, stakerPubKey, vmcommon.UserError)
}

func TestStakingSC_ResetWaitingListUnJailed(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.StakingAccessAddr = stakingAccessAddress
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.StakingSCConfig.MaxNumberOfNodesForStake = 1
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	stakerAddress := []byte("stakerAddr")

	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("firsstKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("secondKey"))

	arguments := CreateVmContractCallInput()
	arguments.Function = "resetLastUnJailedFromQueue"
	arguments.Arguments = [][]byte{}
	arguments.CallerAddr = stakingSmartContract.endOfEpochAccessAddr

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	doSwitchJailedWithWaiting(t, stakingSmartContract, []byte("firsstKey"))
	doUnJail(t, stakingSmartContract, stakingAccessAddress, []byte("firsstKey"), vmcommon.Ok)

	waitingList, _ := stakingSmartContract.getWaitingListHead()
	assert.Equal(t, waitingList.LastJailedKey, []byte("w_firsstKey"))

	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	waitingList, _ = stakingSmartContract.getWaitingListHead()
	assert.Equal(t, len(waitingList.LastJailedKey), 0)

	arguments.CallerAddr = []byte("anotherAddress")
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)

	arguments.CallerAddr = stakingSmartContract.endOfEpochAccessAddr
	arguments.Arguments = [][]byte{[]byte("someArg")}
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)

	stakingSmartContract.flagCorrectLastUnjailed.Reset()
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSc_UnStakeNodeWhenMaxNumIsMoreShouldNotStakeFromWaiting(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.StakingAccessAddr = stakingAccessAddress
	args.StakingSCConfig.MinStakeValue = stakeValue.Text(10)
	args.StakingSCConfig.MaxNumberOfNodesForStake = 2
	args.MinNumNodes = 1
	args.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
	args.EpochConfig.EnableEpochs.CorrectLastUnjailedEnableEpoch = 0
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	stakerAddress := []byte("stakerAddr")

	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("firsstKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("secondKey"))
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("thirddKey"))

	stakingSmartContract.addToStakedNodes(10)

	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("firsstKey"), vmcommon.Ok)
	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, []byte("secondKey"), vmcommon.Ok)

	doGetStatus(t, stakingSmartContract, eei, []byte("thirddKey"), "queued")
}

func TestStakingSc_ChangeRewardAndOwnerAddress(t *testing.T) {
	t.Parallel()

	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.StakingAccessAddr = stakingAccessAddress
	args.Eei = eei
	sc, _ := NewStakingSmartContract(args)

	stakerAddress := []byte("stakerAddr")

	doStake(t, sc, stakingAccessAddress, stakerAddress, []byte("firsstKey"))
	doStake(t, sc, stakingAccessAddress, stakerAddress, []byte("secondKey"))
	doStake(t, sc, stakingAccessAddress, stakerAddress, []byte("thirddKey"))

	sc.flagValidatorToDelegation.Reset()

	arguments := CreateVmContractCallInput()
	arguments.Function = "changeOwnerAndRewardAddress"
	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)

	_ = sc.flagValidatorToDelegation.SetReturningPrevious()
	eei.returnMessage = ""
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "change owner and reward address can be called by validator SC only")

	eei.returnMessage = ""
	arguments.CallerAddr = stakingAccessAddress
	arguments.CallValue.SetUint64(10)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "callValue must be 0")

	arguments.CallValue.SetUint64(0)
	eei.returnMessage = ""
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "number of arguments is 2 at minimum")

	arguments.Arguments = [][]byte{[]byte("key1"), []byte("key2")}
	eei.returnMessage = ""
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "new address must be a smart contract address")

	arguments.Arguments[0] = vm.FirstDelegationSCAddress
	eei.returnMessage = ""
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "cannot change owner and reward address for a key which is not registered")

	arguments.Arguments[1] = []byte("firsstKey")
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	doJail(t, sc, sc.jailAccessAddr, []byte("secondKey"), vmcommon.Ok)

	arguments.Arguments = [][]byte{vm.FirstDelegationSCAddress, []byte("firsstKey"), []byte("secondKey"), []byte("thirddKey")}
	eei.returnMessage = ""
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "can not migrate nodes while jailed nodes exists")

	doUnJail(t, sc, sc.stakeAccessAddr, []byte("secondKey"), vmcommon.Ok)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestStakingSc_RemoveFromWaitingListFirst(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		flag bool
	}{
		{
			name: "BeforeFix",
			flag: false,
		},
		{
			name: "AfterFix",
			flag: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			firstBLS := []byte("first")
			firstKey := createWaitingListKey(firstBLS)
			secondBLS := []byte("second")
			secondKey := createWaitingListKey(secondBLS)

			m := make(map[string]interface{})
			m[string(firstKey)] = &ElementInList{firstBLS, firstKey, secondKey}
			m[string(secondKey)] = &ElementInList{secondBLS, firstKey, nil}
			m[waitingListHeadKey] = &WaitingList{firstKey, secondKey, 2, nil}

			marshalizer := &marshal.JsonMarshalizer{}

			blockChainHook := &mock.BlockChainHookStub{}
			blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
				obj, ok := m[string(index)]
				if ok {
					return marshalizer.Marshal(obj)
				}
				return nil, nil
			}

			eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})

			args := createMockStakingScArguments()
			args.Marshalizer = marshalizer
			args.Eei = eei
			sc, _ := NewStakingSmartContract(args)
			if tt.flag {
				_ = sc.flagCorrectFirstQueued.SetReturningPrevious()
			} else {
				sc.flagCorrectFirstQueued.Reset()
			}
			err := sc.removeFromWaitingList(firstBLS)

			assert.Nil(t, err)
			wlh, err := sc.getWaitingListHead()
			assert.Nil(t, err)
			assert.NotNil(t, wlh)
			assert.Equal(t, secondKey, wlh.FirstKey)
			assert.Equal(t, secondKey, wlh.LastKey)
		})
	}
}

func TestStakingSc_RemoveFromWaitingListSecondThatLooksLikeFirstBeforeFix(t *testing.T) {
	t.Parallel()

	firstBLS := []byte("first")
	firstKey := createWaitingListKey(firstBLS)
	secondBLS := []byte("second")
	secondKey := createWaitingListKey(secondBLS)
	thirdBLS := []byte("third")
	thirdKey := createWaitingListKey(thirdBLS)

	m := make(map[string]interface{})
	m[string(firstKey)] = &ElementInList{firstBLS, firstKey, secondKey}
	// PreviousKey is set to self to look like it was the first
	m[string(secondKey)] = &ElementInList{secondBLS, secondKey, thirdKey}
	m[string(thirdKey)] = &ElementInList{thirdBLS, thirdKey, nil}
	m[waitingListHeadKey] = &WaitingList{firstKey, thirdKey, 3, nil}

	marshalizer := &marshal.JsonMarshalizer{}

	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		obj, ok := m[string(index)]
		if ok {
			return marshalizer.Marshal(obj)
		}
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})

	args := createMockStakingScArguments()
	args.Marshalizer = marshalizer
	args.Eei = eei
	sc, _ := NewStakingSmartContract(args)
	sc.flagCorrectFirstQueued.Reset()

	err := sc.removeFromWaitingList(secondBLS)
	assert.Nil(t, err)
	wlh, err := sc.getWaitingListHead()
	assert.Nil(t, err)
	assert.NotNil(t, wlh)
	// Forgot about the initial first key and now the first is the third
	assert.Equal(t, thirdKey, wlh.FirstKey)
	assert.Equal(t, thirdKey, wlh.LastKey)

	thirdElement, err := sc.getWaitingListElement(wlh.FirstKey)
	assert.Nil(t, err)
	assert.Equal(t, thirdKey, thirdElement.PreviousKey)
	assert.Nil(t, thirdElement.NextKey)
}

func TestStakingSc_RemoveFromWaitingListSecondThatLooksLikeFirstAfterFix(t *testing.T) {
	t.Parallel()

	firstBLS := []byte("first")
	firstKey := createWaitingListKey(firstBLS)
	secondBLS := []byte("second")
	secondKey := createWaitingListKey(secondBLS)
	thirdBLS := []byte("third")
	thirdKey := createWaitingListKey(thirdBLS)

	m := make(map[string]interface{})
	m[string(firstKey)] = &ElementInList{firstBLS, firstKey, secondKey}
	// PreviousKey is set to self to look like it was the first
	m[string(secondKey)] = &ElementInList{secondBLS, secondKey, thirdKey}
	m[string(thirdKey)] = &ElementInList{thirdBLS, thirdKey, nil}
	m[waitingListHeadKey] = &WaitingList{firstKey, thirdKey, 3, nil}

	marshalizer := &marshal.JsonMarshalizer{}

	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		obj, ok := m[string(index)]
		if ok {
			return marshalizer.Marshal(obj)
		}
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})

	args := createMockStakingScArguments()
	args.Marshalizer = marshalizer
	args.Eei = eei
	sc, _ := NewStakingSmartContract(args)
	_ = sc.flagCorrectFirstQueued.SetReturningPrevious()

	err := sc.removeFromWaitingList(secondBLS)
	assert.Nil(t, err)
	wlh, err := sc.getWaitingListHead()
	assert.Nil(t, err)
	assert.NotNil(t, wlh)
	assert.Equal(t, firstKey, wlh.FirstKey)
	assert.Equal(t, thirdKey, wlh.LastKey)

	firstElement, err := sc.getWaitingListElement(firstKey)
	assert.Nil(t, err)
	assert.Equal(t, firstKey, firstElement.PreviousKey)
	assert.Equal(t, thirdKey, firstElement.NextKey)

	thirdElement, err := sc.getWaitingListElement(thirdKey)
	assert.Nil(t, err)
	assert.Equal(t, firstKey, thirdElement.PreviousKey)
	assert.Nil(t, nil, thirdElement.NextKey)
}

func TestStakingSc_RemoveFromWaitingListNotFoundPreviousShouldErrAndFinish(t *testing.T) {
	t.Parallel()

	firstBLS := []byte("first")
	firstKey := createWaitingListKey(firstBLS)
	secondBLS := []byte("second")
	secondKey := createWaitingListKey(secondBLS)
	thirdBLS := []byte("third")
	thirdKey := createWaitingListKey(thirdBLS)
	unknownBLS := []byte("unknown")
	unknownKey := createWaitingListKey(unknownBLS)

	m := make(map[string]interface{})
	m[string(firstKey)] = &ElementInList{firstBLS, firstKey, secondKey}
	m[string(secondKey)] = &ElementInList{secondBLS, secondKey, nil}
	m[string(thirdKey)] = &ElementInList{thirdBLS, unknownKey, nil}
	m[waitingListHeadKey] = &WaitingList{firstKey, thirdKey, 3, nil}

	marshalizer := &marshal.JsonMarshalizer{}

	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		obj, ok := m[string(index)]
		if ok {
			return marshalizer.Marshal(obj)
		}
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})

	args := createMockStakingScArguments()
	args.Marshalizer = marshalizer
	args.Eei = eei
	sc, _ := NewStakingSmartContract(args)
	_ = sc.flagCorrectFirstQueued.SetReturningPrevious()

	err := sc.removeFromWaitingList(thirdBLS)
	assert.Equal(t, vm.ErrElementNotFound, err)
}

func TestStakingSc_InsertAfterLastJailedBeforeFix(t *testing.T) {
	t.Parallel()

	firstBLS := []byte("first")
	firstKey := createWaitingListKey(firstBLS)
	jailedBLS := []byte("jailedBLS")
	jailedKey := createWaitingListKey(jailedBLS)

	m := make(map[string]interface{})
	m[string(firstKey)] = &ElementInList{firstBLS, firstKey, nil}
	waitingListHead := &WaitingList{firstKey, firstKey, 1, nil}
	m[waitingListHeadKey] = waitingListHead

	marshalizer := &marshal.JsonMarshalizer{}

	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		obj, ok := m[string(index)]
		if ok {
			return marshalizer.Marshal(obj)
		}

		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})

	args := createMockStakingScArguments()
	args.Marshalizer = marshalizer
	args.Eei = eei
	sc, _ := NewStakingSmartContract(args)
	sc.flagCorrectFirstQueued.Reset()
	err := sc.insertAfterLastJailed(waitingListHead, jailedBLS)
	assert.Nil(t, err)

	wlh, err := sc.getWaitingListHead()
	assert.Nil(t, err)
	assert.NotNil(t, wlh)
	assert.Equal(t, jailedKey, wlh.FirstKey)
	assert.Equal(t, jailedKey, wlh.LastJailedKey)
	// increase is done in the calling method
	assert.Equal(t, uint32(1), wlh.Length)

	firstElement, err := sc.getWaitingListElement(wlh.FirstKey)
	assert.Nil(t, err)
	assert.NotNil(t, firstElement)
	assert.Equal(t, jailedBLS, firstElement.BLSPublicKey)
	assert.Equal(t, jailedKey, firstElement.PreviousKey)
	assert.Equal(t, firstKey, firstElement.NextKey)

	previousFirstElement, err := sc.getWaitingListElement(firstElement.NextKey)
	assert.Nil(t, err)
	assert.NotNil(t, previousFirstElement)
	assert.Equal(t, firstBLS, previousFirstElement.BLSPublicKey)
	assert.Equal(t, firstKey, previousFirstElement.PreviousKey)
	assert.Nil(t, previousFirstElement.NextKey)
}

func TestStakingSc_InsertAfterLastJailedAfterFix(t *testing.T) {
	t.Parallel()

	firstBLS := []byte("first")
	firstKey := createWaitingListKey(firstBLS)
	jailedBLS := []byte("jailedBLS")
	jailedKey := createWaitingListKey(jailedBLS)

	m := make(map[string]interface{})
	m[string(firstKey)] = &ElementInList{firstBLS, firstKey, nil}
	waitingListHead := &WaitingList{firstKey, firstKey, 1, nil}
	m[waitingListHeadKey] = waitingListHead

	marshalizer := &marshal.JsonMarshalizer{}

	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		obj, ok := m[string(index)]
		if ok {
			return marshalizer.Marshal(obj)
		}

		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})

	args := createMockStakingScArguments()
	args.Marshalizer = marshalizer
	args.Eei = eei
	sc, _ := NewStakingSmartContract(args)
	_ = sc.flagCorrectFirstQueued.SetReturningPrevious()
	err := sc.insertAfterLastJailed(waitingListHead, jailedBLS)
	assert.Nil(t, err)

	wlh, err := sc.getWaitingListHead()
	assert.Nil(t, err)
	assert.NotNil(t, wlh)
	assert.Equal(t, jailedKey, wlh.FirstKey)
	assert.Equal(t, jailedKey, wlh.LastJailedKey)
	// increase is done in the calling method
	assert.Equal(t, uint32(1), wlh.Length)

	firstElement, err := sc.getWaitingListElement(wlh.FirstKey)
	assert.Nil(t, err)
	assert.NotNil(t, firstElement)
	assert.Equal(t, jailedBLS, firstElement.BLSPublicKey)
	assert.Equal(t, jailedKey, firstElement.PreviousKey)
	assert.Equal(t, firstKey, firstElement.NextKey)

	previousFirstElement, err := sc.getWaitingListElement(firstElement.NextKey)
	assert.Nil(t, err)
	assert.NotNil(t, previousFirstElement)
	assert.Equal(t, firstBLS, previousFirstElement.BLSPublicKey)
	assert.Equal(t, jailedKey, previousFirstElement.PreviousKey)
	assert.Nil(t, previousFirstElement.NextKey)
}

func TestStakingSc_InsertAfterLastJailedAfterFixWithEmptyQueue(t *testing.T) {
	t.Parallel()

	jailedBLS := []byte("jailedBLS")
	jailedKey := createWaitingListKey(jailedBLS)

	m := make(map[string]interface{})
	waitingListHead := &WaitingList{nil, nil, 0, nil}
	m[waitingListHeadKey] = waitingListHead

	marshalizer := &marshal.JsonMarshalizer{}

	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		obj, ok := m[string(index)]
		if ok {
			return marshalizer.Marshal(obj)
		}

		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})

	args := createMockStakingScArguments()
	args.Marshalizer = marshalizer
	args.Eei = eei
	sc, _ := NewStakingSmartContract(args)
	_ = sc.flagCorrectFirstQueued.SetReturningPrevious()
	err := sc.insertAfterLastJailed(waitingListHead, jailedBLS)
	assert.Nil(t, err)

	wlh, err := sc.getWaitingListHead()
	assert.Nil(t, err)
	assert.NotNil(t, wlh)
	assert.Equal(t, jailedKey, wlh.FirstKey)
	assert.Equal(t, jailedKey, wlh.LastJailedKey)

	firstElement, err := sc.getWaitingListElement(wlh.FirstKey)
	assert.Nil(t, err)
	assert.NotNil(t, firstElement)
	assert.Equal(t, jailedBLS, firstElement.BLSPublicKey)
	assert.Equal(t, jailedKey, firstElement.PreviousKey)
	assert.Equal(t, 0, len(firstElement.NextKey))
}

func TestStakingSc_getWaitingListRegisterNonceAndRewardAddressWhenLengthIsHigherThanOne(t *testing.T) {
	t.Parallel()

	waitingBlsKeys := [][]byte{
		[]byte("waitingBlsKey1"),
		[]byte("waitingBlsKey2"),
		[]byte("waitingBlsKey3"),
	}
	sc, eei, marshalizer, stakingAccessAddress := makeWrongConfigForWaitingBlsKeysList(t, waitingBlsKeys)
	alterWaitingListLength(t, eei, marshalizer)

	arguments := CreateVmContractCallInput()
	arguments.Function = "getQueueRegisterNonceAndRewardAddress"
	arguments.CallerAddr = stakingAccessAddress
	arguments.Arguments = make([][]byte, 0)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
	assert.Equal(t, 3*len(waitingBlsKeys), len(eei.output))
	for i, waitingKey := range waitingBlsKeys {
		assert.Equal(t, waitingKey, eei.output[i*3])
	}
}

func TestStakingSc_fixWaitingListQueueSize(t *testing.T) {
	t.Parallel()

	t.Run("inactive fix should error", func(t *testing.T) {
		waitingBlsKeys := [][]byte{
			[]byte("waitingBlsKey1"),
			[]byte("waitingBlsKey2"),
			[]byte("waitingBlsKey3"),
		}
		sc, eei, marshalizer, _ := makeWrongConfigForWaitingBlsKeysList(t, waitingBlsKeys)
		alterWaitingListLength(t, eei, marshalizer)
		sc.flagCorrectFirstQueued.Reset()
		eei.SetGasProvided(500000000)

		arguments := CreateVmContractCallInput()
		arguments.Function = "fixWaitingListQueueSize"
		arguments.CallerAddr = []byte("caller")
		arguments.Arguments = make([][]byte, 0)
		arguments.CallValue = big.NewInt(0)

		retCode := sc.Execute(arguments)
		assert.Equal(t, vmcommon.UserError, retCode)
		assert.Equal(t, "invalid method to call", eei.returnMessage)
	})
	t.Run("provided value should error", func(t *testing.T) {
		waitingBlsKeys := [][]byte{
			[]byte("waitingBlsKey1"),
			[]byte("waitingBlsKey2"),
			[]byte("waitingBlsKey3"),
		}
		sc, eei, marshalizer, _ := makeWrongConfigForWaitingBlsKeysList(t, waitingBlsKeys)
		alterWaitingListLength(t, eei, marshalizer)
		eei.SetGasProvided(500000000)

		arguments := CreateVmContractCallInput()
		arguments.Function = "fixWaitingListQueueSize"
		arguments.CallerAddr = []byte("caller")
		arguments.Arguments = make([][]byte, 0)
		arguments.CallValue = big.NewInt(1)

		retCode := sc.Execute(arguments)
		assert.Equal(t, vmcommon.UserError, retCode)
		assert.Equal(t, vm.TransactionValueMustBeZero, eei.returnMessage)
	})
	t.Run("not enough gas should error", func(t *testing.T) {
		waitingBlsKeys := [][]byte{
			[]byte("waitingBlsKey1"),
			[]byte("waitingBlsKey2"),
			[]byte("waitingBlsKey3"),
		}
		sc, eei, marshalizer, _ := makeWrongConfigForWaitingBlsKeysList(t, waitingBlsKeys)
		alterWaitingListLength(t, eei, marshalizer)
		eei.SetGasProvided(499999999)

		arguments := CreateVmContractCallInput()
		arguments.Function = "fixWaitingListQueueSize"
		arguments.CallerAddr = []byte("caller")
		arguments.Arguments = make([][]byte, 0)
		arguments.CallValue = big.NewInt(0)

		retCode := sc.Execute(arguments)
		assert.Equal(t, vmcommon.OutOfGas, retCode)
		assert.Equal(t, "insufficient gas", eei.returnMessage)
	})
	t.Run("should repair", func(t *testing.T) {
		waitingBlsKeys := [][]byte{
			[]byte("waitingBlsKey1"),
			[]byte("waitingBlsKey2"),
			[]byte("waitingBlsKey3"),
		}
		sc, eei, marshalizer, _ := makeWrongConfigForWaitingBlsKeysList(t, waitingBlsKeys)
		alterWaitingListLength(t, eei, marshalizer)
		eei.SetGasProvided(500000000)

		arguments := CreateVmContractCallInput()
		arguments.Function = "fixWaitingListQueueSize"
		arguments.CallerAddr = []byte("caller")
		arguments.Arguments = make([][]byte, 0)
		arguments.CallValue = big.NewInt(0)

		retCode := sc.Execute(arguments)
		assert.Equal(t, vmcommon.Ok, retCode)

		buff := eei.GetStorage([]byte(waitingListHeadKey))
		waitingListHead := &WaitingList{}
		err := marshalizer.Unmarshal(waitingListHead, buff)
		require.Nil(t, err)

		assert.Equal(t, len(waitingBlsKeys), int(waitingListHead.Length))
		assert.Equal(t, waitingBlsKeys[len(waitingBlsKeys)-1], waitingListHead.LastKey[2:])
		assert.Equal(t, waitingBlsKeys[0], waitingListHead.FirstKey[2:])
	})
	t.Run("should not alter if repair is not needed", func(t *testing.T) {
		waitingBlsKeys := [][]byte{
			[]byte("waitingBlsKey1"),
			[]byte("waitingBlsKey2"),
			[]byte("waitingBlsKey3"),
		}
		sc, eei, marshalizer, _ := makeWrongConfigForWaitingBlsKeysList(t, waitingBlsKeys)
		eei.SetGasProvided(500000000)

		arguments := CreateVmContractCallInput()
		arguments.Function = "fixWaitingListQueueSize"
		arguments.CallerAddr = []byte("caller")
		arguments.Arguments = make([][]byte, 0)
		arguments.CallValue = big.NewInt(0)

		retCode := sc.Execute(arguments)
		assert.Equal(t, vmcommon.Ok, retCode)

		buff := eei.GetStorage([]byte(waitingListHeadKey))
		waitingListHead := &WaitingList{}
		err := marshalizer.Unmarshal(waitingListHead, buff)
		require.Nil(t, err)

		assert.Equal(t, len(waitingBlsKeys), int(waitingListHead.Length))
		assert.Equal(t, waitingBlsKeys[len(waitingBlsKeys)-1], waitingListHead.LastKey[2:])
		assert.Equal(t, waitingBlsKeys[0], waitingListHead.FirstKey[2:])
	})
	t.Run("should not alter if the waiting list size is 1", func(t *testing.T) {
		waitingBlsKeys := [][]byte{
			[]byte("waitingBlsKey1"),
		}
		sc, eei, marshalizer, _ := makeWrongConfigForWaitingBlsKeysList(t, waitingBlsKeys)
		eei.SetGasProvided(500000000)

		arguments := CreateVmContractCallInput()
		arguments.Function = "fixWaitingListQueueSize"
		arguments.CallerAddr = []byte("caller")
		arguments.Arguments = make([][]byte, 0)
		arguments.CallValue = big.NewInt(0)

		retCode := sc.Execute(arguments)
		assert.Equal(t, vmcommon.Ok, retCode)

		buff := eei.GetStorage([]byte(waitingListHeadKey))
		waitingListHead := &WaitingList{}
		err := marshalizer.Unmarshal(waitingListHead, buff)
		require.Nil(t, err)

		assert.Equal(t, len(waitingBlsKeys), int(waitingListHead.Length))
		assert.Equal(t, waitingBlsKeys[len(waitingBlsKeys)-1], waitingListHead.LastKey[2:])
		assert.Equal(t, waitingBlsKeys[0], waitingListHead.FirstKey[2:])
	})
	t.Run("should not alter if the waiting list size is 0", func(t *testing.T) {
		waitingBlsKeys := make([][]byte, 0)
		sc, eei, marshalizer, _ := makeWrongConfigForWaitingBlsKeysList(t, waitingBlsKeys)
		eei.SetGasProvided(500000000)

		arguments := CreateVmContractCallInput()
		arguments.Function = "fixWaitingListQueueSize"
		arguments.CallerAddr = []byte("caller")
		arguments.Arguments = make([][]byte, 0)
		arguments.CallValue = big.NewInt(0)

		retCode := sc.Execute(arguments)
		assert.Equal(t, vmcommon.Ok, retCode)

		buff := eei.GetStorage([]byte(waitingListHeadKey))
		waitingListHead := &WaitingList{}
		err := marshalizer.Unmarshal(waitingListHead, buff)
		require.Nil(t, err)

		assert.Equal(t, len(waitingBlsKeys), int(waitingListHead.Length))
		assert.Nil(t, waitingListHead.LastKey)
		assert.Nil(t, waitingListHead.FirstKey)
	})
	t.Run("should not alter lastJailedKey if exists", func(t *testing.T) {
		lastJailedBLSString := "lastJailedKey1"
		waitingBlsKeys := [][]byte{
			[]byte(lastJailedBLSString),
			[]byte("waitingBlsKey2"),
		}
		lastJailedKey := []byte(fmt.Sprintf("w_%s", lastJailedBLSString))
		sc, eei, marshalizer, _ := makeWrongConfigForWaitingBlsKeysListWithLastJailed(t, waitingBlsKeys, lastJailedKey)
		eei.SetGasProvided(500000000)

		arguments := CreateVmContractCallInput()
		arguments.Function = "fixWaitingListQueueSize"
		arguments.CallerAddr = []byte("caller")
		arguments.Arguments = make([][]byte, 0)
		arguments.CallValue = big.NewInt(0)

		beforeBuff := eei.GetStorage([]byte(waitingListHeadKey))
		beforeWaitingListHead := &WaitingList{}
		beforeErr := marshalizer.Unmarshal(beforeWaitingListHead, beforeBuff)
		require.Nil(t, beforeErr)
		assert.Equal(t, lastJailedKey, beforeWaitingListHead.LastJailedKey)

		retCode := sc.Execute(arguments)
		assert.Equal(t, vmcommon.Ok, retCode)

		buff := eei.GetStorage([]byte(waitingListHeadKey))
		waitingListHead := &WaitingList{}
		err := marshalizer.Unmarshal(waitingListHead, buff)
		require.Nil(t, err)

		assert.Equal(t, len(waitingBlsKeys), int(waitingListHead.Length))
		assert.Equal(t, lastJailedKey, waitingListHead.LastJailedKey)
	})
	t.Run("should alter lastJailedKey if NOT exists", func(t *testing.T) {
		waitingBlsKeys := [][]byte{
			[]byte("waitingBlsKey1"),
			[]byte("waitingBlsKey2"),
		}
		lastJailedKey := []byte("lastJailedKey")
		sc, eei, marshalizer, _ := makeWrongConfigForWaitingBlsKeysListWithLastJailed(t, waitingBlsKeys, lastJailedKey)
		eei.SetGasProvided(500000000)

		arguments := CreateVmContractCallInput()
		arguments.Function = "fixWaitingListQueueSize"
		arguments.CallerAddr = []byte("caller")
		arguments.Arguments = make([][]byte, 0)
		arguments.CallValue = big.NewInt(0)

		beforeBuff := eei.GetStorage([]byte(waitingListHeadKey))
		beforeWaitingListHead := &WaitingList{}
		beforeErr := marshalizer.Unmarshal(beforeWaitingListHead, beforeBuff)
		require.Nil(t, beforeErr)
		assert.Equal(t, lastJailedKey, beforeWaitingListHead.LastJailedKey)

		retCode := sc.Execute(arguments)
		assert.Equal(t, vmcommon.Ok, retCode)

		buff := eei.GetStorage([]byte(waitingListHeadKey))
		waitingListHead := &WaitingList{}
		err := marshalizer.Unmarshal(waitingListHead, buff)
		require.Nil(t, err)

		assert.Equal(t, len(waitingBlsKeys), int(waitingListHead.Length))
		assert.Equal(t, 0, len(waitingListHead.LastJailedKey))
	})
}

func makeWrongConfigForWaitingBlsKeysList(t *testing.T, waitingBlsKeys [][]byte) (*stakingSC, *vmContext, marshal.Marshalizer, []byte) {
	return makeWrongConfigForWaitingBlsKeysListWithLastJailed(t, waitingBlsKeys, nil)
}

func makeWrongConfigForWaitingBlsKeysListWithLastJailed(t *testing.T, waitingBlsKeys [][]byte, lastJailedKey []byte) (*stakingSC, *vmContext, marshal.Marshalizer, []byte) {
	blockChainHook := &mock.BlockChainHookStub{}
	marshalizer := &marshal.JsonMarshalizer{}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{}, &stateMock.AccountsStub{}, &mock.RaterMock{})
	m := make(map[string]interface{})
	waitingListHead := &WaitingList{nil, nil, 0, lastJailedKey}
	m[waitingListHeadKey] = waitingListHead

	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		obj, found := m[string(index)]
		if found {
			return marshalizer.Marshal(obj)
		}

		return nil, nil
	}

	args := createMockStakingScArguments()
	args.Marshalizer = marshalizer
	args.Eei = eei
	stakingAccessAddress := []byte("stakingAccessAddress")
	args.StakingAccessAddr = stakingAccessAddress
	args.StakingSCConfig.MaxNumberOfNodesForStake = 2
	args.GasCost.MetaChainSystemSCsCost.FixWaitingListSize = 500000000
	sc, _ := NewStakingSmartContract(args)
	_ = sc.flagCorrectFirstQueued.SetReturningPrevious()
	stakerAddress := []byte("stakerAddr")

	doStake(t, sc, stakingAccessAddress, stakerAddress, []byte("eligibleBlsKey1"))
	doStake(t, sc, stakingAccessAddress, stakerAddress, []byte("eligibleBlsKey2"))
	for _, waitingKey := range waitingBlsKeys {
		doStake(t, sc, stakingAccessAddress, stakerAddress, waitingKey)
	}

	eei.output = make([][]byte, 0)
	eei.returnMessage = ""

	return sc, eei, marshalizer, stakingAccessAddress
}

func alterWaitingListLength(t *testing.T, eei *vmContext, marshalizer marshal.Marshalizer) {
	// manually alter the length
	buff := eei.GetStorage([]byte(waitingListHeadKey))
	existingWaitingListHead := &WaitingList{}
	err := marshalizer.Unmarshal(existingWaitingListHead, buff)
	require.Nil(t, err)
	existingWaitingListHead.Length++
	buff, err = marshalizer.Marshal(existingWaitingListHead)
	require.Nil(t, err)
	eei.SetStorage([]byte(waitingListHeadKey), buff)
}

func doUnStakeAtEndOfEpoch(t *testing.T, sc *stakingSC, blsKey []byte, expectedReturnCode vmcommon.ReturnCode) {
	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = sc.endOfEpochAccessAddr
	arguments.Function = "unStakeAtEndOfEpoch"
	arguments.Arguments = [][]byte{blsKey}

	retCode := sc.Execute(arguments)
	assert.Equal(t, expectedReturnCode, retCode)
}

func doGetRewardAddress(t *testing.T, sc *stakingSC, eei *vmContext, blsKey []byte, expectedAddress string) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "getRewardAddress"
	arguments.Arguments = [][]byte{blsKey}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	lastOutput := eei.output[len(eei.output)-1]
	assert.True(t, bytes.Equal(lastOutput, []byte(hex.EncodeToString([]byte(expectedAddress)))))
}

func doGetRemainingUnbondPeriod(t *testing.T, sc *stakingSC, eei *vmContext, blsKey []byte, expected int) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "getRemainingUnBondPeriod"
	arguments.Arguments = [][]byte{blsKey}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	lastOutput := eei.output[len(eei.output)-1]
	assert.True(t, bytes.Equal(lastOutput, big.NewInt(int64(expected)).Bytes()))
}

func doGetStatus(t *testing.T, sc *stakingSC, eei *vmContext, blsKey []byte, expectedStatus string) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "getBLSKeyStatus"
	arguments.Arguments = [][]byte{blsKey}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	lastOutput := eei.output[len(eei.output)-1]
	assert.True(t, bytes.Equal(lastOutput, []byte(expectedStatus)))
}

func doGetWaitingListSize(t *testing.T, sc *stakingSC, eei *vmContext, expectedSize int) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "getQueueSize"

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	lastOutput := eei.output[len(eei.output)-1]
	assert.True(t, bytes.Equal(lastOutput, []byte(strconv.Itoa(expectedSize))))
}

func doGetWaitingListRegisterNonceAndRewardAddress(t *testing.T, sc *stakingSC, eei *vmContext) [][]byte {
	arguments := CreateVmContractCallInput()
	arguments.Function = "getQueueRegisterNonceAndRewardAddress"
	arguments.CallerAddr = sc.stakeAccessAddr

	currentOutPutIndex := len(eei.output)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	return eei.output[currentOutPutIndex:]
}

func doGetWaitingListIndex(t *testing.T, sc *stakingSC, eei *vmContext, blsKey []byte, expectedCode vmcommon.ReturnCode, expectedIndex int) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "getQueueIndex"
	arguments.CallerAddr = sc.stakeAccessAddr
	arguments.Arguments = [][]byte{blsKey}

	retCode := sc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)

	lastOutput := eei.output[len(eei.output)-1]
	assert.True(t, bytes.Equal(lastOutput, []byte(strconv.Itoa(expectedIndex))))
}

func doUnJail(t *testing.T, sc *stakingSC, callerAddr, addrToUnJail []byte, expectedCode vmcommon.ReturnCode) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "unJail"
	arguments.CallerAddr = callerAddr
	arguments.Arguments = [][]byte{addrToUnJail}

	retCode := sc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)
}

func doJail(t *testing.T, sc *stakingSC, callerAddr, addrToJail []byte, expectedCode vmcommon.ReturnCode) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "jail"
	arguments.CallerAddr = callerAddr
	arguments.Arguments = [][]byte{addrToJail}

	retCode := sc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)
}

func doStake(t *testing.T, sc *stakingSC, callerAddr, stakerAddr, stakerPubKey []byte) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = callerAddr
	arguments.Arguments = [][]byte{stakerPubKey, stakerAddr, stakerAddr}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func doUnStake(t *testing.T, sc *stakingSC, callerAddr, stakerAddr, stakerPubKey []byte, expectedCode vmcommon.ReturnCode) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.CallerAddr = callerAddr
	arguments.Arguments = [][]byte{stakerPubKey, stakerAddr}

	retCode := sc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)
}

func doUnBond(t *testing.T, sc *stakingSC, callerAddr, stakerPubKey []byte, expectedCode vmcommon.ReturnCode) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "unBond"
	arguments.CallerAddr = callerAddr
	arguments.Arguments = [][]byte{stakerPubKey}

	retCode := sc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)
}

func doSwitchJailedWithWaiting(t *testing.T, sc *stakingSC, pubKey []byte) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "switchJailedWithWaiting"
	arguments.CallerAddr = sc.endOfEpochAccessAddr
	arguments.Arguments = [][]byte{pubKey}
	retCode := sc.Execute(arguments)
	assert.Equal(t, retCode, vmcommon.Ok)
}

func checkIsStaked(t *testing.T, sc *stakingSC, callerAddr, stakerPubKey []byte, expectedCode vmcommon.ReturnCode) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "isStaked"
	arguments.CallerAddr = callerAddr
	arguments.Arguments = [][]byte{stakerPubKey}

	retCode := sc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)
}

func TestStakingSc_fixMissingNodeOnQueue(t *testing.T) {
	t.Parallel()

	waitingBlsKeys := [][]byte{
		[]byte("waitingBlsKey1"),
		[]byte("waitingBlsKey2"),
		[]byte("waitingBlsKey3"),
	}
	sc, eei, _, stakingAccessAddress := makeWrongConfigForWaitingBlsKeysList(t, waitingBlsKeys)

	arguments := CreateVmContractCallInput()
	arguments.Function = "addMissingNodeToQueue"
	arguments.CallerAddr = bytes.Repeat([]byte{1}, 32)
	arguments.Arguments = make([][]byte, 0)

	eei.returnMessage = ""
	sc.flagCorrectFirstQueued.Reset()
	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, "invalid method to call", eei.returnMessage)

	eei.returnMessage = ""
	_ = sc.flagCorrectFirstQueued.SetReturningPrevious()
	arguments.CallValue = big.NewInt(10)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, vm.TransactionValueMustBeZero, eei.returnMessage)

	eei.gasRemaining = 1
	sc.gasCost.MetaChainSystemSCsCost.FixWaitingListSize = 50
	eei.returnMessage = ""
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.OutOfGas, retCode)
	assert.Equal(t, "insufficient gas", eei.returnMessage)

	eei.gasRemaining = 50
	eei.returnMessage = ""
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, "invalid number of arguments", eei.returnMessage)

	eei.gasRemaining = 50
	eei.returnMessage = ""
	arguments.Arguments = append(arguments.Arguments, []byte("waitingBlsKey4"))
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, "element was not found", eei.returnMessage)

	doStake(t, sc, stakingAccessAddress, arguments.CallerAddr, []byte("waitingBlsKey4"))

	eei.gasRemaining = 50
	eei.returnMessage = ""
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, "key is in queue, not missing", eei.returnMessage)
}

func TestStakingSc_fixMissingNodeAddOneNodeOnly(t *testing.T) {
	t.Parallel()

	sc, eei, _, _ := makeWrongConfigForWaitingBlsKeysList(t, nil)

	arguments := CreateVmContractCallInput()
	arguments.Function = "addMissingNodeToQueue"
	arguments.CallerAddr = bytes.Repeat([]byte{1}, 32)
	arguments.Arguments = make([][]byte, 0)

	blsKey := []byte("waitingBlsKey1")
	eei.returnMessage = ""
	arguments.Arguments = append(arguments.Arguments, blsKey)
	eei.gasRemaining = 50

	sc.gasCost.MetaChainSystemSCsCost.FixWaitingListSize = 50
	_ = sc.saveWaitingListElement(createWaitingListKey(blsKey), &ElementInList{BLSPublicKey: blsKey})

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	waitingListData, _ := sc.getFirstElementsFromWaitingList(50)
	assert.Equal(t, len(waitingListData.blsKeys), 1)
	assert.Equal(t, waitingListData.blsKeys[0], blsKey)
}

func TestStakingSc_fixMissingNodeAddAsLast(t *testing.T) {
	t.Parallel()

	waitingBlsKeys := [][]byte{
		[]byte("waitingBlsKey1"),
		[]byte("waitingBlsKey2"),
		[]byte("waitingBlsKey3"),
	}
	sc, eei, _, _ := makeWrongConfigForWaitingBlsKeysList(t, waitingBlsKeys)
	sc.gasCost.MetaChainSystemSCsCost.FixWaitingListSize = 50

	arguments := CreateVmContractCallInput()
	arguments.Function = "addMissingNodeToQueue"
	arguments.CallerAddr = bytes.Repeat([]byte{1}, 32)
	arguments.Arguments = make([][]byte, 0)

	blsKey := []byte("waitingBlsKey4")
	eei.returnMessage = ""
	arguments.Arguments = append(arguments.Arguments, blsKey)
	eei.gasRemaining = 50
	_ = sc.saveWaitingListElement(createWaitingListKey(blsKey), &ElementInList{BLSPublicKey: blsKey})

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	waitingListData, _ := sc.getFirstElementsFromWaitingList(50)
	assert.Equal(t, len(waitingListData.blsKeys), 4)
	assert.Equal(t, waitingListData.blsKeys[3], blsKey)
}
