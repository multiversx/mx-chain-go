package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewSoftwareVersionFactory_NilStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	softwareVersionFactory, err := NewSoftwareVersionFactory(nil, config.SoftwareVersionConfig{})

	assert.Equal(t, core.ErrNilAppStatusHandler, err)
	assert.Nil(t, softwareVersionFactory)
}

func TestSoftwareVersionFactory_Create(t *testing.T) {
	t.Parallel()

	statusHandler := &mock.AppStatusHandlerStub{}
	softwareVersionFactory, _ := NewSoftwareVersionFactory(statusHandler, config.SoftwareVersionConfig{PollingIntervalInMinutes: 1})
	softwareVersionChecker, err := softwareVersionFactory.Create()

	assert.Nil(t, err)
	assert.NotNil(t, softwareVersionChecker)
}
