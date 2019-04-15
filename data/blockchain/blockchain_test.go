package blockchain_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewBlockChain_NilBadBlockCacheShouldError(t *testing.T) {
	t.Parallel()

	_, err := blockchain.NewBlockChain(
		nil,
	)

	assert.Equal(t, err, blockchain.ErrBadBlocksCacheNil)
}

func TestBlockChain_IsBadBlock(t *testing.T) {
	t.Parallel()

	badBlocksStub := &mock.CacherStub{}
	hasReturns := true
	badBlocksStub.HasCalled = func(key []byte) bool {
		return hasReturns
	}

	b, _ := blockchain.NewBlockChain(
		badBlocksStub,
	)

	hasBadBlock := b.HasBadBlock([]byte("test"))
	assert.True(t, hasBadBlock)
}

func TestBlockChain_PutBadBlock(t *testing.T) {
	t.Parallel()

	badBlocksStub := &mock.CacherStub{}
	putCalled := false
	badBlocksStub.PutCalled = func(key []byte, value interface{}) bool {
		putCalled = true
		return true
	}

	b, _ := blockchain.NewBlockChain(
		badBlocksStub,
	)

	b.PutBadBlock([]byte("test"))
	assert.True(t, putCalled)
}
