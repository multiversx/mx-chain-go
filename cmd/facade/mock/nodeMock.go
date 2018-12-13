package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"math/big"
)

type NodeMock struct {
	AddressCalled                   func() (string, error)
	StartCalled                     func() error
	StopCalled                      func() error
	IsRunningCalled                 func() bool
	ConnectToInitialAddressesCalled func() error
	ConnectToAddressesCalled        func([]string) error
	StartConsensusCalled            func() error
	GetBalanceCalled                func(address string) (*big.Int, error)
	GenerateTransactionCalled       func(sender string, receiver string, amount big.Int, code string) (string, error)
	GetTransactionCalled            func(hash string) (*transaction.Transaction, error)
}

func (nm *NodeMock) Address() (string, error) {
	return nm.AddressCalled()
}

func (nm *NodeMock) Start() error {
	return nm.StartCalled()
}

func (nm *NodeMock) Stop() error {
	return nm.StopCalled()
}

func (nm *NodeMock) IsRunning() bool {
	return nm.IsRunningCalled()
}

func (nm *NodeMock) ConnectToInitialAddresses() error {
	return nm.ConnectToInitialAddressesCalled()
}

func (nm *NodeMock) ConnectToAddresses(addresses []string) error {
	return nm.ConnectToAddressesCalled(addresses)
}

func (nm *NodeMock) StartConsensus() error {
	return nm.StartConsensusCalled()
}

func (nm *NodeMock) GetBalance(address string) (*big.Int, error) {
	return nm.GetBalanceCalled(address)
}

func (nm *NodeMock) GenerateTransaction(sender string, receiver string, amount big.Int, code string) (string, error) {
	return nm.GenerateTransactionCalled(sender, receiver, amount, code)
}

func (nm *NodeMock) GetTransaction(hash string) (*transaction.Transaction, error) {
	return nm.GetTransactionCalled(hash)
}
