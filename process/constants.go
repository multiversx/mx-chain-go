package process

// BlockHeaderState specifies which is the state of the block header received
type BlockHeaderState int

const (
	// BHReceived defines ID of a received block header
	BHReceived BlockHeaderState = iota
	// BHProcessed defines ID of a processed block header
	BHProcessed
	// BHProposed defines ID of a proposed block header
	BHProposed
)

// MaxRoundsGap defines the maximum expected gap in terms of rounds, between metachain and shardchain, after which
// a block committed and broadcast from shardchain would be visible as notarized in metachain
const MaxRoundsGap = 3
