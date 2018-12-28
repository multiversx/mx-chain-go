package block

import (
	"io"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block/capnp"
	"github.com/glycerine/go-capnproto"
)

// Block structure is the body of a block, holding an array of miniblocks
// each of the miniblocks has a different destination shard
// The body can be transmitted even before having built the heder and go through a prevalidation of each transaction

// BlockType identifies the type of the block
type BlockType uint8

const (
	// TxBlock identifies a block holding transactions
	TxBlock BlockType = 0
	// StateBlock identifies a block holding account state
	StateBlock BlockType = 1
	// PeerBlock identifies a block holding peer assignation
	PeerBlock BlockType = 2
)

// MiniBlock holds the transactions with one of the sender or recipient in node's shard and the other in ShardID
type MiniBlock struct {
	TxHashes [][]byte `capid:"0"`
	ShardID  uint32   `capid:"1"`
}

// StateBlockBody holds the merkle root hash committing to the state of the accounts
type StateBlockBody struct {
	RootHash []byte `capid:"0"`
	ShardID  uint32 `capid:"1"`
}

// PeerChange holds a change in one peer to shard assignation
type PeerChange struct {
	PubKey      []byte `capid:"0"`
	ShardIdDest uint32 `capid:"1"`
}

// PeerBlockBody holds multiple changes of peer to shard assignation
type PeerBlockBody struct {
	StateBlockBody `capid:"0"`
	Changes        []PeerChange `capid:"1"`
}

// TxBlockBody structure is the body of a transaction block, holding an array of miniblocks, each of the
// miniblocks has a different destination shard
// The body can be transmitted even before having built the header and go through a prevalidation of each transaction
type TxBlockBody struct {
	StateBlockBody `capid:"0"`
	MiniBlocks     []MiniBlock `capid:"1"`
}

// Header holds the metadata of a block. This is the part that is being hashed and run through consensus.
// The header holds the hash of the body and also the link to the previous block header hash
type Header struct {
	Nonce         uint64    `capid:"0"`
	PrevHash      []byte    `capid:"1"`
	PubKeysBitmap []byte    `capid:"2"`
	ShardId       uint32    `capid:"3"`
	TimeStamp     uint64    `capid:"4"`
	Round         uint32    `capid:"5"`
	Epoch         uint32    `capid:"6"`
	BlockBodyHash []byte    `capid:"7"`
	BlockBodyType BlockType `capid:"8"`
	Signature     []byte    `capid:"9"`
	Commitment    []byte    `capid:"10"`
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

	// Nonce
	dest.Nonce = src.Nonce()
	// PrevHash
	dest.PrevHash = src.PrevHash()
	// PubKeysBitmap
	dest.PubKeysBitmap = src.PubKeysBitmap()
	// ShardId
	dest.ShardId = src.ShardId()
	// TimeStamp
	dest.TimeStamp = src.TimeStamp()
	// Round
	dest.Round = src.Round()
	// Epoch
	dest.Epoch = src.Epoch()
	// BlockBodyHash
	dest.BlockBodyHash = src.BlockBodyHash()
	// BlockBodyType
	dest.BlockBodyType = BlockType(src.BlockBodyType())
	// Signature
	dest.Signature = src.Signature()
	// Commitment
	dest.Commitment = src.Commitment()

	return dest
}

// HeaderGoToCapn is a helper function to copy fields from a Block Header object to a HeaderCapn object
func HeaderGoToCapn(seg *capn.Segment, src *Header) capnp.HeaderCapn {
	dest := capnp.AutoNewHeaderCapn(seg)

	dest.SetNonce(src.Nonce)
	dest.SetPrevHash(src.PrevHash)
	dest.SetPubKeysBitmap(src.PubKeysBitmap)
	dest.SetShardId(src.ShardId)
	dest.SetTimeStamp(src.TimeStamp)
	dest.SetRound(src.Round)
	dest.SetEpoch(src.Epoch)
	dest.SetBlockBodyHash(src.BlockBodyHash)
	dest.SetBlockBodyType(uint8(src.BlockBodyType))
	dest.SetSignature(src.Signature)
	dest.SetCommitment(src.Commitment)

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

	// TxHashes
	n = src.TxHashes().Len()
	dest.TxHashes = make([][]byte, n)
	for i := 0; i < n; i++ {
		dest.TxHashes[i] = src.TxHashes().At(i)
	}

	dest.ShardID = src.ShardID()

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
	dest.SetShardID(src.ShardID)

	return dest
}

// Save saves the serialized data of a PeerBlockBody into a stream through Capnp protocol
func (s *PeerBlockBody) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	PeerBlockBodyGoToCapn(seg, s)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a PeerBlockBody object through Capnp protocol
func (s *PeerBlockBody) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootPeerBlockBodyCapn(capMsg)
	PeerBlockBodyCapnToGo(z, s)
	return nil
}

// PeerBlockBodyCapnToGo is a helper function to copy fields from a PeerBlockBodyCapn object to a PeerBlockBody object
func PeerBlockBodyCapnToGo(src capnp.PeerBlockBodyCapn, dest *PeerBlockBody) *PeerBlockBody {
	if dest == nil {
		dest = &PeerBlockBody{}
	}
	dest.StateBlockBody = *StateBlockBodyCapnToGo(src.StateBlockBody(), nil)

	var n int

	// Changes
	n = src.Changes().Len()
	dest.Changes = make([]PeerChange, n)
	for i := 0; i < n; i++ {
		dest.Changes[i] = *PeerChangeCapnToGo(src.Changes().At(i), nil)
	}

	return dest
}

// PeerBlockBodyGoToCapn is a helper function to copy fields from a PeerBlockBody object to a PeerBlockBodyCapn object
func PeerBlockBodyGoToCapn(seg *capn.Segment, src *PeerBlockBody) capnp.PeerBlockBodyCapn {
	dest := capnp.AutoNewPeerBlockBodyCapn(seg)
	dest.SetStateBlockBody(StateBlockBodyGoToCapn(seg, &src.StateBlockBody))

	// Changes -> PeerChangeCapn (go slice to capn list)
	if len(src.Changes) > 0 {
		typedList := capnp.NewPeerChangeCapnList(seg, len(src.Changes))
		plist := capn.PointerList(typedList)

		for i, elem := range src.Changes {
			plist.Set(i, capn.Object(PeerChangeGoToCapn(seg, &elem)))
		}
		dest.SetChanges(typedList)
	}

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

	// PubKey
	dest.PubKey = src.PubKey()
	// ShardIdDest
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
func (s *StateBlockBody) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	StateBlockBodyGoToCapn(seg, s)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a StateBlockBody object through Capnp protocol
func (s *StateBlockBody) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootStateBlockBodyCapn(capMsg)
	StateBlockBodyCapnToGo(z, s)
	return nil
}

// StateBlockBodyCapnToGo is a helper function to copy fields from StateBlockBodyCapn object to StateBlockBody object
func StateBlockBodyCapnToGo(src capnp.StateBlockBodyCapn, dest *StateBlockBody) *StateBlockBody {
	if dest == nil {
		dest = &StateBlockBody{}
	}

	// RootHash
	dest.RootHash = src.RootHash()
	// ShardId
	dest.ShardID = src.ShardID()

	return dest
}

// StateBlockBodyGoToCapn is a helper function to copy fields from a StateBlockBody object to a StateBlockBodyCapn object
func StateBlockBodyGoToCapn(seg *capn.Segment, src *StateBlockBody) capnp.StateBlockBodyCapn {
	dest := capnp.AutoNewStateBlockBodyCapn(seg)

	dest.SetRootHash(src.RootHash)
	dest.SetShardID(src.ShardID)

	return dest
}

// Save saves the serialized data of a TxBlockBody into a stream through Capnp protocol
func (txBlk *TxBlockBody) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	TxBlockBodyGoToCapn(seg, txBlk)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a TxBlockBody object through Capnp protocol
func (txBlk *TxBlockBody) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootTxBlockBodyCapn(capMsg)
	TxBlockBodyCapnToGo(z, txBlk)
	return nil
}

// TxBlockBodyCapnToGo is a helper function to copy fields from a TxBlockBodyCapn object to a TxBlockBody object
func TxBlockBodyCapnToGo(src capnp.TxBlockBodyCapn, dest *TxBlockBody) *TxBlockBody {
	if dest == nil {
		dest = &TxBlockBody{}
	}
	dest.StateBlockBody = *StateBlockBodyCapnToGo(src.StateBlockBody(), nil)

	var n int

	// MiniBlocks
	n = src.MiniBlocks().Len()
	dest.MiniBlocks = make([]MiniBlock, n)
	for i := 0; i < n; i++ {
		dest.MiniBlocks[i] = *MiniBlockCapnToGo(src.MiniBlocks().At(i), nil)
	}

	return dest
}

// TxBlockBodyGoToCapn is a helper function to copy fields from a TxBlockBody object to a TxBlockBodyCapn object
func TxBlockBodyGoToCapn(seg *capn.Segment, src *TxBlockBody) capnp.TxBlockBodyCapn {
	dest := capnp.AutoNewTxBlockBodyCapn(seg)
	dest.SetStateBlockBody(StateBlockBodyGoToCapn(seg, &src.StateBlockBody))

	// MiniBlocks -> MiniBlockCapn (go slice to capn list)
	if len(src.MiniBlocks) > 0 {
		typedList := capnp.NewMiniBlockCapnList(seg, len(src.MiniBlocks))
		plist := capn.PointerList(typedList)

		for i, elem := range src.MiniBlocks {
			plist.Set(i, capn.Object(MiniBlockGoToCapn(seg, &elem)))
		}
		dest.SetMiniBlocks(typedList)
	}

	return dest
}
