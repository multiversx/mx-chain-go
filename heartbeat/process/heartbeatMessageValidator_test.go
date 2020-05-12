package process

import (
	"crypto/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/stretchr/testify/require"
)

func TestTrimLengths(t *testing.T) {

	token1 := make([]byte, maxSizeInBytes+2)
	token2 := make([]byte, maxSizeInBytes+2)
	token3 := make([]byte, maxSizeInBytes+2)
	_, _ = rand.Read(token1)
	_, _ = rand.Read(token2)
	_, _ = rand.Read(token3)

	heartBeat := &data.Heartbeat{
		Payload:         token1,
		NodeDisplayName: string(token2),
		VersionNumber:   string(token3),
	}

	trimLengths(heartBeat)
	require.Equal(t, maxSizeInBytes, len(heartBeat.Payload))
	require.Equal(t, maxSizeInBytes, len(heartBeat.NodeDisplayName))
	require.Equal(t, maxSizeInBytes, len(heartBeat.VersionNumber))
}
