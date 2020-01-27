package receipt

import (
	"math/big"
)

// Receipt holds all the data needed for a transaction receipt
type Receipt struct {
	Value   *big.Int
	SndAddr []byte
	Data    []byte
	TxHash  []byte
}

// IsInterfaceNil verifies if underlying object is nil
func (rpt *Receipt) IsInterfaceNil() bool {
	return rpt == nil
}

// GetValue returns the value of the receipt
func (rpt *Receipt) GetValue() *big.Int {
	return rpt.Value
}

// GetNonce returns the nonce of the receipt
func (rpt *Receipt) GetNonce() uint64 {
	return 0
}

// GetData returns the data of the receipt
func (rpt *Receipt) GetData() []byte {
	return rpt.Data
}

// GetRecvAddress returns the receiver address from the receipt
func (rpt *Receipt) GetRecvAddress() []byte {
	return rpt.SndAddr
}

// GetSndAddress returns the sender address from the receipt
func (rpt *Receipt) GetSndAddress() []byte {
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

// SetRecvAddress sets the receiver address of the receipt
func (rpt *Receipt) SetRecvAddress(_ []byte) {
}

// SetSndAddress sets the sender address of the receipt
func (rpt *Receipt) SetSndAddress(addr []byte) {
	rpt.SndAddr = addr
}
