//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. metaBlock.proto
package block

import (
	"bytes"
	"encoding/hex"
	"fmt"
	io "io"
	"io/ioutil"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

// don't break the interface
var _ = data.HeaderHandler(&MetaBlock{})

// GetShardID returns the metachain shard id
func (m *MetaBlock) GetShardID() uint32 {
	return core.MetachainShardId
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

// SetShardID sets header shard ID
func (m *MetaBlock) SetShardID(_ uint32) {
}

// GetMiniBlockHeadersWithDst as a map of hashes and sender IDs
func (m *MetaBlock) GetMiniBlockHeadersWithDst(destId uint32) map[string]uint32 {
	hashDst := make(map[string]uint32)
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

// ----- for compatibility only ----

func (pd *PeerData) Save(w io.Writer) error {
	b, err := pd.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (pd *PeerData) Load(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return pd.Unmarshal(b)
}

func (sd *ShardData) Save(w io.Writer) error {
	b, err := sd.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (sd *ShardData) Load(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return sd.Unmarshal(b)
}

func (mb *MetaBlock) Save(w io.Writer) error {
	b, err := mb.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (mb *MetaBlock) Load(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return mb.Unmarshal(b)
}
