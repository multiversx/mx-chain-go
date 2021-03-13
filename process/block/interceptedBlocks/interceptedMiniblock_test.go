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

func createDefaultMiniblockArgument() *interceptedBlocks.ArgInterceptedMiniblock {
	arg := &interceptedBlocks.ArgInterceptedMiniblock{
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		Hasher:           testHasher,
		Marshalizer:      testMarshalizer,
	}

	miniblock := createMockMiniblock()
	arg.MiniblockBuff, _ = testMarshalizer.Marshal(miniblock)

	return arg
}

func createMockMiniblock() *block.MiniBlock {
	return &block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes: [][]byte{
			[]byte("hash1"),
			[]byte("hash2"),
		},
	}
}

//------- NewInterceptedMiniblock

func TestNewInterceptedMiniblock_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	inMb, err := interceptedBlocks.NewInterceptedMiniblock(nil)

	assert.True(t, check.IfNil(inMb))
	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestNewInterceptedMiniblock_MarshalizerFailShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultMiniblockArgument()
	arg.MiniblockBuff = []byte("invalid buffer")

	inMb, err := interceptedBlocks.NewInterceptedMiniblock(arg)

	assert.True(t, check.IfNil(inMb))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid character")
}

func TestNewInterceptedMiniblock_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultMiniblockArgument()
	inMb, err := interceptedBlocks.NewInterceptedMiniblock(arg)

	assert.False(t, check.IfNil(inMb))
	assert.Nil(t, err)
}

//------- CheckValidity

func TestInterceptedMiniblock_InvalidReceiverShardIdShouldErr(t *testing.T) {
	t.Parallel()

	wrongShardId := uint32(4)
	mb := createMockMiniblock()
	mb.ReceiverShardID = wrongShardId
	buff, _ := testMarshalizer.Marshal(mb)

	arg := createDefaultMiniblockArgument()
	arg.MiniblockBuff = buff
	inMb, _ := interceptedBlocks.NewInterceptedMiniblock(arg)

	err := inMb.CheckValidity()

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestInterceptedMiniblock_InvalidSenderShardIdShouldErr(t *testing.T) {
	t.Parallel()

	wrongShardId := uint32(4)
	mb := createMockMiniblock()
	mb.SenderShardID = wrongShardId
	buff, _ := testMarshalizer.Marshal(mb)

	arg := createDefaultMiniblockArgument()
	arg.MiniblockBuff = buff
	inMb, _ := interceptedBlocks.NewInterceptedMiniblock(arg)

	err := inMb.CheckValidity()

	assert.Equal(t, process.ErrInvalidShardId, err)
}

func TestInterceptedMiniblock_ContainsNilHashShouldErr(t *testing.T) {
	t.Parallel()

	mb := createMockMiniblock()
	mb.TxHashes[1] = nil
	buff, _ := testMarshalizer.Marshal(mb)

	arg := createDefaultMiniblockArgument()
	arg.MiniblockBuff = buff
	inMb, _ := interceptedBlocks.NewInterceptedMiniblock(arg)

	err := inMb.CheckValidity()

	assert.Equal(t, process.ErrNilTxHash, err)
}

//TODO: Remove this unit test after the final decision about reserved field from mini block
//func TestInterceptedMiniblock_ReservedPopulateShouldErr(t *testing.T) {
//	t.Parallel()
//
//	mb := createMockMiniblock()
//	mb.Reserved = []byte("r")
//	buff, _ := testMarshalizer.Marshal(mb)
//
//	arg := createDefaultMiniblockArgument()
//	arg.MiniblockBuff = buff
//	inMb, _ := interceptedBlocks.NewInterceptedMiniblock(arg)
//
//	err := inMb.CheckValidity()
//
//	assert.Equal(t, process.ErrReservedFieldNotSupportedYet, err)
//}

func TestInterceptedMiniblock_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultMiniblockArgument()
	inMb, _ := interceptedBlocks.NewInterceptedMiniblock(arg)

	err := inMb.CheckValidity()

	assert.Nil(t, err)
}

//------- getters

func TestInterceptedMiniblock_Getters(t *testing.T) {
	t.Parallel()

	arg := createDefaultMiniblockArgument()
	inMb, err := interceptedBlocks.NewInterceptedMiniblock(arg)
	assert.Nil(t, err)

	hash := testHasher.Compute(string(arg.MiniblockBuff))

	assert.Equal(t, hash, inMb.Hash())
	assert.Equal(t, createMockMiniblock(), inMb.Miniblock())
}

//------- IsForCurrentShard

func TestInterceptedMiniblock_IsForCurrentShardMiniblockForOtherShardsShouldRetFalse(t *testing.T) {
	t.Parallel()

	mb := &block.MiniBlock{
		ReceiverShardID: 1,
		SenderShardID:   2,
		TxHashes: [][]byte{
			[]byte("hash1"),
			[]byte("hash2"),
		},
	}
	buff, _ := testMarshalizer.Marshal(mb)

	arg := createDefaultMiniblockArgument()
	arg.MiniblockBuff = buff

	inMb, _ := interceptedBlocks.NewInterceptedMiniblock(arg)

	assert.False(t, inMb.IsForCurrentShard())
}

func TestInterceptedMiniblock_IsForCurrentShardMiniblockWithSenderInShardShouldRetTrue(t *testing.T) {
	t.Parallel()

	currentShard := uint32(0)
	mb := &block.MiniBlock{
		ReceiverShardID: 3,
		SenderShardID:   currentShard,
		TxHashes: [][]byte{
			[]byte("hash1"),
			[]byte("hash2"),
		},
	}
	buff, _ := testMarshalizer.Marshal(mb)

	arg := createDefaultMiniblockArgument()
	arg.MiniblockBuff = buff

	inMb, _ := interceptedBlocks.NewInterceptedMiniblock(arg)

	assert.True(t, inMb.IsForCurrentShard())
}

func TestInterceptedMiniblock_IsForCurrentShardMiniblockWithReceiverInShardShouldRetTrue(t *testing.T) {
	t.Parallel()

	currentShard := uint32(0)
	mb := &block.MiniBlock{
		ReceiverShardID: currentShard,
		SenderShardID:   3,
		TxHashes: [][]byte{
			[]byte("hash1"),
			[]byte("hash2"),
		},
	}
	buff, _ := testMarshalizer.Marshal(mb)

	arg := createDefaultMiniblockArgument()
	arg.MiniblockBuff = buff

	inMb, _ := interceptedBlocks.NewInterceptedMiniblock(arg)

	assert.True(t, inMb.IsForCurrentShard())
}
