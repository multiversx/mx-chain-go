package spos

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
)

// ConsensusMetrics keeps track of block delays and computes average delays
type ConsensusMetrics struct {
	blockReceivedOrSentDelay uint64

	blockReceivedDelaySum uint64
	blockReceivedCount    uint64
	proofReceivedDelaySum uint64
	proofReceivedCount    uint64

	isBlockAlreadyReceived bool
	appStatusHandler       core.AppStatusHandler
	mut                    sync.RWMutex
}

// NewConsensusMetrics creates a new instance of ConsensusMetrics
func NewConsensusMetrics(appStatusHandler core.AppStatusHandler) (*ConsensusMetrics, error) {
	if appStatusHandler == nil {
		return nil, ErrNilConsensusMetricsHandler
	}
	return &ConsensusMetrics{appStatusHandler: appStatusHandler}, nil
}

// ResetInstanceValues resets the instance values for the next round.
func (cm *ConsensusMetrics) ResetInstanceValues() {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	cm.blockReceivedOrSentDelay = uint64(0)
	cm.isBlockAlreadyReceived = false
}

// ResetAverages resets the average calculations. It should be called on epoch change.
func (cm *ConsensusMetrics) ResetAverages() {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	avgBlockReceivedDelay := avg(cm.blockReceivedDelaySum, cm.blockReceivedCount)
	avgBlockSignedDelay := avg(cm.proofReceivedDelaySum, cm.proofReceivedCount)
	log.Debug("Resetting consensus metrics averages on epoch change. Values before reset", "avgBlockReceivedDelay", avgBlockReceivedDelay, "avgBlockSignedDelaySum", avgBlockSignedDelay)

	cm.blockReceivedDelaySum = 0
	cm.blockReceivedCount = 0
	cm.proofReceivedDelaySum = 0
	cm.proofReceivedCount = 0
}

// SetBlockReceivedOrSent sets the block body received delay and updates the metrics if both header and body have been received/sent.
func (cm *ConsensusMetrics) SetBlockReceivedOrSent(delayFromRoundStart uint64) {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	cm.isBlockAlreadyReceived = true
	cm.blockReceivedOrSentDelay = delayFromRoundStart

	cm.appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, delayFromRoundStart)
	cm.updateAverages(common.MetricReceivedProposedBlockBody, delayFromRoundStart)
}

// SetProofReceived sets the proof received delay and updates the metrics
func (cm *ConsensusMetrics) SetProofReceived(delayProofFromRoundStart uint64) {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	if !cm.isBlockAlreadyReceived {
		log.Debug("Block body not received yet, cannot compute proof received delay")
		return
	}

	if cm.blockReceivedOrSentDelay > delayProofFromRoundStart {
		log.Debug("Proof received delay is smaller than block body received delay",
			"blockReceivedOrSentDelay", cm.blockReceivedOrSentDelay,
			"delayProofFromRoundStart", delayProofFromRoundStart)
		return
	}

	metricsTime := delayProofFromRoundStart - cm.blockReceivedOrSentDelay
	cm.appStatusHandler.SetUInt64Value(common.MetricReceivedProof, metricsTime)
	cm.updateAverages(common.MetricReceivedProof, metricsTime)
}

func (cm *ConsensusMetrics) updateAverages(metric string, metricsTime uint64) {
	switch metric {
	case common.MetricReceivedProposedBlockBody:
		cm.blockReceivedDelaySum += metricsTime
		cm.blockReceivedCount++
		averageReceivedBlockDelay := avg(cm.blockReceivedDelaySum, cm.blockReceivedCount)
		cm.appStatusHandler.SetUInt64Value(common.MetricAvgReceivedProposedBlockBody, averageReceivedBlockDelay)
		log.Debug("Computed average block header and body received delay", "currentBlockReceivedDelay", metricsTime, "averageBlockReceivedDelay", averageReceivedBlockDelay)
	case common.MetricReceivedProof:
		cm.proofReceivedDelaySum += metricsTime
		cm.proofReceivedCount++
		averageProofDelay := avg(cm.proofReceivedDelaySum, cm.proofReceivedCount)
		cm.appStatusHandler.SetUInt64Value(common.MetricAvgReceivedProof, averageProofDelay)
		log.Debug("Computed average signature received delay from block body received", "currentProofDelay", metricsTime, "averageProofDelay", averageProofDelay)
	}
}

func avg(sum, count uint64) uint64 {
	if count == 0 {
		return 0
	}
	return sum / count
}

func (cm *ConsensusMetrics) GetValuesForTesting() map[string]interface{} {
	cm.mut.RLock()
	defer cm.mut.RUnlock()

	result := make(map[string]interface{})
	result["blockReceivedOrSentDelay"] = cm.blockReceivedOrSentDelay
	result["blockReceivedDelaySum"] = cm.blockReceivedDelaySum
	result["blockReceivedCount"] = cm.blockReceivedCount
	result["blockSignedDelaySum"] = cm.proofReceivedDelaySum
	result["blockSignedCount"] = cm.proofReceivedCount
	result["isBlockAlreadyReceived"] = cm.isBlockAlreadyReceived

	return result
}

// IsInterfaceNil returns true if the interface is nil
func (cm *ConsensusMetrics) IsInterfaceNil() bool {
	return cm == nil
}
