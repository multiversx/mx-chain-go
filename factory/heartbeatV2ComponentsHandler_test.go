package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/stretchr/testify/assert"
)

func TestManagedHeartbeatV2Components(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	mhc, err := factory.NewManagedHeartbeatV2Components(nil)
	assert.True(t, check.IfNil(mhc))
	assert.Equal(t, errors.ErrNilHeartbeatV2ComponentsFactory, err)

	args := createMockHeartbeatV2ComponentsFactoryArgs()
	hcf, _ := factory.NewHeartbeatV2ComponentsFactory(args)
	mhc, err = factory.NewManagedHeartbeatV2Components(hcf)
	assert.False(t, check.IfNil(mhc))
	assert.Nil(t, err)

	err = mhc.Create()
	assert.Nil(t, err)

	err = mhc.CheckSubcomponents()
	assert.Nil(t, err)

	assert.Equal(t, "managedHeartbeatV2Components", mhc.String())

	err = mhc.Close()
	assert.Nil(t, err)
}
