package systemSmartContracts

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	"github.com/stretchr/testify/assert"
)

func createMockArgumentsForDelegation() ArgsNewDelegation {
	return ArgsNewDelegation{
		DelegationSCConfig:     config.DelegationSystemSCConfig{MinStakeAmount: "10"},
		StakingSCConfig:        config.StakingSystemSCConfig{MinStakeValue: "10", UnJailValue: "15", GenesisNodePrice: "100"},
		Eei:                    &mock.SystemEIStub{},
		SigVerifier:            &mock.MessageSignVerifierMock{},
		DelegationMgrSCAddress: []byte("delegMgrScAddr"),
		StakingSCAddress:       []byte("stakingScAddr"),
		AuctionSCAddress:       []byte("auctionScAddr"),
		GasCost:                vm.GasCost{MetaChainSystemSCsCost: vm.MetaChainSystemSCsCost{ESDTIssue: 10}},
		Marshalizer:            &mock.MarshalizerMock{},
		EpochNotifier:          &mock.EpochNotifierStub{},
	}
}

func addAuctionAndStakingScToVmContext(eei *vmContext) {
	auctionArgs := createMockArgumentsForAuction()
	auctionArgs.Eei = eei
	auctionArgs.StakingSCConfig.GenesisNodePrice = "100"
	auctionSc, _ := NewStakingAuctionSmartContract(auctionArgs)

	stakingArgs := createMockStakingScArguments()
	stakingArgs.Eei = eei
	stakingSc, _ := NewStakingSmartContract(stakingArgs)

	eei.inputParser = parsers.NewCallArgsParser()

	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		if bytes.Equal(key, []byte("staking")) {
			return stakingSc, nil
		}

		if bytes.Equal(key, []byte("auctionScAddr")) {
			return auctionSc, nil
		}

		return nil, nil
	}})
}

func getDefaultVmInputForFunc(funcName string, args [][]byte) *vmcommon.ContractCallInput {
	return &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:     []byte("owner"),
			Arguments:      args,
			CallValue:      big.NewInt(0),
			CallType:       0,
			GasPrice:       0,
			GasProvided:    0,
			OriginalTxHash: nil,
			CurrentTxHash:  nil,
		},
		RecipientAddr: []byte("addr"),
		Function:      funcName,
	}
}

func TestNewDelegationSystemSC_NilSystemEnvironmentShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	args.Eei = nil

	d, err := NewDelegationSystemSC(args)
	assert.Nil(t, d)
	assert.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
}

func TestNewDelegationSystemSC_InvalidStakingSCAddrShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("%w for staking sc address", vm.ErrInvalidAddress)
	args := createMockArgumentsForDelegation()
	args.StakingSCAddress = []byte{}

	d, err := NewDelegationSystemSC(args)
	assert.Nil(t, d)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationSystemSC_InvalidAuctionSCAddrShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("%w for auction sc address", vm.ErrInvalidAddress)
	args := createMockArgumentsForDelegation()
	args.AuctionSCAddress = []byte{}

	d, err := NewDelegationSystemSC(args)
	assert.Nil(t, d)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationSystemSC_InvalidDelegationMgrSCAddrShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("%w for delegation sc address", vm.ErrInvalidAddress)
	args := createMockArgumentsForDelegation()
	args.DelegationMgrSCAddress = []byte{}

	d, err := NewDelegationSystemSC(args)
	assert.Nil(t, d)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationSystemSC_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	args.Marshalizer = nil

	d, err := NewDelegationSystemSC(args)
	assert.Nil(t, d)
	assert.Equal(t, vm.ErrNilMarshalizer, err)
}

func TestNewDelegationSystemSC_NilEpochNotifierShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	args.EpochNotifier = nil

	d, err := NewDelegationSystemSC(args)
	assert.Nil(t, d)
	assert.Equal(t, vm.ErrNilEpochNotifier, err)
}

func TestNewDelegationSystemSC_NilSigVerifierShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	args.SigVerifier = nil

	d, err := NewDelegationSystemSC(args)
	assert.Nil(t, d)
	assert.Equal(t, vm.ErrNilMessageSignVerifier, err)
}

func TestNewDelegationSystemSC_InvalidMinStakeAmountShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	args.DelegationSCConfig.MinStakeAmount = "-1"

	d, err := NewDelegationSystemSC(args)
	assert.Nil(t, d)
	assert.Equal(t, vm.ErrInvalidMinStakeValue, err)
}

func TestNewDelegationSystemSC_InvalidUnJailValueShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	args.StakingSCConfig.UnJailValue = "-1"
	expectedErr := fmt.Errorf("%w, value is %v", vm.ErrInvalidUnJailCost, args.StakingSCConfig.UnJailValue)

	d, err := NewDelegationSystemSC(args)
	assert.Nil(t, d)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationSystemSC_InvalidMinStakeValueShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	args.StakingSCConfig.MinStakeValue = "-1"
	expectedErr := fmt.Errorf("%w, value is %v", vm.ErrInvalidMinStakeValue, args.StakingSCConfig.MinStakeValue)

	d, err := NewDelegationSystemSC(args)
	assert.Nil(t, d)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationSystemSC_InvalidGenesisNodePriceShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	args.StakingSCConfig.GenesisNodePrice = "-1"
	expectedErr := fmt.Errorf("%w, value is %v", vm.ErrInvalidNodePrice, args.StakingSCConfig.GenesisNodePrice)

	d, err := NewDelegationSystemSC(args)
	assert.Nil(t, d)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationSystemSC_OkParamsShouldWork(t *testing.T) {
	t.Parallel()

	registerHandler := false
	args := createMockArgumentsForDelegation()
	args.DelegationSCConfig.MinStakeAmount = "10"

	epochNotifier := &mock.EpochNotifierStub{}
	epochNotifier.RegisterNotifyHandlerCalled = func(handler core.EpochSubscriberHandler) {
		registerHandler = true
	}
	args.EpochNotifier = epochNotifier

	d, err := NewDelegationSystemSC(args)
	assert.Nil(t, err)
	assert.True(t, registerHandler)
	assert.Equal(t, big.NewInt(10), d.minDelegationAmount)
}

func TestDelegationSystemSC_ExecuteNilArgsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei
	d, _ := NewDelegationSystemSC(args)

	output := d.Execute(nil)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "nil contract call input"))
}

func TestDelegationSystemSC_ExecuteDelegationDisabledShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei
	d, _ := NewDelegationSystemSC(args)
	d.delegationEnabled.Unset()
	vmInput := getDefaultVmInputForFunc("addNodes", [][]byte{})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "delegation manager contract is not enabled"))
}

func TestDelegationSystemSC_ExecuteInitScAlreadyPresentShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc(core.SCDeployInitFunctionName, [][]byte{})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "smart contract was already initialized"))
}

func TestDelegationSystemSC_ExecuteInitWrongNumOfArgs(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc(core.SCDeployInitFunctionName, [][]byte{[]byte("maxDelegationCap")})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments to init delegation contract"))
}

func TestDelegationSystemSC_ExecuteInitShouldWork(t *testing.T) {
	t.Parallel()

	ownerAddr := []byte("owner")
	maxDelegationCap := []byte{250}
	serviceFee := []byte{10}
	createdNonce := uint64(150)
	callValue := big.NewInt(130)
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{CurrentNonceCalled: func() uint64 {
			return createdNonce
		}},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei
	args.StakingSCConfig.UnBondPeriod = 20

	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc(core.SCDeployInitFunctionName, [][]byte{maxDelegationCap, serviceFee})
	vmInput.CallValue = callValue

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dConf, err := d.getDelegationContractConfig()
	assert.Nil(t, err)
	assert.Equal(t, ownerAddr, dConf.OwnerAddress)
	assert.Equal(t, big.NewInt(250), dConf.MaxDelegationCap)
	assert.Equal(t, big.NewInt(10).Uint64(), dConf.ServiceFee)
	assert.Equal(t, createdNonce, dConf.CreatedNonce)
	assert.Equal(t, big.NewInt(20).Uint64(), dConf.UnBondPeriod)
	assert.True(t, dConf.WithDelegationCap)

	dStatus, err := d.getDelegationStatus()
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), dStatus.NumDelegators)
	assert.Equal(t, 0, len(dStatus.StakedKeys))
	assert.Equal(t, 0, len(dStatus.NotStakedKeys))
	assert.Equal(t, 0, len(dStatus.UnStakedKeys))
	assert.Equal(t, 1, len(dStatus.Delegators))

	ownerFund, err := d.getFund(big.NewInt(0).Bytes())
	assert.Nil(t, err)
	assert.Equal(t, callValue, ownerFund.Value)
	assert.Equal(t, ownerAddr, ownerFund.Address)
	assert.Equal(t, createdNonce, ownerFund.Nonce)
	assert.Equal(t, active, ownerFund.Type)

	dGlobalFund, err := d.getGlobalFundData()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(dGlobalFund.ActiveFunds))
	assert.Equal(t, 0, len(dGlobalFund.UnStakedFunds))
	assert.Equal(t, 0, len(dGlobalFund.WithdrawOnlyFunds))
	assert.Equal(t, big.NewInt(0), dGlobalFund.TotalUnStakedFromNodes)
	assert.Equal(t, big.NewInt(0), dGlobalFund.TotalUnBondedFromNodes)
	assert.Equal(t, callValue, dGlobalFund.TotalActive)
	assert.Equal(t, big.NewInt(0), dGlobalFund.TotalUnStaked)
	assert.Equal(t, big.NewInt(0), dGlobalFund.TotalStaked)

	ok, delegator, err := d.getOrCreateDelegatorData(ownerAddr)
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, 0, len(delegator.UnStakedFunds))
	assert.Equal(t, 0, len(delegator.WithdrawOnlyFunds))
	assert.Equal(t, []byte{}, delegator.ActiveFund)
}

func TestDelegationSystemSC_ExecuteAddNodesUserErrors(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	signature := []byte("sig1")
	callValue := big.NewInt(130)
	vmInputArgs := make([][]byte, 0)
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	sigVerifier := &mock.MessageSignVerifierMock{}
	sigVerifier.VerifyCalled = func(message []byte, signedMessage []byte, pubKey []byte) error {
		return errors.New("verify error")
	}
	args.SigVerifier = sigVerifier

	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc("addNodes", vmInputArgs)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can change delegation config"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.Arguments = append(vmInputArgs, [][]byte{blsKey, blsKey}...)
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.Arguments = append(vmInputArgs, blsKey)
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "arguments must be of pair length - BLSKey and signedMessage"))

	vmInput.Arguments = append(vmInput.Arguments, signature)
	eei.gasRemaining = 10
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	eei.gasRemaining = 100
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidBLSKeys.Error()))
	assert.Equal(t, blsKey, eei.output[0])
	assert.Equal(t, []byte{invalidKey}, eei.output[1])
}

func TestDelegationSystemSC_ExecuteAddNodesStakedKeyAlreadyExistsInStakedKeysShouldErr(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	sig := []byte("sig1")
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	key := &NodesData{
		BLSKey: blsKey,
	}
	dStatus := &DelegationContractStatus{
		StakedKeys: []*NodesData{key},
	}
	_ = d.saveDelegationStatus(dStatus)

	vmInput := getDefaultVmInputForFunc("addNodes", [][]byte{blsKey, sig})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSC_ExecuteAddNodesStakedKeyAlreadyExistsInUnStakedKeysShouldErr(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	sig := []byte("sig1")
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	key := &NodesData{
		BLSKey: blsKey,
	}
	dStatus := &DelegationContractStatus{
		UnStakedKeys: []*NodesData{key},
	}
	_ = d.saveDelegationStatus(dStatus)

	vmInput := getDefaultVmInputForFunc("addNodes", [][]byte{blsKey, sig})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSC_ExecuteAddNodesStakedKeyAlreadyExistsInNotStakedKeysShouldErr(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	sig := []byte("sig1")
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	key := &NodesData{
		BLSKey: blsKey,
	}
	dStatus := &DelegationContractStatus{
		NotStakedKeys: []*NodesData{key},
	}
	_ = d.saveDelegationStatus(dStatus)

	vmInput := getDefaultVmInputForFunc("addNodes", [][]byte{blsKey, sig})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSC_ExecuteAddNodesShouldSaveAddedKeysAsNotStakedKeys(t *testing.T) {
	t.Parallel()

	blsKeys := [][]byte{[]byte("blsKey1"), []byte("blsKey2")}
	signatures := [][]byte{[]byte("sig1"), []byte("sig2")}
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei
	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc("addNodes", [][]byte{blsKeys[0], signatures[0], blsKeys[1], signatures[1]})
	_ = d.saveDelegationStatus(&DelegationContractStatus{})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	delegStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 0, len(delegStatus.StakedKeys))
	assert.Equal(t, 2, len(delegStatus.NotStakedKeys))
	assert.Equal(t, blsKeys[0], delegStatus.NotStakedKeys[0].BLSKey)
	assert.Equal(t, signatures[0], delegStatus.NotStakedKeys[0].SignedMsg)
	assert.Equal(t, blsKeys[1], delegStatus.NotStakedKeys[1].BLSKey)
	assert.Equal(t, signatures[1], delegStatus.NotStakedKeys[1].SignedMsg)
}

func TestDelegationSystemSC_ExecuteAddNodesWithNoArgsShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei
	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc("addNodes", [][]byte{})
	_ = d.saveDelegationStatus(&DelegationContractStatus{})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}

func TestDelegationSystemSC_ExecuteRemoveNodesUserErrors(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	callValue := big.NewInt(130)
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc("removeNodes", [][]byte{})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can change delegation config"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.Arguments = [][]byte{blsKey, blsKey}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.Arguments = [][]byte{blsKey}
	eei.gasRemaining = 10
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	eei.gasRemaining = 100
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationSystemSC_ExecuteRemoveNodesNotPresentInNotStakedShouldErr(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{})

	vmInput := getDefaultVmInputForFunc("removeNodes", [][]byte{blsKey})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSC_ExecuteRemoveNodesShouldRemoveKeyFromNotStakedKeys(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	key1 := &NodesData{BLSKey: blsKey1}
	key2 := &NodesData{BLSKey: blsKey2}
	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		NotStakedKeys: []*NodesData{key1, key2},
	})

	vmInput := getDefaultVmInputForFunc("removeNodes", [][]byte{blsKey1})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	delegStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 0, len(delegStatus.StakedKeys))
	assert.Equal(t, 1, len(delegStatus.NotStakedKeys))
	assert.Equal(t, blsKey2, delegStatus.NotStakedKeys[0].BLSKey)
}

func TestDelegationSystemSC_ExecuteStakeNodesUserErrors(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	callValue := big.NewInt(130)
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc("stakeNodes", [][]byte{})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can change delegation config"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.Arguments = [][]byte{blsKey, blsKey}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough arguments"))

	vmInput.Arguments = [][]byte{blsKey}
	eei.gasRemaining = 100
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationSystemSC_ExecuteStakeNodesNotPresentInNotStakedOrUnStakedShouldErr(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{})

	vmInput := getDefaultVmInputForFunc("stakeNodes", [][]byte{blsKey})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSC_ExecuteStakeNodesVerifiesBothUnStakedAndNotStaked(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	key1 := &NodesData{BLSKey: blsKey1}
	key2 := &NodesData{BLSKey: blsKey2}
	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		NotStakedKeys: []*NodesData{key1},
		UnStakedKeys:  []*NodesData{key2},
	})

	vmInput := getDefaultVmInputForFunc("stakeNodes", [][]byte{blsKey1, blsKey2})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w getGlobalFundData", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	globalFund := &GlobalFundData{
		TotalActive: big.NewInt(10),
	}
	_ = d.saveGlobalFundData(globalFund)
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough in total active to stake"))

	globalFund = &GlobalFundData{
		TotalActive: big.NewInt(200),
		TotalStaked: big.NewInt(0),
	}
	_ = d.saveGlobalFundData(globalFund)
	addAuctionAndStakingScToVmContext(eei)

	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	globalFund, _ = d.getGlobalFundData()
	assert.Equal(t, big.NewInt(200), globalFund.TotalStaked)
	assert.Equal(t, big.NewInt(0), globalFund.TotalActive)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 2, len(dStatus.StakedKeys))
	assert.Equal(t, 0, len(dStatus.UnStakedKeys))
	assert.Equal(t, 0, len(dStatus.NotStakedKeys))
}

func TestDelegationSystemSC_ExecuteUnStakeNodesUserErrors(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	callValue := big.NewInt(130)
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc("unStakeNodes", [][]byte{})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can change delegation config"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.Arguments = [][]byte{blsKey, blsKey}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough arguments"))

	vmInput.Arguments = [][]byte{blsKey}
	eei.gasRemaining = 100
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationSystemSC_ExecuteUnStakeNodesNotPresentInStakedShouldErr(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{})

	vmInput := getDefaultVmInputForFunc("unStakeNodes", [][]byte{blsKey})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSC_ExecuteUnStakeNodes(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	key1 := &NodesData{BLSKey: blsKey1}
	key2 := &NodesData{BLSKey: blsKey2}
	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		StakedKeys: []*NodesData{key1, key2},
	})
	_ = d.saveGlobalFundData(&GlobalFundData{TotalUnStakedFromNodes: big.NewInt(0), TotalActive: big.NewInt(100)})
	addAuctionAndStakingScToVmContext(eei)

	auctionMap := map[string][]byte{}
	registrationDataAuction := &AuctionDataV2{BlsPubKeys: [][]byte{blsKey1, blsKey2}, RewardAddress: []byte("rewardAddr")}
	regData, _ := d.marshalizer.Marshal(registrationDataAuction)
	auctionMap["addr"] = regData

	stakingMap := map[string][]byte{}
	registrationDataStaking := &StakedDataV2_0{RewardAddress: []byte("rewardAddr"), Staked: true}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking)
	stakingMap["blsKey1"] = regData

	registrationDataStaking2 := &StakedDataV2_0{RewardAddress: []byte("rewardAddr"), Staked: true}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking2)
	stakingMap["blsKey2"] = regData

	stakingNodesConfig := &StakingNodesConfig{StakedNodes: 5}
	stkNodes, _ := d.marshalizer.Marshal(stakingNodesConfig)
	stakingMap[nodesConfigKey] = stkNodes

	eei.storageUpdate[string(args.AuctionSCAddress)] = auctionMap
	eei.storageUpdate["staking"] = stakingMap

	vmInput := getDefaultVmInputForFunc("unStakeNodes", [][]byte{blsKey1, blsKey2})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 2, len(dStatus.UnStakedKeys))
	assert.Equal(t, 0, len(dStatus.StakedKeys))

	dGlobalFund, _ := d.getGlobalFundData()
	assert.Equal(t, big.NewInt(200), dGlobalFund.TotalUnStakedFromNodes)
}

func TestDelegationSystemSC_ExecuteUnBondNodesUserErrors(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	callValue := big.NewInt(130)
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc("unBondNodes", [][]byte{})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can change delegation config"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.Arguments = [][]byte{blsKey, blsKey}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough arguments"))

	vmInput.Arguments = [][]byte{blsKey}
	eei.gasRemaining = 100
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationSystemSC_ExecuteUnBondNodesNotPresentInUnStakedShouldErr(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{})

	vmInput := getDefaultVmInputForFunc("unBondNodes", [][]byte{blsKey})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSC_ExecuteUnBondNodes(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	key1 := &NodesData{BLSKey: blsKey1}
	key2 := &NodesData{BLSKey: blsKey2}
	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		UnStakedKeys: []*NodesData{key1, key2},
	})
	_ = d.saveGlobalFundData(&GlobalFundData{TotalUnBondedFromNodes: big.NewInt(0), TotalActive: big.NewInt(100)})
	addAuctionAndStakingScToVmContext(eei)

	auctionMap := map[string][]byte{}
	registrationDataAuction := &AuctionDataV2{
		BlsPubKeys:      [][]byte{blsKey1, blsKey2},
		RewardAddress:   []byte("rewardAddr"),
		LockedStake:     big.NewInt(300),
		TotalStakeValue: big.NewInt(500),
	}
	regData, _ := d.marshalizer.Marshal(registrationDataAuction)
	auctionMap["addr"] = regData

	stakingMap := map[string][]byte{}
	registrationDataStaking := &StakedDataV2_0{RewardAddress: []byte("rewardAddr")}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking)
	stakingMap["blsKey1"] = regData

	registrationDataStaking2 := &StakedDataV2_0{RewardAddress: []byte("rewardAddr")}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking2)
	stakingMap["blsKey2"] = regData

	stakingNodesConfig := &StakingNodesConfig{StakedNodes: 5}
	stkNodes, _ := d.marshalizer.Marshal(stakingNodesConfig)
	stakingMap[nodesConfigKey] = stkNodes

	eei.storageUpdate[string(args.AuctionSCAddress)] = auctionMap
	eei.storageUpdate["staking"] = stakingMap

	vmInput := getDefaultVmInputForFunc("unBondNodes", [][]byte{blsKey1, blsKey2})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 2, len(dStatus.NotStakedKeys))
	assert.Equal(t, 0, len(dStatus.UnStakedKeys))

	dGlobalFund, _ := d.getGlobalFundData()
	assert.Equal(t, big.NewInt(200), dGlobalFund.TotalUnBondedFromNodes)
}
