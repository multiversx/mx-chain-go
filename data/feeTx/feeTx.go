package feeTx

import (
	"io"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/feeTx/capnp"
	"github.com/glycerine/go-capnproto"
)

// FeeTx holds all the data needed for a value transfer
type FeeTx struct {
	Nonce   uint64   `capid:"0" json:"nonce"`
	Value   *big.Int `capid:"1" json:"value"`
	RcvAddr []byte   `capid:"2" json:"receiver"`
	ShardId uint32   `capid:"3" json:"ShardId"`
}

// Save saves the serialized data of a FeeTx into a stream through Capnp protocol
func (scr *FeeTx) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	FeeTxGoToCapn(seg, scr)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a FeeTx object through Capnp protocol
func (scr *FeeTx) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}

	z := capnp.ReadRootFeeTxCapn(capMsg)
	FeeTxCapnToGo(z, scr)
	return nil
}

// FeeTxCapnToGo is a helper function to copy fields from a FeeTxCapn object to a FeeTx object
func FeeTxCapnToGo(src capnp.FeeTxCapn, dest *FeeTx) *FeeTx {
	if dest == nil {
		dest = &FeeTx{}
	}

	if dest.Value == nil {
		dest.Value = big.NewInt(0)
	}

	dest.Nonce = src.Nonce()
	err := dest.Value.GobDecode(src.Value())

	if err != nil {
		return nil
	}

	dest.RcvAddr = src.RcvAddr()
	dest.ShardId = src.ShardId()

	return dest
}

// FeeTxGoToCapn is a helper function to copy fields from a FeeTx object to a FeeTxCapn object
func FeeTxGoToCapn(seg *capn.Segment, src *FeeTx) capnp.FeeTxCapn {
	dest := capnp.AutoNewFeeTxCapn(seg)

	value, _ := src.Value.GobEncode()
	dest.SetNonce(src.Nonce)
	dest.SetValue(value)
	dest.SetRcvAddr(src.RcvAddr)
	dest.SetShardId(src.ShardId)

	return dest
}

// IsInterfaceNil verifies if underlying object is nil
func (scr *FeeTx) IsInterfaceNil() bool {
	return scr == nil
}

// GetValue returns the value of the fee transaction
func (scr *FeeTx) GetValue() *big.Int {
	return scr.Value
}

// GetData returns the data of the fee transaction
func (scr *FeeTx) GetData() string {
	return ""
}

// GetRecvAddress returns the receiver address from the fee transaction
func (scr *FeeTx) GetRecvAddress() []byte {
	return scr.RcvAddr
}

// GetSndAddress returns the sender address from the fee transaction
func (scr *FeeTx) GetSndAddress() []byte {
	return nil
}

// SetValue sets the value of the fee transaction
func (scr *FeeTx) SetValue(value *big.Int) {
	scr.Value = value
}

// SetData sets the data of the fee transaction
func (scr *FeeTx) SetData(data string) {
}

// SetRecvAddress sets the receiver address of the fee transaction
func (scr *FeeTx) SetRecvAddress(addr []byte) {
	scr.RcvAddr = addr
}

// SetSndAddress sets the sender address of the fee transaction
func (scr *FeeTx) SetSndAddress(addr []byte) {
}
