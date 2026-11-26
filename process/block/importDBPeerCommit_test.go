package block

import (
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/state"
	outportMocks "github.com/multiversx/mx-chain-go/testscommon/outport"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/require"
)

func TestImportDBPeerRootRecordedOnlyAfterSuccessfulStateCommit(t *testing.T) {
	for _, hasDrivers := range []bool{false, true} {
		for _, fails := range []bool{false, true} {
			t.Run(fmt.Sprintf("drivers=%t/fails=%t", hasDrivers, fails), func(t *testing.T) {
				var events []string
				var commitErr error
				if fails {
					commitErr = errors.New("failed commit")
				}
				peer := &stateMock.ImportDBPeerAccountsStub{
					AccountsStub: &stateMock.AccountsStub{CommitCalled: func() ([]byte, error) {
						events = append(events, "commit")
						return []byte("root"), commitErr
					}},
					RecordImportDBPeerCommitCalled: func(header data.HeaderHandler, hash []byte) {
						require.Equal(t, uint64(10), header.GetNonce())
						require.Equal(t, []byte("hash"), hash)
						events = append(events, "record")
					},
				}
				bp := &baseProcessor{
					accountsDB:     map[state.AccountsDbIdentifier]state.AccountsAdapter{state.PeerAccountsState: peer},
					outportHandler: &outportMocks.OutportStub{HasDriversCalled: func() bool { return hasDrivers }},
					stateAccessesCollector: &stateMock.StateAccessesCollectorStub{
						BeginExecutionCalled:                func([]byte) { events = append(events, "begin") },
						EndExecutionCalled:                  func([]byte) { events = append(events, "end") },
						DiscardStateAccessesForHeaderCalled: func([]byte) { events = append(events, "discard") },
					},
				}
				err := bp.commitStateForHeader(&block.MetaBlock{Nonce: 10}, []byte("hash"))
				if fails {
					require.ErrorIs(t, err, commitErr)
				} else {
					require.NoError(t, err)
				}
				expected := []string{"commit"}
				if hasDrivers {
					expected = append([]string{"begin"}, expected...)
					if fails {
						expected = append(expected, "discard")
					}
					expected = append(expected, "end")
				}
				if !fails {
					expected = append(expected, "record")
				}
				require.Equal(t, expected, events)
			})
		}
	}
}
