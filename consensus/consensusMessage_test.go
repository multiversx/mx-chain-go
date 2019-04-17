package consensus_test

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsensusMessage_NewConsensusMessageShouldWork(t *testing.T) {
	t.Parallel()

	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		nil,
		nil,
		-1,
		0,
		0)

	assert.NotNil(t, cnsMsg)
}
