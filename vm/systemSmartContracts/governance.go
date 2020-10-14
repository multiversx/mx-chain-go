//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. governance.proto
package systemSmartContracts

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const governanceConfigKey = "governanceConfig"
const hardForkPrefix = "hardFork"
const proposalPrefix = "proposal"
const whiteListPrefix = "whiteList"
const validatorPrefix = "validator"
const hardForkEpochGracePeriod = 2
const githubCommitLength = 40

// ArgsNewGovernanceContract defines the arguments needed for the on-chain governance contract
type ArgsNewGovernanceContract struct {
	Eei                 vm.SystemEI
	GasCost             vm.GasCost
	GovernanceConfig    config.GovernanceSystemSCConfig
	ESDTSCAddress       []byte
	Marshalizer         marshal.Marshalizer
	Hasher              hashing.Hasher
	GovernanceSCAddress []byte
	StakingSCAddress    []byte
	AuctionSCAddress    []byte
	EpochNotifier       vm.EpochNotifier
}

type governanceContract struct {
	eei                 vm.SystemEI
	gasCost             vm.GasCost
	baseProposalCost    *big.Int
	ownerAddress        []byte
	governanceSCAddress []byte
	stakingSCAddress    []byte
	auctionSCAddress    []byte
	marshalizer         marshal.Marshalizer
	hasher              hashing.Hasher
	governanceConfig    config.GovernanceSystemSCConfig
	enabledEpoch        uint32
	flagEnabled         atomic.Flag
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

	baseProposalCost, okConvert := big.NewInt(0).SetString(args.GovernanceConfig.ProposalCost, conversionBase)
	if !okConvert || baseProposalCost.Cmp(big.NewInt(0)) < 0 {
		return nil, vm.ErrInvalidBaseIssuingCost
	}

	g := &governanceContract{
		eei:                 args.Eei,
		gasCost:             args.GasCost,
		baseProposalCost:    baseProposalCost,
		ownerAddress:        nil,
		governanceSCAddress: args.GovernanceSCAddress,
		stakingSCAddress:    args.StakingSCAddress,
		auctionSCAddress:    args.AuctionSCAddress,
		marshalizer:         args.Marshalizer,
		hasher:              args.Hasher,
		governanceConfig:    args.GovernanceConfig,
		enabledEpoch:        args.GovernanceConfig.EnabledEpoch,
	}
	args.EpochNotifier.RegisterNotifyHandler(g)

	return g, nil
}

// Execute calls one of the functions from the governance smart contract and runs the code according to the input
func (g *governanceContract) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
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
	case "whiteList":
		return g.whiteListProposal(args)
	case "hardFork":
		return g.hardForkProposal(args)
	case "proposal":
		return g.proposal(args)
	case "vote":
		return g.vote(args)
	case "delegateVotePower":
		return g.delegateVotePower(args)
	case "revokeVotePower":
		return g.revokeVotePower(args)
	case "changeConfig":
		return g.changeConfig(args)
	case "closeProposal":
		return g.closeProposal(args)
	}

	g.eei.AddReturnMessage("invalid method to call")
	return vmcommon.FunctionNotFound
}

func (g *governanceContract) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	scConfig := &GovernanceConfig{
		NumNodes:         g.governanceConfig.NumNodes,
		MinQuorum:        g.governanceConfig.MinQuorum,
		MinPassThreshold: g.governanceConfig.MinPassThreshold,
		MinVetoThreshold: g.governanceConfig.MinVetoThreshold,
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
	if !okConvert || numNodes.Cmp(big.NewInt(0)) < 0 {
		g.eei.AddReturnMessage("changeConfig first argument is incorrectly formatted")
		return vmcommon.UserError
	}
	minQuorum, okConvert := big.NewInt(0).SetString(string(args.Arguments[1]), conversionBase)
	if !okConvert || minQuorum.Cmp(big.NewInt(0)) < 0 {
		g.eei.AddReturnMessage("changeConfig second argument is incorrectly formatted")
		return vmcommon.UserError
	}
	minVeto, okConvert := big.NewInt(0).SetString(string(args.Arguments[2]), conversionBase)
	if !okConvert || minVeto.Cmp(big.NewInt(0)) < 0 {
		g.eei.AddReturnMessage("changeConfig third argument is incorrectly formatted")
		return vmcommon.UserError
	}
	minPass, okConvert := big.NewInt(0).SetString(string(args.Arguments[3]), conversionBase)
	if !okConvert || minPass.Cmp(big.NewInt(0)) < 0 {
		g.eei.AddReturnMessage("changeConfig fourth argument is incorrectly formatted")
		return vmcommon.UserError
	}

	scConfig, err := g.getConfig()
	if err != nil {
		g.eei.AddReturnMessage("changeConfig error " + err.Error())
		return vmcommon.UserError
	}

	scConfig.NumNodes = numNodes.Int64()
	scConfig.MinQuorum = int32(minQuorum.Int64())
	scConfig.MinVetoThreshold = int32(minVeto.Int64())
	scConfig.MinPassThreshold = int32(minPass.Int64())

	marshaledData, err := g.marshalizer.Marshal(scConfig)
	if err != nil {
		g.eei.AddReturnMessage("changeConfig error " + err.Error())
		return vmcommon.UserError
	}
	g.eei.SetStorage([]byte(governanceConfigKey), marshaledData)

	return vmcommon.Ok
}

func (g *governanceContract) getConfig() (*GovernanceConfig, error) {
	marshaledData := g.eei.GetStorage([]byte(governanceConfigKey))
	scConfig := &GovernanceConfig{}
	err := g.marshalizer.Unmarshal(scConfig, marshaledData)
	if err != nil {
		return nil, err
	}

	return scConfig, nil
}

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
		Yes:            0,
		No:             0,
		Veto:           0,
		DontCare:       0,
		Voted:          false,
		TopReference:   key,
		Voters:         make([][]byte, 0),
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

func (g *governanceContract) saveGeneralProposal(reference []byte, generalProposal *GeneralProposal) error {
	marshaledData, err := g.marshalizer.Marshal(generalProposal)
	if err != nil {
		return err
	}
	key := append([]byte(proposalPrefix), reference...)
	g.eei.SetStorage(key, marshaledData)

	return nil
}

func (g *governanceContract) startEndNonceFromArguments(argStart []byte, argEnd []byte) (uint64, uint64, error) {
	startVoteNonce, okConvert := big.NewInt(0).SetString(string(argStart), conversionBase)
	if !okConvert {
		return 0, 0, vm.ErrInvalidStartEndVoteNonce
	}
	if !startVoteNonce.IsUint64() {
		return 0, 0, vm.ErrInvalidStartEndVoteNonce
	}
	endVoteNonce, okConvert := big.NewInt(0).SetString(string(argEnd), conversionBase)
	if !okConvert {
		return 0, 0, vm.ErrInvalidStartEndVoteNonce
	}
	if !endVoteNonce.IsUint64() {
		return 0, 0, vm.ErrInvalidStartEndVoteNonce
	}
	currentNonce := g.eei.BlockChainHook().CurrentNonce()
	if currentNonce > startVoteNonce.Uint64() || startVoteNonce.Uint64() > endVoteNonce.Uint64() {
		return 0, 0, vm.ErrInvalidStartEndVoteNonce
	}

	return startVoteNonce.Uint64(), endVoteNonce.Uint64(), nil
}

func (g *governanceContract) proposalExists(reference []byte) bool {
	key := append([]byte(proposalPrefix), reference...)
	marshaledData := g.eei.GetStorage(key)
	return len(marshaledData) > 0
}

func (g *governanceContract) getGeneralProposal(reference []byte) (*GeneralProposal, error) {
	key := append([]byte(proposalPrefix), reference...)
	marshaledData := g.eei.GetStorage(key)

	if len(marshaledData) == 0 {
		return nil, vm.ErrEmptyStorage
	}

	generalProposal := &GeneralProposal{}
	err := g.marshalizer.Unmarshal(generalProposal, marshaledData)
	if err != nil {
		return nil, err
	}

	return generalProposal, nil
}

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

	key = append([]byte(whiteListPrefix), args.CallerAddr...)
	generalProposal := &GeneralProposal{
		IssuerAddress:  args.CallerAddr,
		GitHubCommit:   []byte("genesis"),
		StartVoteNonce: 0,
		EndVoteNonce:   0,
		Yes:            int32(g.governanceConfig.NumNodes),
		No:             0,
		Veto:           0,
		DontCare:       0,
		Voted:          true,
		TopReference:   key,
		Voters:         make([][]byte, 0),
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
		Yes:            0,
		No:             0,
		Veto:           0,
		DontCare:       0,
		Voted:          false,
		TopReference:   key,
		Voters:         make([][]byte, 0),
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
		log.Warn("save general proposal", err, "error")
		g.eei.AddReturnMessage("saveGeneralProposal" + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

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
		Yes:            0,
		No:             0,
		Veto:           0,
		DontCare:       0,
		Voted:          false,
		Voters:         make([][]byte, 0),
	}
	err = g.saveGeneralProposal(gitHubCommit, generalProposal)
	if err != nil {
		log.Warn("saveGeneralProposal", "err", err)
		g.eei.AddReturnMessage("saveGeneralProposal" + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (g *governanceContract) vote(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		g.eei.AddReturnMessage("invalid proposal cost, expected 0")
		return vmcommon.OutOfFunds
	}
	err := g.eei.UseGas(g.gasCost.MetaChainSystemSCsCost.Vote)
	if err != nil {
		g.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) < 2 || len(args.Arguments) > 3 {
		g.eei.AddReturnMessage("invalid number of argument expected 2 or 3")
		return vmcommon.FunctionWrongSignature
	}
	if len(args.Arguments) == 3 && len(args.Arguments[2]) != len(args.CallerAddr) {
		g.eei.AddReturnMessage("wrong argument number 3 should be a valid address")
		return vmcommon.FunctionWrongSignature
	}
	if len(args.Arguments) == 3 && bytes.Equal(args.CallerAddr, args.Arguments[2]) {
		g.eei.AddReturnMessage("wrong argument number 3 should be different than caller")
		return vmcommon.FunctionWrongSignature
	}

	proposalToVote := args.Arguments[0]
	if !g.proposalExists(proposalToVote) {
		g.eei.AddReturnMessage("proposal does not exists")
		return vmcommon.UserError
	}

	voteString := string(args.Arguments[1])
	if !g.isValidVoteString(voteString) {
		g.eei.AddReturnMessage("argument 1 is not a valid vote string")
		return vmcommon.UserError
	}

	voterAddress := args.CallerAddr
	validatorAddress := args.CallerAddr
	if len(args.Arguments) == 3 {
		validatorAddress = args.Arguments[2]
	}
	numStakedNodes, err := g.numOfStakedNodes(validatorAddress)
	if err != nil || numStakedNodes == 0 {
		g.eei.AddReturnMessage("address has 0 voting power")
		return vmcommon.UserError
	}

	numNodesToVote := int32(0)
	validatorData, err := g.getOrCreateValidatorData(validatorAddress, int32(numStakedNodes))
	if err != nil {
		log.Warn("getOrCreateValidatorData", "err", err)
		g.eei.AddReturnMessage("getOrCreateValidator data error" + err.Error())
		return vmcommon.UserError
	}

	found := false
	for _, voter := range validatorData.Delegators {
		if bytes.Equal(voter.Address, voterAddress) {
			found = true
			numNodesToVote = voter.NumNodes
			break
		}
	}
	if !found || numNodesToVote <= 0 {
		g.eei.AddReturnMessage("address has 0 voting power")
		return vmcommon.UserError
	}

	err = g.voteForProposal(proposalToVote, voteString, voterAddress, numNodesToVote)
	if err != nil {
		g.eei.AddReturnMessage("voteForProposal " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (g *governanceContract) isValidVoteString(vote string) bool {
	switch vote {
	case "yes":
		return true
	case "no":
		return true
	case "veto":
		return true
	case "dontCare":
		return true
	}
	return false
}

func (g *governanceContract) voteForProposal(
	proposal []byte,
	vote string,
	voter []byte,
	numVotes int32,
) error {
	voteData, err := g.getOrCreateVoteData(proposal, voter)
	if err != nil {
		log.Warn("getOrCreateVoteData", "err", err)
		return err
	}
	if voteData.NumVotes == numVotes && voteData.VoteValue == vote {
		return nil
	}

	oldNum := voteData.NumVotes
	oldValue := voteData.VoteValue

	voteData.NumVotes = numVotes
	voteData.VoteValue = vote
	err = g.saveVoteValue(proposal, voter, voteData)
	if err != nil {
		log.Warn("saveVoteValue", "err", err)
		return err
	}

	generalProposal, err := g.getGeneralProposal(proposal)
	if err != nil {
		return err
	}
	currentNonce := g.eei.BlockChainHook().CurrentNonce()
	if currentNonce < generalProposal.StartVoteNonce {
		return vm.ErrVotedForAProposalThatNotBeginsYet
	}

	if currentNonce > generalProposal.EndVoteNonce {
		return vm.ErrVotedForAnExpiredProposal
	}

	generalProposal.Voters = append(generalProposal.Voters, voter)
	g.addVotedDataToProposal(generalProposal, oldValue, -oldNum)
	g.addVotedDataToProposal(generalProposal, vote, numVotes)

	err = g.saveGeneralProposal(proposal, generalProposal)
	if err != nil {
		log.Warn("saveGeneralProposal", "err", err)
		return err
	}

	return nil
}

func (g *governanceContract) addVotedDataToProposal(generalProposal *GeneralProposal, voteValue string, numVotes int32) {
	switch voteValue {
	case "yes":
		generalProposal.Yes += numVotes
	case "no":
		generalProposal.No += numVotes
	case "veto":
		generalProposal.Veto += numVotes
	case "dontCare":
		generalProposal.DontCare += numVotes
	}
}

func (g *governanceContract) saveVoteValue(proposal []byte, voter []byte, voteData *VoteData) error {
	key := append(proposal, voter...)
	marshaledData, err := g.marshalizer.Marshal(voteData)
	if err != nil {
		return err
	}

	g.eei.SetStorage(key, marshaledData)
	return nil
}

func (g *governanceContract) getOrCreateVoteData(proposal []byte, voter []byte) (*VoteData, error) {
	voteData := &VoteData{}
	key := append(proposal, voter...)
	marshaledData := g.eei.GetStorage(key)
	if len(marshaledData) == 0 {
		return voteData, nil
	}

	err := g.marshalizer.Unmarshal(voteData, marshaledData)
	if err != nil {
		return nil, err
	}

	return voteData, nil
}

func (g *governanceContract) getOrCreateValidatorData(address []byte, numNodes int32) (*ValidatorData, error) {
	validatorData := &ValidatorData{
		Delegators: make([]*VoterData, 1),
		NumNodes:   numNodes,
	}
	validatorData.Delegators[0] = &VoterData{
		Address:  address,
		NumNodes: numNodes,
	}

	key := append([]byte(validatorPrefix), address...)
	marshaledData := g.eei.GetStorage(key)
	if len(marshaledData) == 0 {
		return validatorData, nil
	}

	err := g.marshalizer.Unmarshal(validatorData, marshaledData)
	if err != nil {
		return nil, err
	}

	oldNumNodes := validatorData.NumNodes
	validatorData.NumNodes = numNodes
	if len(validatorData.Delegators) == 1 {
		return validatorData, nil
	}

	log.Trace("difference in old num nodes and new num nodes with delegated voting", oldNumNodes, numNodes)

	return validatorData, nil
}

func (g *governanceContract) delegateVotePower(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	g.eei.AddReturnMessage("delegateVotePower not yet implemented")
	return vmcommon.UserError
}

func (g *governanceContract) revokeVotePower(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	g.eei.AddReturnMessage("revokeVotePower not yet implemented")
	return vmcommon.UserError
}

func (g *governanceContract) numOfStakedNodes(address []byte) (uint32, error) {
	marshaledData := g.eei.GetStorageFromAddress(g.auctionSCAddress, address)
	if len(marshaledData) == 0 {
		return 0, nil
	}

	auctionData := &AuctionDataV2{}
	err := g.marshalizer.Unmarshal(auctionData, marshaledData)
	if err != nil {
		return 0, err
	}

	numStakedNodes := uint32(0)
	for _, blsKey := range auctionData.BlsPubKeys {
		marshaledData = g.eei.GetStorageFromAddress(g.stakingSCAddress, blsKey)
		if len(marshaledData) == 0 {
			continue
		}

		nodeData := &StakedDataV2{}
		err = g.marshalizer.Unmarshal(nodeData, marshaledData)
		if err != nil {
			return 0, err
		}

		if nodeData.Staked {
			numStakedNodes++
		}
	}

	return numStakedNodes, nil
}

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

	for _, voter := range generalProposal.Voters {
		key := append(proposal, voter...)
		g.eei.SetStorage(key, nil)
	}

	return vmcommon.Ok
}

func (g *governanceContract) computeEndResults(proposal *GeneralProposal) error {
	baseConfig, err := g.getConfig()
	if err != nil {
		return err
	}
	totalVotes := proposal.Yes + proposal.No + proposal.DontCare + proposal.Veto
	if totalVotes < baseConfig.MinQuorum {
		proposal.Voted = false
		return nil
	}

	if proposal.Veto > baseConfig.MinVetoThreshold {
		proposal.Voted = false
		return nil
	}

	if proposal.Yes > baseConfig.MinPassThreshold {
		proposal.Voted = true
		return nil
	}

	return nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (g *governanceContract) EpochConfirmed(epoch uint32) {
	g.flagEnabled.Toggle(epoch >= g.enabledEpoch)
	log.Debug("governance contract", "enabled", g.flagEnabled.IsSet())
}

// IsInterfaceNil returns true if underlying object is nil
func (g *governanceContract) IsInterfaceNil() bool {
	return g == nil
}
