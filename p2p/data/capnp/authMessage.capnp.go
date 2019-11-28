package capnp

// AUTO GENERATED - DO NOT EDIT

import (
	"bufio"
	"bytes"
	"encoding/json"
	C "github.com/glycerine/go-capnproto"
	"io"
)

type AuthMessageCapn C.Struct

func NewAuthMessageCapn(s *C.Segment) AuthMessageCapn { return AuthMessageCapn(s.NewStruct(8, 3)) }
func NewRootAuthMessageCapn(s *C.Segment) AuthMessageCapn {
	return AuthMessageCapn(s.NewRootStruct(8, 3))
}
func AutoNewAuthMessageCapn(s *C.Segment) AuthMessageCapn { return AuthMessageCapn(s.NewStructAR(8, 3)) }
func ReadRootAuthMessageCapn(s *C.Segment) AuthMessageCapn {
	return AuthMessageCapn(s.Root(0).ToStruct())
}
func (s AuthMessageCapn) Message() []byte      { return C.Struct(s).GetObject(0).ToData() }
func (s AuthMessageCapn) SetMessage(v []byte)  { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s AuthMessageCapn) Sig() []byte          { return C.Struct(s).GetObject(1).ToData() }
func (s AuthMessageCapn) SetSig(v []byte)      { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s AuthMessageCapn) Pubkey() []byte       { return C.Struct(s).GetObject(2).ToData() }
func (s AuthMessageCapn) SetPubkey(v []byte)   { C.Struct(s).SetObject(2, s.Segment.NewData(v)) }
func (s AuthMessageCapn) Timestamp() int64     { return int64(C.Struct(s).Get64(0)) }
func (s AuthMessageCapn) SetTimestamp(v int64) { C.Struct(s).Set64(0, uint64(v)) }
func (s AuthMessageCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"message\":")
	if err != nil {
		return err
	}
	{
		s := s.Message()
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
	_, err = b.WriteString("\"sig\":")
	if err != nil {
		return err
	}
	{
		s := s.Sig()
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
	_, err = b.WriteString("\"pubkey\":")
	if err != nil {
		return err
	}
	{
		s := s.Pubkey()
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
	_, err = b.WriteString("\"timestamp\":")
	if err != nil {
		return err
	}
	{
		s := s.Timestamp()
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
func (s AuthMessageCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s AuthMessageCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("message = ")
	if err != nil {
		return err
	}
	{
		s := s.Message()
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
	_, err = b.WriteString("sig = ")
	if err != nil {
		return err
	}
	{
		s := s.Sig()
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
	_, err = b.WriteString("pubkey = ")
	if err != nil {
		return err
	}
	{
		s := s.Pubkey()
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
	_, err = b.WriteString("timestamp = ")
	if err != nil {
		return err
	}
	{
		s := s.Timestamp()
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
func (s AuthMessageCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type AuthMessageCapn_List C.PointerList

func NewAuthMessageCapnList(s *C.Segment, sz int) AuthMessageCapn_List {
	return AuthMessageCapn_List(s.NewCompositeList(8, 3, sz))
}
func (s AuthMessageCapn_List) Len() int { return C.PointerList(s).Len() }
func (s AuthMessageCapn_List) At(i int) AuthMessageCapn {
	return AuthMessageCapn(C.PointerList(s).At(i).ToStruct())
}
func (s AuthMessageCapn_List) ToArray() []AuthMessageCapn {
	n := s.Len()
	a := make([]AuthMessageCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s AuthMessageCapn_List) Set(i int, item AuthMessageCapn) {
	C.PointerList(s).Set(i, C.Object(item))
}
