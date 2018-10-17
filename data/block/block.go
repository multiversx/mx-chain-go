package block

import (
	"io"

	"github.com/glycerine/go-capnproto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block/capnproto1"
)

type MiniBlock struct {
	TxHashes    [][]byte
	DestShardID uint32
}

type Block struct {
	MiniBlocks []MiniBlock
}

type Header struct {
	Nonce    []byte
	PrevHash []byte
	// temporary keep list of public keys of signers in header
	// to be removed later
	PubKeys    [][]byte
	ShardId    uint32
	TimeStamp  []byte
	Round      uint32
	BlockHash  []byte
	Signature  []byte
	Commitment []byte
}

func (s *Block) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	BlockGoToCapn(seg, s)
	_, err := seg.WriteTo(w)
	return err
}

func (s *Block) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		//panic(fmt.Errorf("capn.ReadFromStream error: %s", err))
		return err
	}
	z := capnproto1.ReadRootBlockCapn(capMsg)
	BlockCapnToGo(z, s)
	return nil
}

func BlockCapnToGo(src capnproto1.BlockCapn, dest *Block) *Block {
	if dest == nil {
		dest = &Block{}
	}

	var n int

	// MiniBlocks
	n = src.MiniBlocks().Len()
	dest.MiniBlocks = make([]MiniBlock, n)
	for i := 0; i < n; i++ {
		dest.MiniBlocks[i] = *MiniBlockCapnToGo(src.MiniBlocks().At(i), nil)
	}

	return dest
}

func BlockGoToCapn(seg *capn.Segment, src *Block) capnproto1.BlockCapn {
	dest := capnproto1.AutoNewBlockCapn(seg)

	// MiniBlocks -> MiniBlockCapn (go slice to capn list)
	if len(src.MiniBlocks) > 0 {
		typedList := capnproto1.NewMiniBlockCapnList(seg, len(src.MiniBlocks))
		plist := capn.PointerList(typedList)
		i := 0
		for _, ele := range src.MiniBlocks {
			plist.Set(i, capn.Object(MiniBlockGoToCapn(seg, &ele)))
			i++
		}
		dest.SetMiniBlocks(typedList)
	}

	return dest
}

func (s *Header) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	HeaderGoToCapn(seg, s)
	_, err := seg.WriteTo(w)
	return err
}

func (s *Header) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		//panic(fmt.Errorf("capn.ReadFromStream error: %s", err))
		return err
	}
	z := capnproto1.ReadRootHeaderCapn(capMsg)
	HeaderCapnToGo(z, s)
	return nil
}

func HeaderCapnToGo(src capnproto1.HeaderCapn, dest *Header) *Header {
	if dest == nil {
		dest = &Header{}
	}

	var n int

	// Nonce
	n = src.Nonce().Len()
	dest.Nonce = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.Nonce[i] = byte(src.Nonce().At(i))
	}

	// PrevHash
	n = src.PrevHash().Len()
	dest.PrevHash = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.PrevHash[i] = byte(src.PrevHash().At(i))
	}

	// PubKeys
	n = src.PubKeys().Len()
	dest.PubKeys = make([][]byte, n)
	for i := 0; i < n; i++ {
		dest.PubKeys[i] = UInt8ListToSliceByte(capn.UInt8List(src.PubKeys().At(i)))
	}

	dest.ShardId = src.ShardId()

	// TimeStamp
	n = src.TimeStamp().Len()
	dest.TimeStamp = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.TimeStamp[i] = byte(src.TimeStamp().At(i))
	}

	dest.Round = src.Round()

	// BlockHash
	n = src.BlockHash().Len()
	dest.BlockHash = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.BlockHash[i] = byte(src.BlockHash().At(i))
	}

	// Signature
	n = src.Signature().Len()
	dest.Signature = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.Signature[i] = byte(src.Signature().At(i))
	}

	// Commitment
	n = src.Commitment().Len()
	dest.Commitment = make([]byte, n)
	for i := 0; i < n; i++ {
		dest.Commitment[i] = byte(src.Commitment().At(i))
	}

	return dest
}

func HeaderGoToCapn(seg *capn.Segment, src *Header) capnproto1.HeaderCapn {
	dest := capnproto1.AutoNewHeaderCapn(seg)

	mylist1 := seg.NewUInt8List(len(src.Nonce))
	for i := range src.Nonce {
		mylist1.Set(i, uint8(src.Nonce[i]))
	}
	dest.SetNonce(mylist1)

	mylist2 := seg.NewUInt8List(len(src.PrevHash))
	for i := range src.PrevHash {
		mylist2.Set(i, uint8(src.PrevHash[i]))
	}
	dest.SetPrevHash(mylist2)

	mylist3 := seg.NewPointerList(len(src.PubKeys))
	for i := range src.PubKeys {
		mylist3.Set(i, capn.Object(SliceByteToUInt8List(seg, src.PubKeys[i])))
	}
	dest.SetPubKeys(mylist3)
	dest.SetShardId(src.ShardId)

	mylist4 := seg.NewUInt8List(len(src.TimeStamp))
	for i := range src.TimeStamp {
		mylist4.Set(i, uint8(src.TimeStamp[i]))
	}
	dest.SetTimeStamp(mylist4)
	dest.SetRound(src.Round)

	mylist5 := seg.NewUInt8List(len(src.BlockHash))
	for i := range src.BlockHash {
		mylist5.Set(i, uint8(src.BlockHash[i]))
	}
	dest.SetBlockHash(mylist5)

	mylist6 := seg.NewUInt8List(len(src.Signature))
	for i := range src.Signature {
		mylist6.Set(i, uint8(src.Signature[i]))
	}
	dest.SetSignature(mylist6)

	mylist7 := seg.NewUInt8List(len(src.Commitment))
	for i := range src.Commitment {
		mylist7.Set(i, uint8(src.Commitment[i]))
	}
	dest.SetCommitment(mylist7)

	return dest
}

func (s *MiniBlock) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	MiniBlockGoToCapn(seg, s)
	_, err := seg.WriteTo(w)
	return err
}

func (s *MiniBlock) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		//panic(fmt.Errorf("capn.ReadFromStream error: %s", err))
		return err
	}
	z := capnproto1.ReadRootMiniBlockCapn(capMsg)
	MiniBlockCapnToGo(z, s)
	return nil
}

func MiniBlockCapnToGo(src capnproto1.MiniBlockCapn, dest *MiniBlock) *MiniBlock {
	if dest == nil {
		dest = &MiniBlock{}
	}

	var n int

	// TxHashes
	n = src.TxHashes().Len()
	dest.TxHashes = make([][]byte, n)
	for i := 0; i < n; i++ {
		dest.TxHashes[i] = UInt8ListToSliceByte(capn.UInt8List(src.TxHashes().At(i)))
	}

	dest.DestShardID = src.DestShardID()

	return dest
}

func MiniBlockGoToCapn(seg *capn.Segment, src *MiniBlock) capnproto1.MiniBlockCapn {
	dest := capnproto1.AutoNewMiniBlockCapn(seg)

	mylist1 := seg.NewPointerList(len(src.TxHashes))
	for i := range src.TxHashes {
		mylist1.Set(i, capn.Object(SliceByteToUInt8List(seg, src.TxHashes[i])))
	}
	dest.SetTxHashes(mylist1)
	dest.SetDestShardID(src.DestShardID)

	return dest
}

func SliceByteToUInt8List(seg *capn.Segment, m []byte) capn.UInt8List {
	lst := seg.NewUInt8List(len(m))
	for i := range m {
		lst.Set(i, uint8(m[i]))
	}
	return lst
}

func UInt8ListToSliceByte(p capn.UInt8List) []byte {
	v := make([]byte, p.Len())
	for i := range v {
		v[i] = byte(p.At(i))
	}
	return v
}

func SliceMiniBlockToMiniBlockCapnList(seg *capn.Segment, m []MiniBlock) capnproto1.MiniBlockCapn_List {
	lst := capnproto1.NewMiniBlockCapnList(seg, len(m))
	for i := range m {
		lst.Set(i, MiniBlockGoToCapn(seg, &m[i]))
	}
	return lst
}

func MiniBlockCapnListToSliceMiniBlock(p capnproto1.MiniBlockCapn_List) []MiniBlock {
	v := make([]MiniBlock, p.Len())
	for i := range v {
		MiniBlockCapnToGo(p.At(i), &v[i])
	}
	return v
}
