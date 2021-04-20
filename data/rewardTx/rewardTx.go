//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. rewardTx.proto
package rewardTx

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

var _ = data.TransactionHandler(&RewardTx{})

// IsInterfaceNil verifies if underlying object is nil
func (rtx *RewardTx) IsInterfaceNil() bool {
	return rtx == nil
}

// SetValue sets the value of the transaction
func (rtx *RewardTx) SetValue(value *big.Int) {
	rtx.Value = value
}

// GetNonce returns 0 as reward transactions do not have a nonce
func (rtx *RewardTx) GetNonce() uint64 {
	return 0
}

// GetData returns the data of the reward transaction
func (rtx *RewardTx) GetData() []byte {
	return []byte("")
}

// GetSndAddr returns the sender address from the reward transaction
func (rtx *RewardTx) GetSndAddr() []byte {
	return nil
}

// GetGasLimit returns the gas limit of the smart reward transaction
func (rtx *RewardTx) GetGasLimit() uint64 {
	return 0
}

// GetGasPrice returns the gas price of the smart reward transaction
func (rtx *RewardTx) GetGasPrice() uint64 {
	return 0
}

// SetData sets the data of the reward transaction
func (rtx *RewardTx) SetData(_ []byte) {
}

// SetRcvAddr sets the receiver address of the reward transaction
func (rtx *RewardTx) SetRcvAddr(addr []byte) {
	rtx.RcvAddr = addr
}

// SetSndAddr sets the sender address of the reward transaction
func (rtx *RewardTx) SetSndAddr(_ []byte) {
}

// GetRcvUserName returns the receiver user name from the reward transaction
func (rtx *RewardTx) GetRcvUserName() []byte {
	return nil
}

// CheckIntegrity checks for not nil fields and negative value
func (rtx *RewardTx) CheckIntegrity() error {
	if len(rtx.RcvAddr) == 0 {
		return data.ErrNilRcvAddr
	}
	if rtx.Value == nil {
		return data.ErrNilValue
	}
	if rtx.Value.Sign() < 0 {
		return data.ErrNegativeValue
	}

	return nil
}
