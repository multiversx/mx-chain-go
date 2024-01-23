//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf --gogoslick_out=. governance.proto
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
const accumulatedFeeKey = "accumulatedFee"
const noncePrefix = "n_"
const proposalPrefix = "p_"
const yesString = "yes"
const noString = "no"
const vetoString = "veto"
const abstainString = "abstain"
const commitHashLength = 40
const maxPercentage = float64(10000.0)

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
	OwnerAddress           []byte
	UnBondPeriodInEpochs   uint32
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
	unBondPeriodInEpochs   uint32
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
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.GovernanceFlag,
	})
	if err != nil {
		return nil, err
	}

	baseProposalCost, okConvert := big.NewInt(0).SetString(args.GovernanceConfig.V1.ProposalCost, conversionBase)
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
	if len(args.OwnerAddress) < 1 {
		return nil, fmt.Errorf("%w for change config address", vm.ErrInvalidAddress)
	}

	g := &governanceContract{
		eei:                    args.Eei,
		gasCost:                args.GasCost,
		baseProposalCost:       baseProposalCost,
		ownerAddress:           args.OwnerAddress,
		governanceSCAddress:    args.GovernanceSCAddress,
		delegationMgrSCAddress: args.DelegationMgrSCAddress,
		validatorSCAddress:     args.ValidatorSCAddress,
		marshalizer:            args.Marshalizer,
		hasher:                 args.Hasher,
		governanceConfig:       args.GovernanceConfig,
		enableEpochsHandler:    args.EnableEpochsHandler,
		unBondPeriodInEpochs:   args.UnBondPeriodInEpochs,
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

	if !g.enableEpochsHandler.IsFlagEnabled(common.GovernanceFlag) {
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
	case "viewVotingPower":
		return g.viewVotingPower(args)
	case "viewConfig":
		return g.viewConfig(args)
	case "viewUserVoteHistory":
		return g.viewUserVoteHistory(args)
	case "viewDelegatedVoteInfo":
		return g.viewDelegatedVoteInfo(args)
	case "viewProposal":
		return g.viewProposal(args)
	case "claimAccumulatedFees":
		return g.claimAccumulatedFees(args)
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
	return vmcommon.Ok
}

func (g *governanceContract) initV2(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, g.governanceSCAddress) {
		g.eei.AddReturnMessage("invalid caller to switch to V2 config")
		return vmcommon.UserError
	}
	cfg, err := g.convertV2Config(g.governanceConfig)
	if err != nil {
		g.eei.AddReturnMessage("could not create governance V2 config")
		return vmcommon.UserError
	}

	err = g.saveConfig(cfg)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	g.eei.SetStorage([]byte(ownerKey), args.CallerAddr)

	return vmcommon.Ok
}

// changeConfig allows the owner to change the configuration for requesting proposals
//
//	args.Arguments[0] - proposalFee - as string
//	args.Arguments[1] - lostProposalFee - as string
//	args.Arguments[2] - minQuorum - 0-10000 - represents percentage
//	args.Arguments[3] - minVeto   - 0-10000 - represents percentage
//	args.Arguments[4] - minPass   - 0-10000 - represents percentage
func (g *governanceContract) changeConfig(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(g.ownerAddress, args.CallerAddr) {
		g.eei.AddReturnMessage("changeConfig can be called only by owner")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		g.eei.AddReturnMessage("changeConfig can be called only without callValue")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 5 {
		g.eei.AddReturnMessage("changeConfig needs 5 arguments")
		return vmcommon.UserError
	}

	proposalFee, okConvert := big.NewInt(0).SetString(string(args.Arguments[0]), conversionBase)
	if !okConvert || proposalFee.Cmp(zero) <= 0 {
		g.eei.AddReturnMessage("changeConfig first argument is incorrectly formatted")
		return vmcommon.UserError
	}
	lostProposalFee, okConvert := big.NewInt(0).SetString(string(args.Arguments[1]), conversionBase)
	if !okConvert || proposalFee.Cmp(zero) <= 0 {
		g.eei.AddReturnMessage("changeConfig second argument is incorrectly formatted")
		return vmcommon.UserError
	}
	if proposalFee.Cmp(lostProposalFee) < 0 {
		errLocal := fmt.Errorf("%w proposal fee is smaller than lost proposal fee ", vm.ErrIncorrectConfig)
		g.eei.AddReturnMessage(errLocal.Error())
		return vmcommon.UserError
	}

	minQuorum, err := convertDecimalToPercentage(args.Arguments[2])
	if err != nil {
		g.eei.AddReturnMessage(err.Error() + " minQuorum")
		return vmcommon.UserError
	}
	minVeto, err := convertDecimalToPercentage(args.Arguments[3])
	if err != nil {
		g.eei.AddReturnMessage(err.Error() + " minVeto")
		return vmcommon.UserError
	}
	minPass, err := convertDecimalToPercentage(args.Arguments[4])
	if err != nil {
		g.eei.AddReturnMessage(err.Error() + " minPass")
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
	scConfig.LostProposalFee = lostProposalFee

	g.baseProposalCost.Set(proposalFee)
	err = g.saveConfig(scConfig)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

// proposal creates a new proposal from passed arguments
func (g *governanceContract) proposal(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if g.enableEpochsHandler.IsFlagEnabled(common.GovernanceDisableProposeFlag) && !g.enableEpochsHandler.IsFlagEnabled(common.GovernanceFixesFlag) {
		g.eei.AddReturnMessage("proposing is disabled")
		return vmcommon.UserError
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
	generalConfig, err := g.getConfig()
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(generalConfig.ProposalFee) != 0 {
		g.eei.AddReturnMessage("invalid value provided, expected " + generalConfig.ProposalFee.String())
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

	startVoteEpoch, endVoteEpoch, err := g.startEndEpochFromArguments(args.Arguments[1], args.Arguments[2])
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	generalProposal := &GeneralProposal{
		IssuerAddress:  args.CallerAddr,
		CommitHash:     commitHash,
		StartVoteEpoch: startVoteEpoch,
		EndVoteEpoch:   endVoteEpoch,
		Yes:            big.NewInt(0),
		No:             big.NewInt(0),
		Veto:           big.NewInt(0),
		Abstain:        big.NewInt(0),
		QuorumStake:    big.NewInt(0),
		Passed:         false,
		ProposalCost:   generalConfig.ProposalFee,
		Nonce:          nextNonce,
	}
	err = g.saveGeneralProposal(commitHash, generalProposal)
	if err != nil {
		g.eei.AddReturnMessage("saveGeneralProposal " + err.Error())
		return vmcommon.UserError
	}

	nonceAsBytes := big.NewInt(0).SetUint64(nextNonce).Bytes()
	nonceKey := append([]byte(noncePrefix), nonceAsBytes...)
	g.eei.SetStorage(nonceKey, commitHash)

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(args.Function),
		Address:    args.CallerAddr,
		Topics:     [][]byte{nonceAsBytes, commitHash, args.Arguments[1], args.Arguments[2]},
	}
	g.eei.AddLogEntry(logEntry)

	return vmcommon.Ok
}

// vote casts a vote for a validator/delegation. This function receives 2 parameters and will vote with its full delegation + validator amount
//
//	args.Arguments[0] - reference - nonce as string
//	args.Arguments[1] - vote option (yes, no, veto, abstain)
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
		g.eei.AddReturnMessage("only user can call this")
		return vmcommon.UserError
	}

	voterAddress := args.CallerAddr
	proposalToVote := args.Arguments[0]
	totalStake, totalVotingPower, err := g.computeTotalStakeAndVotingPower(voterAddress)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = g.addUserVote(
		voterAddress,
		proposalToVote,
		string(args.Arguments[1]),
		totalVotingPower,
		totalStake,
		true,
		nil)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(args.Function),
		Address:    args.CallerAddr,
		Topics:     [][]byte{proposalToVote, args.Arguments[1], totalStake.Bytes(), totalVotingPower.Bytes()},
	}
	g.eei.AddLogEntry(logEntry)

	return vmcommon.Ok
}

// delegateVote casts a vote from a validator run by WASM SC and delegates it to someone else. This function receives 4 parameters:
//
//	args.Arguments[0] - proposal reference - nonce of proposal
//	args.Arguments[1] - vote option (yes, no, veto)
//	args.Arguments[2] - delegatedTo
//	args.Arguments[3] - balance to vote
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
	voter := args.Arguments[2]
	if len(voter) != len(args.CallerAddr) {
		g.eei.AddReturnMessage("invalid delegator address")
		return vmcommon.UserError
	}

	proposalToVote := args.Arguments[0]
	userStake := big.NewInt(0).SetBytes(args.Arguments[3])

	scDelegatedVoteInfo, votePower, err := g.computeDelegatedVotePower(args.CallerAddr, proposalToVote, userStake)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = g.updateDelegatedContractInfo(args.CallerAddr, proposalToVote, scDelegatedVoteInfo, userStake, votePower)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = g.addUserVote(
		voter,
		proposalToVote,
		string(args.Arguments[1]),
		votePower,
		userStake,
		false,
		args.CallerAddr)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(args.Function),
		Address:    args.CallerAddr,
		Topics:     [][]byte{proposalToVote, args.Arguments[1], voter, userStake.Bytes(), votePower.Bytes()},
	}
	g.eei.AddLogEntry(logEntry)

	return vmcommon.Ok
}

func (g *governanceContract) computeDelegatedVotePower(
	scAddress []byte,
	reference []byte,
	balance *big.Int,
) (*DelegatedSCVoteInfo, *big.Int, error) {
	scVoteInfo, err := g.getDelegatedContractInfo(scAddress, reference)
	if err != nil {
		return nil, nil, err
	}

	totalPower := big.NewInt(0).Set(scVoteInfo.TotalPower)
	votePower := big.NewInt(0).Mul(totalPower, balance)
	votePower.Div(votePower, scVoteInfo.TotalStake)
	return scVoteInfo, votePower, nil
}

func (g *governanceContract) updateDelegatedContractInfo(
	scAddress []byte,
	reference []byte,
	scVoteInfo *DelegatedSCVoteInfo,
	balance *big.Int,
	votePower *big.Int,
) error {
	scVoteInfo.UsedPower.Add(scVoteInfo.UsedPower, votePower)
	if scVoteInfo.TotalPower.Cmp(scVoteInfo.UsedPower) < 0 {
		return vm.ErrNotEnoughVotingPower
	}

	scVoteInfo.UsedStake.Add(scVoteInfo.UsedStake, balance)
	if scVoteInfo.TotalStake.Cmp(scVoteInfo.UsedStake) < 0 {
		return vm.ErrNotEnoughVotingPower
	}

	return g.saveDelegatedContractInfo(scAddress, scVoteInfo, reference)
}

func (g *governanceContract) addUserVote(
	address []byte,
	nonceAsBytes []byte,
	vote string,
	totalVotingPower *big.Int,
	totalStake *big.Int,
	direct bool,
	scAddress []byte,
) error {
	nonce := big.NewInt(0).SetBytes(nonceAsBytes)
	err := g.updateUserVoteList(address, nonce.Uint64(), direct, scAddress)
	if err != nil {
		return err
	}

	proposal, err := g.getValidProposal(nonce)
	if err != nil {
		return err
	}

	err = g.addNewVote(vote, totalVotingPower, proposal)
	if err != nil {
		return err
	}

	proposal.QuorumStake.Add(proposal.QuorumStake, totalStake)
	return g.saveGeneralProposal(proposal.CommitHash, proposal)
}

func (g *governanceContract) updateUserVoteList(address []byte, nonce uint64, direct bool, scAddress []byte) error {
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
		if !g.enableEpochsHandler.IsFlagEnabled(common.GovernanceFixesFlag) {
			userVoteList.Delegated, err = addNewNonce(userVoteList.Delegated, nonce)
			if err != nil {
				return err
			}
		} else {
			userVoteList.DelegatedWithAddress, err = addNewNonceV2(userVoteList.DelegatedWithAddress, scAddress, nonce)
			if err != nil {
				return err
			}
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

	nonceList = append(nonceList, newNonce)
	return nonceList, nil
}

func addNewNonceV2(nonceList []*DelegatedWithAddress, newDelegatedAddress []byte, newNonce uint64) ([]*DelegatedWithAddress, error) {
	for _, delegatedStruct := range nonceList {
		if newNonce == delegatedStruct.Nonce && bytes.Equal(delegatedStruct.DelegatedAddress, newDelegatedAddress) {
			return nil, vm.ErrDoubleVote
		}
	}

	nonceList = append(nonceList, &DelegatedWithAddress{
		Nonce:            newNonce,
		DelegatedAddress: newDelegatedAddress,
	})
	return nonceList, nil
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

	nonce := big.NewInt(0).SetBytes(args.Arguments[0])
	generalProposal, err := g.getProposalFromNonce(nonce)
	if err != nil {
		g.eei.AddReturnMessage("getGeneralProposal error " + err.Error())
		return vmcommon.UserError
	}
	if generalProposal.Closed {
		g.eei.AddReturnMessage("proposal is already closed, do nothing")
		return vmcommon.UserError
	}
	if !bytes.Equal(generalProposal.IssuerAddress, args.CallerAddr) {
		g.eei.AddReturnMessage("only the issuer can close the proposal")
		return vmcommon.UserError
	}

	currentEpoch := g.eei.BlockChainHook().CurrentEpoch()
	if uint64(currentEpoch) <= generalProposal.EndVoteEpoch {
		g.eei.AddReturnMessage(fmt.Sprintf("proposal can be closed only after epoch %d", generalProposal.EndVoteEpoch))
		return vmcommon.UserError
	}

	generalProposal.Closed = true
	baseConfig, err := g.getConfig()
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	generalProposal.Passed = g.computeEndResults(generalProposal, baseConfig)
	if err != nil {
		g.eei.AddReturnMessage("computeEndResults error " + err.Error())
		return vmcommon.UserError
	}

	err = g.saveGeneralProposal(generalProposal.CommitHash, generalProposal)
	if err != nil {
		g.eei.AddReturnMessage("saveGeneralProposal error " + err.Error())
		return vmcommon.UserError
	}

	tokensToReturn := big.NewInt(0).Set(generalProposal.ProposalCost)
	if !generalProposal.Passed {
		tokensToReturn.Sub(tokensToReturn, baseConfig.LostProposalFee)
		g.addToAccumulatedFees(baseConfig.LostProposalFee)
	}

	err = g.eei.Transfer(args.CallerAddr, args.RecipientAddr, tokensToReturn, nil, 0)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(args.Function),
		Address:    args.CallerAddr,
		Topics:     [][]byte{generalProposal.CommitHash, boolToSlice(generalProposal.Passed)},
	}
	g.eei.AddLogEntry(logEntry)

	return vmcommon.Ok
}

func (g *governanceContract) getAccumulatedFees() *big.Int {
	currentData := g.eei.GetStorage([]byte(accumulatedFeeKey))
	return big.NewInt(0).SetBytes(currentData)
}

func (g *governanceContract) setAccumulatedFees(value *big.Int) {
	g.eei.SetStorage([]byte(accumulatedFeeKey), value.Bytes())
}

func (g *governanceContract) addToAccumulatedFees(value *big.Int) {
	currentValue := g.getAccumulatedFees()
	currentValue.Add(currentValue, value)
	g.setAccumulatedFees(currentValue)
}

func (g *governanceContract) claimAccumulatedFees(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		g.eei.AddReturnMessage("callValue expected to be 0")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 0 {
		g.eei.AddReturnMessage("invalid number of arguments, expected 0")
		return vmcommon.UserError
	}
	if !bytes.Equal(args.CallerAddr, g.ownerAddress) {
		g.eei.AddReturnMessage("can be called only by owner")
		return vmcommon.UserError
	}
	err := g.eei.UseGas(g.gasCost.MetaChainSystemSCsCost.CloseProposal)
	if err != nil {
		g.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}

	accumulatedFees := g.getAccumulatedFees()
	g.setAccumulatedFees(big.NewInt(0))

	err = g.eei.Transfer(args.CallerAddr, args.RecipientAddr, accumulatedFees, nil, 0)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

// viewVotingPower returns the total voting power
func (g *governanceContract) viewVotingPower(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := g.checkViewFuncArguments(args, 1)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	validatorAddress := args.Arguments[0]
	if len(validatorAddress) != len(args.CallerAddr) {
		g.eei.AddReturnMessage("invalid address")
		return vmcommon.UserError
	}

	_, votingPower, err := g.computeTotalStakeAndVotingPower(validatorAddress)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	g.eei.Finish(votingPower.Bytes())

	return vmcommon.Ok
}

func (g *governanceContract) viewConfig(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := g.checkViewFuncArguments(args, 0)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	gConfig, err := g.getConfig()
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	g.eei.Finish([]byte(gConfig.ProposalFee.String()))
	g.eei.Finish([]byte(big.NewFloat(float64(gConfig.MinQuorum)).String()))
	g.eei.Finish([]byte(big.NewFloat(float64(gConfig.MinPassThreshold)).String()))
	g.eei.Finish([]byte(big.NewFloat(float64(gConfig.MinVetoThreshold)).String()))
	g.eei.Finish([]byte(big.NewInt(int64(gConfig.LastProposalNonce)).String()))

	return vmcommon.Ok
}

func (g *governanceContract) viewUserVoteHistory(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := g.checkViewFuncArguments(args, 1)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	userVotes, err := g.getUserVotes(args.Arguments[0])
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	g.finishWithIntValue(len(userVotes.Delegated)) // first we send the number of delegated nonces and afterward the nonces
	for _, val := range userVotes.Delegated {
		g.finishWithIntValue(int(val))
	}

	g.finishWithIntValue(len(userVotes.Direct)) // then we send the number of direct nonces and afterward the nonces
	for _, val := range userVotes.Direct {
		g.finishWithIntValue(int(val))
	}

	return vmcommon.Ok
}

func (g *governanceContract) finishWithIntValue(value int) {
	if value == 0 {
		g.eei.Finish([]byte{0})
		return
	}

	g.eei.Finish(big.NewInt(int64(value)).Bytes())
}

func (g *governanceContract) viewProposal(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := g.checkViewFuncArguments(args, 1)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	proposal, err := g.getProposalFromNonce(big.NewInt(0).SetBytes(args.Arguments[0]))
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	g.eei.Finish(proposal.ProposalCost.Bytes())
	g.eei.Finish(proposal.CommitHash)
	g.eei.Finish(big.NewInt(0).SetUint64(proposal.Nonce).Bytes())
	g.eei.Finish(proposal.IssuerAddress)
	g.eei.Finish(big.NewInt(0).SetUint64(proposal.StartVoteEpoch).Bytes())
	g.eei.Finish(big.NewInt(0).SetUint64(proposal.EndVoteEpoch).Bytes())
	g.eei.Finish(proposal.QuorumStake.Bytes())
	g.eei.Finish(proposal.Yes.Bytes())
	g.eei.Finish(proposal.No.Bytes())
	g.eei.Finish(proposal.Veto.Bytes())
	g.eei.Finish(proposal.Abstain.Bytes())
	g.eei.Finish(boolToSlice(proposal.Closed))
	g.eei.Finish(boolToSlice(proposal.Passed))

	return vmcommon.Ok
}

func (g *governanceContract) viewDelegatedVoteInfo(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := g.checkViewFuncArguments(args, 2)
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	scDelegatedInfo, err := g.getDelegatedContractInfo(args.Arguments[0], args.Arguments[1])
	if err != nil {
		g.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	g.eei.Finish(scDelegatedInfo.UsedStake.Bytes())
	g.eei.Finish(scDelegatedInfo.UsedPower.Bytes())
	g.eei.Finish(scDelegatedInfo.TotalStake.Bytes())
	g.eei.Finish(scDelegatedInfo.TotalPower.Bytes())

	return vmcommon.Ok
}

func (g *governanceContract) checkViewFuncArguments(
	args *vmcommon.ContractCallInput,
	numArgs int,
) error {
	if !bytes.Equal(args.CallerAddr, args.RecipientAddr) {
		return vm.ErrInvalidCaller
	}
	if args.CallValue.Cmp(zero) != 0 {
		return vm.ErrCallValueMustBeZero
	}
	if len(args.Arguments) != numArgs {
		return vm.ErrInvalidNumOfArguments
	}

	return nil
}

// addNewVote applies a new vote on a proposal then saves the new information into the storage
func (g *governanceContract) addNewVote(vote string, power *big.Int, proposal *GeneralProposal) error {
	switch vote {
	case yesString:
		proposal.Yes.Add(proposal.Yes, power)
	case noString:
		proposal.No.Add(proposal.No, power)
	case vetoString:
		proposal.Veto.Add(proposal.Veto, power)
	case abstainString:
		proposal.Abstain.Add(proposal.Abstain, power)
	default:
		return fmt.Errorf("%s: %s", vm.ErrInvalidArgument, "invalid vote type")
	}

	return nil
}

// computeVotingPower returns the voting power for a value. The value can be either a balance or
// the staked value for a validator
func (g *governanceContract) computeVotingPower(value *big.Int) (*big.Int, error) {
	minValue, err := g.getMinValueToVote()
	if err != nil {
		return nil, err
	}

	if value.Cmp(minValue) < 0 {
		return nil, vm.ErrNotEnoughStakeToVote
	}

	return big.NewInt(0).Set(value), nil // linear computation
}

// function iterates over all delegation contracts and verifies balances of the given account and makes a sum of it
func (g *governanceContract) computeTotalStakeAndVotingPower(address []byte) (*big.Int, *big.Int, error) {
	totalStake, err := g.getTotalStake(address)
	if err != nil {
		return nil, nil, err
	}

	dContractList, err := getDelegationContractList(g.eei, g.marshalizer, g.delegationMgrSCAddress)
	if err != nil {
		return nil, nil, err
	}

	err = g.eei.UseGas(g.gasCost.MetaChainSystemSCsCost.GetActiveFund * uint64(len(dContractList.Addresses)))
	if err != nil {
		return nil, nil, err
	}

	var activeDelegated *big.Int
	for _, contract := range dContractList.Addresses {
		activeDelegated, err = g.getActiveFundForDelegator(contract, address)
		if err != nil {
			return nil, nil, err
		}

		totalStake.Add(totalStake, activeDelegated)
	}

	votingPower, err := g.computeVotingPower(totalStake)
	if err != nil {
		return nil, nil, err
	}

	return totalStake, votingPower, nil
}

func (g *governanceContract) getTotalStakeInSystem() *big.Int {
	return g.eei.GetBalance(g.validatorSCAddress)
}

// computeEndResults computes if a proposal has passed or not based on votes accumulated
func (g *governanceContract) computeEndResults(proposal *GeneralProposal, baseConfig *GovernanceConfigV2) bool {
	totalVotes := big.NewInt(0).Add(proposal.Yes, proposal.No)
	totalVotes.Add(totalVotes, proposal.Veto)
	totalVotes.Add(totalVotes, proposal.Abstain)

	totalStake := g.getTotalStakeInSystem()
	minQuorumOutOfStake := core.GetIntTrimmedPercentageOfValue(totalStake, float64(baseConfig.MinQuorum))

	if totalVotes.Cmp(minQuorumOutOfStake) == -1 {
		g.eei.Finish([]byte("Proposal did not reach minQuorum"))
		return false
	}

	minVetoOfTotalVotes := core.GetIntTrimmedPercentageOfValue(totalVotes, float64(baseConfig.MinVetoThreshold))
	if proposal.Veto.Cmp(minVetoOfTotalVotes) >= 0 {
		g.eei.Finish([]byte("Proposal vetoed"))
		return false
	}

	minPassOfTotalVotes := core.GetIntTrimmedPercentageOfValue(totalVotes, float64(baseConfig.MinPassThreshold))
	if proposal.Yes.Cmp(minPassOfTotalVotes) >= 0 && proposal.Yes.Cmp(proposal.No) > 0 {
		g.eei.Finish([]byte("Proposal passed"))
		return true
	}

	g.eei.Finish([]byte("Proposal rejected"))
	return false
}

func (g *governanceContract) getActiveFundForDelegator(delegationAddress []byte, address []byte) (*big.Int, error) {
	marshaledData := g.eei.GetStorageFromAddress(delegationAddress, address)
	if len(marshaledData) == 0 {
		return big.NewInt(0), nil
	}

	dData := &DelegatorData{}
	err := g.marshalizer.Unmarshal(dData, marshaledData)
	if err != nil {
		return nil, err
	}

	if len(dData.ActiveFund) == 0 {
		return big.NewInt(0), nil
	}

	marshaledData = g.eei.GetStorageFromAddress(delegationAddress, dData.ActiveFund)
	activeFund := &Fund{}
	err = g.marshalizer.Unmarshal(activeFund, marshaledData)
	if err != nil {
		return nil, err
	}

	if activeFund.Value == nil {
		activeFund.Value = big.NewInt(0)
	}

	return activeFund.Value, nil
}

func (g *governanceContract) getTotalStake(validatorAddress []byte) (*big.Int, error) {
	marshaledData := g.eei.GetStorageFromAddress(g.validatorSCAddress, validatorAddress)
	if len(marshaledData) == 0 {
		return big.NewInt(0), nil
	}

	validatorData := &ValidatorDataV2{}
	err := g.marshalizer.Unmarshal(validatorData, marshaledData)
	if err != nil {
		return nil, err
	}

	if validatorData.TotalStakeValue == nil {
		validatorData.TotalStakeValue = big.NewInt(0)
	}

	return validatorData.TotalStakeValue, nil
}

func (g *governanceContract) saveUserVotes(address []byte, votedList *OngoingVotedListV2) error {
	if !g.enableEpochsHandler.IsFlagEnabled(common.GovernanceFixesFlag) {
		return g.saveUserVotesV1(address, votedList)
	}

	marshaledData, err := g.marshalizer.Marshal(votedList)
	if err != nil {
		return err
	}
	g.eei.SetStorage(address, marshaledData)

	return nil
}

func (g *governanceContract) saveUserVotesV1(address []byte, votedList *OngoingVotedListV2) error {
	votedListV1 := &OngoingVotedList{
		Direct:    votedList.Direct,
		Delegated: votedList.Delegated,
	}

	marshaledData, err := g.marshalizer.Marshal(votedListV1)
	if err != nil {
		return err
	}
	g.eei.SetStorage(address, marshaledData)

	return nil
}

func (g *governanceContract) getUserVotes(address []byte) (*OngoingVotedListV2, error) {
	onGoingList := &OngoingVotedListV2{
		Direct:               make([]uint64, 0),
		Delegated:            make([]uint64, 0),
		DelegatedWithAddress: make([]*DelegatedWithAddress, 0),
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
		TotalStake: big.NewInt(0),
		UsedStake:  big.NewInt(0),
	}

	marshalledData := g.eei.GetStorage(append(scAddress, reference...))
	if len(marshalledData) > 0 {
		err := g.marshalizer.Unmarshal(scVoteInfo, marshalledData)
		if err != nil {
			return nil, err
		}

		return scVoteInfo, nil
	}

	totalStake, totalVotingPower, err := g.computeTotalStakeAndVotingPower(scAddress)
	if err != nil {
		return nil, err
	}

	scVoteInfo.TotalPower.Set(totalVotingPower)
	scVoteInfo.TotalStake.Set(totalStake)

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

func (g *governanceContract) saveConfig(cfg *GovernanceConfigV2) error {
	marshaledData, err := g.marshalizer.Marshal(cfg)
	if err != nil {
		return err
	}

	g.eei.SetStorage([]byte(governanceConfigKey), marshaledData)
	return nil
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

// getValidProposal returns a proposal from storage if it exists, or it is still valid/in-progress
func (g *governanceContract) getValidProposal(nonce *big.Int) (*GeneralProposal, error) {
	proposal, err := g.getProposalFromNonce(nonce)
	if err != nil {
		return nil, err
	}

	currentEpoch := uint64(g.eei.BlockChainHook().CurrentEpoch())
	if currentEpoch < proposal.StartVoteEpoch {
		return nil, vm.ErrVotingNotStartedForProposal
	}

	if currentEpoch > proposal.EndVoteEpoch {
		return nil, vm.ErrVotedForAnExpiredProposal
	}

	return proposal, nil
}

func (g *governanceContract) getProposalFromNonce(nonce *big.Int) (*GeneralProposal, error) {
	nonceKey := append([]byte(noncePrefix), nonce.Bytes()...)
	commitHash := g.eei.GetStorage(nonceKey)
	return g.getGeneralProposal(commitHash)
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

// startEndEpochFromArguments converts the nonce string arguments to uint64
func (g *governanceContract) startEndEpochFromArguments(argStart []byte, argEnd []byte) (uint64, uint64, error) {
	startVoteEpoch := big.NewInt(0).SetBytes(argStart)
	endVoteEpoch := big.NewInt(0).SetBytes(argEnd)

	currentEpoch := uint64(g.eei.BlockChainHook().CurrentEpoch())
	if currentEpoch > startVoteEpoch.Uint64() || startVoteEpoch.Uint64() > endVoteEpoch.Uint64() {
		return 0, 0, vm.ErrInvalidStartEndVoteEpoch
	}
	if endVoteEpoch.Uint64()-startVoteEpoch.Uint64() >= uint64(g.unBondPeriodInEpochs) {
		return 0, 0, vm.ErrInvalidStartEndVoteEpoch
	}

	return startVoteEpoch.Uint64(), endVoteEpoch.Uint64(), nil
}

// convertV2Config converts the passed config file to the correct V2 typed GovernanceConfig
func (g *governanceContract) convertV2Config(config config.GovernanceSystemSCConfig) (*GovernanceConfigV2, error) {
	if config.Active.MinQuorum <= 0.01 {
		return nil, vm.ErrIncorrectConfig
	}
	if config.Active.MinPassThreshold <= 0.01 {
		return nil, vm.ErrIncorrectConfig
	}
	if config.Active.MinVetoThreshold <= 0.01 {
		return nil, vm.ErrIncorrectConfig
	}
	proposalFee, success := big.NewInt(0).SetString(config.Active.ProposalCost, conversionBase)
	if !success {
		return nil, vm.ErrIncorrectConfig
	}

	lostProposalFee, success := big.NewInt(0).SetString(config.Active.LostProposalFee, conversionBase)
	if !success {
		return nil, vm.ErrIncorrectConfig
	}

	if proposalFee.Cmp(lostProposalFee) < 0 {
		return nil, fmt.Errorf("%w proposal fee is smaller than lost proposal fee ", vm.ErrIncorrectConfig)
	}

	return &GovernanceConfigV2{
		MinQuorum:        float32(config.Active.MinQuorum),
		MinPassThreshold: float32(config.Active.MinPassThreshold),
		MinVetoThreshold: float32(config.Active.MinVetoThreshold),
		ProposalFee:      proposalFee,
		LostProposalFee:  lostProposalFee,
	}, nil
}

func convertDecimalToPercentage(arg []byte) (float32, error) {
	value, okConvert := big.NewInt(0).SetString(string(arg), conversionBase)
	if !okConvert {
		return 0.0, vm.ErrIncorrectConfig
	}

	valAsFloat := float64(value.Uint64()) / maxPercentage
	if valAsFloat < 0.001 || valAsFloat > 1.0 {
		return 0.0, vm.ErrIncorrectConfig
	}
	return float32(valAsFloat), nil
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
