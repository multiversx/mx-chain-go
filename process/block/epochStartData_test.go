package block_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	blproc "github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func createMockEpochStartCreatorArguments() blproc.ArgsNewEpochStartDataCreator {
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	startHeaders := createGenesisBlocks(shardCoordinator)
	argsNewEpochStartDataCreator := blproc.ArgsNewEpochStartDataCreator{
		Marshalizer:       &mock.MarshalizerMock{},
		Hasher:            &mock.HasherStub{},
		Store:             &mock.ChainStorerMock{},
		DataPool:          initMetaDataPool(),
		BlockTracker:      mock.NewBlockTrackerMock(shardCoordinator, startHeaders),
		ShardCoordinator:  shardCoordinator,
		EpochStartTrigger: &mock.EpochStartTriggerStub{},
	}
	return argsNewEpochStartDataCreator
}

func TestEpochStartCreator_getLastFinalizedMetaHashForShardMetaHashNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	epoch, _ := blproc.NewEpochStartDataCreator(arguments)
	round := uint64(10)

	shardHdr := &block.Header{Round: round}
	last, lastFinal, shardHdrs, err := epoch.GetLastFinalizedMetaHashForShard(shardHdr)
	assert.Nil(t, last)
	assert.Nil(t, lastFinal)
	assert.Nil(t, shardHdrs)
	assert.Equal(t, dataRetriever.ErrNilHeadersStorage, err)
}

func TestShardProcessor_getLastFinalizedMetaHashForShardShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		IsEpochStartCalled: func() bool {
			return false
		},
	}

	dPool := initMetaDataPool()
	dPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	dPool.HeadersCalled = func() dataRetriever.HeadersPool {
		cs := &mock.HeadersCacherStub{}
		cs.RegisterHandlerCalled = func(i func(header data.HeaderHandler, key []byte)) {
		}
		cs.GetHeaderByHashCalled = func(hash []byte) (handler data.HeaderHandler, e error) {
			return &block.Header{
				PrevHash:         []byte("hash1"),
				Nonce:            2,
				Round:            2,
				PrevRandSeed:     []byte("roothash"),
				MiniBlockHeaders: []block.MiniBlockHeader{{Hash: []byte("hash1"), SenderShardID: 1}},
				MetaBlockHashes:  [][]byte{[]byte("hash1"), []byte("hash2")},
			}, nil
		}

		cs.LenCalled = func() int {
			return 0
		}
		cs.RemoveHeaderByHashCalled = func(key []byte) {}
		cs.NoncesCalled = func(shardId uint32) []uint64 {
			return []uint64{1, 2}
		}
		cs.MaxSizeCalled = func() int {
			return 1000
		}
		return cs
	}

	arguments.DataPool = dPool

	epoch, _ := blproc.NewEpochStartDataCreator(arguments)
	round := uint64(10)
	nonce := uint64(1)

	shardHdr := &block.Header{
		Round:           round,
		Nonce:           nonce,
		MetaBlockHashes: [][]byte{[]byte("mb_hash1")},
	}
	last, lastFinal, shardHdrs, err := epoch.GetLastFinalizedMetaHashForShard(shardHdr)
	assert.NotNil(t, last)
	assert.NotNil(t, lastFinal)
	assert.NotNil(t, shardHdrs)
	assert.Nil(t, err)
}
