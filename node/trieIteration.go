package node

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/checkpoint"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/state"
)

// WithTrieOperations configures the limits of the paginated trie walks served by the API
func WithTrieOperations(cfg config.TrieOperationsConfig) Option {
	return func(n *Node) error {
		n.trieOperationsConfig = cfg.WithDefaults()
		return nil
	}
}

// maxIteratorStateEntries bounds the number of pending nodes accepted in a checkpoint. A DFS frontier holds at most
// 17 siblings (the children of a branch) per trie level, so a genuine checkpoint stays far below this limit
const maxIteratorStateEntries = 4096

// trieIterationRequest describes a page of a paginated walk over the data trie of an account
type trieIterationRequest struct {
	address    string // the bech32 address, decoded when pubKey is not set
	pubKey     []byte
	options    api.AccountQueryOptions
	numKeys    uint
	maxBytes   uint64 // zero: the configured limit
	checkpoint *checkpoint.Checkpoint
	leafParser common.TrieLeafParser // nil: the account's own
}

type dataTrieWithStorage interface {
	GetStorageManager() common.StorageManager
}

// loadedAccount is an account loaded for a page of a paginated walk: the account handler (nil for an account that
// does not exist at that block), the block info the page belongs to and the options the account was loaded with
type loadedAccount struct {
	userAccount state.UserAccountHandler
	blockInfo   api.BlockInfo
	options     api.AccountQueryOptions
}

// iterateDataTrie loads the account at the block recorded in the checkpoint (or at the block selected by the caller
// options for a first page), takes a trie operations slot and walks the data trie from the checkpoint, calling visit
// for every leaf. It returns the block info the leaves belong to and the checkpoint to continue from, nil when the
// walk finished
func (n *Node) iterateDataTrie(ctx context.Context, req trieIterationRequest, visit common.LeafVisitor) (api.BlockInfo, *checkpoint.Checkpoint, error) {
	loaded, err := n.loadAccountForIteration(req)
	if err != nil {
		return api.BlockInfo{}, nil, err
	}

	next, err := n.walkDataTrie(ctx, loaded, req, visit)
	if err != nil {
		return api.BlockInfo{}, nil, err
	}

	return loaded.blockInfo, next, nil
}

// loadAccountForIteration loads the account for a page: at the block recorded in the checkpoint when continuing,
// else at the block selected by the caller options
func (n *Node) loadAccountForIteration(req trieIterationRequest) (loadedAccount, error) {
	options, err := loadOptionsForIteration(req)
	if err != nil {
		return loadedAccount{}, err
	}

	pubKey := req.pubKey
	if len(pubKey) == 0 {
		pubKey, err = n.decodeAddressToPubKey(req.address)
		if err != nil {
			return loadedAccount{}, err
		}
	}

	userAccount, blockInfo, loadedOptions, err := n.loadUserAccountHandlerByPubKeyWithOptions(pubKey, options)
	if err != nil {
		adaptedBlockInfo, isEmptyAccount := extractBlockInfoIfNewAccount(err)
		if !isEmptyAccount {
			if hasRecordedBlock(req.checkpoint) && core.IsGetNodeFromDBError(err) {
				// the recorded state can not be read any more (pruned since the first page): the walk must restart
				return loadedAccount{}, fmt.Errorf("%w: %v", common.ErrIteratorNodeUnavailable, err)
			}
			return loadedAccount{}, err
		}

		// no account at that block: the page is empty, the state it belongs to is kept (the options resolved from
		// the checkpoint or given by the caller, the recorded block below)
		userAccount, blockInfo = nil, adaptedBlockInfo
	} else {
		options = loadedOptions
	}
	if hasRecordedBlock(req.checkpoint) {
		// loaded by the recorded root hash: report the block recorded on the first page
		blockInfo = api.BlockInfo{Nonce: req.checkpoint.BlockNonce, Hash: req.checkpoint.BlockHash, RootHash: req.checkpoint.BlockRootHash}
	}

	return loadedAccount{userAccount: userAccount, blockInfo: blockInfo, options: options}, nil
}

// optionsForSameState returns the options that load another account from the very state the loaded account was
// served from: its root hash and epoch, the ones a checkpoint records. Without a known root hash, the load options
func (n *Node) optionsForSameState(loaded loadedAccount) api.AccountQueryOptions {
	rootHash, err := hex.DecodeString(loaded.blockInfo.RootHash)
	if err != nil || len(rootHash) == 0 {
		return loaded.options
	}

	return api.AccountQueryOptions{
		BlockRootHash: rootHash,
		HintEpoch:     core.OptionalUint32{Value: n.epochForCheckpoint(loaded.blockInfo, loaded.options), HasValue: true},
	}
}

// walkDataTrie takes a trie operations slot and walks the data trie of the loaded account from the checkpoint,
// calling visit for every leaf. It returns the checkpoint to continue from, nil when the walk finished
func (n *Node) walkDataTrie(ctx context.Context, loaded loadedAccount, req trieIterationRequest, visit common.LeafVisitor) (*checkpoint.Checkpoint, error) {
	if check.IfNil(loaded.userAccount) || check.IfNil(loaded.userAccount.DataTrie()) {
		return nil, nil
	}

	dataTrie, ok := loaded.userAccount.DataTrie().(dataTrieWithStorage)
	if !ok {
		return nil, fmt.Errorf("cannot cast data trie to dataTrieWithStorage")
	}

	account, ok := loaded.userAccount.(userAccountWithLeavesParser)
	if !ok {
		return nil, fmt.Errorf("cannot cast user account to userAccountWithLeavesParser")
	}

	iteratorState := [][]byte{loaded.userAccount.GetRootHash()}
	if req.checkpoint != nil && len(req.checkpoint.IteratorState) > 0 {
		iteratorState = req.checkpoint.IteratorState
	} else if common.IsEmptyTrie(iteratorState[0]) {
		return nil, nil // a data trie whose keys were all deleted has the empty trie hash as root and no node
	}

	release, err := n.acquireTrieOperationsSlot(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

	leafParser := account.GetLeavesParser()
	if req.leafParser != nil {
		leafParser = req.leafParser
	}

	nextState, err := n.stateComponents.TrieLeavesRetriever().IterateLeaves(
		n.numLeavesForPage(req.numKeys),
		n.bytesForPage(req.maxBytes),
		iteratorState,
		dataTrie.GetStorageManager(),
		leafParser,
		ctx,
		visit,
	)
	if err != nil {
		return nil, err
	}
	if len(nextState) == 0 {
		return nil, nil
	}

	return newCheckpoint(loaded.blockInfo, n.epochForCheckpoint(loaded.blockInfo, loaded.options), nextState), nil
}

// epochForCheckpoint returns the epoch the state of a page is read from: the one resolved from the block coordinates
// when the account was loaded, else the epoch of the block the state belongs to (the last executed block kept in
// memory, or the stored header), else the current epoch. It is recorded so a continuation page is read from the
// same epoch storage, whatever other queries changed in between
func (n *Node) epochForCheckpoint(blockInfo api.BlockInfo, resolvedOptions api.AccountQueryOptions) uint32 {
	if resolvedOptions.HintEpoch.HasValue {
		return resolvedOptions.HintEpoch.Value
	}

	blockHash, err := hex.DecodeString(blockInfo.Hash)
	if err != nil || len(blockHash) == 0 {
		return n.coreComponents.EpochNotifier().CurrentEpoch()
	}

	lastExecutedHeader, lastExecutedHash := n.getLastExecutedBlock()
	if bytes.Equal(blockHash, lastExecutedHash) && !check.IfNil(lastExecutedHeader) {
		return lastExecutedHeader.GetEpoch()
	}

	header, err := n.getBlockHeaderByHash(blockHash)
	if err == nil {
		return header.GetEpoch()
	}
	log.Debug("epochForCheckpoint: block header not found, using the current epoch", "hash", blockInfo.Hash, "error", err)

	return n.coreComponents.EpochNotifier().CurrentEpoch()
}

func hasRecordedBlock(cp *checkpoint.Checkpoint) bool {
	return cp != nil && cp.BlockRootHash != ""
}

// loadOptionsForIteration returns the options the account is loaded with: the caller options for a first page, the
// root hash recorded in the checkpoint when continuing (the very root the first page was served from, so no block
// coordinate has to be derived again). Caller coordinates that disagree with the checkpoint are rejected
func loadOptionsForIteration(req trieIterationRequest) (api.AccountQueryOptions, error) {
	cp := req.checkpoint
	if cp == nil {
		return req.options, nil
	}
	if len(cp.IteratorState) > maxIteratorStateEntries {
		return api.AccountQueryOptions{}, fmt.Errorf("%w: too many iterator state entries", common.ErrInvalidIteratorState)
	}
	if !hasRecordedBlock(cp) {
		// no block recorded (e.g. the legacy iterate-keys state): continue at the block selected by the caller
		return req.options, nil
	}

	err := checkOptionsAgainstCheckpoint(req.options, cp)
	if err != nil {
		return api.AccountQueryOptions{}, err
	}

	options := api.AccountQueryOptions{}
	options.BlockRootHash, err = hex.DecodeString(cp.BlockRootHash)
	if err != nil {
		return api.AccountQueryOptions{}, fmt.Errorf("%w: block root hash: %v", common.ErrInvalidIteratorState, err)
	}
	options.HintEpoch = req.options.HintEpoch
	if cp.Epoch != nil {
		options.HintEpoch = core.OptionalUint32{Value: *cp.Epoch, HasValue: true}
	}

	return options, nil
}

// checkOptionsAgainstCheckpoint rejects concrete caller block coordinates that contradict the block recorded in the
// checkpoint. A continuation page is always served from the recorded block, so the caller may omit the coordinates or
// repeat the ones of the first page. The block selectors (onFinalBlock, onStartOfEpoch) only apply to the first page:
// on a continuation they are ignored, as the recorded block wins
func checkOptionsAgainstCheckpoint(options api.AccountQueryOptions, cp *checkpoint.Checkpoint) error {
	nonceDiffers := options.BlockNonce.HasValue && options.BlockNonce.Value != cp.BlockNonce
	hashDiffers := len(options.BlockHash) > 0 && hex.EncodeToString(options.BlockHash) != cp.BlockHash
	rootHashDiffers := len(options.BlockRootHash) > 0 && hex.EncodeToString(options.BlockRootHash) != cp.BlockRootHash
	if nonceDiffers || hashDiffers || rootHashDiffers {
		return fmt.Errorf("%w: block coordinates differ from the ones recorded in the iterator state", common.ErrInvalidIteratorState)
	}

	epochDiffers := options.HintEpoch.HasValue && cp.Epoch != nil && options.HintEpoch.Value != *cp.Epoch
	if epochDiffers {
		return fmt.Errorf("%w: hint epoch differs from the one recorded in the iterator state", common.ErrInvalidIteratorState)
	}

	return nil
}

func newCheckpoint(blockInfo api.BlockInfo, epoch uint32, iteratorState [][]byte) *checkpoint.Checkpoint {
	return &checkpoint.Checkpoint{
		BlockNonce:    blockInfo.Nonce,
		BlockHash:     blockInfo.Hash,
		BlockRootHash: blockInfo.RootHash,
		Epoch:         &epoch,
		IteratorState: iteratorState,
	}
}

type lastExecutedBlockProvider interface {
	GetLastExecutedBlockHeaderAndInfo() (data.HeaderHandler, []byte, []byte)
}

// getLastExecutedBlock returns the last executed block header and its hash, read together when the chain handler
// allows it so they always belong to the same block
func (n *Node) getLastExecutedBlock() (data.HeaderHandler, []byte) {
	chainHandler := n.dataComponents.Blockchain()
	provider, ok := chainHandler.(lastExecutedBlockProvider)
	if ok {
		header, headerHash, _ := provider.GetLastExecutedBlockHeaderAndInfo()
		return header, headerHash
	}

	header := chainHandler.GetLastExecutedBlockHeader()
	_, headerHash, _ := chainHandler.GetLastExecutedBlockInfo()

	return header, headerHash
}

// acquireTrieOperationsSlot takes one of the slots that bound the concurrent trie walks; the returned function
// releases it
func (n *Node) acquireTrieOperationsSlot(ctx context.Context) (func(), error) {
	n.trieOperationsOnce.Do(func() {
		n.trieOperationsConfig = n.trieOperationsConfig.WithDefaults()
		n.trieOperationsSlots = make(chan struct{}, n.trieOperationsConfig.MaxConcurrentRequests)
	})
	if ctx.Err() != nil {
		return nil, ErrTrieOperationsTimeout
	}
	select {
	case n.trieOperationsSlots <- struct{}{}:
	default:
		return nil, common.ErrTrieOperationsBusy
	}

	return func() { <-n.trieOperationsSlots }, nil
}

// numLeavesForPage bounds the number of leaves of a page by MaxLeaves, which is also the default
func (n *Node) numLeavesForPage(numKeys uint) int {
	maxLeaves := n.trieOperationsConfig.MaxLeaves
	if numKeys == 0 || uint64(numKeys) > maxLeaves {
		return int(maxLeaves)
	}

	return int(numKeys)
}

// bytesForPage bounds the byte budget of a page by the configured one; zero means the configured one
func (n *Node) bytesForPage(maxBytes uint64) uint64 {
	configured := n.trieOperationsConfig.MaxBytes
	if maxBytes == 0 || maxBytes > configured {
		return configured
	}

	return maxBytes
}

// keyValueCollector stores every leaf as hex key and hex value; a leaf is accepted when its key was not seen before
func keyValueCollector(pairs map[string]string) common.LeafVisitor {
	return func(leaf core.KeyValueHolder) (bool, uint64, error) {
		hexKey := hex.EncodeToString(leaf.Key())
		hexValue := hex.EncodeToString(leaf.Value())
		numPairsBefore := len(pairs)
		pairs[hexKey] = hexValue

		return len(pairs) > numPairsBefore, uint64(len(hexKey) + len(hexValue)), nil
	}
}
