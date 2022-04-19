package rating

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/random"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

func TestNewPeersRatingHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil randomizer should error", func(t *testing.T) {
		t.Parallel()

		prh, err := NewPeersRatingHandler(ArgPeersRatingHandler{nil})
		assert.Equal(t, p2p.ErrNilRandomizer, err)
		assert.True(t, check.IfNil(prh))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		prh, err := NewPeersRatingHandler(ArgPeersRatingHandler{&random.ConcurrentSafeIntRandomizer{}})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(prh))
	})
}

func TestPeersRatingHandler_AddPeer(t *testing.T) {
	t.Parallel()

	prh, _ := NewPeersRatingHandler(ArgPeersRatingHandler{&random.ConcurrentSafeIntRandomizer{}})
	assert.False(t, check.IfNil(prh))

	providedPid := core.PeerID("provided pid")
	prh.AddPeer(providedPid)

	rating, found := prh.peersRatingMap[providedPid]
	assert.True(t, found)
	assert.Equal(t, 0, int(rating))

	peerInTier, found := prh.peersTiersMap[3] // rating 0 should be in tier 3
	assert.True(t, found)
	assert.Equal(t, 1, len(peerInTier))

	_, found = peerInTier[providedPid]
	assert.True(t, found)
}

func TestPeersRatingHandler_IncreaseRating(t *testing.T) {
	t.Parallel()

	prh, _ := NewPeersRatingHandler(ArgPeersRatingHandler{&random.ConcurrentSafeIntRandomizer{}})
	assert.False(t, check.IfNil(prh))

	providedPid := core.PeerID("provided pid")
	numOfCalls := 10
	for i := 0; i < numOfCalls; i++ {
		prh.IncreaseRating(providedPid)
	}

	rating, found := prh.peersRatingMap[providedPid]
	assert.True(t, found)
	assert.Equal(t, numOfCalls*increaseFactor, int(rating))

	// limit exceeded
	for i := 0; i < maxRating; i++ {
		prh.IncreaseRating(providedPid)
	}

	rating, found = prh.peersRatingMap[providedPid]
	assert.True(t, found)
	assert.Equal(t, maxRating, int(rating))

	// peer should be in tier 1
	peersMap, hasPeers := prh.peersTiersMap[1]
	assert.True(t, hasPeers)
	assert.Equal(t, 1, len(peersMap))
	_, found = peersMap[providedPid]
	assert.True(t, found)

	// other tiers should be empty, but providedPeer went from 3 to 1
	for i := uint32(2); i <= numOfTiers; i++ {
		peersMap, hasPeers = prh.peersTiersMap[i]
		assert.True(t, hasPeers)
		assert.Equal(t, 0, len(peersMap))
	}
}

func TestPeersRatingHandler_DecreaseRating(t *testing.T) {
	t.Parallel()

	prh, _ := NewPeersRatingHandler(ArgPeersRatingHandler{&random.ConcurrentSafeIntRandomizer{}})
	assert.False(t, check.IfNil(prh))

	providedPid := core.PeerID("provided pid")
	numOfCalls := 10
	for i := 0; i < numOfCalls; i++ {
		prh.DecreaseRating(providedPid)
	}

	rating, found := prh.peersRatingMap[providedPid]
	assert.True(t, found)
	assert.Equal(t, numOfCalls*decreaseFactor, int(rating))

	// limit exceeded
	for i := 0; i > minRating; i-- {
		prh.DecreaseRating(providedPid)
	}

	rating, found = prh.peersRatingMap[providedPid]
	assert.True(t, found)
	assert.Equal(t, minRating, int(rating))

	// peer should be in tier 4
	peersMap, hasPeers := prh.peersTiersMap[4]
	assert.True(t, hasPeers)
	assert.Equal(t, 1, len(peersMap))
	_, found = peersMap[providedPid]
	assert.True(t, found)

	// other tiers should be empty, but providedPeer went from 3 to 4
	for i := uint32(1); i < 4; i++ {
		peersMap, hasPeers = prh.peersTiersMap[i]
		assert.True(t, hasPeers)
		assert.Equal(t, 0, len(peersMap))
	}
}

func Test_computeRatingTier(t *testing.T) {
	t.Parallel()

	tier1, tier2, tier3, tier4 := uint32(1), uint32(2), uint32(3), uint32(4)
	assert.Equal(t, tier4, computeRatingTier(-100))
	assert.Equal(t, tier4, computeRatingTier(-75))
	assert.Equal(t, tier4, computeRatingTier(-50))
	assert.Equal(t, tier3, computeRatingTier(-49))
	assert.Equal(t, tier3, computeRatingTier(-25))
	assert.Equal(t, tier3, computeRatingTier(0))
	assert.Equal(t, tier2, computeRatingTier(1))
	assert.Equal(t, tier2, computeRatingTier(25))
	assert.Equal(t, tier2, computeRatingTier(50))
	assert.Equal(t, tier1, computeRatingTier(51))
	assert.Equal(t, tier1, computeRatingTier(75))
	assert.Equal(t, tier1, computeRatingTier(100))
}

func TestPeersRatingHandler_GetTopRatedPeersFromList(t *testing.T) {
	t.Parallel()

	t.Run("asking for 0 peers should return empty list", func(t *testing.T) {
		t.Parallel()

		prh, _ := NewPeersRatingHandler(ArgPeersRatingHandler{&random.ConcurrentSafeIntRandomizer{}})
		assert.False(t, check.IfNil(prh))

		res := prh.GetTopRatedPeersFromList([]core.PeerID{"pid"}, 0)
		assert.Equal(t, 0, len(res))
	})
	t.Run("nil provided list should return empty list", func(t *testing.T) {
		t.Parallel()

		prh, _ := NewPeersRatingHandler(ArgPeersRatingHandler{&random.ConcurrentSafeIntRandomizer{}})
		assert.False(t, check.IfNil(prh))

		res := prh.GetTopRatedPeersFromList(nil, 1)
		assert.Equal(t, 0, len(res))
	})
	t.Run("no peers in maps should return empty list", func(t *testing.T) {
		t.Parallel()

		prh, _ := NewPeersRatingHandler(ArgPeersRatingHandler{&random.ConcurrentSafeIntRandomizer{}})
		assert.False(t, check.IfNil(prh))

		providedListOfPeers := []core.PeerID{"pid 1", "pid 2"}
		res := prh.GetTopRatedPeersFromList(providedListOfPeers, 5)
		assert.Equal(t, 0, len(res))
	})
	t.Run("one peer in tier 1 should work", func(t *testing.T) {
		t.Parallel()

		prh, _ := NewPeersRatingHandler(ArgPeersRatingHandler{&random.ConcurrentSafeIntRandomizer{}})
		assert.False(t, check.IfNil(prh))

		providedPid := core.PeerID("provided pid")
		for i := 0; i < maxRating; i++ {
			prh.IncreaseRating(providedPid)
		}

		providedListOfPeers := []core.PeerID{providedPid, "another pid"}
		res := prh.GetTopRatedPeersFromList(providedListOfPeers, 5)
		assert.Equal(t, 1, len(res))
		assert.Equal(t, providedPid, res[0])
	})
	t.Run("one peer in tier one should work", func(t *testing.T) {
		t.Parallel()

		prh, _ := NewPeersRatingHandler(ArgPeersRatingHandler{&random.ConcurrentSafeIntRandomizer{}})
		assert.False(t, check.IfNil(prh))

		providedPid := core.PeerID("provided pid")
		for i := 0; i < maxRating; i++ {
			prh.IncreaseRating(providedPid)
		}

		providedListOfPeers := []core.PeerID{providedPid, "another pid"}
		res := prh.GetTopRatedPeersFromList(providedListOfPeers, 1)
		assert.Equal(t, 1, len(res))
		assert.Equal(t, providedPid, res[0])
	})
	t.Run("all peers in same tier should work", func(t *testing.T) {
		t.Parallel()

		prh, _ := NewPeersRatingHandler(ArgPeersRatingHandler{&random.ConcurrentSafeIntRandomizer{}})
		assert.False(t, check.IfNil(prh))

		providedPid1 := core.PeerID("provided pid 1")
		providedPid2 := core.PeerID("provided pid 2")
		providedPid3 := core.PeerID("provided pid 3")

		prh.AddPeer(providedPid1)
		prh.AddPeer(providedPid2)
		prh.AddPeer(providedPid3)

		providedListOfPeers := []core.PeerID{providedPid1, "extra pid 1", providedPid2, providedPid3, "extra pid 2"}
		requestedNumOfPeers := 2
		res := prh.GetTopRatedPeersFromList(providedListOfPeers, requestedNumOfPeers) // should return 2 random from provided
		assert.Equal(t, requestedNumOfPeers, len(res))

		for _, resEntry := range res {
			println(fmt.Sprintf("got pid: %s", resEntry.Bytes()))
		}
	})
	t.Run("peers from multiple tiers should work", func(t *testing.T) {
		t.Parallel()

		prh, _ := NewPeersRatingHandler(ArgPeersRatingHandler{&random.ConcurrentSafeIntRandomizer{}})
		assert.False(t, check.IfNil(prh))

		providedPid1 := core.PeerID("provided pid 1")
		providedPid2 := core.PeerID("provided pid 2")
		providedPid3 := core.PeerID("provided pid 3")
		prh.AddPeer(providedPid3) // tier 3

		prh.AddPeer(providedPid2)
		prh.IncreaseRating(providedPid2) // tier 2

		for i := 0; i < maxRating; i++ {
			prh.IncreaseRating(providedPid1)
		} // tier 1

		providedListOfPeers := []core.PeerID{providedPid1, "extra pid 1", providedPid2, providedPid3, "extra pid 2"}
		requestedNumOfPeers := 2
		res := prh.GetTopRatedPeersFromList(providedListOfPeers, requestedNumOfPeers) // should return 2 random from provided
		assert.Equal(t, requestedNumOfPeers, len(res))

		for _, resEntry := range res {
			println(fmt.Sprintf("got pid: %s", resEntry.Bytes()))
		}
	})
}

func TestPeerRatingHandler_concurrency_test(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	prh, _ := NewPeersRatingHandler(ArgPeersRatingHandler{&random.ConcurrentSafeIntRandomizer{}})
	assert.False(t, check.IfNil(prh))

	prh.AddPeer("pid0")
	prh.AddPeer("pid1")

	var wg sync.WaitGroup

	numOps := 500
	wg.Add(numOps)

	for i := 1; i <= numOps; i++ {
		go func(i int) {
			defer wg.Done()

			pid1 := core.PeerID(fmt.Sprintf("pid%d", i%2))
			pid2 := core.PeerID(fmt.Sprintf("pid%d", (i+1)%2))

			prh.IncreaseRating(pid1)
			prh.DecreaseRating(pid2)
		}(i)
	}

	wg.Wait()

	// increase factor = 2, decrease factor = 1 so both pids should be in tier 1
	peers := prh.peersTiersMap[1]
	assert.Equal(t, 2, len(peers))
	_, pid0ExistsInTier := peers["pid0"]
	assert.True(t, pid0ExistsInTier)
	_, pid1ExistsInTier := peers["pid1"]
	assert.True(t, pid1ExistsInTier)

	ratingPid0 := prh.peersRatingMap["pid0"]
	assert.True(t, ratingPid0 > 90)
	ratingPid1 := prh.peersRatingMap["pid1"]
	assert.True(t, ratingPid1 > 90)

	numOps = 200
	wg.Add(numOps)

	for i := 1; i <= numOps; i++ {
		go func(i int) {
			defer wg.Done()

			pid1 := core.PeerID(fmt.Sprintf("pid%d", i%2))
			pid2 := core.PeerID(fmt.Sprintf("pid%d", (i+1)%2))

			prh.DecreaseRating(pid1)
			prh.DecreaseRating(pid2)
		}(i)
	}

	wg.Wait()

	// increase factor = 2, decrease factor = 1 so both pids should be in tier 4
	peers = prh.peersTiersMap[4]
	assert.Equal(t, 2, len(peers))
	_, pid0ExistsInTier = peers["pid0"]
	assert.True(t, pid0ExistsInTier)
	_, pid1ExistsInTier = peers["pid1"]
	assert.True(t, pid1ExistsInTier)

	ratingPid0 = prh.peersRatingMap["pid0"]
	assert.True(t, ratingPid0 < -90)
	ratingPid1 = prh.peersRatingMap["pid1"]
	assert.True(t, ratingPid1 < -90)
}
