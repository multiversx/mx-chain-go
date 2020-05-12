package factory

import (
	"context"
	"sync"
)

// CryptoComponentsHandlerArgs holds the arguments required to create a crypto components handler
type CryptoComponentsHandlerArgs CryptoComponentsFactoryArgs

// CryptoComponentsHandler creates the crypto components handler that can create, close and access the crypto components
type CryptoComponentsHandler struct {
	cryptoComponentsFactory *cryptoComponentsFactory
	*CryptoComponents
	mutCryptoComponents sync.RWMutex
	cancelFunc          func()
}

// NewCryptoComponentsHandler creates a new Crypto components handler
func NewCryptoComponentsHandler(args CryptoComponentsHandlerArgs) (*CryptoComponentsHandler, error) {
	cryptoComponentsFactory, err := NewCryptoComponentsFactory(CryptoComponentsFactoryArgs(args))
	if err != nil {
		return nil, err
	}

	return &CryptoComponentsHandler{
		CryptoComponents:        nil,
		cryptoComponentsFactory: cryptoComponentsFactory,
	}, nil
}

// Close closes the managed crypto components
func (cch *CryptoComponentsHandler) Close() error {
	cch.mutCryptoComponents.Lock()
	cch.cancelFunc()
	cch.cancelFunc = nil
	cch.CryptoComponents = nil
	cch.mutCryptoComponents.Unlock()

	return nil
}

// Create creates the crypto components
func (cch *CryptoComponentsHandler) Create() error {
	cryptoComponents, err := cch.cryptoComponentsFactory.Create()
	if err != nil {
		return err
	}

	cch.mutCryptoComponents.Lock()
	cch.CryptoComponents = cryptoComponents
	_, cch.cancelFunc = context.WithCancel(context.Background())
	cch.mutCryptoComponents.Unlock()

	return nil
}
