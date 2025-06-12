package handler

import (
	"encoding/hex"
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
	from                  string
	originator            string
	firstTimeReceivedMile int64
	numReceived           int
	txType                string
	// TODO check what fields to add
}

type interceptorTxDebug struct {
	mutex                sync.RWMutex
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

	id.mutex.Lock()
	defer id.mutex.Unlock()

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

	receivedTxE.numReceived++
}

func getCurrentTimeStampMilli() int64 {
	return time.Now().UnixMilli()
}

func (id *interceptorTxDebug) PrintReceivedTxsBroadcastAndCleanRecords() {
	log.Info("Received Transactions Broadcast Information")

	id.mutex.Lock()
	defer id.mutex.Unlock()

	for hash, et := range id.receivedTxsBroadcast {
		log.Debug("broadcast record",
			"hash", hash,
			"type", et.txType,
			"originator", et.originator,
			"from", et.from,
			"first received", time.Unix(0, et.firstTimeReceivedMile*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000"),
			"times received", et.numReceived)
	}

	id.receivedTxsBroadcast = make(map[string]*receivedTxEvent)
}
