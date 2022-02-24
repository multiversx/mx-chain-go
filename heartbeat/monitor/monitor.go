package monitor

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("heartbeat/monitor")

const minDurationPeerUnresponsive = time.Second

type ArgHeartbeatV2Monitor struct {
	Cache                       storage.Cacher
	PubKeyConverter             core.PubkeyConverter
	Marshaller                  marshal.Marshalizer
	PeerTypeProvider            heartbeat.PeerTypeProviderHandler
	MaxDurationPeerUnresponsive time.Duration
	ShardId                     uint32
}

type heartbeatV2Monitor struct {
	cache                       storage.Cacher
	pubKeyConverter             core.PubkeyConverter
	marshaller                  marshal.Marshalizer
	peerTypeProvider            heartbeat.PeerTypeProviderHandler
	maxDurationPeerUnresponsive time.Duration
	shardId                     uint32
}

func NewHeartbeatV2Monitor(args ArgHeartbeatV2Monitor) (*heartbeatV2Monitor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &heartbeatV2Monitor{
		cache:                       args.Cache,
		pubKeyConverter:             args.PubKeyConverter,
		marshaller:                  args.Marshaller,
		peerTypeProvider:            args.PeerTypeProvider,
		maxDurationPeerUnresponsive: args.MaxDurationPeerUnresponsive,
		shardId:                     args.ShardId,
	}, nil
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
	if check.IfNil(args.PeerTypeProvider) {
		return heartbeat.ErrNilPeerTypeProvider
	}
	if args.MaxDurationPeerUnresponsive < minDurationPeerUnresponsive {
		return fmt.Errorf("%w on MaxDurationPeerUnresponsive, provided %d, min expected %d",
			heartbeat.ErrInvalidTimeDuration, args.MaxDurationPeerUnresponsive, minDurationPeerUnresponsive)
	}

	return nil
}

func (monitor *heartbeatV2Monitor) GetHeartbeats() []data.PubKeyHeartbeat {
	publicKeys := monitor.cache.Keys()

	heartbeatsV2 := make([]data.PubKeyHeartbeat, len(publicKeys))
	for idx, pk := range publicKeys {
		hb, ok := monitor.cache.Get(pk)
		if !ok {
			log.Debug("could not get data from cache for key", "key", monitor.pubKeyConverter.Encode(pk))
			continue
		}

		heartbeatData, err := monitor.parseMessage(pk, hb)
		if err != nil {
			log.Debug("could not parse message for key", "key", monitor.pubKeyConverter.Encode(pk), "error", err.Error())
			continue
		}

		heartbeatsV2[idx] = heartbeatData
	}

	return heartbeatsV2
}

func (monitor *heartbeatV2Monitor) parseMessage(publicKey []byte, message interface{}) (data.PubKeyHeartbeat, error) {
	pubKeyHeartbeat := data.PubKeyHeartbeat{}

	heartbeatV2, ok := message.(heartbeat.HeartbeatV2)
	if !ok {
		return pubKeyHeartbeat, process.ErrWrongTypeAssertion
	}

	payload := heartbeat.Payload{}
	err := monitor.marshaller.Unmarshal(payload, heartbeatV2.Payload)
	if err != nil {
		return pubKeyHeartbeat, err
	}

	peerType, shardId, err := monitor.peerTypeProvider.ComputeForPubKey(publicKey)
	if err != nil {
		return pubKeyHeartbeat, err
	}

	crtTime := time.Now()
	pubKeyHeartbeat = data.PubKeyHeartbeat{
		PublicKey:       monitor.pubKeyConverter.Encode(publicKey),
		TimeStamp:       crtTime,
		IsActive:        monitor.isActive(crtTime, payload.Timestamp),
		ReceivedShardID: monitor.shardId,
		ComputedShardID: shardId,
		VersionNumber:   heartbeatV2.GetVersionNumber(),
		NodeDisplayName: heartbeatV2.GetNodeDisplayName(),
		Identity:        heartbeatV2.GetIdentity(),
		PeerType:        string(peerType),
		Nonce:           heartbeatV2.GetNonce(),
		NumInstances:    0,
		PeerSubType:     heartbeatV2.GetPeerSubType(),
		PidString:       "",
	}

	return pubKeyHeartbeat, nil
}

func (monitor *heartbeatV2Monitor) isActive(crtTime time.Time, messageTimestamp int64) bool {
	messageTime := time.Unix(messageTimestamp, 0)
	msgAge := crtTime.Sub(messageTime)

	if msgAge < 0 {
		return false
	}

	return msgAge <= monitor.maxDurationPeerUnresponsive
}

func (monitor *heartbeatV2Monitor) IsInterfaceNil() bool {
	return monitor == nil
}
