//go:build !race

package transfers

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
)

func TestNotEnoughGas(t *testing.T) {
	context := wasm.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/esdt-safe.wasm", "@c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6")
	require.Nil(t, err)
	vault := context.ScAddress

	context.ScAddress = vault
	err = context.ExecuteSCWithValue(&context.Owner, "exec@c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6000000020000000a53564e2d6136616565660000000000000000000000000906aaf7c8516d0c000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a53564e2d6136616565660000000000000000000000000906aaf7c8516d0c000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f600", big.NewInt(0))
	require.Nil(t, err)
}

func TestTransfers_DuplicatedTransferValueEvents(t *testing.T) {
	context := wasm.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/transferValue/output/transferValue.wasm", "")
	require.Nil(t, err)
	vault := context.ScAddress

	err = context.DeploySC("../testdata/transferValue/output/transferValue.wasm", "")
	require.Nil(t, err)
	forwarder := context.ScAddress

	// Add money to the vault
	context.ScAddress = vault
	err = context.ExecuteSCWithValue(&context.Owner, "fund", big.NewInt(42))
	require.Nil(t, err)

	// Ask money from the vault, via the forwarder
	context.ScAddress = forwarder
	err = context.ExecuteSC(&context.Owner, fmt.Sprintf("forwardAskMoney@%s", hex.EncodeToString(vault)))
	require.Nil(t, err)
	require.Len(t, context.LastLogs, 1)
	require.Len(t, context.LastLogs[0].GetLogEvents(), 5)

	events := context.LastLogs[0].GetLogEvents()

	require.Equal(t, "transferValueOnly", string(events[0].GetIdentifier()))
	require.Equal(t, "AsyncCall", string(events[0].GetData()))
	require.Equal(t, []byte{}, events[0].GetTopics()[0])
	require.Equal(t, forwarder, events[0].GetAddress())
	require.Equal(t, vault, events[0].GetTopics()[1])

	require.Equal(t, "transferValueOnly", string(events[1].GetIdentifier()))
	require.Equal(t, "BackTransfer", string(events[1].GetData()))
	require.Equal(t, []byte{0x01}, events[1].GetTopics()[0])
	require.Equal(t, vault, events[1].GetAddress())
	require.Equal(t, forwarder, events[1].GetTopics()[1])

	// Duplicated "transferValueOnly" events are fixed in #5936.
	require.Equal(t, "transferValueOnly", string(events[2].GetIdentifier()))
	require.Equal(t, "AsyncCallback", string(events[2].GetData()))
	require.Equal(t, []byte{}, events[2].GetTopics()[0])
	require.Equal(t, vault, events[2].GetAddress())
	require.Equal(t, forwarder, events[2].GetTopics()[1])

	require.Equal(t, "writeLog", string(events[3].GetIdentifier()))
	require.Equal(t, "completedTxEvent", string(events[4].GetIdentifier()))
}
