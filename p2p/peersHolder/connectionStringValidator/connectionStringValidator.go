package connectionStringValidator

import (
	"net"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

type connectionStringValidator struct {
}

// NewConnectionStringValidator returns a new connection string validator
func NewConnectionStringValidator() *connectionStringValidator {
	return &connectionStringValidator{}
}

// IsValid checks either a connection string is a valid ip or peer id
func (csv *connectionStringValidator) IsValid(connStr string) bool {
	return csv.isValidIP(connStr) || csv.isValidPeerID(connStr)
}

func (csv *connectionStringValidator) isValidIP(connStr string) bool {
	return net.ParseIP(connStr) != nil
}

func (csv *connectionStringValidator) isValidPeerID(connStr string) bool {
	_, err := core.NewPeerID(connStr)
	return err == nil
}
