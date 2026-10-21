package block_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/pool"

	processBlock "github.com/multiversx/mx-chain-go/process/block"
)

type finalitySignalCalls struct {
	signaledHashes [][]byte
	finalNonce     uint64
	finalHash      []byte
	finalRootHash  []byte
	headerReads    int
}

// metaChain builds a linked prev-hash chain of v3 meta headers, hash "meta<nonce>", each carrying
// an execution result one nonce behind it
func metaChain(fromNonce uint64, toNonce uint64) (map[string]data.MetaHeaderHandler, map[uint64][]byte) {
	headers := make(map[string]data.MetaHeaderHandler)
	hashes := make(map[uint64][]byte)

	for nonce := fromNonce; nonce <= toNonce; nonce++ {
		hashes[nonce] = []byte("meta" + string(rune('0'+nonce)))
	}

	for nonce := fromNonce; nonce <= toNonce; nonce++ {
		prevHash := []byte("meta" + string(rune('0'+nonce-1)))
		headers[string(hashes[nonce])] = &block.MetaBlockV3{
			Nonce:    nonce,
			PrevHash: prevHash,
			LastExecutionResult: &block.MetaExecutionResultInfo{
				ExecutionResult: &block.BaseMetaExecutionResult{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderNonce: nonce - 1,
						HeaderHash:  prevHash,
						RootHash:    []byte("root" + string(rune('0'+nonce-1))),
					},
				},
			},
		}
	}

	return headers, hashes
}

// finalitySignaler is the narrow view of the meta processor these tests drive, since the concrete
// type is unexported
type finalitySignaler interface {
	SignalNewlyFinalBlocks(metaBlock data.MetaHeaderHandler, metaBlockHash []byte)
	SetLastSignaledFinalNonce(nonce uint64)
	GetLastSignaledFinalNonce() uint64
}

func buildFinalitySignalProcessor(
	t *testing.T,
	calls *finalitySignalCalls,
	settledNonce uint64,
	settledHash []byte,
	headers map[string]data.MetaHeaderHandler,
) finalitySignaler {
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()

	poolsStub := dataComponents.DataPool.(*dataRetrieverMock.PoolsHolderStub)
	poolsStub.HeadersCalled = func() dataRetriever.HeadersPool {
		return &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				calls.headerReads++
				header, found := headers[string(hash)]
				if !found {
					return nil, errors.New("header not found")
				}
				return header, nil
			},
		}
	}

	dataComponents.BlockChain = &testscommon.ChainHandlerStub{
		SetFinalBlockInfoCalled: func(nonce uint64, headerHash []byte, rootHash []byte) {
			calls.finalNonce = nonce
			calls.finalHash = headerHash
			calls.finalRootHash = rootHash
		},
	}

	statusComponents.Outport = &outport.OutportStub{
		HasDriversCalled: func() bool {
			return true
		},
		FinalizedBlockCalled: func(finalizedBlock *outportcore.FinalizedBlock) {
			calls.signaledHashes = append(calls.signaledHashes, finalizedBlock.HeaderHash)
		},
	}

	arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.ForkDetector = &mock.ForkDetectorMock{
		GetHighestSettledBlockInfoCalled: func() (uint64, []byte) {
			return settledNonce, settledHash
		},
	}

	mp, err := processBlock.NewMetaProcessor(arguments)
	require.Nil(t, err)

	return mp
}

func TestMetaProcessor_SignalNewlyFinalBlocks(t *testing.T) {
	t.Parallel()

	headers, hashes := metaChain(1, 9)
	committed := headers[string(hashes[9])]
	committedHash := hashes[9]

	t.Run("no settled hash should not signal", func(t *testing.T) {
		t.Parallel()

		calls := &finalitySignalCalls{}
		mp := buildFinalitySignalProcessor(t, calls, 8, nil, headers)
		mp.SetLastSignaledFinalNonce(5)

		mp.SignalNewlyFinalBlocks(committed, committedHash)

		require.Empty(t, calls.signaledHashes)
		require.Equal(t, uint64(5), mp.GetLastSignaledFinalNonce())
	})

	t.Run("settled nonce not above the last signaled should not signal", func(t *testing.T) {
		t.Parallel()

		calls := &finalitySignalCalls{}
		mp := buildFinalitySignalProcessor(t, calls, 5, hashes[5], headers)
		mp.SetLastSignaledFinalNonce(5)

		mp.SignalNewlyFinalBlocks(committed, committedHash)

		require.Empty(t, calls.signaledHashes)
	})

	t.Run("settled exactly one above the last signaled should signal only it", func(t *testing.T) {
		t.Parallel()

		calls := &finalitySignalCalls{}
		mp := buildFinalitySignalProcessor(t, calls, 8, hashes[8], headers)
		mp.SetLastSignaledFinalNonce(7)

		mp.SignalNewlyFinalBlocks(committed, committedHash)

		require.Equal(t, [][]byte{hashes[8]}, calls.signaledHashes)
		require.Equal(t, uint64(8), mp.GetLastSignaledFinalNonce())
	})

	t.Run("the steady state should not walk", func(t *testing.T) {
		t.Parallel()

		calls := &finalitySignalCalls{}
		mp := buildFinalitySignalProcessor(t, calls, 8, hashes[8], headers)
		mp.SetLastSignaledFinalNonce(7)

		mp.SignalNewlyFinalBlocks(committed, committedHash)

		require.Equal(t, [][]byte{hashes[8]}, calls.signaledHashes)
		// the only read is the settled header for the block info, the walk must be short circuited
		require.Equal(t, 1, calls.headerReads)
	})

	t.Run("a gap should be back filled in ascending nonce order", func(t *testing.T) {
		t.Parallel()

		calls := &finalitySignalCalls{}
		mp := buildFinalitySignalProcessor(t, calls, 8, hashes[8], headers)
		mp.SetLastSignaledFinalNonce(4)

		mp.SignalNewlyFinalBlocks(committed, committedHash)

		expected := [][]byte{hashes[5], hashes[6], hashes[7], hashes[8]}
		require.Equal(t, expected, calls.signaledHashes)
		require.Equal(t, uint64(8), mp.GetLastSignaledFinalNonce())
	})

	t.Run("the first signal after a restart should not walk", func(t *testing.T) {
		t.Parallel()

		calls := &finalitySignalCalls{}
		mp := buildFinalitySignalProcessor(t, calls, 8, hashes[8], headers)
		mp.SetLastSignaledFinalNonce(0)

		mp.SignalNewlyFinalBlocks(committed, committedHash)

		require.Equal(t, [][]byte{hashes[8]}, calls.signaledHashes)
		require.Equal(t, uint64(8), mp.GetLastSignaledFinalNonce())
	})

	t.Run("a missing header mid walk should keep the hashes collected so far", func(t *testing.T) {
		t.Parallel()

		partial := make(map[string]data.MetaHeaderHandler)
		for hash, header := range headers {
			if header.GetNonce() == 6 {
				continue
			}
			partial[hash] = header
		}

		calls := &finalitySignalCalls{}
		mp := buildFinalitySignalProcessor(t, calls, 8, hashes[8], partial)
		mp.SetLastSignaledFinalNonce(4)

		mp.SignalNewlyFinalBlocks(committed, committedHash)

		// the walk stops at the missing nonce 6, so 5 is never reached
		require.Equal(t, [][]byte{hashes[7], hashes[8]}, calls.signaledHashes)
		require.Equal(t, uint64(8), mp.GetLastSignaledFinalNonce())
	})

	t.Run("an unloadable settled header should signal only its hash", func(t *testing.T) {
		t.Parallel()

		calls := &finalitySignalCalls{}
		mp := buildFinalitySignalProcessor(t, calls, 8, []byte("unknownHash"), headers)
		mp.SetLastSignaledFinalNonce(4)

		mp.SignalNewlyFinalBlocks(committed, committedHash)

		require.Equal(t, [][]byte{[]byte("unknownHash")}, calls.signaledHashes)
		require.Equal(t, uint64(8), mp.GetLastSignaledFinalNonce())
	})

	t.Run("the block info is anchored on the settled block execution result", func(t *testing.T) {
		t.Parallel()

		calls := &finalitySignalCalls{}
		mp := buildFinalitySignalProcessor(t, calls, 8, hashes[8], headers)
		mp.SetLastSignaledFinalNonce(7)

		mp.SignalNewlyFinalBlocks(committed, committedHash)

		require.Equal(t, uint64(7), calls.finalNonce)
		require.Equal(t, hashes[7], calls.finalHash)
		require.Equal(t, []byte("root7"), calls.finalRootHash)
	})

	t.Run("the committed block is used directly when it is the settled one", func(t *testing.T) {
		t.Parallel()

		calls := &finalitySignalCalls{}
		mp := buildFinalitySignalProcessor(t, calls, 9, committedHash, headers)
		mp.SetLastSignaledFinalNonce(8)

		mp.SignalNewlyFinalBlocks(committed, committedHash)

		require.Equal(t, [][]byte{committedHash}, calls.signaledHashes)
		require.Equal(t, uint64(8), calls.finalNonce)
		require.Equal(t, hashes[8], calls.finalHash)
	})
}
