package endProcess

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDummyEndProcessChannel(t *testing.T) {
	emptyChannel := GetDummyEndProcessChannel()
	require.Empty(t, emptyChannel)

	emptyChannel <- ArgEndProcess{
		Reason: "reason",
	}
	require.NotEmpty(t, emptyChannel)
}
