package sync

import (
	"bytes"
	"math"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
)

const (
	maxResyncRecoveryCandidates = 4
	resyncCandidateTTL          = int64(6)
	resyncCandidateCooldown     = int64(6)
	maxWatchdogBypassRounds     = int64(6)
)

type resyncRecoveryCandidate struct {
	active            bool
	generation        uint64
	parentHash        []byte
	childNonce        uint64
	childEpoch        uint32
	childEpochStart   bool
	firstRound        int64
	lastObservedRound int64
	observations      uint8
	committedNonce    uint64
	probableNonce     uint64
}

type resyncRecoveryCooldown struct {
	parentHash []byte
	untilRound int64
}

type postBootstrapWatchdogBypass struct {
	armed            bool
	generation       uint64
	armedRound       int64
	lastAttemptRound int64
	lastAttemptTime  time.Time
	committedNonce   uint64
	probableNonce    uint64
}

type resyncRecoveryState struct {
	candidates          [maxResyncRecoveryCandidates]resyncRecoveryCandidate
	cooldowns           [maxResyncRecoveryCandidates]resyncRecoveryCooldown
	bypass              postBootstrapWatchdogBypass
	lastChronologyRound int64
	chronologySet       bool
	lastFastActionRound int64
	fastActionSet       bool
	lastFastActionTime  time.Time
	closed              bool
	nextGeneration      uint64
}

type resyncRecoveryAction struct {
	generation      uint64
	parentHash      []byte
	childNonce      uint64
	childEpoch      uint32
	childEpochStart bool
}

func (boot *baseBootstrap) observeRecoveryHeader(header data.HeaderHandler) {
	if check.IfNil(header) || header.GetShardID() != boot.shardCoordinator.SelfId() {
		return
	}
	if !common.IsProofsFlagEnabledForHeader(boot.enableEpochsHandler, header) {
		return
	}

	probableNonce := boot.forkDetector.ProbableHighestNonce()
	if probableNonce == math.MaxUint64 || header.GetNonce() <= probableNonce+1 {
		return
	}
	if header.IsStartOfEpochBlock() && header.GetEpoch() == 0 {
		return
	}

	parentHash := header.GetPrevHash()
	if len(parentHash) == 0 {
		return
	}

	committedNonce := boot.currentCommittedNonce()

	boot.mutRecovery.Lock()
	defer boot.mutRecovery.Unlock()
	round := boot.roundHandler.Index()
	if round < 0 {
		return
	}
	if boot.recoveryState.closed {
		return
	}

	if boot.recoveryState.chronologySet && round < boot.recoveryState.lastChronologyRound {
		boot.resetRecoveryStateLocked()
	}
	boot.recoveryState.chronologySet = true
	boot.recoveryState.lastChronologyRound = round
	boot.cleanupRecoveryCandidatesLocked(round)

	for idx := range boot.recoveryState.candidates {
		candidate := &boot.recoveryState.candidates[idx]
		if !candidate.active || !sameRecoveryCandidate(candidate, header, parentHash) {
			continue
		}
		if candidate.committedNonce != committedNonce || candidate.probableNonce != probableNonce {
			boot.recoveryState.candidates[idx] = resyncRecoveryCandidate{}
			break
		}
		if round > candidate.lastObservedRound {
			candidate.lastObservedRound = round
			if candidate.observations < math.MaxUint8 {
				candidate.observations++
			}
			boot.recoveryEvalSet.Store(false)
		}
		return
	}

	if boot.isRecoveryHashCoolingDownLocked(parentHash, round) {
		return
	}
	for idx := range boot.recoveryState.candidates {
		if boot.recoveryState.candidates[idx].active {
			continue
		}

		boot.recoveryState.nextGeneration++
		if boot.recoveryState.nextGeneration == 0 {
			boot.recoveryState.nextGeneration++
		}
		boot.recoveryState.candidates[idx] = resyncRecoveryCandidate{
			active:            true,
			generation:        boot.recoveryState.nextGeneration,
			parentHash:        append([]byte(nil), parentHash...),
			childNonce:        header.GetNonce(),
			childEpoch:        header.GetEpoch(),
			childEpochStart:   header.IsStartOfEpochBlock(),
			firstRound:        round,
			lastObservedRound: round,
			observations:      1,
			committedNonce:    committedNonce,
			probableNonce:     probableNonce,
		}
		if !boot.recoveryActive.Swap(true) {
			boot.recoveryEvalSet.Store(false)
		}
		return
	}
}

func sameRecoveryCandidate(candidate *resyncRecoveryCandidate, header data.HeaderHandler, parentHash []byte) bool {
	return candidate.childNonce == header.GetNonce() &&
		candidate.childEpoch == header.GetEpoch() &&
		candidate.childEpochStart == header.IsStartOfEpochBlock() &&
		bytes.Equal(candidate.parentHash, parentHash)
}

func (boot *baseBootstrap) evaluateFastRecovery(round int64) {
	if !boot.recoveryActive.Load() {
		return
	}
	if boot.recoveryEvalSet.Load() && boot.recoveryEvalRound.Load() == round {
		return
	}
	boot.recoveryEvalRound.Store(round)
	boot.recoveryEvalSet.Store(true)
	if round < 0 || boot.roundHandler.BeforeGenesis() || !boot.networkWatcher.IsConnectedToTheNetwork() {
		return
	}
	if boot.isInImportMode || boot.pendingV3Realign || boot.pendingV3RollBack != nil {
		return
	}

	committedNonce := boot.currentCommittedNonce()
	probableNonce := boot.forkDetector.ProbableHighestNonce()
	if committedNonce != probableNonce {
		return
	}

	now := time.Now()
	if boot.requestHandler == nil {
		return
	}
	requestInterval := boot.requestHandler.RequestInterval()
	boot.mutRecovery.Lock()
	if boot.recoveryState.closed {
		boot.mutRecovery.Unlock()
		return
	}
	chronologyRound := boot.roundHandler.Index()
	if chronologyRound < 0 {
		boot.mutRecovery.Unlock()
		return
	}
	if boot.recoveryState.chronologySet && chronologyRound < boot.recoveryState.lastChronologyRound {
		boot.resetRecoveryStateLocked()
		boot.mutRecovery.Unlock()
		return
	}
	boot.recoveryState.chronologySet = true
	boot.recoveryState.lastChronologyRound = chronologyRound
	boot.cleanupRecoveryCandidatesLocked(chronologyRound)
	if boot.recoveryState.fastActionSet && boot.recoveryState.lastFastActionRound == round {
		boot.mutRecovery.Unlock()
		return
	}

	selected := -1
	for idx := range boot.recoveryState.candidates {
		candidate := &boot.recoveryState.candidates[idx]
		if !candidate.active || candidate.observations < 2 ||
			candidate.committedNonce != committedNonce || candidate.probableNonce != probableNonce {
			continue
		}
		if selected < 0 || recoveryCandidateLess(candidate, &boot.recoveryState.candidates[selected]) {
			selected = idx
		}
	}
	if selected < 0 {
		boot.mutRecovery.Unlock()
		return
	}
	if !recoveryRequestIntervalElapsed(now, boot.recoveryState.lastFastActionTime, requestInterval) {
		boot.mutRecovery.Unlock()
		return
	}

	candidate := boot.recoveryState.candidates[selected]
	action := resyncRecoveryAction{
		generation:      candidate.generation,
		parentHash:      append([]byte(nil), candidate.parentHash...),
		childNonce:      candidate.childNonce,
		childEpoch:      candidate.childEpoch,
		childEpochStart: candidate.childEpochStart,
	}
	boot.recoveryState.lastFastActionRound = round
	boot.recoveryState.fastActionSet = true
	boot.recoveryState.lastFastActionTime = now
	boot.mutRecovery.Unlock()

	boot.executeFastRecoveryAction(action, chronologyRound)
}

func recoveryCandidateLess(candidate *resyncRecoveryCandidate, selected *resyncRecoveryCandidate) bool {
	if candidate.observations != selected.observations {
		return candidate.observations > selected.observations
	}
	if candidate.childNonce != selected.childNonce {
		return candidate.childNonce < selected.childNonce
	}
	return bytes.Compare(candidate.parentHash, selected.parentHash) < 0
}

func (boot *baseBootstrap) executeFastRecoveryAction(action resyncRecoveryAction, round int64) {
	expectedEpoch, ok := expectedRecoveryParentEpoch(action.childEpoch, action.childEpochStart)
	if !ok {
		boot.expireRecoveryCandidate(action, round)
		return
	}

	parent, err := boot.headers.GetHeaderByHash(action.parentHash)
	if err != nil || check.IfNil(parent) {
		if boot.recoveryActionStillActive(action) {
			boot.requestRecoveryHeader(action.parentHash, expectedEpoch)
		}
		return
	}

	parentHash, err := core.CalculateHash(boot.marshalizer, boot.hasher, parent)
	if err != nil || !bytes.Equal(parentHash, action.parentHash) ||
		!validRecoveryParent(parent, action, expectedEpoch, boot.shardCoordinator.SelfId()) ||
		!common.IsProofsFlagEnabledForHeader(boot.enableEpochsHandler, parent) {
		boot.expireRecoveryCandidate(action, round)
		return
	}

	if boot.proofs.HasProof(parent.GetShardID(), action.parentHash) {
		boot.removeRecoveryCandidate(action)
		return
	}

	if boot.recoveryActionStillActive(action) {
		boot.requestHandler.RequestEquivalentProofByHashForEpoch(parent.GetShardID(), action.parentHash, parent.GetEpoch())
	}
}

func expectedRecoveryParentEpoch(childEpoch uint32, childEpochStart bool) (uint32, bool) {
	if !childEpochStart {
		return childEpoch, true
	}
	if childEpoch == 0 {
		return 0, false
	}
	return childEpoch - 1, true
}

func validRecoveryParent(parent data.HeaderHandler, action resyncRecoveryAction, expectedEpoch uint32, selfShardID uint32) bool {
	if parent.GetShardID() != selfShardID || parent.GetNonce() == math.MaxUint64 {
		return false
	}
	return parent.GetNonce()+1 == action.childNonce && parent.GetEpoch() == expectedEpoch
}

func (boot *baseBootstrap) requestRecoveryHeader(hash []byte, epoch uint32) {
	if boot.shardCoordinator.SelfId() == core.MetachainShardId {
		boot.requestHandler.RequestMetaHeaderForEpoch(hash, epoch)
		return
	}
	boot.requestHandler.RequestShardHeaderForEpoch(boot.shardCoordinator.SelfId(), hash, epoch)
}

func (boot *baseBootstrap) expireRecoveryCandidate(action resyncRecoveryAction, round int64) {
	boot.mutRecovery.Lock()
	defer boot.mutRecovery.Unlock()

	for idx := range boot.recoveryState.candidates {
		candidate := &boot.recoveryState.candidates[idx]
		if !candidate.active || !recoveryActionMatchesCandidate(action, candidate) {
			continue
		}
		boot.addRecoveryCooldownLocked(candidate.parentHash, round)
		boot.recoveryState.candidates[idx] = resyncRecoveryCandidate{}
		boot.updateRecoveryActiveLocked()
		return
	}
}

func (boot *baseBootstrap) removeRecoveryCandidate(action resyncRecoveryAction) {
	boot.mutRecovery.Lock()
	defer boot.mutRecovery.Unlock()

	for idx := range boot.recoveryState.candidates {
		candidate := &boot.recoveryState.candidates[idx]
		if candidate.active && recoveryActionMatchesCandidate(action, candidate) {
			boot.recoveryState.candidates[idx] = resyncRecoveryCandidate{}
			boot.updateRecoveryActiveLocked()
			return
		}
	}
}

func recoveryActionMatchesCandidate(action resyncRecoveryAction, candidate *resyncRecoveryCandidate) bool {
	return action.generation == candidate.generation && action.childNonce == candidate.childNonce && action.childEpoch == candidate.childEpoch &&
		action.childEpochStart == candidate.childEpochStart && bytes.Equal(action.parentHash, candidate.parentHash)
}

func (boot *baseBootstrap) recoveryActionStillActive(action resyncRecoveryAction) bool {
	boot.mutRecovery.Lock()
	defer boot.mutRecovery.Unlock()

	if boot.recoveryState.closed {
		return false
	}
	for idx := range boot.recoveryState.candidates {
		candidate := &boot.recoveryState.candidates[idx]
		if candidate.active && recoveryActionMatchesCandidate(action, candidate) {
			return true
		}
	}
	return false
}

func (boot *baseBootstrap) cleanupRecoveryCandidatesLocked(round int64) {
	for idx := range boot.recoveryState.cooldowns {
		if len(boot.recoveryState.cooldowns[idx].parentHash) > 0 && round >= boot.recoveryState.cooldowns[idx].untilRound {
			boot.recoveryState.cooldowns[idx] = resyncRecoveryCooldown{}
		}
	}
	for idx := range boot.recoveryState.candidates {
		candidate := &boot.recoveryState.candidates[idx]
		if !candidate.active || round < candidate.firstRound || round-candidate.firstRound < resyncCandidateTTL {
			continue
		}
		boot.addRecoveryCooldownLocked(candidate.parentHash, round)
		boot.recoveryState.candidates[idx] = resyncRecoveryCandidate{}
	}
	boot.updateRecoveryActiveLocked()
}

func (boot *baseBootstrap) addRecoveryCooldownLocked(parentHash []byte, round int64) {
	selected := 0
	for idx := range boot.recoveryState.cooldowns {
		if len(boot.recoveryState.cooldowns[idx].parentHash) == 0 ||
			boot.recoveryState.cooldowns[idx].untilRound < boot.recoveryState.cooldowns[selected].untilRound {
			selected = idx
		}
	}
	untilRound := int64(math.MaxInt64)
	if round <= math.MaxInt64-resyncCandidateCooldown {
		untilRound = round + resyncCandidateCooldown
	}
	boot.recoveryState.cooldowns[selected] = resyncRecoveryCooldown{
		parentHash: append([]byte(nil), parentHash...),
		untilRound: untilRound,
	}
}

func (boot *baseBootstrap) isRecoveryHashCoolingDownLocked(parentHash []byte, round int64) bool {
	for idx := range boot.recoveryState.cooldowns {
		cooldown := &boot.recoveryState.cooldowns[idx]
		if round < cooldown.untilRound && bytes.Equal(parentHash, cooldown.parentHash) {
			return true
		}
	}
	return false
}

func (boot *baseBootstrap) armPostBootstrapWatchdogBypass() {
	if boot.roundHandler == nil || boot.networkWatcher == nil || boot.forkDetector == nil || boot.processConfigsHandler == nil {
		return
	}
	if boot.isInImportMode || boot.pendingV3Realign || boot.pendingV3RollBack != nil ||
		boot.roundHandler.BeforeGenesis() || !boot.networkWatcher.IsConnectedToTheNetwork() {
		return
	}

	currentHeader := boot.chainHandler.GetCurrentBlockHeader()
	if check.IfNil(currentHeader) {
		currentHeader = boot.chainHandler.GetGenesisHeader()
	}
	if check.IfNil(currentHeader) {
		return
	}

	committedNonce := currentHeader.GetNonce()
	probableNonce := boot.forkDetector.ProbableHighestNonce()
	currentRound := boot.roundHandler.Index()
	if committedNonce != probableNonce || currentRound < 0 || uint64(currentRound) <= currentHeader.GetRound() {
		return
	}
	if uint64(currentRound)-currentHeader.GetRound() <= boot.getMaxRoundsWithoutBlockReceived(currentHeader.GetRound()) {
		return
	}

	boot.mutRecovery.Lock()
	if boot.recoveryState.closed {
		boot.mutRecovery.Unlock()
		return
	}
	boot.recoveryState.bypass.generation = nextRecoveryGeneration(boot.recoveryState.bypass.generation)
	boot.recoveryState.bypass.armed = true
	boot.recoveryState.bypass.armedRound = currentRound
	boot.recoveryState.bypass.lastAttemptRound = -1
	boot.recoveryState.bypass.lastAttemptTime = time.Time{}
	boot.recoveryState.bypass.committedNonce = committedNonce
	boot.recoveryState.bypass.probableNonce = probableNonce
	boot.recoveryBypass.Store(true)
	boot.mutRecovery.Unlock()
}

func (boot *baseBootstrap) usePostBootstrapWatchdogBypass(round int64) (bool, uint64) {
	if round < 0 || boot.isInImportMode || boot.pendingV3Realign || boot.pendingV3RollBack != nil ||
		!boot.networkWatcher.IsConnectedToTheNetwork() {
		boot.disarmPostBootstrapWatchdogBypass()
		return false, 0
	}

	committedNonce := boot.currentCommittedNonce()
	probableNonce := boot.forkDetector.ProbableHighestNonce()
	now := time.Now()
	if boot.requestHandler == nil {
		boot.disarmPostBootstrapWatchdogBypass()
		return false, 0
	}
	requestInterval := boot.requestHandler.RequestInterval()

	boot.mutRecovery.Lock()
	defer boot.mutRecovery.Unlock()

	bypass := &boot.recoveryState.bypass
	if boot.recoveryState.closed || !bypass.armed {
		return false, 0
	}
	if round < bypass.armedRound || round-bypass.armedRound >= maxWatchdogBypassRounds ||
		committedNonce != bypass.committedNonce || probableNonce != bypass.probableNonce || committedNonce != probableNonce {
		boot.clearWatchdogBypassLocked()
		return false, 0
	}
	if bypass.lastAttemptRound == round ||
		!recoveryRequestIntervalElapsed(now, bypass.lastAttemptTime, requestInterval) {
		return false, 0
	}

	bypass.lastAttemptRound = round
	bypass.lastAttemptTime = now
	return true, bypass.generation
}

func (boot *baseBootstrap) clearRecoveryAfterProgress() {
	if !boot.recoveryActive.Load() && !boot.recoveryBypass.Load() {
		return
	}

	committedNonce := boot.currentCommittedNonce()
	probableNonce := boot.forkDetector.ProbableHighestNonce()

	boot.mutRecovery.Lock()
	defer boot.mutRecovery.Unlock()

	for idx := range boot.recoveryState.candidates {
		candidate := &boot.recoveryState.candidates[idx]
		if candidate.active && (candidate.committedNonce != committedNonce || candidate.probableNonce != probableNonce) {
			boot.recoveryState.candidates[idx] = resyncRecoveryCandidate{}
		}
	}
	boot.updateRecoveryActiveLocked()
	bypass := &boot.recoveryState.bypass
	if bypass.armed && (bypass.committedNonce != committedNonce || bypass.probableNonce != probableNonce || committedNonce != probableNonce) {
		boot.clearWatchdogBypassLocked()
	}
}

func (boot *baseBootstrap) disarmPostBootstrapWatchdogBypass() {
	boot.mutRecovery.Lock()
	boot.clearWatchdogBypassLocked()
	boot.mutRecovery.Unlock()
}

func (boot *baseBootstrap) clearWatchdogBypassLocked() {
	generation := nextRecoveryGeneration(boot.recoveryState.bypass.generation)
	boot.recoveryState.bypass = postBootstrapWatchdogBypass{generation: generation}
	boot.recoveryBypass.Store(false)
}

func (boot *baseBootstrap) isWatchdogBypassGenerationActive(generation uint64) bool {
	if !boot.networkWatcher.IsConnectedToTheNetwork() {
		return false
	}

	committedNonce := boot.currentCommittedNonce()
	probableNonce := boot.forkDetector.ProbableHighestNonce()
	boot.mutRecovery.Lock()
	defer boot.mutRecovery.Unlock()

	bypass := &boot.recoveryState.bypass
	return !boot.recoveryState.closed && bypass.armed && bypass.generation == generation &&
		bypass.committedNonce == committedNonce && bypass.probableNonce == probableNonce && committedNonce == probableNonce
}

func (boot *baseBootstrap) currentCommittedNonce() uint64 {
	currentHeader := boot.chainHandler.GetCurrentBlockHeader()
	if !check.IfNil(currentHeader) {
		return currentHeader.GetNonce()
	}
	genesisHeader := boot.chainHandler.GetGenesisHeader()
	if check.IfNil(genesisHeader) {
		return 0
	}
	return genesisHeader.GetNonce()
}

func (boot *baseBootstrap) resetRecoveryStateLocked() {
	generation := nextRecoveryGeneration(boot.recoveryState.bypass.generation)
	nextGeneration := boot.recoveryState.nextGeneration
	boot.recoveryState = resyncRecoveryState{}
	boot.recoveryState.bypass.generation = generation
	boot.recoveryState.nextGeneration = nextGeneration
	boot.recoveryActive.Store(false)
	boot.recoveryBypass.Store(false)
	boot.recoveryEvalSet.Store(false)
}

func (boot *baseBootstrap) closeRecovery() {
	boot.mutRecovery.Lock()
	generation := nextRecoveryGeneration(boot.recoveryState.bypass.generation)
	boot.recoveryState = resyncRecoveryState{closed: true}
	boot.recoveryState.bypass.generation = generation
	boot.recoveryActive.Store(false)
	boot.recoveryBypass.Store(false)
	boot.recoveryEvalSet.Store(false)
	boot.mutRecovery.Unlock()
}

func (boot *baseBootstrap) updateRecoveryActiveLocked() {
	for idx := range boot.recoveryState.candidates {
		if boot.recoveryState.candidates[idx].active {
			boot.recoveryActive.Store(true)
			return
		}
	}
	boot.recoveryActive.Store(false)
	boot.recoveryEvalSet.Store(false)
}

func recoveryRequestIntervalElapsed(now time.Time, last time.Time, interval time.Duration) bool {
	return last.IsZero() || now.Before(last) || now.Sub(last) >= interval
}

func nextRecoveryGeneration(generation uint64) uint64 {
	generation++
	if generation == 0 {
		return 1
	}
	return generation
}
