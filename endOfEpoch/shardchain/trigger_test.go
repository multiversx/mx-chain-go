package shardchain

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

func createMockShardEndOfEpochTriggerArguments() *ArgsShardEndOfEpochTrigger {
	return &ArgsShardEndOfEpochTrigger{
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
		RequestHandler: &mock.RequestHandlerStub{},
	}
}

func TestNewEndOfEpochTrigger_NilArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	eoet, err := NewEndOfEpochTrigger(nil)

	assert.Nil(t, eoet)
	assert.Equal(t, endOfEpoch.ErrNilArgsNewShardEndOfEpochTrigger, err)
}

func TestNewEndOfEpochTrigger_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEndOfEpochTriggerArguments()
	args.Hasher = nil
	eoet, err := NewEndOfEpochTrigger(args)

	assert.Nil(t, eoet)
	assert.Equal(t, endOfEpoch.ErrNilHasher, err)
}

func TestNewEndOfEpochTrigger_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEndOfEpochTriggerArguments()
	args.Marshalizer = nil
	eoet, err := NewEndOfEpochTrigger(args)

	assert.Nil(t, eoet)
	assert.Equal(t, endOfEpoch.ErrNilMarshalizer, err)
}

func TestNewEndOfEpochTrigger_NilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEndOfEpochTriggerArguments()
	args.HeaderValidator = nil
	eoet, err := NewEndOfEpochTrigger(args)

	assert.Nil(t, eoet)
	assert.Equal(t, endOfEpoch.ErrNilHeaderValidator, err)
}

func TestNewEndOfEpochTrigger_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEndOfEpochTriggerArguments()
	args.DataPool = nil
	eoet, err := NewEndOfEpochTrigger(args)

	assert.Nil(t, eoet)
	assert.Equal(t, endOfEpoch.ErrNilDataPoolsHolder, err)
}

func TestNewEndOfEpochTrigger_NilStorageShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEndOfEpochTriggerArguments()
	args.Storage = nil
	eoet, err := NewEndOfEpochTrigger(args)

	assert.Nil(t, eoet)
	assert.Equal(t, endOfEpoch.ErrNilStorageService, err)
}

func TestNewEndOfEpochTrigger_NilRequestHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEndOfEpochTriggerArguments()
	args.RequestHandler = nil
	eoet, err := NewEndOfEpochTrigger(args)

	assert.Nil(t, eoet)
	assert.Equal(t, endOfEpoch.ErrNilRequestHandler, err)
}

func TestNewEndOfEpochTrigger_NilMetaBlockPoolShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEndOfEpochTriggerArguments()
	args.DataPool = &mock.PoolsHolderStub{
		MetaBlocksCalled: func() storage.Cacher {
			return nil
		},
	}
	eoet, err := NewEndOfEpochTrigger(args)

	assert.Nil(t, eoet)
	assert.Equal(t, endOfEpoch.ErrNilMetaBlocksPool, err)
}

func TestNewEndOfEpochTrigger_NilHeadersNonceShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEndOfEpochTriggerArguments()
	args.DataPool = &mock.PoolsHolderStub{
		MetaBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{}
		},
		HeadersNoncesCalled: func() dataRetriever.Uint64SyncMapCacher {
			return nil
		},
	}
	eoet, err := NewEndOfEpochTrigger(args)

	assert.Nil(t, eoet)
	assert.Equal(t, endOfEpoch.ErrNilHeaderNoncesPool, err)
}

func TestNewEndOfEpochTrigger_NilUint64ConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEndOfEpochTriggerArguments()
	args.Uint64Converter = nil
	eoet, err := NewEndOfEpochTrigger(args)

	assert.Nil(t, eoet)
	assert.Equal(t, endOfEpoch.ErrNilUint64Converter, err)
}

func TestNewEndOfEpochTrigger_NilMetaBlockUnitShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEndOfEpochTriggerArguments()
	args.Storage = &mock.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
		},
	}
	eoet, err := NewEndOfEpochTrigger(args)

	assert.Nil(t, eoet)
	assert.Equal(t, endOfEpoch.ErrNilMetaHdrStorage, err)
}

func TestNewEndOfEpochTrigger_NilMetaNonceHashStorageShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockShardEndOfEpochTriggerArguments()
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
	eoet, err := NewEndOfEpochTrigger(args)

	assert.Nil(t, eoet)
	assert.Equal(t, endOfEpoch.ErrNilMetaNonceHashStorage, err)
}

func TestNewEndOfEpochTrigger_ShouldOk(t *testing.T) {
	t.Parallel()

	args := createMockShardEndOfEpochTriggerArguments()
	eoet, err := NewEndOfEpochTrigger(args)

	assert.NotNil(t, eoet)
	assert.Nil(t, err)
}

func TestTrigger_ReceivedHeaderNotEndOfEpoch(t *testing.T) {
	t.Parallel()

	args := createMockShardEndOfEpochTriggerArguments()
	args.Validity = 2
	args.Finality = 2
	eoet, _ := NewEndOfEpochTrigger(args)

	hash := []byte("hash")
	header := &block.MetaBlock{Nonce: 100}
	header.EndOfEpoch.LastFinalizedHeaders = []block.FinalizedHeaders{{ShardId: 0, RootHash: hash, HeaderHash: hash}}
	eoet.ReceivedHeader(header)

	assert.False(t, eoet.IsEndOfEpoch())
}

func TestTrigger_ReceivedHeaderIsEndOfEpochTrue(t *testing.T) {
	t.Parallel()

	args := createMockShardEndOfEpochTriggerArguments()
	args.Validity = 0
	args.Finality = 2
	eoet, _ := NewEndOfEpochTrigger(args)

	hash := []byte("hash")
	header := &block.MetaBlock{Nonce: 100, Epoch: 1}
	header.EndOfEpoch.LastFinalizedHeaders = []block.FinalizedHeaders{{ShardId: 0, RootHash: hash, HeaderHash: hash}}
	eoet.ReceivedHeader(header)

	header = &block.MetaBlock{Nonce: 101, Epoch: 1}
	eoet.ReceivedHeader(header)

	assert.True(t, eoet.IsEndOfEpoch())
}

func TestTrigger_Epoch(t *testing.T) {
	t.Parallel()

	epoch := uint32(1)
	args := createMockShardEndOfEpochTriggerArguments()
	args.Epoch = epoch
	eoet, _ := NewEndOfEpochTrigger(args)

	currentEpoch := eoet.Epoch()
	assert.Equal(t, epoch, currentEpoch)
}

func TestTrigger_ProcessedAndRevert(t *testing.T) {
	t.Parallel()

	args := createMockShardEndOfEpochTriggerArguments()
	args.Validity = 0
	args.Finality = 0
	et, _ := NewEndOfEpochTrigger(args)

	hash := []byte("hash")
	header := &block.MetaBlock{Nonce: 100, Round: 100}
	header.EndOfEpoch.LastFinalizedHeaders = []block.FinalizedHeaders{{ShardId: 0, RootHash: hash, HeaderHash: hash}}
	et.ReceivedHeader(header)
	assert.True(t, et.IsEndOfEpoch())
	assert.Equal(t, header.Round, et.EpochStartRound())

	et.Processed()
	assert.False(t, et.isEndOfEpoch)
	assert.False(t, et.newEpochHdrReceived)

	et.Revert()
	assert.True(t, et.isEndOfEpoch)
	assert.True(t, et.newEpochHdrReceived)
}
