//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. governance.proto
package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
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

	baseProposalCost, ok := big.NewInt(0).SetString(args.GovernanceConfig.ProposalCost, conversionBase)
	if !ok || baseProposalCost.Cmp(big.NewInt(0)) < 0 {
		return nil, vm.ErrInvalidBaseIssuingCost
	}

	return &governanceContract{
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
	}, nil
}

// Execute calls one of the functions from the esdt smart contract and runs the code according to the input
func (g *governanceContract) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if CheckIfNil(args) != nil {
		return vmcommon.UserError
	}

	switch args.Function {
	case core.SCDeployInitFunctionName:
		return g.init(args)
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
	case "changeNumOfNodes":
		return g.changeNumOfNodes(args)
	}

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
	log.LogIfError(err, "marshal error on esdt init function")

	g.eei.SetStorage([]byte(governanceConfigKey), marshaledData)
	g.eei.SetStorage([]byte(ownerKey), args.CallerAddr)
	g.ownerAddress = make([]byte, 0, len(args.CallerAddr))
	g.ownerAddress = append(g.ownerAddress, args.CallerAddr...)
	return vmcommon.Ok
}

func (g *governanceContract) changeNumOfNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(g.ownerAddress, args.CallerAddr) {
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		return vmcommon.UserError
	}

	numNodes, ok := big.NewInt(0).SetString(string(args.Arguments[0]), conversionBase)
	if !ok || numNodes.Cmp(big.NewInt(0)) < 0 {
		return vmcommon.UserError
	}
	scConfig, err := g.getConfig()
	if err != nil {
		return vmcommon.UserError
	}

	scConfig.NumNodes = numNodes.Int64()
	marshalledData, err := g.marshalizer.Marshal(scConfig)
	if err != nil {
		return vmcommon.UserError
	}
	g.eei.SetStorage([]byte(governanceConfigKey), marshalledData)

	return vmcommon.Ok
}

func (g *governanceContract) getConfig() (*GovernanceConfig, error) {
	marshalledData := g.eei.GetStorage([]byte(governanceConfigKey))
	scConfig := &GovernanceConfig{}
	err := g.marshalizer.Unmarshal(scConfig, marshalledData)
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
		return vmcommon.OutOfFunds
	}
	if len(args.Arguments) != 4 {
		return vmcommon.FunctionWrongSignature
	}
	if len(args.Arguments[0]) != len(args.CallerAddr) {
		return vmcommon.UserError
	}
	if g.proposalExists(args.Arguments[0]) {
		return vmcommon.UserError
	}
	if g.isWhiteListed(args.CallerAddr) {
		return vmcommon.UserError
	}

	startVoteNonce, endVoteNonce, err := g.startEndNonceFromArguments(args.Arguments[2], args.Arguments[3])
	if err != nil {
		return vmcommon.UserError
	}

	key := append([]byte(proposalPrefix), args.Arguments[0]...)
	whiteListAcc := &WhiteListProposal{
		WhiteListAddress: args.Arguments[0],
		ProposalStatus:   key,
	}

	key = append([]byte(whiteListPrefix), args.Arguments[0]...)
	generalProposal := &GeneralProposal{
		IssuerAddress:  args.CallerAddr,
		GitHubCommit:   args.Arguments[1],
		StartVoteNonce: startVoteNonce,
		EndVoteNonce:   endVoteNonce,
		Yes:            0,
		No:             0,
		Veto:           0,
		DontCare:       0,
		Voted:          false,
		TopReference:   key,
	}

	marshalledData, err := g.marshalizer.Marshal(whiteListAcc)
	if err != nil {
		return vmcommon.UserError
	}

	key = append([]byte(whiteListPrefix), args.Arguments[0]...)
	g.eei.SetStorage(key, marshalledData)

	err = g.saveGeneralProposal(args.Arguments[0], generalProposal)
	if err != nil {
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (g *governanceContract) saveGeneralProposal(reference []byte, generalProposal *GeneralProposal) error {
	marshalledData, err := g.marshalizer.Marshal(generalProposal)
	if err != nil {
		return err
	}
	key := append([]byte(proposalPrefix), reference...)
	g.eei.SetStorage(key, marshalledData)

	return nil
}

func (g *governanceContract) startEndNonceFromArguments(argStart []byte, argEnd []byte) (uint64, uint64, error) {
	startVoteNonce, ok := big.NewInt(0).SetString(string(argStart), conversionBase)
	if !ok {
		return 0, 0, vm.ErrInvalidStartEndVoteNonce
	}
	if !startVoteNonce.IsUint64() {
		return 0, 0, vm.ErrInvalidStartEndVoteNonce
	}
	endVoteNonce, ok := big.NewInt(0).SetString(string(argEnd), conversionBase)
	if !ok {
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
	marshalledData := g.eei.GetStorage(key)
	return len(marshalledData) > 0
}

func (g *governanceContract) getGeneralProposal(reference []byte) (*GeneralProposal, error) {
	key := append([]byte(proposalPrefix), reference...)
	marshalledData := g.eei.GetStorage(key)

	if len(marshalledData) == 0 {
		return nil, vm.ErrEmptyStorage
	}

	generalProposal := &GeneralProposal{}
	err := g.marshalizer.Unmarshal(generalProposal, marshalledData)
	if err != nil {
		return nil, err
	}

	return generalProposal, nil
}

func (g *governanceContract) isWhiteListed(address []byte) bool {
	key := append([]byte(whiteListPrefix), address...)
	marshalledData := g.eei.GetStorage(key)
	if len(marshalledData) == 0 {
		return false
	}

	key = append([]byte(proposalPrefix), address...)
	marshalledData = g.eei.GetStorage(key)
	generalProposal := &GeneralProposal{}
	err := g.marshalizer.Unmarshal(generalProposal, marshalledData)
	if err != nil {
		return false
	}

	return generalProposal.Voted
}

func (g *governanceContract) whiteListAtGenesis(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		return vmcommon.UserError
	}
	if g.isWhiteListed(args.CallerAddr) {
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		return vmcommon.UserError
	}
	if len(args.Arguments[0]) != len(args.CallerAddr) {
		return vmcommon.UserError
	}
	if g.proposalExists(args.Arguments[0]) {
		return vmcommon.UserError
	}

	key := append([]byte(proposalPrefix), args.Arguments[0]...)
	whiteListAcc := &WhiteListProposal{
		WhiteListAddress: args.Arguments[0],
		ProposalStatus:   key,
	}

	key = append([]byte(whiteListPrefix), args.Arguments[0]...)
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
	}
	marshalledData, err := g.marshalizer.Marshal(whiteListAcc)
	if err != nil {
		return vmcommon.UserError
	}

	key = append([]byte(whiteListPrefix), args.Arguments[0]...)
	g.eei.SetStorage(key, marshalledData)

	err = g.saveGeneralProposal(args.Arguments[0], generalProposal)
	if err != nil {
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (g *governanceContract) hardForkProposal(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(g.baseProposalCost) != 0 {
		return vmcommon.OutOfFunds
	}
	if len(args.Arguments) != 4 {
		return vmcommon.FunctionWrongSignature
	}
	if !g.isWhiteListed(args.CallerAddr) {
		return vmcommon.UserError
	}
	gitHubCommit := args.Arguments[2]
	if g.proposalExists(gitHubCommit) {
		return vmcommon.UserError
	}

	key := append([]byte(hardForkPrefix), gitHubCommit...)
	marshalledData := g.eei.GetStorage(key)
	if len(marshalledData) != 0 {
		return vmcommon.UserError
	}

	startVoteNonce, endVoteNonce, err := g.startEndNonceFromArguments(args.Arguments[3], args.Arguments[4])
	if err != nil {
		return vmcommon.UserError
	}

	bigIntEpochToHardFork, ok := big.NewInt(0).SetString(string(args.Arguments[0]), conversionBase)
	if !ok || !bigIntEpochToHardFork.IsUint64() {
		return vmcommon.UserError
	}

	epochToHardFork := uint32(bigIntEpochToHardFork.Uint64())
	currentEpoch := g.eei.BlockChainHook().CurrentEpoch()
	if epochToHardFork < currentEpoch && currentEpoch-epochToHardFork < hardForkEpochGracePeriod {
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
	}
	marshalledData, err = g.marshalizer.Marshal(hardForkProposal)
	if err != nil {
		return vmcommon.UserError
	}
	g.eei.SetStorage(key, marshalledData)

	err = g.saveGeneralProposal(args.Arguments[0], generalProposal)
	if err != nil {
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (g *governanceContract) proposal(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(g.baseProposalCost) != 0 {
		return vmcommon.OutOfFunds
	}
	if len(args.Arguments) != 4 {
		return vmcommon.FunctionWrongSignature
	}
	if !g.isWhiteListed(args.CallerAddr) {
		return vmcommon.UserError
	}
	gitHubCommit := args.Arguments[0]
	if g.proposalExists(gitHubCommit) {
		return vmcommon.UserError
	}

	startVoteNonce, endVoteNonce, err := g.startEndNonceFromArguments(args.Arguments[1], args.Arguments[2])
	if err != nil {
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
	}
	err = g.saveGeneralProposal(gitHubCommit, generalProposal)
	if err != nil {
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (g *governanceContract) vote(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		return vmcommon.OutOfFunds
	}
	if len(args.Arguments) < 2 || len(args.Arguments) > 3 {
		return vmcommon.FunctionWrongSignature
	}
	if len(args.Arguments) == 3 && len(args.Arguments[2]) != len(args.CallerAddr) {
		return vmcommon.FunctionWrongSignature
	}
	if len(args.Arguments) == 3 && bytes.Equal(args.CallerAddr, args.Arguments[2]) {
		return vmcommon.FunctionWrongSignature
	}

	proposalToVote := args.Arguments[0]
	if !g.proposalExists(proposalToVote) {
		return vmcommon.UserError
	}

	voteString := string(args.Arguments[1])
	if !g.isValidVoteString(voteString) {
		return vmcommon.UserError
	}

	voterAddress := args.CallerAddr
	validatorAddress := args.CallerAddr
	if len(args.Arguments) == 3 {
		validatorAddress = args.Arguments[2]
	}
	numStakedNodes, err := g.numOfStakedNodes(validatorAddress)
	if err != nil || numStakedNodes == 0 {
		return vmcommon.UserError
	}

	numNodesToVote := int32(0)
	validatorData, err := g.getOrCreateValidatorData(validatorAddress, int32(numStakedNodes))
	if err != nil {
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
		return vmcommon.UserError
	}

	err = g.voteForProposal(proposalToVote, voteString, voterAddress, numNodesToVote)
	if err != nil {
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
		return err
	}

	generalProposal, err := g.getGeneralProposal(proposal)
	if err != nil {
		return err
	}

	g.addVotedDataToProposal(generalProposal, oldValue, -oldNum)
	g.addVotedDataToProposal(generalProposal, vote, numVotes)

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
	marshalledData, err := g.marshalizer.Marshal(voteData)
	if err != nil {
		return err
	}

	g.eei.SetStorage(key, marshalledData)
	return nil
}

func (g *governanceContract) getOrCreateVoteData(proposal []byte, voter []byte) (*VoteData, error) {
	voteData := &VoteData{}
	key := append(proposal, voter...)
	marshalledData := g.eei.GetStorage(key)
	if len(marshalledData) == 0 {
		return voteData, nil
	}

	err := g.marshalizer.Unmarshal(voteData, marshalledData)
	if err != nil {
		return nil, err
	}

	return voteData, nil
}

func (g *governanceContract) getOrCreateValidatorData(address []byte, numNodes int32) (*ValidatorData, error) {
	validatorData := &ValidatorData{
		Delegators: make([]*VoterData, 0, 1),
		NumNodes:   numNodes,
	}
	validatorData.Delegators[0] = &VoterData{
		Address:  address,
		NumNodes: numNodes,
	}

	key := append([]byte(validatorPrefix), address...)
	marshalledData := g.eei.GetStorage(key)
	if len(marshalledData) == 0 {
		return validatorData, nil
	}

	err := g.marshalizer.Unmarshal(validatorData, marshalledData)
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
	log.Trace("delegateVotePower not yet implemented")
	return vmcommon.UserError
}

func (g *governanceContract) revokeVotePower(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	log.Trace("revokeVotePower not yet implemented")
	return vmcommon.UserError
}

func (g *governanceContract) executeOnAuctionSC(data []byte) (*vmcommon.VMOutput, error) {
	return g.eei.ExecuteOnDestContext(g.stakingSCAddress, g.governanceSCAddress, big.NewInt(0), data)
}

func (g *governanceContract) numOfStakedNodes(address []byte) (uint32, error) {
	txData := "get" + "@" + hex.EncodeToString(address)
	vmOutput, err := g.executeOnAuctionSC([]byte(txData))
	if err != nil {
		return 0, err
	}
	if vmOutput.ReturnCode != vmcommon.UserError {
		return 0, vm.ErrNotEnoughQualifiedNodes
	}
	if len(vmOutput.ReturnData) == 0 {
		return 0, vm.ErrNotEnoughQualifiedNodes
	}

	auctionData := &AuctionData{}
	err = g.marshalizer.Unmarshal(auctionData, vmOutput.ReturnData[0])
	if err != nil {
		return 0, err
	}

	return auctionData.NumStaked, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (g *governanceContract) IsInterfaceNil() bool {
	return g == nil
}
