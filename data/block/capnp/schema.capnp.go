package capnp

// AUTO GENERATED - DO NOT EDIT

import (
	"bufio"
	"bytes"
	"encoding/json"
	C "github.com/glycerine/go-capnproto"
	"io"
)

type HeaderCapn C.Struct

func NewHeaderCapn(s *C.Segment) HeaderCapn      { return HeaderCapn(s.NewStruct(32, 5)) }
func NewRootHeaderCapn(s *C.Segment) HeaderCapn  { return HeaderCapn(s.NewRootStruct(32, 5)) }
func AutoNewHeaderCapn(s *C.Segment) HeaderCapn  { return HeaderCapn(s.NewStructAR(32, 5)) }
func ReadRootHeaderCapn(s *C.Segment) HeaderCapn { return HeaderCapn(s.Root(0).ToStruct()) }
func (s HeaderCapn) Nonce() uint64               { return C.Struct(s).Get64(0) }
func (s HeaderCapn) SetNonce(v uint64)           { C.Struct(s).Set64(0, v) }
func (s HeaderCapn) PrevHash() []byte            { return C.Struct(s).GetObject(0).ToData() }
func (s HeaderCapn) SetPrevHash(v []byte)        { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s HeaderCapn) PubKeysBitmap() []byte       { return C.Struct(s).GetObject(1).ToData() }
func (s HeaderCapn) SetPubKeysBitmap(v []byte)   { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s HeaderCapn) ShardId() uint32             { return C.Struct(s).Get32(8) }
func (s HeaderCapn) SetShardId(v uint32)         { C.Struct(s).Set32(8, v) }
func (s HeaderCapn) TimeStamp() uint64           { return C.Struct(s).Get64(16) }
func (s HeaderCapn) SetTimeStamp(v uint64)       { C.Struct(s).Set64(16, v) }
func (s HeaderCapn) Round() uint32               { return C.Struct(s).Get32(12) }
func (s HeaderCapn) SetRound(v uint32)           { C.Struct(s).Set32(12, v) }
func (s HeaderCapn) Epoch() uint32               { return C.Struct(s).Get32(24) }
func (s HeaderCapn) SetEpoch(v uint32)           { C.Struct(s).Set32(24, v) }
func (s HeaderCapn) BlockBodyHash() []byte       { return C.Struct(s).GetObject(2).ToData() }
func (s HeaderCapn) SetBlockBodyHash(v []byte)   { C.Struct(s).SetObject(2, s.Segment.NewData(v)) }
func (s HeaderCapn) BlockBodyType() uint8        { return C.Struct(s).Get8(28) }
func (s HeaderCapn) SetBlockBodyType(v uint8)    { C.Struct(s).Set8(28, v) }
func (s HeaderCapn) Signature() []byte           { return C.Struct(s).GetObject(3).ToData() }
func (s HeaderCapn) SetSignature(v []byte)       { C.Struct(s).SetObject(3, s.Segment.NewData(v)) }
func (s HeaderCapn) Commitment() []byte          { return C.Struct(s).GetObject(4).ToData() }
func (s HeaderCapn) SetCommitment(v []byte)      { C.Struct(s).SetObject(4, s.Segment.NewData(v)) }
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
	_, err = b.WriteString("\"pubKeysBitmap\":")
	if err != nil {
		return err
	}
	{
		s := s.PubKeysBitmap()
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
	_, err = b.WriteString("\"blockBodyHash\":")
	if err != nil {
		return err
	}
	{
		s := s.BlockBodyHash()
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
	_, err = b.WriteString("\"blockBodyType\":")
	if err != nil {
		return err
	}
	{
		s := s.BlockBodyType()
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
	_, err = b.WriteString("pubKeysBitmap = ")
	if err != nil {
		return err
	}
	{
		s := s.PubKeysBitmap()
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
	_, err = b.WriteString("blockBodyHash = ")
	if err != nil {
		return err
	}
	{
		s := s.BlockBodyHash()
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
	_, err = b.WriteString("blockBodyType = ")
	if err != nil {
		return err
	}
	{
		s := s.BlockBodyType()
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
	return HeaderCapn_List(s.NewCompositeList(32, 5, sz))
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
func (s MiniBlockCapn) ShardID() uint32                { return C.Struct(s).Get32(0) }
func (s MiniBlockCapn) SetShardID(v uint32)            { C.Struct(s).Set32(0, v) }
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
	_, err = b.WriteString("\"shardID\":")
	if err != nil {
		return err
	}
	{
		s := s.ShardID()
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
	_, err = b.WriteString("shardID = ")
	if err != nil {
		return err
	}
	{
		s := s.ShardID()
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

type PeerBlockBodyCapn C.Struct

func NewPeerBlockBodyCapn(s *C.Segment) PeerBlockBodyCapn { return PeerBlockBodyCapn(s.NewStruct(0, 2)) }
func NewRootPeerBlockBodyCapn(s *C.Segment) PeerBlockBodyCapn {
	return PeerBlockBodyCapn(s.NewRootStruct(0, 2))
}
func AutoNewPeerBlockBodyCapn(s *C.Segment) PeerBlockBodyCapn {
	return PeerBlockBodyCapn(s.NewStructAR(0, 2))
}
func ReadRootPeerBlockBodyCapn(s *C.Segment) PeerBlockBodyCapn {
	return PeerBlockBodyCapn(s.Root(0).ToStruct())
}
func (s PeerBlockBodyCapn) StateBlockBody() StateBlockBodyCapn {
	return StateBlockBodyCapn(C.Struct(s).GetObject(0).ToStruct())
}
func (s PeerBlockBodyCapn) SetStateBlockBody(v StateBlockBodyCapn) {
	C.Struct(s).SetObject(0, C.Object(v))
}
func (s PeerBlockBodyCapn) Changes() PeerChangeCapn_List {
	return PeerChangeCapn_List(C.Struct(s).GetObject(1))
}
func (s PeerBlockBodyCapn) SetChanges(v PeerChangeCapn_List) { C.Struct(s).SetObject(1, C.Object(v)) }
func (s PeerBlockBodyCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"stateBlockBody\":")
	if err != nil {
		return err
	}
	{
		s := s.StateBlockBody()
		err = s.WriteJSON(b)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"changes\":")
	if err != nil {
		return err
	}
	{
		s := s.Changes()
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
func (s PeerBlockBodyCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s PeerBlockBodyCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("stateBlockBody = ")
	if err != nil {
		return err
	}
	{
		s := s.StateBlockBody()
		err = s.WriteCapLit(b)
		if err != nil {
			return err
		}
	}
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("changes = ")
	if err != nil {
		return err
	}
	{
		s := s.Changes()
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
func (s PeerBlockBodyCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type PeerBlockBodyCapn_List C.PointerList

func NewPeerBlockBodyCapnList(s *C.Segment, sz int) PeerBlockBodyCapn_List {
	return PeerBlockBodyCapn_List(s.NewCompositeList(0, 2, sz))
}
func (s PeerBlockBodyCapn_List) Len() int { return C.PointerList(s).Len() }
func (s PeerBlockBodyCapn_List) At(i int) PeerBlockBodyCapn {
	return PeerBlockBodyCapn(C.PointerList(s).At(i).ToStruct())
}
func (s PeerBlockBodyCapn_List) ToArray() []PeerBlockBodyCapn {
	n := s.Len()
	a := make([]PeerBlockBodyCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s PeerBlockBodyCapn_List) Set(i int, item PeerBlockBodyCapn) {
	C.PointerList(s).Set(i, C.Object(item))
}

type PeerChangeCapn C.Struct

func NewPeerChangeCapn(s *C.Segment) PeerChangeCapn      { return PeerChangeCapn(s.NewStruct(8, 1)) }
func NewRootPeerChangeCapn(s *C.Segment) PeerChangeCapn  { return PeerChangeCapn(s.NewRootStruct(8, 1)) }
func AutoNewPeerChangeCapn(s *C.Segment) PeerChangeCapn  { return PeerChangeCapn(s.NewStructAR(8, 1)) }
func ReadRootPeerChangeCapn(s *C.Segment) PeerChangeCapn { return PeerChangeCapn(s.Root(0).ToStruct()) }
func (s PeerChangeCapn) PubKey() []byte                  { return C.Struct(s).GetObject(0).ToData() }
func (s PeerChangeCapn) SetPubKey(v []byte)              { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s PeerChangeCapn) ShardIdDest() uint32             { return C.Struct(s).Get32(0) }
func (s PeerChangeCapn) SetShardIdDest(v uint32)         { C.Struct(s).Set32(0, v) }
func (s PeerChangeCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"pubKey\":")
	if err != nil {
		return err
	}
	{
		s := s.PubKey()
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
	_, err = b.WriteString("\"shardIdDest\":")
	if err != nil {
		return err
	}
	{
		s := s.ShardIdDest()
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
func (s PeerChangeCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s PeerChangeCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("pubKey = ")
	if err != nil {
		return err
	}
	{
		s := s.PubKey()
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
	_, err = b.WriteString("shardIdDest = ")
	if err != nil {
		return err
	}
	{
		s := s.ShardIdDest()
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
func (s PeerChangeCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type PeerChangeCapn_List C.PointerList

func NewPeerChangeCapnList(s *C.Segment, sz int) PeerChangeCapn_List {
	return PeerChangeCapn_List(s.NewCompositeList(8, 1, sz))
}
func (s PeerChangeCapn_List) Len() int { return C.PointerList(s).Len() }
func (s PeerChangeCapn_List) At(i int) PeerChangeCapn {
	return PeerChangeCapn(C.PointerList(s).At(i).ToStruct())
}
func (s PeerChangeCapn_List) ToArray() []PeerChangeCapn {
	n := s.Len()
	a := make([]PeerChangeCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s PeerChangeCapn_List) Set(i int, item PeerChangeCapn) { C.PointerList(s).Set(i, C.Object(item)) }

type StateBlockBodyCapn C.Struct

func NewStateBlockBodyCapn(s *C.Segment) StateBlockBodyCapn {
	return StateBlockBodyCapn(s.NewStruct(8, 1))
}
func NewRootStateBlockBodyCapn(s *C.Segment) StateBlockBodyCapn {
	return StateBlockBodyCapn(s.NewRootStruct(8, 1))
}
func AutoNewStateBlockBodyCapn(s *C.Segment) StateBlockBodyCapn {
	return StateBlockBodyCapn(s.NewStructAR(8, 1))
}
func ReadRootStateBlockBodyCapn(s *C.Segment) StateBlockBodyCapn {
	return StateBlockBodyCapn(s.Root(0).ToStruct())
}
func (s StateBlockBodyCapn) RootHash() []byte     { return C.Struct(s).GetObject(0).ToData() }
func (s StateBlockBodyCapn) SetRootHash(v []byte) { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s StateBlockBodyCapn) ShardID() uint32      { return C.Struct(s).Get32(0) }
func (s StateBlockBodyCapn) SetShardID(v uint32)  { C.Struct(s).Set32(0, v) }
func (s StateBlockBodyCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"rootHash\":")
	if err != nil {
		return err
	}
	{
		s := s.RootHash()
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
	_, err = b.WriteString("\"shardID\":")
	if err != nil {
		return err
	}
	{
		s := s.ShardID()
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
func (s StateBlockBodyCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s StateBlockBodyCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("rootHash = ")
	if err != nil {
		return err
	}
	{
		s := s.RootHash()
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
	_, err = b.WriteString("shardID = ")
	if err != nil {
		return err
	}
	{
		s := s.ShardID()
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
func (s StateBlockBodyCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type StateBlockBodyCapn_List C.PointerList

func NewStateBlockBodyCapnList(s *C.Segment, sz int) StateBlockBodyCapn_List {
	return StateBlockBodyCapn_List(s.NewCompositeList(8, 1, sz))
}
func (s StateBlockBodyCapn_List) Len() int { return C.PointerList(s).Len() }
func (s StateBlockBodyCapn_List) At(i int) StateBlockBodyCapn {
	return StateBlockBodyCapn(C.PointerList(s).At(i).ToStruct())
}
func (s StateBlockBodyCapn_List) ToArray() []StateBlockBodyCapn {
	n := s.Len()
	a := make([]StateBlockBodyCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s StateBlockBodyCapn_List) Set(i int, item StateBlockBodyCapn) {
	C.PointerList(s).Set(i, C.Object(item))
}

type TxBlockBodyCapn C.Struct

func NewTxBlockBodyCapn(s *C.Segment) TxBlockBodyCapn { return TxBlockBodyCapn(s.NewStruct(0, 2)) }
func NewRootTxBlockBodyCapn(s *C.Segment) TxBlockBodyCapn {
	return TxBlockBodyCapn(s.NewRootStruct(0, 2))
}
func AutoNewTxBlockBodyCapn(s *C.Segment) TxBlockBodyCapn { return TxBlockBodyCapn(s.NewStructAR(0, 2)) }
func ReadRootTxBlockBodyCapn(s *C.Segment) TxBlockBodyCapn {
	return TxBlockBodyCapn(s.Root(0).ToStruct())
}
func (s TxBlockBodyCapn) StateBlockBody() StateBlockBodyCapn {
	return StateBlockBodyCapn(C.Struct(s).GetObject(0).ToStruct())
}
func (s TxBlockBodyCapn) SetStateBlockBody(v StateBlockBodyCapn) {
	C.Struct(s).SetObject(0, C.Object(v))
}
func (s TxBlockBodyCapn) MiniBlocks() MiniBlockCapn_List {
	return MiniBlockCapn_List(C.Struct(s).GetObject(1))
}
func (s TxBlockBodyCapn) SetMiniBlocks(v MiniBlockCapn_List) { C.Struct(s).SetObject(1, C.Object(v)) }
func (s TxBlockBodyCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"stateBlockBody\":")
	if err != nil {
		return err
	}
	{
		s := s.StateBlockBody()
		err = s.WriteJSON(b)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(',')
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
func (s TxBlockBodyCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s TxBlockBodyCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("stateBlockBody = ")
	if err != nil {
		return err
	}
	{
		s := s.StateBlockBody()
		err = s.WriteCapLit(b)
		if err != nil {
			return err
		}
	}
	_, err = b.WriteString(", ")
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
func (s TxBlockBodyCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type TxBlockBodyCapn_List C.PointerList

func NewTxBlockBodyCapnList(s *C.Segment, sz int) TxBlockBodyCapn_List {
	return TxBlockBodyCapn_List(s.NewCompositeList(0, 2, sz))
}
func (s TxBlockBodyCapn_List) Len() int { return C.PointerList(s).Len() }
func (s TxBlockBodyCapn_List) At(i int) TxBlockBodyCapn {
	return TxBlockBodyCapn(C.PointerList(s).At(i).ToStruct())
}
func (s TxBlockBodyCapn_List) ToArray() []TxBlockBodyCapn {
	n := s.Len()
	a := make([]TxBlockBodyCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s TxBlockBodyCapn_List) Set(i int, item TxBlockBodyCapn) {
	C.PointerList(s).Set(i, C.Object(item))
}
