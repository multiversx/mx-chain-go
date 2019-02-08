package block

import (
	"fmt"
	"io"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block/capnp"
	"github.com/glycerine/go-capnproto"
)

// PeerAction type represents the possible events that a node can trigger for the metachain to notarize
type PeerAction int

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
	PublicKey []byte `capid:"0"`
	Action PeerAction `capid:"1"`
	TimeStamp uint64 `capid:"2"`
	Value *big.Int `capid:"3"`
}

// ShardData holds the block information sent by the shards to the metachain
type ShardData struct {
	ShardId uint32 `capid:"0"`
	HeaderHashes [][]byte `capid:"1"`
}

// Proof is a structure that holds inclusion and exclusion proofs for a set of transactions
type Proof struct {
	InclusionProof []byte `capid:"0"`
	ExclusionProof []byte `capid:"1"`
}

// MetaBlock holds the data that will be saved to the metachain each round
type MetaBlock struct {
	Nonce uint64 `capid:"0"`
	Epoch uint32 `capid:"1"`
	Round uint32 `capid:"2"`
	ShardInfo []ShardData `capid:"3"`
	PeerInfo []PeerData `capid:"4"`
	Proof `capid:"5"`
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

// Save saves the serialized data of a Proof into a stream through Capnp protocol
func (p *Proof) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	ProofGoToCapn(seg, p)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a Proof object through Capnp protocol
func (p *Proof) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootProofCapn(capMsg)
	ProofCapnToGo(z, p)
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

// ShardDataGoToCapn is a helper function to copy fields from a ShardData object to a ShardDataCapn object
func ShardDataGoToCapn(seg *capn.Segment, src *ShardData) capnp.ShardDataCapn {
	dest := capnp.AutoNewShardDataCapn(seg)

	dest.SetShardId(src.ShardId)
	mylist1 := seg.NewDataList(len(src.HeaderHashes))
	for i := range src.HeaderHashes {
		mylist1.Set(i, src.HeaderHashes[i])
	}
	dest.SetHeaderHashes(mylist1)

	return dest
}

// ShardDataCapnToGo is a helper function to copy fields from a ShardDataCapn object to a ShardData object
func ShardDataCapnToGo(src capnp.ShardDataCapn, dest *ShardData) *ShardData {
	if dest == nil {
		dest = &ShardData{}
	}
	dest.ShardId = src.ShardId()
	n := src.HeaderHashes().Len()
	dest.HeaderHashes = make([][]byte, n)
	for i := 0; i < n; i++ {
		dest.HeaderHashes[i] = src.HeaderHashes().At(i)
	}

	return dest
}

// ProofGoToCapn is a helper function to copy fields from a Proof object to a ProofCapn object
func ProofGoToCapn(seg *capn.Segment, src *Proof) capnp.ProofCapn {
	dest := capnp.AutoNewProofCapn(seg)

	dest.SetInclusionProof(src.InclusionProof)
	dest.SetExclusionProof(src.ExclusionProof)

	return dest
}

// ProofCapnToGo is a helper function to copy fields from a ProofCapn object to a Proof object
func ProofCapnToGo(src capnp.ProofCapn, dest *Proof) *Proof {
	if dest == nil {
		dest = &Proof{}
	}
	dest.InclusionProof = src.InclusionProof()
	dest.ExclusionProof = src.ExclusionProof()

	return dest
}

// MetaBlockGoToCapn is a helper function to copy fields from a MetaBlock object to a MetaBlockCapn object
func MetaBlockGoToCapn(seg *capn.Segment, src *MetaBlock) capnp.MetaBlockCapn {
	dest := capnp.AutoNewMetaBlockCapn(seg)

	dest.SetNonce(src.Nonce)
	dest.SetEpoch(src.Epoch)
	dest.SetRound(src.Round)

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

	dest.SetProof(ProofGoToCapn(seg, &src.Proof))

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
	dest.Proof = *ProofCapnToGo(src.Proof(), nil)

	return dest
}