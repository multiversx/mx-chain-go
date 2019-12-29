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

func NewHeaderCapn(s *C.Segment) HeaderCapn      { return HeaderCapn(s.NewStruct(40, 13)) }
func NewRootHeaderCapn(s *C.Segment) HeaderCapn  { return HeaderCapn(s.NewRootStruct(40, 13)) }
func AutoNewHeaderCapn(s *C.Segment) HeaderCapn  { return HeaderCapn(s.NewStructAR(40, 13)) }
func ReadRootHeaderCapn(s *C.Segment) HeaderCapn { return HeaderCapn(s.Root(0).ToStruct()) }
func (s HeaderCapn) Nonce() uint64               { return C.Struct(s).Get64(0) }
func (s HeaderCapn) SetNonce(v uint64)           { C.Struct(s).Set64(0, v) }
func (s HeaderCapn) PrevHash() []byte            { return C.Struct(s).GetObject(0).ToData() }
func (s HeaderCapn) SetPrevHash(v []byte)        { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s HeaderCapn) PrevRandSeed() []byte        { return C.Struct(s).GetObject(1).ToData() }
func (s HeaderCapn) SetPrevRandSeed(v []byte)    { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s HeaderCapn) RandSeed() []byte            { return C.Struct(s).GetObject(2).ToData() }
func (s HeaderCapn) SetRandSeed(v []byte)        { C.Struct(s).SetObject(2, s.Segment.NewData(v)) }
func (s HeaderCapn) PubKeysBitmap() []byte       { return C.Struct(s).GetObject(3).ToData() }
func (s HeaderCapn) SetPubKeysBitmap(v []byte)   { C.Struct(s).SetObject(3, s.Segment.NewData(v)) }
func (s HeaderCapn) ShardId() uint32             { return C.Struct(s).Get32(8) }
func (s HeaderCapn) SetShardId(v uint32)         { C.Struct(s).Set32(8, v) }
func (s HeaderCapn) TimeStamp() uint64           { return C.Struct(s).Get64(16) }
func (s HeaderCapn) SetTimeStamp(v uint64)       { C.Struct(s).Set64(16, v) }
func (s HeaderCapn) Round() uint64               { return C.Struct(s).Get64(24) }
func (s HeaderCapn) SetRound(v uint64)           { C.Struct(s).Set64(24, v) }
func (s HeaderCapn) Epoch() uint32               { return C.Struct(s).Get32(12) }
func (s HeaderCapn) SetEpoch(v uint32)           { C.Struct(s).Set32(12, v) }
func (s HeaderCapn) BlockBodyType() uint8        { return C.Struct(s).Get8(32) }
func (s HeaderCapn) SetBlockBodyType(v uint8)    { C.Struct(s).Set8(32, v) }
func (s HeaderCapn) Signature() []byte           { return C.Struct(s).GetObject(4).ToData() }
func (s HeaderCapn) SetSignature(v []byte)       { C.Struct(s).SetObject(4, s.Segment.NewData(v)) }
func (s HeaderCapn) LeaderSignature() []byte     { return C.Struct(s).GetObject(5).ToData() }
func (s HeaderCapn) SetLeaderSignature(v []byte) { C.Struct(s).SetObject(5, s.Segment.NewData(v)) }
func (s HeaderCapn) MiniBlockHeaders() MiniBlockHeaderCapn_List {
	return MiniBlockHeaderCapn_List(C.Struct(s).GetObject(6))
}
func (s HeaderCapn) SetMiniBlockHeaders(v MiniBlockHeaderCapn_List) {
	C.Struct(s).SetObject(6, C.Object(v))
}
func (s HeaderCapn) PeerChanges() PeerChangeCapn_List {
	return PeerChangeCapn_List(C.Struct(s).GetObject(7))
}
func (s HeaderCapn) SetPeerChanges(v PeerChangeCapn_List) { C.Struct(s).SetObject(7, C.Object(v)) }
func (s HeaderCapn) RootHash() []byte                     { return C.Struct(s).GetObject(8).ToData() }
func (s HeaderCapn) SetRootHash(v []byte)                 { C.Struct(s).SetObject(8, s.Segment.NewData(v)) }
func (s HeaderCapn) ValidatorStatsRootHash() []byte       { return C.Struct(s).GetObject(9).ToData() }
func (s HeaderCapn) SetValidatorStatsRootHash(v []byte) {
	C.Struct(s).SetObject(9, s.Segment.NewData(v))
}
func (s HeaderCapn) MetaHdrHashes() C.DataList      { return C.DataList(C.Struct(s).GetObject(10)) }
func (s HeaderCapn) SetMetaHdrHashes(v C.DataList)  { C.Struct(s).SetObject(10, C.Object(v)) }
func (s HeaderCapn) EpochStartMetaHash() []byte     { return C.Struct(s).GetObject(11).ToData() }
func (s HeaderCapn) SetEpochStartMetaHash(v []byte) { C.Struct(s).SetObject(11, s.Segment.NewData(v)) }
func (s HeaderCapn) TxCount() uint32                { return C.Struct(s).Get32(36) }
func (s HeaderCapn) SetTxCount(v uint32)            { C.Struct(s).Set32(36, v) }
func (s HeaderCapn) Chainid() []byte                { return C.Struct(s).GetObject(12).ToData() }
func (s HeaderCapn) SetChainid(v []byte)            { C.Struct(s).SetObject(12, s.Segment.NewData(v)) }
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
	_, err = b.WriteString("\"peerChanges\":")
	if err != nil {
		return err
	}
	{
		s := s.PeerChanges()
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
	_, err = b.WriteString("\"metaHdrHashes\":")
	if err != nil {
		return err
	}
	{
		s := s.MetaHdrHashes()
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
	_, err = b.WriteString("\"epochStartMetaHash\":")
	if err != nil {
		return err
	}
	{
		s := s.EpochStartMetaHash()
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
	_, err = b.WriteString("peerChanges = ")
	if err != nil {
		return err
	}
	{
		s := s.PeerChanges()
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
	_, err = b.WriteString("metaHdrHashes = ")
	if err != nil {
		return err
	}
	{
		s := s.MetaHdrHashes()
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
	_, err = b.WriteString("epochStartMetaHash = ")
	if err != nil {
		return err
	}
	{
		s := s.EpochStartMetaHash()
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
func (s HeaderCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type HeaderCapn_List C.PointerList

func NewHeaderCapnList(s *C.Segment, sz int) HeaderCapn_List {
	return HeaderCapn_List(s.NewCompositeList(40, 13, sz))
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

type MiniBlockHeaderCapn C.Struct

func NewMiniBlockHeaderCapn(s *C.Segment) MiniBlockHeaderCapn {
	return MiniBlockHeaderCapn(s.NewStruct(16, 1))
}
func NewRootMiniBlockHeaderCapn(s *C.Segment) MiniBlockHeaderCapn {
	return MiniBlockHeaderCapn(s.NewRootStruct(16, 1))
}
func AutoNewMiniBlockHeaderCapn(s *C.Segment) MiniBlockHeaderCapn {
	return MiniBlockHeaderCapn(s.NewStructAR(16, 1))
}
func ReadRootMiniBlockHeaderCapn(s *C.Segment) MiniBlockHeaderCapn {
	return MiniBlockHeaderCapn(s.Root(0).ToStruct())
}
func (s MiniBlockHeaderCapn) Hash() []byte                { return C.Struct(s).GetObject(0).ToData() }
func (s MiniBlockHeaderCapn) SetHash(v []byte)            { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s MiniBlockHeaderCapn) ReceiverShardID() uint32     { return C.Struct(s).Get32(0) }
func (s MiniBlockHeaderCapn) SetReceiverShardID(v uint32) { C.Struct(s).Set32(0, v) }
func (s MiniBlockHeaderCapn) SenderShardID() uint32       { return C.Struct(s).Get32(4) }
func (s MiniBlockHeaderCapn) SetSenderShardID(v uint32)   { C.Struct(s).Set32(4, v) }
func (s MiniBlockHeaderCapn) TxCount() uint32             { return C.Struct(s).Get32(8) }
func (s MiniBlockHeaderCapn) SetTxCount(v uint32)         { C.Struct(s).Set32(8, v) }
func (s MiniBlockHeaderCapn) Type() uint8                 { return C.Struct(s).Get8(12) }
func (s MiniBlockHeaderCapn) SetType(v uint8)             { C.Struct(s).Set8(12, v) }
func (s MiniBlockHeaderCapn) WriteJSON(w io.Writer) error {
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
	_, err = b.WriteString("\"receiverShardID\":")
	if err != nil {
		return err
	}
	{
		s := s.ReceiverShardID()
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
	_, err = b.WriteString("\"senderShardID\":")
	if err != nil {
		return err
	}
	{
		s := s.SenderShardID()
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
	_, err = b.WriteString("\"type\":")
	if err != nil {
		return err
	}
	{
		s := s.Type()
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
func (s MiniBlockHeaderCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s MiniBlockHeaderCapn) WriteCapLit(w io.Writer) error {
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
	_, err = b.WriteString("receiverShardID = ")
	if err != nil {
		return err
	}
	{
		s := s.ReceiverShardID()
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
	_, err = b.WriteString("senderShardID = ")
	if err != nil {
		return err
	}
	{
		s := s.SenderShardID()
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
	_, err = b.WriteString("type = ")
	if err != nil {
		return err
	}
	{
		s := s.Type()
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
func (s MiniBlockHeaderCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type MiniBlockHeaderCapn_List C.PointerList

func NewMiniBlockHeaderCapnList(s *C.Segment, sz int) MiniBlockHeaderCapn_List {
	return MiniBlockHeaderCapn_List(s.NewCompositeList(16, 1, sz))
}
func (s MiniBlockHeaderCapn_List) Len() int { return C.PointerList(s).Len() }
func (s MiniBlockHeaderCapn_List) At(i int) MiniBlockHeaderCapn {
	return MiniBlockHeaderCapn(C.PointerList(s).At(i).ToStruct())
}
func (s MiniBlockHeaderCapn_List) ToArray() []MiniBlockHeaderCapn {
	n := s.Len()
	a := make([]MiniBlockHeaderCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s MiniBlockHeaderCapn_List) Set(i int, item MiniBlockHeaderCapn) {
	C.PointerList(s).Set(i, C.Object(item))
}

type MiniBlockCapn C.Struct

func NewMiniBlockCapn(s *C.Segment) MiniBlockCapn      { return MiniBlockCapn(s.NewStruct(16, 1)) }
func NewRootMiniBlockCapn(s *C.Segment) MiniBlockCapn  { return MiniBlockCapn(s.NewRootStruct(16, 1)) }
func AutoNewMiniBlockCapn(s *C.Segment) MiniBlockCapn  { return MiniBlockCapn(s.NewStructAR(16, 1)) }
func ReadRootMiniBlockCapn(s *C.Segment) MiniBlockCapn { return MiniBlockCapn(s.Root(0).ToStruct()) }
func (s MiniBlockCapn) TxHashes() C.DataList           { return C.DataList(C.Struct(s).GetObject(0)) }
func (s MiniBlockCapn) SetTxHashes(v C.DataList)       { C.Struct(s).SetObject(0, C.Object(v)) }
func (s MiniBlockCapn) ReceiverShardID() uint32        { return C.Struct(s).Get32(0) }
func (s MiniBlockCapn) SetReceiverShardID(v uint32)    { C.Struct(s).Set32(0, v) }
func (s MiniBlockCapn) SenderShardID() uint32          { return C.Struct(s).Get32(4) }
func (s MiniBlockCapn) SetSenderShardID(v uint32)      { C.Struct(s).Set32(4, v) }
func (s MiniBlockCapn) Type() uint8                    { return C.Struct(s).Get8(8) }
func (s MiniBlockCapn) SetType(v uint8)                { C.Struct(s).Set8(8, v) }
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
	_, err = b.WriteString("\"receiverShardID\":")
	if err != nil {
		return err
	}
	{
		s := s.ReceiverShardID()
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
	_, err = b.WriteString("\"senderShardID\":")
	if err != nil {
		return err
	}
	{
		s := s.SenderShardID()
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
	_, err = b.WriteString("\"type\":")
	if err != nil {
		return err
	}
	{
		s := s.Type()
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
	_, err = b.WriteString("receiverShardID = ")
	if err != nil {
		return err
	}
	{
		s := s.ReceiverShardID()
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
	_, err = b.WriteString("senderShardID = ")
	if err != nil {
		return err
	}
	{
		s := s.SenderShardID()
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
	_, err = b.WriteString("type = ")
	if err != nil {
		return err
	}
	{
		s := s.Type()
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
	return MiniBlockCapn_List(s.NewCompositeList(16, 1, sz))
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
