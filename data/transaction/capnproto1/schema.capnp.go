package capnproto1

// AUTO GENERATED - DO NOT EDIT

import (
	"bufio"
	"bytes"
	"encoding/json"
	C "github.com/glycerine/go-capnproto"
	"io"
)

type TransactionCapn C.Struct

func NewTransactionCapn(s *C.Segment) TransactionCapn { return TransactionCapn(s.NewStruct(24, 6)) }
func NewRootTransactionCapn(s *C.Segment) TransactionCapn {
	return TransactionCapn(s.NewRootStruct(24, 6))
}
func AutoNewTransactionCapn(s *C.Segment) TransactionCapn {
	return TransactionCapn(s.NewStructAR(24, 6))
}
func ReadRootTransactionCapn(s *C.Segment) TransactionCapn {
	return TransactionCapn(s.Root(0).ToStruct())
}
func (s TransactionCapn) Nonce() uint64         { return C.Struct(s).Get64(0) }
func (s TransactionCapn) SetNonce(v uint64)     { C.Struct(s).Set64(0, v) }
func (s TransactionCapn) Value() []byte         { return C.Struct(s).GetObject(0).ToData() }
func (s TransactionCapn) SetValue(v []byte)     { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s TransactionCapn) RcvAddr() []byte       { return C.Struct(s).GetObject(1).ToData() }
func (s TransactionCapn) SetRcvAddr(v []byte)   { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s TransactionCapn) SndAddr() []byte       { return C.Struct(s).GetObject(2).ToData() }
func (s TransactionCapn) SetSndAddr(v []byte)   { C.Struct(s).SetObject(2, s.Segment.NewData(v)) }
func (s TransactionCapn) GasPrice() uint64      { return C.Struct(s).Get64(8) }
func (s TransactionCapn) SetGasPrice(v uint64)  { C.Struct(s).Set64(8, v) }
func (s TransactionCapn) GasLimit() uint64      { return C.Struct(s).Get64(16) }
func (s TransactionCapn) SetGasLimit(v uint64)  { C.Struct(s).Set64(16, v) }
func (s TransactionCapn) Data() []byte          { return C.Struct(s).GetObject(3).ToData() }
func (s TransactionCapn) SetData(v []byte)      { C.Struct(s).SetObject(3, s.Segment.NewData(v)) }
func (s TransactionCapn) Signature() []byte     { return C.Struct(s).GetObject(4).ToData() }
func (s TransactionCapn) SetSignature(v []byte) { C.Struct(s).SetObject(4, s.Segment.NewData(v)) }
func (s TransactionCapn) Challenge() []byte     { return C.Struct(s).GetObject(5).ToData() }
func (s TransactionCapn) SetChallenge(v []byte) { C.Struct(s).SetObject(5, s.Segment.NewData(v)) }
func (s TransactionCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"nonce\":")
	if err != nil {
		return err
	}
	{
		s := s.Nonce()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"value\":")
	if err != nil {
		return err
	}
	{
		s := s.Value()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"rcvAddr\":")
	if err != nil {
		return err
	}
	{
		s := s.RcvAddr()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"sndAddr\":")
	if err != nil {
		return err
	}
	{
		s := s.SndAddr()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"gasPrice\":")
	if err != nil {
		return err
	}
	{
		s := s.GasPrice()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"gasLimit\":")
	if err != nil {
		return err
	}
	{
		s := s.GasLimit()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"data\":")
	if err != nil {
		return err
	}
	{
		s := s.Data()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"signature\":")
	if err != nil {
		return err
	}
	{
		s := s.Signature()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"challenge\":")
	if err != nil {
		return err
	}
	{
		s := s.Challenge()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte('}')
	if err != nil {
		return err
	}
	err = b.Flush()
	return err
}
func (s TransactionCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s TransactionCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("nonce = ")
	if err != nil {
		return err
	}
	{
		s := s.Nonce()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("value = ")
	if err != nil {
		return err
	}
	{
		s := s.Value()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("rcvAddr = ")
	if err != nil {
		return err
	}
	{
		s := s.RcvAddr()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("sndAddr = ")
	if err != nil {
		return err
	}
	{
		s := s.SndAddr()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("gasPrice = ")
	if err != nil {
		return err
	}
	{
		s := s.GasPrice()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("gasLimit = ")
	if err != nil {
		return err
	}
	{
		s := s.GasLimit()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("data = ")
	if err != nil {
		return err
	}
	{
		s := s.Data()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("signature = ")
	if err != nil {
		return err
	}
	{
		s := s.Signature()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("challenge = ")
	if err != nil {
		return err
	}
	{
		s := s.Challenge()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(')')
	if err != nil {
		return err
	}
	err = b.Flush()
	return err
}
func (s TransactionCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type TransactionCapn_List C.PointerList

func NewTransactionCapnList(s *C.Segment, sz int) TransactionCapn_List {
	return TransactionCapn_List(s.NewCompositeList(24, 6, sz))
}
func (s TransactionCapn_List) Len() int { return C.PointerList(s).Len() }
func (s TransactionCapn_List) At(i int) TransactionCapn {
	return TransactionCapn(C.PointerList(s).At(i).ToStruct())
}
func (s TransactionCapn_List) ToArray() []TransactionCapn {
	n := s.Len()
	a := make([]TransactionCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s TransactionCapn_List) Set(i int, item TransactionCapn) {
	C.PointerList(s).Set(i, C.Object(item))
}
