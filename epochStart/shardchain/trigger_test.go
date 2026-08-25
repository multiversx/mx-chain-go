package shardchain

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	vic "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
)

func createMockShardEpochStartTriggerArguments() *ArgsShardEpochStartTrigger {
	return &ArgsShardEpochStartTrigger{
		Marshalizer: &marshal.GogoProtoMarshalizer{},
		Hasher:      &hashingMocks.HasherMock{},
		HeaderValidator: &mock.HeaderValidatorStub{
			IsHeaderConstructionValidCalled: func(currHdr, prevHdr data.HeaderHandler) error {
				return nil
			},
		},
		Uint64Converter: &mock.Uint64ByteSliceConverterMock{},
		DataPool: &dataRetrieverMock.PoolsHolderStub{
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{}
			},
			MiniBlocksCalled: func() storage.Cacher {
				return cache.NewCacherStub()
			},
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vic.ValidatorInfoCacherStub{}
			},
			ProofsCalled: func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{}
			},
		},
		Storage: &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) (bytes []byte, err error) {
						return []byte("hash"), nil
					},
					PutCalled: func(key, data []byte) error {
						return nil
					},
				}, nil
			},
		},
		RequestHandler:       &testscommon.RequestHandlerStub{},
		EpochStartNotifier:   &mock.EpochStartNotifierStub{},
		PeerMiniBlocksSyncer: &mock.ValidatorInfoSyncerStub{},
		RoundHandler:         &mock.RoundHandlerStub{},
		AppStatusHandler:     &statusHandlerMock.AppStatusHandlerStub{},
		EnableEpochsHandler:  &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ExtraDelayForRequestBlockInfoInMilliseconds: 0,
	}
}

func TestNewEpochStartTrigger_NilArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	epochStartTrigger, err := NewEpochStartTrigger(nil)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilArgsNewShardEpochStartTrigger, err)
}

func TestNewEpochStartTrigger_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.Hasher = nil
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilHasher, err)
}

func TestNewEpochStartTrigger_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.Marshalizer = nil
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilMarshalizer, err)
}

func TestNewEpochStartTrigger_NilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.HeaderValidator = nil
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilHeaderValidator, err)
}

func TestNewEpochStartTrigger_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.DataPool = nil
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilDataPoolsHolder, err)
}

func TestNewEpochStartTrigger_NilStorageShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.Storage = nil
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilStorageService, err)
}

func TestNewEpochStartTrigger_NilRequestHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.RequestHandler = nil
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilRequestHandler, err)
}

func TestNewEpochStartTrigger_NilUint64ConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.Uint64Converter = nil
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilUint64Converter, err)
}

func TestNewEpochStartTrigger_NilEpochStartNotifierShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.EpochStartNotifier = nil
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilEpochStartNotifier, err)
}

func TestNewEpochStartTrigger_GetStorerReturnsErr(t *testing.T) {
	t.Parallel()

	t.Run("missing MetaBlockUnit", testWithMissingStorer(dataRetriever.MetaBlockUnit))
	t.Run("missing BootstrapUnit", testWithMissingStorer(dataRetriever.BootstrapUnit))
	t.Run("missing MetaHdrNonceHashDataUnit", testWithMissingStorer(dataRetriever.MetaHdrNonceHashDataUnit))
	t.Run("missing BlockHeaderUnit", testWithMissingStorer(dataRetriever.BlockHeaderUnit))
}

func testWithMissingStorer(missingUnit dataRetriever.UnitType) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		args := createMockShardEpochStartTriggerArguments()
		args.Storage = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				if unitType == missingUnit {
					return nil, fmt.Errorf("%w for %s", storage.ErrKeyNotFound, missingUnit.String())
				}
				return &storageStubs.StorerStub{}, nil
			},
		}

		epochStartTrigger, err := NewEpochStartTrigger(args)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), storage.ErrKeyNotFound.Error()))
		require.True(t, strings.Contains(err.Error(), missingUnit.String()))
		require.True(t, check.IfNil(epochStartTrigger))
	}
}

func TestNewEpochStartTrigger_NilHeadersPoolShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return nil
		},
		MiniBlocksCalled: func() storage.Cacher {
			return cache.NewCacherStub()
		},
	}
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilMetaBlocksPool, err)
}

func TestNewEpochStartTrigger_NilValidatorInfoProcessorShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.PeerMiniBlocksSyncer = nil
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilValidatorInfoProcessor, err)
}

func TestNewEpochStartTrigger_NilRoundHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.RoundHandler = nil
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilRoundHandler, err)
}

func TestNewEpochStartTrigger_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.EnableEpochsHandler = nil
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilEnableEpochsHandler, err)
}

func TestNewEpochStartTrigger_InvalidEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined()
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
}

func TestNewEpochStartTrigger_ExtraDelayForRequestBlockInfo(t *testing.T) {
	t.Parallel()

	t.Run("negative delay should error", func(t *testing.T) {
		args := createMockShardEpochStartTriggerArguments()
		args.ExtraDelayForRequestBlockInfoInMilliseconds = -1

		trigger, err := NewEpochStartTrigger(args)

		require.Nil(t, trigger)
		require.ErrorIs(t, err, process.ErrNegativeValue)
	})

	t.Run("configured delay should be stored", func(t *testing.T) {
		args := createMockShardEpochStartTriggerArguments()
		args.ExtraDelayForRequestBlockInfoInMilliseconds = 400

		trigger, err := NewEpochStartTrigger(args)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, trigger.Close()) })

		require.Equal(t, 400*time.Millisecond, trigger.getExtraDelayForRequestsBlockInfo())
	})
}

func TestNewEpochStartTrigger_ShouldOk(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.NotNil(t, epochStartTrigger)
	assert.Nil(t, err)
}

func TestTrigger_ReceivedHeaderNotEpochStart(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.Validity = 2
	args.Finality = 2
	epochStartTrigger, _ := NewEpochStartTrigger(args)

	hash := []byte("hash")
	header := &block.MetaBlock{Nonce: 100}
	header.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0, RootHash: hash, HeaderHash: hash}}
	epochStartTrigger.receivedMetaBlock(header, hash)

	assert.False(t, epochStartTrigger.IsEpochStart())
}

func TestTrigger_ReceivedHeaderIsEpochStartTrue(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.Validity = 1
	args.Finality = 2
	epochStartTrigger, _ := NewEpochStartTrigger(args)

	oldEpHeader := &block.MetaBlock{Nonce: 99, Epoch: 0}
	oldHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, oldEpHeader)

	hash := []byte("hash")
	header := &block.MetaBlock{Nonce: 100, Epoch: 1, PrevHash: oldHash}
	header.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0, RootHash: hash, HeaderHash: hash}}

	prevHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, header)
	epochStartTrigger.receivedMetaBlock(header, prevHash)
	epochStartTrigger.receivedMetaBlock(oldEpHeader, oldHash)

	header = &block.MetaBlock{Nonce: 101, Epoch: 1, PrevHash: prevHash}
	prevHash, _ = core.CalculateHash(args.Marshalizer, args.Hasher, header)
	epochStartTrigger.receivedMetaBlock(header, prevHash)

	header = &block.MetaBlock{Nonce: 102, Epoch: 1, PrevHash: prevHash}
	currHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, header)
	epochStartTrigger.receivedMetaBlock(header, currHash)

	assert.True(t, epochStartTrigger.IsEpochStart())
}

func TestTrigger_ReceivedHeaderIsEpochStartTrueWithPeerMiniblocks(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()

	hash := []byte("hash")

	peerMiniblock := &block.MiniBlock{
		TxHashes:        [][]byte{},
		ReceiverShardID: core.AllShardId,
		SenderShardID:   core.MetachainShardId,
	}

	peerMiniBlockHash, _ := args.Marshalizer.Marshal(peerMiniblock)

	miniBlockHeader := block.MiniBlockHeader{
		Hash: peerMiniBlockHash, Type: block.PeerBlock, SenderShardID: core.MetachainShardId, ReceiverShardID: core.AllShardId, TxCount: 1}

	previousHeader99 := &block.MetaBlock{Nonce: 99, Epoch: 0}
	previousHeaderHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, previousHeader99)

	epochStartHeader := &block.MetaBlock{Nonce: 100, Epoch: 1, PrevHash: previousHeaderHash}
	epochStartHeader.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0, RootHash: hash, HeaderHash: hash}}
	epochStartHeader.MiniBlockHeaders = []block.MiniBlockHeader{miniBlockHeader}
	epochStartHeaderHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, epochStartHeader)

	newHeader101 := &block.MetaBlock{Nonce: 101, Epoch: 1, PrevHash: epochStartHeaderHash}
	newHeaderHash101, _ := core.CalculateHash(args.Marshalizer, args.Hasher, newHeader101)

	newHeader102 := &block.MetaBlock{Nonce: 102, Epoch: 1, PrevHash: newHeaderHash101}
	newHeaderHash102, _ := core.CalculateHash(args.Marshalizer, args.Hasher, newHeader102)

	hashesToHeaders := make(map[string]data.HeaderHandler)
	hashesToHeaders[string(previousHeaderHash)] = previousHeader99
	hashesToHeaders[string(epochStartHeaderHash)] = epochStartHeader
	hashesToHeaders[string(newHeaderHash101)] = newHeader101
	hashesToHeaders[string(newHeaderHash102)] = newHeader102

	noncesToHeader := make(map[string][]byte)
	noncesToHeader[fmt.Sprint(previousHeader99.Nonce)] = previousHeaderHash
	noncesToHeader[fmt.Sprint(epochStartHeader.Nonce)] = epochStartHeaderHash
	noncesToHeader[fmt.Sprint(newHeader101.Nonce)] = newHeaderHash101
	noncesToHeader[fmt.Sprint(newHeader102.Nonce)] = newHeaderHash102

	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{
				GetHeaderByHashCalled: func(hash []byte) (handler data.HeaderHandler, err error) {
					header, ok := hashesToHeaders[string(hash)]
					if !ok {
						return nil, nil
					}
					return header, nil
				},
				GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, err error) {
					return nil, nil, nil
				},
			}
		},
		MiniBlocksCalled: func() storage.Cacher {
			return &cache.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					if bytes.Equal(key, peerMiniBlockHash) {
						return peerMiniblock, true
					}
					return nil, false
				},
			}
		},
		CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
			return &vic.ValidatorInfoCacherStub{}
		},
		ProofsCalled: func() dataRetriever.ProofsPool {
			return &dataRetrieverMock.ProofsPoolMock{}
		},
	}
	args.Uint64Converter = &mock.Uint64ByteSliceConverterMock{
		ToByteSliceCalled: func(u uint64) []byte {
			return []byte(fmt.Sprint(u))
		},
	}
	args.Storage = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, err error) {
					return noncesToHeader[string(key)], nil
				},
				PutCalled: func(key, data []byte) error {
					return nil
				},
			}, nil
		},
	}

	args.Validity = 1
	args.Finality = 2

	epochStartTrigger, _ := NewEpochStartTrigger(args)

	currHash, err := core.CalculateHash(args.Marshalizer, args.Hasher, previousHeader99)
	require.Nil(t, err)
	epochStartTrigger.receivedMetaBlock(previousHeader99, currHash)
	require.False(t, epochStartTrigger.IsEpochStart())

	currHash, err = core.CalculateHash(args.Marshalizer, args.Hasher, epochStartHeader)
	require.Nil(t, err)
	epochStartTrigger.receivedMetaBlock(epochStartHeader, currHash)
	require.False(t, epochStartTrigger.IsEpochStart())

	currHash, err = core.CalculateHash(args.Marshalizer, args.Hasher, newHeader101)
	require.Nil(t, err)
	epochStartTrigger.receivedMetaBlock(newHeader101, currHash)
	require.False(t, epochStartTrigger.IsEpochStart())

	currHash, err = core.CalculateHash(args.Marshalizer, args.Hasher, newHeader102)
	require.Nil(t, err)
	epochStartTrigger.receivedMetaBlock(newHeader102, currHash)
	require.True(t, epochStartTrigger.IsEpochStart())
}

func TestTrigger_Epoch(t *testing.T) {
	t.Parallel()

	epoch := uint32(1)
	args := createMockShardEpochStartTriggerArguments()
	args.Epoch = epoch
	epochStartTrigger, _ := NewEpochStartTrigger(args)

	currentEpoch := epochStartTrigger.Epoch()
	assert.Equal(t, epoch, currentEpoch)
}

func TestTrigger_RequestEpochStartIfNeeded(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	called := false
	args.RequestHandler = &testscommon.RequestHandlerStub{
		RequestStartOfEpochMetaBlockCalled: func(_ uint32) {
			called = true
		},
	}
	et, _ := NewEpochStartTrigger(args)
	et.epoch = 2

	hash := []byte("hash")
	et.RequestEpochStartIfNeeded(&block.Header{Epoch: 10})
	assert.False(t, called)

	et.RequestEpochStartIfNeeded(&block.MetaBlock{Epoch: 3,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{ShardID: 0, RootHash: hash, HeaderHash: hash}}}})
	assert.False(t, called)

	et.RequestEpochStartIfNeeded(&block.MetaBlock{Epoch: 2})
	assert.False(t, called)

	et.mapEpochStartHdrs[string(hash)] = &block.MetaBlock{Epoch: 3}
	et.RequestEpochStartIfNeeded(&block.MetaBlock{Epoch: 3})
	assert.False(t, called)

	et.RequestEpochStartIfNeeded(&block.MetaBlock{Epoch: 4})
	assert.True(t, called)
}

func TestTrigger_RevertStateToBlockBehindEpochStart(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()

	prevEpochHdr := &block.Header{Round: 20, Epoch: 2}
	prevEpochHdrBuff, _ := args.Marshalizer.Marshal(prevEpochHdr)

	args.Storage = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, err error) {
					return []byte("hash"), nil
				},
				PutCalled: func(key, data []byte) error {
					return nil
				},
				SearchFirstCalled: func(key []byte) (bytes []byte, err error) {
					return prevEpochHdrBuff, nil
				},
				RemoveCalled: func(key []byte) error {
					return nil
				},
			}, nil
		},
	}
	et, _ := NewEpochStartTrigger(args)

	prevHdr := &block.Header{Round: 29, Epoch: 2}
	prevHash, _ := core.CalculateHash(et.marshaller, et.hasher, prevHdr)

	epochStartShHdr := &block.Header{
		Nonce:              30,
		PrevHash:           prevHash,
		Round:              30,
		EpochStartMetaHash: []byte("metaHash"),
		Epoch:              3,
	}
	et.SetProcessed(epochStartShHdr, nil)

	err := et.RevertStateToBlock(epochStartShHdr)
	assert.Nil(t, err)
	assert.Equal(t, et.epoch, epochStartShHdr.Epoch)
	assert.False(t, et.IsEpochStart())

	err = et.RevertStateToBlock(prevHdr)
	assert.Nil(t, err)
	assert.True(t, et.IsEpochStart())
}

func TestTrigger_LastCommitedShardEpochStartBlock(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	et, _ := NewEpochStartTrigger(args)

	epoch := uint32(37)

	epochStartNonce := uint64(100)
	epochStartRound := uint64(101)
	ecpohStartTimeStamp := uint64(102)

	epochStartShHdr := &block.Header{
		Epoch:              epoch,
		Nonce:              epochStartNonce,
		Round:              epochStartRound,
		TimeStamp:          ecpohStartTimeStamp,
		EpochStartMetaHash: []byte("metaHash"),
	}

	nonce := uint64(200)
	round := uint64(201)
	timeStamp := uint64(202)

	shHdr := &block.Header{
		Epoch:     epoch,
		Nonce:     nonce,
		Round:     round,
		TimeStamp: timeStamp,
	}

	et.SetProcessed(epochStartShHdr, nil)
	et.SetProcessed(shHdr, nil)

	lastCommitedEpochStartBlock, err := et.LastCommitedEpochStartHdr()
	require.Nil(t, err)
	require.Equal(t, epochStartShHdr, lastCommitedEpochStartBlock)
}

func TestTrigger_RevertStateToBlockBehindEpochStartNoBlockInAnEpoch(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()

	prevEpochHdr := &block.Header{Round: 20, Epoch: 1}
	prevEpochHdrBuff, _ := args.Marshalizer.Marshal(prevEpochHdr)

	epochStartKey := core.EpochStartIdentifier(prevEpochHdr.Epoch)

	args.Storage = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, err error) {
					return []byte("hash"), nil
				},
				PutCalled: func(key, data []byte) error {
					return nil
				},
				SearchFirstCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, []byte(epochStartKey)) {
						return prevEpochHdrBuff, nil
					}
					return nil, epochStart.ErrMissingHeader
				},
				RemoveCalled: func(key []byte) error {
					return nil
				},
			}, nil
		},
	}
	et, _ := NewEpochStartTrigger(args)

	// epoch 2 produced no shard block: the last block before the epoch 3 start is still epoch 1
	prevHdr := &block.Header{Round: 29, Epoch: 1}
	prevHash, _ := core.CalculateHash(et.marshaller, et.hasher, prevHdr)

	epochStartShHdr := &block.Header{
		Nonce:              30,
		PrevHash:           prevHash,
		Round:              30,
		EpochStartMetaHash: []byte("metaHash"),
		Epoch:              3,
	}
	et.SetProcessed(epochStartShHdr, nil)

	err := et.RevertStateToBlock(epochStartShHdr)
	assert.Nil(t, err)
	assert.Equal(t, et.epoch, epochStartShHdr.Epoch)
	assert.False(t, et.IsEpochStart())

	err = et.RevertStateToBlock(prevHdr)
	assert.Nil(t, err)
	assert.True(t, et.IsEpochStart())
	assert.Equal(t, et.epochStartShardHeader.GetEpoch(), prevEpochHdr.Epoch)
}

func TestTrigger_RevertStateToBlockMissingEpochStartHeaderErrors(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()

	args.Storage = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return []byte("hash"), nil
				},
				PutCalled: func(key, data []byte) error {
					return nil
				},
				SearchFirstCalled: func(key []byte) ([]byte, error) {
					return nil, epochStart.ErrMissingHeader
				},
				RemoveCalled: func(key []byte) error {
					return nil
				},
			}, nil
		},
	}
	et, _ := NewEpochStartTrigger(args)

	prevHdr := &block.Header{Round: 29, Epoch: 2}
	prevHash, _ := core.CalculateHash(et.marshaller, et.hasher, prevHdr)

	epochStartShHdr := &block.Header{
		Nonce:              30,
		PrevHash:           prevHash,
		Round:              30,
		EpochStartMetaHash: []byte("metaHash"),
		Epoch:              3,
	}
	et.SetProcessed(epochStartShHdr, nil)

	// the target's own epoch start is missing from storage: corruption must surface, not be
	// papered over with fabricated older state
	err := et.RevertStateToBlock(prevHdr)
	assert.Equal(t, epochStart.ErrMissingHeader, err)
	assert.Equal(t, epochStartShHdr.Epoch, et.epochStartShardHeader.GetEpoch())
}

func TestTrigger_ReceivedEpochStartHeaderChangeEpochFinalityAttestingRound(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.Validity = 1
	args.Finality = 1
	epochStartTrigger, _ := NewEpochStartTrigger(args)

	oldEpHeader := &block.MetaBlock{Nonce: 99, Round: 99, Epoch: 0}
	oldHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, oldEpHeader)

	hash := []byte("hash")
	header := &block.MetaBlock{Nonce: 100, Round: 100, Epoch: 1, PrevHash: oldHash}
	header.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0, RootHash: hash, HeaderHash: hash}}

	epochStartHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, header)
	epochStartTrigger.receivedMetaBlock(header, epochStartHash)
	epochStartTrigger.receivedMetaBlock(oldEpHeader, oldHash)

	header102 := &block.MetaBlock{Nonce: 101, Round: 102, Epoch: 1, PrevHash: epochStartHash}
	hash102, _ := core.CalculateHash(args.Marshalizer, args.Hasher, header102)
	epochStartTrigger.receivedMetaBlock(header102, hash102)

	require.True(t, epochStartTrigger.IsEpochStart())
	require.Equal(t, uint64(102), epochStartTrigger.EpochFinalityAttestingRound())

	header = &block.MetaBlock{Nonce: 101, Round: 101, Epoch: 1, PrevHash: epochStartHash}
	currHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, header)
	epochStartTrigger.receivedMetaBlock(header, currHash)

	require.Equal(t, uint64(101), epochStartTrigger.EpochFinalityAttestingRound())

	header103 := &block.MetaBlock{Nonce: 102, Round: 103, Epoch: 1, PrevHash: hash102}
	hash103, _ := core.CalculateHash(args.Marshalizer, args.Hasher, header102)
	epochStartTrigger.receivedMetaBlock(header103, hash103)
	require.Equal(t, uint64(102), epochStartTrigger.EpochFinalityAttestingRound())
}

func TestTrigger_ReceivedHeaderChangeEpochWithoutPrevHeader(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.Validity = 1
	args.Finality = 1

	oldEpHeader := &block.MetaBlock{Nonce: 99, Round: 99, Epoch: 0}
	oldHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, oldEpHeader)

	hash := []byte("hash")
	epochStartHeader := &block.MetaBlock{Nonce: 100, Round: 100, Epoch: 1, PrevHash: oldHash}
	epochStartHeader.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardID: 0, RootHash: hash, HeaderHash: hash}}
	epochStartHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, epochStartHeader)

	nextHeader := &block.MetaBlock{Nonce: 101, Round: 101, Epoch: 1, PrevHash: epochStartHash}
	nextHeaderHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, nextHeader)

	numGetHeadersFromPoolCalls := 0
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					if bytes.Equal(hash, oldHash) {
						if numGetHeadersFromPoolCalls == 0 {
							numGetHeadersFromPoolCalls++
							return nil, errors.New("not found")
						}

						return oldEpHeader, nil
					}

					if bytes.Equal(hash, epochStartHash) {
						return epochStartHeader, nil
					}
					if bytes.Equal(hash, nextHeaderHash) {
						return nextHeader, nil
					}

					return &block.MetaBlock{}, nil
				},
				GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
					if hdrNonce == epochStartHeader.Nonce {
						return []data.HeaderHandler{epochStartHeader}, [][]byte{epochStartHash}, nil
					}

					if hdrNonce == nextHeader.Nonce {
						return []data.HeaderHandler{nextHeader}, [][]byte{nextHeaderHash}, nil
					}

					return make([]data.HeaderHandler, 0), make([][]byte, 0), nil
				},
			}
		},
		MiniBlocksCalled: func() storage.Cacher {
			return cache.NewCacherStub()
		},
		CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
			return &vic.ValidatorInfoCacherStub{}
		},
		ProofsCalled: func() dataRetriever.ProofsPool {
			return &dataRetrieverMock.ProofsPoolMock{}
		},
	}

	args.Storage = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) (b []byte, err error) {
					if bytes.Equal(key, oldHash) {
						return nil, errors.New("failed to get header from storage")
					}

					if bytes.Equal(key, nextHeaderHash) {
						return nextHeaderHash, nil
					}

					return []byte("hash"), nil
				},
				PutCalled: func(key, data []byte) error {
					return nil
				},
			}, nil
		},
	}

	epochStartTrigger, err := NewEpochStartTrigger(args)
	require.Nil(t, err)

	epochStartTrigger.receivedMetaBlock(epochStartHeader, epochStartHash)

	require.False(t, epochStartTrigger.IsEpochStart())

	epochStartTrigger.receivedMetaBlock(epochStartHeader, epochStartHash)

	require.True(t, epochStartTrigger.IsEpochStart())
}

func TestTrigger_ReceivedMetaBlock_WithoutProof(t *testing.T) {
	t.Parallel()

	t.Run("receivedMetaBlock should request proof when missing", func(t *testing.T) {
		t.Parallel()

		var proofRequested atomic.Int32
		var requestedHashMut sync.Mutex
		var requestedHash []byte
		var requestedEpoch uint32

		args := createMockShardEpochStartTriggerArguments()
		args.Epoch = 5
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestEquivalentProofByHashForEpochCalled: func(headerShard uint32, headerHash []byte, epoch uint32) {
				requestedHashMut.Lock()
				requestedHash = headerHash
				requestedEpoch = epoch
				requestedHashMut.Unlock()
				proofRequested.Add(1)
			},
		}

		args.DataPool = &dataRetrieverMock.PoolsHolderStub{
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{}
			},
			MiniBlocksCalled: func() storage.Cacher {
				return cache.NewCacherStub()
			},
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vic.ValidatorInfoCacherStub{}
			},
			ProofsCalled: func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{
					GetProofCalled: func(_ uint32, _ []byte) (data.HeaderProofHandler, error) {
						return nil, errors.New("proof not found")
					},
				}
			},
		}

		et, err := NewEpochStartTrigger(args)
		require.Nil(t, err)
		defer func() {
			_ = et.Close()
		}()

		metaBlockHash := []byte("metablock-hash")
		et.receivedMetaBlock(&block.MetaBlock{
			Nonce:      10,
			Round:      42,
			Epoch:      6,
			EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		}, metaBlockHash)

		time.Sleep(10 * time.Millisecond)

		require.Equal(t, int32(1), proofRequested.Load())

		requestedHashMut.Lock()
		require.Equal(t, metaBlockHash, requestedHash)
		require.Equal(t, uint32(6), requestedEpoch)
		requestedHashMut.Unlock()
	})
}

type pendingProofTestHarness struct {
	trigger        *trigger
	proofRequests  atomic.Int32
	headerRequests atomic.Int32

	mutRequested        sync.Mutex
	lastProofRequested  []byte
	lastProofEpoch      uint32
	lastHeaderEpoch     uint32
	proofRequestsByHash map[string]int

	mutPools      sync.Mutex
	pooledHeaders map[string]data.HeaderHandler
	pooledProofs  map[string]data.HeaderProofHandler
}

func newPendingProofTestHarness(t *testing.T, triggerEpoch uint32) *pendingProofTestHarness {
	h := &pendingProofTestHarness{
		pooledHeaders:       make(map[string]data.HeaderHandler),
		pooledProofs:        make(map[string]data.HeaderProofHandler),
		proofRequestsByHash: make(map[string]int),
	}

	args := createMockShardEpochStartTriggerArguments()
	args.Epoch = triggerEpoch
	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.AndromedaFlag
		},
	}
	// round handler far ahead of header rounds skips the late-broadcast wait in updateTriggerHeaderData
	args.RoundHandler = &mock.RoundHandlerStub{IndexCalled: func() int64 {
		return 1000
	}}
	args.RequestHandler = &testscommon.RequestHandlerStub{
		RequestEquivalentProofByHashForEpochCalled: func(headerShard uint32, headerHash []byte, epoch uint32) {
			h.mutRequested.Lock()
			h.lastProofRequested = headerHash
			h.lastProofEpoch = epoch
			h.proofRequestsByHash[string(headerHash)]++
			h.mutRequested.Unlock()
			h.proofRequests.Add(1)
		},
		RequestStartOfEpochMetaBlockCalled: func(epoch uint32) {
			h.mutRequested.Lock()
			h.lastHeaderEpoch = epoch
			h.mutRequested.Unlock()
			h.headerRequests.Add(1)
		},
	}
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					h.mutPools.Lock()
					defer h.mutPools.Unlock()
					header, found := h.pooledHeaders[string(hash)]
					if !found {
						return nil, errors.New("header not found")
					}
					return header, nil
				},
			}
		},
		MiniBlocksCalled: func() storage.Cacher {
			return cache.NewCacherStub()
		},
		CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
			return &vic.ValidatorInfoCacherStub{}
		},
		ProofsCalled: func() dataRetriever.ProofsPool {
			return &dataRetrieverMock.ProofsPoolMock{
				GetProofCalled: func(_ uint32, headerHash []byte) (data.HeaderProofHandler, error) {
					h.mutPools.Lock()
					defer h.mutPools.Unlock()
					proof, found := h.pooledProofs[string(headerHash)]
					if !found {
						return nil, errors.New("proof not found")
					}
					return proof, nil
				},
			}
		},
	}

	et, err := NewEpochStartTrigger(args)
	require.Nil(t, err)
	t.Cleanup(func() {
		_ = et.Close()
	})
	h.trigger = et

	return h
}

func (h *pendingProofTestHarness) setRetryInterval(interval time.Duration) {
	h.trigger.mutPendingEpochStartData.Lock()
	h.trigger.pendingProofRetryInterval = interval
	h.trigger.mutPendingEpochStartData.Unlock()
}

func (h *pendingProofTestHarness) putHeader(hash []byte, header data.HeaderHandler) {
	h.mutPools.Lock()
	h.pooledHeaders[string(hash)] = header
	h.mutPools.Unlock()
}

func (h *pendingProofTestHarness) dropHeader(hash []byte) {
	h.mutPools.Lock()
	delete(h.pooledHeaders, string(hash))
	h.mutPools.Unlock()
}

func (h *pendingProofTestHarness) putProof(hash []byte) {
	h.mutPools.Lock()
	h.pooledProofs[string(hash)] = &block.HeaderProof{HeaderHash: hash, HeaderShardId: core.MetachainShardId}
	h.mutPools.Unlock()
}

func (h *pendingProofTestHarness) pendingProofs() int {
	h.trigger.mutPendingEpochStartData.Lock()
	defer h.trigger.mutPendingEpochStartData.Unlock()

	return len(h.trigger.pendingEpochStartProofs)
}

func (h *pendingProofTestHarness) pendingRecoveries() int {
	h.trigger.mutPendingEpochStartData.Lock()
	defer h.trigger.mutPendingEpochStartData.Unlock()

	return len(h.trigger.pendingEpochStartProofs) + len(h.trigger.pendingEpochStartHeaders)
}

func (h *pendingProofTestHarness) isPending(hash []byte) bool {
	h.trigger.mutPendingEpochStartData.Lock()
	defer h.trigger.mutPendingEpochStartData.Unlock()

	_, found := h.trigger.pendingEpochStartProofs[string(hash)]
	return found
}

func (h *pendingProofTestHarness) isPendingHeader(epoch uint32) bool {
	h.trigger.mutPendingEpochStartData.Lock()
	defer h.trigger.mutPendingEpochStartData.Unlock()

	_, found := h.trigger.pendingEpochStartHeaders[epoch]
	return found
}

func (h *pendingProofTestHarness) wasEpochStartHdrRecorded(hash []byte) bool {
	h.trigger.mutTrigger.RLock()
	defer h.trigger.mutTrigger.RUnlock()

	_, found := h.trigger.mapEpochStartHdrs[string(hash)]
	return found
}

func (h *pendingProofTestHarness) proofRequestCount(hash []byte) int {
	h.mutRequested.Lock()
	defer h.mutRequested.Unlock()

	return h.proofRequestsByHash[string(hash)]
}

func createEpochStartMetaHdr(epoch uint32, nonce uint64) *block.MetaBlock {
	return &block.MetaBlock{
		Nonce:      nonce,
		Round:      nonce,
		Epoch:      epoch,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
	}
}

func TestTrigger_PendingEpochStartProofRecovery(t *testing.T) {
	t.Parallel()

	metaHash := []byte("epoch-start-meta-hash")

	t.Run("missing proof records pending state and requests immediately", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		metaHdr := createEpochStartMetaHdr(6, 42)
		h.putHeader(metaHash, metaHdr)

		h.trigger.receivedMetaBlock(metaHdr, metaHash)

		require.Eventually(t, func() bool {
			return h.proofRequests.Load() == 1
		}, time.Second, 5*time.Millisecond)
		require.Equal(t, 1, h.pendingProofs())

		h.mutRequested.Lock()
		require.Equal(t, metaHash, h.lastProofRequested)
		require.Equal(t, uint32(6), h.lastProofEpoch)
		h.mutRequested.Unlock()

		time.Sleep(100 * time.Millisecond)
		require.Equal(t, int32(1), h.proofRequests.Load())
	})

	t.Run("pending proof is re-requested without further callbacks", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		h.setRetryInterval(20 * time.Millisecond)
		metaHdr := createEpochStartMetaHdr(6, 42)
		h.putHeader(metaHash, metaHdr)

		h.trigger.receivedMetaBlock(metaHdr, metaHash)

		require.Eventually(t, func() bool {
			return h.proofRequests.Load() >= 3
		}, 2*time.Second, 10*time.Millisecond)
		require.Equal(t, 1, h.pendingProofs())
	})

	t.Run("ordinary metablock notifications do not postpone the retry", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		h.setRetryInterval(30 * time.Millisecond)
		metaHdr := createEpochStartMetaHdr(6, 42)
		h.putHeader(metaHash, metaHdr)
		h.trigger.receivedMetaBlock(metaHdr, metaHash)

		stopSpam := make(chan struct{})
		defer close(stopSpam)
		go func() {
			ordinaryHdr := &block.MetaBlock{Nonce: 43, Round: 43, Epoch: 5}
			for {
				select {
				case <-stopSpam:
					return
				default:
					h.trigger.receivedMetaBlock(ordinaryHdr, []byte("ordinary-hash"))
					time.Sleep(5 * time.Millisecond)
				}
			}
		}()

		require.Eventually(t, func() bool {
			return h.proofRequests.Load() >= 3
		}, 2*time.Second, 10*time.Millisecond)
		require.Equal(t, 1, h.pendingProofs())
	})

	t.Run("proof discovered in pool during retry is actively processed", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		h.setRetryInterval(20 * time.Millisecond)
		metaHdr := createEpochStartMetaHdr(6, 42)
		h.putHeader(metaHash, metaHdr)
		h.trigger.receivedMetaBlock(metaHdr, metaHash)
		require.Equal(t, 1, h.pendingProofs())

		h.putProof(metaHash)

		require.Eventually(t, func() bool {
			return h.pendingProofs() == 0 && h.wasEpochStartHdrRecorded(metaHash)
		}, 2*time.Second, 10*time.Millisecond)
	})

	t.Run("receivedProof clears pending state", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		metaHdr := createEpochStartMetaHdr(6, 42)
		h.putHeader(metaHash, metaHdr)
		h.trigger.receivedMetaBlock(metaHdr, metaHash)
		require.Equal(t, 1, h.pendingProofs())

		h.trigger.receivedProof(&block.HeaderProof{HeaderHash: metaHash, HeaderShardId: core.MetachainShardId})

		require.Equal(t, 0, h.pendingProofs())
		require.True(t, h.wasEpochStartHdrRecorded(metaHash))
	})

	t.Run("evicted header is collapsed to epoch recovery until reacquired", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		h.setRetryInterval(20 * time.Millisecond)
		metaHdr := createEpochStartMetaHdr(6, 42)
		h.putHeader(metaHash, metaHdr)
		h.trigger.receivedMetaBlock(metaHdr, metaHash)
		require.Equal(t, 1, h.pendingProofs())

		h.dropHeader(metaHash)

		require.Eventually(t, func() bool {
			return h.headerRequests.Load() >= 2
		}, 2*time.Second, 10*time.Millisecond)
		require.Equal(t, 1, h.pendingRecoveries())

		h.mutRequested.Lock()
		require.Equal(t, uint32(6), h.lastHeaderEpoch)
		h.mutRequested.Unlock()

		h.putHeader(metaHash, metaHdr)
		h.trigger.receivedMetaBlock(metaHdr, metaHash)

		proofRequestsBeforeReacquisition := h.proofRequests.Load()
		require.Eventually(t, func() bool {
			return h.proofRequests.Load() > proofRequestsBeforeReacquisition
		}, 2*time.Second, 10*time.Millisecond)
		require.Equal(t, 1, h.pendingProofs())

		h.putProof(metaHash)

		require.Eventually(t, func() bool {
			return h.pendingProofs() == 0 && h.wasEpochStartHdrRecorded(metaHash)
		}, 2*time.Second, 10*time.Millisecond)
	})

	t.Run("evicted candidates from the same epoch collapse to one recovery entry", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		h.setRetryInterval(20 * time.Millisecond)
		hashA := []byte("epoch-start-hash-a")
		hashB := []byte("epoch-start-hash-b")
		metaHdrA := createEpochStartMetaHdr(6, 42)
		metaHdrB := createEpochStartMetaHdr(6, 43)
		h.putHeader(hashA, metaHdrA)
		h.putHeader(hashB, metaHdrB)
		h.trigger.receivedMetaBlock(metaHdrA, hashA)
		h.trigger.receivedMetaBlock(metaHdrB, hashB)
		require.Equal(t, 2, h.pendingProofs())

		h.dropHeader(hashA)
		h.dropHeader(hashB)

		require.Eventually(t, func() bool {
			return h.pendingProofs() == 0 && h.pendingRecoveries() == 1
		}, 2*time.Second, 10*time.Millisecond)
	})

	t.Run("proofless replacement candidate does not clear epoch header recovery", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		hashA := []byte("evicted-epoch-start-hash")
		hashB := []byte("replacement-epoch-start-hash")

		h.trigger.addPendingEpochStartProof(hashA, 6)
		h.trigger.movePendingProofToHeaderRecovery(string(hashA), 6)
		h.trigger.addPendingEpochStartProof(hashB, 6)

		require.True(t, h.isPendingHeader(6))
		require.True(t, h.isPending(hashB))
		require.Equal(t, 2, h.pendingRecoveries())
	})

	t.Run("periodic proof requests are bounded and rotate fairly", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		h.setRetryInterval(time.Hour)
		numCandidates := maxPendingProofRequestsPerPass + 5
		hashes := make([][]byte, numCandidates)
		for index := range hashes {
			hashes[index] = []byte(fmt.Sprintf("epoch-start-hash-%d", index))
			metaHdr := createEpochStartMetaHdr(6, uint64(100+index))
			h.putHeader(hashes[index], metaHdr)
			h.trigger.receivedMetaBlock(metaHdr, hashes[index])
		}

		require.Eventually(t, func() bool {
			return h.proofRequests.Load() == int32(numCandidates)
		}, 2*time.Second, 10*time.Millisecond)

		numPasses := (numCandidates + maxPendingProofRequestsPerPass - 1) / maxPendingProofRequestsPerPass
		for range numPasses {
			requestsBeforePass := h.proofRequests.Load()
			h.trigger.retryPendingEpochStartProofs()
			requestsInPass := h.proofRequests.Load() - requestsBeforePass
			require.LessOrEqual(t, requestsInPass, int32(maxPendingProofRequestsPerPass))
		}

		for _, hash := range hashes {
			require.GreaterOrEqual(t, h.proofRequestCount(hash), 2)
		}

		h.mutRequested.Lock()
		require.Equal(t, uint32(6), h.lastProofEpoch)
		h.mutRequested.Unlock()
	})

	t.Run("proof discovery is not limited by the request budget", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		h.setRetryInterval(time.Hour)
		numCandidates := maxPendingProofRequestsPerPass + 1
		var lastHash []byte
		for index := range numCandidates {
			hash := []byte(fmt.Sprintf("epoch-start-hash-%d", index))
			metaHdr := createEpochStartMetaHdr(6, uint64(100+index))
			h.putHeader(hash, metaHdr)
			h.trigger.receivedMetaBlock(metaHdr, hash)
			lastHash = hash
		}

		require.Eventually(t, func() bool {
			return h.proofRequests.Load() == int32(numCandidates)
		}, 2*time.Second, 10*time.Millisecond)

		h.putProof(lastHash)
		h.trigger.retryPendingEpochStartProofs()

		require.False(t, h.isPending(lastHash))
		require.True(t, h.wasEpochStartHdrRecorded(lastHash))
	})

	t.Run("entry is dropped once the trigger epoch reaches it", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		h.setRetryInterval(20 * time.Millisecond)
		metaHdr := createEpochStartMetaHdr(6, 42)
		h.putHeader(metaHash, metaHdr)
		h.trigger.receivedMetaBlock(metaHdr, metaHash)
		require.Equal(t, 1, h.pendingProofs())

		h.trigger.mutTrigger.Lock()
		h.trigger.epoch = 6
		h.trigger.mutTrigger.Unlock()

		require.Eventually(t, func() bool {
			return h.pendingProofs() == 0
		}, 2*time.Second, 10*time.Millisecond)

		requestsAfterDrop := h.proofRequests.Load()
		time.Sleep(150 * time.Millisecond)
		require.Equal(t, requestsAfterDrop, h.proofRequests.Load())
	})

	t.Run("multiple pending hashes are retried independently", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		h.setRetryInterval(20 * time.Millisecond)
		hashA := []byte("epoch-start-hash-a")
		hashB := []byte("epoch-start-hash-b")
		metaHdrA := createEpochStartMetaHdr(6, 42)
		metaHdrB := createEpochStartMetaHdr(7, 142)
		h.putHeader(hashA, metaHdrA)
		h.putHeader(hashB, metaHdrB)
		h.trigger.receivedMetaBlock(metaHdrA, hashA)
		h.trigger.receivedMetaBlock(metaHdrB, hashB)
		require.Equal(t, 2, h.pendingProofs())

		require.Eventually(t, func() bool {
			return h.proofRequests.Load() >= 4
		}, 2*time.Second, 10*time.Millisecond)

		h.putProof(hashA)

		require.Eventually(t, func() bool {
			return h.pendingProofs() == 1 && h.isPending(hashB)
		}, 2*time.Second, 10*time.Millisecond)

		requestsAfterCompletion := h.proofRequests.Load()
		require.Eventually(t, func() bool {
			return h.proofRequests.Load() > requestsAfterCompletion
		}, 2*time.Second, 10*time.Millisecond)
	})

	t.Run("repeated receipt of the same hash creates a single pending entry", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		metaHdr := createEpochStartMetaHdr(6, 42)
		h.putHeader(metaHash, metaHdr)

		h.trigger.receivedMetaBlock(metaHdr, metaHash)
		h.trigger.receivedMetaBlock(metaHdr, metaHash)

		require.Equal(t, 1, h.pendingProofs())
	})

	t.Run("close stops the retries", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		h.setRetryInterval(time.Millisecond)
		metaHdr := createEpochStartMetaHdr(6, 42)
		h.putHeader(metaHash, metaHdr)
		h.trigger.receivedMetaBlock(metaHdr, metaHash)

		require.Eventually(t, func() bool {
			return h.proofRequests.Load() >= 2
		}, 2*time.Second, time.Millisecond)

		_ = h.trigger.Close()
		// drain a pass that may have started before cancellation; the ctx guard blocks new ones
		time.Sleep(20 * time.Millisecond)

		requestsAfterClose := h.proofRequests.Load()
		time.Sleep(150 * time.Millisecond)
		require.Equal(t, requestsAfterClose, h.proofRequests.Load())
	})

	t.Run("concurrent completion by callback and retry pass is safe", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		h.setRetryInterval(time.Millisecond)
		metaHdr := createEpochStartMetaHdr(6, 42)
		h.putHeader(metaHash, metaHdr)
		h.trigger.receivedMetaBlock(metaHdr, metaHash)

		h.putProof(metaHash)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				h.trigger.receivedProof(&block.HeaderProof{HeaderHash: metaHash, HeaderShardId: core.MetachainShardId})
			}()
		}
		wg.Wait()

		require.Eventually(t, func() bool {
			return h.pendingProofs() == 0
		}, 2*time.Second, 10*time.Millisecond)
		require.True(t, h.wasEpochStartHdrRecorded(metaHash))
	})

	t.Run("proof already pooled is processed without entering pending state", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)
		metaHdr := createEpochStartMetaHdr(6, 42)
		h.putHeader(metaHash, metaHdr)
		h.putProof(metaHash)

		h.trigger.receivedMetaBlock(metaHdr, metaHash)

		require.Equal(t, 0, h.pendingProofs())
		require.True(t, h.wasEpochStartHdrRecorded(metaHash))
		require.Equal(t, int32(0), h.proofRequests.Load())
	})

	t.Run("non-epoch-start and current-epoch headers do not enter pending state", func(t *testing.T) {
		h := newPendingProofTestHarness(t, 5)

		ordinaryHdr := &block.MetaBlock{Nonce: 43, Round: 43, Epoch: 6}
		h.trigger.receivedMetaBlock(ordinaryHdr, []byte("ordinary-hash"))

		currentEpochStartHdr := createEpochStartMetaHdr(5, 40)
		h.trigger.receivedMetaBlock(currentEpochStartHdr, []byte("current-epoch-start-hash"))

		require.Equal(t, 0, h.pendingProofs())
		require.Equal(t, int32(0), h.proofRequests.Load())
	})
}

func TestTrigger_ClearMissingValidatorsInfoMapShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	epochStartTrigger, _ := NewEpochStartTrigger(args)

	epochStartTrigger.mutMissingValidatorsInfo.Lock()
	epochStartTrigger.mapMissingValidatorsInfo["a"] = 0
	epochStartTrigger.mapMissingValidatorsInfo["b"] = 0
	epochStartTrigger.mapMissingValidatorsInfo["c"] = 1
	epochStartTrigger.mapMissingValidatorsInfo["d"] = 1
	epochStartTrigger.mutMissingValidatorsInfo.Unlock()

	epochStartTrigger.mutMissingValidatorsInfo.RLock()
	numMissingValidatorsInfo := len(epochStartTrigger.mapMissingValidatorsInfo)
	epochStartTrigger.mutMissingValidatorsInfo.RUnlock()
	assert.Equal(t, 4, numMissingValidatorsInfo)

	epochStartTrigger.clearMissingValidatorsInfoMap(0)

	epochStartTrigger.mutMissingValidatorsInfo.RLock()
	numMissingValidatorsInfo = len(epochStartTrigger.mapMissingValidatorsInfo)
	epochStartTrigger.mutMissingValidatorsInfo.RUnlock()
	assert.Equal(t, 2, numMissingValidatorsInfo)

	assert.Equal(t, uint32(1), epochStartTrigger.mapMissingValidatorsInfo["c"])
	assert.Equal(t, uint32(1), epochStartTrigger.mapMissingValidatorsInfo["d"])
}

func TestTrigger_UpdateMissingValidatorsInfo(t *testing.T) {
	t.Parallel()

	t.Run("update missing validators when there are no missing validators", func(t *testing.T) {
		t.Parallel()

		args := createMockShardEpochStartTriggerArguments()
		epochStartTrigger, _ := NewEpochStartTrigger(args)

		epochStartTrigger.updateMissingValidatorsInfo()

		epochStartTrigger.mutMissingValidatorsInfo.RLock()
		assert.Equal(t, 0, len(epochStartTrigger.mapMissingValidatorsInfo))
		epochStartTrigger.mutMissingValidatorsInfo.RUnlock()
	})

	t.Run("update missing validators when there are missing validators", func(t *testing.T) {
		t.Parallel()

		svi1 := &state.ShardValidatorInfo{PublicKey: []byte("x")}
		svi2 := &state.ShardValidatorInfo{PublicKey: []byte("y")}

		args := createMockShardEpochStartTriggerArguments()

		args.DataPool = &dataRetrieverMock.PoolsHolderStub{
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{}
			},
			MiniBlocksCalled: func() storage.Cacher {
				return cache.NewCacherStub()
			},
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vic.ValidatorInfoCacherStub{}
			},
			ValidatorsInfoCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &testscommon.ShardedDataStub{
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
			},
			ProofsCalled: func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{}
			},
		}

		epochStartTrigger, _ := NewEpochStartTrigger(args)

		epochStartTrigger.mutMissingValidatorsInfo.Lock()
		epochStartTrigger.mapMissingValidatorsInfo["a"] = 1
		epochStartTrigger.mapMissingValidatorsInfo["b"] = 1
		epochStartTrigger.mapMissingValidatorsInfo["c"] = 1
		epochStartTrigger.mutMissingValidatorsInfo.Unlock()

		epochStartTrigger.updateMissingValidatorsInfo()

		epochStartTrigger.mutMissingValidatorsInfo.RLock()
		assert.Equal(t, 1, len(epochStartTrigger.mapMissingValidatorsInfo))
		assert.Equal(t, uint32(1), epochStartTrigger.mapMissingValidatorsInfo["c"])
		epochStartTrigger.mutMissingValidatorsInfo.RUnlock()
	})
}

func TestTrigger_AddMissingValidatorsInfo(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	epochStartTrigger, _ := NewEpochStartTrigger(args)

	missingValidatorsInfoHashes := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
	}

	epochStartTrigger.addMissingValidatorsInfo(1, missingValidatorsInfoHashes)

	epochStartTrigger.mutMissingValidatorsInfo.RLock()
	assert.Equal(t, 3, len(epochStartTrigger.mapMissingValidatorsInfo))
	assert.Equal(t, uint32(1), epochStartTrigger.mapMissingValidatorsInfo["a"])
	assert.Equal(t, uint32(1), epochStartTrigger.mapMissingValidatorsInfo["b"])
	assert.Equal(t, uint32(1), epochStartTrigger.mapMissingValidatorsInfo["c"])
	epochStartTrigger.mutMissingValidatorsInfo.RUnlock()
}

func TestTrigger_ReceivedProof(t *testing.T) {
	t.Parallel()

	t.Run("early exits", func(t *testing.T) {
		t.Parallel()

		args := createMockShardEpochStartTriggerArguments()
		args.DataPool = &dataRetrieverMock.PoolsHolderStub{
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
						require.Fail(t, "should have not been called")
						return nil, nil
					},
				}
			},
		}
		epochStartTrigger, _ := NewEpochStartTrigger(args)

		// nil proof
		epochStartTrigger.receivedProof(nil)

		epochStartTrigger.receivedProof(&block.HeaderProof{
			HeaderShardId: 0, // not meta
		})
	})
	t.Run("GetHeaderByHash error should early exit", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockShardEpochStartTriggerArguments()
		args.DataPool = &dataRetrieverMock.PoolsHolderStub{
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
						return nil, expectedErr
					},
				}
			},
			MiniBlocksCalled: func() storage.Cacher {
				return cache.NewCacherStub()
			},
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vic.ValidatorInfoCacherStub{}
			},
			ProofsCalled: func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{}
			},
		}
		args.EpochStartNotifier = &mock.EpochStartNotifierStub{
			NotifyEpochChangeConfirmedCalled: func(epoch uint32) {
				require.Fail(t, "should not have been called")
			},
		}
		epochStartTrigger, _ := NewEpochStartTrigger(args)

		epochStartTrigger.receivedProof(&block.HeaderProof{
			HeaderShardId: core.MetachainShardId,
		})
	})
	t.Run("not meta block should exit", func(t *testing.T) {
		t.Parallel()

		args := createMockShardEpochStartTriggerArguments()
		args.DataPool = &dataRetrieverMock.PoolsHolderStub{
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
						return &block.Header{}, nil
					},
				}
			},
			MiniBlocksCalled: func() storage.Cacher {
				return cache.NewCacherStub()
			},
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vic.ValidatorInfoCacherStub{}
			},
			ProofsCalled: func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{}
			},
		}
		args.EpochStartNotifier = &mock.EpochStartNotifierStub{
			NotifyEpochChangeConfirmedCalled: func(epoch uint32) {
				require.Fail(t, "should not have been called")
			},
		}
		epochStartTrigger, _ := NewEpochStartTrigger(args)

		epochStartTrigger.receivedProof(&block.HeaderProof{
			HeaderShardId: core.MetachainShardId,
		})
	})
	t.Run("should not update trigger should early exit", func(t *testing.T) {
		t.Parallel()

		args := createMockShardEpochStartTriggerArguments()
		args.DataPool = &dataRetrieverMock.PoolsHolderStub{
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
						return &block.MetaBlock{}, nil
					},
				}
			},
			MiniBlocksCalled: func() storage.Cacher {
				return cache.NewCacherStub()
			},
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vic.ValidatorInfoCacherStub{}
			},
			ProofsCalled: func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{}
			},
		}
		args.EpochStartNotifier = &mock.EpochStartNotifierStub{
			NotifyEpochChangeConfirmedCalled: func(epoch uint32) {
				require.Fail(t, "should not have been called")
			},
		}
		epochStartTrigger, _ := NewEpochStartTrigger(args)

		epochStartTrigger.receivedProof(&block.HeaderProof{
			HeaderShardId: core.MetachainShardId,
		})
	})
	t.Run("should work and notify", func(t *testing.T) {
		t.Parallel()

		args := createMockShardEpochStartTriggerArguments()
		args.Validity = 2
		args.DataPool = &dataRetrieverMock.PoolsHolderStub{
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
						return &block.MetaBlock{
							Epoch: 1,
							Nonce: 3,
							EpochStart: block.EpochStart{
								LastFinalizedHeaders: []block.EpochStartShardData{
									{
										ShardID: 0,
									},
								},
							},
						}, nil
					},
				}
			},
			MiniBlocksCalled: func() storage.Cacher {
				return cache.NewCacherStub()
			},
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vic.ValidatorInfoCacherStub{}
			},
			ProofsCalled: func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{}
			},
		}
		wasCalled := false
		args.EpochStartNotifier = &mock.EpochStartNotifierStub{
			NotifyEpochChangeConfirmedCalled: func(epoch uint32) {
				wasCalled = true
			},
		}
		epochStartTrigger, _ := NewEpochStartTrigger(args)

		epochStartTrigger.receivedProof(&block.HeaderProof{
			HeaderShardId: core.MetachainShardId,
		})

		require.True(t, wasCalled)
	})
}

func TestTrigger_WatchdogRequestEpochStartMetaBlock(t *testing.T) {
	t.Parallel()

	t.Run("fires after timeout", func(t *testing.T) {
		t.Parallel()

		var requestedEpoch atomic.Uint32
		var called atomic.Int32
		args := createMockShardEpochStartTriggerArguments()
		args.RoundHandler = &mock.RoundHandlerStub{
			TimeDurationCalled: func() time.Duration {
				return 10 * time.Millisecond
			},
		}
		args.Epoch = 5
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestStartOfEpochMetaBlockCalled: func(epoch uint32) {
				requestedEpoch.Store(epoch)
				called.Add(1)
			},
		}
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}

		et, err := NewEpochStartTrigger(args)
		require.Nil(t, err)
		defer func() {
			_ = et.Close()
		}()

		time.Sleep(200 * time.Millisecond)

		require.Greater(t, called.Load(), int32(0))
		require.Equal(t, uint32(6), requestedEpoch.Load())
	})

	t.Run("resets timer on any metablock reception", func(t *testing.T) {
		t.Parallel()

		var called atomic.Int32
		args := createMockShardEpochStartTriggerArguments()
		args.RoundHandler = &mock.RoundHandlerStub{
			TimeDurationCalled: func() time.Duration {
				return 30 * time.Millisecond
			},
		}
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestStartOfEpochMetaBlockCalled: func(epoch uint32) {
				called.Add(1)
			},
		}
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}

		et, err := NewEpochStartTrigger(args)
		require.Nil(t, err)
		defer func() {
			_ = et.Close()
		}()

		for i := 0; i < 10; i++ {
			select {
			case et.chanMetaBlockReceived <- struct{}{}:
			default:
			}
			time.Sleep(20 * time.Millisecond)
		}

		require.Equal(t, int32(0), called.Load())
	})

	t.Run("skips when epoch start already detected", func(t *testing.T) {
		t.Parallel()

		var called atomic.Int32
		args := createMockShardEpochStartTriggerArguments()
		args.RoundHandler = &mock.RoundHandlerStub{
			TimeDurationCalled: func() time.Duration {
				return 10 * time.Millisecond
			},
			IndexCalled: func() int64 {
				return 100
			},
		}
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestStartOfEpochMetaBlockCalled: func(epoch uint32) {
				called.Add(1)
			},
		}
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}

		et, err := NewEpochStartTrigger(args)
		require.Nil(t, err)
		defer func() {
			_ = et.Close()
		}()

		et.mutTrigger.Lock()
		et.isEpochStart = true
		et.mutTrigger.Unlock()

		time.Sleep(200 * time.Millisecond)

		require.Equal(t, int32(0), called.Load())
	})

	t.Run("fires even when Andromeda disabled", func(t *testing.T) {
		t.Parallel()

		var requestedEpoch atomic.Uint32
		var called atomic.Int32
		args := createMockShardEpochStartTriggerArguments()
		args.RoundHandler = &mock.RoundHandlerStub{
			TimeDurationCalled: func() time.Duration {
				return 10 * time.Millisecond
			},
			IndexCalled: func() int64 {
				return 100
			},
		}
		args.Epoch = 5
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestStartOfEpochMetaBlockCalled: func(epoch uint32) {
				requestedEpoch.Store(epoch)
				called.Add(1)
			},
		}
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return false
			},
		}

		et, err := NewEpochStartTrigger(args)
		require.Nil(t, err)
		defer func() {
			_ = et.Close()
		}()

		time.Sleep(200 * time.Millisecond)

		require.Greater(t, called.Load(), int32(0))
		require.Equal(t, uint32(6), requestedEpoch.Load())
	})

	t.Run("stops on context cancellation", func(t *testing.T) {
		t.Parallel()

		var called atomic.Int32
		args := createMockShardEpochStartTriggerArguments()
		args.RoundHandler = &mock.RoundHandlerStub{
			TimeDurationCalled: func() time.Duration {
				return 10 * time.Millisecond
			},
			IndexCalled: func() int64 {
				return 100
			},
		}
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestStartOfEpochMetaBlockCalled: func(epoch uint32) {
				called.Add(1)
			},
		}
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}

		et, err := NewEpochStartTrigger(args)
		require.Nil(t, err)

		err = et.Close()
		require.Nil(t, err)

		calledBefore := called.Load()
		time.Sleep(200 * time.Millisecond)

		require.Equal(t, calledBefore, called.Load())
	})

	t.Run("does not start when TimeDuration is zero", func(t *testing.T) {
		t.Parallel()

		var called atomic.Int32
		args := createMockShardEpochStartTriggerArguments()
		args.RoundHandler = &mock.RoundHandlerStub{
			TimeDurationCalled: func() time.Duration {
				return 0
			},
		}
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestStartOfEpochMetaBlockCalled: func(epoch uint32) {
				called.Add(1)
			},
		}
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}

		et, err := NewEpochStartTrigger(args)
		require.Nil(t, err)
		defer func() {
			_ = et.Close()
		}()

		time.Sleep(100 * time.Millisecond)

		require.Equal(t, int32(0), called.Load())
	})

	t.Run("receivedMetaBlock signals watchdog", func(t *testing.T) {
		t.Parallel()

		args := createMockShardEpochStartTriggerArguments()
		// zero TimeDuration prevents the watchdog goroutine from starting
		args.RoundHandler = &mock.RoundHandlerStub{
			TimeDurationCalled: func() time.Duration {
				return 0
			},
		}
		et, err := NewEpochStartTrigger(args)
		require.Nil(t, err)
		defer func() {
			_ = et.Close()
		}()

		et.receivedMetaBlock(&block.MetaBlock{
			Nonce: 10,
			Round: 42,
		}, []byte("hash"))

		select {
		case <-et.chanMetaBlockReceived:
			// expected
		default:
			require.Fail(t, "channel should have been signaled")
		}
	})

	t.Run("receivedMetaBlock requests proof when missing in Andromeda", func(t *testing.T) {
		t.Parallel()

		var proofRequested atomic.Int32
		var requestedHashMut sync.Mutex
		var requestedHash []byte
		var requestedEpoch uint32
		args := createMockShardEpochStartTriggerArguments()
		args.Epoch = 5
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestEquivalentProofByHashForEpochCalled: func(headerShard uint32, headerHash []byte, epoch uint32) {
				requestedHashMut.Lock()
				requestedHash = headerHash
				requestedEpoch = epoch
				requestedHashMut.Unlock()
				proofRequested.Add(1)
			},
		}
		args.DataPool = &dataRetrieverMock.PoolsHolderStub{
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{}
			},
			MiniBlocksCalled: func() storage.Cacher {
				return cache.NewCacherStub()
			},
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vic.ValidatorInfoCacherStub{}
			},
			ProofsCalled: func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{
					GetProofCalled: func(_ uint32, _ []byte) (data.HeaderProofHandler, error) {
						return nil, errors.New("proof not found")
					},
				}
			},
		}

		et, err := NewEpochStartTrigger(args)
		require.Nil(t, err)
		defer func() {
			_ = et.Close()
		}()

		metaBlockHash := []byte("metablock-hash")
		et.receivedMetaBlock(&block.MetaBlock{
			Nonce:      10,
			Round:      42,
			Epoch:      6,
			EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		}, metaBlockHash)

		time.Sleep(50 * time.Millisecond)

		require.Equal(t, int32(1), proofRequested.Load())
		requestedHashMut.Lock()
		require.Equal(t, metaBlockHash, requestedHash)
		require.Equal(t, uint32(6), requestedEpoch)
		requestedHashMut.Unlock()

		proofRequested.Store(0)
		et.receivedMetaBlock(&block.MetaBlock{
			Nonce: 11,
			Round: 43,
			Epoch: 6,
		}, []byte("regular-metablock-hash"))

		time.Sleep(50 * time.Millisecond)
		require.Equal(t, int32(0), proofRequested.Load())

		proofRequested.Store(0)
		et.receivedMetaBlock(&block.MetaBlock{
			Nonce:      5,
			Round:      30,
			Epoch:      5,
			EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		}, []byte("old-epoch-start-hash"))

		time.Sleep(50 * time.Millisecond)
		require.Equal(t, int32(0), proofRequested.Load())
	})
}

// mutHeldFinalPools guards the plain maps backing the held-final pool stubs; the trigger reads them
// from its own goroutines, so post-construction writes must go through putHeldFinalHeader/Proof
var mutHeldFinalPools sync.RWMutex

func putHeldFinalHeader(headersByHash map[string]data.HeaderHandler, hash []byte, header data.HeaderHandler) {
	mutHeldFinalPools.Lock()
	headersByHash[string(hash)] = header
	mutHeldFinalPools.Unlock()
}

func putHeldFinalProof(proofed map[string]struct{}, hash []byte) {
	mutHeldFinalPools.Lock()
	proofed[string(hash)] = struct{}{}
	mutHeldFinalPools.Unlock()
}

func createHeldFinalTriggerArgs(
	headersByHash map[string]data.HeaderHandler,
	proofed map[string]struct{},
	withSupernova bool,
) *ArgsShardEpochStartTrigger {
	args := createMockShardEpochStartTriggerArguments()
	args.Validity = 1
	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			if flag == common.AndromedaFlag {
				return true
			}
			return withSupernova && flag == common.SupernovaFlag
		},
	}
	args.RoundHandler = &mock.RoundHandlerStub{
		IndexCalled: func() int64 {
			return 1000
		},
	}

	headersPool := &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			mutHeldFinalPools.RLock()
			defer mutHeldFinalPools.RUnlock()
			header, ok := headersByHash[string(hash)]
			if !ok {
				return nil, errors.New("header not found")
			}
			return header, nil
		},
		GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
			mutHeldFinalPools.RLock()
			defer mutHeldFinalPools.RUnlock()
			headers := make([]data.HeaderHandler, 0)
			hashes := make([][]byte, 0)
			for hash, header := range headersByHash {
				if header.GetNonce() == hdrNonce && header.GetShardID() == shardId {
					headers = append(headers, header)
					hashes = append(hashes, []byte(hash))
				}
			}
			if len(headers) == 0 {
				return nil, nil, errors.New("no headers at nonce")
			}
			return headers, hashes, nil
		},
	}
	proofsPool := &dataRetrieverMock.ProofsPoolMock{
		HasProofCalled: func(_ uint32, headerHash []byte) bool {
			mutHeldFinalPools.RLock()
			defer mutHeldFinalPools.RUnlock()
			_, ok := proofed[string(headerHash)]
			return ok
		},
		GetProofCalled: func(_ uint32, headerHash []byte) (data.HeaderProofHandler, error) {
			mutHeldFinalPools.RLock()
			defer mutHeldFinalPools.RUnlock()
			if _, ok := proofed[string(headerHash)]; !ok {
				return nil, errors.New("proof not found")
			}
			return &block.HeaderProof{HeaderHash: headerHash}, nil
		},
	}
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return headersPool
		},
		ProofsCalled: func() dataRetriever.ProofsPool {
			return proofsPool
		},
		MiniBlocksCalled: func() storage.Cacher {
			return cache.NewCacherStub()
		},
		CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
			return &vic.ValidatorInfoCacherStub{}
		},
	}

	return args
}

func newEpochStartMetaForTest(epoch uint32, nonce uint64, round uint64, prevHash []byte) *block.MetaBlock {
	return &block.MetaBlock{
		Epoch:    epoch,
		Nonce:    nonce,
		Round:    round,
		PrevHash: prevHash,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{ShardID: 0},
			},
		},
	}
}

func TestTrigger_SupernovaEpochStartActivation(t *testing.T) {
	t.Parallel()

	parentHash := []byte("parentHash")
	esHash := []byte("epochStartHash")
	childHash := []byte("childHash")

	t.Run("contested epoch start defers activation until settled", func(t *testing.T) {
		t.Parallel()

		parent := &block.MetaBlock{Nonce: 9, Round: 15}
		epochStartMeta := newEpochStartMetaForTest(1, 10, 20, parentHash)

		headersByHash := map[string]data.HeaderHandler{
			string(parentHash): parent,
			string(esHash):     epochStartMeta,
		}
		proofed := map[string]struct{}{
			string(parentHash): {},
			string(esHash):     {},
		}
		epochStartTrigger, err := NewEpochStartTrigger(createHeldFinalTriggerArgs(headersByHash, proofed, true))
		require.Nil(t, err)

		epochStartTrigger.receivedMetaBlock(epochStartMeta, esHash)
		require.False(t, epochStartTrigger.IsEpochStart())

		putHeldFinalHeader(headersByHash, childHash, &block.MetaBlock{Epoch: 1, Nonce: 11, Round: 21, PrevHash: esHash})
		putHeldFinalProof(proofed, childHash)
		epochStartTrigger.receivedProof(&block.HeaderProof{
			HeaderShardId: core.MetachainShardId,
			HeaderHash:    childHash,
		})

		require.True(t, epochStartTrigger.IsEpochStart())
		require.Equal(t, uint32(1), epochStartTrigger.MetaEpoch())
		require.Equal(t, esHash, epochStartTrigger.EpochStartMetaHdrHash())
	})

	t.Run("non contended proofed epoch start activates instantly", func(t *testing.T) {
		t.Parallel()

		parent := &block.MetaBlock{Nonce: 9, Round: 15}
		epochStartMeta := newEpochStartMetaForTest(1, 10, 16, parentHash)

		headersByHash := map[string]data.HeaderHandler{
			string(parentHash): parent,
			string(esHash):     epochStartMeta,
		}
		proofed := map[string]struct{}{
			string(parentHash): {},
			string(esHash):     {},
		}
		epochStartTrigger, err := NewEpochStartTrigger(createHeldFinalTriggerArgs(headersByHash, proofed, true))
		require.Nil(t, err)

		epochStartTrigger.receivedMetaBlock(epochStartMeta, esHash)

		require.True(t, epochStartTrigger.IsEpochStart())
		require.Equal(t, esHash, epochStartTrigger.EpochStartMetaHdrHash())
	})

	t.Run("pre Supernova keeps proof only activation", func(t *testing.T) {
		t.Parallel()

		parent := &block.MetaBlock{Nonce: 9, Round: 15}
		epochStartMeta := newEpochStartMetaForTest(1, 10, 20, parentHash)

		headersByHash := map[string]data.HeaderHandler{
			string(parentHash): parent,
			string(esHash):     epochStartMeta,
		}
		proofed := map[string]struct{}{
			string(esHash): {},
		}
		epochStartTrigger, err := NewEpochStartTrigger(createHeldFinalTriggerArgs(headersByHash, proofed, false))
		require.Nil(t, err)

		epochStartTrigger.receivedMetaBlock(epochStartMeta, esHash)

		require.True(t, epochStartTrigger.IsEpochStart())
	})
}

func TestTrigger_DisarmDeadEpochStartActivation(t *testing.T) {
	t.Parallel()

	parentHash := []byte("parentHash")
	deadHash := []byte("deadEpochStartHash")
	canonicalHash := []byte("canonicalEpochStartHash")

	t.Run("no armed activation for the epoch is a no-op", func(t *testing.T) {
		t.Parallel()

		args := createMockShardEpochStartTriggerArguments()
		epochStartTrigger, err := NewEpochStartTrigger(args)
		require.Nil(t, err)

		require.False(t, epochStartTrigger.DisarmDeadEpochStartActivation(1, deadHash))
	})

	t.Run("different armed hash is a no-op", func(t *testing.T) {
		t.Parallel()

		parent := &block.MetaBlock{Nonce: 9, Round: 15}
		epochStartMeta := newEpochStartMetaForTest(1, 10, 16, parentHash)

		headersByHash := map[string]data.HeaderHandler{
			string(parentHash): parent,
			string(deadHash):   epochStartMeta,
		}
		proofed := map[string]struct{}{
			string(parentHash): {},
			string(deadHash):   {},
		}
		epochStartTrigger, err := NewEpochStartTrigger(createHeldFinalTriggerArgs(headersByHash, proofed, true))
		require.Nil(t, err)

		epochStartTrigger.receivedMetaBlock(epochStartMeta, deadHash)
		require.True(t, epochStartTrigger.IsEpochStart())

		require.False(t, epochStartTrigger.DisarmDeadEpochStartActivation(1, []byte("otherHash")))
		require.True(t, epochStartTrigger.IsEpochStart())
	})

	t.Run("disarms, restores state and lets the canonical sibling re-arm", func(t *testing.T) {
		t.Parallel()

		parent := &block.MetaBlock{Nonce: 9, Round: 15}
		deadEpochStart := newEpochStartMetaForTest(1, 10, 16, parentHash)

		headersByHash := map[string]data.HeaderHandler{
			string(parentHash): parent,
			string(deadHash):   deadEpochStart,
		}
		proofed := map[string]struct{}{
			string(parentHash): {},
			string(deadHash):   {},
		}
		args := createHeldFinalTriggerArgs(headersByHash, proofed, true)

		removedKeys := make(map[string]int)
		registryPuts := make([][]byte, 0)
		args.Storage = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					PutCalled: func(key, value []byte) error {
						if strings.HasPrefix(string(key), common.TriggerRegistryKeyPrefix) {
							registryPuts = append(registryPuts, value)
						}
						return nil
					},
					RemoveCalled: func(key []byte) error {
						removedKeys[string(key)]++
						return nil
					},
					SearchFirstCalled: func(key []byte) ([]byte, error) {
						return nil, errors.New("not found")
					},
				}, nil
			},
		}
		epochStartTrigger, err := NewEpochStartTrigger(args)
		require.Nil(t, err)

		epochStartTrigger.receivedMetaBlock(deadEpochStart, deadHash)
		require.True(t, epochStartTrigger.IsEpochStart())
		require.Equal(t, uint32(1), epochStartTrigger.MetaEpoch())

		numRegistryPuts := len(registryPuts)
		require.True(t, epochStartTrigger.DisarmDeadEpochStartActivation(1, deadHash))

		require.False(t, epochStartTrigger.IsEpochStart())
		require.Equal(t, uint32(0), epochStartTrigger.MetaEpoch())
		require.Empty(t, epochStartTrigger.EpochStartMetaHdrHash())
		require.Empty(t, epochStartTrigger.mapFinalizedEpochs)
		require.Empty(t, epochStartTrigger.mapEpochStartHdrs)
		require.NotContains(t, epochStartTrigger.mapHashHdr, string(deadHash))
		require.NotContains(t, epochStartTrigger.mapNonceHashes[10], string(deadHash))

		epochStartIdentifier := core.EpochStartIdentifier(1)
		require.Equal(t, 2, removedKeys[epochStartIdentifier])

		require.Greater(t, len(registryPuts), numRegistryPuts)
		registry, errUnmarshal := epochStart.UnmarshalShardTrigger(args.Marshalizer, registryPuts[len(registryPuts)-1])
		require.Nil(t, errUnmarshal)
		require.False(t, registry.GetIsEpochStart())
		require.Equal(t, uint32(0), registry.GetMetaEpoch())

		require.False(t, epochStartTrigger.DisarmDeadEpochStartActivation(1, deadHash))

		canonicalEpochStart := newEpochStartMetaForTest(1, 10, 16, parentHash)
		canonicalEpochStart.TimeStamp = 1
		putHeldFinalHeader(headersByHash, canonicalHash, canonicalEpochStart)
		putHeldFinalProof(proofed, canonicalHash)
		epochStartTrigger.receivedMetaBlock(canonicalEpochStart, canonicalHash)

		require.True(t, epochStartTrigger.IsEpochStart())
		require.Equal(t, uint32(1), epochStartTrigger.MetaEpoch())
		require.Equal(t, canonicalHash, epochStartTrigger.EpochStartMetaHdrHash())
	})

	t.Run("restores the current epoch start round from storage", func(t *testing.T) {
		t.Parallel()

		parent := &block.MetaBlock{Epoch: 2, Nonce: 9, Round: 15}
		deadEpochStart := newEpochStartMetaForTest(3, 10, 16, parentHash)

		headersByHash := map[string]data.HeaderHandler{
			string(parentHash): parent,
			string(deadHash):   deadEpochStart,
		}
		proofed := map[string]struct{}{
			string(parentHash): {},
			string(deadHash):   {},
		}
		args := createHeldFinalTriggerArgs(headersByHash, proofed, true)
		args.Epoch = 2

		prevEpochStartMeta := &block.MetaBlock{Epoch: 2, Round: 55}
		prevEpochStartMetaBuff, errMarshal := args.Marshalizer.Marshal(prevEpochStartMeta)
		require.Nil(t, errMarshal)
		args.Storage = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					PutCalled: func(key, value []byte) error {
						return nil
					},
					RemoveCalled: func(key []byte) error {
						return nil
					},
					SearchFirstCalled: func(key []byte) ([]byte, error) {
						require.Equal(t, []byte(core.EpochStartIdentifier(2)), key)
						return prevEpochStartMetaBuff, nil
					},
				}, nil
			},
		}
		epochStartTrigger, err := NewEpochStartTrigger(args)
		require.Nil(t, err)

		epochStartTrigger.receivedMetaBlock(deadEpochStart, deadHash)
		require.True(t, epochStartTrigger.IsEpochStart())
		require.Equal(t, uint32(3), epochStartTrigger.MetaEpoch())
		require.Equal(t, uint64(16), epochStartTrigger.EpochStartRound())

		require.True(t, epochStartTrigger.DisarmDeadEpochStartActivation(3, deadHash))

		require.False(t, epochStartTrigger.IsEpochStart())
		require.Equal(t, uint32(2), epochStartTrigger.MetaEpoch())
		require.Equal(t, uint64(55), epochStartTrigger.EpochStartRound())
		require.Equal(t, uint64(55), epochStartTrigger.EpochFinalityAttestingRound())
	})
}

// finalityEvidenceTestHarness drives the Supernova activation gate with a fully controllable
// meta chain neighbourhood, so a proofed epoch start meta block can be presented while its parent
// and its child are absent from the pools
type finalityEvidenceTestHarness struct {
	trigger *trigger

	mutRequested          sync.Mutex
	headerHashRequests    map[string]int
	headerNonceRequests   map[uint64]int
	proofByHashRequests   map[string]int
	startOfEpochRequested int

	mutPools       sync.Mutex
	pooledHeaders  map[string]data.HeaderHandler
	headersByNonce map[uint64][]string
	pooledProofs   map[string]struct{}
}

func newFinalityEvidenceTestHarness(t *testing.T, triggerEpoch uint32) *finalityEvidenceTestHarness {
	h := &finalityEvidenceTestHarness{
		headerHashRequests:  make(map[string]int),
		headerNonceRequests: make(map[uint64]int),
		proofByHashRequests: make(map[string]int),
		pooledHeaders:       make(map[string]data.HeaderHandler),
		headersByNonce:      make(map[uint64][]string),
		pooledProofs:        make(map[string]struct{}),
	}

	args := createMockShardEpochStartTriggerArguments()
	args.Epoch = triggerEpoch
	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.AndromedaFlag || flag == common.SupernovaFlag
		},
	}
	// round handler far ahead of header rounds skips the late-broadcast wait in updateTriggerHeaderData
	args.RoundHandler = &mock.RoundHandlerStub{IndexCalled: func() int64 {
		return 100000
	}}
	args.RequestHandler = &testscommon.RequestHandlerStub{
		RequestMetaHeaderCalled: func(hash []byte) {
			h.mutRequested.Lock()
			h.headerHashRequests[string(hash)]++
			h.mutRequested.Unlock()
		},
		RequestMetaHeaderByNonceCalled: func(nonce uint64) {
			h.mutRequested.Lock()
			h.headerNonceRequests[nonce]++
			h.mutRequested.Unlock()
		},
		RequestEquivalentProofByHashForEpochCalled: func(_ uint32, headerHash []byte, _ uint32) {
			h.mutRequested.Lock()
			h.proofByHashRequests[string(headerHash)]++
			h.mutRequested.Unlock()
		},
		RequestStartOfEpochMetaBlockCalled: func(_ uint32) {
			h.mutRequested.Lock()
			h.startOfEpochRequested++
			h.mutRequested.Unlock()
		},
	}
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					h.mutPools.Lock()
					defer h.mutPools.Unlock()
					header, found := h.pooledHeaders[string(hash)]
					if !found {
						return nil, errors.New("header not found")
					}
					return header, nil
				},
				GetHeaderByNonceAndShardIdCalled: func(nonce uint64, _ uint32) ([]data.HeaderHandler, [][]byte, error) {
					h.mutPools.Lock()
					defer h.mutPools.Unlock()
					hashes, found := h.headersByNonce[nonce]
					if !found {
						return nil, nil, errors.New("no headers at nonce")
					}
					headers := make([]data.HeaderHandler, 0, len(hashes))
					headerHashes := make([][]byte, 0, len(hashes))
					for _, hash := range hashes {
						headers = append(headers, h.pooledHeaders[hash])
						headerHashes = append(headerHashes, []byte(hash))
					}
					return headers, headerHashes, nil
				},
			}
		},
		MiniBlocksCalled: func() storage.Cacher {
			return cache.NewCacherStub()
		},
		CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
			return &vic.ValidatorInfoCacherStub{}
		},
		ProofsCalled: func() dataRetriever.ProofsPool {
			return &dataRetrieverMock.ProofsPoolMock{
				GetProofCalled: func(_ uint32, headerHash []byte) (data.HeaderProofHandler, error) {
					h.mutPools.Lock()
					defer h.mutPools.Unlock()
					if _, found := h.pooledProofs[string(headerHash)]; !found {
						return nil, errors.New("proof not found")
					}
					return &block.HeaderProof{HeaderHash: headerHash, HeaderShardId: core.MetachainShardId}, nil
				},
				HasProofCalled: func(_ uint32, headerHash []byte) bool {
					h.mutPools.Lock()
					defer h.mutPools.Unlock()
					_, found := h.pooledProofs[string(headerHash)]
					return found
				},
			}
		},
	}

	et, err := NewEpochStartTrigger(args)
	require.Nil(t, err)
	t.Cleanup(func() {
		_ = et.Close()
	})
	h.trigger = et

	return h
}

func (h *finalityEvidenceTestHarness) setRetryInterval(interval time.Duration) {
	h.trigger.mutPendingEpochStartData.Lock()
	h.trigger.pendingProofRetryInterval = interval
	h.trigger.mutPendingEpochStartData.Unlock()
}

func (h *finalityEvidenceTestHarness) putHeader(hash []byte, header data.HeaderHandler) {
	h.mutPools.Lock()
	h.pooledHeaders[string(hash)] = header
	h.headersByNonce[header.GetNonce()] = append(h.headersByNonce[header.GetNonce()], string(hash))
	h.mutPools.Unlock()
}

func (h *finalityEvidenceTestHarness) putProof(hash []byte) {
	h.mutPools.Lock()
	h.pooledProofs[string(hash)] = struct{}{}
	h.mutPools.Unlock()
}

func (h *finalityEvidenceTestHarness) numHeaderHashRequests(hash []byte) int {
	h.mutRequested.Lock()
	defer h.mutRequested.Unlock()

	return h.headerHashRequests[string(hash)]
}

func (h *finalityEvidenceTestHarness) numHeaderNonceRequests(nonce uint64) int {
	h.mutRequested.Lock()
	defer h.mutRequested.Unlock()

	return h.headerNonceRequests[nonce]
}

func (h *finalityEvidenceTestHarness) numProofRequests(hash []byte) int {
	h.mutRequested.Lock()
	defer h.mutRequested.Unlock()

	return h.proofByHashRequests[string(hash)]
}

func (h *finalityEvidenceTestHarness) pendingFinalityEvidence() int {
	h.trigger.mutPendingEpochStartData.Lock()
	defer h.trigger.mutPendingEpochStartData.Unlock()

	return len(h.trigger.pendingFinalityEvidence)
}

// TestTrigger_EpochStartNotHeldFinalRequestsNeighbours covers the sync edge case in which a node
// receives a proofed epoch start meta block long after its neighbourhood has left the pools. The
// activation gate reads the pools alone, so without an explicit request the trigger would stay in
// the old epoch forever and every epoch start shard block would fail verification.
func TestTrigger_EpochStartNotHeldFinalRequestsNeighbours(t *testing.T) {
	t.Parallel()

	var (
		parentHash      = []byte("meta-parent-hash")
		epochStartHash  = []byte("epoch-start-meta-hash")
		childHash       = []byte("meta-child-hash")
		epochStartNonce = uint64(6002)
		epochStartRound = uint64(6807)
	)

	newEpochStartMeta := func() *block.MetaBlock {
		return &block.MetaBlock{
			Nonce:      epochStartNonce,
			Round:      epochStartRound,
			Epoch:      8,
			PrevHash:   parentHash,
			EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		}
	}

	t.Run("neighbourhood absent from pools requests parent and child, epoch does not advance", func(t *testing.T) {
		t.Parallel()

		h := newFinalityEvidenceTestHarness(t, 7)
		metaHdr := newEpochStartMeta()
		h.putHeader(epochStartHash, metaHdr)
		h.putProof(epochStartHash)

		h.trigger.receivedMetaBlock(metaHdr, epochStartHash)

		require.Eventually(t, func() bool {
			return h.numHeaderHashRequests(parentHash) >= 1 &&
				h.numHeaderNonceRequests(epochStartNonce+1) >= 1
		}, time.Second, 5*time.Millisecond)

		require.Equal(t, 1, h.pendingFinalityEvidence())
		// the gate is not satisfied, so the trigger must stay put
		require.Equal(t, uint32(7), h.trigger.MetaEpoch())
		require.False(t, h.trigger.IsEpochStart())
	})

	t.Run("parent present without proof requests only the parent proof", func(t *testing.T) {
		t.Parallel()

		h := newFinalityEvidenceTestHarness(t, 7)
		metaHdr := newEpochStartMeta()
		h.putHeader(epochStartHash, metaHdr)
		h.putProof(epochStartHash)
		h.putHeader(parentHash, &block.MetaBlock{Nonce: epochStartNonce - 1, Round: epochStartRound - 1, Epoch: 7})

		h.trigger.receivedMetaBlock(metaHdr, epochStartHash)

		require.Eventually(t, func() bool {
			return h.numProofRequests(parentHash) >= 1
		}, time.Second, 5*time.Millisecond)

		// the header is already held, asking for it again would be wasted traffic
		require.Zero(t, h.numHeaderHashRequests(parentHash))
		require.Equal(t, uint32(7), h.trigger.MetaEpoch())
	})

	t.Run("evidence arriving later lets the retry pass activate the trigger", func(t *testing.T) {
		t.Parallel()

		h := newFinalityEvidenceTestHarness(t, 7)
		h.setRetryInterval(10 * time.Millisecond)

		metaHdr := newEpochStartMeta()
		h.putHeader(epochStartHash, metaHdr)
		h.putProof(epochStartHash)

		h.trigger.receivedMetaBlock(metaHdr, epochStartHash)

		require.Eventually(t, func() bool {
			return h.pendingFinalityEvidence() == 1
		}, time.Second, 5*time.Millisecond)
		require.Equal(t, uint32(7), h.trigger.MetaEpoch())

		// the requested child finally reaches the pools, proofed
		h.putHeader(childHash, &block.MetaBlock{
			Nonce:    epochStartNonce + 1,
			Round:    epochStartRound + 1,
			Epoch:    8,
			PrevHash: epochStartHash,
		})
		h.putProof(childHash)

		require.Eventually(t, func() bool {
			return h.trigger.MetaEpoch() == 8
		}, 2*time.Second, 5*time.Millisecond)

		require.True(t, h.trigger.IsEpochStart())
		require.Equal(t, epochStartRound, h.trigger.EpochStartRound())
		require.Equal(t, epochStartHash, h.trigger.EpochStartMetaHdrHash())

		// activation clears the pending entry, so the retry loop goes back to sleep
		require.Eventually(t, func() bool {
			return h.pendingFinalityEvidence() == 0
		}, time.Second, 5*time.Millisecond)
	})

	t.Run("contended parent is not worth a proof request, only the child is", func(t *testing.T) {
		t.Parallel()

		h := newFinalityEvidenceTestHarness(t, 7)
		metaHdr := newEpochStartMeta()
		h.putHeader(epochStartHash, metaHdr)
		h.putProof(epochStartHash)
		// a round gap across the boundary: the parent settles nothing however well proofed
		h.putHeader(parentHash, &block.MetaBlock{Nonce: epochStartNonce - 1, Round: epochStartRound - 5, Epoch: 7})

		h.trigger.receivedMetaBlock(metaHdr, epochStartHash)

		require.Eventually(t, func() bool {
			return h.numHeaderNonceRequests(epochStartNonce+1) >= 1
		}, time.Second, 5*time.Millisecond)

		require.Zero(t, h.numProofRequests(parentHash))
		require.Zero(t, h.numHeaderHashRequests(parentHash))
	})

	t.Run("a competing candidate settling the epoch clears the pending entry", func(t *testing.T) {
		t.Parallel()

		h := newFinalityEvidenceTestHarness(t, 7)
		siblingHash := []byte("competing-epoch-start-hash")

		// the candidate the node cannot settle
		deadEpochStart := newEpochStartMeta()
		h.putHeader(epochStartHash, deadEpochStart)
		h.putProof(epochStartHash)
		h.trigger.receivedMetaBlock(deadEpochStart, epochStartHash)

		require.Eventually(t, func() bool {
			return h.pendingFinalityEvidence() == 1
		}, time.Second, 5*time.Millisecond)

		// a sibling of the same epoch arrives fully settled and wins
		sibling := &block.MetaBlock{
			Nonce:      epochStartNonce,
			Round:      epochStartRound + 1,
			Epoch:      8,
			PrevHash:   []byte("other-parent"),
			EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		}
		siblingChild := &block.MetaBlock{
			Nonce:    epochStartNonce + 1,
			Round:    epochStartRound + 2,
			Epoch:    8,
			PrevHash: siblingHash,
		}
		h.putHeader(siblingHash, sibling)
		h.putProof(siblingHash)
		h.putHeader([]byte("sibling-child"), siblingChild)
		h.putProof([]byte("sibling-child"))

		h.trigger.receivedMetaBlock(sibling, siblingHash)

		require.Eventually(t, func() bool {
			return h.trigger.MetaEpoch() == 8
		}, time.Second, 5*time.Millisecond)

		// the settled epoch needs no neighbourhood for any of its candidates
		require.Eventually(t, func() bool {
			return h.pendingFinalityEvidence() == 0
		}, time.Second, 5*time.Millisecond)
	})

	t.Run("held final on arrival activates without requesting anything", func(t *testing.T) {
		t.Parallel()

		h := newFinalityEvidenceTestHarness(t, 7)
		metaHdr := newEpochStartMeta()
		h.putHeader(epochStartHash, metaHdr)
		h.putProof(epochStartHash)
		h.putHeader(childHash, &block.MetaBlock{
			Nonce:    epochStartNonce + 1,
			Round:    epochStartRound + 1,
			Epoch:    8,
			PrevHash: epochStartHash,
		})
		h.putProof(childHash)

		h.trigger.receivedMetaBlock(metaHdr, epochStartHash)

		require.Eventually(t, func() bool {
			return h.trigger.MetaEpoch() == 8
		}, time.Second, 5*time.Millisecond)

		require.Zero(t, h.pendingFinalityEvidence())
		require.Zero(t, h.numHeaderHashRequests(parentHash))
		require.Zero(t, h.numHeaderNonceRequests(epochStartNonce+1))
	})
}
