package node

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/resolver"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
)

func (n *Node) createResolvers() error {
	err := n.createTxResolver()
	if err != nil {
		return err
	}

	err = n.createHdrResolver()
	if err != nil {
		return err
	}

	err = n.createTxBlockBodyResolver()
	if err != nil {
		return err
	}

	err = n.createPeerChBlockBodyResolver()
	if err != nil {
		return err
	}

	err = n.createStateBlockBodyResolver()
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) createTxResolver() error {
	resolve, err := resolver.NewTopicResolver(string(TransactionTopic), n.messenger, n.marshalizer)
	if err != nil {
		return err
	}

	txResolver, err := transaction.NewTxResolver(
		resolve,
		n.dataPool.Transactions(),
		n.blkc.GetStorer(blockchain.TransactionUnit),
		n.marshalizer)

	if err != nil {
		return err
	}

	n.resolvers = append(n.resolvers, txResolver)
	return nil
}

func (n *Node) createHdrResolver() error {
	resolve, err := resolver.NewTopicResolver(string(HeadersTopic), n.messenger, n.marshalizer)
	if err != nil {
		return err
	}

	hdrResolver, err := block.NewHeaderResolver(
		resolve,
		n.dataPool,
		n.blkc.GetStorer(blockchain.BlockHeaderUnit),
		n.marshalizer,
		n.uint64ByteSliceConverter)

	if err != nil {
		return err
	}

	n.resolvers = append(n.resolvers, hdrResolver)
	return nil
}

func (n *Node) createTxBlockBodyResolver() error {
	resolve, err := resolver.NewTopicResolver(string(TxBlockBodyTopic), n.messenger, n.marshalizer)
	if err != nil {
		return err
	}

	txBlkResolver, err := block.NewGenericBlockBodyResolver(
		resolve,
		n.dataPool.TxBlocks(),
		n.blkc.GetStorer(blockchain.TxBlockBodyUnit),
		n.marshalizer)

	if err != nil {
		return err
	}

	n.resolvers = append(n.resolvers, txBlkResolver)
	return nil
}

func (n *Node) createPeerChBlockBodyResolver() error {
	resolve, err := resolver.NewTopicResolver(string(PeerChBodyTopic), n.messenger, n.marshalizer)
	if err != nil {
		return err
	}

	peerChBlkResolver, err := block.NewGenericBlockBodyResolver(
		resolve,
		n.dataPool.PeerChangesBlocks(),
		n.blkc.GetStorer(blockchain.PeerBlockBodyUnit),
		n.marshalizer)

	if err != nil {
		return err
	}

	n.resolvers = append(n.resolvers, peerChBlkResolver)
	return nil
}

func (n *Node) createStateBlockBodyResolver() error {
	resolve, err := resolver.NewTopicResolver(string(StateBodyTopic), n.messenger, n.marshalizer)
	if err != nil {
		return err
	}

	stateBlkResolver, err := block.NewGenericBlockBodyResolver(
		resolve,
		n.dataPool.StateBlocks(),
		n.blkc.GetStorer(blockchain.StateBlockBodyUnit),
		n.marshalizer)

	if err != nil {
		return err
	}

	n.resolvers = append(n.resolvers, stateBlkResolver)
	return nil
}
