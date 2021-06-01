package peerHonesty

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("process/rating/peerhonesty")

const float64Size = 8
const defaultTopicSize = 20 //this is an approximate size, used for fast computing
const approximateZero = 0.00001

const minDecayCoefficient = 0.0
const maxDecayCoefficient = 1.0
const minDecayIntervalInSeconds = uint32(1)

type p2pPeerHonesty struct {
	decayCoefficient       float64
	updateIntervalForDecay time.Duration
	maxScore               float64
	minScore               float64
	badPeerThreshold       float64
	unitValue              float64
	cache                  storage.Cacher
	mut                    sync.RWMutex
	blackListedPkCache     process.TimeCacher
	cancelFunc             func()
}

// NewP2pPeerHonesty creates a new peer honesty handler able to manage a provided set of public keys withing
// the provided cache
func NewP2pPeerHonesty(
	peerHonestyConfig config.PeerHonestyConfig,
	blackListedPkCache process.TimeCacher,
	cache storage.Cacher,
) (*p2pPeerHonesty, error) {
	err := checkParams(peerHonestyConfig, blackListedPkCache, cache)
	if err != nil {
		return nil, fmt.Errorf("%w while creating an instance of p2pPeerHonesty", err)
	}

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

	go instance.executeDecayContinuously(ctx, instance.applyDecay)

	return instance, nil
}

func checkParams(
	peerHonestyConfig config.PeerHonestyConfig,
	blackListedPkCache process.TimeCacher,
	cache storage.Cacher,
) error {
	if check.IfNil(blackListedPkCache) {
		return process.ErrNilBlackListedPkCache
	}

	if check.IfNil(cache) {
		return process.ErrNilCacher
	}

	isDecayCoefficientOk := peerHonestyConfig.DecayCoefficient > minDecayCoefficient &&
		peerHonestyConfig.DecayCoefficient < maxDecayCoefficient
	if !isDecayCoefficientOk {
		return fmt.Errorf("%w, decay coefficient should be in interval (%.2f, %.2f)",
			process.ErrInvalidDecayCoefficient,
			minDecayCoefficient,
			maxDecayCoefficient,
		)
	}

	if peerHonestyConfig.DecayUpdateIntervalInSeconds < minDecayIntervalInSeconds {
		return fmt.Errorf("%w, decay interval in seconds should be greater or equal to %d",
			process.ErrInvalidDecayIntervalInSeconds,
			minDecayIntervalInSeconds,
		)
	}
	if peerHonestyConfig.MinScore > 0 {
		return fmt.Errorf("%w, MinScore value should be negative or zero",
			process.ErrInvalidMinScore,
		)
	}
	if peerHonestyConfig.MaxScore < 0 {
		return fmt.Errorf("%w, MaxScore value should be positive or zero",
			process.ErrInvalidMaxScore,
		)
	}
	if peerHonestyConfig.UnitValue < 0 {
		return fmt.Errorf("%w, UnitValue value should be positive or zero",
			process.ErrInvalidUnitValue,
		)
	}

	isBadPeerThresholdOk := peerHonestyConfig.BadPeerThreshold < 0 && peerHonestyConfig.MinScore < peerHonestyConfig.BadPeerThreshold
	if !isBadPeerThresholdOk {
		return fmt.Errorf("%w, BadPeerThreshold value should be in interval (MinScore, 0)",
			process.ErrInvalidBadPeerThreshold,
		)
	}

	return nil
}

func (pph *p2pPeerHonesty) executeDecayContinuously(ctx context.Context, handler func()) {
	for {
		select {
		case <-time.After(pph.updateIntervalForDecay):
			handler()
		case <-ctx.Done():
			log.Debug("closing p2pPeerHonesty.executeDecayContinuously go routine")
			return
		}
	}
}

func (pph *p2pPeerHonesty) applyDecay() {
	pph.mut.Lock()
	defer pph.mut.Unlock()

	keys := pph.cache.Keys()
	for _, key := range keys {
		psObj, ok := pph.cache.Get(key)
		if !ok {
			continue
		}

		ps, ok := psObj.(*peerScore)
		if !ok {
			continue
		}

		for topic, score := range ps.scoresByTopic {
			score = score * pph.decayCoefficient
			if check.IsZeroFloat64(score, approximateZero) {
				score = 0
			}

			ps.scoresByTopic[topic] = score
		}
	}
}

// ChangeScore will change the score of a public key on a provided topic
func (pph *p2pPeerHonesty) ChangeScore(pk string, topic string, units int) {
	pph.mut.Lock()
	defer pph.mut.Unlock()

	ps := pph.getValidPeerScoreNoLock(pk)

	oldValue := ps.scoresByTopic[topic]
	change := float64(units) * pph.unitValue

	if change < 0 {
		//TODO switch this to log.Trace in the future
		log.Debug("p2pPeerHonesty.ChangeScore decrease",
			"pk", core.GetTrimmedPk(hex.EncodeToString([]byte(pk))),
			"current", fmt.Sprintf("%.2f", oldValue),
			"change", fmt.Sprintf("%.2f", change),
		)
	}

	newValue := oldValue + change
	ps.scoresByTopic[topic] = newValue
	if newValue > pph.maxScore {
		ps.scoresByTopic[topic] = pph.maxScore
	}

	if newValue < pph.minScore {
		ps.scoresByTopic[topic] = pph.minScore
	}

	pph.checkBlacklistNoLock(ps)
}

func (pph *p2pPeerHonesty) getValidPeerScoreNoLock(pk string) *peerScore {
	key := []byte(pk)

	var ps *peerScore
	psObj, _ := pph.cache.Get(key)
	if psObj == nil {
		ps = pph.createDefaultPeerScore(pk)
		return ps
	}

	ps, ok := psObj.(*peerScore)
	if !ok {
		ps = pph.createDefaultPeerScore(pk)
	}

	return ps
}

func (pph *p2pPeerHonesty) createDefaultPeerScore(pk string) *peerScore {
	key := []byte(pk)

	ps := newPeerScore(pk)
	pph.cache.Put(key, ps, ps.size())

	return ps
}

func (pph *p2pPeerHonesty) checkBlacklistNoLock(ps *peerScore) {
	shouldBlacklist := false
	for _, score := range ps.scoresByTopic {
		if score < pph.badPeerThreshold {
			shouldBlacklist = true
		}
	}

	if !shouldBlacklist {
		return
	}

	if pph.blackListedPkCache.Has(ps.pk) {
		return
	}

	log.Debug("p2pPeerHonesty.checkBlacklist: added blacklisted pk",
		"pk", core.GetTrimmedPk(hex.EncodeToString([]byte(ps.pk))),
		"duration", core.PublicKeyBlacklistDuration,
	)

	err := pph.blackListedPkCache.Upsert(ps.pk, core.PublicKeyBlacklistDuration)
	if err != nil {
		log.Warn("p2pPeerHonesty.checkBlacklist",
			"pk", core.GetTrimmedPk(hex.EncodeToString([]byte(ps.pk))),
			"error", err)
	}
}

// Close closes the running go routines related to this instance
func (pph *p2pPeerHonesty) Close() error {
	pph.cancelFunc()
	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (pph *p2pPeerHonesty) IsInterfaceNil() bool {
	return pph == nil
}
