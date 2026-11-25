package hooks

type drwaSyncOperationType string

const (
	drwaSyncOpTokenPolicy        drwaSyncOperationType = "token_policy"
	drwaSyncOpHolderMirror       drwaSyncOperationType = "holder_mirror"
	drwaSyncOpHolderMirrorDelete drwaSyncOperationType = "holder_mirror_delete"
)

const (
	drwaSyncCallerPolicyRegistry = "policy_registry"
	drwaSyncCallerAssetManager   = "asset_manager"
	drwaSyncCallerRecoveryAdmin  = "recovery_admin"
)

const (
	drwaSyncRejectUnauthorizedCaller = "DRWA_SYNC_UNAUTHORIZED_CALLER"
	drwaSyncRejectPayloadTooLarge    = "DRWA_SYNC_PAYLOAD_TOO_LARGE"
	drwaSyncRejectHashMismatch       = "DRWA_SYNC_HASH_MISMATCH"
	drwaSyncRejectReplayStale        = "DRWA_SYNC_REPLAY_STALE"
	drwaSyncRejectReplayConflict     = "DRWA_SYNC_REPLAY_CONFLICT"
	drwaSyncRejectBatchAtomicity     = "DRWA_SYNC_BATCH_ATOMICITY_ABORT"
)

const (
	drwaSyncTokenPolicyPrefix  = "drwa:token:"
	drwaSyncHolderMirrorPrefix = "drwa:holder:"
	drwaSyncMaxOperations      = 256
	drwaSyncMaxPayloadBytes    = 1 << 20
)

type drwaSyncEnvelope struct {
	CallerDomain string              `json:"caller_domain"`
	PayloadHash  []byte              `json:"payload_hash"`
	Operations   []drwaSyncOperation `json:"operations"`
	Noop         bool                `json:"noop,omitempty"`
}

type drwaSyncOperation struct {
	OperationType drwaSyncOperationType `json:"operation_type"`
	TokenID       string                `json:"token_id"`
	Holder        string                `json:"holder,omitempty"`
	Version       uint64                `json:"version"`
	Body          []byte                `json:"body"`
}

type drwaSyncStoredValue struct {
	Version uint64 `json:"version"`
	Body    []byte `json:"body"`
}

type drwaSyncApplyResult struct {
	AppliedOperations int
	LastTokenID       string
}
