package node

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

func (n *Node) loadUserAccountHandlerByAddress(address string, options api.AccountQueryOptions) (state.UserAccountHandler, api.BlockInfo, error) {
	pubKey, err := n.decodeAddressToPubKey(address)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	return n.loadUserAccountHandlerByPubKey(pubKey, options)
}

func (n *Node) loadUserAccountHandlerByPubKey(pubKey []byte, options api.AccountQueryOptions) (state.UserAccountHandler, api.BlockInfo, error) {
	repository := n.stateComponents.AccountsRepository()

	var account vmcommon.AccountHandler
	var accountBlockInfo state.AccountBlockInfo
	var err error

	if options.OnFinalBlock {
		account, accountBlockInfo, err = repository.GetAccountOnFinal(pubKey)
	} else if options.OnStartOfEpoch != 0 {
		return nil, api.BlockInfo{}, ErrNotImplemented
	} else {
		account, accountBlockInfo, err = repository.GetAccountOnCurrent(pubKey)
	}
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	userAccount, err := n.castAccountToUserAccount(account)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	return userAccount, accountBlockInfoToApiResource(accountBlockInfo), nil
}

func (n *Node) loadAccountCode(codeHash []byte, options api.AccountQueryOptions) ([]byte, api.BlockInfo) {
	repository := n.stateComponents.AccountsRepository()

	var code []byte
	var accountBlockInfo state.AccountBlockInfo

	if options.OnFinalBlock {
		code, accountBlockInfo = repository.GetCodeOnFinal(codeHash)
	} else if options.OnStartOfEpoch != 0 {
		return nil, api.BlockInfo{}
	} else {
		code, accountBlockInfo = repository.GetCodeOnCurrent(codeHash)
	}

	return code, accountBlockInfoToApiResource(accountBlockInfo)
}

func accountBlockInfoToApiResource(info state.AccountBlockInfo) api.BlockInfo {
	if check.IfNil(info) {
		return api.BlockInfo{}
	}

	return api.BlockInfo{
		Nonce:    info.GetNonce(),
		Hash:     hex.EncodeToString(info.GetHash()),
		RootHash: hex.EncodeToString(info.GetRootHash()),
	}
}
