package blockchain_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewMetachain_ErrorsWhenBlockUnitIsNil(t *testing.T) {
	metachain, err := blockchain.NewMetachain(nil)
	assert.Nil(t, metachain)
	assert.Equal(t, blockchain.ErrMetaBlockUnitNil, err)
}

func TestNewMetachain_ReturnsCorrectly(t *testing.T) {
	metablockUnit := &mock.StorerStub{}
	metachain, err := blockchain.NewMetachain(metablockUnit)
	assert.NotNil(t, metachain)
	assert.Nil(t, err)
}
