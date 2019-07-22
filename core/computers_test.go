package core_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/magiconair/properties/assert"
)

func TestMaxShouldReturnA(t *testing.T) {
	a := uint32(11)
	b := uint32(10)
	assert.Equal(t, a, core.Max(a, b))
}

func TestMaxShouldReturnB(t *testing.T) {
	a := uint32(10)
	b := uint32(11)
	assert.Equal(t, b, core.Max(a, b))
}
