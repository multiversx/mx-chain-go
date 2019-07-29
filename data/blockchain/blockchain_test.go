package blockchain_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestNewBlockChain_NilBadBlockCacheShouldError(t *testing.T) {
	t.Parallel()

	_, err := blockchain.NewBlockChain(
		nil,
		&statusHandler.NilStatusHandler{},
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
		&statusHandler.NilStatusHandler{},
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
		&statusHandler.NilStatusHandler{},
	)

	b.PutBadBlock([]byte("test"))
	assert.True(t, putCalled)
}
