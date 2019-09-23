package dataValidators_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/stretchr/testify/assert"
)

func TestNilTxValidator(t *testing.T) {
	t.Parallel()

	ntv, err := dataValidators.NewNilTxValidator()

	assert.False(t, check.IfNil(ntv))
	assert.Nil(t, err)
}

func TestNilTxValidator_IsTxValidForProcessing(t *testing.T) {
	t.Parallel()

	ntv, _ := dataValidators.NewNilTxValidator()

	assert.True(t, ntv.IsTxValidForProcessing(nil))
}

//------- IsInterfaceNil

func TestNilTxValidator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	ntxv, _ := dataValidators.NewNilTxValidator()
	ntxv = nil

	assert.True(t, check.IfNil(ntxv))
}
