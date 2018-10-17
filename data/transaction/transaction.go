package transaction

import (
	"io"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction/capnproto1"
	"github.com/glycerine/go-capnproto"
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

	var n int

	// Nonce
	n = src.Nonce().Len()
	dest.Nonce = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.Nonce[i] = byte(src.Nonce().At(i))
	}

	// Value
	n = src.Value().Len()
	dest.Value = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.Value[i] = byte(src.Value().At(i))
	}

	// RcvAddr
	n = src.RcvAddr().Len()
	dest.RcvAddr = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.RcvAddr[i] = byte(src.RcvAddr().At(i))
	}

	// SndAddr
	n = src.SndAddr().Len()
	dest.SndAddr = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.SndAddr[i] = byte(src.SndAddr().At(i))
	}

	// GasPrice
	n = src.GasPrice().Len()
	dest.GasPrice = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.GasPrice[i] = byte(src.GasPrice().At(i))
	}

	// GasLimit
	n = src.GasLimit().Len()
	dest.GasLimit = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.GasLimit[i] = byte(src.GasLimit().At(i))
	}

	// Data
	n = src.Data().Len()
	dest.Data = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.Data[i] = byte(src.Data().At(i))
	}

	// Signature
	n = src.Signature().Len()
	dest.Signature = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.Signature[i] = byte(src.Signature().At(i))
	}

	// Challenge
	n = src.Challenge().Len()
	dest.Challenge = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.Challenge[i] = byte(src.Challenge().At(i))
	}

	// PubKey
	n = src.PubKey().Len()
	dest.PubKey = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.PubKey[i] = byte(src.PubKey().At(i))
	}

	return dest
}

func TransactionGoToCapn(seg *capn.Segment, src *Transaction) capnproto1.TransactionCapn {
	dest := capnproto1.AutoNewTransactionCapn(seg)

	mylist1 := seg.NewUInt8List(len(src.Nonce))
	for i := range src.Nonce {
		mylist1.Set(i, uint8(src.Nonce[i]))
	}
	dest.SetNonce(mylist1)

	mylist2 := seg.NewUInt8List(len(src.Value))
	for i := range src.Value {
		mylist2.Set(i, uint8(src.Value[i]))
	}
	dest.SetValue(mylist2)

	mylist3 := seg.NewUInt8List(len(src.RcvAddr))
	for i := range src.RcvAddr {
		mylist3.Set(i, uint8(src.RcvAddr[i]))
	}
	dest.SetRcvAddr(mylist3)

	mylist4 := seg.NewUInt8List(len(src.SndAddr))
	for i := range src.SndAddr {
		mylist4.Set(i, uint8(src.SndAddr[i]))
	}
	dest.SetSndAddr(mylist4)

	mylist5 := seg.NewUInt8List(len(src.GasPrice))
	for i := range src.GasPrice {
		mylist5.Set(i, uint8(src.GasPrice[i]))
	}
	dest.SetGasPrice(mylist5)

	mylist6 := seg.NewUInt8List(len(src.GasLimit))
	for i := range src.GasLimit {
		mylist6.Set(i, uint8(src.GasLimit[i]))
	}
	dest.SetGasLimit(mylist6)

	mylist7 := seg.NewUInt8List(len(src.Data))
	for i := range src.Data {
		mylist7.Set(i, uint8(src.Data[i]))
	}
	dest.SetData(mylist7)

	mylist8 := seg.NewUInt8List(len(src.Signature))
	for i := range src.Signature {
		mylist8.Set(i, uint8(src.Signature[i]))
	}
	dest.SetSignature(mylist8)

	mylist9 := seg.NewUInt8List(len(src.Challenge))
	for i := range src.Challenge {
		mylist9.Set(i, uint8(src.Challenge[i]))
	}
	dest.SetChallenge(mylist9)

	mylist10 := seg.NewUInt8List(len(src.PubKey))
	for i := range src.PubKey {
		mylist10.Set(i, uint8(src.PubKey[i]))
	}
	dest.SetPubKey(mylist10)

	return dest
}

func SliceByteToUInt8List(seg *capn.Segment, m []byte) capn.UInt8List {
	lst := seg.NewUInt8List(len(m))
	for i := range m {
		lst.Set(i, uint8(m[i]))
	}
	return lst
}

func UInt8ListToSliceByte(p capn.UInt8List) []byte {
	v := make([]byte, p.Len())
	for i := range v {
		v[i] = byte(p.At(i))
	}
	return v
}
