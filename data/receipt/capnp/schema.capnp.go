package capnp

// AUTO GENERATED - DO NOT EDIT

import (
	"bufio"
	"bytes"
	"encoding/json"
	C "github.com/glycerine/go-capnproto"
	"io"
)

type ReceiptCapn C.Struct

func NewReceiptCapn(s *C.Segment) ReceiptCapn      { return ReceiptCapn(s.NewStruct(0, 4)) }
func NewRootReceiptCapn(s *C.Segment) ReceiptCapn  { return ReceiptCapn(s.NewRootStruct(0, 4)) }
func AutoNewReceiptCapn(s *C.Segment) ReceiptCapn  { return ReceiptCapn(s.NewStructAR(0, 4)) }
func ReadRootReceiptCapn(s *C.Segment) ReceiptCapn { return ReceiptCapn(s.Root(0).ToStruct()) }
func (s ReceiptCapn) Value() []byte                { return C.Struct(s).GetObject(0).ToData() }
func (s ReceiptCapn) SetValue(v []byte)            { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s ReceiptCapn) SndAddr() []byte              { return C.Struct(s).GetObject(1).ToData() }
func (s ReceiptCapn) SetSndAddr(v []byte)          { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s ReceiptCapn) Data() []byte                 { return C.Struct(s).GetObject(2).ToData() }
func (s ReceiptCapn) SetData(v []byte)             { C.Struct(s).SetObject(2, s.Segment.NewData(v)) }
func (s ReceiptCapn) TxHash() []byte               { return C.Struct(s).GetObject(3).ToData() }
func (s ReceiptCapn) SetTxHash(v []byte)           { C.Struct(s).SetObject(3, s.Segment.NewData(v)) }
func (s ReceiptCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
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
	_, err = b.WriteString("\"txHash\":")
	if err != nil {
		return err
	}
	{
		s := s.TxHash()
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
func (s ReceiptCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s ReceiptCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
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
	_, err = b.WriteString("txHash = ")
	if err != nil {
		return err
	}
	{
		s := s.TxHash()
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
func (s ReceiptCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type ReceiptCapn_List C.PointerList

func NewReceiptCapnList(s *C.Segment, sz int) ReceiptCapn_List {
	return ReceiptCapn_List(s.NewCompositeList(0, 4, sz))
}
func (s ReceiptCapn_List) Len() int             { return C.PointerList(s).Len() }
func (s ReceiptCapn_List) At(i int) ReceiptCapn { return ReceiptCapn(C.PointerList(s).At(i).ToStruct()) }
func (s ReceiptCapn_List) ToArray() []ReceiptCapn {
	n := s.Len()
	a := make([]ReceiptCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s ReceiptCapn_List) Set(i int, item ReceiptCapn) { C.PointerList(s).Set(i, C.Object(item)) }
