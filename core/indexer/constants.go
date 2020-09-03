package indexer

var headerContentTypeJSON = []string{"application/json"}

const headerXSRF = "kbn-xsrf"
const headerContentType = "Content-Type"

const kibanaPluginPath = "_plugin/kibana/api"

const (
	blockIndex      = "blocks"
	miniblocksIndex = "miniblocks"
	txIndex         = "transactions"
	tpsIndex        = "tps"
	validatorsIndex = "validators"
	roundIndex      = "rounds"
	ratingIndex     = "rating"
)

const (
	metachainTpsDocID   = "meta"
	shardTpsDocIDPrefix = "shard"
)

const (
	txPolicy    = "transactions_policy"
	blockPolicy = "blocks_policy"
)

const txsBulkSizeThreshold = 800000 // 0.8MB
