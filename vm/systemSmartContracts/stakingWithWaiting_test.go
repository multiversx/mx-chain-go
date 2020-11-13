package systemSmartContracts

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/parsers"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	"github.com/stretchr/testify/assert"
)

func TestStakingWaitingSC_ExecuteStakeStakeWaitingUnStake(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("address")
	stakerPubKey := []byte("blsPubKey")

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForAuction()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(
		blockChainHook,
		hooks.NewVMCryptoHook(),
		atArgParser,
		&mock.AccountsStub{},
		&mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	argsStaking.StakingSCConfig.MaxNumberOfNodesForStake = 1
	stakingSC, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blsKey2 := []byte("blsKey2")
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), blsKey2, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey)
	stakedData := &StakedDataV2{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.False(t, stakedData.Staked)

	marshaledData = eei.GetStorage(blsKey2)
	stakedData = &StakedDataV2{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)
}

func TestStakingWaitingSC_ExecuteStakeStakeWaitingUnBondFromWaiting(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("address")
	stakerPubKey := []byte("blsPubKey")

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForAuction()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(
		blockChainHook,
		hooks.NewVMCryptoHook(),
		atArgParser,
		&mock.AccountsStub{},
		&mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	argsStaking.StakingSCConfig.MaxNumberOfNodesForStake = 1
	stakingSC, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
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

	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{blsKey2}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	vmOutput := eei.CreateVMOutput()
	assert.NotNil(t, vmOutput)
	outputAccount := vmOutput.OutputAccounts[string(arguments.CallerAddr)]
	assert.True(t, outputAccount.BalanceDelta.Cmp(big.NewInt(10000000)) == 0)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey)
	stakedData := &StakedDataV2{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)

	marshaledData = eei.GetStorage(blsKey2)
	assert.Equal(t, 0, len(marshaledData))
}

func TestStakingWaitingSC_ExecuteStakeStakeUnStakeStakeUnstake(t *testing.T) {
	// Staking fist key will become staked
	// Staking second key gets into the waiting list
	// UnStaking first key will make the second key staked
	// Stake third key will get into the waiting list
	// UnStake the second key must make it unStaked
	// UnBond for that key should not work first
	// Wait unBond period and do unBond should work
	t.Parallel()

	stakerAddress := []byte("address")
	stakerPubKey := []byte("blsPubKey")

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForAuction()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(
		blockChainHook,
		hooks.NewVMCryptoHook(),
		atArgParser,
		&mock.AccountsStub{},
		&mock.RaterMock{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	argsStaking.StakingSCConfig.MaxNumberOfNodesForStake = 1
	stakingSC, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blsKey2 := []byte("blsKey2")
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), blsKey2, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	marshaledData := eei.GetStorageFromAddress(args.StakingSCAddress, stakerPubKey)
	stakedData := &StakedDataV2{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.False(t, stakedData.Staked)

	marshaledData = eei.GetStorageFromAddress(args.StakingSCAddress, blsKey2)
	stakedData = &StakedDataV2{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)

	blsKey3 := []byte("blsKey3")
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), blsKey3, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	marshaledData = eei.GetStorageFromAddress(args.StakingSCAddress, blsKey3)
	stakedData = &StakedDataV2{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.False(t, stakedData.Staked)
	assert.True(t, stakedData.Waiting)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{blsKey2}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	marshaledData = eei.GetStorageFromAddress(args.StakingSCAddress, blsKey2)
	stakedData = &StakedDataV2{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.False(t, stakedData.Staked)
	assert.False(t, stakedData.Waiting)
}
