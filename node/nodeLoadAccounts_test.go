package node_test

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/holders"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	mockState "github.com/ElrondNetwork/elrond-go/testscommon/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestNode_GetAccountWithOptionsShouldWork(t *testing.T) {
	t.Parallel()

	alice, _ := state.NewUserAccount(testscommon.TestPubKeyAlice)
	alice.Balance = big.NewInt(100)

	accountsRepostitory := &mockState.AccountsRepositoryStub{}
	accountsRepostitory.GetAccountWithBlockInfoCalled = func(pubkey []byte, options api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
		if bytes.Equal(pubkey, testscommon.TestPubKeyAlice) {
			return alice, holders.NewBlockInfo([]byte{0xaa}, 1, []byte{0xbb}), nil
		}

		return nil, nil, state.ErrAccNotFound
	}

	coreComponents := getDefaultCoreComponents()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepostitory

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	account, blockInfo, err := n.GetAccount(testscommon.TestAddressAlice, api.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, "100", account.Balance)
	require.Equal(t, uint64(1), blockInfo.Nonce)
	require.Equal(t, "aa", blockInfo.Hash)
	require.Equal(t, "bb", blockInfo.RootHash)
}

func TestNode_GetCodeWithOptionsShouldWork(t *testing.T) {
	t.Parallel()

	goodCodeHash := []byte("code hash")
	code := []byte("code")

	accountsRepostitory := &mockState.AccountsRepositoryStub{}
	accountsRepostitory.GetCodeWithBlockInfoCalled = func(codeHash []byte, options api.AccountQueryOptions) ([]byte, common.BlockInfo, error) {
		if bytes.Equal(codeHash, goodCodeHash) {
			return code, holders.NewBlockInfo([]byte{0xaa}, 1, []byte{0xbb}), nil
		}

		return nil, nil, nil
	}

	coreComponents := getDefaultCoreComponents()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepostitory

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	code, blockInfo := n.GetCode(goodCodeHash, api.AccountQueryOptions{})
	require.Equal(t, code, code)
	require.Equal(t, uint64(1), blockInfo.Nonce)
	require.Equal(t, "aa", blockInfo.Hash)
	require.Equal(t, "bb", blockInfo.RootHash)
}

func TestNode_AddBlockCoordinatesToAccountQueryOptions(t *testing.T) {
	coreComponents := getDefaultCoreComponents()
	stateComponents := getDefaultStateComponents()
	dataComponents := getDefaultDataComponents()
	processComponents := getDefaultProcessComponents()

	blockHeader := &block.Header{
		Nonce:    42,
		RootHash: []byte("blockRootHash"),
	}
	blockHeaderBytes, _ := coreComponents.InternalMarshalizer().Marshal(blockHeader)

	// Setup storage
	chainStorerMock := genericMocks.NewChainStorerMock(0)
	_ = chainStorerMock.BlockHeaders.Put([]byte("blockHash"), blockHeaderBytes)
	nonceAsStorerKey := coreComponents.Uint64ByteSliceConverter().ToByteSlice(42)
	_ = chainStorerMock.ShardHdrNonce.Put(nonceAsStorerKey, []byte("blockHash"))
	dataComponents.Store = chainStorerMock

	// Setup scheduled txs
	getScheduledRootHashForHeaderResult := []byte{}
	getScheduledRootHashForHeaderError := errors.New("missing")
	scheduledTxsStub := &testscommon.ScheduledTxsExecutionStub{
		GetScheduledRootHashForHeaderCalled: func(headerHash []byte) ([]byte, error) {
			return getScheduledRootHashForHeaderResult, getScheduledRootHashForHeaderError
		},
	}
	processComponents.ScheduledTxsExecutionHandlerInternal = scheduledTxsStub

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithDataComponents(dataComponents),
		node.WithProcessComponents(processComponents),
	)

	t.Run("blockRootHash is set", func(t *testing.T) {
		options, err := n.AddBlockCoordinatesToAccountQueryOptions(api.AccountQueryOptions{
			BlockRootHash: []byte("blockRootHash"),
			BlockNonce:    core.OptionalUint64{Value: 7, HasValue: true},
			BlockHash:     []byte("bbbb"),
		})

		expectedOptions := api.AccountQueryOptions{
			BlockRootHash: []byte("blockRootHash"),
			// When "BlockRootHash" is provided, all other coordinates are ignored and reset.
		}

		require.Nil(t, err)
		require.Equal(t, expectedOptions, options)
	})

	t.Run("blockHash is set (without scheduled)", func(t *testing.T) {
		getScheduledRootHashForHeaderResult = []byte{}
		getScheduledRootHashForHeaderError = errors.New("missing")

		options, err := n.AddBlockCoordinatesToAccountQueryOptions(api.AccountQueryOptions{
			BlockHash: []byte("blockHash"),
		})

		expectedOptions := api.AccountQueryOptions{
			BlockHash: []byte("blockHash"),
			// When "BlockHash" is provided, "BlockNonce" and "BlockRootHash" will be populated in the output, as well
			BlockRootHash: []byte("blockRootHash"),
			BlockNonce:    core.OptionalUint64{Value: 42, HasValue: true},
		}

		require.Nil(t, err)
		require.Equal(t, expectedOptions, options)
	})

	t.Run("blockHash is set (with scheduled)", func(t *testing.T) {
		getScheduledRootHashForHeaderResult = []byte("scheduledBlockRootHash")
		getScheduledRootHashForHeaderError = nil

		options, err := n.AddBlockCoordinatesToAccountQueryOptions(api.AccountQueryOptions{
			BlockHash: []byte("blockHash"),
		})

		expectedOptions := api.AccountQueryOptions{
			BlockHash: []byte("blockHash"),
			// When "BlockHash" is provided, "BlockNonce" and "BlockRootHash" will be populated in the output, as well
			BlockRootHash: []byte("scheduledBlockRootHash"),
			BlockNonce:    core.OptionalUint64{Value: 42, HasValue: true},
		}

		require.Nil(t, err)
		require.Equal(t, expectedOptions, options)
	})

	t.Run("blockNonce is set (without scheduled)", func(t *testing.T) {
		getScheduledRootHashForHeaderResult = []byte{}
		getScheduledRootHashForHeaderError = errors.New("missing")

		options, err := n.AddBlockCoordinatesToAccountQueryOptions(api.AccountQueryOptions{
			BlockNonce: core.OptionalUint64{Value: 42, HasValue: true},
		})

		expectedOptions := api.AccountQueryOptions{
			BlockHash: []byte("blockHash"),
			// When "BlockNonce" is provided, "BlockNonce" and "BlockRootHash" will be populated in the output, as well
			BlockRootHash: []byte("blockRootHash"),
			BlockNonce:    core.OptionalUint64{Value: 42, HasValue: true},
		}

		require.Nil(t, err)
		require.Equal(t, expectedOptions, options)
	})

	t.Run("blockNonce is set (with scheduled)", func(t *testing.T) {
		getScheduledRootHashForHeaderResult = []byte("scheduledBlockRootHash")
		getScheduledRootHashForHeaderError = nil

		options, err := n.AddBlockCoordinatesToAccountQueryOptions(api.AccountQueryOptions{
			BlockNonce: core.OptionalUint64{Value: 42, HasValue: true},
		})

		expectedOptions := api.AccountQueryOptions{
			BlockHash: []byte("blockHash"),
			// When "BlockNonce" is provided, "BlockNonce" and "BlockRootHash" will be populated in the output, as well
			BlockRootHash: []byte("scheduledBlockRootHash"),
			BlockNonce:    core.OptionalUint64{Value: 42, HasValue: true},
		}

		require.Nil(t, err)
		require.Equal(t, expectedOptions, options)
	})
}

func TestMergeAccountQueryOptionsIntoBlockInfo(t *testing.T) {
	mergedInfo := node.MergeAccountQueryOptionsIntoBlockInfo(
		api.AccountQueryOptions{
			BlockNonce: core.OptionalUint64{Value: 7, HasValue: true},
		},
		holders.NewBlockInfo(nil, 0, []byte("rootHash")),
	)

	require.Equal(t, holders.NewBlockInfo(nil, 7, []byte("rootHash")), mergedInfo)

	mergedInfo = node.MergeAccountQueryOptionsIntoBlockInfo(
		api.AccountQueryOptions{
			BlockHash:  []byte("blockHash"),
			BlockNonce: core.OptionalUint64{Value: 7, HasValue: true},
		},
		holders.NewBlockInfo(nil, 0, []byte("rootHash")),
	)

	require.Equal(t, holders.NewBlockInfo([]byte("blockHash"), 7, []byte("rootHash")), mergedInfo)

	mergedInfo = node.MergeAccountQueryOptionsIntoBlockInfo(
		api.AccountQueryOptions{
			BlockHash:     []byte("blockHash"),
			BlockNonce:    core.OptionalUint64{Value: 7, HasValue: true},
			BlockRootHash: []byte("rootHash"),
		},
		holders.NewBlockInfo(nil, 0, nil),
	)

	require.Equal(t, holders.NewBlockInfo([]byte("blockHash"), 7, []byte("rootHash")), mergedInfo)
}

func TestExtractApiBlockInfoIfErrAccountNotFoundAtBlock(t *testing.T) {
	arbitraryError := errors.New("arbitraryError")
	errAccountNotFoundAtBlockNil := state.NewErrAccountNotFoundAtBlock(nil)
	errAccountNotFoundAtBlockWithRootHash := state.NewErrAccountNotFoundAtBlock(holders.NewBlockInfo(nil, 0, []byte{0xaa, 0xbb}))
	errAccountNotFoundAtBlockWithAllCoordinates := state.NewErrAccountNotFoundAtBlock(holders.NewBlockInfo([]byte{0xcc, 0xdd}, 7, []byte{0xaa, 0xbb}))

	apiBlockInfo, ok := node.ExtractApiBlockInfoIfErrAccountNotFoundAtBlock(arbitraryError)
	require.Equal(t, api.BlockInfo{}, apiBlockInfo)
	require.False(t, ok)

	apiBlockInfo, ok = node.ExtractApiBlockInfoIfErrAccountNotFoundAtBlock(errAccountNotFoundAtBlockNil)
	require.Equal(t, api.BlockInfo{}, apiBlockInfo)
	require.True(t, ok)

	apiBlockInfo, ok = node.ExtractApiBlockInfoIfErrAccountNotFoundAtBlock(errAccountNotFoundAtBlockWithRootHash)
	require.Equal(t, api.BlockInfo{RootHash: "aabb"}, apiBlockInfo)
	require.True(t, ok)

	apiBlockInfo, ok = node.ExtractApiBlockInfoIfErrAccountNotFoundAtBlock(errAccountNotFoundAtBlockWithAllCoordinates)
	require.Equal(t, api.BlockInfo{Hash: "ccdd", Nonce: 7, RootHash: "aabb"}, apiBlockInfo)
	require.True(t, ok)
}
