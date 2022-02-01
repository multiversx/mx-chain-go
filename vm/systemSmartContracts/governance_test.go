package systemSmartContracts

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func createMockGovernanceArgs() ArgsNewGovernanceContract {
	return ArgsNewGovernanceContract{
		Eei:     &mock.SystemEIStub{},
		GasCost: vm.GasCost{},
		GovernanceConfig: config.GovernanceSystemSCConfig{
			V1: config.GovernanceSystemSCConfigV1{
				NumNodes:         3,
				MinPassThreshold: 1,
				MinQuorum:        2,
				MinVetoThreshold: 2,
				ProposalCost:     "100",
			},
			Active: config.GovernanceSystemSCConfigActive{
				ProposalCost:     "500",
				MinQuorum:        "50",
				MinPassThreshold: "50",
				MinVetoThreshold: "50",
			},
		},
		Marshalizer:                 &mock.MarshalizerMock{},
		Hasher:                      &hashingMocks.HasherMock{},
		GovernanceSCAddress:         vm.GovernanceSCAddress,
		DelegationMgrSCAddress:      vm.DelegationManagerSCAddress,
		ValidatorSCAddress:          vm.ValidatorSCAddress,
		EpochNotifier:               &mock.EpochNotifierStub{},
		InitialWhiteListedAddresses: [][]byte{vm.GovernanceSCAddress},
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

func TestNewGovernanceContract_NilEpochNotifierShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.EpochNotifier = nil

	gsc, err := NewGovernanceContract(args)
	require.Nil(t, gsc)
	require.Equal(t, vm.ErrNilEpochNotifier, err)
}

func TestNewGovernanceContract_ZeroBaseProposerCostShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.GovernanceConfig.Active.ProposalCost = ""

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

func TestNewGovernanceContract_InvalidWhiteList(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.InitialWhiteListedAddresses = nil

	gsc, err := NewGovernanceContract(args)
	require.Nil(t, gsc)
	require.True(t, errors.Is(err, vm.ErrInvalidNumOfInitialWhiteListedAddress))

	args.InitialWhiteListedAddresses = [][]byte{[]byte("someAddress")}
	gsc, err = NewGovernanceContract(args)
	require.Nil(t, gsc)
	require.NotNil(t, err)
	require.True(t, bytes.Contains([]byte(err.Error()), []byte("invalid address length for")))
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
	args.GovernanceConfig.Active.MinQuorum = ""
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
	require.Equal(t, vmcommon.ExecutionFailed, retCode)
}

func TestGovernanceContract_ExecuteInitV2(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)

	callInput := createVMInput(big.NewInt(0), "initV2", vm.GovernanceSCAddress, []byte("addr2"), nil)

	gsc.flagEnabled.Reset()
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)

	_ = gsc.flagEnabled.SetReturningPrevious()

	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, gsc.ownerAddress, vm.GovernanceSCAddress)
}

func TestGovernanceContract_ProposalWrongCallValue(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.GovernanceConfig.Active.ProposalCost = "10"

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(9), "proposal", vm.GovernanceSCAddress, []byte("addr1"), nil)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.OutOfFunds, retCode)
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

func TestGovernanceContract_ProposalInvalidArgumentsLenght(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), nil)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)
}

func TestGovernanceContract_ProposalCallerNptWhitelisted(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{
		[]byte("arg1"),
		[]byte("arg2"),
		[]byte("arg3"),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_ProposalInvalidReferenceLength(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			return []byte("storage item")
		},
	}
	args.Marshalizer = &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			whitelistProposal, proposalOk := obj.(*GeneralProposal)
			if proposalOk {
				whitelistProposal.Passed = true
			}
			return nil
		},
	}
	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{
		[]byte("arg1"),
		[]byte("arg2"),
		[]byte("arg3"),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_ProposalAlreadyExists(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	args := createMockGovernanceArgs()
	args.Eei = createMockStorer(vm.GovernanceSCAddress, proposalIdentifier, &GeneralProposal{})
	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{
		proposalIdentifier,
		[]byte("arg2"),
		[]byte("arg3"),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_ProposalInvalidVoteNonce(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	args := createMockGovernanceArgs()
	args.Eei = createMockStorer(vm.GovernanceSCAddress, proposalIdentifier, nil)
	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{
		proposalIdentifier,
		[]byte("arg2"),
		[]byte("arg3"),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_ProposalOK(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	args := createMockGovernanceArgs()
	args.Eei = createMockStorer(vm.GovernanceSCAddress, proposalIdentifier, nil)
	gsc, _ := NewGovernanceContract(args)
	callInputArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
		[]byte("10"),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_VoteWithFundsNotEnoughGas(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		UseGasCalled: func(_ uint64) error {
			return errors.New("not enough gas")
		},
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(500), "voteWithFunds", vm.GovernanceSCAddress, []byte("addr1"), make([][]byte, 0))
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.OutOfGas, retCode)
}

func TestGovernanceContract_VoteWithFundsInvalidNumOfArguments(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)
	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
		[]byte("10"),
	}
	callInput := createVMInput(big.NewInt(500), "voteWithFunds", vm.GovernanceSCAddress, []byte("addr1"), voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)
}

func TestGovernanceContract_VoteWithFundsProposalNotFound(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)
	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
	}
	callInput := createVMInput(big.NewInt(500), "voteWithFunds", vm.GovernanceSCAddress, []byte("addr1"), voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_VoteWithFundsInvalidVoteType(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	args := createMockGovernanceArgs()
	args.Eei = createMockStorer(vm.GovernanceSCAddress, proposalIdentifier, &GeneralProposal{})
	gsc, _ := NewGovernanceContract(args)
	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
	}
	callInput := createVMInput(big.NewInt(500), "voteWithFunds", vm.GovernanceSCAddress, []byte("addr1"), voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_VoteWithFundsInvalidCallValue(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	args := createMockGovernanceArgs()
	args.Eei = createMockStorer(vm.GovernanceSCAddress, proposalIdentifier, &GeneralProposal{})
	gsc, _ := NewGovernanceContract(args)
	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("yes"),
	}
	callInput := createVMInput(big.NewInt(-500), "voteWithFunds", vm.GovernanceSCAddress, []byte("addr1"), voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_VoteWithFundsAddVoteError(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	defaultMarshalizer := &mock.MarshalizerMock{}

	args := createMockGovernanceArgs()
	args.Eei = createMockStorer(vm.GovernanceSCAddress, proposalIdentifier, &GeneralProposal{
		Yes: big.NewInt(0),
		No:  big.NewInt(0),
	})
	args.Marshalizer = &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return defaultMarshalizer.Unmarshal(obj, buff)
		},
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			_, isVoteSetType := obj.(*VoteSet)
			if isVoteSetType {
				return nil, errors.New("invalid vote set")
			}
			return defaultMarshalizer.Marshal(obj)
		},
	}
	gsc, _ := NewGovernanceContract(args)
	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("yes"),
	}
	callInput := createVMInput(big.NewInt(500), "voteWithFunds", vm.GovernanceSCAddress, []byte("addr1"), voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_VoteWithFundsAddSimpleVote(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	args := createMockGovernanceArgs()
	args.Eei = createMockStorer(vm.GovernanceSCAddress, proposalIdentifier, &GeneralProposal{
		Yes: big.NewInt(0),
		No:  big.NewInt(0),
	})
	gsc, _ := NewGovernanceContract(args)
	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("yes"),
	}
	callInput := createVMInput(big.NewInt(500), "voteWithFunds", vm.GovernanceSCAddress, []byte("addr1"), voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_VoteWithFundsTwiceBadVote(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			isGeneralProposalKey := bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...))
			if isGeneralProposalKey {
				marshaledProposal, _ := args.Marshalizer.Marshal(&GeneralProposal{})
				return marshaledProposal
			}

			marshaledVote, _ := args.Marshalizer.Marshal(&VoteSet{VoteItems: []*VoteDetails{{Value: No}}})
			return marshaledVote
		},
		AddReturnMessageCalled: func(msg string) {
			require.Equal(t, msg, "conflicting votes for same proposal")
		},
		GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
			marshalizer := &mock.MarshalizerMock{}
			if bytes.Equal(address, vm.DelegationManagerSCAddress) && bytes.Equal(key, []byte(delegationManagementKey)) {
				dManagementData := &DelegationManagement{MinDelegationAmount: big.NewInt(10)}
				marshaledData, _ := marshalizer.Marshal(dManagementData)
				return marshaledData
			}

			return nil
		},
	}
	gsc, _ := NewGovernanceContract(args)
	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("yes"),
	}
	callInput := createVMInput(big.NewInt(500), "voteWithFunds", vm.GovernanceSCAddress, []byte("addr1"), voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_VoteWithBadArgsOrCallValue(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "vote", vm.GovernanceSCAddress, []byte("addr1"), [][]byte{[]byte("bad args")})
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)

	callInput.CallValue = big.NewInt(10)
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_ValidatorVoteNotEnoughGas(t *testing.T) {
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

func TestGovernanceContract_ValidatorVoteInvalidProposal(t *testing.T) {
	t.Parallel()

	returnMessage := ""
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce:   15,
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
					return 16
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
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, vm.ErrVotedForAnExpiredProposal.Error(), returnMessage)
}

func TestGovernanceContract_ValidatorVoteInvalidVote(t *testing.T) {
	t.Parallel()

	returnMessage := ""
	errInvalidVoteSubstr := "invalid vote type option"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce:   15,
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
		[]byte("wrong vote"),
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, returnMessage, errInvalidVoteSubstr)
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
		StartVoteNonce: 10,
		EndVoteNonce:   15,
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

func TestGovernanceContract_ValidatorVoteComputePowerError(t *testing.T) {
	t.Parallel()

	returnMessage := ""
	errInvalidVoteSubstr := "could not return total stake for the provided address"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce:   15,
	}
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(generalProposal)
				return proposalBytes
			}

			return nil
		},
		GetStorageFromAddressCalled: func(_ []byte, _ []byte) []byte {
			return []byte("invalid proposal bytes")
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
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, returnMessage, errInvalidVoteSubstr)
}

func TestGovernanceContract_ValidatorVoteInvalidVoteSetError(t *testing.T) {
	t.Parallel()

	mockBlsKey := []byte("bls key")
	mockValidatorBlsKeys := [][]byte{
		mockBlsKey,
		mockBlsKey,
		mockBlsKey,
		mockBlsKey,
	}

	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	votePower := big.NewInt(100).Bytes()

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce:   15,
	}
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(generalProposal)
				return proposalBytes
			}
			if bytes.Equal(key, append(proposalIdentifier, callerAddress...)) {
				return []byte("invalid vote set")
			}

			return nil
		},
		GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
			if bytes.Equal(address, args.ValidatorSCAddress) && bytes.Equal(key, callerAddress) {
				auctionBytes, _ := args.Marshalizer.Marshal(&ValidatorDataV2{
					BlsPubKeys:      mockValidatorBlsKeys,
					TotalStakeValue: big.NewInt(0).SetBytes(votePower),
				})

				return auctionBytes
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
	}
	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("yes"),
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.ExecutionFailed, retCode)
}

func TestGovernanceContract_DelegateVoteVoteNotEnoughPower(t *testing.T) {
	t.Parallel()

	mockBlsKey := []byte("bls key")
	returnMessage := ""
	errInvalidVoteSubstr := "not enough voting power to cast this vote"
	mockValidatorBlsKeys := [][]byte{
		mockBlsKey,
		mockBlsKey,
		mockBlsKey,
		mockBlsKey,
	}

	callerAddress := vm.FirstDelegationSCAddress
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	votePower := big.NewInt(100).Bytes()

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce:   15,
	}
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(generalProposal)
				return proposalBytes
			}

			return nil
		},
		AddReturnMessageCalled: func(msg string) {
			returnMessage = msg
		},
		GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
			if bytes.Equal(address, args.ValidatorSCAddress) && bytes.Equal(key, callerAddress) {
				auctionBytes, _ := args.Marshalizer.Marshal(&ValidatorDataV2{
					BlsPubKeys:      mockValidatorBlsKeys,
					TotalStakeValue: big.NewInt(0).SetBytes(votePower),
				})

				return auctionBytes
			}
			if bytes.Equal(address, vm.DelegationManagerSCAddress) && bytes.Equal(key, []byte(delegationContractsList)) {
				contractList := &DelegationContractList{}
				marshaledData, _ := args.Marshalizer.Marshal(contractList)
				return marshaledData
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
	}

	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("yes"),
		big.NewInt(100000).Bytes(),
		callerAddress,
	}
	gsc, _ := NewGovernanceContract(args)

	callInput := createVMInput(big.NewInt(0), "delegateVote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, returnMessage, errInvalidVoteSubstr)
}

func TestGovernanceContract_DelegateVoteSuccess(t *testing.T) {
	t.Parallel()

	mockBlsKey := []byte("bls key")
	mockValidatorBlsKeys := [][]byte{
		mockBlsKey,
		mockBlsKey,
		mockBlsKey,
		mockBlsKey,
	}

	callerAddress := vm.FirstDelegationSCAddress
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	votePower := big.NewInt(100)

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce:   15,
		Yes:            big.NewInt(10),
	}
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(generalProposal)
				return proposalBytes
			}

			return nil
		},
		GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
			if bytes.Equal(address, args.ValidatorSCAddress) && bytes.Equal(key, callerAddress) {
				auctionBytes, _ := args.Marshalizer.Marshal(&ValidatorDataV2{
					BlsPubKeys:      mockValidatorBlsKeys,
					TotalStakeValue: big.NewInt(0).Set(votePower),
				})

				return auctionBytes
			}
			if bytes.Equal(address, vm.DelegationManagerSCAddress) && bytes.Equal(key, []byte(delegationContractsList)) {
				contractList := &DelegationContractList{}
				marshaledData, _ := args.Marshalizer.Marshal(contractList)
				return marshaledData
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
	}

	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("yes"),
		big.NewInt(10).Bytes(),
		callerAddress,
	}
	gsc, _ := NewGovernanceContract(args)

	callInput := createVMInput(big.NewInt(0), "delegateVote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_ValidatorVote(t *testing.T) {
	t.Parallel()

	mockBlsKey := []byte("bls key")
	mockValidatorBlsKeys := [][]byte{
		mockBlsKey,
		mockBlsKey,
		mockBlsKey,
		mockBlsKey,
	}

	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	votePower := big.NewInt(10)
	proposalKey := append([]byte(proposalPrefix), proposalIdentifier...)
	voteItemKey := append(proposalKey, callerAddress...)

	finalVoteSet := &VoteSet{}
	finalProposal := &GeneralProposal{}

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce:   15,
		Yes:            big.NewInt(0),
	}
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(generalProposal)
				return proposalBytes
			}
			if bytes.Equal(key, append([]byte(stakeLockPrefix), callerAddress...)) {
				return big.NewInt(10).Bytes()
			}

			return nil
		},
		GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
			if bytes.Equal(address, args.ValidatorSCAddress) && bytes.Equal(key, callerAddress) {
				auctionBytes, _ := args.Marshalizer.Marshal(&ValidatorDataV2{
					BlsPubKeys:      mockValidatorBlsKeys,
					TotalStakeValue: big.NewInt(100),
				})

				return auctionBytes
			}
			if bytes.Equal(address, vm.DelegationManagerSCAddress) && bytes.Equal(key, []byte(delegationContractsList)) {
				contractList := &DelegationContractList{Addresses: [][]byte{vm.FirstDelegationSCAddress}}
				marshaledData, _ := args.Marshalizer.Marshal(contractList)
				return marshaledData
			}

			return nil
		},

		SetStorageCalled: func(key []byte, value []byte) {
			if bytes.Equal(key, voteItemKey) {
				_ = args.Marshalizer.Unmarshal(finalVoteSet, value)
			}
			if bytes.Equal(key, proposalKey) {
				_ = args.Marshalizer.Unmarshal(finalProposal, value)
			}
		},
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 14
				},
			}
		},
	}

	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("yes"),
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, votePower, finalProposal.Yes)
	require.Equal(t, 1, len(finalProposal.Votes))
	require.Equal(t, votePower, finalVoteSet.TotalYes)
	require.Equal(t, votePower, finalVoteSet.UsedPower)
	require.Equal(t, big.NewInt(0), finalVoteSet.UsedBalance)
}

func TestGovernanceContract_ValidatorVoteTwice(t *testing.T) {
	t.Parallel()

	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce:   15,
		Yes:            big.NewInt(0),
	}
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(generalProposal)
				return proposalBytes
			}
			if bytes.Equal(key, append(proposalIdentifier, callerAddress...)) {
				voteSetBytes, _ := args.Marshalizer.Marshal(&VoteSet{VoteItems: []*VoteDetails{{Value: 0}}})
				return voteSetBytes
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
			require.Equal(t, msg, "vote only once")
		},
	}

	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("yes"),
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_DelegateVoteUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()

	mockEI := &mock.SystemEIStub{}
	args.Eei = mockEI

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "delegateVote", []byte("address"), vm.GovernanceSCAddress, nil)

	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)

	callInput.Arguments = [][]byte{{1}, {2}, {3}, {4}}
	callInput.CallValue = big.NewInt(10)
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, mockEI.ReturnMessage, "function is not payable")

	mockEI.UseGasCalled = func(_ uint64) error {
		return vm.ErrNotEnoughGas
	}
	callInput.CallValue = big.NewInt(0)
	args.Eei = mockEI
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.OutOfGas, retCode)

	mockEI.AddReturnMessageCalled = func(msg string) {
		require.Equal(t, msg, "only SC can call this")
	}
	mockEI.UseGasCalled = func(gas uint64) error {
		return nil
	}
	args.Eei = mockEI
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)

	mockEI.AddReturnMessageCalled = func(msg string) {
		require.Equal(t, msg, "invalid delegator address")
	}
	callInput.CallerAddr = vm.ESDTSCAddress
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)

	mockEI.AddReturnMessageCalled = func(msg string) {
		require.Equal(t, msg, vm.ErrProposalNotFound.Error())
	}
	args.Eei = mockEI
	callInput.Arguments[3] = vm.ESDTSCAddress
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)

	mockEI.GetStorageCalled = func(key []byte) []byte {
		proposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{})
		return proposalBytes
	}
	mockEI.AddReturnMessageCalled = func(msg string) {
		require.True(t, bytes.Contains([]byte(msg), []byte("invalid vote type option: ")))
	}
	args.Eei = mockEI
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_ClaimFundsWrongCallValue(t *testing.T) {
	t.Parallel()

	returnMessage := ""
	expectedErrorSubstr := "invalid callValue"
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			returnMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(9), "claimFunds", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, returnMessage, expectedErrorSubstr)
}

func TestGovernanceContract_ClaimFundsWrongNumberOfArguments(t *testing.T) {
	t.Parallel()

	returnMessage := ""
	expectedErrorSubstr := "invalid number of arguments"
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			returnMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "claimFunds", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)
	require.Contains(t, returnMessage, expectedErrorSubstr)
}

func TestGovernanceContract_ClaimFundsStillLocked(t *testing.T) {
	t.Parallel()

	returnMessage := ""
	expectedErrorSubstr := "your funds are still locked"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			returnMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			expectedKeyPrefix := append([]byte(fundsLockPrefix), proposalIdentifier...)
			if bytes.Equal(key, append(expectedKeyPrefix, callerAddress...)) {
				voteSetBytes, _ := args.Marshalizer.Marshal(&VoteSet{
					UsedBalance: big.NewInt(100),
				})
				return voteSetBytes
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalByes, _ := args.Marshalizer.Marshal(&GeneralProposal{EndVoteNonce: 100})
				return proposalByes
			}

			return nil
		},
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 11
				},
			}
		},
	}
	claimArgs := [][]byte{
		proposalIdentifier,
	}

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(zero, "claimFunds", callerAddress, vm.GovernanceSCAddress, claimArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, returnMessage, expectedErrorSubstr)
}

func TestGovernanceContract_ClaimFundsNothingToClaim(t *testing.T) {
	t.Parallel()

	returnMessage := ""
	expectedErrorSubstr := "no funds to claim for this proposal"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			returnMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			expectedKeyPrefix := append([]byte(fundsLockPrefix), proposalIdentifier...)
			if bytes.Equal(key, append(expectedKeyPrefix, callerAddress...)) {
				voteSetBytes, _ := args.Marshalizer.Marshal(&VoteSet{
					UsedBalance: zero,
				})
				return voteSetBytes
			}

			return nil
		},
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 11
				},
			}
		},
	}
	claimArgs := [][]byte{
		proposalIdentifier,
	}

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(zero, "claimFunds", callerAddress, vm.GovernanceSCAddress, claimArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, returnMessage, expectedErrorSubstr)
}

func TestGovernanceContract_ClaimFunds(t *testing.T) {
	t.Parallel()

	callerAddress := []byte("address")
	voteValue := big.NewInt(10)
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)

	finalVoteSet := &VoteSet{}
	transferFrom := make([]byte, 0)
	transferTo := make([]byte, 0)
	transferValue := big.NewInt(0)

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			expectedKeyPrefix := append([]byte(fundsLockPrefix), proposalIdentifier...)
			if bytes.Equal(key, append(expectedKeyPrefix, callerAddress...)) {
				voteSetBytes, _ := args.Marshalizer.Marshal(&VoteSet{
					UsedBalance: voteValue,
				})
				return voteSetBytes
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalByes, _ := args.Marshalizer.Marshal(&GeneralProposal{EndVoteNonce: 100})
				return proposalByes
			}

			return nil
		},
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 101
				},
			}
		},
		SetStorageCalled: func(key []byte, value []byte) {
			proposalKey := append([]byte(proposalPrefix), proposalIdentifier...)
			if bytes.Equal(key, append(proposalKey, callerAddress...)) {
				_ = args.Marshalizer.Unmarshal(finalVoteSet, value)
			}
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, _ []byte) error {
			transferTo = destination
			transferFrom = sender
			transferValue.Set(value)

			return nil
		},
	}
	claimArgs := [][]byte{
		proposalIdentifier,
	}

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(zero, "claimFunds", callerAddress, vm.GovernanceSCAddress, claimArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, args.GovernanceSCAddress, transferFrom)
	require.Equal(t, callerAddress, transferTo)
	require.Equal(t, voteValue, transferValue)
}

func TestGovernanceContract_WhiteListProposal(t *testing.T) {
	t.Parallel()

	callerAddress := []byte("address")
	finalWhitelistProposal := &WhiteListProposal{}
	finalProposal := &GeneralProposal{}
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
		SetStorageCalled: func(key []byte, value []byte) {
			if bytes.Equal(key, append([]byte(whiteListPrefix), callerAddress...)) {
				_ = args.Marshalizer.Unmarshal(finalWhitelistProposal, value)
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				_ = args.Marshalizer.Unmarshal(finalProposal, value)
			}
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
		[]byte("10"),
	}
	proposalCost, _ := big.NewInt(0).SetString(args.GovernanceConfig.Active.ProposalCost, conversionBase)
	callInput := createVMInput(proposalCost, "whiteList", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, callerAddress, finalWhitelistProposal.WhiteListAddress)
	require.Equal(t, append([]byte(proposalPrefix), callerAddress...), finalWhitelistProposal.ProposalStatus)
	require.Equal(t, proposalIdentifier, finalProposal.CommitHash)
}

func TestGovernanceContract_WhiteListProposalInvalidCallValue(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "invalid callValue"
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
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
		[]byte("10"),
	}
	callInput := createVMInput(big.NewInt(0), "whiteList", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.OutOfFunds, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_WhiteListProposalNotEnoughGas(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "not enough gas"
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
		UseGasCalled: func(_ uint64) error {
			return errors.New("not enough gas")
		},
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
		[]byte("10"),
	}
	proposalCost, _ := big.NewInt(0).SetString(args.GovernanceConfig.Active.ProposalCost, conversionBase)
	callInput := createVMInput(proposalCost, "whiteList", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.OutOfGas, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_WhiteListInvalidArgumentsLength(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "invalid number of arguments"
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
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
		[]byte("10"),
		[]byte("10"),
	}
	proposalCost, _ := big.NewInt(0).SetString(args.GovernanceConfig.Active.ProposalCost, conversionBase)
	callInput := createVMInput(proposalCost, "whiteList", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_WhiteListProposalAlreadyExists(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "cannot re-propose existing proposal"
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
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				return []byte("proposal exists")
			}

			return nil
		},
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
		[]byte("10"),
	}
	proposalCost, _ := big.NewInt(0).SetString(args.GovernanceConfig.Active.ProposalCost, conversionBase)
	callInput := createVMInput(proposalCost, "whiteList", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_WhiteListAlreadyWhitelisted(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "address is already whitelisted"
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
			if bytes.Equal(key, append([]byte(whiteListPrefix), callerAddress...)) {
				return []byte("whitelisted")
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Passed: true,
				})
				return proposalBytes
			}

			return nil
		},
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
		[]byte("10"),
	}
	proposalCost, _ := big.NewInt(0).SetString(args.GovernanceConfig.Active.ProposalCost, conversionBase)
	callInput := createVMInput(proposalCost, "whiteList", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_WhiteListInvalidProposalLength(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "invalid github commit length"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength-1)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 1
				},
			}
		},
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
		[]byte("10"),
	}
	proposalCost, _ := big.NewInt(0).SetString(args.GovernanceConfig.Active.ProposalCost, conversionBase)
	callInput := createVMInput(proposalCost, "whiteList", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_WhiteListInvalidNonces(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "invalid start/end vote nonce"
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
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
		[]byte("invalid"),
	}
	proposalCost, _ := big.NewInt(0).SetString(args.GovernanceConfig.Active.ProposalCost, conversionBase)
	callInput := createVMInput(proposalCost, "whiteList", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_HardForkProposal(t *testing.T) {
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
			if bytes.Equal(key, append([]byte(whiteListPrefix), callerAddress...)) {
				return []byte("whitelisted caller")
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				whitelistProposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Passed: true,
				})
				return whitelistProposalBytes
			}

			return nil
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		[]byte("1"),
		[]byte("10"),
		proposalIdentifier,
		[]byte("5"),
		[]byte("10"),
	}
	proposalCost, _ := big.NewInt(0).SetString(args.GovernanceConfig.Active.ProposalCost, conversionBase)
	callInput := createVMInput(proposalCost, "hardFork", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_HardForkProposalInvalidCallValue(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "invalid proposal cost"
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
		[]byte("1"),
		[]byte("10"),
		proposalIdentifier,
		[]byte("5"),
		[]byte("10"),
	}
	callInput := createVMInput(big.NewInt(0), "hardFork", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.OutOfFunds, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_HardForkProposalNotEnoughGas(t *testing.T) {
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
		UseGasCalled: func(gas uint64) error {
			return errors.New("not enough gas")
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		[]byte("1"),
		[]byte("10"),
		proposalIdentifier,
		[]byte("5"),
		[]byte("10"),
	}

	proposalCost, _ := big.NewInt(0).SetString(args.GovernanceConfig.Active.ProposalCost, conversionBase)
	callInput := createVMInput(proposalCost, "hardFork", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.OutOfGas, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_HardForkInvalidArgumentsLenght(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "invalid number of arguments"
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
		[]byte("1"),
		[]byte("10"),
		proposalIdentifier,
		[]byte("5"),
	}

	proposalCost, _ := big.NewInt(0).SetString(args.GovernanceConfig.Active.ProposalCost, conversionBase)
	callInput := createVMInput(proposalCost, "hardFork", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_HardForkProposalNotWhitelisted(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "called address is not whiteListed"
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
		[]byte("1"),
		[]byte("10"),
		proposalIdentifier,
		[]byte("5"),
		[]byte("10"),
	}

	proposalCost, _ := big.NewInt(0).SetString(args.GovernanceConfig.Active.ProposalCost, conversionBase)
	callInput := createVMInput(proposalCost, "hardFork", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_HardForkProposalInvalidCommitLength(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "invalid github commit length"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength-1)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(whiteListPrefix), callerAddress...)) {
				return []byte("whitelisted caller")
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				whitelistProposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Passed: true,
				})
				return whitelistProposalBytes
			}

			return nil
		},
	}
	args.EpochConfig.EnableEpochs.GovernanceEnableEpoch = 0
	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		[]byte("1"),
		[]byte("10"),
		proposalIdentifier,
		[]byte("5"),
		[]byte("10"),
	}

	proposalCost, _ := big.NewInt(0).SetString(args.GovernanceConfig.Active.ProposalCost, conversionBase)
	callInput := createVMInput(proposalCost, "hardFork", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_HardForkProposalProposalAlreadyExists(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "proposal already exists"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(whiteListPrefix), callerAddress...)) {
				return []byte("whitelisted caller")
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				whitelistProposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Passed: true,
				})
				return whitelistProposalBytes
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				return []byte("proposal exists")
			}

			return nil
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		[]byte("1"),
		[]byte("10"),
		proposalIdentifier,
		[]byte("5"),
		[]byte("10"),
	}

	proposalCost, _ := big.NewInt(0).SetString(args.GovernanceConfig.Active.ProposalCost, conversionBase)
	callInput := createVMInput(proposalCost, "hardFork", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_HardForkProposalInvalidNonce(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "invalid start/end vote nonce"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(whiteListPrefix), callerAddress...)) {
				return []byte("whitelisted caller")
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				whitelistProposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Passed: true,
				})
				return whitelistProposalBytes
			}

			return nil
		},
	}

	gsc, _ := NewGovernanceContract(args)

	callInputArgs := [][]byte{
		[]byte("1"),
		[]byte("10"),
		proposalIdentifier,
		[]byte("5"),
		[]byte("invalid"),
	}

	proposalCost, _ := big.NewInt(0).SetString(args.GovernanceConfig.Active.ProposalCost, conversionBase)
	callInput := createVMInput(proposalCost, "hardFork", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
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
	callInput := createVMInput(zero, "changeConfig", vm.GovernanceSCAddress, vm.GovernanceSCAddress, callInputArgs)
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
	callInput := createVMInput(big.NewInt(10), "changeConfig", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
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
	callInput := createVMInput(zero, "changeConfig", vm.GovernanceSCAddress, vm.GovernanceSCAddress, nil)
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
	callInput := createVMInput(zero, "changeConfig", vm.GovernanceSCAddress, vm.GovernanceSCAddress, callInputArgs)
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
	callInput = createVMInput(zero, "changeConfig", vm.GovernanceSCAddress, vm.GovernanceSCAddress, callInputArgs)
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
	callInput = createVMInput(zero, "changeConfig", vm.GovernanceSCAddress, vm.GovernanceSCAddress, callInputArgs)
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
	callInput = createVMInput(zero, "changeConfig", vm.GovernanceSCAddress, vm.GovernanceSCAddress, callInputArgs)
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
	callInput := createVMInput(zero, "changeConfig", vm.GovernanceSCAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
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
			if bytes.Equal(key, append([]byte(whiteListPrefix), callerAddress...)) {
				return []byte("whitelisted caller")
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				whitelistProposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Passed: true,
				})
				return whitelistProposalBytes
			}
			if bytes.Equal(key, []byte(governanceConfigKey)) {
				configBytes, _ := args.Marshalizer.Marshal(&GovernanceConfigV2{
					MinQuorum:        big.NewInt(10),
					MinVetoThreshold: big.NewInt(10),
					MinPassThreshold: big.NewInt(10),
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

func TestGovernanceContract_CloseProposalNotWhitelisted(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "caller is not whitelisted"
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

	callInput := createVMInput(zero, "closeProposal", callerAddress, vm.GovernanceSCAddress, callInputArgs)
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
			if bytes.Equal(key, append([]byte(whiteListPrefix), callerAddress...)) {
				return []byte("whitelisted caller")
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				whitelistProposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Passed: true,
				})
				return whitelistProposalBytes
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
			if bytes.Equal(key, append([]byte(whiteListPrefix), callerAddress...)) {
				return []byte("whitelisted caller")
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				whitelistProposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Passed: true,
				})
				return whitelistProposalBytes
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
			if bytes.Equal(key, append([]byte(whiteListPrefix), callerAddress...)) {
				return []byte("whitelisted caller")
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				whitelistProposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Passed: true,
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
			if bytes.Equal(key, append([]byte(whiteListPrefix), callerAddress...)) {
				return []byte("whitelisted caller")
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				whitelistProposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Passed: true,
				})
				return whitelistProposalBytes
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				whitelistProposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Yes:    big.NewInt(10),
					No:     big.NewInt(10),
					Veto:   big.NewInt(10),
					Closed: true,
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
			if bytes.Equal(key, append([]byte(whiteListPrefix), callerAddress...)) {
				return []byte("whitelisted caller")
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				whitelistProposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Passed: true,
				})
				return whitelistProposalBytes
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				whitelistProposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Yes:          big.NewInt(10),
					No:           big.NewInt(10),
					Veto:         big.NewInt(10),
					EndVoteNonce: 10,
				})
				return whitelistProposalBytes
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
			if bytes.Equal(key, append([]byte(whiteListPrefix), callerAddress...)) {
				return []byte("whitelisted caller")
			}
			if bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...)) {
				whitelistProposalBytes, _ := args.Marshalizer.Marshal(&GeneralProposal{
					Passed: true,
				})
				return whitelistProposalBytes
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

	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_GetValidatorVotingPower(t *testing.T) {
	t.Parallel()

	votingPowerResult := make([]byte, 0)
	mockBlsKey := []byte("bls key")
	mockValidatorBlsKeys := [][]byte{
		mockBlsKey,
		mockBlsKey,
		mockBlsKey,
		mockBlsKey,
	}

	callerAddress := []byte("address")

	args := createMockGovernanceArgs()

	args.Eei = &mock.SystemEIStub{
		GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
			if bytes.Equal(address, args.ValidatorSCAddress) && bytes.Equal(key, callerAddress) {
				auctionBytes, _ := args.Marshalizer.Marshal(&ValidatorDataV2{
					BlsPubKeys:      mockValidatorBlsKeys,
					TotalStakeValue: big.NewInt(100),
				})

				return auctionBytes
			}

			return nil
		},
		FinishCalled: func(value []byte) {
			votingPowerResult = value
		},
	}
	callInputArgs := [][]byte{
		callerAddress,
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "getValidatorVotingPower", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, big.NewInt(10).Bytes(), votingPowerResult)
}

func TestGovernanceContract_GetValidatorVotingPowerWrongCallValue(t *testing.T) {
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
	callInput := createVMInput(big.NewInt(10), "getValidatorVotingPower", callerAddress, vm.GovernanceSCAddress, nil)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, vm.TransactionValueMustBeZero)
}

func TestGovernanceContract_GetValidatorVotingPowerWrongArgumentsLength(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "function accepts only one argument, the validator address"
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
	callInput := createVMInput(zero, "getValidatorVotingPower", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_GetValidatorVotingPowerInvalidArgument(t *testing.T) {
	t.Parallel()

	retMessage := ""
	errSubstr := "invalid argument - validator address"
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
	callInput := createVMInput(zero, "getValidatorVotingPower", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, errSubstr)
}

func TestGovernanceContract_GetValidatorVotingPowerComputeErr(t *testing.T) {
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
	callInput := createVMInput(zero, "getValidatorVotingPower", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.ExecutionFailed, retCode)
}

func TestGovernanceContract_GetBalanceVotingPower(t *testing.T) {
	t.Parallel()

	votingPowerResult := make([]byte, 0)

	callerAddress := []byte("address")
	args := createMockGovernanceArgs()

	args.Eei = &mock.SystemEIStub{
		FinishCalled: func(value []byte) {
			votingPowerResult = value
		},
	}
	callInputArgs := [][]byte{
		big.NewInt(400).Bytes(),
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "getBalanceVotingPower", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, big.NewInt(20).Bytes(), votingPowerResult)
}

func TestGovernanceContract_GetBalanceVotingPowerWrongCallValue(t *testing.T) {
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
	callInput := createVMInput(big.NewInt(10), "getBalanceVotingPower", callerAddress, vm.GovernanceSCAddress, nil)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, retMessage, vm.TransactionValueMustBeZero)
}

func TestGovernanceContract_GetBalanceVotingPowerWrongArgumentsLength(t *testing.T) {
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
		big.NewInt(400).Bytes(),
		big.NewInt(400).Bytes(),
	}
	callInput := createVMInput(zero, "getBalanceVotingPower", callerAddress, vm.GovernanceSCAddress, callInputArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)
	require.Contains(t, retMessage, errSubstr)
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

func TestGovernanceContract_GetValidProposalNotFound(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)

	proposal, err := gsc.getValidProposal(proposalIdentifier)
	require.Nil(t, proposal)
	require.Equal(t, vm.ErrProposalNotFound, err)
}

func TestGovernanceContract_GetValidProposalNotStarted(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	generalProposal := &GeneralProposal{
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
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
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 9
				},
			}
		},
	}
	gsc, _ := NewGovernanceContract(args)

	proposal, err := gsc.getValidProposal(proposalIdentifier)
	require.Nil(t, proposal)
	require.Equal(t, vm.ErrVotingNotStartedForProposal, err)
}

func TestGovernanceContract_GetValidProposalVotingFinished(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	generalProposal := &GeneralProposal{
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce:   15,
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
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 16
				},
			}
		},
	}
	gsc, _ := NewGovernanceContract(args)

	proposal, err := gsc.getValidProposal(proposalIdentifier)
	require.Nil(t, proposal)
	require.Equal(t, vm.ErrVotedForAnExpiredProposal, err)
}

func TestGovernanceContract_GetValidProposal(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), commitHashLength)
	generalProposal := &GeneralProposal{
		CommitHash:     proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce:   15,
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
		BlockChainHookCalled: func() vm.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 11
				},
			}
		},
	}
	gsc, _ := NewGovernanceContract(args)

	proposal, err := gsc.getValidProposal(proposalIdentifier)
	require.Nil(t, err)
	require.Equal(t, proposalIdentifier, proposal.CommitHash)
}

func TestGovernanceContract_IsWhitelistedNotFound(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			return nil
		},
	}
	gsc, _ := NewGovernanceContract(args)

	isWhiteListed := gsc.isWhiteListed([]byte("address"))
	require.False(t, isWhiteListed)
}

func TestGovernanceContract_IsWhitelistedUnmarshalErrorReturnsFalse(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			return []byte("invalid proposal")
		},
	}
	gsc, _ := NewGovernanceContract(args)

	isWhiteListed := gsc.isWhiteListed([]byte("address"))
	require.False(t, isWhiteListed)
}

func TestGovernanceContract_IsWhitelistedProposalNotVoted(t *testing.T) {
	t.Parallel()

	address := []byte("address")
	generalProposal := &GeneralProposal{
		Passed: false,
	}

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(whiteListPrefix), address...)) {
				return []byte{1}
			}

			if bytes.Equal(key, append([]byte(proposalPrefix), address...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(generalProposal)
				return proposalBytes
			}

			return nil
		},
	}
	gsc, _ := NewGovernanceContract(args)

	isWhiteListed := gsc.isWhiteListed(address)
	require.False(t, isWhiteListed)
}

func TestGovernanceContract_IsWhitelistedProposalVoted(t *testing.T) {
	t.Parallel()

	address := []byte("address")
	generalProposal := &GeneralProposal{
		Passed: true,
	}

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(whiteListPrefix), address...)) {
				return []byte{1}
			}

			if bytes.Equal(key, append([]byte(proposalPrefix), address...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(generalProposal)
				return proposalBytes
			}
			return nil
		},
	}
	gsc, _ := NewGovernanceContract(args)

	isWhiteListed := gsc.isWhiteListed([]byte("address"))
	require.True(t, isWhiteListed)
}

func TestGovernanceContract_ApplyVoteInvalid(t *testing.T) {
	t.Parallel()

	voteDetails := &VoteDetails{
		Value: 100,
	}

	voteSet := &VoteSet{}
	proposal := &GeneralProposal{}

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)

	_, _, err := gsc.applyVote(voteDetails, voteSet, proposal)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), vm.ErrInvalidArgument.Error())
}

func TestGovernanceContract_ApplyVote(t *testing.T) {
	t.Parallel()

	voteDetails := &VoteDetails{
		Value:   Yes,
		Power:   big.NewInt(10),
		Balance: big.NewInt(100),
	}

	voteSet := &VoteSet{
		UsedPower:   big.NewInt(5),
		UsedBalance: big.NewInt(25),
		TotalYes:    big.NewInt(5),
		VoteItems: []*VoteDetails{
			{
				Value:   Yes,
				Power:   big.NewInt(5),
				Balance: big.NewInt(25),
			},
		},
	}
	proposal := &GeneralProposal{
		Yes:  big.NewInt(10),
		No:   big.NewInt(10),
		Veto: big.NewInt(0),
	}

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)

	voteSetResponse, generalProposalResponse, err := gsc.applyVote(voteDetails, voteSet, proposal)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(20), generalProposalResponse.Yes)
	require.Equal(t, big.NewInt(15), voteSetResponse.TotalYes)
}

func TestGovernanceContract_ComputeAccountLeveledPower(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)

	voteSet := &VoteSet{
		UsedBalance: big.NewInt(0),
	}

	for i := 0; i < 10; i++ {
		balancedPower, _ := gsc.computeAccountLeveledPower(big.NewInt(100), voteSet)

		powerBefore := big.NewInt(0).Sqrt(voteSet.UsedBalance)
		voteSet.UsedBalance.Add(voteSet.UsedBalance, big.NewInt(100))
		powerAfter := big.NewInt(0).Sqrt(voteSet.UsedBalance)
		require.Equal(t, big.NewInt(0).Sub(powerAfter, powerBefore), balancedPower)
	}
}

func TestGovernanceContract_IsValidVoteString(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)

	require.True(t, gsc.isValidVoteString("yes"))
	require.True(t, gsc.isValidVoteString("no"))
	require.True(t, gsc.isValidVoteString("veto"))
	require.False(t, gsc.isValidVoteString("invalid"))
}

func TestComputeEndResults(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, []byte(governanceConfigKey)) {
				configBytes, _ := args.Marshalizer.Marshal(&GovernanceConfigV2{
					MinQuorum:        big.NewInt(100),
					MinPassThreshold: big.NewInt(51),
					MinVetoThreshold: big.NewInt(30),
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

func createMockStorer(callerAddress []byte, proposalIdentifier []byte, proposal *GeneralProposal) *mock.SystemEIStub {
	return &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			marshalizer := &mock.MarshalizerMock{}
			isWhiteListKey := bytes.Equal(key, append([]byte(whiteListPrefix), callerAddress...))
			if isWhiteListKey {
				whiteList, _ := marshalizer.Marshal(&WhiteListProposal{
					WhiteListAddress: callerAddress,
					ProposalStatus:   append([]byte(proposalPrefix), callerAddress...),
				})
				return whiteList
			}
			isWhiteListProposalKey := bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...))
			if isWhiteListProposalKey {
				whiteList, _ := marshalizer.Marshal(&GeneralProposal{
					Passed: true,
				})
				return whiteList
			}

			isGeneralProposalKey := bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...))
			if isGeneralProposalKey && proposal != nil {
				marshaledProposal, _ := marshalizer.Marshal(proposal)

				return marshaledProposal
			}

			return nil
		},
		GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
			marshalizer := &mock.MarshalizerMock{}
			if bytes.Equal(address, vm.DelegationManagerSCAddress) && bytes.Equal(key, []byte(delegationManagementKey)) {
				dManagementData := &DelegationManagement{MinDelegationAmount: big.NewInt(10)}
				marshaledData, _ := marshalizer.Marshal(dManagementData)
				return marshaledData
			}

			return nil
		},
	}
}
