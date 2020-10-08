package multisig

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDisabledMultiSig_MethodsShouldNotPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panic: %v", r))
		}
	}()

	dms := &DisabledMultiSig{}
	recoveredInstance, err := dms.Create(nil, 0)
	assert.NotNil(t, recoveredInstance)
	assert.Nil(t, err)

	recoveredBytes, err := dms.CreateSignatureShare(nil, nil)
	assert.Equal(t, []byte(signature), recoveredBytes)
	assert.Nil(t, err)

	recoveredBytes, err = dms.SignatureShare(0)
	assert.Equal(t, []byte(signature), recoveredBytes)
	assert.Nil(t, err)

	recoveredBytes, err = dms.AggregateSigs(nil)
	assert.Equal(t, []byte(signature), recoveredBytes)
	assert.Nil(t, err)

	assert.Nil(t, dms.SetAggregatedSig(nil))
	assert.Nil(t, dms.Verify(nil, nil))
	assert.False(t, check.IfNil(dms))
	assert.Nil(t, dms.Reset(nil, 0))
	assert.Nil(t, dms.StoreSignatureShare(0, nil))
	assert.Nil(t, dms.VerifySignatureShare(0, nil, nil, nil))
}
