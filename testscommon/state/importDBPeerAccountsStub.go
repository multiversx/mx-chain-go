package state

import "github.com/multiversx/mx-chain-core-go/data"

// ImportDBPeerAccountsStub extends AccountsStub with optional import-db lifecycle hooks.
type ImportDBPeerAccountsStub struct {
	*AccountsStub
	RecordImportDBPeerCommitCalled    func(data.HeaderHandler, []byte)
	ProcessImportDBPeerSnapshotCalled func(data.HeaderHandler, []byte, uint64, []byte)
	InitializeImportDBPeerStateCalled func() error
}

// RecordImportDBPeerCommit records a successfully committed execution.
func (s *ImportDBPeerAccountsStub) RecordImportDBPeerCommit(header data.HeaderHandler, hash []byte) {
	if s.RecordImportDBPeerCommitCalled != nil {
		s.RecordImportDBPeerCommitCalled(header, hash)
	}
}

// ProcessImportDBPeerSnapshot observes canonical commits and execution finality.
func (s *ImportDBPeerAccountsStub) ProcessImportDBPeerSnapshot(header data.HeaderHandler, hash []byte, finalNonce uint64, finalHash []byte) {
	if s.ProcessImportDBPeerSnapshotCalled != nil {
		s.ProcessImportDBPeerSnapshotCalled(header, hash, finalNonce, finalHash)
	}
}

// InitializeImportDBPeerState initializes the restored peer state before replay.
func (s *ImportDBPeerAccountsStub) InitializeImportDBPeerState() error {
	if s.InitializeImportDBPeerStateCalled != nil {
		return s.InitializeImportDBPeerStateCalled()
	}
	return nil
}

// IsInterfaceNil reports whether the import-db accounts stub is nil.
func (s *ImportDBPeerAccountsStub) IsInterfaceNil() bool {
	return s == nil
}
