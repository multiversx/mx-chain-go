package sharding

import (
	"bytes"
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go/core"
)

// InitialNode holds data from json
type InitialNode struct {
	PubKey  string `json:"pubkey"`
	Address string `json:"address"`
	NodeInfo
}

// NodeInfo holds node info
type NodeInfo struct {
	assignedShard uint32
	pubKey        []byte
	address       []byte
}

// AssignedShard gets the node assigned shard
func (ni *NodeInfo) AssignedShard() uint32 {
	return ni.assignedShard
}

// Address gets the node address
func (ni *NodeInfo) Address() []byte {
	return ni.address
}

// PubKey gets the node public key
func (ni *NodeInfo) PubKey() []byte {
	return ni.pubKey
}

// NodesSetup hold data for decoded data from json file
type NodesSetup struct {
	StartTime          int64  `json:"startTime"`
	RoundDuration      uint64 `json:"roundDuration"`
	ConsensusGroupSize uint32 `json:"consensusGroupSize"`
	MinNodesPerShard   uint32 `json:"minNodesPerShard"`

	MetaChainConsensusGroupSize uint32 `json:"metaChainConsensusGroupSize"`
	MetaChainMinNodes           uint32 `json:"metaChainMinNodes"`

	InitialNodes []*InitialNode `json:"initialNodes"`

	nrOfShards         uint32
	nrOfNodes          uint32
	nrOfMetaChainNodes uint32
	allNodesInfo       map[uint32][]*NodeInfo
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

	nodes.processMetaChainAssigment()
	nodes.processShardAssignment()
	nodes.createInitialNodesInfo()

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

func (ns *NodesSetup) createInitialNodesInfo() {
	nrOfShardAndMeta := ns.nrOfShards + 1

	ns.allNodesInfo = make(map[uint32][]*NodeInfo, nrOfShardAndMeta)
	for _, in := range ns.InitialNodes {
		if in.pubKey != nil && in.address != nil {
			ns.allNodesInfo[in.assignedShard] = append(ns.allNodesInfo[in.assignedShard],
				&NodeInfo{in.assignedShard, in.pubKey, in.address})
		}
	}
}

// InitialNodesPubKeys - gets initial nodes public keys
func (ns *NodesSetup) InitialNodesPubKeys() map[uint32][]string {
	allNodesPubKeys := make(map[uint32][]string, 0)
	for shardId, nodesInfo := range ns.allNodesInfo {
		pubKeys := make([]string, len(nodesInfo))
		for i := 0; i < len(nodesInfo); i++ {
			pubKeys[i] = string(nodesInfo[i].pubKey)
		}

		allNodesPubKeys[shardId] = pubKeys
	}

	return allNodesPubKeys
}

// InitialNodesInfo - gets initial nodes info
func (ns *NodesSetup) InitialNodesInfo() map[uint32][]*NodeInfo {
	return ns.allNodesInfo
}

// InitialNodesPubKeysForShard - gets initial nodes public keys for shard
func (ns *NodesSetup) InitialNodesPubKeysForShard(shardId uint32) ([]string, error) {
	if ns.allNodesInfo[shardId] == nil {
		return nil, ErrShardIdOutOfRange
	}
	if len(ns.allNodesInfo[shardId]) == 0 {
		return nil, ErrNoPubKeys
	}

	nodesInfo := ns.allNodesInfo[shardId]
	pubKeys := make([]string, len(nodesInfo))
	for i := 0; i < len(nodesInfo); i++ {
		pubKeys[i] = string(nodesInfo[i].pubKey)
	}

	return pubKeys, nil
}

// InitialNodesInfoForShard - gets initial nodes info for shard
func (ns *NodesSetup) InitialNodesInfoForShard(shardId uint32) ([]*NodeInfo, error) {
	if ns.allNodesInfo[shardId] == nil {
		return nil, ErrShardIdOutOfRange
	}
	if len(ns.allNodesInfo[shardId]) == 0 {
		return nil, ErrNoPubKeys
	}

	return ns.allNodesInfo[shardId], nil
}

// NumberOfShards returns the calculated number of shards
func (ns *NodesSetup) NumberOfShards() uint32 {
	return ns.nrOfShards
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
