package status

import (
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/storage"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const minDuration = time.Second

var log = logger.GetOrCreate("heartbeat/status")

// ArgsMetricsUpdater represents the arguments for the metricsUpdater constructor
type ArgsMetricsUpdater struct {
	PeerAuthenticationCacher            storage.Cacher
	HeartbeatMonitor                    HeartbeatMonitor
	HeartbeatSenderInfoProvider         HeartbeatSenderInfoProvider
	AppStatusHandler                    core.AppStatusHandler
	TimeBetweenConnectionsMetricsUpdate time.Duration
	EpochNotifier                       vmcommon.EpochNotifier
	HeartbeatV1DisableEpoch             uint32
}

type metricsUpdater struct {
	peerAuthenticationCacher            storage.Cacher
	heartbeatMonitor                    HeartbeatMonitor
	heartbeatSenderInfoProvider         HeartbeatSenderInfoProvider
	appStatusHandler                    core.AppStatusHandler
	timeBetweenConnectionsMetricsUpdate time.Duration
	cancelFunc                          func()
	flagHeartbeatV1DisableEpoch         atomic.Flag
	heartbeatV1DisableEpoch             uint32
}

// NewMetricsUpdater creates a new instance of type metricsUpdater
func NewMetricsUpdater(args ArgsMetricsUpdater) (*metricsUpdater, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	updater := &metricsUpdater{
		peerAuthenticationCacher:            args.PeerAuthenticationCacher,
		heartbeatMonitor:                    args.HeartbeatMonitor,
		heartbeatSenderInfoProvider:         args.HeartbeatSenderInfoProvider,
		appStatusHandler:                    args.AppStatusHandler,
		timeBetweenConnectionsMetricsUpdate: args.TimeBetweenConnectionsMetricsUpdate,
		heartbeatV1DisableEpoch:             args.HeartbeatV1DisableEpoch,
	}

	args.EpochNotifier.RegisterNotifyHandler(updater)

	var ctx context.Context
	ctx, updater.cancelFunc = context.WithCancel(context.Background())
	go updater.processMetricsUpdate(ctx)

	return updater, nil
}

func checkArgs(args ArgsMetricsUpdater) error {
	if check.IfNil(args.PeerAuthenticationCacher) {
		return heartbeat.ErrNilCacher
	}
	if check.IfNil(args.AppStatusHandler) {
		return heartbeat.ErrNilAppStatusHandler
	}
	if check.IfNil(args.HeartbeatMonitor) {
		return heartbeat.ErrNilMonitor
	}
	if check.IfNil(args.HeartbeatSenderInfoProvider) {
		return heartbeat.ErrNilSenderInfoProvider
	}
	if args.TimeBetweenConnectionsMetricsUpdate < minDuration {
		return fmt.Errorf("%w on TimeBetweenConnectionsMetricsUpdate, provided %d, min expected %d",
			heartbeat.ErrInvalidTimeDuration, args.TimeBetweenConnectionsMetricsUpdate, minDuration)
	}
	if check.IfNil(args.EpochNotifier) {
		return heartbeat.ErrNilEpochNotifier
	}

	return nil
}

func (updater *metricsUpdater) processMetricsUpdate(ctx context.Context) {
	timer := time.NewTimer(updater.timeBetweenConnectionsMetricsUpdate)
	defer timer.Stop()

	for {
		timer.Reset(updater.timeBetweenConnectionsMetricsUpdate)

		select {
		case <-timer.C:
			updater.updateMetrics()
		case <-ctx.Done():
			log.Debug("closing heartbeat v2 metricsUpdater go routine")
			return
		}
	}
}

func (updater *metricsUpdater) updateMetrics() {
	heartbeatSubsystemV1IsActive := !updater.flagHeartbeatV1DisableEpoch.IsSet()
	if heartbeatSubsystemV1IsActive {
		return
	}

	updater.updateConnectionsMetrics()
	updater.updateSenderMetrics()
}

func (updater *metricsUpdater) updateConnectionsMetrics() {
	heartbeats := updater.heartbeatMonitor.GetHeartbeats()

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

	updater.appStatusHandler.SetUInt64Value(common.MetricNumIntraShardValidatorNodes, uint64(counterActiveValidators))
	updater.appStatusHandler.SetUInt64Value(common.MetricConnectedNodes, uint64(counterConnectedNodes))
	updater.appStatusHandler.SetUInt64Value(common.MetricLiveValidatorNodes, uint64(updater.peerAuthenticationCacher.Len()))
}

func (updater *metricsUpdater) updateSenderMetrics() {
	result, subType, err := updater.heartbeatSenderInfoProvider.GetSenderInfo()
	if err != nil {
		log.Warn("error while updating metrics in heartbeat v2 metricsUpdater", "error", err)
		return
	}

	nodeType := ""
	if result == string(common.ObserverList) {
		nodeType = string(core.NodeTypeObserver)
	} else {
		nodeType = string(core.NodeTypeValidator)
	}

	updater.appStatusHandler.SetStringValue(common.MetricNodeType, nodeType)
	updater.appStatusHandler.SetStringValue(common.MetricPeerType, result)
	updater.appStatusHandler.SetStringValue(common.MetricPeerSubType, subType.String())
}

func isValidator(peerType string) bool {
	return peerType == string(common.EligibleList) || peerType == string(common.WaitingList)
}

// Close closes the internal go routine
func (updater *metricsUpdater) Close() error {
	updater.cancelFunc()

	return nil
}

// EpochConfirmed is called whenever an epoch is confirmed
func (updater *metricsUpdater) EpochConfirmed(epoch uint32, _ uint64) {
	updater.flagHeartbeatV1DisableEpoch.SetValue(epoch >= updater.heartbeatV1DisableEpoch)
	log.Debug("heartbeat v1 subsystem", "enabled", !updater.flagHeartbeatV1DisableEpoch.IsSet())
}

// IsInterfaceNil returns true if there is no value under the interface
func (updater *metricsUpdater) IsInterfaceNil() bool {
	return updater == nil
}
