package txcache

import (
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func Test_newVirtualSelectionSession(t *testing.T) {
	session := txcachemocks.NewSelectionSessionMock()
	vss := newVirtualSelectionSession(session)
	require.NotNil(t, vss)
}
