//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. block.proto
package block

import (
	io "io"
	"io/ioutil"

	"github.com/ElrondNetwork/elrond-go/data"
)

// Body should be used when referring to the full list of mini blocks that forms a block body
type Body []*MiniBlock

// MiniBlockSlice should be used when referring to subset of mini blocks that is not
//  necessarily representing a full block body
type MiniBlockSlice []*MiniBlock

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

// ItemsInHeader gets the number of items(hashes) added in block header
func (h *Header) ItemsInHeader() uint32 {
	itemsInHeader := len(h.MiniBlockHeaders) + len(h.PeerChanges) + len(h.MetaBlockHashes)
	return uint32(itemsInHeader)
}

// ItemsInBody gets the number of items(hashes) added in block body
func (h *Header) ItemsInBody() uint32 {
	return h.TxCount
}

// ----- for compatibility only ----

func (h *Header) Save(w io.Writer) error {
	b, err := h.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (h *Header) Load(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return h.Unmarshal(b)
}

func (m *MiniBlock) Save(w io.Writer) error {
	b, err := m.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (m *MiniBlock) Load(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return m.Unmarshal(b)
}

func (m *MiniBlockHeader) Save(w io.Writer) error {
	b, err := m.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (m *MiniBlockHeader) Load(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return m.Unmarshal(b)
}

func (pc *PeerChange) Save(w io.Writer) error {
	b, err := pc.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (pc *PeerChange) Load(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return pc.Unmarshal(b)
}
