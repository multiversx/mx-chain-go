package handler

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	p2p2 "github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/require"
)

func TestGetReceivedTxsBroadcastTable_With10Records(t *testing.T) {
	dbg := &interceptorTxDebug{
		receivedTxsBroadcast: make(map[string]*receivedTxEvent),
	}

	baseTime := time.Now().UnixMilli()
	for i := 0; i < 10; i++ {
		hash := fmt.Sprintf("hash%02d", i)
		dbg.receivedTxsBroadcast[hash] = &receivedTxEvent{
			from:                  fmt.Sprintf("peer%02d", i),
			numReceived:           i + 1,
			firstTimeReceivedMile: baseTime + int64(i*1000),
			txType:                interceptedTx,
		}
	}

	table := dbg.getReceivedTxsBroadcastTable()
	fmt.Println(table)

	// Check that all hashes are present in the table
	for i := 0; i < 10; i++ {
		txHash := fmt.Sprintf("hash%02d", i)
		if !strings.Contains(table, txHash) {
			t.Errorf("Expected hash %s in table output", hash)
		}
	}
}

func TestProcessAndVerifyTable(t *testing.T) {
	dbg := NewInterceptorTxDebug()

	// Create test data
	testData := []struct {
		hash       string
		txType     string
		originator string
		from       string
	}{
		{"hash1", interceptedTx, "peer1", "p1"},
		{"hash2", interceptedTx, "peer2", "p2"},
		{"hash3", interceptedTx, "peer3", "p3"},
		{"hash4", interceptedTx, "peer4", "p4"},
		{"hash5", interceptedTx, "peer5", "p5"},
	}

	// Process each record
	for _, td := range testData {
		data := testscommon.InterceptedDataStub{
			HashCalled: func() []byte {
				return []byte(td.hash)
			},
			TypeCalled: func() string {
				return td.txType
			},
		}

		originator, err := core.NewPeerID(td.originator)
		require.NoError(t, err)
		msg := &p2pmocks.P2PMessageMock{
			FromField:            []byte(originator),
			BroadcastMethodField: p2p2.Broadcast,
		}
		from, err := core.NewPeerID(td.from)
		require.NoError(t, err)
		dbg.Process(&data, msg, from)
	}

	// Get the table output
	table := dbg.getReceivedTxsBroadcastTable()
	fmt.Println("Table after processing 5 records:")
	fmt.Println(table)

	// Verify all records are present
	for _, td := range testData {
		// The hash is stored as hex in the map
		hexHash := hex.EncodeToString([]byte(td.hash))
		if !strings.Contains(table, hexHash) {
			t.Errorf("Expected hash %s in table output", hexHash)
		}
		if !strings.Contains(table, td.txType) {
			t.Errorf("Expected type %s in table output", td.txType)
		}
		// The peer ID is converted to base58 in the table
		originatorID, _ := core.NewPeerID(td.originator)
		if !strings.Contains(table, originatorID.Pretty()) {
			t.Errorf("Expected peer ID %s in table output", originatorID)
		}

		fromID, _ := core.NewPeerID(td.from)
		if !strings.Contains(table, fromID.Pretty()) {
			t.Errorf("Expected peer ID %s in table output", fromID)
		}
	}

	// Verify the number of records
	expectedRecords := len(testData)
	actualRecords := strings.Count(table, "\n") - 2 // subtract header and separator
	if actualRecords != expectedRecords {
		t.Errorf("Expected %d records in table, got %d", expectedRecords, actualRecords)
	}
}
