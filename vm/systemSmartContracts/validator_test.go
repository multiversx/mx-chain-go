package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/parsers"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgumentsForValidatorSC() ArgsValidatorSmartContract {
	args := ArgsValidatorSmartContract{
		Eei:                &mock.SystemEIStub{},
		SigVerifier:        &mock.MessageSignVerifierMock{},
		ValidatorSCAddress: []byte("validator"),
		StakingSCAddress:   []byte("staking"),
		EndOfEpochAddress:  []byte("endOfEpoch"),
		StakingSCConfig: config.StakingSystemSCConfig{
			GenesisNodePrice: "1000",
			UnJailValue:      "10",
			MinStepValue:     "10",
			MinStakeValue:    "1",
			UnBondPeriod:     1,
			NumRoundsWithoutBleed:                1,
			MaximumPercentageToBleed:             1,
			BleedPercentagePerRound:              1,
			MaxNumberOfNodesForStake:             10,
			ActivateBLSPubKeyMessageVerification: false,
			MinUnstakeTokensValue:                "1",
		},
		Marshalizer:        &mock.MarshalizerMock{},
		GenesisTotalSupply: big.NewInt(100000000),
		EpochNotifier:      &mock.EpochNotifierStub{},
		MinDeposit:         "0",
		EpochConfig: config.EpochConfig{
			EnableEpochs: config.EnableEpochs{
				StakeEnableEpoch:                     0,
				StakingV2Epoch:                       10,
			},
		},
	}

	return args
}

func createABid(totalStakeValue uint64, numBlsKeys uint32, maxStakePerNode uint64) ValidatorDataV2 {
	data := ValidatorDataV2{
		RewardAddress:   []byte("addr"),
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      nil,
		TotalStakeValue: big.NewInt(0).SetUint64(totalStakeValue),
		LockedStake:     big.NewInt(0).SetUint64(totalStakeValue),
		MaxStakePerNode: big.NewInt(0).SetUint64(maxStakePerNode),
		NumRegistered:   numBlsKeys,
	}

	keys := make([][]byte, 0)
	for i := uint32(0); i < numBlsKeys; i++ {
		keys = append(keys, []byte(fmt.Sprintf("%d", rand.Uint32())))
	}
	data.BlsPubKeys = keys

	return data
}

func TestNewStakingValidatorSmartContract_InvalidUnJailValue(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSC()

	arguments.StakingSCConfig.UnJailValue = ""
	asc, err := NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidUnJailCost))

	arguments.StakingSCConfig.UnJailValue = "0"
	asc, err = NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidUnJailCost))

	arguments.StakingSCConfig.UnJailValue = "-5"
	asc, err = NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidUnJailCost))
}

func TestNewStakingValidatorSmartContract_InvalidMinStakeValue(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSC()

	arguments.StakingSCConfig.MinStakeValue = ""
	asc, err := NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidMinStakeValue))

	arguments.StakingSCConfig.MinStakeValue = "0"
	asc, err = NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidMinStakeValue))

	arguments.StakingSCConfig.MinStakeValue = "-5"
	asc, err = NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidMinStakeValue))
}

func TestNewStakingValidatorSmartContract_InvalidGenesisNodePrice(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSC()

	arguments.StakingSCConfig.GenesisNodePrice = ""
	asc, err := NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidNodePrice))

	arguments.StakingSCConfig.GenesisNodePrice = "0"
	asc, err = NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidNodePrice))

	arguments.StakingSCConfig.GenesisNodePrice = "-5"
	asc, err = NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidNodePrice))
}

func TestNewStakingValidatorSmartContract_InvalidMinStepValue(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSC()

	arguments.StakingSCConfig.MinStepValue = ""
	asc, err := NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidMinStepValue))

	arguments.StakingSCConfig.MinStepValue = "0"
	asc, err = NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidMinStepValue))

	arguments.StakingSCConfig.MinStepValue = "-5"
	asc, err = NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidMinStepValue))
}

func TestNewStakingValidatorSmartContract_NilSystemEnvironmentInterface(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSC()
	arguments.Eei = nil

	asc, err := NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
}

func TestNewStakingValidatorSmartContract_NilStakingSmartContractAddress(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSC()
	arguments.StakingSCAddress = nil

	asc, err := NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.Equal(t, vm.ErrNilStakingSmartContractAddress, err)
}

func TestNewStakingValidatorSmartContract_NilValidatorSmartContractAddress(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSC()
	arguments.ValidatorSCAddress = nil

	asc, err := NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.Equal(t, vm.ErrNilValidatorSmartContractAddress, err)
}

func TestNewStakingValidatorSmartContract_NilSigVerifier(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSC()
	arguments.SigVerifier = nil

	asc, err := NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.Equal(t, vm.ErrNilMessageSignVerifier, err)
}

func TestNewStakingValidatorSmartContract_NilMarshalizer(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSC()
	arguments.Marshalizer = nil

	asc, err := NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.Equal(t, vm.ErrNilMarshalizer, err)
}

func TestNewStakingValidatorSmartContract_InvalidGenesisTotalSupply(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSC()
	arguments.GenesisTotalSupply = nil

	asc, err := NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidGenesisTotalSupply))

	arguments.GenesisTotalSupply = big.NewInt(0)

	asc, err = NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidGenesisTotalSupply))

	arguments.GenesisTotalSupply = big.NewInt(-2)

	asc, err = NewValidatorSmartContract(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidGenesisTotalSupply))
}

func TestStakingValidatorSC_ExecuteStakeWithoutArgumentsShouldWork(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInput()
	validatorData := createABid(25000000, 2, 12500000)
	validatorDataBytes, _ := json.Marshal(&validatorData)

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		if bytes.Equal(key, arguments.CallerAddr) {
			return validatorDataBytes
		}
		return nil
	}
	eei.SetStorageCalled = func(key []byte, value []byte) {
		if bytes.Equal(key, arguments.CallerAddr) {
			var validatorDataRecovered ValidatorDataV2
			_ = json.Unmarshal(value, &validatorDataRecovered)
			assert.Equal(t, big.NewInt(26000000), validatorDataRecovered.TotalStakeValue)
		}
	}
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingValidatorSc, _ := NewValidatorSmartContract(args)

	arguments.Function = "stake"
	arguments.CallValue = big.NewInt(1000000)

	errCode := stakingValidatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}

func TestStakingValidatorSC_ExecuteStakeAddedNewPubKeysShouldWork(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInput()
	validatorData := createABid(25000000, 2, 12500000)
	validatorDataBytes, _ := json.Marshal(&validatorData)

	key1 := []byte("Key1")
	key2 := []byte("Key2")
	rewardAddr := []byte("tralala2")
	maxStakePerNonce := big.NewInt(500)

	args := createMockArgumentsForValidatorSC()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("validator"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei
	eei.SetStorage(arguments.CallerAddr, validatorDataBytes)
	args.StakingSCConfig = argsStaking.StakingSCConfig

	stakingValidatorSc, _ := NewValidatorSmartContract(args)

	arguments.Function = "stake"
	arguments.CallValue = big.NewInt(0).Mul(big.NewInt(2), big.NewInt(10000000))
	arguments.Arguments = [][]byte{big.NewInt(2).Bytes(), key1, []byte("msg1"), key2, []byte("msg2"), maxStakePerNonce.Bytes(), rewardAddr}

	errCode := stakingValidatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}

func TestStakingAuctionSC_ExecuteStakeDoubleKeyAndCleanup(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInput()

	key1 := []byte("Key1")
	args := createMockArgumentsForValidatorSC()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("auction"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei
	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.EpochConfig.EnableEpochs.DoubleKeyProtectionEnableEpoch = 1000
	validatorSc, _ := NewValidatorSmartContract(args)

	arguments.Function = "stake"
	arguments.CallValue = big.NewInt(0).Mul(big.NewInt(2), big.NewInt(10000000))
	arguments.Arguments = [][]byte{big.NewInt(2).Bytes(), key1, []byte("msg1"), key1, []byte("msg1")}

	errCode := validatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)

	registeredData := &ValidatorDataV2{}
	_ = validatorSc.marshalizer.Unmarshal(registeredData, eei.GetStorage(arguments.CallerAddr))
	assert.Equal(t, 2, len(registeredData.BlsPubKeys))

	validatorSc.flagDoubleKey.Set()
	arguments.Function = "cleanRegisteredData"
	arguments.CallValue = big.NewInt(0)
	arguments.Arguments = [][]byte{}

	errCode = validatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)

	_ = validatorSc.marshalizer.Unmarshal(registeredData, eei.GetStorage(arguments.CallerAddr))
	assert.Equal(t, 1, len(registeredData.BlsPubKeys))
}

func TestStakingValidatorSC_ExecuteStakeWithRewardAddress(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("staker1")
	stakerPubKey := []byte("stakerBLSPubKey")

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei
	args.StakingSCConfig = argsStaking.StakingSCConfig

	rwdAddress := []byte("reward111")
	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey, []byte("signed"), rwdAddress}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey)
	stakedData := &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, bytes.Equal(stakedData.RewardAddress, rwdAddress))
}

func TestStakingValidatorSC_ExecuteStakeUnJail(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("address")
	stakerPubKey := []byte("blsPubKey")

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	argsStaking.StakingSCConfig.UnJailValue = "1000"

	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei
	args.StakingSCConfig = argsStaking.StakingSCConfig
	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unJail"
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallValue = big.NewInt(1000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.CallerAddr = []byte("otherAddress")
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingValidatorSC_ExecuteStakeUnStakeOneBlsPubKey(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInput()
	validatorData := createABid(25000000, 2, 12500000)
	validatorDataBytes, _ := json.Marshal(&validatorData)

	stakedData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 1,
		UnStakedEpoch: core.DefaultUnstakedEpoch,
		RewardAddress: []byte("tralala1"),
		StakeValue:    big.NewInt(0),
	}
	stakedDataBytes, _ := json.Marshal(&stakedData)

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		if bytes.Equal(key, arguments.CallerAddr) {
			return validatorDataBytes
		}
		if bytes.Equal(key, validatorData.BlsPubKeys[0]) {
			return stakedDataBytes
		}
		return nil
	}
	eei.SetStorageCalled = func(key []byte, value []byte) {
		var stakedDataRecovered StakedDataV2_0
		_ = json.Unmarshal(value, &stakedDataRecovered)

		assert.Equal(t, false, stakedDataRecovered.Staked)
	}

	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingValidatorSc, _ := NewValidatorSmartContract(args)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{validatorData.BlsPubKeys[0]}
	errCode := stakingValidatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}

func TestStakingValidatorSC_ExecuteStakeStakeClaim(t *testing.T) {
	t.Parallel()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress.Bytes()
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "claim"
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	vmOutput := eei.CreateVMOutput()
	assert.NotNil(t, vmOutput)
	outputAccount := vmOutput.OutputAccounts[string(arguments.CallerAddr)]
	assert.True(t, outputAccount.BalanceDelta.Cmp(big.NewInt(10000000)) == 0)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey.Bytes())
	stakedData := &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)
}

func TestStakingValidatorSC_ExecuteStakeStakeTokensUnBondRestakeUnStake(t *testing.T) {
	t.Parallel()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 1
	argsStaking.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	argsStaking.MinNumNodes = 0
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress.Bytes()
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{}
	arguments.CallValue = big.NewInt(100)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStakeTokens"
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes()}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 1
	}

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey.Bytes()}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 100
	}
	blockChainHook.CurrentEpochCalled = func() uint32 {
		return 10
	}

	arguments.Function = "unBondTokens"
	arguments.Arguments = [][]byte{}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress.Bytes()
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	marshaledData := eei.GetStorageFromAddress(args.StakingSCAddress, stakerPubKey.Bytes())
	stakedData := &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{}
	arguments.CallValue = big.NewInt(100)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStakeTokens"
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes()}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey.Bytes()}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestStakingValidatorSC_ExecuteStakeUnStakeStakeClaim(t *testing.T) {
	t.Parallel()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)

	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return 100
		},
		CurrentEpochCalled: func() uint32 {
			return 10
		},
	}
	args := createMockArgumentsForValidatorSC()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	argsStaking.MinNumNodes = 0
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress.Bytes()
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey.Bytes()}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "claim"
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	vmOutput := eei.CreateVMOutput()
	assert.NotNil(t, vmOutput)
	outputAccount := vmOutput.OutputAccounts[string(arguments.CallerAddr)]
	assert.True(t, outputAccount.BalanceDelta.Cmp(big.NewInt(10000000)) == 0)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey.Bytes())
	stakedData := &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)
}

func TestStakingValidatorSC_ExecuteStakeUnStakeOneBlsPubKeyAndRestake(t *testing.T) {
	t.Parallel()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei
	args.StakingSCConfig = argsStaking.StakingSCConfig
	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress.Bytes()
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey.Bytes()}
	arguments.CallValue = big.NewInt(0)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey.Bytes()}
	arguments.CallValue = big.NewInt(0)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	addToStakeValue := big.NewInt(10000)
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = addToStakeValue
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "claim"
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	vmOutput := eei.CreateVMOutput()
	assert.NotNil(t, vmOutput)
	outputAccount := vmOutput.OutputAccounts[string(arguments.CallerAddr)]
	assert.True(t, outputAccount.BalanceDelta.Cmp(addToStakeValue) == 0)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey.Bytes())
	stakedData := &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)
}

func TestStakingValidatorSC_ExecuteStakeUnStake1Stake1More(t *testing.T) {
	t.Parallel()

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.MinNumNodes = 0
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	argsStaking.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 10
	}

	staker := []byte("staker")
	args.Eei = eei
	args.StakingSCConfig = argsStaking.StakingSCConfig
	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = staker
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), []byte("blsKey1"), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{[]byte("blsKey1")}
	arguments.CallValue = big.NewInt(0)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), []byte("blsKey2"), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{[]byte("blsKey2")}
	arguments.CallValue = big.NewInt(0)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(2).Bytes(), []byte("blsKey3"), []byte("blsKey4"), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), []byte("blsKey3"), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), []byte("blsKey4"), []byte("signed")}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)

	marshaledData := eei.GetStorageFromAddress(args.StakingSCAddress, []byte("blsKey3"))
	stakedData := &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)

	marshaledData = eei.GetStorageFromAddress(args.StakingSCAddress, []byte("blsKey2"))
	stakedData = &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.False(t, stakedData.Staked)

	marshaledData = eei.GetStorageFromAddress(args.StakingSCAddress, []byte("blsKey1"))
	stakedData = &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.False(t, stakedData.Staked)
}

func TestStakingValidatorSC_ExecuteStakeUnStakeUnBondUnStakeUnBondOneBlsPubKey(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("staker1")
	stakerPubKey := []byte("bls1")

	unBondPeriod := uint64(5)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = unBondPeriod
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.CallerAddr = []byte("anotherAddress")
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), []byte("anotherKey"), []byte("signed")}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 100
	}
	blockChainHook.CurrentEpochCalled = func() uint32 {
		return 10
	}

	arguments.CallerAddr = stakerAddress
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 103
	}
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 120
	}
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 220
	}
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 320
	}
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey)
	assert.Equal(t, 0, len(marshaledData))
}

func TestStakingValidatorSC_StakeUnStake3XUnBond2xWaitingList(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("address")
	stakerPubKey1 := []byte("blsKey1")
	stakerPubKey2 := []byte("blsKey2")
	stakerPubKey3 := []byte("blsKey3")

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	args.StakingSCConfig.MaxNumberOfNodesForStake = 1
	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey1, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey1}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey2, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey2}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey3, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey3}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey2}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey3}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestStakingValidatorSC_StakeUnStake3XRestake2(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("address")
	stakerPubKey1 := []byte("blsKey1")
	stakerPubKey2 := []byte("blsKey2")
	stakerPubKey3 := []byte("blsKey3")

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	args.StakingSCConfig.MaxNumberOfNodesForStake = 1
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	argsStaking.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey1, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey2, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey3, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 10
	}

	arguments.Function = "unStakeNodes"
	arguments.Arguments = [][]byte{stakerPubKey1}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStakeNodes"
	arguments.Arguments = [][]byte{stakerPubKey2}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "reStakeUnStakedNodes"
	arguments.Arguments = [][]byte{stakerPubKey1, stakerPubKey2}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestStakingValidatorSC_StakeShouldSetOwnerIfStakingV2IsEnabled(t *testing.T) {
	t.Parallel()

	ownerAddress := []byte("owner")
	blsKey := []byte("blsKey")

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	args.StakingSCConfig.MaxNumberOfNodesForStake = 1
	atArgParser := parsers.NewCallArgsParser()

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})
	argsStaking.Eei = eei
	eei.SetSCAddress(args.ValidatorSCAddress)
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	argsStaking.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = ownerAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), blsKey, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	registeredData, err := sc.getStakedData(blsKey)
	require.Nil(t, err)
	assert.Equal(t, ownerAddress, registeredData.OwnerAddress)
}

func TestStakingValidatorSC_ExecuteStakeChangeRewardAddresStakeUnStake(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("staker111")
	stakerPubKey := []byte("bls1")

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 1000
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	rewardAddress := []byte("staker222")
	arguments.Function = "changeRewardAddress"
	arguments.Arguments = [][]byte{rewardAddress}
	arguments.CallValue = big.NewInt(0)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blsKey2 := []byte("blsKey2")
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), blsKey2, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{blsKey2}
	arguments.CallValue = big.NewInt(0)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey)
	stakedData := &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)
	assert.True(t, bytes.Equal(rewardAddress, stakedData.RewardAddress))

	marshaledData = eei.GetStorage(blsKey2)
	stakedData = &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.False(t, stakedData.Staked)
	assert.True(t, bytes.Equal(rewardAddress, stakedData.RewardAddress))
}

func TestStakingValidatorSC_ExecuteStakeUnStakeUnBondBlsPubKeyAndReStake(t *testing.T) {
	t.Parallel()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)
	nonce := uint64(1)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return nonce
		},
	}
	args := createMockArgumentsForValidatorSC()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 1000
	argsStaking.EpochConfig.EnableEpochs.StakingV2Epoch = 100000000
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress.Bytes()
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.CallerAddr = []byte("anotherCaller")
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), []byte("anotherKey"), []byte("signed")}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	nonce += 1
	arguments.CallerAddr = stakerAddress.Bytes()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey.Bytes()}
	arguments.CallValue = big.NewInt(0)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	nonce += args.StakingSCConfig.UnBondPeriod + 1
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey.Bytes()}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)

	arguments.CallValue = big.NewInt(10000000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey.Bytes())
	stakedData := &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)
}

func TestStakingValidatorSC_ExecuteUnBound(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInput()
	totalStake := uint64(25000000)

	validatorData := createABid(totalStake, 2, 12500000)
	validatorDataBytes, _ := json.Marshal(&validatorData)

	stakedData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 1,
		UnStakedEpoch: core.DefaultUnstakedEpoch,
		RewardAddress: []byte("tralala1"),
		StakeValue:    big.NewInt(12500000),
	}
	stakedDataBytes, _ := json.Marshal(&stakedData)

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		if bytes.Equal(arguments.CallerAddr, key) {
			return validatorDataBytes
		}
		if bytes.Equal(key, validatorData.BlsPubKeys[0]) {
			return stakedDataBytes
		}

		return nil
	}

	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingValidatorSc, _ := NewValidatorSmartContract(args)

	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{validatorData.BlsPubKeys[0]}
	errCode := stakingValidatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}

func TestValidatorStakingSC_ExecuteInit(t *testing.T) {
	t.Parallel()

	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContract(args)
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

func TestValidatorStakingSC_ExecuteInitTwoTimeShouldReturnUserError(t *testing.T) {
	t.Parallel()

	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = core.SCDeployInitFunctionName

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestValidatorStakingSC_ExecuteStakeOutOfGasShouldErr(t *testing.T) {
	t.Parallel()

	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return 2
		},
	}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockArgumentsForValidatorSC()
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.Stake = 10
	stakingSmartContract, err := NewValidatorSmartContract(args)
	require.Nil(t, err)
	arguments := CreateVmContractCallInput()
	arguments.GasProvided = 5
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.OutOfGas, retCode)

	assert.Equal(t, vm.InsufficientGasLimit, eei.returnMessage)
}

func TestValidatorStakingSC_ExecuteStakeWrongStakeValueShouldErr(t *testing.T) {
	t.Parallel()

	blockChainHook := &mock.BlockChainHookStub{
		GetUserAccountCalled: func(address []byte) (vmcommon.UserAccountHandler, error) {
			return nil, state.ErrAccNotFound
		},
	}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockArgumentsForValidatorSC()
	args.Eei = eei
	args.StakingSCConfig.GenesisNodePrice = "10"
	stakingSmartContract, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallValue = big.NewInt(2)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, fmt.Sprintf("insufficient stake value: expected %s, got %v",
		args.StakingSCConfig.GenesisNodePrice, arguments.CallValue))

	balance := eei.GetBalance(arguments.CallerAddr)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestValidatorStakingSC_ExecuteStakeWrongUnmarshalDataShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestValidatorStakingSC_ExecuteStakeAlreadyStakedShouldNotErr(t *testing.T) {
	t.Parallel()

	stakerBlsKey1 := big.NewInt(101)
	expectedCallerAddress := []byte("caller")
	stakedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: expectedCallerAddress,
		StakeValue:    nil,
	}

	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)

	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallValue = nodePrice
	arguments.CallerAddr = expectedCallerAddress
	arguments.Arguments = [][]byte{
		big.NewInt(1).Bytes(),
		stakerBlsKey1.Bytes(), []byte("signed"),
	}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(stakerBlsKey1.Bytes(), marshalizedExpectedRegData)

	validatorData := ValidatorDataV2{
		RewardAddress:   arguments.CallerAddr,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{stakerBlsKey1.Bytes()},
		TotalStakeValue: nodePrice,
		LockedStake:     nodePrice,
		MaxStakePerNode: nodePrice,
		NumRegistered:   1,
	}
	marshaledRegistrationData, _ := json.Marshal(validatorData)

	eei.SetSCAddress(args.ValidatorSCAddress)
	eei.SetStorage(arguments.CallerAddr, marshaledRegistrationData)
	retCode := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.Ok, retCode)
	var registrationData ValidatorDataV2
	data := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	_ = json.Unmarshal(data, &registrationData)

	validatorData.TotalStakeValue = big.NewInt(0).Mul(nodePrice, big.NewInt(2))
	validatorData.MaxStakePerNode = big.NewInt(0).Mul(nodePrice, big.NewInt(2))

	assert.Equal(t, validatorData, registrationData)
}

func TestValidatorStakingSC_ExecuteStakeStakedInStakingButNotInValidatorShouldErr(t *testing.T) {
	t.Parallel()

	stakerBlsKey1 := big.NewInt(101)
	expectedCallerAddress := []byte("caller")
	stakedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: expectedCallerAddress,
		StakeValue:    nil,
	}

	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)

	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallValue = nodePrice
	arguments.CallerAddr = expectedCallerAddress
	maxStakePerNode := big.NewInt(0).Mul(nodePrice, big.NewInt(2))
	arguments.Arguments = [][]byte{
		big.NewInt(1).Bytes(),
		stakerBlsKey1.Bytes(),
		[]byte("signed"),
		maxStakePerNode.Bytes(),
	}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(stakerBlsKey1.Bytes(), marshalizedExpectedRegData)

	validatorData := ValidatorDataV2{
		RewardAddress:   arguments.CallerAddr,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      nil,
		TotalStakeValue: big.NewInt(0),
		LockedStake:     big.NewInt(0),
		MaxStakePerNode: big.NewInt(0),
		NumRegistered:   0,
	}
	marshaledRegistrationData, _ := json.Marshal(validatorData)

	eei.SetSCAddress(args.ValidatorSCAddress)
	eei.SetStorage(arguments.CallerAddr, marshaledRegistrationData)
	retCode := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.UserError, retCode)
	assert.True(t, strings.Contains(eei.returnMessage, "bls key already registered"))
	var registrationData ValidatorDataV2
	data := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	_ = json.Unmarshal(data, &registrationData)

	assert.Equal(t, validatorData, registrationData)
}

func TestValidatorStakingSC_ExecuteStakeWithMaxStakePerNode(t *testing.T) {
	t.Parallel()

	stakerBlsKey1 := big.NewInt(101)
	expectedCallerAddress := []byte("caller")

	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)

	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallValue = nodePrice
	arguments.CallerAddr = expectedCallerAddress
	maxStakePerNode := big.NewInt(0).Mul(nodePrice, big.NewInt(2))
	arguments.Arguments = [][]byte{
		big.NewInt(1).Bytes(),
		stakerBlsKey1.Bytes(),
		[]byte("signed"),
		maxStakePerNode.Bytes(),
	}

	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(stakerBlsKey1.Bytes(), nil)

	validatorData := ValidatorDataV2{
		RewardAddress:   arguments.CallerAddr,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      nil,
		TotalStakeValue: big.NewInt(0),
		LockedStake:     big.NewInt(0),
		MaxStakePerNode: big.NewInt(0),
		NumRegistered:   0,
	}
	marshaledRegistrationData, _ := json.Marshal(validatorData)

	eei.SetSCAddress(args.ValidatorSCAddress)
	eei.SetStorage(arguments.CallerAddr, marshaledRegistrationData)
	retCode := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.Ok, retCode)
	var registrationData ValidatorDataV2
	data := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	_ = json.Unmarshal(data, &registrationData)

	assert.Equal(t, maxStakePerNode, registrationData.MaxStakePerNode)
}

func TestValidatorStakingSC_ExecuteStakeNotEnoughArgsShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		registrationDataMarshalized, _ := json.Marshal(&StakedDataV2_0{})
		return registrationDataMarshalized
	}
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestValidatorStakingSC_ExecuteStakeNotEnoughFundsForMultipleNodesShouldErr(t *testing.T) {
	t.Parallel()
	stakerPubKey1 := big.NewInt(101)
	stakerPubKey2 := big.NewInt(102)
	args := createMockArgumentsForValidatorSC()

	blockChainHook := &mock.BlockChainHookStub{}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei
	args.StakingSCConfig.GenesisNodePrice = "10"
	stakingSmartContract, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	eei.SetGasProvided(arguments.GasProvided)
	arguments.CallValue = big.NewInt(15)
	arguments.Arguments = [][]byte{
		big.NewInt(2).Bytes(),
		stakerPubKey1.Bytes(), []byte("signed"),
		stakerPubKey2.Bytes(), []byte("signed"),
	}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.OutOfFunds, retCode)

	balance := eei.GetBalance(arguments.CallerAddr)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestValidatorStakingSC_ExecuteStakeNotEnoughGasForMultipleNodesShouldErr(t *testing.T) {
	t.Parallel()
	stakerPubKey1 := big.NewInt(101)
	stakerPubKey2 := big.NewInt(102)
	args := createMockArgumentsForValidatorSC()

	blockChainHook := &mock.BlockChainHookStub{
		GetUserAccountCalled: func(address []byte) (vmcommon.UserAccountHandler, error) {
			return nil, state.ErrAccNotFound
		},
	}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei
	args.StakingSCConfig.GenesisNodePrice = "10"
	args.GasCost.MetaChainSystemSCsCost.Stake = 10
	stakingSmartContract, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.GasProvided = 15
	eei.SetGasProvided(arguments.GasProvided)
	arguments.CallValue = big.NewInt(15)
	arguments.Arguments = [][]byte{
		big.NewInt(2).Bytes(),
		stakerPubKey1.Bytes(), []byte("signed"),
		stakerPubKey2.Bytes(), []byte("signed"),
	}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.OutOfGas, retCode)

	balance := eei.GetBalance(arguments.CallerAddr)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestValidatorStakingSC_ExecuteStakeOneKeyFailsOneRegisterStakeSCShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForValidatorSC()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)

	blockChainHook := &mock.BlockChainHookStub{}
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	args.Eei = eei
	executeSecond := true
	stakingSc := &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
		if args.Function != "register" {
			return vmcommon.Ok
		}

		if executeSecond {
			executeSecond = false
			return vmcommon.Ok
		}
		return vmcommon.UserError
	}}

	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress.Bytes()
	arguments.CallValue = big.NewInt(nodePrice.Int64() * 2)
	arguments.Arguments = [][]byte{
		big.NewInt(2).Bytes(),
		stakerPubKey.Bytes(), []byte("signed"),
		stakerPubKey.Bytes(), []byte("signed"),
	}

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	sc, _ := NewValidatorSmartContract(args)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot register bls key"))

	data := sc.eei.GetStorage(arguments.CallerAddr)
	assert.Nil(t, data)
}

func TestValidatorStakingSC_ExecuteStakeBeforeValidatorEnableNonce(t *testing.T) {
	t.Parallel()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)

	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 99
		},
	}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 100
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	expectedRegistrationData := ValidatorDataV2{
		RewardAddress:   stakerAddress.Bytes(),
		RegisterNonce:   0,
		Epoch:           99,
		BlsPubKeys:      [][]byte{stakerPubKey.Bytes()},
		TotalStakeValue: nodePrice,
		LockedStake:     nodePrice,
		MaxStakePerNode: nodePrice,
		NumRegistered:   1,
	}

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress.Bytes()
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(0).Set(nodePrice)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData ValidatorDataV2
	data := sc.eei.GetStorage(arguments.CallerAddr)
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)
	assert.Equal(t, expectedRegistrationData, registrationData)
}

func TestValidatorStakingSC_ExecuteStake(t *testing.T) {
	t.Parallel()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	expectedRegistrationData := ValidatorDataV2{
		RewardAddress:   stakerAddress.Bytes(),
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{stakerPubKey.Bytes()},
		TotalStakeValue: nodePrice,
		LockedStake:     nodePrice,
		MaxStakePerNode: nodePrice,
		NumRegistered:   1,
	}

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress.Bytes()
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(100).Set(nodePrice)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData ValidatorDataV2
	data := sc.eei.GetStorage(arguments.CallerAddr)
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)
	assert.Equal(t, expectedRegistrationData, registrationData)
}

func TestValidatorStakingSC_ExecuteUnStakeValueNotZeroShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.CallValue = big.NewInt(1)
	arguments.Function = "unStake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, vm.TransactionValueMustBeZero, eei.ReturnMessage)
}

func TestValidatorStakingSC_ExecuteUnStakeAddressNotStakedShouldErr(t *testing.T) {
	t.Parallel()

	notFoundkey := []byte("abc")
	eei := &mock.SystemEIStub{}
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{notFoundkey}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, vm.CannotGetAllBlsKeysFromRegistrationData+
		fmt.Errorf("%w, key %s not found", vm.ErrBLSPublicKeyMismatch, hex.EncodeToString(notFoundkey)).Error(), eei.ReturnMessage)
}

func TestValidatorStakingSC_ExecuteUnStakeUnmarshalErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei
	args.Marshalizer = &mock.MarshalizerMock{Fail: true}

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{[]byte("abc")}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, vm.CannotGetOrCreateRegistrationData+mock.ErrMockMarshalizer.Error(), eei.ReturnMessage)
}

func TestValidatorStakingSC_ExecuteUnStakeAlreadyUnStakedAddrShouldNotErr(t *testing.T) {
	t.Parallel()

	expectedCallerAddress := []byte("caller")
	stakedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		RewardAddress: expectedCallerAddress,
		StakeValue:    nil,
	}

	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	stakingSmartContract, _ := NewValidatorSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.CallerAddr = expectedCallerAddress
	arguments.Arguments = [][]byte{[]byte("abc")}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)

	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	validatorData := ValidatorDataV2{
		RewardAddress:   arguments.CallerAddr,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{arguments.Arguments[0]},
		TotalStakeValue: nodePrice,
		LockedStake:     nodePrice,
		MaxStakePerNode: nodePrice,
		NumRegistered:   1,
	}
	marshaledRegistrationData, _ := json.Marshal(validatorData)

	eei.SetSCAddress(args.ValidatorSCAddress)
	eei.SetStorage(arguments.CallerAddr, marshaledRegistrationData)
	retCode := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.Ok, retCode)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot unStake node which was already unStaked"))
}

func TestValidatorStakingSC_ExecuteUnStakeFailsWithWrongCaller(t *testing.T) {
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

	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{}, &mock.RaterMock{})
	eei.SetSCAddress([]byte("addr"))
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.CallerAddr = wrongCallerAddress
	arguments.Arguments = [][]byte{[]byte("abc")}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestValidatorStakingSC_ExecuteUnStake(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForValidatorSC()
	args.StakingSCConfig.UnBondPeriod = 10
	callerAddress := []byte("caller")
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	expectedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		RewardAddress: callerAddress,
		StakeValue:    nodePrice,
		JailedRound:   math.MaxUint64,
		SlashValue:    big.NewInt(0),
	}

	stakedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: callerAddress,
		StakeValue:    nil,
	}

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei
	eei.SetSCAddress(args.ValidatorSCAddress)

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{[]byte("abc")}
	arguments.CallerAddr = callerAddress
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)

	validatorData := ValidatorDataV2{
		RewardAddress:   arguments.CallerAddr,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{arguments.Arguments[0]},
		TotalStakeValue: nodePrice,
		LockedStake:     nodePrice,
		MaxStakePerNode: nodePrice,
		NumRegistered:   1,
	}
	marshaledRegistrationData, _ := json.Marshal(validatorData)
	eei.SetStorage(arguments.CallerAddr, marshaledRegistrationData)

	stakedData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		UnStakedEpoch: core.DefaultUnstakedEpoch,
		RewardAddress: arguments.CallerAddr,
		StakeValue:    nodePrice,
		JailedRound:   math.MaxUint64,
		SlashValue:    big.NewInt(0),
	}
	marshaledStakedData, _ := json.Marshal(stakedData)
	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(arguments.Arguments[0], marshaledStakedData)
	stakingSc.setConfig(&StakingNodesConfig{MinNumNodes: 5, StakedNodes: 10})
	eei.SetSCAddress(args.ValidatorSCAddress)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData StakedDataV2_0
	eei.SetSCAddress(args.StakingSCAddress)
	data := eei.GetStorage(arguments.Arguments[0])
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)
	assert.Equal(t, expectedRegistrationData, registrationData)
}

func TestValidatorStakingSC_ExecuteUnBoundUnmarshalErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes(), big.NewInt(200).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestValidatorStakingSC_ExecuteUnBoundValidatorNotUnStakeShouldErr(t *testing.T) {
	t.Parallel()

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
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestValidatorStakingSC_ExecuteStakeUnStakeReturnsErrAsNotEnabled(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.BlockChainHookCalled = func() vm.BlockchainHook {
		return &mock.BlockChainHookStub{
			CurrentEpochCalled: func() uint32 {
				return 100
			},
			CurrentNonceCalled: func() uint64 {
				return 100
			}}
	}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakeEnableEpoch = eei.BlockChainHook().CurrentEpoch() + uint32(1)
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, vm.UnBondNotEnabled, eei.ReturnMessage)

	arguments.Function = "unStake"
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, vm.UnStakeNotEnabled, eei.ReturnMessage)

	arguments.Function = "stake"
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, vm.StakeNotEnabled, eei.ReturnMessage)
}

func TestValidatorStakingSC_ExecuteUnBondBeforePeriodEnds(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(100)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return 10
		},
	}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 10000
	args.StakingSCConfig.UnBondPeriod = 1000
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	validatorSc, _ := NewValidatorSmartContract(args)
	eei.SetSCAddress([]byte("staking"))

	blsPubKey := []byte("pubkey")
	_ = validatorSc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RewardAddress: caller,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedNonce: 1,
					UnstakedValue: big.NewInt(1000),
				},
			},
			BlsPubKeys:      [][]byte{blsPubKey},
			TotalStakeValue: big.NewInt(1000), //in v1 this was still set to the unstaked value
			LockedStake:     big.NewInt(0),
			TotalUnstaked:   big.NewInt(1000),
		},
	)

	systemSc, _ := eei.GetContract(nil)
	stakingSc := systemSc.(*stakingSC)
	registrationData := &StakedDataV2_0{
		RegisterNonce: 0,
		UnStakedNonce: uint64(1),
		RewardAddress: caller,
		StakeValue:    big.NewInt(100),
	}
	_ = stakingSc.saveAsStakingDataV1P1(blsPubKey, registrationData)

	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = caller
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{blsPubKey}

	retCode := validatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
	assert.True(t, strings.Contains(eei.returnMessage, "unBond is not possible"))
	assert.True(t, strings.Contains(eei.returnMessage, "unBond period did not pass"))
	assert.True(t, len(eei.GetStorage(caller)) != 0) //should have not removed the account data
}

func TestValidatorSC_ExecuteUnBondBeforePeriodEndsForV2(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(100)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return 10
		},
	}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	args.StakingSCConfig.UnBondPeriod = 1000
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	validatorSc, _ := NewValidatorSmartContract(args)
	eei.SetSCAddress([]byte("staking"))

	blsPubKey := []byte("pubkey")
	_ = validatorSc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RewardAddress: caller,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedNonce: 1,
					UnstakedValue: big.NewInt(1000),
				},
			},
			BlsPubKeys:      [][]byte{blsPubKey},
			TotalStakeValue: big.NewInt(0),
			LockedStake:     big.NewInt(0),
			TotalUnstaked:   big.NewInt(1000),
		},
	)

	systemSc, _ := eei.GetContract(nil)
	stakingSc := systemSc.(*stakingSC)
	registrationData := &StakedDataV2_0{
		RegisterNonce: 0,
		UnStakedNonce: uint64(1),
		RewardAddress: caller,
		StakeValue:    big.NewInt(100),
	}
	_ = stakingSc.saveAsStakingDataV1P1(blsPubKey, registrationData)

	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = caller
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{blsPubKey}

	retCode := validatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
	assert.True(t, strings.Contains(eei.returnMessage, "unBond is not possible"))
	assert.True(t, strings.Contains(eei.returnMessage, "unBond period did not pass"))
	assert.True(t, len(eei.GetStorage(caller)) != 0) //should have not removed the account data
}

func TestValidatorStakingSC_ExecuteUnBond(t *testing.T) {
	t.Parallel()

	unBondPeriod := uint64(100)
	unStakedNonce := uint64(10)
	stakeValue := big.NewInt(100)
	stakedData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: unStakedNonce,
		RewardAddress: []byte("reward"),
		StakeValue:    big.NewInt(0).Set(stakeValue),
		JailedRound:   math.MaxUint64,
	}

	marshalizedStakedData, _ := json.Marshal(&stakedData)
	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(&mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return unStakedNonce + unBondPeriod + 1
		},
	}, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	scAddress := []byte("owner")
	eei.SetSCAddress(scAddress)
	eei.SetStorage([]byte(ownerKey), scAddress)

	args := createMockArgumentsForValidatorSC()
	args.Eei = eei
	args.StakingSCConfig.GenesisNodePrice = stakeValue.Text(10)
	args.StakingSCConfig.UnBondPeriod = unBondPeriod
	args.EpochConfig.EnableEpochs.StakeEnableEpoch = 0

	argsStaking := createMockStakingScArguments()
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig = args.StakingSCConfig
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	stakingSmartContract, _ := NewValidatorSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("address")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{[]byte("abc")}
	arguments.RecipientAddr = scAddress

	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(arguments.Arguments[0], marshalizedStakedData)
	stakingSc.setConfig(&StakingNodesConfig{MinNumNodes: 5, StakedNodes: 10})
	eei.SetSCAddress(args.ValidatorSCAddress)

	validatorData := ValidatorDataV2{
		RewardAddress:   arguments.CallerAddr,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{arguments.Arguments[0]},
		TotalStakeValue: stakeValue,
		LockedStake:     stakeValue,
		MaxStakePerNode: stakeValue,
		NumRegistered:   1,
	}
	marshaledRegistrationData, _ := json.Marshal(validatorData)
	eei.SetStorage(arguments.CallerAddr, marshaledRegistrationData)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	data := eei.GetStorage(arguments.Arguments[0])
	assert.Equal(t, 0, len(data))

	destinationBalance := stakingSmartContract.eei.GetBalance(arguments.CallerAddr)
	scBalance := stakingSmartContract.eei.GetBalance(scAddress)
	assert.Equal(t, 0, destinationBalance.Cmp(stakeValue))
	assert.Equal(t, 0, scBalance.Cmp(big.NewInt(0).Mul(stakeValue, big.NewInt(-1))))
}

func TestValidatorStakingSC_ExecuteSlashOwnerAddrNotOkShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestValidatorStakingSC_ExecuteUnStakeAndUnBondStake(t *testing.T) {
	t.Parallel()

	// Preparation
	unBondPeriod := uint64(100)
	valueStakedByTheCaller := big.NewInt(100)
	stakerAddress := []byte("address")
	stakerPubKey := []byte("pubKey")
	blockChainHook := &mock.BlockChainHookStub{}
	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	smartcontractAddress := "validator"
	eei.SetSCAddress([]byte(smartcontractAddress))

	args := createMockArgumentsForValidatorSC()
	args.Eei = eei
	args.StakingSCConfig.UnBondPeriod = unBondPeriod
	args.StakingSCConfig.GenesisNodePrice = valueStakedByTheCaller.Text(10)
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	args.EpochConfig.EnableEpochs.StakeEnableEpoch = 0

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	stakingSmartContract, _ := NewValidatorSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallerAddr = stakerAddress
	arguments.RecipientAddr = []byte(smartcontractAddress)

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
	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)
	stakingSc.setConfig(&StakingNodesConfig{MinNumNodes: 5, StakedNodes: 10})

	validatorData := ValidatorDataV2{
		RewardAddress:   arguments.CallerAddr,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{arguments.Arguments[0]},
		TotalStakeValue: valueStakedByTheCaller,
		LockedStake:     valueStakedByTheCaller,
		MaxStakePerNode: valueStakedByTheCaller,
		NumRegistered:   1,
		TotalUnstaked:   big.NewInt(0),
	}
	marshaledRegistrationData, _ := json.Marshal(validatorData)
	eei.SetSCAddress(args.ValidatorSCAddress)
	eei.SetStorage(arguments.CallerAddr, marshaledRegistrationData)

	arguments.Function = "unStake"

	unStakeNonce := uint64(10)
	blockChainHook.CurrentNonceCalled = func() uint64 {
		return unStakeNonce
	}
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData StakedDataV2_0
	eei.SetSCAddress(args.StakingSCAddress)
	data := eei.GetStorage(arguments.Arguments[0])
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
	eei.SetSCAddress(args.ValidatorSCAddress)
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	destinationBalance := eei.GetBalance(arguments.CallerAddr)
	senderBalance := eei.GetBalance([]byte(smartcontractAddress))
	assert.Equal(t, big.NewInt(100), destinationBalance)
	assert.Equal(t, big.NewInt(-100), senderBalance)
}

func TestValidatorStakingSC_ExecuteGetShouldReturnUserErr(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInput()
	arguments.Function = "get"
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{}, &mock.RaterMock{})
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	err := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.UserError, err)
}

func TestValidatorStakingSC_ExecuteGetShouldOk(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInput()
	arguments.Function = "get"
	arguments.Arguments = [][]byte{arguments.CallerAddr}
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{}, &mock.RaterMock{})
	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContract(args)
	err := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.Ok, err)
}

// Test scenario
// 1 -- will call claim from a account that does not stake -> will return error code
// 2 -- will do stake and lock all the stake value and claim should return error code because all the stake value is locked
// 3 -- will do stake and stake value will not be locked and after that claim should work
func TestValidatorStakingSC_Claim(t *testing.T) {
	t.Parallel()

	receiverAddr := []byte("receiverAddress")
	stakerAddress := []byte("stakerAddr")
	stakerPubKey := []byte("stakerPubKey")
	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	nodesToRunBytes := big.NewInt(1).Bytes()

	nonce := uint64(0)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			defer func() {
				nonce++
			}()

			return nonce
		},
	}

	args := createMockArgumentsForValidatorSC()
	args.Eei = createVmContextWithStakingSc(minStakeValue, unboundPeriod, blockChainHook)

	sc, _ := NewValidatorSmartContract(args)

	//do claim should ret error
	doClaim(t, sc, stakerAddress, receiverAddr, vmcommon.UserError)

	//do stake
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	stake(t, sc, nodePrice, receiverAddr, stakerAddress, stakerPubKey, nodesToRunBytes)

	//do claim all stake is locked should return Ok
	doClaim(t, sc, stakerAddress, receiverAddr, vmcommon.Ok)

	// do stake to add more money but not lock the stake
	nonce = 0
	stake(t, sc, big.NewInt(1000), receiverAddr, stakerAddress, stakerPubKey, nodesToRunBytes)

	// do claim should work because not all the stake is locked
	doClaim(t, sc, stakerAddress, receiverAddr, vmcommon.Ok)
}

// Test scenario
// 1 -- call setConfig with wrong owner address should return error
// 2 -- call validator smart contract init and after that call setConfig with wrong number of arguments should return error
// 3 -- call setConfig after init was done successfully should work and config should be set correctly
func TestValidatorStakingSC_SetConfig(t *testing.T) {
	t.Parallel()

	ownerAddr := []byte("ownerAddress")
	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	args.Eei = createVmContextWithStakingSc(minStakeValue, unboundPeriod, blockChainHook)

	sc, _ := NewValidatorSmartContract(args)

	// call setConfig should return error -> wrong owner address
	arguments := CreateVmContractCallInput()
	arguments.Function = "setConfig"
	retCode := sc.Execute(arguments)
	require.Equal(t, vmcommon.UserError, retCode)

	// call validator smart contract init
	arguments.Function = core.SCDeployInitFunctionName
	arguments.CallerAddr = ownerAddr
	retCode = sc.Execute(arguments)
	require.Equal(t, vmcommon.Ok, retCode)

	// call setConfig return error -> wrong number of arguments
	arguments.Function = "setConfig"
	retCode = sc.Execute(arguments)
	require.Equal(t, vmcommon.UserError, retCode)

	// call setConfig
	totalSupply := big.NewInt(10000000)
	minStep := big.NewInt(100)
	nodPrice := big.NewInt(20000)
	epoch := big.NewInt(1)
	unjailPrice := big.NewInt(100)
	arguments.Function = "setConfig"
	arguments.Arguments = [][]byte{minStakeValue.Bytes(), totalSupply.Bytes(), minStep.Bytes(),
		nodPrice.Bytes(), unjailPrice.Bytes(), epoch.Bytes()}
	retCode = sc.Execute(arguments)
	require.Equal(t, vmcommon.Ok, retCode)

	validatorConfig := sc.getConfig(1)
	require.NotNil(t, validatorConfig)
	require.Equal(t, totalSupply, validatorConfig.TotalSupply)
	require.Equal(t, minStep, validatorConfig.MinStep)
	require.Equal(t, nodPrice, validatorConfig.NodePrice)
	require.Equal(t, unjailPrice, validatorConfig.UnJailPrice)
	require.Equal(t, minStakeValue, validatorConfig.MinStakeValue)
}

func TestValidatorStakingSC_SetConfig_InvalidParameters(t *testing.T) {
	t.Parallel()

	ownerAddr := []byte("ownerAddress")
	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	eei := createVmContextWithStakingSc(minStakeValue, unboundPeriod, blockChainHook)
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)

	// call setConfig should return error -> wrong owner address
	arguments := CreateVmContractCallInput()
	arguments.Function = "setConfig"

	// call validator smart contract init
	arguments.Function = core.SCDeployInitFunctionName
	arguments.CallerAddr = ownerAddr
	_ = sc.Execute(arguments)

	totalSupply := big.NewInt(10000000)
	minStep := big.NewInt(100)
	nodPrice := big.NewInt(20000)
	epoch := big.NewInt(1)
	unjailPrice := big.NewInt(100)
	arguments.Function = "setConfig"

	arguments.Arguments = [][]byte{minStakeValue.Bytes(), zero.Bytes(), minStep.Bytes(),
		nodPrice.Bytes(), unjailPrice.Bytes(), epoch.Bytes()}
	retCode := sc.Execute(arguments)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidGenesisTotalSupply.Error()))

	arguments.Arguments = [][]byte{minStakeValue.Bytes(), totalSupply.Bytes(), zero.Bytes(),
		nodPrice.Bytes(), unjailPrice.Bytes(), epoch.Bytes()}
	retCode = sc.Execute(arguments)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidMinStepValue.Error()))

	arguments.Arguments = [][]byte{minStakeValue.Bytes(), totalSupply.Bytes(), minStep.Bytes(),
		zero.Bytes(), unjailPrice.Bytes(), epoch.Bytes()}
	retCode = sc.Execute(arguments)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNodePrice.Error()))

	arguments.Arguments = [][]byte{minStakeValue.Bytes(), totalSupply.Bytes(), minStep.Bytes(),
		nodPrice.Bytes(), zero.Bytes(), epoch.Bytes()}
	retCode = sc.Execute(arguments)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidUnJailCost.Error()))
}

func TestValidatorStakingSC_getBlsStatusWrongCaller(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	eei := createVmContextWithStakingSc(minStakeValue, unboundPeriod, blockChainHook)
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "getBlsKeysStatus"
	arguments.CallerAddr = []byte("wrong caller")

	returnCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.True(t, strings.Contains(eei.returnMessage, "this is only a view function"))
}

func TestValidatorStakingSC_getBlsStatusWrongNumOfArguments(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	eei := createVmContextWithStakingSc(minStakeValue, unboundPeriod, blockChainHook)
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "getBlsKeysStatus"

	returnCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.True(t, strings.Contains(eei.returnMessage, "number of arguments must be equal to 1"))
}

func TestValidatorStakingSC_getBlsStatusWrongRegistrationData(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	eei := createVmContextWithStakingSc(minStakeValue, unboundPeriod, blockChainHook)

	wrongStorageEntry := make(map[string][]byte)
	wrongStorageEntry["erdKey"] = []byte("entry val")
	eei.storageUpdate["addr"] = wrongStorageEntry
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Arguments = append(arguments.Arguments, []byte("erdKey"))
	arguments.Function = "getBlsKeysStatus"

	returnCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot get or create registration data: error "))
}

func TestValidatorStakingSC_getBlsStatusNoBlsKeys(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	eei := createVmContextWithStakingSc(minStakeValue, unboundPeriod, blockChainHook)
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "getBlsKeysStatus"
	arguments.Arguments = append(arguments.Arguments, []byte("erd key"))

	returnCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, returnCode)
	assert.True(t, strings.Contains(eei.returnMessage, "no bls keys"))
}

func TestValidatorStakingSC_getBlsStatusShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	eei := createVmContextWithStakingSc(minStakeValue, unboundPeriod, blockChainHook)

	firstAddr := "addr 1"
	secondAddr := "addr 2"
	validatorData := ValidatorDataV2{
		BlsPubKeys: [][]byte{[]byte(firstAddr), []byte(secondAddr)},
	}
	serializedValidatorData, _ := args.Marshalizer.Marshal(validatorData)

	registrationData1 := &StakedDataV2_0{
		Staked:        true,
		UnStakedEpoch: core.DefaultUnstakedEpoch,
		RewardAddress: []byte("rewards addr"),
		JailedRound:   math.MaxUint64,
		StakedNonce:   math.MaxUint64,
	}
	serializedRegistrationData1, _ := args.Marshalizer.Marshal(registrationData1)

	registrationData2 := &StakedDataV2_0{
		UnStakedEpoch: core.DefaultUnstakedEpoch,
		RewardAddress: []byte("rewards addr"),
		JailedRound:   math.MaxUint64,
		StakedNonce:   math.MaxUint64,
	}
	serializedRegistrationData2, _ := args.Marshalizer.Marshal(registrationData2)

	storageEntry := make(map[string][]byte)
	storageEntry["erdKey"] = serializedValidatorData
	eei.storageUpdate["addr"] = storageEntry

	stakingEntry := make(map[string][]byte)
	stakingEntry[firstAddr] = serializedRegistrationData1
	stakingEntry[secondAddr] = serializedRegistrationData2
	eei.storageUpdate["staking"] = stakingEntry
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Arguments = append(arguments.Arguments, []byte("erdKey"))
	arguments.Function = "getBlsKeysStatus"

	returnCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, returnCode)

	output := eei.CreateVMOutput()
	assert.Equal(t, 4, len(output.ReturnData))
	assert.Equal(t, []byte(firstAddr), output.ReturnData[0])
	assert.Equal(t, []byte("staked"), output.ReturnData[1])
	assert.Equal(t, []byte(secondAddr), output.ReturnData[2])
	assert.Equal(t, []byte("unStaked"), output.ReturnData[3])
}

func TestValidatorStakingSC_getBlsStatusShouldWorkEvenIfAnErrorOccursForOneOfTheBlsKeys(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	eei := createVmContextWithStakingSc(minStakeValue, unboundPeriod, blockChainHook)

	firstAddr := "addr 1"
	secondAddr := "addr 2"
	validatorData := ValidatorDataV2{
		BlsPubKeys: [][]byte{[]byte(firstAddr), []byte(secondAddr)},
	}
	serializedValidatorData, _ := args.Marshalizer.Marshal(validatorData)

	registrationData := &StakedDataV2_0{
		Staked:        true,
		UnStakedEpoch: core.DefaultUnstakedEpoch,
		RewardAddress: []byte("rewards addr"),
		JailedRound:   math.MaxUint64,
		StakedNonce:   math.MaxUint64,
	}
	serializedRegistrationData, _ := args.Marshalizer.Marshal(registrationData)

	storageEntry := make(map[string][]byte)
	storageEntry["erdKey"] = serializedValidatorData
	eei.storageUpdate["addr"] = storageEntry

	stakingEntry := make(map[string][]byte)
	stakingEntry[firstAddr] = []byte("wrong data for first bls key")
	stakingEntry[secondAddr] = serializedRegistrationData
	eei.storageUpdate["staking"] = stakingEntry
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Arguments = append(arguments.Arguments, []byte("erdKey"))
	arguments.Function = "getBlsKeysStatus"

	returnCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, returnCode)

	output := eei.CreateVMOutput()
	assert.Equal(t, 2, len(output.ReturnData))
	assert.Equal(t, []byte(secondAddr), output.ReturnData[0])
	assert.Equal(t, []byte("staked"), output.ReturnData[1])
}

func TestValidatorStakingSC_ChangeRewardAddress(t *testing.T) {
	t.Parallel()

	receiverAddr := []byte("receiverAddress")
	stakerAddress := []byte("stakerA")
	stakerPubKey := []byte("stakerP")
	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	nodesToRunBytes := big.NewInt(1).Bytes()
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	args.Eei = createVmContextWithStakingSc(minStakeValue, unboundPeriod, blockChainHook)

	sc, _ := NewValidatorSmartContract(args)

	//change reward address should error nil arguments
	changeRewardAddress(t, sc, stakerAddress, nil, vmcommon.UserError)
	// change reward address should error wrong address
	changeRewardAddress(t, sc, stakerAddress, []byte("wrongAddress"), vmcommon.UserError)
	// change reward address should error because address is not belongs to any validator
	newRewardAddr := []byte("newAddr11")
	changeRewardAddress(t, sc, stakerAddress, newRewardAddr, vmcommon.UserError)
	//do stake
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	stake(t, sc, nodePrice, receiverAddr, stakerAddress, stakerPubKey, nodesToRunBytes)

	// change reward address should error because new reward address is equal with old reward address
	changeRewardAddress(t, sc, stakerAddress, stakerAddress, vmcommon.UserError)
	// change reward address should work
	changeRewardAddress(t, sc, stakerAddress, newRewardAddr, vmcommon.Ok)
}

func TestStakingValidatorSC_UnstakeTokensNotEnabledShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)

	callFunctionAndCheckResult(t, "unStakeTokens", sc, caller, [][]byte{big.NewInt(1).Bytes()}, zero, vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, "invalid method to call", vmOutput.ReturnMessage)
}

func TestStakingValidatorSC_UnstakeTokensInvalidArgumentsShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)

	registrationData := &ValidatorDataV2{RewardAddress: caller}
	marshaledData, _ := args.Marshalizer.Marshal(registrationData)
	eei.SetStorage(caller, marshaledData)
	callFunctionAndCheckResult(t, "unStakeTokens", sc, caller, nil, zero, vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, "should have specified one argument containing the unstake value", vmOutput.ReturnMessage)

	eei = createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller = []byte("caller")
	sc, _ = NewValidatorSmartContract(args)

	eei.SetStorage(caller, marshaledData)
	callFunctionAndCheckResult(t, "unStakeTokens", sc, caller, [][]byte{[]byte("a"), []byte("b")}, zero, vmcommon.UserError)
	vmOutput = eei.CreateVMOutput()
	assert.Equal(t, "should have specified one argument containing the unstake value", vmOutput.ReturnMessage)
}

func TestStakingValidatorSC_UnstakeTokensWithCallValueShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)

	callFunctionAndCheckResult(t, "unStakeTokens", sc, caller, [][]byte{big.NewInt(1).Bytes()}, big.NewInt(1), vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, vm.TransactionValueMustBeZero, vmOutput.ReturnMessage)
}

func TestStakingValidatorSC_UnstakeTokensOverMaxShouldUnStake(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	startNonce := uint64(56)
	nonce := startNonce
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			nonce++
			return nonce
		},
	}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1010),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo:    nil,
			TotalUnstaked:   nil,
		},
	)

	callFunctionAndCheckResult(t, "unStakeTokens", sc, caller, [][]byte{big.NewInt(11).Bytes()}, zero, vmcommon.Ok)

	registrationData, _ := sc.getOrCreateRegistrationData(caller)
	assert.Equal(t, 1, len(registrationData.UnstakedInfo))
	assert.True(t, big.NewInt(999).Cmp(registrationData.TotalStakeValue) == 0)
	assert.True(t, registrationData.UnstakedInfo[0].UnstakedValue.Cmp(big.NewInt(11)) == 0)
}

func TestStakingValidatorSC_UnstakeTokensUnderMinimumAllowedShouldErr(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	startNonce := uint64(56)
	nonce := startNonce
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			nonce++
			return nonce
		},
	}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	args.StakingSCConfig.MinUnstakeTokensValue = "2"
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1010),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo:    nil,
			TotalUnstaked:   nil,
		},
	)

	callFunctionAndCheckResult(t, "unStakeTokens", sc, caller, [][]byte{big.NewInt(1).Bytes()}, zero, vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.True(t, strings.Contains(vmOutput.ReturnMessage, "can not unstake the provided value either because is under the minimum threshold"))
}

func TestStakingValidatorSC_UnstakeAllTokensWithActiveNodesShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	startNonce := uint64(56)
	nonce := startNonce
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			nonce++
			return nonce
		},
	}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	args.MinDeposit = "1000"
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1010),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo:    nil,
			TotalUnstaked:   nil,
		},
	)

	callFunctionAndCheckResult(t, "unStakeTokens", sc, caller, [][]byte{big.NewInt(11).Bytes()}, zero, vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.True(t, strings.Contains(vmOutput.ReturnMessage, "cannot unStake tokens, the validator would remain without min deposit, nodes are still active"))
}

func TestStakingValidatorSC_UnstakeTokensShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	startNonce := uint64(56)
	nonce := startNonce
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			nonce++
			return nonce
		},
	}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1010),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo:    nil,
			TotalUnstaked:   nil,
		},
	)

	callFunctionAndCheckResult(t, "unStakeTokens", sc, caller, [][]byte{big.NewInt(1).Bytes()}, zero, vmcommon.Ok)
	callFunctionAndCheckResult(t, "unStakeTokens", sc, caller, [][]byte{big.NewInt(2).Bytes()}, zero, vmcommon.Ok)

	expected := &ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   caller,
		TotalStakeValue: big.NewInt(1007),
		LockedStake:     big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0),
		BlsPubKeys:      [][]byte{[]byte("key")},
		NumRegistered:   1,
		UnstakedInfo: []*UnstakedValue{
			{
				UnstakedNonce: startNonce + 1,
				UnstakedValue: big.NewInt(1),
			},
			{
				UnstakedNonce: startNonce + 2,
				UnstakedValue: big.NewInt(2),
			},
		},
		TotalUnstaked: big.NewInt(3),
	}

	recovered, err := sc.getOrCreateRegistrationData(caller)
	require.Nil(t, err)

	assert.Equal(t, expected, recovered)
}

func TestStakingValidatorSC_UnstakeTokensHavingUnstakedShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	startNonce := uint64(56)
	nonce := startNonce
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			nonce++
			return nonce
		},
	}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1010),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedNonce: 1,
					UnstakedValue: big.NewInt(5),
				},
			},
			TotalUnstaked: big.NewInt(5),
		},
	)

	callFunctionAndCheckResult(t, "unStakeTokens", sc, caller, [][]byte{big.NewInt(6).Bytes()}, zero, vmcommon.Ok)

	expected := &ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   caller,
		TotalStakeValue: big.NewInt(1004),
		LockedStake:     big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0),
		BlsPubKeys:      [][]byte{[]byte("key")},
		NumRegistered:   1,
		UnstakedInfo: []*UnstakedValue{
			{
				UnstakedNonce: 1,
				UnstakedValue: big.NewInt(5),
			},
			{
				UnstakedNonce: startNonce + 1,
				UnstakedValue: big.NewInt(6),
			},
		},
		TotalUnstaked: big.NewInt(11),
	}

	recovered, err := sc.getOrCreateRegistrationData(caller)
	require.Nil(t, err)

	assert.Equal(t, expected, recovered)
}

func TestStakingValidatorSC_UnstakeAllTokensShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	startNonce := uint64(56)
	nonce := startNonce
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			nonce++
			return nonce
		},
	}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1010),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo:    nil,
			TotalUnstaked:   nil,
		},
	)

	callFunctionAndCheckResult(t, "unStakeTokens", sc, caller, [][]byte{big.NewInt(10).Bytes()}, zero, vmcommon.Ok)

	expected := &ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   caller,
		TotalStakeValue: big.NewInt(1000),
		LockedStake:     big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0),
		BlsPubKeys:      [][]byte{[]byte("key")},
		NumRegistered:   1,
		UnstakedInfo: []*UnstakedValue{
			{
				UnstakedNonce: startNonce + 1,
				UnstakedValue: big.NewInt(10),
			},
		},
		TotalUnstaked: big.NewInt(10),
	}

	recovered, err := sc.getOrCreateRegistrationData(caller)
	require.Nil(t, err)

	assert.Equal(t, expected, recovered)
}

func TestStakingValidatorSC_UnbondTokensNotEnabledShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)

	callFunctionAndCheckResult(t, "unBondTokens", sc, caller, nil, zero, vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, "invalid method to call", vmOutput.ReturnMessage)
}

func TestStakingValidatorSC_UnbondTokensOneArgument(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	startNonce := uint64(56)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return startNonce + unbondPeriod
		},
	}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	args.StakingSCConfig.UnBondPeriod = unbondPeriod
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1000),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedNonce: startNonce - 2,
					UnstakedValue: big.NewInt(1),
				},
				{
					UnstakedNonce: startNonce - 1,
					UnstakedValue: big.NewInt(2),
				},
				{
					UnstakedNonce: startNonce,
					UnstakedValue: big.NewInt(3),
				},
				{
					UnstakedNonce: startNonce + 1,
					UnstakedValue: big.NewInt(4),
				},
			},
			TotalUnstaked: big.NewInt(10),
		},
	)

	unBondRequest := big.NewInt(4)
	callFunctionAndCheckResult(t, "unBondTokens", sc, caller, [][]byte{unBondRequest.Bytes()}, zero, vmcommon.Ok)

	expected := &ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   caller,
		TotalStakeValue: big.NewInt(1000),
		LockedStake:     big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0),
		BlsPubKeys:      [][]byte{[]byte("key")},
		NumRegistered:   1,
		UnstakedInfo: []*UnstakedValue{
			{
				UnstakedNonce: startNonce,
				UnstakedValue: big.NewInt(2),
			},
			{
				UnstakedNonce: startNonce + 1,
				UnstakedValue: big.NewInt(4),
			},
		},
		TotalUnstaked: big.NewInt(6),
	}

	recovered, err := sc.getOrCreateRegistrationData(caller)
	require.Nil(t, err)
	assert.Equal(t, expected, recovered)

	outTransferValue := eei.outputAccounts[string(caller)].OutputTransfers[0].Value
	assert.True(t, unBondRequest.Cmp(outTransferValue) == 0)
}

func TestStakingValidatorSC_UnbondTokensWithCallValueShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)

	callFunctionAndCheckResult(t, "unBondTokens", sc, caller, nil, big.NewInt(1), vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, vm.TransactionValueMustBeZero, vmOutput.ReturnMessage)
}

func TestStakingValidatorSC_UnBondTokensShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	startNonce := uint64(56)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return startNonce + unbondPeriod
		},
	}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	args.StakingSCConfig.UnBondPeriod = unbondPeriod
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1000),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedNonce: startNonce - 2,
					UnstakedValue: big.NewInt(1),
				},
				{
					UnstakedNonce: startNonce - 1,
					UnstakedValue: big.NewInt(2),
				},
				{
					UnstakedNonce: startNonce,
					UnstakedValue: big.NewInt(3),
				},
				{
					UnstakedNonce: startNonce + 1,
					UnstakedValue: big.NewInt(4),
				},
			},
			TotalUnstaked: big.NewInt(10),
		},
	)

	callFunctionAndCheckResult(t, "unBondTokens", sc, caller, nil, zero, vmcommon.Ok)

	expected := &ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   caller,
		TotalStakeValue: big.NewInt(1000),
		LockedStake:     big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0),
		BlsPubKeys:      [][]byte{[]byte("key")},
		NumRegistered:   1,
		UnstakedInfo: []*UnstakedValue{
			{
				UnstakedNonce: startNonce + 1,
				UnstakedValue: big.NewInt(4),
			},
		},
		TotalUnstaked: big.NewInt(4),
	}

	recovered, err := sc.getOrCreateRegistrationData(caller)
	require.Nil(t, err)

	assert.Equal(t, expected, recovered)
}

func TestStakingValidatorSC_UnBondAllTokensWithMinDepositShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	startNonce := uint64(56)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return startNonce + unbondPeriod + 1
		},
	}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	args.MinDeposit = "1000"
	args.StakingSCConfig.UnBondPeriod = unbondPeriod
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(999),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedNonce: startNonce - 2,
					UnstakedValue: big.NewInt(1),
				},
			},
			TotalUnstaked: big.NewInt(10),
		},
	)

	callFunctionAndCheckResult(t, "unBondTokens", sc, caller, nil, zero, vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.True(t, strings.Contains(vmOutput.ReturnMessage, "cannot unBond tokens, the validator would remain without min deposit, nodes are still active"))
}

func TestStakingValidatorSC_UnBondAllTokensShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	startNonce := uint64(56)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return startNonce + unbondPeriod + 1
		},
	}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	args.StakingSCConfig.UnBondPeriod = unbondPeriod
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1000),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedNonce: startNonce - 2,
					UnstakedValue: big.NewInt(1),
				},
				{
					UnstakedNonce: startNonce - 1,
					UnstakedValue: big.NewInt(2),
				},
				{
					UnstakedNonce: startNonce,
					UnstakedValue: big.NewInt(3),
				},
				{
					UnstakedNonce: startNonce + 1,
					UnstakedValue: big.NewInt(4),
				},
			},
			TotalUnstaked: big.NewInt(10),
		},
	)

	callFunctionAndCheckResult(t, "unBondTokens", sc, caller, nil, zero, vmcommon.Ok)

	expected := &ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   caller,
		TotalStakeValue: big.NewInt(1000),
		LockedStake:     big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0),
		BlsPubKeys:      [][]byte{[]byte("key")},
		NumRegistered:   1,
		UnstakedInfo:    make([]*UnstakedValue, 0),
		TotalUnstaked:   big.NewInt(0),
	}

	recovered, err := sc.getOrCreateRegistrationData(caller)
	require.Nil(t, err)

	assert.Equal(t, expected, recovered)
}

func TestStakingValidatorSC_UpdateStakingV2NotEnabledShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 10
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	sc, _ := NewValidatorSmartContract(args)

	callFunctionAndCheckResult(t, "updateStakingV2", sc, args.ValidatorSCAddress, make([][]byte, 0), zero, vmcommon.UserError)

	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, "invalid method to call", vmOutput.ReturnMessage)
}

func TestStakingValidatorSC_GetTopUpNotEnabledShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)

	callFunctionAndCheckResult(t, "getTopUp", sc, caller, nil, zero, vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, "invalid method to call", vmOutput.ReturnMessage)
}

func TestStakingValidatorSC_GetTopUpTotalStakedWithValueShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)

	callFunctionAndCheckResult(t, "getTotalStakedTopUpStakedBlsKeys", sc, caller, [][]byte{caller}, big.NewInt(1), vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, vm.TransactionValueMustBeZero, vmOutput.ReturnMessage)
}

func TestStakingValidatorSC_GetTopUpTotalStakedInsufficientGasShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.Get = 1
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)

	callFunctionAndCheckResult(t, "getTotalStakedTopUpStakedBlsKeys", sc, caller, [][]byte{caller}, big.NewInt(0), vmcommon.OutOfGas)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, vm.InsufficientGasLimit, vmOutput.ReturnMessage)
}

func TestStakingValidatorSC_GetTopUpTotalStakedCallerDoesNotExistShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)

	callFunctionAndCheckResult(t, "getTotalStakedTopUpStakedBlsKeys", sc, caller, [][]byte{caller}, big.NewInt(0), vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, "caller not registered in staking/validator sc", vmOutput.ReturnMessage)
}

func TestStakingValidatorSC_GetTopUpTotalStakedShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	eei := createVmContextWithStakingSc(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContract(args)

	totalStake := big.NewInt(33827)
	lockedStake := big.NewInt(4564)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: totalStake,
			LockedStake:     lockedStake,
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      make([][]byte, 0),
			NumRegistered:   1,
			UnstakedInfo:    nil,
			TotalUnstaked:   nil,
		},
	)

	callFunctionAndCheckResult(t, "getTotalStakedTopUpStakedBlsKeys", sc, caller, [][]byte{caller}, big.NewInt(0), vmcommon.Ok)
	vmOutput := eei.CreateVMOutput()

	assert.Equal(t, totalStake.Bytes(), vmOutput.ReturnData[0])
	assert.Equal(t, totalStake.Bytes(), vmOutput.ReturnData[1])
}

func TestMarshalingBetweenValidatorV1AndValidatorV2(t *testing.T) {
	t.Parallel()

	validatorV1 := &ValidatorDataV1{
		RegisterNonce:   1,
		Epoch:           2,
		RewardAddress:   []byte("reward address"),
		TotalStakeValue: big.NewInt(3),
		LockedStake:     big.NewInt(4),
		MaxStakePerNode: big.NewInt(5),
		BlsPubKeys:      [][]byte{[]byte("bls1"), []byte("bls2")},
		NumRegistered:   6,
	}

	marshalizer := &marshal.GogoProtoMarshalizer{}

	buff, err := marshalizer.Marshal(validatorV1)
	require.Nil(t, err)

	validatorV2 := &ValidatorDataV2{}

	err = marshalizer.Unmarshal(validatorV2, buff)
	require.Nil(t, err)
}

func TestStakingValidatorSC_UnStakeUnBondFromWaitingList(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("address")
	stakerPubKey1 := []byte("blsKey1")
	stakerPubKey2 := []byte("blsKey2")
	stakerPubKey3 := []byte("blsKey3")

	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return 100000
		},
	}
	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	argsStaking.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	argsStaking.StakingSCConfig.MaxNumberOfNodesForStake = 1
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args := createMockArgumentsForValidatorSC()
	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey1, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey2, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey3, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey2, stakerPubKey3}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey2, stakerPubKey3}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	registrationData, _ := sc.getOrCreateRegistrationData(arguments.CallerAddr)
	assert.Equal(t, len(registrationData.UnstakedInfo), 0)

	eei.returnMessage = ""
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey1}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Contains(t, eei.returnMessage, "could not unstake any nodes")
}

func TestStakingValidatorSC_StakeUnStakeUnBondTokensNoNodes(t *testing.T) {
	t.Parallel()

	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return 100000
		},
	}
	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	argsStaking.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	argsStaking.StakingSCConfig.MaxNumberOfNodesForStake = 1
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args := createMockArgumentsForValidatorSC()
	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = vm.DelegationManagerSCAddress
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStakeTokens"
	arguments.Arguments = [][]byte{big.NewInt(10000000).Bytes()}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unBondTokens"
	arguments.Arguments = [][]byte{big.NewInt(10000000).Bytes()}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestValidatorStakingSC_UnStakeUnBondPaused(t *testing.T) {
	t.Parallel()

	receiverAddr := []byte("receiverAddress")
	stakerAddress := []byte("stakerAddr")
	stakerPubKey := []byte("stakerPubKey")
	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	nodesToRunBytes := big.NewInt(1).Bytes()

	nonce := uint64(0)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			defer func() {
				nonce++
			}()
			return nonce
		},
	}

	args := createMockArgumentsForValidatorSC()
	args.EpochConfig.EnableEpochs.StakingV2Epoch = 0
	eei := createVmContextWithStakingSc(minStakeValue, unboundPeriod, blockChainHook)
	args.Eei = eei

	sc, _ := NewValidatorSmartContract(args)

	//do stake
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	stake(t, sc, nodePrice, receiverAddr, stakerAddress, stakerPubKey, nodesToRunBytes)

	togglePauseUnStakeUnBond(t, sc, true)

	expectedReturnMessage := "unStake/unBond is paused as not enough total staked in protocol"
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{stakerPubKey}

	eei.returnMessage = ""
	errCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)
	assert.Equal(t, eei.returnMessage, expectedReturnMessage)

	arguments.Function = "unStakeNodes"
	eei.returnMessage = ""
	errCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)
	assert.Equal(t, eei.returnMessage, expectedReturnMessage)

	arguments.Function = "unStakeTokens"
	eei.returnMessage = ""
	errCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)
	assert.Equal(t, eei.returnMessage, expectedReturnMessage)

	arguments.Function = "unBond"
	eei.returnMessage = ""
	errCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)
	assert.Equal(t, eei.returnMessage, expectedReturnMessage)

	arguments.Function = "unBondNodes"
	eei.returnMessage = ""
	errCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)
	assert.Equal(t, eei.returnMessage, expectedReturnMessage)

	arguments.Function = "unBondTokens"
	eei.returnMessage = ""
	errCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)
	assert.Equal(t, eei.returnMessage, expectedReturnMessage)

	togglePauseUnStakeUnBond(t, sc, false)
}

func TestValidatorSC_getUnStakedTokensList_InvalidArgumentsCountShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInput()
	args := createMockArgumentsForValidatorSC()

	retMessage := ""

	eei := &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}
	args.Eei = eei

	stakingValidatorSc, _ := NewValidatorSmartContract(args)

	stakingValidatorSc.flagEnableTopUp.Set()

	arguments.Function = "getUnStakedTokensList"
	arguments.CallValue = big.NewInt(10)
	arguments.Arguments = make([][]byte, 0)

	errCode := stakingValidatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)

	assert.Contains(t, retMessage, "number of arguments")
}

func TestValidatorSC_getUnStakedTokensList_CallValueNotZeroShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInput()
	args := createMockArgumentsForValidatorSC()

	retMessage := ""

	eei := &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}
	args.Eei = eei

	stakingValidatorSc, _ := NewValidatorSmartContract(args)

	stakingValidatorSc.flagEnableTopUp.Set()

	arguments.Function = "getUnStakedTokensList"
	arguments.CallValue = big.NewInt(10)
	arguments.Arguments = append(arguments.Arguments, []byte("pubKey"))

	errCode := stakingValidatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)

	assert.Equal(t, vm.TransactionValueMustBeZero, retMessage)
}

func TestValidatorSC_getUnStakedTokensList(t *testing.T) {
	t.Parallel()

	currentNonce := uint64(12)
	unBondPeriod := uint64(5)

	eeiFinishedValues := make([][]byte, 0)
	arguments := CreateVmContractCallInput()
	validatorData := createABid(25000000, 2, 12500000)
	validatorData.UnstakedInfo = append(validatorData.UnstakedInfo,
		&UnstakedValue{
			UnstakedNonce: 6,
			UnstakedValue: big.NewInt(10),
		},
		&UnstakedValue{
			UnstakedNonce: 10,
			UnstakedValue: big.NewInt(11),
		},
	)
	validatorDataBytes, _ := json.Marshal(&validatorData)

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return validatorDataBytes
	}
	eei.SetStorageCalled = func(key []byte, value []byte) {
		if bytes.Equal(key, arguments.CallerAddr) {
			var validatorDataRecovered ValidatorDataV2
			_ = json.Unmarshal(value, &validatorDataRecovered)
			assert.Equal(t, big.NewInt(26000000), validatorDataRecovered.TotalStakeValue)
		}
	}
	eei.BlockChainHookCalled = func() vm.BlockchainHook {
		return &mock.BlockChainHookStub{
			CurrentNonceCalled: func() uint64 {
				return currentNonce
			},
		}
	}
	eei.FinishCalled = func(value []byte) {
		eeiFinishedValues = append(eeiFinishedValues, value)
	}

	args := createMockArgumentsForValidatorSC()
	args.Eei = eei

	stakingValidatorSc, _ := NewValidatorSmartContract(args)

	stakingValidatorSc.unBondPeriod = unBondPeriod
	stakingValidatorSc.flagEnableTopUp.Set()

	arguments.Function = "getUnStakedTokensList"
	arguments.CallValue = big.NewInt(0)
	arguments.Arguments = append(arguments.Arguments, []byte("pubKey"))

	errCode := stakingValidatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)

	assert.Equal(t, 4, len(eeiFinishedValues))

	expectedValues := [][]byte{
		{10}, // value 10
		{},   // elapsed nonces > unbond period
		{11}, // value 11
		{3},  // number of nonces remaining
	}

	assert.Equal(t, expectedValues, eeiFinishedValues)
}

func createVmContextWithStakingSc(stakeValue *big.Int, unboundPeriod uint64, blockChainHook vm.BlockchainHook) *vmContext {
	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{}, &mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = stakeValue.Text(10)
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = unboundPeriod
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	return eei
}

func doClaim(t *testing.T, asc *validatorSC, stakerAddr, receiverAdd []byte, expectedCode vmcommon.ReturnCode) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "claim"
	arguments.RecipientAddr = receiverAdd
	arguments.CallerAddr = stakerAddr

	retCode := asc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)
}

func stake(t *testing.T, asc *validatorSC, stakeValue *big.Int, receiverAdd, stakerAddr, stakerPubKey, nodesToRunBytes []byte) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.RecipientAddr = receiverAdd
	arguments.CallerAddr = stakerAddr
	arguments.Arguments = [][]byte{nodesToRunBytes, stakerPubKey, []byte("signed")}
	arguments.CallValue = big.NewInt(0).Set(stakeValue)

	retCode := asc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func togglePauseUnStakeUnBond(t *testing.T, v *validatorSC, value bool) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "unPauseUnStakeUnBond"
	arguments.CallerAddr = v.endOfEpochAddress
	arguments.CallValue = big.NewInt(0)

	if value {
		arguments.Function = "pauseUnStakeUnBond"
	}

	retCode := v.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	assert.Equal(t, value, v.isUnStakeUnBondPaused())
}

func changeRewardAddress(t *testing.T, asc *validatorSC, callerAddr, newRewardAddr []byte, expectedCode vmcommon.ReturnCode) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "changeRewardAddress"
	arguments.CallerAddr = callerAddr
	if newRewardAddr == nil {
		arguments.Arguments = nil
	} else {
		arguments.Arguments = [][]byte{newRewardAddr}
	}

	retCode := asc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)
}

func callFunctionAndCheckResult(
	t *testing.T,
	function string,
	asc *validatorSC,
	callerAddr []byte,
	args [][]byte,
	callValue *big.Int,
	expectedCode vmcommon.ReturnCode,
) {
	arguments := CreateVmContractCallInput()
	arguments.Function = function
	arguments.CallerAddr = callerAddr
	arguments.Arguments = args
	arguments.CallValue = callValue

	retCode := asc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)
}
