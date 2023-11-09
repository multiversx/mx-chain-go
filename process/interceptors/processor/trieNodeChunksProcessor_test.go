package processor

import (
	"bytes"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

var reference = bytes.Repeat([]byte{1}, 32)

func createMockWhiteLister(isWhiteListed bool) process.WhiteListHandler {
	return &testscommon.WhiteListHandlerStub{
		IsWhiteListedAtLeastOneCalled: func(identifiers [][]byte) bool {
			return isWhiteListed
		},
	}
}

func createMockTrieNodesChunksProcessorArgs() TrieNodesChunksProcessorArgs {
	return TrieNodesChunksProcessorArgs{
		Hasher: &testscommon.HasherStub{
			SizeCalled: func() int {
				return 32
			},
		},
		ChunksCacher:    testscommon.NewCacherMock(),
		RequestInterval: time.Second,
		RequestHandler:  &testscommon.RequestHandlerStub{},
		Topic:           "topic",
	}
}

func TestNewTrieNodeChunksProcessor_NilHasher(t *testing.T) {
	t.Parallel()

	args := createMockTrieNodesChunksProcessorArgs()
	args.Hasher = nil
	tncp, err := NewTrieNodeChunksProcessor(args)
	assert.True(t, errors.Is(err, process.ErrNilHasher))
	assert.True(t, check.IfNil(tncp))
}

func TestNewTrieNodeChunksProcessor_NilCacher(t *testing.T) {
	t.Parallel()

	args := createMockTrieNodesChunksProcessorArgs()
	args.ChunksCacher = nil
	tncp, err := NewTrieNodeChunksProcessor(args)
	assert.True(t, errors.Is(err, process.ErrNilCacher))
	assert.True(t, check.IfNil(tncp))
}

func TestNewTrieNodeChunksProcessor_InvalidInterval(t *testing.T) {
	t.Parallel()

	args := createMockTrieNodesChunksProcessorArgs()
	args.RequestInterval = minimumRequestTimeInterval - time.Nanosecond
	tncp, err := NewTrieNodeChunksProcessor(args)
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
	assert.True(t, check.IfNil(tncp))
}

func TestNewTrieNodeChunksProcessor_NilRequestHandler(t *testing.T) {
	t.Parallel()

	args := createMockTrieNodesChunksProcessorArgs()
	args.RequestHandler = nil
	tncp, err := NewTrieNodeChunksProcessor(args)
	assert.True(t, errors.Is(err, process.ErrNilRequestHandler))
	assert.True(t, check.IfNil(tncp))
}

func TestNewTrieNodeChunksProcessor_EmptyTopic(t *testing.T) {
	t.Parallel()

	args := createMockTrieNodesChunksProcessorArgs()
	args.Topic = ""
	tncp, err := NewTrieNodeChunksProcessor(args)
	assert.True(t, errors.Is(err, process.ErrEmptyTopic))
	assert.True(t, check.IfNil(tncp))
}

func TestNewTrieNodeChunksProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockTrieNodesChunksProcessorArgs()
	tncp, err := NewTrieNodeChunksProcessor(args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(tncp))
	assert.Nil(t, tncp.Close())
}

func TestTrieNodeChunksProcessor_CheckBatchInvalidBatch(t *testing.T) {
	t.Parallel()

	emptyCheckedChunkResult := process.CheckedChunkResult{}

	args := createMockTrieNodesChunksProcessorArgs()
	tncp, _ := NewTrieNodeChunksProcessor(args)
	chunkResult, err := tncp.CheckBatch(
		&batch.Batch{
			Data:       make([][]byte, 1),
			Reference:  make([]byte, 32),
			ChunkIndex: 0,
			MaxChunks:  1,
		},
		createMockWhiteLister(true),
	)
	assert.Nil(t, err)
	assert.Equal(t, emptyCheckedChunkResult, chunkResult)

	chunkResult, err = tncp.CheckBatch(
		&batch.Batch{
			Data:       make([][]byte, 1),
			Reference:  nil,
			ChunkIndex: 0,
			MaxChunks:  2,
		},
		createMockWhiteLister(true),
	)
	assert.Equal(t, err, process.ErrIncompatibleReference)
	assert.Equal(t, emptyCheckedChunkResult, chunkResult)

	chunkResult, err = tncp.CheckBatch(
		&batch.Batch{
			Data:       nil,
			Reference:  make([]byte, 32),
			ChunkIndex: 0,
			MaxChunks:  2,
		},
		createMockWhiteLister(true),
	)
	assert.Nil(t, err)
	assert.Equal(t, emptyCheckedChunkResult, chunkResult)
	_ = tncp.Close()
}

func TestTrieNodeChunksProcessor_NilWhitelistHandler(t *testing.T) {
	t.Parallel()

	emptyCheckedChunkResult := process.CheckedChunkResult{}

	args := createMockTrieNodesChunksProcessorArgs()
	tncp, _ := NewTrieNodeChunksProcessor(args)
	chunkResult, err := tncp.CheckBatch(
		&batch.Batch{
			Data:       [][]byte{[]byte("buff1")},
			Reference:  reference,
			ChunkIndex: 0,
			MaxChunks:  2,
		},
		nil,
	)
	assert.Equal(t, process.ErrNilWhiteListHandler, err)
	assert.Equal(t, emptyCheckedChunkResult, chunkResult)

	_ = tncp.Close()
}

func TestTrieNodeChunksProcessor_NotWhiteListed(t *testing.T) {
	t.Parallel()

	emptyCheckedChunkResult := process.CheckedChunkResult{}

	args := createMockTrieNodesChunksProcessorArgs()
	tncp, _ := NewTrieNodeChunksProcessor(args)
	chunkResult, err := tncp.CheckBatch(
		&batch.Batch{
			Data:       [][]byte{[]byte("buff1")},
			Reference:  reference,
			ChunkIndex: 0,
			MaxChunks:  2,
		},
		createMockWhiteLister(false),
	)
	assert.Equal(t, process.ErrTrieNodeIsNotWhitelisted, err)
	assert.Equal(t, emptyCheckedChunkResult, chunkResult)

	_ = tncp.Close()
}

func TestTrieNodeChunksProcessor_CheckBatchShouldWork(t *testing.T) {
	t.Parallel()

	expectedCheckedChunkResult := process.CheckedChunkResult{
		IsChunk:        true,
		HaveAllChunks:  false,
		CompleteBuffer: nil,
	}

	args := createMockTrieNodesChunksProcessorArgs()
	tncp, _ := NewTrieNodeChunksProcessor(args)
	chunkResult, err := tncp.CheckBatch(
		&batch.Batch{
			Data:       [][]byte{[]byte("buff1")},
			Reference:  reference,
			ChunkIndex: 0,
			MaxChunks:  2,
		},
		createMockWhiteLister(true),
	)
	assert.Nil(t, err)
	assert.Equal(t, expectedCheckedChunkResult, chunkResult)
	assert.Equal(t, 1, args.ChunksCacher.Len())

	chunkResult, err = tncp.CheckBatch(
		&batch.Batch{
			Data:       [][]byte{[]byte("buff2")},
			Reference:  reference,
			ChunkIndex: 1,
			MaxChunks:  2,
		},
		createMockWhiteLister(true),
	)
	assert.Nil(t, err)

	expectedCheckedChunkResult = process.CheckedChunkResult{
		IsChunk:        true,
		HaveAllChunks:  true,
		CompleteBuffer: []byte("buff1buff2"),
	}
	assert.Equal(t, expectedCheckedChunkResult, chunkResult)
	assert.Equal(t, 0, args.ChunksCacher.Len())

	_ = tncp.Close()
}

func TestTrieNodeChunksProcessor_CheckBatchNotTheFirstBatch(t *testing.T) {
	t.Parallel()

	expectedCheckedChunkResult := process.CheckedChunkResult{
		IsChunk:        true,
		HaveAllChunks:  false,
		CompleteBuffer: nil,
	}

	args := createMockTrieNodesChunksProcessorArgs()
	tncp, _ := NewTrieNodeChunksProcessor(args)
	chunkResult, err := tncp.CheckBatch(
		&batch.Batch{
			Data:       [][]byte{[]byte("buff1")},
			Reference:  reference,
			ChunkIndex: 1,
			MaxChunks:  2,
		},
		createMockWhiteLister(true),
	)
	assert.Nil(t, err)
	assert.Equal(t, expectedCheckedChunkResult, chunkResult)
	assert.Equal(t, 0, args.ChunksCacher.Len())

	tncp.chunksCacher.Put(reference, "wrong object", 0)

	chunkResult, err = tncp.CheckBatch(
		&batch.Batch{
			Data:       [][]byte{[]byte("buff1")},
			Reference:  reference,
			ChunkIndex: 1,
			MaxChunks:  2,
		},
		createMockWhiteLister(true),
	)
	assert.Nil(t, err)
	assert.Equal(t, expectedCheckedChunkResult, chunkResult)
	assert.Equal(t, 1, args.ChunksCacher.Len())
}

func TestTrieNodeChunksProcessor_CheckBatchComponentClosed(t *testing.T) {
	t.Parallel()

	expectedCheckedChunkResult := process.CheckedChunkResult{
		IsChunk:        false,
		HaveAllChunks:  false,
		CompleteBuffer: nil,
	}

	args := createMockTrieNodesChunksProcessorArgs()
	tncp, _ := NewTrieNodeChunksProcessor(args)

	_ = tncp.Close()

	chunkResult, err := tncp.CheckBatch(
		&batch.Batch{
			Data:       [][]byte{[]byte("buff1")},
			Reference:  reference,
			ChunkIndex: 0,
			MaxChunks:  2,
		},
		createMockWhiteLister(true),
	)
	assert.Equal(t, process.ErrProcessClosed, err)
	assert.Equal(t, expectedCheckedChunkResult, chunkResult)
}

func TestTrieNodeChunksProcessor_RequestShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockTrieNodesChunksProcessorArgs()
	numRequested := uint32(0)
	args.RequestHandler = &testscommon.RequestHandlerStub{
		RequestTrieNodeCalled: func(requestHash []byte, topic string, chunkIndex uint32) {
			atomic.AddUint32(&numRequested, 1)
			assert.Equal(t, reference, requestHash)
			assert.Equal(t, args.Topic, topic)
			assert.True(t, chunkIndex >= 1 && chunkIndex <= 2)
		},
	}
	tncp, _ := NewTrieNodeChunksProcessor(args)
	_, err := tncp.CheckBatch(
		&batch.Batch{
			Data:       [][]byte{[]byte("buff1")},
			Reference:  reference,
			ChunkIndex: 0,
			MaxChunks:  3,
		},
		createMockWhiteLister(true),
	)
	assert.Nil(t, err)

	time.Sleep(time.Second + time.Millisecond*500)

	assert.Equal(t, 1, args.ChunksCacher.Len())
	assert.Equal(t, uint32(2), atomic.LoadUint32(&numRequested))

	_ = tncp.Close()
}
