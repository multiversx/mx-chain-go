package dataValidators_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/stretchr/testify/assert"
)

func TestNilHeaderValidator(t *testing.T) {
	t.Parallel()

	nhhv, err := dataValidators.NewNilHeaderValidator()

	assert.NotNil(t, nhhv)
	assert.Nil(t, err)
}

func TestNilHeaderValidator_IsHeaderValidForProcessing(t *testing.T) {
	t.Parallel()

	nhv, _ := dataValidators.NewNilHeaderValidator()

	assert.True(t, nhv.IsHeaderValidForProcessing(nil))
}
