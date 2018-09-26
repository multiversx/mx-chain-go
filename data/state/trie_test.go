package state

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

//go rules:
// "" = []byte{}
// "" = string(nil)
func TestByteSlicesStrings(t *testing.T) {
	b := make([]byte, 0)

	assert.Equal(t, "", string(b))

	b = nil

	assert.Equal(t, "", string(b))
}
