package ed25519

import (
	"crypto/ed25519"
)

func NewScalar(key ed25519.PrivateKey) *ed25519Scalar {
	return &ed25519Scalar{key}
}

func IsKeyValid(key ed25519.PrivateKey) error {
	return isKeyValid(key)
}
