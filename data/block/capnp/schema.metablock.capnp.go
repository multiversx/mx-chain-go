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

func NewPeerDataCapn(s *C.Segment) PeerDataCapn      { return PeerDataCapn(s.NewStruct(16, 3)) }
func NewRootPeerDataCapn(s *C.Segment) PeerDataCapn  { return PeerDataCapn(s.NewRootStruct(16, 3)) }
func AutoNewPeerDataCapn(s *C.Segment) PeerDataCapn  { return PeerDataCapn(s.NewStructAR(16, 3)) }
func ReadRootPeerDataCapn(s *C.Segment) PeerDataCapn { return PeerDataCapn(s.Root(0).ToStruct()) }
func (s PeerDataCapn) Address() []byte               { return C.Struct(s).GetObject(0).ToData() }
func (s PeerDataCapn) SetAddress(v []byte)           { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s PeerDataCapn) PublicKey() []byte             { return C.Struct(s).GetObject(1).ToData() }
func (s PeerDataCapn) SetPublicKey(v []byte)         { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s PeerDataCapn) Action() uint8                 { return C.Struct(s).Get8(0) }
func (s PeerDataCapn) SetAction(v uint8)             { C.Struct(s).Set8(0, v) }
func (s PeerDataCapn) Timestamp() uint64             { return C.Struct(s).Get64(8) }
func (s PeerDataCapn) SetTimestamp(v uint64)         { C.Struct(s).Set64(8, v) }
func (s PeerDataCapn) Value() []byte                 { return C.Struct(s).GetObject(2).ToData() }
func (s PeerDataCapn) SetValue(v []byte)             { C.Struct(s).SetObject(2, s.Segment.NewData(v)) }
func (s PeerDataCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"address\":")
	if err != nil {
		return err
	}
	{
		s := s.Address()
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
	_, err = b.WriteString("address = ")
	if err != nil {
		return err
	}
	{
		s := s.Address()
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
	return PeerDataCapn_List(s.NewCompositeList(16, 3, sz))
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
	return ShardMiniBlockHeaderCapn(s.NewStruct(16, 1))
}
func NewRootShardMiniBlockHeaderCapn(s *C.Segment) ShardMiniBlockHeaderCapn {
	return ShardMiniBlockHeaderCapn(s.NewRootStruct(16, 1))
}
func AutoNewShardMiniBlockHeaderCapn(s *C.Segment) ShardMiniBlockHeaderCapn {
	return ShardMiniBlockHeaderCapn(s.NewStructAR(16, 1))
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
func (s ShardMiniBlockHeaderCapn) TxCount() uint32             { return C.Struct(s).Get32(8) }
func (s ShardMiniBlockHeaderCapn) SetTxCount(v uint32)         { C.Struct(s).Set32(8, v) }
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
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"txCount\":")
	if err != nil {
		return err
	}
	{
		s := s.TxCount()
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
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("txCount = ")
	if err != nil {
		return err
	}
	{
		s := s.TxCount()
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
	return ShardMiniBlockHeaderCapn_List(s.NewCompositeList(16, 1, sz))
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

func NewShardDataCapn(s *C.Segment) ShardDataCapn      { return ShardDataCapn(s.NewStruct(24, 6)) }
func NewRootShardDataCapn(s *C.Segment) ShardDataCapn  { return ShardDataCapn(s.NewRootStruct(24, 6)) }
func AutoNewShardDataCapn(s *C.Segment) ShardDataCapn  { return ShardDataCapn(s.NewStructAR(24, 6)) }
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
func (s ShardDataCapn) PrevRandSeed() []byte      { return C.Struct(s).GetObject(2).ToData() }
func (s ShardDataCapn) SetPrevRandSeed(v []byte)  { C.Struct(s).SetObject(2, s.Segment.NewData(v)) }
func (s ShardDataCapn) PubKeysBitmap() []byte     { return C.Struct(s).GetObject(3).ToData() }
func (s ShardDataCapn) SetPubKeysBitmap(v []byte) { C.Struct(s).SetObject(3, s.Segment.NewData(v)) }
func (s ShardDataCapn) Signature() []byte         { return C.Struct(s).GetObject(4).ToData() }
func (s ShardDataCapn) SetSignature(v []byte)     { C.Struct(s).SetObject(4, s.Segment.NewData(v)) }
func (s ShardDataCapn) TxCount() uint32           { return C.Struct(s).Get32(4) }
func (s ShardDataCapn) SetTxCount(v uint32)       { C.Struct(s).Set32(4, v) }
func (s ShardDataCapn) Round() uint64             { return C.Struct(s).Get64(8) }
func (s ShardDataCapn) SetRound(v uint64)         { C.Struct(s).Set64(8, v) }
func (s ShardDataCapn) PrevHash() []byte          { return C.Struct(s).GetObject(5).ToData() }
func (s ShardDataCapn) SetPrevHash(v []byte)      { C.Struct(s).SetObject(5, s.Segment.NewData(v)) }
func (s ShardDataCapn) Nonce() uint64             { return C.Struct(s).Get64(16) }
func (s ShardDataCapn) SetNonce(v uint64)         { C.Struct(s).Set64(16, v) }
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
	_, err = b.WriteString("\"txCount\":")
	if err != nil {
		return err
	}
	{
		s := s.TxCount()
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
	_, err = b.WriteString("txCount = ")
	if err != nil {
		return err
	}
	{
		s := s.TxCount()
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
	return ShardDataCapn_List(s.NewCompositeList(24, 6, sz))
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

type FinalizedHeadersCapn C.Struct

func NewFinalizedHeadersCapn(s *C.Segment) FinalizedHeadersCapn {
	return FinalizedHeadersCapn(s.NewStruct(8, 5))
}
func NewRootFinalizedHeadersCapn(s *C.Segment) FinalizedHeadersCapn {
	return FinalizedHeadersCapn(s.NewRootStruct(8, 5))
}
func AutoNewFinalizedHeadersCapn(s *C.Segment) FinalizedHeadersCapn {
	return FinalizedHeadersCapn(s.NewStructAR(8, 5))
}
func ReadRootFinalizedHeadersCapn(s *C.Segment) FinalizedHeadersCapn {
	return FinalizedHeadersCapn(s.Root(0).ToStruct())
}
func (s FinalizedHeadersCapn) ShardId() uint32               { return C.Struct(s).Get32(0) }
func (s FinalizedHeadersCapn) SetShardId(v uint32)           { C.Struct(s).Set32(0, v) }
func (s FinalizedHeadersCapn) HeaderHash() []byte            { return C.Struct(s).GetObject(0).ToData() }
func (s FinalizedHeadersCapn) SetHeaderHash(v []byte)        { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s FinalizedHeadersCapn) RootHash() []byte              { return C.Struct(s).GetObject(1).ToData() }
func (s FinalizedHeadersCapn) SetRootHash(v []byte)          { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s FinalizedHeadersCapn) FirstPendingMetaBlock() []byte { return C.Struct(s).GetObject(2).ToData() }
func (s FinalizedHeadersCapn) SetFirstPendingMetaBlock(v []byte) {
	C.Struct(s).SetObject(2, s.Segment.NewData(v))
}
func (s FinalizedHeadersCapn) LastFinishedMetaBlock() []byte { return C.Struct(s).GetObject(3).ToData() }
func (s FinalizedHeadersCapn) SetLastFinishedMetaBlock(v []byte) {
	C.Struct(s).SetObject(3, s.Segment.NewData(v))
}
func (s FinalizedHeadersCapn) PendingMiniBlockHeaders() ShardMiniBlockHeaderCapn_List {
	return ShardMiniBlockHeaderCapn_List(C.Struct(s).GetObject(4))
}
func (s FinalizedHeadersCapn) SetPendingMiniBlockHeaders(v ShardMiniBlockHeaderCapn_List) {
	C.Struct(s).SetObject(4, C.Object(v))
}
func (s FinalizedHeadersCapn) WriteJSON(w io.Writer) error {
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
	_, err = b.WriteString("\"firstPendingMetaBlock\":")
	if err != nil {
		return err
	}
	{
		s := s.FirstPendingMetaBlock()
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
	_, err = b.WriteString("\"lastFinishedMetaBlock\":")
	if err != nil {
		return err
	}
	{
		s := s.LastFinishedMetaBlock()
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
	_, err = b.WriteString("\"pendingMiniBlockHeaders\":")
	if err != nil {
		return err
	}
	{
		s := s.PendingMiniBlockHeaders()
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
func (s FinalizedHeadersCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s FinalizedHeadersCapn) WriteCapLit(w io.Writer) error {
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
	_, err = b.WriteString("firstPendingMetaBlock = ")
	if err != nil {
		return err
	}
	{
		s := s.FirstPendingMetaBlock()
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
	_, err = b.WriteString("lastFinishedMetaBlock = ")
	if err != nil {
		return err
	}
	{
		s := s.LastFinishedMetaBlock()
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
	_, err = b.WriteString("pendingMiniBlockHeaders = ")
	if err != nil {
		return err
	}
	{
		s := s.PendingMiniBlockHeaders()
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
func (s FinalizedHeadersCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type FinalizedHeadersCapn_List C.PointerList

func NewFinalizedHeadersCapnList(s *C.Segment, sz int) FinalizedHeadersCapn_List {
	return FinalizedHeadersCapn_List(s.NewCompositeList(8, 5, sz))
}
func (s FinalizedHeadersCapn_List) Len() int { return C.PointerList(s).Len() }
func (s FinalizedHeadersCapn_List) At(i int) FinalizedHeadersCapn {
	return FinalizedHeadersCapn(C.PointerList(s).At(i).ToStruct())
}
func (s FinalizedHeadersCapn_List) ToArray() []FinalizedHeadersCapn {
	n := s.Len()
	a := make([]FinalizedHeadersCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s FinalizedHeadersCapn_List) Set(i int, item FinalizedHeadersCapn) {
	C.PointerList(s).Set(i, C.Object(item))
}

type EpochStartCapn C.Struct

func NewEpochStartCapn(s *C.Segment) EpochStartCapn      { return EpochStartCapn(s.NewStruct(0, 1)) }
func NewRootEpochStartCapn(s *C.Segment) EpochStartCapn  { return EpochStartCapn(s.NewRootStruct(0, 1)) }
func AutoNewEpochStartCapn(s *C.Segment) EpochStartCapn  { return EpochStartCapn(s.NewStructAR(0, 1)) }
func ReadRootEpochStartCapn(s *C.Segment) EpochStartCapn { return EpochStartCapn(s.Root(0).ToStruct()) }
func (s EpochStartCapn) LastFinalizedHeaders() FinalizedHeadersCapn_List {
	return FinalizedHeadersCapn_List(C.Struct(s).GetObject(0))
}
func (s EpochStartCapn) SetLastFinalizedHeaders(v FinalizedHeadersCapn_List) {
	C.Struct(s).SetObject(0, C.Object(v))
}
func (s EpochStartCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"lastFinalizedHeaders\":")
	if err != nil {
		return err
	}
	{
		s := s.LastFinalizedHeaders()
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
func (s EpochStartCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s EpochStartCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("lastFinalizedHeaders = ")
	if err != nil {
		return err
	}
	{
		s := s.LastFinalizedHeaders()
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
func (s EpochStartCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type EpochStartCapn_List C.PointerList

func NewEpochStartCapnList(s *C.Segment, sz int) EpochStartCapn_List {
	return EpochStartCapn_List(s.NewCompositeList(0, 1, sz))
}
func (s EpochStartCapn_List) Len() int { return C.PointerList(s).Len() }
func (s EpochStartCapn_List) At(i int) EpochStartCapn {
	return EpochStartCapn(C.PointerList(s).At(i).ToStruct())
}
func (s EpochStartCapn_List) ToArray() []EpochStartCapn {
	n := s.Len()
	a := make([]EpochStartCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s EpochStartCapn_List) Set(i int, item EpochStartCapn) { C.PointerList(s).Set(i, C.Object(item)) }

type MetaBlockCapn C.Struct

func NewMetaBlockCapn(s *C.Segment) MetaBlockCapn      { return MetaBlockCapn(s.NewStruct(32, 13)) }
func NewRootMetaBlockCapn(s *C.Segment) MetaBlockCapn  { return MetaBlockCapn(s.NewRootStruct(32, 13)) }
func AutoNewMetaBlockCapn(s *C.Segment) MetaBlockCapn  { return MetaBlockCapn(s.NewStructAR(32, 13)) }
func ReadRootMetaBlockCapn(s *C.Segment) MetaBlockCapn { return MetaBlockCapn(s.Root(0).ToStruct()) }
func (s MetaBlockCapn) Nonce() uint64                  { return C.Struct(s).Get64(0) }
func (s MetaBlockCapn) SetNonce(v uint64)              { C.Struct(s).Set64(0, v) }
func (s MetaBlockCapn) Epoch() uint32                  { return C.Struct(s).Get32(8) }
func (s MetaBlockCapn) SetEpoch(v uint32)              { C.Struct(s).Set32(8, v) }
func (s MetaBlockCapn) Round() uint64                  { return C.Struct(s).Get64(16) }
func (s MetaBlockCapn) SetRound(v uint64)              { C.Struct(s).Set64(16, v) }
func (s MetaBlockCapn) TimeStamp() uint64              { return C.Struct(s).Get64(24) }
func (s MetaBlockCapn) SetTimeStamp(v uint64)          { C.Struct(s).Set64(24, v) }
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
func (s MetaBlockCapn) LeaderSignature() []byte         { return C.Struct(s).GetObject(3).ToData() }
func (s MetaBlockCapn) SetLeaderSignature(v []byte)     { C.Struct(s).SetObject(3, s.Segment.NewData(v)) }
func (s MetaBlockCapn) PubKeysBitmap() []byte           { return C.Struct(s).GetObject(4).ToData() }
func (s MetaBlockCapn) SetPubKeysBitmap(v []byte)       { C.Struct(s).SetObject(4, s.Segment.NewData(v)) }
func (s MetaBlockCapn) PrevHash() []byte                { return C.Struct(s).GetObject(5).ToData() }
func (s MetaBlockCapn) SetPrevHash(v []byte)            { C.Struct(s).SetObject(5, s.Segment.NewData(v)) }
func (s MetaBlockCapn) PrevRandSeed() []byte            { return C.Struct(s).GetObject(6).ToData() }
func (s MetaBlockCapn) SetPrevRandSeed(v []byte)        { C.Struct(s).SetObject(6, s.Segment.NewData(v)) }
func (s MetaBlockCapn) RandSeed() []byte                { return C.Struct(s).GetObject(7).ToData() }
func (s MetaBlockCapn) SetRandSeed(v []byte)            { C.Struct(s).SetObject(7, s.Segment.NewData(v)) }
func (s MetaBlockCapn) RootHash() []byte                { return C.Struct(s).GetObject(8).ToData() }
func (s MetaBlockCapn) SetRootHash(v []byte)            { C.Struct(s).SetObject(8, s.Segment.NewData(v)) }
func (s MetaBlockCapn) ValidatorStatsRootHash() []byte  { return C.Struct(s).GetObject(9).ToData() }
func (s MetaBlockCapn) SetValidatorStatsRootHash(v []byte) {
	C.Struct(s).SetObject(9, s.Segment.NewData(v))
}
func (s MetaBlockCapn) TxCount() uint32     { return C.Struct(s).Get32(12) }
func (s MetaBlockCapn) SetTxCount(v uint32) { C.Struct(s).Set32(12, v) }
func (s MetaBlockCapn) MiniBlockHeaders() MiniBlockHeaderCapn_List {
	return MiniBlockHeaderCapn_List(C.Struct(s).GetObject(10))
}
func (s MetaBlockCapn) SetMiniBlockHeaders(v MiniBlockHeaderCapn_List) {
	C.Struct(s).SetObject(10, C.Object(v))
}
func (s MetaBlockCapn) EpochStart() EpochStartCapn {
	return EpochStartCapn(C.Struct(s).GetObject(11).ToStruct())
}
func (s MetaBlockCapn) SetEpochStart(v EpochStartCapn) { C.Struct(s).SetObject(11, C.Object(v)) }
func (s MetaBlockCapn) Chainid() []byte                { return C.Struct(s).GetObject(12).ToData() }
func (s MetaBlockCapn) SetChainid(v []byte)            { C.Struct(s).SetObject(12, s.Segment.NewData(v)) }
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
	_, err = b.WriteString("\"leaderSignature\":")
	if err != nil {
		return err
	}
	{
		s := s.LeaderSignature()
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
	_, err = b.WriteString("\"validatorStatsRootHash\":")
	if err != nil {
		return err
	}
	{
		s := s.ValidatorStatsRootHash()
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
	_, err = b.WriteString("\"txCount\":")
	if err != nil {
		return err
	}
	{
		s := s.TxCount()
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
	_, err = b.WriteString("\"miniBlockHeaders\":")
	if err != nil {
		return err
	}
	{
		s := s.MiniBlockHeaders()
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
	_, err = b.WriteString("\"epochStart\":")
	if err != nil {
		return err
	}
	{
		s := s.EpochStart()
		err = s.WriteJSON(b)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"chainid\":")
	if err != nil {
		return err
	}
	{
		s := s.Chainid()
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
	_, err = b.WriteString("leaderSignature = ")
	if err != nil {
		return err
	}
	{
		s := s.LeaderSignature()
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
	_, err = b.WriteString("validatorStatsRootHash = ")
	if err != nil {
		return err
	}
	{
		s := s.ValidatorStatsRootHash()
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
	_, err = b.WriteString("txCount = ")
	if err != nil {
		return err
	}
	{
		s := s.TxCount()
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
	_, err = b.WriteString("miniBlockHeaders = ")
	if err != nil {
		return err
	}
	{
		s := s.MiniBlockHeaders()
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
	_, err = b.WriteString("epochStart = ")
	if err != nil {
		return err
	}
	{
		s := s.EpochStart()
		err = s.WriteCapLit(b)
		if err != nil {
			return err
		}
	}
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("chainid = ")
	if err != nil {
		return err
	}
	{
		s := s.Chainid()
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
	return MetaBlockCapn_List(s.NewCompositeList(32, 13, sz))
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
