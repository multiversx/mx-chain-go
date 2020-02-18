//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. smartContractResult.proto
package smartContractResult

import (
	io "io"
	"io/ioutil"
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

// TrimSlicePtr creates a copy of the provided slice without the excess capacity
func TrimSlicePtr(in []*SmartContractResult) []*SmartContractResult {
	if len(in) == 0 {
		return []*SmartContractResult{}
	}
	ret := make([]*SmartContractResult, len(in))
	copy(ret, in)
	return ret
}

// ----- for compatibility only ----

func (scr *SmartContractResult) Save(w io.Writer) error {
	b, err := scr.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (scr *SmartContractResult) Load(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return scr.Unmarshal(b)
}
