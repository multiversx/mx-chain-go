package indexer

var headerContentTypeJSON = []string{"application/json"}

const (
	headerXSRF        = "kbn-xsrf"
	headerContentType = "Content-Type"
	kibanaPluginPath  = "_plugin/kibana/api"

	blockIndex      = "blocks"
	miniblocksIndex = "miniblocks"
	txIndex         = "transactions"
	tpsIndex        = "tps"
	validatorsIndex = "validators"
	roundIndex      = "rounds"
	ratingIndex     = "rating"
	accountsIndex   = "accounts"

	txPolicy         = "transactions_policy"
	blockPolicy      = "blocks_policy"
	miniblocksPolicy = "miniblocks_policy"
	tpsPolicy        = "tps_policy"
	validatorsPolicy = "validators_policy"
	roundPolicy      = "rounds_policy"
	ratingPolicy     = "rating_policy"
	accountsPolicy   = "accounts_policy"

	metachainTpsDocID   = "meta"
	shardTpsDocIDPrefix = "shard"

	txsBulkSizeThreshold = 800000 // 0.8MB
)
