package sharding

import (
	"bytes"
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
)

// InitialNode holds data from json
type InitialNode struct {
	PubKey        string `json:"pubkey"`
	assignedShard uint32
	pubKey        []byte
}

// Nodes hold data for decoded data from json file
type Nodes struct {
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
	allNodesPubKeys    map[uint32][]string
}

// NewNodesConfig creates a new decoded nodes structure from json config file
func NewNodesConfig(nodesFilePath string) (*Nodes, error) {
	nodes := &Nodes{}

	err := core.LoadJsonFile(nodes, nodesFilePath, log)
	if err != nil {
		return nil, err
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

func (n *Nodes) processConfig() error {
	var err error

	n.nrOfNodes = 0
	n.nrOfMetaChainNodes = 0
	for i := 0; i < len(n.InitialNodes); i++ {
		n.InitialNodes[i].pubKey, err = hex.DecodeString(n.InitialNodes[i].PubKey)

		// decoder treats empty string as correct, it is not allowed to have empty string as public key
		if n.InitialNodes[i].PubKey == "" || err != nil {
			n.InitialNodes[i].pubKey = nil
			return ErrCouldNotParsePubKey
		}

		n.nrOfNodes++
	}

	if n.ConsensusGroupSize < 1 {
		return ErrNegativeOrZeroConsensusGroupSize
	}
	if n.MinNodesPerShard < n.ConsensusGroupSize {
		return ErrMinNodesPerShardSmallerThanConsensusSize
	}
	if n.nrOfNodes < n.MinNodesPerShard {
		return ErrNodesSizeSmallerThanMinNoOfNodes
	}

	if n.MetaChainActive {
		if n.MetaChainConsensusGroupSize < 1 {
			return ErrNegativeOrZeroConsensusGroupSize
		}
		if n.MetaChainMinNodes < n.MetaChainConsensusGroupSize {
			return ErrMinNodesPerShardSmallerThanConsensusSize
		}

		totalMinNodes := n.MetaChainMinNodes + n.MinNodesPerShard
		if n.nrOfNodes < totalMinNodes {
			return ErrNodesSizeSmallerThanMinNoOfNodes
		}
	}

	return nil
}

func (n *Nodes) processMetaChainAssigment() {
	n.nrOfMetaChainNodes = 0
	for id := uint32(0); id < n.MetaChainMinNodes; id++ {
		if n.InitialNodes[id].pubKey != nil {
			n.InitialNodes[id].assignedShard = MetachainShardId
			n.nrOfMetaChainNodes++
		}
	}
}

func (n *Nodes) processShardAssignment() {
	// initial implementation - as there is no other info than public key, we allocate first nodes in FIFO order to shards
	n.nrOfShards = (n.nrOfNodes - n.nrOfMetaChainNodes) / n.MinNodesPerShard

	currentShard := uint32(0)
	countSetNodes := n.nrOfMetaChainNodes
	for ; currentShard < n.nrOfShards; currentShard++ {
		for id := countSetNodes; id < n.nrOfMetaChainNodes+(currentShard+1)*n.MinNodesPerShard; id++ {
			// consider only nodes with valid public key
			if n.InitialNodes[id].pubKey != nil {
				n.InitialNodes[id].assignedShard = currentShard
				countSetNodes++
			}
		}
	}

	// allocate the rest
	currentShard = 0
	for i := countSetNodes; i < n.nrOfNodes; i++ {
		n.InitialNodes[i].assignedShard = currentShard
		currentShard = (currentShard + 1) % n.nrOfShards
	}
}

func (n *Nodes) createInitialNodesPubKeys() {
	nrOfShardAndMeta := n.nrOfShards
	if n.MetaChainActive {
		nrOfShardAndMeta += 1
	}

	n.allNodesPubKeys = make(map[uint32][]string, nrOfShardAndMeta)
	for _, in := range n.InitialNodes {
		if in.pubKey != nil {
			n.allNodesPubKeys[in.assignedShard] = append(n.allNodesPubKeys[in.assignedShard], string(in.pubKey))
		}
	}
}

// InitialNodesPubKeys - gets initial public keys
func (n *Nodes) InitialNodesPubKeys() map[uint32][]string {
	return n.allNodesPubKeys
}

// InitialNodesPubKeysForShard - gets initial public keys
func (n *Nodes) InitialNodesPubKeysForShard(shardId uint32) ([]string, error) {
	if n.allNodesPubKeys[shardId] == nil {
		return nil, ErrShardIdOutOfRange
	}

	if len(n.allNodesPubKeys[shardId]) == 0 {
		return nil, ErrNoPubKeys
	}

	return n.allNodesPubKeys[shardId], nil
}

// NumberOfShards returns the calculated number of shards
func (n *Nodes) NumberOfShards() uint32 {
	return n.nrOfShards
}

// IsMetaChainActive returns if MetaChain is active
func (n *Nodes) IsMetaChainActive() bool {
	return n.MetaChainActive
}

// GetShardIDForPubKey returns the allocated shard ID from public key
func (n *Nodes) GetShardIDForPubKey(pubKey []byte) (uint32, error) {
	for _, in := range n.InitialNodes {
		if in.pubKey != nil && bytes.Equal(pubKey, in.pubKey) {
			return in.assignedShard, nil
		}
	}
	return 0, ErrNoValidPublicKey
}
