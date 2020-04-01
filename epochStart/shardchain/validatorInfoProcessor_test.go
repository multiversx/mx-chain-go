package shardchain

import (
	"bytes"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/require"
)

func createDefaultArguments() ArgValidatorInfoProcessor {
	defaultArgs := ArgValidatorInfoProcessor{
		MiniBlocksPool:               &mock.CacherStub{},
		Marshalizer:                  &mock.MarshalizerMock{},
		ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorStub{},
		Requesthandler:               &mock.RequestHandlerStub{},
		Hasher:                       &mock.HasherMock{},
	}

	return defaultArgs
}

func TestNewValidatorInfoProcessor_NilValidatorStatisticsProcessorShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultArguments()
	args.ValidatorStatisticsProcessor = nil
	validatorInfoProcessor, err := NewValidatorInfoProcessor(args)

	require.Nil(t, validatorInfoProcessor)
	require.Equal(t, epochStart.ErrNilValidatorStatistics, err)
}

func TestNewValidatorInfoProcessor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultArguments()
	args.Marshalizer = nil
	validatorInfoProcessor, err := NewValidatorInfoProcessor(args)

	require.Nil(t, validatorInfoProcessor)
	require.Equal(t, epochStart.ErrNilMarshalizer, err)
}

func TestNewValidatorInfoProcessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultArguments()
	args.Hasher = nil
	validatorInfoProcessor, err := NewValidatorInfoProcessor(args)

	require.Nil(t, validatorInfoProcessor)
	require.Equal(t, epochStart.ErrNilHasher, err)
}

func TestNewValidatorInfoProcessor_NilMiniBlocksPoolErr(t *testing.T) {
	t.Parallel()

	args := createDefaultArguments()
	args.MiniBlocksPool = nil
	validatorInfoProcessor, err := NewValidatorInfoProcessor(args)

	require.Nil(t, validatorInfoProcessor)
	require.Equal(t, epochStart.ErrNilMiniBlockPool, err)
}

func TestNewValidatorInfoProcessor_NilRequestHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultArguments()
	args.Requesthandler = nil
	validatorInfoProcessor, err := NewValidatorInfoProcessor(args)

	require.Nil(t, validatorInfoProcessor)
	require.Equal(t, epochStart.ErrNilRequestHandler, err)
}

func TestValidatorInfoProcessor_IsInterfaceNil(t *testing.T) {
	args := createDefaultArguments()
	args.MiniBlocksPool = &mock.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte)) {
		},
	}

	validatorInfoProcessor, err := NewValidatorInfoProcessor(args)

	require.Nil(t, err)
	require.False(t, check.IfNil(validatorInfoProcessor))
}

func TestValidatorInfoProcessor_ShouldWork(t *testing.T) {
	args := createDefaultArguments()
	args.MiniBlocksPool = &mock.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte)) {
		},
	}

	validatorInfoProcessor, err := NewValidatorInfoProcessor(args)

	require.Nil(t, err)
	require.NotNil(t, validatorInfoProcessor)
}

func TestValidatorInfoProcessor_ProcessMetaBlockThatIsNoStartOfEpochShouldWork(t *testing.T) {
	args := createDefaultArguments()
	args.MiniBlocksPool = &mock.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte)) {
		},
	}

	previousHeader99 := &block.MetaBlock{Nonce: 99, Epoch: 0}
	previousHeaderHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, previousHeader99)

	validatorInfoProcessor, _ := NewValidatorInfoProcessor(args)
	_, processError := validatorInfoProcessor.ProcessMetaBlock(previousHeader99, previousHeaderHash)

	require.Nil(t, processError)
}

func TestValidatorInfoProcessor_ProcesStartOfEpochWithNoResolvedPeerMiniblocksShouldErr(t *testing.T) {
	args := createDefaultArguments()
	args.MiniBlocksPool = &mock.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte)) {
		},
	}

	previousHeader99 := &block.MetaBlock{Nonce: 99, Epoch: 0}
	previousHeaderHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, previousHeader99)

	validatorInfoProcessor, _ := NewValidatorInfoProcessor(args)
	_, processError := validatorInfoProcessor.ProcessMetaBlock(previousHeader99, previousHeaderHash)

	require.Nil(t, processError)
}

func TestValidatorInfoProcessor_ProcesStartOfEpochWithNoPeerMiniblocksShouldWork(t *testing.T) {
	args := createDefaultArguments()
	args.MiniBlocksPool = &mock.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte)) {
		},
	}

	hash := []byte("hash")
	previousHash := []byte("prevHash")

	peerMiniblock := &block.MiniBlock{
		TxHashes:        [][]byte{},
		ReceiverShardID: core.AllShardId,
		SenderShardID:   core.MetachainShardId,
		Type:            block.RewardsBlock,
	}

	peerMiniBlockHash, _ := args.Marshalizer.Marshal(peerMiniblock)

	miniBlockHeader := block.MiniBlockHeader{
		Hash: peerMiniBlockHash, Type: block.RewardsBlock, SenderShardID: core.MetachainShardId, ReceiverShardID: core.AllShardId, TxCount: 1}

	epochStartHeader := &block.MetaBlock{Nonce: 100, Epoch: 1, PrevHash: previousHash}
	epochStartHeader.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0, RootHash: hash, HeaderHash: hash}}
	epochStartHeader.MiniBlockHeaders = []block.MiniBlockHeader{miniBlockHeader}
	epochStartHeaderHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, epochStartHeader)

	peekCalled := false
	args.MiniBlocksPool = &mock.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte)) {

		},
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			peekCalled = true
			return nil, false
		},
	}

	processCalled := false
	args.ValidatorStatisticsProcessor = &mock.ValidatorStatisticsProcessorStub{
		ProcessCalled: func(info data.ShardValidatorInfoHandler) error {
			processCalled = true

			return nil
		},
	}

	validatorInfoProcessor, _ := NewValidatorInfoProcessor(args)
	_, processError := validatorInfoProcessor.ProcessMetaBlock(epochStartHeader, epochStartHeaderHash)

	require.Nil(t, processError)
	require.False(t, processCalled)
	require.False(t, peekCalled)
}

func TestValidatorInfoProcessor_ProcesStartOfEpochWithPeerMiniblocksInPoolShouldProcess(t *testing.T) {
	args := createDefaultArguments()

	hash := []byte("hash")
	previousHash := []byte("prevHash")

	pk := []byte("publicKey")
	list := "eligible"
	tempRating := uint32(53)
	rating := uint32(51)
	vi := state.ValidatorInfo{
		PublicKey:                  pk,
		ShardId:                    0,
		List:                       list,
		Index:                      0,
		TempRating:                 tempRating,
		Rating:                     rating,
		RewardAddress:              nil,
		LeaderSuccess:              0,
		LeaderFailure:              0,
		ValidatorSuccess:           0,
		ValidatorFailure:           0,
		NumSelectedInSuccessBlocks: 0,
		AccumulatedFees:            nil,
	}

	marshalizedVi, _ := args.Marshalizer.Marshal(vi)

	peerMiniblock := &block.MiniBlock{
		TxHashes:        [][]byte{marshalizedVi},
		ReceiverShardID: core.AllShardId,
		SenderShardID:   core.MetachainShardId,
		Type:            block.PeerBlock,
	}

	peerMiniBlockHash, _ := args.Marshalizer.Marshal(peerMiniblock)

	miniBlockHeader := block.MiniBlockHeader{
		Hash: peerMiniBlockHash, Type: block.PeerBlock, SenderShardID: core.MetachainShardId, ReceiverShardID: core.AllShardId, TxCount: 1}

	epochStartHeader := &block.MetaBlock{Nonce: 100, Epoch: 1, PrevHash: previousHash}
	epochStartHeader.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0, RootHash: hash, HeaderHash: hash}}
	epochStartHeader.MiniBlockHeaders = []block.MiniBlockHeader{miniBlockHeader}
	epochStartHeaderHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, epochStartHeader)

	args.MiniBlocksPool = &mock.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte)) {

		},
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(key, peerMiniBlockHash) {
				return peerMiniblock, true
			}
			return nil, false
		},
	}

	processCalled := false
	args.ValidatorStatisticsProcessor = &mock.ValidatorStatisticsProcessorStub{
		ProcessCalled: func(info data.ShardValidatorInfoHandler) error {
			processCalled = true

			require.Equal(t, pk, info.GetPublicKey())
			require.Equal(t, tempRating, info.GetTempRating())

			return nil
		},
	}

	validatorInfoProcessor, _ := NewValidatorInfoProcessor(args)

	_, processError := validatorInfoProcessor.ProcessMetaBlock(epochStartHeader, epochStartHeaderHash)

	require.Nil(t, processError)
	require.True(t, processCalled)
}

func TestValidatorInfoProcessor_ProcesStartOfEpochWithMissinPeerMiniblocksShouldProcess(t *testing.T) {
	args := createDefaultArguments()

	hash := []byte("hash")
	previousHash := []byte("prevHash")

	pk := []byte("publicKey")
	list := "eligible"
	tempRating := uint32(53)
	rating := uint32(51)
	vi := state.ValidatorInfo{
		PublicKey:                  pk,
		ShardId:                    0,
		List:                       list,
		Index:                      0,
		TempRating:                 tempRating,
		Rating:                     rating,
		RewardAddress:              nil,
		LeaderSuccess:              0,
		LeaderFailure:              0,
		ValidatorSuccess:           0,
		ValidatorFailure:           0,
		NumSelectedInSuccessBlocks: 0,
		AccumulatedFees:            nil,
	}

	marshalizedVi, _ := args.Marshalizer.Marshal(vi)

	peerMiniblock := &block.MiniBlock{
		TxHashes:        [][]byte{marshalizedVi},
		ReceiverShardID: core.AllShardId,
		SenderShardID:   core.MetachainShardId,
		Type:            block.PeerBlock,
	}

	peerMiniBlockHash, _ := args.Marshalizer.Marshal(peerMiniblock)

	miniBlockHeader := block.MiniBlockHeader{
		Hash: peerMiniBlockHash, Type: block.PeerBlock, SenderShardID: core.MetachainShardId, ReceiverShardID: core.AllShardId, TxCount: 1}

	epochStartHeader := &block.MetaBlock{Nonce: 100, Epoch: 1, PrevHash: previousHash}
	epochStartHeader.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0, RootHash: hash, HeaderHash: hash}}
	epochStartHeader.MiniBlockHeaders = []block.MiniBlockHeader{miniBlockHeader}
	epochStartHeaderHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, epochStartHeader)

	var receivedMiniblock func(key []byte)
	args.MiniBlocksPool = &mock.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte)) {
			receivedMiniblock = f
		},
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(key, peerMiniBlockHash) {
				return peerMiniblock, true
			}
			return nil, false
		},
		PutCalled: func(key []byte, value interface{}) (evicted bool) {
			if bytes.Equal(key, peerMiniBlockHash) {
				receivedMiniblock(key)
				return false
			}
			return false
		},
	}

	processCalled := false
	args.ValidatorStatisticsProcessor = &mock.ValidatorStatisticsProcessorStub{
		ProcessCalled: func(info data.ShardValidatorInfoHandler) error {
			processCalled = true

			require.Equal(t, pk, info.GetPublicKey())
			require.Equal(t, tempRating, info.GetTempRating())

			return nil
		},
	}

	args.Requesthandler = &mock.RequestHandlerStub{
		RequestMiniBlocksHandlerCalled: func(destShardID uint32, miniblockHashes [][]byte) {
			if destShardID == core.MetachainShardId &&
				bytes.Equal(miniblockHashes[0], peerMiniBlockHash) {
				args.MiniBlocksPool.Put(peerMiniBlockHash, miniBlockHeader)
			}
		},
	}

	validatorInfoProcessor, _ := NewValidatorInfoProcessor(args)

	_, processError := validatorInfoProcessor.ProcessMetaBlock(epochStartHeader, epochStartHeaderHash)

	require.Nil(t, processError)
	require.True(t, processCalled)
}

func TestValidatorInfoProcessor_ProcesStartOfEpochWithMissinPeerMiniblocksTimeoutShouldErr(t *testing.T) {
	args := createDefaultArguments()

	hash := []byte("hash")
	previousHash := []byte("prevHash")

	vi := state.ValidatorInfo{}

	marshalizedVi, _ := args.Marshalizer.Marshal(vi)

	peerMiniblock := &block.MiniBlock{
		TxHashes:        [][]byte{marshalizedVi},
		ReceiverShardID: core.AllShardId,
		SenderShardID:   core.MetachainShardId,
		Type:            block.PeerBlock,
	}

	peerMiniBlockHash, _ := args.Marshalizer.Marshal(peerMiniblock)

	miniBlockHeader := block.MiniBlockHeader{
		Hash: peerMiniBlockHash, Type: block.PeerBlock, SenderShardID: core.MetachainShardId, ReceiverShardID: core.AllShardId, TxCount: 1}

	epochStartHeader := &block.MetaBlock{Nonce: 100, Epoch: 1, PrevHash: previousHash}
	epochStartHeader.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0, RootHash: hash, HeaderHash: hash}}
	epochStartHeader.MiniBlockHeaders = []block.MiniBlockHeader{miniBlockHeader}
	epochStartHeaderHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, epochStartHeader)

	var receivedMiniblock func(key []byte)
	args.MiniBlocksPool = &mock.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte)) {
			receivedMiniblock = f
		},
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(key, peerMiniBlockHash) {
				return peerMiniblock, true
			}
			return nil, false
		},
		PutCalled: func(key []byte, value interface{}) (evicted bool) {
			if bytes.Equal(key, peerMiniBlockHash) {
				receivedMiniblock(key)
				return false
			}
			return false
		},
	}

	args.ValidatorStatisticsProcessor = &mock.ValidatorStatisticsProcessorStub{
		ProcessCalled: func(info data.ShardValidatorInfoHandler) error {
			return nil
		},
	}

	args.Requesthandler = &mock.RequestHandlerStub{
		RequestMiniBlocksHandlerCalled: func(destShardID uint32, miniblockHashes [][]byte) {
			if destShardID == core.MetachainShardId &&
				bytes.Equal(miniblockHashes[0], peerMiniBlockHash) {
				time.Sleep(5100 * time.Millisecond)
				args.MiniBlocksPool.Put(peerMiniBlockHash, miniBlockHeader)
			}
		},
	}

	validatorInfoProcessor, _ := NewValidatorInfoProcessor(args)

	_, processError := validatorInfoProcessor.ProcessMetaBlock(epochStartHeader, epochStartHeaderHash)

	require.Equal(t, process.ErrTimeIsOut, processError)
}
