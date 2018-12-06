package interceptors

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type transactionInterceptor struct {
	intercept  *interceptor
	blockchain *blockchain.BlockChain
}

func NewTransactionInterceptor(messenger p2p.Messenger, blockchain *blockchain.BlockChain) (*transactionInterceptor, error) {
	if messenger == nil {
		return nil, ErrNilMessenger
	}

	if blockchain == nil {
		return nil, ErrNilInterceptor
	}

	intercept, errD := NewInterceptor("tx", messenger, &transaction.Transaction{})

	txIntercept := transactionInterceptor{}

}
