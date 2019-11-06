package systemSmartContracts

import (
	"bytes"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func CreateVmContractCallInput() *vmcommon.ContractCallInput {
	return &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("tralala1"),
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    big.NewInt(0),
			GasProvided: big.NewInt(0),
			Header:      &vmcommon.SCCallHeader{},
		},
		RecipientAddr: []byte("tralala2"),
		Function:      "something",
	}
}

func TestNewStakingSmartContract_NilStakeValueShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	stakingSmartContract, err := NewStakingSmartContract(nil, 0, eei)

	assert.Nil(t, stakingSmartContract)
	assert.Equal(t, vm.ErrNilInitialStakeValue, err)
}

func TestNewStakingSmartContract_NilSystemEIShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	stakingSmartContract, err := NewStakingSmartContract(stakeValue, 0, nil)

	assert.Nil(t, stakingSmartContract)
	assert.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
}

func TestNewStakingSmartContract_NegativeStakeValueShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(-100)
	eei := &mock.SystemEIStub{}
	stakingSmartContract, err := NewStakingSmartContract(stakeValue, 0, eei)

	assert.Nil(t, stakingSmartContract)
	assert.Equal(t, vm.ErrNegativeInitialStakeValue, err)
}

func TestNewStakingSmartContract(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	stakingSmartContract, err := NewStakingSmartContract(stakeValue, 0, eei)

	assert.NotNil(t, stakingSmartContract)
	assert.Nil(t, err)
}

func TestStakingSC_ExecuteInit(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, &mock.CryptoHookStub{})
	eei.SetSCAddress([]byte("addr"))
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
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
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, &mock.CryptoHookStub{})
	eei.SetSCAddress([]byte("addr"))
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
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
	eei, _ := NewVMContext(blockChainHook, &mock.CryptoHookStub{})
	eei.SetSCAddress([]byte("addr"))
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
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
		registrationDataMarshalized, _ := json.Marshal(&StakingData{Staked: true})
		return registrationDataMarshalized
	}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
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
		registrationDataMarshalized, _ := json.Marshal(&StakingData{})
		return registrationDataMarshalized
	}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteStake(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)

	expectedRegistrationData := StakingData{
		StartNonce:    0,
		Staked:        true,
		UnStakedNonce: 0,
		BlsPubKey:     []byte{100},
		StakeValue:    big.NewInt(0).Set(stakeValue),
	}

	blockChainHook := &mock.BlockChainHookStub{}
	blockChainHook.GetStorageDataCalled = func(accountsAddress []byte, index []byte) (i []byte, e error) {
		switch {
		case bytes.Equal(index, []byte(initialStakeKey)):
			return stakeValue.Bytes(), nil
		default:
			return nil, nil
		}
	}

	eei, _ := NewVMContext(blockChainHook, &mock.CryptoHookStub{})
	eei.SetSCAddress([]byte("addr"))

	stakerAddress := big.NewInt(100)
	slashValue := big.NewInt(200)
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.Arguments = []*big.Int{stakerAddress, slashValue}
	arguments.CallValue = big.NewInt(100)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData StakingData
	data := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)
	assert.Equal(t, expectedRegistrationData, registrationData)

	receiverBalance := stakingSmartContract.eei.GetBalance(arguments.RecipientAddr)
	senderBalance := stakingSmartContract.eei.GetBalance(arguments.CallerAddr)
	assert.Equal(t, big.NewInt(-100), senderBalance)
	assert.Equal(t, big.NewInt(100), receiverBalance)
}

func TestStakingSC_ExecuteUnStakeAddressNotStakedShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"

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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteUnStakeAlreadyUnStakedAddrShouldErr(t *testing.T) {
	t.Parallel()

	stakedRegistrationData := StakingData{
		StartNonce:    0,
		Staked:        false,
		UnStakedNonce: 0,
		BlsPubKey:     nil,
		StakeValue:    nil,
	}

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, &mock.CryptoHookStub{})
	eei.SetSCAddress([]byte("addr"))

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.Arguments = []*big.Int{big.NewInt(100), big.NewInt(200)}
	arguments.Header = &vmcommon.SCCallHeader{Number: big.NewInt(100)}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.CallerAddr, marshalizedExpectedRegData)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteUnStake(t *testing.T) {
	t.Parallel()

	expectedRegistrationData := StakingData{
		StartNonce:    0,
		Staked:        false,
		UnStakedNonce: 0,
		BlsPubKey:     nil,
		StakeValue:    nil,
	}

	stakedRegistrationData := StakingData{
		StartNonce:    0,
		Staked:        true,
		UnStakedNonce: 0,
		BlsPubKey:     nil,
		StakeValue:    nil,
	}

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, &mock.CryptoHookStub{})
	eei.SetSCAddress([]byte("addr"))

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.Arguments = []*big.Int{big.NewInt(100), big.NewInt(200)}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.CallerAddr, marshalizedExpectedRegData)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData StakingData
	data := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "unBound"
	arguments.Arguments = []*big.Int{big.NewInt(100), big.NewInt(200)}

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
			registrationDataMarshalized, _ := json.Marshal(&StakingData{UnStakedNonce: 0})
			return registrationDataMarshalized
		}
	}
	eei.BlockChainHookCalled = func() vmcommon.BlockchainHook {
		return &mock.BlockChainHookStub{CurrentNonceCalled: func() uint64 {
			return 10000
		}}
	}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 100, eei)
	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "unBound"
	arguments.Arguments = []*big.Int{big.NewInt(100)}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteFinalizeUnBoundBeforePeriodEnds(t *testing.T) {
	t.Parallel()

	unstakedNonce := uint64(10)
	registrationData := StakingData{
		StartNonce:    0,
		Staked:        true,
		UnStakedNonce: unstakedNonce,
		BlsPubKey:     nil,
		StakeValue:    big.NewInt(100),
	}
	blsPubKey := big.NewInt(100)
	stakeValue := big.NewInt(100)
	marshalizedRegData, _ := json.Marshal(&registrationData)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return unstakedNonce + 1
		},
	}, &mock.CryptoHookStub{})
	eei.SetSCAddress([]byte("addr"))
	eei.SetStorage([]byte(ownerKey), []byte("data"))
	eei.SetStorage(blsPubKey.Bytes(), marshalizedRegData)
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 100, eei)
	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "finalizeUnStake"
	arguments.Arguments = []*big.Int{blsPubKey}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteUnBound(t *testing.T) {
	t.Parallel()

	unBoundPeriod := uint64(100)
	unstakedNonce := uint64(10)
	registrationData := StakingData{
		StartNonce:    0,
		Staked:        false,
		UnStakedNonce: unstakedNonce,
		BlsPubKey:     nil,
		StakeValue:    big.NewInt(100),
	}

	stakeValue := big.NewInt(100)
	marshalizedRegData, _ := json.Marshal(&registrationData)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return unstakedNonce + unBoundPeriod + 1
		},
	}, &mock.CryptoHookStub{})
	scAddress := []byte("owner")
	eei.SetSCAddress(scAddress)
	eei.SetStorage([]byte(ownerKey), scAddress)

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, unBoundPeriod, eei)

	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("address")
	arguments.Function = "unBound"

	stakingSmartContract.eei.SetStorage(arguments.CallerAddr, marshalizedRegData)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	data := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	assert.Equal(t, 0, len(data))

	destinationBalance := stakingSmartContract.eei.GetBalance(arguments.CallerAddr)
	scBalance := stakingSmartContract.eei.GetBalance(scAddress)
	assert.Equal(t, 0, destinationBalance.Cmp(stakeValue))
	assert.Equal(t, 0, scBalance.Cmp(big.NewInt(0).Mul(stakeValue, big.NewInt(-1))))
}

func TestStakingSC_ExecuteSlashOwnerAddrNotOkShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
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
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"
	arguments.CallerAddr = []byte("data")
	arguments.Arguments = []*big.Int{big.NewInt(100), big.NewInt(100)}

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
			registrationDataMarshalized, _ := json.Marshal(&StakingData{StakeValue: big.NewInt(100)})
			return registrationDataMarshalized
		}
	}

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"
	arguments.CallerAddr = []byte("data")
	arguments.Arguments = []*big.Int{big.NewInt(100), big.NewInt(100)}

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
			registrationDataMarshalized, _ := json.Marshal(&StakingData{StakeValue: big.NewInt(100), Staked: true})
			return registrationDataMarshalized
		}
	}

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"
	arguments.CallerAddr = []byte("data")
	arguments.Arguments = []*big.Int{big.NewInt(100), big.NewInt(100)}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestStakingSC_ExecuteUnStakeAndUnBoundStake(t *testing.T) {
	t.Parallel()

	// Preparation
	unBoundPeriod := uint64(100)
	stakeValue := big.NewInt(100)
	valueStakedByTheCaller := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	eei, _ := NewVMContext(blockChainHook, &mock.CryptoHookStub{})

	smartcontractAddress := "smartcontractAddress"
	eei.SetSCAddress([]byte(smartcontractAddress))

	ownerAddress := "ownerAddress"
	eei.SetStorage([]byte(ownerKey), []byte(ownerAddress))

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, unBoundPeriod, eei)

	arguments := CreateVmContractCallInput()

	stakedRegistrationData := StakingData{
		StartNonce:    0,
		Staked:        true,
		UnStakedNonce: 0,
		BlsPubKey:     nil,
		StakeValue:    valueStakedByTheCaller,
	}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.CallerAddr, marshalizedExpectedRegData)

	arguments.Function = "unStake"

	unStakeNonce := uint64(10)
	blockChainHook.CurrentNonceCalled = func() uint64 {
		return unStakeNonce
	}
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData StakingData
	data := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)

	expectedRegistrationData := StakingData{
		StartNonce:    0,
		Staked:        false,
		UnStakedNonce: unStakeNonce,
		BlsPubKey:     nil,
		StakeValue:    valueStakedByTheCaller,
	}
	assert.Equal(t, expectedRegistrationData, registrationData)

	arguments.Function = "unBound"

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return unStakeNonce + unBoundPeriod + 1
	}
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	destinationBalance := stakingSmartContract.eei.GetBalance(arguments.CallerAddr)
	senderBalance := stakingSmartContract.eei.GetBalance([]byte(ownerAddress))
	assert.Equal(t, big.NewInt(100), destinationBalance)
	assert.Equal(t, big.NewInt(-100), senderBalance)
}

func TestStakingSC_ExecuteGetShouldReturnUserErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	arguments := CreateVmContractCallInput()
	arguments.Function = "get"
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, &mock.CryptoHookStub{})
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
	err := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.UserError, err)
}

func TestStakingSC_ExecuteGetShouldOk(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	arguments := CreateVmContractCallInput()
	arguments.Function = "get"
	callerAddress := big.NewInt(0).SetBytes(arguments.CallerAddr)
	arguments.Arguments = []*big.Int{callerAddress}
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, &mock.CryptoHookStub{})
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
	err := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.Ok, err)
}

func TestStakingSc_ExecuteSlashTwoTime(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, &mock.CryptoHookStub{})

	stakedRegistrationData := StakingData{
		StartNonce:    50,
		Staked:        true,
		UnStakedNonce: 0,
		BlsPubKey:     nil,
		StakeValue:    stakeValue,
	}

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, 0, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"
	arguments.CallerAddr = []byte("data")
	marshalizedStakedData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.CallerAddr, marshalizedStakedData)
	stakingSmartContract.eei.SetStorage([]byte(ownerKey), arguments.CallerAddr)

	callerAddress := big.NewInt(0).SetBytes(arguments.CallerAddr)
	slashValue := big.NewInt(70)
	arguments.Arguments = []*big.Int{callerAddress, slashValue}
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	dataBytes := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	var registrationData StakingData
	err := json.Unmarshal(dataBytes, &registrationData)
	assert.Nil(t, err)

	expectedStake := big.NewInt(0).Sub(stakeValue, slashValue)
	assert.Equal(t, expectedStake, registrationData.StakeValue)

	arguments.Arguments = []*big.Int{callerAddress, slashValue}
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	dataBytes = stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	err = json.Unmarshal(dataBytes, &registrationData)
	assert.Nil(t, err)

	expectedStake = big.NewInt(0).Sub(expectedStake, slashValue)
	assert.Equal(t, expectedStake, registrationData.StakeValue)
}
