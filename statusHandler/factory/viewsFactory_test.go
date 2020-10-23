package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestViewsFactory_NewViewFactoryShouldError(t *testing.T) {
	t.Parallel()

	vf, err := NewViewsFactory(nil, 1)

	assert.Nil(t, vf)
	assert.Error(t, statusHandler.ErrNilPresenterInterface, err)
}

func TestViewsFactory_Create(t *testing.T) {
	t.Parallel()

	pf := NewPresenterFactory()
	presenterStatusHandler := pf.Create()

	vf, err := NewViewsFactory(presenterStatusHandler, 1)
	assert.Nil(t, err)
	assert.NotNil(t, vf)

	views, err := vf.Create()

	assert.NotNil(t, views)
	assert.Nil(t, err)
}
