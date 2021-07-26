package track_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBlockNotarizer_ShouldErrNilHasher(t *testing.T) {
	t.Parallel()

	bn, err := track.NewBlockNotarizer(nil, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	assert.Equal(t, process.ErrNilHasher, err)
	assert.True(t, check.IfNil(bn))
}

func TestNewBlockNotarizer_ShouldErrNilMarshalizer(t *testing.T) {
	t.Parallel()

	bn, err := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, nil, mock.NewMultipleShardsCoordinatorMock())

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.True(t, check.IfNil(bn))
}

func TestNewBlockNotarizer_ShouldErrNilShardCoordinator(t *testing.T) {
	t.Parallel()

	bn, err := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, nil)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.True(t, check.IfNil(bn))
}

func TestNewBlockNotarizer_ShouldWork(t *testing.T) {
	t.Parallel()

	bn, err := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	assert.Nil(t, err)
	assert.False(t, check.IfNil(bn))
}

func TestAddNotarizedHeader_ShouldNotAddNilHeader(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	bn.AddNotarizedHeader(0, nil, nil)

	assert.Equal(t, 0, len(bn.GetNotarizedHeaders()[0]))
}

func TestAddNotarizedHeader_ShouldWork(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	hdr1 := block.Header{Nonce: 1}
	hdr2 := block.Header{Nonce: 2}
	bn.AddNotarizedHeader(0, &hdr2, nil)
	bn.AddNotarizedHeader(0, &hdr1, nil)

	require.Equal(t, 2, len(bn.GetNotarizedHeaders()[0]))
	assert.Equal(t, &hdr1, bn.GetNotarizedHeaderWithIndex(0, 0))
	assert.Equal(t, &hdr2, bn.GetNotarizedHeaderWithIndex(0, 1))
}

func TestCleanupNotarizedHeadersBehindNonce_ShouldNotCleanWhenGivenNonceIsZero(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	bn.AddNotarizedHeader(0, &block.Header{}, nil)
	bn.CleanupNotarizedHeadersBehindNonce(0, 0)

	assert.Equal(t, 1, len(bn.GetNotarizedHeaders()[0]))
}

func TestCleanupNotarizedHeadersBehindNonce_ShouldNotCleanWhenGivenShardIsInvalid(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	bn.AddNotarizedHeader(0, &block.Header{}, nil)
	bn.CleanupNotarizedHeadersBehindNonce(1, 1)

	assert.Equal(t, 1, len(bn.GetNotarizedHeaders()[0]))
}

func TestCleanupNotarizedHeadersBehindNonce_ShouldKeepAtLeastOneHeader(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	bn.AddNotarizedHeader(0, &block.Header{}, nil)
	bn.CleanupNotarizedHeadersBehindNonce(0, 1)

	assert.Equal(t, 1, len(bn.GetNotarizedHeaders()[0]))
}

func TestCleanupNotarizedHeadersBehindNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	hdr1 := block.Header{Nonce: 1}
	hdr2 := block.Header{Nonce: 2}
	bn.AddNotarizedHeader(0, &hdr2, nil)
	bn.AddNotarizedHeader(0, &hdr1, nil)

	assert.Equal(t, 2, len(bn.GetNotarizedHeaders()[0]))

	bn.CleanupNotarizedHeadersBehindNonce(0, 2)
	assert.Equal(t, 1, len(bn.GetNotarizedHeaders()[0]))

	header, _, _ := bn.GetNotarizedHeader(0, 0)
	assert.Equal(t, &hdr2, header)
}

func TestNotarizedHeaders_ShouldNotPanicWhenGivenShardIsInvalid(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should not have paniced %v", r))
		}
	}()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	bn.AddNotarizedHeader(0, &block.Header{Nonce: 1}, nil)
	_ = logger.SetLogLevel("track:DEBUG")
	bn.DisplayNotarizedHeaders(1, "test")
}

func TestNotarizedHeaders_ShouldNotPanicWhenTheOnlyNotarizedHeaderHasNonceZero(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should not have paniced %v", r))
		}
	}()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	bn.AddNotarizedHeader(0, &block.Header{}, nil)
	_ = logger.SetLogLevel("track:DEBUG")
	bn.DisplayNotarizedHeaders(0, "test")
}

func TestNotarizedHeaders_ShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should not have paniced %v", r))
		}
	}()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	hdr1 := block.Header{Nonce: 1}
	hdr2 := block.Header{Nonce: 2}
	bn.AddNotarizedHeader(0, &hdr2, nil)
	bn.AddNotarizedHeader(0, &hdr1, nil)
	_ = logger.SetLogLevel("track:DEBUG")
	bn.DisplayNotarizedHeaders(0, "test")
}

func TestGetLastNotarizedHeader_ShouldErrNotarizedHeadersSliceForShardIsNil(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	_, _, err := bn.GetLastNotarizedHeader(0)

	assert.Equal(t, process.ErrNotarizedHeadersSliceForShardIsNil, err)
}

func TestGetLastNotarizedHeader_ShouldWork(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	hdr1 := block.Header{Nonce: 1}
	hash1 := []byte("hash1")
	bn.AddNotarizedHeader(0, &hdr1, hash1)
	lastNotarizedHeader, lastNotarizedHeaderHash, _ := bn.GetLastNotarizedHeader(0)

	assert.Equal(t, &hdr1, lastNotarizedHeader)
	assert.Equal(t, hash1, lastNotarizedHeaderHash)
}

func TestGetFirstNotarizedHeader_ShouldErrNotarizedHeadersSliceForShardIsNil(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	_, _, err := bn.GetFirstNotarizedHeader(0)

	assert.Equal(t, process.ErrNotarizedHeadersSliceForShardIsNil, err)
}

func TestGetFirstNotarizedHeader_ShouldWork(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	hdr1 := block.Header{Nonce: 1}
	hash1 := []byte("hash1")
	hdr2 := block.Header{Nonce: 2}
	hash2 := []byte("hash1")
	shardId := uint32(0)
	bn.AddNotarizedHeader(shardId, &hdr1, hash1)
	bn.AddNotarizedHeader(shardId, &hdr2, hash2)
	lastNotarizedHeader, lastNotarizedHeaderHash, _ := bn.GetFirstNotarizedHeader(shardId)

	assert.Equal(t, &hdr1, lastNotarizedHeader)
	assert.Equal(t, hash1, lastNotarizedHeaderHash)
}

func TestGetLastNotarizedHeaderNonce_ShouldReturnZeroWhenNotarizedHeadersForGivenShardDoesNotExist(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	lastNotarizedHeaderNonce := bn.GetLastNotarizedHeaderNonce(1)

	assert.Equal(t, uint64(0), lastNotarizedHeaderNonce)
}

func TestGetLastNotarizedHeaderNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	hdr1 := block.Header{Nonce: 3}
	bn.AddNotarizedHeader(0, &hdr1, nil)
	lastNotarizedHeaderNonce := bn.GetLastNotarizedHeaderNonce(0)

	assert.Equal(t, hdr1.Nonce, lastNotarizedHeaderNonce)
}

func TestLastNotarizedHeaderInfo_ShouldReturnNilWhenNotarizedHeadersForGivenShardDoesNotExist(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	assert.Nil(t, bn.LastNotarizedHeaderInfo(0))
}

func TestLastNotarizedHeaderInfo_ShouldWork(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	bn.AddNotarizedHeader(0, &block.Header{}, nil)

	assert.NotNil(t, bn.LastNotarizedHeaderInfo(0))
}

func TestGetNotarizedHeader_ShouldErrNotarizedHeadersSliceForShardIsNil(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	_, _, err := bn.GetNotarizedHeader(0, 0)

	assert.Equal(t, process.ErrNotarizedHeadersSliceForShardIsNil, err)
}

func TestGetNotarizedHeader_ShouldErrNotarizedHeaderOffsetIsOutOfBound(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	bn.AddNotarizedHeader(0, &block.Header{}, nil)
	_, _, err := bn.GetNotarizedHeader(0, 1)

	assert.Equal(t, track.ErrNotarizedHeaderOffsetIsOutOfBound, err)
}

func TestGetNotarizedHeader_ShouldWork(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	hdr1 := block.Header{Nonce: 1}
	hdr2 := block.Header{Nonce: 2}
	hdr3 := block.Header{Nonce: 3}
	bn.AddNotarizedHeader(0, &hdr3, nil)
	bn.AddNotarizedHeader(0, &hdr2, nil)
	bn.AddNotarizedHeader(0, &hdr1, nil)

	notarizedHeaderWithOffset, _, _ := bn.GetNotarizedHeader(0, 0)
	assert.Equal(t, &hdr3, notarizedHeaderWithOffset)

	notarizedHeaderWithOffset, _, _ = bn.GetNotarizedHeader(0, 1)
	assert.Equal(t, &hdr2, notarizedHeaderWithOffset)

	notarizedHeaderWithOffset, _, _ = bn.GetNotarizedHeader(0, 2)
	assert.Equal(t, &hdr1, notarizedHeaderWithOffset)
}

func TestInitNotarizedHeaders_ShouldErrWhenStartHeadersSliceIsNil(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	err := bn.InitNotarizedHeaders(nil)

	assert.Equal(t, process.ErrNotarizedHeadersSliceIsNil, err)
}

func TestInitNotarizedHeaders_ShouldErrWhenCalculateHashFail(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{Fail: true}, mock.NewMultipleShardsCoordinatorMock())

	startHeaders := make(map[uint32]data.HeaderHandler)
	startHeaders[0] = &block.Header{}
	err := bn.InitNotarizedHeaders(startHeaders)

	assert.NotNil(t, err)
}

func TestInitNotarizedHeaders_ShouldInitNotarizedHeaders(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	startHeaders := make(map[uint32]data.HeaderHandler)
	startHeaders[0] = &block.Header{Nonce: 1}
	err := bn.InitNotarizedHeaders(startHeaders)

	assert.Nil(t, err)

	lastNotarizedHeader, _, _ := bn.GetLastNotarizedHeader(0)

	assert.Equal(t, startHeaders[0], lastNotarizedHeader)
}

func TestRemoveLastNotarizedHeader_ShouldNotRemoveIfExistOnlyOneNotarizedHeader(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	bn.AddNotarizedHeader(0, &block.Header{}, nil)
	bn.RemoveLastNotarizedHeader()

	assert.Equal(t, 1, len(bn.GetNotarizedHeaders()[0]))
}

func TestRemoveLastNotarizedHeader_ShouldWork(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	hdr1 := block.Header{Nonce: 1}
	hdr2 := block.Header{Nonce: 2}
	hdr3 := block.Header{Nonce: 3}
	bn.AddNotarizedHeader(0, &hdr3, nil)
	bn.AddNotarizedHeader(0, &hdr2, nil)
	bn.AddNotarizedHeader(0, &hdr1, nil)

	assert.Equal(t, 3, len(bn.GetNotarizedHeaders()[0]))

	bn.RemoveLastNotarizedHeader()
	lastNotarizedHeader, _, _ := bn.GetLastNotarizedHeader(0)

	assert.Equal(t, 2, len(bn.GetNotarizedHeaders()[0]))
	assert.Equal(t, &hdr2, lastNotarizedHeader)
}

func TestRestoreNotarizedHeadersToGenesis_ShouldNotRestoreIfExistOnlyOneNotarizedHeader(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	bn.AddNotarizedHeader(0, &block.Header{}, nil)
	bn.RestoreNotarizedHeadersToGenesis()

	assert.Equal(t, 1, len(bn.GetNotarizedHeaders()[0]))
}

func TestRestoreNotarizedHeadersToGenesis_ShouldWork(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotarizer(&hashingMocks.HasherMock{}, &mock.MarshalizerMock{}, mock.NewMultipleShardsCoordinatorMock())

	hdr1 := block.Header{Nonce: 1}
	hdr2 := block.Header{Nonce: 2}
	hdr3 := block.Header{Nonce: 3}
	bn.AddNotarizedHeader(0, &hdr3, nil)
	bn.AddNotarizedHeader(0, &hdr2, nil)
	bn.AddNotarizedHeader(0, &hdr1, nil)

	assert.Equal(t, 3, len(bn.GetNotarizedHeaders()[0]))

	bn.RestoreNotarizedHeadersToGenesis()
	lastNotarizedHeader, _, _ := bn.GetLastNotarizedHeader(0)

	assert.Equal(t, 1, len(bn.GetNotarizedHeaders()[0]))
	assert.Equal(t, &hdr1, lastNotarizedHeader)
}
