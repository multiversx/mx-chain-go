package poolsCleaner

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
	"github.com/stretchr/testify/assert"
)

func TestNewCrossTxsPoolsCleaner_NilAddrConverterErr(t *testing.T) {
	t.Parallel()

	crossTxCleaner, err := NewCrossTxsPoolsCleaner(
		nil, &mock.PoolsHolderMock{}, &mock.RounderMock{}, mock.NewMultipleShardsCoordinatorMock(),
	)
	assert.Nil(t, crossTxCleaner)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewCrossTxsPoolsCleaner_NilDataPoolHolderErr(t *testing.T) {
	t.Parallel()

	crossTxCleaner, err := NewCrossTxsPoolsCleaner(
		&mock.PubkeyConverterStub{}, nil, &mock.RounderMock{}, mock.NewMultipleShardsCoordinatorMock(),
	)
	assert.Nil(t, crossTxCleaner)
	assert.Equal(t, process.ErrNilPoolsHolder, err)
}

func TestNewCrossTxsPoolsCleaner_NilTxsPoolErr(t *testing.T) {
	t.Parallel()

	dataPool := &mock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return nil
		},
	}
	crossTxCleaner, err := NewCrossTxsPoolsCleaner(
		&mock.PubkeyConverterStub{}, dataPool, &mock.RounderMock{}, mock.NewMultipleShardsCoordinatorMock(),
	)
	assert.Nil(t, crossTxCleaner)
	assert.Equal(t, process.ErrNilTransactionPool, err)
}

func TestNewCrossTxsPoolsCleaner_NilRewardTxsPoolErr(t *testing.T) {
	t.Parallel()

	dataPool := &mock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return nil
		},
	}
	crossTxCleaner, err := NewCrossTxsPoolsCleaner(
		&mock.PubkeyConverterStub{}, dataPool, &mock.RounderMock{}, mock.NewMultipleShardsCoordinatorMock(),
	)
	assert.Nil(t, crossTxCleaner)
	assert.Equal(t, process.ErrNilRewardTxDataPool, err)
}

func TestNewCrossTxsPoolsCleaner_NilUnsignedTxsPoolErr(t *testing.T) {
	t.Parallel()

	dataPool := &mock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return nil
		},
	}
	crossTxCleaner, err := NewCrossTxsPoolsCleaner(
		&mock.PubkeyConverterStub{}, dataPool, &mock.RounderMock{}, mock.NewMultipleShardsCoordinatorMock(),
	)
	assert.Nil(t, crossTxCleaner)
	assert.Equal(t, process.ErrNilUnsignedTxDataPool, err)
}

func TestNewCrossTxsPoolsCleaner_NilRounderErr(t *testing.T) {
	t.Parallel()

	dataPool := &mock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
	}
	crossTxCleaner, err := NewCrossTxsPoolsCleaner(
		&mock.PubkeyConverterStub{}, dataPool, nil, mock.NewMultipleShardsCoordinatorMock(),
	)
	assert.Nil(t, crossTxCleaner)
	assert.Equal(t, process.ErrNilRounder, err)
}

func TestNewCrossTxsPoolsCleaner_NilShardCoordinatorErr(t *testing.T) {
	t.Parallel()

	dataPool := &mock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
	}
	crossTxCleaner, err := NewCrossTxsPoolsCleaner(
		&mock.PubkeyConverterStub{}, dataPool, &mock.RounderMock{}, nil,
	)
	assert.Nil(t, crossTxCleaner)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewCrossTxsPoolsCleaner_ShouldWork(t *testing.T) {
	t.Parallel()

	dataPool := &mock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
	}

	crossTxCleaner, err := NewCrossTxsPoolsCleaner(
		&mock.PubkeyConverterStub{}, dataPool, &mock.RounderMock{}, mock.NewMultipleShardsCoordinatorMock(),
	)
	assert.Nil(t, err)
	assert.NotNil(t, crossTxCleaner)
}

func TestGetShardFromAddress(t *testing.T) {
	t.Parallel()

	addrLen := 64
	expectedErr := errors.New("error")
	testAddr := []byte("addr")
	addrConverter := &mock.PubkeyConverterStub{
		LenCalled: func() int {
			return addrLen
		},
		CreateAddressFromBytesCalled: func(pubKey []byte) (state.AddressContainer, error) {
			if bytes.Equal(pubKey, testAddr) {
				return nil, expectedErr
			}
			return nil, nil
		},
	}
	expectedShard := uint32(2)
	crossTxCleaner, _ := NewCrossTxsPoolsCleaner(
		addrConverter, &mock.PoolsHolderStub{}, &mock.RounderMock{}, &mock.CoordinatorStub{
			ComputeIdCalled: func(address state.AddressContainer) uint32 {
				return expectedShard
			},
		},
	)

	emptyAddr := make([]byte, addrLen)
	result, err := crossTxCleaner.getShardFromAddress(emptyAddr)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), result)

	_, err = crossTxCleaner.getShardFromAddress(testAddr)
	assert.Equal(t, expectedErr, err)

	result, err = crossTxCleaner.getShardFromAddress([]byte("123"))
	assert.Nil(t, err)
	assert.Equal(t, expectedShard, result)
}

func TestReceivedBlockTx_ShouldBeAddedInMapCrossTxsRounds(t *testing.T) {
	t.Parallel()

	crossTxCleaner, _ := NewCrossTxsPoolsCleaner(
		&mock.PubkeyConverterStub{}, &mock.PoolsHolderStub{
			TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &mock.ShardedDataStub{
					ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
						return &mock.CacherMock{}
					},
				}
			},
		}, &mock.RounderMock{}, &mock.CoordinatorStub{},
	)

	txWrap := &txcache.WrappedTransaction{
		Tx:            &transaction.Transaction{},
		SenderShardID: 2,
	}
	txBlockKey := []byte("key")
	crossTxCleaner.receivedBlockTx(txBlockKey, txWrap)
	assert.NotNil(t, crossTxCleaner.mapCrossTxsRounds[string(txBlockKey)])
}

func TestReceivedRewardTx_ShouldBeAddedInMapCrossTxsRounds(t *testing.T) {
	t.Parallel()

	crossTxCleaner, _ := NewCrossTxsPoolsCleaner(
		&mock.PubkeyConverterStub{}, &mock.PoolsHolderStub{
			RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &mock.ShardedDataStub{
					ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
						return &mock.CacherMock{}
					},
				}
			},
		}, &mock.RounderMock{}, &mock.CoordinatorStub{},
	)

	txKey := []byte("key")
	crossTxCleaner.receivedRewardTx(txKey, nil)
	assert.NotNil(t, crossTxCleaner.mapCrossTxsRounds[string(txKey)])
}

func TestReceivedUnsignedTx_ShouldBeAddedInMapCrossTxsRounds(t *testing.T) {
	t.Parallel()

	sndAddr := []byte("sndAddr")
	crossTxCleaner, _ := NewCrossTxsPoolsCleaner(
		&mock.PubkeyConverterStub{
			CreateAddressFromBytesCalled: func(pubKey []byte) (state.AddressContainer, error) {
				return nil, nil
			},
		}, &mock.PoolsHolderStub{
			UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &mock.ShardedDataStub{
					ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
						return &mock.CacherMock{}
					},
				}
			},
		}, &mock.RounderMock{}, &mock.CoordinatorStub{
			ComputeIdCalled: func(address state.AddressContainer) uint32 {
				return 2
			},
		},
	)

	txKey := []byte("key")
	tx := &transaction.Transaction{
		SndAddr: sndAddr,
	}
	crossTxCleaner.receivedUnsignedTx(txKey, tx)
	assert.NotNil(t, crossTxCleaner.mapCrossTxsRounds[string(txKey)])
}

func TestCleanCrossTxsPoolsIfNeeded_CannotFindTxInPoolShouldBeRemovedFromMap(t *testing.T) {
	t.Parallel()

	sndAddr := []byte("sndAddr")
	crossTxCleaner, _ := NewCrossTxsPoolsCleaner(
		&mock.PubkeyConverterStub{
			CreateAddressFromBytesCalled: func(pubKey []byte) (state.AddressContainer, error) {
				return nil, nil
			},
		}, &mock.PoolsHolderStub{
			UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &mock.ShardedDataStub{
					ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
						return &mock.CacherMock{}
					},
				}
			},
		}, &mock.RounderMock{}, &mock.CoordinatorStub{
			ComputeIdCalled: func(address state.AddressContainer) uint32 {
				return 2
			},
		},
	)

	txKey := []byte("key")
	tx := &transaction.Transaction{
		SndAddr: sndAddr,
	}
	crossTxCleaner.receivedUnsignedTx(txKey, tx)

	numTxsInMap := crossTxCleaner.cleanCrossTxsPoolsIfNeeded()
	assert.Equal(t, 0, numTxsInMap)
}

func TestCleanCrossTxsPoolsIfNeeded_RoundDiffTooSmallShouldNotBeRemoved(t *testing.T) {
	t.Parallel()

	sndAddr := []byte("sndAddr")
	crossTxCleaner, _ := NewCrossTxsPoolsCleaner(
		&mock.PubkeyConverterStub{
			CreateAddressFromBytesCalled: func(pubKey []byte) (state.AddressContainer, error) {
				return nil, nil
			},
		}, &mock.PoolsHolderStub{
			UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &mock.ShardedDataStub{
					ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
						return &mock.CacherStub{
							GetCalled: func(key []byte) (value interface{}, ok bool) {
								return nil, true
							},
						}
					},
				}
			},
		}, &mock.RounderMock{}, &mock.CoordinatorStub{
			ComputeIdCalled: func(address state.AddressContainer) uint32 {
				return 2
			},
		},
	)

	txKey := []byte("key")
	tx := &transaction.Transaction{
		SndAddr: sndAddr,
	}
	crossTxCleaner.receivedUnsignedTx(txKey, tx)

	numTxsInMap := crossTxCleaner.cleanCrossTxsPoolsIfNeeded()
	assert.Equal(t, 1, numTxsInMap)
}

func TestCleanCrossTxsPoolsIfNeeded_RoundDiffTooBigShouldBeRemoved(t *testing.T) {
	t.Parallel()

	rounder := &mock.RoundStub{IndexCalled: func() int64 {
		return 0
	}}
	called := false
	sndAddr := []byte("sndAddr")
	crossTxCleaner, _ := NewCrossTxsPoolsCleaner(
		&mock.PubkeyConverterStub{
			CreateAddressFromBytesCalled: func(pubKey []byte) (state.AddressContainer, error) {
				return nil, nil
			},
		}, &mock.PoolsHolderStub{
			UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &mock.ShardedDataStub{
					ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
						return &mock.CacherStub{
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
		}, rounder, &mock.CoordinatorStub{
			ComputeIdCalled: func(address state.AddressContainer) uint32 {
				return 2
			},
		},
	)

	txKey := []byte("key")
	tx := &transaction.Transaction{
		SndAddr: sndAddr,
	}
	crossTxCleaner.receivedUnsignedTx(txKey, tx)

	rounder.IndexCalled = func() int64 {
		return process.MaxRoundsToKeepUnprocessedTransactions + 1
	}
	numTxsInMap := crossTxCleaner.cleanCrossTxsPoolsIfNeeded()
	assert.Equal(t, 0, numTxsInMap)
	assert.Nil(t, crossTxCleaner.mapCrossTxsRounds[string(txKey)])
	assert.True(t, called)
}
