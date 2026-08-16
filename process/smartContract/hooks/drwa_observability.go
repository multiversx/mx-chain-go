package hooks

import "sync"

const (
	drwaMetricSyncApplySuccess          = "sync_apply_success"
	drwaMetricSyncApplyFailure          = "sync_apply_failure"
	drwaMetricSyncUnauthorizedCaller    = "sync_unauthorized_caller"
	drwaMetricSyncHashMismatch          = "sync_hash_mismatch"
	drwaMetricSyncReplayRejected        = "sync_replay_rejected"
	drwaMetricSyncDecodeFailure         = "sync_decode_failure"
	drwaMetricRecoverySafeModeReport    = "recovery_safe_mode_report"
	drwaMetricRecoveryNonRepairable     = "recovery_non_repairable_report"
	drwaMetricRolloutVerificationPass   = "rollout_verification_pass"
	drwaMetricRolloutVerificationReject = "rollout_verification_reject"
)

type drwaObservability struct {
	mut      sync.Mutex
	counters map[string]uint64
}

var drwaMetrics = newDRWAObservability()

func newDRWAObservability() *drwaObservability {
	return &drwaObservability{
		counters: make(map[string]uint64),
	}
}

func (d *drwaObservability) increment(metric string) {
	d.mut.Lock()
	d.counters[metric]++
	d.mut.Unlock()
}

func (d *drwaObservability) snapshot() map[string]uint64 {
	d.mut.Lock()
	defer d.mut.Unlock()

	snapshot := make(map[string]uint64, len(d.counters))
	for key, value := range d.counters {
		snapshot[key] = value
	}

	return snapshot
}

func (d *drwaObservability) reset() {
	d.mut.Lock()
	d.counters = make(map[string]uint64)
	d.mut.Unlock()
}

func recordDRWAMetric(metric string) {
	drwaMetrics.increment(metric)
}

func snapshotDRWAMetrics() map[string]uint64 {
	return drwaMetrics.snapshot()
}

func resetDRWAMetrics() {
	drwaMetrics.reset()
}
