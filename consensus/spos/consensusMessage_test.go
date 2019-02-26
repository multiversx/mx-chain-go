package spos_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/stretchr/testify/assert"
)

func TestConsensusMessage_NewConsensusMessageShouldWork(t *testing.T) {
	t.Parallel()

	cnsMsg := spos.NewConsensusMessage(
		nil,
		nil,
		nil,
		nil,
		-1,
		0,
		0)

	assert.NotNil(t, cnsMsg)
}

func TestConsensusMessage_ConsensusMessageCreateShouldReturnTheSameObject(t *testing.T) {
	t.Parallel()

	cnsMsg := spos.NewConsensusMessage(
		nil,
		nil,
		nil,
		nil,
		0,
		0,
		0)

	assert.Equal(t, cnsMsg, cnsMsg.Create())
}

func TestConsensusMessage_ConsensusMessageIDShouldReturnID(t *testing.T) {
	t.Parallel()

	cnsMsg := spos.NewConsensusMessage(
		nil,
		nil,
		nil,
		[]byte("sig"),
		6,
		0,
		1)

	id := fmt.Sprintf("1-sig-6")

	assert.Equal(t, id, cnsMsg.ID())
}
