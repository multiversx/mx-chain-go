package hooks

import "testing"

func TestApplyDRWASyncEnvelopeRecordsUnauthorizedCallerMetric(t *testing.T) {
	resetDRWAMetrics()
	adapter := newMockDRWASyncStateAdapter()
	envelope := &drwaSyncEnvelope{
		CallerDomain: "random_caller",
		Operations: []drwaSyncOperation{{
			OperationType: drwaSyncOpTokenPolicy,
			TokenID:       "CARBON-1",
			Version:       1,
			Body:          []byte(`{}`),
		}},
	}
	hash, _ := computeDRWASyncHash(envelope.CallerDomain, envelope.Operations)
	envelope.PayloadHash = hash

	_, err := applyDRWASyncEnvelope(adapter, envelope, 16, []byte("random_caller"))
	if err == nil {
		t.Fatalf("expected unauthorized caller rejection")
	}

	metrics := snapshotDRWAMetrics()
	if metrics[drwaMetricSyncUnauthorizedCaller] != 1 {
		t.Fatalf("expected unauthorized caller metric, got %+v", metrics)
	}
	if metrics[drwaMetricSyncApplyFailure] != 1 {
		t.Fatalf("expected sync apply failure metric, got %+v", metrics)
	}
}

func TestApplyDRWASyncEnvelopeRecordsHashMismatchMetric(t *testing.T) {
	resetDRWAMetrics()
	adapter := newMockDRWASyncStateAdapter()
	envelope := &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerPolicyRegistry,
		PayloadHash:  []byte("wrong"),
		Operations: []drwaSyncOperation{{
			OperationType: drwaSyncOpTokenPolicy,
			TokenID:       "CARBON-1",
			Version:       1,
			Body:          []byte(`{}`),
		}},
	}

	_, err := applyDRWASyncEnvelope(adapter, envelope, 16, []byte("policy_registry"))
	if err == nil {
		t.Fatalf("expected hash mismatch rejection")
	}

	metrics := snapshotDRWAMetrics()
	if metrics[drwaMetricSyncHashMismatch] != 1 {
		t.Fatalf("expected hash mismatch metric, got %+v", metrics)
	}
	if metrics[drwaMetricSyncApplyFailure] != 1 {
		t.Fatalf("expected sync apply failure metric, got %+v", metrics)
	}
}

func TestApplyDRWASyncEnvelopeRecordsSuccessMetric(t *testing.T) {
	resetDRWAMetrics()
	adapter := newMockDRWASyncStateAdapter()
	envelope := &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerPolicyRegistry,
		Operations: []drwaSyncOperation{{
			OperationType: drwaSyncOpTokenPolicy,
			TokenID:       "CARBON-1",
			Version:       1,
			Body:          []byte(`{}`),
		}},
	}
	hash, _ := computeDRWASyncHash(envelope.CallerDomain, envelope.Operations)
	envelope.PayloadHash = hash

	_, err := applyDRWASyncEnvelope(adapter, envelope, 16, []byte("policy_registry"))
	if err != nil {
		t.Fatalf("expected successful sync apply, got %v", err)
	}

	metrics := snapshotDRWAMetrics()
	if metrics[drwaMetricSyncApplySuccess] != 1 {
		t.Fatalf("expected sync apply success metric, got %+v", metrics)
	}
}

func TestValidateDRWASyncVersionRecordsReplayMetric(t *testing.T) {
	resetDRWAMetrics()
	err := validateDRWASyncVersion(2, 2)
	if err == nil {
		t.Fatalf("expected replay conflict")
	}

	metrics := snapshotDRWAMetrics()
	if metrics[drwaMetricSyncReplayRejected] != 1 {
		t.Fatalf("expected replay rejection metric, got %+v", metrics)
	}
}

func TestInspectDRWARecoveryStateRecordsSafeModeAndNonRepairableMetrics(t *testing.T) {
	resetDRWAMetrics()
	adapter := newMockDRWASyncStateAdapter()
	adapter.tokenVersions["CARBON-1"] = 2
	adapter.tokenBodies["CARBON-1"] = []byte(`{"regulated":false}`)
	adapter.holderVersions["CARBON-1|erd1a"] = 2
	adapter.holderBodies["CARBON-1|erd1a"] = nil

	manifest := &drwaRecoveryManifest{
		TokenID:       "CARBON-1",
		PolicyVersion: 3,
		PolicyBody:    []byte(`{"regulated":true}`),
		Holders: []drwaRecoveryHolder{
			{Address: "erd1a", Version: 2, Body: []byte(`{"kyc":"approved"}`)},
		},
	}

	_, err := inspectDRWARecoveryState(adapter, manifest)
	if err != nil {
		t.Fatalf("inspect recovery: %v", err)
	}

	metrics := snapshotDRWAMetrics()
	if metrics[drwaMetricRecoverySafeModeReport] != 1 {
		t.Fatalf("expected safe-mode metric, got %+v", metrics)
	}
	if metrics[drwaMetricRecoveryNonRepairable] != 1 {
		t.Fatalf("expected non-repairable metric, got %+v", metrics)
	}
}

func TestBuildDRWARolloutVerificationReportRecordsPassAndRejectMetrics(t *testing.T) {
	resetDRWAMetrics()
	healthyManifest := &drwaRolloutManifest{
		TokenID:                "CARBON-1",
		Issuer:                 "issuer-1",
		Stage:                  drwaRolloutStageCanary,
		MaxSyncFailureRateBps:  10,
		MaxAPIErrorRateBps:     10,
		MaxDenialMismatchCount: 1,
	}

	_, err := buildDRWARolloutVerificationReport(healthyManifest, &drwaRolloutObservedMetrics{
		SyncFailureRateBps:  1,
		APIErrorRateBps:     1,
		DenialMismatchCount: 0,
	})
	if err != nil {
		t.Fatalf("build healthy verification report: %v", err)
	}

	_, err = buildDRWARolloutVerificationReport(healthyManifest, &drwaRolloutObservedMetrics{
		SyncFailureRateBps:  11,
		APIErrorRateBps:     1,
		DenialMismatchCount: 0,
	})
	if err != nil {
		t.Fatalf("build rejected verification report: %v", err)
	}

	metrics := snapshotDRWAMetrics()
	if metrics[drwaMetricRolloutVerificationPass] != 1 {
		t.Fatalf("expected rollout pass metric, got %+v", metrics)
	}
	if metrics[drwaMetricRolloutVerificationReject] != 1 {
		t.Fatalf("expected rollout reject metric, got %+v", metrics)
	}
}
