package rewardTx

import (
	"math/big"
)

// RewardTx holds the data for a reward transaction
type RewardTx struct {
	Round   uint64
	Value   *big.Int
	RcvAddr []byte
	ShardId uint32
	Epoch   uint32
}

// IsInterfaceNil verifies if underlying object is nil
func (rtx *RewardTx) IsInterfaceNil() bool {
	return rtx == nil
}

// GetValue returns the value of the reward transaction
func (rtx *RewardTx) GetValue() *big.Int {
	return rtx.Value
}

// GetNonce returns 0 as reward transactions do not have a nonce
func (rtx *RewardTx) GetNonce() uint64 {
	return 0
}

// GetData returns the data of the reward transaction
func (rtx *RewardTx) GetData() []byte {
	return []byte("")
}

// GetRecvAddress returns the receiver address from the reward transaction
func (rtx *RewardTx) GetRecvAddress() []byte {
	return rtx.RcvAddr
}

// GetSndAddress returns the sender address from the reward transaction
func (rtx *RewardTx) GetSndAddress() []byte {
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

// SetValue sets the value of the reward transaction
func (rtx *RewardTx) SetValue(value *big.Int) {
	rtx.Value = value
}

// SetData sets the data of the reward transaction
func (rtx *RewardTx) SetData(data []byte) {
}

// SetRecvAddress sets the receiver address of the reward transaction
func (rtx *RewardTx) SetRecvAddress(addr []byte) {
	rtx.RcvAddr = addr
}

// SetSndAddress sets the sender address of the reward transaction
func (rtx *RewardTx) SetSndAddress(addr []byte) {
}
