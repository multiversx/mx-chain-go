package spos_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/stretchr/testify/assert"
)

func TestConsensusData_NewConsensusDataShouldWork(t *testing.T) {
	cnsdta := spos.NewConsensusData(
		nil,
		nil,
		nil,
		nil,
		-1,
		0,
		0)

	assert.NotNil(t, cnsdta)
}

func TestConsensusData_ConsensusDataCreateShouldReturnTheSameObject(t *testing.T) {
	cnsdta := spos.NewConsensusData(
		nil,
		nil,
		nil,
		nil,
		0,
		0,
		0)

	assert.Equal(t, cnsdta, cnsdta.Create())
}

func TestConsensusData_ConsensusDataIDShouldReturnID(t *testing.T) {
	cnsdta := spos.NewConsensusData(
		nil,
		nil,
		nil,
		[]byte("sig"),
		6,
		0,
		1)

	id := fmt.Sprintf("1-sig-6")

	assert.Equal(t, id, cnsdta.ID())
}
