package statusHandler

import (
	"sync"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("statusHandler")

type processStatusHandler struct {
	mutStatus sync.RWMutex
	isIdle    bool
	// counted separately from isIdle: these spans must not take the block processing exclusion, and
	// they can overlap each other, so a single flag would be released by whichever span ends first
	numBackgroundJobBlockers uint32
	numBlockingSuspensions   uint32
}

// NewProcessStatusHandler creates a new instance of type processStatusHandler
func NewProcessStatusHandler() *processStatusHandler {
	return &processStatusHandler{
		isIdle: true, // always start as idle so the initial snapshots (if required) will work
	}
}

// SetBusy will set the internal state to "busy"
func (psh *processStatusHandler) SetBusy(reason string) {
	log.Debug("processStatusHandler.SetBusy", "reason", reason)

	psh.mutStatus.Lock()
	psh.isIdle = false
	psh.mutStatus.Unlock()
}

// TrySetBusy will atomically check if idle and set the internal state to "busy".
// Returns true if the state was successfully set to busy, false if already busy.
func (psh *processStatusHandler) TrySetBusy(reason string) bool {
	psh.mutStatus.Lock()
	defer psh.mutStatus.Unlock()

	if !psh.isIdle {
		log.Debug("processStatusHandler.TrySetBusy: already busy", "reason", reason)
		return false
	}

	log.Debug("processStatusHandler.TrySetBusy", "reason", reason)
	psh.isIdle = false
	return true
}

// SetIdle will set the internal state to "idle"
func (psh *processStatusHandler) SetIdle() {
	log.Debug("processStatusHandler.SetIdle")

	psh.mutStatus.Lock()
	psh.isIdle = true
	psh.mutStatus.Unlock()
}

// BlockBackgroundJobs marks the start of latency critical work that background jobs must yield to.
// Unlike SetBusy it does not claim the block processing exclusion, so callers are never rejected
func (psh *processStatusHandler) BlockBackgroundJobs(reason string) {
	log.Debug("processStatusHandler.BlockBackgroundJobs", "reason", reason)

	psh.mutStatus.Lock()
	psh.numBackgroundJobBlockers++
	psh.mutStatus.Unlock()
}

// UnblockBackgroundJobs marks the end of one span started by BlockBackgroundJobs
func (psh *processStatusHandler) UnblockBackgroundJobs() {
	psh.mutStatus.Lock()
	isUnmatched := psh.numBackgroundJobBlockers == 0
	if !isUnmatched {
		psh.numBackgroundJobBlockers--
	}
	psh.mutStatus.Unlock()

	// logged outside the lock: a commit holds this mutex and the trie snapshot polls it every ms
	if isUnmatched {
		log.Warn("processStatusHandler.UnblockBackgroundJobs: called more times than BlockBackgroundJobs")
	}
}

// SuspendBackgroundJobBlocking makes IsIdle ignore active blockers: for a caller that waits on a
// background job while itself holding a blocker, which would otherwise deadlock the wait
func (psh *processStatusHandler) SuspendBackgroundJobBlocking(reason string) {
	log.Debug("processStatusHandler.SuspendBackgroundJobBlocking", "reason", reason)

	psh.mutStatus.Lock()
	psh.numBlockingSuspensions++
	psh.mutStatus.Unlock()
}

// ResumeBackgroundJobBlocking ends one suspension started by SuspendBackgroundJobBlocking
func (psh *processStatusHandler) ResumeBackgroundJobBlocking() {
	psh.mutStatus.Lock()
	isUnmatched := psh.numBlockingSuspensions == 0
	if !isUnmatched {
		psh.numBlockingSuspensions--
	}
	psh.mutStatus.Unlock()

	// logged outside the lock: a commit holds this mutex and the trie snapshot polls it every ms
	if isUnmatched {
		log.Warn("processStatusHandler.ResumeBackgroundJobBlocking: called more times than SuspendBackgroundJobBlocking")
	}
}

// IsIdle returns true if the node is idle
func (psh *processStatusHandler) IsIdle() bool {
	psh.mutStatus.RLock()
	defer psh.mutStatus.RUnlock()

	return psh.isIdle && (psh.numBackgroundJobBlockers == 0 || psh.numBlockingSuspensions > 0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (psh *processStatusHandler) IsInterfaceNil() bool {
	return psh == nil
}
