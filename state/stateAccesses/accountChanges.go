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

func getAccountChanges(oldAcc, newAcc vmcommon.AccountHandler) uint32 {
	baseNewAcc, newAccOk := newAcc.(userAccountHandler)
	if !newAccOk {
		return stateChange.NoChange
	}
	baseOldAccount, oldAccOk := oldAcc.(userAccountHandler)
	if !oldAccOk {
		return stateChange.NoChange
	}

	accountChanges := stateChange.NoChange

	if baseNewAcc.GetNonce() != baseOldAccount.GetNonce() {
		accountChanges = stateChange.AddAccountChange(accountChanges, stateChange.NonceChanged)
	}

	if baseNewAcc.GetBalance().Uint64() != baseOldAccount.GetBalance().Uint64() {
		accountChanges = stateChange.AddAccountChange(accountChanges, stateChange.BalanceChanged)
	}

	if !bytes.Equal(baseNewAcc.GetCodeHash(), baseOldAccount.GetCodeHash()) {
		accountChanges = stateChange.AddAccountChange(accountChanges, stateChange.CodeHashChanged)
	}

	if !bytes.Equal(baseNewAcc.GetRootHash(), baseOldAccount.GetRootHash()) {
		accountChanges = stateChange.AddAccountChange(accountChanges, stateChange.RootHashChanged)
	}

	if !bytes.Equal(baseNewAcc.GetDeveloperReward().Bytes(), baseOldAccount.GetDeveloperReward().Bytes()) {
		accountChanges = stateChange.AddAccountChange(accountChanges, stateChange.DeveloperRewardChanged)
	}

	if !bytes.Equal(baseNewAcc.GetOwnerAddress(), baseOldAccount.GetOwnerAddress()) {
		accountChanges = stateChange.AddAccountChange(accountChanges, stateChange.OwnerAddressChanged)
	}

	if !bytes.Equal(baseNewAcc.GetUserName(), baseOldAccount.GetUserName()) {
		accountChanges = stateChange.AddAccountChange(accountChanges, stateChange.UserNameChanged)
	}

	if !bytes.Equal(baseNewAcc.GetCodeMetadata(), baseOldAccount.GetCodeMetadata()) {
		accountChanges = stateChange.AddAccountChange(accountChanges, stateChange.CodeMetadataChanged)
	}

	return accountChanges
}
