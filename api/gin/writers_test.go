package gin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGinWriter_Write(t *testing.T) {
	t.Parallel()

	gw := ginWriter{}

	providedBuff := []byte("provided buff")
	l, err := gw.Write(providedBuff)
	assert.Nil(t, err)
	assert.Equal(t, len(providedBuff), l)
}

func TestGinErrorWriter_Write(t *testing.T) {
	t.Parallel()

	gew := ginErrorWriter{}

	providedBuff := []byte("provided buff")
	l, err := gew.Write(providedBuff)
	assert.Nil(t, err)
	assert.Equal(t, len(providedBuff), l)
}
