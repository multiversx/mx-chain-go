package statusHandler_test

import (
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTermuiStatusHandler_NewTermuiStatusHandler(t *testing.T) {
	t.Parallel()

	termuiStatusHandler, err := statusHandler.NewTermuiStatusHandler()
	assert.NotNil(t, termuiStatusHandler)
	assert.Nil(t, err)
}

func TestTermuiStatusHandler_TermuiShouldPass(t *testing.T) {
	t.Parallel()

	termuiStatusHandler, err := statusHandler.NewTermuiStatusHandler()
	termuiConsole := termuiStatusHandler.Termui()

	assert.NotNil(t, termuiConsole)
	assert.Nil(t, err)
}
