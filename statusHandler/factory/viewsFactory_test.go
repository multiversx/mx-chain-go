package factory

import (
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestViewsFactory_NewViewFactoryShouldError(t *testing.T) {
	t.Parallel()

	viewsFactory, err := NewViewsFactory(nil)

	assert.Nil(t, viewsFactory)
	assert.Error(t, statusHandler.ErrorNilPresenterInterface, err)
}

func TestViewsFactory_Create(t *testing.T) {
	t.Parallel()

	presenterFactory := NewPresenterFactory()
	presenterStatusHandler := presenterFactory.Create()

	viewsFactory, err := NewViewsFactory(presenterStatusHandler)
	assert.Nil(t, err)
	assert.NotNil(t, viewsFactory)

	views, err := viewsFactory.Create()

	assert.NotNil(t, views)
	assert.Nil(t, err)
}
