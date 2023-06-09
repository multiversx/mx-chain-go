package coordinator

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/assert"
)

func TestNewSovereignChainTransactionCoordinator_ShouldErrNilTransactionCoordinator(t *testing.T) {
	t.Parallel()

	sctc, err := NewSovereignChainTransactionCoordinator(nil)
	assert.Nil(t, sctc)
	assert.Equal(t, process.ErrNilTransactionCoordinator, err)
}

func TestNewSovereignChainTransactionCoordinator_ShouldWork(t *testing.T) {
	t.Parallel()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments()
	tc, _ := NewTransactionCoordinator(argsTransactionCoordinator)

	sctc, err := NewSovereignChainTransactionCoordinator(tc)
	assert.NotNil(t, sctc)
	assert.Nil(t, err)
}

//TODO: More unit tests should be added. Created PR https://multiversxlabs.atlassian.net/browse/MX-14149
