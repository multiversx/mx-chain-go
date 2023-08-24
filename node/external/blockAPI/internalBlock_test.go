package blockAPI

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	storageMocks "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockInternalBlockProcessor(
	shardID uint32,
	blockHeaderHash []byte,
	storerMock storage.Storer,
	withKey bool,
) *internalBlockProcessor {
	return newInternalBlockProcessor(
		&ArgAPIBlockProcessor{
			SelfShardID: shardID,
			Marshalizer: &mock.MarshalizerFake{},
			Store: &storageMocks.ChainStorerStub{
				GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
					return storerMock, nil
				},
				GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
					if withKey {
						return storerMock.Get(key)
					}
					return blockHeaderHash, nil
				},
			},
			Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
			HistoryRepo: &dblookupext.HistoryRepositoryStub{
				GetEpochByHashCalled: func(hash []byte) (uint32, error) {
					return 1, nil
				},
				IsEnabledCalled: func() bool {
					return false
				},
			},
			EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		}, nil)
}

// -------- ShardBlock --------

func TestInternalBlockProcessor_ConvertShardBlockBytesToInternalBlockShouldFail(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("failed to unmarshal err")

	ibp := newInternalBlockProcessor(
		&ArgAPIBlockProcessor{
			Marshalizer: &marshallerMock.MarshalizerStub{
				UnmarshalCalled: func(_ interface{}, buff []byte) error {
					return expectedErr
				},
			},
			HistoryRepo:         &dblookupext.HistoryRepositoryStub{},
			EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		}, nil)

	wrongBytes := []byte{0, 1, 2}

	blockHeader, err := ibp.convertShardBlockBytesToInternalBlock(wrongBytes)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, blockHeader)
}

func TestInternalBlockProcessor_ConvertShardBlockBytesToInternalBlockShouldWork(t *testing.T) {
	t.Parallel()

	ibp := newInternalBlockProcessor(
		&ArgAPIBlockProcessor{
			Marshalizer:         &marshallerMock.MarshalizerMock{},
			HistoryRepo:         &dblookupext.HistoryRepositoryStub{},
			EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		}, nil)

	header := &block.Header{
		Nonce: uint64(15),
		Round: uint64(14),
	}
	headerBytes, _ := json.Marshal(header)

	blockHeader, err := ibp.convertShardBlockBytesToInternalBlock(headerBytes)
	assert.Nil(t, err)
	assert.Equal(t, header, blockHeader)
}

func TestInternalBlockProcessor_ConvertShardBlockBytesByOutputFormat(t *testing.T) {
	t.Parallel()

	ibp := createMockInternalBlockProcessor(
		1,
		nil,
		nil,
		false,
	)
	header := &block.Header{
		Nonce: uint64(15),
		Round: uint64(14),
	}
	headerBytes, _ := json.Marshal(header)

	t.Run("invalid output format, should fail", func(t *testing.T) {
		t.Parallel()

		headerOutput, err := ibp.convertShardBlockBytesByOutputFormat(2, headerBytes)
		assert.Equal(t, ErrInvalidOutputFormat, err)
		assert.Nil(t, headerOutput)
	})

	t.Run("internal format, should work", func(t *testing.T) {
		t.Parallel()

		headerOutput, err := ibp.convertShardBlockBytesByOutputFormat(common.ApiOutputFormatJSON, headerBytes)
		require.Nil(t, err)
		assert.Equal(t, header, headerOutput)
	})

	t.Run("proto format, should work", func(t *testing.T) {
		t.Parallel()

		headerOutput, err := ibp.convertShardBlockBytesByOutputFormat(common.ApiOutputFormatProto, headerBytes)
		require.Nil(t, err)
		assert.Equal(t, headerBytes, headerOutput)
	})

}

func TestInternalBlockProcessor_GetInternalShardBlockShouldFail(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	expectedErr := errors.New("key not found err")
	storerMock := &storageMocks.StorerStub{
		GetCalled: func(_ []byte) ([]byte, error) {
			return nil, expectedErr
		},
	}

	ibp := createMockInternalBlockProcessor(
		0,
		headerHash,
		storerMock,
		true,
	)

	t.Run("storer not found", func(t *testing.T) {
		t.Parallel()

		ibpTmp := createMockInternalBlockProcessor(
			0,
			headerHash,
			storerMock,
			true,
		)
		ibpTmp.store = &storageMocks.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return nil, expectedErr
			},
		}
		ibpTmp.hasDbLookupExtensions = true
		blk, err := ibpTmp.GetInternalShardBlockByHash(common.ApiOutputFormatJSON, headerHash)
		assert.Nil(t, blk)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("provided hash not in storer", func(t *testing.T) {
		t.Parallel()

		blk, err := ibp.GetInternalShardBlockByHash(common.ApiOutputFormatJSON, []byte("invalidHash"))
		assert.Nil(t, blk)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("provided nonce not in storer", func(t *testing.T) {
		t.Parallel()

		blk, err := ibp.GetInternalShardBlockByNonce(common.ApiOutputFormatJSON, 100)
		assert.Nil(t, blk)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("provided round not in storer", func(t *testing.T) {
		t.Parallel()

		blk, err := ibp.GetInternalShardBlockByRound(common.ApiOutputFormatJSON, 100)
		assert.Nil(t, blk)
		assert.Equal(t, expectedErr, err)
	})
}

func TestInternalBlockProcessor_GetInternalShardBlockByHash(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(1)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	ibp, headerBytes := preapreShardBlockProcessor(nonce, round, headerHash)
	header := &block.Header{}
	err := json.Unmarshal(headerBytes, header)
	require.Nil(t, err)

	blk, err := ibp.GetInternalShardBlockByHash(common.ApiOutputFormatJSON, headerHash)
	assert.Nil(t, err)
	assert.Equal(t, header, blk)
}

func TestInternalBlockProcessor_GetProtoShardBlockByHash(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(1)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	ibp, headerBytes := preapreShardBlockProcessor(nonce, round, headerHash)

	blk, err := ibp.GetInternalShardBlockByHash(common.ApiOutputFormatProto, headerHash)
	assert.Nil(t, err)
	assert.Equal(t, headerBytes, blk)
}

func TestInternalBlockProcessor_GetInternalShardBlockByNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(1)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	ibp, headerBytes := preapreShardBlockProcessor(nonce, round, headerHash)
	header := &block.Header{}
	err := json.Unmarshal(headerBytes, header)
	require.Nil(t, err)

	blk, err := ibp.GetInternalShardBlockByNonce(common.ApiOutputFormatJSON, nonce)
	assert.Nil(t, err)
	assert.Equal(t, header, blk)
}

func TestInternalBlockProcessor_GetProtoShardBlockByNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(1)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	ibp, headerBytes := preapreShardBlockProcessor(nonce, round, headerHash)

	blk, err := ibp.GetInternalShardBlockByNonce(common.ApiOutputFormatProto, nonce)
	assert.Nil(t, err)
	assert.Equal(t, headerBytes, blk)
}

func TestInternalBlockProcessor_GetInternalShardBlockByRound(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(1)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	ibp, headerBytes := preapreShardBlockProcessor(nonce, round, headerHash)
	header := &block.Header{}
	err := json.Unmarshal(headerBytes, header)
	require.Nil(t, err)

	blk, err := ibp.GetInternalShardBlockByRound(common.ApiOutputFormatJSON, round)
	assert.Nil(t, err)
	assert.Equal(t, header, blk)
}

func TestInternalBlockProcessor_GetProtoShardBlockByRound(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(1)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	ibp, headerBytes := preapreShardBlockProcessor(nonce, round, headerHash)

	blk, err := ibp.GetInternalShardBlockByRound(common.ApiOutputFormatProto, round)
	assert.Nil(t, err)
	assert.Equal(t, headerBytes, blk)
}

func preapreShardBlockProcessor(nonce uint64, round uint64, headerHash []byte) (*internalBlockProcessor, []byte) {
	storerMock := genericMocks.NewStorerMock()
	uint64Converter := mock.NewNonceHashConverterMock()

	ibp := createMockInternalBlockProcessor(
		0,
		headerHash,
		storerMock,
		true,
	)

	header := &block.Header{
		Nonce: nonce,
		Round: round,
	}
	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, headerBytes)

	nonceBytes := uint64Converter.ToByteSlice(nonce)
	_ = storerMock.Put(nonceBytes, headerHash)

	return ibp, headerBytes
}

// -------- MetaBlock --------

func TestInternalBlockProcessor_ConvertMetaBlockBytesToInternalBlock_ShouldFail(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("failed to unmarshal err")

	ibp := newInternalBlockProcessor(
		&ArgAPIBlockProcessor{
			Marshalizer: &marshallerMock.MarshalizerStub{
				UnmarshalCalled: func(_ interface{}, buff []byte) error {
					return expectedErr
				},
			},
			HistoryRepo:         &dblookupext.HistoryRepositoryStub{},
			EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		}, nil)

	wrongBytes := []byte{0, 1, 2}

	blockHeader, err := ibp.convertMetaBlockBytesToInternalBlock(wrongBytes)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, blockHeader)
}

func TestInternalBlockProcessor_ConvertMetaBlockBytesToInternalBlockShouldWork(t *testing.T) {
	t.Parallel()

	ibp := newInternalBlockProcessor(
		&ArgAPIBlockProcessor{
			Marshalizer:         &marshallerMock.MarshalizerMock{},
			HistoryRepo:         &dblookupext.HistoryRepositoryStub{},
			EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		}, nil)

	header := &block.MetaBlock{
		Nonce: uint64(15),
		Round: uint64(14),
	}
	headerBytes, _ := json.Marshal(header)

	blockHeader, err := ibp.convertMetaBlockBytesToInternalBlock(headerBytes)
	assert.Nil(t, err)
	assert.Equal(t, header, blockHeader)
}

func TestInternalBlockProcessor_ConvertMetaBlockBytesByOutputFormat(t *testing.T) {
	t.Parallel()

	ibp := createMockInternalBlockProcessor(
		1,
		nil,
		nil,
		false,
	)
	header := &block.MetaBlock{
		Nonce: uint64(15),
		Round: uint64(14),
	}
	headerBytes, _ := json.Marshal(header)

	t.Run("invalid output format, should fail", func(t *testing.T) {
		t.Parallel()

		headerOutput, err := ibp.convertMetaBlockBytesByOutputFormat(2, headerBytes)
		assert.Equal(t, ErrInvalidOutputFormat, err)
		assert.Nil(t, headerOutput)
	})

	t.Run("internal format, should work", func(t *testing.T) {
		t.Parallel()

		headerOutput, err := ibp.convertMetaBlockBytesByOutputFormat(common.ApiOutputFormatJSON, headerBytes)
		require.Nil(t, err)
		assert.Equal(t, header, headerOutput)
	})

	t.Run("proto format, should work", func(t *testing.T) {
		t.Parallel()

		headerOutput, err := ibp.convertMetaBlockBytesByOutputFormat(common.ApiOutputFormatProto, headerBytes)
		require.Nil(t, err)
		assert.Equal(t, headerBytes, headerOutput)
	})

}

func TestInternalBlockProcessor_GetInternalMetaBlockShouldFail(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	expectedErr := errors.New("key not found err")
	storerMock := &storageMocks.StorerStub{
		GetCalled: func(_ []byte) ([]byte, error) {
			return nil, expectedErr
		},
	}

	ibp := createMockInternalBlockProcessor(
		core.MetachainShardId,
		headerHash,
		storerMock,
		true,
	)

	t.Run("storer not found", func(t *testing.T) {
		t.Parallel()

		ibpTmp := createMockInternalBlockProcessor(
			core.MetachainShardId,
			headerHash,
			storerMock,
			true,
		)
		ibpTmp.store = &storageMocks.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return nil, expectedErr
			},
		}
		ibpTmp.hasDbLookupExtensions = true
		blk, err := ibpTmp.GetInternalMetaBlockByHash(common.ApiOutputFormatJSON, []byte("invalidHash"))
		assert.Nil(t, blk)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("provided hash not in storer", func(t *testing.T) {
		t.Parallel()

		blk, err := ibp.GetInternalMetaBlockByHash(common.ApiOutputFormatJSON, []byte("invalidHash"))
		assert.Nil(t, blk)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("provided nonce not in storer", func(t *testing.T) {
		t.Parallel()

		blk, err := ibp.GetInternalMetaBlockByNonce(common.ApiOutputFormatJSON, 100)
		assert.Nil(t, blk)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("provided round not in storer", func(t *testing.T) {
		t.Parallel()

		blk, err := ibp.GetInternalMetaBlockByRound(common.ApiOutputFormatJSON, 100)
		assert.Nil(t, blk)
		assert.Equal(t, expectedErr, err)
	})
}

func TestInternalBlockProcessor_GetInternalMetaBlockByHash(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(1)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	ibp, headerBytes := prepareMetaBlockProcessor(nonce, round, headerHash)
	header := &block.MetaBlock{}
	err := json.Unmarshal(headerBytes, header)
	require.Nil(t, err)

	blk, err := ibp.GetInternalMetaBlockByHash(common.ApiOutputFormatJSON, headerHash)
	assert.Nil(t, err)
	assert.Equal(t, header, blk)
}

func TestInternalBlockProcessor_GetProtoMetaBlockByHash(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(1)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	ibp, headerBytes := prepareMetaBlockProcessor(nonce, round, headerHash)

	blk, err := ibp.GetInternalMetaBlockByHash(common.ApiOutputFormatProto, headerHash)
	assert.Nil(t, err)
	assert.Equal(t, headerBytes, blk)
}

func TestInternalBlockProcessor_GetInternalMetaBlockByNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(1)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	ibp, headerBytes := prepareMetaBlockProcessor(nonce, round, headerHash)
	header := &block.MetaBlock{}
	err := json.Unmarshal(headerBytes, header)
	require.Nil(t, err)

	blk, err := ibp.GetInternalMetaBlockByNonce(common.ApiOutputFormatJSON, nonce)
	assert.Nil(t, err)
	assert.Equal(t, header, blk)
}

func TestInternalBlockProcessor_GetProtoMetaBlockByNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(1)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	ibp, headerBytes := prepareMetaBlockProcessor(nonce, round, headerHash)

	blk, err := ibp.GetInternalMetaBlockByNonce(common.ApiOutputFormatProto, nonce)
	assert.Nil(t, err)
	assert.Equal(t, headerBytes, blk)
}

func TestInternalBlockProcessor_GetInternalMetaBlockByRound(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(1)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	ibp, headerBytes := prepareMetaBlockProcessor(nonce, round, headerHash)
	header := &block.MetaBlock{}
	err := json.Unmarshal(headerBytes, header)
	require.Nil(t, err)

	blk, err := ibp.GetInternalMetaBlockByRound(common.ApiOutputFormatJSON, nonce)
	assert.Nil(t, err)
	assert.Equal(t, header, blk)
}

func TestInternalBlockProcessor_GetProtoMetaBlockByRound(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(1)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	ibp, headerBytes := prepareMetaBlockProcessor(nonce, round, headerHash)

	blk, err := ibp.GetInternalMetaBlockByRound(common.ApiOutputFormatProto, nonce)
	assert.Nil(t, err)
	assert.Equal(t, headerBytes, blk)
}

func prepareMetaBlockProcessor(nonce uint64, round uint64, headerHash []byte) (*internalBlockProcessor, []byte) {
	storerMock := genericMocks.NewStorerMock()
	uint64Converter := mock.NewNonceHashConverterMock()

	ibp := createMockInternalBlockProcessor(
		core.MetachainShardId,
		headerHash,
		storerMock,
		true,
	)

	header := &block.MetaBlock{
		Nonce: nonce,
		Round: round,
	}
	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, headerBytes)

	nonceBytes := uint64Converter.ToByteSlice(nonce)
	_ = storerMock.Put(nonceBytes, headerHash)

	return ibp, headerBytes
}

// ------- MiniBlock -------

func TestInternalBlockProcessor_GetInternalMiniBlockByHash(t *testing.T) {
	t.Parallel()

	miniBlockHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")
	txHash := []byte("dummyhash")
	expEpoch := uint32(1)

	t.Run("storer not found", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("key not found err")
		ibp := newInternalBlockProcessor(
			&ArgAPIBlockProcessor{
				SelfShardID: 1,
				Marshalizer: &mock.MarshalizerFake{},
				Store: &storageMocks.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return nil, expectedErr
					},
				},
				Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
				HistoryRepo:              &dblookupext.HistoryRepositoryStub{},
				EnableEpochsHandler:      &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			}, nil)

		blk, err := ibp.GetInternalMiniBlock(common.ApiOutputFormatJSON, []byte("invalidHash"), 1)
		assert.Nil(t, blk)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("provided hash not in storer", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("key not found err")
		storerMock := &storageMocks.StorerStub{
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				return nil, expectedErr
			},
		}

		ibp := newInternalBlockProcessor(
			&ArgAPIBlockProcessor{
				SelfShardID: 1,
				Marshalizer: &mock.MarshalizerFake{},
				Store: &storageMocks.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return storerMock, nil
					},
				},
				Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
				HistoryRepo:              &dblookupext.HistoryRepositoryStub{},
				EnableEpochsHandler:      &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			}, nil)

		blk, err := ibp.GetInternalMiniBlock(common.ApiOutputFormatJSON, []byte("invalidHash"), 1)
		assert.Nil(t, blk)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("provided hash not in storer", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("key not found err")
		storerMock := &storageMocks.StorerStub{
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				return nil, expectedErr
			},
		}

		ibp := newInternalBlockProcessor(
			&ArgAPIBlockProcessor{
				SelfShardID: 1,
				Marshalizer: &mock.MarshalizerFake{},
				Store: &storageMocks.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return storerMock, nil
					},
				},
				Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
				HistoryRepo:              &dblookupext.HistoryRepositoryStub{},
				EnableEpochsHandler:      &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			}, nil)

		blk, err := ibp.GetInternalMiniBlock(common.ApiOutputFormatJSON, []byte("invalidHash"), 1)
		assert.Nil(t, blk)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("proto raw data mini block, should work", func(t *testing.T) {
		t.Parallel()

		mb := &block.MiniBlock{
			TxHashes: [][]byte{txHash},
		}
		mbBytes, _ := json.Marshal(mb)
		storerMock := &storageMocks.StorerStub{
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				assert.Equal(t, miniBlockHash, key)
				assert.Equal(t, expEpoch, epoch)
				return mbBytes, nil
			},
		}

		ibp := newInternalBlockProcessor(
			&ArgAPIBlockProcessor{
				SelfShardID: 1,
				Marshalizer: &mock.MarshalizerFake{},
				Store: &storageMocks.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return storerMock, nil
					},
				},
				Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
				HistoryRepo:              &dblookupext.HistoryRepositoryStub{},
				EnableEpochsHandler:      &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			}, nil)

		blk, err := ibp.GetInternalMiniBlock(common.ApiOutputFormatProto, miniBlockHash, 1)
		assert.Nil(t, err)
		assert.Equal(t, mbBytes, blk)
	})

	t.Run("internal data mini block, should work", func(t *testing.T) {
		t.Parallel()

		mb := &block.MiniBlock{
			TxHashes: [][]byte{txHash},
		}
		mbBytes, _ := json.Marshal(mb)

		storerMock := &storageMocks.StorerStub{
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				assert.Equal(t, miniBlockHash, key)
				assert.Equal(t, expEpoch, epoch)
				return mbBytes, nil
			},
		}

		ibp := newInternalBlockProcessor(
			&ArgAPIBlockProcessor{
				SelfShardID: 1,
				Marshalizer: &mock.MarshalizerFake{},
				Store: &storageMocks.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return storerMock, nil
					},
				},
				Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
				HistoryRepo:              &dblookupext.HistoryRepositoryStub{},
				EnableEpochsHandler:      &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			}, nil)

		blk, err := ibp.GetInternalMiniBlock(common.ApiOutputFormatJSON, miniBlockHash, expEpoch)
		assert.Nil(t, err)
		assert.Equal(t, mb, blk)
	})
}

func TestInternalBlockProcessor_GetInternalStartOfEpochMetaBlock(t *testing.T) {
	t.Parallel()

	expEpoch := uint32(1)

	header := &block.MetaBlock{
		Nonce: 1,
	}
	headerBytes, _ := json.Marshal(header)

	t.Run("not metachain shard, should fail", func(t *testing.T) {
		t.Parallel()

		ibp := newInternalBlockProcessor(
			&ArgAPIBlockProcessor{
				SelfShardID:              1,
				Marshalizer:              &mock.MarshalizerFake{},
				Store:                    &storageMocks.ChainStorerStub{},
				Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
				HistoryRepo:              &dblookupext.HistoryRepositoryStub{},
				EnableEpochsHandler:      &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			}, nil)

		blk, err := ibp.GetInternalStartOfEpochMetaBlock(common.ApiOutputFormatJSON, expEpoch)
		assert.Nil(t, blk)
		assert.Equal(t, ErrMetachainOnlyEndpoint, err)
	})

	t.Run("fail to get from storer", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("key not found err")
		storerMock := &storageMocks.StorerStub{
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				return nil, expectedErr
			},
		}

		ibp := newInternalBlockProcessor(
			&ArgAPIBlockProcessor{
				SelfShardID: core.MetachainShardId,
				Marshalizer: &mock.MarshalizerFake{},
				Store: &storageMocks.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return storerMock, nil
					},
				},
				Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
				HistoryRepo:              &dblookupext.HistoryRepositoryStub{},
				EnableEpochsHandler:      &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			}, nil)

		blk, err := ibp.GetInternalStartOfEpochMetaBlock(common.ApiOutputFormatJSON, expEpoch)
		assert.Nil(t, blk)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("proto raw data meta block, should work", func(t *testing.T) {
		t.Parallel()

		storerMock := &storageMocks.StorerStub{
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				assert.Equal(t, expEpoch, epoch)
				return headerBytes, nil
			},
		}

		ibp := newInternalBlockProcessor(
			&ArgAPIBlockProcessor{
				SelfShardID: core.MetachainShardId,
				Marshalizer: &mock.MarshalizerFake{},
				Store: &storageMocks.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return storerMock, nil
					},
				},
				Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
				HistoryRepo:              &dblookupext.HistoryRepositoryStub{},
				EnableEpochsHandler:      &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			}, nil)

		blk, err := ibp.GetInternalStartOfEpochMetaBlock(common.ApiOutputFormatProto, expEpoch)
		assert.Nil(t, err)
		assert.Equal(t, headerBytes, blk)
	})

	t.Run("internal data meta block, should work", func(t *testing.T) {
		t.Parallel()

		storerMock := &storageMocks.StorerStub{
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				assert.Equal(t, expEpoch, epoch)
				return headerBytes, nil
			},
		}

		ibp := newInternalBlockProcessor(
			&ArgAPIBlockProcessor{
				SelfShardID: core.MetachainShardId,
				Marshalizer: &mock.MarshalizerFake{},
				Store: &storageMocks.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return storerMock, nil
					},
				},
				Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
				HistoryRepo:              &dblookupext.HistoryRepositoryStub{},
				EnableEpochsHandler:      &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			}, nil)

		blk, err := ibp.GetInternalStartOfEpochMetaBlock(common.ApiOutputFormatJSON, expEpoch)
		assert.Nil(t, err)
		assert.Equal(t, header, blk)
	})
}

func TestInternalBlockProcessor_GetInternalStartOfEpochValidatorsInfo(t *testing.T) {
	t.Parallel()

	expEpoch := uint32(1)

	t.Run("not metachain shard, should fail", func(t *testing.T) {
		t.Parallel()

		ibp := newInternalBlockProcessor(
			&ArgAPIBlockProcessor{
				SelfShardID:              1,
				Marshalizer:              &mock.MarshalizerFake{},
				Store:                    &storageMocks.ChainStorerStub{},
				Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
				HistoryRepo:              &dblookupext.HistoryRepositoryStub{},
				EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
					IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
						return flag == common.RefactorPeersMiniBlocksFlag
					},
				},
			}, nil)

		blk, err := ibp.GetInternalStartOfEpochValidatorsInfo(expEpoch)
		assert.Nil(t, blk)
		assert.Equal(t, ErrMetachainOnlyEndpoint, err)
	})

	t.Run("fail to get from storer", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("key not found err")
		storerMock := &storageMocks.StorerStub{
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				return nil, expectedErr
			},
		}

		ibp := newInternalBlockProcessor(
			&ArgAPIBlockProcessor{
				SelfShardID: core.MetachainShardId,
				Marshalizer: &mock.MarshalizerFake{},
				Store: &storageMocks.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return storerMock, nil
					},
				},
				Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
				HistoryRepo:              &dblookupext.HistoryRepositoryStub{},
				EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
					IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
						return flag == common.RefactorPeersMiniBlocksFlag
					},
				},
			}, nil)

		blk, err := ibp.GetInternalStartOfEpochValidatorsInfo(expEpoch)
		assert.Nil(t, blk)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("without refactor peers miniblock activated, should work", func(t *testing.T) {
		t.Parallel()

		marshaller := &mock.MarshalizerFake{}

		mbHeader1Hash := []byte("miniBlockHeaderHash1")
		mbHeader1 := block.MiniBlockHeader{
			Hash:            mbHeader1Hash,
			SenderShardID:   0,
			ReceiverShardID: 0,
			TxCount:         2,
			Type:            block.PeerBlock,
		}

		svi := &state.ShardValidatorInfo{
			PublicKey:  []byte("pubkey1"),
			ShardId:    0,
			Index:      1,
			TempRating: 500,
		}
		sviBytes, _ := marshaller.Marshal(svi)

		txHashes := [][]byte{sviBytes}
		mb1 := block.MiniBlock{
			TxHashes:        txHashes,
			ReceiverShardID: 1,
			SenderShardID:   1,
			Type:            block.PeerBlock,
		}
		mb1Bytes, _ := marshaller.Marshal(mb1)

		header := &block.MetaBlock{
			Nonce: 1,
			Epoch: 1,
			MiniBlockHeaders: []block.MiniBlockHeader{
				mbHeader1,
			},
		}
		headerBytes, _ := marshaller.Marshal(header)

		firstRun := true
		storerMock := &storageMocks.StorerStub{
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				if firstRun {
					firstRun = false
					return headerBytes, nil
				}

				require.Equal(t, mbHeader1Hash, key)
				return mb1Bytes, nil
			},
		}

		ibp := newInternalBlockProcessor(
			&ArgAPIBlockProcessor{
				SelfShardID: core.MetachainShardId,
				Marshalizer: marshaller,
				Store: &storageMocks.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return storerMock, nil
					},
				},
				Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
				HistoryRepo:              &dblookupext.HistoryRepositoryStub{},
				EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
					RefactorPeersMiniBlocksEnableEpochField: 5,
				},
			}, nil)

		validatorsInfo, err := ibp.GetInternalStartOfEpochValidatorsInfo(expEpoch)

		expectedValidatorsInfo := []*state.ShardValidatorInfo{
			svi,
		}

		assert.Nil(t, err)
		assert.Equal(t, expectedValidatorsInfo, validatorsInfo)
	})

	t.Run("with refactor peers miniblock activated, should work", func(t *testing.T) {
		t.Parallel()

		marshaller := &mock.MarshalizerFake{}

		mbHeader1Hash := []byte("miniBlockHeaderHash1")
		mbHeader1 := block.MiniBlockHeader{
			Hash:            mbHeader1Hash,
			SenderShardID:   0,
			ReceiverShardID: 0,
			TxCount:         2,
			Type:            block.PeerBlock,
		}

		txHashes := [][]byte{[]byte("txHash1")}
		mb1 := block.MiniBlock{
			TxHashes:        txHashes,
			ReceiverShardID: 1,
			SenderShardID:   1,
			Type:            block.PeerBlock,
		}
		mb1Bytes, _ := marshaller.Marshal(mb1)

		header := &block.MetaBlock{
			Nonce: 1,
			Epoch: 5,
			MiniBlockHeaders: []block.MiniBlockHeader{
				mbHeader1,
			},
		}
		headerBytes, _ := marshaller.Marshal(header)

		firstRun := true
		storerMock := &storageMocks.StorerStub{
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				if firstRun {
					firstRun = false
					return headerBytes, nil
				}

				require.Equal(t, mbHeader1Hash, key)
				return mb1Bytes, nil
			},
		}

		svi := &state.ShardValidatorInfo{
			PublicKey:  []byte("pubkey1"),
			ShardId:    0,
			Index:      1,
			TempRating: 500,
		}
		sviBytes, _ := marshaller.Marshal(svi)

		ibp := newInternalBlockProcessor(
			&ArgAPIBlockProcessor{
				SelfShardID: core.MetachainShardId,
				Marshalizer: marshaller,
				Store: &storageMocks.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return storerMock, nil
					},
					GetAllCalled: func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
						require.Equal(t, txHashes, keys)
						allData := make(map[string][]byte)
						allData["hash1"] = sviBytes
						return allData, nil
					},
				},
				Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
				HistoryRepo:              &dblookupext.HistoryRepositoryStub{},
				EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
					IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
						if flag == common.RefactorPeersMiniBlocksFlag {
							return epoch >= 5
						}
						return false
					},
				},
			}, nil)

		validatorsInfo, err := ibp.GetInternalStartOfEpochValidatorsInfo(expEpoch)

		expectedValidatorsInfo := []*state.ShardValidatorInfo{
			svi,
		}

		assert.Nil(t, err)
		assert.Equal(t, expectedValidatorsInfo, validatorsInfo)
	})
}
