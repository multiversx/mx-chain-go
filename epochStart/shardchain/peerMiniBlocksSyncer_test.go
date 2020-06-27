package shardchain

import (
	"bytes"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func createDefaultArguments() ArgPeerMiniBlockSyncer {
	defaultArgs := ArgPeerMiniBlockSyncer{
		MiniBlocksPool: testscommon.NewCacherStub(),
		Requesthandler: &mock.RequestHandlerStub{},
	}

	return defaultArgs
}

func TestNewValidatorInfoProcessor_NilMiniBlocksPoolErr(t *testing.T) {
	t.Parallel()

	args := createDefaultArguments()
	args.MiniBlocksPool = nil
	syncer, err := NewPeerMiniBlockSyncer(args)

	require.Nil(t, syncer)
	require.Equal(t, epochStart.ErrNilMiniBlockPool, err)
}

func TestNewValidatorInfoProcessor_NilRequestHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultArguments()
	args.Requesthandler = nil
	syncer, err := NewPeerMiniBlockSyncer(args)

	require.Nil(t, syncer)
	require.Equal(t, epochStart.ErrNilRequestHandler, err)
}

func TestValidatorInfoProcessor_IsInterfaceNil(t *testing.T) {
	args := createDefaultArguments()
	args.MiniBlocksPool = &testscommon.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte, value interface{})) {
		},
	}

	syncer, err := NewPeerMiniBlockSyncer(args)

	require.Nil(t, err)
	require.False(t, check.IfNil(syncer))
}

func TestValidatorInfoProcessor_ShouldWork(t *testing.T) {
	args := createDefaultArguments()
	args.MiniBlocksPool = &testscommon.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte, value interface{})) {
		},
	}

	syncer, err := NewPeerMiniBlockSyncer(args)

	require.Nil(t, err)
	require.NotNil(t, syncer)
}

func TestValidatorInfoProcessor_ProcessMetaBlockThatIsNoStartOfEpochShouldWork(t *testing.T) {
	args := createDefaultArguments()
	args.MiniBlocksPool = &testscommon.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte, value interface{})) {
		},
	}

	previousHeader99 := &block.MetaBlock{Nonce: 99, Epoch: 0}

	syncer, _ := NewPeerMiniBlockSyncer(args)
	_, _, processError := syncer.SyncMiniBlocks(previousHeader99)

	require.Nil(t, processError)
}

func TestValidatorInfoProcessor_ProcesStartOfEpochWithNoPeerMiniblocksShouldWork(t *testing.T) {
	args := createDefaultArguments()
	args.MiniBlocksPool = &testscommon.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte, value interface{})) {
		},
	}

	hash := []byte("hash")
	previousHash := []byte("prevHash")

	miniBlockHeader := block.MiniBlockHeader{
		Hash: []byte("hash"), Type: block.RewardsBlock, SenderShardID: core.MetachainShardId, ReceiverShardID: core.AllShardId, TxCount: 1}

	epochStartHeader := &block.MetaBlock{Nonce: 100, Epoch: 1, PrevHash: previousHash}
	epochStartHeader.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0, RootHash: hash, HeaderHash: hash}}
	epochStartHeader.MiniBlockHeaders = []block.MiniBlockHeader{miniBlockHeader}

	peekCalled := false
	args.MiniBlocksPool = &testscommon.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte, value interface{})) {

		},
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			peekCalled = true
			return nil, false
		},
	}

	syncer, _ := NewPeerMiniBlockSyncer(args)
	_, _, processError := syncer.SyncMiniBlocks(epochStartHeader)

	require.Nil(t, processError)
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

	marshalizer := &mock.MarshalizerMock{}
	marshalizedVi, _ := marshalizer.Marshal(vi)

	peerMiniblock := &block.MiniBlock{
		TxHashes:        [][]byte{marshalizedVi},
		ReceiverShardID: core.AllShardId,
		SenderShardID:   core.MetachainShardId,
		Type:            block.PeerBlock,
	}

	peerMiniBlockHash, _ := marshalizer.Marshal(peerMiniblock)

	miniBlockHeader := block.MiniBlockHeader{
		Hash: peerMiniBlockHash, Type: block.PeerBlock, SenderShardID: core.MetachainShardId, ReceiverShardID: core.AllShardId, TxCount: 1}

	epochStartHeader := &block.MetaBlock{Nonce: 100, Epoch: 1, PrevHash: previousHash}
	epochStartHeader.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0, RootHash: hash, HeaderHash: hash}}
	epochStartHeader.MiniBlockHeaders = []block.MiniBlockHeader{miniBlockHeader}

	args.MiniBlocksPool = &testscommon.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte, value interface{})) {

		},
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(key, peerMiniBlockHash) {
				return peerMiniblock, true
			}
			return nil, false
		},
	}

	syncer, _ := NewPeerMiniBlockSyncer(args)
	_, _, processError := syncer.SyncMiniBlocks(epochStartHeader)

	require.Nil(t, processError)
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
	marshalizer := &mock.MarshalizerMock{}
	marshalizedVi, _ := marshalizer.Marshal(vi)

	peerMiniblock := &block.MiniBlock{
		TxHashes:        [][]byte{marshalizedVi},
		ReceiverShardID: core.AllShardId,
		SenderShardID:   core.MetachainShardId,
		Type:            block.PeerBlock,
	}

	peerMiniBlockHash, _ := marshalizer.Marshal(peerMiniblock)

	miniBlockHeader := block.MiniBlockHeader{
		Hash: peerMiniBlockHash, Type: block.PeerBlock, SenderShardID: core.MetachainShardId, ReceiverShardID: core.AllShardId, TxCount: 1}

	epochStartHeader := &block.MetaBlock{Nonce: 100, Epoch: 1, PrevHash: previousHash}
	epochStartHeader.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0, RootHash: hash, HeaderHash: hash}}
	epochStartHeader.MiniBlockHeaders = []block.MiniBlockHeader{miniBlockHeader}

	var receivedMiniblock func(key []byte, value interface{})
	args.MiniBlocksPool = &testscommon.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte, value interface{})) {
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
		PutCalled: func(key []byte, value interface{}, _ int) (evicted bool) {
			if bytes.Equal(key, peerMiniBlockHash) {
				receivedMiniblock(key, value)
				return false
			}
			return false
		},
	}

	args.Requesthandler = &mock.RequestHandlerStub{
		RequestMiniBlocksHandlerCalled: func(destShardID uint32, miniblockHashes [][]byte) {
			if destShardID == core.MetachainShardId &&
				bytes.Equal(miniblockHashes[0], peerMiniBlockHash) {
				args.MiniBlocksPool.Put(peerMiniBlockHash, peerMiniblock, peerMiniblock.Size())
			}
		},
	}

	syncer, _ := NewPeerMiniBlockSyncer(args)
	_, _, processError := syncer.SyncMiniBlocks(epochStartHeader)

	require.Nil(t, processError)
}

func TestValidatorInfoProcessor_ProcesStartOfEpochWithMissinPeerMiniblocksTimeoutShouldErr(t *testing.T) {
	args := createDefaultArguments()

	hash := []byte("hash")
	previousHash := []byte("prevHash")

	vi := state.ValidatorInfo{}
	marshalizer := &mock.MarshalizerMock{}
	marshalizedVi, _ := marshalizer.Marshal(vi)

	peerMiniblock := &block.MiniBlock{
		TxHashes:        [][]byte{marshalizedVi},
		ReceiverShardID: core.AllShardId,
		SenderShardID:   core.MetachainShardId,
		Type:            block.PeerBlock,
	}

	peerMiniBlockHash, _ := marshalizer.Marshal(peerMiniblock)

	miniBlockHeader := block.MiniBlockHeader{
		Hash: peerMiniBlockHash, Type: block.PeerBlock, SenderShardID: core.MetachainShardId, ReceiverShardID: core.AllShardId, TxCount: 1}

	epochStartHeader := &block.MetaBlock{Nonce: 100, Epoch: 1, PrevHash: previousHash}
	epochStartHeader.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0, RootHash: hash, HeaderHash: hash}}
	epochStartHeader.MiniBlockHeaders = []block.MiniBlockHeader{miniBlockHeader}

	var receivedMiniblock func(key []byte, value interface{})
	args.MiniBlocksPool = &testscommon.CacherStub{
		RegisterHandlerCalled: func(f func(key []byte, value interface{})) {
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
		PutCalled: func(key []byte, value interface{}, _ int) (evicted bool) {
			if bytes.Equal(key, peerMiniBlockHash) {
				receivedMiniblock(key, value)
				return false
			}
			return false
		},
	}

	args.Requesthandler = &mock.RequestHandlerStub{
		RequestMiniBlocksHandlerCalled: func(destShardID uint32, miniblockHashes [][]byte) {
			if destShardID == core.MetachainShardId &&
				bytes.Equal(miniblockHashes[0], peerMiniBlockHash) {
				time.Sleep(5100 * time.Millisecond)
				args.MiniBlocksPool.Put(peerMiniBlockHash, miniBlockHeader, miniBlockHeader.Size())
			}
		},
	}

	syncer, _ := NewPeerMiniBlockSyncer(args)
	_, _, processError := syncer.SyncMiniBlocks(epochStartHeader)

	require.Equal(t, process.ErrTimeIsOut, processError)
}
