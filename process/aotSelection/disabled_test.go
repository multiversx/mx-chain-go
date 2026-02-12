package aotSelection

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestNewDisabledAOTSelector(t *testing.T) {
	t.Parallel()

	sel := NewDisabledAOTSelector()
	assert.NotNil(t, sel)
	assert.False(t, sel.IsInterfaceNil())
}

func TestDisabledAOTSelector_TriggerAOTSelectionDoesNothing(t *testing.T) {
	t.Parallel()

	sel := NewDisabledAOTSelector()
	header := &testscommon.HeaderHandlerStub{
		GetNonceCalled:    func() uint64 { return 10 },
		GetRandSeedCalled: func() []byte { return []byte("randomness") },
		EpochField:        1,
	}
	// Should not panic
	sel.TriggerAOTSelection(header, 100)
	sel.TriggerAOTSelection(nil, 100)
}

func TestDisabledAOTSelector_GetPreSelectedTransactionsReturnsFalse(t *testing.T) {
	t.Parallel()

	sel := NewDisabledAOTSelector()
	result, found := sel.GetPreSelectedTransactions(42)
	assert.Nil(t, result)
	assert.False(t, found)
}

func TestDisabledAOTSelector_CancelOngoingSelectionDoesNothing(t *testing.T) {
	t.Parallel()

	sel := NewDisabledAOTSelector()
	// Should not panic
	sel.CancelOngoingSelection()
}

func TestDisabledAOTSelector_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var sel *disabledAOTSelector
	assert.True(t, sel.IsInterfaceNil())

	sel = NewDisabledAOTSelector()
	assert.False(t, sel.IsInterfaceNil())
}

func TestDisabledAOTSelector_ImplementsInterface(t *testing.T) {
	t.Parallel()

	var _ process.AOTTransactionSelector = (*disabledAOTSelector)(nil)
}
