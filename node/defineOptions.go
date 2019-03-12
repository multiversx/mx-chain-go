package node

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// WithMessenger sets up the messenger option for the Node
func WithMessenger(mes p2p.Messenger) Option {
	return func(n *Node) error {
		if mes == nil {
			return ErrNilMessenger
		}
		n.messenger = mes
		return nil
	}
}

// WithMarshalizer sets up the marshalizer option for the Node
func WithMarshalizer(marshalizer marshal.Marshalizer) Option {
	return func(n *Node) error {
		if marshalizer == nil {
			return ErrNilMarshalizer
		}
		n.marshalizer = marshalizer
		return nil
	}
}

// WithHasher sets up the hasher option for the Node
func WithHasher(hasher hashing.Hasher) Option {
	return func(n *Node) error {
		if hasher == nil {
			return ErrNilHasher
		}
		n.hasher = hasher
		return nil
	}
}

// WithAccountsAdapter sets up the accounts adapter option for the Node
func WithAccountsAdapter(accounts state.AccountsAdapter) Option {
	return func(n *Node) error {
		if accounts == nil {
			return ErrNilAccountsAdapter
		}
		n.accounts = accounts
		return nil
	}
}

// WithAddressConverter sets up the address converter adapter option for the Node
func WithAddressConverter(addrConverter state.AddressConverter) Option {
	return func(n *Node) error {
		if addrConverter == nil {
			return ErrNilAddressConverter
		}
		n.addrConverter = addrConverter
		return nil
	}
}

// WithBlockChain sets up the blockchain option for the Node
func WithBlockChain(blkc *blockchain.BlockChain) Option {
	return func(n *Node) error {
		if blkc == nil {
			return ErrNilBlockchain
		}
		n.blkc = blkc
		return nil
	}
}

// WithPrivateKey sets up the private key option for the Node
func WithPrivateKey(sk crypto.PrivateKey) Option {
	return func(n *Node) error {
		if sk == nil {
			return ErrNilPrivateKey
		}
		n.privateKey = sk
		return nil
	}
}

// WithKeyGenerator sets up the single sign key generator option for the Node
func WithKeyGenerator(keyGen crypto.KeyGenerator) Option {
	return func(n *Node) error {
		if keyGen == nil {
			return ErrNilSingleSignKeyGen
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
			return ErrNilPublicKey
		}

		n.publicKey = pk
		return nil
	}
}

// WithRoundDuration sets up the round duration option for the Node
func WithRoundDuration(roundDuration uint64) Option {
	return func(n *Node) error {
		if roundDuration == 0 {
			return ErrZeroRoundDurationNotSupported
		}
		n.roundDuration = roundDuration
		return nil
	}
}

// WithConsensusGroupSize sets up the consensus group size option for the Node
func WithConsensusGroupSize(consensusGroupSize int) Option {
	return func(n *Node) error {
		if consensusGroupSize < 1 {
			return ErrNegativeOrZeroConsensusGroupSize
		}
		n.consensusGroupSize = consensusGroupSize
		return nil
	}
}

// WithSyncer sets up the syncer option for the Node
func WithSyncer(syncer ntp.SyncTimer) Option {
	return func(n *Node) error {
		if syncer == nil {
			return ErrNilSyncTimer
		}
		n.syncer = syncer
		return nil
	}
}

// WithBlockProcessor sets up the block processor option for the Node
func WithBlockProcessor(blockProcessor process.BlockProcessor) Option {
	return func(n *Node) error {
		if blockProcessor == nil {
			return ErrNilBlockProcessor
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
			return ErrNilDataPool
		}
		n.dataPool = dataPool
		return nil
	}
}

// WithShardCoordinator sets up the transient shard coordinator for the Node
func WithShardCoordinator(shardCoordinator sharding.ShardCoordinator) Option {
	return func(n *Node) error {
		if shardCoordinator == nil {
			return ErrNilShardCoordinator
		}
		n.shardCoordinator = shardCoordinator
		return nil
	}
}

// WithUint64ByteSliceConverter sets up the uint64 <-> []byte converter
func WithUint64ByteSliceConverter(converter typeConverters.Uint64ByteSliceConverter) Option {
	return func(n *Node) error {
		if converter == nil {
			return ErrNilUint64ByteSliceConverter
		}
		n.uint64ByteSliceConverter = converter
		return nil
	}
}

// WithInitialNodesBalances sets up the initial map of nodes public keys and their respective balances
func WithInitialNodesBalances(balances map[string]*big.Int) Option {
	return func(n *Node) error {
		if balances == nil {
			return ErrNilBalances
		}
		n.initialNodesBalances = balances
		return nil
	}
}

// WithSinglesig sets up the singlesig option for the Node
func WithSinglesig(singlesig crypto.SingleSigner) Option {
	return func(n *Node) error {
		if singlesig == nil {
			return ErrNilSingleSig
		}
		n.singlesig = singlesig
		return nil
	}
}

// WithMultisig sets up the multisig option for the Node
func WithMultisig(multisig crypto.MultiSigner) Option {
	return func(n *Node) error {
		if multisig == nil {
			return ErrNilMultiSig
		}
		n.multisig = multisig
		return nil
	}
}

// WithForkDetector sets up the multisig option for the Node
func WithForkDetector(forkDetector process.ForkDetector) Option {
	return func(n *Node) error {
		if forkDetector == nil {
			return ErrNilForkDetector
		}
		n.forkDetector = forkDetector
		return nil
	}
}

// WithInterceptorsContainer sets up the interceptors container option for the Node
func WithInterceptorsContainer(interceptorsContainer process.InterceptorsContainer) Option {
	return func(n *Node) error {
		if interceptorsContainer == nil {
			return ErrNilInterceptorsContainer
		}
		n.interceptorsContainer = interceptorsContainer
		return nil
	}
}

// WithResolversContainer sets up the interceptors container option for the Node
func WithResolversContainer(resolversContainer process.ResolversContainer) Option {
	return func(n *Node) error {
		if resolversContainer == nil {
			return ErrNilResolversContainer
		}
		n.resolversContainer = resolversContainer
		return nil
	}
}
