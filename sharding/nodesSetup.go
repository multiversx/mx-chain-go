package sharding

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
)

var _ GenesisNodesSetupHandler = (*NodesSetup)(nil)
var _ GenesisNodeInfoHandler = (*NodeInfo)(nil)

const defaultInitialRating = uint32(5000001)

// InitialNode holds data from json
type InitialNode struct {
	PubKey        string `json:"pubkey"`
	Address       string `json:"address"`
	InitialRating uint32 `json:"initialRating"`
	NodeInfo
}

// NodeInfo holds node info
type NodeInfo struct {
	assignedShard uint32
	eligible      bool
	pubKey        []byte
	address       []byte
	initialRating uint32
}

// AssignedShard gets the node assigned shard
func (ni *NodeInfo) AssignedShard() uint32 {
	return ni.assignedShard
}

// AddressBytes gets the node address as bytes
func (ni *NodeInfo) AddressBytes() []byte {
	return ni.address
}

// PubKeyBytes gets the node public key as bytes
func (ni *NodeInfo) PubKeyBytes() []byte {
	return ni.pubKey
}

// GetInitialRating gets the initial rating for a node
func (ni *NodeInfo) GetInitialRating() uint32 {
	return ni.initialRating
}

// IsInterfaceNil returns true if underlying object is nil
func (ni *NodeInfo) IsInterfaceNil() bool {
	return ni == nil
}

// NodesSetup hold data for decoded data from json file
type NodesSetup struct {
	StartTime             int64  `json:"startTime"`
	RoundDuration         uint64 `json:"roundDuration"`
	ConsensusGroupSize    uint32 `json:"consensusGroupSize"`
	MinNodesPerShard      uint32 `json:"minNodesPerShard"`
	ChainID               string `json:"chainID"`
	MinTransactionVersion uint32 `json:"minTransactionVersion"`

	MetaChainConsensusGroupSize uint32  `json:"metaChainConsensusGroupSize"`
	MetaChainMinNodes           uint32  `json:"metaChainMinNodes"`
	Hysteresis                  float32 `json:"hysteresis"`
	Adaptivity                  bool    `json:"adaptivity"`

	InitialNodes []*InitialNode `json:"initialNodes"`

	genesisMaxNumShards      uint32
	nrOfShards               uint32
	nrOfNodes                uint32
	nrOfMetaChainNodes       uint32
	eligible                 map[uint32][]GenesisNodeInfoHandler
	waiting                  map[uint32][]GenesisNodeInfoHandler
	validatorPubkeyConverter core.PubkeyConverter
	addressPubkeyConverter   core.PubkeyConverter
}

// NewNodesSetup creates a new decoded nodes structure from json config file
func NewNodesSetup(
	nodesFilePath string,
	addressPubkeyConverter core.PubkeyConverter,
	validatorPubkeyConverter core.PubkeyConverter,
	genesisMaxNumShards uint32,
) (*NodesSetup, error) {

	if check.IfNil(addressPubkeyConverter) {
		return nil, fmt.Errorf("%w for addressPubkeyConverter", ErrNilPubkeyConverter)
	}
	if check.IfNil(validatorPubkeyConverter) {
		return nil, fmt.Errorf("%w for validatorPubkeyConverter", ErrNilPubkeyConverter)
	}
	if genesisMaxNumShards < 1 {
		return nil, fmt.Errorf("%w for genesisMaxNumShards", ErrInvalidMaximumNumberOfShards)
	}

	nodes := &NodesSetup{
		addressPubkeyConverter:   addressPubkeyConverter,
		validatorPubkeyConverter: validatorPubkeyConverter,
		genesisMaxNumShards:      genesisMaxNumShards,
	}

	err := core.LoadJsonFile(nodes, nodesFilePath)
	if err != nil {
		return nil, err
	}

	err = nodes.processConfig()
	if err != nil {
		return nil, err
	}

	nodes.processMetaChainAssigment()
	nodes.processShardAssignment()
	nodes.createInitialNodesInfo()

	//TODO: delete this log before merging:
	log.Debug("nodes setup",
		"start time", nodes.StartTime,
		"chain id", nodes.ChainID,
		"round duration", nodes.RoundDuration,
		"min tx version", nodes.MinTransactionVersion)
	return nodes, nil
}

func (ns *NodesSetup) processConfig() error {
	var err error

	ns.nrOfNodes = 0
	ns.nrOfMetaChainNodes = 0
	for i := 0; i < len(ns.InitialNodes); i++ {
		pubKey := ns.InitialNodes[i].PubKey
		ns.InitialNodes[i].pubKey, err = ns.validatorPubkeyConverter.Decode(pubKey)
		if err != nil {
			return fmt.Errorf("%w, %s for string %s", ErrCouldNotParsePubKey, err.Error(), pubKey)
		}

		address := ns.InitialNodes[i].Address
		ns.InitialNodes[i].address, err = ns.addressPubkeyConverter.Decode(address)
		if err != nil {
			return fmt.Errorf("%w, %s for string %s", ErrCouldNotParseAddress, err.Error(), address)
		}

		// decoder treats empty string as correct, it is not allowed to have empty string as public key
		if ns.InitialNodes[i].PubKey == "" {
			ns.InitialNodes[i].pubKey = nil
			return ErrCouldNotParsePubKey
		}

		// decoder treats empty string as correct, it is not allowed to have empty string as address
		if ns.InitialNodes[i].Address == "" {
			ns.InitialNodes[i].address = nil
			return ErrCouldNotParseAddress
		}

		initialRating := ns.InitialNodes[i].InitialRating
		if initialRating == uint32(0) {
			initialRating = defaultInitialRating
		}
		ns.InitialNodes[i].initialRating = initialRating

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
			ns.InitialNodes[id].assignedShard = core.MetachainShardId
			ns.InitialNodes[id].eligible = true
			ns.nrOfMetaChainNodes++
		}
	}

	hystMeta := uint32(float32(ns.MetaChainMinNodes) * ns.Hysteresis)
	hystShard := uint32(float32(ns.MinNodesPerShard) * ns.Hysteresis)

	ns.nrOfShards = (ns.nrOfNodes - ns.nrOfMetaChainNodes - hystMeta) / (ns.MinNodesPerShard + hystShard)

	if ns.nrOfShards > ns.genesisMaxNumShards {
		ns.nrOfShards = ns.genesisMaxNumShards
	}
}

func (ns *NodesSetup) processShardAssignment() {
	// initial implementation - as there is no other info than public key, we allocate first nodes in FIFO order to shards
	currentShard := uint32(0)
	countSetNodes := ns.nrOfMetaChainNodes
	for ; currentShard < ns.nrOfShards; currentShard++ {
		for id := countSetNodes; id < ns.nrOfMetaChainNodes+(currentShard+1)*ns.MinNodesPerShard; id++ {
			// consider only nodes with valid public key
			if ns.InitialNodes[id].pubKey != nil {
				ns.InitialNodes[id].assignedShard = currentShard
				ns.InitialNodes[id].eligible = true
				countSetNodes++
			}
		}
	}

	// allocate the rest to waiting lists
	currentShard = 0
	for i := countSetNodes; i < ns.nrOfNodes; i++ {
		currentShard = (currentShard + 1) % (ns.nrOfShards + 1)
		if currentShard == ns.nrOfShards {
			currentShard = core.MetachainShardId
		}

		if ns.InitialNodes[i].pubKey != nil {
			ns.InitialNodes[i].assignedShard = currentShard
			ns.InitialNodes[i].eligible = false
		}
	}
}

func (ns *NodesSetup) createInitialNodesInfo() {
	nrOfShardAndMeta := ns.nrOfShards + 1

	ns.eligible = make(map[uint32][]GenesisNodeInfoHandler, nrOfShardAndMeta)
	ns.waiting = make(map[uint32][]GenesisNodeInfoHandler, nrOfShardAndMeta)
	for _, in := range ns.InitialNodes {
		if in.pubKey != nil && in.address != nil {
			nodeInfo := &NodeInfo{
				assignedShard: in.assignedShard,
				eligible:      in.eligible,
				pubKey:        in.pubKey,
				address:       in.address,
				initialRating: in.initialRating,
			}
			if in.eligible {
				ns.eligible[in.assignedShard] = append(ns.eligible[in.assignedShard], nodeInfo)
			} else {
				ns.waiting[in.assignedShard] = append(ns.waiting[in.assignedShard], nodeInfo)
			}
		}
	}
}

// InitialNodesPubKeys - gets initial nodes public keys
func (ns *NodesSetup) InitialNodesPubKeys() map[uint32][]string {
	allNodesPubKeys := make(map[uint32][]string)
	for shardId, nodesInfo := range ns.eligible {
		pubKeys := make([]string, len(nodesInfo))
		for i := 0; i < len(nodesInfo); i++ {
			pubKeys[i] = string(nodesInfo[i].PubKeyBytes())
		}

		allNodesPubKeys[shardId] = pubKeys
	}

	return allNodesPubKeys
}

// InitialNodesInfo - gets initial nodes info
func (ns *NodesSetup) InitialNodesInfo() (map[uint32][]GenesisNodeInfoHandler, map[uint32][]GenesisNodeInfoHandler) {
	return ns.eligible, ns.waiting
}

// AllInitialNodes returns all initial nodes loaded
func (ns *NodesSetup) AllInitialNodes() []GenesisNodeInfoHandler {
	list := make([]GenesisNodeInfoHandler, len(ns.InitialNodes))
	for idx, initialNode := range ns.InitialNodes {
		list[idx] = initialNode
	}

	return list
}

// InitialEligibleNodesPubKeysForShard - gets initial nodes public keys for shard
func (ns *NodesSetup) InitialEligibleNodesPubKeysForShard(shardId uint32) ([]string, error) {
	if ns.eligible[shardId] == nil {
		return nil, ErrShardIdOutOfRange
	}
	if len(ns.eligible[shardId]) == 0 {
		return nil, ErrNoPubKeys
	}

	nodesInfo := ns.eligible[shardId]
	pubKeys := make([]string, len(nodesInfo))
	for i := 0; i < len(nodesInfo); i++ {
		pubKeys[i] = string(nodesInfo[i].PubKeyBytes())
	}

	return pubKeys, nil
}

// InitialNodesInfoForShard - gets initial nodes info for shard
func (ns *NodesSetup) InitialNodesInfoForShard(shardId uint32) ([]GenesisNodeInfoHandler, []GenesisNodeInfoHandler, error) {
	if ns.eligible[shardId] == nil {
		return nil, nil, ErrShardIdOutOfRange
	}
	if len(ns.eligible[shardId]) == 0 {
		return nil, nil, ErrNoPubKeys
	}

	return ns.eligible[shardId], ns.waiting[shardId], nil
}

// NumberOfShards returns the calculated number of shards
func (ns *NodesSetup) NumberOfShards() uint32 {
	return ns.nrOfShards
}

// MinNumberOfNodes returns the minimum number of nodes
func (ns *NodesSetup) MinNumberOfNodes() uint32 {
	return ns.nrOfShards*ns.MinNodesPerShard + ns.MetaChainMinNodes
}

// MinShardHysteresisNodes returns the minimum number of hysteresis nodes per shard
func (ns *NodesSetup) MinShardHysteresisNodes() uint32 {
	return uint32(float32(ns.MinNodesPerShard) * ns.Hysteresis)
}

// MinMetaHysteresisNodes returns the minimum number of hysteresis nodes in metachain
func (ns *NodesSetup) MinMetaHysteresisNodes() uint32 {
	return uint32(float32(ns.MetaChainMinNodes) * ns.Hysteresis)
}

// MinNumberOfNodesWithHysteresis returns the minimum number of nodes with hysteresis
func (ns *NodesSetup) MinNumberOfNodesWithHysteresis() uint32 {
	hystNodesMeta := ns.MinMetaHysteresisNodes()
	hystNodesShard := ns.MinShardHysteresisNodes()
	minNumberOfNodes := ns.MinNumberOfNodes()

	return minNumberOfNodes + hystNodesMeta + ns.nrOfShards*hystNodesShard
}

// MinNumberOfShardNodes returns the minimum number of nodes per shard
func (ns *NodesSetup) MinNumberOfShardNodes() uint32 {
	return ns.MinNodesPerShard
}

// MinNumberOfMetaNodes returns the minimum number of nodes in metachain
func (ns *NodesSetup) MinNumberOfMetaNodes() uint32 {
	return ns.MetaChainMinNodes
}

// GetHysteresis returns the hysteresis value
func (ns *NodesSetup) GetHysteresis() float32 {
	return ns.Hysteresis
}

// GetAdaptivity returns the value of the adaptivity boolean flag
func (ns *NodesSetup) GetAdaptivity() bool {
	return ns.Adaptivity
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

// GetStartTime returns the start time
func (ns *NodesSetup) GetStartTime() int64 {
	return ns.StartTime
}

// GetRoundDuration returns the round duration
func (ns *NodesSetup) GetRoundDuration() uint64 {
	return ns.RoundDuration
}

// GetChainId returns the chain ID
func (ns *NodesSetup) GetChainId() string {
	return ns.ChainID
}

// GetMinTransactionVersion returns the minimum transaction version
func (ns *NodesSetup) GetMinTransactionVersion() uint32 {
	return ns.MinTransactionVersion
}

// GetShardConsensusGroupSize returns the shard consensus group size
func (ns *NodesSetup) GetShardConsensusGroupSize() uint32 {
	return ns.ConsensusGroupSize
}

// GetMetaConsensusGroupSize returns the metachain consensus group size
func (ns *NodesSetup) GetMetaConsensusGroupSize() uint32 {
	return ns.MetaChainConsensusGroupSize
}

// IsInterfaceNil returns true if underlying object is nil
func (ns *NodesSetup) IsInterfaceNil() bool {
	return ns == nil
}
