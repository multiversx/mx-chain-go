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
	"github.com/ElrondNetwork/elrond-go/vm/factory"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const governanceConfigKey = "governanceConfig"

// ArgsNewGovernanceContract defines the arguments needed for the on-chain governance contract
type ArgsNewGovernanceContract struct {
	Eei              vm.SystemEI
	GasCost          vm.GasCost
	GovernanceConfig config.GovernanceSystemSCConfig
	ESDTSCAddress    []byte
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
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
		governanceSCAddress: factory.GovernanceSCAddress,
		stakingSCAddress:    factory.StakingSCAddress,
		auctionSCAddress:    factory.AuctionSCAddress,
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
	if len(args.Arguments) != 1 {
		return vmcommon.FunctionWrongSignature
	}

	return vmcommon.Ok
}

func (g *governanceContract) isWhiteListed(address []byte) bool {
	return false
}

func (g *governanceContract) whiteListAtGenesis(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (g *governanceContract) hardForkProposal(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (g *governanceContract) proposal(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (g *governanceContract) vote(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (g *governanceContract) delegateVotePower(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (g *governanceContract) revokeVotePower(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
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
