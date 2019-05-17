package core_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/stretchr/testify/assert"
)

func TestSubroundStartRound_GetPkToDisplayShouldTrim(t *testing.T) {
	pk := "1234567891234"
	pkToDisplay := core.GetTrimmedPk(pk)
	assert.Equal(t, "123456789123...", pkToDisplay)
}

func TestSubroundStartRound_GetPkToDisplayShouldNotTrim(t *testing.T) {
	pk := "123456789123"
	pkToDisplay := core.GetTrimmedPk(pk)
	assert.Equal(t, pk, pkToDisplay)
}
