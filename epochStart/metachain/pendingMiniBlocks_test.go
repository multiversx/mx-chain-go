package metachain

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/stretchr/testify/assert"
)

func TestNewPendingMiniBlocks_ShouldWork(t *testing.T) {
	t.Parallel()

	pmb, err := NewPendingMiniBlocks()
	assert.NotNil(t, pmb)
	assert.Nil(t, err)
}

func TestPendingMiniBlockHeaders_AddCommittedHeaderNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	pmb, _ := NewPendingMiniBlocks()

	err := pmb.AddProcessedHeader(nil)
	assert.Equal(t, epochStart.ErrNilHeaderHandler, err)
}

func TestPendingMiniBlockHeaders_AddProcessedHeaderWrongHeaderShouldErr(t *testing.T) {
	t.Parallel()

	pmb, _ := NewPendingMiniBlocks()
	header := &block.Header{}

	err := pmb.AddProcessedHeader(header)
	assert.Equal(t, epochStart.ErrWrongTypeAssertion, err)
}

func TestPendingMiniBlockHeaders_RevertHeaderNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	pmb, _ := NewPendingMiniBlocks()

	err := pmb.RevertHeader(nil)
	assert.Equal(t, epochStart.ErrNilHeaderHandler, err)
}

func TestPendingMiniBlockHeaders_RevertHeaderWrongHeaderTypeShouldErr(t *testing.T) {
	t.Parallel()

	pmb, _ := NewPendingMiniBlocks()
	header := &block.Header{}

	err := pmb.RevertHeader(header)
	assert.Equal(t, epochStart.ErrWrongTypeAssertion, err)
}
