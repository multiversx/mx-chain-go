package sync

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/update/mock"
)

func TestMissingHeadersProofUsesHeaderEpoch(t *testing.T) {
	t.Parallel()

	for _, source := range []string{"pool", "already collected", "received callback"} {
		t.Run(source, func(t *testing.T) {
			header := &block.Header{Epoch: 2238, ShardID: 0}
			requests := make(chan uint32, 1)
			args := getMisingHeadersByHashSyncerArgs()
			args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledInEpochCalled: func(core.EnableEpochFlag, uint32) bool { return true },
			}
			args.Cache = &mock.HeadersCacherStub{
				GetHeaderByHashCalled: func([]byte) (data.HeaderHandler, error) {
					if source == "pool" {
						return header, nil
					}
					return nil, errors.New("header absent from pool")
				},
			}
			args.RequestHandler = &testscommon.RequestHandlerStub{
				RequestEquivalentProofByHashForEpochCalled: func(_ uint32, _ []byte, epoch uint32) {
					requests <- epoch
				},
			}
			syncer, err := NewMissingheadersByHashSyncer(args)
			require.NoError(t, err)
			if source == "already collected" {
				syncer.mapHeaders["hash"] = header
			}
			if source == "received callback" {
				syncer.stopSyncing = false
				syncer.mapHashes["hash"] = struct{}{}
				syncer.receivedHeader(header, []byte("hash"))
			} else {
				syncer.updateMapsAndRequestIfNeeded(0, "hash", map[string]uint32{"hash": 0})
			}
			select {
			case epoch := <-requests:
				require.Equal(t, header.Epoch, epoch)
			case <-time.After(time.Second):
				t.Fatal("expected a proof request using the header epoch")
			}
		})
	}
}

func TestEpochStartShardProofUsesHeaderEpoch(t *testing.T) {
	t.Parallel()

	args := createPendingEpochStartShardHeaderSyncerArgs()
	var requestedEpoch uint32
	args.RequestHandler = &testscommon.RequestHandlerStub{
		RequestEquivalentProofByHashForEpochCalled: func(_ uint32, _ []byte, epoch uint32) {
			requestedEpoch = epoch
		},
	}
	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.NoError(t, err)
	syncer.stopSyncing = false
	syncer.targetShardId = 0
	syncer.expectedNonce = 100
	syncer.receivedHeader(&block.Header{Epoch: 2238, ShardID: 0, Nonce: 100}, []byte("hash"))
	require.Equal(t, uint32(2238), requestedEpoch)
}
