package networkSharding

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// peerShardMapper stores the mappings between peer IDs and shard IDs
type peerShardMapper struct {
	mutPeerIdPkMap sync.RWMutex
	peerIdPkMap    map[p2p.PeerID][]byte

	mutFallbackPkShardMap sync.RWMutex
	fallbackPkShardMap    map[string]uint32

	nodesCoordinator sharding.NodesCoordinator
}

// NewPeerShardMapper creates a new peerShardMapper instance
func NewPeerShardMapper(nodesCoordinator sharding.NodesCoordinator) (*peerShardMapper, error) {
	if check.IfNil(nodesCoordinator) {
		return nil, sharding.ErrNilNodesCoordinator
	}

	return &peerShardMapper{
		peerIdPkMap:        make(map[p2p.PeerID][]byte),
		fallbackPkShardMap: make(map[string]uint32),
		nodesCoordinator:   nodesCoordinator,
	}, nil
}

// ByID returns the corresponding shard ID of a given peer ID.
// If the pid is unknown, it will return sharding.UnknownShardId
func (psm *peerShardMapper) ByID(pid p2p.PeerID) (shardId uint32) {
	psm.mutPeerIdPkMap.RLock()
	pk, ok := psm.peerIdPkMap[pid]
	psm.mutPeerIdPkMap.RUnlock()

	if !ok {
		return sharding.UnknownShardId
	}

	_, shardId, err := psm.nodesCoordinator.GetValidatorWithPublicKey(pk)
	if err == nil {
		return shardId
	}

	psm.mutFallbackPkShardMap.RLock()
	shard, ok := psm.fallbackPkShardMap[string(pk)]
	psm.mutFallbackPkShardMap.RUnlock()

	if ok {
		return shard
	}

	return sharding.UnknownShardId
}

// UpdatePeerIdPublicKey updates the peer ID - public key pair in the corresponding map
func (psm *peerShardMapper) UpdatePeerIdPublicKey(pid p2p.PeerID, pk []byte) {
	psm.mutPeerIdPkMap.Lock()
	psm.peerIdPkMap[pid] = pk
	psm.mutPeerIdPkMap.Unlock()
}

// UpdatePublicKeyShardId updates the fallback search map containing public key and shard IDs
func (psm *peerShardMapper) UpdatePublicKeyShardId(pk []byte, shardId uint32) {
	psm.mutFallbackPkShardMap.Lock()
	psm.fallbackPkShardMap[string(pk)] = shardId
	psm.mutFallbackPkShardMap.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (psm *peerShardMapper) IsInterfaceNil() bool {
	return psm == nil
}
