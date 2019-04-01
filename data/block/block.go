package block

import (
	"fmt"
	"io"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block/capnp"
	"github.com/glycerine/go-capnproto"
)

// This file holds the data structures related with the functionality of a shard block
//
// MiniBlock structure represents the body of a transaction block, holding an array of miniblocks
// each of the miniblocks has a different destination shard
// The body can be transmitted even before having built the heder and go through a prevalidation of each transaction

// Type identifies the type of the block
type Type uint8

// Body should be used when referring to the full list of mini blocks that forms a block body
type Body []*MiniBlock

// MiniBlockSlice should be used when referring to subset of mini blocks that is not
//  necessarily representing a full block body
type MiniBlockSlice []*MiniBlock

const (
	// TxBlock identifies a block holding transactions
	TxBlock Type = 0
	// StateBlock identifies a block holding account state
	StateBlock Type = 1
	// PeerBlock identifies a block holding peer assignation
	PeerBlock Type = 2
)

// String returns the string representation of the Type
func (bType Type) String() string {
	switch bType {
	case TxBlock:
		return "TxBody"
	case StateBlock:
		return "StateBody"
	case PeerBlock:
		return "PeerBody"
	default:
		return fmt.Sprintf("Unknown(%d)", bType)
	}
}

// MiniBlock holds the transactions and the sender/destination shard ids
type MiniBlock struct {
	TxHashes        [][]byte `capid:"0"`
	ReceiverShardID uint32   `capid:"1"`
	SenderShardID   uint32   `capid:"2"`
}

// MiniBlockHeader holds the hash of a miniblock together with sender/deastination shard id pair.
// The shard ids are both kept in order to differentiate between cross and single shard transactions
type MiniBlockHeader struct {
	Hash            []byte `capid:"0"`
	SenderShardID   uint32 `capid:"1"`
	ReceiverShardID uint32 `capid:"2"`
	TxCount         uint32 `capid:"3"`
}

// PeerChange holds a change in one peer to shard assignation
type PeerChange struct {
	PubKey      []byte `capid:"0"`
	ShardIdDest uint32 `capid:"1"`
}

// Header holds the metadata of a block. This is the part that is being hashed and run through consensus.
// The header holds the hash of the body and also the link to the previous block header hash
type Header struct {
	Nonce            uint64            `capid:"0"`
	PrevHash         []byte            `capid:"1"`
	PrevRandSeed     []byte            `capid:"2"`
	RandSeed         []byte            `capid:"3"`
	PubKeysBitmap    []byte            `capid:"4"`
	ShardId          uint32            `capid:"5"`
	TimeStamp        uint64            `capid:"6"`
	Round            uint32            `capid:"7"`
	Epoch            uint32            `capid:"8"`
	BlockBodyType    Type              `capid:"9"`
	Signature        []byte            `capid:"10"`
	MiniBlockHeaders []MiniBlockHeader `capid:"11"`
	PeerChanges      []PeerChange      `capid:"12"`
	RootHash         []byte            `capid:"13"`
	TxCount          uint32            `capid:"14"`
	processedMBs     map[string]bool   // TODO remove this field when metachain processing is running
}

// Save saves the serialized data of a Block Header into a stream through Capnp protocol
func (h *Header) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	HeaderGoToCapn(seg, h)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a Block Header object through Capnp protocol
func (h *Header) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootHeaderCapn(capMsg)
	HeaderCapnToGo(z, h)
	return nil
}

// HeaderCapnToGo is a helper function to copy fields from a HeaderCapn object to a Header object
func HeaderCapnToGo(src capnp.HeaderCapn, dest *Header) *Header {
	if dest == nil {
		dest = &Header{}
	}

	dest.Nonce = src.Nonce()
	dest.PrevHash = src.PrevHash()
	dest.PrevRandSeed = src.PrevRandSeed()
	dest.RandSeed = src.RandSeed()
	dest.PubKeysBitmap = src.PubKeysBitmap()
	dest.ShardId = src.ShardId()
	dest.TimeStamp = src.TimeStamp()
	dest.Round = src.Round()
	dest.Epoch = src.Epoch()
	dest.BlockBodyType = Type(src.BlockBodyType())
	dest.Signature = src.Signature()

	mbLength := src.MiniBlockHeaders().Len()
	dest.MiniBlockHeaders = make([]MiniBlockHeader, mbLength)
	for i := 0; i < mbLength; i++ {
		dest.MiniBlockHeaders[i] = *MiniBlockHeaderCapnToGo(src.MiniBlockHeaders().At(i), nil)
	}

	peerChangesLen := src.PeerChanges().Len()
	dest.PeerChanges = make([]PeerChange, peerChangesLen)
	for i := 0; i < peerChangesLen; i++ {
		dest.PeerChanges[i] = *PeerChangeCapnToGo(src.PeerChanges().At(i), nil)
	}

	dest.RootHash = src.RootHash()
	dest.TxCount = src.TxCount()

	return dest
}

// HeaderGoToCapn is a helper function to copy fields from a Block Header object to a HeaderCapn object
func HeaderGoToCapn(seg *capn.Segment, src *Header) capnp.HeaderCapn {
	dest := capnp.AutoNewHeaderCapn(seg)

	dest.SetNonce(src.Nonce)
	dest.SetPrevHash(src.PrevHash)
	dest.SetPrevRandSeed(src.PrevRandSeed)
	dest.SetRandSeed(src.RandSeed)
	dest.SetPubKeysBitmap(src.PubKeysBitmap)
	dest.SetShardId(src.ShardId)
	dest.SetTimeStamp(src.TimeStamp)
	dest.SetRound(src.Round)
	dest.SetEpoch(src.Epoch)
	dest.SetBlockBodyType(uint8(src.BlockBodyType))
	dest.SetSignature(src.Signature)
	if len(src.MiniBlockHeaders) > 0 {
		miniBlockList := capnp.NewMiniBlockHeaderCapnList(seg, len(src.MiniBlockHeaders))
		pList := capn.PointerList(miniBlockList)

		for i, elem := range src.MiniBlockHeaders {
			_ = pList.Set(i, capn.Object(MiniBlockHeaderGoToCapn(seg, &elem)))
		}
		dest.SetMiniBlockHeaders(miniBlockList)
	}

	if len(src.PeerChanges) > 0 {
		peerChangeList := capnp.NewPeerChangeCapnList(seg, len(src.PeerChanges))
		plist := capn.PointerList(peerChangeList)

		for i, elem := range src.PeerChanges {
			_ = plist.Set(i, capn.Object(PeerChangeGoToCapn(seg, &elem)))
		}
		dest.SetPeerChanges(peerChangeList)
	}

	dest.SetRootHash(src.RootHash)
	dest.SetTxCount(src.TxCount)

	return dest
}

// Save saves the serialized data of a MiniBlock into a stream through Capnp protocol
func (s *MiniBlock) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	MiniBlockGoToCapn(seg, s)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a MiniBlock object through Capnp protocol
func (s *MiniBlock) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootMiniBlockCapn(capMsg)
	MiniBlockCapnToGo(z, s)
	return nil
}

// MiniBlockCapnToGo is a helper function to copy fields from a MiniBlockCapn object to a MiniBlock object
func MiniBlockCapnToGo(src capnp.MiniBlockCapn, dest *MiniBlock) *MiniBlock {
	if dest == nil {
		dest = &MiniBlock{}
	}

	var n int

	n = src.TxHashes().Len()
	dest.TxHashes = make([][]byte, n)
	for i := 0; i < n; i++ {
		dest.TxHashes[i] = src.TxHashes().At(i)
	}

	dest.ReceiverShardID = src.ReceiverShardID()
	dest.SenderShardID = src.SenderShardID()

	return dest
}

// MiniBlockGoToCapn is a helper function to copy fields from a MiniBlock object to a MiniBlockCapn object
func MiniBlockGoToCapn(seg *capn.Segment, src *MiniBlock) capnp.MiniBlockCapn {
	dest := capnp.AutoNewMiniBlockCapn(seg)

	mylist1 := seg.NewDataList(len(src.TxHashes))
	for i := range src.TxHashes {
		mylist1.Set(i, src.TxHashes[i])
	}
	dest.SetTxHashes(mylist1)
	dest.SetReceiverShardID(src.ReceiverShardID)
	dest.SetSenderShardID(src.SenderShardID)

	return dest
}

// Save saves the serialized data of a PeerChange into a stream through Capnp protocol
func (s *PeerChange) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	PeerChangeGoToCapn(seg, s)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a PeerChange object through Capnp protocol
func (s *PeerChange) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootPeerChangeCapn(capMsg)
	PeerChangeCapnToGo(z, s)
	return nil
}

// PeerChangeCapnToGo is a helper function to copy fields from a PeerChangeCapn object to a PeerChange object
func PeerChangeCapnToGo(src capnp.PeerChangeCapn, dest *PeerChange) *PeerChange {
	if dest == nil {
		dest = &PeerChange{}
	}

	dest.PubKey = src.PubKey()
	dest.ShardIdDest = src.ShardIdDest()

	return dest
}

// PeerChangeGoToCapn is a helper function to copy fields from a PeerChange object to a PeerChangeGoToCapn object
func PeerChangeGoToCapn(seg *capn.Segment, src *PeerChange) capnp.PeerChangeCapn {
	dest := capnp.AutoNewPeerChangeCapn(seg)
	dest.SetPubKey(src.PubKey)
	dest.SetShardIdDest(src.ShardIdDest)

	return dest
}

// Save saves the serialized data of a StateBlockBody into a stream through Capnp protocol
func (s *MiniBlockHeader) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	MiniBlockHeaderGoToCapn(seg, s)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a StateBlockBody object through Capnp protocol
func (s *MiniBlockHeader) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootMiniBlockHeaderCapn(capMsg)
	MiniBlockHeaderCapnToGo(z, s)
	return nil
}

// MiniBlockHeaderCapnToGo is a helper function to copy fields from a MiniBlockHeaderCapn object to a MiniBlockHeader object
func MiniBlockHeaderCapnToGo(src capnp.MiniBlockHeaderCapn, dest *MiniBlockHeader) *MiniBlockHeader {
	if dest == nil {
		dest = &MiniBlockHeader{}
	}
	dest.Hash = src.Hash()
	dest.ReceiverShardID = src.ReceiverShardID()
	dest.SenderShardID = src.SenderShardID()
	dest.TxCount = src.TxCount()

	return dest
}

// MiniBlockHeaderGoToCapn is a helper function to copy fields from a MiniBlockHeader object to a MiniBlockHeaderCapn object
func MiniBlockHeaderGoToCapn(seg *capn.Segment, src *MiniBlockHeader) capnp.MiniBlockHeaderCapn {
	dest := capnp.AutoNewMiniBlockHeaderCapn(seg)

	dest.SetHash(src.Hash)
	dest.SetReceiverShardID(src.ReceiverShardID)
	dest.SetSenderShardID(src.SenderShardID)
	dest.SetTxCount(src.TxCount)

	return dest
}

// GetNonce returns header nonce
func (h *Header) GetNonce() uint64 {
	return h.Nonce
}

// GetEpoch returns header epoch
func (h *Header) GetEpoch() uint32 {
	return h.Epoch
}

// GetRound returns round from header
func (h *Header) GetRound() uint32 {
	return h.Round
}

// GetRootHash returns the roothash from header
func (h *Header) GetRootHash() []byte {
	return h.RootHash
}

// GetPrevHash returns previous block header hash
func (h *Header) GetPrevHash() []byte {
	return h.PrevHash
}

// GetPrevRandSeed returns previous random seed
func (h *Header) GetPrevRandSeed() []byte {
	return h.PrevRandSeed
}

// GetRandSeed returns the random seed
func (h *Header) GetRandSeed() []byte {
	return h.RandSeed
}

// GetPubKeysBitmap return signers bitmap
func (h *Header) GetPubKeysBitmap() []byte {
	return h.PubKeysBitmap
}

// GetSignature returns signed data
func (h *Header) GetSignature() []byte {
	return h.Signature
}

// GetTimestamp returns the time stamp
func (h *Header) GetTimestamp() uint64 {
	return h.TimeStamp
}

// GetTxCount returns transaction count in the block associated with this header
func (h *Header) GetTxCount() uint32 {
	return h.TxCount
}

// SetNonce sets header nonce
func (h *Header) SetNonce(n uint64) {
	h.Nonce = n
}

// SetEpoch sets header epoch
func (h *Header) SetEpoch(e uint32) {
	h.Epoch = e
}

// SetRound sets header round
func (h *Header) SetRound(r uint32) {
	h.Round = r
}

// SetRootHash sets root hash
func (h *Header) SetRootHash(rHash []byte) {
	h.RootHash = rHash
}

// SetPrevHash sets prev hash
func (h *Header) SetPrevHash(pvHash []byte) {
	h.PrevHash = pvHash
}

// SetPrevRandSeed sets previous random seed
func (h *Header) SetPrevRandSeed(pvRandSeed []byte) {
	h.PrevRandSeed = pvRandSeed
}

// SetRandSeed sets previous random seed
func (h *Header) SetRandSeed(randSeed []byte) {
	h.RandSeed = randSeed
}

// SetPubKeysBitmap sets publick key bitmap
func (h *Header) SetPubKeysBitmap(pkbm []byte) {
	h.PubKeysBitmap = pkbm
}

// SetSignature sets header signature
func (h *Header) SetSignature(sg []byte) {
	h.Signature = sg
}

// SetTimeStamp sets header timestamp
func (h *Header) SetTimeStamp(ts uint64) {
	h.TimeStamp = ts
}

// SetTxCount sets the transaction count of the block associated with this header
func (h *Header) SetTxCount(txCount uint32) {
	h.TxCount = txCount
}

// GetMiniBlockHeadersWithDst as a map of hashes and sender IDs
func (h *Header) GetMiniBlockHeadersWithDst(destId uint32) map[string]uint32 {
	hashDst := make(map[string]uint32, 0)
	for _, val := range h.MiniBlockHeaders {
		if val.ReceiverShardID == destId && val.SenderShardID != destId {
			hashDst[string(val.Hash)] = val.SenderShardID
		}
	}
	return hashDst
}

// GetMiniBlockProcessed verifies if miniblock from header was processed
func (h *Header) GetMiniBlockProcessed(hash []byte) bool {
	if h.processedMBs == nil {
		h.processedMBs = make(map[string]bool, 0)
	}
	return h.processedMBs[string(hash)]
}

// SetMiniBlockProcessed set that miniblock with hash to processed
func (h *Header) SetMiniBlockProcessed(hash []byte) {
	if h.processedMBs == nil {
		h.processedMBs = make(map[string]bool, 0)
	}
	h.processedMBs[string(hash)] = true
}

// IntegrityAndValidity checks if data is valid
func (b Body) IntegrityAndValidity() error {
	if b == nil {
		return data.ErrNilBlockBody
	}

	for i := 0; i < len(b); i++ {
		if len(b[i].TxHashes) == 0 {
			return data.ErrMiniBlockEmpty
		}
	}

	return nil
}
