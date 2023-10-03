package processor

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	mxAtomic "github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/random"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgPeerAuthenticationRequestsProcessor() ArgPeerAuthenticationRequestsProcessor {
	return ArgPeerAuthenticationRequestsProcessor{
		RequestHandler:          &testscommon.RequestHandlerStub{},
		NodesCoordinator:        &shardingMocks.NodesCoordinatorStub{},
		PeerAuthenticationPool:  &testscommon.CacherMock{},
		ShardId:                 0,
		Epoch:                   0,
		MinPeersThreshold:       0.8,
		DelayBetweenRequests:    time.Second,
		MaxTimeoutForRequests:   5 * time.Second,
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
	t.Run("invalid max missing keys should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.MaxMissingKeysInRequest = 0

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "MaxMissingKeysInRequest"))
		assert.True(t, check.IfNil(processor))
	})
	t.Run("invalid max timeout for requests should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.MaxTimeoutForRequests = time.Second - time.Nanosecond

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "MaxTimeoutForRequests"))
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

	t.Run("should work: <-requestsTimer.C", func(t *testing.T) {
		t.Parallel()

		providedEligibleKeys := [][]byte{[]byte("pk1"), []byte("pk2"), []byte("pk3"), []byte("pk4")}
		providedEligibleKeysMap := make(map[uint32][][]byte, 2)
		providedEligibleKeysMap[0] = providedEligibleKeys[:len(providedEligibleKeys)/2]
		providedEligibleKeysMap[1] = providedEligibleKeys[len(providedEligibleKeys)/2:]

		providedWaitingKeys := [][]byte{[]byte("pk5"), []byte("pk6"), []byte("pk7"), []byte("pk8")}
		providedWaitingKeysMap := make(map[uint32][][]byte, 2)
		providedWaitingKeysMap[0] = providedWaitingKeys[:len(providedWaitingKeys)/2]
		providedWaitingKeysMap[1] = providedWaitingKeys[len(providedWaitingKeys)/2:]

		args := createMockArgPeerAuthenticationRequestsProcessor()
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
				return providedEligibleKeysMap, nil
			},
			GetAllWaitingValidatorsPublicKeysCalled: func(_ uint32) (map[uint32][][]byte, error) {
				return providedWaitingKeysMap, nil
			},
		}

		args.MinPeersThreshold = 1 // need messages from all peers

		mutRequestedHashes := sync.Mutex{}
		requestedHashes := make(map[string]struct{})
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestPeerAuthenticationsByHashesCalled: func(destShardID uint32, hashes [][]byte) {
				mutRequestedHashes.Lock()
				defer mutRequestedHashes.Unlock()

				insertSliceInMap(hashes, requestedHashes)
			},
		}

		args.PeerAuthenticationPool = &testscommon.CacherStub{
			KeysCalled: func() [][]byte {
				return providedEligibleKeysMap[0]
			},
		}

		processor, err := NewPeerAuthenticationRequestsProcessor(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(processor))

		time.Sleep(3 * time.Second)
		_ = processor.Close()

		expectedRequestedHashes := make(map[string]struct{})
		insertSliceInMap(providedEligibleKeysMap[1], expectedRequestedHashes) // providedEligibleKeysMap[0] was already in pool
		insertSliceInMap(providedWaitingKeysMap[0], expectedRequestedHashes)
		insertSliceInMap(providedWaitingKeysMap[1], expectedRequestedHashes)
		mutRequestedHashes.Lock()
		assert.Equal(t, expectedRequestedHashes, requestedHashes)
		mutRequestedHashes.Unlock()
	})
}

func insertSliceInMap(hashesSlice [][]byte, hashesMap map[string]struct{}) {
	for _, hash := range hashesSlice {
		hashesMap[string(hash)] = struct{}{}
	}
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

	processor, err := NewPeerAuthenticationRequestsProcessorWithoutGoRoutine(args)
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

		processor, err := NewPeerAuthenticationRequestsProcessorWithoutGoRoutine(args)
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
	processor, err := NewPeerAuthenticationRequestsProcessorWithoutGoRoutine(args)
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

func TestPeerAuthenticationRequestsProcessor_goRoutineIsWorkingAndCloseShouldStopIt(t *testing.T) {
	t.Parallel()

	args := createMockArgPeerAuthenticationRequestsProcessor()
	args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
		GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
			return map[uint32][][]byte{
				0: {[]byte("pk0")},
			}, nil
		},
	}
	keysCalled := &mxAtomic.Flag{}
	args.PeerAuthenticationPool = &testscommon.CacherStub{
		KeysCalled: func() [][]byte {
			keysCalled.SetValue(true)
			return make([][]byte, 0)
		},
	}

	processor, _ := NewPeerAuthenticationRequestsProcessor(args)
	time.Sleep(args.DelayBetweenRequests*2 + time.Millisecond*300) // wait for the go routine to start and execute at least once
	assert.True(t, keysCalled.IsSet())

	err := processor.Close()
	assert.Nil(t, err)

	time.Sleep(time.Second) // wait for the go routine to stop
	keysCalled.SetValue(false)

	time.Sleep(args.DelayBetweenRequests*2 + time.Millisecond*300) // if the go routine did not stop it will set again the flag
	assert.False(t, keysCalled.IsSet())
}

func TestPeerAuthenticationRequestsProcessor_CloseCalledTwiceShouldNotPanicNorError(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	args := createMockArgPeerAuthenticationRequestsProcessor()
	processor, _ := NewPeerAuthenticationRequestsProcessor(args)

	time.Sleep(args.DelayBetweenRequests*2 + time.Millisecond*300) // wait for the go routine to start and execute at least once

	err := processor.Close()
	assert.Nil(t, err)

	err = processor.Close()
	assert.Nil(t, err)
}
