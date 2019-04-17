package spos

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

//ConsensusDataContainerInterface encapsulates all needed Data for the Consensus
type ConsensusDataContainerInterface interface {
	//Blockchain gets the ChainHandler stored in the ConsensusDataContainer
	Blockchain() data.ChainHandler
	//BlockProcessor gets the BlockProcessor stored in the ConsensusDataContainer
	BlockProcessor() process.BlockProcessor
	//BootStrapper gets the Bootstrapper stored in the ConsensusDataContainer
	BootStrapper() process.Bootstrapper
	//Chronology gets the ChronologyHandler stored in the ConsensusDataContainer
	Chronology() consensus.ChronologyHandler
	//ConsensusState gets the ConsensusState stored in the ConsensusDataContainer
	ConsensusState() *ConsensusState
	//Hasher gets the Hasher stored in the ConsensusDataContainer
	Hasher() hashing.Hasher
	//Marshalizer gets the Marshalizer stored in the ConsensusDataContainer
	Marshalizer() marshal.Marshalizer
	//MultiSigner gets the MultiSigner stored in the ConsensusDataContainer
	MultiSigner() crypto.MultiSigner
	//Rounder gets the Rounder stored in the ConsensusDataContainer
	Rounder() consensus.Rounder
	//ShardCoordinator gets the Coordinator stored in the ConsensusDataContainer
	ShardCoordinator() sharding.Coordinator
	//SyncTimer gets the SyncTimer stored in the ConsensusDataContainer
	SyncTimer() ntp.SyncTimer
	//ValidatorGroupSelector gets the ValidatorGroupSelector stored in the ConsensusDataContainer
	ValidatorGroupSelector() consensus.ValidatorGroupSelector
}
