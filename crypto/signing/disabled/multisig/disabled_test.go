package multisig

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDisabled_MethodsShouldNotPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panic: %v", r))
		}
	}()

	d := &Disabled{}
	recovInstance, err := d.Create(nil, 0)
	assert.NotNil(t, recovInstance)
	assert.Nil(t, err)

	recovBytes, err := d.CreateSignatureShare(nil, nil)
	assert.Equal(t, []byte(signature), recovBytes)
	assert.Nil(t, err)

	recovBytes, err = d.SignatureShare(0)
	assert.Equal(t, []byte(signature), recovBytes)
	assert.Nil(t, err)

	recovBytes, err = d.AggregateSigs(nil)
	assert.Equal(t, []byte(signature), recovBytes)
	assert.Nil(t, err)

	assert.Nil(t, d.SetAggregatedSig(nil))
	assert.Nil(t, d.Verify(nil, nil))
	assert.False(t, check.IfNil(d))
	assert.Nil(t, d.Reset(nil, 0))
	assert.Nil(t, d.StoreSignatureShare(0, nil))
	assert.Nil(t, d.VerifySignatureShare(0, nil, nil, nil))
}
