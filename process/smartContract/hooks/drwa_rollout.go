package hooks

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	errDRWANilRolloutManifest          = errors.New("nil DRWA rollout manifest")
	errDRWAEmptyRolloutTokenID         = errors.New("empty DRWA rollout token id")
	errDRWAEmptyRolloutIssuer          = errors.New("empty DRWA rollout issuer")
	errDRWAInvalidRolloutStage         = errors.New("invalid DRWA rollout stage")
	errDRWANilRolloutStateReader       = errors.New("nil DRWA rollout state reader")
	errDRWARolloutPolicyMissing        = errors.New("DRWA rollout token policy missing")
	errDRWARolloutNotCanaryReady       = errors.New("DRWA rollout not canary ready")
	errDRWARolloutSafeModeRequired     = errors.New("DRWA rollout requires safe mode clearance")
	errDRWANilRolloutEvidenceSink      = errors.New("nil DRWA rollout evidence sink")
	errDRWANilRolloutVerificationSink  = errors.New("nil DRWA rollout verification sink")
	errDRWANilRolloutPreflightReport   = errors.New("nil DRWA rollout preflight report")
	errDRWANilRolloutMetrics           = errors.New("nil DRWA rollout metrics")
	errDRWANilRolloutVerification      = errors.New("nil DRWA rollout verification report")
	errDRWARolloutVerificationRequired = errors.New("DRWA rollout verification required")
	errDRWARolloutVerificationRejected = errors.New("DRWA rollout verification rejected")
	errDRWARolloutReportTokenMismatch  = errors.New("DRWA rollout report token mismatch")
	errDRWARolloutReportStageMismatch  = errors.New("DRWA rollout report stage mismatch")
	errDRWAInvalidRolloutThresholds    = errors.New("invalid DRWA rollout thresholds")
)

const (
	drwaRolloutStageCanary     = "canary"
	drwaRolloutStageLimited    = "limited"
	drwaRolloutStageProduction = "production"
	drwaRolloutMaxFailureRateBpsUpperBound  = 100
	drwaRolloutMaxAPIErrorRateBpsUpperBound = 100
	drwaRolloutMaxDenialMismatchUpperBound  = 10
)

type drwaRolloutManifest struct {
	TokenID                string   `json:"token_id"`
	Issuer                 string   `json:"issuer"`
	Stage                  string   `json:"stage"`
	ExpectedPolicyVersion  uint64   `json:"expected_policy_version"`
	RequiredHolders        []string `json:"required_holders"`
	AllowSafeModeRollout   bool     `json:"allow_safe_mode_rollout"`
	MaxSyncFailureRateBps  uint64   `json:"max_sync_failure_rate_bps"`
	MaxAPIErrorRateBps     uint64   `json:"max_api_error_rate_bps"`
	MaxDenialMismatchCount uint64   `json:"max_denial_mismatch_count"`
}

type drwaRolloutStateReader interface {
	GetTokenPolicyStored(tokenID string) (*drwaSyncStoredValue, error)
	GetHolderMirrorStored(tokenID, holder string) (*drwaSyncStoredValue, error)
}

type drwaRolloutPreflightReport struct {
	TokenID            string   `json:"token_id"`
	Stage              string   `json:"stage"`
	Checks             []string `json:"checks"`
	MissingHolders     []string `json:"missing_holders"`
	Warnings           []string `json:"warnings"`
	Ready              bool     `json:"ready"`
	RequiresCanaryOnly bool     `json:"requires_canary_only"`
}

type drwaRolloutObservedMetrics struct {
	SyncFailureRateBps  uint64 `json:"sync_failure_rate_bps"`
	APIErrorRateBps     uint64 `json:"api_error_rate_bps"`
	DenialMismatchCount uint64 `json:"denial_mismatch_count"`
}

type drwaRolloutVerificationReport struct {
	TokenID      string   `json:"token_id"`
	Stage        string   `json:"stage"`
	PassedChecks []string `json:"passed_checks"`
	FailedChecks []string `json:"failed_checks"`
	Accepted     bool     `json:"accepted"`
}

type drwaRolloutEvidenceSink interface {
	PersistRolloutEvidence(tokenID string, payload []byte) error
}

type drwaRolloutVerificationSink interface {
	PersistRolloutVerification(tokenID string, payload []byte) error
}

func validateDRWARolloutManifest(manifest *drwaRolloutManifest) error {
	if manifest == nil {
		return errDRWANilRolloutManifest
	}
	if strings.TrimSpace(manifest.TokenID) == "" {
		return errDRWAEmptyRolloutTokenID
	}
	if strings.TrimSpace(manifest.Issuer) == "" {
		return errDRWAEmptyRolloutIssuer
	}
	switch manifest.Stage {
	case drwaRolloutStageCanary, drwaRolloutStageLimited, drwaRolloutStageProduction:
	default:
		return errDRWAInvalidRolloutStage
	}
	if manifest.MaxSyncFailureRateBps == 0 || manifest.MaxAPIErrorRateBps == 0 || manifest.MaxDenialMismatchCount == 0 {
		return errDRWAInvalidRolloutThresholds
	}
	if manifest.MaxSyncFailureRateBps > drwaRolloutMaxFailureRateBpsUpperBound ||
		manifest.MaxAPIErrorRateBps > drwaRolloutMaxAPIErrorRateBpsUpperBound ||
		manifest.MaxDenialMismatchCount > drwaRolloutMaxDenialMismatchUpperBound {
		return errDRWAInvalidRolloutThresholds
	}

	return nil
}

func inspectDRWARolloutPreflight(reader drwaRolloutStateReader, manifest *drwaRolloutManifest, recoveryReport *drwaRecoveryReport) (*drwaRolloutPreflightReport, error) {
	if reader == nil {
		return nil, errDRWANilRolloutStateReader
	}

	err := validateDRWARolloutManifest(manifest)
	if err != nil {
		return nil, err
	}

	report := &drwaRolloutPreflightReport{
		TokenID: manifest.TokenID,
		Stage:   manifest.Stage,
	}

	policy, err := reader.GetTokenPolicyStored(manifest.TokenID)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, errDRWARolloutPolicyMissing
	}
	if policy.Version != manifest.ExpectedPolicyVersion {
		report.Warnings = append(report.Warnings, fmt.Sprintf("policy version mismatch: expected=%d observed=%d", manifest.ExpectedPolicyVersion, policy.Version))
		report.RequiresCanaryOnly = true
	}
	report.Checks = append(report.Checks, "token policy present")

	for _, holder := range manifest.RequiredHolders {
		holder = strings.TrimSpace(holder)
		if holder == "" {
			continue
		}
		stored, readErr := reader.GetHolderMirrorStored(manifest.TokenID, holder)
		if readErr != nil {
			return nil, readErr
		}
		if stored == nil {
			report.MissingHolders = append(report.MissingHolders, holder)
			continue
		}
		report.Checks = append(report.Checks, "holder mirror present:"+holder)
	}

	if recoveryReport != nil && recoveryReport.RequiresSafeMode {
		if !manifest.AllowSafeModeRollout {
			return nil, errDRWARolloutSafeModeRequired
		}
		report.Warnings = append(report.Warnings, "safe mode still required by latest recovery report")
	}

	if manifest.MaxSyncFailureRateBps > 0 {
		report.Checks = append(report.Checks, fmt.Sprintf("sync failure threshold configured:%d", manifest.MaxSyncFailureRateBps))
	}
	if manifest.MaxAPIErrorRateBps > 0 {
		report.Checks = append(report.Checks, fmt.Sprintf("api error threshold configured:%d", manifest.MaxAPIErrorRateBps))
	}
	if manifest.MaxDenialMismatchCount > 0 {
		report.Checks = append(report.Checks, fmt.Sprintf("denial mismatch threshold configured:%d", manifest.MaxDenialMismatchCount))
	}

	switch manifest.Stage {
	case drwaRolloutStageCanary:
		report.Ready = len(report.MissingHolders) == 0
	case drwaRolloutStageLimited, drwaRolloutStageProduction:
		if report.RequiresCanaryOnly {
			return nil, errDRWARolloutNotCanaryReady
		}
		report.Ready = len(report.MissingHolders) == 0
	}

	sort.Strings(report.Checks)
	sort.Strings(report.MissingHolders)
	sort.Strings(report.Warnings)

	return report, nil
}

func marshalDRWARolloutEvidence(report *drwaRolloutPreflightReport) ([]byte, error) {
	if report == nil {
		return nil, errDRWANilRolloutPreflightReport
	}

	payload := map[string]interface{}{
		"token_id":             report.TokenID,
		"stage":                report.Stage,
		"checks":               report.Checks,
		"missing_holders":      report.MissingHolders,
		"warnings":             report.Warnings,
		"ready":                report.Ready,
		"requires_canary_only": report.RequiresCanaryOnly,
	}

	return json.MarshalIndent(payload, "", "  ")
}

func persistDRWARolloutEvidence(sink drwaRolloutEvidenceSink, tokenID string, payload []byte) error {
	if sink == nil {
		return errDRWANilRolloutEvidenceSink
	}

	return sink.PersistRolloutEvidence(tokenID, payload)
}

func buildDRWARolloutVerificationReport(manifest *drwaRolloutManifest, metrics *drwaRolloutObservedMetrics) (*drwaRolloutVerificationReport, error) {
	if metrics == nil {
		return nil, errDRWANilRolloutMetrics
	}
	err := validateDRWARolloutManifest(manifest)
	if err != nil {
		return nil, err
	}

	report := &drwaRolloutVerificationReport{
		TokenID: manifest.TokenID,
		Stage:   manifest.Stage,
	}

	if metrics.SyncFailureRateBps <= manifest.MaxSyncFailureRateBps {
		report.PassedChecks = append(report.PassedChecks, "sync failure rate within threshold")
	} else {
		report.FailedChecks = append(report.FailedChecks, fmt.Sprintf("sync failure rate exceeded: observed=%d threshold=%d", metrics.SyncFailureRateBps, manifest.MaxSyncFailureRateBps))
	}

	if metrics.APIErrorRateBps <= manifest.MaxAPIErrorRateBps {
		report.PassedChecks = append(report.PassedChecks, "api error rate within threshold")
	} else {
		report.FailedChecks = append(report.FailedChecks, fmt.Sprintf("api error rate exceeded: observed=%d threshold=%d", metrics.APIErrorRateBps, manifest.MaxAPIErrorRateBps))
	}

	if metrics.DenialMismatchCount <= manifest.MaxDenialMismatchCount {
		report.PassedChecks = append(report.PassedChecks, "denial mismatch count within threshold")
	} else {
		report.FailedChecks = append(report.FailedChecks, fmt.Sprintf("denial mismatch count exceeded: observed=%d threshold=%d", metrics.DenialMismatchCount, manifest.MaxDenialMismatchCount))
	}

	sort.Strings(report.PassedChecks)
	sort.Strings(report.FailedChecks)
	report.Accepted = len(report.FailedChecks) == 0
	if report.Accepted {
		recordDRWAMetric(drwaMetricRolloutVerificationPass)
	} else {
		recordDRWAMetric(drwaMetricRolloutVerificationReject)
	}

	return report, nil
}

func marshalDRWARolloutVerificationReport(report *drwaRolloutVerificationReport) ([]byte, error) {
	if report == nil {
		return nil, errDRWANilRolloutVerification
	}

	return json.MarshalIndent(report, "", "  ")
}

func persistDRWARolloutVerification(sink drwaRolloutVerificationSink, tokenID string, payload []byte) error {
	if sink == nil {
		return errDRWANilRolloutVerificationSink
	}

	return sink.PersistRolloutVerification(tokenID, payload)
}

func validateDRWARolloutAdmission(
	manifest *drwaRolloutManifest,
	preflight *drwaRolloutPreflightReport,
	verification *drwaRolloutVerificationReport,
) error {
	err := validateDRWARolloutManifest(manifest)
	if err != nil {
		return err
	}
	if preflight == nil {
		return errDRWANilRolloutPreflightReport
	}
	if preflight.TokenID != manifest.TokenID {
		return errDRWARolloutReportTokenMismatch
	}
	if preflight.Stage != manifest.Stage {
		return errDRWARolloutReportStageMismatch
	}
	if !preflight.Ready {
		return errDRWARolloutNotCanaryReady
	}

	switch manifest.Stage {
	case drwaRolloutStageCanary:
		return nil
	case drwaRolloutStageLimited, drwaRolloutStageProduction:
		if verification == nil {
			return errDRWARolloutVerificationRequired
		}
		if verification.TokenID != manifest.TokenID {
			return errDRWARolloutReportTokenMismatch
		}
		if verification.Stage != manifest.Stage {
			return errDRWARolloutReportStageMismatch
		}
		if !verification.Accepted {
			return errDRWARolloutVerificationRejected
		}
		return nil
	default:
		return errDRWAInvalidRolloutStage
	}
}
