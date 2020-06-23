package consensus_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/stretchr/testify/assert"
)

func TestConsensusMessage_NewConsensusMessageShouldWork(t *testing.T) {
	t.Parallel()

	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		-1,
		0,
		[]byte("chain ID"),
		nil,
		nil,
		nil,
		"pid",
	)

	assert.NotNil(t, cnsMsg)
}
