package handler

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

const separator = "_"

type receivedEvent struct {
	from                   string
	firstTimeReceivedMilli int64
	numReceived            int
	isCross                bool
}

type broadcastDebugHandler struct {
	ntpTime           NTPTime
	mutex             sync.RWMutex
	messageTypes      map[string]struct{}
	receivedBroadcast map[string]map[string]*receivedEvent
}

// NewBroadcastDebug will create a new instance of *interceptorTxDebug
func NewBroadcastDebug(config config.BroadcastStatisticsConfig, ntpTime NTPTime) (*broadcastDebugHandler, error) {
	if check.IfNil(ntpTime) {
		return nil, process.ErrNilSyncTimer
	}

	messageTypes := make(map[string]struct{})
	for _, messageType := range config.Messages {
		messageTypes[messageType] = struct{}{}
	}

	return &broadcastDebugHandler{
		receivedBroadcast: make(map[string]map[string]*receivedEvent),
		messageTypes:      messageTypes,
		ntpTime:           ntpTime,
	}, nil
}

// Process will process the intercept data and add statistics about p2p message
func (bd *broadcastDebugHandler) Process(data process.InterceptedData, msg p2p.MessageP2P, fromConnectedPeer core.PeerID) {
	if msg.BroadcastMethod() != p2p.Broadcast {
		return
	}

	bd.mutex.Lock()
	defer bd.mutex.Unlock()
	messageType := data.Type()
	if _, ok := bd.messageTypes[messageType]; !ok {
		return
	}

	_, found := bd.receivedBroadcast[messageType]
	if !found {
		bd.receivedBroadcast[messageType] = make(map[string]*receivedEvent)
	}

	hexHash := hex.EncodeToString(data.Hash())
	originatorPretty := msg.Peer().Pretty()
	mapID := computeMapID(hexHash, originatorPretty, msg.Topic())

	receivedE, found := bd.receivedBroadcast[messageType][mapID]
	if !found {
		bd.receivedBroadcast[messageType][mapID] = &receivedEvent{
			from:                   fromConnectedPeer.Pretty(),
			numReceived:            1,
			firstTimeReceivedMilli: bd.getCurrentTimeStampMilli(),
			isCross:                isCross(msg.Topic()),
		}
		return
	}

	receivedE.numReceived++
}

func (bd *broadcastDebugHandler) getCurrentTimeStampMilli() int64 {
	return bd.ntpTime.CurrentTime().UnixMilli()
}

// PrintReceivedTxsBroadcastAndCleanRecords will print information about received transactions from current epoch and clean records
func (bd *broadcastDebugHandler) PrintReceivedTxsBroadcastAndCleanRecords() {
	log.Info("Received broadcast information")

	bd.mutex.Lock()
	defer bd.mutex.Unlock()

	for messageType := range bd.receivedBroadcast {
		mapHashEvent := bd.receivedBroadcast[messageType]
		for id, et := range mapHashEvent {
			hash, originator := getHashAndOriginatorFromID(id)

			log.Debug("broadcast record",
				"hash", hash,
				"type", messageType,
				"originator", originator,
				"from", et.from,
				"first received", time.Unix(0, et.firstTimeReceivedMilli*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000"),
				"times received", et.numReceived,
				"is-cross", et.isCross)
		}
	}

	bd.receivedBroadcast = make(map[string]map[string]*receivedEvent)
}

func isCross(topic string) bool {
	split := strings.Split(topic, separator)

	return len(split) > 2
}

func computeMapID(key1, key2, key3 string) string {
	return fmt.Sprintf("%s%s%s%s%s", key1, separator, key2, separator, key3)
}

func getHashAndOriginatorFromID(id string) (string, string) {
	split := strings.Split(id, separator)
	return split[0], split[1]
}

// IsInterfaceNil returns true if there is no value under the interface
func (bd *broadcastDebugHandler) IsInterfaceNil() bool {
	return bd == nil
}
