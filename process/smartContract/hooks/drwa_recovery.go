package hooks

import (
	"bytes"
	"encoding/json"
	"errors"
	"sort"
	"strings"
)

var (
	errDRWANilRecoveryManifest     = errors.New("nil DRWA recovery manifest")
	errDRWANilRecoveryStateReader  = errors.New("nil DRWA recovery state reader")
	errDRWANilRecoveryReport       = errors.New("nil DRWA recovery report")
	errDRWARecoveryNotRepairable   = errors.New("DRWA recovery report not repairable")
	errDRWARecoveryEnumeratorError = errors.New("DRWA recovery mirror enumeration failed")
	errDRWANilRecoveryEvidenceSink = errors.New("nil DRWA recovery evidence sink")
	errDRWARecoveryUnexpectedHolderConflict = errors.New("DRWA recovery report contains unexpected holder also present in manifest")
)

const (
	drwaRecoveryStatusInSync                  = "in_sync"
	drwaRecoveryStatusTokenPolicyMissing      = "token_policy_missing"
	drwaRecoveryStatusTokenPolicyCorrupt      = "token_policy_corrupt"
	drwaRecoveryStatusTokenPolicyVersionDrift = "token_policy_version_drift"
	drwaRecoveryStatusTokenPolicyBodyDrift    = "token_policy_body_drift"
	drwaRecoveryStatusHolderMirrorMissing     = "holder_mirror_missing"
	drwaRecoveryStatusHolderMirrorCorrupt     = "holder_mirror_corrupt"
	drwaRecoveryStatusHolderVersionDrift      = "holder_mirror_version_drift"
	drwaRecoveryStatusHolderBodyDrift         = "holder_mirror_body_drift"
	drwaRecoveryStatusUnexpectedHolderMirror  = "unexpected_holder_mirror"
)

type drwaRecoveryManifest = drwaMigrationManifest
type drwaRecoveryHolder = drwaMigrationHolder

type drwaRecoveryFinding struct {
	Status          string `json:"status"`
	TokenID         string `json:"token_id"`
	Holder          string `json:"holder,omitempty"`
	ExpectedVersion uint64 `json:"expected_version"`
	ObservedVersion uint64 `json:"observed_version"`
}

type drwaRecoveryReport struct {
	TokenID          string                `json:"token_id"`
	Findings         []drwaRecoveryFinding `json:"findings"`
	RequiresSafeMode bool                  `json:"requires_safe_mode"`
	Repairable       bool                  `json:"repairable"`
}

type drwaRecoveryMirrorEnumerator interface {
	ListHolderMirrorAddresses(tokenID string) ([]string, error)
}

type drwaRecoveryEvidenceSink interface {
	PersistRecoveryEvidence(tokenID string, payload []byte) error
}

func validateDRWARecoveryManifest(manifest *drwaRecoveryManifest) error {
	if manifest == nil {
		return errDRWANilRecoveryManifest
	}

	return validateDRWAMigrationManifest((*drwaMigrationManifest)(manifest))
}

func inspectDRWARecoveryState(reader drwaMigrationStateReader, manifest *drwaRecoveryManifest) (*drwaRecoveryReport, error) {
	if reader == nil {
		return nil, errDRWANilRecoveryStateReader
	}

	err := validateDRWARecoveryManifest(manifest)
	if err != nil {
		return nil, err
	}

	report := &drwaRecoveryReport{
		TokenID:    manifest.TokenID,
		Repairable: true,
	}

	tokenPolicy, err := reader.GetTokenPolicyStored(manifest.TokenID)
	if err != nil {
		return nil, err
	}

	switch {
	case tokenPolicy == nil:
		report.Findings = append(report.Findings, drwaRecoveryFinding{
			Status:          drwaRecoveryStatusTokenPolicyMissing,
			TokenID:         manifest.TokenID,
			ExpectedVersion: manifest.PolicyVersion,
		})
		report.RequiresSafeMode = true
	case tokenPolicy.Version > 0 && len(tokenPolicy.Body) == 0:
		report.Findings = append(report.Findings, drwaRecoveryFinding{
			Status:          drwaRecoveryStatusTokenPolicyCorrupt,
			TokenID:         manifest.TokenID,
			ExpectedVersion: manifest.PolicyVersion,
			ObservedVersion: tokenPolicy.Version,
		})
		report.RequiresSafeMode = true
		report.Repairable = false
	case tokenPolicy.Version != manifest.PolicyVersion:
		report.Findings = append(report.Findings, drwaRecoveryFinding{
			Status:          drwaRecoveryStatusTokenPolicyVersionDrift,
			TokenID:         manifest.TokenID,
			ExpectedVersion: manifest.PolicyVersion,
			ObservedVersion: tokenPolicy.Version,
		})
		report.RequiresSafeMode = true
	case !bytes.Equal(tokenPolicy.Body, manifest.PolicyBody):
		report.Findings = append(report.Findings, drwaRecoveryFinding{
			Status:          drwaRecoveryStatusTokenPolicyBodyDrift,
			TokenID:         manifest.TokenID,
			ExpectedVersion: manifest.PolicyVersion,
			ObservedVersion: tokenPolicy.Version,
		})
		report.RequiresSafeMode = true
	}

	for _, holder := range manifest.Holders {
		stored, readErr := reader.GetHolderMirrorStored(manifest.TokenID, holder.Address)
		if readErr != nil {
			return nil, readErr
		}

		switch {
		case stored == nil:
			report.Findings = append(report.Findings, drwaRecoveryFinding{
				Status:          drwaRecoveryStatusHolderMirrorMissing,
				TokenID:         manifest.TokenID,
				Holder:          holder.Address,
				ExpectedVersion: holder.Version,
			})
		case stored.Version > 0 && len(stored.Body) == 0:
			report.Findings = append(report.Findings, drwaRecoveryFinding{
				Status:          drwaRecoveryStatusHolderMirrorCorrupt,
				TokenID:         manifest.TokenID,
				Holder:          holder.Address,
				ExpectedVersion: holder.Version,
				ObservedVersion: stored.Version,
			})
			report.Repairable = false
		case stored.Version != holder.Version:
			report.Findings = append(report.Findings, drwaRecoveryFinding{
				Status:          drwaRecoveryStatusHolderVersionDrift,
				TokenID:         manifest.TokenID,
				Holder:          holder.Address,
				ExpectedVersion: holder.Version,
				ObservedVersion: stored.Version,
			})
		case !bytes.Equal(stored.Body, holder.Body):
			report.Findings = append(report.Findings, drwaRecoveryFinding{
				Status:          drwaRecoveryStatusHolderBodyDrift,
				TokenID:         manifest.TokenID,
				Holder:          holder.Address,
				ExpectedVersion: holder.Version,
				ObservedVersion: stored.Version,
			})
		}
	}

	enumerator, ok := reader.(drwaRecoveryMirrorEnumerator)
	if ok {
		addresses, listErr := enumerator.ListHolderMirrorAddresses(manifest.TokenID)
		if listErr != nil {
			return nil, errors.Join(errDRWARecoveryEnumeratorError, listErr)
		}

		expected := make(map[string]struct{}, len(manifest.Holders))
		for _, holder := range manifest.Holders {
			expected[holder.Address] = struct{}{}
		}

		for _, address := range addresses {
			if _, exists := expected[address]; exists {
				continue
			}

			stored, readErr := reader.GetHolderMirrorStored(manifest.TokenID, address)
			if readErr != nil {
				return nil, readErr
			}

			finding := drwaRecoveryFinding{
				Status:  drwaRecoveryStatusUnexpectedHolderMirror,
				TokenID: manifest.TokenID,
				Holder:  address,
			}
			if stored != nil {
				finding.ObservedVersion = stored.Version
			}

			report.Findings = append(report.Findings, finding)
		}
	}

	if len(report.Findings) == 0 {
		report.Findings = append(report.Findings, drwaRecoveryFinding{
			Status:          drwaRecoveryStatusInSync,
			TokenID:         manifest.TokenID,
			ExpectedVersion: manifest.PolicyVersion,
			ObservedVersion: manifest.PolicyVersion,
		})
		report.RequiresSafeMode = false
	}

	sort.Slice(report.Findings, func(i, j int) bool {
		left := report.Findings[i].Status + "|" + report.Findings[i].Holder
		right := report.Findings[j].Status + "|" + report.Findings[j].Holder
		return left < right
	})

	if report.RequiresSafeMode {
		recordDRWAMetric(drwaMetricRecoverySafeModeReport)
	}
	if !report.Repairable {
		recordDRWAMetric(drwaMetricRecoveryNonRepairable)
	}

	return report, nil
}

func buildDRWARecoveryEnvelope(manifest *drwaRecoveryManifest, report *drwaRecoveryReport) (*drwaSyncEnvelope, error) {
	if report == nil {
		return nil, errDRWANilRecoveryReport
	}
	if !report.Repairable {
		return nil, errDRWARecoveryNotRepairable
	}

	if len(report.Findings) == 1 && report.Findings[0].Status == drwaRecoveryStatusInSync {
		payloadHash, err := computeDRWASyncHash(drwaSyncCallerRecoveryAdmin, nil)
		if err != nil {
			return nil, err
		}
		return &drwaSyncEnvelope{
			CallerDomain: drwaSyncCallerRecoveryAdmin,
			PayloadHash:  payloadHash,
			Operations:   nil,
			Noop:         true,
		}, nil
	}
	operations := make([]drwaSyncOperation, 0, 1+len(manifest.Holders)+len(report.Findings))
	repairEnvelope, err := buildDRWAMigrationEnvelope((*drwaMigrationManifest)(manifest))
	if err != nil {
		return nil, err
	}
	operations = append(operations, repairEnvelope.Operations...)

	for _, finding := range report.Findings {
		if finding.Status != drwaRecoveryStatusUnexpectedHolderMirror {
			continue
		}
		for _, holder := range manifest.Holders {
			if holder.Address == finding.Holder {
				return nil, errDRWARecoveryUnexpectedHolderConflict
			}
		}
		nextVersion := finding.ObservedVersion + 1
		if nextVersion == 0 {
			nextVersion = 1
		}
		operations = append(operations, drwaSyncOperation{
			OperationType: drwaSyncOpHolderMirrorDelete,
			TokenID:       finding.TokenID,
			Holder:        finding.Holder,
			Version:       nextVersion,
		})
	}

	sort.SliceStable(operations, func(i, j int) bool {
		return drwaRecoveryOperationOrder(operations[i]) < drwaRecoveryOperationOrder(operations[j])
	})

	hash, err := computeDRWASyncHash(drwaSyncCallerRecoveryAdmin, operations)
	if err != nil {
		return nil, err
	}

	return &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerRecoveryAdmin,
		PayloadHash:  hash,
		Operations:   operations,
	}, nil
}

func sanitizeDRWARecoveryManifest(manifest *drwaRecoveryManifest) {
	if manifest == nil {
		return
	}

	manifest.TokenID = strings.TrimSpace(manifest.TokenID)
	for idx := range manifest.Holders {
		manifest.Holders[idx].Address = strings.TrimSpace(manifest.Holders[idx].Address)
	}
}

func marshalDRWARecoveryEvidence(report *drwaRecoveryReport) ([]byte, error) {
	if report == nil {
		return nil, errDRWANilRecoveryReport
	}

	return json.MarshalIndent(report, "", "  ")
}

func persistDRWARecoveryEvidence(sink drwaRecoveryEvidenceSink, report *drwaRecoveryReport) error {
	if sink == nil {
		return errDRWANilRecoveryEvidenceSink
	}

	payload, err := marshalDRWARecoveryEvidence(report)
	if err != nil {
		return err
	}

	return sink.PersistRecoveryEvidence(report.TokenID, payload)
}

func drwaRecoveryOperationOrder(operation drwaSyncOperation) string {
	switch operation.OperationType {
	case drwaSyncOpTokenPolicy:
		return "0|"
	case drwaSyncOpHolderMirror:
		return "1|" + operation.Holder
	case drwaSyncOpHolderMirrorDelete:
		return "2|" + operation.Holder
	default:
		return "9|" + operation.Holder
	}
}
