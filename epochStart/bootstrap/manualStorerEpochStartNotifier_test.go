package bootstrap

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewManualStorerEpochStartNotifier_NilManualHandlerShouldErr(t *testing.T) {
	t.Parallel()

	msesn, err := NewManualStorerEpochStartNotifier(nil)

	assert.True(t, check.IfNil(msesn))
	assert.Equal(t, dataRetriever.ErrNilManualEpochStartNotifier, err)
}

func TestNewManualStorerEpochStartNotifier_ShouldWork(t *testing.T) {
	t.Parallel()

	msesn, err := NewManualStorerEpochStartNotifier(&mock.ManualEpochStartNotifierStub{})

	assert.False(t, check.IfNil(msesn))
	assert.Nil(t, err)
}

func TestManualStorerEpochStartNotifier_RegisterHandlerShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not paniced %v", r))
		}
	}()

	msesn, _ := NewManualStorerEpochStartNotifier(&mock.ManualEpochStartNotifierStub{})

	msesn.RegisterHandler(nil)
}
