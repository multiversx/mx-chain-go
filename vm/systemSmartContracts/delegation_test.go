package systemSmartContracts

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
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
		EndOfEpochAddress:      vm.EndOfEpochAddress,
		GovernanceSCAddress:    vm.GovernanceSCAddress,
		AddTokensAddress:       bytes.Repeat([]byte{1}, 32),
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsDelegationSmartContractFlagEnabledField:              true,
			IsStakingV2FlagEnabledForActivationEpochCompletedField: true,
			IsAddTokensToDelegationFlagEnabledField:                true,
			IsDeleteDelegatorAfterClaimRewardsFlagEnabledField:     true,
			IsComputeRewardCheckpointFlagEnabledField:              true,
			IsValidatorToDelegationFlagEnabledField:                true,
			IsReDelegateBelowMinCheckFlagEnabledField:              true,
			IsMultiClaimOnDelegationEnabledField:                   true,
		},
	}
}

func addValidatorAndStakingScToVmContext(eei *vmContext) {
	validatorArgs := createMockArgumentsForValidatorSC()
	validatorArgs.Eei = eei
	validatorArgs.StakingSCConfig.GenesisNodePrice = "100"
	validatorArgs.StakingSCAddress = vm.StakingSCAddress
	enableEpochsHandler, _ := validatorArgs.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
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
			enableEpochsHandler.IsStakingV2FlagEnabledField = true
			_ = validatorSc.saveRegistrationData([]byte("addr"), &ValidatorDataV2{
				RewardAddress:   []byte("rewardAddr"),
				TotalStakeValue: big.NewInt(1000),
				LockedStake:     big.NewInt(500),
				BlsPubKeys:      [][]byte{[]byte("blsKey1"), []byte("blsKey2")},
				TotalUnstaked:   big.NewInt(150),
				UnstakedInfo: []*UnstakedValue{
					{
						UnstakedEpoch: 10,
						UnstakedValue: big.NewInt(60),
					},
					{
						UnstakedEpoch: 50,
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

func createDelegationContractAndEEI() (*delegation, *vmContext) {
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(VMContextArgs{
		BlockChainHook: &mock.BlockChainHookStub{
			CurrentEpochCalled: func() uint32 {
				return 2
			},
		},
		CryptoHook:          hooks.NewVMCryptoHook(),
		InputParser:         &mock.ArgumentParserMock{},
		ValidatorAccountsDB: &stateMock.AccountsStub{},
		UserAccountsDB:      &stateMock.AccountsStub{},
		ChanceComputer:      &mock.RaterMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ShardCoordinator:    &mock.ShardCoordinatorStub{},
	})
	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}}
	_ = eei.SetSystemSCContainer(systemSCContainerStub)

	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	args.DelegationSCConfig.MinServiceFee = 0
	d, _ := NewDelegationSystemSC(args)
	return d, eei
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

func TestNewDelegationSystemSC_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	args.EnableEpochsHandler = nil

	d, err := NewDelegationSystemSC(args)
	assert.Nil(t, d)
	assert.Equal(t, vm.ErrNilEnableEpochsHandler, err)
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

	d, err := NewDelegationSystemSC(createMockArgumentsForDelegation())
	assert.Nil(t, err)
	assert.False(t, check.IfNil(d))
}

func TestDelegationSystemSC_ExecuteNilArgsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
	args.Eei = eei
	d, _ := NewDelegationSystemSC(args)

	output := d.Execute(nil)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInputArgsIsNil.Error()))
}

func TestDelegationSystemSC_ExecuteDelegationDisabledShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	d, _ := NewDelegationSystemSC(args)
	enableEpochsHandler.IsDelegationSmartContractFlagEnabledField = false
	vmInput := getDefaultVmInputForFunc("addNodes", [][]byte{})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "delegation contract is not enabled"))
}

func TestDelegationSystemSC_ExecuteInitScAlreadyPresentShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	createdEpoch := uint32(150)
	callValue := big.NewInt(130)
	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return createdEpoch
		},
		CurrentNonceCalled: func() uint64 {
			return uint64(createdEpoch)
		},
	}
	createDelegationManagerConfig(eei, args.Marshalizer, callValue)
	args.Eei = eei
	args.StakingSCConfig.UnBondPeriod = 20
	args.StakingSCConfig.UnBondPeriodInEpochs = 20
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
	assert.Equal(t, uint64(createdEpoch), dConf.CreatedNonce)
	assert.Equal(t, big.NewInt(20).Uint64(), uint64(dConf.UnBondPeriodInEpochs))

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
	assert.Equal(t, createdEpoch, ownerFund.Epoch)
	assert.Equal(t, active, ownerFund.Type)

	dGlobalFund, err := d.getGlobalFundData()
	assert.Nil(t, err)
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()

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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
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
	enableEpochsHandler, _ := validatorArgs.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.IsStakingV2FlagEnabledField = true
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()

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
	eei := createDefaultEei()

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
	eei := createDefaultEei()

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
	eei := createDefaultEei()

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
	eei := createDefaultEei()

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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, uint64(1), dStatus.NumUsers)

	_, dData, _ := d.getOrCreateDelegatorData(delegator1)
	assert.Equal(t, fundKey, dData.ActiveFund)
}

func TestDelegationSystemSC_ExecuteDelegateFailsWhenGettingDelegationManagement(t *testing.T) {
	t.Parallel()

	delegator1 := []byte("delegator1")
	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
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
	assert.Equal(t, big.NewInt(20), globalFund.TotalActive)
	assert.Equal(t, big.NewInt(80), globalFund.TotalUnStaked)

	_, dData, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.Equal(t, 1, len(dData.UnStakedFunds))
	assert.Equal(t, nextFundKey, dData.UnStakedFunds[0])

	_ = d.saveDelegationContractConfig(&DelegationConfig{
		UnBondPeriodInEpochs: 50,
	})

	blockChainHook.CurrentEpochCalled = func() uint32 {
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
	eei := createDefaultEei()
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
		TotalActive:   big.NewInt(100),
		TotalUnStaked: big.NewInt(0),
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
	assert.Equal(t, big.NewInt(0), globalFund.TotalActive)
	assert.Equal(t, big.NewInt(100), globalFund.TotalUnStaked)

	_, dData, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.Equal(t, 1, len(dData.UnStakedFunds))
	assert.Equal(t, nextFundKey, dData.UnStakedFunds[0])
}

func TestDelegationSystemSC_ExecuteUnDelegateAllFundsAsOwner(t *testing.T) {
	t.Parallel()

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	nextFundKey := append([]byte(fundKeyPrefix), []byte{2}...)
	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
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
		TotalActive:   big.NewInt(100),
		TotalUnStaked: big.NewInt(0),
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
	assert.Equal(t, big.NewInt(0), globalFund.TotalActive)
	assert.Equal(t, big.NewInt(100), globalFund.TotalUnStaked)

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

func TestDelegationSystemSC_ExecuteUnDelegateMultipleTimesSameAndDiffEpochAndWithdraw(t *testing.T) {
	t.Parallel()

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	nextFundKey := append([]byte(fundKeyPrefix), []byte{2}...)
	thirdFundKey := append([]byte(fundKeyPrefix), []byte{3}...)
	args := createMockArgumentsForDelegation()
	currentEpoch := uint32(10)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return currentEpoch
		},
	}
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	args.Eei = eei
	args.StakingSCConfig.UnBondPeriodInEpochs = 10
	addValidatorAndStakingScToVmContext(eei)
	createDelegationManagerConfig(eei, args.Marshalizer, big.NewInt(10))

	vmInput := getDefaultVmInputForFunc("unDelegate", [][]byte{{10}})
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
		TotalActive:   big.NewInt(100),
		TotalUnStaked: big.NewInt(0),
	})
	_ = d.saveDelegationContractConfig(&DelegationConfig{
		MaxDelegationCap:            big.NewInt(0),
		InitialOwnerFunds:           big.NewInt(0),
		AutomaticActivation:         false,
		ChangeableServiceFee:        false,
		CreatedNonce:                0,
		UnBondPeriodInEpochs:        args.StakingSCConfig.UnBondPeriodInEpochs,
		CheckCapOnReDelegateRewards: false,
	})
	_ = d.saveDelegationStatus(&DelegationContractStatus{NumUsers: 10})
	d.eei.SetStorage([]byte(lastFundKey), fundKey)

	for i := 0; i < 5; i++ {
		output := d.Execute(vmInput)
		assert.Equal(t, vmcommon.Ok, output)
	}
	currentEpoch += 1
	for i := 0; i < 5; i++ {
		output := d.Execute(vmInput)
		assert.Equal(t, vmcommon.Ok, output)
	}

	dFund, _ := d.getFund(fundKey)
	assert.Nil(t, dFund)

	dFund, _ = d.getFund(nextFundKey)
	assert.Equal(t, big.NewInt(50), dFund.Value)
	assert.Equal(t, currentEpoch-1, dFund.Epoch)
	assert.Equal(t, unStaked, dFund.Type)
	assert.Equal(t, vmInput.CallerAddr, dFund.Address)

	dFund, _ = d.getFund(thirdFundKey)
	assert.Equal(t, big.NewInt(50), dFund.Value)
	assert.Equal(t, currentEpoch, dFund.Epoch)
	assert.Equal(t, unStaked, dFund.Type)
	assert.Equal(t, vmInput.CallerAddr, dFund.Address)

	globalFund, _ := d.getGlobalFundData()
	assert.Equal(t, big.NewInt(0), globalFund.TotalActive)
	assert.Equal(t, big.NewInt(100), globalFund.TotalUnStaked)

	_, dData, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.Equal(t, 2, len(dData.UnStakedFunds))
	assert.Equal(t, nextFundKey, dData.UnStakedFunds[0])
	assert.Equal(t, thirdFundKey, dData.UnStakedFunds[1])

	currentEpoch += d.unBondPeriodInEpochs - 1
	vmInput.Function = "withdraw"
	vmInput.Arguments = [][]byte{}
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	_, dData, _ = d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.Equal(t, 1, len(dData.UnStakedFunds))
	assert.Equal(t, thirdFundKey, dData.UnStakedFunds[0])

	dFund, _ = d.getFund(thirdFundKey)
	assert.Equal(t, big.NewInt(50), dFund.Value)
	assert.Equal(t, currentEpoch-d.unBondPeriodInEpochs+1, dFund.Epoch)
	assert.Equal(t, unStaked, dFund.Type)
	assert.Equal(t, vmInput.CallerAddr, dFund.Address)

	currentEpoch += 1
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	isNew, _, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.True(t, isNew)
}

func TestDelegationSystemSC_ExecuteWithdrawUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return currentNonce
		},
		CurrentEpochCalled: func() uint32 {
			return uint32(currentNonce)
		},
	}
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
		Epoch:   10,
		Type:    unStaked,
	})
	_ = d.saveFund(fundKey2, &Fund{
		Value:   big.NewInt(80),
		Address: vmInput.CallerAddr,
		Epoch:   50,
		Type:    unStaked,
	})
	_ = d.saveDelegationContractConfig(&DelegationConfig{
		UnBondPeriodInEpochs: 50,
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalUnStaked: big.NewInt(140),
		TotalActive:   big.NewInt(0),
	})

	_ = d.saveDelegationStatus(&DelegationContractStatus{
		NumUsers: 10,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "nothing to unBond")

	gFundData, _ := d.getGlobalFundData()
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
	eei := createDefaultEei()

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
	eei := createDefaultEei()

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
	eei := createDefaultEei()

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
	eei := createDefaultEei()

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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return currentEpoch
		},
	}
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = &mock.ArgumentParserMock{}
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

func TestDelegation_ExecuteClaimRewardsShouldDeleteDelegator(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 10
		},
	}
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	args.Eei = eei

	args.DelegationSCConfig.MaxServiceFee = 10000
	vmInput := getDefaultVmInputForFunc("claimRewards", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegatorData(vmInput.CallerAddr, &DelegatorData{
		ActiveFund:            nil,
		RewardsCheckpoint:     0,
		UnClaimedRewards:      big.NewInt(135),
		TotalCumulatedRewards: big.NewInt(0),
	})

	_ = d.saveDelegationStatus(&DelegationContractStatus{
		NumUsers: 10,
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

	vmInput = getDefaultVmInputForFunc("getTotalCumulatedRewardsForUser", [][]byte{vmInput.CallerAddr})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	res := d.eei.GetStorage(vmInput.CallerAddr)
	require.Len(t, res, 0)
}

func TestDelegation_ExecuteReDelegateRewardsNoExtraCheck(t *testing.T) {
	t.Parallel()

	d, eei := prepareReDelegateRewardsComponents(t, 1000, big.NewInt(1156))
	vmInput := getDefaultVmInputForFunc("reDelegateRewards", [][]byte{})
	vmInput.CallerAddr = []byte("stakingProvider")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	_, exists := eei.outputAccounts[string(vmInput.CallerAddr)]
	assert.False(t, exists)
	_, exists = eei.outputAccounts[string(vmInput.RecipientAddr)]
	assert.True(t, exists)

	_, delegatorData, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.Equal(t, uint32(3), delegatorData.RewardsCheckpoint)
	assert.Equal(t, uint64(0), delegatorData.UnClaimedRewards.Uint64())
	assert.Equal(t, uint64(155), delegatorData.TotalCumulatedRewards.Uint64())

	activeFund, _ := d.getFund(delegatorData.ActiveFund)
	assert.Equal(t, big.NewInt(155+1000), activeFund.Value)
}

func TestDelegation_ExecuteReDelegateRewardsWithExtraCheckReDelegateIsAboveMinimum(t *testing.T) {
	t.Parallel()

	d, eei := prepareReDelegateRewardsComponents(t, 0, big.NewInt(155))
	vmInput := getDefaultVmInputForFunc("reDelegateRewards", [][]byte{})
	vmInput.CallerAddr = []byte("stakingProvider")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	_, exists := eei.outputAccounts[string(vmInput.CallerAddr)]
	assert.False(t, exists)
	_, exists = eei.outputAccounts[string(vmInput.RecipientAddr)]
	assert.True(t, exists)

	_, delegatorData, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.Equal(t, uint32(3), delegatorData.RewardsCheckpoint)
	assert.Equal(t, uint64(0), delegatorData.UnClaimedRewards.Uint64())
	assert.Equal(t, uint64(155), delegatorData.TotalCumulatedRewards.Uint64())

	activeFund, _ := d.getFund(delegatorData.ActiveFund)
	assert.Equal(t, big.NewInt(155+1000), activeFund.Value)
}

func TestDelegation_ExecuteReDelegateRewardsWithExtraCheckReDelegateIsBelowMinimum(t *testing.T) {
	t.Parallel()

	d, eei := prepareReDelegateRewardsComponents(t, 0, big.NewInt(1000))
	createDelegationManagerConfig(eei, &mock.MarshalizerMock{}, big.NewInt(1156))
	vmInput := getDefaultVmInputForFunc("reDelegateRewards", [][]byte{})
	vmInput.CallerAddr = []byte("stakingProvider")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func prepareReDelegateRewardsComponents(
	t *testing.T,
	extraCheckEpoch uint32,
	minDelegation *big.Int,
) (*delegation, *vmContext) {
	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}})

	createDelegationManagerConfig(eei, args.Marshalizer, minDelegation)
	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	args.DelegationSCConfig.MinServiceFee = 0
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.IsReDelegateBelowMinCheckFlagEnabledField = extraCheckEpoch == 0
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

	return d, eei
}

func TestDelegation_ExecuteGetRewardDataUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFunc("getUserUnBondable", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		UnStakedFunds: nil,
	})

	_ = d.saveDelegationContractConfig(&DelegationConfig{
		UnBondPeriodInEpochs: 10,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, big.NewInt(0), big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegation_ExecuteGetUserUnBondable(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return 500
		},
		CurrentEpochCalled: func() uint32 {
			return 500
		},
	}
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
		Epoch: 400,
	})

	_ = d.saveFund(fundKey2, &Fund{
		Value: fundValue,
		Epoch: 495,
	})

	_ = d.saveDelegationContractConfig(&DelegationConfig{
		UnBondPeriodInEpochs: 10,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, fundValue, big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegation_ExecuteGetNumNodesUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFunc("getContractConfig", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	ownerAddress := []byte("owner")
	maxDelegationCap := big.NewInt(200)
	serviceFee := uint64(10000)
	initialOwnerFunds := big.NewInt(500)
	createdNonce := uint64(100)
	unBondPeriodInEpoch := uint32(144000)
	_ = d.saveDelegationContractConfig(&DelegationConfig{
		MaxDelegationCap:            maxDelegationCap,
		InitialOwnerFunds:           initialOwnerFunds,
		AutomaticActivation:         true,
		ChangeableServiceFee:        true,
		CheckCapOnReDelegateRewards: true,
		CreatedNonce:                createdNonce,
		UnBondPeriodInEpochs:        unBondPeriodInEpoch,
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
	assert.Equal(t, big.NewInt(0).SetUint64(uint64(unBondPeriodInEpoch)), big.NewInt(0).SetBytes(eei.output[9]))
}

func TestDelegation_ExecuteUnknownFunc(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 1
		},
	}
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
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 1
		},
	}
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
	eei := createDefaultEei()
	args.Eei = eei

	delegatorAddress := []byte("address which didn't delegate")
	vmInput := getDefaultVmInputForFunc("isDelegator", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSC(args)

	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestDelegation_isDelegatorShouldWork(t *testing.T) {
	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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

	delegationManagement, err := getDelegationManagement(d.eei, d.marshalizer, d.delegationMgrSCAddress)

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

	delegationManagement, err := getDelegationManagement(d.eei, d.marshalizer, d.delegationMgrSCAddress)

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

	delegationManagement, err := getDelegationManagement(d.eei, d.marshalizer, d.delegationMgrSCAddress)

	assert.Nil(t, err)
	require.NotNil(t, delegationManagement)
	assert.Equal(t, minDeposit, delegationManagement.MinDeposit)
	assert.Equal(t, minDelegationAmount, delegationManagement.MinDelegationAmount)
}

func TestDelegation_ExecuteInitFromValidatorData(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}})

	createDelegationManagerConfig(eei, args.Marshalizer, big.NewInt(1000))
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
}

func TestDelegation_checkArgumentsForValidatorToDelegation(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}})
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)

	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	args.DelegationSCConfig.MinServiceFee = 0
	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc(initFromValidatorData, [][]byte{big.NewInt(0).Bytes(), big.NewInt(0).Bytes()})

	enableEpochsHandler.IsValidatorToDelegationFlagEnabledField = false
	returnCode := d.checkArgumentsForValidatorToDelegation(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, initFromValidatorData+" is an unknown function")

	enableEpochsHandler.IsValidatorToDelegationFlagEnabledField = true
	eei.returnMessage = ""
	returnCode = d.checkArgumentsForValidatorToDelegation(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "only delegation manager sc can call this function")

	eei.returnMessage = ""
	vmInput.CallerAddr = d.delegationMgrSCAddress
	vmInput.CallValue.SetUint64(10)
	vmInput.Arguments = [][]byte{}
	returnCode = d.checkArgumentsForValidatorToDelegation(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "call value must be 0")

	eei.returnMessage = ""
	vmInput.CallValue.SetUint64(0)
	vmInput.Arguments = [][]byte{}
	returnCode = d.checkArgumentsForValidatorToDelegation(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "not enough arguments")

	eei.returnMessage = ""
	vmInput.Arguments = [][]byte{[]byte("key")}
	returnCode = d.checkArgumentsForValidatorToDelegation(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "invalid arguments, first must be an address")
}

func TestDelegation_getAndVerifyValidatorData(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}})

	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	args.DelegationSCConfig.MinServiceFee = 0
	d, _ := NewDelegationSystemSC(args)

	addr := []byte("address")
	_, returnCode := d.getAndVerifyValidatorData(addr)
	assert.Equal(t, eei.returnMessage, vm.ErrEmptyStorage.Error())
	assert.Equal(t, returnCode, vmcommon.UserError)

	eei.SetStorageForAddress(d.validatorSCAddr, addr, addr)
	_, returnCode = d.getAndVerifyValidatorData(addr)
	assert.Equal(t, returnCode, vmcommon.UserError)

	validatorData := &ValidatorDataV2{
		RewardAddress:   []byte("randomAddress"),
		TotalSlashed:    big.NewInt(0),
		TotalUnstaked:   big.NewInt(0),
		TotalStakeValue: big.NewInt(0),
		UnstakedInfo:    []*UnstakedValue{{UnstakedValue: big.NewInt(10)}},
		NumRegistered:   3,
		BlsPubKeys:      [][]byte{[]byte("firsstKey"), []byte("secondKey"), []byte("thirddKey")},
	}
	marshaledData, _ := d.marshalizer.Marshal(validatorData)
	eei.SetStorageForAddress(d.validatorSCAddr, addr, marshaledData)

	eei.returnMessage = ""
	_, returnCode = d.getAndVerifyValidatorData(addr)
	assert.Equal(t, returnCode, vmcommon.UserError)
	assert.Equal(t, eei.returnMessage, "invalid reward address on validator data")

	validatorData.RewardAddress = addr
	marshaledData, _ = d.marshalizer.Marshal(validatorData)
	eei.SetStorageForAddress(d.validatorSCAddr, addr, marshaledData)

	eei.returnMessage = ""
	_, returnCode = d.getAndVerifyValidatorData(addr)
	assert.Equal(t, returnCode, vmcommon.UserError)

	managementData := &DelegationManagement{
		NumOfContracts:      0,
		LastAddress:         vm.FirstDelegationSCAddress,
		MinServiceFee:       0,
		MaxServiceFee:       100,
		MinDeposit:          big.NewInt(100),
		MinDelegationAmount: big.NewInt(100),
	}
	marshaledData, _ = d.marshalizer.Marshal(managementData)
	eei.SetStorageForAddress(d.delegationMgrSCAddress, []byte(delegationManagementKey), marshaledData)

	eei.returnMessage = ""
	_, returnCode = d.getAndVerifyValidatorData(addr)
	assert.Equal(t, returnCode, vmcommon.UserError)
	assert.Equal(t, eei.returnMessage, "not enough stake to make delegation contract")

	validatorData.TotalStakeValue.SetUint64(10000)
	marshaledData, _ = d.marshalizer.Marshal(validatorData)
	eei.SetStorageForAddress(d.validatorSCAddr, addr, marshaledData)

	eei.returnMessage = ""
	_, returnCode = d.getAndVerifyValidatorData(addr)
	assert.Equal(t, returnCode, vmcommon.UserError)
	assert.Equal(t, eei.returnMessage, "clean unStaked info before changing validator to delegation contract")
}

func TestDelegation_initFromValidatorData(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}}
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)

	_ = eei.SetSystemSCContainer(systemSCContainerStub)

	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	args.DelegationSCConfig.MinServiceFee = 0
	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc(initFromValidatorData, [][]byte{big.NewInt(0).Bytes(), big.NewInt(0).Bytes()})

	enableEpochsHandler.IsValidatorToDelegationFlagEnabledField = false
	returnCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, initFromValidatorData+" is an unknown function")

	enableEpochsHandler.IsValidatorToDelegationFlagEnabledField = true

	eei.returnMessage = ""
	vmInput.CallerAddr = d.delegationMgrSCAddress
	vmInput.CallValue.SetUint64(0)
	oldAddress := bytes.Repeat([]byte{1}, len(vmInput.CallerAddr))
	vmInput.Arguments = [][]byte{oldAddress}
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "invalid number of arguments")

	eei.returnMessage = ""
	vmInput.Arguments = [][]byte{oldAddress, big.NewInt(0).Bytes(), big.NewInt(0).SetUint64(d.maxServiceFee + 1).Bytes()}
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "service fee out of bounds")

	systemSCContainerStub.GetCalled = func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.UserError
		}}, vm.ErrEmptyStorage
	}

	eei.returnMessage = ""
	vmInput.Arguments = [][]byte{oldAddress, big.NewInt(0).SetUint64(d.maxServiceFee).Bytes(), big.NewInt(0).Bytes()}
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "storage is nil for given key@storage is nil for given key")

	systemSCContainerStub.GetCalled = func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.UserError
		}}, nil
	}
	eei.returnMessage = ""
	vmInput.Arguments = [][]byte{oldAddress, big.NewInt(0).SetUint64(d.maxServiceFee).Bytes(), big.NewInt(0).Bytes()}
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)

	systemSCContainerStub.GetCalled = func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}
	eei.returnMessage = ""
	vmInput.Arguments = [][]byte{oldAddress, big.NewInt(0).SetUint64(d.maxServiceFee).Bytes(), big.NewInt(0).Bytes()}
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, vm.ErrEmptyStorage.Error())

	validatorData := &ValidatorDataV2{
		RewardAddress:   vmInput.RecipientAddr,
		TotalSlashed:    big.NewInt(0),
		TotalUnstaked:   big.NewInt(0),
		TotalStakeValue: big.NewInt(1000000),
		NumRegistered:   3,
		BlsPubKeys:      [][]byte{[]byte("firsstKey"), []byte("secondKey"), []byte("thirddKey")},
	}
	marshaledData, _ := d.marshalizer.Marshal(validatorData)
	eei.SetStorageForAddress(d.validatorSCAddr, vmInput.RecipientAddr, marshaledData)

	managementData := &DelegationManagement{
		NumOfContracts:      0,
		LastAddress:         vm.FirstDelegationSCAddress,
		MinServiceFee:       0,
		MaxServiceFee:       100,
		MinDeposit:          big.NewInt(100),
		MinDelegationAmount: big.NewInt(100),
	}
	marshaledData, _ = d.marshalizer.Marshal(managementData)
	eei.SetStorageForAddress(d.delegationMgrSCAddress, []byte(delegationManagementKey), marshaledData)

	eei.returnMessage = ""
	vmInput.Arguments = [][]byte{oldAddress, big.NewInt(0).SetUint64(d.maxServiceFee).Bytes(), big.NewInt(0).Bytes()}
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, vm.ErrEmptyStorage.Error())

	for i, blsKey := range validatorData.BlsPubKeys {
		stakedData := &StakedDataV2_0{
			Staked: true,
		}
		if i == 0 {
			stakedData.Staked = false
		}
		marshaledData, _ = d.marshalizer.Marshal(stakedData)
		eei.SetStorageForAddress(d.stakingSCAddr, blsKey, marshaledData)
	}

	eei.returnMessage = ""
	vmInput.Arguments = [][]byte{oldAddress, big.NewInt(1).Bytes(), big.NewInt(0).SetUint64(d.maxServiceFee).Bytes()}
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "total delegation cap reached")

	eei.returnMessage = ""
	vmInput.Arguments = [][]byte{oldAddress, validatorData.TotalStakeValue.Bytes(), big.NewInt(0).SetUint64(d.maxServiceFee).Bytes()}
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, returnCode)
}

func TestDelegation_mergeValidatorDataToDelegation(t *testing.T) {
	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}}
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)

	_ = eei.SetSystemSCContainer(systemSCContainerStub)

	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	args.DelegationSCConfig.MinServiceFee = 0
	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc(mergeValidatorDataToDelegation, [][]byte{big.NewInt(0).Bytes(), big.NewInt(0).Bytes()})

	enableEpochsHandler.IsValidatorToDelegationFlagEnabledField = false
	returnCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, mergeValidatorDataToDelegation+" is an unknown function")

	enableEpochsHandler.IsValidatorToDelegationFlagEnabledField = true

	eei.returnMessage = ""
	vmInput.CallerAddr = d.delegationMgrSCAddress
	vmInput.CallValue.SetUint64(0)
	oldAddress := bytes.Repeat([]byte{1}, len(vmInput.CallerAddr))
	vmInput.Arguments = [][]byte{oldAddress, oldAddress}
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "invalid number of arguments")

	eei.returnMessage = ""
	vmInput.Arguments = [][]byte{oldAddress}
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, vm.ErrEmptyStorage.Error())

	validatorData := &ValidatorDataV2{
		RewardAddress:   oldAddress,
		TotalSlashed:    big.NewInt(0),
		TotalUnstaked:   big.NewInt(0),
		TotalStakeValue: big.NewInt(1000000),
		NumRegistered:   3,
		BlsPubKeys:      [][]byte{[]byte("firsstKey"), []byte("secondKey"), []byte("thirddKey")},
	}
	marshaledData, _ := d.marshalizer.Marshal(validatorData)
	eei.SetStorageForAddress(d.validatorSCAddr, oldAddress, marshaledData)

	managementData := &DelegationManagement{
		NumOfContracts:      0,
		LastAddress:         vm.FirstDelegationSCAddress,
		MinServiceFee:       0,
		MaxServiceFee:       100,
		MinDeposit:          big.NewInt(100),
		MinDelegationAmount: big.NewInt(100),
	}
	marshaledData, _ = d.marshalizer.Marshal(managementData)
	eei.SetStorageForAddress(d.delegationMgrSCAddress, []byte(delegationManagementKey), marshaledData)

	systemSCContainerStub.GetCalled = func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.UserError
		}}, vm.ErrEmptyStorage
	}

	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "storage is nil for given key@storage is nil for given key")

	systemSCContainerStub.GetCalled = func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.UserError
		}}, nil
	}
	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)

	systemSCContainerStub.GetCalled = func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}
	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "data was not found under requested key delegation status")

	_ = d.saveDelegationStatus(createNewDelegationContractStatus())
	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, vm.ErrEmptyStorage.Error())

	for i, blsKey := range validatorData.BlsPubKeys {
		stakedData := &StakedDataV2_0{
			Staked: true,
		}
		if i == 2 {
			stakedData.Staked = false
		}
		marshaledData, _ = d.marshalizer.Marshal(stakedData)
		eei.SetStorageForAddress(d.stakingSCAddr, blsKey, marshaledData)
	}

	createNewContractInput := getDefaultVmInputForFunc(core.SCDeployInitFunctionName, [][]byte{big.NewInt(1000000).Bytes(), big.NewInt(0).Bytes()})
	createNewContractInput.CallValue = big.NewInt(1000000)
	createNewContractInput.CallerAddr = d.delegationMgrSCAddress
	returnCode = d.Execute(createNewContractInput)
	assert.Equal(t, vmcommon.Ok, returnCode)

	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "total delegation cap reached")

	dConfig, _ := d.getDelegationContractConfig()
	dConfig.MaxDelegationCap.SetUint64(0)
	_ = d.saveDelegationContractConfig(dConfig)

	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, returnCode)

	dStatus, err := d.getDelegationStatus()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(dStatus.UnStakedKeys))
	assert.Equal(t, 2, len(dStatus.StakedKeys))
}

func TestDelegation_whitelistForMerge(t *testing.T) {
	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}}
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)

	_ = eei.SetSystemSCContainer(systemSCContainerStub)

	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	args.DelegationSCConfig.MinServiceFee = 0
	d, _ := NewDelegationSystemSC(args)
	d.eei.SetStorage([]byte(ownerKey), []byte("address0"))

	vmInput := getDefaultVmInputForFunc("whitelistForMerge", [][]byte{[]byte("address")})

	enableEpochsHandler.IsValidatorToDelegationFlagEnabledField = false
	returnCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "whitelistForMerge"+" is an unknown function")

	enableEpochsHandler.IsValidatorToDelegationFlagEnabledField = true

	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "can be called by owner or the delegation manager")

	vmInput.CallerAddr = []byte("address0")
	vmInput.GasProvided = 0
	eei.gasRemaining = 0
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 1
	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, returnCode)

	vmInput.GasProvided = 1000
	eei.gasRemaining = vmInput.GasProvided
	vmInput.Arguments = [][]byte{}
	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "invalid number of arguments")

	vmInput.Arguments = [][]byte{[]byte("a")}
	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "invalid argument, wanted an address")

	vmInput.Arguments = [][]byte{[]byte("address0")}
	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "cannot whitelist own address")

	vmInput.Arguments = [][]byte{[]byte("address1")}
	vmInput.CallValue = big.NewInt(10)
	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "non-payable function")

	vmInput.CallValue = big.NewInt(0)
	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, returnCode)
	assert.Equal(t, []byte("address1"), d.eei.GetStorage([]byte(whitelistedAddress)))
}

func TestDelegation_deleteWhitelistForMerge(t *testing.T) {
	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}}
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)

	_ = eei.SetSystemSCContainer(systemSCContainerStub)

	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	args.DelegationSCConfig.MinServiceFee = 0
	d, _ := NewDelegationSystemSC(args)
	d.eei.SetStorage([]byte(ownerKey), []byte("address0"))

	vmInput := getDefaultVmInputForFunc("deleteWhitelistForMerge", [][]byte{[]byte("address")})

	enableEpochsHandler.IsValidatorToDelegationFlagEnabledField = false
	returnCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "deleteWhitelistForMerge"+" is an unknown function")

	enableEpochsHandler.IsValidatorToDelegationFlagEnabledField = true
	d.eei.SetStorage([]byte(ownerKey), []byte("address0"))
	vmInput.CallerAddr = []byte("address0")

	vmInput.GasProvided = 1000
	eei.gasRemaining = vmInput.GasProvided
	vmInput.Arguments = [][]byte{[]byte("a")}
	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "invalid number of arguments")

	d.eei.SetStorage([]byte(whitelistedAddress), []byte("address"))
	vmInput.Arguments = [][]byte{}
	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, returnCode)
	assert.Equal(t, 0, len(d.eei.GetStorage([]byte(whitelistedAddress))))

	d.eei.SetStorage([]byte(whitelistedAddress), []byte("address"))
	vmInput.Arguments = [][]byte{}
	vmInput.CallerAddr = vm.DelegationManagerSCAddress
	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, returnCode)
	assert.Equal(t, 0, len(d.eei.GetStorage([]byte(whitelistedAddress))))
}

func TestDelegation_GetWhitelistForMerge(t *testing.T) {
	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}}
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)

	_ = eei.SetSystemSCContainer(systemSCContainerStub)

	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	args.DelegationSCConfig.MinServiceFee = 0
	d, _ := NewDelegationSystemSC(args)
	d.eei.SetStorage([]byte(ownerKey), []byte("address0"))

	vmInput := getDefaultVmInputForFunc("getWhitelistForMerge", make([][]byte, 0))

	enableEpochsHandler.IsValidatorToDelegationFlagEnabledField = false
	returnCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "getWhitelistForMerge"+" is an unknown function")

	enableEpochsHandler.IsValidatorToDelegationFlagEnabledField = true

	addr := []byte("address1")
	vmInput = getDefaultVmInputForFunc("whitelistForMerge", [][]byte{addr})
	vmInput.CallValue = big.NewInt(0)
	vmInput.CallerAddr = []byte("address0")
	eei.returnMessage = ""
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, returnCode)

	vmInput = getDefaultVmInputForFunc("getWhitelistForMerge", make([][]byte, 0))
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, returnCode)
	require.Equal(t, 1, len(eei.output))
	assert.Equal(t, addr, eei.output[0])
}

func TestDelegation_OptimizeRewardsComputation(t *testing.T) {
	args := createMockArgumentsForDelegation()
	currentEpoch := uint32(2)
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return currentEpoch
		},
	}
	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}}

	_ = eei.SetSystemSCContainer(systemSCContainerStub)
	createDelegationManagerConfig(eei, args.Marshalizer, big.NewInt(10))

	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	args.DelegationSCConfig.MinServiceFee = 0
	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{})
	_ = d.saveDelegationContractConfig(&DelegationConfig{
		MaxDelegationCap:  big.NewInt(10000),
		InitialOwnerFunds: big.NewInt(1000),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive: big.NewInt(1000),
	})

	d.eei.SetStorage([]byte(ownerKey), []byte("address0"))

	delegator := []byte("delegator")
	_ = d.saveDelegatorData(delegator, &DelegatorData{
		ActiveFund:            nil,
		UnStakedFunds:         [][]byte{},
		UnClaimedRewards:      big.NewInt(1000),
		TotalCumulatedRewards: big.NewInt(0),
		RewardsCheckpoint:     0,
	})

	vmInput := getDefaultVmInputForFunc("updateRewards", [][]byte{})
	vmInput.CallValue = big.NewInt(20)
	vmInput.CallerAddr = vm.EndOfEpochAddress

	for i := 0; i < 10; i++ {
		currentEpoch++
		output := d.Execute(vmInput)
		assert.Equal(t, vmcommon.Ok, output)
	}

	vmInput = getDefaultVmInputForFunc("delegate", [][]byte{})
	vmInput.CallValue = big.NewInt(1000)
	vmInput.CallerAddr = delegator

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	currentEpoch++
	vmInput = getDefaultVmInputForFunc("updateRewards", [][]byte{})
	vmInput.CallValue = big.NewInt(20)
	vmInput.CallerAddr = vm.EndOfEpochAddress
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmInput = getDefaultVmInputForFunc("claimRewards", [][]byte{})
	vmInput.CallerAddr = delegator

	output = d.Execute(vmInput)
	fmt.Println(eei.returnMessage)
	assert.Equal(t, vmcommon.Ok, output)

	destAcc, exists := eei.outputAccounts[string(vmInput.CallerAddr)]
	assert.True(t, exists)
	_, exists = eei.outputAccounts[string(vmInput.RecipientAddr)]
	assert.True(t, exists)

	assert.Equal(t, 1, len(destAcc.OutputTransfers))
	outputTransfer := destAcc.OutputTransfers[0]
	assert.Equal(t, big.NewInt(1010), outputTransfer.Value)

	_, delegatorData, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
	assert.Equal(t, uint32(14), delegatorData.RewardsCheckpoint)
	assert.Equal(t, uint64(0), delegatorData.UnClaimedRewards.Uint64())
	assert.Equal(t, 1010, int(delegatorData.TotalCumulatedRewards.Uint64()))
}

func TestDelegation_AddTokens(t *testing.T) {
	args := createMockArgumentsForDelegation()
	eei := createDefaultEei()
	eei.inputParser = &mock.ArgumentParserMock{}
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	args.Eei = eei
	d, _ := NewDelegationSystemSC(args)

	vmInput := getDefaultVmInputForFunc("addTokens", [][]byte{})
	vmInput.CallValue = big.NewInt(20)
	vmInput.CallerAddr = vm.EndOfEpochAddress

	enableEpochsHandler.IsAddTokensToDelegationFlagEnabledField = false
	returnCode := d.Execute(vmInput)
	assert.Equal(t, returnCode, vmcommon.UserError)
	assert.Equal(t, eei.returnMessage, vmInput.Function+" is an unknown function")

	eei.returnMessage = ""
	enableEpochsHandler.IsAddTokensToDelegationFlagEnabledField = true
	returnCode = d.Execute(vmInput)
	assert.Equal(t, returnCode, vmcommon.UserError)
	assert.Equal(t, eei.returnMessage, vmInput.Function+" can be called by whitelisted address only")

	vmInput.CallerAddr = args.AddTokensAddress
	returnCode = d.Execute(vmInput)
	assert.Equal(t, returnCode, vmcommon.Ok)
}

func TestDelegation_correctNodesStatus(t *testing.T) {
	d, eei := createDelegationContractAndEEI()
	vmInput := getDefaultVmInputForFunc("correctNodesStatus", nil)

	enableEpochsHandler, _ := d.enableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.IsAddTokensToDelegationFlagEnabledField = false
	returnCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "correctNodesStatus is an unknown function")

	enableEpochsHandler.IsAddTokensToDelegationFlagEnabledField = true
	eei.returnMessage = ""
	vmInput.CallValue.SetUint64(10)
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "call value must be zero")

	eei.returnMessage = ""
	eei.gasRemaining = 1
	d.gasCost.MetaChainSystemSCsCost.GetAllNodeStates = 10
	vmInput.CallValue.SetUint64(0)
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, returnCode)

	eei.returnMessage = ""
	eei.gasRemaining = 11
	vmInput.CallValue.SetUint64(0)
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "data was not found under requested key delegation status")

	wrongStatus := &DelegationContractStatus{
		StakedKeys:    []*NodesData{{BLSKey: []byte("key1")}, {BLSKey: []byte("key2")}, {BLSKey: []byte("key3")}},
		NotStakedKeys: []*NodesData{{BLSKey: []byte("key4")}, {BLSKey: []byte("key5")}, {BLSKey: []byte("key3")}},
		UnStakedKeys:  []*NodesData{{BLSKey: []byte("key6")}, {BLSKey: []byte("key7")}, {BLSKey: []byte("key3")}},
		NumUsers:      0,
	}
	_ = d.saveDelegationStatus(wrongStatus)

	stakedKeys := [][]byte{[]byte("key1"), []byte("key4"), []byte("key7")}
	unStakedKeys := [][]byte{[]byte("key2"), []byte("key6")}
	for i, blsKey := range stakedKeys {
		stakedData := &StakedDataV2_0{
			Staked: true,
		}
		if i == 2 {
			stakedData.Staked = false
			stakedData.Jailed = true
		}
		marshaledData, _ := d.marshalizer.Marshal(stakedData)
		eei.SetStorageForAddress(d.stakingSCAddr, blsKey, marshaledData)
	}

	for _, blsKey := range unStakedKeys {
		stakedData := &StakedDataV2_0{
			Staked: false,
		}
		marshaledData, _ := d.marshalizer.Marshal(stakedData)
		eei.SetStorageForAddress(d.stakingSCAddr, blsKey, marshaledData)
	}

	eei.returnMessage = ""
	eei.gasRemaining = 11
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "storage is nil for given key")

	validatorData := &ValidatorDataV2{BlsPubKeys: [][]byte{[]byte("key8")}}
	marshaledData, _ := d.marshalizer.Marshal(validatorData)
	eei.SetStorageForAddress(d.validatorSCAddr, vmInput.RecipientAddr, marshaledData)

	stakedData := &StakedDataV2_0{
		Staked: false,
		Jailed: true,
	}
	marshaledData, _ = d.marshalizer.Marshal(stakedData)
	eei.SetStorageForAddress(d.stakingSCAddr, []byte("key8"), marshaledData)
	stakedKeys = append(stakedKeys, []byte("key8"))

	eei.returnMessage = ""
	eei.gasRemaining = 11
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, returnCode)

	correctedStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 4, len(correctedStatus.StakedKeys))
	assert.Equal(t, 2, len(correctedStatus.UnStakedKeys))
	assert.Equal(t, 2, len(correctedStatus.NotStakedKeys))

	for _, stakedKey := range stakedKeys {
		found := false
		for _, stakedNode := range correctedStatus.StakedKeys {
			if bytes.Equal(stakedNode.BLSKey, stakedKey) {
				found = true
				break
			}
		}
		assert.True(t, found)
	}

	for _, unStakedKey := range unStakedKeys {
		found := false
		for _, unStakedNode := range correctedStatus.UnStakedKeys {
			if bytes.Equal(unStakedNode.BLSKey, unStakedKey) {
				found = true
				break
			}
		}
		assert.True(t, found)
	}

	notStakedKeys := [][]byte{[]byte("key3"), []byte("key5")}
	for _, notStakedKey := range notStakedKeys {
		found := false
		for _, notStakedNode := range correctedStatus.NotStakedKeys {
			if bytes.Equal(notStakedNode.BLSKey, notStakedKey) {
				found = true
				break
			}
		}
		assert.True(t, found)
	}
}

func createDefaultEei() *vmContext {
	eei, _ := NewVMContext(createDefaultEeiArgs())

	return eei
}

func createDefaultEeiArgs() VMContextArgs {
	return VMContextArgs{
		BlockChainHook:      &mock.BlockChainHookStub{},
		CryptoHook:          hooks.NewVMCryptoHook(),
		InputParser:         parsers.NewCallArgsParser(),
		ValidatorAccountsDB: &stateMock.AccountsStub{},
		UserAccountsDB:      &stateMock.AccountsStub{},
		ChanceComputer:      &mock.RaterMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsMultiClaimOnDelegationEnabledField: true,
		},
		ShardCoordinator: &mock.ShardCoordinatorStub{},
	}
}

func TestDelegationSystemSC_ExecuteChangeOwnerUserErrors(t *testing.T) {
	t.Parallel()

	vmInputArgs := make([][]byte, 0)
	args := createMockArgumentsForDelegation()
	argsVmContext := VMContextArgs{
		BlockChainHook:      &mock.BlockChainHookStub{},
		CryptoHook:          hooks.NewVMCryptoHook(),
		InputParser:         &mock.ArgumentParserMock{},
		ValidatorAccountsDB: &stateMock.AccountsStub{},
		UserAccountsDB:      &stateMock.AccountsStub{},
		ChanceComputer:      &mock.RaterMock{},
		EnableEpochsHandler: args.EnableEpochsHandler,
		ShardCoordinator:    &mock.ShardCoordinatorStub{},
	}
	eei, err := NewVMContext(argsVmContext)
	require.Nil(t, err)

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub).IsChangeDelegationOwnerFlagEnabledField = false
	vmInput := getDefaultVmInputForFunc("changeOwner", vmInputArgs)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vmInput.Function+" is an unknown function"))

	args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub).IsChangeDelegationOwnerFlagEnabledField = true
	vmInput.CallValue = big.NewInt(0)
	vmInput.CallerAddr = []byte("aaa")
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	eei.returnMessage = ""
	vmInput.CallerAddr = delegationsMap[ownerKey]
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "wrong number of arguments, expected 1"))

	eei.returnMessage = ""
	vmInput.Arguments = append(vmInput.Arguments, []byte("aaa"))
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid argument, wanted an address"))

	eei.returnMessage = ""
	vmInput.Arguments[0] = []byte("second123")
	delegationMgrMap := map[string][]byte{}
	delegationMgrMap["second123"] = []byte("info")
	eei.storageUpdate[string(d.delegationMgrSCAddress)] = delegationMgrMap
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "destination already deployed a delegation sc"))

	eei.storageUpdate[string(d.delegationMgrSCAddress)] = map[string][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "owner is new delegator"))

	marshalledData, _ := d.marshalizer.Marshal(&DelegatorData{RewardsCheckpoint: 10})
	delegationsMap["second123"] = marshalledData
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "destination should be a new account"))
}

func TestDelegationSystemSC_ExecuteChangeOwnerWithoutAccountUpdate(t *testing.T) {
	t.Parallel()

	vmInputArgs := make([][]byte, 0)
	args := createMockArgumentsForDelegation()
	epochHandler := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	epochHandler.IsMultiClaimOnDelegationEnabledField = false
	argsVmContext := VMContextArgs{
		BlockChainHook:      &mock.BlockChainHookStub{},
		CryptoHook:          hooks.NewVMCryptoHook(),
		InputParser:         &mock.ArgumentParserMock{},
		ValidatorAccountsDB: &stateMock.AccountsStub{},
		UserAccountsDB:      &stateMock.AccountsStub{},
		ChanceComputer:      &mock.RaterMock{},
		EnableEpochsHandler: args.EnableEpochsHandler,
		ShardCoordinator:    &mock.ShardCoordinatorStub{},
	}
	args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub).IsChangeDelegationOwnerFlagEnabledField = true
	eei, err := NewVMContext(argsVmContext)
	require.Nil(t, err)

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	marshalledData, _ := args.Marshalizer.Marshal(&DelegatorData{RewardsCheckpoint: 10})
	delegationsMap["ownerAddr"] = marshalledData
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	vmInput := getDefaultVmInputForFunc("changeOwner", vmInputArgs)
	vmInput.CallValue = big.NewInt(0)
	vmInput.CallerAddr = delegationsMap[ownerKey]
	vmInput.Arguments = append(vmInput.Arguments, []byte("second123"))

	returnCode := d.Execute(vmInput)
	assert.Equal(t, returnCode, vmcommon.Ok)
	assert.Equal(t, delegationsMap[ownerKey], []byte("second123"))
	assert.Equal(t, eei.storageUpdate[string(d.delegationMgrSCAddress)]["ownerAddr"], []byte{})
	assert.Equal(t, eei.storageUpdate[string(d.delegationMgrSCAddress)]["second123"], vmInput.RecipientAddr)
	returnCode = d.Execute(vmInput)
	assert.Equal(t, returnCode, vmcommon.UserError)

	assert.Len(t, eei.logs, 3)
	assert.Equal(t, []byte("delegate"), eei.logs[0].Identifier)
	assert.Equal(t, []byte("second123"), eei.logs[0].Address)

	assert.Equal(t, []byte(withdraw), eei.logs[1].Identifier)
	assert.Equal(t, []byte("ownerAddr"), eei.logs[1].Address)
	assert.Equal(t, boolToSlice(true), eei.logs[1].Topics[4])

	assert.Equal(t, []byte(core.BuiltInFunctionChangeOwnerAddress), eei.logs[2].Identifier)
	assert.Equal(t, []byte("addr"), eei.logs[2].Address)
	assert.Equal(t, []byte("second123"), eei.logs[2].Topics[0])

	eei.logs = nil
	vmInput.CallerAddr = []byte("second123")
	vmInput.Arguments[0] = []byte("ownerAddr")
	returnCode = d.Execute(vmInput)
	assert.Equal(t, returnCode, vmcommon.Ok)
	assert.Equal(t, delegationsMap[ownerKey], []byte("ownerAddr"))

	assert.Equal(t, eei.storageUpdate[string(d.delegationMgrSCAddress)]["ownerAddr"], vmInput.RecipientAddr)
	assert.Equal(t, eei.storageUpdate[string(d.delegationMgrSCAddress)]["second123"], []byte{})

	assert.Len(t, eei.logs, 3)
	assert.Equal(t, []byte("delegate"), eei.logs[0].Identifier)
	assert.Equal(t, []byte("ownerAddr"), eei.logs[0].Address)

	assert.Equal(t, []byte(withdraw), eei.logs[1].Identifier)
	assert.Equal(t, []byte("second123"), eei.logs[1].Address)
	assert.Equal(t, boolToSlice(true), eei.logs[1].Topics[4])

	assert.Equal(t, []byte(core.BuiltInFunctionChangeOwnerAddress), eei.logs[2].Identifier)
	assert.Equal(t, []byte("addr"), eei.logs[2].Address)
	assert.Equal(t, []byte("ownerAddr"), eei.logs[2].Topics[0])
}

func TestDelegationSystemSC_ExecuteChangeOwnerWithAccountUpdate(t *testing.T) {
	t.Parallel()

	vmInputArgs := make([][]byte, 0)
	args := createMockArgumentsForDelegation()
	epochHandler := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	epochHandler.FixDelegationChangeOwnerOnAccountEnabledField = true
	account := &stateMock.AccountWrapMock{}
	argsVmContext := VMContextArgs{
		BlockChainHook:      &mock.BlockChainHookStub{},
		CryptoHook:          hooks.NewVMCryptoHook(),
		InputParser:         &mock.ArgumentParserMock{},
		ValidatorAccountsDB: &stateMock.AccountsStub{},
		UserAccountsDB: &stateMock.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
				return account, nil
			},
		},
		ChanceComputer:      &mock.RaterMock{},
		EnableEpochsHandler: args.EnableEpochsHandler,
		ShardCoordinator:    &mock.ShardCoordinatorStub{},
	}
	args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub).IsChangeDelegationOwnerFlagEnabledField = true
	eei, err := NewVMContext(argsVmContext)
	require.Nil(t, err)

	vmInput := getDefaultVmInputForFunc("changeOwner", vmInputArgs)
	ownerAddress := bytes.Repeat([]byte{1}, len(vmInput.RecipientAddr))
	newOwnerAddress := bytes.Repeat([]byte{2}, len(vmInput.RecipientAddr))
	scAddress := bytes.Repeat([]byte{1}, len(vmInput.RecipientAddr))
	eei.SetSCAddress(scAddress)

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = ownerAddress
	marshalledData, _ := args.Marshalizer.Marshal(&DelegatorData{RewardsCheckpoint: 10})
	delegationsMap[string(ownerAddress)] = marshalledData
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSC(args)
	vmInput.CallValue = big.NewInt(0)
	vmInput.CallerAddr = delegationsMap[ownerKey]
	vmInput.Arguments = append(vmInput.Arguments, newOwnerAddress)

	returnCode := d.Execute(vmInput)
	assert.Equal(t, returnCode, vmcommon.Ok)
	assert.Equal(t, newOwnerAddress, delegationsMap[ownerKey])
	assert.Equal(t, eei.storageUpdate[string(d.delegationMgrSCAddress)][string(ownerAddress)], []byte{})
	assert.Equal(t, eei.storageUpdate[string(d.delegationMgrSCAddress)][string(newOwnerAddress)], vmInput.RecipientAddr)
	assert.Equal(t, newOwnerAddress, account.Owner)
}

func TestDelegationSystemSC_SynchronizeOwner(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegation()
	epochHandler := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	epochHandler.FixDelegationChangeOwnerOnAccountEnabledField = false

	account := &stateMock.AccountWrapMock{}

	argsVmContext := VMContextArgs{
		BlockChainHook:      &mock.BlockChainHookStub{},
		CryptoHook:          hooks.NewVMCryptoHook(),
		InputParser:         &mock.ArgumentParserMock{},
		ValidatorAccountsDB: &stateMock.AccountsStub{},
		UserAccountsDB: &stateMock.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
				return account, nil
			},
		},
		ChanceComputer:      &mock.RaterMock{},
		EnableEpochsHandler: args.EnableEpochsHandler,
		ShardCoordinator:    &mock.ShardCoordinatorStub{},
	}
	eei, err := NewVMContext(argsVmContext)
	require.Nil(t, err)

	delegationsMap := map[string][]byte{}
	ownerAddress := []byte("1111")
	scAddress := bytes.Repeat([]byte{1}, len(ownerAddress))
	eei.SetSCAddress(scAddress)
	delegationsMap[ownerKey] = ownerAddress
	marshalledData, _ := args.Marshalizer.Marshal(&DelegatorData{RewardsCheckpoint: 10})
	delegationsMap[string(ownerAddress)] = marshalledData
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	vmInputArgs := make([][]byte, 0)

	d, _ := NewDelegationSystemSC(args)

	// do not run these tests in parallel
	t.Run("function is disabled", func(t *testing.T) {
		vmInput := getDefaultVmInputForFunc("synchronizeOwner", vmInputArgs)
		returnCode := d.Execute(vmInput)
		assert.Equal(t, vmcommon.UserError, returnCode)
		assert.Equal(t, "synchronizeOwner is an unknown function", eei.GetReturnMessage())
	})

	epochHandler.FixDelegationChangeOwnerOnAccountEnabledField = true
	eei.ResetReturnMessage()

	t.Run("transfer value is not zero", func(t *testing.T) {
		vmInput := getDefaultVmInputForFunc("synchronizeOwner", vmInputArgs)
		vmInput.CallValue = big.NewInt(1)
		returnCode := d.Execute(vmInput)
		assert.Equal(t, vmcommon.UserError, returnCode)
		assert.Equal(t, vm.ErrCallValueMustBeZero.Error(), eei.GetReturnMessage())
		eei.ResetReturnMessage()
	})
	t.Run("wrong arguments", func(t *testing.T) {
		vmInput := getDefaultVmInputForFunc("synchronizeOwner", [][]byte{[]byte("argument")})
		returnCode := d.Execute(vmInput)
		assert.Equal(t, vmcommon.UserError, returnCode)
		assert.Equal(t, "invalid number of arguments, expected 0", eei.GetReturnMessage())
		eei.ResetReturnMessage()
	})
	t.Run("wrong stored address", func(t *testing.T) {
		vmInput := getDefaultVmInputForFunc("synchronizeOwner", vmInputArgs)
		eei.SetSCAddress(scAddress[:1])
		returnCode := d.Execute(vmInput)
		assert.Equal(t, vmcommon.UserError, returnCode)
		assert.Equal(t, "wrong new owner address", eei.GetReturnMessage())
		assert.Equal(t, 0, len(account.Owner))
		eei.ResetReturnMessage()
	})
	t.Run("should work", func(t *testing.T) {
		vmInput := getDefaultVmInputForFunc("synchronizeOwner", vmInputArgs)
		eei.SetSCAddress(scAddress)
		returnCode := d.Execute(vmInput)
		assert.Equal(t, vmcommon.Ok, returnCode)
		assert.Equal(t, "", eei.GetReturnMessage())
		assert.Equal(t, ownerAddress, account.Owner)
		eei.ResetReturnMessage()
	})
}
