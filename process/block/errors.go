package block

import "errors"

var errInvalidNumOutGoingMBInMetaHdrProposal = errors.New("invalid number of outgoing miniblocks in meta header proposal, should be zero")

var errInvalidNumOutGoingTxsInMetaHdrProposal = errors.New("invalid number of outgoing transactions in meta header proposal, should be zero")

var errNilPreviousHeader = errors.New("nil previous header")

var errInvalidMiniBlocks = errors.New("invalid mini blocks")

var errIncludedContendedUnsettledHeader = errors.New("included contended header not yet settled")

var errContendedHeaderWithBetterCompetitor = errors.New("included contended header has a better proofed competitor")

var errContendedHeaderInsideArbitrationWindow = errors.New("included contended header before the arbitration discovery window elapsed")

var errReferencedNonAncestorMetaHeader = errors.New("shard header references a meta block that is not an ancestor of the built block")

var errNilMetaAncestryView = errors.New("nil meta ancestry view")

var errReferencedDeadMetaHeader = errors.New("shard header references a meta block the authority built past")
