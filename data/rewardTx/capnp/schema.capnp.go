package capnp

// AUTO GENERATED - DO NOT EDIT

import (
	"bufio"
	"bytes"
	"encoding/json"
	C "github.com/glycerine/go-capnproto"
	"io"
)

type RewardTxCapn C.Struct

func NewRewardTxCapn(s *C.Segment) RewardTxCapn      { return RewardTxCapn(s.NewStruct(16, 2)) }
func NewRootRewardTxCapn(s *C.Segment) RewardTxCapn  { return RewardTxCapn(s.NewRootStruct(16, 2)) }
func AutoNewRewardTxCapn(s *C.Segment) RewardTxCapn  { return RewardTxCapn(s.NewStructAR(16, 2)) }
func ReadRootRewardTxCapn(s *C.Segment) RewardTxCapn { return RewardTxCapn(s.Root(0).ToStruct()) }
func (s RewardTxCapn) Round() uint64                 { return C.Struct(s).Get64(0) }
func (s RewardTxCapn) SetRound(v uint64)             { C.Struct(s).Set64(0, v) }
func (s RewardTxCapn) Epoch() uint32                 { return C.Struct(s).Get32(8) }
func (s RewardTxCapn) SetEpoch(v uint32)             { C.Struct(s).Set32(8, v) }
func (s RewardTxCapn) Value() []byte                 { return C.Struct(s).GetObject(0).ToData() }
func (s RewardTxCapn) SetValue(v []byte)             { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s RewardTxCapn) RcvAddr() []byte               { return C.Struct(s).GetObject(1).ToData() }
func (s RewardTxCapn) SetRcvAddr(v []byte)           { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s RewardTxCapn) ShardId() uint32               { return C.Struct(s).Get32(12) }
func (s RewardTxCapn) SetShardId(v uint32)           { C.Struct(s).Set32(12, v) }
func (s RewardTxCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"round\":")
	if err != nil {
		return err
	}
	{
		s := s.Round()
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
	_, err = b.WriteString("\"epoch\":")
	if err != nil {
		return err
	}
	{
		s := s.Epoch()
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
func (s RewardTxCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s RewardTxCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("round = ")
	if err != nil {
		return err
	}
	{
		s := s.Round()
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
	_, err = b.WriteString("epoch = ")
	if err != nil {
		return err
	}
	{
		s := s.Epoch()
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
func (s RewardTxCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type RewardTxCapn_List C.PointerList

func NewRewardTxCapnList(s *C.Segment, sz int) RewardTxCapn_List {
	return RewardTxCapn_List(s.NewCompositeList(16, 2, sz))
}
func (s RewardTxCapn_List) Len() int { return C.PointerList(s).Len() }
func (s RewardTxCapn_List) At(i int) RewardTxCapn {
	return RewardTxCapn(C.PointerList(s).At(i).ToStruct())
}
func (s RewardTxCapn_List) ToArray() []RewardTxCapn {
	n := s.Len()
	a := make([]RewardTxCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s RewardTxCapn_List) Set(i int, item RewardTxCapn) { C.PointerList(s).Set(i, C.Object(item)) }
