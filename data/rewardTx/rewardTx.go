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
func (rtx *RewardTx) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	RewardTxGoToCapn(seg, rtx)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a RewardTx object through Capnp protocol
func (rtx *RewardTx) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}

	z := capnp.ReadRootRewardTxCapn(capMsg)
	RewardTxCapnToGo(z, rtx)
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
func (rtx *RewardTx) IsInterfaceNil() bool {
	return rtx == nil
}

// GetValue returns the value of the reward transaction
func (rtx *RewardTx) GetValue() *big.Int {
	return rtx.Value
}

// GetNonce returns 0 as reward transactions do not have a nonce
func (scr *RewardTx) GetNonce() uint64 {
	return 0
}

// GetData returns the data of the reward transaction
func (rtx *RewardTx) GetData() string {
	return ""
}

// GetRecvAddress returns the receiver address from the reward transaction
func (rtx *RewardTx) GetRecvAddress() []byte {
	return rtx.RcvAddr
}

// GetSndAddress returns the sender address from the reward transaction
func (rtx *RewardTx) GetSndAddress() []byte {
	return nil
}

// GetGasLimit returns the gas limit of the smart reward transaction
func (scr *RewardTx) GetGasLimit() uint64 {
	return 0
}

// GetGasPrice returns the gas price of the smart reward transaction
func (scr *RewardTx) GetGasPrice() uint64 {
	return 0
}

// SetValue sets the value of the reward transaction
func (rtx *RewardTx) SetValue(value *big.Int) {
	rtx.Value = value
}

// SetData sets the data of the reward transaction
func (rtx *RewardTx) SetData(data string) {
}

// SetRecvAddress sets the receiver address of the reward transaction
func (rtx *RewardTx) SetRecvAddress(addr []byte) {
	rtx.RcvAddr = addr
}

// SetSndAddress sets the sender address of the reward transaction
func (rtx *RewardTx) SetSndAddress(addr []byte) {
}
