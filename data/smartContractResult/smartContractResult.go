//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. smartContractResult.proto
package smartContractResult

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

var _ = data.TransactionHandler(&SmartContractResult{})

// IsInterfaceNil verifies if underlying object is nil
func (scr *SmartContractResult) IsInterfaceNil() bool {
	return scr == nil
}

// SetValue sets the value of the smart contract result
func (scr *SmartContractResult) SetValue(value *big.Int) {
	scr.Value = value
}

// SetData sets the data of the smart contract result
func (scr *SmartContractResult) SetData(data []byte) {
	scr.Data = data
}

// SetRcvAddr sets the receiver address of the smart contract result
func (scr *SmartContractResult) SetRcvAddr(addr []byte) {
	scr.RcvAddr = addr
}

// SetSndAddr sets the sender address of the smart contract result
func (scr *SmartContractResult) SetSndAddr(addr []byte) {
	scr.SndAddr = addr
}

// GetRcvUserName returns the receiver user name from the smart contract result
func (_ *SmartContractResult) GetRcvUserName() []byte {
	return nil
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
