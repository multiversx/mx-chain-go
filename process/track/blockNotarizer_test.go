package track_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/stretchr/testify/assert"
)

func TestNewBlockNotarizer_ShouldErrNilHasher(t *testing.T) {
	bn, err := track.NewBlockNotarizer(nil, &mock.MarshalizerMock{})
	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, bn)
}

func TestNewBlockNotarizer_ShouldErrNilMarshalizer(t *testing.T) {
	bn, err := track.NewBlockNotarizer(&mock.HasherMock{}, nil)
	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, bn)
}

func TestNewBlockNotarizer_ShouldWork(t *testing.T) {
	bn, err := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})
	assert.Nil(t, err)
	assert.NotNil(t, bn)
}

func TestAddNotarizedHeader_ShouldNotAddNilHeader(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})
	bn.AddNotarizedHeader(0, nil, nil)
	assert.Equal(t, 0, len(bn.GetNotarizedHeaders()[0]))
}

func TestAddNotarizedHeader_ShouldWork(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})

	hdr1 := block.Header{Nonce: 1}
	hdr2 := block.Header{Nonce: 2}
	bn.AddNotarizedHeader(0, &hdr2, nil)
	bn.AddNotarizedHeader(0, &hdr1, nil)

	assert.Equal(t, 2, len(bn.GetNotarizedHeaders()[0]))
	assert.Equal(t, &hdr1, bn.GetNotarizedHeader(0, 0))
	assert.Equal(t, &hdr2, bn.GetNotarizedHeader(0, 1))
}

func TestCleanupNotarizedHeadersBehindNonce_ShouldNotCleanWhenGivenNonceIsZero(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})
	bn.AddNotarizedHeader(0, &block.Header{}, nil)
	bn.CleanupNotarizedHeadersBehindNonce(0, 0)
	assert.Equal(t, 1, len(bn.GetNotarizedHeaders()[0]))
}

func TestCleanupNotarizedHeadersBehindNonce_ShouldNotCleanWhenGivenShardIsInvalid(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})
	bn.AddNotarizedHeader(0, &block.Header{}, nil)
	bn.CleanupNotarizedHeadersBehindNonce(1, 1)
	assert.Equal(t, 1, len(bn.GetNotarizedHeaders()[0]))
}

func TestCleanupNotarizedHeadersBehindNonce_ShouldWork(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})

	hdr1 := block.Header{Nonce: 1}
	hdr2 := block.Header{Nonce: 2}
	bn.AddNotarizedHeader(0, &hdr2, nil)
	bn.AddNotarizedHeader(0, &hdr1, nil)

	assert.Equal(t, 2, len(bn.GetNotarizedHeaders()[0]))

	bn.CleanupNotarizedHeadersBehindNonce(0, 2)

	assert.Equal(t, 1, len(bn.GetNotarizedHeaders()[0]))
	assert.Equal(t, &hdr2, bn.GetNotarizedHeader(0, 0))
}

func TestNotarizedHeaders_ShouldNotDisplayWhenGivenShardIsInvalid(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})

	bn.AddNotarizedHeader(0, &block.Header{Nonce: 1}, nil)
	logger.SetLogLevel("track:TRACE")
	bn.DisplayNotarizedHeaders(1, "test")
}

func TestNotarizedHeaders_ShouldNotDisplayWhenTheOnlyNotarizedHeaderHasNonceZero(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})

	bn.AddNotarizedHeader(0, &block.Header{}, nil)
	logger.SetLogLevel("track:TRACE")
	bn.DisplayNotarizedHeaders(0, "test")
}

func TestNotarizedHeaders_ShouldWork(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})

	hdr1 := block.Header{Nonce: 1}
	hdr2 := block.Header{Nonce: 2}
	bn.AddNotarizedHeader(0, &hdr2, nil)
	bn.AddNotarizedHeader(0, &hdr1, nil)
	logger.SetLogLevel("track:TRACE")
	bn.DisplayNotarizedHeaders(0, "test")
}

func TestGetLastNotarizedHeader_ShouldErrNotarizedHeadersSliceForShardIsNil(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})
	_, _, err := bn.GetLastNotarizedHeader(0)
	assert.Equal(t, process.ErrNotarizedHeadersSliceForShardIsNil, err)
}

func TestGetLastNotarizedHeader_ShouldWork(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})

	hdr1 := block.Header{Nonce: 1}
	hash1 := []byte("hash1")
	bn.AddNotarizedHeader(0, &hdr1, hash1)
	lastNotarizedHeader, lastNotarizedHeaderHash, _ := bn.GetLastNotarizedHeader(0)

	assert.Equal(t, &hdr1, lastNotarizedHeader)
	assert.Equal(t, hash1, lastNotarizedHeaderHash)
}

func TestGetLastNotarizedHeaderNonce_ShouldReturnZeroWhenNotarizedHeadersForGivenShardDoesNotExist(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})
	lastNotarizedHeaderNonce := bn.GetLastNotarizedHeaderNonce(1)

	assert.Equal(t, uint64(0), lastNotarizedHeaderNonce)
}

func TestGetLastNotarizedHeaderNonce_ShouldWork(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})

	hdr1 := block.Header{Nonce: 3}
	bn.AddNotarizedHeader(0, &hdr1, nil)
	lastNotarizedHeaderNonce := bn.GetLastNotarizedHeaderNonce(0)

	assert.Equal(t, hdr1.Nonce, lastNotarizedHeaderNonce)
}

func TestLastNotarizedHeaderInfo_ShouldReturnNilWhenNotarizedHeadersForGivenShardDoesNotExist(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})
	assert.Nil(t, bn.LastNotarizedHeaderInfo(0))
}

func TestLastNotarizedHeaderInfo_ShouldWork(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})
	bn.AddNotarizedHeader(0, &block.Header{}, nil)
	assert.NotNil(t, bn.LastNotarizedHeaderInfo(0))
}

func TestGetNotarizedHeader_ShouldErrNotarizedHeadersSliceForShardIsNil(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})
	_, _, err := bn.GetNotarizedHeaderWithOffset(0, 0)
	assert.Equal(t, process.ErrNotarizedHeadersSliceForShardIsNil, err)
}

func TestGetNotarizedHeader_ShouldErrNotarizedHeaderOffsetIsOutOfBound(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})
	bn.AddNotarizedHeader(0, &block.Header{}, nil)
	_, _, err := bn.GetNotarizedHeaderWithOffset(0, 1)
	assert.Equal(t, track.ErrNotarizedHeaderOffsetIsOutOfBound, err)
}

func TestGetNotarizedHeader_ShouldWork(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})

	hdr1 := block.Header{Nonce: 1}
	hdr2 := block.Header{Nonce: 2}
	hdr3 := block.Header{Nonce: 3}
	bn.AddNotarizedHeader(0, &hdr3, nil)
	bn.AddNotarizedHeader(0, &hdr2, nil)
	bn.AddNotarizedHeader(0, &hdr1, nil)

	notarizedHeaderWithOffset, _, _ := bn.GetNotarizedHeaderWithOffset(0, 0)
	assert.Equal(t, &hdr3, notarizedHeaderWithOffset)

	notarizedHeaderWithOffset, _, _ = bn.GetNotarizedHeaderWithOffset(0, 1)
	assert.Equal(t, &hdr2, notarizedHeaderWithOffset)

	notarizedHeaderWithOffset, _, _ = bn.GetNotarizedHeaderWithOffset(0, 2)
	assert.Equal(t, &hdr1, notarizedHeaderWithOffset)
}

func TestInitNotarizedHeaders_ShouldErrWhenStartHeadersSliceIsNil(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})
	err := bn.InitNotarizedHeaders(nil)
	assert.Equal(t, process.ErrNotarizedHeadersSliceIsNil, err)
}

func TestInitNotarizedHeaders_ShouldErrWhenCalculateHashFail(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{
		Fail: true,
	})

	startHeaders := make(map[uint32]data.HeaderHandler)
	startHeaders[0] = &block.Header{}

	err := bn.InitNotarizedHeaders(startHeaders)
	assert.NotNil(t, err)
}

func TestInitNotarizedHeaders_ShouldInitNotarizedHeaders(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})

	startHeaders := make(map[uint32]data.HeaderHandler)
	startHeaders[0] = &block.Header{Nonce: 1}

	err := bn.InitNotarizedHeaders(startHeaders)

	assert.Nil(t, err)

	lastNotarizedHeader, _, _ := bn.GetLastNotarizedHeader(0)
	assert.Equal(t, startHeaders[0], lastNotarizedHeader)
}

func TestRemoveLastNotarizedHeader_ShouldNotRemoveIfExistOnlyOneNotarizedHeader(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})
	bn.AddNotarizedHeader(0, &block.Header{}, nil)
	bn.RemoveLastNotarizedHeader()
	assert.Equal(t, 1, len(bn.GetNotarizedHeaders()[0]))
}

func TestRemoveLastNotarizedHeader_ShouldWork(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})

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
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})
	bn.AddNotarizedHeader(0, &block.Header{}, nil)
	bn.RestoreNotarizedHeadersToGenesis()
	assert.Equal(t, 1, len(bn.GetNotarizedHeaders()[0]))
}

func TestRestoreNotarizedHeadersToGenesis_ShouldWork(t *testing.T) {
	bn, _ := track.NewBlockNotarizer(&mock.HasherMock{}, &mock.MarshalizerMock{})

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
