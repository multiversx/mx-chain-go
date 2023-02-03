package consensus_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/consensus"
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
		nil,
	)

	assert.NotNil(t, cnsMsg)
}
