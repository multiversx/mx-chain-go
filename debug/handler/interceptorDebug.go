package handler

import (
	"encoding/hex"
	"sync"
	"time"

	p2p2 "github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

type receivedEvent struct {
	from                   string
	originator             string
	firstTimeReceivedMilli int64
	numReceived            int
}

type interceptorDebug struct {
	mutex             sync.RWMutex
	messageTypes      map[string]struct{}
	receivedBroadcast map[string]map[string]*receivedEvent
}

// NewInterceptorDebug will create a new instance of *interceptorTxDebug
func NewInterceptorDebug(config config.BroadcastStatisticsConfig) *interceptorDebug {
	messageTypes := make(map[string]struct{})
	for _, messageType := range config.Messages {
		messageTypes[messageType] = struct{}{}
	}

	return &interceptorDebug{
		receivedBroadcast: make(map[string]map[string]*receivedEvent),
		messageTypes:      messageTypes,
	}
}

// Process will process the intercept data and add statistics about p2p message
func (id *interceptorDebug) Process(data process.InterceptedData, msg p2p.MessageP2P, fromConnectedPeer core.PeerID) {
	if msg.BroadcastMethod() != p2p2.Broadcast {
		return
	}

	id.mutex.Lock()
	defer id.mutex.Unlock()
	messageType := data.Type()
	if _, ok := id.messageTypes[messageType]; !ok {
		return
	}

	_, found := id.receivedBroadcast[messageType]
	if !found {
		id.receivedBroadcast[messageType] = make(map[string]*receivedEvent)
	}

	hexHash := hex.EncodeToString(data.Hash())
	receivedE, found := id.receivedBroadcast[messageType][hexHash]
	if !found {
		id.receivedBroadcast[messageType][hexHash] = &receivedEvent{
			originator:             core.PeerID(msg.From()).Pretty(),
			from:                   fromConnectedPeer.Pretty(),
			numReceived:            1,
			firstTimeReceivedMilli: getCurrentTimeStampMilli(),
		}
		return
	}

	receivedE.numReceived++
}

func getCurrentTimeStampMilli() int64 {
	return time.Now().UnixMilli()
}

// PrintReceivedTxsBroadcastAndCleanRecords will print information about received transactions from current epoch and clean records
func (id *interceptorDebug) PrintReceivedTxsBroadcastAndCleanRecords() {
	log.Info("Received Transactions Broadcast Information")

	id.mutex.Lock()
	defer id.mutex.Unlock()

	for messageType := range id.receivedBroadcast {
		mapHashEvent := id.receivedBroadcast[messageType]
		for hash, et := range mapHashEvent {
			log.Debug("broadcast record",
				"hash", hash,
				"type", messageType,
				"originator", et.originator,
				"from", et.from,
				"first received", time.Unix(0, et.firstTimeReceivedMilli*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000"),
				"times received", et.numReceived)
		}
	}

	id.receivedBroadcast = make(map[string]map[string]*receivedEvent)
}
