package capnproto1

// AUTO GENERATED - DO NOT EDIT

import (
	"bufio"
	"bytes"
	"encoding/json"
	C "github.com/glycerine/go-capnproto"
	"io"
)

type BlockCapn C.Struct

func NewBlockCapn(s *C.Segment) BlockCapn      { return BlockCapn(s.NewStruct(0, 1)) }
func NewRootBlockCapn(s *C.Segment) BlockCapn  { return BlockCapn(s.NewRootStruct(0, 1)) }
func AutoNewBlockCapn(s *C.Segment) BlockCapn  { return BlockCapn(s.NewStructAR(0, 1)) }
func ReadRootBlockCapn(s *C.Segment) BlockCapn { return BlockCapn(s.Root(0).ToStruct()) }
func (s BlockCapn) MiniBlocks() MiniBlockCapn_List {
	return MiniBlockCapn_List(C.Struct(s).GetObject(0))
}
func (s BlockCapn) SetMiniBlocks(v MiniBlockCapn_List) { C.Struct(s).SetObject(0, C.Object(v)) }
func (s BlockCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"miniBlocks\":")
	if err != nil {
		return err
	}
	{
		s := s.MiniBlocks()
		{
			err = b.WriteByte('[')
			if err != nil {
				return err
			}
			for i, s := range s.ToArray() {
				if i != 0 {
					_, err = b.WriteString(", ")
				}
				if err != nil {
					return err
				}
				err = s.WriteJSON(b)
				if err != nil {
					return err
				}
			}
			err = b.WriteByte(']')
		}
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
func (s BlockCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s BlockCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("miniBlocks = ")
	if err != nil {
		return err
	}
	{
		s := s.MiniBlocks()
		{
			err = b.WriteByte('[')
			if err != nil {
				return err
			}
			for i, s := range s.ToArray() {
				if i != 0 {
					_, err = b.WriteString(", ")
				}
				if err != nil {
					return err
				}
				err = s.WriteCapLit(b)
				if err != nil {
					return err
				}
			}
			err = b.WriteByte(']')
		}
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
func (s BlockCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type BlockCapn_List C.PointerList

func NewBlockCapnList(s *C.Segment, sz int) BlockCapn_List {
	return BlockCapn_List(s.NewCompositeList(0, 1, sz))
}
func (s BlockCapn_List) Len() int           { return C.PointerList(s).Len() }
func (s BlockCapn_List) At(i int) BlockCapn { return BlockCapn(C.PointerList(s).At(i).ToStruct()) }
func (s BlockCapn_List) ToArray() []BlockCapn {
	n := s.Len()
	a := make([]BlockCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s BlockCapn_List) Set(i int, item BlockCapn) { C.PointerList(s).Set(i, C.Object(item)) }

type HeaderCapn C.Struct

func NewHeaderCapn(s *C.Segment) HeaderCapn      { return HeaderCapn(s.NewStruct(8, 7)) }
func NewRootHeaderCapn(s *C.Segment) HeaderCapn  { return HeaderCapn(s.NewRootStruct(8, 7)) }
func AutoNewHeaderCapn(s *C.Segment) HeaderCapn  { return HeaderCapn(s.NewStructAR(8, 7)) }
func ReadRootHeaderCapn(s *C.Segment) HeaderCapn { return HeaderCapn(s.Root(0).ToStruct()) }
func (s HeaderCapn) Nonce() []byte               { return C.Struct(s).GetObject(0).ToData() }
func (s HeaderCapn) SetNonce(v []byte)           { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s HeaderCapn) PrevHash() []byte            { return C.Struct(s).GetObject(1).ToData() }
func (s HeaderCapn) SetPrevHash(v []byte)        { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s HeaderCapn) PubKeys() C.DataList         { return C.DataList(C.Struct(s).GetObject(2)) }
func (s HeaderCapn) SetPubKeys(v C.DataList)     { C.Struct(s).SetObject(2, C.Object(v)) }
func (s HeaderCapn) ShardId() uint32             { return C.Struct(s).Get32(0) }
func (s HeaderCapn) SetShardId(v uint32)         { C.Struct(s).Set32(0, v) }
func (s HeaderCapn) TimeStamp() []byte           { return C.Struct(s).GetObject(3).ToData() }
func (s HeaderCapn) SetTimeStamp(v []byte)       { C.Struct(s).SetObject(3, s.Segment.NewData(v)) }
func (s HeaderCapn) Round() uint32               { return C.Struct(s).Get32(4) }
func (s HeaderCapn) SetRound(v uint32)           { C.Struct(s).Set32(4, v) }
func (s HeaderCapn) BlockHash() []byte           { return C.Struct(s).GetObject(4).ToData() }
func (s HeaderCapn) SetBlockHash(v []byte)       { C.Struct(s).SetObject(4, s.Segment.NewData(v)) }
func (s HeaderCapn) Signature() []byte           { return C.Struct(s).GetObject(5).ToData() }
func (s HeaderCapn) SetSignature(v []byte)       { C.Struct(s).SetObject(5, s.Segment.NewData(v)) }
func (s HeaderCapn) Commitment() []byte          { return C.Struct(s).GetObject(6).ToData() }
func (s HeaderCapn) SetCommitment(v []byte)      { C.Struct(s).SetObject(6, s.Segment.NewData(v)) }
func (s HeaderCapn) WriteJSON(w io.Writer) error {
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
	_, err = b.WriteString("\"prevHash\":")
	if err != nil {
		return err
	}
	{
		s := s.PrevHash()
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
	_, err = b.WriteString("\"pubKeys\":")
	if err != nil {
		return err
	}
	{
		s := s.PubKeys()
		{
			err = b.WriteByte('[')
			if err != nil {
				return err
			}
			for i, s := range s.ToArray() {
				if i != 0 {
					_, err = b.WriteString(", ")
				}
				if err != nil {
					return err
				}
				buf, err = json.Marshal(s)
				if err != nil {
					return err
				}
				_, err = b.Write(buf)
				if err != nil {
					return err
				}
			}
			err = b.WriteByte(']')
		}
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
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"timeStamp\":")
	if err != nil {
		return err
	}
	{
		s := s.TimeStamp()
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
	_, err = b.WriteString("\"blockHash\":")
	if err != nil {
		return err
	}
	{
		s := s.BlockHash()
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
	_, err = b.WriteString("\"commitment\":")
	if err != nil {
		return err
	}
	{
		s := s.Commitment()
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
func (s HeaderCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s HeaderCapn) WriteCapLit(w io.Writer) error {
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
	_, err = b.WriteString("prevHash = ")
	if err != nil {
		return err
	}
	{
		s := s.PrevHash()
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
	_, err = b.WriteString("pubKeys = ")
	if err != nil {
		return err
	}
	{
		s := s.PubKeys()
		{
			err = b.WriteByte('[')
			if err != nil {
				return err
			}
			for i, s := range s.ToArray() {
				if i != 0 {
					_, err = b.WriteString(", ")
				}
				if err != nil {
					return err
				}
				buf, err = json.Marshal(s)
				if err != nil {
					return err
				}
				_, err = b.Write(buf)
				if err != nil {
					return err
				}
			}
			err = b.WriteByte(']')
		}
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
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("timeStamp = ")
	if err != nil {
		return err
	}
	{
		s := s.TimeStamp()
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
	_, err = b.WriteString("blockHash = ")
	if err != nil {
		return err
	}
	{
		s := s.BlockHash()
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
	_, err = b.WriteString("commitment = ")
	if err != nil {
		return err
	}
	{
		s := s.Commitment()
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
func (s HeaderCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type HeaderCapn_List C.PointerList

func NewHeaderCapnList(s *C.Segment, sz int) HeaderCapn_List {
	return HeaderCapn_List(s.NewCompositeList(8, 7, sz))
}
func (s HeaderCapn_List) Len() int            { return C.PointerList(s).Len() }
func (s HeaderCapn_List) At(i int) HeaderCapn { return HeaderCapn(C.PointerList(s).At(i).ToStruct()) }
func (s HeaderCapn_List) ToArray() []HeaderCapn {
	n := s.Len()
	a := make([]HeaderCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s HeaderCapn_List) Set(i int, item HeaderCapn) { C.PointerList(s).Set(i, C.Object(item)) }

type MiniBlockCapn C.Struct

func NewMiniBlockCapn(s *C.Segment) MiniBlockCapn      { return MiniBlockCapn(s.NewStruct(8, 1)) }
func NewRootMiniBlockCapn(s *C.Segment) MiniBlockCapn  { return MiniBlockCapn(s.NewRootStruct(8, 1)) }
func AutoNewMiniBlockCapn(s *C.Segment) MiniBlockCapn  { return MiniBlockCapn(s.NewStructAR(8, 1)) }
func ReadRootMiniBlockCapn(s *C.Segment) MiniBlockCapn { return MiniBlockCapn(s.Root(0).ToStruct()) }
func (s MiniBlockCapn) TxHashes() C.DataList           { return C.DataList(C.Struct(s).GetObject(0)) }
func (s MiniBlockCapn) SetTxHashes(v C.DataList)       { C.Struct(s).SetObject(0, C.Object(v)) }
func (s MiniBlockCapn) DestShardID() uint32            { return C.Struct(s).Get32(0) }
func (s MiniBlockCapn) SetDestShardID(v uint32)        { C.Struct(s).Set32(0, v) }
func (s MiniBlockCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"txHashes\":")
	if err != nil {
		return err
	}
	{
		s := s.TxHashes()
		{
			err = b.WriteByte('[')
			if err != nil {
				return err
			}
			for i, s := range s.ToArray() {
				if i != 0 {
					_, err = b.WriteString(", ")
				}
				if err != nil {
					return err
				}
				buf, err = json.Marshal(s)
				if err != nil {
					return err
				}
				_, err = b.Write(buf)
				if err != nil {
					return err
				}
			}
			err = b.WriteByte(']')
		}
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"destShardID\":")
	if err != nil {
		return err
	}
	{
		s := s.DestShardID()
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
func (s MiniBlockCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s MiniBlockCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("txHashes = ")
	if err != nil {
		return err
	}
	{
		s := s.TxHashes()
		{
			err = b.WriteByte('[')
			if err != nil {
				return err
			}
			for i, s := range s.ToArray() {
				if i != 0 {
					_, err = b.WriteString(", ")
				}
				if err != nil {
					return err
				}
				buf, err = json.Marshal(s)
				if err != nil {
					return err
				}
				_, err = b.Write(buf)
				if err != nil {
					return err
				}
			}
			err = b.WriteByte(']')
		}
		if err != nil {
			return err
		}
	}
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("destShardID = ")
	if err != nil {
		return err
	}
	{
		s := s.DestShardID()
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
func (s MiniBlockCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type MiniBlockCapn_List C.PointerList

func NewMiniBlockCapnList(s *C.Segment, sz int) MiniBlockCapn_List {
	return MiniBlockCapn_List(s.NewCompositeList(8, 1, sz))
}
func (s MiniBlockCapn_List) Len() int { return C.PointerList(s).Len() }
func (s MiniBlockCapn_List) At(i int) MiniBlockCapn {
	return MiniBlockCapn(C.PointerList(s).At(i).ToStruct())
}
func (s MiniBlockCapn_List) ToArray() []MiniBlockCapn {
	n := s.Len()
	a := make([]MiniBlockCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s MiniBlockCapn_List) Set(i int, item MiniBlockCapn) { C.PointerList(s).Set(i, C.Object(item)) }
