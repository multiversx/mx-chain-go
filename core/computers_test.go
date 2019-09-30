package core_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
)

func TestMaxShouldReturnA(t *testing.T) {
	a := int32(11)
	b := int32(10)
	assert.Equal(t, a, core.Max(a, b))
}

func TestMaxShouldReturnB(t *testing.T) {
	a := int32(10)
	b := int32(11)
	assert.Equal(t, b, core.Max(a, b))
}

func TestMinShouldReturnB(t *testing.T) {
	a := int32(11)
	b := int32(10)
	assert.Equal(t, b, core.Min(a, b))
}

func TestMinShouldReturnA(t *testing.T) {
	a := int32(10)
	b := int32(11)
	assert.Equal(t, a, core.Min(a, b))
}
