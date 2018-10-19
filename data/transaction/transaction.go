package transaction

import (
	"io"

	"math/rand"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction/capnproto1"
	"github.com/glycerine/go-capnproto"
)

// Transaction holds all the data needed for a value transfer
type Transaction struct {
	Nonce     uint64
	Value     uint64
	RcvAddr   []byte
	SndAddr   []byte
	GasPrice  uint64
	GasLimit  uint64
	Data      []byte
	Signature []byte
	Challenge []byte
}

// Save saves the serialized data of a Transaction into a stream through Capnp protocol
func (tx *Transaction) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	TransactionGoToCapn(seg, tx)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a Transaction object through Capnp protocol
func (tx *Transaction) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		//panic(fmt.Errorf("capn.ReadFromStream error: %s", err))
		return err
	}
	z := capnproto1.ReadRootTransactionCapn(capMsg)
	TransactionCapnToGo(z, tx)
	return nil
}

// TransactionCapnToGo is a helper function to copy fields from a TransactionCapn object to a Transaction object
func TransactionCapnToGo(src capnproto1.TransactionCapn, dest *Transaction) *Transaction {
	if dest == nil {
		dest = &Transaction{}
	}

	// Nonce
	dest.Nonce = src.Nonce()
	// Value
	dest.Value = src.Value()
	// RcvAddr
	dest.RcvAddr = src.RcvAddr()
	// SndAddr
	dest.SndAddr = src.SndAddr()
	// GasPrice
	dest.GasPrice = src.GasPrice()
	// GasLimit
	dest.GasLimit = src.GasLimit()
	// Data
	dest.Data = src.Data()
	// Signature
	dest.Signature = src.Signature()
	// Challenge
	dest.Challenge = src.Challenge()

	return dest
}

// TransactionGoToCapn is a helper function to copy fields from a Transaction object to a TransactionCapn object
func TransactionGoToCapn(seg *capn.Segment, src *Transaction) capnproto1.TransactionCapn {
	dest := capnproto1.AutoNewTransactionCapn(seg)

	dest.SetNonce(src.Nonce)
	dest.SetValue(src.Value)
	dest.SetRcvAddr(src.RcvAddr)
	dest.SetSndAddr(src.SndAddr)
	dest.SetGasPrice(src.GasPrice)
	dest.SetGasLimit(src.GasLimit)
	dest.SetData(src.Data)
	dest.SetSignature(src.Signature)
	dest.SetChallenge(src.Challenge)

	return dest
}

// GenerateDummyArray is used for tests to generate an array of transactions with dummy data
func (tx *Transaction) GenerateDummyArray() []data.CapnpHelper {
	transactions := make([]data.CapnpHelper, 0, 1000)

	for i := 0; i < 1000; i++ {
		transactions = append(transactions, &Transaction{
			Nonce:     uint64(rand.Int63n(10000)),
			Value:     uint64(rand.Int63n(10000)),
			RcvAddr:   []byte(data.RandomStr(32)),
			SndAddr:   []byte(data.RandomStr(32)),
			GasPrice:  uint64(rand.Int63n(10000)),
			GasLimit:  uint64(rand.Int63n(10000)),
			Data:      []byte(data.RandomStr(32)),
			Signature: []byte(data.RandomStr(32)),
			Challenge: []byte(data.RandomStr(32)),
		})
	}

	return transactions
}
