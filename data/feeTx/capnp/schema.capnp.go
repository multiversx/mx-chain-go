package capnp

// AUTO GENERATED - DO NOT EDIT

import (
	"bufio"
	"bytes"
	"encoding/json"
	C "github.com/glycerine/go-capnproto"
	"io"
)

type FeeTxCapn C.Struct

func NewFeeTxCapn(s *C.Segment) FeeTxCapn      { return FeeTxCapn(s.NewStruct(16, 2)) }
func NewRootFeeTxCapn(s *C.Segment) FeeTxCapn  { return FeeTxCapn(s.NewRootStruct(16, 2)) }
func AutoNewFeeTxCapn(s *C.Segment) FeeTxCapn  { return FeeTxCapn(s.NewStructAR(16, 2)) }
func ReadRootFeeTxCapn(s *C.Segment) FeeTxCapn { return FeeTxCapn(s.Root(0).ToStruct()) }
func (s FeeTxCapn) Nonce() uint64              { return C.Struct(s).Get64(0) }
func (s FeeTxCapn) SetNonce(v uint64)          { C.Struct(s).Set64(0, v) }
func (s FeeTxCapn) Value() []byte              { return C.Struct(s).GetObject(0).ToData() }
func (s FeeTxCapn) SetValue(v []byte)          { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s FeeTxCapn) RcvAddr() []byte            { return C.Struct(s).GetObject(1).ToData() }
func (s FeeTxCapn) SetRcvAddr(v []byte)        { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s FeeTxCapn) ShardId() uint32            { return C.Struct(s).Get32(8) }
func (s FeeTxCapn) SetShardId(v uint32)        { C.Struct(s).Set32(8, v) }
func (s FeeTxCapn) WriteJSON(w io.Writer) error {
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
	_, err = b.WriteString("\"shardId\":")
	if err != nil {
		return err
	}
	{
		s := s.ShardId()
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
func (s FeeTxCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s FeeTxCapn) WriteCapLit(w io.Writer) error {
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
	_, err = b.WriteString("shardId = ")
	if err != nil {
		return err
	}
	{
		s := s.ShardId()
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
func (s FeeTxCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type FeeTxCapn_List C.PointerList

func NewFeeTxCapnList(s *C.Segment, sz int) FeeTxCapn_List {
	return FeeTxCapn_List(s.NewCompositeList(16, 2, sz))
}
func (s FeeTxCapn_List) Len() int           { return C.PointerList(s).Len() }
func (s FeeTxCapn_List) At(i int) FeeTxCapn { return FeeTxCapn(C.PointerList(s).At(i).ToStruct()) }
func (s FeeTxCapn_List) ToArray() []FeeTxCapn {
	n := s.Len()
	a := make([]FeeTxCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s FeeTxCapn_List) Set(i int, item FeeTxCapn) { C.PointerList(s).Set(i, C.Object(item)) }
