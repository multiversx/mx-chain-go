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
	require.Len(t, context.LastLogs[0].GetLogEvents(), 3)

	events := context.LastLogs[0].GetLogEvents()

	// Duplicated "transferValueOnly" events are fixed in #5936.
	require.Equal(t, "transferValueOnly", string(events[0].GetIdentifier()))
	require.Equal(t, "BackTransfer", string(events[0].GetData()))
	require.Equal(t, []byte{0x01}, events[0].GetTopics()[0])

	require.Equal(t, "writeLog", string(events[1].GetIdentifier()))
	require.Len(t, events[1].GetTopics(), 2)
	require.Contains(t, string(events[1].GetTopics()[1]), "too much gas provided for processing")
	require.Equal(t, "completedTxEvent", string(events[2].GetIdentifier()))
}
