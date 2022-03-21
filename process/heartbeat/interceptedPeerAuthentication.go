package heartbeat

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ArgInterceptedPeerAuthentication is the argument used in the intercepted peer authentication constructor
type ArgInterceptedPeerAuthentication struct {
	ArgBaseInterceptedHeartbeat
	NodesCoordinator     NodesCoordinator
	SignaturesHandler    SignaturesHandler
	PeerSignatureHandler crypto.PeerSignatureHandler
	ExpiryTimespanInSec  int64
}

// interceptedPeerAuthentication is a wrapper over PeerAuthentication
type interceptedPeerAuthentication struct {
	peerAuthentication   heartbeat.PeerAuthentication
	payload              heartbeat.Payload
	peerId               core.PeerID
	nodesCoordinator     NodesCoordinator
	signaturesHandler    SignaturesHandler
	peerSignatureHandler crypto.PeerSignatureHandler
	expiryTimespanInSec  int64
}

// NewInterceptedPeerAuthentication tries to create a new intercepted peer authentication instance
func NewInterceptedPeerAuthentication(arg ArgInterceptedPeerAuthentication) (*interceptedPeerAuthentication, error) {
	err := checkArg(arg)
	if err != nil {
		return nil, err
	}

	peerAuthentication, payload, err := createPeerAuthentication(arg.Marshalizer, arg.DataBuff)
	if err != nil {
		return nil, err
	}

	intercepted := &interceptedPeerAuthentication{
		peerAuthentication:   *peerAuthentication,
		payload:              *payload,
		nodesCoordinator:     arg.NodesCoordinator,
		signaturesHandler:    arg.SignaturesHandler,
		peerSignatureHandler: arg.PeerSignatureHandler,
		expiryTimespanInSec:  arg.ExpiryTimespanInSec,
	}
	intercepted.peerId = core.PeerID(intercepted.peerAuthentication.Pid)

	return intercepted, nil
}

func checkArg(arg ArgInterceptedPeerAuthentication) error {
	err := checkBaseArg(arg.ArgBaseInterceptedHeartbeat)
	if err != nil {
		return err
	}
	if check.IfNil(arg.NodesCoordinator) {
		return process.ErrNilNodesCoordinator
	}
	if check.IfNil(arg.SignaturesHandler) {
		return process.ErrNilSignaturesHandler
	}
	if arg.ExpiryTimespanInSec < minDurationInSec {
		return process.ErrInvalidExpiryTimespan
	}
	if check.IfNil(arg.PeerSignatureHandler) {
		return process.ErrNilPeerSignatureHandler
	}
	return nil
}

func createPeerAuthentication(marshalizer marshal.Marshalizer, buff []byte) (*heartbeat.PeerAuthentication, *heartbeat.Payload, error) {
	peerAuthentication := &heartbeat.PeerAuthentication{}
	err := marshalizer.Unmarshal(peerAuthentication, buff)
	if err != nil {
		return nil, nil, err
	}
	payload := &heartbeat.Payload{}
	err = marshalizer.Unmarshal(payload, peerAuthentication.Payload)
	if err != nil {
		return nil, nil, err
	}

	return peerAuthentication, payload, nil
}

// CheckValidity checks the validity of the received peer authentication. This call won't trigger the signature validation.
func (ipa *interceptedPeerAuthentication) CheckValidity() error {
	// Verify properties len
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

	// Verify validator
	_, _, err = ipa.nodesCoordinator.GetValidatorWithPublicKey(ipa.peerAuthentication.Pubkey)
	if err != nil {
		return err
	}

	// Verify payload signature
	err = ipa.signaturesHandler.Verify(ipa.peerAuthentication.Payload, ipa.peerId, ipa.peerAuthentication.PayloadSignature)
	if err != nil {
		return err
	}

	// Verify payload
	err = ipa.verifyPayload()
	if err != nil {
		return err
	}

	// Verify message bls signature
	err = ipa.peerSignatureHandler.VerifyPeerSignature(ipa.peerAuthentication.Pubkey, ipa.peerId, ipa.peerAuthentication.Signature)
	if err != nil {
		return err
	}

	return nil
}

// IsForCurrentShard always returns true
func (ipa *interceptedPeerAuthentication) IsForCurrentShard() bool {
	return true
}

// Hash always returns an empty string
func (ipa *interceptedPeerAuthentication) Hash() []byte {
	return []byte("")
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

// Message returns the peer authentication message
func (ipa *interceptedPeerAuthentication) Message() interface{} {
	return ipa.peerAuthentication
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

func (ipa *interceptedPeerAuthentication) verifyPayload() error {
	currentTimeStamp := time.Now().Unix()
	messageTimeStamp := ipa.payload.Timestamp
	minTimestampAllowed := currentTimeStamp - ipa.expiryTimespanInSec
	maxTimestampAllowed := currentTimeStamp + payloadExpiryThresholdInSec
	if messageTimeStamp < minTimestampAllowed || messageTimeStamp > maxTimestampAllowed {
		return process.ErrMessageExpired
	}
	// TODO: check for payload hardfork

	return nil
}

// SizeInBytes returns the size in bytes held by this instance
func (ipa *interceptedPeerAuthentication) SizeInBytes() int {
	return len(ipa.peerAuthentication.Pubkey) +
		len(ipa.peerAuthentication.Signature) +
		len(ipa.peerAuthentication.Pid) +
		len(ipa.peerAuthentication.Payload) +
		len(ipa.peerAuthentication.PayloadSignature)
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
