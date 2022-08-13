package node

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/common"
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
	options, err := n.transformAccountQueryOptions(options)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	repository := n.stateComponents.AccountsRepository()

	account, blockInfo, err := repository.GetAccountWithBlockInfo(pubKey, options)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	userAccount, err := n.castAccountToUserAccount(account)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	return userAccount, accountBlockInfoToApiResource(blockInfo), nil
}

func (n *Node) loadAccountCode(codeHash []byte, options api.AccountQueryOptions) ([]byte, api.BlockInfo) {
	options, err := n.transformAccountQueryOptions(options)
	if err != nil {
		return nil, api.BlockInfo{}
	}

	repository := n.stateComponents.AccountsRepository()

	code, blockInfo, err := repository.GetCodeWithBlockInfo(codeHash, options)
	if err != nil {
		log.Warn("Node.loadAccountCode", "error", err)
		return nil, api.BlockInfo{}
	}

	return code, accountBlockInfoToApiResource(blockInfo)
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

func (n *Node) transformAccountQueryOptions(options api.AccountQueryOptions) (api.AccountQueryOptions, error) {
	blockRootHash, err := hex.DecodeString(options.BlockRootHash)
	if err != nil {
		return api.AccountQueryOptions{}, err
	}

	blockHash, err := hex.DecodeString(options.BlockHash)
	if err != nil {
		return api.AccountQueryOptions{}, err
	}

	blockNonce := options.BlockNonce

	if len(blockRootHash) > 0 {
		// We cannot infer other block coordinates (hash, nonce) at this moment
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
			BlockNonce:    blockHeader.GetNonce(),
			BlockRootHash: hex.EncodeToString(blockRootHash),
		}, nil
	}

	// Workaround: at this moment, we cannot check whether "blockNonce" is set in other way than comparing it with zero.
	if blockNonce > 0 {
		blockHeader, blockHash, err := n.getBlockHeaderByNonce(blockNonce)
		if err != nil {
			return api.AccountQueryOptions{}, err
		}

		blockRootHash := n.getBlockRootHash(blockHash, blockHeader)

		return api.AccountQueryOptions{
			BlockHash:     options.BlockHash,
			BlockNonce:    blockHeader.GetNonce(),
			BlockRootHash: hex.EncodeToString(blockRootHash),
		}, nil
	}

	return options, nil
}
