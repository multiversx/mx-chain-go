package receipt

import (
	"io"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/receipt/capnp"
	capn "github.com/glycerine/go-capnproto"
)

// Receipt holds all the data needed for a transaction receipt
type Receipt struct {
	Value   *big.Int `capid:"0" json:"value"`
	SndAddr []byte   `capid:"1" json:"sender"`
	Data    []byte   `capid:"2" json:"data,omitempty"`
	TxHash  []byte   `capid:"3" json:"txHash"`
}

// Save saves the serialized data of a Receipt into a stream through Capnp protocol
func (rpt *Receipt) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	ReceiptGoToCapn(seg, rpt)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a Receipt object through Capnp protocol
func (rpt *Receipt) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}

	z := capnp.ReadRootReceiptCapn(capMsg)
	ReceiptCapnToGo(z, rpt)
	return nil
}

// ReceiptCapnToGo is a helper function to copy fields from a ReceiptCapn object to a Receipt object
func ReceiptCapnToGo(src capnp.ReceiptCapn, dest *Receipt) *Receipt {
	if dest == nil {
		dest = &Receipt{}
	}

	if dest.Value == nil {
		dest.Value = big.NewInt(0)
	}

	err := dest.Value.GobDecode(src.Value())
	if err != nil {
		return nil
	}

	dest.SndAddr = src.SndAddr()
	dest.Data = src.Data()
	dest.TxHash = src.TxHash()

	return dest
}

// ReceiptGoToCapn is a helper function to copy fields from a Receipt object to a ReceiptCapn object
func ReceiptGoToCapn(seg *capn.Segment, src *Receipt) capnp.ReceiptCapn {
	dest := capnp.AutoNewReceiptCapn(seg)

	value, _ := src.Value.GobEncode()
	dest.SetValue(value)
	dest.SetSndAddr(src.SndAddr)
	dest.SetData(src.Data)
	dest.SetTxHash(src.TxHash)

	return dest
}

// IsInterfaceNil verifies if underlying object is nil
func (rpt *Receipt) IsInterfaceNil() bool {
	return rpt == nil
}

// GetValue returns the value of the receipt
func (rpt *Receipt) GetValue() *big.Int {
	return rpt.Value
}

// GetNonce returns the nonce of the receipt
func (rpt *Receipt) GetNonce() uint64 {
	return 0
}

// GetData returns the data of the receipt
func (rpt *Receipt) GetData() []byte {
	return rpt.Data
}

// GetRecvAddress returns the receiver address from the receipt
func (rpt *Receipt) GetRecvAddress() []byte {
	return rpt.SndAddr
}

// GetSndAddress returns the sender address from the receipt
func (rpt *Receipt) GetSndAddress() []byte {
	return rpt.SndAddr
}

// GetGasLimit returns the gas limit of the receipt
func (rpt *Receipt) GetGasLimit() uint64 {
	return 0
}

// GetGasPrice returns the gas price of the receipt
func (rpt *Receipt) GetGasPrice() uint64 {
	return 0
}

// SetValue sets the value of the receipt
func (rpt *Receipt) SetValue(value *big.Int) {
	rpt.Value = value
}

// SetData sets the data of the receipt
func (rpt *Receipt) SetData(data []byte) {
	rpt.Data = data
}

// SetRecvAddress sets the receiver address of the receipt
func (rpt *Receipt) SetRecvAddress(_ []byte) {
}

// SetSndAddress sets the sender address of the receipt
func (rpt *Receipt) SetSndAddress(addr []byte) {
	rpt.SndAddr = addr
}
