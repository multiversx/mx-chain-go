//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. governance.proto
package systemSmartContracts

import (
	"bytes"
	"fmt"
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const governanceConfigKey = "governanceConfig"
const proposalPrefix = "p_"
const stakeLockPrefix = "s_"
const yesString = "yes"
const noString = "no"
const vetoString = "veto"
const abstainString = "abstain"
const commitHashLength = 40

// ArgsNewGovernanceContract defines the arguments needed for the on-chain governance contract
type ArgsNewGovernanceContract struct {
	Eei                    vm.SystemEI
	GasCost                vm.GasCost
	GovernanceConfig       config.GovernanceSystemSCConfig
	Marshalizer            marshal.Marshalizer
	Hasher                 hashing.Hasher
	GovernanceSCAddress    []byte
	DelegationMgrSCAddress []byte
	ValidatorSCAddress     []byte
	EnableEpochsHandler    common.EnableEpochsHandler
}

type governanceContract struct {
	eei                    vm.SystemEI
	gasCost                vm.GasCost
	baseProposalCost       *big.Int
	ownerAddress           []byte
	governanceSCAddress    []byte
	delegationMgrSCAddress []byte
	validatorSCAddress     []byte
	marshalizer            marshal.Marshalizer
	hasher                 hashing.Hasher
	governanceConfig       config.GovernanceSystemSCConfig
	enableEpochsHandler    common.EnableEpochsHandler
	mutExecution           sync.RWMutex
}

// NewGovernanceContract creates a new governance smart contract
func NewGovernanceContract(args ArgsNewGovernanceContract) (*governanceContract, error) {
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, vm.ErrNilHasher
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, vm.ErrNilEnableEpochsHandler
	}

	activeConfig := args.GovernanceConfig.Active
	baseProposalCost, okConvert := big.NewInt(0).SetString(activeConfig.ProposalCost, conversionBase)
	if !okConvert || baseProposalCost.Cmp(zero) < 0 {
		return nil, vm.ErrInvalidBaseIssuingCost
	}

	if len(args.ValidatorSCAddress) < 1 {
		return nil, fmt.Errorf("%w for validator sc address", vm.ErrInvalidAddress)
	}
	if len(args.DelegationMgrSCAddress) < 1 {
		return nil, fmt.Errorf("%w for delegation sc address", vm.ErrInvalidAddress)
	}
	if len(args.GovernanceSCAddress) < 1 {
		return nil, fmt.Errorf("%w for governance sc address", vm.ErrInvalidAddress)
	}

	g := &governanceContract{
		eei:                    args.Eei,
		gasCost:                args.GasCost,
		baseProposalCost:       baseProposalCost,
		ownerAddress:           nil,
		governanceSCAddress:    args.GovernanceSCAddress,
		delegationMgrSCAddress: args.DelegationMgrSCAddress,
		validatorSCAddress:     args.ValidatorSCAddress,
		marshalizer:            args.Marshalizer,
		hasher:                 args.Hasher,
		governanceConfig:       args.GovernanceConfig,
		enableEpochsHandler:    args.EnableEpochsHandler,
	}

	return g, nil
}

// Execute calls one of the functions from the governance smart contract and runs the code according to the input
func (g *governanceContract) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	g.mutExecution.RLock()
	defer g.mutExecution.RUnlock()
	if CheckIfNil(args) != nil {
		return vmcommon.UserError
	}

	if args.Function == core.SCDeployInitFunctionName {
		return g.init(args)
	}

	if !g.enableEpochsHandler.IsGovernanceFlagEnabled() {
		g.eei.AddReturnMessage("Governance SC disabled")
		return vmcommon.UserError
	}

	if len(args.ESDTTransfers) > 0 {
		g.eei.AddReturnMessage("cannot transfer ESDT to system SCs")
		return vmcommon.UserError
	}

	switch args.Function {
	case "initV2":
		return g.initV2(args)
	case "proposal":
		return g.proposal(args)
	case "vote":
		return g.vote(args)
	case "delegateVote":
		return g.delegateVote(args)
	case "changeConfig":
		return g.changeConfig(args)
	case "closeProposal":
		return g.closeProposal(args)
	case "getValidatorVotingPower":
		return g.getValidatorVotingPower(args)
	}

	g.eei.AddReturnMessage("invalid method to call")
	return vmcommon.FunctionNotFound
}

func (g *governanceContract) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	scConfig := &GovernanceConfig{
		NumNodes:         g.governanceConfig.V1.NumNodes,
		MinQuorum:        g.governanceConfig.V1.MinQuorum,
		MinPassThreshold: g.governanceConfig.V1.MinPassThreshold,
		MinVetoThreshold: g.governanceConfig.V1.MinVetoThreshold,
		ProposalFee:      g.baseProposalCost,
	}
	marshaledData, err := g.marshalizer.Marshal(scConfig)
	log.LogIfError(err, "function", "governanceContract.init")

	g.eei.SetStorage([]byte(governanceConfigKey), marshaledData)
	g.eei.SetStorage([]byte(ownerKey), args.CallerAddr)
	g.ownerAddress = make([]byte, 0, len(args.CallerAddr))
	g.ownerAddress = append(g.ownerAddress, args.CallerAddr...)
	return vmcommon.Ok
}

func (g *governanceContract) initV2(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, g.governanceSCAddress) {
		log.Error("invalid caller to switch to V2 config")
		return vmcommon.UserError
	}
	cfg, err := g.convertV2Config(g.governanceConfig)
	if err != nil {
		log.Error("could not create governance V2 config")
		return vmcommon.UserError
	}

	err = g.saveConfig(cfg)
	if err != nil {
		log.Error(err.Error())
		return vmcommon.UserError
	}

	g.eei.SetStorage([]byte(ownerKey), args.CallerAddr)
	g.ownerAddress = make([]byte, 0, len(args.CallerAddr))
	g.ownerAddress = append(g.ownerAddress, args.CallerAddr...)

	return vmcommon.Ok
}

// changeConfig allows the owner to change the configuration for requesting proposals
func (g *governanceContract) changeConfig(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(g.ownerAddress, args.CallerAddr) {
		g.eei.AddReturnMessage("changeConfig can be called only by owner")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		g.eei.AddReturnMessage("changeConfig can be called only without callValue")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 4 {
		g.eei.AddReturnMessage("changeConfig needs 4 arguments")
		return vmcommon.UserError
	}

	proposalFee, okConvert := big.NewInt(0).SetString(string(args.Arguments[0]), conversionBase)
	if !okConvert || proposalFee.Cmp(zero) < 0 {
		g.eei.AddReturnMessage("changeConfig first argument is incorrectly formatted")
		return vmcommon.UserError
	}
	minQuorum, okConvert := big.NewInt(0).SetString(string(args.Arguments[1]), conversionBase)
	if !okConvert || minQuorum.Cmp(zero) < 0 {
		g.eei.AddReturnMessage("changeConfig second argument is incorrectly formatted")
		return vmcommon.UserError
	}
	minVeto, okConvert := big.NewInt(0).SetString(string(args.Arguments[2]), conversionBase)
	if !okConvert || minVeto.Cmp(zero) < 0 {
		g.eei.AddReturnMessage("changeConfig third argument is incorrectly formatted")
		return vmcommon.UserError
	}
	minPass, okConvert := big.NewInt(0).SetString(string(args.Arguments[3]), conversionBase)
	if !okConvert || minPass.Cmp(zero) < 0 {
		g.eei.AddReturnMessage("changeConfig fourth argument is incorrectly formatted")
		return vmcommon.UserError
	}

	scConfig, err := g.getConfig()
	if err != nil {
		g.eei.AddReturnMessage("changeConfig error " + err.Error())
		return vmcommon.UserError
	}

	scConfig.MinQuorum = minQuorum
	scConfig.MinVetoThreshold = minVeto
	scConfig.MinPassThreshold = minPass
	scConfig.ProposalFee = proposalFee

	err = g.saveConfig(scConfig)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (g *governanceContract) saveConfig(cfg *GovernanceConfigV2) error {
	marshaledData, err := g.marshalizer.Marshal(cfg)
	if err != nil {
		return err
	}

	g.eei.SetStorage([]byte(governanceConfigKey), marshaledData)
	return nil
}

// proposal creates a new proposal from passed arguments
func (g *governanceContract) proposal(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := g.eei.UseGas(g.gasCost.MetaChainSystemSCsCost.Proposal)
	if err != nil {
		g.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 3 {
		g.eei.AddReturnMessage("invalid number of arguments, expected 3")
		return vmcommon.FunctionWrongSignature
	}
	generalConfig, err := g.getConfig()
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(generalConfig.ProposalFee) != 0 {
		g.eei.AddReturnMessage("invalid proposal cost, expected " + g.baseProposalCost.String())
		return vmcommon.OutOfFunds
	}

	generalConfig.LastProposalNonce++
	nextNonce := generalConfig.LastProposalNonce
	err = g.saveConfig(generalConfig)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	commitHash := args.Arguments[0]
	if len(commitHash) != commitHashLength {
		g.eei.AddReturnMessage(fmt.Sprintf("invalid github commit length, wanted exactly %d", commitHashLength))
		return vmcommon.UserError
	}
	if g.proposalExists(commitHash) {
		g.eei.AddReturnMessage("proposal already exists")
		return vmcommon.UserError
	}

	startVoteNonce, endVoteNonce, err := g.startEndNonceFromArguments(args.Arguments[1], args.Arguments[2])
	if err != nil {
		g.eei.AddReturnMessage("invalid start/end vote nonce " + err.Error())
		return vmcommon.UserError
	}

	generalProposal := &GeneralProposal{
		IssuerAddress:  args.CallerAddr,
		CommitHash:     commitHash,
		StartVoteNonce: startVoteNonce,
		EndVoteNonce:   endVoteNonce,
		Yes:            big.NewInt(0),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Abstain:        big.NewInt(0),
		Passed:         false,
		ProposalCost:   generalConfig.ProposalFee,
	}
	err = g.saveGeneralProposal(commitHash, generalProposal)
	if err != nil {
		g.eei.AddReturnMessage("saveGeneralProposal " + err.Error())
		return vmcommon.UserError
	}

	g.eei.SetStorage(big.NewInt(0).SetUint64(nextNonce).Bytes(), commitHash)

	return vmcommon.Ok
}

// vote casts a vote for a validator/delegation. This function receives 2 parameters and will vote with its full delegation + validator amount
//  args.Arguments[0] - reference - nonce as string
//  args.Arguments[1] - vote option (yes, no, veto, abstain)
func (g *governanceContract) vote(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		g.eei.AddReturnMessage("function is not payable")
		return vmcommon.UserError
	}
	err := g.eei.UseGas(g.gasCost.MetaChainSystemSCsCost.Vote)
	if err != nil {
		g.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 2 {
		g.eei.AddReturnMessage("invalid number of arguments, expected 2")
		return vmcommon.FunctionWrongSignature
	}
	if core.IsSmartContractAddress(args.CallerAddr) {
		g.eei.AddReturnMessage("only SC can call this")
		return vmcommon.UserError
	}

	voterAddress := args.CallerAddr
	proposalToVote := args.Arguments[0]
	totalVotingPower, err := g.computeVotingPowerFromTotalStake(voterAddress)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = g.addUserVote(
		voterAddress,
		proposalToVote,
		string(args.Arguments[1]),
		totalVotingPower,
		proposalToVote,
		true)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

// delegateVote casts a vote from a validator run by WASM SC and delegates it to someone else. This function receives 4 parameters:
//  args.Arguments[0] - proposal reference - nonce of proposal
//  args.Arguments[1] - vote option (yes, no, veto)
//  args.Arguments[2] - delegatedTo
//  args.Arguments[3] - balance to vote
func (g *governanceContract) delegateVote(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) != 4 {
		g.eei.AddReturnMessage("invalid number of arguments")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		g.eei.AddReturnMessage("function is not payable")
		return vmcommon.UserError
	}
	err := g.eei.UseGas(g.gasCost.MetaChainSystemSCsCost.DelegateVote)
	if err != nil {
		g.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}
	if !core.IsSmartContractAddress(args.CallerAddr) {
		g.eei.AddReturnMessage("only SC can call this")
		return vmcommon.UserError
	}
	voter := args.Arguments[3]
	if len(voter) != len(args.CallerAddr) {
		g.eei.AddReturnMessage("invalid delegator address")
		return vmcommon.UserError
	}

	proposalToVote := args.Arguments[0]
	votePower, err := g.updateDelegatedContractInfo(args.CallerAddr, proposalToVote, args.Arguments[2])
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = g.addUserVote(
		voter,
		proposalToVote,
		string(args.Arguments[1]),
		votePower,
		proposalToVote,
		false)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (g *governanceContract) updateDelegatedContractInfo(scAddress []byte, reference []byte, balance []byte) (*big.Int, error) {
	scVoteInfo, err := g.getDelegatedContractInfo(scAddress, reference)
	if err != nil {
		return nil, err
	}
	votePower, err := g.computeVotingPower(big.NewInt(0).SetBytes(balance))
	if err != nil {
		return nil, err
	}

	scVoteInfo.UsedPower.Add(scVoteInfo.UsedPower, votePower)
	if scVoteInfo.TotalPower.Cmp(scVoteInfo.UsedPower) < 0 {
		return nil, fmt.Errorf("not enough voting power to cast this vote")
	}
	err = g.saveDelegatedContractInfo(scAddress, scVoteInfo, reference)
	if err != nil {
		return nil, err
	}

	return votePower, nil
}

func (g *governanceContract) addUserVote(
	address []byte,
	nonceAsBytes []byte,
	vote string,
	totalVotingPower *big.Int,
	proposalToVote []byte,
	direct bool,
) error {
	nonce, err := nonceFromBytes(nonceAsBytes)
	if err != nil {
		return err
	}

	err = g.updateUserVoteList(address, nonce.Uint64(), direct)
	if err != nil {
		return err
	}

	voteOption, err := g.castVoteType(vote)
	if err != nil {
		return err
	}

	proposal, err := g.getValidProposal(proposalToVote)
	if err != nil {
		return err
	}

	err = g.addNewVote(voteOption, totalVotingPower, proposal)
	if err != nil {
		return err
	}

	g.lockStake(address, proposal.EndVoteNonce)
	return g.saveGeneralProposal(proposalToVote, proposal)
}

func (g *governanceContract) updateUserVoteList(address []byte, nonce uint64, direct bool) error {
	userVoteList, err := g.getUserVotes(address)
	if err != nil {
		return err
	}

	if direct {
		userVoteList.Direct, err = addNewNonce(userVoteList.Direct, nonce)
		if err != nil {
			return err
		}
	} else {
		userVoteList.Delegated, err = addNewNonce(userVoteList.Delegated, nonce)
		if err != nil {
			return err
		}
	}

	return g.saveUserVotes(address, userVoteList)
}

func addNewNonce(nonceList []uint64, newNonce uint64) ([]uint64, error) {
	for _, nonce := range nonceList {
		if newNonce == nonce {
			return nil, vm.ErrDoubleVote
		}
	}

	return nonceList, nil
}

//TODO: I would delete lockStake - if we put a voting period less than 10 epochs, we do not need this.
func (g *governanceContract) lockStake(address []byte, endNonce uint64) {
	stakeLockKey := append([]byte(stakeLockPrefix), address...)
	lastData := g.eei.GetStorage(stakeLockKey)
	lastEndNonce := uint64(0)
	if len(lastData) > 0 {
		lastEndNonce = big.NewInt(0).SetBytes(lastData).Uint64()
	}

	if lastEndNonce < endNonce {
		g.eei.SetStorage(stakeLockKey, big.NewInt(0).SetUint64(endNonce).Bytes())
	}
}

func isStakeLocked(eei vm.SystemEI, governanceAddress []byte, address []byte) bool {
	stakeLockKey := append([]byte(stakeLockPrefix), address...)
	lastData := eei.GetStorageFromAddress(governanceAddress, stakeLockKey)
	if len(lastData) == 0 {
		return false
	}

	lastEndNonce := big.NewInt(0).SetBytes(lastData).Uint64()
	return eei.BlockChainHook().CurrentNonce() < lastEndNonce
}

func (g *governanceContract) getMinValueToVote() (*big.Int, error) {
	delegationManagement, err := getDelegationManagement(g.eei, g.marshalizer, g.delegationMgrSCAddress)
	if err != nil {
		return nil, err
	}

	return delegationManagement.MinDelegationAmount, nil
}

// closeProposal generates and saves end results for a proposal
func (g *governanceContract) closeProposal(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		g.eei.AddReturnMessage("closeProposal callValue expected to be 0")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		g.eei.AddReturnMessage("invalid number of arguments expected 1")
		return vmcommon.UserError
	}
	err := g.eei.UseGas(g.gasCost.MetaChainSystemSCsCost.CloseProposal)
	if err != nil {
		g.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}

	proposal := args.Arguments[0]
	generalProposal, err := g.getGeneralProposal(proposal)
	if err != nil {
		g.eei.AddReturnMessage("getGeneralProposal error " + err.Error())
		return vmcommon.UserError
	}
	if generalProposal.Closed {
		g.eei.AddReturnMessage("proposal is already closed, do nothing")
		return vmcommon.Ok
	}

	currentNonce := g.eei.BlockChainHook().CurrentNonce()
	if currentNonce < generalProposal.EndVoteNonce {
		g.eei.AddReturnMessage(fmt.Sprintf("proposal can be closed only after nonce %d", generalProposal.EndVoteNonce))
		return vmcommon.UserError
	}

	generalProposal.Closed = true
	err = g.computeEndResults(generalProposal)
	if err != nil {
		g.eei.AddReturnMessage("computeEndResults error" + err.Error())
		return vmcommon.UserError
	}

	err = g.saveGeneralProposal(proposal, generalProposal)
	if err != nil {
		g.eei.AddReturnMessage("saveGeneralProposal error" + err.Error())
		return vmcommon.UserError
	}

	err = g.eei.Transfer(args.RecipientAddr, args.CallerAddr, generalProposal.ProposalCost, nil, 0)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

// getValidatorVotingPower returns the total voting power for a validator. Un-staked nodes are not
//  taken into consideration
func (g *governanceContract) getValidatorVotingPower(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		g.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	err := g.eei.UseGas(g.gasCost.MetaChainSystemSCsCost.Vote)
	if err != nil {
		g.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 1 {
		g.eei.AddReturnMessage("function accepts only one argument, the validator address")
		return vmcommon.FunctionWrongSignature
	}
	validatorAddress := args.Arguments[0]
	if len(validatorAddress) != len(args.CallerAddr) {
		g.eei.AddReturnMessage("invalid argument - validator address")
		return vmcommon.UserError
	}

	votingPower, err := g.computeValidatorVotingPower(validatorAddress)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.ExecutionFailed
	}

	g.eei.Finish(votingPower.Bytes())

	return vmcommon.Ok
}

// getValidProposal returns a proposal from storage if it exists, or it is still valid/in-progress
func (g *governanceContract) getValidProposal(reference []byte) (*GeneralProposal, error) {
	proposal, err := g.getGeneralProposal(reference)
	if err != nil {
		return nil, err
	}

	currentNonce := g.eei.BlockChainHook().CurrentNonce()
	if currentNonce < proposal.StartVoteNonce {
		return nil, vm.ErrVotingNotStartedForProposal
	}

	if currentNonce > proposal.EndVoteNonce {
		return nil, vm.ErrVotedForAnExpiredProposal
	}

	return proposal, nil
}

// addNewVote applies a new vote on a proposal then saves the new information into the storage
func (g *governanceContract) addNewVote(voteValueType VoteValueType, power *big.Int, proposal *GeneralProposal) error {
	switch voteValueType {
	case Yes:
		proposal.Yes.Add(proposal.Yes, power)
	case No:
		proposal.No.Add(proposal.No, power)
	case Veto:
		proposal.Veto.Add(proposal.Veto, power)
	case Abstain:
		proposal.Abstain.Add(proposal.Abstain, power)
	default:
		return fmt.Errorf("%s: %s", vm.ErrInvalidArgument, "invalid vote type")
	}

	return g.saveGeneralProposal(proposal.CommitHash, proposal)
}

// computeVotingPower returns the voting power for a value. The value can be either a balance or
//  the staked value for a validator
func (g *governanceContract) computeVotingPower(value *big.Int) (*big.Int, error) {
	if value.Cmp(zero) < 0 {
		return nil, fmt.Errorf("cannot compute voting power on a negative value")
	}

	return big.NewInt(0).Sqrt(value), nil
}

// castVoteType casts a valid string vote passed as an argument to the actual mapped value
func (g *governanceContract) castVoteType(vote string) (VoteValueType, error) {
	switch vote {
	case yesString:
		return Yes, nil
	case noString:
		return No, nil
	case vetoString:
		return Veto, nil
	case abstainString:
		return Abstain, nil
	default:
		return 0, fmt.Errorf("%s: %s%s", vm.ErrInvalidArgument, "invalid vote type option: ", vote)
	}
}

// computeValidatorVotingPower returns the total voting power of a validator
func (g *governanceContract) computeValidatorVotingPower(validatorAddress []byte) (*big.Int, error) {
	totalStake, err := g.getTotalStake(validatorAddress)
	if err != nil {
		return nil, fmt.Errorf("could not return total stake for the provided address, thus cannot compute voting power")
	}

	votingPower, err := g.computeVotingPower(totalStake)
	if err != nil {
		return nil, fmt.Errorf("could not return total stake for the provided address, thus cannot compute voting power")
	}

	return votingPower, nil
}

// function iterates over all delegation contracts and verifies balances of the given account and makes a sum of it
func (g *governanceContract) computeVotingPowerFromTotalStake(address []byte) (*big.Int, error) {
	totalStake, err := g.getTotalStake(address)
	if err != nil && err != vm.ErrEmptyStorage {
		return nil, fmt.Errorf("could not return total stake for the provided address, thus cannot compute voting power")
	}
	if totalStake == nil {
		totalStake = big.NewInt(0)
	}

	dContractList, err := getDelegationContractList(g.eei, g.marshalizer, g.delegationMgrSCAddress)
	if err != nil {
		return nil, err
	}

	err = g.eei.UseGas(g.gasCost.MetaChainSystemSCsCost.Get * uint64(len(dContractList.Addresses)))
	if err != nil {
		return nil, err
	}

	var activeDelegated *big.Int
	for _, contract := range dContractList.Addresses {
		activeDelegated, err = g.getActiveFundForDelegator(contract, address)
		if err != nil {
			return nil, err
		}

		totalStake.Add(totalStake, activeDelegated)
	}

	votingPower, err := g.computeVotingPower(totalStake)
	if err != nil {
		return nil, err
	}

	return votingPower, nil
}

// computeEndResults computes if a proposal has passed or not based on votes accumulated
func (g *governanceContract) computeEndResults(proposal *GeneralProposal) error {
	baseConfig, err := g.getConfig()
	if err != nil {
		return err
	}

	totalVotes := big.NewInt(0).Add(proposal.Yes, proposal.No)
	totalVotes.Add(totalVotes, proposal.Veto)

	if totalVotes.Cmp(baseConfig.MinQuorum) == -1 {
		proposal.Passed = false
		return nil
	}

	if proposal.Veto.Cmp(baseConfig.MinVetoThreshold) >= 0 {
		proposal.Passed = false
		return nil
	}

	if proposal.Yes.Cmp(baseConfig.MinPassThreshold) >= 0 && proposal.Yes.Cmp(proposal.No) == 1 {
		proposal.Passed = true
		return nil
	}

	proposal.Passed = false
	return nil
}

func (g *governanceContract) getActiveFundForDelegator(delegationAddress []byte, address []byte) (*big.Int, error) {
	dData := &DelegatorData{
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	}
	marshaledData := g.eei.GetStorageFromAddress(delegationAddress, address)
	if len(marshaledData) == 0 {
		return big.NewInt(0), nil
	}

	err := g.marshalizer.Unmarshal(dData, marshaledData)
	if err != nil {
		return nil, err
	}

	if len(dData.ActiveFund) == 0 {
		return big.NewInt(0), nil
	}

	marshaledData = g.eei.GetStorageFromAddress(delegationAddress, dData.ActiveFund)
	activeFund := &Fund{Value: big.NewInt(0)}
	err = g.marshalizer.Unmarshal(activeFund, marshaledData)
	if err != nil {
		return nil, err
	}

	return activeFund.Value, nil
}

func (g *governanceContract) getTotalStake(validatorAddress []byte) (*big.Int, error) {
	marshaledData := g.eei.GetStorageFromAddress(g.validatorSCAddress, validatorAddress)
	if len(marshaledData) == 0 {
		return nil, vm.ErrEmptyStorage
	}

	validatorData := &ValidatorDataV2{}
	err := g.marshalizer.Unmarshal(validatorData, marshaledData)
	if err != nil {
		return nil, err
	}

	return validatorData.TotalStakeValue, nil
}

func (g *governanceContract) saveUserVotes(address []byte, votedList *OngoingVotedList) error {
	marshaledData, err := g.marshalizer.Marshal(votedList)
	if err != nil {
		return err
	}
	g.eei.SetStorage(address, marshaledData)

	return nil
}

func (g *governanceContract) getUserVotes(address []byte) (*OngoingVotedList, error) {
	onGoingList := &OngoingVotedList{
		Direct:    make([]uint64, 0),
		Delegated: make([]uint64, 0),
	}
	marshaledData := g.eei.GetStorage(address)
	if len(marshaledData) == 0 {
		return onGoingList, nil
	}

	err := g.marshalizer.Unmarshal(onGoingList, marshaledData)
	if err != nil {
		return nil, err
	}

	return onGoingList, nil
}

func (g *governanceContract) getDelegatedContractInfo(scAddress []byte, reference []byte) (*DelegatedSCVoteInfo, error) {
	scVoteInfo := &DelegatedSCVoteInfo{
		TotalPower: big.NewInt(0),
		UsedPower:  big.NewInt(0),
	}

	marshalledData := g.eei.GetStorage(append(scAddress, reference...))
	if len(marshalledData) > 0 {
		err := g.marshalizer.Unmarshal(scVoteInfo, marshalledData)
		if err != nil {
			return nil, err
		}

		return scVoteInfo, nil
	}

	totalVotingPower, err := g.computeVotingPowerFromTotalStake(scAddress)
	if err != nil {
		return nil, err
	}
	scVoteInfo.TotalPower.Set(totalVotingPower)

	return scVoteInfo, nil
}

func (g *governanceContract) saveDelegatedContractInfo(
	scAddress []byte,
	scVoteInfo *DelegatedSCVoteInfo,
	reference []byte,
) error {
	marshalledData, err := g.marshalizer.Marshal(scVoteInfo)
	if err != nil {
		return err
	}

	g.eei.SetStorage(append(scAddress, reference...), marshalledData)
	return nil
}

// getConfig returns the current system smart contract configuration
func (g *governanceContract) getConfig() (*GovernanceConfigV2, error) {
	scConfig := &GovernanceConfigV2{}

	marshaledData := g.eei.GetStorage([]byte(governanceConfigKey))
	if len(marshaledData) == 0 {
		return nil, vm.ErrElementNotFound
	}

	err := g.marshalizer.Unmarshal(scConfig, marshaledData)
	if err != nil {
		return nil, err
	}

	return scConfig, nil
}

// saveGeneralProposal saves a proposal into the storage
func (g *governanceContract) saveGeneralProposal(reference []byte, generalProposal *GeneralProposal) error {
	marshaledData, err := g.marshalizer.Marshal(generalProposal)
	if err != nil {
		return err
	}
	key := append([]byte(proposalPrefix), reference...)
	g.eei.SetStorage(key, marshaledData)

	return nil
}

// getGeneralProposal returns a proposal from storage
func (g *governanceContract) getGeneralProposal(reference []byte) (*GeneralProposal, error) {
	commitHash := g.eei.GetStorage(reference)
	key := append([]byte(proposalPrefix), commitHash...)
	marshaledData := g.eei.GetStorage(key)

	if len(marshaledData) == 0 {
		return nil, vm.ErrProposalNotFound
	}

	generalProposal := &GeneralProposal{}
	err := g.marshalizer.Unmarshal(generalProposal, marshaledData)
	if err != nil {
		return nil, err
	}

	return generalProposal, nil
}

// proposalExists returns true if a proposal already exists
func (g *governanceContract) proposalExists(reference []byte) bool {
	key := append([]byte(proposalPrefix), reference...)
	marshaledData := g.eei.GetStorage(key)
	return len(marshaledData) > 0
}

// startEndNonceFromArguments converts the nonce string arguments to uint64
func (g *governanceContract) startEndNonceFromArguments(argStart []byte, argEnd []byte) (uint64, uint64, error) {
	startVoteNonce, err := nonceFromBytes(argStart)
	if err != nil {
		return 0, 0, err
	}
	endVoteNonce, err := nonceFromBytes(argEnd)
	if err != nil {
		return 0, 0, err
	}

	currentNonce := g.eei.BlockChainHook().CurrentNonce()
	if currentNonce > startVoteNonce.Uint64() || startVoteNonce.Uint64() > endVoteNonce.Uint64() {
		return 0, 0, vm.ErrInvalidStartEndVoteNonce
	}

	return startVoteNonce.Uint64(), endVoteNonce.Uint64(), nil
}

// nonceFromBytes converts a byte array to a big.Int. Returns ErrInvalidStartEndVoteNonce for invalid values
func nonceFromBytes(nonce []byte) (*big.Int, error) {
	voteNonce, okConvert := big.NewInt(0).SetString(string(nonce), conversionBase)
	if !okConvert {
		return nil, vm.ErrInvalidStartEndVoteNonce
	}
	if !voteNonce.IsUint64() {
		return nil, vm.ErrInvalidStartEndVoteNonce
	}

	return voteNonce, nil
}

// convertV2Config converts the passed config file to the correct V2 typed GovernanceConfig
func (g *governanceContract) convertV2Config(config config.GovernanceSystemSCConfig) (*GovernanceConfigV2, error) {
	minQuorum, success := big.NewInt(0).SetString(config.Active.MinQuorum, conversionBase)
	if !success {
		return nil, vm.ErrIncorrectConfig
	}
	minPass, success := big.NewInt(0).SetString(config.Active.MinPassThreshold, conversionBase)
	if !success {
		return nil, vm.ErrIncorrectConfig
	}
	minVeto, success := big.NewInt(0).SetString(config.Active.MinVetoThreshold, conversionBase)
	if !success {
		return nil, vm.ErrIncorrectConfig
	}
	proposalFee, success := big.NewInt(0).SetString(config.Active.ProposalCost, conversionBase)
	if !success {
		return nil, vm.ErrIncorrectConfig
	}

	return &GovernanceConfigV2{
		MinQuorum:        minQuorum,
		MinPassThreshold: minPass,
		MinVetoThreshold: minVeto,
		ProposalFee:      proposalFee,
	}, nil
}

// CanUseContract returns true if contract is enabled
func (g *governanceContract) CanUseContract() bool {
	return true
}

// SetNewGasCost is called whenever a gas cost was changed
func (g *governanceContract) SetNewGasCost(gasCost vm.GasCost) {
	g.mutExecution.Lock()
	g.gasCost = gasCost
	g.mutExecution.Unlock()
}

// IsInterfaceNil returns true if underlying object is nil
func (g *governanceContract) IsInterfaceNil() bool {
	return g == nil
}
