package consensus

// SubRoundBlockType defines the types of subround block type
type SubRoundBlockType string

const (
	// SubRoundBlockTypeV1 defines the first variant of the subround block type
	SubRoundBlockTypeV1 SubRoundBlockType = "subround v1"
	// SubRoundBlockTypeV2 defines the second variant of the subround block type
	SubRoundBlockTypeV2 SubRoundBlockType = "subround v2"
)
