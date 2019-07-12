package sharding

import (
	"bytes"
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go/core"
)

// InitialNode holds data from json
type InitialNode struct {
	PubKey        string `json:"pubkey"`
	Address       string `json:"address"`
	assignedShard uint32
	nodeInfo
}

// nodeInfo holds node info
type nodeInfo struct {
	pubKey  []byte
	address []byte
}

// NodesSetup hold data for decoded data from json file
type NodesSetup struct {
	StartTime          int64  `json:"startTime"`
	RoundDuration      uint64 `json:"roundDuration"`
	ConsensusGroupSize uint32 `json:"consensusGroupSize"`
	MinNodesPerShard   uint32 `json:"minNodesPerShard"`

	MetaChainActive             bool   `json:"metaChainActive"`
	MetaChainConsensusGroupSize uint32 `json:"metaChainConsensusGroupSize"`
	MetaChainMinNodes           uint32 `json:"metaChainMinNodes"`

	InitialNodes []*InitialNode `json:"initialNodes"`

	nrOfShards         uint32
	nrOfNodes          uint32
	nrOfMetaChainNodes uint32
	allNodesInfos      map[uint32][]*nodeInfo
}

// NewNodesSetup creates a new decoded nodes structure from json config file
func NewNodesSetup(nodesFilePath string, numOfNodes uint64) (*NodesSetup, error) {
	nodes := &NodesSetup{}

	err := core.LoadJsonFile(nodes, nodesFilePath, log)
	if err != nil {
		return nil, err
	}

	if numOfNodes < uint64(len(nodes.InitialNodes)) {
		nodes.InitialNodes = nodes.InitialNodes[:numOfNodes]
	}

	err = nodes.processConfig()
	if err != nil {
		return nil, err
	}

	if nodes.MetaChainActive {
		nodes.processMetaChainAssigment()
	}

	nodes.processShardAssignment()
	nodes.createInitialNodesPubKeys()

	return nodes, nil
}

func (ns *NodesSetup) processConfig() error {
	var err error

	ns.nrOfNodes = 0
	ns.nrOfMetaChainNodes = 0
	for i := 0; i < len(ns.InitialNodes); i++ {
		ns.InitialNodes[i].pubKey, err = hex.DecodeString(ns.InitialNodes[i].PubKey)
		ns.InitialNodes[i].address, err = hex.DecodeString(ns.InitialNodes[i].Address)

		// decoder treats empty string as correct, it is not allowed to have empty string as public key
		if ns.InitialNodes[i].PubKey == "" || err != nil {
			ns.InitialNodes[i].pubKey = nil
			return ErrCouldNotParsePubKey
		}

		// decoder treats empty string as correct, it is not allowed to have empty string as address
		if ns.InitialNodes[i].Address == "" || err != nil {
			ns.InitialNodes[i].address = nil
			return ErrCouldNotParseAddress
		}

		ns.nrOfNodes++
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

	if ns.MetaChainActive {
		if ns.MetaChainConsensusGroupSize < 1 {
			return ErrNegativeOrZeroConsensusGroupSize
		}
		if ns.MetaChainMinNodes < ns.MetaChainConsensusGroupSize {
			return ErrMinNodesPerShardSmallerThanConsensusSize
		}

		totalMinNodes := ns.MetaChainMinNodes + ns.MinNodesPerShard
		if ns.nrOfNodes < totalMinNodes {
			return ErrNodesSizeSmallerThanMinNoOfNodes
		}
	}

	return nil
}

func (ns *NodesSetup) processMetaChainAssigment() {
	ns.nrOfMetaChainNodes = 0
	for id := uint32(0); id < ns.MetaChainMinNodes; id++ {
		if ns.InitialNodes[id].pubKey != nil {
			ns.InitialNodes[id].assignedShard = MetachainShardId
			ns.nrOfMetaChainNodes++
		}
	}
}

func (ns *NodesSetup) processShardAssignment() {
	// initial implementation - as there is no other info than public key, we allocate first nodes in FIFO order to shards
	ns.nrOfShards = (ns.nrOfNodes - ns.nrOfMetaChainNodes) / ns.MinNodesPerShard

	currentShard := uint32(0)
	countSetNodes := ns.nrOfMetaChainNodes
	for ; currentShard < ns.nrOfShards; currentShard++ {
		for id := countSetNodes; id < ns.nrOfMetaChainNodes+(currentShard+1)*ns.MinNodesPerShard; id++ {
			// consider only nodes with valid public key
			if ns.InitialNodes[id].pubKey != nil {
				ns.InitialNodes[id].assignedShard = currentShard
				countSetNodes++
			}
		}
	}

	// allocate the rest
	currentShard = 0
	for i := countSetNodes; i < ns.nrOfNodes; i++ {
		ns.InitialNodes[i].assignedShard = currentShard
		currentShard = (currentShard + 1) % ns.nrOfShards
	}
}

func (ns *NodesSetup) createInitialNodesPubKeys() {
	nrOfShardAndMeta := ns.nrOfShards
	if ns.MetaChainActive {
		nrOfShardAndMeta += 1
	}

	ns.allNodesInfos = make(map[uint32][]*nodeInfo, nrOfShardAndMeta)
	for _, in := range ns.InitialNodes {
		if in.pubKey != nil && in.address != nil {
			ns.allNodesInfos[in.assignedShard] = append(ns.allNodesInfos[in.assignedShard],
				&nodeInfo{in.pubKey, in.address})
		}
	}
}

// InitialNodesInfos - gets initial public keys
func (ns *NodesSetup) InitialNodesInfos() map[uint32][]*nodeInfo {
	return ns.allNodesInfos
}

// InitialNodesInfosForShard - gets initial public keys
func (ns *NodesSetup) InitialNodesInfosForShard(shardId uint32) ([]*nodeInfo, error) {
	if ns.allNodesInfos[shardId] == nil {
		return nil, ErrShardIdOutOfRange
	}
	if len(ns.allNodesInfos[shardId]) == 0 {
		return nil, ErrNoPubKeys
	}

	return ns.allNodesInfos[shardId], nil
}

// NumberOfShards returns the calculated number of shards
func (ns *NodesSetup) NumberOfShards() uint32 {
	return ns.nrOfShards
}

// IsMetaChainActive returns if MetaChain is active
func (ns *NodesSetup) IsMetaChainActive() bool {
	return ns.MetaChainActive
}

// GetShardIDForPubKey returns the allocated shard ID from public key
func (ns *NodesSetup) GetShardIDForPubKey(pubKey []byte) (uint32, error) {
	for _, in := range ns.InitialNodes {
		if in.pubKey != nil && bytes.Equal(pubKey, in.pubKey) {
			return in.assignedShard, nil
		}
	}
	return 0, ErrPublicKeyNotFoundInGenesis
}
