package peerHonesty

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

func NewP2pPeerHonestyWithCustomExecuteDelayFunction(
	peerHonestyConfig config.PeerHonestyConfig,
	blackListedPkCache process.TimeCacher,
	cache storage.Cacher,
	handler func(),
) (*p2pPeerHonesty, error) {
	instance := &p2pPeerHonesty{
		decayCoefficient:       peerHonestyConfig.DecayCoefficient,
		updateIntervalForDecay: time.Duration(peerHonestyConfig.DecayUpdateIntervalInSeconds) * time.Second,
		maxScore:               peerHonestyConfig.MaxScore,
		minScore:               peerHonestyConfig.MinScore,
		badPeerThreshold:       peerHonestyConfig.BadPeerThreshold,
		unitValue:              peerHonestyConfig.UnitValue,
		cache:                  cache,
		blackListedPkCache:     blackListedPkCache,
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	instance.cancelFunc = cancelFunc

	go instance.executeDecayContinuously(ctx, handler)

	return instance, nil
}

func (ps *peerScore) clone() *peerScore {
	psCloned := newPeerScore(ps.pk)

	for t, s := range ps.scoresByTopic {
		psCloned.scoresByTopic[t] = s
	}

	return psCloned
}

func (pph *p2pPeerHonesty) Get(key string) *peerScore {
	pph.mut.RLock()
	defer pph.mut.RUnlock()

	val, _ := pph.cache.Get([]byte(key))
	if val == nil {
		return nil
	}

	ps, ok := val.(*peerScore)
	if !ok {
		return nil
	}

	return ps.clone()
}

func (pph *p2pPeerHonesty) Put(pk string, topic string, value float64) {
	pph.mut.Lock()
	defer pph.mut.Unlock()

	ps := pph.getValidPeerScoreNoLock(pk)
	ps.scoresByTopic[topic] = value
}
