package preprocess_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/stretchr/testify/assert"
)

func TestNewGasConsumption_NilEconomicsFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	gc, err := preprocess.NewGasComputation(
		nil,
	)

	assert.Nil(t, gc)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewGasConsumption_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, err := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, gc)
	assert.Nil(t, err)
}
