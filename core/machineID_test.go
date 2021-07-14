package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAnonymizedMachineID(t *testing.T) {
	t.Parallel()

	firstVariant := GetAnonymizedMachineID("first")
	secondVariant := GetAnonymizedMachineID("second")

	assert.NotEqual(t, firstVariant, secondVariant)
	assert.Equal(t, maxMachineIDLen, len(firstVariant))
	assert.Equal(t, maxMachineIDLen, len(secondVariant))
}
