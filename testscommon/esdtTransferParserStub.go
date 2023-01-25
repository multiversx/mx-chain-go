package testscommon

import (
	"errors"

	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// ESDTTransferParserStub -
type ESDTTransferParserStub struct {
	ParseESDTTransfersCalled func(sndAddr []byte, rcvAddr []byte, function string, args [][]byte) (*vmcommon.ParsedESDTTransfers, error)
}

// ParseESDTTransfers -
func (stub *ESDTTransferParserStub) ParseESDTTransfers(sndAddr []byte, rcvAddr []byte, function string, args [][]byte) (*vmcommon.ParsedESDTTransfers, error) {
	if stub.ParseESDTTransfersCalled != nil {
		return stub.ParseESDTTransfersCalled(sndAddr, rcvAddr, function, args)
	}

	return nil, errors.New("not implemented")
}

// IsInterfaceNil -
func (stub *ESDTTransferParserStub) IsInterfaceNil() bool {
	return stub == nil
}
