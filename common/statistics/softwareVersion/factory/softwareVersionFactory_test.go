package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/config"
	statusHandlerMock "github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestNewSoftwareVersionFactory_NilStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	factory, err := NewSoftwareVersionFactory(nil, config.SoftwareVersionConfig{})

	assert.Equal(t, core.ErrNilAppStatusHandler, err)
	assert.Nil(t, factory)
}

func TestSoftwareVersionFactory_Create(t *testing.T) {
	t.Parallel()

	statusHandler := &statusHandlerMock.AppStatusHandlerStub{}
	factory, _ := NewSoftwareVersionFactory(statusHandler, config.SoftwareVersionConfig{PollingIntervalInMinutes: 1})
	softwareVersionChecker, err := factory.Create()

	assert.Nil(t, err)
	assert.NotNil(t, softwareVersionChecker)
}
