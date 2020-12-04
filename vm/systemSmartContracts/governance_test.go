package systemSmartContracts

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
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
		Marshalizer:         &mock.MarshalizerMock{},
		Hasher:              &mock.HasherMock{},
		GovernanceSCAddress: []byte("governanceSC"),
		StakingSCAddress:    []byte("stakingSC"),
		AuctionSCAddress:    []byte("auctionSC"),
		EpochNotifier:       &mock.EpochNotifierStub{},
	}
}

func createVMInput(callValue *big.Int, funcName string, callerAddr, recipientAddr []byte, arguments [][]byte) *vmcommon.ContractCallInput {
	return &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:  callValue,
			CallerAddr: callerAddr,
			Arguments: arguments,
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

func TestGovernanceContract_ExecuteNilVMInputShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)

	retCode := gsc.Execute(nil)
	require.Equal(t, vmcommon.UserError, retCode)
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
	retCode := gsc.Execute(callInput)
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
	proposalArgs := [][]byte{
		[]byte("arg1"),
		[]byte("arg2"),
		[]byte("arg3"),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), proposalArgs)
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
				whitelistProposal.Voted = true
			}
			return nil
		},
	}
	gsc, _ := NewGovernanceContract(args)
	proposalArgs := [][]byte{
		[]byte("arg1"),
		[]byte("arg2"),
		[]byte("arg3"),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), proposalArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_ProposalAlreadyExists(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)

	args := createMockGovernanceArgs()
	args.Eei = createMockStorer(vm.GovernanceSCAddress, proposalIdentifier, &GeneralProposal{})
	gsc, _ := NewGovernanceContract(args)
	proposalArgs := [][]byte{
		proposalIdentifier,
		[]byte("arg2"),
		[]byte("arg3"),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), proposalArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_ProposalInvalidVoteNonce(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)

	args := createMockGovernanceArgs()
	args.Eei = createMockStorer(vm.GovernanceSCAddress, proposalIdentifier, nil)
	gsc, _ := NewGovernanceContract(args)
	proposalArgs := [][]byte{
		proposalIdentifier,
		[]byte("arg2"),
		[]byte("arg3"),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), proposalArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_ProposalOK(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)

	args := createMockGovernanceArgs()
	args.Eei = createMockStorer(vm.GovernanceSCAddress, proposalIdentifier, nil)
	gsc, _ := NewGovernanceContract(args)
	proposalArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
		[]byte("10"),
	}
	callInput := createVMInput(big.NewInt(500), "proposal", vm.GovernanceSCAddress, []byte("addr1"), proposalArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_AccountVoteNotEnoughGas(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		UseGasCalled: func(_ uint64) error {
			return errors.New("not enough gas")
		},
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(500), "vote", vm.GovernanceSCAddress, []byte("addr1"), make([][]byte, 0))
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.OutOfGas, retCode)
}

func TestGovernanceContract_AccountVoteInvalidNrOfArguments(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)
	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
		[]byte("10"),
	}
	callInput := createVMInput(big.NewInt(500), "vote", vm.GovernanceSCAddress, []byte("addr1"), voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)
}

func TestGovernanceContract_AccountVoteProposalNotFound(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)
	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
	}
	callInput := createVMInput(big.NewInt(500), "vote", vm.GovernanceSCAddress, []byte("addr1"), voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_AccountVoteInvalidVoteType(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)

	args := createMockGovernanceArgs()
	args.Eei = createMockStorer(vm.GovernanceSCAddress, proposalIdentifier, &GeneralProposal{})
	gsc, _ := NewGovernanceContract(args)
	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
	}
	callInput := createVMInput(big.NewInt(500), "vote", vm.GovernanceSCAddress, []byte("addr1"), voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_AccountVoteInvalidCallValue(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)

	args := createMockGovernanceArgs()
	args.Eei = createMockStorer(vm.GovernanceSCAddress, proposalIdentifier, &GeneralProposal{})
	gsc, _ := NewGovernanceContract(args)
	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("yes"),
	}
	callInput := createVMInput(big.NewInt(-500), "vote", vm.GovernanceSCAddress, []byte("addr1"), voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_AccountVoteAddVoteError(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	defaultMarshalizer := &mock.MarshalizerMock{}

	args := createMockGovernanceArgs()
	args.Eei = createMockStorer(vm.GovernanceSCAddress, proposalIdentifier, &GeneralProposal{
		Yes: big.NewInt(0),
		No: big.NewInt(0),
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
	callInput := createVMInput(big.NewInt(500), "vote", vm.GovernanceSCAddress, []byte("addr1"), voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_AccountVoteAddSimpleVote(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)

	args := createMockGovernanceArgs()
	args.Eei = createMockStorer(vm.GovernanceSCAddress, proposalIdentifier, &GeneralProposal{
		Yes: big.NewInt(0),
		No: big.NewInt(0),
	})
	gsc, _ := NewGovernanceContract(args)
	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("yes"),
	}
	callInput := createVMInput(big.NewInt(500), "vote", vm.GovernanceSCAddress, []byte("addr1"), voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_ValidatorVoteNotEnoughGas(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	args := createMockGovernanceArgs()
	voteArgs := [][]byte{
		proposalIdentifier,
		[]byte("yes"),
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "vote", vm.GovernanceSCAddress, []byte("addr1"), voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)

	voteArgs = [][]byte{
		proposalIdentifier,
		[]byte("yes"),
		[]byte("third"),
		[]byte("fourth"),
		[]byte("fifth"),
	}

	callInput = createVMInput(big.NewInt(0), "vote", vm.GovernanceSCAddress, []byte("addr1"), voteArgs)
	retCode = gsc.Execute(callInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)
}

func TestGovernanceContract_ValidatorVoteNotEnoughArguments(t *testing.T) {
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
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		GitHubCommit: proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce: 15,
	}
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(generalProposal)
				return proposalBytes
			}

			return nil
		},
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
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
		[]byte("third"),
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
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		GitHubCommit: proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce: 15,
	}
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(generalProposal)
				return proposalBytes
			}

			return nil
		},
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
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
		[]byte("third"),
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
	errInvalidVoteSubstr := "invalid delegator address length"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		GitHubCommit: proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce: 15,
	}
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...)) {
				proposalBytes, _ := args.Marshalizer.Marshal(generalProposal)
				return proposalBytes
			}

			return nil
		},
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
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
		[]byte("third"),
		[]byte("fourth"),
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, returnMessage, errInvalidVoteSubstr)
}

func TestGovernanceContract_ValidatorVoteComputePowerError(t *testing.T) {
	t.Parallel()

	returnMessage := ""
	errInvalidVoteSubstr := "could not return total stake for the provided address"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		GitHubCommit: proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce: 15,
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
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
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
		[]byte("third"),
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, returnMessage, errInvalidVoteSubstr)
}

func TestGovernanceContract_ValidatorVoteVoteSetError(t *testing.T) {
	t.Parallel()

	mockBlsKey := []byte("bls key")
	mockValidatorBlsKeys := [][]byte{
		mockBlsKey,
		mockBlsKey,
		mockBlsKey,
		mockBlsKey,
	}

	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	votePower := big.NewInt(100).Bytes()

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		GitHubCommit: proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce: 15,
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
			if bytes.Equal(address, args.AuctionSCAddress) && bytes.Equal(key, callerAddress) {
				auctionBytes, _ := args.Marshalizer.Marshal(&AuctionDataV2{
					BlsPubKeys: mockValidatorBlsKeys,
				})

				return auctionBytes
			}

			if bytes.Equal(address, args.StakingSCAddress) && bytes.Equal(key, mockBlsKey) {
				stakeDataBytes, _ := args.Marshalizer.Marshal(&StakedDataV2_0{
					Staked: true,
					StakeValue: big.NewInt(0).SetBytes(votePower),
				})

				return stakeDataBytes
			}

			return nil
		},
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
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
		votePower,
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.ExecutionFailed, retCode)
}

func TestGovernanceContract_ValidatorVoteVoteNotEnoughPower(t *testing.T) {
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

	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	votePower := big.NewInt(100).Bytes()

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		GitHubCommit: proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce: 15,
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
			if bytes.Equal(address, args.AuctionSCAddress) && bytes.Equal(key, callerAddress) {
				auctionBytes, _ := args.Marshalizer.Marshal(&AuctionDataV2{
					BlsPubKeys: mockValidatorBlsKeys,
				})

				return auctionBytes
			}

			if bytes.Equal(address, args.StakingSCAddress) && bytes.Equal(key, mockBlsKey) {
				stakeDataBytes, _ := args.Marshalizer.Marshal(&StakedDataV2_0{
					Staked: true,
					StakeValue: big.NewInt(0).SetBytes(votePower),
				})

				return stakeDataBytes
			}

			return nil
		},
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
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
		votePower,
	}
	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "vote", callerAddress, vm.GovernanceSCAddress, voteArgs)
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, returnMessage, errInvalidVoteSubstr)
}

func TestGovernanceContract_ValidatorVoteVote(t *testing.T) {
	t.Parallel()

	mockBlsKey := []byte("bls key")
	mockValidatorBlsKeys := [][]byte{
		mockBlsKey,
		mockBlsKey,
		mockBlsKey,
		mockBlsKey,
	}

	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	votePower := big.NewInt(10)
	proposalKey := append([]byte(proposalPrefix), proposalIdentifier...)
	voteItemKey := append(proposalKey, callerAddress...)

	finalVoteSet := &VoteSet{}
	finalProposal := &GeneralProposal{}

	args := createMockGovernanceArgs()

	generalProposal := &GeneralProposal{
		GitHubCommit: proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce: 15,
		Yes: big.NewInt(0),
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
			if bytes.Equal(address, args.AuctionSCAddress) && bytes.Equal(key, callerAddress) {
				auctionBytes, _ := args.Marshalizer.Marshal(&AuctionDataV2{
					BlsPubKeys: mockValidatorBlsKeys,
				})

				return auctionBytes
			}

			if bytes.Equal(address, args.StakingSCAddress) && bytes.Equal(key, mockBlsKey) {
				stakeDataBytes, _ := args.Marshalizer.Marshal(&StakedDataV2_0{
					Staked: true,
					StakeValue: big.NewInt(100),
				})

				return stakeDataBytes
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
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
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
		votePower.Bytes(),
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

func TestGovernanceContract_ClaimFundsStillLocked(t *testing.T) {
	t.Parallel()

	returnMessage := ""
	expectedErrorSubstr := "your funds are still locked"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			returnMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			expectedKeyPrefix := append([]byte(accountLockPrefix), proposalIdentifier...)
			if bytes.Equal(key, append(expectedKeyPrefix, callerAddress...)) {
				return big.NewInt(10).Bytes()
			}

			return nil
		},
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
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

func TestGovernanceContract_ClaimFundsAlreadyClaimed(t *testing.T) {
	t.Parallel()

	returnMessage := ""
	expectedErrorSubstr := "you already claimed back your funds"
	callerAddress := []byte("address")
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			returnMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			expectedKeyPrefix := append([]byte(accountLockPrefix), proposalIdentifier...)
			if bytes.Equal(key, append(expectedKeyPrefix, callerAddress...)) {
				return big.NewInt(10).Bytes()
			}

			if bytes.Equal(key, append(proposalIdentifier, callerAddress...)) {
				voteSetBytes, _ := args.Marshalizer.Marshal(&VoteSet{
					Claimed: true,
				})
				return voteSetBytes
			}

			return nil
		},
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 9
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
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			returnMessage = msg
		},
		GetStorageCalled: func(key []byte) []byte {
			expectedKeyPrefix := append([]byte(accountLockPrefix), proposalIdentifier...)
			if bytes.Equal(key, append(expectedKeyPrefix, callerAddress...)) {
				return big.NewInt(10).Bytes()
			}

			if bytes.Equal(key, append(proposalIdentifier, callerAddress...)) {
				voteSetBytes, _ := args.Marshalizer.Marshal(&VoteSet{
					Claimed: false,
					UsedBalance: zero,
				})
				return voteSetBytes
			}

			return nil
		},
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 9
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
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)

	finalVoteSet := &VoteSet{}
	transferFrom := make([]byte, 0)
	transferTo := make([]byte, 0)
	transferValue := big.NewInt(0)

	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			expectedKeyPrefix := append([]byte(accountLockPrefix), proposalIdentifier...)
			if bytes.Equal(key, append(expectedKeyPrefix, callerAddress...)) {
				return big.NewInt(10).Bytes()
			}

			if bytes.Equal(key, append(proposalIdentifier, callerAddress...)) {
				voteSetBytes, _ := args.Marshalizer.Marshal(&VoteSet{
					Claimed: false,
					UsedBalance: voteValue,
				})
				return voteSetBytes
			}

			return nil
		},
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 9
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
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
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

	proposalArgs := [][]byte{
		proposalIdentifier,
		[]byte("1"),
		[]byte("10"),
	}
	proposalCost, _ := big.NewInt(0).SetString(args.GovernanceConfig.Active.ProposalCost, conversionBase)
	callInput := createVMInput(proposalCost, "whiteList", callerAddress, vm.GovernanceSCAddress, proposalArgs)
	retCode := gsc.Execute(callInput)

	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, callerAddress, finalWhitelistProposal.WhiteListAddress)
	require.Equal(t, append([]byte(proposalPrefix), callerAddress...), finalWhitelistProposal.ProposalStatus)
	require.Equal(t, proposalIdentifier, finalProposal.GitHubCommit)
}

// ========  Begin testing of helper functions

func TestGovernanceContract_GetGeneralProposalNotFound(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)

	proposal, err := gsc.getGeneralProposal(proposalIdentifier)
	require.Nil(t, proposal)
	require.Equal(t, vm.ErrProposalNotFound, err)
}

func TestGovernanceContract_GetGeneralProposalUnmarshalErr(t *testing.T) {
	t.Parallel()

	unmarshalErr := errors.New("unmarshal error")
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
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

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	generalProposal := &GeneralProposal{
		GitHubCommit: proposalIdentifier,
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
	require.Equal(t, proposalIdentifier, proposal.GitHubCommit)
}

func TestGovernanceContract_SaveGeneralProposalUnmarshalErr(t *testing.T) {
	t.Parallel()

	customErr := errors.New("unmarshal error")
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	generalProposal := &GeneralProposal{
		Yes: big.NewInt(10),
		No: big.NewInt(0),
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
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	generalProposal := &GeneralProposal{
		Yes: big.NewInt(10),
		No: big.NewInt(0),
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

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)

	proposal, err := gsc.getValidProposal(proposalIdentifier)
	require.Nil(t, proposal)
	require.Equal(t, vm.ErrProposalNotFound, err)
}

func TestGovernanceContract_GetValidProposalNotStarted(t *testing.T) {
	t.Parallel()

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	generalProposal := &GeneralProposal{
		GitHubCommit: proposalIdentifier,
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
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
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

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	generalProposal := &GeneralProposal{
		GitHubCommit: proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce: 15,
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
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
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

	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	generalProposal := &GeneralProposal{
		GitHubCommit: proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce: 15,
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
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
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
	require.Equal(t, proposalIdentifier, proposal.GitHubCommit)
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
		Voted: false,
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
		Voted: true,
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
		Value: Yes,
		Power: big.NewInt(10),
		Balance: big.NewInt(100),
		Type: Account,
	}

	voteSet := &VoteSet{
		UsedPower: big.NewInt(5),
		UsedBalance: big.NewInt(25),
		TotalYes: big.NewInt(5),
		VoteItems: []*VoteDetails{
			{
				Value: Yes,
				Power: big.NewInt(5),
				Balance: big.NewInt(25),
				Type: Account,
			},
		},
	}
	proposal := &GeneralProposal{
		Yes: big.NewInt(10),
		No: big.NewInt(10),
		Veto: big.NewInt(0),
	}

	args := createMockGovernanceArgs()
	gsc, _ := NewGovernanceContract(args)

	voteSetResponse, generalProposalResponse, err := gsc.applyVote(voteDetails, voteSet, proposal)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(20), generalProposalResponse.Yes)
	require.Equal(t, big.NewInt(15), voteSetResponse.TotalYes)
}

func TestGovernanceContract_SetLock(t *testing.T) {
	t.Parallel()
	args := createMockGovernanceArgs()

	storageKeyCalled := make([]byte, 0)
	storageValueCalled := make([]byte, 0)
	args.Eei = &mock.SystemEIStub{
		SetStorageCalled: func(key []byte, value []byte) {
			storageKeyCalled = key
			storageValueCalled = value
		},
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 14
				},
			}
		},
	}
	gsc, _ := NewGovernanceContract(args)

	voter := []byte("voter")
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	proposal := &GeneralProposal{
		GitHubCommit: proposalIdentifier,
		StartVoteNonce: 10,
		EndVoteNonce: 15,
	}

	gsc.setLock(voter, Validator, proposal)
	require.Equal(t, append([]byte(validatorLockPrefix), voter...), storageKeyCalled)
	require.Equal(t, big.NewInt(0).SetUint64(19).Bytes(), storageValueCalled)


	accPrefix := append([]byte(accountLockPrefix), proposalIdentifier...)
	gsc.setLock(voter, Account, proposal)
	require.Equal(t, append(accPrefix, voter...), storageKeyCalled)
	require.Equal(t, big.NewInt(0).SetUint64(19).Bytes(), storageValueCalled)
}

func TestGovernanceContract_GetLock(t *testing.T) {
	t.Parallel()
	args := createMockGovernanceArgs()

	voter := []byte("voter")
	proposalIdentifier := bytes.Repeat([]byte("a"), githubCommitLength)
	lockValue := uint64(10)
	storageKeyCalled := make([]byte, 0)
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			storageKeyCalled = key
			return big.NewInt(0).SetUint64(lockValue).Bytes()
		},
	}
	gsc, _ := NewGovernanceContract(args)
	lockReturned := gsc.getLock(voter, Validator, proposalIdentifier)
	require.Equal(t, append([]byte(validatorLockPrefix), voter...), storageKeyCalled)
	require.Equal(t, lockValue, lockReturned)

	accPrefix := append([]byte(accountLockPrefix), proposalIdentifier...)
	lockReturned = gsc.getLock(voter, Account, proposalIdentifier)
	require.Equal(t, append(accPrefix, voter...), storageKeyCalled)
	require.Equal(t, lockValue, lockReturned)
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

func createMockStorer(callerAddress []byte, proposalIdentifier []byte, proposal *GeneralProposal) *mock.SystemEIStub {
	return &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			marshalizer := &mock.MarshalizerMock{}
			isWhiteListKey := bytes.Equal(key, append([]byte(whiteListPrefix), callerAddress...))
			if isWhiteListKey {
				whiteList, _ := marshalizer.Marshal(&WhiteListProposal{
					WhiteListAddress: callerAddress,
					ProposalStatus: append([]byte(proposalPrefix), callerAddress...),
				})
				return whiteList
			}
			isWhiteListProposalKey := bytes.Equal(key, append([]byte(proposalPrefix), callerAddress...))
			if isWhiteListProposalKey {
				whiteList, _ := marshalizer.Marshal(&GeneralProposal{
					Voted: true,
				})
				return whiteList
			}

			isGeneralProposalKey := bytes.Equal(key, append([]byte(proposalPrefix), proposalIdentifier...))
			if isGeneralProposalKey && proposal != nil{
				marshaledProposal, _ := marshalizer.Marshal(proposal)

				return marshaledProposal
			}

			return nil
		},
	}
}
