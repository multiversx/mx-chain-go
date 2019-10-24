package interceptedBlocks_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func createDefaultTxBodyArgument() *interceptedBlocks.ArgInterceptedTxBlockBody {
	arg := &interceptedBlocks.ArgInterceptedTxBlockBody{
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		Hasher:           testHasher,
		Marshalizer:      testMarshalizer,
	}

	txBlockBody := createMockTxBlockBody()
	arg.TxBlockBodyBuff, _ = testMarshalizer.Marshal(txBlockBody)

	return arg
}

func createMockTxBlockBody() block.Body {
	return block.Body{
		{
			ReceiverShardID: 0,
			SenderShardID:   0,
			TxHashes: [][]byte{
				[]byte("hash1"),
				[]byte("hash2"),
			},
		},
		{
			ReceiverShardID: 0,
			SenderShardID:   0,
			TxHashes:        make([][]byte, 0),
		},
	}
}

//------- NewInterceptedTxBlockBody

func TestNewInterceptedTxBlockBody_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	inTxBody, err := interceptedBlocks.NewInterceptedTxBlockBody(nil)

	assert.Nil(t, inTxBody)
	assert.Equal(t, process.ErrNilArguments, err)
}

func TestNewInterceptedTxBlockBody_MarshalizerFailShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultTxBodyArgument()
	arg.TxBlockBodyBuff = []byte("invalid buffer")

	inTxBody, err := interceptedBlocks.NewInterceptedTxBlockBody(arg)

	assert.Nil(t, inTxBody)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid character")
}

func TestNewInterceptedTxBlockBody_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultTxBodyArgument()

	inTxBody, err := interceptedBlocks.NewInterceptedTxBlockBody(arg)

	assert.False(t, check.IfNil(inTxBody))
	assert.Nil(t, err)
}

//------- CheckValidity

func TestInterceptedTxBlockBody_CheckValidityNilTxHashesShouldErr(t *testing.T) {
	t.Parallel()

	txBody := createMockTxBlockBody()
	txBody[0].TxHashes = nil
	buff, _ := testMarshalizer.Marshal(txBody)

	arg := createDefaultTxBodyArgument()
	arg.TxBlockBodyBuff = buff
	inTxBody, _ := interceptedBlocks.NewInterceptedTxBlockBody(arg)

	err := inTxBody.CheckValidity()

	assert.Equal(t, process.ErrNilTxHashes, err)
}

func TestInterceptedTxBlockBody_InvalidReceiverShardIdShouldErr(t *testing.T) {
	t.Parallel()

	wrongShardId := uint32(4)
	txBody := createMockTxBlockBody()
	txBody[0].ReceiverShardID = wrongShardId
	buff, _ := testMarshalizer.Marshal(txBody)

	arg := createDefaultTxBodyArgument()
	arg.TxBlockBodyBuff = buff
	inTxBody, _ := interceptedBlocks.NewInterceptedTxBlockBody(arg)

	err := inTxBody.CheckValidity()

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestInterceptedTxBlockBody_InvalidSenderShardIdShouldErr(t *testing.T) {
	t.Parallel()

	wrongShardId := uint32(4)
	txBody := createMockTxBlockBody()
	txBody[0].SenderShardID = wrongShardId
	buff, _ := testMarshalizer.Marshal(txBody)

	arg := createDefaultTxBodyArgument()
	arg.TxBlockBodyBuff = buff
	inTxBody, _ := interceptedBlocks.NewInterceptedTxBlockBody(arg)

	err := inTxBody.CheckValidity()

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestInterceptedTxBlockBody_ContainsNilHashShouldErr(t *testing.T) {
	t.Parallel()

	txBody := createMockTxBlockBody()
	txBody[0].TxHashes[1] = nil
	buff, _ := testMarshalizer.Marshal(txBody)

	arg := createDefaultTxBodyArgument()
	arg.TxBlockBodyBuff = buff
	inTxBody, _ := interceptedBlocks.NewInterceptedTxBlockBody(arg)

	err := inTxBody.CheckValidity()

	assert.Equal(t, process.ErrNilTxHash, err)
}

func TestInterceptedTxBlockBody_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultTxBodyArgument()
	inTxBody, _ := interceptedBlocks.NewInterceptedTxBlockBody(arg)

	err := inTxBody.CheckValidity()

	assert.Nil(t, err)
}

//------- getters

func TestInterceptedTxBlockBody_Getters(t *testing.T) {
	t.Parallel()

	arg := createDefaultTxBodyArgument()
	inTxBody, _ := interceptedBlocks.NewInterceptedTxBlockBody(arg)

	hash := testHasher.Compute(string(arg.TxBlockBodyBuff))

	assert.Equal(t, hash, inTxBody.Hash())
	assert.Equal(t, 2, len(inTxBody.TxBlockBody()))
}

//------- IsForCurrentShard

func TestInterceptedTxBlockBody_IsForCurrentShardNoMiniblocksShouldRetFalse(t *testing.T) {
	t.Parallel()

	txBody := make([]*block.MiniBlock, 0)
	buff, _ := testMarshalizer.Marshal(txBody)

	arg := createDefaultTxBodyArgument()
	arg.TxBlockBodyBuff = buff

	inTxBody, _ := interceptedBlocks.NewInterceptedTxBlockBody(arg)

	assert.False(t, inTxBody.IsForCurrentShard())
}

func TestInterceptedTxBlockBody_IsForCurrentShardMiniblockForOtherShardsShouldRetFalse(t *testing.T) {
	t.Parallel()

	txBody := []*block.MiniBlock{
		{
			ReceiverShardID: 1,
			SenderShardID:   2,
			TxHashes: [][]byte{
				[]byte("hash1"),
				[]byte("hash2"),
			},
		},
		{
			ReceiverShardID: 3,
			SenderShardID:   4,
			TxHashes: [][]byte{
				[]byte("hash1"),
				[]byte("hash2"),
			},
		},
	}
	buff, _ := testMarshalizer.Marshal(txBody)

	arg := createDefaultTxBodyArgument()
	arg.TxBlockBodyBuff = buff

	inTxBody, _ := interceptedBlocks.NewInterceptedTxBlockBody(arg)

	assert.False(t, inTxBody.IsForCurrentShard())
}

func TestInterceptedTxBlockBody_IsForCurrentShardMiniblockOneMiniBlockWithSenderShouldRetTrue(t *testing.T) {
	t.Parallel()

	currentShard := uint32(0)
	txBody := []*block.MiniBlock{
		{
			ReceiverShardID: 1,
			SenderShardID:   2,
			TxHashes: [][]byte{
				[]byte("hash1"),
				[]byte("hash2"),
			},
		},
		{
			ReceiverShardID: 3,
			SenderShardID:   currentShard,
			TxHashes: [][]byte{
				[]byte("hash1"),
				[]byte("hash2"),
			},
		},
	}
	buff, _ := testMarshalizer.Marshal(txBody)

	arg := createDefaultTxBodyArgument()
	arg.TxBlockBodyBuff = buff

	inTxBody, _ := interceptedBlocks.NewInterceptedTxBlockBody(arg)

	assert.True(t, inTxBody.IsForCurrentShard())
}

func TestInterceptedTxBlockBody_IsForCurrentShardMiniblockOneMiniBlockWithReceiverShouldRetTrue(t *testing.T) {
	t.Parallel()

	currentShard := uint32(0)
	txBody := []*block.MiniBlock{
		{
			ReceiverShardID: 1,
			SenderShardID:   2,
			TxHashes: [][]byte{
				[]byte("hash1"),
				[]byte("hash2"),
			},
		},
		{
			ReceiverShardID: currentShard,
			SenderShardID:   3,
			TxHashes: [][]byte{
				[]byte("hash1"),
				[]byte("hash2"),
			},
		},
	}
	buff, _ := testMarshalizer.Marshal(txBody)

	arg := createDefaultTxBodyArgument()
	arg.TxBlockBodyBuff = buff

	inTxBody, _ := interceptedBlocks.NewInterceptedTxBlockBody(arg)

	assert.True(t, inTxBody.IsForCurrentShard())
}

//------- IsInterfaceNil

func TestInterceptedTxBlockBody_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var inTxBody *interceptedBlocks.InterceptedTxBlockBody

	assert.True(t, check.IfNil(inTxBody))
}
