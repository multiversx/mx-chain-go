package data

import (
	"io"

	"github.com/ElrondNetwork/elrond-go/p2p/data/capnp"
	protobuf "github.com/ElrondNetwork/elrond-go/p2p/data/proto"
	capn "github.com/glycerine/go-capnproto"
)

// AuthMessage represents the authentication message used in the handshake process of 2 peers
type AuthMessage struct {
	protobuf.AuthMessagePb
}

// Save saves the serialized data of a log line message into a stream through Capnp protocol
func (am *AuthMessage) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	authMessageGoToCapn(seg, am)
	_, err := seg.WriteTo(w)

	return err
}

// Load loads the data from the stream into a log line message object through Capnp protocol
func (am *AuthMessage) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootAuthMessageCapn(capMsg)
	authMessageCapnToGo(z, am)

	return nil
}

func authMessageGoToCapn(seg *capn.Segment, src *AuthMessage) capnp.AuthMessageCapn {
	dest := capnp.AutoNewAuthMessageCapn(seg)

	dest.SetMessage(src.Message)
	dest.SetSig(src.Sig)
	dest.SetPubkey(src.Pubkey)
	dest.SetTimestamp(src.Timestamp)

	return dest
}

func authMessageCapnToGo(src capnp.AuthMessageCapn, dest *AuthMessage) {
	dest.Message = src.Message()
	dest.Sig = src.Sig()
	dest.Pubkey = src.Pubkey()
	dest.Timestamp = src.Timestamp()

	return
}

// IsInterfaceNil returns true if there is no value under the interface
func (am *AuthMessage) IsInterfaceNil() bool {
	return am == nil
}
