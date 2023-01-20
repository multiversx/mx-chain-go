package systemSmartContracts

import (
	"bytes"
	"errors"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func createMockGovernanceArgs() ArgsNewGovernanceContract {
	eei := createEEIWithBlockchainHook(&mock.BlockChainHookStub{CurrentEpochCalled: func() uint32 {
		return 2
	}})

	return createArgsWithEEI(eei)
}

func createArgsWithEEI(eei vm.SystemEI) ArgsNewGovernanceContract {
	return ArgsNewGovernanceContract{
		Eei:     eei,
		GasCost: vm.GasCost{},
		GovernanceConfig: config.GovernanceSystemSCConfig{
			V1: config.GovernanceSystemSCConfigV1{
				NumNodes:         3,
				MinPassThreshold: 1,
				MinQuorum:        2,
				MinVetoThreshold: 2,
				ProposalCost:     "500",
			},
			Active: config.GovernanceSystemSCConfigActive{
				ProposalCost:     "500",
				MinQuorum:        0.5,
				MinPassThreshold: 0.5,
				MinVetoThreshold: 0.5,
			},
			ChangeConfigAddress: "erd1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awqgqky80",
		},
		Marshalizer:            &mock.MarshalizerMock{},
		Hasher:                 &hashingMocks.HasherMock{},
		GovernanceSCAddress:    vm.GovernanceSCAddress,
		DelegationMgrSCAddress: vm.DelegationManagerSCAddress,
		ValidatorSCAddress:     vm.ValidatorSCAddress,
		ConfigChangeAddress:    bytes.Repeat([]byte{1}, 32),
		UnBondPeriodInEpochs:   10,
		EnableEpochsHandler: &testscommon.EnableEpochsHandlerStub{
			IsGovernanceFlagEnabledField: true,
		},
	}
}

func createEEIWithBlockchainHook(blockchainHook vm.BlockchainHook) vm.ContextHandler {
	eei, _ := NewVMContext(VMContextArgs{
		BlockChainHook:      blockchainHook,
		CryptoHook:          hooks.NewVMCryptoHook(),
		InputParser:         &mock.ArgumentParserMock{},
		ValidatorAccountsDB: &stateMock.AccountsStub{},
		ChanceComputer:      &mock.RaterMock{},
		EnableEpochsHandler: &testscommon.EnableEpochsHandlerStub{},
	})
	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}}
	_ = eei.SetSystemSCContainer(systemSCContainerStub)

	return eei
}

func createGovernanceBlockChainHookStubContextHandler() (*governanceContract, *mock.BlockChainHookStub, vm.ContextHandler) {
	blockChainHook := &mock.BlockChainHookStub{CurrentEpochCalled: func() uint32 {
		return 2
	}}
	eei := createEEIWithBlockchainHook(blockChainHook)
	gsc, _ := NewGovernanceContract(createArgsWithEEI(eei))
	gsc.initV2(&vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallerAddr: gsc.governanceSCAddress}})

	addressList := [][]byte{vm.FirstDelegationSCAddress, vm.StakingSCAddress}
	marshaledData, _ := gsc.marshalizer.Marshal(&DelegationContractList{addressList})

	gsc.eei.SetStorageForAddress(gsc.delegationMgrSCAddress, []byte(delegationContractsList), marshaledData)
	_ = saveDelegationManagementData(eei, gsc.marshalizer, gsc.delegationMgrSCAddress, &DelegationManagement{MinDelegationAmount: big.NewInt(10)})

	userAddress := bytes.Repeat([]byte{2}, 32)
	addStakeAndDelegationForAddress(gsc, userAddress)

	return gsc, blockChainHook, eei
}

func addStakeAndDelegationForAddress(gsc *governanceContract, userAddress []byte) {
	marshaledData, _ := gsc.marshalizer.Marshal(&ValidatorDataV2{TotalStakeValue: big.NewInt(100)})
	gsc.eei.SetStorageForAddress(gsc.validatorSCAddress, userAddress, marshaledData)

	addressList, _ := getDelegationContractList(gsc.eei, gsc.marshalizer, gsc.delegationMgrSCAddress)

	for index, delegationAddress := range addressList.Addresses {
		fundKey := append([]byte(fundKeyPrefix), big.NewInt(int64(index)).Bytes()...)

		marshaledData, _ = gsc.marshalizer.Marshal(&DelegatorData{ActiveFund: fundKey})
		gsc.eei.SetStorageForAddress(delegationAddress, userAddress, marshaledData)

		marshaledData, _ = gsc.marshalizer.Marshal(&Fund{Value: big.NewInt(10)})
		gsc.eei.SetStorageForAddress(delegationAddress, fundKey, marshaledData)
	}
}

func createVMInput(callValue *big.Int, funcName string, callerAddr, recipientAddr []byte, arguments [][]byte) *vmcommon.ContractCallInput {
	return &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:  callValue,
			CallerAddr: callerAddr,
			Arguments:  arguments,
		},
		Function:      funcName,
		RecipientAddr: recipientAddr,
	}
}

func TestNewGovernanceContract_NilEeiShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.Eei = nil

	gsc, err := NewGovernanceContract(args)
	require.Nil(t, gsc)
	require.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
}

func TestNewGovernanceContract_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.Marshalizer = nil

	gsc, err := NewGovernanceContract(args)
	require.Nil(t, gsc)
	require.Equal(t, vm.ErrNilMarshalizer, err)
}

func TestNewGovernanceContract_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.Hasher = nil

	gsc, err := NewGovernanceContract(args)
	require.Nil(t, gsc)
	require.Equal(t, vm.ErrNilHasher, err)
}

func TestNewGovernanceContract_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.EnableEpochsHandler = nil

	gsc, err := NewGovernanceContract(args)
	require.Nil(t, gsc)
	require.Equal(t, vm.ErrNilEnableEpochsHandler, err)
}

func TestNewGovernanceContract_ZeroBaseProposerCostShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.GovernanceConfig.V1.ProposalCost = ""

	gsc, err := NewGovernanceContract(args)
	require.Nil(t, gsc)
	require.Equal(t, vm.ErrInvalidBaseIssuingCost, err)
}

func TestNewGovernanceContract_InvalidValidatorAddress(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.ValidatorSCAddress = nil

	gsc, err := NewGovernanceContract(args)
	require.Nil(t, gsc)
	require.True(t, errors.Is(err, vm.ErrInvalidAddress))
}

func TestNewGovernanceContract_InvalidDelegationMgrAddress(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.GovernanceSCAddress = nil

	gsc, err := NewGovernanceContract(args)
	require.Nil(t, gsc)
	require.True(t, errors.Is(err, vm.ErrInvalidAddress))
}

func TestNewGovernanceContract_InvalidGovernanceAddress(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.DelegationMgrSCAddress = nil

	gsc, err := NewGovernanceContract(args)
	require.Nil(t, gsc)
	require.True(t, errors.Is(err, vm.ErrInvalidAddress))
}

func TestGovernanceContract_ExecuteNilVMInputShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)

	retCode := gsc.Execute(nil)
	require.Equal(t, vmcommon.UserError, retCode)

	callInput := createVMInput(big.NewInt(0), "unknownFunction", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.FunctionNotFound, retCode)
}

func TestGovernanceContract_ExecuteInit(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)

	callerAddr := []byte("addr1")
	callInput := createVMInput(big.NewInt(0), core.SCDeployInitFunctionName, callerAddr, vm.GovernanceSCAddress, nil)

	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, gsc.ownerAddress, callerAddr)
}

func TestGovernanceContract_ExecuteInitV2InvalidCaller(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)
	callerAddr := []byte("addr1")
	callInput := createVMInput(big.NewInt(0), "initV2", callerAddr, vm.GovernanceSCAddress, nil)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_ExecuteInitV2InvalidConfig(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.GovernanceConfig.Active.MinQuorum = 0.0
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "initV2", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_ExecuteInitV2MarshalError(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.Marshalizer = &mock.MarshalizerMock{
		Fail: true,
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "initV2", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_ExecuteInitV2(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*testscommon.EnableEpochsHandlerStub)
	gsc, _ := NewGovernanceContract(args)

	callInput := createVMInput(big.NewInt(0), "initV2", vm.GovernanceSCAddress, []byte("addr2"), nil)

	enableEpochsHandler.IsGovernanceFlagEnabledField = false
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)

	enableEpochsHandler.IsGovernanceFlagEnabledField = true

	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, gsc.ownerAddress, vm.GovernanceSCAddress)
}

func TestGovernanceContract_ChangeConfig(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 1
				},
			}
		},
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, []byte(governanceConfigKey)) {
				configBytes, _ := args.Marshalizer.Marshal(&GovernanceConfigV2{})
				return configBytes
			}

			return nil
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		[]byte("1"),
		[]byte("10"),
		[]byte("10"),
		[]byte("5"),
	}
	initInput := createVMInput(zero, "initV2", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
	_ = gsc.Execute(initInput)
	callInput := createVMInput(zero, "changeConfig", args.ConfigChangeAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_ChangeConfigWrongCaller(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "changeConfig can be called only by owner"
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)
	initInput := createVMInput(zero, "initV2", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
	_ = gsc.Execute(initInput)
	callInput := createVMInput(zero, "changeConfig", []byte("wrong caller"), vm.GovernanceSCAddress, nil)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_ChangeConfigWrongCallValue(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "changeConfig can be called only without callValue"
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)

	initInput := createVMInput(zero, "initV2", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
	_ = gsc.Execute(initInput)
	callInput := createVMInput(big.NewInt(10), "changeConfig", args.ConfigChangeAddress, vm.GovernanceSCAddress, nil)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_ChangeConfigWrongArgumentsLength(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "changeConfig needs 4 arguments"
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)

	initInput := createVMInput(zero, "initV2", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
	_ = gsc.Execute(initInput)
	callInput := createVMInput(zero, "changeConfig", args.ConfigChangeAddress, vm.GovernanceSCAddress, nil)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_ChangeConfigInvalidParams(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "changeConfig first argument is incorrectly formatted"
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)

	initInput := createVMInput(zero, "initV2", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
	_ = gsc.Execute(initInput)

	callInputArgs := [][]byte{
		[]byte("invalid"),
		[]byte("10"),
		[]byte("10"),
		[]byte("5"),
	}
	callInput := createVMInput(zero, "changeConfig", args.ConfigChangeAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)

	errSubstr = "changeConfig second argument is incorrectly formatted"
	callInputArgs = [][]byte{
		[]byte("1"),
		[]byte("invalid"),
		[]byte("10"),
		[]byte("5"),
	}
	callInput = createVMInput(zero, "changeConfig", args.ConfigChangeAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode = gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)

	errSubstr = "changeConfig third argument is incorrectly formatted"
	callInputArgs = [][]byte{
		[]byte("1"),
		[]byte("10"),
		[]byte("invalid"),
		[]byte("5"),
	}
	callInput = createVMInput(zero, "changeConfig", args.ConfigChangeAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode = gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)

	errSubstr = "changeConfig fourth argument is incorrectly formatted"
	callInputArgs = [][]byte{
		[]byte("1"),
		[]byte("10"),
		[]byte("10"),
		[]byte("invalid"),
	}
	callInput = createVMInput(zero, "changeConfig", args.ConfigChangeAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode = gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_ChangeConfigGetConfigErr(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "changeConfig error"
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, []byte(governanceConfigKey)) {
				return []byte("invalid config")
			}

			return nil
		},
	}

	gsc, _ := NewGovernanceContract(args)

	initInput := createVMInput(zero, "initV2", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
	_ = gsc.Execute(initInput)

	callInputArgs := [][]byte{
		[]byte("1"),
		[]byte("10"),
		[]byte("10"),
		[]byte("5"),
	}
	callInput := createVMInput(zero, "changeConfig", args.ConfigChangeAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_ProposalNotEnoughGas(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		UseGasCalled: func(gas uint64) error {
			return errors.New("not enough gas")
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), nil)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.OutOfGas, retCode)
}

func TestGovernanceContract_ProposalInvalidArgumentsLength(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), nil)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)
}

func TestGovernanceContract_ProposalWrongCallValue(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()

	gsc, _ := NewGovernanceContract(args)

	callInput := createVMInput(big.NewInt(9), "proposal", vm.GovernanceSCAddress, []byte("addr1"), [][]byte{{1}, {1}, {1}})
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)

	gsc.initV2(&vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallerAddr: gsc.governanceSCAddress}})
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.OutOfFunds, retCode)
}

func TestGovernanceContract_ProposalInvalidReferenceLength(t *testing.T) {
	t.Parallel()

	gsc, _, eei := createGovernanceBlockChainHookStubContextHandler()
	callInputArgs := [][]byte{
		[]byte("arg1"),
		[]byte("arg2"),
		[]byte("arg3"),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "invalid github commit"))
}

func TestGovernanceContract_ProposalAlreadyExists(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	gsc, _, eei := createGovernanceBlockChainHookStubContextHandler()
	callInputArgs := [][]byte{
		proposalIdentifier,
		[]byte("arg2"),
		[]byte("arg3"),
	}

	gsc.eei.SetStorage([]byte(proposalPrefix+string(proposalIdentifier)), []byte("1"))
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.GetReturnMessage(), "proposal already exists")
}

func TestGovernanceContract_ProposalInvalidVoteNonce(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	gsc, _, eei := createGovernanceBlockChainHookStubContextHandler()
	callInputArgs := [][]byte{
		proposalIdentifier,
		[]byte("arg2"),
		[]byte("arg3"),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.GetReturnMessage(), vm.ErrInvalidStartEndVoteNonce.Error())
}

func TestGovernanceContract_ProposalOK(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	gsc, _, _ := createGovernanceBlockChainHookStubContextHandler()

	callInputArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
		[]byte("10"),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_VoteWithBadArgsOrCallValue(t *testing.T) {
	t.Parallel()

	gsc, _, eei := createGovernanceBlockChainHookStubContextHandler()
	callInput := createVMInput(big.NewInt(0), "vote", vm.GovernanceSCAddress, []byte("addr1"), [][]byte{[]byte("bad args")})
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)

	callInput.CallValue = big.NewInt(10)
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "function is not payable"))

	callInput.CallValue = big.NewInt(0)
	callInput.Arguments = [][]byte{{1}, {2}}
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "only user can call this"))

	callInput.CallerAddr = bytes.Repeat([]byte{1}, 32)
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "not enough stake/delegate to vote"))
}

func TestGovernanceContract_VoteNotEnoughGas(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		UseGasCalled: func(_ uint64) error {
			return errors.New("not enough gas")
		},
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "vote", vm.GovernanceSCAddress, []byte("addr1"), make([][]byte, 0))
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.OutOfGas, retCode)
}

func TestGovernanceContract_VoteInvalidProposal(t *testing.T) {
	t.Parallel()

	callerAddress := bytes.Repeat([]byte{2}, 32)
	proposalIdentifier := []byte("aaaaaaaaa")
	generalProposal := &GeneralProposal{
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce:   15,
	}

	voteArgs := [][]byte{
		[]byte("1"),
		[]byte("yes"),
	}
	gsc, blockchainHook, eei := createGovernanceBlockChainHookStubContextHandler()
	blockchainHook.CurrentNonceCalled = func() uint64 {
		return 16
	}

	nonce, _ := nonceFromBytes(voteArgs[0])
	gsc.eei.SetStorage(nonce.Bytes(), proposalIdentifier)
	_ = gsc.saveGeneralProposal(proposalIdentifier, generalProposal)

	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.GetReturnMessage(), vm.ErrVotedForAnExpiredProposal.Error())
}

func TestGovernanceContract_VoteInvalidVote(t *testing.T) {
	t.Parallel()

	callerAddress := bytes.Repeat([]byte{2}, 32)
	proposalIdentifier := []byte("aaaaaaaaa")
	generalProposal := &GeneralProposal{
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce:   15,
	}

	voteArgs := [][]byte{
		[]byte("1"),
		[]byte("invalid"),
	}
	gsc, blockchainHook, eei := createGovernanceBlockChainHookStubContextHandler()
	blockchainHook.CurrentNonceCalled = func() uint64 {
		return 16
	}

	nonce, _ := nonceFromBytes(voteArgs[0])
	gsc.eei.SetStorage(nonce.Bytes(), proposalIdentifier)
	_ = gsc.saveGeneralProposal(proposalIdentifier, generalProposal)

	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.GetReturnMessage(), "invalid argument: invalid vote type option: invalid")
}

func TestGovernanceContract_VoteTwice(t *testing.T) {
	t.Parallel()

	gsc, blockchainHook, eei := createGovernanceBlockChainHookStubContextHandler()
	blockchainHook.CurrentNonceCalled = func() uint64 {
		return 12
	}

	callerAddress := bytes.Repeat([]byte{2}, 32)
	proposalIdentifier := []byte("aaaaaaaaa")
	generalProposal := &GeneralProposal{
		ProposalCost:   gsc.baseProposalCost,
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce:   15,
		Yes:            big.NewInt(0),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Abstain:        big.NewInt(0),
	}

	voteArgs := [][]byte{
		[]byte("1"),
		[]byte("yes"),
	}

	nonce, _ := nonceFromBytes(voteArgs[0])
	gsc.eei.SetStorage(nonce.Bytes(), proposalIdentifier)
	_ = gsc.saveGeneralProposal(proposalIdentifier, generalProposal)

	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	voteArgs[1] = []byte("no")
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.GetReturnMessage(), "double vote is not allowed")
}

func TestGovernanceContract_DelegateVoteUserErrors(t *testing.T) {
	t.Parallel()

	gsc, blockchainHook, eei := createGovernanceBlockChainHookStubContextHandler()
	blockchainHook.CurrentNonceCalled = func() uint64 {
		return 12
	}

	callerAddress := bytes.Repeat([]byte{2}, 32)
	proposalIdentifier := []byte("aaaaaaaaa")
	generalProposal := &GeneralProposal{
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce:   15,
		Yes:            big.NewInt(0),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Abstain:        big.NewInt(0),
	}

	voteArgs := [][]byte{
		[]byte("1"),
		[]byte("yes"),
	}
	nonce, _ := nonceFromBytes(voteArgs[0])
	gsc.eei.SetStorage(nonce.Bytes(), proposalIdentifier)
	_ = gsc.saveGeneralProposal(proposalIdentifier, generalProposal)

	callInput := createVMInput(big.NewInt(0), "delegateVote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.GetReturnMessage(), "invalid number of arguments")

	callInput.Arguments = append(callInput.Arguments, []byte{1}, []byte{2})
	callInput.CallValue = big.NewInt(10)
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "function is not payable"))

	callInput.CallValue = big.NewInt(0)
	callInput.GasProvided = 0
	gsc.gasCost.MetaChainSystemSCsCost.DelegateVote = 10
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.OutOfGas, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "not enough gas"))
}

func TestGovernanceContract_DelegateVoteMoreErrors(t *testing.T) {
	t.Parallel()

	gsc, blockchainHook, eei := createGovernanceBlockChainHookStubContextHandler()
	blockchainHook.CurrentNonceCalled = func() uint64 {
		return 12
	}

	callerAddress := bytes.Repeat([]byte{2}, 32)
	proposalIdentifier := []byte("aaaaaaaaa")
	generalProposal := &GeneralProposal{
		ProposalCost:   gsc.baseProposalCost,
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce:   15,
		Yes:            big.NewInt(0),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Abstain:        big.NewInt(0),
	}

	voteArgs := [][]byte{
		[]byte("1"),
		[]byte("yes"),
		{1},
		big.NewInt(10000).Bytes(),
	}
	nonce, _ := nonceFromBytes(voteArgs[0])
	gsc.eei.SetStorage(nonce.Bytes(), proposalIdentifier)
	_ = gsc.saveGeneralProposal(proposalIdentifier, generalProposal)

	callInput := createVMInput(big.NewInt(0), "delegateVote", callerAddress, vm.GovernanceSCAddress, voteArgs)

	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "only SC can call this"))

	callInput.CallerAddr = vm.ESDTSCAddress
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "invalid delegator address"))

	callInput.Arguments[2] = callerAddress
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "not enough stake/delegate to vote"))

	addStakeAndDelegationForAddress(gsc, callInput.CallerAddr)

	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "not enough voting power to cast this vote"))

	callInput.Arguments[3] = big.NewInt(12).Bytes()
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "double vote is not allowed"))
}

func TestGovernanceContract_CloseProposal(t *testing.T) {
	t.Parallel()

	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 1
				},
			}
		},
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				whitelistProposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Passed: true,
				})
				return whitelistProposalBytes
			}
			if bytes.Equal(key, []byte(governanceConfigKey)) {
				configBytes, _ := args.Marshalizer.Marshal(&GovernanceConfigV2{
					MinQuorum:        0.1,
					MinVetoThreshold: 0.1,
					MinPassThreshold: 0.1,
				})
				return configBytes
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				whitelistProposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Yes:  big.NewInt(10),
					No:   big.NewInt(10),
					Veto: big.NewInt(10),
				})
				return whitelistProposalBytes
			}

			return nil
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		proposalIdentifier,
	}

	callInput := createVMInput(zero, "closeProposal", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_CloseProposalWrongCallValue(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "closeProposal callValue expected to be 0"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{
		proposalIdentifier,
	}

	callInput := createVMInput(big.NewInt(10), "closeProposal", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_CloseProposalWrongArgumentsLength(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "invalid number of arguments expected 1"
	callerAddress := []byte("address")
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Passed: true,
				})
				return proposalBytes
			}

			return nil
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(zero, "closeProposal", callerAddress, vm.GovernanceSCAddress, nil)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_CloseProposalNotEnoughGas(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "not enough gas"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Passed: true,
				})
				return proposalBytes
			}

			return nil
		},
		UseGasCalled: func(_ uint64) error {
			return errors.New("not enough gas")
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{
		proposalIdentifier,
	}
	callInput := createVMInput(zero, "closeProposal", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.OutOfGas, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_CloseProposalGetProposalErr(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "getGeneralProposal error"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			return nil
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{
		proposalIdentifier,
	}
	callInput := createVMInput(zero, "closeProposal", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_CloseProposalAlreadyClosed(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "proposal is already closed, do nothing"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Yes:    big.NewInt(10),
					No:     big.NewInt(10),
					Veto:   big.NewInt(10),
					Closed: true,
				})
				return proposalBytes
			}

			return nil
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{
		proposalIdentifier,
	}
	callInput := createVMInput(zero, "closeProposal", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.Ok, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_CloseProposalVoteNotfinished(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "proposal can be closed only after nonce"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Yes:          big.NewInt(10),
					No:           big.NewInt(10),
					Veto:         big.NewInt(10),
					EndVoteNonce: 10,
				})
				return proposalBytes
			}

			return nil
		},
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 1
				},
			}
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{
		proposalIdentifier,
	}
	callInput := createVMInput(zero, "closeProposal", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_CloseProposalComputeResultsErr(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "computeEndResults error"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Yes:  big.NewInt(10),
					No:   big.NewInt(10),
					Veto: big.NewInt(10),
				})
				return proposalBytes
			}

			return nil
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{
		proposalIdentifier,
	}
	callInput := createVMInput(zero, "closeProposal", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_GetVotingPower(t *testing.T) {
	t.Parallel()

	gsc, _, eei := createGovernanceBlockChainHookStubContextHandler()
	callerAddress := bytes.Repeat([]byte{2}, 32)
	callInputArgs := [][]byte{
		callerAddress,
	}

	callInput := createVMInput(big.NewInt(0), "getVotingPower", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	vmOutput := eei.CreateVMOutput()
	require.Equal(t, big.NewInt(10).Bytes(), vmOutput.ReturnData[0])
}

func TestGovernanceContract_GetVVotingPowerWrongCallValue(t *testing.T) {
	t.Parallel()

	retMessage := ""
	callerAddress := []byte("address")
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(10), "getVotingPower", callerAddress, vm.GovernanceSCAddress, nil)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, vm.TransactionValueMustBeZero)
}

func TestGovernanceContract_GetVotingPowerWrongArgumentsLength(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "function accepts only one argument"
	callerAddress := []byte("address")
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{
		callerAddress,
		callerAddress,
	}
	callInput := createVMInput(zero, "getVotingPower", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_GetVotingPowerInvalidArgument(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "invalid address"
	callerAddress := []byte("address")
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{
		[]byte("address_wrong"),
	}
	callInput := createVMInput(zero, "getVotingPower", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_GetVotingPowerComputeErr(t *testing.T) {
	t.Parallel()

	callerAddress := []byte("address")
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageFromAddressCalled: func(_ []byte, _ []byte) []byte {
			return []byte("invalid data")
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{
		callerAddress,
	}
	callInput := createVMInput(zero, "getVotingPower", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

// ========  Begin testing of helper functions

func TestGovernanceContract_GetGeneralProposalNotFound(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)

	proposal, err := gsc.getGeneralProposal(proposalIdentifier)
	require.Nil(t, proposal)
	require.Equal(t, vm.ErrProposalNotFound, err)
}

func TestGovernanceContract_GetGeneralProposalUnmarshalErr(t *testing.T) {
	t.Parallel()

	unmarshalErr := errors.New("unmarshal error")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(_ []byte) []byte {
			return []byte("storage proposal")
		},
	}
	args.Marshalizer = &mock.MarshalizerStub{
		UnmarshalCalled: func(_ interface{}, _ []byte) error {
			return unmarshalErr
		},
	}
	gsc, _ := NewGovernanceContract(args)

	proposal, err := gsc.getGeneralProposal(proposalIdentifier)
	require.Nil(t, proposal)
	require.Equal(t, unmarshalErr, err)
}

func TestGovernanceContract_GetGeneralProposal(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	generalProposal := &GeneralProposal{
		CommitHash: proposalIdentifier,
	}

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(generalProposal)
				return proposalBytes
			}
			return nil
		},
	}
	gsc, _ := NewGovernanceContract(args)

	proposal, err := gsc.getGeneralProposal(proposalIdentifier)
	require.Nil(t, err)
	require.Equal(t, proposalIdentifier, proposal.CommitHash)
}

func TestGovernanceContract_SaveGeneralProposalUnmarshalErr(t *testing.T) {
	t.Parallel()

	customErr := errors.New("unmarshal error")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	generalProposal := &GeneralProposal{
		Yes: big.NewInt(10),
		No:  big.NewInt(0),
	}
	args := createMockGovernanceArgs()
	args.Marshalizer = &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, customErr
		},
	}

	gsc, _ := NewGovernanceContract(args)
	err := gsc.saveGeneralProposal(proposalIdentifier, generalProposal)

	require.Equal(t, customErr, err)
}

func TestGovernanceContract_SaveGeneralProposal(t *testing.T) {
	t.Parallel()

	marshaledProposal := []byte("general proposal")
	setStorageCalledKey := make([]byte, 0)
	setStorageCalledBytes := make([]byte, 0)
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	generalProposal := &GeneralProposal{
		Yes: big.NewInt(10),
		No:  big.NewInt(0),
	}

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		SetStorageCalled: func(key []byte, value []byte) {
			setStorageCalledKey = key
			setStorageCalledBytes = value
		},
	}
	args.Marshalizer = &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return marshaledProposal, nil
		},
	}

	gsc, _ := NewGovernanceContract(args)
	err := gsc.saveGeneralProposal(proposalIdentifier, generalProposal)

	require.Nil(t, err)
	require.Equal(t, append([]byte(proposalPrefix), proposalIdentifier...), setStorageCalledKey)
	require.Equal(t, marshaledProposal, setStorageCalledBytes)

}

func TestGovernanceContract_ProposalExists(t *testing.T) {
	t.Parallel()

	proposalReference := []byte("general proposal")

	correctKeyCalled := false
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalReference...)) {
				correctKeyCalled = true
			}
			return nil
		},
	}

	gsc, _ := NewGovernanceContract(args)
	proposalExists := gsc.proposalExists(proposalReference)

	require.False(t, proposalExists)
	require.True(t, correctKeyCalled)
}

func TestComputeEndResults(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, []byte(governanceConfigKey)) {
				configBytes, _ := args.Marshalizer.Marshal(&GovernanceConfigV2{
					MinQuorum:        0.4,
					MinPassThreshold: 0.5,
					MinVetoThreshold: 0.3,
				})
				return configBytes
			}

			return nil
		},
	}
	gsc, _ := NewGovernanceContract(args)

	didNotPassQuorum := &GeneralProposal{
		Yes:  big.NewInt(50),
		No:   big.NewInt(0),
		Veto: big.NewInt(0),
	}
	err := gsc.computeEndResults(didNotPassQuorum)
	require.Nil(t, err)
	require.False(t, didNotPassQuorum.Passed)

	didNotPassVotes := &GeneralProposal{
		Yes:  big.NewInt(50),
		No:   big.NewInt(50),
		Veto: big.NewInt(0),
	}
	err = gsc.computeEndResults(didNotPassVotes)
	require.Nil(t, err)
	require.False(t, didNotPassVotes.Passed)

	didNotPassVotes2 := &GeneralProposal{
		Yes:  big.NewInt(50),
		No:   big.NewInt(51),
		Veto: big.NewInt(0),
	}
	err = gsc.computeEndResults(didNotPassVotes2)
	require.Nil(t, err)
	require.False(t, didNotPassVotes2.Passed)

	didNotPassVeto := &GeneralProposal{
		Yes:  big.NewInt(51),
		No:   big.NewInt(50),
		Veto: big.NewInt(30),
	}
	err = gsc.computeEndResults(didNotPassVeto)
	require.Nil(t, err)
	require.False(t, didNotPassVeto.Passed)

	pass := &GeneralProposal{
		Yes:  big.NewInt(51),
		No:   big.NewInt(50),
		Veto: big.NewInt(29),
	}
	err = gsc.computeEndResults(pass)
	require.Nil(t, err)
	require.True(t, pass.Passed)
}
