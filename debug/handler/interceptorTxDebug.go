package handler

import (
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"

	p2p2 "github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

const (
	interceptedTx         = "intercepted tx"
	interceptedRewardTx   = "intercepted reward tx"
	interceptedUnsignedTx = "intercepted unsigned tx"
)

type receivedTxEvent struct {
	mutex                 sync.Mutex
	from                  string
	originator            string
	firstTimeReceivedMile int64
	numReceived           int
	txType                string
	// TODO check what fields to add
}

type interceptorTxDebug struct {
	receivedTxsBroadcast map[string]*receivedTxEvent
}

func NewInterceptorTxDebug() *interceptorTxDebug {
	return &interceptorTxDebug{
		receivedTxsBroadcast: make(map[string]*receivedTxEvent),
	}
}

func (id *interceptorTxDebug) Process(data process.InterceptedData, msg p2p.MessageP2P, fromConnectedPeer core.PeerID) {
	if msg.BroadcastMethod() != p2p2.Broadcast {
		return
	}

	isCorrectType := data.Type() == interceptedTx // && data.Type() == interceptedRewardTx && data.Type() == interceptedUnsignedTx
	if !isCorrectType {
		return
	}

	receivedTxE, found := id.receivedTxsBroadcast[string(data.Hash())]
	if !found {
		hexHash := hex.EncodeToString(data.Hash())
		id.receivedTxsBroadcast[hexHash] = &receivedTxEvent{
			originator:            core.PeerID(msg.From()).Pretty(),
			from:                  fromConnectedPeer.Pretty(),
			numReceived:           1,
			firstTimeReceivedMile: getCurrentTimeStampMilli(),
			txType:                data.Type(),
		}
		return
	}

	receivedTxE.mutex.Lock()
	receivedTxE.numReceived++
	receivedTxE.mutex.Unlock()
}

func getCurrentTimeStampMilli() int64 {
	return time.Now().UnixMilli()
}

func (id *interceptorTxDebug) PrintReceivedTxsBroadcastAndCleanRecords() {
	log.Info("Received Transactions Broadcast Information",
		"table", fmt.Sprintf("\n%s", id.getReceivedTxsBroadcastTable()))

	id.receivedTxsBroadcast = make(map[string]*receivedTxEvent)
}

func (id *interceptorTxDebug) getReceivedTxsBroadcastTable() string {
	// Create header with fixed widths
	header := fmt.Sprintf("%-64s %-25s %-60s  %-60s %-25s %-15s\n",
		"Hash",
		"Type",
		"Originator",
		"From",
		"First Received",
		"Times Received")
	separator := fmt.Sprintf("%-64s %-25s %-60s %-60s %-25s %-15s\n",
		"----------------------------------------------------------------",
		"-------------------------",
		"----------------------------------------",
		"----------------------------------------",
		"-------------------------",
		"---------------")

	// Create rows with consistent formatting
	var rows string
	for hash, et := range id.receivedTxsBroadcast {
		et.mutex.Lock()
		row := fmt.Sprintf("%-64s %-25s %-60s %-60s %-25s %-15d\n",
			hash,
			et.txType,
			et.originator,
			et.from,
			time.Unix(0, et.firstTimeReceivedMile*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000"),
			et.numReceived)
		et.mutex.Unlock()
		rows += row
	}

	return header + separator + rows
}
