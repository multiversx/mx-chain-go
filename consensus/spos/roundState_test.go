package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/stretchr/testify/assert"
)

func TestNewRoundState(t *testing.T) {
	roundState := spos.NewRoundState()
	assert.NotNil(t, roundState)
}
