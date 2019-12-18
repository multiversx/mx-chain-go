//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. smartContractResult.proto
package smartContractResult

import (
	io "io"
	"io/ioutil"

	"github.com/ElrondNetwork/elrond-go/data"
)

var _ = data.TransactionHandler(&SmartContractResult{})

// IsInterfaceNil verifies if underlying object is nil
func (scr *SmartContractResult) IsInterfaceNil() bool {
	return scr == nil
}

// GetValue returns the value of the smart contract result
func (scr *SmartContractResult) GetValue() *data.ProtoBigInt {
	return scr.Value
}
// SetValue sets the value of the smart contract result
func (scr *SmartContractResult) SetValue(value *data.ProtoBigInt) {
	scr.Value = value
}

// SetData sets the data of the smart contract result
func (scr *SmartContractResult) SetData(data string) {
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
