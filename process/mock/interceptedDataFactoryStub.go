package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

// InterceptedDataFactoryStub -
type InterceptedDataFactoryStub struct {
	CreateCalled func(buff []byte) (process.InterceptedData, error)
}

// Create -
func (idfs *InterceptedDataFactoryStub) Create(buff []byte, messageOriginator core.PeerID, broadcastMethod p2p.BroadcastMethod) (process.InterceptedData, error) {
	return idfs.CreateCalled(buff)
}

// IsInterfaceNil -
func (idfs *InterceptedDataFactoryStub) IsInterfaceNil() bool {
	return idfs == nil
}
