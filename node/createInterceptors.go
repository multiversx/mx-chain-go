package node

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/interceptor"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
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

	txStorer := n.blkc.GetStorer(blockchain.TransactionUnit)

	txInterceptor, err := transaction.NewTxInterceptor(
		intercept,
		n.dataPool.Transactions(),
		txStorer,
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

	headerStorer := n.blkc.GetStorer(blockchain.BlockHeaderUnit)

	hdrInterceptor, err := block.NewHeaderInterceptor(
		intercept,
		n.dataPool.Headers(),
		n.dataPool.HeadersNonces(),
		headerStorer,
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

	txBlockBodyStorer := n.blkc.GetStorer(blockchain.TxBlockBodyUnit)

	txBlockBodyInterceptor, err := block.NewGenericBlockBodyInterceptor(
		intercept,
		n.dataPool.TxBlocks(),
		txBlockBodyStorer,
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

	peerBlockBodyStorer := n.blkc.GetStorer(blockchain.PeerBlockBodyUnit)

	peerChBodyInterceptor, err := block.NewGenericBlockBodyInterceptor(
		intercept,
		n.dataPool.PeerChangesBlocks(),
		peerBlockBodyStorer,
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

	stateBlockBodyStorer := n.blkc.GetStorer(blockchain.StateBlockBodyUnit)

	stateBodyInterceptor, err := block.NewGenericBlockBodyInterceptor(
		intercept,
		n.dataPool.StateBlocks(),
		stateBlockBodyStorer,
		n.hasher,
		n.shardCoordinator,
	)

	if err != nil {
		return err
	}

	n.interceptors = append(n.interceptors, stateBodyInterceptor)
	return nil
}
