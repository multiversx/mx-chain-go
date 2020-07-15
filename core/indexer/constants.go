package indexer

var headerContentTypeJSON = []string{"application/json"}
const headerXSRF = "kbn-xsrf"
const headerContentType = "Content-Type"

const awsRegion = "us-east-1"
const kibanaPluginPath = "_plugin/kibana/api"
const txsBulkSizeThreshold = 800000 // 0.8MB

const maxNumberOfDocumentsGet = 5000
const txIndex = "transactions"
const txPolicy = "transactions_policy"
const blockIndex = "blocks"
const miniblocksIndex = "miniblocks"
const tpsIndex = "tps"
const validatorsIndex = "validators"
const roundIndex = "rounds"
const ratingIndex = "rating"

const metachainTpsDocID = "meta"
const shardTpsDocIDPrefix = "shard"

const dataNodeIdentifier = "data"

