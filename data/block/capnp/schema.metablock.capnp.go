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

func NewPeerDataCapn(s *C.Segment) PeerDataCapn      { return PeerDataCapn(s.NewStruct(24, 1)) }
func NewRootPeerDataCapn(s *C.Segment) PeerDataCapn  { return PeerDataCapn(s.NewRootStruct(24, 1)) }
func AutoNewPeerDataCapn(s *C.Segment) PeerDataCapn  { return PeerDataCapn(s.NewStructAR(24, 1)) }
func ReadRootPeerDataCapn(s *C.Segment) PeerDataCapn { return PeerDataCapn(s.Root(0).ToStruct()) }
func (s PeerDataCapn) PublicKey() []byte             { return C.Struct(s).GetObject(0).ToData() }
func (s PeerDataCapn) SetPublicKey(v []byte)         { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s PeerDataCapn) Action() uint8                 { return C.Struct(s).Get8(0) }
func (s PeerDataCapn) SetAction(v uint8)             { C.Struct(s).Set8(0, v) }
func (s PeerDataCapn) Timestamp() uint64             { return C.Struct(s).Get64(8) }
func (s PeerDataCapn) SetTimestamp(v uint64)         { C.Struct(s).Set64(8, v) }
func (s PeerDataCapn) Value() uint64                 { return C.Struct(s).Get64(16) }
func (s PeerDataCapn) SetValue(v uint64)             { C.Struct(s).Set64(16, v) }
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
	return PeerDataCapn_List(s.NewCompositeList(24, 1, sz))
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

type ShardDataCapn C.Struct

func NewShardDataCapn(s *C.Segment) ShardDataCapn      { return ShardDataCapn(s.NewStruct(8, 1)) }
func NewRootShardDataCapn(s *C.Segment) ShardDataCapn  { return ShardDataCapn(s.NewRootStruct(8, 1)) }
func AutoNewShardDataCapn(s *C.Segment) ShardDataCapn  { return ShardDataCapn(s.NewStructAR(8, 1)) }
func ReadRootShardDataCapn(s *C.Segment) ShardDataCapn { return ShardDataCapn(s.Root(0).ToStruct()) }
func (s ShardDataCapn) ShardId() uint32                { return C.Struct(s).Get32(0) }
func (s ShardDataCapn) SetShardId(v uint32)            { C.Struct(s).Set32(0, v) }
func (s ShardDataCapn) HeaderHashes() C.DataList       { return C.DataList(C.Struct(s).GetObject(0)) }
func (s ShardDataCapn) SetHeaderHashes(v C.DataList)   { C.Struct(s).SetObject(0, C.Object(v)) }
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
	_, err = b.WriteString("\"headerHashes\":")
	if err != nil {
		return err
	}
	{
		s := s.HeaderHashes()
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
	_, err = b.WriteString("headerHashes = ")
	if err != nil {
		return err
	}
	{
		s := s.HeaderHashes()
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
	return ShardDataCapn_List(s.NewCompositeList(8, 1, sz))
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

type ProofCapn C.Struct

func NewProofCapn(s *C.Segment) ProofCapn      { return ProofCapn(s.NewStruct(0, 2)) }
func NewRootProofCapn(s *C.Segment) ProofCapn  { return ProofCapn(s.NewRootStruct(0, 2)) }
func AutoNewProofCapn(s *C.Segment) ProofCapn  { return ProofCapn(s.NewStructAR(0, 2)) }
func ReadRootProofCapn(s *C.Segment) ProofCapn { return ProofCapn(s.Root(0).ToStruct()) }
func (s ProofCapn) InclusionProof() []byte     { return C.Struct(s).GetObject(0).ToData() }
func (s ProofCapn) SetInclusionProof(v []byte) { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s ProofCapn) ExclusionProof() []byte     { return C.Struct(s).GetObject(1).ToData() }
func (s ProofCapn) SetExclusionProof(v []byte) { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s ProofCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"inclusionProof\":")
	if err != nil {
		return err
	}
	{
		s := s.InclusionProof()
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
	_, err = b.WriteString("\"exclusionProof\":")
	if err != nil {
		return err
	}
	{
		s := s.ExclusionProof()
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
func (s ProofCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s ProofCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("inclusionProof = ")
	if err != nil {
		return err
	}
	{
		s := s.InclusionProof()
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
	_, err = b.WriteString("exclusionProof = ")
	if err != nil {
		return err
	}
	{
		s := s.ExclusionProof()
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
func (s ProofCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type ProofCapn_List C.PointerList

func NewProofCapnList(s *C.Segment, sz int) ProofCapn_List {
	return ProofCapn_List(s.NewCompositeList(0, 2, sz))
}
func (s ProofCapn_List) Len() int           { return C.PointerList(s).Len() }
func (s ProofCapn_List) At(i int) ProofCapn { return ProofCapn(C.PointerList(s).At(i).ToStruct()) }
func (s ProofCapn_List) ToArray() []ProofCapn {
	n := s.Len()
	a := make([]ProofCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s ProofCapn_List) Set(i int, item ProofCapn) { C.PointerList(s).Set(i, C.Object(item)) }

type MetaBlockCapn C.Struct

func NewMetaBlockCapn(s *C.Segment) MetaBlockCapn      { return MetaBlockCapn(s.NewStruct(16, 3)) }
func NewRootMetaBlockCapn(s *C.Segment) MetaBlockCapn  { return MetaBlockCapn(s.NewRootStruct(16, 3)) }
func AutoNewMetaBlockCapn(s *C.Segment) MetaBlockCapn  { return MetaBlockCapn(s.NewStructAR(16, 3)) }
func ReadRootMetaBlockCapn(s *C.Segment) MetaBlockCapn { return MetaBlockCapn(s.Root(0).ToStruct()) }
func (s MetaBlockCapn) Nonce() uint64                  { return C.Struct(s).Get64(0) }
func (s MetaBlockCapn) SetNonce(v uint64)              { C.Struct(s).Set64(0, v) }
func (s MetaBlockCapn) Epoch() uint32                  { return C.Struct(s).Get32(8) }
func (s MetaBlockCapn) SetEpoch(v uint32)              { C.Struct(s).Set32(8, v) }
func (s MetaBlockCapn) Round() uint32                  { return C.Struct(s).Get32(12) }
func (s MetaBlockCapn) SetRound(v uint32)              { C.Struct(s).Set32(12, v) }
func (s MetaBlockCapn) ShardInfo() ShardDataCapn_List {
	return ShardDataCapn_List(C.Struct(s).GetObject(0))
}
func (s MetaBlockCapn) SetShardInfo(v ShardDataCapn_List) { C.Struct(s).SetObject(0, C.Object(v)) }
func (s MetaBlockCapn) PeerInfo() PeerDataCapn_List {
	return PeerDataCapn_List(C.Struct(s).GetObject(1))
}
func (s MetaBlockCapn) SetPeerInfo(v PeerDataCapn_List) { C.Struct(s).SetObject(1, C.Object(v)) }
func (s MetaBlockCapn) Proof() ProofCapn                { return ProofCapn(C.Struct(s).GetObject(2).ToStruct()) }
func (s MetaBlockCapn) SetProof(v ProofCapn)            { C.Struct(s).SetObject(2, C.Object(v)) }
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
	_, err = b.WriteString("\"proof\":")
	if err != nil {
		return err
	}
	{
		s := s.Proof()
		err = s.WriteJSON(b)
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
	_, err = b.WriteString("proof = ")
	if err != nil {
		return err
	}
	{
		s := s.Proof()
		err = s.WriteCapLit(b)
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
	return MetaBlockCapn_List(s.NewCompositeList(16, 3, sz))
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
