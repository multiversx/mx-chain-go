package hooks

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

func TestValidateDRWAMigrationManifestRejectsDuplicateHolder(t *testing.T) {
	t.Parallel()

	err := validateDRWAMigrationManifest(&drwaMigrationManifest{
		TokenID:       "CARBON-1",
		PolicyVersion: 1,
		PolicyBody:    []byte(`{"drwa_enabled":true}`),
		Holders: []drwaMigrationHolder{
			{Address: "erd1holder", Version: 1, Body: []byte(`{}`)},
			{Address: "erd1holder", Version: 2, Body: []byte(`{}`)},
		},
	})

	if err == nil {
		t.Fatalf("expected duplicate holder rejection")
	}
}

func TestBuildDRWAMigrationEnvelopeOrdersPolicyThenSortedHolders(t *testing.T) {
	t.Parallel()

	envelope, err := buildDRWAMigrationEnvelope(&drwaMigrationManifest{
		TokenID:       "CARBON-1",
		PolicyVersion: 7,
		PolicyBody:    []byte(`{"drwa_enabled":true}`),
		Holders: []drwaMigrationHolder{
			{Address: "erd1z", Version: 3, Body: []byte(`{"kyc":"approved"}`)},
			{Address: "erd1a", Version: 2, Body: []byte(`{"kyc":"approved"}`)},
		},
	})
	if err != nil {
		t.Fatalf("build envelope: %v", err)
	}

	if envelope.CallerDomain != drwaSyncCallerRecoveryAdmin {
		t.Fatalf("unexpected caller domain: %s", envelope.CallerDomain)
	}
	if len(envelope.Operations) != 3 {
		t.Fatalf("unexpected operations len: %d", len(envelope.Operations))
	}
	if envelope.Operations[0].OperationType != drwaSyncOpTokenPolicy {
		t.Fatalf("expected token policy first")
	}
	if envelope.Operations[1].Holder != "erd1a" || envelope.Operations[2].Holder != "erd1z" {
		t.Fatalf("holders not sorted: %+v", envelope.Operations)
	}
	if len(envelope.PayloadHash) == 0 {
		t.Fatalf("expected payload hash")
	}
}

func TestCaptureDRWAMigrationSnapshotAndBuildRollbackEnvelope(t *testing.T) {
	t.Parallel()

	adapter := newMockDRWASyncStateAdapter()
	adapter.tokenVersions["CARBON-1"] = 5
	adapter.tokenBodies["CARBON-1"] = []byte(`{"drwa_enabled":false}`)
	adapter.holderVersions["CARBON-1|erd1a"] = 4
	adapter.holderBodies["CARBON-1|erd1a"] = []byte(`{"kyc":"approved"}`)

	manifest := &drwaMigrationManifest{
		TokenID:       "CARBON-1",
		PolicyVersion: 6,
		PolicyBody:    []byte(`{"drwa_enabled":true}`),
		Holders: []drwaMigrationHolder{
			{Address: "erd1a", Version: 5, Body: []byte(`{"kyc":"approved"}`)},
			{Address: "erd1b", Version: 1, Body: []byte(`{"kyc":"approved"}`)},
		},
	}

	snapshot, err := captureDRWAMigrationSnapshot(adapter, manifest)
	if err != nil {
		t.Fatalf("capture snapshot: %v", err)
	}

	if snapshot.TokenPolicy == nil || snapshot.TokenPolicy.Version != 5 {
		t.Fatalf("unexpected token policy snapshot: %+v", snapshot.TokenPolicy)
	}
	if snapshot.Holders["erd1a"] == nil || snapshot.Holders["erd1a"].Version != 4 {
		t.Fatalf("unexpected holder snapshot: %+v", snapshot.Holders["erd1a"])
	}
	if snapshot.Holders["erd1b"] != nil {
		t.Fatalf("expected nil snapshot for new holder")
	}

	rollbackEnvelope, err := buildDRWARollbackEnvelope(snapshot)
	if err != nil {
		t.Fatalf("build rollback envelope: %v", err)
	}

	if len(rollbackEnvelope.Operations) != 3 {
		t.Fatalf("unexpected rollback operations: %d", len(rollbackEnvelope.Operations))
	}
	if rollbackEnvelope.Operations[0].Version != 5 {
		t.Fatalf("unexpected rollback token version: %d", rollbackEnvelope.Operations[0].Version)
	}
	if rollbackEnvelope.Operations[1].Holder != "erd1a" {
		t.Fatalf("unexpected rollback holder: %s", rollbackEnvelope.Operations[1].Holder)
	}
	if !bytes.Equal(rollbackEnvelope.Operations[1].Body, []byte(`{"kyc":"approved"}`)) {
		t.Fatalf("unexpected rollback body: %s", string(rollbackEnvelope.Operations[1].Body))
	}
	if rollbackEnvelope.Operations[2].OperationType != drwaSyncOpHolderMirrorDelete {
		t.Fatalf("expected delete operation for new holder, got %+v", rollbackEnvelope.Operations[2])
	}
	if rollbackEnvelope.Operations[2].Holder != "erd1b" {
		t.Fatalf("unexpected rollback delete holder: %s", rollbackEnvelope.Operations[2].Holder)
	}
	if rollbackEnvelope.Operations[2].Version != 2 {
		t.Fatalf("unexpected rollback delete version: %d", rollbackEnvelope.Operations[2].Version)
	}
}

func TestPersistDRWAMigrationAuthorizedCallers(t *testing.T) {
	t.Parallel()

	adapter := newMockDRWASyncStateAdapter()
	policyRaw := strings.Repeat("p", 32)
	assetRaw := strings.Repeat("a", 32)
	recoveryRaw := strings.Repeat("r", 32)
	manifest := &drwaMigrationManifest{
		TokenID:       "CARBON-1",
		PolicyVersion: 1,
		PolicyBody:    []byte(`{"drwa_enabled":true}`),
		AuthorizedCallers: map[string]string{
			drwaSyncCallerPolicyRegistry: policyRaw,
			drwaSyncCallerAssetManager:   assetRaw,
			drwaSyncCallerRecoveryAdmin:  hex.EncodeToString([]byte(recoveryRaw)),
		},
	}

	err := persistDRWAMigrationAuthorizedCallers(adapter, manifest)
	if err != nil {
		t.Fatalf("persist authorized callers: %v", err)
	}

	if got := string(adapter.authorizedCallers[drwaSyncCallerPolicyRegistry]); got != policyRaw {
		t.Fatalf("unexpected policy caller: %s", got)
	}
	if got := string(adapter.authorizedCallers[drwaSyncCallerAssetManager]); got != assetRaw {
		t.Fatalf("unexpected asset caller: %s", got)
	}
	if got := string(adapter.authorizedCallers[drwaSyncCallerRecoveryAdmin]); got != recoveryRaw {
		t.Fatalf("unexpected recovery caller: %q", got)
	}
}

func TestValidateDRWAMigrationManifestRejectsInvalidAuthorizedCallerFormat(t *testing.T) {
	t.Parallel()

	err := validateDRWAMigrationManifest(&drwaMigrationManifest{
		TokenID:       "CARBON-1",
		PolicyVersion: 1,
		PolicyBody:    []byte(`{"drwa_enabled":true}`),
		AuthorizedCallers: map[string]string{
			drwaSyncCallerPolicyRegistry: "erd1policy",
			drwaSyncCallerAssetManager:   strings.Repeat("a", 32),
			drwaSyncCallerRecoveryAdmin:  strings.Repeat("r", 32),
		},
	})

	if err == nil {
		t.Fatalf("expected invalid authorized caller format rejection")
	}
}

func TestBuildDRWAAuthorizedCallerKeyUsesDedicatedAuthPrefix(t *testing.T) {
	t.Parallel()

	if got := string(buildDRWAAuthorizedCallerKey(drwaSyncCallerPolicyRegistry)); got != "drwa:auth:"+drwaSyncCallerPolicyRegistry {
		t.Fatalf("unexpected authorized caller key: %s", got)
	}
}
