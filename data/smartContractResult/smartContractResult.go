package smartContractResult

import (
	"io"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/smartContractResult/capnp"
	"github.com/glycerine/go-capnproto"
)

// SmartContractResult holds all the data needed for a value transfer
type SmartContractResult struct {
	Nonce   uint64   `capid:"0" json:"nonce"`
	Value   *big.Int `capid:"1" json:"value"`
	RcvAddr []byte   `capid:"2" json:"receiver"`
	SndAddr []byte   `capid:"3" json:"sender"`
	Code    []byte   `capid:"4" json:"code,omitempty"`
	Data    []byte   `capid:"5" json:"data,omitempty"`
	TxHash  []byte   `capid:"6" json:"txHash"`
}

// Save saves the serialized data of a SmartContractResult into a stream through Capnp protocol
func (scr *SmartContractResult) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	SmartContractResultGoToCapn(seg, scr)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a SmartContractResult object through Capnp protocol
func (scr *SmartContractResult) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}

	z := capnp.ReadRootSmartContractResultCapn(capMsg)
	SmartContractResultCapnToGo(z, scr)
	return nil
}

// SmartContractResultCapnToGo is a helper function to copy fields from a SmartContractResultCapn object to a SmartContractResult object
func SmartContractResultCapnToGo(src capnp.SmartContractResultCapn, dest *SmartContractResult) *SmartContractResult {
	if dest == nil {
		dest = &SmartContractResult{}
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
	dest.SndAddr = src.SndAddr()
	dest.Data = src.Data()
	dest.Code = src.Code()
	dest.TxHash = src.TxHash()

	return dest
}

// SmartContractResultGoToCapn is a helper function to copy fields from a SmartContractResult object to a SmartContractResultCapn object
func SmartContractResultGoToCapn(seg *capn.Segment, src *SmartContractResult) capnp.SmartContractResultCapn {
	dest := capnp.AutoNewSmartContractResultCapn(seg)

	value, _ := src.Value.GobEncode()
	dest.SetNonce(src.Nonce)
	dest.SetValue(value)
	dest.SetRcvAddr(src.RcvAddr)
	dest.SetSndAddr(src.SndAddr)
	dest.SetData(src.Data)
	dest.SetCode(src.Code)
	dest.SetTxHash(src.TxHash)

	return dest
}

// IsInterfaceNil verifies if underlying object is nil
func (scr *SmartContractResult) IsInterfaceNil() bool {
	return scr == nil
}

// GetValue returns the value of the smart contract result
func (scr *SmartContractResult) GetValue() *big.Int {
	return scr.Value
}

// GetData returns the data of the smart contract result
func (scr *SmartContractResult) GetData() []byte {
	return scr.Data
}

// GetRecvAddress returns the receiver address from the smart contract result
func (scr *SmartContractResult) GetRecvAddress() []byte {
	return scr.RcvAddr
}

// GetSndAddress returns the sender address from the smart contract result
func (scr *SmartContractResult) GetSndAddress() []byte {
	return scr.SndAddr
}

// SetValue sets the value of the smart contract result
func (scr *SmartContractResult) SetValue(value *big.Int) {
	scr.Value = value
}

// SetData sets the data of the smart contract result
func (scr *SmartContractResult) SetData(data []byte) {
	scr.Data = data
}

// SetRecvAddress sets the receiver address of the smart contract result
func (scr *SmartContractResult) SetRecvAddress(addr []byte) {
	scr.RcvAddr = addr
}

// SetSndAddress sets the sender address of the smart contract result
func (scr *SmartContractResult) SetSndAddress(addr []byte) {
	scr.SndAddr = addr
}
