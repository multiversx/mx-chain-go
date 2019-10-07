package rewardTx

import (
	"io"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/rewardTx/capnp"
	capn "github.com/glycerine/go-capnproto"
)

// RewardTx holds the data for a reward transaction
type RewardTx struct {
	Round   uint64   `capid:"1" json:"round"`
	Epoch   uint32   `capid:"2" json:"epoch"`
	Value   *big.Int `capid:"3" json:"value"`
	RcvAddr []byte   `capid:"4" json:"receiver"`
	ShardId uint32   `capid:"5" json:"shardId"`
}

// Save saves the serialized data of a RewardTx into a stream through Capnp protocol
func (scr *RewardTx) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	RewardTxGoToCapn(seg, scr)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a RewardTx object through Capnp protocol
func (scr *RewardTx) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}

	z := capnp.ReadRootRewardTxCapn(capMsg)
	RewardTxCapnToGo(z, scr)
	return nil
}

// RewardTxCapnToGo is a helper function to copy fields from a RewardTxCapn object to a RewardTx object
func RewardTxCapnToGo(src capnp.RewardTxCapn, dest *RewardTx) *RewardTx {
	if dest == nil {
		dest = &RewardTx{}
	}

	if dest.Value == nil {
		dest.Value = big.NewInt(0)
	}

	dest.Epoch = src.Epoch()
	dest.Round = src.Round()
	err := dest.Value.GobDecode(src.Value())

	if err != nil {
		return nil
	}

	dest.RcvAddr = src.RcvAddr()
	dest.ShardId = src.ShardId()

	return dest
}

// RewardTxGoToCapn is a helper function to copy fields from a RewardTx object to a RewardTxCapn object
func RewardTxGoToCapn(seg *capn.Segment, src *RewardTx) capnp.RewardTxCapn {
	dest := capnp.AutoNewRewardTxCapn(seg)

	value, _ := src.Value.GobEncode()
	dest.SetEpoch(src.Epoch)
	dest.SetRound(src.Round)
	dest.SetValue(value)
	dest.SetRcvAddr(src.RcvAddr)
	dest.SetShardId(src.ShardId)

	return dest
}

// IsInterfaceNil verifies if underlying object is nil
func (scr *RewardTx) IsInterfaceNil() bool {
	return scr == nil
}

// GetValue returns the value of the reward transaction
func (scr *RewardTx) GetValue() *big.Int {
	return scr.Value
}

// GetNonce returns 0 as reward transactions do not have a nonce
func (scr *RewardTx) GetNonce() uint64 {
	return 0
}

// GetData returns the data of the reward transaction
func (scr *RewardTx) GetData() string {
	return ""
}

// GetRecvAddress returns the receiver address from the reward transaction
func (scr *RewardTx) GetRecvAddress() []byte {
	return scr.RcvAddr
}

// GetSndAddress returns the sender address from the reward transaction
func (scr *RewardTx) GetSndAddress() []byte {
	return nil
}

// SetValue sets the value of the reward transaction
func (scr *RewardTx) SetValue(value *big.Int) {
	scr.Value = value
}

// SetData sets the data of the reward transaction
func (scr *RewardTx) SetData(data string) {
}

// SetRecvAddress sets the receiver address of the reward transaction
func (scr *RewardTx) SetRecvAddress(addr []byte) {
	scr.RcvAddr = addr
}

// SetSndAddress sets the sender address of the reward transaction
func (scr *RewardTx) SetSndAddress(addr []byte) {
}
