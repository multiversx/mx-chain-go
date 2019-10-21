package block

import (
	"fmt"
	"io"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/block/capnp"
	"github.com/ElrondNetwork/elrond-go/sharding"
	capn "github.com/glycerine/go-capnproto"
)

// PeerAction type represents the possible events that a node can trigger for the metachain to notarize
type PeerAction uint8

// Constants mapping the actions that a node can take
const (
	PeerRegistrantion PeerAction = iota + 1
	PeerDeregistration
)

func (pa PeerAction) String() string {
	switch pa {
	case PeerRegistrantion:
		return "PeerRegistration"
	case PeerDeregistration:
		return "PeerDeregistration"
	default:
		return fmt.Sprintf("Unknown type (%d)", pa)
	}
}

// PeerData holds information about actions taken by a peer:
//  - a peer can register with an amount to become a validator
//  - a peer can choose to deregister and get back the deposited value
type PeerData struct {
	PublicKey []byte     `capid:"0"`
	Action    PeerAction `capid:"1"`
	TimeStamp uint64     `capid:"2"`
	Value     *big.Int   `capid:"3"`
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
	TxCount               uint32                 `capid:"3"`
}

// MetaBlock holds the data that will be saved to the metachain each round
type MetaBlock struct {
	Nonce            uint64            `capid:"0"`
	Epoch            uint32            `capid:"1"`
	Round            uint64            `capid:"2"`
	TimeStamp        uint64            `capid:"3"`
	ShardInfo        []ShardData       `capid:"4"`
	PeerInfo         []PeerData        `capid:"5"`
	Signature        []byte            `capid:"6"`
	PubKeysBitmap    []byte            `capid:"7"`
	PrevHash         []byte            `capid:"8"`
	PrevRandSeed     []byte            `capid:"9"`
	RandSeed         []byte            `capid:"10"`
	RootHash         []byte            `capid:"11"`
	TxCount          uint32            `capid:"12"`
	MiniBlockHeaders []MiniBlockHeader `capid:"13"`
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

// PeerDataGoToCapn is a helper function to copy fields from a Peer Data object to a PeerDataCapn object
func PeerDataGoToCapn(seg *capn.Segment, src *PeerData) capnp.PeerDataCapn {
	dest := capnp.AutoNewPeerDataCapn(seg)
	value, _ := src.Value.GobEncode()
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
	if dest.Value == nil {
		dest.Value = big.NewInt(0)
	}
	dest.PublicKey = src.PublicKey()
	dest.Action = PeerAction(src.Action())
	dest.TimeStamp = src.Timestamp()
	err := dest.Value.GobDecode(src.Value())
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

	n := src.ShardMiniBlockHeaders().Len()
	dest.ShardMiniBlockHeaders = make([]ShardMiniBlockHeader, n)
	for i := 0; i < n; i++ {
		dest.ShardMiniBlockHeaders[i] = *ShardMiniBlockHeaderCapnToGo(src.ShardMiniBlockHeaders().At(i), nil)
	}
	dest.TxCount = src.TxCount()

	return dest
}

// MetaBlockGoToCapn is a helper function to copy fields from a MetaBlock object to a MetaBlockCapn object
func MetaBlockGoToCapn(seg *capn.Segment, src *MetaBlock) capnp.MetaBlockCapn {
	dest := capnp.AutoNewMetaBlockCapn(seg)

	dest.SetNonce(src.Nonce)
	dest.SetEpoch(src.Epoch)
	dest.SetRound(src.Round)
	dest.SetTimeStamp(src.TimeStamp)

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
	dest.SetTxCount(src.TxCount)

	return dest
}

// MetaBlockCapnToGo is a helper function to copy fields from a MetaBlockCapn object to a MetaBlock object
func MetaBlockCapnToGo(src capnp.MetaBlockCapn, dest *MetaBlock) *MetaBlock {
	if dest == nil {
		dest = &MetaBlock{}
	}
	dest.Nonce = src.Nonce()
	dest.Epoch = src.Epoch()
	dest.Round = src.Round()
	dest.TimeStamp = src.TimeStamp()

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
	dest.TxCount = src.TxCount()

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

// GetTxCount returns transaction count in the current meta block
func (m *MetaBlock) GetTxCount() uint32 {
	return m.TxCount
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

// GetMiniBlockHeaders returns all the miniblock headers saved in the current metablock
func (m *MetaBlock) GetMiniBlockHeaders() []MiniBlockHeader {
	return m.MiniBlockHeaders
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

// ItemsInBody gets the number of items(hashes) added in block body
func (m *MetaBlock) ItemsInBody() uint32 {
	return m.TxCount
}
