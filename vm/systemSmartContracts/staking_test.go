package systemSmartContracts

import (
	"bytes"
	"encoding/json"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"math"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func createMockStakingScArguments() ArgsNewStakingSmartContract {
	return ArgsNewStakingSmartContract{
		MinStakeValue:            big.NewInt(100),
		UnBondPeriod:             0,
		Eei:                      &mock.SystemEIStub{},
		StakingAccessAddr:        []byte("auction"),
		JailAccessAddr:           []byte("jail"),
		NumRoundsWithoutBleed:    0,
		BleedPercentagePerRound:  0,
		MaximumPercentageToBleed: 0,
	}
}

func CreateVmContractCallInput() *vmcommon.ContractCallInput {
	return &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("auction"),
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    0,
			GasProvided: 0,
		},
		RecipientAddr: []byte("rcpntaddr"),
		Function:      "something",
	}
}

func TestNewStakingSmartContract_NilStakeValueShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockStakingScArguments()
	args.MinStakeValue = nil
	stakingSmartContract, err := NewStakingSmartContract(args)

	assert.Nil(t, stakingSmartContract)
	assert.Equal(t, vm.ErrNilInitialStakeValue, err)
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

func TestNewStakingSmartContract_NegativeStakeValueShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockStakingScArguments()
	args.MinStakeValue = big.NewInt(-100)
	stakingSmartContract, err := NewStakingSmartContract(args)

	assert.Nil(t, stakingSmartContract)
	assert.Equal(t, vm.ErrNegativeInitialStakeValue, err)
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
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	eei.SetSCAddress([]byte("addr"))
	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "_init"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	ownerAddr := stakingSmartContract.eei.GetStorage([]byte(OwnerKey))
	assert.Equal(t, arguments.CallerAddr, ownerAddr)

	ownerBalanceBytes := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	ownerBalance := big.NewInt(0).SetBytes(ownerBalanceBytes)
	assert.Equal(t, big.NewInt(0), ownerBalance)

}

func TestStakingSC_ExecuteInitTwoTimeShouldReturnUserError(t *testing.T) {
	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	eei.SetSCAddress([]byte("addr"))
	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "_init"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteStakeWrongStakeValueShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	eei.SetSCAddress([]byte("addr"))
	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
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
	args.MinStakeValue = stakeValue
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
		registrationDataMarshalized, _ := json.Marshal(&StakedData{Staked: true})
		return registrationDataMarshalized
	}
	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
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
		registrationDataMarshalized, _ := json.Marshal(&StakedData{})
		return registrationDataMarshalized
	}
	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
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
	expectedRegistrationData := StakedData{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: []byte{100},
		StakeValue:    big.NewInt(0).Set(stakeValue),
		JailedRound:   math.MaxUint64,
	}

	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = []byte("auction")
	arguments.Arguments = [][]byte{stakerPubKey.Bytes(), stakerAddress.Bytes()}
	arguments.CallValue = big.NewInt(100)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData StakedData
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
	args.MinStakeValue = stakeValue
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
	args.MinStakeValue = stakeValue
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake@abc"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteUnStakeAlreadyUnStakedAddrShouldErr(t *testing.T) {
	t.Parallel()

	stakedRegistrationData := StakedData{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		RewardAddress: nil,
		StakeValue:    nil,
	}

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
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

	stakedRegistrationData := StakedData{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: expectedCallerAddress,
		StakeValue:    nil,
	}

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
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

func TestStakingSC_ExecuteUnStake(t *testing.T) {
	t.Parallel()

	callerAddress := []byte("caller")

	expectedRegistrationData := StakedData{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		RewardAddress: callerAddress,
		StakeValue:    nil,
		JailedRound:   math.MaxUint64,
	}

	stakedRegistrationData := StakedData{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: callerAddress,
		StakeValue:    nil,
		JailedRound:   math.MaxUint64,
	}

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{[]byte("abc"), callerAddress}
	arguments.CallerAddr = []byte("auction")
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData StakedData
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
	args.MinStakeValue = stakeValue
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
		case bytes.Equal(key, []byte(OwnerKey)):
			return []byte("data")
		default:
			registrationDataMarshalized, _ := json.Marshal(&StakedData{UnStakedNonce: 0})
			return registrationDataMarshalized
		}
	}
	eei.BlockChainHookCalled = func() vmcommon.BlockchainHook {
		return &mock.BlockChainHookStub{CurrentNonceCalled: func() uint64 {
			return 10000
		}}
	}
	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
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
	registrationData := StakedData{
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
	}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	eei.SetSCAddress([]byte("addr"))
	eei.SetStorage([]byte(OwnerKey), []byte("data"))
	eei.SetStorage(blsPubKey.Bytes(), marshalizedRegData)
	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "finalizeUnStake"
	arguments.Arguments = [][]byte{blsPubKey.Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteUnBound(t *testing.T) {
	t.Parallel()

	unBondPeriod := uint64(100)
	unstakedNonce := uint64(10)
	registrationData := StakedData{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: unstakedNonce,
		RewardAddress: []byte("auction"),
		StakeValue:    big.NewInt(100),
		JailedRound:   math.MaxUint64,
	}

	stakeValue := big.NewInt(100)
	marshalizedRegData, _ := json.Marshal(&registrationData)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return unstakedNonce + unBondPeriod + 1
		},
	}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	scAddress := []byte("owner")
	eei.SetSCAddress(scAddress)
	eei.SetStorage([]byte(OwnerKey), scAddress)

	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("auction")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{[]byte("abc")}

	stakingSmartContract.eei.SetStorage(arguments.Arguments[0], marshalizedRegData)

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
	args.MinStakeValue = stakeValue
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
	args.MinStakeValue = stakeValue
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
	args.MinStakeValue = stakeValue
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
		case bytes.Equal(key, []byte(OwnerKey)):
			return []byte("data")
		default:
			registrationDataMarshalized, _ := json.Marshal(&StakedData{StakeValue: big.NewInt(100)})
			return registrationDataMarshalized
		}
	}

	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"
	arguments.CallerAddr = []byte("data")
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes(), big.NewInt(100).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteSlashStaked(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		switch {
		case bytes.Equal(key, []byte(OwnerKey)):
			return []byte("data")
		default:
			registrationDataMarshalized, _ := json.Marshal(&StakedData{StakeValue: big.NewInt(100), Staked: true, RewardAddress: []byte("reward")})
			return registrationDataMarshalized
		}
	}

	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"
	arguments.CallerAddr = []byte("data")
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes(), big.NewInt(100).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
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
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})

	smartcontractAddress := "smartcontractAddress"
	eei.SetSCAddress([]byte(smartcontractAddress))

	ownerAddress := "ownerAddress"
	eei.SetStorage([]byte(OwnerKey), []byte(ownerAddress))

	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Arguments = [][]byte{stakerPubKey, stakerAddress}
	arguments.CallerAddr = []byte("auction")

	stakedRegistrationData := StakedData{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: stakerAddress,
		StakeValue:    valueStakedByTheCaller,
		JailedRound:   math.MaxUint64,
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

	var registrationData StakedData
	data := stakingSmartContract.eei.GetStorage(arguments.Arguments[0])
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)

	expectedRegistrationData := StakedData{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: unStakeNonce,
		RewardAddress: stakerAddress,
		StakeValue:    valueStakedByTheCaller,
		JailedRound:   math.MaxUint64,
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
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
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
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	err := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.Ok, err)
}

func TestStakingSc_ExecuteSlashTwoTime(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})

	stakedRegistrationData := StakedData{
		RegisterNonce: 50,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: []byte("auction"),
		StakeValue:    stakeValue,
		JailedRound:   math.MaxUint64,
	}

	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"
	arguments.CallerAddr = []byte("data")
	marshalizedStakedData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.CallerAddr, marshalizedStakedData)
	stakingSmartContract.eei.SetStorage([]byte(OwnerKey), arguments.CallerAddr)

	slashValue := big.NewInt(70)
	arguments.Arguments = [][]byte{arguments.CallerAddr, slashValue.Bytes()}
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	dataBytes := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	var registrationData StakedData
	err := json.Unmarshal(dataBytes, &registrationData)
	assert.Nil(t, err)

	expectedStake := big.NewInt(0).Sub(stakeValue, slashValue)
	assert.Equal(t, expectedStake, registrationData.StakeValue)

	arguments.Arguments = [][]byte{arguments.CallerAddr, slashValue.Bytes()}
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	dataBytes = stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	err = json.Unmarshal(dataBytes, &registrationData)
	assert.Nil(t, err)

	expectedStake = big.NewInt(0).Sub(expectedStake, slashValue)
	assert.Equal(t, expectedStake, registrationData.StakeValue)
}

func TestStakingSc_ExecuteNilArgs(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})

	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	retCode := stakingSmartContract.Execute(nil)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func doStake(t *testing.T, sc *stakingSC, callerAddr, stakerAddr, stakerPubKey []byte, stakeValue *big.Int) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = callerAddr
	arguments.Arguments = [][]byte{stakerPubKey, stakerAddr}
	arguments.CallValue = stakeValue

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func doUnStake(t *testing.T, sc *stakingSC, callerAddr, stakerAddr, stakerPubKey []byte) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.CallerAddr = callerAddr
	arguments.Arguments = [][]byte{stakerPubKey, stakerAddr}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func checkIsStaked(t *testing.T, sc *stakingSC, callerAddr, stakerPubKey []byte, expectedCode vmcommon.ReturnCode) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "isStaked"
	arguments.CallerAddr = callerAddr
	arguments.Arguments = [][]byte{stakerPubKey}

	retCode := sc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)
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

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.StakingAccessAddr = stakingAccessAddress
	args.MinStakeValue = stakeValue
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
	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, stakerPubKey, stakeValue)
	// check again isStaked should return vmcommon.Ok
	checkIsStaked(t, stakingSmartContract, callerAddress, stakerPubKey, vmcommon.Ok)
	//do unStake
	doUnStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, stakerPubKey)
	// check if account is staked should return error code
	checkIsStaked(t, stakingSmartContract, callerAddress, stakerPubKey, vmcommon.UserError)
}

func setStakeValueCurrentEpoch(t *testing.T, sc *stakingSC, callerAddr []byte, stakeValue *big.Int, expectedCode vmcommon.ReturnCode) {
	arguments := CreateVmContractCallInput()

	arguments.CallerAddr = callerAddr
	arguments.Function = "setStakeValue"
	if stakeValue != nil {
		arguments.Arguments = [][]byte{stakeValue.Bytes()}
	} else {
		arguments.Arguments = nil
	}

	retCode := sc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)
}

func TestStakingSc_SetStakeValueForCurrentEpoch(t *testing.T) {
	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingAccessAddress := []byte("stakingAccessAddress")
	args := createMockStakingScArguments()
	args.MinStakeValue = stakeValue
	args.StakingAccessAddr = stakingAccessAddress
	args.Eei = eei
	stakingSmartContract, _ := NewStakingSmartContract(args)

	stakerAddress := []byte("stakerAddr")
	stakerPubKey := []byte("stakerPublicKey")

	//setStakeValueCurrentEpoch will return UserError -> wrong caller address
	currentEpochStakeValue := big.NewInt(1000)
	callerAddress := []byte("data")
	setStakeValueCurrentEpoch(t, stakingSmartContract, callerAddress, currentEpochStakeValue, vmcommon.UserError)
	//setStakeValueCurrentEpoch will return UserError -> nil current epoch stake value
	callerAddress = stakingAccessAddress
	setStakeValueCurrentEpoch(t, stakingSmartContract, callerAddress, nil, vmcommon.UserError)
	//setStakeValueCurrentEpoch will return Ok
	setStakeValueCurrentEpoch(t, stakingSmartContract, callerAddress, currentEpochStakeValue, vmcommon.Ok)

	doStake(t, stakingSmartContract, stakingAccessAddress, stakerAddress, stakerPubKey, stakeValue)
}
