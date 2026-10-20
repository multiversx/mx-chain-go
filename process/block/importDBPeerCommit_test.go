package block

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportDisabled "github.com/multiversx/mx-chain-go/outport/disabled"
	"github.com/multiversx/mx-chain-go/state"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/require"
)

type importDBPeerCommitObserver struct {
	*stateMock.AccountsStub
	record func(data.HeaderHandler, []byte)
}

func (o *importDBPeerCommitObserver) RecordImportDBPeerCommit(header data.HeaderHandler, hash []byte) {
	o.record(header, hash)
}

func TestImportDBPeerRootRecordedOnlyAfterSuccessfulStateCommit(t *testing.T) {
	committed, recorded := false, false
	var commitErr error
	peer := &importDBPeerCommitObserver{AccountsStub: &stateMock.AccountsStub{
		CommitCalled: func() ([]byte, error) { committed = true; return []byte("root"), commitErr },
	}}
	peer.record = func(header data.HeaderHandler, hash []byte) {
		require.True(t, committed)
		require.Equal(t, uint64(10), header.GetNonce())
		require.Equal(t, []byte("hash"), hash)
		recorded = true
	}
	bp := &baseProcessor{
		accountsDB:     map[state.AccountsDbIdentifier]state.AccountsAdapter{state.PeerAccountsState: peer},
		outportHandler: outportDisabled.NewDisabledOutport(),
	}
	require.NoError(t, bp.commitStateForHeader(&block.MetaBlock{Nonce: 10}, []byte("hash")))
	require.True(t, recorded)
	recorded, commitErr = false, errors.New("failed commit")
	require.Error(t, bp.commitStateForHeader(&block.MetaBlock{Nonce: 10}, []byte("hash")))
	require.False(t, recorded)
}
