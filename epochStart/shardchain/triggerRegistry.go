package shardchain

import (
	"fmt"
	"strconv"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/hardfork"
	"github.com/multiversx/mx-chain-go/epochStart"
)

// maxLoadStateFallbackSteps bounds the walk-back when the saved key's round is inside a
// hardfork exclusion interval. Should be well above the longest configured interval.
const maxLoadStateFallbackSteps = 2_000_000

// LoadState loads into trigger the saved state.
// If the key decodes to a round inside a configured hardfork exclusion interval, the state at that
// round is considered poisoned: the trigger walks back round-by-round (skipping any other excluded
// intervals) and loads the first saved state whose key falls outside every exclusion.
func (t *trigger) LoadState(key []byte) error {
	resolvedKey, err := t.resolveLoadStateKey(key)
	if err != nil {
		return err
	}

	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), resolvedKey...)
	log.Debug("getting start of epoch trigger state", "key", trigInternalKey)

	triggerData, err := t.triggerStorage.Get(trigInternalKey)
	if err != nil {
		return err
	}

	state, err := epochStart.UnmarshalShardTrigger(t.marshaller, triggerData)
	if err != nil {
		return err
	}

	t.mutTrigger.Lock()
	t.triggerStateKey = resolvedKey
	t.epoch = state.GetEpoch()
	t.metaEpoch = state.GetMetaEpoch()
	t.currentRoundIndex = state.GetCurrentRoundIndex()
	t.epochStartRound = state.GetEpochStartRound()
	t.epochMetaBlockHash = state.GetEpochMetaBlockHash()
	t.isEpochStart = state.GetIsEpochStart()
	t.newEpochHdrReceived = state.GetNewEpochHeaderReceived()
	t.epochFinalityAttestingRound = state.GetEpochFinalityAttestingRound()
	t.epochStartShardHeader = state.GetEpochStartHeaderHandler()
	t.mutTrigger.Unlock()

	log.Debug("loaded start of epoch trigger state",
		"epoch", t.epoch,
		"metaEpoch", t.metaEpoch,
		"currentRoundIndex", t.currentRoundIndex,
		"epochStartRound", t.epochStartRound,
		"epochMetaBlockHash", t.epochMetaBlockHash,
		"isEpochStart", t.isEpochStart,
		"newEpochHdrReceived", t.newEpochHdrReceived,
		"epochFinalityAttestingRound", t.epochFinalityAttestingRound,
		"epochStartShardHeader", t.epochStartShardHeader)

	return nil
}

// resolveLoadStateKey returns `key` unchanged when it does not decode as a pure round (e.g. the
// initial epoch-prefixed key built in NewEpochStartTrigger) or when the decoded round sits outside
// every configured exclusion interval. Otherwise it walks backward to the nearest round that has a
// state saved in triggerStorage and is not itself excluded.
func (t *trigger) resolveLoadStateKey(key []byte) ([]byte, error) {
	round, err := strconv.ParseUint(string(key), 10, 64)
	if err != nil {
		return key, nil
	}

	iv := hardfork.IntervalForRoundAnyShard(round)
	if iv == nil {
		return key, nil
	}

	log.Warn("epoch start trigger load state: key is inside hardfork exclusion interval, walking back",
		"round", round,
		"excluded low", iv.Low,
		"excluded high", iv.High)

	candidate := iv.Low
	for step := 0; step < maxLoadStateFallbackSteps; step++ {
		if candidate == 0 {
			break
		}
		candidate--

		if skip := hardfork.IntervalForRoundAnyShard(candidate); skip != nil {
			// jump to the round just below the lower bound of the interval we just entered
			if skip.Low == 0 {
				break
			}
			candidate = skip.Low
			continue
		}

		candidateKey := []byte(fmt.Sprint(candidate))
		trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), candidateKey...)
		if _, getErr := t.triggerStorage.Get(trigInternalKey); getErr == nil {
			log.Info("epoch start trigger load state: using earlier key outside exclusion",
				"original round", round,
				"loaded round", candidate)
			return candidateKey, nil
		}
	}

	return nil, fmt.Errorf("no epoch start trigger state found before excluded round %d", round)
}

// saveState saves the trigger state. Needs to be called under mutex
func (t *trigger) saveState(key []byte) error {
	registry := epochStart.CreateShardRegistryHandler(t.epochStartShardHeader)
	_ = registry.SetMetaEpoch(t.metaEpoch)
	_ = registry.SetEpoch(t.epoch)
	_ = registry.SetCurrentRoundIndex(t.currentRoundIndex)
	_ = registry.SetEpochStartRound(t.epochStartRound)
	_ = registry.SetEpochMetaBlockHash(t.epochMetaBlockHash)
	_ = registry.SetIsEpochStart(t.isEpochStart)
	_ = registry.SetNewEpochHeaderReceived(t.newEpochHdrReceived)
	_ = registry.SetEpochFinalityAttestingRound(t.epochFinalityAttestingRound)
	_ = registry.SetEpochStartHeaderHandler(t.epochStartShardHeader)

	marshalledRegistry, err := t.marshaller.Marshal(registry)
	if err != nil {
		return err
	}

	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)
	log.Debug("saving start of epoch trigger state", "key", trigInternalKey)

	return t.triggerStorage.Put(trigInternalKey, marshalledRegistry)
}
