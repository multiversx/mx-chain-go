package block

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block/capnp"
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
	// TxBlock identifies a miniblock holding transactions
	TxBlock Type = 0
	// StateBlock identifies a miniblock holding account state
	StateBlock Type = 1
	// PeerBlock identifies a miniblock holding peer assignation
	PeerBlock Type = 2
	// SmartContractResultBlock identifies a miniblock holding smartcontractresults
	SmartContractResultBlock Type = 3
	// RewardsBlock identifies a miniblock holding accumulated rewards, both system generated and from tx fees
	RewardsBlock Type = 4
	// InvalidBlock identifies identifies an invalid miniblock
	InvalidBlock Type = 5
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
	case SmartContractResultBlock:
		return "SmartContractResultBody"
	case RewardsBlock:
		return "RewardsBody"
	case InvalidBlock:
		return "InvalidBlock"
	default:
		return fmt.Sprintf("Unknown(%d)", bType)
	}
}

// MiniBlock holds the transactions and the sender/destination shard ids
type MiniBlock struct {
	TxHashes        [][]byte `capid:"0"`
	ReceiverShardID uint32   `capid:"1"`
	SenderShardID   uint32   `capid:"2"`
	Type            Type     `capid:"3"`
}

// MiniBlockHeader holds the hash of a miniblock together with sender/deastination shard id pair.
// The shard ids are both kept in order to differentiate between cross and single shard transactions
type MiniBlockHeader struct {
	Hash            []byte `capid:"0"`
	SenderShardID   uint32 `capid:"1"`
	ReceiverShardID uint32 `capid:"2"`
	TxCount         uint32 `capid:"3"`
	Type            Type   `capid:"4"`
}

// PeerChange holds a change in one peer to shard assignation
type PeerChange struct {
	PubKey      []byte `capid:"0"`
	ShardIdDest uint32 `capid:"1"`
}

// Header holds the metadata of a block. This is the part that is being hashed and run through consensus.
// The header holds the hash of the body and also the link to the previous block header hash
type Header struct {
	Nonce                  uint64            `capid:"0"`
	PrevHash               []byte            `capid:"1"`
	PrevRandSeed           []byte            `capid:"2"`
	RandSeed               []byte            `capid:"3"`
	PubKeysBitmap          []byte            `capid:"4"`
	ShardId                uint32            `capid:"5"`
	TimeStamp              uint64            `capid:"6"`
	Round                  uint64            `capid:"7"`
	Epoch                  uint32            `capid:"8"`
	BlockBodyType          Type              `capid:"9"`
	Signature              []byte            `capid:"10"`
	LeaderSignature        []byte            `capid:"11"`
	MiniBlockHeaders       []MiniBlockHeader `capid:"12"`
	PeerChanges            []PeerChange      `capid:"13"`
	RootHash               []byte            `capid:"14"`
	ValidatorStatsRootHash []byte            `capid:"15"`
	MetaBlockHashes        [][]byte          `capid:"16"`
	EpochStartMetaHash     []byte            `capid:"17"`
	TxCount                uint32            `capid:"18"`
	ChainID                []byte            `capid:"19"`
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
	dest.LeaderSignature = src.LeaderSignature()
	dest.EpochStartMetaHash = src.EpochStartMetaHash()
	dest.ChainID = src.Chainid()

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
	dest.ValidatorStatsRootHash = src.ValidatorStatsRootHash()

	var n int
	n = src.MetaHdrHashes().Len()
	dest.MetaBlockHashes = make([][]byte, n)
	for i := 0; i < n; i++ {
		dest.MetaBlockHashes[i] = src.MetaHdrHashes().At(i)
	}

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
	dest.SetLeaderSignature(src.LeaderSignature)
	dest.SetChainid(src.ChainID)
	dest.SetEpochStartMetaHash(src.EpochStartMetaHash)

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
	dest.SetValidatorStatsRootHash(src.ValidatorStatsRootHash)

	mylist1 := seg.NewDataList(len(src.MetaBlockHashes))
	for i := range src.MetaBlockHashes {
		mylist1.Set(i, src.MetaBlockHashes[i])
	}
	dest.SetMetaHdrHashes(mylist1)

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
	dest.Type = Type(src.Type())

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
	dest.SetType(uint8(src.Type))

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
	dest.Type = Type(src.Type())

	return dest
}

// MiniBlockHeaderGoToCapn is a helper function to copy fields from a MiniBlockHeader object to a MiniBlockHeaderCapn object
func MiniBlockHeaderGoToCapn(seg *capn.Segment, src *MiniBlockHeader) capnp.MiniBlockHeaderCapn {
	dest := capnp.AutoNewMiniBlockHeaderCapn(seg)

	dest.SetHash(src.Hash)
	dest.SetReceiverShardID(src.ReceiverShardID)
	dest.SetSenderShardID(src.SenderShardID)
	dest.SetTxCount(src.TxCount)
	dest.SetType(uint8(src.Type))

	return dest
}

// GetShardID returns header shard id
func (h *Header) GetShardID() uint32 {
	return h.ShardId
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
func (h *Header) GetRound() uint64 {
	return h.Round
}

// GetRootHash returns the roothash from header
func (h *Header) GetRootHash() []byte {
	return h.RootHash
}

// GetValidatorStatsRootHash returns the root hash for the validator statistics trie at this current block
func (h *Header) GetValidatorStatsRootHash() []byte {
	return h.ValidatorStatsRootHash
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

// GetLeaderSignature returns the leader's signature
func (h *Header) GetLeaderSignature() []byte {
	return h.LeaderSignature
}

// GetChainID gets the chain ID on which this block is valid on
func (h *Header) GetChainID() []byte {
	return h.ChainID
}

// GetTimeStamp returns the time stamp
func (h *Header) GetTimeStamp() uint64 {
	return h.TimeStamp
}

// GetTxCount returns transaction count in the block associated with this header
func (h *Header) GetTxCount() uint32 {
	return h.TxCount
}

// SetShardID sets header shard ID
func (h *Header) SetShardID(shId uint32) {
	h.ShardId = shId
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
func (h *Header) SetRound(r uint64) {
	h.Round = r
}

// SetRootHash sets root hash
func (h *Header) SetRootHash(rHash []byte) {
	h.RootHash = rHash
}

// SetValidatorStatsRootHash set's the root hash for the validator statistics trie
func (h *Header) SetValidatorStatsRootHash(rHash []byte) {
	h.ValidatorStatsRootHash = rHash
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

// SetLeaderSignature will set the leader's signature
func (h *Header) SetLeaderSignature(sg []byte) {
	h.LeaderSignature = sg
}

// SetChainID sets the chain ID on which this block is valid on
func (h *Header) SetChainID(chainID []byte) {
	h.ChainID = chainID
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

// MapMiniBlockHashesToShards is a map of mini block hashes and sender IDs
func (h *Header) MapMiniBlockHashesToShards() map[string]uint32 {
	hashDst := make(map[string]uint32, 0)
	for _, val := range h.MiniBlockHeaders {
		hashDst[string(val.Hash)] = val.SenderShardID
	}
	return hashDst
}

// Clone returns a clone of the object
func (h *Header) Clone() data.HeaderHandler {
	headerCopy := *h
	return &headerCopy
}

// IntegrityAndValidity checks if data is valid
func (b Body) IntegrityAndValidity() error {
	if b == nil || b.IsInterfaceNil() {
		return data.ErrNilBlockBody
	}

	for i := 0; i < len(b); i++ {
		if len(b[i].TxHashes) == 0 {
			return data.ErrMiniBlockEmpty
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (b Body) IsInterfaceNil() bool {
	if b == nil {
		return true
	}
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *Header) IsInterfaceNil() bool {
	if h == nil {
		return true
	}
	return false
}

// IsStartOfEpochBlock verifies if the block is of type start of epoch
func (h *Header) IsStartOfEpochBlock() bool {
	return len(h.EpochStartMetaHash) > 0
}

// ItemsInHeader gets the number of items(hashes) added in block header
func (h *Header) ItemsInHeader() uint32 {
	itemsInHeader := len(h.MiniBlockHeaders) + len(h.PeerChanges) + len(h.MetaBlockHashes)
	return uint32(itemsInHeader)
}

// ItemsInBody gets the number of items(hashes) added in block body
func (h *Header) ItemsInBody() uint32 {
	return h.TxCount
}

// CheckChainID returns nil if the header's chain ID matches the one provided
// otherwise, it will error
func (h *Header) CheckChainID(reference []byte) error {
	if !bytes.Equal(h.ChainID, reference) {
		return fmt.Errorf(
			"%w, expected: %s, got %s",
			data.ErrInvalidChainID,
			hex.EncodeToString(reference),
			hex.EncodeToString(h.ChainID),
		)
	}

	return nil
}
