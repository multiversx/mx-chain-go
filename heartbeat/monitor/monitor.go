package monitor

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("heartbeat/monitor")

const minDuration = time.Second

// ArgHeartbeatV2Monitor holds the arguments needed to create a new instance of heartbeatV2Monitor
type ArgHeartbeatV2Monitor struct {
	Cache                         storage.Cacher
	PubKeyConverter               core.PubkeyConverter
	Marshaller                    marshal.Marshalizer
	PeerShardMapper               process.PeerShardMapper
	MaxDurationPeerUnresponsive   time.Duration
	HideInactiveValidatorInterval time.Duration
	ShardId                       uint32
	PeerTypeProvider              heartbeat.PeerTypeProviderHandler
	AppStatusHandler              core.AppStatusHandler
}

type heartbeatV2Monitor struct {
	cache                         storage.Cacher
	pubKeyConverter               core.PubkeyConverter
	marshaller                    marshal.Marshalizer
	peerShardMapper               process.PeerShardMapper
	maxDurationPeerUnresponsive   time.Duration
	hideInactiveValidatorInterval time.Duration
	shardId                       uint32
	peerTypeProvider              heartbeat.PeerTypeProviderHandler
	appStatusHandler              core.AppStatusHandler
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
		peerShardMapper:               args.PeerShardMapper,
		maxDurationPeerUnresponsive:   args.MaxDurationPeerUnresponsive,
		hideInactiveValidatorInterval: args.HideInactiveValidatorInterval,
		shardId:                       args.ShardId,
		peerTypeProvider:              args.PeerTypeProvider,
		appStatusHandler:              args.AppStatusHandler,
	}

	hbv2Monitor.cache.RegisterHandler(hbv2Monitor.updateMetrics, core.UniqueIdentifier())

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
	if check.IfNil(args.PeerShardMapper) {
		return heartbeat.ErrNilPeerShardMapper
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
	if check.IfNil(args.AppStatusHandler) {
		return heartbeat.ErrNilAppStatusHandler
	}

	return nil
}

// GetHeartbeats returns the heartbeat status
func (monitor *heartbeatV2Monitor) GetHeartbeats() []data.PubKeyHeartbeat {
	numInstances := make(map[string]uint64)

	pids := monitor.cache.Keys()

	heartbeatsV2 := make([]data.PubKeyHeartbeat, 0)
	for idx := 0; idx < len(pids); idx++ {
		pid := pids[idx]
		hb, ok := monitor.cache.Get(pid)
		if !ok {
			continue
		}

		peerId := core.PeerID(pid)
		heartbeatData, err := monitor.parseMessage(peerId, hb, numInstances)
		if err != nil {
			monitor.cache.Remove(pid)
			log.Trace("could not parse message for pid, removed message", "pid", peerId.Pretty(), "error", err.Error())
			continue
		}

		heartbeatsV2 = append(heartbeatsV2, heartbeatData)
	}

	for idx := range heartbeatsV2 {
		hbData := &heartbeatsV2[idx]
		pk := hbData.PublicKey
		hbData.NumInstances = numInstances[pk]
	}

	sort.Slice(heartbeatsV2, func(i, j int) bool {
		return strings.Compare(heartbeatsV2[i].PublicKey, heartbeatsV2[j].PublicKey) < 0
	})

	return heartbeatsV2
}

func (monitor *heartbeatV2Monitor) parseMessage(pid core.PeerID, message interface{}, numInstances map[string]uint64) (data.PubKeyHeartbeat, error) {
	pubKeyHeartbeat := data.PubKeyHeartbeat{}

	heartbeatV2, ok := message.(*heartbeat.HeartbeatV2)
	if !ok {
		return pubKeyHeartbeat, process.ErrWrongTypeAssertion
	}

	payload := heartbeat.Payload{}
	err := monitor.marshaller.Unmarshal(&payload, heartbeatV2.Payload)
	if err != nil {
		return pubKeyHeartbeat, err
	}

	peerInfo := monitor.peerShardMapper.GetPeerInfo(pid)

	crtTime := time.Now()
	messageAge := monitor.getMessageAge(crtTime, payload.Timestamp)
	stringType := monitor.computePeerType(peerInfo.PkBytes)
	if monitor.shouldSkipMessage(messageAge, stringType) {
		return pubKeyHeartbeat, heartbeat.ErrShouldSkipValidator
	}

	pk := monitor.pubKeyConverter.Encode(peerInfo.PkBytes)
	numInstances[pk]++

	pubKeyHeartbeat = data.PubKeyHeartbeat{
		PublicKey:       pk,
		TimeStamp:       crtTime,
		IsActive:        monitor.isActive(messageAge),
		ReceivedShardID: monitor.shardId,
		ComputedShardID: peerInfo.ShardID,
		VersionNumber:   heartbeatV2.GetVersionNumber(),
		NodeDisplayName: heartbeatV2.GetNodeDisplayName(),
		Identity:        heartbeatV2.GetIdentity(),
		PeerType:        stringType,
		Nonce:           heartbeatV2.GetNonce(),
		PeerSubType:     heartbeatV2.GetPeerSubType(),
		PidString:       pid.Pretty(),
	}

	return pubKeyHeartbeat, nil
}

func (monitor *heartbeatV2Monitor) computePeerType(pk []byte) string {
	peerType, _, err := monitor.peerTypeProvider.ComputeForPubKey(pk)
	if err != nil {
		log.Warn("heartbeatV2Monitor: computePeerType", "error", err)
		return string(common.ObserverList)
	}

	return string(peerType)
}

func (monitor *heartbeatV2Monitor) getMessageAge(crtTime time.Time, messageTimestamp int64) time.Duration {
	messageTime := time.Unix(messageTimestamp, 0)
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

func (monitor *heartbeatV2Monitor) shouldSkipMessage(messageAge time.Duration, peerType string) bool {
	isActive := monitor.isActive(messageAge)
	isInactiveObserver := !isActive &&
		peerType != string(common.EligibleList) &&
		peerType != string(common.WaitingList)
	if isInactiveObserver {
		return messageAge > monitor.hideInactiveValidatorInterval
	}

	return false
}

func (monitor *heartbeatV2Monitor) updateMetrics(_ []byte, _ interface{}) {
	heartbeats := monitor.GetHeartbeats()

	counterActiveValidators := 0
	counterConnectedNodes := 0
	for _, heartbeatMessage := range heartbeats {
		if heartbeatMessage.IsActive {
			counterConnectedNodes++

			if isValidator(heartbeatMessage.PeerType) {
				counterActiveValidators++
			}
		}
	}

	monitor.appStatusHandler.SetUInt64Value(common.MetricLiveValidatorNodes, uint64(counterActiveValidators))
	monitor.appStatusHandler.SetUInt64Value(common.MetricConnectedNodes, uint64(counterConnectedNodes))
}

func isValidator(peerType string) bool {
	return peerType == string(common.EligibleList) || peerType == string(common.WaitingList)
}

// IsInterfaceNil returns true if there is no value under the interface
func (monitor *heartbeatV2Monitor) IsInterfaceNil() bool {
	return monitor == nil
}
