package heartbeat

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrimLengths(t *testing.T) {

	token1 := make([]byte, maxSizeInBytes+2)
	token2 := make([]byte, maxSizeInBytes+2)
	token3 := make([]byte, maxSizeInBytes+2)
	rand.Read(token1)
	rand.Read(token2)
	rand.Read(token3)

	heartBeat := &Heartbeat{
		Payload:         token1,
		NodeDisplayName: string(token2),
		VersionNumber:   string(token3),
	}

	trimLengths(heartBeat)
	require.Equal(t, maxSizeInBytes, len(heartBeat.Payload))
	require.Equal(t, maxSizeInBytes, len(heartBeat.NodeDisplayName))
	require.Equal(t, maxSizeInBytes, len(heartBeat.VersionNumber))
}
