package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewSystemSCFactory_NilSystemEI(t *testing.T) {
	t.Parallel()

	scFactory, err := NewSystemSCFactory(nil, &mock.ValidatorSettingsStub{})

	assert.Nil(t, scFactory)
	assert.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
}

func TestNewSystemSCFactory_NilEconomicsData(t *testing.T) {
	t.Parallel()

	scFactory, err := NewSystemSCFactory(&mock.SystemEIStub{}, nil)

	assert.Nil(t, scFactory)
	assert.Equal(t, vm.ErrNilEconomicsData, err)
}

func TestNewSystemSCFactory_Ok(t *testing.T) {
	t.Parallel()

	scFactory, err := NewSystemSCFactory(&mock.SystemEIStub{}, &mock.ValidatorSettingsStub{})

	assert.Nil(t, err)
	assert.NotNil(t, scFactory)
}

func TestSystemSCFactory_Create(t *testing.T) {
	t.Parallel()

	scFactory, _ := NewSystemSCFactory(&mock.SystemEIStub{}, &mock.ValidatorSettingsStub{})

	container, err := scFactory.Create()
	assert.Nil(t, err)
	assert.Equal(t, 1, container.Len())
}

func TestSystemSCFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	scFactory, _ := NewSystemSCFactory(&mock.SystemEIStub{}, &mock.ValidatorSettingsStub{})
	assert.False(t, scFactory.IsInterfaceNil())

	scFactory = nil
	assert.True(t, scFactory.IsInterfaceNil())
}
