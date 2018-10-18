package transaction

import (
	"io"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction/capnproto1"
	"github.com/glycerine/go-capnproto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

// Transaction holds all the data needed for a value transfer
type Transaction struct {
	Nonce     []byte
	Value     []byte
	RcvAddr   []byte
	SndAddr   []byte
	GasPrice  []byte
	GasLimit  []byte
	Data      []byte
	Signature []byte
	Challenge []byte
	PubKey    []byte
}

func (s *Transaction) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	TransactionGoToCapn(seg, s)
	_, err := seg.WriteTo(w)
	return err
}

func (s *Transaction) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		//panic(fmt.Errorf("capn.ReadFromStream error: %s", err))
		return err
	}
	z := capnproto1.ReadRootTransactionCapn(capMsg)
	TransactionCapnToGo(z, s)
	return nil
}

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

	// PubKey
	dest.PubKey = src.PubKey()

	return dest
}

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

	dest.SetPubKey(src.PubKey)

	return dest
}

// GenerateDummyArray is used for tests to generate an array of transactions with dummy data
func (tx *Transaction) GenerateDummyArray() []data.CapnpHelper {
	transactions := make([]data.CapnpHelper, 0, 1000)

	for i := 0; i < 1000; i++ {
		transactions = append(transactions, &Transaction{
			Nonce:     []byte(data.RandomStr(6)),
			Value:     []byte(data.RandomStr(6)),
			RcvAddr:   []byte(data.RandomStr(32)),
			SndAddr:   []byte(data.RandomStr(32)),
			GasPrice:  []byte(data.RandomStr(6)),
			GasLimit:  []byte(data.RandomStr(6)),
			Data:      []byte(data.RandomStr(32)),
			Signature: []byte(data.RandomStr(32)),
			Challenge: []byte(data.RandomStr(32)),
			PubKey:    []byte(data.RandomStr(32)),
		})
	}

	return transactions
}
