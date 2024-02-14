//go:build !race

package transfers

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/stretchr/testify/require"
)

func TestTransfers_DuplicatedTransferValueEvents(t *testing.T) {
	context := wasm.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/transferValue/output/transferValue.wasm", "")
	require.Nil(t, err)

	err = context.ExecuteSCWithValue(&context.Owner, "receive", big.NewInt(1))
	require.Nil(t, err)
	require.Len(t, context.LastLogs, 1)
	require.Len(t, context.LastLogs[0].GetLogEvents(), 4)

	events := context.LastLogs[0].GetLogEvents()

	// There are duplicated events, to be fixed here:
	// https://github.com/multiversx/mx-chain-go/pull/5936
	require.Equal(t, "transferValueOnly", string(events[0].GetIdentifier()))
	require.Equal(t, "AsyncCall", string(events[0].GetData()))
	require.Equal(t, []byte{0x01}, events[0].GetTopics()[0])

	require.Equal(t, "transferValueOnly", string(events[1].GetIdentifier()))
	require.Equal(t, "BackTransfer", string(events[1].GetData()))
	require.Equal(t, []byte{0x01}, events[1].GetTopics()[0])
}
