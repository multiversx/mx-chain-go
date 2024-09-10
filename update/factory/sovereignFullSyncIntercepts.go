package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type sovereignFullSyncInterceptorsContainerFactory struct {
	*fullSyncInterceptorsContainerFactory
}

// Create returns an interceptor container that will hold all interceptors in the system
func (ficf *sovereignFullSyncInterceptorsContainerFactory) Create() (process.InterceptorsContainer, process.InterceptorsContainer, error) {
	err := ficf.generateTxInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = ficf.generateUnsignedTxsInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = ficf.generateMiniBlocksInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = ficf.generateShardHeaderInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = ficf.generateTrieNodesInterceptors()
	if err != nil {
		return nil, nil, err
	}

	return ficf.mainContainer, ficf.fullArchiveContainer, nil
}

func (ficf *sovereignFullSyncInterceptorsContainerFactory) generateTxInterceptors() error {
	keys := make([]string, 0, 1)
	interceptorSlice := make([]process.Interceptor, 0, 1)

	identifierTx := factory.TransactionTopic + ficf.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	if !ficf.checkIfInterceptorExists(identifierTx) {
		interceptor, err := ficf.createOneTxInterceptor(identifierTx)
		if err != nil {
			return err
		}

		keys = append(keys, identifierTx)
		interceptorSlice = append(interceptorSlice, interceptor)
	}

	return ficf.addInterceptorsToContainers(keys, interceptorSlice)
}

func (ficf *sovereignFullSyncInterceptorsContainerFactory) generateUnsignedTxsInterceptors() error {
	keys := make([]string, 0, 1)
	interceptorsSlice := make([]process.Interceptor, 0, 1)

	identifierScr := factory.UnsignedTransactionTopic + ficf.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	if !ficf.checkIfInterceptorExists(identifierScr) {
		interceptor, err := ficf.createOneUnsignedTxInterceptor(identifierScr)
		if err != nil {
			return err
		}

		keys = append(keys, identifierScr)
		interceptorsSlice = append(interceptorsSlice, interceptor)
	}

	return ficf.addInterceptorsToContainers(keys, interceptorsSlice)
}

func (ficf *sovereignFullSyncInterceptorsContainerFactory) generateMiniBlocksInterceptors() error {
	keys := make([]string, 0, 1)
	interceptorsSlice := make([]process.Interceptor, 0, 1)

	identifierMiniBlocks := factory.MiniBlocksTopic + ficf.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	if !ficf.checkIfInterceptorExists(identifierMiniBlocks) {
		interceptor, err := ficf.createOneMiniBlocksInterceptor(identifierMiniBlocks)
		if err != nil {
			return err
		}

		keys = append(keys, identifierMiniBlocks)
		interceptorsSlice = append(interceptorsSlice, interceptor)
	}

	return ficf.addInterceptorsToContainers(keys, interceptorsSlice)
}

func (ficf *sovereignFullSyncInterceptorsContainerFactory) generateShardHeaderInterceptors() error {
	keys := make([]string, 0, 1)
	interceptorsSlice := make([]process.Interceptor, 0, 1)

	identifierHeader := factory.ShardBlocksTopic + ficf.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	if !ficf.checkIfInterceptorExists(identifierHeader) {
		interceptor, errCreate := ficf.createOneShardHeaderInterceptor(identifierHeader)
		if errCreate != nil {
			return errCreate
		}

		keys = append(keys, identifierHeader)
		interceptorsSlice = append(interceptorsSlice, interceptor)
	}

	return ficf.addInterceptorsToContainers(keys, interceptorsSlice)
}

func (ficf *sovereignFullSyncInterceptorsContainerFactory) generateTrieNodesInterceptors() error {
	keys := make([]string, 0, 1)
	trieInterceptors := make([]process.Interceptor, 0, 1)

	identifierTrieNodes := factory.ValidatorTrieNodesTopic + ficf.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	if !ficf.checkIfInterceptorExists(identifierTrieNodes) {
		interceptor, err := ficf.createOneTrieNodesInterceptor(identifierTrieNodes)
		if err != nil {
			return err
		}

		keys = append(keys, identifierTrieNodes)
		trieInterceptors = append(trieInterceptors, interceptor)
	}

	identifierTrieNodes = factory.AccountTrieNodesTopic + ficf.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	if !ficf.checkIfInterceptorExists(identifierTrieNodes) {
		interceptor, err := ficf.createOneTrieNodesInterceptor(identifierTrieNodes)
		if err != nil {
			return err
		}

		keys = append(keys, identifierTrieNodes)
		trieInterceptors = append(trieInterceptors, interceptor)
	}

	return ficf.addInterceptorsToContainers(keys, trieInterceptors)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ficf *sovereignFullSyncInterceptorsContainerFactory) IsInterfaceNil() bool {
	return ficf == nil
}
