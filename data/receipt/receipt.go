//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. receipt.proto
package receipt

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

var _ = data.TransactionHandler(&Receipt{})

// IsInterfaceNil verifies if underlying object is nil
func (rpt *Receipt) IsInterfaceNil() bool {
	return rpt == nil
}

// GetNonce returns the nonce of the receipt
func (rpt *Receipt) GetNonce() uint64 {
	return 0
}

// GetRcvAddr returns the receiver address from the receipt
func (rpt *Receipt) GetRcvAddr() []byte {
	return rpt.SndAddr
}

// GetGasLimit returns the gas limit of the receipt
func (rpt *Receipt) GetGasLimit() uint64 {
	return 0
}

// GetGasPrice returns the gas price of the receipt
func (rpt *Receipt) GetGasPrice() uint64 {
	return 0
}

// SetValue sets the value of the receipt
func (rpt *Receipt) SetValue(value *big.Int) {
	rpt.Value = value
}

// SetData sets the data of the receipt
func (rpt *Receipt) SetData(data []byte) {
	rpt.Data = data
}

// SetRcvAddr sets the receiver address of the receipt
func (rpt *Receipt) SetRcvAddr(_ []byte) {
}

// SetSndAddr sets the sender address of the receipt
func (rpt *Receipt) SetSndAddr(addr []byte) {
	rpt.SndAddr = addr
}

// GetRcvUserName returns the receiver user name from the receipt
func (_ *Receipt) GetRcvUserName() []byte {
	return nil
}

func (rpt *Receipt) CheckIntegrity() error {
	return nil
}
