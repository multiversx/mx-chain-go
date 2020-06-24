package systemSmartContracts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
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
			ProposalCost: "100",
		},
		ESDTSCAddress:       nil,
		Marshalizer:         &mock.MarshalizerMock{},
		Hasher:              &mock.HasherMock{},
		GovernanceSCAddress: nil,
		StakingSCAddress:    nil,
		AuctionSCAddress:    nil,
	}
}

func createVMInput(callValue *big.Int, funcName string, callerAddr, recipientAddr []byte) *vmcommon.ContractCallInput {
	return &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:  callValue,
			CallerAddr: callerAddr,
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

func TestNewGovernanceContract_ZeroBaseProposerCostShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockGovernanceArgs()
	args.GovernanceConfig.ProposalCost = ""

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
	callInput := createVMInput(big.NewInt(0), core.SCDeployInitFunctionName, callerAddr, []byte("addr2"))
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, gsc.ownerAddress, callerAddr)
}

func TestGovernanceContract_ExecuteChangeConfigCallerIsNotTheOwner(t *testing.T) {
	t.Parallel()

	callerAddr := []byte("addr1")
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			require.Equal(t, "changeConfig can be called only by owner", msg)
		},
	}
	gsc, _ := NewGovernanceContract(args)
	gsc.ownerAddress = []byte("owner")

	callInput := createVMInput(big.NewInt(0), "changeConfig", callerAddr, []byte("addr2"))

	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestGovernanceContract_ExecuteChangeConfigShouldWork(t *testing.T) {
	t.Parallel()

	numNodes := int64(10)
	minQuorum := int32(5)
	minVeto := int32(2)
	minPas := int32(3)

	callerAddr := []byte("addr1")
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			configBytes, _ := json.Marshal(&GovernanceConfig{})
			return configBytes
		},
		SetStorageCalled: func(key []byte, value []byte) {
			newConfig := &GovernanceConfig{}
			_ = json.Unmarshal(value, newConfig)
			require.Equal(t, numNodes, newConfig.NumNodes)
			require.Equal(t, minQuorum, newConfig.MinQuorum)
			require.Equal(t, minVeto, newConfig.MinVetoThreshold)
			require.Equal(t, minPas, newConfig.MinPassThreshold)
		},
	}
	gsc, _ := NewGovernanceContract(args)
	gsc.ownerAddress = callerAddr

	callInput := createVMInput(big.NewInt(0), "changeConfig", callerAddr, []byte("addr2"))
	callInput.Arguments = [][]byte{
		[]byte(fmt.Sprintf("%d", numNodes)),
		[]byte(fmt.Sprintf("%d", minQuorum)),
		[]byte(fmt.Sprintf("%d", minVeto)),
		[]byte(fmt.Sprintf("%d", minPas)),
	}

	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_ExecuteWhiteListProposalInvalidValueShouldErr(t *testing.T) {
	t.Parallel()

	callerAddr := []byte("addr1")
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 1
				},
			}
		},
		AddReturnMessageCalled: func(msg string) {
			require.True(t, strings.Contains(msg, "invalid callValue, needs exactly"))
		},
	}
	gsc, _ := NewGovernanceContract(args)
	gsc.ownerAddress = callerAddr

	callInput := createVMInput(big.NewInt(10), "whiteList", callerAddr, []byte("addr2"))
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.OutOfFunds, retCode)
}

func TestGovernanceContract_ExecuteWhiteListProposalAtGenesisShouldWork(t *testing.T) {
	t.Parallel()

	callerAddr := []byte("addr1")
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		BlockChainHookCalled: func() vmcommon.BlockchainHook {
			return &mock.BlockChainHookStub{
				CurrentNonceCalled: func() uint64 {
					return 0
				},
			}
		},
		SetStorageCalled: func(key []byte, value []byte) {
			if !strings.Contains(string(key), whiteListPrefix) {
				genProposal := &GeneralProposal{}
				_ = json.Unmarshal(value, genProposal)
				require.Equal(t, []byte("genesis"), genProposal.GitHubCommit)
				require.True(t, genProposal.Voted)
				require.Equal(t, callerAddr, genProposal.IssuerAddress)

				return
			}

			whiteListProp := &WhiteListProposal{}
			_ = json.Unmarshal(value, whiteListProp)
			require.Equal(t, whiteListProp.WhiteListAddress, callerAddr)
		},
	}
	gsc, _ := NewGovernanceContract(args)
	gsc.ownerAddress = callerAddr

	callInput := createVMInput(big.NewInt(0), "whiteList", callerAddr, []byte("addr2"))
	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_ExecuteWhiteListProposalShouldWork(t *testing.T) {
	t.Parallel()

	gitHubCommit := []byte("0123456789012345678901234567890123456789")
	startNonce := uint64(100)
	stopNonce := uint64(1000)
	callerAddr := []byte("addr1")
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
			if strings.Contains(string(key), string(gitHubCommit)) {
				genProposal := &GeneralProposal{}
				_ = json.Unmarshal(value, genProposal)
				require.Equal(t, gitHubCommit, genProposal.GitHubCommit)
				require.Equal(t, startNonce, genProposal.StartVoteNonce)
				require.Equal(t, stopNonce, genProposal.EndVoteNonce)
				require.Equal(t, callerAddr, genProposal.IssuerAddress)
				require.False(t, genProposal.Voted)

				return
			}

			whiteListProp := &WhiteListProposal{}
			_ = json.Unmarshal(value, whiteListProp)
			require.Equal(t, whiteListProp.WhiteListAddress, callerAddr)
		},
	}
	gsc, _ := NewGovernanceContract(args)
	gsc.ownerAddress = callerAddr

	callInput := createVMInput(big.NewInt(100), "whiteList", callerAddr, []byte("addr2"))
	callInput.Arguments = [][]byte{
		gitHubCommit,
		[]byte(fmt.Sprintf("%d", startNonce)),
		[]byte(fmt.Sprintf("%d", stopNonce)),
	}

	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_ExecuteHardfork(t *testing.T) {
	t.Parallel()

	gitHubCommit := []byte("0123456789012345678901234567890123456789")
	startNonce := uint64(100)
	stopNonce := uint64(1000)
	epochStartHardfork := []byte("1")
	newSoftwareVersion := []byte("version")
	callerAddr := []byte("addr1")
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if strings.Contains(string(key), string(gitHubCommit)) {
				return []byte("")
			}

			generalProposal := &GeneralProposal{
				Voted: true,
			}
			generalProposalBytes, _ := json.Marshal(generalProposal)
			return generalProposalBytes
		},
		SetStorageCalled: func(key []byte, value []byte) {
			if bytes.Equal(key, append([]byte(hardForkPrefix), gitHubCommit...)) {
				hardForkProposal := &HardForkProposal{}
				_ = json.Unmarshal(value, hardForkProposal)
				require.Equal(t, uint32(1), hardForkProposal.EpochToHardFork)
				require.Equal(t, newSoftwareVersion, hardForkProposal.NewSoftwareVersion)

				return
			}

			genProposal := &GeneralProposal{}
			_ = json.Unmarshal(value, genProposal)
			require.Equal(t, gitHubCommit, genProposal.GitHubCommit)
			require.Equal(t, startNonce, genProposal.StartVoteNonce)
			require.Equal(t, stopNonce, genProposal.EndVoteNonce)
			require.Equal(t, callerAddr, genProposal.IssuerAddress)
			require.False(t, genProposal.Voted)

			return
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(100), "hardFork", callerAddr, []byte("addr2"))
	callInput.Arguments = [][]byte{
		epochStartHardfork,
		newSoftwareVersion,
		gitHubCommit,
		[]byte(fmt.Sprintf("%d", startNonce)),
		[]byte(fmt.Sprintf("%d", stopNonce)),
	}

	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_ExecuteProposal(t *testing.T) {
	t.Parallel()

	gitHubCommit := []byte("0123456789012345678901234567890123456789")
	startNonce := uint64(100)
	stopNonce := uint64(1000)
	callerAddr := []byte("addr1")
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if strings.Contains(string(key), string(gitHubCommit)) {
				return []byte("")
			}

			generalProposal := &GeneralProposal{
				Voted: true,
			}
			generalProposalBytes, _ := json.Marshal(generalProposal)
			return generalProposalBytes
		},
		SetStorageCalled: func(key []byte, value []byte) {
			genProposal := &GeneralProposal{}
			_ = json.Unmarshal(value, genProposal)
			require.Equal(t, gitHubCommit, genProposal.GitHubCommit)
			require.Equal(t, startNonce, genProposal.StartVoteNonce)
			require.Equal(t, stopNonce, genProposal.EndVoteNonce)
			require.Equal(t, callerAddr, genProposal.IssuerAddress)
			require.False(t, genProposal.Voted)

			return
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(100), "proposal", callerAddr, []byte("addr2"))
	callInput.Arguments = [][]byte{
		gitHubCommit,
		[]byte(fmt.Sprintf("%d", startNonce)),
		[]byte(fmt.Sprintf("%d", stopNonce)),
	}

	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestGovernanceContract_ExecuteVoteYes(t *testing.T) {
	testExecuteVote(t, []byte("yes"))
}

func TestGovernanceContract_ExecuteVoteNo(t *testing.T) {
	testExecuteVote(t, []byte("no"))
}

func TestGovernanceContract_ExecuteVoteVeto(t *testing.T) {
	testExecuteVote(t, []byte("veto"))
}

func TestGovernanceContract_ExecuteVoteDontCare(t *testing.T) {
	testExecuteVote(t, []byte("dontCare"))
}

func testExecuteVote(t *testing.T, vote []byte) {
	t.Parallel()

	proposalToVote := []byte("proposalToVote")
	validatorAddr := []byte("addr2")
	callerAddr := []byte("addr1")
	args := createMockGovernanceArgs()
	args.Eei = &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			if bytes.Equal(key, append([]byte(validatorPrefix), validatorAddr...)) {
				validatorData := &ValidatorData{
					Delegators: []*VoterData{
						{Address: callerAddr, NumNodes: 1},
					},
				}

				validatorDataBytes, _ := json.Marshal(validatorData)
				return validatorDataBytes
			}
			generalProposal := &GeneralProposal{
				Voted: true,
			}
			generalProposalBytes, _ := json.Marshal(generalProposal)
			return generalProposalBytes
		},
		ExecuteOnDestContextCalled: func(destination, sender []byte, value *big.Int, input []byte) (output *vmcommon.VMOutput, err error) {
			autionData := &AuctionData{
				NumStaked: 1,
			}
			auctionDataBytes, _ := json.Marshal(autionData)

			return &vmcommon.VMOutput{
				ReturnCode: vmcommon.Ok,
				ReturnData: [][]byte{auctionDataBytes},
			}, nil
		},
	}

	gsc, _ := NewGovernanceContract(args)
	callInput := createVMInput(big.NewInt(0), "vote", callerAddr, []byte("addr2"))
	callInput.Arguments = [][]byte{
		proposalToVote,
		vote,
		validatorAddr,
	}

	retCode := gsc.Execute(callInput)
	require.Equal(t, vmcommon.Ok, retCode)
}
