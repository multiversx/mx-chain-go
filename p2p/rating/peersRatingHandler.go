package rating

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/random"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const (
	topRatedTier   = "top rated tier"
	badRatedTier   = "bad rated tier"
	defaultRating  = 0
	minRating      = -100
	maxRating      = 100
	increaseFactor = 2
	decreaseFactor = -1
	minNumOfPeers  = 1
	int32Size      = 4
)

// ArgPeersRatingHandler is the DTO used to create a new peers rating handler
type ArgPeersRatingHandler struct {
	TopRatedCache storage.Cacher
	BadRatedCache storage.Cacher
	Randomizer    p2p.IntRandomizer
}

type peersRatingHandler struct {
	topRatedCache storage.Cacher
	badRatedCache storage.Cacher
	randomizer    dataRetriever.IntRandomizer
	mut           sync.Mutex
}

// NewPeersRatingHandler returns a new peers rating handler
func NewPeersRatingHandler(args ArgPeersRatingHandler) (*peersRatingHandler, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	prh := &peersRatingHandler{
		topRatedCache: args.TopRatedCache,
		badRatedCache: args.BadRatedCache,
		randomizer:    args.Randomizer,
	}

	return prh, nil
}

func checkArgs(args ArgPeersRatingHandler) error {
	if check.IfNil(args.Randomizer) {
		return p2p.ErrNilRandomizer
	}
	if check.IfNil(args.TopRatedCache) {
		return fmt.Errorf("%w for TopRatedCache", p2p.ErrNilCacher)
	}
	if check.IfNil(args.BadRatedCache) {
		return fmt.Errorf("%w for BadRatedCache", p2p.ErrNilCacher)
	}

	return nil
}

// AddPeer adds a new peer to the cache with rating 0
// this is called when a new peer is detected
func (prh *peersRatingHandler) AddPeer(pid core.PeerID) {
	prh.mut.Lock()
	defer prh.mut.Unlock()

	_, found := prh.getOldRating(pid)
	if found {
		return
	}

	prh.topRatedCache.Put(pid.Bytes(), defaultRating, int32Size)
}

// IncreaseRating increases the rating of a peer with the increase factor
func (prh *peersRatingHandler) IncreaseRating(pid core.PeerID) {
	prh.mut.Lock()
	defer prh.mut.Unlock()

	oldRating, found := prh.getOldRating(pid)
	if !found {
		// new pid, add it with default rating
		prh.topRatedCache.Put(pid.Bytes(), defaultRating, int32Size)
		return
	}

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

	oldRating, found := prh.getOldRating(pid)
	if !found {
		// new pid, add it with default rating
		prh.topRatedCache.Put(pid.Bytes(), defaultRating, int32Size)
		return
	}

	newRating := oldRating + decreaseFactor
	if newRating < minRating {
		return
	}

	prh.updateRating(pid, oldRating, newRating)
}

func (prh *peersRatingHandler) getOldRating(pid core.PeerID) (int32, bool) {
	oldRating, found := prh.topRatedCache.Get(pid.Bytes())
	if found {
		oldRatingInt, _ := oldRating.(int32)
		return oldRatingInt, found
	}

	oldRating, found = prh.badRatedCache.Get(pid.Bytes())
	if found {
		oldRatingInt, _ := oldRating.(int32)
		return oldRatingInt, found
	}

	return defaultRating, found
}

func (prh *peersRatingHandler) updateRating(pid core.PeerID, oldRating, newRating int32) {
	oldTier := computeRatingTier(oldRating)
	newTier := computeRatingTier(newRating)
	if newTier == oldTier {
		if newTier == topRatedTier {
			prh.topRatedCache.Put(pid.Bytes(), newRating, int32Size)
		} else {
			prh.badRatedCache.Put(pid.Bytes(), newRating, int32Size)
		}

		return
	}

	prh.movePeerToNewTier(newRating, pid)
}

func computeRatingTier(peerRating int32) string {
	if peerRating >= defaultRating {
		return topRatedTier
	}

	return badRatedTier
}

func (prh *peersRatingHandler) movePeerToNewTier(newRating int32, pid core.PeerID) {
	newTier := computeRatingTier(newRating)
	if newTier == topRatedTier {
		prh.badRatedCache.Remove(pid.Bytes())
		prh.topRatedCache.Put(pid.Bytes(), newRating, int32Size)
	} else {
		prh.topRatedCache.Remove(pid.Bytes())
		prh.badRatedCache.Put(pid.Bytes(), newRating, int32Size)
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

	if prh.hasEnoughTopRated(peers, numOfPeers) {
		return prh.extractRandomPeers(prh.topRatedCache.Keys(), numOfPeers)
	}

	peersForExtraction := make([][]byte, 0)
	peersForExtraction = append(peersForExtraction, prh.topRatedCache.Keys()...)
	peersForExtraction = append(peersForExtraction, prh.badRatedCache.Keys()...)

	return prh.extractRandomPeers(peersForExtraction, numOfPeers)
}

func (prh *peersRatingHandler) hasEnoughTopRated(peers []core.PeerID, numOfPeers int) bool {
	counter := 0

	for _, peer := range peers {
		if prh.topRatedCache.Has(peer.Bytes()) {
			counter++
			if counter >= numOfPeers {
				return true
			}
		}
	}

	return false
}

func (prh *peersRatingHandler) extractRandomPeers(peersBytes [][]byte, numOfPeers int) []core.PeerID {
	peersLen := len(peersBytes)
	if peersLen <= numOfPeers {
		return peersBytesToPeerIDs(peersBytes)
	}

	indexes := createIndexList(peersLen)
	shuffledIndexes := random.FisherYatesShuffle(indexes, prh.randomizer)

	randomPeers := make([]core.PeerID, numOfPeers)
	for i := 0; i < numOfPeers; i++ {
		peerBytes := peersBytes[shuffledIndexes[i]]
		randomPeers[i] = core.PeerID(peerBytes)
	}

	return randomPeers
}

func peersBytesToPeerIDs(peersBytes [][]byte) []core.PeerID {
	peerIDs := make([]core.PeerID, len(peersBytes))
	for idx, peerBytes := range peersBytes {
		peerIDs[idx] = core.PeerID(peerBytes)
	}

	return peerIDs
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
