package pendingMb_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/pendingMb"
	"github.com/stretchr/testify/assert"
)

func TestNewPendingMiniBlocks_ShouldWork(t *testing.T) {
	t.Parallel()

	pmb, err := pendingMb.NewPendingMiniBlocks()

	assert.NotNil(t, pmb)
	assert.Nil(t, err)
}

func TestPendingMiniBlockHeaders_AddCommittedHeaderNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	pmb, _ := pendingMb.NewPendingMiniBlocks()
	err := pmb.AddProcessedHeader(nil)

	assert.Equal(t, process.ErrNilHeaderHandler, err)
}

func TestPendingMiniBlockHeaders_AddProcessedHeaderWrongHeaderShouldErr(t *testing.T) {
	t.Parallel()

	pmb, _ := pendingMb.NewPendingMiniBlocks()
	header := &block.Header{}
	err := pmb.AddProcessedHeader(header)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestPendingMiniBlockHeaders_RevertHeaderNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	pmb, _ := pendingMb.NewPendingMiniBlocks()
	err := pmb.RevertHeader(nil)

	assert.Equal(t, process.ErrNilHeaderHandler, err)
}

func TestPendingMiniBlockHeaders_RevertHeaderWrongHeaderTypeShouldErr(t *testing.T) {
	t.Parallel()

	pmb, _ := pendingMb.NewPendingMiniBlocks()
	header := &block.Header{}
	err := pmb.RevertHeader(header)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}
