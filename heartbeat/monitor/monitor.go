package monitor

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/heartbeat/data"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("heartbeat/monitor")

const minDuration = time.Second

// ArgHeartbeatV2Monitor holds the arguments needed to create a new instance of heartbeatV2Monitor
type ArgHeartbeatV2Monitor struct {
	Cache                         storage.Cacher
	PubKeyConverter               core.PubkeyConverter
	Marshaller                    marshal.Marshalizer
	MaxDurationPeerUnresponsive   time.Duration
	HideInactiveValidatorInterval time.Duration
	ShardId                       uint32
	PeerTypeProvider              heartbeat.PeerTypeProviderHandler
}

type heartbeatV2Monitor struct {
	cache                         storage.Cacher
	pubKeyConverter               core.PubkeyConverter
	marshaller                    marshal.Marshalizer
	maxDurationPeerUnresponsive   time.Duration
	hideInactiveValidatorInterval time.Duration
	shardId                       uint32
	peerTypeProvider              heartbeat.PeerTypeProviderHandler
}

// NewHeartbeatV2Monitor creates a new instance of heartbeatV2Monitor
func NewHeartbeatV2Monitor(args ArgHeartbeatV2Monitor) (*heartbeatV2Monitor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	hbv2Monitor := &heartbeatV2Monitor{
		cache:                         args.Cache,
		pubKeyConverter:               args.PubKeyConverter,
		marshaller:                    args.Marshaller,
		maxDurationPeerUnresponsive:   args.MaxDurationPeerUnresponsive,
		hideInactiveValidatorInterval: args.HideInactiveValidatorInterval,
		shardId:                       args.ShardId,
		peerTypeProvider:              args.PeerTypeProvider,
	}

	return hbv2Monitor, nil
}

func checkArgs(args ArgHeartbeatV2Monitor) error {
	if check.IfNil(args.Cache) {
		return heartbeat.ErrNilCacher
	}
	if check.IfNil(args.PubKeyConverter) {
		return heartbeat.ErrNilPubkeyConverter
	}
	if check.IfNil(args.Marshaller) {
		return heartbeat.ErrNilMarshaller
	}
	if args.MaxDurationPeerUnresponsive < minDuration {
		return fmt.Errorf("%w on MaxDurationPeerUnresponsive, provided %d, min expected %d",
			heartbeat.ErrInvalidTimeDuration, args.MaxDurationPeerUnresponsive, minDuration)
	}
	if args.HideInactiveValidatorInterval < minDuration {
		return fmt.Errorf("%w on HideInactiveValidatorInterval, provided %d, min expected %d",
			heartbeat.ErrInvalidTimeDuration, args.HideInactiveValidatorInterval, minDuration)
	}
	if check.IfNil(args.PeerTypeProvider) {
		return heartbeat.ErrNilPeerTypeProvider
	}

	return nil
}

// GetHeartbeats returns the heartbeat status
func (monitor *heartbeatV2Monitor) GetHeartbeats() []data.PubKeyHeartbeat {
	heartbeatMessagesMap := monitor.processRAWDataFromCache()

	heartbeatsV2 := make([]data.PubKeyHeartbeat, 0)
	for _, heartbeatMessagesInstance := range heartbeatMessagesMap {
		message, err := heartbeatMessagesInstance.getHeartbeat()
		if err != nil {
			log.Warn("heartbeatV2Monitor.GetHeartbeats", "error", err)
			continue
		}

		message.NumInstances = heartbeatMessagesInstance.numActivePids
		heartbeatsV2 = append(heartbeatsV2, *message)

		monitor.removeInactive(heartbeatMessagesInstance.getInactivePids())
	}

	sort.Slice(heartbeatsV2, func(i, j int) bool {
		return strings.Compare(heartbeatsV2[i].PublicKey, heartbeatsV2[j].PublicKey) < 0
	})

	return heartbeatsV2
}

func (monitor *heartbeatV2Monitor) removeInactive(pids []core.PeerID) {
	for _, pid := range pids {
		monitor.cache.Remove([]byte(pid))
	}
}

func (monitor *heartbeatV2Monitor) processRAWDataFromCache() map[string]*heartbeatMessages {
	pids := monitor.cache.Keys()

	heartbeatsV2 := make(map[string]*heartbeatMessages)
	for idx := 0; idx < len(pids); idx++ {
		pid := pids[idx]
		hb, ok := monitor.cache.Get(pid)
		if !ok {
			continue
		}

		peerId := core.PeerID(pid)
		heartbeatData, err := monitor.parseMessage(peerId, hb)
		if err != nil {
			monitor.cache.Remove(pid)
			log.Trace("could not parse message for pid, removed message", "pid", peerId.Pretty(), "error", err.Error())
			continue
		}

		heartbeatMessagesInstance, found := heartbeatsV2[heartbeatData.PublicKey]
		if !found {
			heartbeatMessagesInstance = newHeartbeatMessages()
			heartbeatsV2[heartbeatData.PublicKey] = heartbeatMessagesInstance
		}

		heartbeatMessagesInstance.addMessage(peerId, heartbeatData)
	}

	return heartbeatsV2
}

func (monitor *heartbeatV2Monitor) parseMessage(pid core.PeerID, message interface{}) (*data.PubKeyHeartbeat, error) {
	heartbeatV2, ok := message.(*heartbeat.HeartbeatV2)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	payload := &heartbeat.Payload{}
	err := monitor.marshaller.Unmarshal(payload, heartbeatV2.Payload)
	if err != nil {
		return nil, err
	}

	crtTime := time.Now()
	messageTime := time.Unix(payload.Timestamp, 0)
	messageAge := monitor.getMessageAge(crtTime, messageTime)
	computedShardID, stringType := monitor.computePeerTypeAndShardID(heartbeatV2)
	if monitor.shouldSkipMessage(messageAge) {
		return nil, fmt.Errorf("%w, messageAge %v", heartbeat.ErrShouldSkipValidator, messageAge)
	}

	encodedPubKey, err := monitor.pubKeyConverter.Encode(heartbeatV2.GetPubkey())
	if err != nil {
		return nil, err
	}

	pubKeyHeartbeat := &data.PubKeyHeartbeat{
		PublicKey:            encodedPubKey,
		TimeStamp:            messageTime,
		IsActive:             monitor.isActive(messageAge),
		ReceivedShardID:      monitor.shardId,
		ComputedShardID:      computedShardID,
		VersionNumber:        heartbeatV2.GetVersionNumber(),
		NodeDisplayName:      heartbeatV2.GetNodeDisplayName(),
		Identity:             heartbeatV2.GetIdentity(),
		PeerType:             stringType,
		Nonce:                heartbeatV2.GetNonce(),
		PeerSubType:          heartbeatV2.GetPeerSubType(),
		PidString:            pid.Pretty(),
		NumTrieNodesReceived: heartbeatV2.NumTrieNodesSynced,
	}

	return pubKeyHeartbeat, nil
}

func (monitor *heartbeatV2Monitor) computePeerTypeAndShardID(hbMessage *heartbeat.HeartbeatV2) (uint32, string) {
	peerType, shardID, err := monitor.peerTypeProvider.ComputeForPubKey(hbMessage.Pubkey)
	if err != nil {
		log.Warn("heartbeatV2Monitor: computePeerType", "error", err)
		return monitor.shardId, string(common.ObserverList)
	}
	if peerType == common.ObserverList {
		return monitor.shardId, string(common.ObserverList)
	}

	return shardID, string(peerType)
}

func (monitor *heartbeatV2Monitor) getMessageAge(crtTime time.Time, messageTime time.Time) time.Duration {
	msgAge := crtTime.Sub(messageTime)
	return monitor.maxDuration(0, msgAge)
}

func (monitor *heartbeatV2Monitor) maxDuration(first, second time.Duration) time.Duration {
	if first > second {
		return first
	}

	return second
}

func (monitor *heartbeatV2Monitor) isActive(messageAge time.Duration) bool {
	return messageAge <= monitor.maxDurationPeerUnresponsive
}

func (monitor *heartbeatV2Monitor) shouldSkipMessage(messageAge time.Duration) bool {
	isActive := monitor.isActive(messageAge)
	if !isActive {
		return messageAge > monitor.hideInactiveValidatorInterval
	}

	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (monitor *heartbeatV2Monitor) IsInterfaceNil() bool {
	return monitor == nil
}
