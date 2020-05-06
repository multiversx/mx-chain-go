package txcache

import (
	"github.com/ElrondNetwork/elrond-go/core/atomic"
)

// ChaosState -
var ChaosState *Chaos

func init() {
	ChaosState = &Chaos{}
}

// Chaos -
type Chaos struct {
	Shard     atomic.Uint32
	Round     atomic.Int64
	Epoch     atomic.Uint32
	Validator atomic.Int64
}

func (chaos *Chaos) shouldApplyOnGetTxByHash() bool {
	if chaos.Shard.Get() == 0 {
		return false
	}

	if chaos.Epoch.Get()%2 == 1 {
		return false
	}

	if chaos.Round.Get()%2 == 0 {
		return false
	}

	return true
}

func (chaos *Chaos) shouldApplyOnRemoveTxByHash() bool {
	if chaos.Shard.Get() == 0 {
		return false
	}

	return true
}

func (chaos *Chaos) isExtremelyMalicious() bool {
	return chaos.Validator.Get() == 3
}

func (chaos *Chaos) display() {
	log.Warn("Chaos state:", "shard", chaos.Shard.Get(), "epoch", chaos.Epoch.Get(), "round", chaos.Round.Get(), "validator", chaos.Validator.Get())
}
