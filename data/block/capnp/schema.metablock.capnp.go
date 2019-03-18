package capnp

// AUTO GENERATED - DO NOT EDIT

import (
	"bufio"
	"bytes"
	"encoding/json"
	C "github.com/glycerine/go-capnproto"
	"io"
)

type PeerDataCapn C.Struct

func NewPeerDataCapn(s *C.Segment) PeerDataCapn      { return PeerDataCapn(s.NewStruct(16, 2)) }
func NewRootPeerDataCapn(s *C.Segment) PeerDataCapn  { return PeerDataCapn(s.NewRootStruct(16, 2)) }
func AutoNewPeerDataCapn(s *C.Segment) PeerDataCapn  { return PeerDataCapn(s.NewStructAR(16, 2)) }
func ReadRootPeerDataCapn(s *C.Segment) PeerDataCapn { return PeerDataCapn(s.Root(0).ToStruct()) }
func (s PeerDataCapn) PublicKey() []byte             { return C.Struct(s).GetObject(0).ToData() }
func (s PeerDataCapn) SetPublicKey(v []byte)         { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s PeerDataCapn) Action() uint8                 { return C.Struct(s).Get8(0) }
func (s PeerDataCapn) SetAction(v uint8)             { C.Struct(s).Set8(0, v) }
func (s PeerDataCapn) Timestamp() uint64             { return C.Struct(s).Get64(8) }
func (s PeerDataCapn) SetTimestamp(v uint64)         { C.Struct(s).Set64(8, v) }
func (s PeerDataCapn) Value() []byte                 { return C.Struct(s).GetObject(1).ToData() }
func (s PeerDataCapn) SetValue(v []byte)             { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s PeerDataCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"publicKey\":")
	if err != nil {
		return err
	}
	{
		s := s.PublicKey()
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
	_, err = b.WriteString("\"action\":")
	if err != nil {
		return err
	}
	{
		s := s.Action()
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
	err = b.WriteByte('}')
	if err != nil {
		return err
	}
	err = b.Flush()
	return err
}
func (s PeerDataCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s PeerDataCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("publicKey = ")
	if err != nil {
		return err
	}
	{
		s := s.PublicKey()
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
	_, err = b.WriteString("action = ")
	if err != nil {
		return err
	}
	{
		s := s.Action()
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
	err = b.WriteByte(')')
	if err != nil {
		return err
	}
	err = b.Flush()
	return err
}
func (s PeerDataCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type PeerDataCapn_List C.PointerList

func NewPeerDataCapnList(s *C.Segment, sz int) PeerDataCapn_List {
	return PeerDataCapn_List(s.NewCompositeList(16, 2, sz))
}
func (s PeerDataCapn_List) Len() int { return C.PointerList(s).Len() }
func (s PeerDataCapn_List) At(i int) PeerDataCapn {
	return PeerDataCapn(C.PointerList(s).At(i).ToStruct())
}
func (s PeerDataCapn_List) ToArray() []PeerDataCapn {
	n := s.Len()
	a := make([]PeerDataCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s PeerDataCapn_List) Set(i int, item PeerDataCapn) { C.PointerList(s).Set(i, C.Object(item)) }

type ShardMiniBlockHeaderCapn C.Struct

func NewShardMiniBlockHeaderCapn(s *C.Segment) ShardMiniBlockHeaderCapn {
	return ShardMiniBlockHeaderCapn(s.NewStruct(8, 1))
}
func NewRootShardMiniBlockHeaderCapn(s *C.Segment) ShardMiniBlockHeaderCapn {
	return ShardMiniBlockHeaderCapn(s.NewRootStruct(8, 1))
}
func AutoNewShardMiniBlockHeaderCapn(s *C.Segment) ShardMiniBlockHeaderCapn {
	return ShardMiniBlockHeaderCapn(s.NewStructAR(8, 1))
}
func ReadRootShardMiniBlockHeaderCapn(s *C.Segment) ShardMiniBlockHeaderCapn {
	return ShardMiniBlockHeaderCapn(s.Root(0).ToStruct())
}
func (s ShardMiniBlockHeaderCapn) Hash() []byte                { return C.Struct(s).GetObject(0).ToData() }
func (s ShardMiniBlockHeaderCapn) SetHash(v []byte)            { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s ShardMiniBlockHeaderCapn) ReceiverShardId() uint32     { return C.Struct(s).Get32(0) }
func (s ShardMiniBlockHeaderCapn) SetReceiverShardId(v uint32) { C.Struct(s).Set32(0, v) }
func (s ShardMiniBlockHeaderCapn) SenderShardId() uint32       { return C.Struct(s).Get32(4) }
func (s ShardMiniBlockHeaderCapn) SetSenderShardId(v uint32)   { C.Struct(s).Set32(4, v) }
func (s ShardMiniBlockHeaderCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"hash\":")
	if err != nil {
		return err
	}
	{
		s := s.Hash()
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
	_, err = b.WriteString("\"receiverShardId\":")
	if err != nil {
		return err
	}
	{
		s := s.ReceiverShardId()
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
	_, err = b.WriteString("\"senderShardId\":")
	if err != nil {
		return err
	}
	{
		s := s.SenderShardId()
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
func (s ShardMiniBlockHeaderCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s ShardMiniBlockHeaderCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("hash = ")
	if err != nil {
		return err
	}
	{
		s := s.Hash()
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
	_, err = b.WriteString("receiverShardId = ")
	if err != nil {
		return err
	}
	{
		s := s.ReceiverShardId()
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
	_, err = b.WriteString("senderShardId = ")
	if err != nil {
		return err
	}
	{
		s := s.SenderShardId()
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
func (s ShardMiniBlockHeaderCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type ShardMiniBlockHeaderCapn_List C.PointerList

func NewShardMiniBlockHeaderCapnList(s *C.Segment, sz int) ShardMiniBlockHeaderCapn_List {
	return ShardMiniBlockHeaderCapn_List(s.NewCompositeList(8, 1, sz))
}
func (s ShardMiniBlockHeaderCapn_List) Len() int { return C.PointerList(s).Len() }
func (s ShardMiniBlockHeaderCapn_List) At(i int) ShardMiniBlockHeaderCapn {
	return ShardMiniBlockHeaderCapn(C.PointerList(s).At(i).ToStruct())
}
func (s ShardMiniBlockHeaderCapn_List) ToArray() []ShardMiniBlockHeaderCapn {
	n := s.Len()
	a := make([]ShardMiniBlockHeaderCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s ShardMiniBlockHeaderCapn_List) Set(i int, item ShardMiniBlockHeaderCapn) {
	C.PointerList(s).Set(i, C.Object(item))
}

type ShardDataCapn C.Struct

func NewShardDataCapn(s *C.Segment) ShardDataCapn      { return ShardDataCapn(s.NewStruct(8, 2)) }
func NewRootShardDataCapn(s *C.Segment) ShardDataCapn  { return ShardDataCapn(s.NewRootStruct(8, 2)) }
func AutoNewShardDataCapn(s *C.Segment) ShardDataCapn  { return ShardDataCapn(s.NewStructAR(8, 2)) }
func ReadRootShardDataCapn(s *C.Segment) ShardDataCapn { return ShardDataCapn(s.Root(0).ToStruct()) }
func (s ShardDataCapn) ShardId() uint32                { return C.Struct(s).Get32(0) }
func (s ShardDataCapn) SetShardId(v uint32)            { C.Struct(s).Set32(0, v) }
func (s ShardDataCapn) HeaderHash() []byte             { return C.Struct(s).GetObject(0).ToData() }
func (s ShardDataCapn) SetHeaderHash(v []byte)         { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s ShardDataCapn) ShardMiniBlockHeaders() ShardMiniBlockHeaderCapn_List {
	return ShardMiniBlockHeaderCapn_List(C.Struct(s).GetObject(1))
}
func (s ShardDataCapn) SetShardMiniBlockHeaders(v ShardMiniBlockHeaderCapn_List) {
	C.Struct(s).SetObject(1, C.Object(v))
}
func (s ShardDataCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
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
	_, err = b.WriteString("\"headerHash\":")
	if err != nil {
		return err
	}
	{
		s := s.HeaderHash()
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
	_, err = b.WriteString("\"shardMiniBlockHeaders\":")
	if err != nil {
		return err
	}
	{
		s := s.ShardMiniBlockHeaders()
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
func (s ShardDataCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s ShardDataCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
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
	_, err = b.WriteString("headerHash = ")
	if err != nil {
		return err
	}
	{
		s := s.HeaderHash()
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
	_, err = b.WriteString("shardMiniBlockHeaders = ")
	if err != nil {
		return err
	}
	{
		s := s.ShardMiniBlockHeaders()
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
func (s ShardDataCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type ShardDataCapn_List C.PointerList

func NewShardDataCapnList(s *C.Segment, sz int) ShardDataCapn_List {
	return ShardDataCapn_List(s.NewCompositeList(8, 2, sz))
}
func (s ShardDataCapn_List) Len() int { return C.PointerList(s).Len() }
func (s ShardDataCapn_List) At(i int) ShardDataCapn {
	return ShardDataCapn(C.PointerList(s).At(i).ToStruct())
}
func (s ShardDataCapn_List) ToArray() []ShardDataCapn {
	n := s.Len()
	a := make([]ShardDataCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s ShardDataCapn_List) Set(i int, item ShardDataCapn) { C.PointerList(s).Set(i, C.Object(item)) }

type MetaBlockCapn C.Struct

func NewMetaBlockCapn(s *C.Segment) MetaBlockCapn      { return MetaBlockCapn(s.NewStruct(24, 8)) }
func NewRootMetaBlockCapn(s *C.Segment) MetaBlockCapn  { return MetaBlockCapn(s.NewRootStruct(24, 8)) }
func AutoNewMetaBlockCapn(s *C.Segment) MetaBlockCapn  { return MetaBlockCapn(s.NewStructAR(24, 8)) }
func ReadRootMetaBlockCapn(s *C.Segment) MetaBlockCapn { return MetaBlockCapn(s.Root(0).ToStruct()) }
func (s MetaBlockCapn) Nonce() uint64                  { return C.Struct(s).Get64(0) }
func (s MetaBlockCapn) SetNonce(v uint64)              { C.Struct(s).Set64(0, v) }
func (s MetaBlockCapn) Epoch() uint32                  { return C.Struct(s).Get32(8) }
func (s MetaBlockCapn) SetEpoch(v uint32)              { C.Struct(s).Set32(8, v) }
func (s MetaBlockCapn) Round() uint32                  { return C.Struct(s).Get32(12) }
func (s MetaBlockCapn) SetRound(v uint32)              { C.Struct(s).Set32(12, v) }
func (s MetaBlockCapn) TimeStamp() uint64              { return C.Struct(s).Get64(16) }
func (s MetaBlockCapn) SetTimeStamp(v uint64)          { C.Struct(s).Set64(16, v) }
func (s MetaBlockCapn) ShardInfo() ShardDataCapn_List {
	return ShardDataCapn_List(C.Struct(s).GetObject(0))
}
func (s MetaBlockCapn) SetShardInfo(v ShardDataCapn_List) { C.Struct(s).SetObject(0, C.Object(v)) }
func (s MetaBlockCapn) PeerInfo() PeerDataCapn_List {
	return PeerDataCapn_List(C.Struct(s).GetObject(1))
}
func (s MetaBlockCapn) SetPeerInfo(v PeerDataCapn_List) { C.Struct(s).SetObject(1, C.Object(v)) }
func (s MetaBlockCapn) Signature() []byte               { return C.Struct(s).GetObject(2).ToData() }
func (s MetaBlockCapn) SetSignature(v []byte)           { C.Struct(s).SetObject(2, s.Segment.NewData(v)) }
func (s MetaBlockCapn) PubKeysBitmap() []byte           { return C.Struct(s).GetObject(3).ToData() }
func (s MetaBlockCapn) SetPubKeysBitmap(v []byte)       { C.Struct(s).SetObject(3, s.Segment.NewData(v)) }
func (s MetaBlockCapn) PreviousHash() []byte            { return C.Struct(s).GetObject(4).ToData() }
func (s MetaBlockCapn) SetPreviousHash(v []byte)        { C.Struct(s).SetObject(4, s.Segment.NewData(v)) }
func (s MetaBlockCapn) PrevRandSeed() []byte            { return C.Struct(s).GetObject(5).ToData() }
func (s MetaBlockCapn) SetPrevRandSeed(v []byte)        { C.Struct(s).SetObject(5, s.Segment.NewData(v)) }
func (s MetaBlockCapn) RandSeed() []byte                { return C.Struct(s).GetObject(6).ToData() }
func (s MetaBlockCapn) SetRandSeed(v []byte)            { C.Struct(s).SetObject(6, s.Segment.NewData(v)) }
func (s MetaBlockCapn) StateRootHash() []byte           { return C.Struct(s).GetObject(7).ToData() }
func (s MetaBlockCapn) SetStateRootHash(v []byte)       { C.Struct(s).SetObject(7, s.Segment.NewData(v)) }
func (s MetaBlockCapn) WriteJSON(w io.Writer) error {
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
	_, err = b.WriteString("\"shardInfo\":")
	if err != nil {
		return err
	}
	{
		s := s.ShardInfo()
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
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"peerInfo\":")
	if err != nil {
		return err
	}
	{
		s := s.PeerInfo()
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
	_, err = b.WriteString("\"previousHash\":")
	if err != nil {
		return err
	}
	{
		s := s.PreviousHash()
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
	_, err = b.WriteString("\"prevRandSeed\":")
	if err != nil {
		return err
	}
	{
		s := s.PrevRandSeed()
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
	_, err = b.WriteString("\"randSeed\":")
	if err != nil {
		return err
	}
	{
		s := s.RandSeed()
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
	_, err = b.WriteString("\"stateRootHash\":")
	if err != nil {
		return err
	}
	{
		s := s.StateRootHash()
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
func (s MetaBlockCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s MetaBlockCapn) WriteCapLit(w io.Writer) error {
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
	_, err = b.WriteString("shardInfo = ")
	if err != nil {
		return err
	}
	{
		s := s.ShardInfo()
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
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("peerInfo = ")
	if err != nil {
		return err
	}
	{
		s := s.PeerInfo()
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
	_, err = b.WriteString("previousHash = ")
	if err != nil {
		return err
	}
	{
		s := s.PreviousHash()
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
	_, err = b.WriteString("prevRandSeed = ")
	if err != nil {
		return err
	}
	{
		s := s.PrevRandSeed()
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
	_, err = b.WriteString("randSeed = ")
	if err != nil {
		return err
	}
	{
		s := s.RandSeed()
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
	_, err = b.WriteString("stateRootHash = ")
	if err != nil {
		return err
	}
	{
		s := s.StateRootHash()
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
func (s MetaBlockCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type MetaBlockCapn_List C.PointerList

func NewMetaBlockCapnList(s *C.Segment, sz int) MetaBlockCapn_List {
	return MetaBlockCapn_List(s.NewCompositeList(24, 8, sz))
}
func (s MetaBlockCapn_List) Len() int { return C.PointerList(s).Len() }
func (s MetaBlockCapn_List) At(i int) MetaBlockCapn {
	return MetaBlockCapn(C.PointerList(s).At(i).ToStruct())
}
func (s MetaBlockCapn_List) ToArray() []MetaBlockCapn {
	n := s.Len()
	a := make([]MetaBlockCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s MetaBlockCapn_List) Set(i int, item MetaBlockCapn) { C.PointerList(s).Set(i, C.Object(item)) }
