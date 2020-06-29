package indexer

const txsBulkSizeThreshold = 900000 // 0.9MB

const maxNumberOfDocumentsGet = 5000
const txIndex = "transactions"
const blockIndex = "blocks"
const miniblocksIndex = "miniblocks"
const tpsIndex = "tps"
const validatorsIndex = "validators"
const roundIndex = "rounds"
const ratingIndex = "rating"

const metachainTpsDocID = "meta"
const shardTpsDocIDPrefix = "shard"
