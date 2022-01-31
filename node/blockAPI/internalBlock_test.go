package blockAPI

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockInternalBlockProcessor(
	shardID uint32,
	blockHeaderHash []byte,
	storerMock *mock.StorerMock,
	withKey bool,
) *internalBlockProcessor {
	return NewInternalBlockProcessor(
		&APIBlockProcessorArg{
			SelfShardID: shardID,
			Marshalizer: &mock.MarshalizerFake{},
			Store: &mock.ChainStorerMock{
				GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
					return storerMock
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
			},
		},
	)
}

// -------- ShardBlock --------

func TestInternalBlockProcessor_ConvertShardBlockBytesToInternalBlock_ShouldFail(t *testing.T) {
	t.Parallel()

	ibp := NewInternalBlockProcessor(
		&APIBlockProcessorArg{
			Marshalizer: &mock.MarshalizerFake{},
			HistoryRepo: &dblookupext.HistoryRepositoryStub{},
		},
	)

	blockHeader, err := ibp.convertShardBlockBytesToInternalBlock([]byte{0})
	assert.NotNil(t, err)
	assert.Nil(t, blockHeader)
}

func TestInternalBlockProcessor_ConvertShardBlockBytesToOutportFormat_ShouldFail(t *testing.T) {
	t.Parallel()

	ibp := createMockInternalBlockProcessor(
		core.MetachainShardId,
		nil,
		nil,
		false,
	)

	headerBytes := bytes.Repeat([]byte("1"), 10)

	headerOutput, err := ibp.convertShardBlockBytesByOutportFormat(2, headerBytes)
	assert.Equal(t, ErrInvalidOutportFormat, err)
	assert.Nil(t, headerOutput)
}

func TestInternalBlockProcessor_ConvertShardBlockBytesToInternalOutportFormat(t *testing.T) {
	t.Parallel()

	ibp := createMockInternalBlockProcessor(
		core.MetachainShardId,
		nil,
		nil,
		false,
	)

	header := &block.Header{
		Nonce: uint64(15),
		Round: uint64(14),
	}
	headerBytes, _ := json.Marshal(header)

	headerOutput, err := ibp.convertShardBlockBytesByOutportFormat(common.ApiOutputFormatInternal, headerBytes)
	require.Nil(t, err)
	assert.Equal(t, header, headerOutput)
}

func TestInternalBlockProcessor_ConvertShardBlockBytesToProtoOutportFormat(t *testing.T) {
	t.Parallel()

	ibp := createMockInternalBlockProcessor(
		core.MetachainShardId,
		nil,
		nil,
		false,
	)

	header := &block.Header{
		Nonce: uint64(15),
		Round: uint64(14),
	}
	headerBytes, _ := json.Marshal(header)

	headerOutput, err := ibp.convertShardBlockBytesByOutportFormat(common.ApiOutputFormatProto, headerBytes)
	require.Nil(t, err)
	assert.Equal(t, headerBytes, headerOutput)
}

func TestInternalBlockProcessor_GetInternalShardBlockByHashInvalidHashShouldErr(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := mock.NewStorerMock()

	ibp := createMockInternalBlockProcessor(
		0,
		headerHash,
		storerMock,
		false,
	)

	blk, err := ibp.GetInternalShardBlockByHash(common.ApiOutputFormatInternal, []byte("invalidHash"))
	assert.Nil(t, blk)
	assert.Error(t, err)
}

func TestInternalBlockProcessor_GetInternalShardBlockByNonceInvalidNonceShouldErr(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := mock.NewStorerMock()

	ibp := createMockInternalBlockProcessor(
		0,
		headerHash,
		storerMock,
		true,
	)

	blk, err := ibp.GetInternalShardBlockByNonce(common.ApiOutputFormatInternal, 100)
	assert.Nil(t, blk)
	assert.Error(t, err)
}

func TestInternalBlockProcessor_GetInternalShardBlockByRoundInvalidRoundShouldErr(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := mock.NewStorerMock()

	ibp := createMockInternalBlockProcessor(
		0,
		headerHash,
		storerMock,
		true,
	)

	blk, err := ibp.GetInternalShardBlockByRound(common.ApiOutputFormatInternal, 100)
	assert.Nil(t, blk)
	assert.Error(t, err)
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

	blk, err := ibp.GetInternalShardBlockByHash(common.ApiOutputFormatInternal, headerHash)
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

	blk, err := ibp.GetInternalShardBlockByNonce(common.ApiOutputFormatInternal, nonce)
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

	blk, err := ibp.GetInternalShardBlockByRound(common.ApiOutputFormatInternal, round)
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
	storerMock := mock.NewStorerMock()
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

	ibp := NewInternalBlockProcessor(
		&APIBlockProcessorArg{
			Marshalizer: &mock.MarshalizerFake{},
			HistoryRepo: &dblookupext.HistoryRepositoryStub{},
		},
	)

	blockHeader, err := ibp.convertMetaBlockBytesToInternalBlock([]byte{0})
	assert.NotNil(t, err)
	assert.Nil(t, blockHeader)
}

func TestInternalBlockProcessor_ConvertMetaBlockBytesToOutportFormat_ShouldFail(t *testing.T) {
	t.Parallel()

	ibp := createMockInternalBlockProcessor(
		core.MetachainShardId,
		nil,
		nil,
		false,
	)

	headerBytes := bytes.Repeat([]byte("1"), 10)

	headerOutput, err := ibp.convertMetaBlockBytesByOutportFormat(2, headerBytes)
	assert.Equal(t, ErrInvalidOutportFormat, err)
	assert.Nil(t, headerOutput)
}

func TestInternalBlockProcessor_ConvertMetaBlockBytesToInternalOutportFormat(t *testing.T) {
	t.Parallel()

	ibp := createMockInternalBlockProcessor(
		core.MetachainShardId,
		nil,
		nil,
		false,
	)

	header := &block.MetaBlock{
		Nonce: uint64(15),
		Round: uint64(14),
	}
	headerBytes, _ := json.Marshal(header)

	headerOutput, err := ibp.convertMetaBlockBytesByOutportFormat(common.ApiOutputFormatInternal, headerBytes)
	require.Nil(t, err)
	assert.Equal(t, header, headerOutput)
}

func TestInternalBlockProcessor_ConvertMetaBlockBytesToProtoOutportFormat(t *testing.T) {
	t.Parallel()

	ibp := createMockInternalBlockProcessor(
		core.MetachainShardId,
		nil,
		nil,
		false,
	)

	header := &block.MetaBlock{
		Nonce: uint64(15),
		Round: uint64(14),
	}
	headerBytes, _ := json.Marshal(header)

	headerOutput, err := ibp.convertMetaBlockBytesByOutportFormat(common.ApiOutputFormatProto, headerBytes)
	require.Nil(t, err)
	assert.Equal(t, headerBytes, headerOutput)
}

func TestInternalBlockProcessor_GetInternalMetaBlockByHashInvalidHashShouldErr(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := mock.NewStorerMock()

	ibp := createMockInternalBlockProcessor(
		core.MetachainShardId,
		headerHash,
		storerMock,
		false,
	)

	blk, err := ibp.GetInternalMetaBlockByHash(common.ApiOutputFormatInternal, []byte("invalidHash"))
	assert.Nil(t, blk)
	assert.Error(t, err)
}

func TestInternalBlockProcessor_GetInternalMetaBlockByNonceInvalidNonceShouldErr(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := mock.NewStorerMock()

	ibp := createMockInternalBlockProcessor(
		core.MetachainShardId,
		headerHash,
		storerMock,
		true,
	)

	blk, err := ibp.GetInternalMetaBlockByNonce(common.ApiOutputFormatInternal, 100)
	assert.Nil(t, blk)
	assert.Error(t, err)
}

func TestInternalBlockProcessor_GetInternalMetaBlockByRoundInvalidRoundShouldErr(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := mock.NewStorerMock()

	ibp := createMockInternalBlockProcessor(
		core.MetachainShardId,
		headerHash,
		storerMock,
		true,
	)

	blk, err := ibp.GetInternalMetaBlockByRound(common.ApiOutputFormatInternal, 100)
	assert.Nil(t, blk)
	assert.Error(t, err)
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

	blk, err := ibp.GetInternalMetaBlockByHash(common.ApiOutputFormatInternal, headerHash)
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

	blk, err := ibp.GetInternalMetaBlockByNonce(common.ApiOutputFormatInternal, nonce)
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

	blk, err := ibp.GetInternalMetaBlockByRound(common.ApiOutputFormatInternal, nonce)
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
	storerMock := mock.NewStorerMock()
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
