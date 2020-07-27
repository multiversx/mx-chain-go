package systemSmartContracts

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	"github.com/stretchr/testify/assert"
)

func TestStakingWaitingSC_ExecuteStakeStakeWaitingUnStake(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("address")
	stakerPubKey := []byte("blsPubKey")

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForAuction()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{})

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
	marshalledData := eei.GetStorage(stakerPubKey)
	stakedData := &StakedData{}
	_ = json.Unmarshal(marshalledData, stakedData)
	assert.False(t, stakedData.Staked)

	marshalledData = eei.GetStorage(blsKey2)
	stakedData = &StakedData{}
	_ = json.Unmarshal(marshalledData, stakedData)
	assert.True(t, stakedData.Staked)
}

func TestStakingWaitingSC_ExecuteStakeStakeWaitingUnBondTheWaiting(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("address")
	stakerPubKey := []byte("blsPubKey")

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForAuction()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{})

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
	marshalledData := eei.GetStorage(stakerPubKey)
	stakedData := &StakedData{}
	_ = json.Unmarshal(marshalledData, stakedData)
	assert.True(t, stakedData.Staked)

	marshalledData = eei.GetStorage(blsKey2)
	assert.Equal(t, 0, len(marshalledData))
}
