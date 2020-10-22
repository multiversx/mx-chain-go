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
		DelegationSCConfig: config.DelegationSystemSCConfig{
			MinStakeAmount: "10",
			MinServiceFee:  10,
			MaxServiceFee:  200,
		},
		StakingSCConfig: config.StakingSystemSCConfig{
			MinStakeValue:    "10",
			UnJailValue:      "15",
			GenesisNodePrice: "100",
		},
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
			auctionSc.flagEnableTopUp.Set()
			_ = auctionSc.saveRegistrationData([]byte("addr"), &AuctionDataV2{
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
			})
			auctionSc.unBondPeriod = 50
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

	delegatorDataPresent, delegator, err := d.getOrCreateDelegatorData(ownerAddr)
	assert.Nil(t, err)
	assert.False(t, delegatorDataPresent)
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
	_ = d.saveGlobalFundData(&GlobalFundData{TotalUnStakedFromNodes: big.NewInt(0)})
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
	_ = d.saveGlobalFundData(&GlobalFundData{TotalUnBondedFromNodes: big.NewInt(0)})
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

	key1 := &NodesData{BLSKey: blsKey1}
	key2 := &NodesData{BLSKey: blsKey2}
	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		StakedKeys:   []*NodesData{key1},
		UnStakedKeys: []*NodesData{key2},
		Delegators:   [][]byte{vmInput.CallerAddr},
	})
	addAuctionAndStakingScToVmContext(eei)

	auctionMap := map[string][]byte{}
	registrationDataAuction := &AuctionDataV2{
		BlsPubKeys:    [][]byte{blsKey1, blsKey2},
		RewardAddress: []byte("rewardAddr"),
	}
	regData, _ := d.marshalizer.Marshal(registrationDataAuction)
	auctionMap["addr"] = regData

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

	eei.storageUpdate[string(args.AuctionSCAddress)] = auctionMap
	eei.storageUpdate["staking"] = stakingMap

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

	vmInput := getDefaultVmInputForFunc("delegate", [][]byte{})
	vmInput.CallValue = big.NewInt(15)
	d, _ := NewDelegationSystemSC(args)

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
	addAuctionAndStakingScToVmContext(eei)

	vmInput := getDefaultVmInputForFunc("delegate", [][]byte{})
	vmInput.CallValue = big.NewInt(15)
	vmInput.CallerAddr = delegator1
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegationStatus(&DelegationContractStatus{})
	_ = d.saveDelegationContractConfig(&DelegationConfig{
		WithDelegationCap: true,
		MaxDelegationCap:  big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive: big.NewInt(100),
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "total delegation cap reached, no more space to accept"))

	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive: big.NewInt(0),
	})

	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dFund, _ := d.getFund([]byte{})
	assert.Equal(t, big.NewInt(15), dFund.Value)
	assert.Equal(t, delegator1, dFund.Address)
	assert.Equal(t, active, dFund.Type)

	dGlobalFund, _ := d.getGlobalFundData()
	assert.Equal(t, big.NewInt(15), dGlobalFund.TotalActive)
	assert.Equal(t, []byte{}, dGlobalFund.ActiveFunds[0])

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 1, len(dStatus.Delegators))
	assert.Equal(t, delegator1, dStatus.Delegators[0])

	_, dData, _ := d.getOrCreateDelegatorData(delegator1)
	assert.Equal(t, []byte{}, dData.ActiveFund)
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

func TestDelegationSystemSC_ExecuteUnDelegateUserNotDelegatorOrNoActiveFundShouldErr(t *testing.T) {
	t.Parallel()

	fundKey := []byte{1}
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei

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
		Value: big.NewInt(5),
	})
	vmInput.Arguments = [][]byte{{5}}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid value to undelegate - need to undelegate all - do not leave dust behind"))
}

func TestDelegationSystemSC_ExecuteUnDelegatePartOfFunds(t *testing.T) {
	t.Parallel()

	fundKey := []byte{1}
	nextFundKey := []byte{2}
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei
	addAuctionAndStakingScToVmContext(eei)

	vmInput := getDefaultVmInputForFunc("unDelegate", [][]byte{{80}})
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegatorData(vmInput.CallerAddr, &DelegatorData{
		ActiveFund:    fundKey,
		UnStakedFunds: [][]byte{},
	})
	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		UnStakedFunds:          [][]byte{},
		TotalActive:            big.NewInt(100),
		TotalUnStaked:          big.NewInt(0),
		TotalUnStakedFromNodes: big.NewInt(0),
	})
	d.eei.SetStorage([]byte(lastFundKey), []byte{1})

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
}

func TestDelegationSystemSC_ExecuteUnDelegateAllFunds(t *testing.T) {
	t.Parallel()

	fundKey := []byte{1}
	nextFundKey := []byte{2}
	args := createMockArgumentsForDelegation()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{},
	)
	args.Eei = eei
	addAuctionAndStakingScToVmContext(eei)

	vmInput := getDefaultVmInputForFunc("unDelegate", [][]byte{{100}})
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegatorData(vmInput.CallerAddr, &DelegatorData{
		ActiveFund:    fundKey,
		UnStakedFunds: [][]byte{},
	})
	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		UnStakedFunds:          [][]byte{},
		TotalActive:            big.NewInt(100),
		TotalUnStaked:          big.NewInt(0),
		TotalUnStakedFromNodes: big.NewInt(0),
		ActiveFunds:            [][]byte{fundKey},
	})
	d.eei.SetStorage([]byte(lastFundKey), []byte{1})

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
	addAuctionAndStakingScToVmContext(eei)

	vmInput := getDefaultVmInputForFunc("withdraw", [][]byte{})
	d, _ := NewDelegationSystemSC(args)

	_ = d.saveDelegatorData(vmInput.CallerAddr, &DelegatorData{
		UnStakedFunds: [][]byte{fundKey1, fundKey2},
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
		UnStakedFunds:          [][]byte{fundKey1, fundKey2},
		TotalUnBondedFromNodes: big.NewInt(0),
		TotalUnStaked:          big.NewInt(140),
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
	vmInput.Arguments = [][]byte{[]byte("service fee")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid new service fee"))

	vmInput.Arguments = [][]byte{[]byte("5")}
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

	vmInput := getDefaultVmInputForFunc("changeServiceFee", [][]byte{[]byte("70")})
	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationContractConfig(&DelegationConfig{})
	_ = d.saveGlobalFundData(&GlobalFundData{})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dConfig, _ := d.getDelegationContractConfig()
	assert.Equal(t, uint64(70), dConfig.ServiceFee)
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
	vmInput.Arguments = [][]byte{[]byte("new delegation cap")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid new total delegation cap"))

	vmInput.Arguments = [][]byte{[]byte("70")}
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

	vmInput := getDefaultVmInputForFunc("modifyTotalDelegationCap", [][]byte{[]byte("500")})
	d, _ := NewDelegationSystemSC(args)
	_ = d.saveDelegationContractConfig(&DelegationConfig{})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive: big.NewInt(1000),
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot make total delegation cap smaller than active"))

	vmInput.Arguments = [][]byte{[]byte("1500")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dConfig, _ := d.getDelegationContractConfig()
	assert.Equal(t, big.NewInt(1500), dConfig.MaxDelegationCap)
	assert.Equal(t, true, dConfig.WithDelegationCap)
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
