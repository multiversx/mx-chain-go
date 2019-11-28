package shardchain

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

func createMockShardEpochStartTriggerArguments() *ArgsShardEpochStartTrigger {
	return &ArgsShardEpochStartTrigger{
		Marshalizer: &mock.MarshalizerMock{},
		Hasher:      &mock.HasherMock{},
		HeaderValidator: &mock.HeaderValidatorStub{
			IsHeaderConstructionValidCalled: func(currHdr, prevHdr data.HeaderHandler) error {
				return nil
			},
		},
		Uint64Converter: &mock.Uint64ByteSliceConverterMock{},
		DataPool: &mock.PoolsHolderStub{
			MetaBlocksCalled: func() storage.Cacher {
				return &mock.CacherStub{
					PeekCalled: func(key []byte) (value interface{}, ok bool) {
						return nil, true
					},
				}
			},
			HeadersNoncesCalled: func() dataRetriever.Uint64SyncMapCacher {
				return &mock.Uint64SyncMapCacherStub{
					GetCalled: func(nonce uint64) (hashMap dataRetriever.ShardIdHashMap, b bool) {
						return &mock.ShardIdHasMapStub{LoadCalled: func(shardId uint32) (bytes []byte, b bool) {
							return []byte("hash"), true
						}}, true
					},
				}
			},
		},
		Storage: &mock.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
				return &mock.StorerStub{
					GetCalled: func(key []byte) (bytes []byte, err error) {
						return []byte("hash"), nil
					},
				}
			},
		},
		RequestHandler:     &mock.RequestHandlerStub{},
		EpochStartNotifier: &mock.EpochStartNotifierStub{},
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

func TestNewEpochStartTrigger_NilMetaBlockPoolShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.DataPool = &mock.PoolsHolderStub{
		MetaBlocksCalled: func() storage.Cacher {
			return nil
		},
	}
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilMetaBlocksPool, err)
}

func TestNewEpochStartTrigger_NilHeadersNonceShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.DataPool = &mock.PoolsHolderStub{
		MetaBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{}
		},
		HeadersNoncesCalled: func() dataRetriever.Uint64SyncMapCacher {
			return nil
		},
	}
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilHeaderNoncesPool, err)
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

func TestNewEpochStartTrigger_NilMetaBlockUnitShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.Storage = &mock.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
		},
	}
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilMetaHdrStorage, err)
}

func TestNewEpochStartTrigger_NilMetaNonceHashStorageShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.Storage = &mock.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			switch unitType {
			case dataRetriever.MetaHdrNonceHashDataUnit:
				return nil
			default:
				return &mock.StorerStub{}
			}
		},
	}
	epochStartTrigger, err := NewEpochStartTrigger(args)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilMetaNonceHashStorage, err)
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
	header.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardId: 0, RootHash: hash, HeaderHash: hash}}
	epochStartTrigger.ReceivedHeader(header)

	assert.False(t, epochStartTrigger.IsEpochStart())
}

func TestTrigger_ReceivedHeaderIsEpochStartTrue(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.Validity = 1
	args.Finality = 2
	epochStartTrigger, _ := NewEpochStartTrigger(args)

	oldEpHeader := &block.MetaBlock{Nonce: 99, Epoch: 0}
	prevHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, oldEpHeader)

	hash := []byte("hash")
	header := &block.MetaBlock{Nonce: 100, Epoch: 1, PrevHash: prevHash}
	header.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardId: 0, RootHash: hash, HeaderHash: hash}}
	epochStartTrigger.ReceivedHeader(header)
	epochStartTrigger.ReceivedHeader(oldEpHeader)

	prevHash, _ = core.CalculateHash(args.Marshalizer, args.Hasher, header)
	header = &block.MetaBlock{Nonce: 101, Epoch: 1, PrevHash: prevHash}
	epochStartTrigger.ReceivedHeader(header)

	prevHash, _ = core.CalculateHash(args.Marshalizer, args.Hasher, header)
	header = &block.MetaBlock{Nonce: 102, Epoch: 1, PrevHash: prevHash}
	epochStartTrigger.ReceivedHeader(header)

	assert.True(t, epochStartTrigger.IsEpochStart())
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

func TestTrigger_ProcessedAndRevert(t *testing.T) {
	t.Parallel()

	args := createMockShardEpochStartTriggerArguments()
	args.Validity = 0
	args.Finality = 0
	args.EpochStartNotifier = &mock.EpochStartNotifierStub{NotifyAllCalled: func(hdr data.HeaderHandler) {}}
	et, _ := NewEpochStartTrigger(args)

	hash := []byte("hash")
	epochStartRound := uint64(100)
	header := &block.MetaBlock{Nonce: 100, Round: epochStartRound, Epoch: 1}
	header.EpochStart.LastFinalizedHeaders = []block.EpochStartShardData{{ShardId: 0, RootHash: hash, HeaderHash: hash}}
	et.ReceivedHeader(header)
	header = &block.MetaBlock{Nonce: 101, Round: epochStartRound + 1, Epoch: 1}
	et.ReceivedHeader(header)

	assert.True(t, et.IsEpochStart())
	assert.Equal(t, epochStartRound, et.EpochStartRound())

	et.SetProcessed(&block.Header{EpochStartMetaHash: []byte("metahash")})
	assert.False(t, et.isEpochStart)
	assert.False(t, et.newEpochHdrReceived)

	et.Revert()
	assert.True(t, et.isEpochStart)
	assert.True(t, et.newEpochHdrReceived)
}
