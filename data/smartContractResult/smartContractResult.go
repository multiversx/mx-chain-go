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
