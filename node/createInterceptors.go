package node

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/interceptor"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
)

func (n *Node) createInterceptors() error {
	err := n.createTxInterceptor()
	if err != nil {
		return err
	}

	err = n.createHdrInterceptor()
	if err != nil {
		return err
	}

	err = n.createTxBlockBodyInterceptor()
	if err != nil {
		return err
	}

	err = n.createPeerChBlockBodyInterceptor()
	if err != nil {
		return err
	}

	err = n.createStateBlockBodyInterceptor()
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) createTxInterceptor() error {
	intercept, err := interceptor.NewTopicInterceptor(string(TransactionTopic), n.messenger, transaction.NewInterceptedTransaction())
	if err != nil {
		return err
	}
	txInterceptor, err := transaction.NewTxInterceptor(
		intercept,
		n.dataPool.Transactions(),
		n.addrConverter,
		n.hasher,
		n.singleSignKeyGen,
		n.shardCoordinator)

	if err != nil {
		return err
	}

	n.interceptors = append(n.interceptors, txInterceptor)
	return nil
}

func (n *Node) createHdrInterceptor() error {
	intercept, err := interceptor.NewTopicInterceptor(string(HeadersTopic), n.messenger, block.NewInterceptedHeader())
	if err != nil {
		return err
	}
	hdrInterceptor, err := block.NewHeaderInterceptor(
		intercept,
		n.dataPool.Headers(),
		n.dataPool.HeadersNonces(),
		n.hasher,
		n.shardCoordinator,
	)

	if err != nil {
		return err
	}

	n.interceptors = append(n.interceptors, hdrInterceptor)
	return nil
}

func (n *Node) createTxBlockBodyInterceptor() error {
	intercept, err := interceptor.NewTopicInterceptor(string(TxBlockBodyTopic), n.messenger, block.NewInterceptedTxBlockBody())
	if err != nil {
		return err
	}
	txBlockBodyInterceptor, err := block.NewGenericBlockBodyInterceptor(
		intercept,
		n.dataPool.TxBlocks(),
		n.hasher,
		n.shardCoordinator,
	)

	if err != nil {
		return err
	}

	n.interceptors = append(n.interceptors, txBlockBodyInterceptor)
	return nil
}

func (n *Node) createPeerChBlockBodyInterceptor() error {
	intercept, err := interceptor.NewTopicInterceptor(string(PeerChBodyTopic), n.messenger, block.NewInterceptedPeerBlockBody())
	if err != nil {
		return err
	}
	peerChBodyInterceptor, err := block.NewGenericBlockBodyInterceptor(
		intercept,
		n.dataPool.PeerChangesBlocks(),
		n.hasher,
		n.shardCoordinator,
	)

	if err != nil {
		return err
	}

	n.interceptors = append(n.interceptors, peerChBodyInterceptor)
	return nil
}

func (n *Node) createStateBlockBodyInterceptor() error {
	intercept, err := interceptor.NewTopicInterceptor(string(StateBodyTopic), n.messenger, block.NewInterceptedStateBlockBody())
	if err != nil {
		return err
	}
	stateBodyInterceptor, err := block.NewGenericBlockBodyInterceptor(
		intercept,
		n.dataPool.StateBlocks(),
		n.hasher,
		n.shardCoordinator,
	)

	if err != nil {
		return err
	}

	n.interceptors = append(n.interceptors, stateBodyInterceptor)
	return nil
}
