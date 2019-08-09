package dataValidators_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/stretchr/testify/assert"
)

func TestNilHeaderHandlerProcessValidator(t *testing.T) {
	t.Parallel()

	nhhv, err := dataValidators.NewNilHeaderHandlerProcessValidator()

	assert.NotNil(t, nhhv)
	assert.Nil(t, err)
}

func TestNilHeaderHandlerProcessValidator_CheckHeaderHandlerValid(t *testing.T) {
	t.Parallel()

	nhhv, _ := dataValidators.NewNilHeaderHandlerProcessValidator()

	assert.True(t, nhhv.CheckHeaderHandlerValid(nil))
}
