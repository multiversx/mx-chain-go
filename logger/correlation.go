package logger

import (
	atomic "github.com/ElrondNetwork/elrond-go/core/atomic"
)

var globalCorrelation logCorrelation
var globalCorrelationFlag atomic.Flag

// logCorrelation holds log correlation elements
type logCorrelation struct {
	shard    atomic.String
	epoch    atomic.Uint32
	round    atomic.Int64
	subRound atomic.String
}

func (correlation *logCorrelation) setShard(shardID string) {
	correlation.shard.Set(shardID)
}

func (correlation *logCorrelation) getShard() string {
	return correlation.shard.Get()
}

func (correlation *logCorrelation) setEpoch(epoch uint32) {
	correlation.epoch.Set(epoch)
}

func (correlation *logCorrelation) getEpoch() uint32 {
	return correlation.epoch.Get()
}

func (correlation *logCorrelation) setRound(round int64) {
	correlation.round.Set(round)
}

func (correlation *logCorrelation) getRound() int64 {
	return correlation.round.Get()
}

func (correlation *logCorrelation) setSubRound(subRound string) {
	correlation.subRound.Set(subRound)
}

func (correlation *logCorrelation) getSubRound() string {
	return correlation.subRound.Get()
}

// ToggleCorrelation enables or disables correlation elements for log lines
func ToggleCorrelation(enable bool) {
	globalCorrelationFlag.Toggle(enable)
}

// IsEnabledCorrelation returns whether correlation elements are enabled
func IsEnabledCorrelation() bool {
	return globalCorrelationFlag.IsSet()
}

// SetCorrelationShard sets the current shard ID as a log correlation element
func SetCorrelationShard(shardID string) {
	globalCorrelation.setShard(shardID)
}

// SetCorrelationEpoch sets the current epoch as a log correlation element
func SetCorrelationEpoch(epoch uint32) {
	globalCorrelation.setEpoch(epoch)
}

// SetCorrelationRound sets the current round as a log correlation element
func SetCorrelationRound(round int64) {
	globalCorrelation.setRound(round)
}

// SetCorrelationSubround sets the current sub-round as a log correlation element
func SetCorrelationSubround(subRound string) {
	globalCorrelation.setSubRound(subRound)
}

// GetCorrelation gets global correlation elements
func GetCorrelation() LogCorrelationMessage {
	return LogCorrelationMessage{
		Shard:    globalCorrelation.getShard(),
		Epoch:    globalCorrelation.getEpoch(),
		Round:    globalCorrelation.getRound(),
		SubRound: globalCorrelation.getSubRound(),
	}
}
