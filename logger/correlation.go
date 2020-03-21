package logger

import (
	atomic "github.com/ElrondNetwork/elrond-go/core/atomic"
)

var correlation logCorrelation

type logCorrelation struct {
	enabled  bool
	epoch    atomic.Uint32
	round    atomic.Int64
	subRound atomic.String
}

func (correlation *logCorrelation) enable() {
	// No need for atomic updates (it should only be called once, in main)
	correlation.enabled = true
}

func (correlation *logCorrelation) isEnabled() bool {
	// No need for atomic reads here (no updates are ever performed)
	return correlation.enabled
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

// EnableCorrelationElements enables correlation elements (such as round number) for log lines
func EnableCorrelationElements() {
	correlation.enable()
}

// SetCorrelationEpoch sets the current round as a log correlation element
func SetCorrelationEpoch(epoch uint32) {
	correlation.setEpoch(epoch)
}

// SetCorrelationRound sets the current round as a log correlation element
func SetCorrelationRound(round int64) {
	correlation.setRound(round)
}

// SetCorrelationSubround sets the current subRound as a log correlation element
func SetCorrelationSubround(subRound string) {
	correlation.setSubRound(subRound)
}
