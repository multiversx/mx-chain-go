package node_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/checkpoint"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/config"
	factoryMock "github.com/multiversx/mx-chain-go/factory/mock"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/dataTrieValue"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/trie/leavesRetriever"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

// warningsCounter counts the log lines with the given message written at warn level or above. It is the formatter
// and the (discarding) writer of its log observer, both safe to be called concurrently
type warningsCounter struct {
	message string
	num     atomic.Int32
}

// Output counts the line
func (wc *warningsCounter) Output(line logger.LogLineHandler) []byte {
	if line.GetLogLevel() >= int32(logger.LogWarning) && line.GetMessage() == wc.message {
		wc.num.Add(1)
	}

	return nil
}

// Write discards the output
func (wc *warningsCounter) Write(p []byte) (int, error) {
	return len(p), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (wc *warningsCounter) IsInterfaceNil() bool {
	return wc == nil
}

// countWarnings observes the logs until the end of the test, counting the warnings with the given message
func countWarnings(t *testing.T, message string) *warningsCounter {
	counter := &warningsCounter{message: message}
	require.Nil(t, logger.AddLogObserver(counter, counter))
	t.Cleanup(func() {
		_ = logger.RemoveLogObserver(counter)
	})

	return counter
}

type trieIterationTestEnv struct {
	node          *node.Node
	address       string
	pubKey        []byte
	expectedPairs map[string]string
	loadOptions   []api.AccountQueryOptions
	codeOptions   []api.AccountQueryOptions
	blockInfo     common.BlockInfo
	loadErr       error
	headerEpoch   uint32
	afterLoad     func() // called after every account load, e.g. to advance the current state
	process       *factoryMock.ProcessComponentsMock
	systemAccount state.UserAccountHandler                         // served for the system account address when set
	userAccount   state.UserAccountHandler                         // served instead of the built account when set
	enableEpochs  *enableEpochsHandlerMock.EnableEpochsHandlerStub // the activation flags of the node
	hasher        *hashingMocks.HasherMock                         // the hasher of the node
}

func newTrieIterationTestEnv(t *testing.T, numLeaves int, cfg config.TrieOperationsConfig) *trieIterationTestEnv {
	leaves := make(map[string][]byte)
	for i := 0; i < numLeaves; i++ {
		leaves[fmt.Sprintf("key%d", i)] = []byte(fmt.Sprintf("value%d", i))
	}

	return newTrieIterationTestEnvWithLeaves(t, leaves, cfg)
}

// newTrieIterationTestEnvWithLeaves builds a node whose test account has a real data trie holding the given leaves
// (value stored as value||key||address, the data trie leaf format). An account without leaves has no data trie
func newTrieIterationTestEnvWithLeaves(t *testing.T, leaves map[string][]byte, cfg config.TrieOperationsConfig) *trieIterationTestEnv {
	env := &trieIterationTestEnv{expectedPairs: make(map[string]string)}
	addressBytes := bytes.Repeat([]byte{1}, 32)
	env.pubKey = addressBytes
	env.address = testscommonAddress(addressBytes)
	acc := createAcc(addressBytes)
	acc.SetCodeHash([]byte("code hash"))

	tr := stateMock.GetNewTrie()
	for key, value := range leaves {
		stored := append(append(append([]byte{}, value...), []byte(key)...), addressBytes...)
		require.Nil(t, tr.Update([]byte(key), stored))
		env.expectedPairs[hex.EncodeToString([]byte(key))] = hex.EncodeToString(value)
	}
	require.Nil(t, tr.Commit())
	rootHash, err := tr.RootHash()
	require.Nil(t, err)
	if len(leaves) > 0 {
		acc.SetDataTrie(tr)
		acc.SetRootHash(rootHash)
	}

	env.blockInfo = holders.NewBlockInfo(nil, 0, []byte("main root"))
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = &stateMock.AccountsRepositoryStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
			env.loadOptions = append(env.loadOptions, options)
			if env.afterLoad != nil {
				defer env.afterLoad()
			}
			if env.loadErr != nil {
				return nil, nil, env.loadErr
			}
			served := acc
			if env.userAccount != nil {
				served = env.userAccount
			}
			if env.systemAccount != nil && bytes.Equal(address, core.SystemAccountAddress) {
				served = env.systemAccount
			}
			if len(options.BlockRootHash) > 0 {
				// as the historical accounts adapter: only the root hash is known
				return served, holders.NewBlockInfo(nil, 0, options.BlockRootHash), nil
			}
			return served, env.blockInfo, nil
		},
		GetCodeWithBlockInfoCalled: func(codeHash []byte, options api.AccountQueryOptions) ([]byte, common.BlockInfo, error) {
			env.codeOptions = append(env.codeOptions, options)
			if len(codeHash) == 0 {
				return nil, env.blockInfo, nil
			}
			return []byte("code"), env.blockInfo, nil
		},
	}
	stateComponents.LeavesRetriever, err = leavesRetriever.NewLeavesRetriever(&marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}, cfg.WithDefaults().MaxBytes)
	require.Nil(t, err)

	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	env.enableEpochs = enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
	coreComponents.EnableEpochsHandlerField = env.enableEpochs
	env.hasher = &hashingMocks.HasherMock{}
	coreComponents.Hash = env.hasher
	coreComponents.EpochChangeNotifier = &epochNotifier.EpochNotifierStub{CurrentEpochCalled: func() uint32 {
		return 9
	}}
	// "header hash" is the last executed block, kept in memory with epoch 6; the header of "final hash" is stored
	// with epoch 7 and is also the current block (its hash must not be matched with the last executed header);
	// headers of other hashes are not found
	env.headerEpoch = 7
	dataComponents := getDefaultDataComponents()
	dataComponents.BlockChain = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderHashCalled: func() []byte {
			return []byte("final hash")
		},
		GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 42, Epoch: 6}
		},
		GetLastExecutedBlockInfoCalled: func() (uint64, []byte, []byte) {
			return 42, []byte("header hash"), []byte("main root")
		},
	}
	chainStorerMock := genericMocks.NewChainStorerMock(0)
	dataComponents.Store = chainStorerMock
	processComponents := getDefaultProcessComponents()
	processComponents.HistoryRepositoryInternal = &dblookupext.HistoryRepositoryStub{
		IsEnabledCalled: func() bool {
			return false
		},
	}
	env.process = processComponents
	env.node, err = node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithProcessComponents(processComponents),
		node.WithStateComponents(stateComponents),
		node.WithTrieOperations(cfg),
	)
	require.Nil(t, err)
	headerBytes, err := coreComponents.InternalMarshalizer().Marshal(&block.Header{Nonce: 41, Epoch: env.headerEpoch})
	require.Nil(t, err)
	require.Nil(t, chainStorerMock.BlockHeaders.Put([]byte("final hash"), headerBytes))

	return env
}

// onMetachain makes the node a metachain one, where the ESDT system smart contract lives
func (env *trieIterationTestEnv) onMetachain(t *testing.T) {
	env.process.ShardCoord = &mock.ShardCoordinatorMock{SelfShardId: core.MetachainShardId}
	require.Nil(t, env.node.ApplyOptions(node.WithProcessComponents(env.process)))
}

func testscommonAddress(addressBytes []byte) string {
	encoded, _ := createMockPubkeyConverter().Encode(addressBytes)
	return encoded
}

func collectPairs(pairs map[string]string) common.LeafVisitor {
	return func(leaf core.KeyValueHolder) (bool, uint64, error) {
		pairs[hex.EncodeToString(leaf.Key())] = hex.EncodeToString(leaf.Value())
		return true, uint64(len(leaf.Key()) + len(leaf.Value())), nil
	}
}

func TestNode_IterateDataTrie(t *testing.T) {
	t.Parallel()

	t.Run("pages through the trie, continuing at the recorded root hash with the first page block info", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{})
		env.blockInfo = holders.NewBlockInfo([]byte("header hash"), 42, []byte("main root"))
		expectedBlockInfo := api.BlockInfo{Nonce: 42, Hash: hex.EncodeToString([]byte("header hash")), RootHash: hex.EncodeToString([]byte("main root"))}
		collected := make(map[string]string)
		var cp *checkpoint.Checkpoint
		numPages := 0
		for {
			// a client repeats the same request, selectors included, with the new iterator state
			pagePairs := make(map[string]string)
			blockInfo, next, err := env.node.IterateDataTrie(context.Background(), env.address, api.AccountQueryOptions{OnFinalBlock: true}, 10, cp, collectPairs(pagePairs))
			require.Nil(t, err)
			require.Equal(t, expectedBlockInfo, blockInfo)
			require.LessOrEqual(t, len(pagePairs), 10)
			for k, v := range pagePairs {
				require.NotContains(t, collected, k)
				collected[k] = v
			}
			numPages++
			require.Less(t, numPages, 10)
			if next == nil {
				break
			}
			require.Equal(t, expectedBlockInfo.RootHash, next.BlockRootHash)
			require.Equal(t, expectedBlockInfo.Hash, next.BlockHash)
			require.Equal(t, expectedBlockInfo.Nonce, next.BlockNonce)
			require.NotNil(t, next.Epoch)
			require.Equal(t, uint32(6), *next.Epoch) // the epoch of the last executed block, the page was served from it
			require.NotEmpty(t, next.IteratorState)
			cp = next
		}
		require.Equal(t, 3, numPages)
		require.Equal(t, env.expectedPairs, collected)
		require.Len(t, env.loadOptions, 3)
		require.Equal(t, api.AccountQueryOptions{OnFinalBlock: true}, env.loadOptions[0])
		for _, loadOptions := range env.loadOptions[1:] {
			require.Equal(t, api.AccountQueryOptions{BlockRootHash: []byte("main root"), HintEpoch: core.OptionalUint32{Value: 6, HasValue: true}}, loadOptions)
		}
	})
	t.Run("stored header epoch is recorded for a block that is not the last executed one", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{})
		env.blockInfo = holders.NewBlockInfo([]byte("final hash"), 41, []byte("main root"))
		pairs := make(map[string]string)
		_, next, err := env.node.IterateDataTrie(context.Background(), env.address, api.AccountQueryOptions{}, 10, nil, collectPairs(pairs))
		require.Nil(t, err)
		require.NotNil(t, next)
		require.Equal(t, uint32(7), *next.Epoch)
	})
	t.Run("current epoch is recorded when the block header is not available", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{})
		env.blockInfo = holders.NewBlockInfo([]byte("unknown header"), 43, []byte("main root"))
		pairs := make(map[string]string)
		_, next, err := env.node.IterateDataTrie(context.Background(), env.address, api.AccountQueryOptions{}, 10, nil, collectPairs(pairs))
		require.Nil(t, err)
		require.NotNil(t, next)
		require.Equal(t, uint32(9), *next.Epoch)

		env.blockInfo = holders.NewBlockInfo(nil, 0, []byte("main root"))
		_, next, err = env.node.IterateDataTrie(context.Background(), env.address, api.AccountQueryOptions{}, 10, nil, collectPairs(pairs))
		require.Nil(t, err)
		require.NotNil(t, next)
		require.Equal(t, uint32(9), *next.Epoch)
	})
	t.Run("historical first page records the hint epoch for the continuation", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{})
		options := api.AccountQueryOptions{BlockRootHash: []byte("main root"), HintEpoch: core.OptionalUint32{Value: 5, HasValue: true}}
		pairs := make(map[string]string)
		blockInfo, next, err := env.node.IterateDataTrie(context.Background(), env.address, options, 10, nil, collectPairs(pairs))
		require.Nil(t, err)
		require.Equal(t, api.BlockInfo{RootHash: hex.EncodeToString([]byte("main root"))}, blockInfo)
		require.NotNil(t, next)
		require.Equal(t, uint32(5), *next.Epoch)

		blockInfo, next, err = env.node.IterateDataTrie(context.Background(), env.address, api.AccountQueryOptions{}, 10, next, collectPairs(pairs))
		require.Nil(t, err)
		require.Equal(t, api.BlockInfo{RootHash: hex.EncodeToString([]byte("main root"))}, blockInfo)
		require.NotNil(t, next)
		require.Equal(t, uint32(5), *next.Epoch)
		require.Len(t, pairs, 20)
		require.Len(t, env.loadOptions, 2)
		for _, loadOptions := range env.loadOptions {
			require.Equal(t, options, loadOptions)
		}
	})
	t.Run("legacy IterateKeys continues with the raw iterator state at the caller block", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{})
		env.blockInfo = holders.NewBlockInfo([]byte("header hash"), 42, []byte("main root"))
		collected := make(map[string]string)
		var state [][]byte
		numPages := 0
		for {
			pairs, next, blockInfo, err := env.node.IterateKeys(env.address, 10, state, api.AccountQueryOptions{}, context.Background())
			require.Nil(t, err)
			require.Equal(t, api.BlockInfo{Nonce: 42, Hash: hex.EncodeToString([]byte("header hash")), RootHash: hex.EncodeToString([]byte("main root"))}, blockInfo)
			for k, v := range pairs {
				require.NotContains(t, collected, k)
				collected[k] = v
			}
			numPages++
			if len(next) == 0 {
				break
			}
			state = next
		}
		require.Equal(t, 3, numPages)
		require.Equal(t, env.expectedPairs, collected)
		for _, loadOptions := range env.loadOptions {
			require.Equal(t, api.AccountQueryOptions{}, loadOptions)
		}
	})
	t.Run("numKeys zero and above the maximum are bounded by MaxLeaves", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{MaxLeaves: 7})
		for _, numKeys := range []uint{0, 100} {
			pairs := make(map[string]string)
			_, next, err := env.node.IterateDataTrie(context.Background(), env.address, api.AccountQueryOptions{}, numKeys, nil, collectPairs(pairs))
			require.Nil(t, err)
			require.Len(t, pairs, 7)
			require.NotNil(t, next)
		}
	})
	t.Run("byte budget stops the page", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{MaxBytes: 1})
		pairs := make(map[string]string)
		_, next, err := env.node.IterateDataTrie(context.Background(), env.address, api.AccountQueryOptions{}, 0, nil, collectPairs(pairs))
		require.Nil(t, err)
		require.NotEmpty(t, pairs)
		require.Less(t, len(pairs), 25)
		require.NotNil(t, next)
	})
	t.Run("empty account and account without data trie return an empty page", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{})
		env.loadErr = state.NewErrAccountNotFoundAtBlock(holders.NewBlockInfo([]byte("h"), 3, []byte("r")))
		pairs := make(map[string]string)
		blockInfo, next, err := env.node.IterateDataTrie(context.Background(), env.address, api.AccountQueryOptions{}, 0, nil, collectPairs(pairs))
		require.Nil(t, err)
		require.Equal(t, api.BlockInfo{Nonce: 3, Hash: hex.EncodeToString([]byte("h")), RootHash: hex.EncodeToString([]byte("r"))}, blockInfo)
		require.Nil(t, next)
		require.Empty(t, pairs)

		env = newTrieIterationTestEnv(t, 0, config.TrieOperationsConfig{})
		blockInfo, next, err = env.node.IterateDataTrie(context.Background(), env.address, api.AccountQueryOptions{}, 0, nil, collectPairs(pairs))
		require.Nil(t, err)
		require.Equal(t, api.BlockInfo{RootHash: hex.EncodeToString([]byte("main root"))}, blockInfo)
		require.Nil(t, next)
		require.Empty(t, pairs)
	})
	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{MaxConcurrentRequests: 1})
		pairs := make(map[string]string)

		missing := &checkpoint.Checkpoint{IteratorState: [][]byte{bytes.Repeat([]byte{7}, 32)}}
		_, _, err := env.node.IterateDataTrie(context.Background(), env.address, api.AccountQueryOptions{}, 0, missing, collectPairs(pairs))
		require.ErrorIs(t, err, common.ErrIteratorNodeUnavailable)

		conflicting := &checkpoint.Checkpoint{BlockRootHash: "aa", IteratorState: [][]byte{bytes.Repeat([]byte{7}, 32)}}
		_, _, err = env.node.IterateDataTrie(context.Background(), env.address, api.AccountQueryOptions{BlockRootHash: []byte{0xbb}}, 0, conflicting, collectPairs(pairs))
		require.ErrorIs(t, err, common.ErrInvalidIteratorState)
		require.Empty(t, env.loadOptions[1:]) // rejected before loading the account

		// an entry with an oversized path is invalid, whatever its node: it is rejected before any node is read
		oversized := &checkpoint.Checkpoint{IteratorState: [][]byte{append(bytes.Repeat([]byte{7}, 32), make([]byte, 4097)...)}}
		_, newState, err := env.node.IterateDataTrie(context.Background(), env.address, api.AccountQueryOptions{}, 0, oversized, collectPairs(pairs))
		require.ErrorIs(t, err, common.ErrInvalidIteratorState)
		require.Nil(t, newState)
		require.Empty(t, pairs)

		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		_, _, err = env.node.IterateDataTrie(cancelledCtx, env.address, api.AccountQueryOptions{}, 0, nil, collectPairs(pairs))
		require.ErrorIs(t, err, node.ErrTrieOperationsTimeout)

		numSlots := 1
		started, release := make(chan struct{}, numSlots), make(chan struct{})
		done := make(chan error, numSlots)
		for i := 0; i < numSlots; i++ {
			go func() {
				_, _, err := env.node.IterateDataTrie(context.Background(), env.address, api.AccountQueryOptions{}, 0, nil, func(core.KeyValueHolder) (bool, uint64, error) {
					select {
					case started <- struct{}{}: // the first leaf of each walk, the others find the buffer full
					default:
					}
					<-release
					return true, 0, nil
				})
				done <- err
			}()
			select { // one walk at a time through the test loader, which is not meant for concurrent calls
			case <-started:
			case <-time.After(5 * time.Second):
				t.Fatal("walk did not start")
			}
		}
		_, _, err = env.node.IterateDataTrie(context.Background(), env.address, api.AccountQueryOptions{}, 0, nil, collectPairs(pairs))
		require.ErrorIs(t, err, common.ErrTrieOperationsBusy)
		close(release)
		for i := 0; i < numSlots; i++ {
			require.Nil(t, <-done)
		}
	})
}

func TestNode_GetKeyValuePairs(t *testing.T) {
	t.Parallel()

	t.Run("pages through the pairs, reporting the first page block info", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{})
		env.blockInfo = holders.NewBlockInfo([]byte("header hash"), 42, []byte("main root"))
		expectedBlockInfo := api.BlockInfo{Nonce: 42, Hash: hex.EncodeToString([]byte("header hash")), RootHash: hex.EncodeToString([]byte("main root"))}
		collected := make(map[string]string)
		var cp *checkpoint.Checkpoint
		numPages := 0
		for {
			pairs, next, blockInfo, err := env.node.GetKeyValuePairs(env.address, api.AccountQueryOptions{}, 10, cp, context.Background())
			require.Nil(t, err)
			require.Equal(t, expectedBlockInfo, blockInfo)
			require.LessOrEqual(t, len(pairs), 10)
			for k, v := range pairs {
				require.NotContains(t, collected, k)
				collected[k] = v
			}
			numPages++
			if next == nil {
				break
			}
			require.Less(t, numPages, 10)
			cp = next
		}
		require.Equal(t, 3, numPages)
		require.Equal(t, env.expectedPairs, collected)

		pairs, next, _, err := env.node.GetKeyValuePairs(env.address, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Nil(t, next) // numKeys 0: bounded by MaxLeaves, all the pairs fit in one page
		require.Equal(t, env.expectedPairs, pairs)
	})
	t.Run("empty account and account without data trie return no pairs", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{})
		env.loadErr = state.NewErrAccountNotFoundAtBlock(holders.NewBlockInfo([]byte("h"), 3, []byte("r")))
		pairs, next, blockInfo, err := env.node.GetKeyValuePairs(env.address, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Equal(t, api.BlockInfo{Nonce: 3, Hash: hex.EncodeToString([]byte("h")), RootHash: hex.EncodeToString([]byte("r"))}, blockInfo)
		require.Nil(t, next)
		require.Empty(t, pairs)
		require.NotNil(t, pairs)

		env = newTrieIterationTestEnv(t, 0, config.TrieOperationsConfig{})
		pairs, next, blockInfo, err = env.node.GetKeyValuePairs(env.address, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Equal(t, api.BlockInfo{RootHash: hex.EncodeToString([]byte("main root"))}, blockInfo)
		require.Nil(t, next)
		require.Empty(t, pairs)

		// a data trie whose keys were all deleted has the empty trie hash as root, which is not a stored node
		env.userAccount = accountWithLeaves(t, env.pubKey, nil)
		require.True(t, common.IsEmptyTrie(env.userAccount.GetRootHash()))
		pairs, next, _, err = env.node.GetKeyValuePairs(env.address, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Nil(t, next)
		require.Empty(t, pairs)
	})
	t.Run("a recorded state that can not be read any more is an unavailable node, a first page keeps the error", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{})
		env.loadErr = core.NewGetNodeFromDBErrWithKey([]byte("main root"), errors.New("key not found"), "id")
		recorded := &checkpoint.Checkpoint{BlockNonce: 3, BlockHash: "aa", BlockRootHash: hex.EncodeToString([]byte("main root")), IteratorState: [][]byte{bytes.Repeat([]byte{7}, 32)}}
		_, _, _, err := env.node.GetKeyValuePairs(env.address, api.AccountQueryOptions{}, 0, recorded, context.Background())
		require.ErrorIs(t, err, common.ErrIteratorNodeUnavailable)

		_, _, _, err = env.node.GetKeyValuePairs(env.address, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Equal(t, env.loadErr, err)
	})
	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{})
		missing := &checkpoint.Checkpoint{IteratorState: [][]byte{bytes.Repeat([]byte{7}, 32)}}
		pairs, next, blockInfo, err := env.node.GetKeyValuePairs(env.address, api.AccountQueryOptions{}, 0, missing, context.Background())
		require.ErrorIs(t, err, common.ErrIteratorNodeUnavailable)
		require.Nil(t, pairs)
		require.Nil(t, next)
		require.Equal(t, api.BlockInfo{}, blockInfo)

		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		_, _, _, err = env.node.GetKeyValuePairs(env.address, api.AccountQueryOptions{}, 0, nil, cancelledCtx)
		require.ErrorIs(t, err, node.ErrTrieOperationsTimeout)
	})
}

func TestNode_GetAccountWithKeys(t *testing.T) {
	t.Parallel()

	t.Run("returns the account and pages through its keys", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{})
		env.blockInfo = holders.NewBlockInfo([]byte("header hash"), 42, []byte("main root"))
		expectedBlockInfo := api.BlockInfo{Nonce: 42, Hash: hex.EncodeToString([]byte("header hash")), RootHash: hex.EncodeToString([]byte("main root"))}
		collected := make(map[string]string)
		var cp *checkpoint.Checkpoint
		numPages := 0
		for {
			account, next, blockInfo, err := env.node.GetAccountWithKeys(env.address, api.AccountQueryOptions{WithKeys: true}, 10, cp, context.Background())
			require.Nil(t, err)
			require.Equal(t, expectedBlockInfo, blockInfo)
			require.Equal(t, env.address, account.Address)
			require.Equal(t, "0", account.Balance)
			require.NotEmpty(t, account.RootHash)
			require.Equal(t, hex.EncodeToString([]byte("code")), account.Code)
			require.NotEmpty(t, account.Pairs)
			require.LessOrEqual(t, len(account.Pairs), 10)
			for k, v := range account.Pairs {
				require.NotContains(t, collected, k)
				collected[k] = v
			}
			numPages++
			if next == nil {
				break
			}
			require.Less(t, numPages, 10)
			cp = next
		}
		require.Equal(t, 3, numPages)
		require.Equal(t, env.expectedPairs, collected)
		require.Len(t, env.loadOptions, 3)                 // one load per page: the account response and the keys share it
		require.Equal(t, env.loadOptions, env.codeOptions) // the code is read from the state of the page
		for _, codeOptions := range env.codeOptions[1:] {
			require.Equal(t, api.AccountQueryOptions{BlockRootHash: []byte("main root"), HintEpoch: core.OptionalUint32{Value: 6, HasValue: true}}, codeOptions)
		}
	})
	t.Run("account that does not exist and account without data trie return no pairs", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{})
		env.loadErr = state.NewErrAccountNotFoundAtBlock(holders.NewBlockInfo([]byte("h"), 3, []byte("r")))
		account, next, blockInfo, err := env.node.GetAccountWithKeys(env.address, api.AccountQueryOptions{WithKeys: true}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Equal(t, api.AccountResponse{Address: env.address, Balance: "0", DeveloperReward: "0"}, account)
		require.Equal(t, api.BlockInfo{Nonce: 3, Hash: hex.EncodeToString([]byte("h")), RootHash: hex.EncodeToString([]byte("r"))}, blockInfo)
		require.Nil(t, next)

		env = newTrieIterationTestEnv(t, 0, config.TrieOperationsConfig{})
		account, next, _, err = env.node.GetAccountWithKeys(env.address, api.AccountQueryOptions{WithKeys: true}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Equal(t, env.address, account.Address)
		require.Nil(t, account.Pairs)
		require.Nil(t, next)
	})
	t.Run("walk error returns no account", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{})
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		account, next, blockInfo, err := env.node.GetAccountWithKeys(env.address, api.AccountQueryOptions{WithKeys: true}, 0, nil, cancelledCtx)
		require.ErrorIs(t, err, node.ErrTrieOperationsTimeout)
		require.Equal(t, api.AccountResponse{}, account)
		require.Nil(t, next)
		require.Equal(t, api.BlockInfo{}, blockInfo)
	})
}

// accountWithLeaves builds a user account with a data trie holding the given leaves (data trie leaf format)
func accountWithLeaves(t *testing.T, addressBytes []byte, leaves map[string][]byte) state.UserAccountHandler {
	acc := createAcc(addressBytes)
	tr := stateMock.GetNewTrie()
	for key, value := range leaves {
		stored := append(append(append([]byte{}, value...), []byte(key)...), addressBytes...)
		require.Nil(t, tr.Update([]byte(key), stored))
	}
	require.Nil(t, tr.Commit())
	rootHash, err := tr.RootHash()
	require.Nil(t, err)
	acc.SetDataTrie(tr)
	acc.SetRootHash(rootHash)
	return acc
}

func marshalledToken(t *testing.T, token *esdt.ESDigitalToken) []byte {
	buff, err := (&marshallerMock.MarshalizerMock{}).Marshal(token)
	require.Nil(t, err)
	return buff
}

func TestNode_GetAllESDTTokens(t *testing.T) {
	t.Parallel()

	esdtKey := func(name string) string { return core.ProtectedKeyPrefix + core.ESDTKeyIdentifier + name }
	nftKey := func(name string, nonce int64) string { return esdtKey(name) + string(big.NewInt(nonce).Bytes()) }
	creator := bytes.Repeat([]byte{2}, 32)
	fungible := &esdt.ESDigitalToken{Value: big.NewInt(10), Type: uint32(core.Fungible)}
	// the metadata of a v1 NFT and of a dynamic SFT lives in the system account, the one of a v2 or dynamic NFT on
	// the user account
	nftV1 := &esdt.ESDigitalToken{Type: uint32(core.NonFungible), Value: big.NewInt(1)}
	nftV1Metadata := &esdt.MetaData{Nonce: 1, Creator: creator, Name: []byte("v1")}
	nftV2 := &esdt.ESDigitalToken{Type: uint32(core.NonFungibleV2), Value: big.NewInt(1), TokenMetaData: &esdt.MetaData{Nonce: 2, Creator: creator, Name: []byte("v2")}}
	dynamicSFT := &esdt.ESDigitalToken{Type: uint32(core.DynamicSFT), Value: big.NewInt(5)}
	dynamicSFTMetadata := &esdt.MetaData{Nonce: 3, Creator: creator, Name: []byte("dynamic sft")}
	emptyNft := &esdt.ESDigitalToken{Type: uint32(core.DynamicNFT), Value: big.NewInt(0), TokenMetaData: &esdt.MetaData{Nonce: 100}}
	userLeaves := map[string][]byte{
		esdtKey("TKKR-7q8w9e"):     marshalledToken(t, fungible),
		nftKey("TCKR-67tgv3", 1):   marshalledToken(t, nftV1),
		nftKey("TCKR-v2v2v2", 2):   marshalledToken(t, nftV2),
		nftKey("SFTD-dyndyn", 3):   marshalledToken(t, dynamicSFT),
		nftKey("TCKR-abcdef", 100): marshalledToken(t, emptyNft),
		esdtKey("BAD-123456"):      []byte("not a token"),
		"not an esdt key":          []byte("ignored"),
	}
	systemLeaves := map[string][]byte{
		nftKey("TCKR-67tgv3", 1): marshalledToken(t, &esdt.ESDigitalToken{TokenMetaData: nftV1Metadata, Reserved: []byte{1}}),
		nftKey("TCKR-v2v2v2", 2): marshalledToken(t, &esdt.ESDigitalToken{TokenMetaData: &esdt.MetaData{Nonce: 2, Name: []byte("stale")}}),
		nftKey("SFTD-dyndyn", 3): marshalledToken(t, &esdt.ESDigitalToken{TokenMetaData: dynamicSFTMetadata, Reserved: []byte{2}}),
	}
	systemAccountAddress := core.SystemAccountAddress
	expectedCreator := testscommonAddress(creator)

	t.Run("tokens decoded from the leaves, metadata from the system account only when it lives there", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnvWithLeaves(t, userLeaves, config.TrieOperationsConfig{})
		env.enableEpochs.AddActiveFlags(common.SaveToSystemAccountFlag)
		env.systemAccount = accountWithLeaves(t, systemAccountAddress, systemLeaves)
		env.afterLoad = func() {
			// the current state advances after the user account is loaded: the system account must not follow it
			env.blockInfo = holders.NewBlockInfo(nil, 1, []byte("newer root"))
		}

		tokens, next, blockInfo, err := env.node.GetAllESDTTokens(env.address, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Nil(t, next)
		require.Equal(t, api.BlockInfo{RootHash: hex.EncodeToString([]byte("main root"))}, blockInfo)
		require.Len(t, tokens, 4)
		require.Equal(t, fungible, tokens["TKKR-7q8w9e"])
		require.Equal(t, &esdt.ESDigitalToken{Type: uint32(core.NonFungible), Value: big.NewInt(1), TokenMetaData: &esdt.MetaData{Nonce: 1, Creator: []byte(expectedCreator), Name: []byte("v1")}}, tokens["TCKR-67tgv3-01"])
		require.Equal(t, &esdt.ESDigitalToken{Type: uint32(core.NonFungibleV2), Value: big.NewInt(1), TokenMetaData: &esdt.MetaData{Nonce: 2, Creator: []byte(expectedCreator), Name: []byte("v2")}}, tokens["TCKR-v2v2v2-02"])
		require.Equal(t, &esdt.ESDigitalToken{Type: uint32(core.DynamicSFT), Value: big.NewInt(5), TokenMetaData: &esdt.MetaData{Nonce: 3, Creator: []byte(expectedCreator), Name: []byte("dynamic sft")}, Reserved: []byte{2}}, tokens["SFTD-dyndyn-03"])
		require.Len(t, env.loadOptions, 2) // the user account at the caller options, the system account pinned to its state
		require.Equal(t, api.AccountQueryOptions{}, env.loadOptions[0])
		require.Equal(t, api.AccountQueryOptions{BlockRootHash: []byte("main root"), HintEpoch: core.OptionalUint32{Value: 9, HasValue: true}}, env.loadOptions[1]) // no block hash known: current epoch
	})
	t.Run("auto balanced leaves: the envelope is validated as a lookup of the key would", func(t *testing.T) {
		t.Parallel()

		marshaller := &marshallerMock.MarshalizerMock{}
		envelope := func(key string, address []byte) []byte {
			buff, err := marshaller.Marshal(&dataTrieValue.TrieLeafData{Key: []byte(key), Value: marshalledToken(t, fungible), Address: address})
			require.Nil(t, err)
			return buff
		}
		for _, fixActive := range []bool{true, false} {
			env := newTrieIterationTestEnv(t, 0, config.TrieOperationsConfig{})
			env.enableEpochs.AddActiveFlags(common.AutoBalanceDataTriesFlag)
			if fixActive {
				env.enableEpochs.AddActiveFlags(common.FixDataTrieLeafValidationFlag)
			}
			acc := createAcc(env.pubKey)
			tr := stateMock.GetNewTrie()
			versioned := tr.(interface {
				UpdateWithVersion(key []byte, value []byte, version core.TrieNodeVersion) error
			})
			hashed := func(key string) []byte { return env.hasher.Compute(key) }
			require.Nil(t, versioned.UpdateWithVersion(hashed(esdtKey("GOOD-aaaaaa")), envelope(esdtKey("GOOD-aaaaaa"), env.pubKey), core.AutoBalanceEnabled))
			require.Nil(t, versioned.UpdateWithVersion(hashed(esdtKey("ADDR-bbbbbb")), envelope(esdtKey("ADDR-bbbbbb"), bytes.Repeat([]byte{9}, 32)), core.AutoBalanceEnabled))
			require.Nil(t, versioned.UpdateWithVersion(hashed(esdtKey("ELSE-cccccc")), envelope(esdtKey("LIES-dddddd"), env.pubKey), core.AutoBalanceEnabled))
			require.Nil(t, versioned.UpdateWithVersion([]byte(esdtKey("LEGA-eeeeee")), append(append(marshalledToken(t, fungible), []byte(esdtKey("LEGA-eeeeee"))...), env.pubKey...), core.NotSpecified))
			require.Nil(t, tr.Commit())
			rootHash, err := tr.RootHash()
			require.Nil(t, err)
			acc.SetDataTrie(tr)
			acc.SetRootHash(rootHash)
			env.userAccount = acc

			tokens, next, _, err := env.node.GetAllESDTTokens(env.address, api.AccountQueryOptions{}, 0, nil, context.Background())
			require.Nil(t, err)
			require.Nil(t, next)
			expected := []string{"GOOD-aaaaaa", "LEGA-eeeeee"}
			if !fixActive {
				expected = append(expected, "ADDR-bbbbbb") // the address binding is only checked once the fix is active
			}
			found := make([]string, 0, len(tokens))
			for name := range tokens {
				found = append(found, name)
			}
			require.ElementsMatch(t, expected, found, "fix active: %v", fixActive) // the envelope lying about its key is never a token
		}
	})
	t.Run("before the save-to-system-account activation the leaves are the whole token", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnvWithLeaves(t, userLeaves, config.TrieOperationsConfig{})
		env.systemAccount = accountWithLeaves(t, systemAccountAddress, systemLeaves)

		tokens, _, _, err := env.node.GetAllESDTTokens(env.address, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Len(t, tokens, 4)
		require.Equal(t, nftV1, tokens["TCKR-67tgv3"+string([]byte{1})]) // no metadata: the identifier keeps the raw nonce, as before
		require.Equal(t, dynamicSFT, tokens["SFTD-dyndyn"+string([]byte{3})])
	})
	t.Run("pages through the tokens, the system account loaded from the recorded state", func(t *testing.T) {
		t.Parallel()

		leaves := make(map[string][]byte)
		for i := 0; i < 25; i++ {
			leaves[esdtKey(fmt.Sprintf("TOKEN-%06d", i))] = marshalledToken(t, fungible)
		}
		env := newTrieIterationTestEnvWithLeaves(t, leaves, config.TrieOperationsConfig{})
		env.blockInfo = holders.NewBlockInfo([]byte("header hash"), 42, []byte("main root"))
		expectedBlockInfo := api.BlockInfo{Nonce: 42, Hash: hex.EncodeToString([]byte("header hash")), RootHash: hex.EncodeToString([]byte("main root"))}

		collected := make(map[string]*esdt.ESDigitalToken)
		var cp *checkpoint.Checkpoint
		numPages := 0
		for {
			tokens, next, blockInfo, err := env.node.GetAllESDTTokens(env.address, api.AccountQueryOptions{}, 10, cp, context.Background())
			require.Nil(t, err)
			require.Equal(t, expectedBlockInfo, blockInfo)
			require.LessOrEqual(t, len(tokens), 10)
			for name, token := range tokens {
				require.NotContains(t, collected, name)
				collected[name] = token
			}
			numPages++
			if next == nil {
				break
			}
			require.Less(t, numPages, 10)
			cp = next
		}
		require.Equal(t, 3, numPages)
		require.Len(t, collected, 25)
		require.Len(t, env.loadOptions, 6) // per page: the user account, then the system account pinned to its state
		require.Equal(t, api.AccountQueryOptions{}, env.loadOptions[0])
		for _, loadOptions := range env.loadOptions[1:] {
			require.Equal(t, api.AccountQueryOptions{BlockRootHash: []byte("main root"), HintEpoch: core.OptionalUint32{Value: 6, HasValue: true}}, loadOptions)
		}
	})
	t.Run("the metadata read from the system account is charged against the byte budget", func(t *testing.T) {
		t.Parallel()

		heavyMetadata := &esdt.MetaData{Nonce: 1, Attributes: make([]byte, 1024)}
		leaves := map[string][]byte{
			nftKey("NFT-000001", 1): marshalledToken(t, &esdt.ESDigitalToken{Type: uint32(core.NonFungible), Value: big.NewInt(1)}),
			nftKey("NFT-000002", 1): marshalledToken(t, &esdt.ESDigitalToken{Type: uint32(core.NonFungible), Value: big.NewInt(1)}),
		}
		env := newTrieIterationTestEnvWithLeaves(t, leaves, config.TrieOperationsConfig{MaxBytes: 1000})
		env.enableEpochs.AddActiveFlags(common.SaveToSystemAccountFlag)
		env.systemAccount = accountWithLeaves(t, systemAccountAddress, map[string][]byte{
			nftKey("NFT-000001", 1): marshalledToken(t, &esdt.ESDigitalToken{TokenMetaData: heavyMetadata}),
			nftKey("NFT-000002", 1): marshalledToken(t, &esdt.ESDigitalToken{TokenMetaData: heavyMetadata}),
		})

		tokens, next, _, err := env.node.GetAllESDTTokens(env.address, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Nil(t, err)
		require.NotNil(t, next) // the leaves are tiny, the metadata is not: one token fills the page
		require.Len(t, tokens, 1)

		more, next, _, err := env.node.GetAllESDTTokens(env.address, api.AccountQueryOptions{}, 0, next, context.Background())
		require.Nil(t, err)
		require.Nil(t, next)
		require.Len(t, more, 1)
		for name := range tokens {
			require.NotContains(t, more, name)
		}
	})
	t.Run("empty account and account without data trie return no tokens", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnv(t, 25, config.TrieOperationsConfig{})
		env.loadErr = state.NewErrAccountNotFoundAtBlock(holders.NewBlockInfo([]byte("h"), 3, []byte("r")))
		tokens, next, blockInfo, err := env.node.GetAllESDTTokens(env.address, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Equal(t, api.BlockInfo{Nonce: 3, Hash: hex.EncodeToString([]byte("h")), RootHash: hex.EncodeToString([]byte("r"))}, blockInfo)
		require.Nil(t, next)
		require.NotNil(t, tokens)
		require.Empty(t, tokens)

		env = newTrieIterationTestEnv(t, 0, config.TrieOperationsConfig{})
		tokens, next, _, err = env.node.GetAllESDTTokens(env.address, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Nil(t, next)
		require.Empty(t, tokens)
		require.Len(t, env.loadOptions, 1) // the system account is not loaded when there is nothing to walk
	})
	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnvWithLeaves(t, map[string][]byte{esdtKey("TKKR-7q8w9e"): marshalledToken(t, fungible)}, config.TrieOperationsConfig{})

		missing := &checkpoint.Checkpoint{IteratorState: [][]byte{bytes.Repeat([]byte{7}, 32)}}
		tokens, next, _, err := env.node.GetAllESDTTokens(env.address, api.AccountQueryOptions{}, 0, missing, context.Background())
		require.ErrorIs(t, err, common.ErrIteratorNodeUnavailable)
		require.Nil(t, tokens)
		require.Nil(t, next)

		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		_, _, _, err = env.node.GetAllESDTTokens(env.address, api.AccountQueryOptions{}, 0, nil, cancelledCtx)
		require.ErrorIs(t, err, node.ErrTrieOperationsTimeout)
	})
	t.Run("legacy ESDT leaves under a shifted path are skipped without a warning", func(t *testing.T) {
		t.Parallel()

		warnings := countWarnings(t, "cannot get ESDT token")
		env := newTrieIterationTestEnvWithLeaves(t, nil, config.TrieOperationsConfig{})
		account := accountWithLeaves(t, env.pubKey, map[string][]byte{
			esdtKey("TKKR-7q8w9e"): marshalledToken(t, fungible),
			esdtKey("TKKR-aaaaaa"): marshalledToken(t, fungible),
			esdtKey("TKKR-bbbbbb"): marshalledToken(t, fungible),
		})
		env.userAccount = account

		tokens, next, _, err := env.node.GetAllESDTTokens(env.address, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Nil(t, next)
		require.Len(t, tokens, 3)

		// two more symbols in front of the root make every key one byte longer, so the value of a legacy leaf is
		// trimmed one byte short and the token can not be decoded
		shifted := &checkpoint.Checkpoint{IteratorState: [][]byte{append(append([]byte{}, account.GetRootHash()...), 0, 0)}}
		tokens, next, _, err = env.node.GetAllESDTTokens(env.address, api.AccountQueryOptions{}, 0, shifted, context.Background())
		require.Nil(t, err)
		require.Nil(t, next)
		require.Empty(t, tokens)
		require.Zero(t, warnings.num.Load())
	})
	t.Run("a stop while the metadata of a token is read happens after the token is complete", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnvWithLeaves(t, userLeaves, config.TrieOperationsConfig{})
		env.enableEpochs.AddActiveFlags(common.SaveToSystemAccountFlag)
		systemAccount := accountWithLeaves(t, systemAccountAddress, systemLeaves)
		env.systemAccount = systemAccount
		all, next, _, err := env.node.GetAllESDTTokens(env.address, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Nil(t, next)

		// the context is done in the middle of the first system account read: the walk may stop only after it
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		env.systemAccount = &cancellingAccount{UserAccountHandler: systemAccount, cancel: cancel}
		firstPage, next, _, err := env.node.GetAllESDTTokens(env.address, api.AccountQueryOptions{}, 0, nil, ctx)
		require.Nil(t, err)
		require.NotNil(t, next) // cut by the context
		require.Less(t, len(firstPage), len(all))
		numWithSystemMetadata := 0
		for identifier, token := range firstPage {
			require.Equal(t, all[identifier], token, identifier) // complete, metadata included
			if token.TokenMetaData != nil {
				numWithSystemMetadata++
			}
		}
		require.Greater(t, numWithSystemMetadata, 0) // the token whose read ended the context is on the page

		rest, next, _, err := env.node.GetAllESDTTokens(env.address, api.AccountQueryOptions{}, 0, next, context.Background())
		require.Nil(t, err)
		require.Nil(t, next)
		for identifier, token := range rest {
			require.NotContains(t, firstPage, identifier)
			firstPage[identifier] = token
		}
		require.Equal(t, all, firstPage)
	})
}

// esdtRegistryLeaves builds the leaves of the ESDT system smart contract data trie: token identifier → marshalled
// token data
func esdtRegistryLeaves(t *testing.T, tokens map[string]*systemSmartContracts.ESDTDataV2) map[string][]byte {
	leaves := make(map[string][]byte)
	for identifier, data := range tokens {
		marshalled, err := getMarshalizer().Marshal(data)
		require.Nil(t, err)
		leaves[identifier] = marshalled
	}
	leaves["no dash so not a token"] = []byte("ignored")
	return leaves
}

func TestNode_GetAllIssuedESDTs(t *testing.T) {
	t.Parallel()

	tokens := map[string]*systemSmartContracts.ESDTDataV2{
		"TCK-RANDOM":    {TokenName: []byte("fungible"), TokenType: []byte(core.FungibleESDT)},
		"SFT-RANDOM":    {TokenName: []byte("semi fungible"), TokenType: []byte(core.SemiFungibleESDT)},
		"SFT-Dynamic":   {TokenName: []byte("semi fungible dynamic"), TokenType: []byte(core.DynamicSFTESDT)},
		"NFT-RANDOM":    {TokenName: []byte("non fungible"), TokenType: []byte(core.NonFungibleESDT)},
		"NFT-RANDOM-V2": {TokenName: []byte("non fungible v2"), TokenType: []byte(core.NonFungibleESDTv2)},
		"NFT-Dynamic":   {TokenName: []byte("non fungible dynamic"), TokenType: []byte(core.DynamicNFTESDT)},
	}

	t.Run("filters by type, every dynamic equivalence included", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnvWithLeaves(t, esdtRegistryLeaves(t, tokens), config.TrieOperationsConfig{})
		env.onMetachain(t)
		for tokenType, expected := range map[string][]string{
			core.FungibleESDT:      {"TCK-RANDOM"},
			core.SemiFungibleESDT:  {"SFT-RANDOM", "SFT-Dynamic"},
			core.NonFungibleESDT:   {"NFT-RANDOM", "NFT-RANDOM-V2", "NFT-Dynamic"},
			core.NonFungibleESDTv2: {"NFT-RANDOM", "NFT-RANDOM-V2", "NFT-Dynamic"},
			"":                     {"TCK-RANDOM", "SFT-RANDOM", "SFT-Dynamic", "NFT-RANDOM", "NFT-RANDOM-V2", "NFT-Dynamic"},
		} {
			identifiers, next, blockInfo, err := env.node.GetAllIssuedESDTs(tokenType, api.AccountQueryOptions{}, 0, nil, context.Background())
			require.Nil(t, err, tokenType)
			require.Nil(t, next)
			require.Equal(t, api.BlockInfo{RootHash: hex.EncodeToString([]byte("main root"))}, blockInfo)
			require.ElementsMatch(t, expected, identifiers, tokenType)
		}
	})
	t.Run("leaves that are not tokens are skipped without a warning", func(t *testing.T) {
		t.Parallel()

		warnings := countWarnings(t, "cannot unmarshal esdt data")
		leaves := esdtRegistryLeaves(t, tokens)
		leaves["NOT-ATOKEN"] = []byte("not the data of a token")
		leaves["NOT-ATOKEN2"] = []byte("not the data of a token either")
		env := newTrieIterationTestEnvWithLeaves(t, leaves, config.TrieOperationsConfig{})
		env.onMetachain(t)

		identifiers, next, _, err := env.node.GetAllIssuedESDTs(core.FungibleESDT, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Nil(t, next)
		require.Equal(t, []string{"TCK-RANDOM"}, identifiers)
		require.Zero(t, warnings.num.Load())
	})
	t.Run("pages through the identifiers at the recorded state", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnvWithLeaves(t, esdtRegistryLeaves(t, tokens), config.TrieOperationsConfig{})
		env.onMetachain(t)
		collected := make([]string, 0)
		var cp *checkpoint.Checkpoint
		numPages := 0
		for {
			identifiers, next, _, err := env.node.GetAllIssuedESDTs("", api.AccountQueryOptions{}, 2, cp, context.Background())
			require.Nil(t, err)
			require.LessOrEqual(t, len(identifiers), 2)
			collected = append(collected, identifiers...)
			numPages++
			if next == nil {
				break
			}
			require.Less(t, numPages, 10)
			cp = next
		}
		require.GreaterOrEqual(t, numPages, 3) // a page can be empty when the remaining leaves are all filtered out
		require.ElementsMatch(t, []string{"TCK-RANDOM", "SFT-RANDOM", "SFT-Dynamic", "NFT-RANDOM", "NFT-RANDOM-V2", "NFT-Dynamic"}, collected)
		for _, loadOptions := range env.loadOptions[1:] {
			require.Equal(t, api.AccountQueryOptions{BlockRootHash: []byte("main root"), HintEpoch: core.OptionalUint32{Value: 9, HasValue: true}}, loadOptions)
		}
	})
	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnvWithLeaves(t, esdtRegistryLeaves(t, tokens), config.TrieOperationsConfig{})
		_, _, _, err := env.node.GetAllIssuedESDTs("", api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Equal(t, node.ErrMetachainOnlyEndpoint, err)

		env.onMetachain(t)
		env.loadErr = state.NewErrAccountNotFoundAtBlock(holders.NewBlockInfo([]byte("h"), 3, []byte("r")))
		_, _, _, err = env.node.GetAllIssuedESDTs("", api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Equal(t, node.ErrSystemSCAccountNotFound, err)

		env.loadErr = nil
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		_, _, _, err = env.node.GetAllIssuedESDTs("", api.AccountQueryOptions{}, 0, nil, cancelledCtx)
		require.ErrorIs(t, err, node.ErrTrieOperationsTimeout)
	})
}

func TestNode_GetESDTsRegisteredAndRoles(t *testing.T) {
	t.Parallel()

	alice := bytes.Repeat([]byte{3}, 32)
	bob := bytes.Repeat([]byte{4}, 32)
	tokens := map[string]*systemSmartContracts.ESDTDataV2{
		"FNG-aaaaaa": {TokenType: []byte(core.FungibleESDT), OwnerAddress: alice, SpecialRoles: []*systemSmartContracts.ESDTRoles{{Address: alice, Roles: [][]byte{[]byte(core.ESDTRoleLocalMint), []byte(core.ESDTRoleLocalBurn)}}}},
		"NFT-bbbbbb": {TokenType: []byte(core.NonFungibleESDT), OwnerAddress: alice, SpecialRoles: []*systemSmartContracts.ESDTRoles{{Address: bob, Roles: [][]byte{[]byte(core.ESDTRoleNFTCreate)}}}},
		"SFT-cccccc": {TokenType: []byte(core.SemiFungibleESDT), OwnerAddress: bob, SpecialRoles: []*systemSmartContracts.ESDTRoles{{Address: alice, Roles: [][]byte{[]byte(core.ESDTRoleNFTCreate)}}}},
		"NFT-dddddd": {TokenType: []byte(core.DynamicNFTESDT), OwnerAddress: alice},
	}
	aliceAddress := testscommonAddress(alice)

	t.Run("registered NFTs, tokens with a role and the roles of an address", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnvWithLeaves(t, esdtRegistryLeaves(t, tokens), config.TrieOperationsConfig{})
		env.onMetachain(t)

		registered, next, _, err := env.node.GetNFTTokenIDsRegisteredByAddress(aliceAddress, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Nil(t, next)
		require.ElementsMatch(t, []string{"NFT-bbbbbb", "NFT-dddddd"}, registered)

		withRole, next, _, err := env.node.GetESDTsWithRole(aliceAddress, core.ESDTRoleNFTCreate, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Nil(t, next)
		require.Equal(t, []string{"SFT-cccccc"}, withRole)

		roles, next, _, err := env.node.GetESDTsRoles(aliceAddress, api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Nil(t, err)
		require.Nil(t, next)
		require.Equal(t, map[string][]string{
			"FNG-aaaaaa": {core.ESDTRoleLocalMint, core.ESDTRoleLocalBurn},
			"SFT-cccccc": {core.ESDTRoleNFTCreate},
		}, roles)
	})
	t.Run("roles are paged, each page carrying the roles of its tokens", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnvWithLeaves(t, esdtRegistryLeaves(t, tokens), config.TrieOperationsConfig{})
		env.onMetachain(t)
		collected := make(map[string][]string)
		var cp *checkpoint.Checkpoint
		numPages := 0
		for {
			roles, next, _, err := env.node.GetESDTsRoles(aliceAddress, api.AccountQueryOptions{}, 1, cp, context.Background())
			require.Nil(t, err)
			require.LessOrEqual(t, len(roles), 1)
			for identifier, tokenRoles := range roles {
				require.NotContains(t, collected, identifier)
				collected[identifier] = tokenRoles
			}
			numPages++
			if next == nil {
				break
			}
			require.Less(t, numPages, 10)
			cp = next
		}
		require.GreaterOrEqual(t, numPages, 2)
		require.Equal(t, map[string][]string{
			"FNG-aaaaaa": {core.ESDTRoleLocalMint, core.ESDTRoleLocalBurn},
			"SFT-cccccc": {core.ESDTRoleNFTCreate},
		}, collected)
	})
	t.Run("the roles count against the byte budget", func(t *testing.T) {
		t.Parallel()

		mintRole := []*systemSmartContracts.ESDTRoles{{Address: alice, Roles: [][]byte{[]byte(core.ESDTRoleLocalMint)}}}
		// the identifiers share a prefix so the leaves hang from more than one trie node: a page can end between them
		withRoles := map[string]*systemSmartContracts.ESDTDataV2{
			"FNG-aaaaaa": {TokenType: []byte(core.FungibleESDT), SpecialRoles: mintRole},
			"FNG-aaaaab": {TokenType: []byte(core.FungibleESDT), SpecialRoles: mintRole},
			"FNG-cccccc": {TokenType: []byte(core.FungibleESDT), SpecialRoles: mintRole},
		}
		// one token with its role is 10 + 17 bytes: a page holds the leaves of one node only
		env := newTrieIterationTestEnvWithLeaves(t, esdtRegistryLeaves(t, withRoles), config.TrieOperationsConfig{MaxBytes: 20})
		env.onMetachain(t)
		collected := make(map[string][]string)
		var cp *checkpoint.Checkpoint
		numPages := 0
		for {
			roles, next, _, err := env.node.GetESDTsRoles(aliceAddress, api.AccountQueryOptions{}, 0, cp, context.Background())
			require.Nil(t, err)
			require.Less(t, len(roles), 3)
			for identifier, tokenRoles := range roles {
				collected[identifier] = tokenRoles
			}
			numPages++
			if next == nil {
				break
			}
			require.Less(t, numPages, 10)
			cp = next
		}
		require.GreaterOrEqual(t, numPages, 2)
		require.Len(t, collected, 3)
	})
	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		env := newTrieIterationTestEnvWithLeaves(t, esdtRegistryLeaves(t, tokens), config.TrieOperationsConfig{})
		env.onMetachain(t)

		_, _, _, err := env.node.GetESDTsWithRole(aliceAddress, "not a role", api.AccountQueryOptions{}, 0, nil, context.Background())
		require.Equal(t, node.ErrInvalidESDTRole, err)

		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		_, _, _, err = env.node.GetNFTTokenIDsRegisteredByAddress(aliceAddress, api.AccountQueryOptions{}, 0, nil, cancelledCtx)
		require.ErrorIs(t, err, node.ErrTrieOperationsTimeout)
		_, _, _, err = env.node.GetESDTsRoles(aliceAddress, api.AccountQueryOptions{}, 0, nil, cancelledCtx)
		require.ErrorIs(t, err, node.ErrTrieOperationsTimeout)
	})
}

// cancellingAccount cancels a context when its data is read: it stands for the system account whose metadata read
// outlasts the request deadline
type cancellingAccount struct {
	state.UserAccountHandler
	cancel context.CancelFunc
}

// AccountDataHandler returns the data handler of the account, cancelling the context on every read
func (a *cancellingAccount) AccountDataHandler() vmcommon.AccountDataHandler {
	return &cancellingDataHandler{AccountDataHandler: a.UserAccountHandler.(vmcommon.UserAccountHandler).AccountDataHandler(), cancel: a.cancel}
}

type cancellingDataHandler struct {
	vmcommon.AccountDataHandler
	cancel context.CancelFunc
}

// RetrieveValue cancels the context then reads the value
func (h *cancellingDataHandler) RetrieveValue(key []byte) ([]byte, uint32, error) {
	h.cancel()
	return h.AccountDataHandler.RetrieveValue(key)
}
