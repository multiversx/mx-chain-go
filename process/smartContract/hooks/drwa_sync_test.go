package hooks

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
	teststate "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type mockDRWASyncStateAdapter struct {
	tokenVersions       map[string]uint64
	holderVersions      map[string]uint64
	tokenBodies         map[string][]byte
	holderBodies        map[string][]byte
	authorizedCallers   map[string][]byte
	recoveryEvidence    map[string][]byte
	rolloutEvidence     map[string][]byte
	rolloutVerification map[string][]byte
	holderIndex         map[string][]string
	snapshots           int
	failPut             bool
	rolledBack          bool
	putHolderHook       func(tokenID, holder string, version uint64, body []byte) error
}

func newMockDRWASyncStateAdapter() *mockDRWASyncStateAdapter {
	return &mockDRWASyncStateAdapter{
		tokenVersions:  make(map[string]uint64),
		holderVersions: make(map[string]uint64),
		tokenBodies:    make(map[string][]byte),
		holderBodies:   make(map[string][]byte),
		authorizedCallers: map[string][]byte{
			drwaSyncCallerPolicyRegistry: []byte("policy_registry"),
			drwaSyncCallerAssetManager:   []byte("asset_manager"),
			drwaSyncCallerRecoveryAdmin:  []byte("recovery_admin"),
		},
		recoveryEvidence:    make(map[string][]byte),
		rolloutEvidence:     make(map[string][]byte),
		rolloutVerification: make(map[string][]byte),
		holderIndex:         make(map[string][]string),
	}
}

func (m *mockDRWASyncStateAdapter) GetTokenPolicyVersion(tokenID string) (uint64, error) {
	return m.tokenVersions[tokenID], nil
}

func (m *mockDRWASyncStateAdapter) GetHolderMirrorVersion(tokenID, holder string) (uint64, error) {
	return m.holderVersions[tokenID+"|"+holder], nil
}

func (m *mockDRWASyncStateAdapter) GetAuthorizedCallerAddress(domain string) ([]byte, error) {
	return append([]byte(nil), m.authorizedCallers[domain]...), nil
}

func (m *mockDRWASyncStateAdapter) PutAuthorizedCallerAddress(domain string, address []byte) error {
	m.authorizedCallers[domain] = append([]byte(nil), address...)
	return nil
}

func (m *mockDRWASyncStateAdapter) GetTokenPolicyStored(tokenID string) (*drwaSyncStoredValue, error) {
	version, ok := m.tokenVersions[tokenID]
	if !ok {
		return nil, nil
	}
	return &drwaSyncStoredValue{
		Version: version,
		Body:    m.tokenBodies[tokenID],
	}, nil
}

func (m *mockDRWASyncStateAdapter) GetHolderMirrorStored(tokenID, holder string) (*drwaSyncStoredValue, error) {
	key := tokenID + "|" + holder
	version, ok := m.holderVersions[key]
	if !ok {
		return nil, nil
	}
	return &drwaSyncStoredValue{
		Version: version,
		Body:    m.holderBodies[key],
	}, nil
}

func (m *mockDRWASyncStateAdapter) PutTokenPolicyBody(tokenID string, version uint64, body []byte) error {
	if m.failPut {
		return errDRWATestFailPut
	}
	m.tokenVersions[tokenID] = version
	m.tokenBodies[tokenID] = body
	return nil
}

func (m *mockDRWASyncStateAdapter) PutHolderMirrorBody(tokenID, holder string, version uint64, body []byte) error {
	if m.putHolderHook != nil {
		return m.putHolderHook(tokenID, holder, version, body)
	}
	if m.failPut {
		return errDRWATestFailPut
	}
	key := tokenID + "|" + holder
	m.holderVersions[key] = version
	m.holderBodies[key] = body
	m.ensureHolderIndexed(tokenID, holder)
	return nil
}

func (m *mockDRWASyncStateAdapter) DeleteHolderMirror(tokenID, holder string, version uint64) error {
	if m.failPut {
		return errDRWATestFailPut
	}
	key := tokenID + "|" + holder
	delete(m.holderVersions, key)
	delete(m.holderBodies, key)
	m.removeHolderIndexed(tokenID, holder)
	return nil
}

func (m *mockDRWASyncStateAdapter) PersistRecoveryEvidence(tokenID string, payload []byte) error {
	if m.failPut {
		return errDRWATestFailPut
	}
	m.recoveryEvidence[tokenID] = append([]byte(nil), payload...)
	return nil
}

func (m *mockDRWASyncStateAdapter) PersistRolloutEvidence(tokenID string, payload []byte) error {
	if m.failPut {
		return errDRWATestFailPut
	}
	m.rolloutEvidence[tokenID] = append([]byte(nil), payload...)
	return nil
}

func (m *mockDRWASyncStateAdapter) PersistRolloutVerification(tokenID string, payload []byte) error {
	if m.failPut {
		return errDRWATestFailPut
	}
	m.rolloutVerification[tokenID] = append([]byte(nil), payload...)
	return nil
}

func (m *mockDRWASyncStateAdapter) ListHolderMirrorAddresses(tokenID string) ([]string, error) {
	addresses := append([]string(nil), m.holderIndex[tokenID]...)
	return addresses, nil
}

func (m *mockDRWASyncStateAdapter) Snapshot() int {
	m.snapshots++
	return m.snapshots
}

func (m *mockDRWASyncStateAdapter) Rollback(snapshot int) error {
	m.rolledBack = snapshot > 0
	return nil
}

func (m *mockDRWASyncStateAdapter) ensureHolderIndexed(tokenID, holder string) {
	addresses := m.holderIndex[tokenID]
	for _, existing := range addresses {
		if existing == holder {
			return
		}
	}
	m.holderIndex[tokenID] = append(addresses, holder)
}

func (m *mockDRWASyncStateAdapter) removeHolderIndexed(tokenID, holder string) {
	addresses := m.holderIndex[tokenID]
	filtered := addresses[:0]
	for _, existing := range addresses {
		if existing != holder {
			filtered = append(filtered, existing)
		}
	}
	m.holderIndex[tokenID] = append([]string(nil), filtered...)
}

const drwaTestMsgOneApplied = "expected 1 applied operation, got %d"

var errDRWATestFailPut = &drwaTestError{"fail put"}

type drwaTestError struct {
	msg string
}

func (e *drwaTestError) Error() string { return e.msg }

func TestDecodeDRWASyncEnvelopeRejectsOversizedPayload(t *testing.T) {
	payload := make([]byte, drwaSyncMaxPayloadBytes+1)

	_, err := decodeDRWASyncEnvelope(payload)
	if err == nil || err.Error() != drwaSyncRejectPayloadTooLarge {
		t.Fatalf("expected oversized payload rejection, got %v", err)
	}
}

func TestDecodeDRWASyncEnvelopeRejectsMalformedPayload(t *testing.T) {
	payload := []byte(`{"caller_domain":"asset_manager","operations":[`)

	_, err := decodeDRWASyncEnvelope(payload)
	if err == nil {
		t.Fatalf("expected malformed payload rejection")
	}
}

func TestApplyDRWASyncEnvelopeRejectsUnauthorizedCaller(t *testing.T) {
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
	if err == nil || err.Error() != drwaSyncRejectUnauthorizedCaller {
		t.Fatalf("expected unauthorized caller rejection, got %v", err)
	}
}

func TestSerializeDRWASyncEnvelopePayloadTokenPolicyUsesZeroAddressHolder(t *testing.T) {
	operations := []drwaSyncOperation{{
		OperationType: drwaSyncOpTokenPolicy,
		TokenID:       "CARBON-1",
		Version:       7,
		Body:          []byte{0xAA, 0xBB},
	}}

	payload, err := serializeDRWASyncEnvelopePayload(drwaSyncCallerPolicyRegistry, operations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := bytes.NewBuffer(nil)
	expected.WriteByte(0)
	expected.WriteByte(0)
	writeLengthPrefixedTest(expected, []byte("CARBON-1"))
	writeLengthPrefixedTest(expected, make([]byte, 32))
	versionBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(versionBytes, 7)
	expected.Write(versionBytes)
	writeLengthPrefixedTest(expected, []byte{0xAA, 0xBB})

	if !bytes.Equal(expected.Bytes(), payload) {
		t.Fatalf("unexpected payload bytes: %x", payload)
	}
}

func TestSerializeDRWASyncEnvelopePayloadHolderMirrorUsesHolderBytes(t *testing.T) {
	operations := []drwaSyncOperation{{
		OperationType: drwaSyncOpHolderMirror,
		TokenID:       "CARBON-1",
		Holder:        "holder-address",
		Version:       2,
		Body:          []byte{0x01},
	}}

	payload, err := serializeDRWASyncEnvelopePayload(drwaSyncCallerAssetManager, operations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := bytes.NewBuffer(nil)
	expected.WriteByte(1)
	expected.WriteByte(1)
	writeLengthPrefixedTest(expected, []byte("CARBON-1"))
	writeLengthPrefixedTest(expected, []byte("holder-address"))
	versionBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(versionBytes, 2)
	expected.Write(versionBytes)
	writeLengthPrefixedTest(expected, []byte{0x01})

	if !bytes.Equal(expected.Bytes(), payload) {
		t.Fatalf("unexpected payload bytes: %x", payload)
	}
}

func TestApplyDRWASyncEnvelopeRejectsHashMismatch(t *testing.T) {
	adapter := newMockDRWASyncStateAdapter()
	envelope := &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerPolicyRegistry,
		PayloadHash:  []byte("wrong-hash"),
		Operations: []drwaSyncOperation{{
			OperationType: drwaSyncOpTokenPolicy,
			TokenID:       "CARBON-1",
			Version:       1,
			Body:          []byte(`{}`),
		}},
	}

	_, err := applyDRWASyncEnvelope(adapter, envelope, 16, []byte("policy_registry"))
	if err == nil || err.Error() != drwaSyncRejectHashMismatch {
		t.Fatalf("expected hash mismatch rejection, got %v", err)
	}
}

func writeLengthPrefixedTest(buffer *bytes.Buffer, value []byte) {
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, uint32(len(value)))
	buffer.Write(lengthBytes)
	buffer.Write(value)
}

func TestApplyDRWASyncEnvelopeRejectsCallerAddressMismatch(t *testing.T) {
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

	_, err := applyDRWASyncEnvelope(adapter, envelope, 16, []byte("asset_manager"))
	if err == nil || err.Error() != drwaSyncRejectUnauthorizedCaller {
		t.Fatalf("expected caller-address mismatch rejection, got %v", err)
	}
}

func TestComputeDRWASyncHashBindsCallerDomain(t *testing.T) {
	operations := []drwaSyncOperation{{
		OperationType: drwaSyncOpTokenPolicy,
		TokenID:       "CARBON-1",
		Version:       1,
		Body:          []byte(`{}`),
	}}

	left, err := computeDRWASyncHash(drwaSyncCallerPolicyRegistry, operations)
	if err != nil {
		t.Fatalf("policy hash: %v", err)
	}
	right, err := computeDRWASyncHash(drwaSyncCallerAssetManager, operations)
	if err != nil {
		t.Fatalf("asset hash: %v", err)
	}
	if string(left) == string(right) {
		t.Fatalf("expected caller domain to affect sync hash")
	}
}

func TestApplyDRWASyncEnvelopeRejectsStaleReplay(t *testing.T) {
	adapter := newMockDRWASyncStateAdapter()
	adapter.tokenVersions["CARBON-1"] = 7
	envelope := &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerPolicyRegistry,
		Operations: []drwaSyncOperation{{
			OperationType: drwaSyncOpTokenPolicy,
			TokenID:       "CARBON-1",
			Version:       6,
			Body:          []byte(`{}`),
		}},
	}
	hash, _ := computeDRWASyncHash(envelope.CallerDomain, envelope.Operations)
	envelope.PayloadHash = hash

	_, err := applyDRWASyncEnvelope(adapter, envelope, 16, []byte("policy_registry"))
	if err == nil || err.Error() != drwaSyncRejectReplayStale {
		t.Fatalf("expected stale replay rejection, got %v", err)
	}
}

func TestApplyDRWASyncEnvelopeRejectsEqualVersionConflict(t *testing.T) {
	adapter := newMockDRWASyncStateAdapter()
	adapter.tokenVersions["CARBON-1"] = 7
	envelope := &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerPolicyRegistry,
		Operations: []drwaSyncOperation{{
			OperationType: drwaSyncOpTokenPolicy,
			TokenID:       "CARBON-1",
			Version:       7,
			Body:          []byte(`{}`),
		}},
	}
	hash, _ := computeDRWASyncHash(envelope.CallerDomain, envelope.Operations)
	envelope.PayloadHash = hash

	_, err := applyDRWASyncEnvelope(adapter, envelope, 16, []byte("policy_registry"))
	if err == nil || err.Error() != drwaSyncRejectReplayConflict {
		t.Fatalf("expected conflict rejection, got %v", err)
	}
}

func TestApplyDRWASyncEnvelopeRejectsVersionSkip(t *testing.T) {
	adapter := newMockDRWASyncStateAdapter()
	adapter.tokenVersions["CARBON-1"] = 1
	envelope := &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerPolicyRegistry,
		Operations: []drwaSyncOperation{{
			OperationType: drwaSyncOpTokenPolicy,
			TokenID:       "CARBON-1",
			Version:       3,
			Body:          []byte(`{}`),
		}},
	}
	hash, _ := computeDRWASyncHash(envelope.CallerDomain, envelope.Operations)
	envelope.PayloadHash = hash

	_, err := applyDRWASyncEnvelope(adapter, envelope, 16, []byte("policy_registry"))
	if err == nil || err.Error() != drwaSyncRejectReplayConflict {
		t.Fatalf("expected version-skip rejection, got %v", err)
	}
}

func TestApplyDRWASyncEnvelopeRollsBackAtomically(t *testing.T) {
	adapter := newMockDRWASyncStateAdapter()
	adapter.failPut = true
	envelope := &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerAssetManager,
		Operations: []drwaSyncOperation{{
			OperationType: drwaSyncOpHolderMirror,
			TokenID:       "CARBON-1",
			Holder:        "erd1holder",
			Version:       1,
			Body:          []byte(`{}`),
		}},
	}
	hash, _ := computeDRWASyncHash(envelope.CallerDomain, envelope.Operations)
	envelope.PayloadHash = hash

	_, err := applyDRWASyncEnvelope(adapter, envelope, 16, []byte("asset_manager"))
	if err == nil {
		t.Fatalf("expected batch failure")
	}
	if !adapter.rolledBack {
		t.Fatalf("expected rollback on failed batch")
	}
}

func TestApplyDRWASyncEnvelopeRejectsOversizedPayload(t *testing.T) {
	adapter := newMockDRWASyncStateAdapter()
	envelope := &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerAssetManager,
		Operations: []drwaSyncOperation{
			{OperationType: drwaSyncOpHolderMirror, TokenID: "A", Holder: "h1", Version: 1, Body: []byte(`{}`)},
			{OperationType: drwaSyncOpHolderMirror, TokenID: "B", Holder: "h2", Version: 1, Body: []byte(`{}`)},
		},
	}
	hash, _ := computeDRWASyncHash(envelope.CallerDomain, envelope.Operations)
	envelope.PayloadHash = hash

	_, err := applyDRWASyncEnvelope(adapter, envelope, 1, []byte("asset_manager"))
	if err == nil || err.Error() != drwaSyncRejectPayloadTooLarge {
		t.Fatalf("expected oversized payload rejection, got %v", err)
	}
}

func TestApplyDRWASyncEnvelopeRejectsPolicyRegistryOnMixedBatch(t *testing.T) {
	adapter := newMockDRWASyncStateAdapter()
	envelope := &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerPolicyRegistry,
		Operations: []drwaSyncOperation{
			{OperationType: drwaSyncOpTokenPolicy, TokenID: "CARBON-1", Version: 1, Body: []byte(`{}`)},
			{OperationType: drwaSyncOpHolderMirror, TokenID: "CARBON-1", Holder: "erd1holder", Version: 1, Body: []byte(`{}`)},
		},
	}
	hash, _ := computeDRWASyncHash(envelope.CallerDomain, envelope.Operations)
	envelope.PayloadHash = hash

	_, err := applyDRWASyncEnvelope(adapter, envelope, 16, []byte("policy_registry"))
	if err == nil || err.Error() != drwaSyncRejectUnauthorizedCaller {
		t.Fatalf("expected mixed-batch caller rejection, got %v", err)
	}
}

func TestApplyDRWASyncEnvelopeRollsBackAfterPartialProgress(t *testing.T) {
	adapter := newMockDRWASyncStateAdapter()
	callCount := 0
	envelope := &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerAssetManager,
		Operations: []drwaSyncOperation{
			{OperationType: drwaSyncOpHolderMirror, TokenID: "CARBON-1", Holder: "erd1holder", Version: 1, Body: []byte(`{"kyc":"approved"}`)},
			{OperationType: drwaSyncOpHolderMirror, TokenID: "CARBON-2", Holder: "erd1holder", Version: 1, Body: []byte(`{"kyc":"approved"}`)},
		},
	}
	hash, _ := computeDRWASyncHash(envelope.CallerDomain, envelope.Operations)
	envelope.PayloadHash = hash

	adapter.putHolderHook = func(tokenID, holder string, version uint64, body []byte) error {
		callCount++
		if callCount == 2 {
			return errDRWATestFailPut
		}
		key := tokenID + "|" + holder
		adapter.holderVersions[key] = version
		adapter.holderBodies[key] = body
		return nil
	}

	_, err := applyDRWASyncEnvelope(adapter, envelope, 16, []byte("asset_manager"))
	if err == nil {
		t.Fatalf("expected rollback-triggering failure")
	}
	if !adapter.rolledBack {
		t.Fatalf("expected rollback after partial progress")
	}
}

func TestApplyDRWASyncEnvelopeAppliesHigherVersion(t *testing.T) {
	adapter := newMockDRWASyncStateAdapter()
	adapter.holderVersions["CARBON-1|erd1holder"] = 1
	envelope := &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerAssetManager,
		Operations: []drwaSyncOperation{{
			OperationType: drwaSyncOpHolderMirror,
			TokenID:       "CARBON-1",
			Holder:        "erd1holder",
			Version:       2,
			Body:          []byte(`{"kyc":"approved"}`),
		}},
	}
	hash, _ := computeDRWASyncHash(envelope.CallerDomain, envelope.Operations)
	envelope.PayloadHash = hash

	result, err := applyDRWASyncEnvelope(adapter, envelope, 16, []byte("asset_manager"))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if result.AppliedOperations != 1 {
		t.Fatalf(drwaTestMsgOneApplied, result.AppliedOperations)
	}
}

func TestApplyDRWASyncEnvelopeDeletesUnexpectedHolderMirror(t *testing.T) {
	adapter := newMockDRWASyncStateAdapter()
	adapter.holderVersions["CARBON-1|erd1ghost"] = 5
	adapter.holderBodies["CARBON-1|erd1ghost"] = []byte(`{"kyc":"approved"}`)
	adapter.ensureHolderIndexed("CARBON-1", "erd1ghost")
	envelope := &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerRecoveryAdmin,
		Operations: []drwaSyncOperation{{
			OperationType: drwaSyncOpHolderMirrorDelete,
			TokenID:       "CARBON-1",
			Holder:        "erd1ghost",
			Version:       6,
		}},
	}
	hash, _ := computeDRWASyncHash(envelope.CallerDomain, envelope.Operations)
	envelope.PayloadHash = hash

	result, err := applyDRWASyncEnvelope(adapter, envelope, 16, []byte("recovery_admin"))
	if err != nil {
		t.Fatalf("expected delete success, got %v", err)
	}
	if result.AppliedOperations != 1 {
		t.Fatalf(drwaTestMsgOneApplied, result.AppliedOperations)
	}
	if _, exists := adapter.holderVersions["CARBON-1|erd1ghost"]; exists {
		t.Fatalf("expected holder mirror deletion")
	}
}

func TestApplyDRWASyncEnvelopeRejectsAssetManagerDeleteOperation(t *testing.T) {
	adapter := newMockDRWASyncStateAdapter()
	envelope := &drwaSyncEnvelope{
		CallerDomain: drwaSyncCallerAssetManager,
		Operations: []drwaSyncOperation{{
			OperationType: drwaSyncOpHolderMirrorDelete,
			TokenID:       "CARBON-1",
			Holder:        "erd1ghost",
			Version:       2,
		}},
	}
	hash, _ := computeDRWASyncHash(envelope.CallerDomain, envelope.Operations)
	envelope.PayloadHash = hash

	_, err := applyDRWASyncEnvelope(adapter, envelope, 16, []byte("asset_manager"))
	if err == nil || err.Error() != drwaSyncRejectUnauthorizedCaller {
		t.Fatalf("expected unauthorized delete rejection, got %v", err)
	}
}

func TestDRWAHookStateAdapterWritesPolicyAndHolder(t *testing.T) {
	systemAccount := teststate.NewAccountWrapMock(core.SystemAccountAddress)
	holderAddress := []byte("erd1holder")
	holderAccount := teststate.NewAccountWrapMock(holderAddress)

	accountsStub := &teststate.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			switch string(address) {
			case string(core.SystemAccountAddress):
				return systemAccount, nil
			case string(holderAddress):
				return holderAccount, nil
			default:
				return nil, nil
			}
		},
		SaveAccountCalled: func(account vmcommon.AccountHandler) error {
			return nil
		},
		JournalLenCalled: func() int {
			return 1
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}

	adapter := newDRWAHookStateAdapter(accountsStub)
	err := adapter.PutTokenPolicyBody("CARBON-1", 2, []byte(`{"regulated":true}`))
	if err != nil {
		t.Fatalf("PutTokenPolicyBody failed: %v", err)
	}

	err = adapter.PutHolderMirrorBody("CARBON-1", string(holderAddress), 3, []byte(`{"kyc":"approved"}`))
	if err != nil {
		t.Fatalf("PutHolderMirrorBody failed: %v", err)
	}

	policyVersion, err := adapter.GetTokenPolicyVersion("CARBON-1")
	if err != nil {
		t.Fatalf("GetTokenPolicyVersion failed: %v", err)
	}
	if policyVersion != 2 {
		t.Fatalf("expected policy version 2, got %d", policyVersion)
	}

	holderVersion, err := adapter.GetHolderMirrorVersion("CARBON-1", string(holderAddress))
	if err != nil {
		t.Fatalf("GetHolderMirrorVersion failed: %v", err)
	}
	if holderVersion != 3 {
		t.Fatalf("expected holder version 3, got %d", holderVersion)
	}
}

// TestDecodeDRWASyncEnvelopeBinaryPathAndApply proves that a binary hook payload
// produced by the Rust managedDRWASyncMirror call (format: [32-byte keccak256] ||
// [canonical binary payload]) is correctly decoded and atomically applied to the
// native mirror.  This validates N-03 from AuditV3: the invoke_drwa_sync_hook
// wrapper is a no-op in unit builds, so this test exercises the full Go-side path
// that runs on every real block.
func TestDecodeDRWASyncEnvelopeBinaryPathAndApply(t *testing.T) {
	policyBody := []byte(`{"drwa_enabled":true,"global_pause":false,"strict_auditor_mode":false,"metadata_protection_enabled":false}`)
	operations := []drwaSyncOperation{{
		OperationType: drwaSyncOpTokenPolicy,
		TokenID:       "BOND-1",
		Version:       1,
		Body:          policyBody,
	}}

	// Build the canonical binary payload (mirrors Rust serialize_sync_envelope_payload).
	canonical, err := serializeDRWASyncEnvelopePayload(drwaSyncCallerPolicyRegistry, operations)
	if err != nil {
		t.Fatalf("serialize failed: %v", err)
	}

	// Compute keccak256 hash of canonical payload (mirrors Rust crypto().keccak256()).
	hash, err := computeDRWASyncHash(drwaSyncCallerPolicyRegistry, operations)
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	if len(hash) != drwaBinaryHashSize {
		t.Fatalf("expected %d-byte hash, got %d", drwaBinaryHashSize, len(hash))
	}

	// Assemble the full binary hook payload: [32-byte hash] || [canonical].
	// This is exactly what build_sync_hook_payload in Rust produces.
	binaryPayload := append(hash, canonical...)

	// The first byte of a binary payload is never '{' (it is the caller tag byte 0x00
	// for PolicyRegistry), so decodeDRWASyncEnvelope must route to the binary path.
	if binaryPayload[0] == '{' {
		t.Fatalf("test invariant broken: binary payload starts with '{'")
	}

	envelope, err := decodeDRWASyncEnvelope(binaryPayload)
	if err != nil {
		t.Fatalf("binary decode failed: %v", err)
	}
	if envelope.CallerDomain != drwaSyncCallerPolicyRegistry {
		t.Fatalf("wrong caller domain: %q", envelope.CallerDomain)
	}
	if len(envelope.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(envelope.Operations))
	}
	op := envelope.Operations[0]
	if op.TokenID != "BOND-1" || op.Version != 1 {
		t.Fatalf("wrong operation fields: %+v", op)
	}
	if !bytes.Equal(op.Body, policyBody) {
		t.Fatalf("body mismatch: %q", op.Body)
	}

	// Apply to state adapter — proves the mirror is atomically updated.
	adapter := newMockDRWASyncStateAdapter()
	result, err := applyDRWASyncEnvelope(adapter, envelope, 16, []byte("policy_registry"))
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if result.AppliedOperations != 1 {
		t.Fatalf(drwaTestMsgOneApplied, result.AppliedOperations)
	}
	if adapter.tokenVersions["BOND-1"] != 1 {
		t.Fatalf("expected mirror version 1, got %d", adapter.tokenVersions["BOND-1"])
	}
	if !bytes.Equal(adapter.tokenBodies["BOND-1"], policyBody) {
		t.Fatalf("mirror body mismatch")
	}
}

func TestDRWAHookStateAdapterRejectsWrongTypeAssertion(t *testing.T) {
	accountsStub := &teststate.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return &teststate.StateUserAccountHandlerStub{}, nil
		},
	}

	adapter := newDRWAHookStateAdapter(accountsStub)
	_, err := adapter.GetTokenPolicyVersion("CARBON-1")
	if err != process.ErrWrongTypeAssertion {
		t.Fatalf("expected wrong type assertion, got %v", err)
	}
}
