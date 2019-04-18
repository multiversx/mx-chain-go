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

//ConsensusCore encapsulates all needed Data for the Consensus
type ConsensusCore interface {
	//Blockchain gets the ChainHandler stored in the SPOSConsensusCore
	Blockchain() data.ChainHandler
	//BlockProcessor gets the BlockProcessor stored in the SPOSConsensusCore
	BlockProcessor() process.BlockProcessor
	//BootStrapper gets the Bootstrapper stored in the SPOSConsensusCore
	BootStrapper() process.Bootstrapper
	//Chronology gets the ChronologyHandler stored in the SPOSConsensusCore
	Chronology() consensus.ChronologyHandler
	//Hasher gets the Hasher stored in the SPOSConsensusCore
	Hasher() hashing.Hasher
	//Marshalizer gets the Marshalizer stored in the SPOSConsensusCore
	Marshalizer() marshal.Marshalizer
	//MultiSigner gets the MultiSigner stored in the SPOSConsensusCore
	MultiSigner() crypto.MultiSigner
	//Rounder gets the Rounder stored in the SPOSConsensusCore
	Rounder() consensus.Rounder
	//ShardCoordinator gets the Coordinator stored in the SPOSConsensusCore
	ShardCoordinator() sharding.Coordinator
	//SyncTimer gets the SyncTimer stored in the SPOSConsensusCore
	SyncTimer() ntp.SyncTimer
	//ValidatorGroupSelector gets the ValidatorGroupSelector stored in the SPOSConsensusCore
	ValidatorGroupSelector() consensus.ValidatorGroupSelector
}
