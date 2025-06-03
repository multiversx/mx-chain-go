package handler

import (
	"testing"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)


func TestPrintReceivedTxsBroadcastAndCleanRecords(t *testing.T) {
	dbg := &interceptorTxDebug{
		receivedTxsBroadcast: make(map[string]*receivedTxEvent),
	}

	// Add test records
	baseTime := time.Now().UnixMilli()
	testRecords := []struct {
		hash        string
		txType      string
		originator  string
		from        string
		numReceived int
	}{
		{"hash1", interceptedTx, "originator1", "from1", 1},
		{"hash2", interceptedTx, "originator2", "from2", 2},
		{"hash3", interceptedTx, "originator3", "from3", 3},
	}

	for _, tr := range testRecords {
		dbg.receivedTxsBroadcast[tr.hash] = &receivedTxEvent{
			from:                  tr.from,
			originator:            tr.originator,
			numReceived:           tr.numReceived,
			firstTimeReceivedMile: baseTime,
			txType:                tr.txType,
		}
	}

	_ = logger.SetLogLevel("*:DEBUG")

	// Call the function
	dbg.PrintReceivedTxsBroadcastAndCleanRecords()

	// Verify that the map is cleared
	require.Equal(t, 0, len(dbg.receivedTxsBroadcast))

	// Verify that all records were processed
	for _, tr := range testRecords {
		_, exists := dbg.receivedTxsBroadcast[tr.hash]
		require.False(t, exists, "Record %s should have been cleared", tr.hash)
	}
}
