package capnp

// AUTO GENERATED - DO NOT EDIT
import (
	"bufio"
	"bytes"
	"encoding/json"
	C "github.com/glycerine/go-capnproto"
	"io"
)

type SmartContractResultCapn C.Struct

func NewSmartContractResultCapn(s *C.Segment) SmartContractResultCapn {
	return SmartContractResultCapn(s.NewStruct(8, 6))
}
func NewRootSmartContractResultCapn(s *C.Segment) SmartContractResultCapn {
	return SmartContractResultCapn(s.NewRootStruct(8, 6))
}
func AutoNewSmartContractResultCapn(s *C.Segment) SmartContractResultCapn {
	return SmartContractResultCapn(s.NewStructAR(8, 6))
}
func ReadRootSmartContractResultCapn(s *C.Segment) SmartContractResultCapn {
	return SmartContractResultCapn(s.Root(0).ToStruct())
}
func (s SmartContractResultCapn) Nonce() uint64       { return C.Struct(s).Get64(0) }
func (s SmartContractResultCapn) SetNonce(v uint64)   { C.Struct(s).Set64(0, v) }
func (s SmartContractResultCapn) Value() []byte       { return C.Struct(s).GetObject(0).ToData() }
func (s SmartContractResultCapn) SetValue(v []byte)   { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s SmartContractResultCapn) RcvAddr() []byte     { return C.Struct(s).GetObject(1).ToData() }
func (s SmartContractResultCapn) SetRcvAddr(v []byte) { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s SmartContractResultCapn) SndAddr() []byte     { return C.Struct(s).GetObject(2).ToData() }
func (s SmartContractResultCapn) SetSndAddr(v []byte) { C.Struct(s).SetObject(2, s.Segment.NewData(v)) }
func (s SmartContractResultCapn) Code() []byte        { return C.Struct(s).GetObject(3).ToData() }
func (s SmartContractResultCapn) SetCode(v []byte)    { C.Struct(s).SetObject(3, s.Segment.NewData(v)) }
func (s SmartContractResultCapn) Data() []byte        { return C.Struct(s).GetObject(4).ToData() }
func (s SmartContractResultCapn) SetData(v []byte)    { C.Struct(s).SetObject(4, s.Segment.NewData(v)) }
func (s SmartContractResultCapn) TxHash() []byte      { return C.Struct(s).GetObject(5).ToData() }
func (s SmartContractResultCapn) SetTxHash(v []byte)  { C.Struct(s).SetObject(5, s.Segment.NewData(v)) }
func (s SmartContractResultCapn) WriteJSON(w io.Writer) error {
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
	_, err = b.WriteString("\"code\":")
	if err != nil {
		return err
	}
	{
		s := s.Code()
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
func (s SmartContractResultCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s SmartContractResultCapn) WriteCapLit(w io.Writer) error {
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
	_, err = b.WriteString("code = ")
	if err != nil {
		return err
	}
	{
		s := s.Code()
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
func (s SmartContractResultCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type SmartContractResultCapn_List C.PointerList

func NewSmartContractResultCapnList(s *C.Segment, sz int) SmartContractResultCapn_List {
	return SmartContractResultCapn_List(s.NewCompositeList(8, 6, sz))
}
func (s SmartContractResultCapn_List) Len() int { return C.PointerList(s).Len() }
func (s SmartContractResultCapn_List) At(i int) SmartContractResultCapn {
	return SmartContractResultCapn(C.PointerList(s).At(i).ToStruct())
}
func (s SmartContractResultCapn_List) ToArray() []SmartContractResultCapn {
	n := s.Len()
	a := make([]SmartContractResultCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s SmartContractResultCapn_List) Set(i int, item SmartContractResultCapn) {
	C.PointerList(s).Set(i, C.Object(item))
}
