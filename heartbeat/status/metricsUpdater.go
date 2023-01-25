package status

import (
	"context"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
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
}

type metricsUpdater struct {
	peerAuthenticationCacher            storage.Cacher
	heartbeatMonitor                    HeartbeatMonitor
	heartbeatSenderInfoProvider         HeartbeatSenderInfoProvider
	appStatusHandler                    core.AppStatusHandler
	timeBetweenConnectionsMetricsUpdate time.Duration
	cancelFunc                          func()
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
	}

	args.PeerAuthenticationCacher.RegisterHandler(updater.onAddedPeerAuthenticationMessage, "metricsUpdater")

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
		return heartbeat.ErrNilHeartbeatMonitor
	}
	if check.IfNil(args.HeartbeatSenderInfoProvider) {
		return heartbeat.ErrNilHeartbeatSenderInfoProvider
	}
	if args.TimeBetweenConnectionsMetricsUpdate < minDuration {
		return fmt.Errorf("%w on TimeBetweenConnectionsMetricsUpdate, provided %d, min expected %d",
			heartbeat.ErrInvalidTimeDuration, args.TimeBetweenConnectionsMetricsUpdate, minDuration)
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
}

func (updater *metricsUpdater) updateSenderMetrics() {
	result, subType, err := updater.heartbeatSenderInfoProvider.GetCurrentNodeType()
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

func (updater *metricsUpdater) onAddedPeerAuthenticationMessage(_ []byte, _ interface{}) {
	updater.appStatusHandler.SetUInt64Value(common.MetricLiveValidatorNodes, uint64(updater.peerAuthenticationCacher.Len()))
}

// IsInterfaceNil returns true if there is no value under the interface
func (updater *metricsUpdater) IsInterfaceNil() bool {
	return updater == nil
}
