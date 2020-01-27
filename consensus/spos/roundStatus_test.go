package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/stretchr/testify/assert"
)

func TestRoundStatus_NewRoundStatusShouldWork(t *testing.T) {
	t.Parallel()

	rstatus := spos.NewRoundStatus()
	assert.NotNil(t, rstatus)
}
