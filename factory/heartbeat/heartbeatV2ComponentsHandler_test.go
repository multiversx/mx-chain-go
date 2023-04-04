package heartbeat_test

import (
	"testing"

	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	heartbeatComp "github.com/multiversx/mx-chain-go/factory/heartbeat"
	"github.com/stretchr/testify/assert"
)

func TestNewManagedHeartbeatV2Components(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	mhc, err := heartbeatComp.NewManagedHeartbeatV2Components(nil)
	assert.Nil(t, mhc)
	assert.Equal(t, errorsMx.ErrNilHeartbeatV2ComponentsFactory, err)

	args := createMockHeartbeatV2ComponentsFactoryArgs()
	hcf, _ := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
	mhc, err = heartbeatComp.NewManagedHeartbeatV2Components(hcf)
	assert.NotNil(t, mhc)
	assert.NoError(t, err)
}

func TestManagedHeartbeatV2Components_Create(t *testing.T) {
	t.Parallel()

	t.Run("invalid config should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.Config.HeartbeatV2.PeerAuthenticationTimeBetweenSendsInSec = 0 // Create will fail
		hcf, _ := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		mhc, _ := heartbeatComp.NewManagedHeartbeatV2Components(hcf)
		assert.NotNil(t, mhc)
		err := mhc.Create()
		assert.Error(t, err)
	})
	t.Run("should work with getters", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		hcf, _ := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		mhc, _ := heartbeatComp.NewManagedHeartbeatV2Components(hcf)
		assert.NotNil(t, mhc)
		assert.Nil(t, mhc.Monitor())

		err := mhc.Create()
		assert.NoError(t, err)
		assert.NotNil(t, mhc.Monitor())

		assert.Equal(t, factory.HeartbeatV2ComponentsName, mhc.String())

		assert.NoError(t, mhc.Close())
	})
}

func TestManagedHeartbeatV2Components_CheckSubcomponents(t *testing.T) {
	t.Parallel()

	args := createMockHeartbeatV2ComponentsFactoryArgs()
	hcf, _ := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
	mhc, _ := heartbeatComp.NewManagedHeartbeatV2Components(hcf)
	assert.NotNil(t, mhc)
	assert.Equal(t, errorsMx.ErrNilHeartbeatV2Components, mhc.CheckSubcomponents())

	err := mhc.Create()
	assert.NoError(t, err)
	assert.Nil(t, mhc.CheckSubcomponents())

	assert.NoError(t, mhc.Close())
}

func TestManagedHeartbeatV2Components_Close(t *testing.T) {
	t.Parallel()

	args := createMockHeartbeatV2ComponentsFactoryArgs()
	hcf, _ := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
	mhc, _ := heartbeatComp.NewManagedHeartbeatV2Components(hcf)
	assert.NotNil(t, mhc)
	assert.NoError(t, mhc.Close())

	err := mhc.Create()
	assert.NoError(t, err)
	assert.NoError(t, mhc.Close())
}

func TestManagedHeartbeatV2Components_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	mhc, _ := heartbeatComp.NewManagedHeartbeatV2Components(nil)
	assert.True(t, mhc.IsInterfaceNil())

	args := createMockHeartbeatV2ComponentsFactoryArgs()
	hcf, _ := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
	mhc, _ = heartbeatComp.NewManagedHeartbeatV2Components(hcf)
	assert.False(t, mhc.IsInterfaceNil())
}
