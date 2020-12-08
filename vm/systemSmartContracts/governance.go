//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. governance.proto
package systemSmartContracts

import (
	"bytes"
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/vm"
)

const governanceConfigKey = "governanceConfig"
const hardForkPrefix = "hardFork_"
const proposalPrefix = "proposal_"
const accountLockPrefix = "accountLock_"
const validatorLockPrefix = "validatorLock_"
const whiteListPrefix = "whiteList_"
const yesString = "yes"
const noString = "no"
const vetoString = "veto"
const hardForkEpochGracePeriod = 2
const githubCommitLength = 40

// ArgsNewGovernanceContract defines the arguments needed for the on-chain governance contract
type ArgsNewGovernanceContract struct {
	Eei                        vm.SystemEI
	GasCost                    vm.GasCost
	GovernanceConfig           config.GovernanceSystemSCConfig
	Marshalizer                marshal.Marshalizer
	Hasher                     hashing.Hasher
	GovernanceSCAddress        []byte
	StakingSCAddress           []byte
	ValidatorSCAddress  []byte
	InitalWhiteListedAddresses [][]byte
	EpochNotifier              vm.EpochNotifier
}

type governanceContract struct {
	eei                         vm.SystemEI
	gasCost                     vm.GasCost
	baseProposalCost            *big.Int
	ownerAddress                []byte
	governanceSCAddress         []byte
	stakingSCAddress            []byte
	validatorSCAddress  []byte
	marshalizer                 marshal.Marshalizer
	hasher                      hashing.Hasher
	governanceConfig            config.GovernanceSystemSCConfig
	initialWhiteListedAddresses [][]byte
	enabledEpoch                uint32
	flagEnabled                 atomic.Flag
	mutExecution        sync.RWMutex
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
	if check.IfNil(args.EpochNotifier) {
		return nil, vm.ErrNilEpochNotifier
	}

	activeConfig := args.GovernanceConfig.Active
	baseProposalCost, okConvert := big.NewInt(0).SetString(activeConfig.ProposalCost, conversionBase)
	if !okConvert || baseProposalCost.Cmp(zero) < 0 {
		return nil, vm.ErrInvalidBaseIssuingCost
	}

	g := &governanceContract{
		eei:                 args.Eei,
		gasCost:             args.GasCost,
		baseProposalCost:    baseProposalCost,
		ownerAddress:        nil,
		governanceSCAddress: args.GovernanceSCAddress,
		stakingSCAddress:    args.StakingSCAddress,
		validatorSCAddress:  args.ValidatorSCAddress,
		marshalizer:         args.Marshalizer,
		hasher:              args.Hasher,
		governanceConfig:    args.GovernanceConfig,
		enabledEpoch:        activeConfig.EnabledEpoch,
	}

	err := g.validateInitialWhiteListedAddresses(args.InitalWhiteListedAddresses)
	if err != nil {
		return nil, err
	}
	g.initialWhiteListedAddresses = args.InitalWhiteListedAddresses

	args.EpochNotifier.RegisterNotifyHandler(g)

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

	if !g.flagEnabled.IsSet() {
		g.eei.AddReturnMessage("Governance SC disabled")
		return vmcommon.UserError
	}

	switch args.Function {
	case "initV2":
		return g.initV2(args)
	case "proposal":
		return g.proposal(args)
	case "vote":
		return g.vote(args)
	case "claimFunds":
		return g.claimFunds(args)
	case "whiteList":
		return g.whiteListProposal(args)
	case "hardFork":
		return g.hardForkProposal(args)
	case "changeConfig":
		return g.changeConfig(args)
	case "closeProposal":
		return g.closeProposal(args)
	case "getValidatorVotingPower":
		return g.getValidatorVotingPower(args)
	case "getBalanceVotingPower":
		return g.getBalanceVotingPower(args)
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
	log.LogIfError(err, "marshal error on governance init function")

	g.eei.SetStorage([]byte(governanceConfigKey), marshaledData)
	g.eei.SetStorage([]byte(ownerKey), args.CallerAddr)
	g.ownerAddress = make([]byte, 0, len(args.CallerAddr))
	g.ownerAddress = append(g.ownerAddress, args.CallerAddr...)
	return vmcommon.Ok
}

func (g *governanceContract) initV2(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, vm.GovernanceSCAddress) {
		log.Error("invalid caller to switch to V2 config")
		return vmcommon.UserError
	}
	cfg, err := g.convertV2Config(g.governanceConfig)
	if err != nil {
		log.Error("could not create governance V2 config")
		return vmcommon.UserError
	}

	marshaledData, err := g.marshalizer.Marshal(cfg)
	if err != nil {
		log.Error("marshal error on governance init function")
		return vmcommon.ExecutionFailed
	}

	g.eei.SetStorage([]byte(governanceConfigKey), marshaledData)
	g.eei.SetStorage([]byte(ownerKey), args.CallerAddr)
	g.ownerAddress = make([]byte, 0, len(args.CallerAddr))
	g.ownerAddress = append(g.ownerAddress, args.CallerAddr...)

	return vmcommon.Ok
}

// proposal creates a new proposal from passed arguments
func (g *governanceContract) proposal(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(g.baseProposalCost) != 0 {
		g.eei.AddReturnMessage("invalid proposal cost, expected " + g.baseProposalCost.String())
		return vmcommon.OutOfFunds
	}
	err := g.eei.UseGas(g.gasCost.MetaChainSystemSCsCost.Proposal)
	if err != nil {
		g.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 3 {
		g.eei.AddReturnMessage("invalid number of arguments, expected 3")
		return vmcommon.FunctionWrongSignature
	}
	if !g.isWhiteListed(args.CallerAddr) {
		g.eei.AddReturnMessage("called address is not whiteListed")
		return vmcommon.UserError
	}
	gitHubCommit := args.Arguments[0]
	if len(gitHubCommit) != githubCommitLength {
		g.eei.AddReturnMessage(fmt.Sprintf("invalid github commit length, wanted exactly %d", githubCommitLength))
		return vmcommon.UserError
	}
	if g.proposalExists(gitHubCommit) {
		g.eei.AddReturnMessage("proposal already exists")
		return vmcommon.UserError
	}

	startVoteNonce, endVoteNonce, err := g.startEndNonceFromArguments(args.Arguments[1], args.Arguments[2])
	if err != nil {
		g.eei.AddReturnMessage("invalid start/end vote nonce" + err.Error())
		return vmcommon.UserError
	}

	generalProposal := &GeneralProposal{
		IssuerAddress:  args.CallerAddr,
		GitHubCommit:   gitHubCommit,
		StartVoteNonce: startVoteNonce,
		EndVoteNonce:   endVoteNonce,
		Yes:            big.NewInt(0),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Voted:          false,
		Votes:          make([][]byte, 0),
	}
	err = g.saveGeneralProposal(gitHubCommit, generalProposal)
	if err != nil {
		log.Warn("saveGeneralProposal", "err", err)
		g.eei.AddReturnMessage("saveGeneralProposal" + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

// vote will cast a new vote
func (g *governanceContract) vote(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		return g.accountVote(args)
	}
	return g.validatorVote(args)
}

// accountVote casts a vote taking the transaction value as input for the vote power. It receives 2 arguments:
//  args.Arguments[0] - proposal reference (github commit)
//  args.Arguments[1] - vote option (yes, no, veto)
func (g *governanceContract) accountVote(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := g.eei.UseGas(g.gasCost.MetaChainSystemSCsCost.Vote)
	if err != nil {
		g.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}

	if len(args.Arguments) != 2 {
		g.eei.AddReturnMessage("invalid number of arguments, expected 2")
		return vmcommon.FunctionWrongSignature
	}

	voterAddress := args.CallerAddr
	proposalToVote := args.Arguments[0]
	proposal, err := g.getValidProposal(proposalToVote)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	voteOption, err := g.castVoteType(string(args.Arguments[1]))
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	currentVoteSet, err := g.getOrCreateVoteSet(proposalToVote, voterAddress)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.ExecutionFailed
	}

	votePower, err := g.computeAccountLeveledPower(args.CallValue, currentVoteSet)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	currentVote := &VoteDetails{
		Value:       voteOption,
		Power:       votePower,
		Balance:     args.CallValue,
		Type:        Account,
	}

	err = g.addNewVote(voterAddress, currentVote, currentVoteSet, proposal)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

// validatorVote casts a vote for a validator. This function can receive 3 or 4 parameters:
//  args.Arguments[0] - proposal reference (github commit)
//  args.Arguments[1] - vote option (yes, no, veto)
//  args.Arguments[2] - vote power used for this vote
//  args.Arguments[3] (optional) - an address that identifies if the vote was made on behalf of someone else - this
//   only helps for statistical and view purposes, it does not afftect the logic in any way
func (g *governanceContract) validatorVote(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := g.eei.UseGas(g.gasCost.MetaChainSystemSCsCost.Vote)
	if err != nil {
		g.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}

	if len(args.Arguments) < 3 || len(args.Arguments) > 4 {
		g.eei.AddReturnMessage("invalid number of arguments, expected 3 or 4")
		return vmcommon.FunctionWrongSignature
	}

	voterAddress := args.CallerAddr
	proposalToVote := args.Arguments[0]
	proposal, err := g.getValidProposal(proposalToVote)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	voteOption, err := g.castVoteType(string(args.Arguments[1]))
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	votePower := big.NewInt(0).SetBytes(args.Arguments[2])
	delegatedTo, err := g.getDelegatedToAddress(args)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	currentVote := &VoteDetails{
		Value:       voteOption,
		Power:       votePower,
		DelegatedTo: delegatedTo,
		Balance:     big.NewInt(0),
		Type:        Validator,
	}

	totalVotingPower, err := g.computeValidatorVotingPower(voterAddress)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	currentVoteSet, err := g.getOrCreateVoteSet(proposalToVote, voterAddress)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.ExecutionFailed
	}
	if totalVotingPower.Cmp(big.NewInt(0).Add(votePower, currentVoteSet.UsedPower)) == -1 {
		g.eei.AddReturnMessage("not enough voting power to cast this vote")
		return vmcommon.UserError
	}

	err = g.addNewVote(voterAddress, currentVote, currentVoteSet, proposal)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}


	return vmcommon.Ok
}

// claimFunds returns back the used funds for a particular proposal if they are unlocked. Accepts a single parameter:
//  args.Arguments[0] - proposal reference
func (g *governanceContract) claimFunds(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(big.NewInt(0)) != 0 {
		g.eei.AddReturnMessage("invalid callValue, should be 0")
		return vmcommon.UserError
	}

	if len(args.Arguments) != 1 {
		g.eei.AddReturnMessage("invalid number of arguments, expected 1")
		return vmcommon.FunctionWrongSignature
	}

	lock := g.getLock(args.CallerAddr, Account, args.Arguments[0])
	currentNonce := g.eei.BlockChainHook().CurrentNonce()

	if lock < currentNonce {
		g.eei.AddReturnMessage("your funds are still locked")
		return vmcommon.UserError
	}

	currentVoteSet, err := g.getOrCreateVoteSet(args.Arguments[0], args.CallerAddr)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.ExecutionFailed
	}
	if currentVoteSet.Claimed == true {
		g.eei.AddReturnMessage("you already claimed back your funds")
		return vmcommon.UserError
	}

	if currentVoteSet.UsedBalance.Cmp(zero) != 1 {
		g.eei.AddReturnMessage("no funds to claim for this proposal")
		return vmcommon.UserError
	}

	currentVoteSet.Claimed = true
	err = g.storeVoteSet(args.CallerAddr, currentVoteSet, args.Arguments[0])
	if err != nil {
		g.eei.AddReturnMessage("could not save vote data as claimed")
		return vmcommon.ExecutionFailed
	}

	err = g.eei.Transfer(args.CallerAddr, g.governanceSCAddress, currentVoteSet.UsedBalance, nil, 0)
	if err != nil {
		g.eei.AddReturnMessage("transfer error on claimFunds function")
		return vmcommon.ExecutionFailed
	}

	return vmcommon.Ok
}

// whiteListProposal will create a new proposal to white list an address
func (g *governanceContract) whiteListProposal(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	currentNonce := g.eei.BlockChainHook().CurrentNonce()
	if currentNonce == 0 {
		return g.whiteListAtGenesis(args)
	}
	if args.CallValue.Cmp(g.baseProposalCost) != 0 {
		g.eei.AddReturnMessage("invalid callValue, needs exactly " + g.baseProposalCost.String())
		return vmcommon.OutOfFunds
	}
	err := g.eei.UseGas(g.gasCost.MetaChainSystemSCsCost.Proposal)
	if err != nil {
		g.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 3 {
		g.eei.AddReturnMessage("invalid number of arguments")
		return vmcommon.FunctionWrongSignature
	}
	if g.proposalExists(args.Arguments[0]) {
		g.eei.AddReturnMessage("cannot re-propose existing proposal")
		return vmcommon.UserError
	}
	if g.isWhiteListed(args.CallerAddr) {
		g.eei.AddReturnMessage("address is already whitelisted")
		return vmcommon.UserError
	}
	if len(args.Arguments[0]) != githubCommitLength {
		g.eei.AddReturnMessage(fmt.Sprintf("invalid github commit length, wanted exactly %d", githubCommitLength))
		return vmcommon.UserError
	}

	startVoteNonce, endVoteNonce, err := g.startEndNonceFromArguments(args.Arguments[1], args.Arguments[2])
	if err != nil {
		g.eei.AddReturnMessage("invalid start/end vote nonce " + err.Error())
		return vmcommon.UserError
	}

	key := append([]byte(proposalPrefix), args.CallerAddr...)
	whiteListAcc := &WhiteListProposal{
		WhiteListAddress: args.CallerAddr,
		ProposalStatus:   key,
	}

	key = append([]byte(whiteListPrefix), args.CallerAddr...)
	generalProposal := &GeneralProposal{
		IssuerAddress:  args.CallerAddr,
		GitHubCommit:   args.Arguments[0],
		StartVoteNonce: startVoteNonce,
		EndVoteNonce:   endVoteNonce,
		Yes:            big.NewInt(0),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Voted:          false,
		Votes:          make([][]byte, 0),
	}

	marshaledData, err := g.marshalizer.Marshal(whiteListAcc)
	if err != nil {
		g.eei.AddReturnMessage("marshall error " + err.Error())
		return vmcommon.UserError
	}
	g.eei.SetStorage(key, marshaledData)

	err = g.saveGeneralProposal(args.CallerAddr, generalProposal)
	if err != nil {
		g.eei.AddReturnMessage("save proposal error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

// hardForkProposal creates a new proposal for a hard-fork
func (g *governanceContract) hardForkProposal(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(g.baseProposalCost) != 0 {
		g.eei.AddReturnMessage("invalid proposal cost, expected " + g.baseProposalCost.String())
		return vmcommon.OutOfFunds
	}
	err := g.eei.UseGas(g.gasCost.MetaChainSystemSCsCost.Proposal)
	if err != nil {
		g.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 5 {
		g.eei.AddReturnMessage("invalid number of arguments, expected 5")
		return vmcommon.FunctionWrongSignature
	}
	if !g.isWhiteListed(args.CallerAddr) {
		g.eei.AddReturnMessage("called address is not whiteListed")
		return vmcommon.UserError
	}
	gitHubCommit := args.Arguments[2]
	if len(gitHubCommit) != githubCommitLength {
		g.eei.AddReturnMessage(fmt.Sprintf("invalid github commit length, wanted exactly %d", githubCommitLength))
		return vmcommon.UserError
	}
	if g.proposalExists(gitHubCommit) {
		g.eei.AddReturnMessage("proposal already exists")
		return vmcommon.UserError
	}

	key := append([]byte(hardForkPrefix), gitHubCommit...)
	marshaledData := g.eei.GetStorage(key)
	if len(marshaledData) != 0 {
		g.eei.AddReturnMessage("hardFork proposal already exists")
		return vmcommon.UserError
	}

	startVoteNonce, endVoteNonce, err := g.startEndNonceFromArguments(args.Arguments[3], args.Arguments[4])
	if err != nil {
		g.eei.AddReturnMessage("invalid start/end vote nonce" + err.Error())
		return vmcommon.UserError
	}

	bigIntEpochToHardFork, okConvert := big.NewInt(0).SetString(string(args.Arguments[0]), conversionBase)
	if !okConvert || !bigIntEpochToHardFork.IsUint64() {
		g.eei.AddReturnMessage("invalid argument for epoch")
		return vmcommon.UserError
	}

	epochToHardFork := uint32(bigIntEpochToHardFork.Uint64())
	currentEpoch := g.eei.BlockChainHook().CurrentEpoch()
	if epochToHardFork < currentEpoch && currentEpoch-epochToHardFork < hardForkEpochGracePeriod {
		g.eei.AddReturnMessage("invalid epoch to hardFork")
		return vmcommon.UserError
	}

	key = append([]byte(proposalPrefix), gitHubCommit...)
	hardForkProposal := &HardForkProposal{
		EpochToHardFork:    epochToHardFork,
		NewSoftwareVersion: args.Arguments[1],
		ProposalStatus:     key,
	}

	key = append([]byte(hardForkPrefix), gitHubCommit...)
	generalProposal := &GeneralProposal{
		IssuerAddress:  args.CallerAddr,
		GitHubCommit:   gitHubCommit,
		StartVoteNonce: startVoteNonce,
		EndVoteNonce:   endVoteNonce,
		Yes:            big.NewInt(0),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Voted:          false,
		Votes:         make([][]byte, 0),
	}
	marshaledData, err = g.marshalizer.Marshal(hardForkProposal)
	if err != nil {
		log.Warn("hardFork proposal marshal", "err", err)
		g.eei.AddReturnMessage("marshal proposal" + err.Error())
		return vmcommon.UserError
	}
	g.eei.SetStorage(key, marshaledData)

	err = g.saveGeneralProposal(args.Arguments[0], generalProposal)
	if err != nil {
		log.Warn("save general proposal", "error", err)
		g.eei.AddReturnMessage("saveGeneralProposal" + err.Error())
		return vmcommon.UserError
	}

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

	numNodes, okConvert := big.NewInt(0).SetString(string(args.Arguments[0]), conversionBase)
	if !okConvert || numNodes.Cmp(zero) < 0 {
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

	marshaledData, err := g.marshalizer.Marshal(scConfig)
	if err != nil {
		g.eei.AddReturnMessage("changeConfig error " + err.Error())
		return vmcommon.UserError
	}
	g.eei.SetStorage([]byte(governanceConfigKey), marshaledData)

	return vmcommon.Ok
}

// closeProposal generates and saves end results for a proposal
func (g *governanceContract) closeProposal(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		g.eei.AddReturnMessage("closeProposal callValue expected to be 0")
		return vmcommon.UserError
	}
	if !g.isWhiteListed(args.CallerAddr) {
		g.eei.AddReturnMessage("caller is not whitelisted")
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

	return vmcommon.Ok
}

// getConfig returns the curent system smart contract configuration
func (g *governanceContract) getConfig() (*GovernanceConfigV2, error) {
	marshaledData := g.eei.GetStorage([]byte(governanceConfigKey))
	scConfig := &GovernanceConfigV2{}
	err := g.marshalizer.Unmarshal(scConfig, marshaledData)
	if err != nil {
		return nil, err
	}

	return scConfig, nil
}

// getValidatorVotingPower returns the total voting power for a validator. Un-staked nodes are not
//  taken into consideration
func (g *governanceContract) getValidatorVotingPower(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		g.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
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

// getBalanceVotingPower returns the voting power associated with the value sent in the transaction by the user
func (g *governanceContract) getBalanceVotingPower(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		g.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		g.eei.AddReturnMessage("function accepts only one argument, the balance for computing the power")
		return vmcommon.FunctionWrongSignature
	}

	balance := big.NewInt(0).SetBytes(args.Arguments[0])
	votingPower, err := g.computeVotingPower(balance)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	g.eei.Finish(votingPower.Bytes())
	return vmcommon.Ok
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
	key := append([]byte(proposalPrefix), reference...)
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

// getValidProposal returns a proposal from storage if it exists or it is still valid/in-progress
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

// isWhiteListed checks if an address is whitelisted
func (g *governanceContract) isWhiteListed(address []byte) bool {
	key := append([]byte(whiteListPrefix), address...)
	marshaledData := g.eei.GetStorage(key)
	if len(marshaledData) == 0 {
		return false
	}

	key = append([]byte(proposalPrefix), address...)
	marshaledData = g.eei.GetStorage(key)
	generalProposal := &GeneralProposal{}
	err := g.marshalizer.Unmarshal(generalProposal, marshaledData)
	if err != nil {
		return false
	}

	return generalProposal.Voted
}

func (g *governanceContract) whiteListAtGenesis(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		log.Warn("whiteList at genesis should be without callValue")
		return vmcommon.UserError
	}
	if g.isWhiteListed(args.CallerAddr) {
		log.Warn("address is already whiteListed")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 0 {
		log.Warn("excepted argument number is 0")
		return vmcommon.UserError
	}
	if g.proposalExists(args.CallerAddr) {
		log.Warn("proposal with this key already exists")
		return vmcommon.UserError
	}

	key := append([]byte(proposalPrefix), args.CallerAddr...)
	whiteListAcc := &WhiteListProposal{
		WhiteListAddress: args.CallerAddr,
		ProposalStatus:   key,
	}

	minQuorum, success := big.NewInt(0).SetString(g.governanceConfig.Active.MinQuorum, conversionBase)
	if !success {
		log.Warn("could not convert min quorum to bigInt")
		return vmcommon.UserError
	}

	key = append([]byte(whiteListPrefix), args.CallerAddr...)
	generalProposal := &GeneralProposal{
		IssuerAddress:  args.CallerAddr,
		GitHubCommit:   []byte("genesis"),
		StartVoteNonce: 0,
		EndVoteNonce:   0,
		Yes:            minQuorum,
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Voted:          true,
		Votes:          make([][]byte, 0),
	}
	marshaledData, err := g.marshalizer.Marshal(whiteListAcc)
	if err != nil {
		log.Warn("marshal error in whiteListAtGenesis", "err", err)
		return vmcommon.UserError
	}
	g.eei.SetStorage(key, marshaledData)

	err = g.saveGeneralProposal(args.CallerAddr, generalProposal)
	if err != nil {
		log.Warn("save general proposal ", "err", err)
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

// applyVote takes in a vote and a full VoteSet object and correctly applies the new vote, then returning
//  the new full VoteSet object. In the same way applies the vote to the general proposal
func (g *governanceContract) applyVote(vote *VoteDetails, voteData *VoteSet, proposal *GeneralProposal) (*VoteSet, *GeneralProposal, error) {
	switch vote.Value {
	case Yes:
		voteData.TotalYes.Add(voteData.TotalYes, vote.Power)
		proposal.Yes.Add(proposal.Yes, vote.Power)
		break
	case No:
		voteData.TotalNo.Add(voteData.TotalNo, vote.Power)
		proposal.No.Add(proposal.No, vote.Power)
		break
	case Veto:
		voteData.TotalVeto.Add(voteData.TotalVeto, vote.Power)
		proposal.Veto.Add(proposal.Veto, vote.Power)
		break
	default:
		return nil, nil, fmt.Errorf("%s: %s", vm.ErrInvalidArgument, "invalid vote type")
	}

	voteData.UsedPower.Add(voteData.UsedPower, vote.Power)
	voteData.UsedBalance.Add(voteData.UsedBalance, vote.Balance)
	voteData.VoteItems = append(voteData.VoteItems, vote)

	return voteData, proposal, nil
}

// addNewVote applies a new vote on a proposal then saves the new infonrmation into the storage
func (g *governanceContract) addNewVote(voterAddress []byte, currentVote *VoteDetails, currentVoteSet *VoteSet, proposal *GeneralProposal) error {
	newVoteSet, updatedProposal, err := g.applyVote(currentVote, currentVoteSet, proposal)
	if err != nil {
		return err
	}

	err = g.saveVoteSet(voterAddress, newVoteSet, updatedProposal, currentVote.Type)
	if err != nil {
		return err
	}

	return nil
}

// saveVoteSet first saves the main vote data of the voter, then updates the proposal with the new voteer information
func (g *governanceContract) saveVoteSet(voter []byte, voteData *VoteSet, proposal *GeneralProposal, voteType VoteType) error {
	proposalKey := append([]byte(proposalPrefix), proposal.GitHubCommit...)
	voteItemKey := append(proposalKey, voter...)

	marshaledVoteItem, err := g.marshalizer.Marshal(voteData)
	if err != nil {
		return err
	}
	g.eei.SetStorage(voteItemKey, marshaledVoteItem)

	if !g.proposalContainsVoter(proposal, voteItemKey) {
		proposal.Votes = append(proposal.Votes, voteItemKey)
	}

	marshaledProposal, err := g.marshalizer.Marshal(proposal)
	g.eei.SetStorage(proposalKey, marshaledProposal)

	g.setLock(voter, voteType, proposal)

	return nil
}

// storeVoteSet saves the provided vote data into the storage
func (g *governanceContract) storeVoteSet(voter []byte, voteData *VoteSet, proposalReference []byte) error {
	proposalKey := append([]byte(proposalPrefix), proposalReference...)
	voteItemKey := append(proposalKey, voter...)

	marshaledVoteItem, err := g.marshalizer.Marshal(voteData)
	if err != nil {
		return err
	}
	g.eei.SetStorage(voteItemKey, marshaledVoteItem)

	return nil
}

// setLock will set a storage key with the nonce until the funds for a specific voter are locked
func (g *governanceContract) setLock(voter []byte, voteType VoteType, proposal *GeneralProposal) {
	prefix := []byte(validatorLockPrefix)
	if voteType == Account {
		prefix = append([]byte(accountLockPrefix), proposal.GitHubCommit...)
	}
	lockKey := append(prefix, voter...)

	proposalDuration := proposal.EndVoteNonce - proposal.StartVoteNonce
	currentNonce := g.eei.BlockChainHook().CurrentNonce()
	lockNonce := big.NewInt(0).SetUint64(currentNonce + proposalDuration)

	g.eei.SetStorage(lockKey, lockNonce.Bytes())
}

// getLock returns the lock nonce for a voter
func (g *governanceContract) getLock(voter []byte, voteType VoteType, proposalReferance []byte) uint64 {
	prefix := []byte(validatorLockPrefix)
	if voteType == Account {
		prefix = append([]byte(accountLockPrefix), proposalReferance...)
	}
	lockKey := append(prefix, voter...)

	lock := g.eei.GetStorage(lockKey)

	return big.NewInt(0).SetBytes(lock).Uint64()
}

// proposalContainsVoter iterates through all the votes on a proposal and returns if it already contains a
//  vote from a certain address
func (g *governanceContract) proposalContainsVoter(proposal *GeneralProposal, voteKey []byte) bool {
	for _, vote := range proposal.Votes {
		if bytes.Equal(vote, voteKey) {
			return true
		}
	}

	return false
}

// computeVotingPower returns the voting power for a value. The value can be either a balance or
//  the staked value for a validator
func (g *governanceContract) computeVotingPower(value *big.Int) (*big.Int, error) {
	if value.Cmp(zero) == -1 {
		return nil, fmt.Errorf("cannot compute voting power on a negative value")
	}

	return big.NewInt(0).Sqrt(value), nil
}

// computeAccountLeveledPower takes a value and some voter data and returns the voting power of that value in
//  the following way: the power of all votes combined has to be sqrt(sum(allAccountVotes)). So, the new
//  vote will have a smaller power depending on how much existed previously
func (g *governanceContract) computeAccountLeveledPower(value *big.Int, voteData *VoteSet) (*big.Int, error) {
	previousAccountPower, err := g.computeVotingPower(voteData.UsedBalance)
	if err != nil {
		return nil, err
	}

	fullAccountBalance := big.NewInt(0).Add(voteData.UsedBalance, value)
	newAccountPower, err := g.computeVotingPower(fullAccountBalance)
	if err != nil {
		return nil, err
	}

	return big.NewInt(0).Sub(newAccountPower, previousAccountPower), nil
}

// getDelegatedToAddress looks into the arguments passed and returns the optional delegatedTo address
func (g *governanceContract) getDelegatedToAddress(args *vmcommon.ContractCallInput) ([]byte, error) {
	if len(args.Arguments) < 4 {
		return make([]byte, 0), nil
	}
	if len(args.Arguments[3]) != len(args.CallerAddr) {
		return nil, fmt.Errorf("%s: %s", vm.ErrInvalidArgument, "invalid delegator address length")
	}

	return args.Arguments[3], nil
}

// isValidVoteString checks if a certain string represents a valid vote string
func (g *governanceContract) isValidVoteString(vote string) bool {
	switch vote {
	case yesString, noString, vetoString:
		return true
	default:
		return false
	}
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
	default: return 0, fmt.Errorf("%s: %s%s", vm.ErrInvalidArgument, "invalid vote type option: ", vote)
	}
}

// getOrCreateVoteSet returns the vote data from storage for a goven proposer/validator pair.
//  If no vote data exists, it returns a new instance of VoteSet
func (g *governanceContract) getOrCreateVoteSet(proposal []byte, voter []byte) (*VoteSet, error) {
	key := append(proposal, voter...)
	marshaledData := g.eei.GetStorage(key)
	if len(marshaledData) == 0 {
		return g.getEmptyVoteSet(), nil
	}

	voteData := &VoteSet{}
	err := g.marshalizer.Unmarshal(voteData, marshaledData)
	if err != nil {
		return nil, err
	}

	return voteData, nil
}

// getEmptyVoteSet returns a new  VoteSet instance with it's members initialised with their 0 value
func (g *governanceContract) getEmptyVoteSet() *VoteSet {
	return &VoteSet{
		UsedPower: big.NewInt(0),
		UsedBalance: big.NewInt(0),
		TotalYes: big.NewInt(0),
		TotalNo: big.NewInt(0),
		TotalVeto: big.NewInt(0),
		VoteItems: make([]*VoteDetails, 0),
	}
}

// getTotalStake returns the total stake for a given address. It does not
//  include values from nodes that were unstaked.
// TODO: Take into account TopUp value, should discuss how this will work
//  since unBond for this funds follow a different mechanism
func (g *governanceContract) getTotalStake(address []byte) (*big.Int, error) {
	totalStake := big.NewInt(0)
	marshaledData := g.eei.GetStorageFromAddress(g.validatorSCAddress, address)
	if len(marshaledData) == 0 {
		return totalStake, nil
	}

	validatorData := &ValidatorDataV2{}
	err := g.marshalizer.Unmarshal(validatorData, marshaledData)
	if err != nil {
		return totalStake, err
	}

	for _, blsKey := range validatorData.BlsPubKeys {
		marshaledData = g.eei.GetStorageFromAddress(g.stakingSCAddress, blsKey)
		if len(marshaledData) == 0 {
			continue
		}

		nodeData := &StakedDataV2_0{}
		err = g.marshalizer.Unmarshal(nodeData, marshaledData)
		if err != nil {
			return big.NewInt(0), err
		}

		if !nodeData.Staked {
			continue
		}

		totalStake.Add(totalStake, nodeData.StakeValue)
	}

	return totalStake, nil
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

// validateInitialWhiteListedAddresses makes basic checks that the provided initial whitelisted
//  addresses have the correct format
func (g *governanceContract) validateInitialWhiteListedAddresses(addresses [][]byte) error {
	if len(addresses) == 0 {
		log.Debug("0 initial whiteListed addresses provided to the governance contract")
		return nil
	}

	for _, addr := range addresses {
		if len(addr) != len(g.ownerAddress) {
			return fmt.Errorf("invalid address length for %s", string(addr))
		}
	}

	return nil
}

// startEndNonceFromArguments converts the nonce string arguments to uint64
func (g *governanceContract) startEndNonceFromArguments(argStart []byte, argEnd []byte) (uint64, uint64, error) {
	startVoteNonce, err := g.nonceFromBytes(argStart)
	if err != nil{
		return 0, 0, err
	}
	endVoteNonce, err := g.nonceFromBytes(argEnd)
	if err != nil{
		return 0, 0, err
	}

	currentNonce := g.eei.BlockChainHook().CurrentNonce()
	if currentNonce > startVoteNonce.Uint64() || startVoteNonce.Uint64() > endVoteNonce.Uint64() {
		return 0, 0, vm.ErrInvalidStartEndVoteNonce
	}

	return startVoteNonce.Uint64(), endVoteNonce.Uint64(), nil
}

// nonceFromBytes converts a byte array to a big.Int. Returns ErrInvalidStartEndVoteNonce for invalid values
func (g *governanceContract) nonceFromBytes(nonce []byte) (*big.Int, error) {
	voteNonce, okConvert := big.NewInt(0).SetString(string(nonce), conversionBase)
	if !okConvert {
		return nil, vm.ErrInvalidStartEndVoteNonce
	}
	if !voteNonce.IsUint64() {
		return nil, vm.ErrInvalidStartEndVoteNonce
	}

	return voteNonce, nil
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
		proposal.Voted = false
		return nil
	}

	if proposal.Veto.Cmp(baseConfig.MinVetoThreshold) >= 0 {
		proposal.Voted = false
		return nil
	}

	if proposal.Yes.Cmp(baseConfig.MinPassThreshold) >= 0 && proposal.Yes.Cmp(proposal.No) == 1 {
		proposal.Voted = true
		return nil
	}

	proposal.Voted = false
	return nil
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
		MinQuorum: minQuorum,
		MinPassThreshold: minPass,
		MinVetoThreshold: minVeto,
		ProposalFee: proposalFee,
	}, nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (g *governanceContract) EpochConfirmed(epoch uint32) {
	g.flagEnabled.Toggle(epoch >= g.enabledEpoch)
	log.Debug("governance contract", "enabled", g.flagEnabled.IsSet())
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
