package smartContractResult

import (
	"math/big"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// SmartContractResult holds all the data needed for results coming from smart contract processing
type SmartContractResult struct {
	Nonce    uint64            `json:"nonce"`
	Value    *big.Int          `json:"value"`
	RcvAddr  []byte            `json:"receiver"`
	SndAddr  []byte            `json:"sender"`
	Code     []byte            `json:"code,omitempty"`
	Data     []byte            `json:"data,omitempty"`
	TxHash   []byte            `json:"txHash"`
	GasLimit uint64            `json:"gasLimit"`
	GasPrice uint64            `json:"gasPrice"`
	CallType vmcommon.CallType `json:"callType"`
}

// IsInterfaceNil verifies if underlying object is nil
func (scr *SmartContractResult) IsInterfaceNil() bool {
	return scr == nil
}

// GetValue returns the value of the smart contract result
func (scr *SmartContractResult) GetValue() *big.Int {
	return scr.Value
}

// GetNonce returns the nonce of the smart contract result
func (scr *SmartContractResult) GetNonce() uint64 {
	return scr.Nonce
}

// GetData returns the data of the smart contract result
func (scr *SmartContractResult) GetData() []byte {
	return scr.Data
}

// GetRecvAddress returns the receiver address from the smart contract result
func (scr *SmartContractResult) GetRecvAddress() []byte {
	return scr.RcvAddr
}

// GetSndAddress returns the sender address from the smart contract result
func (scr *SmartContractResult) GetSndAddress() []byte {
	return scr.SndAddr
}

// GetGasLimit returns the gas limit of the smart contract result
func (scr *SmartContractResult) GetGasLimit() uint64 {
	return scr.GasLimit
}

// GetGasPrice returns the gas price of the smart contract result
func (scr *SmartContractResult) GetGasPrice() uint64 {
	return scr.GasPrice
}

// SetValue sets the value of the smart contract result
func (scr *SmartContractResult) SetValue(value *big.Int) {
	scr.Value = value
}

// SetData sets the data of the smart contract result
func (scr *SmartContractResult) SetData(data []byte) {
	scr.Data = data
}

// SetRecvAddress sets the receiver address of the smart contract result
func (scr *SmartContractResult) SetRecvAddress(addr []byte) {
	scr.RcvAddr = addr
}

// SetSndAddress sets the sender address of the smart contract result
func (scr *SmartContractResult) SetSndAddress(addr []byte) {
	scr.SndAddr = addr
}

// TrimSlicePtr creates a copy of the provided slice without the excess capacity
func TrimSlicePtr(in []*SmartContractResult) []*SmartContractResult {
	if len(in) == 0 {
		return []*SmartContractResult{}
	}
	ret := make([]*SmartContractResult, len(in))
	copy(ret, in)
	return ret
}
