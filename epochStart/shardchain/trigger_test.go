package shardchain

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	vic "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				return testscommon.NewCacherStub()
			},
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vic.ValidatorInfoCacherStub{}
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
			return testscommon.NewCacherStub()
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
			return &testscommon.CacherStub{
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
	assert.Equal(t, et.epochStartShardHeader.GetEpoch(), prevEpochHdr.Epoch)
}

func TestTrigger_ReceivedHeaderChangeEpochFinalityAttestingRound(t *testing.T) {
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
				return testscommon.NewCacherStub()
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
