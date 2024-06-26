package testscommon

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-crypto-go"
)

type keysHandlerSingleSignerMock struct {
	privateKey crypto.PrivateKey
	publicKey  crypto.PublicKey
	pkBytes    []byte
	pid        core.PeerID
}

// NewKeysHandlerSingleSignerMock -
func NewKeysHandlerSingleSignerMock(
	privateKey crypto.PrivateKey,
	pid core.PeerID,
) *keysHandlerSingleSignerMock {
	pk := privateKey.GeneratePublic()
	pkBytes, _ := pk.ToByteArray()

	return &keysHandlerSingleSignerMock{
		privateKey: privateKey,
		publicKey:  pk,
		pkBytes:    pkBytes,
		pid:        pid,
	}
}

// GetHandledPrivateKey -
func (mock *keysHandlerSingleSignerMock) GetHandledPrivateKey(_ []byte) crypto.PrivateKey {
	return mock.privateKey
}

// GetP2PIdentity -
func (mock *keysHandlerSingleSignerMock) GetP2PIdentity(_ []byte) ([]byte, core.PeerID, error) {
	return make([]byte, 0), "", nil
}

// IsKeyManagedByCurrentNode -
func (mock *keysHandlerSingleSignerMock) IsKeyManagedByCurrentNode(_ []byte) bool {
	return false
}

// IncrementRoundsWithoutReceivedMessages -
func (mock *keysHandlerSingleSignerMock) IncrementRoundsWithoutReceivedMessages(_ []byte) {
}

// GetAssociatedPid -
func (mock *keysHandlerSingleSignerMock) GetAssociatedPid(pkBytes []byte) core.PeerID {
	if bytes.Equal(mock.pkBytes, pkBytes) {
		return mock.pid
	}

	return ""
}

// IsOriginalPublicKeyOfTheNode -
func (mock *keysHandlerSingleSignerMock) IsOriginalPublicKeyOfTheNode(pkBytes []byte) bool {
	return bytes.Equal(mock.pkBytes, pkBytes)
}

// ResetRoundsWithoutReceivedMessages -
func (mock *keysHandlerSingleSignerMock) ResetRoundsWithoutReceivedMessages(_ []byte, _ core.PeerID) {
}

// GetRedundancyStepInReason -
func (mock *keysHandlerSingleSignerMock) GetRedundancyStepInReason() string {
	return ""
}

// IsInterfaceNil -
func (mock *keysHandlerSingleSignerMock) IsInterfaceNil() bool {
	return mock == nil
}
