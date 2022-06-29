package rating

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func createMockArgs() ArgPeersRatingHandler {
	return ArgPeersRatingHandler{
		TopRatedCache: &testscommon.CacherStub{},
		BadRatedCache: &testscommon.CacherStub{},
	}
}

func TestNewPeersRatingHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil top rated cache should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TopRatedCache = nil

		prh, err := NewPeersRatingHandler(args)
		assert.True(t, errors.Is(err, p2p.ErrNilCacher))
		assert.True(t, strings.Contains(err.Error(), "TopRatedCache"))
		assert.True(t, check.IfNil(prh))
	})
	t.Run("nil bad rated cache should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.BadRatedCache = nil

		prh, err := NewPeersRatingHandler(args)
		assert.True(t, errors.Is(err, p2p.ErrNilCacher))
		assert.True(t, strings.Contains(err.Error(), "BadRatedCache"))
		assert.True(t, check.IfNil(prh))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		prh, err := NewPeersRatingHandler(createMockArgs())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(prh))
	})
}

func TestPeersRatingHandler_AddPeer(t *testing.T) {
	t.Parallel()

	t.Run("new peer should add", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		providedPid := core.PeerID("provided pid")
		args := createMockArgs()
		args.TopRatedCache = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				assert.True(t, bytes.Equal(providedPid.Bytes(), key))

				wasCalled = true
				return false
			},
		}
		args.BadRatedCache = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}

		prh, _ := NewPeersRatingHandler(args)
		assert.False(t, check.IfNil(prh))

		prh.AddPeer(providedPid)
		assert.True(t, wasCalled)
	})
	t.Run("peer in top rated should not add", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		providedPid := core.PeerID("provided pid")
		args := createMockArgs()
		args.TopRatedCache = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, true
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				wasCalled = true
				return false
			},
		}
		args.BadRatedCache = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}

		prh, _ := NewPeersRatingHandler(args)
		assert.False(t, check.IfNil(prh))

		prh.AddPeer(providedPid)
		assert.False(t, wasCalled)
	})
	t.Run("peer in bad rated should not add", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		providedPid := core.PeerID("provided pid")
		args := createMockArgs()
		args.TopRatedCache = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				wasCalled = true
				return false
			},
		}
		args.BadRatedCache = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, true
			},
		}

		prh, _ := NewPeersRatingHandler(args)
		assert.False(t, check.IfNil(prh))

		prh.AddPeer(providedPid)
		assert.False(t, wasCalled)
	})
}

func TestPeersRatingHandler_IncreaseRating(t *testing.T) {
	t.Parallel()

	t.Run("new peer should add to cache", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		providedPid := core.PeerID("provided pid")
		args := createMockArgs()
		args.TopRatedCache = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				assert.True(t, bytes.Equal(providedPid.Bytes(), key))

				wasCalled = true
				return false
			},
		}
		args.BadRatedCache = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}
		prh, _ := NewPeersRatingHandler(args)
		assert.False(t, check.IfNil(prh))

		prh.IncreaseRating(providedPid)
		assert.True(t, wasCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cacheMap := make(map[string]interface{})
		providedPid := core.PeerID("provided pid")
		args := createMockArgs()
		args.TopRatedCache = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				val, found := cacheMap[string(key)]
				return val, found
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				cacheMap[string(key)] = value
				return false
			},
		}

		prh, _ := NewPeersRatingHandler(args)
		assert.False(t, check.IfNil(prh))

		prh.IncreaseRating(providedPid)
		val, found := cacheMap[string(providedPid.Bytes())]
		assert.True(t, found)
		assert.Equal(t, defaultRating, val)

		// exceed the limit
		numOfCalls := 100
		for i := 0; i < numOfCalls; i++ {
			prh.IncreaseRating(providedPid)
		}
		val, found = cacheMap[string(providedPid.Bytes())]
		assert.True(t, found)
		assert.Equal(t, int32(maxRating), val)
	})
}

func TestPeersRatingHandler_DecreaseRating(t *testing.T) {
	t.Parallel()

	t.Run("new peer should add to cache", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		providedPid := core.PeerID("provided pid")
		args := createMockArgs()
		args.TopRatedCache = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				assert.True(t, bytes.Equal(providedPid.Bytes(), key))

				wasCalled = true
				return false
			},
		}
		args.BadRatedCache = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}
		prh, _ := NewPeersRatingHandler(args)
		assert.False(t, check.IfNil(prh))

		prh.DecreaseRating(providedPid)
		assert.True(t, wasCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		topRatedCacheMap := make(map[string]interface{})
		badRatedCacheMap := make(map[string]interface{})
		providedPid := core.PeerID("provided pid")
		args := createMockArgs()
		args.TopRatedCache = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				val, found := topRatedCacheMap[string(key)]
				return val, found
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				topRatedCacheMap[string(key)] = value
				return false
			},
			RemoveCalled: func(key []byte) {
				delete(topRatedCacheMap, string(key))
			},
		}
		args.BadRatedCache = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				val, found := badRatedCacheMap[string(key)]
				return val, found
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				badRatedCacheMap[string(key)] = value
				return false
			},
			RemoveCalled: func(key []byte) {
				delete(badRatedCacheMap, string(key))
			},
		}

		prh, _ := NewPeersRatingHandler(args)
		assert.False(t, check.IfNil(prh))

		// first call just adds it with default rating
		prh.DecreaseRating(providedPid)
		val, found := topRatedCacheMap[string(providedPid.Bytes())]
		assert.True(t, found)
		assert.Equal(t, defaultRating, val)

		// exceed the limit
		numOfCalls := 200
		for i := 0; i < numOfCalls; i++ {
			prh.DecreaseRating(providedPid)
		}
		val, found = badRatedCacheMap[string(providedPid.Bytes())]
		assert.True(t, found)
		assert.Equal(t, int32(minRating), val)

		// move back to top tier
		for i := 0; i < numOfCalls; i++ {
			prh.IncreaseRating(providedPid)
		}
		_, found = badRatedCacheMap[string(providedPid.Bytes())]
		assert.False(t, found)

		val, found = topRatedCacheMap[string(providedPid.Bytes())]
		assert.True(t, found)
		assert.Equal(t, int32(maxRating), val)
	})
}

func TestPeersRatingHandler_GetTopRatedPeersFromList(t *testing.T) {
	t.Parallel()

	t.Run("asking for 0 peers should return empty list", func(t *testing.T) {
		t.Parallel()

		prh, _ := NewPeersRatingHandler(createMockArgs())
		assert.False(t, check.IfNil(prh))

		res := prh.GetTopRatedPeersFromList([]core.PeerID{"pid"}, 0)
		assert.Equal(t, 0, len(res))
	})
	t.Run("nil provided list should return empty list", func(t *testing.T) {
		t.Parallel()

		prh, _ := NewPeersRatingHandler(createMockArgs())
		assert.False(t, check.IfNil(prh))

		res := prh.GetTopRatedPeersFromList(nil, 1)
		assert.Equal(t, 0, len(res))
	})
	t.Run("no peers in maps should return empty list", func(t *testing.T) {
		t.Parallel()

		prh, _ := NewPeersRatingHandler(createMockArgs())
		assert.False(t, check.IfNil(prh))

		providedListOfPeers := []core.PeerID{"pid 1", "pid 2"}
		res := prh.GetTopRatedPeersFromList(providedListOfPeers, 5)
		assert.Equal(t, 0, len(res))
	})
	t.Run("one peer in top rated, asking for one should work", func(t *testing.T) {
		t.Parallel()

		providedPid := core.PeerID("provided pid")
		args := createMockArgs()
		args.TopRatedCache = &testscommon.CacherStub{
			LenCalled: func() int {
				return 1
			},
			KeysCalled: func() [][]byte {
				return [][]byte{providedPid.Bytes()}
			},
			HasCalled: func(key []byte) bool {
				return bytes.Equal(key, providedPid.Bytes())
			},
		}
		prh, _ := NewPeersRatingHandler(args)
		assert.False(t, check.IfNil(prh))

		providedListOfPeers := []core.PeerID{providedPid, "another pid"}
		res := prh.GetTopRatedPeersFromList(providedListOfPeers, 1)
		assert.Equal(t, 1, len(res))
		assert.Equal(t, providedPid, res[0])
	})
	t.Run("one peer in each, asking for two should work", func(t *testing.T) {
		t.Parallel()

		providedTopPid := core.PeerID("provided top pid")
		providedBadPid := core.PeerID("provided bad pid")
		args := createMockArgs()
		args.TopRatedCache = &testscommon.CacherStub{
			LenCalled: func() int {
				return 1
			},
			KeysCalled: func() [][]byte {
				return [][]byte{providedTopPid.Bytes()}
			},
			HasCalled: func(key []byte) bool {
				return bytes.Equal(key, providedTopPid.Bytes())
			},
		}
		args.BadRatedCache = &testscommon.CacherStub{
			LenCalled: func() int {
				return 1
			},
			KeysCalled: func() [][]byte {
				return [][]byte{providedBadPid.Bytes()}
			},
			HasCalled: func(key []byte) bool {
				return bytes.Equal(key, providedBadPid.Bytes())
			},
		}
		prh, _ := NewPeersRatingHandler(args)
		assert.False(t, check.IfNil(prh))

		providedListOfPeers := []core.PeerID{providedTopPid, providedBadPid, "another pid"}
		expectedListOfPeers := []core.PeerID{providedTopPid, providedBadPid}
		res := prh.GetTopRatedPeersFromList(providedListOfPeers, 2)
		assert.Equal(t, expectedListOfPeers, res)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedPid1, providedPid2, providedPid3 := core.PeerID("provided pid 1"), core.PeerID("provided pid 2"), core.PeerID("provided pid 3")
		args := createMockArgs()
		args.TopRatedCache = &testscommon.CacherStub{
			LenCalled: func() int {
				return 3
			},
			KeysCalled: func() [][]byte {
				return [][]byte{providedPid1.Bytes(), providedPid2.Bytes(), providedPid3.Bytes()}
			},
			HasCalled: func(key []byte) bool {
				has := bytes.Equal(key, providedPid1.Bytes()) ||
					bytes.Equal(key, providedPid2.Bytes()) ||
					bytes.Equal(key, providedPid3.Bytes())
				return has
			},
		}
		prh, _ := NewPeersRatingHandler(args)
		assert.False(t, check.IfNil(prh))

		providedListOfPeers := []core.PeerID{providedPid1, providedPid2, providedPid3, "another pid 1", "another pid 2"}
		expectedListOfPeers := []core.PeerID{providedPid1, providedPid2, providedPid3}
		res := prh.GetTopRatedPeersFromList(providedListOfPeers, 2)
		assert.Equal(t, expectedListOfPeers, res)
	})
}
