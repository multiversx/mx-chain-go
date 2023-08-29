package sharding

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

// SovereignNodesSetup holds the sovereign nodes setup
type SovereignNodesSetup struct {
	*NodesSetup
}

// SovereignNodesSetupArgs is a struct placeholder for sovereign nodes setup args
type SovereignNodesSetupArgs struct {
	NodesFilePath            string
	AddressPubKeyConverter   core.PubkeyConverter
	ValidatorPubKeyConverter core.PubkeyConverter
}

// NewSovereignNodesSetup  creates a new decoded sovereign nodes structure from json config file
func NewSovereignNodesSetup(args *SovereignNodesSetupArgs) (*SovereignNodesSetup, error) {
	if check.IfNil(args.AddressPubKeyConverter) {
		return nil, fmt.Errorf("%w for addressPubKeyConverter", ErrNilPubkeyConverter)
	}
	if check.IfNil(args.ValidatorPubKeyConverter) {
		return nil, fmt.Errorf("%w for validatorPubKeyConverter", ErrNilPubkeyConverter)
	}

	nodes := &NodesSetup{
		addressPubkeyConverter:   args.AddressPubKeyConverter,
		validatorPubkeyConverter: args.ValidatorPubKeyConverter,
	}

	err := core.LoadJsonFile(nodes, args.NodesFilePath)
	if err != nil {
		return nil, err
	}

	sovereignNodes := &SovereignNodesSetup{
		NodesSetup: nodes,
	}

	err = sovereignNodes.processSovereignConfig()
	if err != nil {
		return nil, err
	}

	sovereignNodes.processSovereignShardAssignment()
	sovereignNodes.createSovereignInitialNodesInfo()

	sovereignNodes.nrOfMetaChainNodes = 0
	return sovereignNodes, nil
}

func (ns *SovereignNodesSetup) processSovereignConfig() error {
	var err error

	ns.nrOfNodes = 0
	ns.nrOfMetaChainNodes = 0
	ns.nrOfShards = 1
	err = ns.processInitialNodes()
	if err != nil {
		return err
	}

	if ns.ConsensusGroupSize < 1 {
		return ErrNegativeOrZeroConsensusGroupSize
	}
	if ns.MinNodesPerShard < ns.ConsensusGroupSize {
		return ErrMinNodesPerShardSmallerThanConsensusSize
	}
	if ns.nrOfNodes < ns.MinNodesPerShard {
		return ErrNodesSizeSmallerThanMinNoOfNodes
	}

	if ns.MetaChainMinNodes != 0 || ns.MetaChainConsensusGroupSize != 0 {
		return fmt.Errorf("%w, min nodes and consensus size should be set to", errSovereignInvalidMetaConsensusSize)
	}

	return nil
}

func (ns *SovereignNodesSetup) processSovereignShardAssignment() {
	for id := uint32(0); id < ns.nrOfNodes; id++ {
		// consider only nodes with valid public key
		if ns.InitialNodes[id].pubKey != nil {
			ns.InitialNodes[id].assignedShard = core.SovereignChainShardId
			ns.InitialNodes[id].eligible = true
		}
	}
}

func (ns *SovereignNodesSetup) createSovereignInitialNodesInfo() {
	ns.eligible = make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, ns.nrOfShards)
	ns.waiting = make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, 0)
	for _, in := range ns.InitialNodes {
		if in.pubKey != nil && in.address != nil {
			ni := &nodeInfo{
				assignedShard: in.assignedShard,
				eligible:      in.eligible,
				pubKey:        in.pubKey,
				address:       in.address,
				initialRating: in.initialRating,
			}
			ns.eligible[in.assignedShard] = append(ns.eligible[in.assignedShard], ni)
		}
	}
}
