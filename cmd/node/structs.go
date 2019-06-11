package main

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type Core struct {
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
	tr          trie.PatriciaMerkelTree
}

type State struct {
	addressConverter state.AddressConverter
	accountsAdapter  state.AccountsAdapter
}

type Data struct {
	blkc         data.ChainHandler
	store        dataRetriever.StorageService
	datapool     dataRetriever.PoolsHolder
	metaDatapool dataRetriever.MetaPoolsHolder
}

type Crypto struct {
	txSingleSigner crypto.SingleSigner
	singleSigner   crypto.SingleSigner
	multiSigner    crypto.MultiSigner
	txSignKeyGen   crypto.KeyGenerator
	txSignPrivKey  crypto.PrivateKey
	txSignPubKey   crypto.PublicKey
}

type Process struct {
	interceptorsContainer process.InterceptorsContainer
	resolversFinder       dataRetriever.ResolversFinder
	rounder               consensus.Rounder
	forkDetector          process.ForkDetector
	blockProcessor        process.BlockProcessor
	blockTracker          process.BlocksTracker
}

//func initCoreComponents(config *config.Config) (*Core, error) {
//	hasher, err := getHasherFromConfig(config)
//	if err != nil {
//		return nil, errors.New("could not create hasher: " + err.Error())
//	}
//
//	marshalizer, err := getMarshalizerFromConfig(config)
//	if err != nil {
//		return nil, errors.New("could not create marshalizer: " + err.Error())
//	}
//
//	tr, err := getTrie(config.AccountsTrieStorage, hasher)
//	if err != nil {
//		return nil, errors.New("error creating trie: " + err.Error())
//	}
//
//	return &Core{
//		hasher:      hasher,
//		marshalizer: marshalizer,
//		tr:          tr,
//	}, nil
//}
