package poolsCleaner

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/assert"
)

func createMockArgTxsPoolsCleaner() ArgTxsPoolsCleaner {
	return ArgTxsPoolsCleaner{
		ArgBasePoolsCleaner: ArgBasePoolsCleaner{
			RoundHandler:                   &mock.RoundHandlerMock{},
			ShardCoordinator:               mock.NewMultipleShardsCoordinatorMock(),
			MaxRoundsToKeepUnprocessedData: 1,
		},
		AddressPubkeyConverter: &testscommon.PubkeyConverterStub{},
		DataPool:               dataRetrieverMock.NewPoolsHolderMock(),
	}
}

func TestNewTxsPoolsCleaner_NilAddrConverterErr(t *testing.T) {
	t.Parallel()

	args := createMockArgTxsPoolsCleaner()
	args.AddressPubkeyConverter = nil
	txsPoolsCleaner, err := NewTxsPoolsCleaner(args)
	assert.Nil(t, txsPoolsCleaner)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewTxsPoolsCleaner_NilDataPoolHolderErr(t *testing.T) {
	t.Parallel()

	args := createMockArgTxsPoolsCleaner()
	args.DataPool = nil
	txsPoolsCleaner, err := NewTxsPoolsCleaner(args)
	assert.Nil(t, txsPoolsCleaner)
	assert.Equal(t, process.ErrNilPoolsHolder, err)
}

func TestNewTxsPoolsCleaner_NilTxsPoolErr(t *testing.T) {
	t.Parallel()

	args := createMockArgTxsPoolsCleaner()
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return nil
		},
	}
	txsPoolsCleaner, err := NewTxsPoolsCleaner(args)
	assert.Nil(t, txsPoolsCleaner)
	assert.Equal(t, process.ErrNilTransactionPool, err)
}

func TestNewTxsPoolsCleaner_NilRewardTxsPoolErr(t *testing.T) {
	t.Parallel()

	args := createMockArgTxsPoolsCleaner()
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return nil
		},
	}
	txsPoolsCleaner, err := NewTxsPoolsCleaner(args)
	assert.Nil(t, txsPoolsCleaner)
	assert.Equal(t, process.ErrNilRewardTxDataPool, err)
}

func TestNewTxsPoolsCleaner_NilUnsignedTxsPoolErr(t *testing.T) {
	t.Parallel()

	args := createMockArgTxsPoolsCleaner()
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return nil
		},
	}
	txsPoolsCleaner, err := NewTxsPoolsCleaner(args)
	assert.Nil(t, txsPoolsCleaner)
	assert.Equal(t, process.ErrNilUnsignedTxDataPool, err)
}

func TestNewTxsPoolsCleaner_NilRoundHandlerErr(t *testing.T) {
	t.Parallel()

	args := createMockArgTxsPoolsCleaner()
	args.RoundHandler = nil
	txsPoolsCleaner, err := NewTxsPoolsCleaner(args)
	assert.Nil(t, txsPoolsCleaner)
	assert.Equal(t, process.ErrNilRoundHandler, err)
}

func TestNewTxsPoolsCleaner_NilShardCoordinatorErr(t *testing.T) {
	t.Parallel()

	args := createMockArgTxsPoolsCleaner()
	args.ShardCoordinator = nil
	txsPoolsCleaner, err := NewTxsPoolsCleaner(args)
	assert.Nil(t, txsPoolsCleaner)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewTxsPoolsCleaner_InvalidMaxRoundsToKeepUnprocessedDataShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgTxsPoolsCleaner()
	args.MaxRoundsToKeepUnprocessedData = 0
	txsPoolsCleaner, err := NewTxsPoolsCleaner(args)
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
	assert.True(t, strings.Contains(err.Error(), "MaxRoundsToKeepUnprocessedData"))
	assert.Nil(t, txsPoolsCleaner)
}

func TestNewTxsPoolsCleaner_ShouldWork(t *testing.T) {
	t.Parallel()

	txsPoolsCleaner, err := NewTxsPoolsCleaner(createMockArgTxsPoolsCleaner())
	assert.Nil(t, err)
	assert.NotNil(t, txsPoolsCleaner)
}

func TestGetShardFromAddress(t *testing.T) {
	t.Parallel()

	args := createMockArgTxsPoolsCleaner()
	addrLen := 64
	args.AddressPubkeyConverter = &testscommon.PubkeyConverterStub{
		LenCalled: func() int {
			return addrLen
		},
	}
	expectedShard := uint32(2)
	args.ShardCoordinator = &mock.CoordinatorStub{
		ComputeIdCalled: func(address []byte) uint32 {
			return expectedShard
		},
	}
	txsPoolsCleaner, _ := NewTxsPoolsCleaner(args)

	emptyAddr := make([]byte, addrLen)
	result, err := txsPoolsCleaner.getShardFromAddress(emptyAddr)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), result)

	result, err = txsPoolsCleaner.getShardFromAddress([]byte("123"))
	assert.Nil(t, err)
	assert.Equal(t, expectedShard, result)
}

func TestReceivedBlockTx_ShouldBeAddedInMapTxsRounds(t *testing.T) {
	t.Parallel()

	args := createMockArgTxsPoolsCleaner()
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
					return testscommon.NewCacherMock()
				},
			}
		},
	}
	txsPoolsCleaner, _ := NewTxsPoolsCleaner(args)

	txWrap := &txcache.WrappedTransaction{
		Tx:            &transaction.Transaction{},
		SenderShardID: 2,
	}
	txBlockKey := []byte("key")
	txsPoolsCleaner.receivedBlockTx(txBlockKey, txWrap)
	assert.NotNil(t, txsPoolsCleaner.mapTxsRounds[string(txBlockKey)])
}

func TestReceivedRewardTx_ShouldBeAddedInMapTxsRounds(t *testing.T) {
	t.Parallel()

	sndAddr := []byte("sndAddr")
	args := createMockArgTxsPoolsCleaner()
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
					return testscommon.NewCacherMock()
				},
			}
		},
	}
	txsPoolsCleaner, _ := NewTxsPoolsCleaner(args)

	txKey := []byte("key")
	tx := &transaction.Transaction{
		SndAddr: sndAddr,
	}
	txsPoolsCleaner.receivedRewardTx(txKey, tx)
	assert.NotNil(t, txsPoolsCleaner.mapTxsRounds[string(txKey)])
}

func TestReceivedUnsignedTx_ShouldBeAddedInMapTxsRounds(t *testing.T) {
	t.Parallel()

	sndAddr := []byte("sndAddr")
	args := createMockArgTxsPoolsCleaner()
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
					return testscommon.NewCacherMock()
				},
			}
		},
	}
	args.ShardCoordinator = &mock.CoordinatorStub{
		ComputeIdCalled: func(address []byte) uint32 {
			return 2
		},
	}
	txsPoolsCleaner, _ := NewTxsPoolsCleaner(args)

	txKey := []byte("key")
	tx := &transaction.Transaction{
		SndAddr: sndAddr,
	}
	txsPoolsCleaner.receivedUnsignedTx(txKey, tx)
	assert.NotNil(t, txsPoolsCleaner.mapTxsRounds[string(txKey)])
}

func TestCleanTxsPoolsIfNeeded_CannotFindTxInPoolShouldBeRemovedFromMap(t *testing.T) {
	t.Parallel()

	sndAddr := []byte("sndAddr")
	args := createMockArgTxsPoolsCleaner()
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
					return testscommon.NewCacherMock()
				},
			}
		},
	}
	args.ShardCoordinator = &mock.CoordinatorStub{
		ComputeIdCalled: func(address []byte) uint32 {
			return 2
		},
	}
	txsPoolsCleaner, _ := NewTxsPoolsCleaner(args)

	txKey := []byte("key")
	tx := &transaction.Transaction{
		SndAddr: sndAddr,
	}
	txsPoolsCleaner.receivedUnsignedTx(txKey, tx)

	numTxsInMap := txsPoolsCleaner.cleanTxsPoolsIfNeeded()
	assert.Equal(t, 0, numTxsInMap)
}

func TestCleanTxsPoolsIfNeeded_RoundDiffTooSmallShouldNotBeRemoved(t *testing.T) {
	t.Parallel()

	sndAddr := []byte("sndAddr")
	args := createMockArgTxsPoolsCleaner()
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
					return &testscommon.CacherStub{
						GetCalled: func(key []byte) (value interface{}, ok bool) {
							return nil, true
						},
					}
				},
			}
		},
	}
	args.ShardCoordinator = &mock.CoordinatorStub{
		ComputeIdCalled: func(address []byte) uint32 {
			return 2
		},
	}
	txsPoolsCleaner, _ := NewTxsPoolsCleaner(args)

	txKey := []byte("key")
	tx := &transaction.Transaction{
		SndAddr: sndAddr,
	}
	txsPoolsCleaner.receivedUnsignedTx(txKey, tx)

	numTxsInMap := txsPoolsCleaner.cleanTxsPoolsIfNeeded()
	assert.Equal(t, 1, numTxsInMap)
}

func TestCleanTxsPoolsIfNeeded_RoundDiffTooBigShouldBeRemoved(t *testing.T) {
	t.Parallel()

	sndAddr := []byte("sndAddr")
	args := createMockArgTxsPoolsCleaner()
	roundHandler := &mock.RoundStub{IndexCalled: func() int64 {
		return 0
	}}
	args.RoundHandler = roundHandler
	called := false
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
					return &testscommon.CacherStub{
						GetCalled: func(key []byte) (value interface{}, ok bool) {
							return nil, true
						},
						RemoveCalled: func(key []byte) {
							called = true
						},
					}
				},
			}
		},
	}
	args.ShardCoordinator = &mock.CoordinatorStub{
		ComputeIdCalled: func(address []byte) uint32 {
			return 2
		},
	}
	txsPoolsCleaner, _ := NewTxsPoolsCleaner(args)

	txKey := []byte("key")
	tx := &transaction.Transaction{
		SndAddr: sndAddr,
	}
	txsPoolsCleaner.receivedUnsignedTx(txKey, tx)

	roundHandler.IndexCalled = func() int64 {
		return args.MaxRoundsToKeepUnprocessedData + 1
	}
	numTxsInMap := txsPoolsCleaner.cleanTxsPoolsIfNeeded()
	assert.Equal(t, 0, numTxsInMap)
	assert.Nil(t, txsPoolsCleaner.mapTxsRounds[string(txKey)])
	assert.True(t, called)
}
