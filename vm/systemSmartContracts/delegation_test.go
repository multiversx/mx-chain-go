package systemSmartContracts

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/parsers"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgumentsForDelegation() ArgsNewDelegation {
	return ArgsNewDelegation{
		DelegationSCConfig: config.DelegationSystemSCConfig{
			MinServiceFee: 10,
			MaxServiceFee: 200,
		},
		StakingSCConfig: config.StakingSystemSCConfig{
			MinStakeValue:    "10",
			UnJailValue:      "15",
			GenesisNodePrice: "100",
		},
		Eei:                    &mock.SystemEIStub{},
		SigVerifier:            &mock.MessageSignVerifierMock{},
		DelegationMgrSCAddress: vm.DelegationManagerSCAddress,
		StakingSCAddress:       vm.StakingSCAddress,
		ValidatorSCAddress:     vm.ValidatorSCAddress,
		GasCost:                vm.GasCost{MetaChainSystemSCsCost: vm.MetaChainSystemSCsCost{ESDTIssue: 10}},
		Marshalizer:            &mock.MarshalizerMock{},
		EpochNotifier:          &mock.EpochNotifierStub{},
		EndOfEpochAddress:      vm.EndOfEpochAddress,
	}
}

func addValidatorAndStakingScToVmContext(eei *vmContext) {
	validatorArgs := createMockArgumentsForValidatorSC()
	validatorArgs.Eei = eei
	validatorArgs.StakingSCConfig.GenesisNodePrice = "100"
	validatorArgs.StakingSCAddress = vm.StakingSCAddress
	validatorSc, _ := NewValidatorSmartContract(validatorArgs)

	stakingArgs := createMockStakingScArguments()
	stakingArgs.Eei = eei
	stakingSc, _ := NewStakingSmartContract(stakingArgs)

	eei.inputParser = parsers.NewCallArgsParser()

	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		if bytes.Equal(key, vm.StakingSCAddress) {
			return stakingSc, nil
		}

		if bytes.Equal(key, vm.ValidatorSCAddress) {
			validatorSc.flagEnableTopUp.Set()
			_ = validatorSc.saveRegistrationData([]byte("addr"), &ValidatorDataV2{
				RewardAddress:   []byte("rewardAddr"),
				TotalStakeValue: big.NewInt(1000),
				LockedStake:     big.NewInt(500),
				BlsPubKeys:      [][]byte{[]byte("blsKey1"), []byte("blsKey2")},
				TotalUnstaked:   big.NewInt(150),
				UnstakedInfo: []*UnstakedValue{
					{
						UnstakedNonce: 10,
						UnstakedValue: big.NewInt(60),
					},
					{
						UnstakedNonce: 50,
						UnstakedValue: big.NewInt(80),
					},
				},
				NumRegistered: 2,
			})
			validatorSc.unBondPeriod = 50
			return validatorSc, nil
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

func createDelegationManagerConfig(eei *vmContext, marshalizer marshal.Marshalizer, minDelegationAmount *big.Int) {
	cfg := &DelegationManagement{
		MinDelegationAmount: minDelegationAmount,
	}

	marshaledData, _ := marshalizer.Marshal(cfg)
	eei.SetStorageForAddress(vm.DelegationManagerSCAddress, []byte(delegationManagementKey), marshaledData)
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

func TestNewDelegationSystemSC_InvalidValidatorSCAddrShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("%w for validator sc address", vm.ErrInvalidAddress)
	args := createMockArgumentsForDelegation()
	args.ValidatorSCAddress = []byte{}

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

	epochNotifier := &mock.EpochNotifierStub{}
	epochNotifier.RegisterNotifyHandlerCalled = func(handler core.EpochSubscriberHandler) {
		registerHandler = true
	}
	args.EpochNotifier = epochNotifier

	d, err := NewDelegationSystemSC(args)
	assert.Nil(t, err)
	assert.True(t, registerHandler)
	assert.False(t, check.IfNil(d))
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
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInputArgsIsNil.Error()))
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
	assert.True(t, strings.Contains(eei.returnMessage, "delegation contract is not enabled"))
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

func TestDelegationSystemSC_ExecuteInitCallValueHigherThanMaxDelegationCapShouldErr(t *testing.T) {
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
	vmInput := getDefaultVmInputForFunc(core.SCDeployInitFunctionName, [][]byte{big.NewInt(250).Bytes(), big.NewInt(10).Bytes()})
	vmInput.CallerAddr = []byte("ownerAddress")
	vmInput.CallValue = big.NewInt(300)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "total delegation cap reached"))
}

func TestDelegationSystemSC_ExecuteInitShouldWork(t *testing.T) {
	t.Parallel()

	ownerAddr := []byte("ownerAddr")
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
		parsers.NewCallArgsParser(),
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei
	args.StakingSCConfig.UnBondPeriod = 20
	_ = eei.SetSystemSCContainer(
		createSystemSCContainer(eei),
	)

	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc(core.SCDeployInitFunctionName, [][]byte{maxDelegationCap, serviceFee})
	vmInput.CallValue = callValue
	vmInput.RecipientAddr = createNewAddress(vm.FirstDelegationSCAddress)
	vmInput.CallerAddr = ownerAddr
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	retrievedOwnerAddress := d.eei.GetStorage([]byte(ownerKey))
	retrievedServiceFee := d.eei.GetStorage([]byte(serviceFeeKey))

	dConf, err := d.getDelegationContractConfig()
	assert.Nil(t, err)
	assert.Equal(t, ownerAddr, retrievedOwnerAddress)
	assert.Equal(t, big.NewInt(250), dConf.MaxDelegationCap)
	assert.Equal(t, []byte{10}, retrievedServiceFee)
	assert.Equal(t, createdNonce, dConf.CreatedNonce)
	assert.Equal(t, big.NewInt(20).Uint64(), dConf.UnBondPeriod)

	dStatus, err := d.getDelegationStatus()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(dStatus.StakedKeys))
	assert.Equal(t, 0, len(dStatus.NotStakedKeys))
	assert.Equal(t, 0, len(dStatus.UnStakedKeys))
	assert.Equal(t, uint64(1), dStatus.NumUsers)

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	ownerFund, err := d.getFund(fundKey)
	assert.Nil(t, err)
	assert.Equal(t, callValue, ownerFund.Value)
	assert.Equal(t, ownerAddr, ownerFund.Address)
	assert.Equal(t, createdNonce, ownerFund.Nonce)
	assert.Equal(t, active, ownerFund.Type)

	dGlobalFund, err := d.getGlobalFundData()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(dGlobalFund.ActiveFunds))
	assert.Equal(t, 0, len(dGlobalFund.UnStakedFunds))
	assert.Equal(t, callValue, dGlobalFund.TotalActive)
	assert.Equal(t, big.NewInt(0), dGlobalFund.TotalUnStaked)

	delegatorDataPresent, delegator, err := d.getOrCreateDelegatorData(ownerAddr)
	assert.Nil(t, err)
	assert.False(t, delegatorDataPresent)
	assert.Equal(t, 0, len(delegator.UnStakedFunds))
	assert.Equal(t, fundKey, delegator.ActiveFund)
}

func TestDelegationSystemSC_ExecuteAddNodesUserErrors(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	blsKey3 := []byte("blsKey3")
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
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can call this method"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.Arguments = append(vmInputArgs, [][]byte{blsKey1, blsKey1}...)
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.Arguments = append(vmInputArgs, [][]byte{blsKey1, blsKey2, blsKey3}...)
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "arguments must be of pair length - BLSKey and signedMessage"))

	vmInput.Arguments = append(vmInputArgs, [][]byte{blsKey1, signature}...)
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
	assert.Equal(t, blsKey1, eei.output[0])
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

func TestDelegationSystemSC_ExecuteAddNodesWithNoArgsShouldErr(t *testing.T) {
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
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough arguments"))
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
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can call this method"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

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

func TestDelegationSystemSC_ExecuteRemoveNodesWithNoArgsShouldErr(t *testing.T) {
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
	vmInput := getDefaultVmInputForFunc("removeNodes", [][]byte{})
	_ = d.saveDelegationStatus(&DelegationContractStatus{})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough arguments"))
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
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can call this method"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

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
	}
	_ = d.saveGlobalFundData(globalFund)
	addValidatorAndStakingScToVmContext(eei)

	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	globalFund, _ = d.getGlobalFundData()
	assert.Equal(t, big.NewInt(200), globalFund.TotalActive)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 2, len(dStatus.StakedKeys))
	assert.Equal(t, 0, len(dStatus.UnStakedKeys))
	assert.Equal(t, 0, len(dStatus.NotStakedKeys))
}

func TestDelegationSystemSC_ExecuteDelegateStakeNodes(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	args := createMockArgumentsForDelegation()
	args.GasCost.MetaChainSystemSCsCost.Stake = 1
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)

	createDelegationManagerConfig(eei, args.Marshalizer, big.NewInt(10))
	eei.SetSCAddress(vm.FirstDelegationSCAddress)
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
	globalFund := &GlobalFundData{
		TotalActive: big.NewInt(0),
	}
	_ = d.saveGlobalFundData(globalFund)
	_ = d.saveDelegationContractConfig(&DelegationConfig{
		MaxDelegationCap:    big.NewInt(1000),
		InitialOwnerFunds:   big.NewInt(100),
		AutomaticActivation: true,
	})
	addValidatorAndStakingScToVmContext(eei)

	vmInput := getDefaultVmInputForFunc("delegate", [][]byte{})
	vmInput.CallerAddr = []byte("delegator")
	vmInput.CallValue = big.NewInt(500)
	vmInput.GasProvided = 10000
	eei.gasRemaining = vmInput.GasProvided

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	globalFund, _ = d.getGlobalFundData()
	assert.Equal(t, big.NewInt(500), globalFund.TotalActive)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 2, len(dStatus.StakedKeys))
	assert.Equal(t, 0, len(dStatus.UnStakedKeys))
	assert.Equal(t, 0, len(dStatus.NotStakedKeys))

	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, 6, len(vmOutput.OutputAccounts))
	assert.Equal(t, 2, len(vmOutput.OutputAccounts[string(vm.StakingSCAddress)].OutputTransfers))

	output = d.Execute(vmInput)
	eei.gasRemaining = vmInput.GasProvided
	assert.Equal(t, vmcommon.Ok, output)

	globalFund, _ = d.getGlobalFundData()
	assert.Equal(t, big.NewInt(1000), globalFund.TotalActive)

	_, delegator, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
	fund, _ := d.getFund(delegator.ActiveFund)
	assert.Equal(t, fund.Value, big.NewInt(1000))
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
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can call this method"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

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
	_ = d.saveGlobalFundData(&GlobalFundData{TotalActive: big.NewInt(100)})
	addValidatorAndStakingScToVmContext(eei)

	validatorMap := map[string][]byte{}
	registrationDataValidator := &ValidatorDataV2{BlsPubKeys: [][]byte{blsKey1, blsKey2}, RewardAddress: []byte("rewardAddr")}
	regData, _ := d.marshalizer.Marshal(registrationDataValidator)
	validatorMap["addr"] = regData

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

	eei.storageUpdate[string(args.ValidatorSCAddress)] = validatorMap
	eei.storageUpdate[string(args.StakingSCAddress)] = stakingMap

	vmInput := getDefaultVmInputForFunc("unStakeNodes", [][]byte{blsKey1, blsKey2})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 2, len(dStatus.UnStakedKeys))
	assert.Equal(t, 0, len(dStatus.StakedKeys))
}

func TestDelegationSystemSC_ExecuteUnStakeNodesAtEndOfEpoch(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		blockChainHook,
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
	_ = d.saveGlobalFundData(&GlobalFundData{TotalActive: big.NewInt(100)})
	validatorArgs := createMockArgumentsForValidatorSC()
	validatorArgs.Eei = eei
	validatorArgs.StakingSCConfig.GenesisNodePrice = "100"
	validatorArgs.StakingSCConfig.StakingV2Epoch = 0
	validatorArgs.StakingSCAddress = vm.StakingSCAddress
	validatorSc, _ := NewValidatorSmartContract(validatorArgs)

	stakingArgs := createMockStakingScArguments()
	stakingArgs.Eei = eei
	stakingSc, _ := NewStakingSmartContract(stakingArgs)

	eei.inputParser = parsers.NewCallArgsParser()
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		if bytes.Equal(key, vm.StakingSCAddress) {
			return stakingSc, nil
		}

		if bytes.Equal(key, vm.ValidatorSCAddress) {
			return validatorSc, nil
		}

		return nil, vm.ErrUnknownSystemSmartContract
	}})

	validatorMap := map[string][]byte{}
	registrationDataValidator := &ValidatorDataV2{
		BlsPubKeys:      [][]byte{blsKey1, blsKey2},
		RewardAddress:   []byte("rewardAddr"),
		TotalStakeValue: big.NewInt(1000000),
		NumRegistered:   2,
	}
	regData, _ := d.marshalizer.Marshal(registrationDataValidator)
	validatorMap["addr"] = regData

	stakingMap := map[string][]byte{}
	registrationDataStaking := &StakedDataV2_0{RewardAddress: []byte("rewardAddr"), Staked: false, UnStakedNonce: 5, StakeValue: big.NewInt(0)}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking)
	stakingMap["blsKey1"] = regData

	registrationDataStaking2 := &StakedDataV2_0{RewardAddress: []byte("rewardAddr"), Staked: false, UnStakedNonce: 5, StakeValue: big.NewInt(0)}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking2)
	stakingMap["blsKey2"] = regData

	stakingNodesConfig := &StakingNodesConfig{StakedNodes: 5}
	stkNodes, _ := d.marshalizer.Marshal(stakingNodesConfig)
	stakingMap[nodesConfigKey] = stkNodes

	eei.storageUpdate[string(args.ValidatorSCAddress)] = validatorMap
	eei.storageUpdate[string(args.StakingSCAddress)] = stakingMap

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 10
	}

	vmInput := getDefaultVmInputForFunc("unStakeAtEndOfEpoch", [][]byte{blsKey1, blsKey2})
	vmInput.CallerAddr = args.EndOfEpochAddress
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 2, len(dStatus.UnStakedKeys))
	assert.Equal(t, 0, len(dStatus.StakedKeys))

	vmInput = getDefaultVmInputForFunc("reStakeUnStakedNodes", [][]byte{blsKey1, blsKey2})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dStatus, _ = d.getDelegationStatus()
	assert.Equal(t, 0, len(dStatus.UnStakedKeys))
	assert.Equal(t, 2, len(dStatus.StakedKeys))
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
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can call this method"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

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
	_ = d.saveGlobalFundData(&GlobalFundData{TotalActive: big.NewInt(100)})
	addValidatorAndStakingScToVmContext(eei)

	registrationDataValidator := &ValidatorDataV2{
		BlsPubKeys:      [][]byte{blsKey1, blsKey2},
		RewardAddress:   []byte("rewardAddr"),
		LockedStake:     big.NewInt(300),
		TotalStakeValue: big.NewInt(1000),
		NumRegistered:   2,
	}
	regData, _ := d.marshalizer.Marshal(registrationDataValidator)
	eei.SetStorageForAddress(vm.ValidatorSCAddress, []byte("addr"), regData)

	registrationDataStaking := &StakedDataV2_0{RewardAddress: []byte("rewardAddr")}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking)
	eei.SetStorageForAddress(vm.StakingSCAddress, []byte("blsKey1"), regData)

	registrationDataStaking2 := &StakedDataV2_0{RewardAddress: []byte("rewardAddr")}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking2)
	eei.SetStorageForAddress(vm.StakingSCAddress, []byte("blsKey2"), regData)

	stakingNodesConfig := &StakingNodesConfig{StakedNodes: 5}
	stkNodes, _ := d.marshalizer.Marshal(stakingNodesConfig)
	eei.SetStorageForAddress(vm.StakingSCAddress, []byte(nodesConfigKey), stkNodes)

	vmInput := getDefaultVmInputForFunc("unBondNodes", [][]byte{blsKey1, blsKey2})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 2, len(dStatus.NotStakedKeys))
	assert.Equal(t, 0, len(dStatus.UnStakedKeys))
}

func TestDelegationSystemSC_ExecuteUnJailNodesUserErrors(t *testing.T) {
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
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc("unJailNodes", [][]byte{})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough arguments"))

	vmInput.Arguments = append(vmInput.Arguments, blsKey)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	_ = d.saveDelegationStatus(&DelegationContractStatus{})
	vmInput.Arguments = [][]byte{blsKey, blsKey}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))
}

func TestDelegationSystemSC_ExecuteUnJailNodesNotPresentInStakedOrUnStakedShouldErr(t *testing.T) {
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

	vmInput := getDefaultVmInputForFunc("unJailNodes", [][]byte{blsKey})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSC_ExecuteUnJailNodesNotDelegatorShouldErr(t *testing.T) {
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
	vmInput := getDefaultVmInputForFunc("unJailNodes", [][]byte{blsKey1, blsKey2})
	vmInput.CallerAddr = []byte("notDelegator")
	key1 := &NodesData{BLSKey: blsKey1}
	key2 := &NodesData{BLSKey: blsKey2}
	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		StakedKeys:   []*NodesData{key1},
		UnStakedKeys: []*NodesData{key2},
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not a delegator"))
}

func TestDelegationSystemSC_ExecuteUnJailNodes(t *testing.T) {
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

	vmInput := getDefaultVmInputForFunc("unJailNodes", [][]byte{blsKey1, blsKey2})
	vmInput.CallValue = big.NewInt(20)
	vmInput.CallerAddr = []byte("delegator")

	key1 := &NodesData{BLSKey: blsKey1}
	key2 := &NodesData{BLSKey: blsKey2}
	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		StakedKeys:   []*NodesData{key1},
		UnStakedKeys: []*NodesData{key2},
		NumUsers:     1,
	})
	addValidatorAndStakingScToVmContext(eei)

	_ = d.saveDelegatorData(vmInput.CallerAddr, &DelegatorData{ActiveFund: []byte("someFund"), UnClaimedRewards: big.NewInt(0), TotalCumulatedRewards: big.NewInt(0)})

	validatorMap := map[string][]byte{}
	registrationDataValidator := &ValidatorDataV2{
		BlsPubKeys:    [][]byte{blsKey1, blsKey2},
		RewardAddress: []byte("rewardAddr"),
	}
	regData, _ := d.marshalizer.Marshal(registrationDataValidator)
	validatorMap["addr"] = regData

	stakingMap := map[string][]byte{}
	registrationDataStaking := &StakedDataV2_0{
		RewardAddress: []byte("rewardAddr"),
		Jailed:        true,
	}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking)
	stakingMap["blsKey1"] = regData

	registrationDataStaking2 := &StakedDataV2_0{
		RewardAddress: []byte("rewardAddr"),
		Jailed:        true,
	}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking2)
	stakingMap["blsKey2"] = regData

	eei.storageUpdate[string(args.ValidatorSCAddress)] = validatorMap
	eei.storageUpdate[string(args.StakingSCAddress)] = stakingMap

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}

func TestDelegationSystemSC_ExecuteDelegateUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei
	createDelegationManagerConfig(eei, args.Marshalizer, big.NewInt(10))

	vmInput := getDefaultVmInputForFunc("delegate", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "delegate value must be higher than minDelegationAmount"))

	vmInput.CallValue = big.NewInt(15)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))
}

func TestDelegationSystemSC_ExecuteDelegateWrongInit(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei
	createDelegationManagerConfig(eei, args.Marshalizer, big.NewInt(10))

	vmInput := getDefaultVmInputForFunc("delegate", [][]byte{})
	vmInput.CallValue = big.NewInt(15)
	d, _ := NewDelegationSystemSC(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	_ = d.saveDelegationStatus(&DelegationContractStatus{})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr = fmt.Errorf("%w delegation contract config", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	_ = d.saveDelegationContractConfig(&DelegationConfig{})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr = fmt.Errorf("%w getGlobalFundData", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationSystemSC_ExecuteDelegate(t *testing.T) {
	t.Parallel()

	delegator1 := []byte("delegator1")
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei
	addValidatorAndStakingScToVmContext(eei)
	createDelegationManagerConfig(eei, args.Marshalizer, big.NewInt(10))

	vmInput := getDefaultVmInputForFunc("delegate", [][]byte{})
	vmInput.CallValue = big.NewInt(15)
	vmInput.CallerAddr = delegator1
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegationStatus(&DelegationContractStatus{})
	_ = d.saveDelegationContractConfig(&DelegationConfig{
		MaxDelegationCap:  big.NewInt(100),
		InitialOwnerFunds: big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive: big.NewInt(100),
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "total delegation cap reached"))

	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive: big.NewInt(0),
	})

	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	dFund, _ := d.getFund(fundKey)
	assert.Equal(t, big.NewInt(15), dFund.Value)
	assert.Equal(t, delegator1, dFund.Address)
	assert.Equal(t, active, dFund.Type)

	dGlobalFund, _ := d.getGlobalFundData()
	assert.Equal(t, big.NewInt(15), dGlobalFund.TotalActive)
	assert.Equal(t, fundKey, dGlobalFund.ActiveFunds[0])

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, uint64(1), dStatus.NumUsers)

	_, dData, _ := d.getOrCreateDelegatorData(delegator1)
	assert.Equal(t, fundKey, dData.ActiveFund)
}

func TestDelegationSystemSC_ExecuteDelegateFailsWhenGettingDelegationManagement(t *testing.T) {
	t.Parallel()

	delegator1 := []byte("delegator1")
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei
	addValidatorAndStakingScToVmContext(eei)

	vmInput := getDefaultVmInputForFunc("delegate", [][]byte{})
	vmInput.CallValue = big.NewInt(15)
	vmInput.CallerAddr = delegator1
	d, _ := NewDelegationSystemSC(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "error getting minimum delegation amount data was not found under requested key"))
}

func TestDelegationSystemSC_ExecuteUnDelegateUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei
	createDelegationManagerConfig(eei, args.Marshalizer, big.NewInt(10))

	vmInput := getDefaultVmInputForFunc("unDelegate", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "wrong number of arguments"))
}

func TestDelegationSystemSC_ExecuteUnDelegateUserErrorsWhenAnInvalidValueToDelegateWasProvided(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	negativeValueToUndelegate := big.NewInt(0)
	vmInput := getDefaultVmInputForFunc("unDelegate", [][]byte{negativeValueToUndelegate.Bytes()})

	d, _ := NewDelegationSystemSC(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid value to undelegate"))
}

func TestDelegationSystemSC_ExecuteUnDelegateUserErrorsWhenGettingMinimumDelegationAmount(t *testing.T) {
	t.Parallel()

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("unDelegate", [][]byte{{80}})
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegatorData(vmInput.CallerAddr, &DelegatorData{
		ActiveFund:            fundKey,
		UnStakedFunds:         [][]byte{},
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})
	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		UnStakedFunds: [][]byte{},
		TotalActive:   big.NewInt(100),
		TotalUnStaked: big.NewInt(0),
	})
	d.eei.SetStorage([]byte(lastFundKey), fundKey)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "error getting minimum delegation amount"))
}

func TestDelegationSystemSC_ExecuteUnDelegateUserNotDelegatorOrNoActiveFundShouldErr(t *testing.T) {
	t.Parallel()

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei
	createDelegationManagerConfig(eei, args.Marshalizer, big.NewInt(10))

	vmInput := getDefaultVmInputForFunc("unDelegate", [][]byte{{100}})
	d, _ := NewDelegationSystemSC(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "caller is not a delegator"))

	_ = d.saveDelegatorData(vmInput.CallerAddr, &DelegatorData{
		ActiveFund: fundKey,
	})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w getFund %s", vm.ErrDataNotFoundUnderKey, string(fundKey))
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(50),
	})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid value to undelegate"))

	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(11),
	})
	vmInput.Arguments = [][]byte{{10}}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid value to undelegate - need to undelegate all - do not leave dust behind"))
}

func TestDelegationSystemSC_ExecuteUnDelegatePartOfFunds(t *testing.T) {
	t.Parallel()

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	nextFundKey := append([]byte(fundKeyPrefix), []byte{2}...)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		blockChainHook,
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei
	addValidatorAndStakingScToVmContext(eei)
	createDelegationManagerConfig(eei, args.Marshalizer, big.NewInt(10))

	vmInput := getDefaultVmInputForFunc("unDelegate", [][]byte{{80}})
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegatorData(vmInput.CallerAddr, &DelegatorData{
		ActiveFund:            fundKey,
		UnStakedFunds:         [][]byte{},
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})
	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		UnStakedFunds: [][]byte{},
		TotalActive:   big.NewInt(100),
		TotalUnStaked: big.NewInt(0),
	})
	d.eei.SetStorage([]byte(lastFundKey), fundKey)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dFund, _ := d.getFund(fundKey)
	assert.Equal(t, big.NewInt(20), dFund.Value)
	assert.Equal(t, active, dFund.Type)

	dFund, _ = d.getFund(nextFundKey)
	assert.Equal(t, big.NewInt(80), dFund.Value)
	assert.Equal(t, unStaked, dFund.Type)
	assert.Equal(t, vmInput.CallerAddr, dFund.Address)

	globalFund, _ := d.getGlobalFundData()
	assert.Equal(t, 1, len(globalFund.UnStakedFunds))
	assert.Equal(t, nextFundKey, globalFund.UnStakedFunds[0])
	assert.Equal(t, big.NewInt(20), globalFund.TotalActive)
	assert.Equal(t, big.NewInt(80), globalFund.TotalUnStaked)

	_, dData, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.Equal(t, 1, len(dData.UnStakedFunds))
	assert.Equal(t, nextFundKey, dData.UnStakedFunds[0])

	_ = d.saveDelegationContractConfig(&DelegationConfig{
		UnBondPeriod: 50,
	})

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 100
	}

	vmInput.Arguments = [][]byte{{20}}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	eei.output = make([][]byte, 0)
	vmInput = getDefaultVmInputForFunc("getUserUnDelegatedList", [][]byte{})
	vmInput.Arguments = [][]byte{vmInput.CallerAddr}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	assert.Equal(t, 4, len(eei.output))
	assert.Equal(t, eei.output[0], []byte{80})
	assert.Equal(t, eei.output[1], []byte{})
	assert.Equal(t, eei.output[2], []byte{20})
	assert.Equal(t, eei.output[3], []byte{50})
}

func TestDelegationSystemSC_ExecuteUnDelegateAllFunds(t *testing.T) {
	t.Parallel()

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	nextFundKey := append([]byte(fundKeyPrefix), []byte{2}...)
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei
	addValidatorAndStakingScToVmContext(eei)
	createDelegationManagerConfig(eei, args.Marshalizer, big.NewInt(10))

	vmInput := getDefaultVmInputForFunc("unDelegate", [][]byte{{100}})
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegatorData(vmInput.CallerAddr, &DelegatorData{
		ActiveFund:            fundKey,
		UnStakedFunds:         [][]byte{},
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})
	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		UnStakedFunds: [][]byte{},
		TotalActive:   big.NewInt(100),
		TotalUnStaked: big.NewInt(0),
		ActiveFunds:   [][]byte{fundKey},
	})
	d.eei.SetStorage([]byte(lastFundKey), fundKey)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dFund, _ := d.getFund(fundKey)
	assert.Nil(t, dFund)

	dFund, _ = d.getFund(nextFundKey)
	assert.Equal(t, big.NewInt(100), dFund.Value)
	assert.Equal(t, unStaked, dFund.Type)
	assert.Equal(t, vmInput.CallerAddr, dFund.Address)

	globalFund, _ := d.getGlobalFundData()
	assert.Equal(t, 1, len(globalFund.UnStakedFunds))
	assert.Equal(t, nextFundKey, globalFund.UnStakedFunds[0])
	assert.Equal(t, big.NewInt(0), globalFund.TotalActive)
	assert.Equal(t, big.NewInt(100), globalFund.TotalUnStaked)
	assert.Equal(t, 0, len(globalFund.ActiveFunds))

	_, dData, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.Equal(t, 1, len(dData.UnStakedFunds))
	assert.Equal(t, nextFundKey, dData.UnStakedFunds[0])
}

func TestDelegationSystemSC_ExecuteUnDelegateAllFundsAsOwner(t *testing.T) {
	t.Parallel()

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	nextFundKey := append([]byte(fundKeyPrefix), []byte{2}...)
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei
	addValidatorAndStakingScToVmContext(eei)
	minDelegationAmount := big.NewInt(10)
	createDelegationManagerConfig(eei, args.Marshalizer, minDelegationAmount)

	vmInput := getDefaultVmInputForFunc("unDelegate", [][]byte{{100}})
	d, _ := NewDelegationSystemSC(args)

	vmInput.CallerAddr = []byte("ownerAsDelegator")
	d.eei.SetStorage([]byte(ownerKey), vmInput.CallerAddr)
	_ = d.saveDelegationContractConfig(&DelegationConfig{InitialOwnerFunds: big.NewInt(100), MaxDelegationCap: big.NewInt(0)})
	_ = d.saveDelegatorData(vmInput.CallerAddr, &DelegatorData{
		ActiveFund:            fundKey,
		UnStakedFunds:         [][]byte{},
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})
	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		UnStakedFunds: [][]byte{},
		TotalActive:   big.NewInt(100),
		TotalUnStaked: big.NewInt(0),
		ActiveFunds:   [][]byte{fundKey},
	})
	d.eei.SetStorage([]byte(lastFundKey), fundKey)

	_ = d.saveDelegationStatus(&DelegationContractStatus{StakedKeys: []*NodesData{{BLSKey: []byte("blsKey"), SignedMsg: []byte("someMsg")}}})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	_ = d.saveDelegationStatus(&DelegationContractStatus{})
	vmInput.Arguments = [][]byte{{50}}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	vmInput.Arguments = [][]byte{{100}}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dFund, _ := d.getFund(fundKey)
	assert.Nil(t, dFund)

	dFund, _ = d.getFund(nextFundKey)
	assert.Equal(t, big.NewInt(100), dFund.Value)
	assert.Equal(t, unStaked, dFund.Type)
	assert.Equal(t, vmInput.CallerAddr, dFund.Address)

	globalFund, _ := d.getGlobalFundData()
	assert.Equal(t, 1, len(globalFund.UnStakedFunds))
	assert.Equal(t, nextFundKey, globalFund.UnStakedFunds[0])
	assert.Equal(t, big.NewInt(0), globalFund.TotalActive)
	assert.Equal(t, big.NewInt(100), globalFund.TotalUnStaked)
	assert.Equal(t, 0, len(globalFund.ActiveFunds))

	_, dData, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.Equal(t, 1, len(dData.UnStakedFunds))
	assert.Equal(t, nextFundKey, dData.UnStakedFunds[0])

	managementData := &DelegationManagement{
		MinDeposit:          big.NewInt(10),
		MinDelegationAmount: minDelegationAmount,
	}
	marshaledData, _ := d.marshalizer.Marshal(managementData)
	eei.SetStorageForAddress(d.delegationMgrSCAddress, []byte(delegationManagementKey), marshaledData)

	vmInput.Function = "delegate"
	vmInput.Arguments = [][]byte{}
	vmInput.CallValue = big.NewInt(1000)
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}

func TestDelegationSystemSC_ExecuteWithdrawUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("withdraw", [][]byte{[]byte("wrong arg")})
	d, _ := NewDelegationSystemSC(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "wrong number of arguments"))

	vmInput.Arguments = [][]byte{}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "caller is not a delegator"))
}

func TestDelegationSystemSC_ExecuteWithdrawWrongInit(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("withdraw", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegatorData(vmInput.CallerAddr, &DelegatorData{})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation contract config", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	_ = d.saveDelegationContractConfig(&DelegationConfig{})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr = fmt.Errorf("%w getGlobalFundData", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationSystemSC_ExecuteWithdraw(t *testing.T) {
	t.Parallel()

	fundKey1 := []byte{1}
	fundKey2 := []byte{2}
	currentNonce := uint64(60)
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{CurrentNonceCalled: func() uint64 {
			return currentNonce
		}},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei
	addValidatorAndStakingScToVmContext(eei)

	vmInput := getDefaultVmInputForFunc("withdraw", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegatorData(vmInput.CallerAddr, &DelegatorData{
		UnStakedFunds:         [][]byte{fundKey1, fundKey2},
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})
	_ = d.saveFund(fundKey1, &Fund{
		Value:   big.NewInt(60),
		Address: vmInput.CallerAddr,
		Nonce:   10,
		Type:    unStaked,
	})
	_ = d.saveFund(fundKey2, &Fund{
		Value:   big.NewInt(80),
		Address: vmInput.CallerAddr,
		Nonce:   50,
		Type:    unStaked,
	})
	_ = d.saveDelegationContractConfig(&DelegationConfig{
		UnBondPeriod: 50,
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		UnStakedFunds: [][]byte{fundKey1, fundKey2},
		TotalUnStaked: big.NewInt(140),
		TotalActive:   big.NewInt(0),
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	gFundData, _ := d.getGlobalFundData()
	assert.Equal(t, 1, len(gFundData.UnStakedFunds))
	assert.Equal(t, fundKey2, gFundData.UnStakedFunds[0])
	assert.Equal(t, big.NewInt(80), gFundData.TotalUnStaked)

	_, dData, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.Equal(t, 1, len(dData.UnStakedFunds))
	assert.Equal(t, fundKey2, dData.UnStakedFunds[0])

	fundKey, _ := d.getFund(fundKey1)
	assert.Nil(t, fundKey)

	_ = d.saveDelegationStatus(&DelegationContractStatus{NumUsers: 2})
	currentNonce = 150
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	isNew, _, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.True(t, isNew)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, uint64(1), dStatus.NumUsers)
}

func TestDelegationSystemSC_ExecuteChangeServiceFeeUserErrors(t *testing.T) {
	t.Parallel()

	newServiceFee := []byte{50}
	callValue := big.NewInt(15)
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("changeServiceFee", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can call this method"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.Arguments = [][]byte{newServiceFee, newServiceFee}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.Arguments = [][]byte{newServiceFee, []byte("wrong arg")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments"))

	vmInput.Arguments = [][]byte{big.NewInt(5).Bytes()}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "new service fee out of bounds"))

	vmInput.Arguments = [][]byte{[]byte("210")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "new service fee out of bounds"))
}

func TestDelegationSystemSC_ExecuteChangeServiceFee(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("changeServiceFee", [][]byte{big.NewInt(70).Bytes()})
	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationContractConfig(&DelegationConfig{})
	_ = d.saveGlobalFundData(&GlobalFundData{TotalActive: big.NewInt(0)})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	retrievedServiceFee := d.eei.GetStorage([]byte(serviceFeeKey))
	assert.Equal(t, []byte{70}, retrievedServiceFee)
}

func TestDelegationSystemSC_ExecuteModifyTotalDelegationCapUserErrors(t *testing.T) {
	t.Parallel()

	newServiceFee := []byte{50}
	callValue := big.NewInt(15)
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("modifyTotalDelegationCap", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can call this method"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.Arguments = [][]byte{newServiceFee, newServiceFee}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.Arguments = [][]byte{newServiceFee, []byte("wrong arg")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments"))

	vmInput.Arguments = [][]byte{newServiceFee}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation contract config", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	_ = d.saveDelegationContractConfig(&DelegationConfig{})
	vmInput.Arguments = [][]byte{big.NewInt(70).Bytes()}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr = fmt.Errorf("%w getGlobalFundData", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationSystemSC_ExecuteModifyTotalDelegationCap(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("modifyTotalDelegationCap", [][]byte{big.NewInt(500).Bytes()})
	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationContractConfig(&DelegationConfig{})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive: big.NewInt(1000),
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot make total delegation cap smaller than active"))

	vmInput.Arguments = [][]byte{big.NewInt(1500).Bytes()}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dConfig, _ := d.getDelegationContractConfig()
	assert.Equal(t, big.NewInt(1500), dConfig.MaxDelegationCap)

	vmInput.Arguments = [][]byte{big.NewInt(0).Bytes()}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dConfig, _ = d.getDelegationContractConfig()
	assert.Equal(t, big.NewInt(0), dConfig.MaxDelegationCap)
}

func TestDelegation_getSuccessAndUnSuccessKeysAllUnSuccess(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("bls1")
	blsKey2 := []byte("bls2")
	returnData := [][]byte{blsKey1, {failed}, blsKey2, {failed}}
	blsKeys := [][]byte{blsKey1, blsKey2}

	okKeys, failedKeys := getSuccessAndUnSuccessKeys(returnData, blsKeys)
	assert.Nil(t, okKeys)
	assert.Equal(t, 2, len(failedKeys))
	assert.Equal(t, blsKey1, failedKeys[0])
	assert.Equal(t, blsKey2, failedKeys[1])
}

func TestDelegation_getSuccessAndUnSuccessKeysAllSuccess(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("bls1")
	blsKey2 := []byte("bls2")
	returnData := [][]byte{blsKey1, {ok}, blsKey2, {ok}}
	blsKeys := [][]byte{blsKey1, blsKey2}

	okKeys, failedKeys := getSuccessAndUnSuccessKeys(returnData, blsKeys)
	assert.Equal(t, 0, len(failedKeys))
	assert.Equal(t, 2, len(okKeys))
	assert.Equal(t, blsKey1, okKeys[0])
	assert.Equal(t, blsKey2, okKeys[1])
}

func TestDelegation_getSuccessAndUnSuccessKeys(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("bls1")
	blsKey2 := []byte("bls2")
	blsKey3 := []byte("bls3")
	returnData := [][]byte{blsKey1, {ok}, blsKey2, {failed}, blsKey3, {waiting}}
	blsKeys := [][]byte{blsKey1, blsKey2, blsKey3}

	okKeys, failedKeys := getSuccessAndUnSuccessKeys(returnData, blsKeys)
	assert.Equal(t, 2, len(okKeys))
	assert.Equal(t, blsKey1, okKeys[0])
	assert.Equal(t, blsKey3, okKeys[1])

	assert.Equal(t, 1, len(failedKeys))
	assert.Equal(t, blsKey2, failedKeys[0])
}

func TestDelegation_ExecuteUpdateRewardsUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("updateRewards", [][]byte{})
	vmInput.CallerAddr = []byte("eoeAddress")
	d, _ := NewDelegationSystemSC(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only end of epoch address can call this function"))

	vmInput.CallerAddr = vm.EndOfEpochAddress
	vmInput.Arguments = [][]byte{[]byte("arg")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "must call without arguments"))

	vmInput.Arguments = [][]byte{}
	vmInput.CallValue = big.NewInt(-10)
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot call with negative value"))
}

func TestDelegation_ExecuteUpdateRewards(t *testing.T) {
	t.Parallel()

	currentEpoch := uint32(15)
	callValue := big.NewInt(20)
	totalActive := big.NewInt(200)
	serviceFee := big.NewInt(100)
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{
			CurrentEpochCalled: func() uint32 {
				return currentEpoch
			},
		},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("updateRewards", [][]byte{})
	vmInput.CallValue = callValue
	vmInput.CallerAddr = vm.EndOfEpochAddress
	d, _ := NewDelegationSystemSC(args)

	d.eei.SetStorage([]byte(totalActiveKey), totalActive.Bytes())
	d.eei.SetStorage([]byte(serviceFeeKey), serviceFee.Bytes())

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	wasPresent, rewardData, err := d.getRewardComputationData(currentEpoch)
	assert.True(t, wasPresent)
	assert.Nil(t, err)
	assert.Equal(t, serviceFee.Uint64(), rewardData.ServiceFee)
	assert.Equal(t, totalActive, rewardData.TotalActive)
	assert.Equal(t, callValue, rewardData.RewardsToDistribute)
}

func TestDelegation_ExecuteClaimRewardsUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("claimRewards", [][]byte{{10}})
	d, _ := NewDelegationSystemSC(args)

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "wrong number of arguments"))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "caller is not a delegator"))
}

func TestDelegation_ExecuteClaimRewards(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	eei, _ := NewVMContext(
		blockChainHook,
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	args.DelegationSCConfig.MaxServiceFee = 10000
	vmInput := getDefaultVmInputForFunc("claimRewards", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	fundKey := []byte{1}
	_ = d.saveDelegatorData(vmInput.CallerAddr, &DelegatorData{
		ActiveFund:            fundKey,
		RewardsCheckpoint:     0,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})

	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(1000),
	})

	_ = d.saveRewardData(0, &RewardComputationData{
		RewardsToDistribute: big.NewInt(100),
		TotalActive:         big.NewInt(1000),
		ServiceFee:          1000,
	})

	_ = d.saveRewardData(1, &RewardComputationData{
		RewardsToDistribute: big.NewInt(100),
		TotalActive:         big.NewInt(2000),
		ServiceFee:          1000,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	destAcc, exists := eei.outputAccounts[string(vmInput.CallerAddr)]
	assert.True(t, exists)
	_, exists = eei.outputAccounts[string(vmInput.RecipientAddr)]
	assert.True(t, exists)

	assert.Equal(t, 1, len(destAcc.OutputTransfers))
	outputTransfer := destAcc.OutputTransfers[0]
	assert.Equal(t, big.NewInt(135), outputTransfer.Value)

	_, delegatorData, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.Equal(t, uint32(3), delegatorData.RewardsCheckpoint)
	assert.Equal(t, uint64(0), delegatorData.UnClaimedRewards.Uint64())
	assert.Equal(t, uint64(135), delegatorData.TotalCumulatedRewards.Uint64())

	blockChainHook.CurrentEpochCalled = func() uint32 {
		return 3
	}

	_ = d.saveRewardData(3, &RewardComputationData{
		RewardsToDistribute: big.NewInt(100),
		TotalActive:         big.NewInt(2000),
		ServiceFee:          1000,
	})

	vmInput = getDefaultVmInputForFunc("getTotalCumulatedRewardsForUser", [][]byte{vmInput.CallerAddr})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	_, delegatorData, _ = d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.Equal(t, uint64(0), delegatorData.UnClaimedRewards.Uint64())
	assert.Equal(t, uint64(135), delegatorData.TotalCumulatedRewards.Uint64())
	lastValue := eei.output[len(eei.output)-1]
	assert.Equal(t, big.NewInt(0).SetBytes(lastValue).Uint64(), uint64(180))
}

func TestDelegation_ExecuteReDelegateRewards(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{
			CurrentEpochCalled: func() uint32 {
				return 2
			},
		},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}})

	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	args.DelegationSCConfig.MinServiceFee = 0
	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc(core.SCDeployInitFunctionName, [][]byte{big.NewInt(0).Bytes(), big.NewInt(0).Bytes()})
	vmInput.CallValue = big.NewInt(1000)
	vmInput.RecipientAddr = createNewAddress(vm.FirstDelegationSCAddress)
	vmInput.CallerAddr = []byte("stakingProvider")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	fundKey := []byte{1}
	_ = d.saveDelegatorData(vmInput.CallerAddr, &DelegatorData{
		ActiveFund:            fundKey,
		RewardsCheckpoint:     0,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})

	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(1000),
	})

	_ = d.saveRewardData(0, &RewardComputationData{
		RewardsToDistribute: big.NewInt(100),
		TotalActive:         big.NewInt(1000),
		ServiceFee:          1000,
	})

	_ = d.saveRewardData(1, &RewardComputationData{
		RewardsToDistribute: big.NewInt(100),
		TotalActive:         big.NewInt(2000),
		ServiceFee:          1000,
	})
	vmInput = getDefaultVmInputForFunc("reDelegateRewards", [][]byte{})
	vmInput.CallerAddr = []byte("stakingProvider")
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	_, exists := eei.outputAccounts[string(vmInput.CallerAddr)]
	assert.False(t, exists)
	_, exists = eei.outputAccounts[string(vmInput.RecipientAddr)]
	assert.True(t, exists)

	_, delegatorData, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.Equal(t, uint32(3), delegatorData.RewardsCheckpoint)
	assert.Equal(t, uint64(0), delegatorData.UnClaimedRewards.Uint64())

	activeFund, _ := d.getFund(delegatorData.ActiveFund)
	assert.Equal(t, big.NewInt(155+1000), activeFund.Value)
}

func TestDelegation_ExecuteGetRewardDataUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getRewardData", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "must call with 1 arguments"))

	vmInput.Arguments = [][]byte{{2}}
	vmInput.CallValue = big.NewInt(-10)
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "reward not found"))
}

func TestDelegation_ExecuteGetRewardData(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getRewardData", [][]byte{{2}})
	d, _ := NewDelegationSystemSC(args)

	rewardsToDistribute := big.NewInt(100)
	totalActive := big.NewInt(2000)
	serviceFee := uint64(10000)
	_ = d.saveRewardData(2, &RewardComputationData{
		RewardsToDistribute: rewardsToDistribute,
		TotalActive:         totalActive,
		ServiceFee:          serviceFee,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 3, len(eei.output))
	assert.Equal(t, rewardsToDistribute, big.NewInt(0).SetBytes(eei.output[0]))
	assert.Equal(t, totalActive, big.NewInt(0).SetBytes(eei.output[1]))
	assert.Equal(t, uint16(serviceFee), binary.BigEndian.Uint16(eei.output[2]))
}

func TestDelegation_ExecuteGetClaimableRewardsUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getClaimableRewards", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "view function works only for existing delegators"))
}

func TestDelegation_ExecuteGetClaimableRewards(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{
			CurrentEpochCalled: func() uint32 {
				return 2
			},
		},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	delegatorAddr := []byte("address")
	vmInput := getDefaultVmInputForFunc("getClaimableRewards", [][]byte{delegatorAddr})
	d, _ := NewDelegationSystemSC(args)

	fundKey := []byte{1}
	_ = d.saveDelegatorData(delegatorAddr, &DelegatorData{
		ActiveFund:            fundKey,
		RewardsCheckpoint:     0,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})

	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(1000),
	})

	_ = d.saveRewardData(0, &RewardComputationData{
		RewardsToDistribute: big.NewInt(100),
		TotalActive:         big.NewInt(1000),
		ServiceFee:          1000,
	})

	_ = d.saveRewardData(1, &RewardComputationData{
		RewardsToDistribute: big.NewInt(100),
		TotalActive:         big.NewInt(2000),
		ServiceFee:          1000,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, big.NewInt(135), big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegation_ExecuteGetTotalCumulatedRewardsUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getTotalCumulatedRewards", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "this is a view function only"))
}

func TestDelegation_ExecuteGetTotalCumulatedRewards(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{
			CurrentEpochCalled: func() uint32 {
				return 2
			},
		},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getTotalCumulatedRewards", [][]byte{})
	vmInput.CallerAddr = vm.EndOfEpochAddress
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveRewardData(0, &RewardComputationData{
		RewardsToDistribute: big.NewInt(100),
		TotalActive:         big.NewInt(1000),
		ServiceFee:          10000,
	})

	_ = d.saveRewardData(1, &RewardComputationData{
		RewardsToDistribute: big.NewInt(200),
		TotalActive:         big.NewInt(2000),
		ServiceFee:          10000,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, big.NewInt(300), big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegation_ExecuteGetNumUsersUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getNumUsers", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegation_ExecuteGetNumUsers(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{
			CurrentEpochCalled: func() uint32 {
				return 2
			},
		},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getNumUsers", [][]byte{})
	vmInput.CallerAddr = vm.EndOfEpochAddress
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegationStatus(&DelegationContractStatus{
		NumUsers: 3,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, []byte{3}, eei.output[0])
}

func TestDelegation_ExecuteGetTotalUnStakedUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getTotalUnStaked", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w getGlobalFundData", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegation_ExecuteGetTotalUnStaked(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{
			CurrentEpochCalled: func() uint32 {
				return 2
			},
		},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getTotalUnStaked", [][]byte{})
	vmInput.CallerAddr = vm.EndOfEpochAddress
	d, _ := NewDelegationSystemSC(args)

	totalUnstaked := big.NewInt(1100)
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalUnStaked: totalUnstaked,
		TotalActive:   big.NewInt(0),
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, totalUnstaked, big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegation_ExecuteGetTotalActiveStakeUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getTotalActiveStake", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w getGlobalFundData", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegation_ExecuteGetTotalActiveStake(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{
			CurrentEpochCalled: func() uint32 {
				return 2
			},
		},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getTotalActiveStake", [][]byte{})
	vmInput.CallerAddr = vm.EndOfEpochAddress
	d, _ := NewDelegationSystemSC(args)

	totalActive := big.NewInt(5000)
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive: totalActive,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, totalActive, big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegation_ExecuteGetUserActiveStakeUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getUserActiveStake", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "view function works only for existing delegators"))
}

func TestDelegation_ExecuteGetUserActiveStakeNoActiveFund(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFunc("getUserActiveStake", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		ActiveFund: nil,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, big.NewInt(0), big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegation_ExecuteGetUserActiveStake(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFunc("getUserActiveStake", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSC(args)

	fundKey := []byte{2}
	fundValue := big.NewInt(150)
	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		ActiveFund: fundKey,
	})

	_ = d.saveFund(fundKey, &Fund{
		Value: fundValue,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, fundValue, big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegation_ExecuteGetUserUnStakedValueUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getUserUnStakedValue", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "view function works only for existing delegators"))
}

func TestDelegation_ExecuteGetUserUnStakedValueNoUnStakedFund(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFunc("getUserUnStakedValue", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		UnStakedFunds: nil,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, big.NewInt(0), big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegation_ExecuteGetUserUnStakedValue(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFunc("getUserUnStakedValue", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSC(args)

	fundKey1 := []byte{2}
	fundKey2 := []byte{3}
	fundValue := big.NewInt(150)
	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		UnStakedFunds: [][]byte{fundKey1, fundKey2},
	})

	_ = d.saveFund(fundKey1, &Fund{
		Value: fundValue,
	})

	_ = d.saveFund(fundKey2, &Fund{
		Value: fundValue,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, big.NewInt(300), big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegation_ExecuteGetUserUnBondableUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getUserUnBondable", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "view function works only for existing delegators"))

	_ = d.saveDelegatorData([]byte("address"), &DelegatorData{})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation contract config", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegation_ExecuteGetUserUnBondableNoUnStakedFund(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFunc("getUserUnBondable", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		UnStakedFunds: nil,
	})

	_ = d.saveDelegationContractConfig(&DelegationConfig{
		UnBondPeriod: 10,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, big.NewInt(0), big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegation_ExecuteGetUserUnBondable(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{
			CurrentNonceCalled: func() uint64 {
				return 500
			},
		},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFunc("getUserUnBondable", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSC(args)

	fundKey1 := []byte{2}
	fundKey2 := []byte{3}
	fundValue := big.NewInt(150)
	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		UnStakedFunds: [][]byte{fundKey1, fundKey2},
	})

	_ = d.saveFund(fundKey1, &Fund{
		Value: fundValue,
		Nonce: 400,
	})

	_ = d.saveFund(fundKey2, &Fund{
		Value: fundValue,
		Nonce: 495,
	})

	_ = d.saveDelegationContractConfig(&DelegationConfig{
		UnBondPeriod: 10,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, fundValue, big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegation_ExecuteGetNumNodesUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getNumNodes", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegation_ExecuteGetNumNodes(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getNumNodes", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegationStatus(&DelegationContractStatus{
		StakedKeys:    []*NodesData{{}},
		NotStakedKeys: []*NodesData{{}, {}},
		UnStakedKeys:  []*NodesData{{}, {}, {}},
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, []byte{6}, eei.output[0])
}

func TestDelegation_ExecuteGetAllNodeStatesUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getAllNodeStates", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegation_ExecuteGetAllNodeStates(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getAllNodeStates", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	blsKey3 := []byte("blsKey3")
	blsKey4 := []byte("blsKey4")
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		StakedKeys:    []*NodesData{{BLSKey: blsKey1}},
		NotStakedKeys: []*NodesData{{BLSKey: blsKey2}, {BLSKey: blsKey3}},
		UnStakedKeys:  []*NodesData{{BLSKey: blsKey4}},
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 7, len(eei.output))
	assert.Equal(t, []byte("staked"), eei.output[0])
	assert.Equal(t, blsKey1, eei.output[1])
	assert.Equal(t, []byte("notStaked"), eei.output[2])
	assert.Equal(t, blsKey2, eei.output[3])
	assert.Equal(t, blsKey3, eei.output[4])
	assert.Equal(t, []byte("unStaked"), eei.output[5])
	assert.Equal(t, blsKey4, eei.output[6])
}

func TestDelegation_ExecuteGetContractConfigUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getContractConfig", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation contract config", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegation_ExecuteGetContractConfig(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getContractConfig", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	ownerAddress := []byte("owner")
	maxDelegationCap := big.NewInt(200)
	serviceFee := uint64(10000)
	initialOwnerFunds := big.NewInt(500)
	createdNonce := uint64(100)
	unBondPeriod := uint64(144000)
	_ = d.saveDelegationContractConfig(&DelegationConfig{
		MaxDelegationCap:            maxDelegationCap,
		InitialOwnerFunds:           initialOwnerFunds,
		AutomaticActivation:         true,
		ChangeableServiceFee:        true,
		CheckCapOnReDelegateRewards: true,
		CreatedNonce:                createdNonce,
		UnBondPeriod:                unBondPeriod,
	})
	eei.SetStorage([]byte(ownerKey), ownerAddress)
	eei.SetStorage([]byte(serviceFeeKey), big.NewInt(0).SetUint64(serviceFee).Bytes())

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	require.Equal(t, 10, len(eei.output))
	assert.Equal(t, ownerAddress, eei.output[0])
	assert.Equal(t, big.NewInt(0).SetUint64(serviceFee), big.NewInt(0).SetBytes(eei.output[1]))
	assert.Equal(t, maxDelegationCap, big.NewInt(0).SetBytes(eei.output[2]))
	assert.Equal(t, initialOwnerFunds, big.NewInt(0).SetBytes(eei.output[3]))
	assert.Equal(t, []byte("true"), eei.output[4])
	assert.Equal(t, []byte("true"), eei.output[5])
	assert.Equal(t, []byte("true"), eei.output[6])
	assert.Equal(t, []byte("true"), eei.output[7])
	assert.Equal(t, big.NewInt(0).SetUint64(createdNonce), big.NewInt(0).SetBytes(eei.output[8]))
	assert.Equal(t, big.NewInt(0).SetUint64(unBondPeriod), big.NewInt(0).SetBytes(eei.output[9]))
}

func TestDelegation_ExecuteUnknownFunc(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	invalidFunc := "invalid func"
	vmInput := getDefaultVmInputForFunc(invalidFunc, [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := invalidFunc + " is an unknown function"
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr))
}

func TestDelegation_computeAndUpdateRewardsWithTotalActiveZeroDoesNotPanic(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{
			CurrentEpochCalled: func() uint32 {
				return 1
			},
		},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei
	d, _ := NewDelegationSystemSC(args)

	fundKey := []byte{2}
	dData := &DelegatorData{
		ActiveFund:            fundKey,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	}

	rewards := big.NewInt(1000)
	_ = d.saveFund(fundKey, &Fund{Value: big.NewInt(1)})
	_ = d.saveRewardData(1, &RewardComputationData{
		TotalActive:         big.NewInt(0),
		RewardsToDistribute: rewards,
	})

	ownerAddr := []byte("ownerAddress")
	eei.SetStorage([]byte(ownerKey), ownerAddr)

	err := d.computeAndUpdateRewards([]byte("other address"), dData)
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(0), dData.UnClaimedRewards)
}

func TestDelegation_computeAndUpdateRewardsWithTotalActiveZeroSendsAllRewardsToOwner(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{
			CurrentEpochCalled: func() uint32 {
				return 1
			},
		},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei
	d, _ := NewDelegationSystemSC(args)

	fundKey := []byte{2}
	dData := &DelegatorData{
		ActiveFund:            fundKey,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	}

	rewards := big.NewInt(1000)
	_ = d.saveFund(fundKey, &Fund{Value: big.NewInt(1)})
	_ = d.saveRewardData(1, &RewardComputationData{
		TotalActive:         big.NewInt(0),
		RewardsToDistribute: rewards,
	})

	ownerAddr := []byte("ownerAddress")
	eei.SetStorage([]byte(ownerKey), ownerAddr)

	err := d.computeAndUpdateRewards(ownerAddr, dData)
	assert.Nil(t, err)
	assert.Equal(t, rewards, dData.UnClaimedRewards)
}

func TestDelegation_isDelegatorShouldErrBecauseAddressIsNotFound(t *testing.T) {
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	delegatorAddress := []byte("address which didn't delegate")
	vmInput := getDefaultVmInputForFunc("isDelegator", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSC(args)

	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestDelegation_isDelegatorShouldWork(t *testing.T) {
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFunc("isDelegator", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSC(args)

	fundKey := []byte{2}

	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		ActiveFund: fundKey,
	})

	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestDelegation_getDelegatorFundsDataDelegatorNotFoundShouldErr(t *testing.T) {
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFunc("getDelegatorFundsData", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSC(args)

	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Contains(t, eei.returnMessage, "existing delegators")
}

func TestDelegation_getDelegatorFundsDataCannotLoadFundsShouldErr(t *testing.T) {
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFunc("getDelegatorFundsData", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSC(args)

	fundKey := []byte{2}
	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		ActiveFund:            fundKey,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})

	_ = d.saveDelegationContractConfig(&DelegationConfig{
		AutomaticActivation:  false,
		ChangeableServiceFee: true,
	})

	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Contains(t, eei.returnMessage, vm.ErrDataNotFoundUnderKey.Error())
}

func TestDelegation_getDelegatorFundsDataCannotFindConfigShouldErr(t *testing.T) {
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFunc("getDelegatorFundsData", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSC(args)

	fundKey := []byte{2}
	fundValue := big.NewInt(150)
	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		ActiveFund:            fundKey,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})

	_ = d.saveFund(fundKey, &Fund{
		Value: fundValue,
	})

	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Contains(t, eei.returnMessage, vm.ErrDataNotFoundUnderKey.Error())
}

func TestDelegation_getDelegatorFundsDataShouldWork(t *testing.T) {
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFunc("getDelegatorFundsData", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSC(args)

	fundKey := []byte{2}
	fundValue := big.NewInt(150)
	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		ActiveFund:            fundKey,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})

	_ = d.saveFund(fundKey, &Fund{
		Value: fundValue,
	})

	_ = d.saveDelegationContractConfig(&DelegationConfig{
		AutomaticActivation:  false,
		ChangeableServiceFee: true,
	})

	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, retCode)

	assert.Equal(t, fundValue.Bytes(), eei.output[0])
}

func TestDelegation_setAndGetDelegationMetadata(t *testing.T) {
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)

	vmInput := getDefaultVmInputForFunc("setMetaData", [][]byte{[]byte("name"), []byte("website"), []byte("identifier")})
	d.eei.SetStorage([]byte(ownerKey), vmInput.CallerAddr)
	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, retCode)

	vmInputErr := getDefaultVmInputForFunc("setMetaData", [][]byte{[]byte("one")})
	retCode = d.Execute(vmInputErr)
	assert.Equal(t, vmcommon.UserError, retCode)

	vmInputGet := getDefaultVmInputForFunc("getMetaData", [][]byte{})
	retCode = d.Execute(vmInputGet)
	assert.Equal(t, vmcommon.Ok, retCode)

	assert.Equal(t, eei.output[0], vmInput.Arguments[0])
	assert.Equal(t, eei.output[1], vmInput.Arguments[1])
	assert.Equal(t, eei.output[2], vmInput.Arguments[2])
}

func TestDelegation_setAutomaticActivation(t *testing.T) {
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationContractConfig(&DelegationConfig{})

	vmInput := getDefaultVmInputForFunc("setAutomaticActivation", [][]byte{[]byte("true")})
	d.eei.SetStorage([]byte(ownerKey), vmInput.CallerAddr)
	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, retCode)

	dConfig, _ := d.getDelegationContractConfig()
	assert.Equal(t, dConfig.AutomaticActivation, true)

	vmInput = getDefaultVmInputForFunc("setAutomaticActivation", [][]byte{[]byte("abcd")})
	retCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, retCode)

	vmInput = getDefaultVmInputForFunc("setAutomaticActivation", [][]byte{[]byte("false")})
	retCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, retCode)

	dConfig, _ = d.getDelegationContractConfig()
	assert.Equal(t, dConfig.AutomaticActivation, false)

	vmInput = getDefaultVmInputForFunc("setCheckCapOnReDelegateRewards", [][]byte{[]byte("true")})
	retCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, retCode)

	vmInput = getDefaultVmInputForFunc("setCheckCapOnReDelegateRewards", [][]byte{[]byte("abcd")})
	retCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, retCode)

	dConfig, _ = d.getDelegationContractConfig()
	assert.Equal(t, dConfig.CheckCapOnReDelegateRewards, true)

	vmInput = getDefaultVmInputForFunc("setCheckCapOnReDelegateRewards", [][]byte{[]byte("false")})
	retCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, retCode)

	dConfig, _ = d.getDelegationContractConfig()
	assert.Equal(t, dConfig.CheckCapOnReDelegateRewards, false)
}

func TestDelegation_GetDelegationManagementNoDataShouldError(t *testing.T) {
	t.Parallel()

	d := &delegation{
		eei: &mock.SystemEIStub{
			GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
				return nil
			},
		},
	}

	delegationManagement, err := d.getDelegationManagement()

	assert.Nil(t, delegationManagement)
	assert.True(t, errors.Is(err, vm.ErrDataNotFoundUnderKey))
}

func TestDelegation_GetDelegationManagementMarshalizerFailsShouldError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	d := &delegation{
		eei: &mock.SystemEIStub{
			GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
				return make([]byte, 1)
			},
		},
		marshalizer: &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		},
	}

	delegationManagement, err := d.getDelegationManagement()

	assert.Nil(t, delegationManagement)
	assert.True(t, errors.Is(err, expectedErr))
}

func TestDelegation_GetDelegationManagementShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	minDelegationAmount := big.NewInt(45)
	minDeposit := big.NewInt(2232)
	cfg := &DelegationManagement{
		MinDelegationAmount: minDelegationAmount,
		MinDeposit:          minDeposit,
	}

	buff, err := marshalizer.Marshal(cfg)
	require.Nil(t, err)

	d := &delegation{
		eei: &mock.SystemEIStub{
			GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
				return buff
			},
		},
		marshalizer: marshalizer,
	}

	delegationManagement, err := d.getDelegationManagement()

	assert.Nil(t, err)
	require.NotNil(t, delegationManagement)
	assert.Equal(t, minDeposit, delegationManagement.MinDeposit)
	assert.Equal(t, minDelegationAmount, delegationManagement.MinDelegationAmount)
}
