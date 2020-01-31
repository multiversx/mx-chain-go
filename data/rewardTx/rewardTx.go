//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. rewardTx.proto
package rewardTx

import (
	io "io"
	"io/ioutil"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

var _ = data.TransactionHandler(&RewardTx{})

// IsInterfaceNil verifies if underlying object is nil
func (tx *RewardTx) IsInterfaceNil() bool {
	return tx == nil
}

// SetValue sets the value of the transaction
func (tx *RewardTx) SetValue(value *big.Int) {
	tx.Value = value
}

// GetNonce returns 0 as reward transactions do not have a nonce
func (_ *RewardTx) GetNonce() uint64 {
	return 0
}

// GetData returns the data of the reward transaction
func (_ *RewardTx) GetData() string {
	return ""
}

// GetSndAddr returns the sender address from the reward transaction
func (_ *RewardTx) GetSndAddr() []byte {
	return nil
}

// GetGasLimit returns the gas limit of the smart reward transaction
func (_ *RewardTx) GetGasLimit() uint64 {
	return 0
}

// GetGasPrice returns the gas price of the smart reward transaction
func (_ *RewardTx) GetGasPrice() uint64 {
	return 0
}

// SetData sets the data of the reward transaction
func (_ *RewardTx) SetData(data string) {
}

// SetRcvAddr sets the receiver address of the reward transaction
func (rtx *RewardTx) SetRcvAddr(addr []byte) {
	rtx.RcvAddr = addr
}

// SetSndAddr sets the sender address of the reward transaction
func (_ *RewardTx) SetSndAddr(addr []byte) {
}

// ----- for compatibility only ----

func (t *RewardTx) Save(w io.Writer) error {
	b, err := t.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (t *RewardTx) Load(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return t.Unmarshal(b)
}
