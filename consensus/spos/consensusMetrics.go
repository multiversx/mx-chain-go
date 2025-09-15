package spos

import (
	"bytes"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
)

// ConsensusyMetrics keeps track of block delays and computes average delays
type ConsensusMetrics struct {
	blockHeaderReceivedOrSentDelay uint64
	blockBodyReceivedOrSentDelay   uint64
	blockHash                      []byte

	blockReceivedDelaySum uint64
	blockReceivedCount    uint64
	blockSignedDelaySum   uint64
	blockSignedCount      uint64

	isBlockBodyAlreadyReceived bool
	isProofAlreadyReceived     bool
	appStatusHandler           core.AppStatusHandler
	mut                        sync.RWMutex
}

// NewConsensusMetrics creates a new instance of ConsensusMetrics
func NewConsensusMetrics(appStatusHandler core.AppStatusHandler) *ConsensusMetrics {
	if appStatusHandler == nil {
		return nil
	}
	return &ConsensusMetrics{appStatusHandler: appStatusHandler}
}

// ResetInstanceValues resets the instance values for the next round.
func (cm *ConsensusMetrics) ResetInstanceValues() {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	cm.blockHeaderReceivedOrSentDelay = uint64(0)
	cm.blockBodyReceivedOrSentDelay = uint64(0)

	cm.blockHash = nil
	cm.isBlockBodyAlreadyReceived = false
	cm.isProofAlreadyReceived = false
}

// ResetAverages resets the average calculations. It should be called on epoch change.
func (cm *ConsensusMetrics) ResetAverages() {
	cm.mut.Lock()

	avgBlockReceivedDelay := avg(cm.blockReceivedDelaySum, cm.blockReceivedCount)
	avgBlockSignedDelay := avg(cm.blockSignedDelaySum, cm.blockSignedCount)
	log.Debug("Resetting consensus metrics averages on epoch change. Values before reset", "avgBlockReceivedDelay", avgBlockReceivedDelay, "avgBlockSignedDelaySum", avgBlockSignedDelay)

	cm.blockReceivedDelaySum = 0
	cm.blockReceivedCount = 0
	cm.blockSignedDelaySum = 0
	cm.blockSignedCount = 0
	cm.mut.Unlock()
}

// SetBlockHeaderReceived sets the block header received delay and updates the metrics if both header and body have been received/sent.
func (cm *ConsensusMetrics) SetBlockHeaderReceived(blockHash []byte, delayFromRoundStart uint64) {
	cm.mut.Lock()

	if cm.blockHash == nil && !cm.isBlockBodyAlreadyReceived {
		// first block information received
		cm.blockHash = blockHash
		cm.blockHeaderReceivedOrSentDelay = delayFromRoundStart
	} else if (bytes.Equal(cm.blockHash, blockHash) || cm.blockHash == nil && cm.isBlockBodyAlreadyReceived) && cm.blockHeaderReceivedOrSentDelay == 0 {
		// all block information received, update metrics
		cm.blockHeaderReceivedOrSentDelay = delayFromRoundStart
		// if block body comes first, it may not have a hash yet
		cm.blockHash = blockHash

		metricValue := max(cm.blockHeaderReceivedOrSentDelay, cm.blockBodyReceivedOrSentDelay)
		defer cm.appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, metricValue)
		cm.updateAverages(common.MetricReceivedProposedBlockBody, delayFromRoundStart)
	}

	cm.mut.Unlock()
}

// SetBlockBodyReceived sets the block body received delay and updates the metrics if both header and body have been received/sent.
func (cm *ConsensusMetrics) SetBlockBodyReceived(blockHash []byte, delayFromRoundStart uint64) {
	cm.mut.Lock()

	if cm.blockHash == nil {
		// first block information received
		cm.blockHash = blockHash
		cm.blockBodyReceivedOrSentDelay = delayFromRoundStart
	} else if bytes.Equal(cm.blockHash, blockHash) && cm.blockBodyReceivedOrSentDelay == 0 {
		// all block information received, update metrics
		cm.blockBodyReceivedOrSentDelay = delayFromRoundStart
		metricValue := max(cm.blockHeaderReceivedOrSentDelay, cm.blockBodyReceivedOrSentDelay)
		defer cm.appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, metricValue)
		cm.updateAverages(common.MetricReceivedProposedBlockBody, delayFromRoundStart)
	}

	cm.isBlockBodyAlreadyReceived = true
	cm.mut.Unlock()
}

// SetSignaturesReceived sets the proof received delay and updates the metrics
func (cm *ConsensusMetrics) SetSignaturesReceived(blockHash []byte, delayProofFromRoundStart uint64) error {
	cm.mut.Lock()
	defer cm.mut.Unlock()

	delayBlockFromRoundStart := max(cm.blockHeaderReceivedOrSentDelay, cm.blockBodyReceivedOrSentDelay)

	if cm.blockHash == nil || delayBlockFromRoundStart == 0 {
		return ErrHeaderProofNotExpected
	}

	if bytes.Equal(cm.blockHash, blockHash) {

		if cm.blockHeaderReceivedOrSentDelay == 0 || cm.blockBodyReceivedOrSentDelay == 0 {
			// update metrics for incomplete block received
			defer cm.appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, delayBlockFromRoundStart)
			cm.updateAverages(common.MetricReceivedProposedBlockBody, delayBlockFromRoundStart)
			log.Debug("Signatures received: Incomplete Block header and body received")
		}

		if delayBlockFromRoundStart > delayProofFromRoundStart {
			// proof received before block header and body
			return ErrHeaderProofNotExpected
		}

		metricsTime := delayProofFromRoundStart - delayBlockFromRoundStart
		defer cm.appStatusHandler.SetUInt64Value(common.MetricReceivedSignatures, metricsTime)
		cm.updateAverages(common.MetricReceivedSignatures, metricsTime)
		cm.isProofAlreadyReceived = true
	}

	return nil
}

func (cm *ConsensusMetrics) updateAverages(metric string, metricsTime uint64) {
	switch metric {
	case common.MetricReceivedProposedBlockBody:
		cm.blockReceivedDelaySum += metricsTime
		cm.blockReceivedCount++
		averageReceivedBlockDelay := avg(cm.blockReceivedDelaySum, cm.blockReceivedCount)
		cm.appStatusHandler.SetUInt64Value(common.MetricAvgReceivedProposedBlockBody, averageReceivedBlockDelay)
		log.Debug("Computed average block header and body received delay", "currentBlockReceivedDelay", metricsTime, "averageBlockReceivedDelay", averageReceivedBlockDelay)
	case common.MetricReceivedSignatures:
		cm.blockSignedDelaySum += metricsTime
		cm.blockSignedCount++
		averageProofDelay := avg(cm.blockSignedDelaySum, cm.blockSignedCount)
		defer cm.appStatusHandler.SetUInt64Value(common.MetricAvgReceivedSignatures, averageProofDelay)
		log.Debug("Computed average signature received delay from block body received", "currentSignaturesDelay", metricsTime, "averageSignaturesDelay", averageProofDelay)
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
	result["blockHeaderReceivedOrSentDelay"] = cm.blockHeaderReceivedOrSentDelay
	result["blockBodyReceivedOrSentDelay"] = cm.blockBodyReceivedOrSentDelay
	result["blockHash"] = cm.blockHash
	result["blockReceivedDelaySum"] = cm.blockReceivedDelaySum
	result["blockReceivedCount"] = cm.blockReceivedCount
	result["blockSignedDelaySum"] = cm.blockSignedDelaySum
	result["blockSignedCount"] = cm.blockSignedCount
	result["isBlockBodyAlreadyReceived"] = cm.isBlockBodyAlreadyReceived
	result["isProofAlreadyReceived"] = cm.isProofAlreadyReceived
	return result
}

func (cm *ConsensusMetrics) IsProofForCurrentConsensusSet() bool {
	cm.mut.RLock()
	defer cm.mut.RUnlock()

	return cm.isProofAlreadyReceived
}

// IsInterfaceNil returns true if the interface is nil
func (cm *ConsensusMetrics) IsInterfaceNil() bool {
	return cm == nil
}
