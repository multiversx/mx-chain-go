package spos

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
)

// BlockDelayMetrics keeps track of block delays and computes average delays
type ConsensusMetrics struct {
	blockHeaderReceivedOrSentDelay uint64
	blockBodyReceivedOrSentDelay   uint64
	blockHash                      []byte

	blockReceivedDelaySum uint64
	blockReceivedCount    uint64
	blockSignedDelaySum   uint64
	blockSignedCount      uint64

	appStatusHandler core.AppStatusHandler
	mut              sync.RWMutex
}

// NewConsensusMetrics creates a new instance of ConsensusMetrics
func NewConsensusMetrics(appStatusHandler core.AppStatusHandler) *ConsensusMetrics {
	if appStatusHandler == nil {
		return nil
	}
	return &ConsensusMetrics{appStatusHandler: appStatusHandler}
}

// resetInstanceValues resets the instance values for the next round.
func (cm *ConsensusMetrics) resetInstanceValues() {
	cm.blockHeaderReceivedOrSentDelay = uint64(0)
	cm.blockBodyReceivedOrSentDelay = uint64(0)
	cm.blockHash = nil
}

// ResetAverages resets the average calculations. It should be called on epoch change.
func (cm *ConsensusMetrics) ResetAverages() {
	cm.mut.Lock()
	cm.blockReceivedDelaySum = 0
	cm.blockReceivedCount = 0
	cm.blockSignedDelaySum = 0
	cm.blockSignedCount = 0
	cm.mut.Unlock()
}

// SetBlockReceived sets the block received delay and updates the mettics if both header and body have been received/sent.
func (cm *ConsensusMetrics) SetBlockHeaderReceived(blockHash []byte, delayFromRoundStart uint64) {
	cm.mut.Lock()

	if cm.blockHash == nil {
		// first block information received
		cm.blockHash = blockHash
		cm.blockHeaderReceivedOrSentDelay = delayFromRoundStart
	} else if string(cm.blockHash) == string(blockHash) && cm.blockHeaderReceivedOrSentDelay == 0 {
		// all block information received, update metrics
		cm.blockHeaderReceivedOrSentDelay = delayFromRoundStart
		metricValue := max(cm.blockHeaderReceivedOrSentDelay, cm.blockBodyReceivedOrSentDelay)
		defer cm.appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, metricValue)
		cm.updateAverage(common.MetricReceivedProposedBlockBody, delayFromRoundStart)
	}

	cm.mut.Unlock()
}

// SetBlockReceived sets the block received delay and updates the metrics if both header and body have been received/sent.
func (cm *ConsensusMetrics) SetBlockBodyReceived(blockHash []byte, delayFromRoundStart uint64) {
	cm.mut.Lock()

	if cm.blockHash == nil {
		// first block information received
		cm.blockHash = blockHash
		cm.blockBodyReceivedOrSentDelay = delayFromRoundStart
	} else if string(cm.blockHash) == string(blockHash) && cm.blockBodyReceivedOrSentDelay == 0 {
		// all block information received, update metrics
		cm.blockBodyReceivedOrSentDelay = delayFromRoundStart
		metricValue := max(cm.blockHeaderReceivedOrSentDelay, cm.blockBodyReceivedOrSentDelay)
		defer cm.appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, metricValue)
		cm.updateAverage(common.MetricReceivedProposedBlockBody, delayFromRoundStart)
	}

	cm.mut.Unlock()
}

// SetProofReceived sets the proof received delay and updates the metrics
func (cm *ConsensusMetrics) SetProofReceived(blockHash []byte, delayProofFromRoundStart uint64) error {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	delayBlockFromRoundStart := max(cm.blockHeaderReceivedOrSentDelay, cm.blockBodyReceivedOrSentDelay)

	if cm.blockHash == nil || delayBlockFromRoundStart == 0 {
		return ErrHeaderProofNotExpected
	}

	if string(cm.blockHash) == string(blockHash) {

		if cm.blockHeaderReceivedOrSentDelay == 0 || cm.blockBodyReceivedOrSentDelay == 0 {
			// update metrics for incomplete block received
			defer cm.appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, delayBlockFromRoundStart)
			cm.updateAverage(common.MetricReceivedProposedBlockBody, delayBlockFromRoundStart)
			log.Debug("Proof received: Incomplete Block header and body received")
		}

		if delayBlockFromRoundStart > delayProofFromRoundStart {
			// proof received before block header and body
			return ErrHeaderProofNotExpected
		}

		metricsTime := delayProofFromRoundStart - delayBlockFromRoundStart
		defer cm.appStatusHandler.SetUInt64Value(common.MetricReceivedProof, metricsTime)
		cm.updateAverage(common.MetricReceivedProof, metricsTime)

		// reset instance values for the next round
		cm.resetInstanceValues()
	}

	return nil
}

func (cm *ConsensusMetrics) updateAverage(metric string, metricsTime uint64) {
	switch metric {
	case common.MetricReceivedProposedBlockBody:
		cm.blockReceivedDelaySum += metricsTime
		cm.blockReceivedCount++
		averageReceivedBlockDelay := cm.blockReceivedDelaySum / cm.blockReceivedCount
		log.Debug("Block header and body received", "averageBlockReceivedDelay", averageReceivedBlockDelay)
	case common.MetricReceivedProof:
		cm.blockSignedDelaySum += metricsTime
		cm.blockSignedCount++
		averageProofDelay := cm.blockSignedDelaySum / cm.blockSignedCount
		log.Debug("Proof received", "averageProofDelay", averageProofDelay)
	}
}

func (cm *ConsensusMetrics) IsInterfaceNil() bool {
	return cm == nil
}
