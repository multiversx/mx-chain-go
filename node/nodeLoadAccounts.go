package node

import (
	"encoding/hex"
	"errors"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/holders"
	"github.com/ElrondNetwork/elrond-go/state"
)

func (n *Node) loadUserAccountHandlerByAddress(address string, options api.AccountQueryOptions) (state.UserAccountHandler, api.BlockInfo, error) {
	pubKey, err := n.decodeAddressToPubKey(address)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	return n.loadUserAccountHandlerByPubKey(pubKey, options)
}

func (n *Node) loadUserAccountHandlerByPubKey(pubKey []byte, options api.AccountQueryOptions) (state.UserAccountHandler, api.BlockInfo, error) {
	options, err := n.addBlockCoordinatesToAccountQueryOptions(options)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	repository := n.stateComponents.AccountsRepository()

	account, blockInfo, err := repository.GetAccountWithBlockInfo(pubKey, options)
	if err != nil {
		blockInfo, ok := extractBlockInfoIfErrAccountNotFoundAtBlock(err)
		if ok {
			blockInfo = mergeAccountQueryOptionsIntoBlockInfo(options, blockInfo)
			// Return the same error (now with additional block info)
			return nil, api.BlockInfo{}, state.NewErrAccountNotFoundAtBlock(blockInfo)
		}

		return nil, api.BlockInfo{}, err
	}

	userAccount, err := n.castAccountToUserAccount(account)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	blockInfo = mergeAccountQueryOptionsIntoBlockInfo(options, blockInfo)
	return userAccount, accountBlockInfoToApiResource(blockInfo), nil
}

func (n *Node) loadAccountCode(codeHash []byte, options api.AccountQueryOptions) ([]byte, api.BlockInfo) {
	options, err := n.addBlockCoordinatesToAccountQueryOptions(options)
	if err != nil {
		return nil, api.BlockInfo{}
	}

	repository := n.stateComponents.AccountsRepository()

	code, blockInfo, err := repository.GetCodeWithBlockInfo(codeHash, options)
	if err != nil {
		log.Warn("Node.loadAccountCode", "error", err)
		return nil, api.BlockInfo{}
	}

	blockInfo = mergeAccountQueryOptionsIntoBlockInfo(options, blockInfo)
	return code, accountBlockInfoToApiResource(blockInfo)
}

func mergeAccountQueryOptionsIntoBlockInfo(options api.AccountQueryOptions, info common.BlockInfo) common.BlockInfo {
	if check.IfNil(info) {
		return nil
	}

	blockNonce := info.GetNonce()
	blockHash := info.GetHash()
	blockRootHash := info.GetRootHash()

	if blockNonce == 0 && options.BlockNonce.HasValue {
		blockNonce = options.BlockNonce.Value
	}
	if len(blockHash) == 0 && len(options.BlockHash) > 0 {
		blockHash = options.BlockHash
	}
	if len(blockRootHash) == 0 && len(options.BlockRootHash) > 0 {
		blockRootHash = options.BlockRootHash
	}

	return holders.NewBlockInfo(blockHash, blockNonce, blockRootHash)
}

func accountBlockInfoToApiResource(info common.BlockInfo) api.BlockInfo {
	if check.IfNil(info) {
		return api.BlockInfo{}
	}

	return api.BlockInfo{
		Nonce:    info.GetNonce(),
		Hash:     hex.EncodeToString(info.GetHash()),
		RootHash: hex.EncodeToString(info.GetRootHash()),
	}
}

// Important note about "AccountQueryOptions.HintEpoch": for blocks right after an epoch change, we will actually need `epoch - 1`
// for the purpose of historical lookup (which involves recreation of tries).
// However, since the current implementation of "recreate trie in epoch N" also looks for data in epoch N - 1 (on different purposes),
// it's sufficient (and non-ambigous) to set "HintEpoch" = N here.
func (n *Node) addBlockCoordinatesToAccountQueryOptions(options api.AccountQueryOptions) (api.AccountQueryOptions, error) {
	blockNonce := options.BlockNonce
	blockHash := options.BlockHash
	blockRootHash := options.BlockRootHash

	if len(blockRootHash) > 0 {
		// We cannot infer other block coordinates (hash, nonce, hint for epoch) at this moment
		return api.AccountQueryOptions{
			BlockRootHash: options.BlockRootHash,
		}, nil
	}

	if len(blockHash) > 0 {
		blockHeader, err := n.getBlockHeaderByHash(blockHash)
		if err != nil {
			return api.AccountQueryOptions{}, err
		}

		blockRootHash := n.getBlockRootHash(blockHash, blockHeader)

		return api.AccountQueryOptions{
			BlockHash:     options.BlockHash,
			BlockNonce:    core.OptionalUint64{Value: blockHeader.GetNonce(), HasValue: true},
			BlockRootHash: blockRootHash,
			HintEpoch:     core.OptionalUint32{Value: blockHeader.GetEpoch(), HasValue: true},
		}, nil
	}

	if blockNonce.HasValue {
		blockHeader, blockHash, err := n.getBlockHeaderByNonce(blockNonce.Value)
		if err != nil {
			return api.AccountQueryOptions{}, err
		}

		blockRootHash := n.getBlockRootHash(blockHash, blockHeader)

		return api.AccountQueryOptions{
			BlockHash:     blockHash,
			BlockNonce:    core.OptionalUint64{Value: blockHeader.GetNonce(), HasValue: true},
			BlockRootHash: blockRootHash,
			HintEpoch:     core.OptionalUint32{Value: blockHeader.GetEpoch(), HasValue: true},
		}, nil
	}

	return options, nil
}

func extractApiBlockInfoIfErrAccountNotFoundAtBlock(err error) (api.BlockInfo, bool) {
	blockInfo, ok := extractBlockInfoIfErrAccountNotFoundAtBlock(err)
	if ok {
		return accountBlockInfoToApiResource(blockInfo), true
	}
	return api.BlockInfo{}, false
}

func extractBlockInfoIfErrAccountNotFoundAtBlock(err error) (common.BlockInfo, bool) {
	var accountNotFoundErr *state.ErrAccountNotFoundAtBlock
	if errors.As(err, &accountNotFoundErr) {
		return accountNotFoundErr.BlockInfo, true
	}
	return nil, false
}
