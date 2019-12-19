package block

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block/capnp"
	"github.com/ElrondNetwork/elrond-go/sharding"
	capn "github.com/glycerine/go-capnproto"
)

// PeerAction type represents the possible events that a node can trigger for the metachain to notarize
type PeerAction uint8

// Constants mapping the actions that a node can take
const (
	PeerRegistrantion PeerAction = iota + 1
	PeerUnstaking
	PeerDeregistration
	PeerJailed
	PeerUnJailed
	PeerSlashed
	PeerReStake
)

func (pa PeerAction) String() string {
	switch pa {
	case PeerRegistrantion:
		return "PeerRegistration"
	case PeerUnstaking:
		return "PeerUnstaking"
	case PeerDeregistration:
		return "PeerDeregistration"
	case PeerJailed:
		return "PeerJailed"
	case PeerUnJailed:
		return "PeerUnjailed"
	case PeerSlashed:
		return "PeerSlashed"
	case PeerReStake:
		return "PeerReStake"
	default:
		return fmt.Sprintf("Unknown type (%d)", pa)
	}
}

// PeerData holds information about actions taken by a peer:
//  - a peer can register with an amount to become a validator
//  - a peer can choose to deregister and get back the deposited value
type PeerData struct {
	Address     []byte     `capid:"0"`
	PublicKey   []byte     `capid:"1"`
	Action      PeerAction `capid:"2"`
	TimeStamp   uint64     `capid:"3"`
	ValueChange *big.Int   `capid:"4"`
}

// ShardMiniBlockHeader holds data for one shard miniblock header
type ShardMiniBlockHeader struct {
	Hash            []byte `capid:"0"`
	ReceiverShardID uint32 `capid:"1"`
	SenderShardID   uint32 `capid:"2"`
	TxCount         uint32 `capid:"3"`
}

// ShardData holds the block information sent by the shards to the metachain
type ShardData struct {
	ShardID               uint32                 `capid:"0"`
	HeaderHash            []byte                 `capid:"1"`
	ShardMiniBlockHeaders []ShardMiniBlockHeader `capid:"2"`
	PrevRandSeed          []byte                 `capid:"3"`
	PubKeysBitmap         []byte                 `capid:"4"`
	Signature             []byte                 `capid:"5"`
	TxCount               uint32                 `capid:"6"`
	Round                 uint64                 `capid:"7"`
	PrevHash              []byte                 `capid:"8"`
	Nonce                 uint64                 `capid:"9"`
}

// EpochStartShardData hold the last finalized headers hash and state root hash
type EpochStartShardData struct {
	ShardId                 uint32                 `capid:"0"`
	HeaderHash              []byte                 `capid:"1"`
	RootHash                []byte                 `capid:"2"`
	FirstPendingMetaBlock   []byte                 `capid:"3"`
	LastFinishedMetaBlock   []byte                 `capid:"4"`
	PendingMiniBlockHeaders []ShardMiniBlockHeader `capid:"5"`
}

// EpochStart holds the block information for end-of-epoch
type EpochStart struct {
	LastFinalizedHeaders []EpochStartShardData `capid:"1"`
}

// MetaBlock holds the data that will be saved to the metachain each round
type MetaBlock struct {
	Nonce                  uint64            `capid:"0"`
	Epoch                  uint32            `capid:"1"`
	Round                  uint64            `capid:"2"`
	TimeStamp              uint64            `capid:"3"`
	ShardInfo              []ShardData       `capid:"4"`
	PeerInfo               []PeerData        `capid:"5"`
	Signature              []byte            `capid:"6"`
	LeaderSignature        []byte            `capid:"7"`
	PubKeysBitmap          []byte            `capid:"8"`
	PrevHash               []byte            `capid:"9"`
	PrevRandSeed           []byte            `capid:"10"`
	RandSeed               []byte            `capid:"11"`
	RootHash               []byte            `capid:"12"`
	ValidatorStatsRootHash []byte            `capid:"13"`
	TxCount                uint32            `capid:"14"`
	MiniBlockHeaders       []MiniBlockHeader `capid:"15"`
	EpochStart             EpochStart        `capid:"15"`
	ChainID                []byte            `capid:"16"`
}


// Save saves the serialized data of a PeerData into a stream through Capnp protocol
func (p *PeerData) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	PeerDataGoToCapn(seg, p)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a PeerData object through Capnp protocol
func (p *PeerData) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootPeerDataCapn(capMsg)
	PeerDataCapnToGo(z, p)
	return nil
}

// Save saves the serialized data of a ShardData into a stream through Capnp protocol
func (s *ShardData) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	ShardDataGoToCapn(seg, s)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a ShardData object through Capnp protocol
func (s *ShardData) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootShardDataCapn(capMsg)
	ShardDataCapnToGo(z, s)
	return nil
}

// Save saves the serialized data of a MetaBlock into a stream through Capnp protocol
func (m *MetaBlock) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	MetaBlockGoToCapn(seg, m)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a MetaBlock object through Capnp protocol
func (m *MetaBlock) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootMetaBlockCapn(capMsg)
	MetaBlockCapnToGo(z, m)
	return nil
}

// Save saves the serialized data of a ShardData into a stream through Capnp protocol
func (e *EpochStart) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	EpochStartGoToCapn(seg, *e)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a EpochStart object through Capnp protocol
func (e *EpochStart) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootEpochStartCapn(capMsg)
	EpochStartCapnToGo(z, e)
	return nil
}

// PeerDataGoToCapn is a helper function to copy fields from a Peer Data object to a PeerDataCapn object
func PeerDataGoToCapn(seg *capn.Segment, src *PeerData) capnp.PeerDataCapn {
	dest := capnp.AutoNewPeerDataCapn(seg)
	value, _ := src.ValueChange.GobEncode()
	dest.SetAddress(src.Address)
	dest.SetPublicKey(src.PublicKey)
	dest.SetAction(uint8(src.Action))
	dest.SetTimestamp(src.TimeStamp)
	dest.SetValue(value)

	return dest
}

// PeerDataCapnToGo is a helper function to copy fields from a PeerDataCapn object to a PeerData object
func PeerDataCapnToGo(src capnp.PeerDataCapn, dest *PeerData) *PeerData {
	if dest == nil {
		dest = &PeerData{}
	}
	if dest.ValueChange == nil {
		dest.ValueChange = big.NewInt(0)
	}
	dest.Address = src.Address()
	dest.PublicKey = src.PublicKey()
	dest.Action = PeerAction(src.Action())
	dest.TimeStamp = src.Timestamp()
	err := dest.ValueChange.GobDecode(src.Value())
	if err != nil {
		return nil
	}
	return dest
}

// ShardMiniBlockHeaderGoToCapn is a helper function to copy fields from a ShardMiniBlockHeader object to a
// ShardMiniBlockHeaderCapn object
func ShardMiniBlockHeaderGoToCapn(seg *capn.Segment, src *ShardMiniBlockHeader) capnp.ShardMiniBlockHeaderCapn {
	dest := capnp.AutoNewShardMiniBlockHeaderCapn(seg)

	dest.SetHash(src.Hash)
	dest.SetReceiverShardId(src.ReceiverShardID)
	dest.SetSenderShardId(src.SenderShardID)
	dest.SetTxCount(src.TxCount)

	return dest
}

// ShardMiniBlockHeaderCapnToGo is a helper function to copy fields from a ShardMiniBlockHeaderCapn object to a
// ShardMiniBlockHeader object
func ShardMiniBlockHeaderCapnToGo(src capnp.ShardMiniBlockHeaderCapn, dest *ShardMiniBlockHeader) *ShardMiniBlockHeader {
	if dest == nil {
		dest = &ShardMiniBlockHeader{}
	}
	dest.Hash = src.Hash()
	dest.ReceiverShardID = src.ReceiverShardId()
	dest.SenderShardID = src.SenderShardId()
	dest.TxCount = src.TxCount()

	return dest
}

// ShardDataGoToCapn is a helper function to copy fields from a ShardData object to a ShardDataCapn object
func ShardDataGoToCapn(seg *capn.Segment, src *ShardData) capnp.ShardDataCapn {
	dest := capnp.AutoNewShardDataCapn(seg)

	dest.SetShardId(src.ShardID)
	dest.SetHeaderHash(src.HeaderHash)
	dest.SetPrevRandSeed(src.PrevRandSeed)
	dest.SetPubKeysBitmap(src.PubKeysBitmap)
	dest.SetSignature(src.Signature)

	// create the list of shardMiniBlockHeaders
	if len(src.ShardMiniBlockHeaders) > 0 {
		typedList := capnp.NewShardMiniBlockHeaderCapnList(seg, len(src.ShardMiniBlockHeaders))
		plist := capn.PointerList(typedList)

		for i, elem := range src.ShardMiniBlockHeaders {
			_ = plist.Set(i, capn.Object(ShardMiniBlockHeaderGoToCapn(seg, &elem)))
		}
		dest.SetShardMiniBlockHeaders(typedList)
	}
	dest.SetTxCount(src.TxCount)

	return dest
}

// ShardDataCapnToGo is a helper function to copy fields from a ShardDataCapn object to a ShardData object
func ShardDataCapnToGo(src capnp.ShardDataCapn, dest *ShardData) *ShardData {
	if dest == nil {
		dest = &ShardData{}
	}
	dest.ShardID = src.ShardId()
	dest.HeaderHash = src.HeaderHash()
	dest.PrevRandSeed = src.PrevRandSeed()
	dest.PubKeysBitmap = src.PubKeysBitmap()
	dest.Signature = src.Signature()

	n := src.ShardMiniBlockHeaders().Len()
	dest.ShardMiniBlockHeaders = make([]ShardMiniBlockHeader, n)
	for i := 0; i < n; i++ {
		dest.ShardMiniBlockHeaders[i] = *ShardMiniBlockHeaderCapnToGo(src.ShardMiniBlockHeaders().At(i), nil)
	}
	dest.TxCount = src.TxCount()

	return dest
}

// EpochStartShardDataGoToCapn is a helper function to copy fields from a FinalizedHeaderHeader object to a
// EpochStartShardDataCapn object
func EpochStartShardDataGoToCapn(seg *capn.Segment, src *EpochStartShardData) capnp.FinalizedHeadersCapn {
	dest := capnp.AutoNewFinalizedHeadersCapn(seg)

	dest.SetRootHash(src.RootHash)
	dest.SetHeaderHash(src.HeaderHash)
	dest.SetShardId(src.ShardId)
	dest.SetFirstPendingMetaBlock(src.FirstPendingMetaBlock)
	dest.SetLastFinishedMetaBlock(src.LastFinishedMetaBlock)

	if len(src.PendingMiniBlockHeaders) > 0 {
		typedList := capnp.NewShardMiniBlockHeaderCapnList(seg, len(src.PendingMiniBlockHeaders))
		plist := capn.PointerList(typedList)

		for i, elem := range src.PendingMiniBlockHeaders {
			_ = plist.Set(i, capn.Object(ShardMiniBlockHeaderGoToCapn(seg, &elem)))
		}
		dest.SetPendingMiniBlockHeaders(typedList)
	}

	return dest
}

// EpochStartShardDataCapnToGo is a helper function to copy fields from a FinalizedHeaderCapn object to a
// EpochStartShardData object
func EpochStartShardDataCapnToGo(src capnp.FinalizedHeadersCapn, dest *EpochStartShardData) *EpochStartShardData {
	if dest == nil {
		dest = &EpochStartShardData{}
	}

	dest.RootHash = src.RootHash()
	dest.HeaderHash = src.HeaderHash()
	dest.ShardId = src.ShardId()
	dest.FirstPendingMetaBlock = src.FirstPendingMetaBlock()
	dest.LastFinishedMetaBlock = src.LastFinishedMetaBlock()

	n := src.PendingMiniBlockHeaders().Len()
	dest.PendingMiniBlockHeaders = make([]ShardMiniBlockHeader, n)
	for i := 0; i < n; i++ {
		dest.PendingMiniBlockHeaders[i] = *ShardMiniBlockHeaderCapnToGo(src.PendingMiniBlockHeaders().At(i), nil)
	}

	return dest
}

// EpochStartGoToCapn is a helper function to copy fields from a ShardData object to a ShardDataCapn object
func EpochStartGoToCapn(seg *capn.Segment, src EpochStart) capnp.EpochStartCapn {
	dest := capnp.AutoNewEpochStartCapn(seg)

	if len(src.LastFinalizedHeaders) > 0 {
		typedList := capnp.NewFinalizedHeadersCapnList(seg, len(src.LastFinalizedHeaders))
		pList := capn.PointerList(typedList)

		for i, elem := range src.LastFinalizedHeaders {
			_ = pList.Set(i, capn.Object(EpochStartShardDataGoToCapn(seg, &elem)))
		}
		dest.SetLastFinalizedHeaders(typedList)
	}

	return dest
}

// EpochStartCapnToGo is a helper function to copy fields from a ShardDataCapn object to a ShardData object
func EpochStartCapnToGo(src capnp.EpochStartCapn, dest *EpochStart) *EpochStart {
	if dest == nil {
		dest = &EpochStart{}
	}

	n := src.LastFinalizedHeaders().Len()
	dest.LastFinalizedHeaders = make([]EpochStartShardData, n)
	for i := 0; i < n; i++ {
		dest.LastFinalizedHeaders[i] = *EpochStartShardDataCapnToGo(src.LastFinalizedHeaders().At(i), nil)
	}

	return dest
}

// MetaBlockGoToCapn is a helper function to copy fields from a MetaBlock object to a MetaBlockCapn object
func MetaBlockGoToCapn(seg *capn.Segment, src *MetaBlock) capnp.MetaBlockCapn {
	dest := capnp.AutoNewMetaBlockCapn(seg)

	if len(src.ShardInfo) > 0 {
		typedList := capnp.NewShardDataCapnList(seg, len(src.ShardInfo))
		plist := capn.PointerList(typedList)

		for i, elem := range src.ShardInfo {
			_ = plist.Set(i, capn.Object(ShardDataGoToCapn(seg, &elem)))
		}
		dest.SetShardInfo(typedList)
	}

	if len(src.PeerInfo) > 0 {
		typedList := capnp.NewPeerDataCapnList(seg, len(src.PeerInfo))
		plist := capn.PointerList(typedList)

		for i, elem := range src.PeerInfo {
			_ = plist.Set(i, capn.Object(PeerDataGoToCapn(seg, &elem)))
		}
		dest.SetPeerInfo(typedList)
	}

	if len(src.MiniBlockHeaders) > 0 {
		miniBlockList := capnp.NewMiniBlockHeaderCapnList(seg, len(src.MiniBlockHeaders))
		pList := capn.PointerList(miniBlockList)

		for i, elem := range src.MiniBlockHeaders {
			_ = pList.Set(i, capn.Object(MiniBlockHeaderGoToCapn(seg, &elem)))
		}
		dest.SetMiniBlockHeaders(miniBlockList)
	}

	dest.SetSignature(src.Signature)
	dest.SetPubKeysBitmap(src.PubKeysBitmap)
	dest.SetPrevHash(src.PrevHash)
	dest.SetPrevRandSeed(src.PrevRandSeed)
	dest.SetRandSeed(src.RandSeed)
	dest.SetRootHash(src.RootHash)
	dest.SetValidatorStatsRootHash(src.ValidatorStatsRootHash)
	dest.SetTxCount(src.TxCount)
	dest.SetNonce(src.Nonce)
	dest.SetEpoch(src.Epoch)
	dest.SetRound(src.Round)
	dest.SetTimeStamp(src.TimeStamp)
	dest.SetEpochStart(EpochStartGoToCapn(seg, src.EpochStart))
	dest.SetLeaderSignature(src.LeaderSignature)
	dest.SetChainid(src.ChainID)

	return dest
}

// MetaBlockCapnToGo is a helper function to copy fields from a MetaBlockCapn object to a MetaBlock object
func MetaBlockCapnToGo(src capnp.MetaBlockCapn, dest *MetaBlock) *MetaBlock {
	if dest == nil {
		dest = &MetaBlock{}
	}

	n := src.ShardInfo().Len()
	dest.ShardInfo = make([]ShardData, n)
	for i := 0; i < n; i++ {
		dest.ShardInfo[i] = *ShardDataCapnToGo(src.ShardInfo().At(i), nil)
	}
	n = src.PeerInfo().Len()
	dest.PeerInfo = make([]PeerData, n)
	for i := 0; i < n; i++ {
		dest.PeerInfo[i] = *PeerDataCapnToGo(src.PeerInfo().At(i), nil)
	}

	mbLength := src.MiniBlockHeaders().Len()
	dest.MiniBlockHeaders = make([]MiniBlockHeader, mbLength)
	for i := 0; i < mbLength; i++ {
		dest.MiniBlockHeaders[i] = *MiniBlockHeaderCapnToGo(src.MiniBlockHeaders().At(i), nil)
	}

	dest.Signature = src.Signature()
	dest.PubKeysBitmap = src.PubKeysBitmap()
	dest.PrevHash = src.PrevHash()
	dest.PrevRandSeed = src.PrevRandSeed()
	dest.RandSeed = src.RandSeed()
	dest.RootHash = src.RootHash()
	dest.ValidatorStatsRootHash = src.ValidatorStatsRootHash()
	dest.TxCount = src.TxCount()
	dest.LeaderSignature = src.LeaderSignature()
	dest.Nonce = src.Nonce()
	dest.Epoch = src.Epoch()
	dest.Round = src.Round()
	dest.TimeStamp = src.TimeStamp()
	dest.EpochStart = *EpochStartCapnToGo(src.EpochStart(), nil)
	dest.ChainID = src.Chainid()

	return dest
}

// GetShardID returns the metachain shard id
func (m *MetaBlock) GetShardID() uint32 {
	return sharding.MetachainShardId
}

// GetNonce return header nonce
func (m *MetaBlock) GetNonce() uint64 {
	return m.Nonce
}

// GetEpoch return header epoch
func (m *MetaBlock) GetEpoch() uint32 {
	return m.Epoch
}

// GetRound return round from header
func (m *MetaBlock) GetRound() uint64 {
	return m.Round
}

// GetTimeStamp returns the time stamp
func (m *MetaBlock) GetTimeStamp() uint64 {
	return m.TimeStamp
}

// GetRootHash returns the roothash from header
func (m *MetaBlock) GetRootHash() []byte {
	return m.RootHash
}

// GetValidatorStatsRootHash returns the root hash for the validator statistics trie at this current block
func (m *MetaBlock) GetValidatorStatsRootHash() []byte {
	return m.ValidatorStatsRootHash
}

// GetPrevHash returns previous block header hash
func (m *MetaBlock) GetPrevHash() []byte {
	return m.PrevHash
}

// GetPrevRandSeed gets the previous random seed
func (m *MetaBlock) GetPrevRandSeed() []byte {
	return m.PrevRandSeed
}

// GetRandSeed gets the current random seed
func (m *MetaBlock) GetRandSeed() []byte {
	return m.RandSeed
}

// GetPubKeysBitmap return signers bitmap
func (m *MetaBlock) GetPubKeysBitmap() []byte {
	return m.PubKeysBitmap
}

// GetSignature return signed data
func (m *MetaBlock) GetSignature() []byte {
	return m.Signature
}

// GetLeaderSignature returns the signature of the leader
func (m *MetaBlock) GetLeaderSignature() []byte {
	return m.LeaderSignature
}

// GetChainID gets the chain ID on which this block is valid on
func (m *MetaBlock) GetChainID() []byte {
	return m.ChainID
}

// GetTxCount returns transaction count in the current meta block
func (m *MetaBlock) GetTxCount() uint32 {
	return m.TxCount
}

// SetShardID sets header shard ID
func (m *MetaBlock) SetShardID(shId uint32) {
}

// SetNonce sets header nonce
func (m *MetaBlock) SetNonce(n uint64) {
	m.Nonce = n
}

// SetEpoch sets header epoch
func (m *MetaBlock) SetEpoch(e uint32) {
	m.Epoch = e
}

// SetRound sets header round
func (m *MetaBlock) SetRound(r uint64) {
	m.Round = r
}

// SetRootHash sets root hash
func (m *MetaBlock) SetRootHash(rHash []byte) {
	m.RootHash = rHash
}

// SetValidatorStatsRootHash set's the root hash for the validator statistics trie
func (m *MetaBlock) SetValidatorStatsRootHash(rHash []byte) {
	m.ValidatorStatsRootHash = rHash
}

// SetPrevHash sets prev hash
func (m *MetaBlock) SetPrevHash(pvHash []byte) {
	m.PrevHash = pvHash
}

// SetPrevRandSeed sets the previous randomness seed
func (m *MetaBlock) SetPrevRandSeed(pvRandSeed []byte) {
	m.PrevRandSeed = pvRandSeed
}

// SetRandSeed sets the current random seed
func (m *MetaBlock) SetRandSeed(randSeed []byte) {
	m.RandSeed = randSeed
}

// SetPubKeysBitmap sets publick key bitmap
func (m *MetaBlock) SetPubKeysBitmap(pkbm []byte) {
	m.PubKeysBitmap = pkbm
}

// SetSignature set header signature
func (m *MetaBlock) SetSignature(sg []byte) {
	m.Signature = sg
}

// SetLeaderSignature will set the leader's signature
func (m *MetaBlock) SetLeaderSignature(sg []byte) {
	m.LeaderSignature = sg
}

// SetChainID sets the chain ID on which this block is valid on
func (m *MetaBlock) SetChainID(chainID []byte) {
	m.ChainID = chainID
}

// SetTimeStamp sets header timestamp
func (m *MetaBlock) SetTimeStamp(ts uint64) {
	m.TimeStamp = ts
}

// SetTxCount sets the transaction count of the current meta block
func (m *MetaBlock) SetTxCount(txCount uint32) {
	m.TxCount = txCount
}

// GetMiniBlockHeadersWithDst as a map of hashes and sender IDs
func (m *MetaBlock) GetMiniBlockHeadersWithDst(destId uint32) map[string]uint32 {
	hashDst := make(map[string]uint32, 0)
	for i := 0; i < len(m.ShardInfo); i++ {
		if m.ShardInfo[i].ShardID == destId {
			continue
		}

		for _, val := range m.ShardInfo[i].ShardMiniBlockHeaders {
			if val.ReceiverShardID == destId && val.SenderShardID != destId {
				hashDst[string(val.Hash)] = val.SenderShardID
			}
		}
	}

	for _, val := range m.MiniBlockHeaders {
		if val.ReceiverShardID == destId && val.SenderShardID != destId {
			hashDst[string(val.Hash)] = val.SenderShardID
		}
	}

	return hashDst
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *MetaBlock) IsInterfaceNil() bool {
	if m == nil {
		return true
	}
	return false
}

// ItemsInHeader gets the number of items(hashes) added in block header
func (m *MetaBlock) ItemsInHeader() uint32 {
	itemsInHeader := len(m.ShardInfo)
	for i := 0; i < len(m.ShardInfo); i++ {
		itemsInHeader += len(m.ShardInfo[i].ShardMiniBlockHeaders)
	}

	itemsInHeader += len(m.PeerInfo)
	itemsInHeader += len(m.MiniBlockHeaders)

	return uint32(itemsInHeader)
}

// IsStartOfEpochBlock verifies if the block is of type start of epoch
func (m *MetaBlock) IsStartOfEpochBlock() bool {
	return len(m.EpochStart.LastFinalizedHeaders) > 0
}

// ItemsInBody gets the number of items(hashes) added in block body
func (m *MetaBlock) ItemsInBody() uint32 {
	return m.TxCount
}

// Clone will return a clone of the object
func (m *MetaBlock) Clone() data.HeaderHandler {
	metaBlockCopy := *m
	return &metaBlockCopy
}

// CheckChainID returns nil if the header's chain ID matches the one provided
// otherwise, it will error
func (m *MetaBlock) CheckChainID(reference []byte) error {
	if !bytes.Equal(m.ChainID, reference) {
		return fmt.Errorf(
			"%w, expected: %s, got %s",
			data.ErrInvalidChainID,
			hex.EncodeToString(reference),
			hex.EncodeToString(m.ChainID),
		)
	}

	return nil
}
