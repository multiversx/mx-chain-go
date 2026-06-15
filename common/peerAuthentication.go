package common

// MaxPeerAuthenticationPublicKeyIdentifierLen is the maximum number of public key bytes used as
// a peer authentication request/whitelist identifier.
const MaxPeerAuthenticationPublicKeyIdentifierLen = 32

// PeerAuthenticationPublicKeyIdentifier returns the public key prefix used as request/whitelist identifier.
func PeerAuthenticationPublicKeyIdentifier(publicKey []byte) []byte {
	if len(publicKey) > MaxPeerAuthenticationPublicKeyIdentifierLen {
		publicKey = publicKey[:MaxPeerAuthenticationPublicKeyIdentifierLen]
	}

	return append([]byte(nil), publicKey...)
}
