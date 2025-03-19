package stateChanges

import (
	"bytes"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type dataAnalysisStateChangeDTO struct {
	state.StateChange
	Operation       string `json:"operation"`
	Nonce           bool   `json:"nonceChanged"`
	Balance         bool   `json:"balanceChanged"`
	CodeHash        bool   `json:"codeHashChanged"`
	RootHash        bool   `json:"rootHashChanged"`
	DeveloperReward bool   `json:"developerRewardChanged"`
	OwnerAddress    bool   `json:"ownerAddressChanged"`
	UserName        bool   `json:"userNameChanged"`
	CodeMetadata    bool   `json:"codeMetadataChanged"`
}

type dataAnalysisStateChangesForTx struct {
	StateChangesForTx
	Tx data.TransactionHandler `json:"tx"`
}

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

func checkAccountChanges(oldAcc, newAcc vmcommon.AccountHandler, stateChange *dataAnalysisStateChangeDTO) {
	baseNewAcc, newAccOk := newAcc.(userAccountHandler)
	if !newAccOk {
		return
	}
	baseOldAccount, oldAccOk := oldAcc.(userAccountHandler)
	if !oldAccOk {
		return
	}

	if baseNewAcc.GetNonce() != baseOldAccount.GetNonce() {
		stateChange.Nonce = true
	}

	if baseNewAcc.GetBalance().Uint64() != baseOldAccount.GetBalance().Uint64() {
		stateChange.Balance = true
	}

	if !bytes.Equal(baseNewAcc.GetCodeHash(), baseOldAccount.GetCodeHash()) {
		stateChange.CodeHash = true
	}

	if !bytes.Equal(baseNewAcc.GetRootHash(), baseOldAccount.GetRootHash()) {
		stateChange.RootHash = true
	}

	if !bytes.Equal(baseNewAcc.GetDeveloperReward().Bytes(), baseOldAccount.GetDeveloperReward().Bytes()) {
		stateChange.DeveloperReward = true
	}

	if !bytes.Equal(baseNewAcc.GetOwnerAddress(), baseOldAccount.GetOwnerAddress()) {
		stateChange.OwnerAddress = true
	}

	if !bytes.Equal(baseNewAcc.GetUserName(), baseOldAccount.GetUserName()) {
		stateChange.UserName = true
	}

	if !bytes.Equal(baseNewAcc.GetCodeMetadata(), baseOldAccount.GetCodeMetadata()) {
		stateChange.CodeMetadata = true
	}
}
