package heartbeat

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ArgInterceptedPeerAuthentication is the argument used in the intercepted peer authentication constructor
type ArgInterceptedPeerAuthentication struct {
	argBaseInterceptedHeartbeat
}

// interceptedPeerAuthentication is a wrapper over PeerAuthentication
type interceptedPeerAuthentication struct {
	peerAuthentication heartbeat.PeerAuthentication
	peerId             core.PeerID
	hash               []byte
}

// NewInterceptedPeerAuthentication tries to create a new intercepted peer authentication instance
func NewInterceptedPeerAuthentication(arg ArgInterceptedPeerAuthentication) (*interceptedPeerAuthentication, error) {
	err := checkBaseArg(arg.argBaseInterceptedHeartbeat)
	if err != nil {
		return nil, err
	}

	peerAuthentication, err := createPeerAuthentication(arg.Marshalizer, arg.DataBuff)
	if err != nil {
		return nil, err
	}

	intercepted := &interceptedPeerAuthentication{
		peerAuthentication: *peerAuthentication,
	}

	intercepted.processFields(arg.Hasher, arg.DataBuff)

	return intercepted, nil
}

func createPeerAuthentication(marshalizer marshal.Marshalizer, buff []byte) (*heartbeat.PeerAuthentication, error) {
	peerAuthentication := &heartbeat.PeerAuthentication{}
	err := marshalizer.Unmarshal(peerAuthentication, buff)
	if err != nil {
		return nil, err
	}

	return peerAuthentication, nil
}

func (ipa *interceptedPeerAuthentication) processFields(hasher hashing.Hasher, buff []byte) {
	ipa.hash = hasher.Compute(string(buff))
	ipa.peerId = core.PeerID(ipa.peerAuthentication.Pid)
}

// CheckValidity will check the validity of the received peer authentication. This call won't trigger the signature validation.
func (ipa *interceptedPeerAuthentication) CheckValidity() error {
	err := verifyPropertyLen(publicKeyProperty, ipa.peerAuthentication.Pubkey)
	if err != nil {
		return err
	}
	err = verifyPropertyLen(signatureProperty, ipa.peerAuthentication.Signature)
	if err != nil {
		return err
	}
	err = verifyPropertyLen(peerIdProperty, ipa.peerId.Bytes())
	if err != nil {
		return err
	}
	err = verifyPropertyLen(payloadProperty, ipa.peerAuthentication.Payload)
	if err != nil {
		return err
	}
	err = verifyPropertyLen(payloadSignatureProperty, ipa.peerAuthentication.PayloadSignature)
	if err != nil {
		return err
	}

	return nil
}

// IsForCurrentShard always returns true
func (ipa *interceptedPeerAuthentication) IsForCurrentShard() bool {
	return true
}

// Hash returns the hash of this intercepted peer authentication
func (ipa *interceptedPeerAuthentication) Hash() []byte {
	return ipa.hash
}

// Type returns the type of this intercepted data
func (ipa *interceptedPeerAuthentication) Type() string {
	return interceptedPeerAuthenticationType
}

// Identifiers returns the identifiers used in requests
func (ipa *interceptedPeerAuthentication) Identifiers() [][]byte {
	return [][]byte{ipa.peerAuthentication.Pubkey, ipa.peerAuthentication.Pid}
}

// PeerID returns the peer ID
func (ipa *interceptedPeerAuthentication) PeerID() core.PeerID {
	return core.PeerID(ipa.peerAuthentication.Pid)
}

// Signature returns the signature for the peer authentication
func (ipa *interceptedPeerAuthentication) Signature() []byte {
	return ipa.peerAuthentication.Signature
}

// Payload returns the payload data
func (ipa *interceptedPeerAuthentication) Payload() []byte {
	return ipa.peerAuthentication.Payload
}

// PayloadSignature returns the signature done on the payload
func (ipa *interceptedPeerAuthentication) PayloadSignature() []byte {
	return ipa.peerAuthentication.PayloadSignature
}

// String returns the most important fields as string
func (ipa *interceptedPeerAuthentication) String() string {
	return fmt.Sprintf("pk=%s, pid=%s, sig=%s, payload=%s, payloadSig=%s",
		logger.DisplayByteSlice(ipa.peerAuthentication.Pubkey),
		ipa.peerId.Pretty(),
		logger.DisplayByteSlice(ipa.peerAuthentication.Signature),
		logger.DisplayByteSlice(ipa.peerAuthentication.Payload),
		logger.DisplayByteSlice(ipa.peerAuthentication.PayloadSignature),
	)
}

// verifyPropertyLen returns an error if the provided value is longer than accepted by the network
func verifyPropertyLen(property string, value []byte) error {
	if len(value) > maxSizeInBytes {
		return fmt.Errorf("%w for %s", process.ErrPropertyTooLong, property)
	}
	if len(value) < minSizeInBytes {
		return fmt.Errorf("%w for %s", process.ErrPropertyTooShort, property)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ipa *interceptedPeerAuthentication) IsInterfaceNil() bool {
	return ipa == nil
}
