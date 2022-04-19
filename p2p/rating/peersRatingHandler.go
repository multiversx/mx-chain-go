package rating

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/random"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

const (
	defaultRating      = 0
	minRating          = -100
	maxRating          = 100
	increaseFactor     = 2
	decreaseFactor     = -1
	numOfTiers         = 4
	tierRatingTreshold = 50
	minNumOfPeers      = 1
)

// ArgPeersRatingHandler is the DTO used to create a new peers rating handler
type ArgPeersRatingHandler struct {
	Randomizer p2p.IntRandomizer
}

type peersRatingHandler struct {
	peersRatingMap map[core.PeerID]int32
	peersTiersMap  map[uint32]map[core.PeerID]struct{}
	randomizer     dataRetriever.IntRandomizer
	mut            sync.Mutex
}

// NewPeersRatingHandler returns a new peers rating handler
func NewPeersRatingHandler(args ArgPeersRatingHandler) (*peersRatingHandler, error) {
	if check.IfNil(args.Randomizer) {
		return nil, p2p.ErrNilRandomizer
	}

	prh := &peersRatingHandler{
		peersRatingMap: make(map[core.PeerID]int32),
		randomizer:     args.Randomizer,
	}

	prh.mut.Lock()
	prh.createTiersMap()
	prh.mut.Unlock()

	return prh, nil
}

// AddPeer adds a new peer to the maps with rating 0
// this is called when a new peer is connected, so if peer is known, its rating is reset
func (prh *peersRatingHandler) AddPeer(pid core.PeerID) {
	prh.mut.Lock()
	defer prh.mut.Unlock()

	oldRating := prh.peersRatingMap[pid]
	prh.updateRating(pid, oldRating, defaultRating)
}

// IncreaseRating increases the rating of a peer with the increase factor
func (prh *peersRatingHandler) IncreaseRating(pid core.PeerID) {
	prh.mut.Lock()
	defer prh.mut.Unlock()

	oldRating := prh.peersRatingMap[pid]
	newRating := oldRating + increaseFactor
	if newRating > maxRating {
		return
	}

	prh.updateRating(pid, oldRating, newRating)
}

// DecreaseRating decreases the rating of a peer with the decrease factor
func (prh *peersRatingHandler) DecreaseRating(pid core.PeerID) {
	prh.mut.Lock()
	defer prh.mut.Unlock()

	oldRating := prh.peersRatingMap[pid]
	newRating := oldRating + decreaseFactor
	if newRating < minRating {
		return
	}

	prh.updateRating(pid, oldRating, newRating)
}

// this method must be called under mutex protection
func (prh *peersRatingHandler) updateRating(pid core.PeerID, oldRating, newRating int32) {
	prh.peersRatingMap[pid] = newRating

	oldTier := computeRatingTier(oldRating)
	newTier := computeRatingTier(newRating)
	if newTier == oldTier {
		// if pid is not in tier, add it
		// this happens when a new peer is added
		_, isInTier := prh.peersTiersMap[newTier][pid]
		if !isInTier {
			prh.peersTiersMap[newTier][pid] = struct{}{}
		}

		return
	}

	prh.movePeerToNewTier(oldTier, newTier, pid)
}

func computeRatingTier(peerRating int32) uint32 {
	// [100,   51] -> tier 1
	// [ 50,    1] -> tier 2
	// [  0,  -49] -> tier 3
	// [-50, -100] -> tier 4

	tempPositiveRating := peerRating + 2*tierRatingTreshold
	tempTier := (tempPositiveRating - 1) / tierRatingTreshold

	return uint32(numOfTiers - tempTier)
}

// this method must be called under mutex protection
func (prh *peersRatingHandler) movePeerToNewTier(oldTier, newTier uint32, pid core.PeerID) {
	delete(prh.peersTiersMap[oldTier], pid)
	prh.peersTiersMap[newTier][pid] = struct{}{}
}

// this method must be called under mutex protection
func (prh *peersRatingHandler) createTiersMap() {
	prh.peersTiersMap = make(map[uint32]map[core.PeerID]struct{})
	for tier := uint32(numOfTiers); tier > 0; tier-- {
		prh.peersTiersMap[tier] = make(map[core.PeerID]struct{})
	}
}

// GetTopRatedPeersFromList returns a list of random peers, searching them in the order of rating tiers
func (prh *peersRatingHandler) GetTopRatedPeersFromList(peers []core.PeerID, numOfPeers int) []core.PeerID {
	prh.mut.Lock()
	defer prh.mut.Unlock()

	isListEmpty := len(peers) == 0
	if numOfPeers < minNumOfPeers || isListEmpty {
		return make([]core.PeerID, 0)
	}

	peersForExtraction := make([]core.PeerID, 0)
	for tier := uint32(numOfTiers); tier > 0; tier-- {
		peersInCurrentTier, found := prh.extractPeersForTier(tier, peers)
		if !found {
			continue
		}

		peersForExtraction = append(peersForExtraction, peersInCurrentTier...)

		if len(peersForExtraction) > numOfPeers {
			return prh.extractRandomPeers(peersForExtraction, numOfPeers)
		}
	}

	return prh.extractRandomPeers(peersForExtraction, numOfPeers)
}

// this method must be called under mutex protection
func (prh *peersRatingHandler) extractPeersForTier(tier uint32, peers []core.PeerID) ([]core.PeerID, bool) {
	peersInTier := make([]core.PeerID, 0)
	knownPeersInTier, found := prh.peersTiersMap[tier]
	isListEmpty := len(knownPeersInTier) == 0
	if !found || isListEmpty {
		return peersInTier, false
	}

	for _, peer := range peers {
		_, found = knownPeersInTier[peer]
		if found {
			peersInTier = append(peersInTier, peer)
		}
	}

	return peersInTier, true
}

// this method must be called under mutex protection
func (prh *peersRatingHandler) extractRandomPeers(peers []core.PeerID, numOfPeers int) []core.PeerID {
	peersLen := len(peers)
	if peersLen < numOfPeers {
		return peers
	}

	indexes := createIndexList(peersLen)
	shuffledIndexes := random.FisherYatesShuffle(indexes, prh.randomizer)

	randomPeers := make([]core.PeerID, numOfPeers)
	for i := 0; i < numOfPeers; i++ {
		randomPeers[i] = peers[shuffledIndexes[i]]
	}

	return randomPeers
}

func createIndexList(listLength int) []int {
	indexes := make([]int, listLength)
	for i := 0; i < listLength; i++ {
		indexes[i] = i
	}

	return indexes
}

// IsInterfaceNil returns true if there is no value under the interface
func (prh *peersRatingHandler) IsInterfaceNil() bool {
	return prh == nil
}
