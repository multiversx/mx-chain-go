package staking

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/stretchr/testify/assert"
)

func TestNewTestMetaProcessor(t *testing.T) {
	node := NewTestMetaProcessor(1, 1, 1, 1, 1)
	metaHdr := &block.MetaBlock{}
	headerHandler, bodyHandler, err := node.MetaBlockProcessor.CreateBlock(metaHdr, func() bool { return true })
	assert.Nil(t, err)

	err = headerHandler.SetRound(uint64(1))
	assert.Nil(t, err)

	err = headerHandler.SetNonce(1)
	assert.Nil(t, err)

	err = headerHandler.SetPrevHash([]byte("hash"))
	assert.Nil(t, err)

	err = headerHandler.SetAccumulatedFees(big.NewInt(0))
	assert.Nil(t, err)

	_ = bodyHandler
	/*
		metaHeaderHandler, _ := headerHandler.(data.MetaHeaderHandler)
		err = metaHeaderHandler.SetAccumulatedFeesInEpoch(big.NewInt(0))
		assert.Nil(t, err)

		err = node.MetaBlockProcessor.ProcessBlock(headerHandler, bodyHandler, func() time.Duration { return time.Second })
		assert.Nil(t, err)
	*/
}
