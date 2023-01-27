package shardchain

import (
	"bytes"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDefaultArguments() ArgPeerMiniBlockSyncer {
	defaultArgs := ArgPeerMiniBlockSyncer{
		MiniBlocksPool:     testscommon.NewCacherStub(),
		ValidatorsInfoPool: testscommon.NewShardedDataStub(),
		RequestHandler:     &testscommon.RequestHandlerStub{},
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

func TestNewValidatorInfoProcessor_NilValidatorsInfoPoolShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultArguments()
	args.ValidatorsInfoPool = nil
	syncer, err := NewPeerMiniBlockSyncer(args)

	require.Nil(t, syncer)
	require.Equal(t, epochStart.ErrNilValidatorsInfoPool, err)
}

func TestNewValidatorInfoProcessor_NilRequestHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultArguments()
	args.RequestHandler = nil
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

	args.RequestHandler = &testscommon.RequestHandlerStub{
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

	args.RequestHandler = &testscommon.RequestHandlerStub{
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

func TestValidatorInfoProcessor_SyncValidatorsInfo(t *testing.T) {
	t.Parallel()

	t.Run("sync validators info with nil block body", func(t *testing.T) {
		t.Parallel()

		args := createDefaultArguments()
		syncer, _ := NewPeerMiniBlockSyncer(args)

		missingValidatorsInfoHashes, validatorsInfo, err := syncer.SyncValidatorsInfo(nil)
		assert.Nil(t, missingValidatorsInfoHashes)
		assert.Nil(t, validatorsInfo)
		assert.Equal(t, epochStart.ErrNilBlockBody, err)
	})

	t.Run("sync validators info with missing validators info", func(t *testing.T) {
		t.Parallel()

		args := createDefaultArguments()
		args.ValidatorsInfoPool = &testscommon.ShardedDataStub{
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}
		syncer, _ := NewPeerMiniBlockSyncer(args)

		body := &block.Body{}
		body.MiniBlocks = append(body.MiniBlocks, createMockMiniBlock(core.MetachainShardId, 0, block.TxBlock))
		body.MiniBlocks = append(body.MiniBlocks, createMockMiniBlock(core.MetachainShardId, 0, block.PeerBlock))
		missingValidatorsInfoHashes, validatorsInfo, err := syncer.SyncValidatorsInfo(body)

		assert.Equal(t, 3, len(missingValidatorsInfoHashes))
		assert.Nil(t, validatorsInfo)
		assert.Equal(t, process.ErrTimeIsOut, err)
	})

	t.Run("sync validators info without missing validators info", func(t *testing.T) {
		t.Parallel()

		svi1 := &state.ShardValidatorInfo{PublicKey: []byte("x")}
		svi2 := &state.ShardValidatorInfo{PublicKey: []byte("y")}
		svi3 := &state.ShardValidatorInfo{PublicKey: []byte("z")}

		args := createDefaultArguments()
		args.ValidatorsInfoPool = &testscommon.ShardedDataStub{
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal(key, []byte("a")) {
					return svi1, true
				}
				if bytes.Equal(key, []byte("b")) {
					return svi2, true
				}
				if bytes.Equal(key, []byte("c")) {
					return svi3, true
				}
				return nil, false
			},
		}
		syncer, _ := NewPeerMiniBlockSyncer(args)

		body := &block.Body{}
		body.MiniBlocks = append(body.MiniBlocks, createMockMiniBlock(core.MetachainShardId, 0, block.TxBlock))
		body.MiniBlocks = append(body.MiniBlocks, createMockMiniBlock(core.MetachainShardId, 0, block.PeerBlock))
		missingValidatorsInfoHashes, validatorsInfo, err := syncer.SyncValidatorsInfo(body)

		assert.Nil(t, err)
		assert.Nil(t, missingValidatorsInfoHashes)
		assert.Equal(t, 3, len(validatorsInfo))
		assert.Equal(t, svi1, validatorsInfo["a"])
		assert.Equal(t, svi2, validatorsInfo["b"])
		assert.Equal(t, svi3, validatorsInfo["c"])
	})
}

func TestValidatorInfoProcessor_ReceivedValidatorInfo(t *testing.T) {
	t.Parallel()

	t.Run("received validators info with wrong type assertion", func(t *testing.T) {
		t.Parallel()

		args := createDefaultArguments()
		syncer, _ := NewPeerMiniBlockSyncer(args)
		syncer.initValidatorsInfo()

		syncer.mutValidatorsInfoForBlock.Lock()
		syncer.mapAllValidatorsInfo["a"] = nil
		syncer.numMissingValidatorsInfo = 1
		syncer.mutValidatorsInfoForBlock.Unlock()

		syncer.receivedValidatorInfo([]byte("a"), nil)

		syncer.mutValidatorsInfoForBlock.RLock()
		numMissingValidatorsInfo := syncer.numMissingValidatorsInfo
		syncer.mutValidatorsInfoForBlock.RUnlock()

		assert.Equal(t, uint32(1), numMissingValidatorsInfo)
	})

	t.Run("received validators info with not requested validator info", func(t *testing.T) {
		t.Parallel()

		svi := &state.ShardValidatorInfo{PublicKey: []byte("x")}

		args := createDefaultArguments()
		syncer, _ := NewPeerMiniBlockSyncer(args)
		syncer.initValidatorsInfo()

		syncer.mutValidatorsInfoForBlock.Lock()
		syncer.mapAllValidatorsInfo["a"] = nil
		syncer.numMissingValidatorsInfo = 1
		syncer.mutValidatorsInfoForBlock.Unlock()

		syncer.receivedValidatorInfo([]byte("b"), svi)

		syncer.mutValidatorsInfoForBlock.RLock()
		numMissingValidatorsInfo := syncer.numMissingValidatorsInfo
		syncer.mutValidatorsInfoForBlock.RUnlock()

		assert.Equal(t, uint32(1), numMissingValidatorsInfo)
	})

	t.Run("received validators info with already received validator info", func(t *testing.T) {
		t.Parallel()

		svi := &state.ShardValidatorInfo{PublicKey: []byte("x")}

		args := createDefaultArguments()
		syncer, _ := NewPeerMiniBlockSyncer(args)
		syncer.initValidatorsInfo()

		syncer.mutValidatorsInfoForBlock.Lock()
		syncer.mapAllValidatorsInfo["a"] = svi
		syncer.numMissingValidatorsInfo = 1
		syncer.mutValidatorsInfoForBlock.Unlock()

		syncer.receivedValidatorInfo([]byte("a"), svi)

		syncer.mutValidatorsInfoForBlock.RLock()
		numMissingValidatorsInfo := syncer.numMissingValidatorsInfo
		syncer.mutValidatorsInfoForBlock.RUnlock()

		assert.Equal(t, uint32(1), numMissingValidatorsInfo)
	})

	t.Run("received validators info with missing validator info", func(t *testing.T) {
		t.Parallel()

		svi := &state.ShardValidatorInfo{PublicKey: []byte("x")}

		args := createDefaultArguments()
		syncer, _ := NewPeerMiniBlockSyncer(args)
		syncer.initValidatorsInfo()

		syncer.mutValidatorsInfoForBlock.Lock()
		syncer.mapAllValidatorsInfo["a"] = nil
		syncer.numMissingValidatorsInfo = 1
		syncer.mutValidatorsInfoForBlock.Unlock()

		wasWithTimeOut := atomic.Flag{}
		go func() {
			select {
			case <-syncer.chRcvAllValidatorsInfo:
				return
			case <-time.After(time.Second):
				wasWithTimeOut.SetValue(true)
				return
			}
		}()

		syncer.receivedValidatorInfo([]byte("a"), svi)

		syncer.mutValidatorsInfoForBlock.RLock()
		numMissingValidatorsInfo := syncer.numMissingValidatorsInfo
		syncer.mutValidatorsInfoForBlock.RUnlock()

		assert.False(t, wasWithTimeOut.IsSet())
		assert.Equal(t, uint32(0), numMissingValidatorsInfo)
	})
}

func TestValidatorInfoProcessor_GetAllValidatorsInfoShouldWork(t *testing.T) {
	t.Parallel()

	svi1 := &state.ShardValidatorInfo{PublicKey: []byte("x")}
	svi2 := &state.ShardValidatorInfo{PublicKey: []byte("y")}
	svi3 := &state.ShardValidatorInfo{PublicKey: []byte("z")}

	args := createDefaultArguments()
	syncer, _ := NewPeerMiniBlockSyncer(args)
	syncer.initValidatorsInfo()

	syncer.mutValidatorsInfoForBlock.Lock()
	syncer.mapAllValidatorsInfo["a"] = svi1
	syncer.mapAllValidatorsInfo["b"] = svi2
	syncer.mapAllValidatorsInfo["c"] = svi3
	syncer.mutValidatorsInfoForBlock.Unlock()

	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, createMockMiniBlock(core.MetachainShardId, 0, block.TxBlock))
	body.MiniBlocks = append(body.MiniBlocks, createMockMiniBlock(core.MetachainShardId, 0, block.PeerBlock))
	validatorsInfo := syncer.getAllValidatorsInfo(body)

	assert.Equal(t, 3, len(validatorsInfo))
	assert.Equal(t, svi1, validatorsInfo["a"])
	assert.Equal(t, svi2, validatorsInfo["b"])
	assert.Equal(t, svi3, validatorsInfo["c"])
}

func TestValidatorInfoProcessor_ComputeMissingValidatorsInfoShouldWork(t *testing.T) {
	t.Parallel()

	svi1 := &state.ShardValidatorInfo{PublicKey: []byte("x")}
	svi2 := &state.ShardValidatorInfo{PublicKey: []byte("y")}

	args := createDefaultArguments()
	args.ValidatorsInfoPool = &testscommon.ShardedDataStub{
		SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(key, []byte("a")) {
				return svi1, true
			}
			if bytes.Equal(key, []byte("b")) {
				return svi2, true
			}
			return nil, false
		},
	}
	syncer, _ := NewPeerMiniBlockSyncer(args)
	syncer.initValidatorsInfo()

	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, createMockMiniBlock(core.MetachainShardId, 0, block.TxBlock))
	body.MiniBlocks = append(body.MiniBlocks, createMockMiniBlock(core.MetachainShardId, 0, block.PeerBlock))
	syncer.computeMissingValidatorsInfo(body)

	syncer.mutValidatorsInfoForBlock.RLock()
	assert.Equal(t, uint32(1), syncer.numMissingValidatorsInfo)
	assert.Equal(t, 3, len(syncer.mapAllValidatorsInfo))
	assert.Equal(t, svi1, syncer.mapAllValidatorsInfo["a"])
	assert.Equal(t, svi2, syncer.mapAllValidatorsInfo["b"])
	assert.Nil(t, syncer.mapAllValidatorsInfo["c"])
	syncer.mutValidatorsInfoForBlock.RUnlock()
}

func TestValidatorInfoProcessor_RetrieveMissingValidatorsInfo(t *testing.T) {
	t.Parallel()

	t.Run("retrieve missing validators info without missing validators info", func(t *testing.T) {
		t.Parallel()

		args := createDefaultArguments()
		syncer, _ := NewPeerMiniBlockSyncer(args)

		missingValidatorsInfoHashes, err := syncer.retrieveMissingValidatorsInfo()
		assert.Nil(t, missingValidatorsInfoHashes)
		assert.Nil(t, err)
	})

	t.Run("retrieve missing validators info with not all validators info received", func(t *testing.T) {
		t.Parallel()

		svi1 := &state.ShardValidatorInfo{PublicKey: []byte("x")}
		svi2 := &state.ShardValidatorInfo{PublicKey: []byte("y")}
		svi3 := &state.ShardValidatorInfo{PublicKey: []byte("z")}

		args := createDefaultArguments()
		syncer, _ := NewPeerMiniBlockSyncer(args)
		syncer.initValidatorsInfo()
		syncer.requestHandler = &testscommon.RequestHandlerStub{
			RequestValidatorsInfoCalled: func(hashes [][]byte) {
				syncer.mutValidatorsInfoForBlock.Lock()
				for _, hash := range hashes {
					if bytes.Equal(hash, []byte("a")) {
						syncer.mapAllValidatorsInfo["a"] = svi1
					}
					if bytes.Equal(hash, []byte("b")) {
						syncer.mapAllValidatorsInfo["b"] = svi2
					}
					if bytes.Equal(hash, []byte("c")) {
						syncer.mapAllValidatorsInfo["c"] = svi3
					}
				}
				syncer.mutValidatorsInfoForBlock.Unlock()
			},
		}

		syncer.mutValidatorsInfoForBlock.Lock()
		syncer.mapAllValidatorsInfo["a"] = nil
		syncer.mapAllValidatorsInfo["b"] = nil
		syncer.mapAllValidatorsInfo["c"] = nil
		syncer.mapAllValidatorsInfo["d"] = nil
		syncer.mutValidatorsInfoForBlock.Unlock()

		missingValidatorsInfoHashes, err := syncer.retrieveMissingValidatorsInfo()
		assert.Equal(t, process.ErrTimeIsOut, err)
		require.Equal(t, 1, len(missingValidatorsInfoHashes))
		assert.Equal(t, []byte("d"), missingValidatorsInfoHashes[0])
	})

	t.Run("retrieve missing validators info with all validators info received", func(t *testing.T) {
		t.Parallel()

		svi1 := &state.ShardValidatorInfo{PublicKey: []byte("x")}
		svi2 := &state.ShardValidatorInfo{PublicKey: []byte("y")}
		svi3 := &state.ShardValidatorInfo{PublicKey: []byte("z")}

		args := createDefaultArguments()
		syncer, _ := NewPeerMiniBlockSyncer(args)
		syncer.initValidatorsInfo()
		syncer.requestHandler = &testscommon.RequestHandlerStub{
			RequestValidatorsInfoCalled: func(hashes [][]byte) {
				syncer.mutValidatorsInfoForBlock.Lock()
				for _, hash := range hashes {
					if bytes.Equal(hash, []byte("a")) {
						syncer.mapAllValidatorsInfo["a"] = svi1
					}
					if bytes.Equal(hash, []byte("b")) {
						syncer.mapAllValidatorsInfo["b"] = svi2
					}
					if bytes.Equal(hash, []byte("c")) {
						syncer.mapAllValidatorsInfo["c"] = svi3
					}
				}
				syncer.mutValidatorsInfoForBlock.Unlock()
			},
		}

		syncer.mutValidatorsInfoForBlock.Lock()
		syncer.mapAllValidatorsInfo["a"] = nil
		syncer.mapAllValidatorsInfo["b"] = nil
		syncer.mapAllValidatorsInfo["c"] = nil
		syncer.mutValidatorsInfoForBlock.Unlock()

		go func() {
			time.Sleep(waitTime / 2)
			syncer.chRcvAllValidatorsInfo <- struct{}{}
		}()

		missingValidatorsInfoHashes, err := syncer.retrieveMissingValidatorsInfo()

		assert.Nil(t, err)
		assert.Nil(t, missingValidatorsInfoHashes)
	})
}

func TestValidatorInfoProcessor_GetAllMissingValidatorsInfoHashesShouldWork(t *testing.T) {
	t.Parallel()

	svi1 := &state.ShardValidatorInfo{PublicKey: []byte("x")}
	svi2 := &state.ShardValidatorInfo{PublicKey: []byte("y")}
	svi3 := &state.ShardValidatorInfo{PublicKey: []byte("z")}

	args := createDefaultArguments()
	syncer, _ := NewPeerMiniBlockSyncer(args)
	syncer.initValidatorsInfo()

	syncer.mutValidatorsInfoForBlock.Lock()
	syncer.mapAllValidatorsInfo["a"] = svi1
	syncer.mapAllValidatorsInfo["b"] = svi2
	syncer.mapAllValidatorsInfo["c"] = svi3
	syncer.mapAllValidatorsInfo["d"] = nil
	syncer.mutValidatorsInfoForBlock.Unlock()

	missingValidatorsInfoHashes := syncer.getAllMissingValidatorsInfoHashes()
	require.Equal(t, 1, len(missingValidatorsInfoHashes))
	assert.Equal(t, []byte("d"), missingValidatorsInfoHashes[0])
}

func createMockMiniBlock(senderShardID, receiverShardID uint32, blockType block.Type) *block.MiniBlock {
	return &block.MiniBlock{
		SenderShardID:   senderShardID,
		ReceiverShardID: receiverShardID,
		Type:            blockType,
		TxHashes: [][]byte{
			[]byte("a"),
			[]byte("b"),
			[]byte("c"),
		},
	}
}
