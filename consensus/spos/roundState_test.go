package spos_test

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRoundState(t *testing.T) {
	roundState := spos.NewRoundState()
	assert.NotNil(t, roundState)
}
