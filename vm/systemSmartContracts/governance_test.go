package systemSmartContracts

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
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
				LostProposalFee:  "1",
			},
			OwnerAddress:                 "erd1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awqgqky80",
			MaxVotingDelayPeriodInEpochs: 30,
		},
		Marshalizer:                  &mock.MarshalizerMock{},
		Hasher:                       &hashingMocks.HasherMock{},
		GovernanceSCAddress:          vm.GovernanceSCAddress,
		DelegationMgrSCAddress:       vm.DelegationManagerSCAddress,
		ValidatorSCAddress:           vm.ValidatorSCAddress,
		OwnerAddress:                 bytes.Repeat([]byte{1}, 32),
		UnBondPeriodInEpochs:         10,
		MaxVotingDelayPeriodInEpochs: 30,
		EnableEpochsHandler:          enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GovernanceFlag, common.GovernanceFixesFlag, common.GovernanceDisableProposeFlag),
	}
}

func createEEIWithBlockchainHook(blockchainHook vm.BlockchainHook) vm.ContextHandler {
	eei, _ := NewVMContext(VMContextArgs{
		BlockChainHook:      blockchainHook,
		CryptoHook:          hooks.NewVMCryptoHook(),
		InputParser:         &mock.ArgumentParserMock{},
		ValidatorAccountsDB: &stateMock.AccountsStub{},
		UserAccountsDB:      &stateMock.AccountsStub{},
		ChanceComputer:      &mock.RaterMock{},
		EnableEpochsHandler: enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GovernanceFlag, common.GovernanceFixesFlag, common.GovernanceDisableProposeFlag),
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

func TestNewGovernanceContract_InvalidEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined()

	gsc, err := NewGovernanceContract(args)
	require.Nil(t, gsc)
	require.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
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

func TestGovernanceContract_SetNewGasCost(t *testing.T) {
	args := createMockGovernanceArgs()

	gsc, _ := NewGovernanceContract(args)
	require.False(t, gsc.IsInterfaceNil())
	require.True(t, gsc.CanUseContract())

	gasCost := vm.GasCost{MetaChainSystemSCsCost: vm.MetaChainSystemSCsCost{Vote: 1000000}}
	gsc.SetNewGasCost(gasCost)

	assert.Equal(t, gsc.gasCost.MetaChainSystemSCsCost.Vote, gasCost.MetaChainSystemSCsCost.Vote)
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
}

func TestGovernanceContract_ExecuteSendESDT(t *testing.T) {
	t.Parallel()

	gsc, _, eei := createGovernanceBlockChainHookStubContextHandler()

	callerAddr := []byte("addr1")
	callInput := createVMInput(big.NewInt(0), "vote", callerAddr, vm.GovernanceSCAddress, nil)
	callInput.ESDTTransfers = []*vmcommon.ESDTTransfer{{ESDTValue: big.NewInt(10), ESDTTokenName: []byte("tokenName"), ESDTTokenType: 0, ESDTTokenNonce: 0}}
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "cannot transfer ESDT to system SCs"))
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
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	gsc, _ := NewGovernanceContract(args)

	callInput := createVMInput(big.NewInt(0), "initV2", vm.GovernanceSCAddress, []byte("addr2"), nil)

	enableEpochsHandler.RemoveActiveFlags(common.GovernanceFlag)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)

	enableEpochsHandler.AddActiveFlags(common.GovernanceFlag)

	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_ChangeConfig(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	errorMsg := ""
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
				configBytes, _ := args.Marshalizer.Marshal(&GovernanceConfigV2{
					ProposalFee:     big.NewInt(10),
					LostProposalFee: big.NewInt(1),
				})
				return configBytes
			}

			return nil
		},
		AddReturnMessageCalled: func(msg string) {
			errorMsg = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		[]byte("1"),
		[]byte("1"),
		[]byte("10"),
		[]byte("10"),
		[]byte("5001"),
	}
	initInput := createVMInput(zero, "initV2", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
	_ = gsc.Execute(initInput)
	callInput := createVMInput(zero, "changeConfig", args.OwnerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	callInputArgs[4] = []byte("300")
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, errorMsg, "min pass should be higher than 50%")
}

func TestGovernanceContract_ChangeConfigOutOfGas(t *testing.T) {
	t.Parallel()

	gsc, _, eei := createGovernanceBlockChainHookStubContextHandler()

	callInputArgs := [][]byte{
		[]byte("1"),
		[]byte("1"),
		[]byte("10"),
		[]byte("10"),
		[]byte("5001"),
	}
	callInput := createVMInput(zero, "changeConfig", gsc.ownerAddress, vm.GovernanceSCAddress, callInputArgs)

	gsc.gasCost.MetaChainSystemSCsCost.ChangeConfig = 100
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.OutOfGas, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "not enough gas"))

	gsc.gasCost.MetaChainSystemSCsCost.ChangeConfig = 0
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_ValidatorVoteInvalidDelegated(t *testing.T) {
	t.Parallel()

	returnMessage := ""
	errInvalidVoteSubstr := "invalid delegator address"
	callerAddress := vm.FirstDelegationSCAddress
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		CommitHash:     proposalIdentifier,
		StartVoteEpoch: 10,
		EndVoteEpoch:   15,
	}
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(generalProposal)
				return proposalBytes
			}

			return nil
		},
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 14
				},
			}
		},
		AddReturnMessageCalled: func(msg string) {
			returnMessage = msg
		},
	}
	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("yes"),
		[]byte("delegatedToWrongAddress"),
		big.NewInt(1000).Bytes(),
	}

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "delegateVote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, returnMessage, errInvalidVoteSubstr)
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
		StartVoteEpoch: 10,
		EndVoteEpoch:   15,
		Yes:            big.NewInt(0),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Abstain:        big.NewInt(0),
	}

	voteArgs := [][]byte{
		[]byte("1"),
		[]byte("yes"),
	}
	gsc.eei.SetStorage(voteArgs[0], proposalIdentifier)
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
	callInput := createVMInput(big.NewInt(10), "changeConfig", args.OwnerAddress, vm.GovernanceSCAddress, nil)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_ChangeConfigWrongArgumentsLength(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "changeConfig needs 5 arguments"
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)

	initInput := createVMInput(zero, "initV2", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
	_ = gsc.Execute(initInput)
	callInput := createVMInput(zero, "changeConfig", args.OwnerAddress, vm.GovernanceSCAddress, nil)
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
		[]byte("invalid"),
		[]byte("10"),
		[]byte("10"),
		[]byte("5"),
	}
	callInput := createVMInput(zero, "changeConfig", args.OwnerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)

	errSubstr = "changeConfig second argument is incorrectly formatted"
	callInputArgs = [][]byte{
		[]byte("1"),
		[]byte("invalid"),
		[]byte("10"),
		[]byte("10"),
		[]byte("5"),
	}
	callInput = createVMInput(zero, "changeConfig", args.OwnerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode = gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)

	errSubstr = vm.ErrIncorrectConfig.Error() + " proposal fee is smaller than lost proposal fee "
	callInputArgs = [][]byte{
		[]byte("1"),
		[]byte("10"),
		[]byte("10"),
		[]byte("10"),
		[]byte("5"),
	}
	callInput = createVMInput(zero, "changeConfig", args.OwnerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode = gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)

	errSubstr = "config incorrect minQuorum"
	callInputArgs = [][]byte{
		[]byte("1"),
		[]byte("1"),
		[]byte("invalid"),
		[]byte("10"),
		[]byte("5"),
	}
	callInput = createVMInput(zero, "changeConfig", args.OwnerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode = gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)

	errSubstr = "config incorrect minVeto"
	callInputArgs = [][]byte{
		[]byte("1"),
		[]byte("1"),
		[]byte("10"),
		[]byte("invalid"),
		[]byte("5"),
	}
	callInput = createVMInput(zero, "changeConfig", args.OwnerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode = gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)

	errSubstr = "config incorrect minPass"
	callInputArgs = [][]byte{
		[]byte("1"),
		[]byte("1"),
		[]byte("10"),
		[]byte("10"),
		[]byte("invalid"),
	}
	callInput = createVMInput(zero, "changeConfig", args.OwnerAddress, vm.GovernanceSCAddress, callInputArgs)
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
		[]byte("1"),
		[]byte("10"),
		[]byte("10"),
		[]byte("5001"),
	}
	callInput := createVMInput(zero, "changeConfig", args.OwnerAddress, vm.GovernanceSCAddress, callInputArgs)
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

func TestGovernanceContract_ProposalInvalidStartEndVoteEpochTooLong(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	gsc, _, eei := createGovernanceBlockChainHookStubContextHandler()
	callInputArgs := [][]byte{
		proposalIdentifier,
		[]byte("50"),
		[]byte("61"),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.GetReturnMessage(), vm.ErrInvalidStartEndVoteEpoch.Error())
}

func TestGovernanceContract_ProposalInvalidStartEndVoteEpochTooFarV1(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	gsc, _, eei := createGovernanceBlockChainHookStubContextHandler()
	callInputArgs := [][]byte{
		proposalIdentifier,
		{byte(50)},
		{byte(54)},
	}
	gsc.enableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GovernanceFlag)
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
	logsEntry := eei.GetLogs()
	assert.Equal(t, 1, len(logsEntry))
	expectedTopics := [][]byte{{1}, proposalIdentifier, {byte(50)}, {byte(54)}}
	assert.Equal(t, expectedTopics, logsEntry[0].Topics)
}

func TestGovernanceContract_ProposalInvalidStartEndVoteEpochTooFarV2(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	gsc, _, eei := createGovernanceBlockChainHookStubContextHandler()
	gsc.enableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GovernanceFlag, common.GovernanceDisableProposeFlag, common.GovernanceFixesFlag)
	callInputArgs := [][]byte{
		proposalIdentifier,
		{byte(50)},
		{byte(54)},
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.GetReturnMessage(), vm.ErrInvalidStartEndVoteEpoch.Error())
}

func TestGovernanceContract_ProposalOK(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	gsc, _, _ := createGovernanceBlockChainHookStubContextHandler()

	callInputArgs := [][]byte{
		proposalIdentifier,
		{byte(5)},
		{byte(8)},
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.Ok, retCode)
	logsEntry := gsc.eei.GetLogs()
	assert.Equal(t, 1, len(logsEntry))
	expectedTopics := [][]byte{{1}, proposalIdentifier, {byte(5)}, {byte(8)}}
	assert.Equal(t, expectedTopics, logsEntry[0].Topics)
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
		StartVoteEpoch: 10,
		EndVoteEpoch:   15,
	}

	voteArgs := [][]byte{
		[]byte("1"),
		[]byte("yes"),
	}
	gsc, blockchainHook, eei := createGovernanceBlockChainHookStubContextHandler()
	blockchainHook.CurrentEpochCalled = func() uint32 {
		return 16
	}

	gsc.eei.SetStorage(append([]byte(noncePrefix), voteArgs[0]...), proposalIdentifier)
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
		StartVoteEpoch: 10,
		EndVoteEpoch:   15,
	}

	voteArgs := [][]byte{
		[]byte("1"),
		[]byte("invalid"),
	}
	gsc, blockchainHook, eei := createGovernanceBlockChainHookStubContextHandler()
	blockchainHook.CurrentEpochCalled = func() uint32 {
		return 14
	}

	gsc.eei.SetStorage(append([]byte(noncePrefix), voteArgs[0]...), proposalIdentifier)
	_ = gsc.saveGeneralProposal(proposalIdentifier, generalProposal)

	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.GetReturnMessage(), "invalid argument: invalid vote type")
}

func TestGovernanceContract_VoteTwice(t *testing.T) {
	t.Parallel()

	gsc, blockchainHook, eei := createGovernanceBlockChainHookStubContextHandler()
	blockchainHook.CurrentEpochCalled = func() uint32 {
		return 12
	}

	callerAddress := bytes.Repeat([]byte{2}, 32)
	proposalIdentifier := []byte("aaaaaaaaa")
	generalProposal := &GeneralProposal{
		ProposalCost:   gsc.baseProposalCost,
		CommitHash:     proposalIdentifier,
		StartVoteEpoch: 10,
		EndVoteEpoch:   15,
		Yes:            big.NewInt(0),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Abstain:        big.NewInt(0),
		QuorumStake:    big.NewInt(0),
	}

	voteArgs := [][]byte{
		[]byte("1"),
		[]byte("yes"),
	}

	gsc.eei.SetStorage(append([]byte(noncePrefix), voteArgs[0]...), proposalIdentifier)
	_ = gsc.saveGeneralProposal(proposalIdentifier, generalProposal)

	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	voteArgs[1] = []byte("no")
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.GetReturnMessage(), "double vote is not allowed")
}

func TestGovernanceContract_VoteTwiceV2(t *testing.T) {
	t.Parallel()

	gsc, blockchainHook, eei := createGovernanceBlockChainHookStubContextHandler()
	gsc.enableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GovernanceFlag, common.GovernanceDisableProposeFlag, common.GovernanceFixesFlag)
	blockchainHook.CurrentEpochCalled = func() uint32 {
		return 12
	}

	callerAddress := bytes.Repeat([]byte{2}, 32)
	proposalIdentifier := []byte("aaaaaaaaa")
	generalProposal := &GeneralProposal{
		ProposalCost:   gsc.baseProposalCost,
		CommitHash:     proposalIdentifier,
		StartVoteEpoch: 10,
		EndVoteEpoch:   15,
		Yes:            big.NewInt(0),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Abstain:        big.NewInt(0),
		QuorumStake:    big.NewInt(0),
	}

	voteArgs := [][]byte{
		[]byte("1"),
		[]byte("yes"),
	}

	gsc.eei.SetStorage(append([]byte(noncePrefix), voteArgs[0]...), proposalIdentifier)
	_ = gsc.saveGeneralProposal(proposalIdentifier, generalProposal)

	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	voteArgs[1] = []byte("no")
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.GetReturnMessage(), "double vote is not allowed")
}

func TestGovernanceContract_VoteAfterGovernanceFixesActivationWithOngoingListV1(t *testing.T) {
	t.Parallel()

	gsc, blockchainHook, _ := createGovernanceBlockChainHookStubContextHandler()
	blockchainHook.CurrentEpochCalled = func() uint32 {
		return 12
	}

	delegatedAddress := bytes.Repeat([]byte{2}, 32)
	proposalIdentifier := []byte("aaaaaaaaa")
	generalProposal := &GeneralProposal{
		ProposalCost:   gsc.baseProposalCost,
		CommitHash:     proposalIdentifier,
		StartVoteEpoch: 10,
		EndVoteEpoch:   15,
		Yes:            big.NewInt(0),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Abstain:        big.NewInt(0),
		QuorumStake:    big.NewInt(0),
	}

	voteArgs := [][]byte{
		[]byte("1"),
		[]byte("yes"),
	}
	delegateVoteArgs := [][]byte{
		[]byte("1"),
		[]byte("yes"),
		delegatedAddress,
		big.NewInt(30).Bytes(),
	}

	gsc.eei.SetStorage(append([]byte(noncePrefix), delegateVoteArgs[0]...), proposalIdentifier)
	_ = gsc.saveGeneralProposal(proposalIdentifier, generalProposal)

	callInput := createVMInput(big.NewInt(0), "delegateVote", vm.ESDTSCAddress, vm.GovernanceSCAddress, delegateVoteArgs)
	addStakeAndDelegationForAddress(gsc, callInput.CallerAddr)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	callInput = createVMInput(big.NewInt(0), "vote", delegatedAddress, vm.GovernanceSCAddress, voteArgs)
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	delegateVoteArgs[0] = []byte("2")
	generalProposal.CommitHash = []byte("bbbbbbbbb")

	gsc.eei.SetStorage(append([]byte(noncePrefix), delegateVoteArgs[0]...), proposalIdentifier)
	_ = gsc.saveGeneralProposal(proposalIdentifier, generalProposal)
	callInput = createVMInput(big.NewInt(0), "delegateVote", vm.ESDTSCAddress, vm.GovernanceSCAddress, delegateVoteArgs)

	gsc.enableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GovernanceFlag, common.GovernanceDisableProposeFlag, common.GovernanceFixesFlag)
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_DelegateVoteMoreErrors(t *testing.T) {
	t.Parallel()

	gsc, blockchainHook, eei := createGovernanceBlockChainHookStubContextHandler()
	blockchainHook.CurrentEpochCalled = func() uint32 {
		return 12
	}

	callerAddress := bytes.Repeat([]byte{2}, 32)
	proposalIdentifier := []byte("aaaaaaaaa")
	generalProposal := &GeneralProposal{
		ProposalCost:   gsc.baseProposalCost,
		CommitHash:     proposalIdentifier,
		StartVoteEpoch: 10,
		EndVoteEpoch:   15,
		Yes:            big.NewInt(0),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Abstain:        big.NewInt(0),
		QuorumStake:    big.NewInt(0),
	}

	voteArgs := [][]byte{
		[]byte("1"),
		[]byte("yes"),
		{1},
		big.NewInt(10000).Bytes(),
	}

	gsc.eei.SetStorage(append([]byte(noncePrefix), voteArgs[0]...), proposalIdentifier)
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

	callInput.Arguments[3] = []byte("0156464564645646846846456121684641")
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "too long argument for user stake"))

	callInput.Arguments[0] = []byte("01")
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "double vote is not allowed"))

	addStakeAndDelegationForAddress(gsc, vm.FirstDelegationSCAddress)
	callInput.CallerAddr = vm.FirstDelegationSCAddress
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "double vote is not allowed"))

	callInput.Arguments[0] = []byte("01")
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "double vote is not allowed"))

	callInput.Arguments[0] = []byte("001")
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "double vote is not allowed"))

	callInput.Arguments[0] = []byte("0001")
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "double vote is not allowed"))
}

func TestGovernanceContract_ProposalDuringDisableFlag(t *testing.T) {
	t.Parallel()

	gsc, _, eei := createGovernanceBlockChainHookStubContextHandler()
	gsc.enableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GovernanceFlag, common.GovernanceDisableProposeFlag)
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	callInputArgs := [][]byte{
		proposalIdentifier,
		{byte(5)},
		{byte(14)},
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), callInputArgs)
	retCode := gsc.Execute(callInput)

	require.True(t, strings.Contains(eei.GetReturnMessage(), "proposing is disabled"))
	require.Equal(t, vmcommon.UserError, retCode)

	gsc.enableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GovernanceFlag, common.GovernanceDisableProposeFlag, common.GovernanceFixesFlag)
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_DelegateVoteMultiple(t *testing.T) {
	t.Parallel()

	gsc, blockchainHook, eei := createGovernanceBlockChainHookStubContextHandler()
	gsc.enableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GovernanceFlag, common.GovernanceFixesFlag)
	blockchainHook.CurrentEpochCalled = func() uint32 {
		return 12
	}

	voter := bytes.Repeat([]byte{2}, 32)
	proposalIdentifier := []byte("aaaaaaaaa")
	generalProposal := &GeneralProposal{
		ProposalCost:   gsc.baseProposalCost,
		CommitHash:     proposalIdentifier,
		StartVoteEpoch: 10,
		EndVoteEpoch:   15,
		Yes:            big.NewInt(0),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Abstain:        big.NewInt(0),
		QuorumStake:    big.NewInt(0),
	}

	voteArgs := [][]byte{
		[]byte("1"),
		[]byte("yes"),
		voter,
		big.NewInt(12).Bytes(),
	}

	gsc.eei.SetStorage(append([]byte(noncePrefix), voteArgs[0]...), proposalIdentifier)
	_ = gsc.saveGeneralProposal(proposalIdentifier, generalProposal)

	addStakeAndDelegationForAddress(gsc, vm.ESDTSCAddress)
	addStakeAndDelegationForAddress(gsc, vm.FirstDelegationSCAddress)
	callInput := createVMInput(big.NewInt(0), "delegateVote", vm.ESDTSCAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "double vote is not allowed"))

	callInput.CallerAddr = vm.FirstDelegationSCAddress
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
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
				CurrentEpochCalled: func() uint32 {
					return 1
				},
			}
		},
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, []byte(governanceConfigKey)) {
				configBytes, _ := args.Marshalizer.Marshal(&GovernanceConfigV2{
					MinQuorum:        0.1,
					MinVetoThreshold: 0.1,
					MinPassThreshold: 0.1,
					ProposalFee:      big.NewInt(10),
					LostProposalFee:  big.NewInt(1),
				})
				return configBytes
			}
			if bytes.Equal(key, append([]byte(noncePrefix), byte(1))) {
				return proposalIdentifier
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					ProposalCost:  big.NewInt(10),
					Yes:           big.NewInt(10),
					No:            big.NewInt(10),
					Veto:          big.NewInt(10),
					Abstain:       big.NewInt(10),
					IssuerAddress: callerAddress,
				})
				return proposalBytes
			}

			return nil
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{{1}}

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
			if bytes.Equal(key, append([]byte(noncePrefix), byte(1))) {
				return proposalIdentifier
			}
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
	callInputArgs := [][]byte{{1}}
	callInput := createVMInput(zero, "closeProposal", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_CloseProposalVoteNotfinished(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "proposal can be closed only after epoch"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(noncePrefix), byte(1))) {
				return proposalIdentifier
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Yes:           big.NewInt(10),
					No:            big.NewInt(10),
					Veto:          big.NewInt(10),
					EndVoteEpoch:  10,
					IssuerAddress: callerAddress,
				})
				return proposalBytes
			}

			return nil
		},
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentEpochCalled: func() uint32 {
					return 1
				},
			}
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{{1}}
	callInput := createVMInput(zero, "closeProposal", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_CloseProposalCallerNotIssuer(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "only the issuer can close the proposal"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(noncePrefix), byte(1))) {
				return proposalIdentifier
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Yes:            big.NewInt(10),
					No:             big.NewInt(10),
					Veto:           big.NewInt(10),
					EndVoteEpoch:   10,
					StartVoteEpoch: 5,
				})
				return proposalBytes
			}

			return nil
		},
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentEpochCalled: func() uint32 {
					return 1
				},
			}
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{{1}}
	callInput := createVMInput(zero, "closeProposal", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)

	gsc.enableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GovernanceFlag, common.GovernanceDisableProposeFlag, common.GovernanceFixesFlag)

	retCode = gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, "only the issuer can close the proposal before start vote epoch")
}

func TestGovernanceContract_CloseProposalComputeResultsErr(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "element was not found"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 1
				},
				CurrentEpochCalled: func() uint32 {
					return 1
				},
			}
		},
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(noncePrefix), byte(1))) {
				return proposalIdentifier
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					ProposalCost:  big.NewInt(10),
					Yes:           big.NewInt(10),
					No:            big.NewInt(10),
					Veto:          big.NewInt(10),
					IssuerAddress: callerAddress,
				})
				return proposalBytes
			}

			return nil
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{{1}}
	callInput := createVMInput(zero, "closeProposal", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_CloseProposalAfterGovernanceFixesShouldSetLastEndedNonce(t *testing.T) {
	t.Parallel()

	callerAddress := []byte("address")
	gsc, blockchainHook, _ := createGovernanceBlockChainHookStubContextHandler()
	gsc.enableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GovernanceFlag, common.GovernanceDisableProposeFlag, common.GovernanceFixesFlag)
	currentEpoch := uint32(17)
	blockchainHook.CurrentEpochCalled = func() uint32 {
		return currentEpoch
	}

	generalProposal := &GeneralProposal{
		Nonce:          1,
		CommitHash:     bytes.Repeat([]byte("a"), commitHashLength),
		StartVoteEpoch: 10,
		EndVoteEpoch:   15,
		Yes:            big.NewInt(0),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Abstain:        big.NewInt(0),
		ProposalCost:   big.NewInt(10),
	}
	nonceKey := append([]byte(noncePrefix), big.NewInt(int64(generalProposal.Nonce)).Bytes()...)
	gsc.eei.SetStorage(nonceKey, generalProposal.CommitHash)
	_ = gsc.saveGeneralProposal(generalProposal.CommitHash, generalProposal)

	callInputArgs := [][]byte{{1}}
	callInput := createVMInput(zero, "closeProposal", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.Ok, retCode)
	lastEndedNonce := gsc.getLastEndedNonce()
	require.Equal(t, generalProposal.Nonce, lastEndedNonce)

	generalProposal.Nonce = 2
	generalProposal.CommitHash = bytes.Repeat([]byte("b"), commitHashLength)
	nonceKey = append([]byte(noncePrefix), big.NewInt(int64(generalProposal.Nonce)).Bytes()...)
	gsc.eei.SetStorage(nonceKey, generalProposal.CommitHash)
	_ = gsc.saveGeneralProposal(generalProposal.CommitHash, generalProposal)

	generalProposal.Nonce = 3
	generalProposal.CommitHash = bytes.Repeat([]byte("c"), commitHashLength)
	nonceKey = append([]byte(noncePrefix), big.NewInt(int64(generalProposal.Nonce)).Bytes()...)
	gsc.eei.SetStorage(nonceKey, generalProposal.CommitHash)
	_ = gsc.saveGeneralProposal(generalProposal.CommitHash, generalProposal)

	generalProposal.Nonce = 4
	generalProposal.CommitHash = bytes.Repeat([]byte("d"), commitHashLength)
	generalProposal.EndVoteEpoch = 18
	nonceKey = append([]byte(noncePrefix), big.NewInt(int64(generalProposal.Nonce)).Bytes()...)
	gsc.eei.SetStorage(nonceKey, generalProposal.CommitHash)
	_ = gsc.saveGeneralProposal(generalProposal.CommitHash, generalProposal)

	generalProposal.Nonce = 5
	generalProposal.CommitHash = bytes.Repeat([]byte("e"), commitHashLength)
	generalProposal.EndVoteEpoch = 15
	nonceKey = append([]byte(noncePrefix), big.NewInt(int64(generalProposal.Nonce)).Bytes()...)
	gsc.eei.SetStorage(nonceKey, generalProposal.CommitHash)
	_ = gsc.saveGeneralProposal(generalProposal.CommitHash, generalProposal)

	generalProposal.Nonce = 6
	generalProposal.CommitHash = bytes.Repeat([]byte("f"), commitHashLength)
	generalProposal.EndVoteEpoch = 15
	nonceKey = append([]byte(noncePrefix), big.NewInt(int64(generalProposal.Nonce)).Bytes()...)
	gsc.eei.SetStorage(nonceKey, generalProposal.CommitHash)
	_ = gsc.saveGeneralProposal(generalProposal.CommitHash, generalProposal)

	generalProposal.Nonce = 7
	generalProposal.CommitHash = bytes.Repeat([]byte("g"), commitHashLength)
	generalProposal.StartVoteEpoch = 15
	generalProposal.EndVoteEpoch = 20
	nonceKey = append([]byte(noncePrefix), big.NewInt(int64(generalProposal.Nonce)).Bytes()...)
	gsc.eei.SetStorage(nonceKey, generalProposal.CommitHash)
	_ = gsc.saveGeneralProposal(generalProposal.CommitHash, generalProposal)

	currentEpoch = 18
	callInput.Arguments = [][]byte{{3}}
	retCode = gsc.Execute(callInput)

	require.Equal(t, vmcommon.Ok, retCode)
	lastEndedNonce = gsc.getLastEndedNonce()
	require.Equal(t, uint64(3), lastEndedNonce)

	callInput.Arguments = [][]byte{{6}}
	retCode = gsc.Execute(callInput)

	require.Equal(t, vmcommon.Ok, retCode)
	lastEndedNonce = gsc.getLastEndedNonce()
	require.Equal(t, uint64(3), lastEndedNonce)

	currentEpoch = 19
	callInput.Arguments = [][]byte{{4}}
	retCode = gsc.Execute(callInput)

	require.Equal(t, vmcommon.Ok, retCode)
	lastEndedNonce = gsc.getLastEndedNonce()
	require.Equal(t, uint64(6), lastEndedNonce)
}

func TestGovernanceContract_CannotClose(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)

	currentEpoch := uint64(10)
	startVoteEpoch := currentEpoch + 1
	endVoteEpoch := currentEpoch + 10
	closedBeforeStart := &GeneralProposal{
		Yes:            big.NewInt(20),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Abstain:        big.NewInt(10),
		StartVoteEpoch: startVoteEpoch,
		EndVoteEpoch:   endVoteEpoch,
	}

	t.Run("fixes flag disabled, should return false only after end vote epoch", func(t *testing.T) {
		gsc.enableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GovernanceFlag)

		t.Run("current epoch is less than start epoch", func(t *testing.T) {
			t.Parallel()

			cannotClose := gsc.cannotClose(currentEpoch, closedBeforeStart)
			require.True(t, cannotClose)
		})
		t.Run("current epoch is equal with start epoch", func(t *testing.T) {
			t.Parallel()

			cannotClose := gsc.cannotClose(startVoteEpoch, closedBeforeStart)
			require.True(t, cannotClose)
		})
		t.Run("current epoch is between start and end epochs", func(t *testing.T) {
			t.Parallel()

			cannotClose := gsc.cannotClose(currentEpoch+5, closedBeforeStart)
			require.True(t, cannotClose)
		})
		t.Run("current epoch is equal with end epoch", func(t *testing.T) {
			t.Parallel()

			cannotClose := gsc.cannotClose(endVoteEpoch, closedBeforeStart)
			require.True(t, cannotClose)
		})
		t.Run("current epoch is larger than end epoch", func(t *testing.T) {
			t.Parallel()

			cannotClose := gsc.cannotClose(endVoteEpoch+1, closedBeforeStart)
			require.False(t, cannotClose)
		})
	})
	t.Run("fixes flag enabled, should return false only between start and end epoch", func(t *testing.T) {
		gsc.enableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GovernanceFixesFlag)

		t.Run("current epoch is less than start epoch", func(t *testing.T) {
			t.Parallel()

			cannotClose := gsc.cannotClose(currentEpoch, closedBeforeStart)
			require.False(t, cannotClose)
		})
		t.Run("current epoch is equal with start epoch", func(t *testing.T) {
			t.Parallel()

			cannotClose := gsc.cannotClose(startVoteEpoch, closedBeforeStart)
			require.True(t, cannotClose)
		})
		t.Run("current epoch is between start and end epochs", func(t *testing.T) {
			t.Parallel()

			cannotClose := gsc.cannotClose(currentEpoch+5, closedBeforeStart)
			require.True(t, cannotClose)
		})
		t.Run("current epoch is equal with end epoch", func(t *testing.T) {
			t.Parallel()

			cannotClose := gsc.cannotClose(endVoteEpoch, closedBeforeStart)
			require.True(t, cannotClose)
		})
		t.Run("current epoch is larger than end epoch", func(t *testing.T) {
			t.Parallel()

			cannotClose := gsc.cannotClose(endVoteEpoch+1, closedBeforeStart)
			require.False(t, cannotClose)
		})
	})
}

func TestGovernanceContract_ClearEndedProposalsCallValue(t *testing.T) {
	t.Parallel()

	gsc, _, eei := createGovernanceBlockChainHookStubContextHandler()
	callInput := createVMInput(big.NewInt(10), "clearEndedProposals", []byte("address"), vm.GovernanceSCAddress, nil)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.GetReturnMessage(), "clearEndedProposals callValue expected to be 0")
}

func TestGovernanceContract_ClearEndedProposalsNoArguments(t *testing.T) {
	t.Parallel()

	callerAddress := []byte("address")
	args := createMockGovernanceArgs()
	args.GasCost.MetaChainSystemSCsCost.ClearProposal = 10
	isCalled := false
	args.Eei = &mock.SystemEIStub{
		UseGasCalled: func(gas uint64) error {
			isCalled = true
			return nil
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(zero, "clearEndedProposals", callerAddress, vm.GovernanceSCAddress, make([][]byte, 0))
	retCode := gsc.Execute(callInput)

	// Assert
	require.Equal(t, vmcommon.Ok, retCode)
	require.True(t, isCalled)
}

func TestGovernanceContract_ClearEndedProposalsBadArguments(t *testing.T) {
	t.Parallel()

	callerAddress := bytes.Repeat([]byte{2}, 32)
	gsc, _, eei := createGovernanceBlockChainHookStubContextHandler()
	callInputArgs := [][]byte{
		bytes.Repeat([]byte{2}, 31),
	}
	callInput := createVMInput(zero, "clearEndedProposals", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.GetReturnMessage(), "invalid delegator address")
}

func TestGovernanceContract_ClearEndedProposalsOutOfGas(t *testing.T) {
	t.Parallel()

	callerAddress := []byte("address")
	args := createMockGovernanceArgs()
	args.GasCost.MetaChainSystemSCsCost.ClearProposal = 10
	args.Eei = &mock.SystemEIStub{
		UseGasCalled: func(gas uint64) error {
			if gas <= args.GasCost.MetaChainSystemSCsCost.ClearProposal*3 {
				return nil
			}
			return vm.ErrNotEnoughGas
		},
	}
	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		callerAddress,
	}
	callInput := createVMInput(zero, "clearEndedProposals", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	callInput.Arguments = [][]byte{
		callerAddress,
		callerAddress,
		callerAddress,
		callerAddress,
	}
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.OutOfGas, retCode)
}

func TestGovernanceContract_GetVotingPower(t *testing.T) {
	t.Parallel()

	gsc, _, eei := createGovernanceBlockChainHookStubContextHandler()
	callerAddress := bytes.Repeat([]byte{2}, 32)
	callInputArgs := [][]byte{
		callerAddress,
	}

	callInput := createVMInput(big.NewInt(0), "viewVotingPower", vm.GovernanceSCAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	fmt.Println(eei.GetReturnMessage())
	require.Equal(t, vmcommon.Ok, retCode)

	vmOutput := eei.CreateVMOutput()
	require.Equal(t, big.NewInt(120).Bytes(), vmOutput.ReturnData[0])
}

func TestGovernanceContract_GetVVotingPowerWrongCallValue(t *testing.T) {
	t.Parallel()

	retMessage := ""
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(10), "viewVotingPower", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, vm.ErrCallValueMustBeZero.Error())
}

func TestGovernanceContract_GetVotingPowerWrongArgumentsLength(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := vm.ErrInvalidNumOfArguments.Error()
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
	callInput := createVMInput(zero, "viewVotingPower", vm.GovernanceSCAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_GetVotingPowerInvalidCaller(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := vm.ErrInvalidCaller.Error()
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
	callInput := createVMInput(zero, "viewVotingPower", callerAddress, vm.GovernanceSCAddress, callInputArgs)
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
	callInput := createVMInput(zero, "viewVotingPower", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_ViewConfig(t *testing.T) {
	t.Parallel()

	callerAddress := []byte("address")
	args := createMockGovernanceArgs()
	returnMessage := ""
	returnedValues := make([]string, 0)
	mockEEI := &mock.SystemEIStub{
		GetStorageFromAddressCalled: func(_ []byte, _ []byte) []byte {
			return []byte("invalid data")
		},
		AddReturnMessageCalled: func(msg string) {
			returnMessage = msg
		},
		FinishCalled: func(value []byte) {
			returnedValues = append(returnedValues, string(value))
		},
	}
	args.Eei = mockEEI

	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{
		callerAddress,
	}
	callInput := createVMInput(zero, "viewConfig", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, returnMessage, vm.ErrInvalidCaller.Error())

	callInput.CallerAddr = callInput.RecipientAddr
	callInput.Arguments = [][]byte{}
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, returnMessage, "element was not found")

	mockEEI.GetStorageCalled = func(key []byte) []byte {
		proposalBytes, _ := args.Marshalizer.Marshal(&GovernanceConfigV2{
			ProposalFee:       big.NewInt(10),
			LostProposalFee:   big.NewInt(1),
			LastProposalNonce: 10,
			MinQuorum:         0.4,
			MinPassThreshold:  0.4,
			MinVetoThreshold:  0.4,
		})
		return proposalBytes
	}

	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	require.Equal(t, []string{"10", "1", "0.4000", "0.4000", "0.4000", "10"}, returnedValues)
}

func TestGovernanceContract_ViewProposal(t *testing.T) {
	t.Parallel()

	callerAddress := []byte("address")
	args := createMockGovernanceArgs()
	returnMessage := ""
	mockEEI := &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			returnMessage = msg
		},
	}
	args.Eei = mockEEI

	gsc, _ := NewGovernanceContract(args)

	callInput := createVMInput(zero, "viewProposal", callerAddress, vm.GovernanceSCAddress, [][]byte{})
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, returnMessage, vm.ErrInvalidCaller.Error())

	callInput.CallerAddr = callInput.RecipientAddr
	callInput.Arguments = [][]byte{callerAddress}
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, returnMessage, vm.ErrProposalNotFound.Error())

	mockEEI.GetStorageCalled = func(key []byte) []byte {
		proposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
			Yes:          big.NewInt(10),
			No:           big.NewInt(10),
			Veto:         big.NewInt(10),
			Abstain:      big.NewInt(10),
			ProposalCost: big.NewInt(10),
			QuorumStake:  big.NewInt(10),
			Closed:       true,
		})
		return proposalBytes
	}

	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_ViewDelegatedVoteInfo(t *testing.T) {
	t.Parallel()

	callerAddress := []byte("address")
	args := createMockGovernanceArgs()
	returnMessage := ""
	mockEEI := &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			returnMessage = msg
		},
	}
	args.Eei = mockEEI

	gsc, _ := NewGovernanceContract(args)

	callInput := createVMInput(zero, "viewDelegatedVoteInfo", callerAddress, vm.GovernanceSCAddress, [][]byte{})
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, returnMessage, vm.ErrInvalidCaller.Error())

	callInput.CallerAddr = callInput.RecipientAddr
	callInput.Arguments = [][]byte{callerAddress}
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, returnMessage, vm.ErrInvalidNumOfArguments.Error())

	callInput.Arguments = [][]byte{callerAddress, callerAddress}

	mockEEI.GetStorageCalled = func(key []byte) []byte {
		delegatedVoteInfo, _ := args.Marshalizer.Marshal(&DelegatedSCVoteInfo{
			UsedPower:  big.NewInt(10),
			UsedStake:  big.NewInt(100),
			TotalPower: big.NewInt(1000),
			TotalStake: big.NewInt(10000),
		})
		return delegatedVoteInfo
	}

	retCode = gsc.Execute(callInput)
	fmt.Println(returnMessage)
	require.Equal(t, vmcommon.Ok, retCode)
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

func TestGovernanceContract_addNewVote(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)

	proposal := &GeneralProposal{
		Yes:     big.NewInt(0),
		No:      big.NewInt(0),
		Abstain: big.NewInt(0),
		Veto:    big.NewInt(0),
	}

	_ = gsc.addNewVote(yesString, big.NewInt(9), proposal)
	require.Equal(t, proposal.Yes, big.NewInt(9))

	_ = gsc.addNewVote(noString, big.NewInt(99), proposal)
	require.Equal(t, proposal.No, big.NewInt(99))

	_ = gsc.addNewVote(vetoString, big.NewInt(999), proposal)
	require.Equal(t, proposal.Veto, big.NewInt(999))

	_ = gsc.addNewVote(abstainString, big.NewInt(9999), proposal)
	require.Equal(t, proposal.Abstain, big.NewInt(9999))
}

func TestComputeEndResults(t *testing.T) {
	t.Parallel()

	baseConfig := &GovernanceConfigV2{
		MinQuorum:        0.4,
		MinPassThreshold: 0.5,
		MinVetoThreshold: 0.3,
		ProposalFee:      big.NewInt(10),
		LostProposalFee:  big.NewInt(1),
	}

	retMessage := ""
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, []byte(governanceConfigKey)) {
				configBytes, _ := args.Marshalizer.Marshal(baseConfig)
				return configBytes
			}

			return nil
		},
		GetBalanceCalled: func(_ []byte) *big.Int {
			return big.NewInt(100)
		},
		FinishCalled: func(value []byte) {
			retMessage = string(value)
		},
	}
	gsc, _ := NewGovernanceContract(args)

	startVoteEpoch := uint64(10)
	gsc.enableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GovernanceFixesFlag)
	closedBeforeStart := &GeneralProposal{
		Yes:            big.NewInt(20),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Abstain:        big.NewInt(10),
		StartVoteEpoch: startVoteEpoch,
	}
	passed := gsc.computeEndResults(startVoteEpoch-1, closedBeforeStart, baseConfig)
	require.False(t, passed)
	require.Equal(t, "Proposal closed before voting started", retMessage)
	require.False(t, closedBeforeStart.Passed)

	gsc.enableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
	didNotPassQuorum := &GeneralProposal{
		Yes:     big.NewInt(20),
		No:      big.NewInt(0),
		Veto:    big.NewInt(0),
		Abstain: big.NewInt(10),
	}
	passed = gsc.computeEndResults(startVoteEpoch, didNotPassQuorum, baseConfig)
	require.False(t, passed)
	require.Equal(t, "Proposal did not reach minQuorum", retMessage)
	require.False(t, didNotPassQuorum.Passed)

	didNotPassVotes := &GeneralProposal{
		Yes:     big.NewInt(50),
		No:      big.NewInt(50),
		Veto:    big.NewInt(0),
		Abstain: big.NewInt(10),
	}
	passed = gsc.computeEndResults(startVoteEpoch, didNotPassVotes, baseConfig)
	require.False(t, passed)
	require.Equal(t, "Proposal rejected", retMessage)
	require.False(t, didNotPassVotes.Passed)

	didNotPassVotes2 := &GeneralProposal{
		Yes:     big.NewInt(50),
		No:      big.NewInt(51),
		Veto:    big.NewInt(0),
		Abstain: big.NewInt(10),
	}
	passed = gsc.computeEndResults(startVoteEpoch, didNotPassVotes2, baseConfig)
	require.False(t, passed)
	require.Equal(t, "Proposal rejected", retMessage)
	require.False(t, didNotPassVotes2.Passed)

	didNotPassVeto := &GeneralProposal{
		Yes:     big.NewInt(51),
		No:      big.NewInt(50),
		Veto:    big.NewInt(70),
		Abstain: big.NewInt(10),
	}
	passed = gsc.computeEndResults(startVoteEpoch, didNotPassVeto, baseConfig)
	require.False(t, passed)
	require.Equal(t, "Proposal vetoed", retMessage)
	require.False(t, didNotPassVeto.Passed)

	pass := &GeneralProposal{
		Yes:     big.NewInt(70),
		No:      big.NewInt(50),
		Veto:    big.NewInt(10),
		Abstain: big.NewInt(10),
	}
	passed = gsc.computeEndResults(startVoteEpoch, pass, baseConfig)
	require.True(t, passed)
	require.Equal(t, "Proposal passed", retMessage)
}

func TestGovernanceContract_ProposeVoteClose(t *testing.T) {
	t.Parallel()

	callerAddress := bytes.Repeat([]byte{2}, 32)
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	gsc, blockchainHook, eei := createGovernanceBlockChainHookStubContextHandler()

	callInputArgs := [][]byte{
		proposalIdentifier,
		big.NewInt(5).Bytes(),
		big.NewInt(8).Bytes(),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	currentEpoch := uint32(7)
	blockchainHook.CurrentEpochCalled = func() uint32 {
		return currentEpoch
	}

	callInput = createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, [][]byte{big.NewInt(1).Bytes(), []byte("yes")})
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	currentEpoch = 56
	callInput = createVMInput(big.NewInt(0), "closeProposal", callerAddress, vm.GovernanceSCAddress, [][]byte{big.NewInt(1).Bytes()})
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	proposal, _ := gsc.getProposalFromNonce(big.NewInt(1))
	require.True(t, proposal.Closed)
	require.True(t, proposal.Passed)
	require.Equal(t, big.NewInt(500), eei.GetTotalSentToUser(callInput.CallerAddr))
}

func TestGovernanceContract_ProposeClosePayFee(t *testing.T) {
	t.Parallel()

	callerAddress := bytes.Repeat([]byte{2}, 32)
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	gsc, blockchainHook, eei := createGovernanceBlockChainHookStubContextHandler()

	callInputArgs := [][]byte{
		proposalIdentifier,
		big.NewInt(5).Bytes(),
		big.NewInt(8).Bytes(),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	currentEpoch := uint32(6)
	blockchainHook.CurrentEpochCalled = func() uint32 {
		return currentEpoch
	}

	currentEpoch = 56
	callInput = createVMInput(big.NewInt(0), "closeProposal", callerAddress, vm.GovernanceSCAddress, [][]byte{big.NewInt(1).Bytes()})
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)

	proposal, _ := gsc.getProposalFromNonce(big.NewInt(1))
	require.True(t, proposal.Closed)
	require.False(t, proposal.Passed)
	require.Equal(t, big.NewInt(499), eei.GetTotalSentToUser(callInput.CallerAddr))
}

func TestGovernanceContract_ClaimAccumulatedFees(t *testing.T) {
	t.Parallel()

	gsc, _, eei := createGovernanceBlockChainHookStubContextHandler()
	callInput := createVMInput(big.NewInt(500), "claimAccumulatedFees", []byte("addr1"), vm.GovernanceSCAddress, [][]byte{{1}})

	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.GetReturnMessage(), "callValue expected to be 0")

	callInput.CallValue = big.NewInt(0)
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "invalid number of arguments, expected 0"))

	callInput.Arguments = [][]byte{}
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "can be called only by owner"))

	gsc.gasCost.MetaChainSystemSCsCost.ClaimAccumulatedFees = 100
	callInput.CallerAddr = gsc.ownerAddress
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.OutOfGas, retCode)
	require.True(t, strings.Contains(eei.GetReturnMessage(), "not enough gas"))

	gsc.gasCost.MetaChainSystemSCsCost.ClaimAccumulatedFees = 0
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, big.NewInt(0), eei.GetTotalSentToUser(callInput.CallerAddr))

	gsc.addToAccumulatedFees(big.NewInt(100))

	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, big.NewInt(100), eei.GetTotalSentToUser(callInput.CallerAddr))

	require.Equal(t, big.NewInt(0), gsc.getAccumulatedFees())
}
