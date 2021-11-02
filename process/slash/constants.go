package slash

import coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"

// Used by slashing notifier to create a slashing transaction
// from a proof. Each transaction identifies a different
// slashing event based on this ID
const (
	// MultipleProposalProofID = MultipleProposal's ID
	MultipleProposalProofID byte = 0x1
	// MultipleSigningProofID = MultipleSigning's ID
	MultipleSigningProofID byte = 0x2
)

// ProofIDs is a convenience map containing all slashing events
// with their corresponding ID
var ProofIDs = map[coreSlash.SlashingType]byte{
	coreSlash.MultipleProposal: MultipleProposalProofID,
	coreSlash.MultipleSigning:  MultipleSigningProofID,
}
