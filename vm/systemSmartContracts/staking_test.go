package systemSmartContracts

import (
	"bytes"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

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

	eei := &mock.SystemEIStub{}
	stakingSmartContract, err := NewStakingSmartContract(nil, 0, eei, []byte("auction"))

	assert.Nil(t, stakingSmartContract)
	assert.Equal(t, vm.ErrNilInitialStakeValue, err)
}

func TestNewStakingSmartContract_NilSystemEIShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	stakingSmartContract, err := NewStakingSmartContract(stakeValue, 0, nil, []byte("auction"))

	assert.Nil(t, stakingSmartContract)
	assert.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
}

func TestNewStakingSmartContract_NegativeStakeValueShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(-100)
	eei := &mock.SystemEIStub{}
	stakingSmartContract, err := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))

	assert.Nil(t, stakingSmartContract)
	assert.Equal(t, vm.ErrNegativeInitialStakeValue, err)
}

func TestNewStakingSmartContract(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	stakingSmartContract, err := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))

	assert.NotNil(t, stakingSmartContract)
	assert.Nil(t, err)
}

func TestStakingSC_ExecuteInit(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	eei.SetSCAddress([]byte("addr"))
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
	arguments := CreateVmContractCallInput()
	arguments.Function = "_init"

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
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	eei.SetSCAddress([]byte("addr"))
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
	}

	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		return nil, nil
	}

	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
	}

	stakedRegistrationData := StakedData{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: callerAddress,
		StakeValue:    nil,
	}

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), &mock.ArgumentParserMock{})
	eei.SetSCAddress([]byte("addr"))

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
			registrationDataMarshalized, _ := json.Marshal(&StakedData{UnStakedNonce: 0})
			return registrationDataMarshalized
		}
	}
	eei.BlockChainHookCalled = func() vmcommon.BlockchainHook {
		return &mock.BlockChainHookStub{CurrentNonceCalled: func() uint64 {
			return 10000
		}}
	}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 100, eei, []byte("auction"))
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
	eei.SetStorage([]byte(ownerKey), []byte("data"))
	eei.SetStorage(blsPubKey.Bytes(), marshalizedRegData)
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 100, eei, []byte("auction"))
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
		RewardAddress: nil,
		StakeValue:    big.NewInt(100),
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
	eei.SetStorage([]byte(ownerKey), scAddress)

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, unBondPeriod, eei, []byte("auction"))

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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
			registrationDataMarshalized, _ := json.Marshal(&StakedData{StakeValue: big.NewInt(100)})
			return registrationDataMarshalized
		}
	}

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
		case bytes.Equal(key, []byte(ownerKey)):
			return []byte("data")
		default:
			registrationDataMarshalized, _ := json.Marshal(&StakedData{StakeValue: big.NewInt(100), Staked: true})
			return registrationDataMarshalized
		}
	}

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
	eei.SetStorage([]byte(ownerKey), []byte(ownerAddress))

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, unBondPeriod, eei, []byte("auction"))

	arguments := CreateVmContractCallInput()
	arguments.Arguments = [][]byte{stakerPubKey, stakerAddress}
	arguments.CallerAddr = []byte("auction")

	stakedRegistrationData := StakedData{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: stakerAddress,
		StakeValue:    valueStakedByTheCaller,
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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
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
		RewardAddress: nil,
		StakeValue:    stakeValue,
	}

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei, []byte("auction"))
	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"
	arguments.CallerAddr = []byte("data")
	marshalizedStakedData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.CallerAddr, marshalizedStakedData)
	stakingSmartContract.eei.SetStorage([]byte(ownerKey), arguments.CallerAddr)

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
