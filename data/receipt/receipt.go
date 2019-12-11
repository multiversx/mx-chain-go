package receipt

import (
	"io"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/receipt/capnp"
	capn "github.com/glycerine/go-capnproto"
)

// Receipt holds all the data needed for a transaction receipt
type Receipt struct {
	Value   *big.Int `capid:"1" json:"value"`
	SndAddr []byte   `capid:"2" json:"sender"`
	Data    string   `capid:"3" json:"data,omitempty"`
	TxHash  []byte   `capid:"4" json:"txHash"`
}

// Save saves the serialized data of a Receipt into a stream through Capnp protocol
func (scr *Receipt) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	ReceiptGoToCapn(seg, scr)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a Receipt object through Capnp protocol
func (scr *Receipt) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}

	z := capnp.ReadRootReceiptCapn(capMsg)
	ReceiptCapnToGo(z, scr)
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
	dest.Data = string(src.Data())
	dest.TxHash = src.TxHash()

	return dest
}

// ReceiptGoToCapn is a helper function to copy fields from a Receipt object to a ReceiptCapn object
func ReceiptGoToCapn(seg *capn.Segment, src *Receipt) capnp.ReceiptCapn {
	dest := capnp.AutoNewReceiptCapn(seg)

	value, _ := src.Value.GobEncode()
	dest.SetValue(value)
	dest.SetSndAddr(src.SndAddr)
	dest.SetData([]byte(src.Data))
	dest.SetTxHash(src.TxHash)

	return dest
}

// IsInterfaceNil verifies if underlying object is nil
func (scr *Receipt) IsInterfaceNil() bool {
	return scr == nil
}

// GetValue returns the value of the smart contract result
func (scr *Receipt) GetValue() *big.Int {
	return scr.Value
}

// GetNonce returns the nonce of the smart contract result
func (scr *Receipt) GetNonce() uint64 {
	return 0
}

// GetData returns the data of the smart contract result
func (scr *Receipt) GetData() string {
	return scr.Data
}

// GetRecvAddress returns the receiver address from the smart contract result
func (scr *Receipt) GetRecvAddress() []byte {
	return scr.SndAddr
}

// GetSndAddress returns the sender address from the smart contract result
func (scr *Receipt) GetSndAddress() []byte {
	return scr.SndAddr
}

// GetGasLimit returns the gas limit of the smart contract result
func (scr *Receipt) GetGasLimit() uint64 {
	return 0
}

// GetGasPrice returns the gas price of the smart contract result
func (scr *Receipt) GetGasPrice() uint64 {
	return 0
}

// SetValue sets the value of the smart contract result
func (scr *Receipt) SetValue(value *big.Int) {
	scr.Value = value
}

// SetData sets the data of the smart contract result
func (scr *Receipt) SetData(data string) {
	scr.Data = data
}

// SetRecvAddress sets the receiver address of the smart contract result
func (scr *Receipt) SetRecvAddress(_ []byte) {
}

// SetSndAddress sets the sender address of the smart contract result
func (scr *Receipt) SetSndAddress(addr []byte) {
	scr.SndAddr = addr
}
