package slash

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// SlashingDetectorResultHandler - contains the result of a slashing detection
type SlashingDetectorResultHandler interface {
	//GetType - contains the type of slashing detection
	GetType() string
	//GetData1 - contains the first data of slashing detection
	GetData1() process.InterceptedData
	//GetData2 - contains the second data of slashing detection
	GetData2() process.InterceptedData
}

// SlashingProofHandler - contains a proof for a slashing event and can be wrapped in a transaction
type SlashingProofHandler interface {
	// GetLevel - contains the slashing level for the current slashing type
	// multiple colluding parties should have a higher leve
	GetLevel() string
	//GetType - contains the type of slashing detection
	GetType() string
	//GetData1 - contains the first data of slashing detection
	GetData1() process.InterceptedData
	//GetData2 - contains the second data of slashing detection
	GetData2() process.InterceptedData
}

// SlashingDetector - checks for slashable events and generates proofs to be used for slash
type SlashingDetector interface {
	// VerifyData - checks if an intercepted data represents a slashable event
	VerifyData(data process.InterceptedData) SlashingDetectorResultHandler
	// GenerateProof - creates the SlashingProofHandler for the DetectorResult to be added to the Tx Data Field
	GenerateProof(result SlashingDetectorResultHandler) SlashingProofHandler
}

// SlashingNotifier - creates a transaction from the generated proof of the slash detector and sends it to the network
type SlashingNotifier interface {
	// CreateShardSlashingTransaction - creates a slash transaction from the generated SlashingProofHandler
	CreateShardSlashingTransaction(proof SlashingProofHandler) data.TransactionHandler
	// CreateMetaSlashingEscalatedTransaction - creates a transaction for the metachain if x rounds passed
	// and no slash transaction has been created by any of the previous x proposers
	CreateMetaSlashingEscalatedTransaction(proof SlashingProofHandler) data.TransactionHandler
}

// SlashingTxProcessor - processes the proofs from the SlashingNotifier inside shards
type SlashingTxProcessor interface {
	// ProcessTx - processes a slash transaction that contains a proof from the SlashingNotifier
	// if the proof is valid, a SCResult with destination metachain is created,
	// where the actual slash actions are taken (jail, inactivate, remove balance etc)
	ProcessTx(transaction data.TransactionHandler) data.TransactionHandler
}

// Slasher - processes the validated slash proof from the shards
// and applies the necessary actions (jail, inactivate, remove balance etc)
type Slasher interface {
	// ExecuteSlash - processes a slash SCResult that contains information about the slashable event
	// validator could be jailed, inactivated, balance can be decreased
	ExecuteSlash(transaction data.TransactionHandler) data.TransactionHandler
}
