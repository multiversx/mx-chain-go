package peerHonesty

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
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

	go instance.executeDecay(ctx, handler)

	return instance, nil
}

func (ps *peerScore) clone() *peerScore {
	psCloned := &peerScore{
		pk:     ps.pk,
		scores: make(map[string]float64),
	}

	for t, s := range ps.scores {
		psCloned.scores[t] = s
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

	ps := pph.getValidPeerScore(pk)
	ps.scores[topic] = value
}
