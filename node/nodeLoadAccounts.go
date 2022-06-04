package node

import (
	"encoding/hex"
	"errors"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type blockInfo struct {
	nonce    uint64
	hash     []byte
	rootHash []byte
}

func (info *blockInfo) toApiResource() api.BlockInfo {
	return api.BlockInfo{
		Nonce:    info.nonce,
		Hash:     hex.EncodeToString(info.hash),
		RootHash: hex.EncodeToString(info.rootHash),
	}
}

func (n *Node) loadUserAccountHandlerByAddress(address string, options api.AccountQueryOptions) (state.UserAccountHandler, api.BlockInfo, error) {
	pubKey, err := n.decodeAddressToPubKey(address)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	return n.loadUserAccountHandlerByPubKey(pubKey, options)
}

func (n *Node) loadUserAccountHandlerByPubKey(pubKey []byte, options api.AccountQueryOptions) (state.UserAccountHandler, api.BlockInfo, error) {
	blockInfo, err := n.getBlockInfoGivenOptions(options)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	account, err := n.loadUserAccountOnRootHash(pubKey, blockInfo.rootHash)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	userAccount, ok := n.castAccountToUserAccount(account)
	if !ok {
		return nil, api.BlockInfo{}, ErrCannotCastAccountHandlerToUserAccountHandler
	}

	return userAccount, blockInfo.toApiResource(), nil
}

func (n *Node) loadUserAccountOnRootHash(pubKey []byte, rootHash []byte) (vmcommon.AccountHandler, error) {
	repository, err := n.getAccountsRepository()
	if err != nil {
		return nil, err
	}

	account, err := repository.GetExistingAccount(pubKey, rootHash)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (n *Node) loadAccountCode(codeHash []byte, options api.AccountQueryOptions) ([]byte, api.BlockInfo) {
	blockInfo, err := n.getBlockInfoGivenOptions(options)
	if err != nil {
		return nil, api.BlockInfo{}
	}

	repository, err := n.getAccountsRepository()
	if err != nil {
		return nil, api.BlockInfo{}
	}

	code := repository.GetCode(codeHash, blockInfo.rootHash)
	return code, blockInfo.toApiResource()
}

func (n *Node) getAccountsRepository() (state.AccountsRepository, error) {
	// Question for review: do we still require the special lazy initialization logic (used on some components in node.go)?
	notInitialized := check.IfNil(n.stateComponents.AccountsRepository())
	if notInitialized {
		return nil, errors.New("AccountsRepository not initialized")
	}

	return n.stateComponents.AccountsRepository(), nil
}

func (n *Node) getBlockInfoGivenOptions(options api.AccountQueryOptions) (blockInfo, error) {
	// Get info of "highest final block"
	if options.OnFinalBlock {
		return n.getFinalBlockInfo()
	}

	// Get info of "start of epoch" block
	if options.OnStartOfEpoch != 0 {
		// TODO: Implement this feature at a later time.
		return blockInfo{}, ErrNotImplemented
	}

	// Fallback to info of "current block"
	return n.getCurrentBlockInfo()
}

func (n *Node) getFinalBlockInfo() (blockInfo, error) {
	nonce, hash, rootHash := n.dataComponents.Blockchain().GetFinalBlockInfo()
	if len(hash) == 0 || len(rootHash) == 0 {
		return blockInfo{}, ErrBlockInfoNotAvailable
	}

	return blockInfo{
		nonce:    nonce,
		hash:     hash,
		rootHash: rootHash,
	}, nil
}

func (n *Node) getCurrentBlockInfo() (blockInfo, error) {
	// Question for review: do we still require the special lazy initialization logic (used on some components in node.go)?
	notInitialized := check.IfNil(n.dataComponents.Blockchain())
	if notInitialized {
		return blockInfo{}, errors.New("blockchain (chain handler) not initialized")
	}

	// TODO: Fix possible race conditions (blockchain.go does not provide a load x 3 method; the following loads do not take place in a critical section).
	// A possible fix would be to add a function such as blockchain.GetCurrentBlockInfo().
	block := n.dataComponents.Blockchain().GetCurrentBlockHeader()
	if check.IfNil(block) {
		return blockInfo{}, ErrBlockInfoNotAvailable
	}

	hash := n.dataComponents.Blockchain().GetCurrentBlockHeaderHash()
	if len(hash) == 0 {
		return blockInfo{}, ErrBlockInfoNotAvailable
	}

	rootHash := n.dataComponents.Blockchain().GetCurrentBlockRootHash()
	if len(rootHash) == 0 {
		return blockInfo{}, ErrBlockInfoNotAvailable
	}

	return blockInfo{
		nonce:    block.GetNonce(),
		hash:     hash,
		rootHash: rootHash,
	}, nil
}
