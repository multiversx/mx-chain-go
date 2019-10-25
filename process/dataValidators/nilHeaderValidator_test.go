package dataValidators_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/stretchr/testify/assert"
)

func TestNilHeaderValidator(t *testing.T) {
	t.Parallel()

	nhhv, err := dataValidators.NewNilHeaderValidator()

	assert.False(t, check.IfNil(nhhv))
	assert.Nil(t, err)
}

func TestNilHeaderValidator_IsHeaderValidForProcessing(t *testing.T) {
	t.Parallel()

	nhv, _ := dataValidators.NewNilHeaderValidator()

	assert.Nil(t, nhv.HeaderValidForProcessing(nil))
}

//------- IsInterfaceNil

func TestNilHeaderValidator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	hdrValidator, _ := dataValidators.NewNilHeaderValidator()
	hdrValidator = nil

	assert.True(t, check.IfNil(hdrValidator))
}
