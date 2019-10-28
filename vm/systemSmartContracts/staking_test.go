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
	stakingSmartContract, err := NewStakingSmartContract(nil, eei)

	assert.Nil(t, stakingSmartContract)
	assert.Equal(t, vm.ErrNilInitialStakeValue, err)
}

func TestNewStakingSmartContract_NilSystemEIShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	stakingSmartContract, err := NewStakingSmartContract(stakeValue, nil)

	assert.Nil(t, stakingSmartContract)
	assert.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
}

func TestNewStakingSmartContract_NegativeStakeValueShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(-100)
	eei := &mock.SystemEIStub{}
	stakingSmartContract, err := NewStakingSmartContract(stakeValue, eei)

	assert.Nil(t, stakingSmartContract)
	assert.Equal(t, vm.ErrNegativeInitialStakeValue, err)
}

func TestNewStakingSmartContract(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei := &mock.SystemEIStub{}
	stakingSmartContract, err := NewStakingSmartContract(stakeValue, eei)

	assert.NotNil(t, stakingSmartContract)
	assert.Nil(t, err)
}

func TestStakingSC_ExecuteInit(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, &mock.CryptoHookStub{})
	eei.SetSCAddress([]byte("addr"))
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "_init"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	data := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	dataCallerAddr := stakingSmartContract.eei.GetStorage([]byte(ownerKey))
	assert.Equal(t, 0, len(data))
	assert.Equal(t, arguments.CallerAddr, dataCallerAddr)
}

func TestStakingSC_ExecuteStakeWrongStakeValueShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(100)
	blockChainHook := &mock.BlockChainHookStub{}
	eei, _ := NewVMContext(blockChainHook, &mock.CryptoHookStub{})
	eei.SetSCAddress([]byte("addr"))
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)

	balance := eei.GetBalance(arguments.CallerAddr)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestStakingSC_ExecuteStakeWrongUnmarshalDataShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(0)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteStakeRegistrationDataStakedShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(0)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		registrationDataMarshalized, _ := json.Marshal(&stakingData{Staked: true})
		return registrationDataMarshalized
	}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteStakeNotEnoughArgsShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(0)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		registrationDataMarshalized, _ := json.Marshal(&stakingData{})
		return registrationDataMarshalized
	}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteStake(t *testing.T) {
	t.Parallel()

	expectedRegistrationData := stakingData{
		StartNonce:    100,
		Staked:        true,
		UnStakedNonce: 0,
		BlsPubKey:     []byte{100},
		StakeValue:    big.NewInt(0),
	}

	stakeValue := big.NewInt(100)
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

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.Arguments = []*big.Int{big.NewInt(100), big.NewInt(200)}
	arguments.Header = &vmcommon.SCCallHeader{Number: big.NewInt(100)}
	arguments.CallValue = big.NewInt(100)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData stakingData
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

	stakeValue := big.NewInt(0)
	eei := &mock.SystemEIStub{}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteUnStakeUnmarshalErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(0)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteUnStakeAlreadyUnStakedAddrShouldErr(t *testing.T) {
	t.Parallel()

	stakedRegistrationData := stakingData{
		StartNonce:    0,
		Staked:        false,
		UnStakedNonce: 0,
		BlsPubKey:     nil,
		StakeValue:    nil,
	}

	stakeValue := big.NewInt(0)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, &mock.CryptoHookStub{})
	eei.SetSCAddress([]byte("addr"))

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
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

	expectedRegistrationData := stakingData{
		StartNonce:    0,
		Staked:        false,
		UnStakedNonce: 100,
		BlsPubKey:     nil,
		StakeValue:    nil,
	}

	stakedRegistrationData := stakingData{
		StartNonce:    0,
		Staked:        true,
		UnStakedNonce: 0,
		BlsPubKey:     nil,
		StakeValue:    nil,
	}

	stakeValue := big.NewInt(0)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, &mock.CryptoHookStub{})
	eei.SetSCAddress([]byte("addr"))

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.Arguments = []*big.Int{big.NewInt(100), big.NewInt(200)}
	arguments.Header = &vmcommon.SCCallHeader{Number: big.NewInt(100)}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.CallerAddr, marshalizedExpectedRegData)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData stakingData
	data := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)
	assert.Equal(t, expectedRegistrationData, registrationData)
}

func TestStakingSC_ExecuteFinalizeUnStakeOwnerAddrNotOkShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(0)
	eei := &mock.SystemEIStub{}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "finalizeUnStake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteFinalizeUnStakeUnmarshalErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(0)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "finalizeUnStake"
	arguments.Arguments = []*big.Int{big.NewInt(100), big.NewInt(200)}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteFinalizeUnStakeValidatorNotUnStakeShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(0)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		switch {
		case bytes.Equal(key, []byte(ownerKey)):
			return []byte("data")
		default:
			registrationDataMarshalized, _ := json.Marshal(&stakingData{UnStakedNonce: 0})
			return registrationDataMarshalized
		}
	}

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "finalizeUnStake"
	arguments.Arguments = []*big.Int{big.NewInt(100)}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteFinalizeUnStake(t *testing.T) {
	t.Parallel()

	registrationData := stakingData{
		StartNonce:    0,
		Staked:        true,
		UnStakedNonce: 10,
		BlsPubKey:     nil,
		StakeValue:    big.NewInt(100),
	}
	blsPubKey := big.NewInt(100)
	stakeValue := big.NewInt(0)
	marshalizedRegData, _ := json.Marshal(&registrationData)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, &mock.CryptoHookStub{})
	eei.SetSCAddress([]byte("addr"))
	eei.SetStorage([]byte(ownerKey), []byte("data"))
	eei.SetStorage(blsPubKey.Bytes(), marshalizedRegData)
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "finalizeUnStake"
	arguments.Arguments = []*big.Int{blsPubKey}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	data := stakingSmartContract.eei.GetStorage(blsPubKey.Bytes())
	assert.Equal(t, 0, len(data))

	destinationBalance := stakingSmartContract.eei.GetBalance(arguments.CallerAddr)
	senderBalance := stakingSmartContract.eei.GetBalance(blsPubKey.Bytes())
	assert.Equal(t, big.NewInt(100), destinationBalance)
	assert.Equal(t, big.NewInt(-100), senderBalance)
}

func TestStakingSC_ExecuteSlashOwnerAddrNotOkShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(0)
	eei := &mock.SystemEIStub{}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteSlashArgumentsNotOkShouldErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(0)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"
	arguments.CallerAddr = []byte("data")

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteSlashUnmarhsalErr(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(0)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"
	arguments.CallerAddr = []byte("data")
	arguments.Arguments = []*big.Int{big.NewInt(100), big.NewInt(100)}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingSC_ExecuteSlashNotStake(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(0)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		switch {
		case bytes.Equal(key, []byte(ownerKey)):
			return []byte("data")
		default:
			registrationDataMarshalized, _ := json.Marshal(&stakingData{StakeValue: big.NewInt(100)})
			return registrationDataMarshalized
		}
	}

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"
	arguments.CallerAddr = []byte("data")
	arguments.Arguments = []*big.Int{big.NewInt(100), big.NewInt(100)}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestStakingSC_ExecuteSlashStaked(t *testing.T) {
	t.Parallel()

	stakeValue := big.NewInt(0)
	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		switch {
		case bytes.Equal(key, []byte(ownerKey)):
			return []byte("data")
		default:
			registrationDataMarshalized, _ := json.Marshal(&stakingData{StakeValue: big.NewInt(100), Staked: true})
			return registrationDataMarshalized
		}
	}

	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"
	arguments.CallerAddr = []byte("data")
	arguments.Arguments = []*big.Int{big.NewInt(100), big.NewInt(100)}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestStakingSC_ExecuteUnStakeAndFinalizeUnStake(t *testing.T) {
	t.Parallel()

	expectedRegistrationData := stakingData{
		StartNonce:    0,
		Staked:        false,
		UnStakedNonce: 100,
		BlsPubKey:     nil,
		StakeValue:    big.NewInt(100),
	}

	stakedRegistrationData := stakingData{
		StartNonce:    0,
		Staked:        true,
		UnStakedNonce: 0,
		BlsPubKey:     nil,
		StakeValue:    big.NewInt(100),
	}

	stakeValue := big.NewInt(0)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, &mock.CryptoHookStub{})
	eei.SetSCAddress([]byte("addr"))
	eei.SetStorage([]byte(ownerKey), []byte("addr2"))
	stakingSmartContract, _ := NewStakingSmartContract(stakeValue, eei)
	arguments := CreateVmContractCallInput()
	addr1 := big.NewInt(0).SetBytes(arguments.CallerAddr)
	arguments.Function = "unStake"
	arguments.Arguments = []*big.Int{addr1}
	arguments.Header = &vmcommon.SCCallHeader{Number: big.NewInt(100)}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.CallerAddr, marshalizedExpectedRegData)
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData stakingData
	data := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)
	assert.Equal(t, expectedRegistrationData, registrationData)

	arguments.CallerAddr = []byte("addr2")
	arguments.Function = "finalizeUnStake"
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
	destinationBalance := stakingSmartContract.eei.GetBalance(arguments.CallerAddr)
	senderBalance := stakingSmartContract.eei.GetBalance(addr1.Bytes())
	assert.Equal(t, big.NewInt(100), destinationBalance)
	assert.Equal(t, big.NewInt(-100), senderBalance)
}
