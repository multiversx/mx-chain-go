package processor

import (
	"bytes"
	"errors"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	coreAtomic "github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/random"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
)

func createMockArgPeerAuthenticationRequestsProcessor() ArgPeerAuthenticationRequestsProcessor {
	return ArgPeerAuthenticationRequestsProcessor{
		RequestHandler:          &testscommon.RequestHandlerStub{},
		NodesCoordinator:        &shardingMocks.NodesCoordinatorStub{},
		PeerAuthenticationPool:  &testscommon.CacherMock{},
		ShardId:                 0,
		Epoch:                   0,
		MessagesInChunk:         5,
		MinPeersThreshold:       0.8,
		DelayBetweenRequests:    time.Second,
		MaxTimeout:              5 * time.Second,
		MaxMissingKeysInRequest: 10,
		Randomizer:              &random.ConcurrentSafeIntRandomizer{},
	}
}

func getSortedSlice(slice [][]byte) [][]byte {
	sort.Slice(slice, func(i, j int) bool {
		return bytes.Compare(slice[i], slice[j]) < 0
	})

	return slice
}

func TestNewPeerAuthenticationRequestsProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil request handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.RequestHandler = nil

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.Equal(t, heartbeat.ErrNilRequestHandler, err)
		assert.True(t, check.IfNil(processor))
	})
	t.Run("nil nodes coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.NodesCoordinator = nil

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.Equal(t, heartbeat.ErrNilNodesCoordinator, err)
		assert.True(t, check.IfNil(processor))
	})
	t.Run("nil peer auth pool should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.PeerAuthenticationPool = nil

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.Equal(t, heartbeat.ErrNilPeerAuthenticationPool, err)
		assert.True(t, check.IfNil(processor))
	})
	t.Run("invalid messages in chunk should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.MessagesInChunk = 0

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "MessagesInChunk"))
		assert.True(t, check.IfNil(processor))
	})
	t.Run("invalid min peers threshold should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.MinPeersThreshold = 0.1

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "MinPeersThreshold"))
		assert.True(t, check.IfNil(processor))
	})
	t.Run("min peers threshold too big should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.MinPeersThreshold = 1.001

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "MinPeersThreshold"))
		assert.True(t, check.IfNil(processor))
	})
	t.Run("invalid delay between requests should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.DelayBetweenRequests = time.Second - time.Nanosecond

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "DelayBetweenRequests"))
		assert.True(t, check.IfNil(processor))
	})
	t.Run("invalid max timeout should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.MaxMissingKeysInRequest = 0

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "MaxMissingKeysInRequest"))
		assert.True(t, check.IfNil(processor))
	})
	t.Run("invalid max missing keys should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.MaxTimeout = time.Second - time.Nanosecond

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "MaxTimeout"))
		assert.True(t, check.IfNil(processor))
	})
	t.Run("nil randomizer should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.Randomizer = nil

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.Equal(t, heartbeat.ErrNilRandomizer, err)
		assert.True(t, check.IfNil(processor))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		processor, err := NewPeerAuthenticationRequestsProcessor(createMockArgPeerAuthenticationRequestsProcessor())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(processor))

		err = processor.Close()
		assert.Nil(t, err)
	})
}

func TestPeerAuthenticationRequestsProcessor_startRequestingMessages(t *testing.T) {
	t.Parallel()

	t.Run("threshold reached from requestKeysChunks", func(t *testing.T) {
		t.Parallel()

		providedKeys := [][]byte{[]byte("pk3"), []byte("pk2"), []byte("pk0"), []byte("pk1")}
		providedKeysMap := make(map[uint32][][]byte, 2)
		providedKeysMap[0] = providedKeys[:len(providedKeys)/2]
		providedKeysMap[1] = providedKeys[len(providedKeys)/2:]
		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
				return providedKeysMap, nil
			},
		}

		args.MessagesInChunk = 5 // all provided keys in one chunk

		wasRequestPeerAuthenticationsChunkCalled := coreAtomic.Flag{}
		wasRequestPeerAuthenticationsByHashesCalled := coreAtomic.Flag{}
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestPeerAuthenticationsChunkCalled: func(destShardID uint32, chunkIndex uint32) {
				wasRequestPeerAuthenticationsChunkCalled.SetValue(true)
				assert.Equal(t, uint32(0), chunkIndex)
			},
			RequestPeerAuthenticationsByHashesCalled: func(destShardID uint32, hashes [][]byte) {
				wasRequestPeerAuthenticationsByHashesCalled.SetValue(true)
			},
		}

		args.PeerAuthenticationPool = &testscommon.CacherStub{
			KeysCalled: func() [][]byte {
				return providedKeys // all keys requested available in cache
			},
		}

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(processor))

		time.Sleep(3 * time.Second)
		_ = processor.Close()

		assert.False(t, wasRequestPeerAuthenticationsByHashesCalled.IsSet())
		assert.True(t, wasRequestPeerAuthenticationsChunkCalled.IsSet())
	})
	t.Run("should work: <-requestsTimer.C", func(t *testing.T) {
		t.Parallel()

		providedKeys := [][]byte{[]byte("pk3"), []byte("pk2"), []byte("pk0"), []byte("pk1")}
		providedKeysMap := make(map[uint32][][]byte, 2)
		providedKeysMap[0] = providedKeys[:len(providedKeys)/2]
		providedKeysMap[1] = providedKeys[len(providedKeys)/2:]
		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
				return providedKeysMap, nil
			},
		}

		args.MessagesInChunk = 5   // all provided keys in one chunk
		args.MinPeersThreshold = 1 // need messages from all peers

		wasRequestPeerAuthenticationsChunkCalled := coreAtomic.Flag{}
		wasRequestPeerAuthenticationsByHashesCalled := coreAtomic.Flag{}
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestPeerAuthenticationsChunkCalled: func(destShardID uint32, chunkIndex uint32) {
				wasRequestPeerAuthenticationsChunkCalled.SetValue(true)
				assert.Equal(t, uint32(0), chunkIndex)
			},
			RequestPeerAuthenticationsByHashesCalled: func(destShardID uint32, hashes [][]byte) {
				wasRequestPeerAuthenticationsByHashesCalled.SetValue(true)
				assert.Equal(t, getSortedSlice(providedKeys[len(providedKeys)/2:]), getSortedSlice(hashes))
			},
		}

		args.PeerAuthenticationPool = &testscommon.CacherStub{
			KeysCalled: func() [][]byte {
				return providedKeys[:len(providedKeys)/2]
			},
		}

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(processor))

		time.Sleep(3 * time.Second)
		_ = processor.Close()

		assert.True(t, wasRequestPeerAuthenticationsByHashesCalled.IsSet())
		assert.True(t, wasRequestPeerAuthenticationsChunkCalled.IsSet())
	})
}

func TestPeerAuthenticationRequestsProcessor_requestKeysChunks(t *testing.T) {
	t.Parallel()

	providedKeys := [][]byte{[]byte("pk3"), []byte("pk2"), []byte("pk0"), []byte("pk1")} // 2 chunks of 2
	counter := uint32(0)
	args := createMockArgPeerAuthenticationRequestsProcessor()
	args.MessagesInChunk = 2
	args.RequestHandler = &testscommon.RequestHandlerStub{
		RequestPeerAuthenticationsChunkCalled: func(destShardID uint32, chunkIndex uint32) {
			assert.Equal(t, counter, chunkIndex)
			atomic.AddUint32(&counter, 1)
		},
	}

	processor, err := NewPeerAuthenticationRequestsProcessor(args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(processor))

	processor.requestKeysChunks(providedKeys)
}

func TestPeerAuthenticationRequestsProcessor_getMaxChunks(t *testing.T) {
	t.Parallel()

	args := createMockArgPeerAuthenticationRequestsProcessor()
	args.MessagesInChunk = 2

	processor, err := NewPeerAuthenticationRequestsProcessor(args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(processor))

	maxChunks := processor.getMaxChunks(nil)
	assert.Equal(t, uint32(0), maxChunks)

	providedBuff := [][]byte{[]byte("msg")}
	maxChunks = processor.getMaxChunks(providedBuff)
	assert.Equal(t, uint32(1), maxChunks)

	providedBuff = [][]byte{[]byte("msg"), []byte("msg")}
	maxChunks = processor.getMaxChunks(providedBuff)
	assert.Equal(t, uint32(1), maxChunks)

	providedBuff = [][]byte{[]byte("msg"), []byte("msg"), []byte("msg")}
	maxChunks = processor.getMaxChunks(providedBuff)
	assert.Equal(t, uint32(2), maxChunks)
}

func TestPeerAuthenticationRequestsProcessor_isThresholdReached(t *testing.T) {
	t.Parallel()

	providedPks := [][]byte{[]byte("pk0"), []byte("pk1"), []byte("pk2"), []byte("pk3")}
	args := createMockArgPeerAuthenticationRequestsProcessor()
	args.MinPeersThreshold = 0.6
	counter := uint32(0)
	args.PeerAuthenticationPool = &testscommon.CacherStub{
		KeysCalled: func() [][]byte {
			var keys = make([][]byte, 0)
			switch atomic.LoadUint32(&counter) {
			case 0:
				keys = [][]byte{[]byte("pk0")}
			case 1:
				keys = [][]byte{[]byte("pk0"), []byte("pk2")}
			case 2:
				keys = [][]byte{[]byte("pk0"), []byte("pk1"), []byte("pk2")}
			case 3:
				keys = [][]byte{[]byte("pk0"), []byte("pk1"), []byte("pk2"), []byte("pk3")}
			}

			atomic.AddUint32(&counter, 1)
			return keys
		},
	}

	processor, err := NewPeerAuthenticationRequestsProcessor(args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(processor))

	assert.False(t, processor.isThresholdReached(providedPks)) // counter 0
	assert.False(t, processor.isThresholdReached(providedPks)) // counter 1
	assert.True(t, processor.isThresholdReached(providedPks))  // counter 2
	assert.True(t, processor.isThresholdReached(providedPks))  // counter 3
}

func TestPeerAuthenticationRequestsProcessor_requestMissingKeys(t *testing.T) {
	t.Parallel()

	t.Run("get missing keys returns nil", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestPeerAuthenticationsByHashesCalled: func(destShardID uint32, hashes [][]byte) {
				wasCalled = true
			},
		}

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(processor))

		processor.requestMissingKeys(nil)
		assert.False(t, wasCalled)
	})
}

func TestPeerAuthenticationRequestsProcessor_getRandMaxMissingKeys(t *testing.T) {
	t.Parallel()

	providedPks := [][]byte{[]byte("pk0"), []byte("pk1"), []byte("pk2"), []byte("pk3"), []byte("pk5"),
		[]byte("pk8"), []byte("pk4"), []byte("pk7"), []byte("pk6")}

	args := createMockArgPeerAuthenticationRequestsProcessor()
	args.MaxMissingKeysInRequest = 3
	processor, err := NewPeerAuthenticationRequestsProcessor(args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(processor))

	for i := 0; i < 100; i++ {
		randMissingKeys := processor.getRandMaxMissingKeys(providedPks)
		assert.Equal(t, int(args.MaxMissingKeysInRequest), len(randMissingKeys))

		randMissingKeys = getSortedSlice(randMissingKeys)
		for j := 0; j < len(randMissingKeys)-1; j++ {
			assert.NotEqual(t, randMissingKeys[j], randMissingKeys[j+1])
		}
	}
}
