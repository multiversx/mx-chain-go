package consensus

// SubroundBlockType defines the types of subround block type
type SubroundBlockType string

const (
	// SubroundBlockTypeV1 defines the first variant of the subround block type
	SubroundBlockTypeV1 SubroundBlockType = "subround v1"
	// SubroundBlockTypeV2 defines the second variant of the subround block type
	SubroundBlockTypeV2 SubroundBlockType = "subround v2"
)
