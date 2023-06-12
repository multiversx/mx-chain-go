package track_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	processBlock "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CreateSovereignChainShardTrackerMockArguments -
func CreateSovereignChainShardTrackerMockArguments() track.ArgShardTracker {
	shardBlockTrackArguments := CreateShardTrackerMockArguments()

	shardBlockTrackArguments.RequestHandler = &testscommon.ExtendedShardHeaderRequestHandlerStub{}

	argsHeaderValidator := processBlock.ArgsHeaderValidator{
		Hasher:      &hashingMocks.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
	}
	headerValidator, _ := processBlock.NewHeaderValidator(argsHeaderValidator)
	sovereignChainHeaderValidator, _ := processBlock.NewSovereignChainHeaderValidator(headerValidator)
	shardBlockTrackArguments.HeaderValidator = sovereignChainHeaderValidator

	return shardBlockTrackArguments
}

func TestNewSovereignChainShardBlockTrack_ShouldErrNilBlockTracker(t *testing.T) {
	t.Parallel()

	scsbt, err := track.NewSovereignChainShardBlockTrack(nil)
	assert.Nil(t, scsbt)
	assert.Equal(t, process.ErrNilBlockTracker, err)
}

func TestNewSovereignChainShardBlockTrack_ShouldErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	shardBlockTrackArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)

	sbt.SetBlockProcessor(nil)

	scsbt, err := track.NewSovereignChainShardBlockTrack(sbt)
	assert.Nil(t, scsbt)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestNewSovereignChainShardBlockTrack_ShouldErrWhenInitCrossNotarizedStartHeadersFails(t *testing.T) {
	t.Parallel()

	shardBlockTrackArguments := CreateSovereignChainShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)

	sbt.ClearStartHeaders()

	scsbt, err := track.NewSovereignChainShardBlockTrack(sbt)
	assert.Nil(t, scsbt)
	assert.ErrorIs(t, err, process.ErrMissingHeader)
}

func TestNewSovereignChainShardBlockTrack_ShouldWork(t *testing.T) {
	t.Parallel()

	shardBlockTrackArguments := CreateSovereignChainShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)

	scsbt, err := track.NewSovereignChainShardBlockTrack(sbt)
	assert.NotNil(t, scsbt)
	assert.Nil(t, err)
}

func TestSovereignChainShardBlockTrack_ComputeLongestSelfChainShouldWork(t *testing.T) {
	t.Parallel()

	lastNotarizedHeader := &block.Header{Nonce: 1}
	lastNotarizedHash := []byte("hash")
	shardBlockTrackArguments := CreateSovereignChainShardTrackerMockArguments()
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinatorMock.CurrentShard = 1
	shardBlockTrackArguments.ShardCoordinator = shardCoordinatorMock
	sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)
	selfNotarizer := &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			if shardID != shardCoordinatorMock.CurrentShard {
				return nil, nil, errors.New("wrong shard ID")
			}
			return lastNotarizedHeader, lastNotarizedHash, nil
		},
	}
	sbt.SetSelfNotarizer(selfNotarizer)
	scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

	header, hash, headers, hashes := scsbt.ComputeLongestSelfChain()

	assert.Equal(t, lastNotarizedHeader, header)
	assert.NotNil(t, lastNotarizedHash, hash)
	assert.Equal(t, 0, len(headers))
	assert.Equal(t, 0, len(hashes))
}

func TestSovereignChainShardBlockTrack_GetSelfNotarizedHeaderShouldWork(t *testing.T) {
	t.Parallel()

	lastNotarizedHeader := &block.Header{Nonce: 1}
	lastNotarizedHash := []byte("hash")
	shardBlockTrackArguments := CreateSovereignChainShardTrackerMockArguments()
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinatorMock.CurrentShard = 1
	shardBlockTrackArguments.ShardCoordinator = shardCoordinatorMock
	sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)
	selfNotarizer := &mock.BlockNotarizerHandlerMock{
		GetNotarizedHeaderCalled: func(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
			if shardID != shardCoordinatorMock.CurrentShard {
				return nil, nil, errors.New("wrong shard ID")
			}
			return lastNotarizedHeader, lastNotarizedHash, nil
		},
	}
	sbt.SetSelfNotarizer(selfNotarizer)
	scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

	header, hash, err := scsbt.GetSelfNotarizedHeader(core.MetachainShardId, 0)

	assert.Equal(t, lastNotarizedHeader, header)
	assert.NotNil(t, lastNotarizedHash, hash)
	assert.Nil(t, err)
}

func TestSovereignChainShardBlockTrack_ReceivedHeaderShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should add extended shard header to sovereign chain tracked headers", func(t *testing.T) {
		t.Parallel()

		shardBlockTrackArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		header := &block.Header{Nonce: 1}
		headerV2 := &block.HeaderV2{Header: header}
		extendedShardHeader := &block.ShardHeaderExtended{
			Header: headerV2,
		}
		extendedShardHeaderHash := []byte("hash")
		scsbt.ReceivedHeader(extendedShardHeader, extendedShardHeaderHash)
		headers, _ := scsbt.GetTrackedHeaders(core.SovereignChainShardId)

		require.Equal(t, 1, len(headers))
		assert.Equal(t, extendedShardHeader, headers[0])
	})

	t.Run("should add shard header to sovereign chain tracked headers", func(t *testing.T) {
		t.Parallel()

		shardBlockTrackArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		header := &block.Header{Nonce: 1}
		headerHash := []byte("hash")
		scsbt.ReceivedHeader(header, headerHash)
		headers, _ := scsbt.GetTrackedHeaders(header.GetShardID())

		require.Equal(t, 1, len(headers))
		assert.Equal(t, header, headers[0])
	})
}

func TestSovereignChainShardBlockTrack_ReceivedExtendedShardHeaderShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should not add to tracked headers when extended shard header is out of range", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{Nonce: 1},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: 1002,
				},
			},
		}
		shardHeaderExtendedHash := []byte("hash")

		scsbt.ReceivedExtendedShardHeader(shardHeaderExtended, shardHeaderExtendedHash)
		headers, _ := scsbt.GetTrackedHeaders(core.SovereignChainShardId)
		assert.Zero(t, len(headers))
	})

	t.Run("should add to tracked headers when extended shard header is in range", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: 1000,
				},
			},
		}
		shardHeaderExtendedHash := []byte("hash")

		scsbt.ReceivedExtendedShardHeader(shardHeaderExtended, shardHeaderExtendedHash)
		headers, _ := scsbt.GetTrackedHeaders(core.SovereignChainShardId)

		require.Equal(t, 1, len(headers))
		assert.Equal(t, shardHeaderExtended, headers[0])
	})
}

func TestSovereignChainShardBlockTrack_ShouldAddExtendedShardHeaderShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should return false when extended shard header is out of range", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		maxNumHeadersToKeepPerShard := uint64(scsbt.GetMaxNumHeadersToKeepPerShard())

		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{Nonce: 1},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: maxNumHeadersToKeepPerShard + 2,
				},
			},
		}

		result := scsbt.ShouldAddExtendedShardHeader(shardHeaderExtended)
		assert.False(t, result)
	})

	t.Run("should return true when extended shard header is in range", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		maxNumHeadersToKeepPerShard := uint64(scsbt.GetMaxNumHeadersToKeepPerShard())

		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: maxNumHeadersToKeepPerShard,
				},
			},
		}

		result := scsbt.ShouldAddExtendedShardHeader(shardHeaderExtended)
		assert.True(t, result)
	})
}

func TestSovereignChainShardBlockTrack_DoWhitelistWithExtendedShardHeaderIfNeededShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should not whitelist when extended shard header is out of range", func(t *testing.T) {
		t.Parallel()

		cache := make(map[string]struct{})
		mutCache := sync.Mutex{}
		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		shardArguments.WhitelistHandler = &testscommon.WhiteListHandlerStub{
			AddCalled: func(keys [][]byte) {
				mutCache.Lock()
				for _, key := range keys {
					cache[string(key)] = struct{}{}
				}
				mutCache.Unlock()
			},
		}
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{Nonce: 1},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		txHash := []byte("hash")
		incomingMiniBlocks := []*block.MiniBlock{
			{
				TxHashes: [][]byte{
					txHash,
				},
			},
		}
		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: process.MaxHeadersToWhitelistInAdvance + 2,
				},
			},
			IncomingMiniBlocks: incomingMiniBlocks,
		}

		scsbt.DoWhitelistWithExtendedShardHeaderIfNeeded(shardHeaderExtended)
		_, ok := cache[string(txHash)]

		assert.False(t, ok)
	})

	t.Run("should whitelist when extended shard header is in range", func(t *testing.T) {
		t.Parallel()

		cache := make(map[string]struct{})
		mutCache := sync.Mutex{}
		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		shardArguments.WhitelistHandler = &testscommon.WhiteListHandlerStub{
			AddCalled: func(keys [][]byte) {
				mutCache.Lock()
				for _, key := range keys {
					cache[string(key)] = struct{}{}
				}
				mutCache.Unlock()
			},
		}
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		txHash := []byte("hash")
		incomingMiniBlocks := []*block.MiniBlock{
			{
				TxHashes: [][]byte{
					txHash,
				},
			},
		}
		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: process.MaxHeadersToWhitelistInAdvance,
				},
			},
			IncomingMiniBlocks: incomingMiniBlocks,
		}

		scsbt.DoWhitelistWithExtendedShardHeaderIfNeeded(shardHeaderExtended)
		_, ok := cache[string(txHash)]

		assert.True(t, ok)
	})
}

func TestSovereignChainShardBlockTrack_IsExtendedShardHeaderOutOfRangeShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should return true when extended shard header is out of range", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		nonce := uint64(8)
		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: nonce,
				},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: nonce + process.MaxHeadersToWhitelistInAdvance + 1,
				},
			},
		}

		assert.True(t, scsbt.IsExtendedShardHeaderOutOfRange(shardHeaderExtended))
	})

	t.Run("should return false when extended shard header is in range", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		nonce := uint64(8)
		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: nonce,
				},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: nonce + process.MaxHeadersToWhitelistInAdvance,
				},
			},
		}

		assert.False(t, scsbt.IsExtendedShardHeaderOutOfRange(shardHeaderExtended))
	})
}

func TestSovereignChainShardBlockTrack_ComputeLongestExtendedShardChainFromLastNotarizedShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Round:    1,
					Nonce:    1,
					RandSeed: []byte("rand seed init"),
				},
			},
		}

		shardHeaderExtendedInitHash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, shardHeaderExtendedInit)

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		headerInitHash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, shardHeaderExtendedInit.Header)
		shardHeaderExtended1 := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Round:        2,
					Nonce:        2,
					PrevHash:     headerInitHash,
					PrevRandSeed: shardHeaderExtendedInit.GetRandSeed(),
					RandSeed:     []byte("rand seed 1"),
				},
			},
		}

		shardHeaderExtendedHash1, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, shardHeaderExtended1)

		headerHash1, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, shardHeaderExtended1.Header)
		shardHeaderExtended2 := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Round:        3,
					Nonce:        3,
					PrevHash:     headerHash1,
					PrevRandSeed: shardHeaderExtended1.GetRandSeed(),
					RandSeed:     []byte("rand seed 2"),
				},
			},
		}

		shardHeaderExtendedHash2, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, shardHeaderExtended2)

		headerHash2, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, shardHeaderExtended2.Header)
		shardHeaderExtended3 := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Round:        4,
					Nonce:        4,
					PrevHash:     headerHash2,
					PrevRandSeed: shardHeaderExtended2.GetRandSeed(),
					RandSeed:     []byte("rand seed 3"),
				},
			},
		}

		shardHeaderExtendedHash3, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, shardHeaderExtended3)

		scsbt.AddTrackedHeader(shardHeaderExtended1, shardHeaderExtendedHash1)
		scsbt.AddTrackedHeader(shardHeaderExtended2, shardHeaderExtendedHash2)
		scsbt.AddTrackedHeader(shardHeaderExtended3, shardHeaderExtendedHash3)

		headers, _, _ := scsbt.ComputeLongestExtendedShardChainFromLastNotarized()

		require.Equal(t, 2, len(headers))
		assert.Equal(t, shardHeaderExtended1, headers[0])
		assert.Equal(t, shardHeaderExtended2, headers[1])
	})
}

func TestSovereignChainShardBlockTrack_CleanupHeadersBehindNonceShouldWork(t *testing.T) {
	t.Parallel()

	shardArguments := CreateSovereignChainShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

	header := &block.Header{
		ShardID: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   1,
	}
	headerHash := []byte("hash")

	scsbt.AddSelfNotarizedHeader(header.GetShardID(), header, headerHash)
	scsbt.AddTrackedHeader(header, headerHash)

	shardHeaderExtended := &block.ShardHeaderExtended{
		Header: &block.HeaderV2{
			Header: &block.Header{
				Nonce: 1,
			},
		},
	}
	shardHeaderExtendedHash := []byte("extended_hash")

	scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtended, shardHeaderExtendedHash)
	scsbt.AddTrackedHeader(shardHeaderExtended, shardHeaderExtendedHash)

	scsbt.CleanupHeadersBehindNonce(shardArguments.ShardCoordinator.SelfId(), 2, 2)

	lastSelfNotarizedHeader, _, _ := scsbt.GetLastSelfNotarizedHeader(header.GetShardID())
	lastCrossNotarizedHeader, _, _ := scsbt.GetLastCrossNotarizedHeader(core.SovereignChainShardId)
	trackedHeadersForSelfShard, _ := scsbt.GetTrackedHeaders(header.GetShardID())
	trackedHeadersForCrossShard, _ := scsbt.GetTrackedHeaders(core.SovereignChainShardId)

	assert.Equal(t, header, lastSelfNotarizedHeader)
	assert.Equal(t, shardHeaderExtended, lastCrossNotarizedHeader)
	assert.Zero(t, len(trackedHeadersForSelfShard))
	assert.Zero(t, len(trackedHeadersForCrossShard))
}

func TestSovereignChainShardBlockTrack_DisplayTrackedHeadersShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should not have paniced %v", r))
		}
	}()

	shardArguments := CreateSovereignChainShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

	header := &block.Header{
		ShardID: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   1,
	}
	headerHash := []byte("hash")
	scsbt.AddSelfNotarizedHeader(header.GetShardID(), header, headerHash)
	scsbt.AddTrackedHeader(header, headerHash)

	shardHeaderExtended := &block.ShardHeaderExtended{
		Header: &block.HeaderV2{
			Header: &block.Header{
				Nonce: 1,
			},
		},
	}
	shardHeaderExtendedHash := []byte("extended_hash")
	scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtended, shardHeaderExtendedHash)
	scsbt.AddTrackedHeader(shardHeaderExtended, shardHeaderExtendedHash)

	_ = logger.SetLogLevel("track:DEBUG")
	scsbt.DisplayTrackedHeaders()
}

func TestSovereignChainShardBlockTrack_GetFinalHeaderShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should error for shard header or extended shard header", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		expectedSelfNotarizerErr := errors.New("expected self notarizer err")
		scsbt.SetSelfNotarizer(
			&mock.BlockNotarizerHandlerMock{
				GetFirstNotarizedHeaderCalled: func(shardID uint32) (handler data.HeaderHandler, bytes []byte, err error) {
					return nil, nil, expectedSelfNotarizerErr
				},
			},
		)

		expectedCrossNotarizerErr := errors.New("expected cross notarizer err")
		scsbt.SetCrossNotarizer(
			&mock.BlockNotarizerHandlerMock{
				GetFirstNotarizedHeaderCalled: func(shardID uint32) (handler data.HeaderHandler, bytes []byte, err error) {
					return nil, nil, expectedCrossNotarizerErr
				},
			},
		)

		header := &block.Header{}
		finalHeader, err := scsbt.GetFinalHeader(header)

		assert.Nil(t, finalHeader)
		assert.True(t, errors.Is(err, expectedSelfNotarizerErr))

		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{},
			},
		}
		finalHeader, err = scsbt.GetFinalHeader(shardHeaderExtended)

		assert.Nil(t, finalHeader)
		assert.True(t, errors.Is(err, expectedCrossNotarizerErr))
	})

	t.Run("should work for shard header or extended shard header", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		expectedShardHeader := &block.Header{Nonce: 69}
		scsbt.SetSelfNotarizer(
			&mock.BlockNotarizerHandlerMock{
				GetFirstNotarizedHeaderCalled: func(shardID uint32) (handler data.HeaderHandler, bytes []byte, err error) {
					return expectedShardHeader, []byte("hash"), nil
				},
			},
		)

		expectedExtendedShardHeader := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: 69,
				},
			},
		}
		scsbt.SetCrossNotarizer(
			&mock.BlockNotarizerHandlerMock{
				GetFirstNotarizedHeaderCalled: func(shardID uint32) (handler data.HeaderHandler, bytes []byte, err error) {
					return expectedExtendedShardHeader, []byte("extended_hash"), nil
				},
			},
		)

		header := &block.Header{}
		finalHeader, err := scsbt.GetFinalHeader(header)

		assert.Equal(t, expectedShardHeader, finalHeader)
		assert.Nil(t, err)

		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{},
			},
		}
		finalHeader, err = scsbt.GetFinalHeader(shardHeaderExtended)

		assert.Equal(t, expectedExtendedShardHeader, finalHeader)
		assert.Nil(t, err)
	})
}

func TestSovereignChainShardBlockTrack_InitCrossNotarizedStartHeadersShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("init cross notarized start headers should return error when self start header is missing", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		scsbt.ClearStartHeaders()

		err := scsbt.InitCrossNotarizedStartHeaders()

		assert.ErrorIs(t, err, process.ErrMissingHeader)
	})

	t.Run("init cross notarized start headers should return error when wrong type assertion is occurred", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		startHeaders := make(map[uint32]data.HeaderHandler)
		startHeaders[shardArguments.ShardCoordinator.SelfId()] = &block.MetaBlock{}
		scsbt.SetStartHeaders(startHeaders)

		err := scsbt.InitCrossNotarizedStartHeaders()

		assert.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})

	t.Run("init cross notarized start headers should work", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		_ = scsbt.InitNotarizedHeaders(make(map[uint32]data.HeaderHandler))

		_, _, err := scsbt.GetLastCrossNotarizedHeader(core.SovereignChainShardId)
		assert.Equal(t, process.ErrNotarizedHeadersSliceForShardIsNil, err)

		err = scsbt.InitCrossNotarizedStartHeaders()
		assert.Nil(t, err)

		lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, err := scsbt.GetLastCrossNotarizedHeader(core.SovereignChainShardId)
		assert.Nil(t, err)

		selfStartHeader := shardArguments.StartHeaders[shardArguments.ShardCoordinator.SelfId()]
		header := selfStartHeader.(*block.Header)
		extendedSelfStartHeader := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: header,
			},
		}
		extendedSelfStartHeaderHash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, extendedSelfStartHeader)

		assert.Equal(t, extendedSelfStartHeader, lastCrossNotarizedHeader)
		assert.Equal(t, extendedSelfStartHeaderHash, lastCrossNotarizedHeaderHash)
	})
}
