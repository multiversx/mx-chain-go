//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. transaction.proto
package transaction

import (
	io "io"
	"io/ioutil"

	"github.com/ElrondNetwork/elrond-go/data"
)

var _ = data.TransactionHandler(&Transaction{})

// IsInterfaceNil verifies if underlying object is nil
func (tx *Transaction) IsInterfaceNil() bool {
	return tx == nil
}

// SetValue sets the value of the transaction
func (tx *Transaction) GetValue() *data.ProtoBigInt {
	if tx == nil {
		return nil
	}
	return tx.Value
}

// SetValue sets the value of the transaction
func (tx *Transaction) SetValue(value *data.ProtoBigInt) {
	tx.Value = value
}

// SetData sets the data of the transaction
func (tx *Transaction) SetData(data string) {
	tx.Data = data
}

// SetRcvAddr sets the receiver address of the transaction
func (tx *Transaction) SetRcvAddr(addr []byte) {
	tx.RcvAddr = addr
}

// SetSndAddr sets the sender address of the transaction
func (tx *Transaction) SetSndAddr(addr []byte) {
	tx.SndAddr = addr
}

// ----- for compatibility only ----

func (t *Transaction) Save(w io.Writer) error {
	b, err := t.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (t *Transaction) Load(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return t.Unmarshal(b)
}
