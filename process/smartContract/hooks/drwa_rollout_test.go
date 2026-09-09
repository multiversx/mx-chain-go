package hooks

import "testing"

func TestValidateDRWARolloutManifestRejectsInvalidStage(t *testing.T) {
	t.Parallel()

	err := validateDRWARolloutManifest(&drwaRolloutManifest{
		TokenID: "CARBON-1",
		Issuer:  "issuer-1",
		Stage:   "invalid",
	})
	if err == nil {
		t.Fatalf("expected invalid stage rejection")
	}
}

func TestValidateDRWARolloutManifestRejectsZeroThresholds(t *testing.T) {
	t.Parallel()

	err := validateDRWARolloutManifest(&drwaRolloutManifest{
		TokenID:                "CARBON-1",
		Issuer:                 "issuer-1",
		Stage:                  drwaRolloutStageCanary,
		MaxSyncFailureRateBps:  0,
		MaxAPIErrorRateBps:     10,
		MaxDenialMismatchCount: 1,
	})
	if err != errDRWAInvalidRolloutThresholds {
		t.Fatalf("expected zero-threshold rejection, got %v", err)
	}
}

func TestValidateDRWARolloutManifestRejectsThresholdsAboveUpperBounds(t *testing.T) {
	t.Parallel()

	err := validateDRWARolloutManifest(&drwaRolloutManifest{
		TokenID:                "CARBON-1",
		Issuer:                 "issuer-1",
		Stage:                  drwaRolloutStageCanary,
		MaxSyncFailureRateBps:  drwaRolloutMaxFailureRateBpsUpperBound + 1,
		MaxAPIErrorRateBps:     10,
		MaxDenialMismatchCount: 1,
	})
	if err != errDRWAInvalidRolloutThresholds {
		t.Fatalf("expected upper-bound rejection, got %v", err)
	}
}

func TestInspectDRWARolloutPreflightRequiresPolicy(t *testing.T) {
	t.Parallel()

	adapter := newMockDRWASyncStateAdapter()
	_, err := inspectDRWARolloutPreflight(adapter, &drwaRolloutManifest{
		TokenID:                "CARBON-1",
		Issuer:                 "issuer-1",
		Stage:                  drwaRolloutStageCanary,
		ExpectedPolicyVersion:  2,
		MaxSyncFailureRateBps:  10,
		MaxAPIErrorRateBps:     10,
		MaxDenialMismatchCount: 1,
	}, nil)
	if err != errDRWARolloutPolicyMissing {
		t.Fatalf("expected policy missing, got %v", err)
	}
}

func TestInspectDRWARolloutPreflightFlagsMissingHolders(t *testing.T) {
	t.Parallel()

	adapter := newMockDRWASyncStateAdapter()
	adapter.tokenVersions["CARBON-1"] = 2
	adapter.tokenBodies["CARBON-1"] = []byte(`{"regulated":true}`)
	adapter.holderVersions["CARBON-1|erd1a"] = 1
	adapter.holderBodies["CARBON-1|erd1a"] = []byte(`{"kyc":"approved"}`)

	report, err := inspectDRWARolloutPreflight(adapter, &drwaRolloutManifest{
		TokenID:                "CARBON-1",
		Issuer:                 "issuer-1",
		Stage:                  drwaRolloutStageCanary,
		ExpectedPolicyVersion:  2,
		RequiredHolders:        []string{"erd1a", "erd1b"},
		MaxSyncFailureRateBps:  10,
		MaxAPIErrorRateBps:     10,
		MaxDenialMismatchCount: 1,
	}, nil)
	if err != nil {
		t.Fatalf("inspect rollout: %v", err)
	}
	if len(report.MissingHolders) != 1 || report.MissingHolders[0] != "erd1b" {
		t.Fatalf("unexpected missing holders: %+v", report.MissingHolders)
	}
	if report.Ready {
		t.Fatalf("expected rollout not ready with missing holders")
	}
}

func TestInspectDRWARolloutPreflightBlocksLimitedWhenCanaryRequired(t *testing.T) {
	t.Parallel()

	adapter := newMockDRWASyncStateAdapter()
	adapter.tokenVersions["CARBON-1"] = 1
	adapter.tokenBodies["CARBON-1"] = []byte(`{"regulated":true}`)

	_, err := inspectDRWARolloutPreflight(adapter, &drwaRolloutManifest{
		TokenID:                "CARBON-1",
		Issuer:                 "issuer-1",
		Stage:                  drwaRolloutStageLimited,
		ExpectedPolicyVersion:  2,
		MaxSyncFailureRateBps:  10,
		MaxAPIErrorRateBps:     10,
		MaxDenialMismatchCount: 1,
	}, nil)
	if err != errDRWARolloutNotCanaryReady {
		t.Fatalf("expected canary-ready rejection, got %v", err)
	}
}

func TestMarshalDRWARolloutEvidenceAndPersist(t *testing.T) {
	t.Parallel()

	report := &drwaRolloutPreflightReport{
		TokenID: "CARBON-1",
		Stage:   drwaRolloutStageCanary,
		Checks:  []string{"token policy present"},
		Ready:   true,
	}

	payload, err := marshalDRWARolloutEvidence(report)
	if err != nil {
		t.Fatalf("marshal rollout evidence: %v", err)
	}
	if len(payload) == 0 {
		t.Fatalf("expected rollout evidence payload")
	}

	adapter := newMockDRWASyncStateAdapter()
	err = persistDRWARolloutEvidence(adapter, "CARBON-1", payload)
	if err != nil {
		t.Fatalf("persist rollout evidence: %v", err)
	}
	if len(adapter.rolloutEvidence["CARBON-1"]) == 0 {
		t.Fatalf("expected persisted rollout evidence")
	}
}

func TestBuildDRWARolloutVerificationReportAcceptsHealthyMetrics(t *testing.T) {
	t.Parallel()

	report, err := buildDRWARolloutVerificationReport(&drwaRolloutManifest{
		TokenID:                "CARBON-1",
		Issuer:                 "issuer-1",
		Stage:                  drwaRolloutStageCanary,
		MaxSyncFailureRateBps:  10,
		MaxAPIErrorRateBps:     20,
		MaxDenialMismatchCount: 1,
	}, &drwaRolloutObservedMetrics{
		SyncFailureRateBps:  5,
		APIErrorRateBps:     10,
		DenialMismatchCount: 0,
	})
	if err != nil {
		t.Fatalf("build verification report: %v", err)
	}
	if !report.Accepted {
		t.Fatalf("expected verification acceptance, got %+v", report)
	}
}

func TestBuildDRWARolloutVerificationReportRejectsBadMetrics(t *testing.T) {
	t.Parallel()

	report, err := buildDRWARolloutVerificationReport(&drwaRolloutManifest{
		TokenID:                "CARBON-1",
		Issuer:                 "issuer-1",
		Stage:                  drwaRolloutStageLimited,
		MaxSyncFailureRateBps:  10,
		MaxAPIErrorRateBps:     20,
		MaxDenialMismatchCount: 1,
	}, &drwaRolloutObservedMetrics{
		SyncFailureRateBps:  50,
		APIErrorRateBps:     21,
		DenialMismatchCount: 2,
	})
	if err != nil {
		t.Fatalf("build verification report: %v", err)
	}
	if report.Accepted {
		t.Fatalf("expected verification rejection, got %+v", report)
	}
	if len(report.FailedChecks) != 3 {
		t.Fatalf("expected 3 failed checks, got %+v", report.FailedChecks)
	}
}

func TestMarshalDRWARolloutVerificationReport(t *testing.T) {
	t.Parallel()

	payload, err := marshalDRWARolloutVerificationReport(&drwaRolloutVerificationReport{
		TokenID:      "CARBON-1",
		Stage:        drwaRolloutStageProduction,
		PassedChecks: []string{"api error rate within threshold"},
		Accepted:     true,
	})
	if err != nil {
		t.Fatalf("marshal verification report: %v", err)
	}
	if len(payload) == 0 {
		t.Fatalf("expected verification payload")
	}
}

func TestPersistDRWARolloutVerification(t *testing.T) {
	t.Parallel()

	adapter := newMockDRWASyncStateAdapter()
	payload, err := marshalDRWARolloutVerificationReport(&drwaRolloutVerificationReport{
		TokenID:      "CARBON-1",
		Stage:        drwaRolloutStageLimited,
		FailedChecks: []string{"sync failure rate exceeded: observed=50 threshold=10"},
		Accepted:     false,
	})
	if err != nil {
		t.Fatalf("marshal verification report: %v", err)
	}

	err = persistDRWARolloutVerification(adapter, "CARBON-1", payload)
	if err != nil {
		t.Fatalf("persist verification report: %v", err)
	}
	if len(adapter.rolloutVerification["CARBON-1"]) == 0 {
		t.Fatalf("expected persisted verification report")
	}
}

func TestValidateDRWARolloutAdmissionRequiresVerificationBeyondCanary(t *testing.T) {
	t.Parallel()

	err := validateDRWARolloutAdmission(&drwaRolloutManifest{
		TokenID:                "CARBON-1",
		Issuer:                 "issuer-1",
		Stage:                  drwaRolloutStageLimited,
		MaxSyncFailureRateBps:  10,
		MaxAPIErrorRateBps:     10,
		MaxDenialMismatchCount: 1,
	}, &drwaRolloutPreflightReport{
		TokenID: "CARBON-1",
		Stage:   drwaRolloutStageLimited,
		Ready:   true,
	}, nil)
	if err != errDRWARolloutVerificationRequired {
		t.Fatalf("expected verification-required error, got %v", err)
	}
}

func TestValidateDRWARolloutAdmissionRejectsFailedVerification(t *testing.T) {
	t.Parallel()

	err := validateDRWARolloutAdmission(&drwaRolloutManifest{
		TokenID:                "CARBON-1",
		Issuer:                 "issuer-1",
		Stage:                  drwaRolloutStageProduction,
		MaxSyncFailureRateBps:  10,
		MaxAPIErrorRateBps:     10,
		MaxDenialMismatchCount: 1,
	}, &drwaRolloutPreflightReport{
		TokenID: "CARBON-1",
		Stage:   drwaRolloutStageProduction,
		Ready:   true,
	}, &drwaRolloutVerificationReport{
		TokenID:  "CARBON-1",
		Stage:    drwaRolloutStageProduction,
		Accepted: false,
	})
	if err != errDRWARolloutVerificationRejected {
		t.Fatalf("expected verification-rejected error, got %v", err)
	}
}

func TestValidateDRWARolloutAdmissionRejectsPreflightTokenMismatch(t *testing.T) {
	t.Parallel()

	err := validateDRWARolloutAdmission(&drwaRolloutManifest{
		TokenID:                "CARBON-1",
		Issuer:                 "issuer-1",
		Stage:                  drwaRolloutStageCanary,
		MaxSyncFailureRateBps:  10,
		MaxAPIErrorRateBps:     10,
		MaxDenialMismatchCount: 1,
	}, &drwaRolloutPreflightReport{
		TokenID: "OTHER-1",
		Stage:   drwaRolloutStageCanary,
		Ready:   true,
	}, nil)
	if err != errDRWARolloutReportTokenMismatch {
		t.Fatalf("expected token-mismatch error, got %v", err)
	}
}

func TestValidateDRWARolloutAdmissionRejectsVerificationStageMismatch(t *testing.T) {
	t.Parallel()

	err := validateDRWARolloutAdmission(&drwaRolloutManifest{
		TokenID:                "CARBON-1",
		Issuer:                 "issuer-1",
		Stage:                  drwaRolloutStageProduction,
		MaxSyncFailureRateBps:  10,
		MaxAPIErrorRateBps:     10,
		MaxDenialMismatchCount: 1,
	}, &drwaRolloutPreflightReport{
		TokenID: "CARBON-1",
		Stage:   drwaRolloutStageProduction,
		Ready:   true,
	}, &drwaRolloutVerificationReport{
		TokenID:  "CARBON-1",
		Stage:    drwaRolloutStageLimited,
		Accepted: true,
	})
	if err != errDRWARolloutReportStageMismatch {
		t.Fatalf("expected stage-mismatch error, got %v", err)
	}
}

func TestValidateDRWARolloutAdmissionAllowsCanaryWithoutVerification(t *testing.T) {
	t.Parallel()

	err := validateDRWARolloutAdmission(&drwaRolloutManifest{
		TokenID:                "CARBON-1",
		Issuer:                 "issuer-1",
		Stage:                  drwaRolloutStageCanary,
		MaxSyncFailureRateBps:  10,
		MaxAPIErrorRateBps:     10,
		MaxDenialMismatchCount: 1,
	}, &drwaRolloutPreflightReport{
		TokenID: "CARBON-1",
		Stage:   drwaRolloutStageCanary,
		Ready:   true,
	}, nil)
	if err != nil {
		t.Fatalf("expected canary admission, got %v", err)
	}
}
