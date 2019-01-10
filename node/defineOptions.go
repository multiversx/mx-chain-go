package node

import (
	"context"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// WithPort sets up the port option for the Node
func WithPort(port int) Option {
	return func(n *Node) error {
		n.port = port
		return nil
	}
}

// WithMarshalizer sets up the marshalizer option for the Node
func WithMarshalizer(marshalizer marshal.Marshalizer) Option {
	return func(n *Node) error {
		if marshalizer == nil {
			return errNilMarshalizer
		}
		n.marshalizer = marshalizer
		return nil
	}
}

// WithContext sets up the context option for the Node
func WithContext(ctx context.Context) Option {
	return func(n *Node) error {
		if ctx == nil {
			return errNilContext
		}
		n.ctx = ctx
		return nil
	}
}

// WithHasher sets up the hasher option for the Node
func WithHasher(hasher hashing.Hasher) Option {
	return func(n *Node) error {
		if hasher == nil {
			return errNilHasher
		}
		n.hasher = hasher
		return nil
	}
}

// WithMaxAllowedPeers sets up the maxAllowedPeers option for the Node
func WithMaxAllowedPeers(maxAllowedPeers int) Option {
	return func(n *Node) error {
		n.maxAllowedPeers = maxAllowedPeers
		return nil
	}
}

// WithPubSubStrategy sets up the strategy option for the Node
func WithPubSubStrategy(strategy p2p.PubSubStrategy) Option {
	return func(n *Node) error {
		n.pubSubStrategy = strategy
		return nil
	}
}

// WithAccountsAdapter sets up the accounts adapter option for the Node
func WithAccountsAdapter(accounts state.AccountsAdapter) Option {
	return func(n *Node) error {
		if accounts == nil {
			return errNilAccountsAdapter
		}
		n.accounts = accounts
		return nil
	}
}

// WithAddressConverter sets up the address converter adapter option for the Node
func WithAddressConverter(addrConverter state.AddressConverter) Option {
	return func(n *Node) error {
		if addrConverter == nil {
			return errNilAddressConverter
		}
		n.addrConverter = addrConverter
		return nil
	}
}

// WithBlockChain sets up the blockchain option for the Node
func WithBlockChain(blkc *blockchain.BlockChain) Option {
	return func(n *Node) error {
		if blkc == nil {
			return errNilBlockchain
		}
		n.blkc = blkc
		return nil
	}
}

// WithPrivateKey sets up the private key option for the Node
func WithPrivateKey(sk crypto.PrivateKey) Option {
	return func(n *Node) error {
		if sk == nil {
			return errNilPrivateKey
		}
		n.privateKey = sk
		return nil
	}
}

// WithSingleSignKeyGenerator sets up the single sign key generator option for the Node
func WithSingleSignKeyGenerator(keyGen crypto.KeyGenerator) Option {
	return func(n *Node) error {
		if keyGen == nil {
			return errNilSingleSignKeyGen
		}
		n.singleSignKeyGen = keyGen
		return nil
	}
}

// WithInitialNodesPubKeys sets up the initial nodes public key option for the Node
func WithInitialNodesPubKeys(pubKeys []string) Option {
	return func(n *Node) error {
		n.initialNodesPubkeys = pubKeys
		return nil
	}
}

// WithPublicKey sets up the public key option for the Node
func WithPublicKey(pk crypto.PublicKey) Option {
	return func(n *Node) error {
		if pk == nil {
			return errNilPublicKey
		}

		n.publicKey = pk
		return nil
	}
}

// WithRoundDuration sets up the round duration option for the Node
func WithRoundDuration(roundDuration uint64) Option {
	return func(n *Node) error {
		if roundDuration == 0 {
			return errZeroRoundDurationNotSupported
		}
		n.roundDuration = roundDuration
		return nil
	}
}

// WithConsensusGroupSize sets up the consensus group size option for the Node
func WithConsensusGroupSize(consensusGroupSize int) Option {
	return func(n *Node) error {
		if consensusGroupSize < 1 {
			return errNegativeOrZeroConsensusGroupSize
		}
		n.consensusGroupSize = consensusGroupSize
		return nil
	}
}

// WithSyncer sets up the syncer option for the Node
func WithSyncer(syncer ntp.SyncTimer) Option {
	return func(n *Node) error {
		if syncer == nil {
			return errNilSyncTimer
		}
		n.syncer = syncer
		return nil
	}
}

// WithBlockProcessor sets up the block processor option for the Node
func WithBlockProcessor(blockProcessor process.BlockProcessor) Option {
	return func(n *Node) error {
		if blockProcessor == nil {
			return errNilBlockProcessor
		}
		n.blockProcessor = blockProcessor
		return nil
	}
}

// WithGenesisTime sets up the genesis time option for the Node
func WithGenesisTime(genesisTime time.Time) Option {
	return func(n *Node) error {
		n.genesisTime = genesisTime
		return nil
	}
}

// WithElasticSubrounds sets up the elastic subround option for the Node
func WithElasticSubrounds(elasticSubrounds bool) Option {
	return func(n *Node) error {
		n.elasticSubrounds = elasticSubrounds
		return nil
	}
}

// WithDataPool sets up the transient data pool option for the Node
func WithDataPool(dataPool data.TransientDataHolder) Option {
	return func(n *Node) error {
		if dataPool == nil {
			return errNilDataPool
		}
		n.dataPool = dataPool
		return nil
	}
}

// WithShardCoordinator sets up the transient shard coordinator for the Node
func WithShardCoordinator(shardCoordinator sharding.ShardCoordinator) Option {
	return func(n *Node) error {
		if shardCoordinator == nil {
			return errNilShardCoordinator
		}
		n.shardCoordinator = shardCoordinator
		return nil
	}
}

// WithUint64ByteSliceConverter sets up the uint64 <-> []byte converter
func WithUint64ByteSliceConverter(converter typeConverters.Uint64ByteSliceConverter) Option {
	return func(n *Node) error {
		if converter == nil {
			return errNilUint64ByteSliceConverter
		}
		n.uint64ByteSliceConverter = converter
		return nil
	}
}

// WithInitialNodesBalances sets up the initial map of nodes public keys and their respective balances
func WithInitialNodesBalances(balances map[string]big.Int) Option {
	return func(n *Node) error {
		if balances == nil {
			return errNilBalances
		}
		n.initialNodesBalances = balances
		return nil
	}
}

// WithMultisig sets up the multisig option for the Node
func WithMultisig(multisig crypto.MultiSigner) Option {
	return func(n *Node) error {
		if multisig == nil {
			return errNilMultiSig
		}
		n.multisig = multisig
		return nil
	}
}
