package block

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	poolMock "github.com/multiversx/mx-chain-go/testscommon/pool"

	"github.com/stretchr/testify/require"
)

func TestShardInfo_IsInterfaceNil(t *testing.T) {
	t.Parallel()
	var si *ShardInfoCreateData
	require.True(t, si.IsInterfaceNil())
	si = &ShardInfoCreateData{}
	require.False(t, si.IsInterfaceNil())
}

func TestShardInfo_NewShardInfoCreateData(t *testing.T) {
	t.Parallel()

	t.Run("nil enableEpochsHandler", func(t *testing.T) {
		t.Parallel()

		args := createDefaultShardInfoCreateDataArgs()
		sicd, err := NewShardInfoCreateData(
			nil,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, sicd)
		require.Equal(t, process.ErrNilEnableEpochsHandler, err)
	})

	t.Run("nil headersPool", func(t *testing.T) {
		t.Parallel()
		args := createDefaultShardInfoCreateDataArgs()
		sicd, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			nil,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, sicd)
		require.Equal(t, process.ErrNilHeadersDataPool, err)
	})

	t.Run("nil proofsPool", func(t *testing.T) {
		t.Parallel()

		args := createDefaultShardInfoCreateDataArgs()

		sicd, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			nil,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, sicd)
		require.Equal(t, process.ErrNilProofsPool, err)
	})

	t.Run("nil pendingMiniBlocksHandler", func(t *testing.T) {
		t.Parallel()

		args := createDefaultShardInfoCreateDataArgs()

		sicd, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			nil,
			args.blockTracker,
		)
		require.Nil(t, sicd)
		require.Equal(t, process.ErrNilPendingMiniBlocksHandler, err)
	})

	t.Run("nil blockTracker", func(t *testing.T) {
		t.Parallel()

		args := createDefaultShardInfoCreateDataArgs()
		sicd, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			nil,
		)
		require.Nil(t, sicd)
		require.Equal(t, process.ErrNilBlockTracker, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createDefaultShardInfoCreateDataArgs()

		sicd, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.NotNil(t, sicd)
		require.Nil(t, err)
	})
}

func TestShardInfoCreateData_CreateShardInfoV3(t *testing.T) {
	t.Parallel()

	hdrHash0 := []byte("header hash for shard 0")
	hdrHash1 := []byte("header hash for shard 1")
	hdrHash2 := []byte("header hash for shard 2")
	hdrHash3 := []byte("header hash for shard 2 V3")

	header0 := getHeaderV3ForShard(uint32(0), hdrHash0)
	header1 := getShardHeaderForShard(uint32(1))
	header2 := getShardHeaderForShard(uint32(2))
	header3 := getHeaderV3ForShard(uint32(2), hdrHash3)
	headers := []data.HeaderHandler{header0, header1, header2, header3}

	headerHashes := [][]byte{hdrHash0, hdrHash1, hdrHash2, hdrHash3}

	pool := dataRetrieverMock.NewPoolsHolderMock()
	pool.Headers().AddHeader(hdrHash0, header0)
	pool.Headers().AddHeader(hdrHash1, header1)
	pool.Headers().AddHeader(hdrHash2, header2)
	pool.Headers().AddHeader(hdrHash3, header3)

	round := uint64(10)
	metaHdrV3 := &block.MetaBlockV3{Round: round}
	args := createDefaultShardInfoCreateDataArgs()
	sic, err := NewShardInfoCreateData(
		args.enableEpochsHandler,
		pool.Headers(),
		args.proofsPool,
		args.pendingMiniBlocksHandler,
		args.blockTracker,
	)
	require.Nil(t, err)

	t.Run("should fail with nil meta header", func(t *testing.T) {
		t.Parallel()

		shardDataProposalHandlers, shardDataHandlers, err := sic.CreateShardInfoV3(nil, headers, headerHashes)
		require.ErrorIs(t, err, process.ErrNilHeaderHandler)
		require.Nil(t, shardDataProposalHandlers)
		require.Nil(t, shardDataHandlers)
	})

	t.Run("should fail with legacy meta header", func(t *testing.T) {
		t.Parallel()

		metaHdr := &block.MetaBlock{Round: round}

		shardDataProposalHandlers, shardDataHandlers, err := sic.CreateShardInfoV3(metaHdr, headers, headerHashes)
		require.Equal(t, process.ErrInvalidHeader, err)
		require.Nil(t, shardDataProposalHandlers)
		require.Nil(t, shardDataHandlers)
	})
	t.Run("should fail with inconsistent headers and hashes", func(t *testing.T) {
		t.Parallel()

		shardDataProposalHandlers, shardDataHandlers, err := sic.CreateShardInfoV3(metaHdrV3, headers, headerHashes[:2])
		require.Equal(t, process.ErrInconsistentShardHeadersAndHashes, err)
		require.Nil(t, shardDataProposalHandlers)
		require.Nil(t, shardDataHandlers)
	})
	t.Run("should fail when createShardInfoFromHeader errors", func(t *testing.T) {
		t.Parallel()

		shardDataProposalHandlers, shardDataHandlers, err := sic.CreateShardInfoV3(metaHdrV3, headers, headerHashes)
		require.Contains(t, err.Error(), process.ErrMissingHeaderProof.Error())
		require.Nil(t, shardDataProposalHandlers)
		require.Nil(t, shardDataHandlers)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createDefaultShardInfoCreateDataArgs()
		args.pendingMiniBlocksHandler.GetPendingMiniBlocksCalled = func(shardID uint32) [][]byte {
			return [][]byte{[]byte("hash1"), []byte("hash2")}
		}
		args.blockTracker.GetLastSelfNotarizedHeaderCalled = func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.Header{Nonce: headers[shardID].GetNonce()}, []byte("selfNotarizedHash"), nil
		}
		args.enableEpochsHandler.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
			return true
		}
		args.proofsPool.HasProofCalled = func(shardID uint32, headerHash []byte) bool {
			return true
		}

		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			pool.Headers(),
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		shardDataProposalHandlers, shardDataHandlers, err := sic.CreateShardInfoV3(metaHdrV3, headers, headerHashes)
		require.Nil(t, err)
		require.NotNil(t, shardDataHandlers)
		require.NotNil(t, shardDataProposalHandlers)
		require.Equal(t, 4, len(shardDataHandlers))
		require.Equal(t, 4, len(shardDataProposalHandlers))
	})
}

func TestShardInfoCreateData_CreateShardInfoFromLegacyMeta(t *testing.T) {
	t.Parallel()

	pool := dataRetrieverMock.NewPoolsHolderMock()
	// we will have a 3 hdrs in pool
	hdrHash1 := []byte("hdr hash 1")
	hdrHash2 := []byte("hdr hash 2")
	hdrHash3 := []byte("hdr hash 3")

	mbHash1 := []byte("mb hash 1")
	mbHash2 := []byte("mb hash 2")
	mbHash3 := []byte("mb hash 3")

	miniBlockHeader1 := block.MiniBlockHeader{Hash: mbHash1}
	miniBlockHeader2 := block.MiniBlockHeader{Hash: mbHash2}
	miniBlockHeader3 := block.MiniBlockHeader{Hash: mbHash3}

	miniBlockHeaders1 := []block.MiniBlockHeader{miniBlockHeader1, miniBlockHeader2, miniBlockHeader3}
	miniBlockHeaders2 := []block.MiniBlockHeader{miniBlockHeader1, miniBlockHeader2}
	miniBlockHeaders3 := []block.MiniBlockHeader{miniBlockHeader1}

	header1 := &block.Header{
		Round:            1,
		Nonce:            45,
		ShardID:          0,
		MiniBlockHeaders: miniBlockHeaders1}
	header2 := &block.Header{
		Round:            2,
		Nonce:            45,
		ShardID:          1,
		MiniBlockHeaders: miniBlockHeaders2}
	header3 := &block.Header{
		Round:            3,
		Nonce:            45,
		ShardID:          2,
		MiniBlockHeaders: miniBlockHeaders3}
	// put the existing headers inside datapool
	pool.Headers().AddHeader(hdrHash1, header1)
	pool.Headers().AddHeader(hdrHash2, header2)
	pool.Headers().AddHeader(hdrHash3, header3)

	headerHashes := [][]byte{hdrHash1, hdrHash2, hdrHash3}
	headers := []data.ShardHeaderHandler{header1, header2, header3}
	round := uint64(10)
	metaHdr := &block.MetaBlock{Round: round}

	args := createDefaultShardInfoCreateDataArgs()
	sic, err := NewShardInfoCreateData(
		args.enableEpochsHandler,
		args.headersPool,
		args.proofsPool,
		args.pendingMiniBlocksHandler,
		args.blockTracker,
	)
	require.Nil(t, err)
	t.Run("should fail with nil meta header", func(t *testing.T) {
		t.Parallel()

		shardInfo, err := sic.CreateShardInfoFromLegacyMeta(nil, headers, headerHashes)
		require.Equal(t, process.ErrNilHeaderHandler, err)
		require.Nil(t, shardInfo)
	})
	t.Run("should fail with V3 meta header", func(t *testing.T) {
		t.Parallel()

		metaHdrV3 := &block.MetaBlockV3{Round: round}
		shardInfo, err := sic.CreateShardInfoFromLegacyMeta(metaHdrV3, headers, headerHashes)
		require.Equal(t, process.ErrInvalidHeader, err)
		require.Nil(t, shardInfo)
	})
	t.Run("should fail with inconsistent headers and hashes", func(t *testing.T) {
		t.Parallel()

		shardInfo, err := sic.CreateShardInfoFromLegacyMeta(metaHdr, headers, headerHashes[:1])
		require.Equal(t, process.ErrInconsistentShardHeadersAndHashes, err)
		require.Nil(t, shardInfo)

	})
	t.Run("should fail when createShardDataFromLegacyHeader errors", func(t *testing.T) {
		t.Parallel()

		args := createDefaultShardInfoCreateDataArgs()
		args.pendingMiniBlocksHandler.GetPendingMiniBlocksCalled = func(shardID uint32) [][]byte {
			return [][]byte{[]byte("hash1"), []byte("hash2")}
		}
		args.blockTracker.GetLastSelfNotarizedHeaderCalled = func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return nil, nil, errExpected
		}
		args.enableEpochsHandler.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
			return true
		}
		args.proofsPool.HasProofCalled = func(shardID uint32, headerHash []byte) bool {
			return true
		}
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			pool.Headers(),
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		shardInfo, err := sic.CreateShardInfoFromLegacyMeta(metaHdr, headers, headerHashes)
		require.NotNil(t, err)
		require.Nil(t, shardInfo)
		require.Equal(t, errExpected, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		args := createDefaultShardInfoCreateDataArgs()
		args.pendingMiniBlocksHandler.GetPendingMiniBlocksCalled = func(shardID uint32) [][]byte {
			return [][]byte{[]byte("hash1"), []byte("hash2")}
		}
		args.blockTracker.GetLastSelfNotarizedHeaderCalled = func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.Header{Nonce: headers[shardID].GetNonce()}, []byte("selfNotarizedHash"), nil
		}
		args.enableEpochsHandler.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
			return true
		}
		args.proofsPool.HasProofCalled = func(shardID uint32, headerHash []byte) bool {
			return true
		}
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			pool.Headers(),
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		shardInfo, err := sic.CreateShardInfoFromLegacyMeta(metaHdr, headers, headerHashes)
		require.Nil(t, err)
		require.NotNil(t, shardInfo)
		require.Equal(t, 3, len(shardInfo))
	})
}

func TestShardInfoCreateData_createShardInfoFromHeader(t *testing.T) {
	t.Parallel()

	t.Run("should fail with nil header", func(t *testing.T) {
		t.Parallel()
		args := createDefaultShardInfoCreateDataArgs()
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)

		shardDataProposalHandlers, shardDataHandlers, err := sic.createShardInfoFromHeader(nil, nil)
		require.Nil(t, shardDataProposalHandlers)
		require.Nil(t, shardDataHandlers)
		require.ErrorIs(t, err, process.ErrNilHeaderHandler)
	})
	t.Run("should fail with invalid hash", func(t *testing.T) {
		t.Parallel()
		args := createDefaultShardInfoCreateDataArgs()
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)

		shardDataProposalHandlers, shardDataHandlers, err := sic.createShardInfoFromHeader(&block.Header{}, []byte{})
		require.Nil(t, shardDataProposalHandlers)
		require.Nil(t, shardDataHandlers)
		require.ErrorIs(t, err, process.ErrInvalidHash)
	})

	t.Run("should fail with missing shard header proof", func(t *testing.T) {
		t.Parallel()
		args := createDefaultShardInfoCreateDataArgs()
		args.proofsPool.HasProofCalled = func(shardID uint32, headerHash []byte) bool {
			return false
		}
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)

		shardDataProposalHandlers, shardDataHandlers, err := sic.createShardInfoFromHeader(&block.Header{Nonce: 1}, []byte("hash"))
		require.Nil(t, shardDataProposalHandlers)
		require.Nil(t, shardDataHandlers)
		require.ErrorIs(t, err, process.ErrMissingHeaderProof)
	})
	t.Run("should work with Legacy", func(t *testing.T) {
		t.Parallel()
		header := getShardHeaderForShard(uint32(1))
		args := createDefaultShardInfoCreateDataArgs()
		args.pendingMiniBlocksHandler.GetPendingMiniBlocksCalled = func(shardID uint32) [][]byte {
			return [][]byte{[]byte("hash1"), []byte("hash2")}
		}
		args.blockTracker.GetLastSelfNotarizedHeaderCalled = func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.Header{Nonce: header.GetNonce()}, []byte("selfNotarizedHash"), nil
		}
		args.enableEpochsHandler.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
			return true
		}
		args.proofsPool.HasProofCalled = func(shardID uint32, headerHash []byte) bool {
			return true
		}
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		expectedProposalHandler := &block.ShardDataProposal{
			HeaderHash:           []byte("hash"),
			Round:                10,
			Nonce:                45,
			ShardID:              1,
			Epoch:                0,
			NumPendingMiniBlocks: 2,
		}

		shardDataProposalHandler, shardDataHandlers, err := sic.createShardInfoFromHeader(header, []byte("hash"))
		require.Nil(t, err)
		require.NotNil(t, shardDataProposalHandler)
		require.NotNil(t, shardDataHandlers)
		require.Equal(t, expectedProposalHandler, shardDataProposalHandler)
	})
	t.Run("should work with Legacy no proof for nonce < 1", func(t *testing.T) {
		t.Parallel()
		header := getShardHeaderForShard(uint32(1))
		_ = header.SetNonce(0)
		args := createDefaultShardInfoCreateDataArgs()
		args.pendingMiniBlocksHandler.GetPendingMiniBlocksCalled = func(shardID uint32) [][]byte {
			return [][]byte{[]byte("hash1"), []byte("hash2")}
		}
		args.blockTracker.GetLastSelfNotarizedHeaderCalled = func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.Header{Nonce: header.GetNonce()}, []byte("selfNotarizedHash"), nil
		}
		args.enableEpochsHandler.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
			return true
		}
		args.proofsPool.HasProofCalled = func(shardID uint32, headerHash []byte) bool {
			return false
		}
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		expectedProposalHandler := &block.ShardDataProposal{
			HeaderHash:           []byte("hash"),
			Round:                10,
			Nonce:                0,
			ShardID:              1,
			Epoch:                0,
			NumPendingMiniBlocks: 2,
		}

		shardDataProposalHandler, shardDataHandlers, err := sic.createShardInfoFromHeader(header, []byte("hash"))
		require.Nil(t, err)
		require.NotNil(t, shardDataProposalHandler)
		require.NotNil(t, shardDataHandlers)
		require.Equal(t, expectedProposalHandler, shardDataProposalHandler)
	})
	t.Run("should work with V3", func(t *testing.T) {
		t.Parallel()
		header := getHeaderV3ForShard(uint32(1), []byte("header hash for shard 1"))
		args := createDefaultShardInfoCreateDataArgs()
		args.headersPool.GetHeaderByHashCalled = func(hash []byte) (data.HeaderHandler, error) {
			return header, nil
		}

		args.pendingMiniBlocksHandler.GetPendingMiniBlocksCalled = func(shardID uint32) [][]byte {
			return [][]byte{[]byte("hash1"), []byte("hash2")}
		}
		args.blockTracker.GetLastSelfNotarizedHeaderCalled = func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.Header{Nonce: header.GetNonce()}, []byte("selfNotarizedHash"), nil
		}
		args.enableEpochsHandler.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
			return true
		}
		args.proofsPool.HasProofCalled = func(shardID uint32, headerHash []byte) bool {
			return true
		}
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		expectedProposalHandler := &block.ShardDataProposal{
			HeaderHash:           []byte("hash"),
			Round:                10,
			Nonce:                45,
			ShardID:              1,
			Epoch:                7,
			NumPendingMiniBlocks: 2,
		}

		shardDataProposalHandler, shardDataHandlers, err := sic.createShardInfoFromHeader(header, []byte("hash"))
		require.Nil(t, err)
		require.NotNil(t, shardDataProposalHandler)
		require.NotNil(t, shardDataHandlers)
		require.Equal(t, expectedProposalHandler, shardDataProposalHandler)
	})
	t.Run("should work with V3 no proof for nonce < 1", func(t *testing.T) {
		t.Parallel()
		header := getHeaderV3ForShard(uint32(1), []byte("header hash for shard 1"))
		expectedNonce := uint64(0)
		_ = header.SetNonce(expectedNonce)
		args := createDefaultShardInfoCreateDataArgs()
		args.headersPool.GetHeaderByHashCalled = func(hash []byte) (data.HeaderHandler, error) {
			return header, nil
		}
		args.pendingMiniBlocksHandler.GetPendingMiniBlocksCalled = func(shardID uint32) [][]byte {
			return [][]byte{[]byte("hash1"), []byte("hash2")}
		}
		args.blockTracker.GetLastSelfNotarizedHeaderCalled = func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.Header{Nonce: header.GetNonce()}, []byte("selfNotarizedHash"), nil
		}
		args.enableEpochsHandler.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
			return true
		}
		args.proofsPool.HasProofCalled = func(shardID uint32, headerHash []byte) bool {
			return false
		}
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		expectedProposalHandler := &block.ShardDataProposal{
			HeaderHash:           []byte("hash"),
			Round:                10,
			Nonce:                0,
			ShardID:              1,
			Epoch:                7,
			NumPendingMiniBlocks: 2,
		}

		shardDataProposalHandler, shardDataHandlers, err := sic.createShardInfoFromHeader(header, []byte("hash"))
		require.Nil(t, err)
		require.NotNil(t, shardDataProposalHandler)
		require.NotNil(t, shardDataHandlers)
		require.Equal(t, expectedProposalHandler, shardDataProposalHandler)
	})
}

func TestShardInfoCreateData_createShardDataFromLegacyHeader(t *testing.T) {
	t.Parallel()

	t.Run("should fail with updateShardDataWithCrossShardInfo error", func(t *testing.T) {
		t.Parallel()
		header := getShardHeaderForShard(uint32(1))
		args := createDefaultShardInfoCreateDataArgs()
		args.pendingMiniBlocksHandler.GetPendingMiniBlocksCalled = func(shardID uint32) [][]byte {
			return [][]byte{[]byte("hash1"), []byte("hash2")}
		}
		args.blockTracker.GetLastSelfNotarizedHeaderCalled = func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return nil, nil, fmt.Errorf("GetLastSelfNotarizedHeader error")
		}
		args.enableEpochsHandler.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
			return false
		}
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		shardDataList, err := sic.createShardDataFromLegacyHeader(header, []byte("headerHash"))
		require.Contains(t, err.Error(), "GetLastSelfNotarizedHeader error")
		require.Nil(t, shardDataList)
	})

	t.Run("should work with enable epoch flag disabled", func(t *testing.T) {
		t.Parallel()
		header := getShardHeaderForShard(uint32(1))
		args := createDefaultShardInfoCreateDataArgs()
		args.pendingMiniBlocksHandler.GetPendingMiniBlocksCalled = func(shardID uint32) [][]byte {
			return [][]byte{[]byte("hash1"), []byte("hash2")}
		}
		args.blockTracker.GetLastSelfNotarizedHeaderCalled = func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.Header{Nonce: header.GetNonce()}, []byte("selfNotarizedHash"), nil
		}
		args.enableEpochsHandler.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
			return false
		}
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		shardDataList, err := sic.createShardDataFromLegacyHeader(header, []byte("headerHash"))
		require.Nil(t, err)
		require.NotNil(t, shardDataList)
		require.Equal(t, 1, len(shardDataList))
		require.Equal(t, header.GetNonce(), shardDataList[0].GetNonce())
		require.Equal(t, uint32(0), shardDataList[0].(*block.ShardData).GetEpoch())
	})
	t.Run("should work with enable epoch flag enabled", func(t *testing.T) {
		t.Parallel()
		header := getShardHeaderForShard(uint32(1))
		args := createDefaultShardInfoCreateDataArgs()
		args.pendingMiniBlocksHandler.GetPendingMiniBlocksCalled = func(shardID uint32) [][]byte {
			return [][]byte{[]byte("hash1"), []byte("hash2")}
		}
		args.blockTracker.GetLastSelfNotarizedHeaderCalled = func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.Header{Nonce: header.GetNonce()}, []byte("selfNotarizedHash"), nil
		}
		args.enableEpochsHandler.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
			return true
		}
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		shardDataList, err := sic.createShardDataFromLegacyHeader(header, []byte("headerHash"))
		require.Nil(t, err)
		require.NotNil(t, shardDataList)
		require.Equal(t, 1, len(shardDataList))
		require.Equal(t, header.GetNonce(), shardDataList[0].GetNonce())
		require.Equal(t, header.GetEpoch(), shardDataList[0].(*block.ShardData).GetEpoch())
	})
}
func TestShardInfoCreateData_createShardDataFromV3Header(t *testing.T) {
	t.Parallel()

	t.Run("should fail with nil header", func(t *testing.T) {
		t.Parallel()
		args := createDefaultShardInfoCreateDataArgs()
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		shardDataProposalHandler, shardDataHandlers, err := sic.createShardDataFromV3Header(nil, nil)
		require.Nil(t, shardDataHandlers)
		require.Nil(t, shardDataProposalHandler)
		require.ErrorIs(t, err, process.ErrNilHeaderHandler)
	})
	t.Run("should return early if no execution results", func(t *testing.T) {
		t.Parallel()
		header := &block.HeaderV3{
			Nonce: 0,
		}
		args := createDefaultShardInfoCreateDataArgs()
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		shardDataProposalHandler, shardDataHandlers, err := sic.createShardDataFromV3Header(header, []byte("headerHash"))
		require.Nil(t, err)
		require.NotNil(t, shardDataHandlers)
		require.NotNil(t, shardDataProposalHandler)
		require.Equal(t, 0, len(shardDataHandlers))
	})
	t.Run("should fail with createShardDataFromExecutionResult error", func(t *testing.T) {
		t.Parallel()
		expectedNonce := uint64(12345)
		header := getHeaderV3ForShard(uint32(1), []byte("header hash for shard 1"))
		args := createDefaultShardInfoCreateDataArgs()
		// GetHeaderByHash error will fail createShardDataFromExecutionResult
		args.headersPool.GetHeaderByHashCalled = func(hash []byte) (data.HeaderHandler, error) {
			return nil, fmt.Errorf("GetHeaderByHash error")
		}

		args.pendingMiniBlocksHandler.GetPendingMiniBlocksCalled = func(shardID uint32) [][]byte {
			return [][]byte{[]byte("hash1"), []byte("hash2")}
		}
		args.blockTracker.GetLastSelfNotarizedHeaderCalled = func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.Header{Nonce: expectedNonce}, []byte("selfNotarizedHash"), nil
		}
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		shardDataProposalHandler, shardDataHandlers, err := sic.createShardDataFromV3Header(header, []byte("headerHash"))
		require.Nil(t, shardDataHandlers)
		require.Nil(t, shardDataProposalHandler)
		require.Equal(t, fmt.Errorf("GetHeaderByHash error"), err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		expectedNonce := uint64(12345)
		header := getHeaderV3ForShard(uint32(1), []byte("header hash for shard 1"))
		args := createDefaultShardInfoCreateDataArgs()
		args.headersPool.GetHeaderByHashCalled = func(hash []byte) (data.HeaderHandler, error) {
			return header, nil
		}
		args.pendingMiniBlocksHandler.GetPendingMiniBlocksCalled = func(shardID uint32) [][]byte {
			return [][]byte{[]byte("hash1"), []byte("hash2")}
		}
		args.blockTracker.GetLastSelfNotarizedHeaderCalled = func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.Header{Nonce: expectedNonce}, []byte("selfNotarizedHash"), nil
		}
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		shardDataProposalHandler, shardDataHandlers, err := sic.createShardDataFromV3Header(header, []byte("headerHash"))
		require.Nil(t, err)
		require.NotNil(t, shardDataHandlers)
		require.Equal(t, 1, len(shardDataHandlers))
		require.Equal(t, expectedNonce, shardDataHandlers[0].GetNonce())
		require.NotNil(t, shardDataProposalHandler)
		require.Equal(t, int(header.GetNonce()), int(shardDataProposalHandler.GetNonce()))
	})
}

func TestShardInfoCreateData_createShardDataFromExecutionResult(t *testing.T) {
	t.Parallel()

	t.Run("should fail with nil execution result", func(t *testing.T) {
		t.Parallel()
		args := createDefaultShardInfoCreateDataArgs()
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		shardData, err := sic.createShardDataFromExecutionResult(nil)
		require.ErrorIs(t, err, process.ErrNilExecutionResultHandler)
		require.Nil(t, shardData)
	})

	t.Run("should fail with wrong type of execution result", func(t *testing.T) {
		t.Parallel()
		args := createDefaultShardInfoCreateDataArgs()
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		shardData, err := sic.createShardDataFromExecutionResult(&block.BaseExecutionResult{})
		require.ErrorIs(t, err, process.ErrWrongTypeAssertion)
		require.Nil(t, shardData)
	})

	t.Run("should fail when GetHeaderByHash errors", func(t *testing.T) {
		t.Parallel()
		args := createDefaultShardInfoCreateDataArgs()
		args.headersPool.GetHeaderByHashCalled = func(hash []byte) (data.HeaderHandler, error) {
			return nil, fmt.Errorf("GetHeaderByHash error")
		}

		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		execResult := getExecutionResultForShard(uint32(1), []byte("header hash for shard 1"))
		shardData, err := sic.createShardDataFromExecutionResult(execResult)
		require.Equal(t, fmt.Errorf("GetHeaderByHash error"), err)
		require.Nil(t, shardData)
	})

	t.Run("should fail when updateShardDataWithCrossShardInfo fails", func(t *testing.T) {
		t.Parallel()
		args := createDefaultShardInfoCreateDataArgs()
		args.headersPool.GetHeaderByHashCalled = func(hash []byte) (data.HeaderHandler, error) {
			return getShardHeaderForShard(uint32(1)), nil
		}

		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		execResult := getExecutionResultForShard(uint32(1), []byte("header hash for shard 1"))
		shardData, err := sic.createShardDataFromExecutionResult(execResult)
		require.Equal(t, process.ErrNotarizedHeadersSliceIsNil, err)
		require.Nil(t, shardData)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		expectedNonce := uint64(1)
		header := getShardHeaderForShard(uint32(1))
		args := createDefaultShardInfoCreateDataArgs()
		args.headersPool.GetHeaderByHashCalled = func(hash []byte) (data.HeaderHandler, error) {
			return header, nil
		}
		args.pendingMiniBlocksHandler.GetPendingMiniBlocksCalled = func(shardID uint32) [][]byte {
			return [][]byte{[]byte("hash1"), []byte("hash2")}
		}
		args.blockTracker.GetLastSelfNotarizedHeaderCalled = func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.Header{Nonce: expectedNonce}, []byte("selfNotarizedHash"), nil
		}
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		execResult := getExecutionResultForShard(uint32(1), []byte("header hash for shard 1"))
		shardData, err := sic.createShardDataFromExecutionResult(execResult)
		require.Nil(t, err)
		require.NotNil(t, shardData)
		require.Equal(t, uint32(2), shardData.GetNumPendingMiniBlocks())
		require.Equal(t, uint32(execResult.GetExecutedTxCount()), shardData.GetTxCount())
		require.Equal(t, uint32(1), shardData.GetShardID())
		require.Equal(t, execResult.GetAccumulatedFees(), shardData.GetAccumulatedFees())
		require.Equal(t, execResult.GetHeaderHash(), shardData.GetHeaderHash())
		require.Equal(t, execResult.GetHeaderRound(), shardData.GetRound())
		require.Equal(t, header.GetPrevHash(), shardData.GetPrevHash())
		require.Equal(t, execResult.GetHeaderNonce(), shardData.GetNonce())
		require.Equal(t, header.GetPrevRandSeed(), shardData.GetPrevRandSeed())
		require.Equal(t, header.GetPubKeysBitmap(), shardData.GetPubKeysBitmap())
		require.Equal(t, execResult.GetAccumulatedFees(), shardData.GetAccumulatedFees())
		require.Equal(t, execResult.GetDeveloperFees(), shardData.GetDeveloperFees())
		require.Equal(t, header.GetEpoch(), shardData.(*block.ShardData).GetEpoch())
	})
}

func TestShardInfoCreateData_miniBlockHeaderFromMiniBlockHeader(t *testing.T) {
	t.Parallel()

	t.Run("ScheduledMiniBlocksFlag disabled", func(t *testing.T) {
		t.Parallel()
		headerHandler := getShardHeaderForShard(uint32(1))
		enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()

		miniblockHeaders := createShardMiniBlockHeaderFromHeader(headerHandler, enableEpochsHandler)
		require.NotNil(t, miniblockHeaders)
		require.Equal(t, 3, len(miniblockHeaders))
		checkMiniBlockHeadersComparable(t, headerHandler.GetMiniBlockHeaderHandlers()[0], miniblockHeaders[0])
		checkMiniBlockHeadersComparable(t, headerHandler.GetMiniBlockHeaderHandlers()[1], miniblockHeaders[1])
		checkMiniBlockHeadersComparable(t, headerHandler.GetMiniBlockHeaderHandlers()[2], miniblockHeaders[2])
	})
	t.Run("ScheduledMiniBlocksFlag enabled, all miniblocks final", func(t *testing.T) {
		t.Parallel()
		headerHandler := getShardHeaderForShard(uint32(1))
		enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
		enableEpochsHandler.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
			return flag == common.ScheduledMiniBlocksFlag
		}

		miniblockHeaders := createShardMiniBlockHeaderFromHeader(headerHandler, enableEpochsHandler)
		require.NotNil(t, miniblockHeaders)
		require.Equal(t, 3, len(miniblockHeaders))
		checkMiniBlockHeadersComparable(t, headerHandler.GetMiniBlockHeaderHandlers()[0], miniblockHeaders[0])
		checkMiniBlockHeadersComparable(t, headerHandler.GetMiniBlockHeaderHandlers()[1], miniblockHeaders[1])
		checkMiniBlockHeadersComparable(t, headerHandler.GetMiniBlockHeaderHandlers()[2], miniblockHeaders[2])
	})
	t.Run("ScheduledMiniBlocksFlag enabled, not all miniblocks final", func(t *testing.T) {
		t.Parallel()
		headerHandler := getShardHeaderForShard(uint32(1))
		enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
		enableEpochsHandler.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
			return flag == common.ScheduledMiniBlocksFlag
		}
		_ = headerHandler.GetMiniBlockHeaderHandlers()[1].SetConstructionState(int32(block.Proposed))
		require.False(t, headerHandler.GetMiniBlockHeaderHandlers()[1].IsFinal())

		miniblockHeaders := createShardMiniBlockHeaderFromHeader(headerHandler, enableEpochsHandler)
		require.NotNil(t, miniblockHeaders)
		require.Equal(t, 2, len(miniblockHeaders))
		checkMiniBlockHeadersComparable(t, headerHandler.GetMiniBlockHeaderHandlers()[0], miniblockHeaders[0])
		checkMiniBlockHeadersComparable(t, headerHandler.GetMiniBlockHeaderHandlers()[2], miniblockHeaders[1])
	})
	t.Run("ScheduledMiniBlocksFlag enabled, no final miniblocks", func(t *testing.T) {
		t.Parallel()
		headerHandler := getShardHeaderForShard(uint32(1))
		enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
		enableEpochsHandler.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
			return flag == common.ScheduledMiniBlocksFlag
		}
		for i := 0; i < len(headerHandler.GetMiniBlockHeaderHandlers()); i++ {
			_ = headerHandler.GetMiniBlockHeaderHandlers()[i].SetConstructionState(int32(block.Proposed))
			require.False(t, headerHandler.GetMiniBlockHeaderHandlers()[i].IsFinal())
		}

		miniblockHeaders := createShardMiniBlockHeaderFromHeader(headerHandler, enableEpochsHandler)
		require.NotNil(t, miniblockHeaders)
		require.Equal(t, 0, len(miniblockHeaders))
	})
}

func TestShardInfoCreateData_createShardMiniBlockHeaderFromExecutionResultHandler(t *testing.T) {
	t.Parallel()

	execResultHandler := getExecutionResultForShard(uint32(1), []byte("header hash for shard 1"))
	shardMiniBlockHeaders := createShardMiniBlockHeaderFromExecutionResultHandler(execResultHandler)
	require.NotNil(t, shardMiniBlockHeaders)
	require.Equal(t, 3, len(shardMiniBlockHeaders))
	for i := 0; i < len(shardMiniBlockHeaders); i++ {
		require.Equal(t, execResultHandler.MiniBlockHeaders[i].Hash, shardMiniBlockHeaders[i].Hash)
		require.Equal(t, execResultHandler.MiniBlockHeaders[i].Type, shardMiniBlockHeaders[i].Type)
		require.Equal(t, execResultHandler.MiniBlockHeaders[i].TxCount, shardMiniBlockHeaders[i].TxCount)
		require.Equal(t, execResultHandler.MiniBlockHeaders[i].SenderShardID, shardMiniBlockHeaders[i].SenderShardID)
		require.Equal(t, execResultHandler.MiniBlockHeaders[i].ReceiverShardID, shardMiniBlockHeaders[i].ReceiverShardID)
	}
}

// TODO modify when the function is updated
func TestShardInfoCreateData_updateShardDataWithCrossShardInfo(t *testing.T) {
	t.Parallel()
	header := block.Header{ShardID: 1}
	shardData := &block.ShardData{}

	t.Run("should fail with GetLastSelfNotarizedHeader error", func(t *testing.T) {
		t.Parallel()
		args := createDefaultShardInfoCreateDataArgs()
		args.blockTracker.GetLastSelfNotarizedHeaderCalled = func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return nil, nil, fmt.Errorf("GetLastSelfNotarizedHeader error")
		}
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		err = sic.updateShardDataWithCrossShardInfo(shardData, &header)
		require.Contains(t, err.Error(), "GetLastSelfNotarizedHeader error")
	})

	t.Run("should work with no data", func(t *testing.T) {
		t.Parallel()
		expectedNonce := uint64(12345)
		args := createDefaultShardInfoCreateDataArgs()
		args.pendingMiniBlocksHandler.GetPendingMiniBlocksCalled = func(shardID uint32) [][]byte {
			return [][]byte{[]byte("hash1"), []byte("hash2")}
		}
		args.blockTracker.GetLastSelfNotarizedHeaderCalled = func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.Header{Nonce: expectedNonce}, []byte("selfNotarizedHash"), nil
		}
		sic, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		require.Nil(t, err)
		err = sic.updateShardDataWithCrossShardInfo(shardData, &header)
		require.NotNil(t, shardData)
		require.Nil(t, err)
		require.Equal(t, uint32(2), shardData.NumPendingMiniBlocks)
		require.Equal(t, expectedNonce, shardData.LastIncludedMetaNonce)
	})
}

type shardInfoCreateDataTestArgs struct {
	headersPool              *poolMock.HeadersPoolStub
	proofsPool               *dataRetrieverMock.ProofsPoolMock
	pendingMiniBlocksHandler *mock.PendingMiniBlocksHandlerStub
	blockTracker             *mock.BlockTrackerMock
	enableEpochsHandler      *enableEpochsHandlerMock.EnableEpochsHandlerStub
}

func createDefaultShardInfoCreateDataArgs() *shardInfoCreateDataTestArgs {
	return &shardInfoCreateDataTestArgs{
		headersPool:              &poolMock.HeadersPoolStub{},
		proofsPool:               &dataRetrieverMock.ProofsPoolMock{},
		pendingMiniBlocksHandler: &mock.PendingMiniBlocksHandlerStub{},
		blockTracker:             &mock.BlockTrackerMock{},
		enableEpochsHandler:      enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
	}
}

func getMiniBlockHeadersForShard(shardID uint32) []block.MiniBlockHeader {
	mbHash1 := []byte(fmt.Sprintf("mb hash 1 for shard %d", shardID))
	mbHash2 := []byte(fmt.Sprintf("mb hash 2 for shard %d", shardID))
	mbHash3 := []byte(fmt.Sprintf("mb hash 3 for shard %d", shardID))

	miniBlockHeader1 := block.MiniBlockHeader{
		Hash:            mbHash1,
		Type:            block.TxBlock,
		TxCount:         10,
		SenderShardID:   shardID,
		ReceiverShardID: 2,
	}
	miniBlockHeader2 := block.MiniBlockHeader{
		Hash:            mbHash2,
		Type:            block.InvalidBlock,
		TxCount:         1,
		SenderShardID:   shardID,
		ReceiverShardID: 2,
	}
	miniBlockHeader3 := block.MiniBlockHeader{
		Hash:            mbHash3,
		Type:            block.SmartContractResultBlock,
		TxCount:         5,
		SenderShardID:   shardID,
		ReceiverShardID: 0,
	}

	miniBlockHeaders := []block.MiniBlockHeader{miniBlockHeader1, miniBlockHeader2, miniBlockHeader3}
	return miniBlockHeaders
}

func getShardHeaderForShard(shardID uint32) data.HeaderHandler {
	prevHash := []byte(fmt.Sprintf("prevHash for shard %d", shardID))
	prevRandSeed := []byte(fmt.Sprintf("prevRandSeed for shard %d", shardID))
	currRandSeed := []byte(fmt.Sprintf("currRandSeed for shard %d", shardID))
	return &block.HeaderV2{
		Header: &block.Header{
			Round:            10,
			Nonce:            45,
			ShardID:          shardID,
			PrevRandSeed:     prevRandSeed,
			RandSeed:         currRandSeed,
			PrevHash:         prevHash,
			MiniBlockHeaders: getMiniBlockHeadersForShard(shardID),
		},
	}
}

func getHeaderV3ForShard(shardID uint32, hash []byte) data.HeaderHandler {
	prevHash := []byte(fmt.Sprintf("prevHash for shard %d", shardID))
	prevRandSeed := []byte(fmt.Sprintf("prevRandSeed for shard %d", shardID))
	currRandSeed := []byte(fmt.Sprintf("currRandSeed for shard %d", shardID))
	return &block.HeaderV3{
		Epoch:            7,
		Round:            10,
		Nonce:            45,
		ShardID:          shardID,
		PrevRandSeed:     prevRandSeed,
		RandSeed:         currRandSeed,
		PrevHash:         prevHash,
		MiniBlockHeaders: getMiniBlockHeadersForShard(shardID),
		ExecutionResults: []*block.ExecutionResult{getExecutionResultForShard(shardID, hash)},
	}
}

func getExecutionResultForShard(shardID uint32, hash []byte) *block.ExecutionResult {
	return &block.ExecutionResult{
		ExecutedTxCount:  100,
		MiniBlockHeaders: getMiniBlockHeadersForShard(shardID),
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  hash,
			HeaderNonce: 12345,
			HeaderEpoch: 7,
			HeaderRound: 15,
			RootHash:    []byte(fmt.Sprintf("root hash for shard %d", shardID)),
			GasUsed:     5000,
		},
	}
}

// compares only the fields that are set in createShardMiniBlockHeaderFromExecutionResultHandler
func checkMiniBlockHeadersComparable(t *testing.T, mbHeader1 data.MiniBlockHeaderHandler, mbHeader2 block.MiniBlockHeader) {
	require.Equal(t, mbHeader1.GetHash(), mbHeader2.Hash)
	require.Equal(t, block.Type(mbHeader1.GetTypeInt32()), mbHeader2.Type)
	require.Equal(t, mbHeader1.GetTxCount(), mbHeader2.TxCount)
	require.Equal(t, mbHeader1.GetSenderShardID(), mbHeader2.SenderShardID)
	require.Equal(t, mbHeader1.GetReceiverShardID(), mbHeader2.ReceiverShardID)
}
