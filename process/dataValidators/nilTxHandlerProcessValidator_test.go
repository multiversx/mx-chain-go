package dataValidators_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/stretchr/testify/assert"
)

func TestNilTxHandlerProcessValidator(t *testing.T) {
	t.Parallel()

	nthv, err := dataValidators.NewNilTxHandlerProcessValidator()

	assert.NotNil(t, nthv)
	assert.Nil(t, err)
}

func TestNilTxHandlerProcessValidator_CheckTxHandlerValid(t *testing.T) {
	t.Parallel()

	nthv, _ := dataValidators.NewNilTxHandlerProcessValidator()

	assert.True(t, nthv.CheckTxHandlerValid(nil))
}
