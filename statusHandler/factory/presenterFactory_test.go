package factory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPresenterFactory_Create(t *testing.T) {
	t.Parallel()

	presenterFactory := NewPresenterFactory()
	presenterStatusHandler := presenterFactory.Create()

	assert.NotNil(t, presenterStatusHandler)
}
