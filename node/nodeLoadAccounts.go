package node

import (
	"encoding/hex"
	"errors"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/state"
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

func (n *Node) loadUserAccountHandler(address string, options api.AccountQueryOptions) (state.UserAccountHandler, api.BlockInfo, error) {
	// Question for review: do we require this special lazy initialization logic?
	componentsNotInitialized := check.IfNil(n.stateComponents.AccountsRepository())
	if componentsNotInitialized {
		return nil, api.BlockInfo{}, errors.New("AccountsRepository not initialized")
	}

	publicKey, err := n.decodeAddressToPubKey(address)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	blockInfo, err := n.getBlockInfoGivenOptions(options)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	account, err := n.stateComponents.AccountsRepository().GetExistingAccount(publicKey, blockInfo.rootHash)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	userAccount, ok := n.castAccountToUserAccount(account)
	if !ok {
		return nil, api.BlockInfo{}, ErrCannotCastAccountHandlerToUserAccountHandler
	}

	return userAccount, blockInfo.toApiResource(), nil
}

func (n *Node) getBlockInfoGivenOptions(options api.AccountQueryOptions) (blockInfo, error) {
	// Get info of "highest final block"
	if options.OnFinalBlock {
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

	// Get info of "start of epoch" block
	if options.OnStartOfEpoch != 0 {
		// TODO: Implement this feature at a later time.
		return blockInfo{}, ErrNotImplemented
	}

	// Fallback to info of "current block"

	// TODO: Fix possible race conditions (blockchain.go does not provide a load x 3 method; the following loads do not take place in a critical section).
	// A possible fix would be to add a function such as blockchain.GetCurrentBlockInfo().
	block := n.dataComponents.Blockchain().GetCurrentBlockHeader()
	if check.IfNil(block) {
		return blockInfo{}, ErrBlockInfoNotAvailable
	}

	hash := n.dataComponents.Blockchain().GetCurrentBlockHeaderHash()
	rootHash := n.dataComponents.Blockchain().GetCurrentBlockRootHash()
	if len(hash) == 0 || len(rootHash) == 0 {
		return blockInfo{}, ErrBlockInfoNotAvailable
	}

	return blockInfo{
		nonce:    block.GetNonce(),
		hash:     hash,
		rootHash: rootHash,
	}, nil
}
