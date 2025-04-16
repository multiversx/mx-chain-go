package stateAccesses

import (
	"bytes"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data/stateChange"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type userAccountHandler interface {
	GetCodeMetadata() []byte
	GetCodeHash() []byte
	GetRootHash() []byte
	GetBalance() *big.Int
	GetDeveloperReward() *big.Int
	GetOwnerAddress() []byte
	GetUserName() []byte
	vmcommon.AccountHandler
}

func getAccountChanges(oldAcc, newAcc vmcommon.AccountHandler) *stateChange.AccountChanges {
	baseNewAcc, newAccOk := newAcc.(userAccountHandler)
	if !newAccOk {
		return nil
	}
	baseOldAccount, oldAccOk := oldAcc.(userAccountHandler)
	if !oldAccOk {
		return nil
	}

	accountChanges := &stateChange.AccountChanges{}

	if baseNewAcc.GetNonce() != baseOldAccount.GetNonce() {
		accountChanges.Nonce = true
	}

	if baseNewAcc.GetBalance().Uint64() != baseOldAccount.GetBalance().Uint64() {
		accountChanges.Balance = true
	}

	if !bytes.Equal(baseNewAcc.GetCodeHash(), baseOldAccount.GetCodeHash()) {
		accountChanges.CodeHash = true
	}

	if !bytes.Equal(baseNewAcc.GetRootHash(), baseOldAccount.GetRootHash()) {
		accountChanges.RootHash = true
	}

	if !bytes.Equal(baseNewAcc.GetDeveloperReward().Bytes(), baseOldAccount.GetDeveloperReward().Bytes()) {
		accountChanges.DeveloperReward = true
	}

	if !bytes.Equal(baseNewAcc.GetOwnerAddress(), baseOldAccount.GetOwnerAddress()) {
		accountChanges.OwnerAddress = true
	}

	if !bytes.Equal(baseNewAcc.GetUserName(), baseOldAccount.GetUserName()) {
		accountChanges.UserName = true
	}

	if !bytes.Equal(baseNewAcc.GetCodeMetadata(), baseOldAccount.GetCodeMetadata()) {
		accountChanges.CodeMetadata = true
	}

	return accountChanges
}
