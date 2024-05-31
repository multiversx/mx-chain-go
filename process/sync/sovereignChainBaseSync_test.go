package sync

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func haveTimeAlways() time.Duration {
	return time.Hour
}

func TestBaseSync_sovereignChainProcessAndCommit(t *testing.T) {
	t.Parallel()

	t.Run("sovereignChainProcessAndCommit with process error", func(t *testing.T) {
		t.Parallel()

		errProcess := errors.New("process error")
		boot := &baseBootstrap{
			blockProcessor: &testscommon.BlockProcessorStub{
				ProcessBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error) {
					return nil, nil, errProcess
				},
			},
		}

		header := &block.Header{}
		body := &block.Body{}
		err := boot.sovereignChainProcessAndCommit(header, body, haveTimeAlways)
		assert.Equal(t, errProcess, err)
	})

	t.Run("sovereignChainProcessAndCommit with commit error", func(t *testing.T) {
		t.Parallel()

		errCommit := errors.New("commit error")
		boot := &baseBootstrap{
			blockProcessor: &testscommon.BlockProcessorStub{
				ProcessBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error) {
					return &block.Header{}, &block.Body{}, nil
				},
				CommitBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
					return errCommit
				},
			},
		}

		header := &block.Header{}
		body := &block.Body{}
		err := boot.sovereignChainProcessAndCommit(header, body, haveTimeAlways)
		assert.Equal(t, errCommit, err)
	})

	t.Run("sovereignChainProcessAndCommit without error", func(t *testing.T) {
		t.Parallel()

		boot := &baseBootstrap{
			blockProcessor: &testscommon.BlockProcessorStub{
				ProcessBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error) {
					return &block.Header{}, &block.Body{}, nil
				},
				CommitBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
					return nil
				},
			},
		}

		header := &block.Header{}
		body := &block.Body{}
		err := boot.sovereignChainProcessAndCommit(header, body, haveTimeAlways)
		assert.Nil(t, err)
	})
}

func TestBaseSync_sovereignChainHandleScheduledRollBackToHeader(t *testing.T) {
	t.Parallel()

	t.Run("sovereignChainHandleScheduledRollBackToHeader with nil header should return genesis root hash", func(t *testing.T) {
		t.Parallel()

		genesisRootHash := []byte("genesis root hash")
		boot := &baseBootstrap{
			chainHandler: &testscommon.ChainHandlerStub{
				GetGenesisHeaderCalled: func() data.HeaderHandler {
					return &block.Header{
						RootHash: genesisRootHash,
					}
				},
			},
		}

		rootHash := boot.sovereignChainHandleScheduledRollBackToHeader(nil, nil)
		assert.Equal(t, genesisRootHash, rootHash)
	})

	t.Run("sovereignChainHandleScheduledRollBackToHeader with not nil header should return header root hash", func(t *testing.T) {
		t.Parallel()

		genesisRootHash := []byte("genesis root hash")
		boot := &baseBootstrap{
			chainHandler: &testscommon.ChainHandlerStub{
				GetGenesisHeaderCalled: func() data.HeaderHandler {
					return &block.Header{
						RootHash: genesisRootHash,
					}
				},
			},
		}

		headerRootHash := []byte("header root hash")
		header := &block.Header{
			RootHash: headerRootHash,
		}

		rootHash := boot.sovereignChainHandleScheduledRollBackToHeader(header, nil)
		assert.Equal(t, headerRootHash, rootHash)
	})
}
