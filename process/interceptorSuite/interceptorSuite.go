package interceptorSuite

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
)

// InterceptorSuite is used to hold all the interceptors defined
type InterceptorSuite struct {
	txInterceptor         *transaction.TransactionInterceptor
	headerInterceptor     *block.HeaderInterceptor
	blockbodyInterceptors []*block.GenericBlockBodyInterceptor

	messenger p2p.Messenger
	hasher    hashing.Hasher
}

// NewInterceptorSuite creates a new suite of interceptors
func NewInterceptorSuite(messenger p2p.Messenger, hasher hashing.Hasher) (*InterceptorSuite, error) {
	if messenger == nil {
		return nil, process.ErrNilMessenger
	}

	if hasher == nil {
		return nil, process.ErrNilHasher
	}

	return &InterceptorSuite{
		messenger:             messenger,
		hasher:                hasher,
		blockbodyInterceptors: make([]*block.GenericBlockBodyInterceptor, 0),
	}, nil
}

// SetTransactionInterceptor sets the transaction interceptor
func (is *InterceptorSuite) SetTransactionInterceptor(txPool *dataPool.DataPool, addrConv state.AddressConverter) error {
	ti, err := transaction.NewTransactionInterceptor(is.messenger, txPool, addrConv, is.hasher)

	if err != nil {
		return err
	}

	is.txInterceptor = ti
	return nil
}

// SetHeaderInterceptor sets the header interceptor
func (is *InterceptorSuite) SetHeaderInterceptor(headerPool *dataPool.DataPool) error {
	hi, err := block.NewHeaderInterceptor(is.messenger, headerPool, is.hasher)

	if err != nil {
		return err
	}

	is.headerInterceptor = hi
	return nil
}

// AppendBlockBodyInterceptor appends a block body interceptor to the internal block bodies list
func (is *InterceptorSuite) AppendBlockBodyInterceptor(name string, bbp *dataPool.DataPool, templateObj process.BlockBodyInterceptorAdapter) error {
	bbi, err := block.NewGenericBlockBodyInterceptor(name, is.messenger, bbp, is.hasher, templateObj)

	if err != nil {
		return err
	}

	is.blockbodyInterceptors = append(is.blockbodyInterceptors, bbi)
	return nil
}

// MakeDefaultInterceptors makes the default list
func (is *InterceptorSuite) MakeDefaultInterceptors(
	txPool *dataPool.DataPool,
	headerPool *dataPool.DataPool,
	txBlockPool *dataPool.DataPool,
	stateBlockPool *dataPool.DataPool,
	peerBlockPool *dataPool.DataPool,
	addrConv state.AddressConverter,
) error {

	err := is.SetTransactionInterceptor(txPool, addrConv)
	if err != nil {
		return err
	}

	err = is.SetHeaderInterceptor(headerPool)
	if err != nil {
		return err
	}

	err = is.AppendBlockBodyInterceptor("txBlock", txBlockPool, block.NewInterceptedTxBlockBody())
	if err != nil {
		return err
	}

	err = is.AppendBlockBodyInterceptor("stateBlock", stateBlockPool, block.NewInterceptedStateBlockBody())
	if err != nil {
		return err
	}

	err = is.AppendBlockBodyInterceptor("peerBlock", peerBlockPool, block.NewInterceptedPeerBlockBody())
	if err != nil {
		return err
	}

	return nil
}
