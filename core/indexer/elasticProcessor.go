package indexer

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

const numDecimalsInFloatBalance = 10

type elasticProcessor struct {
	*txDatabaseProcessor

	elasticClient  DatabaseClientHandler
	parser         *dataParser
	enabledIndexes map[string]struct{}
	accountsDB     state.AccountsAdapter
	denomination   int
}

// NewElasticProcessor creates an elasticsearch es and handles saving
func NewElasticProcessor(arguments ArgElasticProcessor) (ElasticProcessor, error) {
	err := checkArgElasticProcessor(arguments)
	if err != nil {
		return nil, err
	}

	ei := &elasticProcessor{
		elasticClient: arguments.DBClient,
		parser: &dataParser{
			hasher:      arguments.Hasher,
			marshalizer: arguments.Marshalizer,
		},
		enabledIndexes: arguments.EnabledIndexes,
		accountsDB:     arguments.AccountsDB,
		denomination:   arguments.Denomination,
	}

	ei.txDatabaseProcessor = newTxDatabaseProcessor(
		arguments.Hasher,
		arguments.Marshalizer,
		arguments.AddressPubkeyConverter,
		arguments.ValidatorPubkeyConverter,
	)

	if arguments.Options.UseKibana {
		err = ei.initWithKibana(arguments.IndexTemplates, arguments.IndexPolicies)
		if err != nil {
			return nil, err
		}
	} else {
		err = ei.initNoKibana(arguments.IndexTemplates)
		if err != nil {
			return nil, err
		}
	}

	return ei, nil
}

func checkArgElasticProcessor(arguments ArgElasticProcessor) error {
	if check.IfNil(arguments.DBClient) {
		return ErrNilDatabaseClient
	}
	if check.IfNil(arguments.Marshalizer) {
		return core.ErrNilMarshalizer
	}
	if check.IfNil(arguments.Hasher) {
		return core.ErrNilHasher
	}
	if check.IfNil(arguments.AddressPubkeyConverter) {
		return ErrNilPubkeyConverter
	}
	if check.IfNil(arguments.ValidatorPubkeyConverter) {
		return ErrNilPubkeyConverter
	}
	if check.IfNil(arguments.AccountsDB) {
		return ErrNilAccountsDB
	}
	if arguments.Options == nil {
		return ErrNilOptions
	}

	return nil
}

func (ei *elasticProcessor) initWithKibana(indexTemplates, indexPolicies map[string]*bytes.Buffer) error {
	err := ei.createOpenDistroTemplates(indexTemplates)
	if err != nil {
		return err
	}

	err = ei.createIndexPolicies(indexPolicies)
	if err != nil {
		return err
	}

	err = ei.createIndexTemplates(indexTemplates)
	if err != nil {
		return err
	}

	err = ei.createIndexes()
	if err != nil {
		return err
	}

	err = ei.createAliases()
	if err != nil {
		return err
	}

	return nil
}

func (ei *elasticProcessor) initNoKibana(indexTemplates map[string]*bytes.Buffer) error {
	err := ei.createOpenDistroTemplates(indexTemplates)
	if err != nil {
		return err
	}

	return ei.createIndexes()
}

func (ei *elasticProcessor) createIndexPolicies(indexPolicies map[string]*bytes.Buffer) error {

	indexesPolicies := []string{txPolicy, blockPolicy, miniblocksPolicy, tpsPolicy, ratingPolicy, roundPolicy, validatorsPolicy, accountsPolicy}
	for _, indexPolicyName := range indexesPolicies {
		indexPolicy := getTemplateByName(indexPolicyName, indexPolicies)
		if indexPolicy != nil {
			err := ei.elasticClient.CheckAndCreatePolicy(indexPolicyName, indexPolicy)
			if err != nil {
				log.Error("cehck and create policy", "err", err)
				return err
			}
		}
	}

	return nil
}

func (ei *elasticProcessor) createOpenDistroTemplates(indexTemplates map[string]*bytes.Buffer) error {
	opendistroTemplate := getTemplateByName("opendistro", indexTemplates)
	if opendistroTemplate != nil {
		err := ei.elasticClient.CheckAndCreateTemplate("opendistro", opendistroTemplate)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ei *elasticProcessor) createIndexTemplates(indexTemplates map[string]*bytes.Buffer) error {
	indexes := []string{txIndex, blockIndex, miniblocksIndex, tpsIndex, ratingIndex, roundIndex, validatorsIndex, accountsIndex}
	for _, index := range indexes {
		indexTemplate := getTemplateByName(index, indexTemplates)
		if indexTemplate != nil {
			err := ei.elasticClient.CheckAndCreateTemplate(index, indexTemplate)
			if err != nil {
				log.Error("cehck and create template", "err", err)
				return err
			}
		}
	}
	return nil
}

func (ei *elasticProcessor) createIndexes() error {
	indexes := []string{txIndex, blockIndex, miniblocksIndex, tpsIndex, ratingIndex, roundIndex, validatorsIndex, accountsIndex}
	for _, index := range indexes {
		indexName := fmt.Sprintf("%s-000001", index)
		err := ei.elasticClient.CheckAndCreateIndex(indexName)
		if err != nil {
			log.Error("cehck and create index", "err", err)
			return err
		}
	}
	return nil
}

func (ei *elasticProcessor) createAliases() error {
	indexes := []string{txIndex, blockIndex, miniblocksIndex, tpsIndex, ratingIndex, roundIndex, validatorsIndex, accountsIndex}
	for _, index := range indexes {
		indexName := fmt.Sprintf("%s-000001", index)
		err := ei.elasticClient.CheckAndCreateAlias(index, indexName)
		if err != nil {
			log.Error("cehck and create alias", "err", err)
			return err
		}
	}

	return nil
}

func (ei *elasticProcessor) getExistingObjMap(hashes []string, index string) (map[string]bool, error) {
	if len(hashes) == 0 {
		return make(map[string]bool), nil
	}

	response, err := ei.elasticClient.DoMultiGet(getDocumentsByIDsQuery(hashes), index)
	if err != nil {
		return nil, err
	}

	return getDecodedResponseMultiGet(response), nil
}

func getTemplateByName(templateName string, templateList map[string]*bytes.Buffer) *bytes.Buffer {
	if template, ok := templateList[templateName]; ok {
		return template
	}

	log.Debug("elasticProcessor.getTemplateByName", "could not find template", templateName)
	return nil
}

// SaveHeader will prepare and save information about a header in elasticsearch server
func (ei *elasticProcessor) SaveHeader(
	header data.HeaderHandler,
	signersIndexes []uint64,
	body *block.Body,
	notarizedHeadersHashes []string,
	txsSize int,
) error {
	if _, ok := ei.enabledIndexes[blockIndex]; !ok {
		return nil
	}

	var buff bytes.Buffer

	serializedBlock, headerHash, err := ei.parser.getSerializedElasticBlockAndHeaderHash(header, signersIndexes, body, notarizedHeadersHashes, txsSize)
	if err != nil {
		return err
	}

	buff.Grow(len(serializedBlock))
	_, err = buff.Write(serializedBlock)
	if err != nil {
		return err
	}

	req := &esapi.IndexRequest{
		Index:      blockIndex,
		DocumentID: hex.EncodeToString(headerHash),
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	return ei.elasticClient.DoRequest(req)
}

// RemoveHeader will remove a block from elasticsearch server
func (ei *elasticProcessor) RemoveHeader(header data.HeaderHandler) error {
	headerHash, err := core.CalculateHash(ei.marshalizer, ei.hasher, header)
	if err != nil {
		return err
	}

	return ei.elasticClient.DoBulkRemove(blockIndex, []string{hex.EncodeToString(headerHash)})
}

// RemoveMiniblocks will remove all miniblocks that are in header from elasticsearch server
func (ei *elasticProcessor) RemoveMiniblocks(header data.HeaderHandler, body *block.Body) error {
	if body == nil || len(header.GetMiniBlockHeadersHashes()) == 0 {
		return nil
	}

	encodedMiniblocksHashes := make([]string, 0)
	selfShardID := header.GetShardID()
	for _, miniblock := range body.MiniBlocks {
		if miniblock.Type == block.PeerBlock {
			continue
		}

		isDstMe := selfShardID == miniblock.ReceiverShardID
		isCrossShard := miniblock.ReceiverShardID != miniblock.SenderShardID
		if isDstMe && isCrossShard {
			continue
		}

		miniblockHash, err := core.CalculateHash(ei.marshalizer, ei.hasher, miniblock)
		if err != nil {
			log.Debug("indexer.RemoveMiniblocks cannot calculate miniblock hash",
				"error", err.Error())
			continue
		}
		encodedMiniblocksHashes = append(encodedMiniblocksHashes, hex.EncodeToString(miniblockHash))

	}

	return ei.elasticClient.DoBulkRemove(miniblocksIndex, encodedMiniblocksHashes)
}

// SetTxLogsProcessor will set tx logs processor
func (ei *elasticProcessor) SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase) {
	ei.txLogsProcessor = txLogsProc
}

// SaveMiniblocks will prepare and save information about miniblocks in elasticsearch server
func (ei *elasticProcessor) SaveMiniblocks(header data.HeaderHandler, body *block.Body) (map[string]bool, error) {
	if _, ok := ei.enabledIndexes[miniblocksIndex]; !ok {
		return map[string]bool{}, nil
	}

	miniblocks := ei.parser.getMiniblocks(header, body)
	if len(miniblocks) == 0 {
		return make(map[string]bool), nil
	}

	buff, mbHashDb := serializeBulkMiniBlocks(header.GetShardID(), miniblocks, ei.getExistingObjMap)
	return mbHashDb, ei.elasticClient.DoBulkRequest(&buff, miniblocksIndex)
}

// SaveTransactions will prepare and save information about a transactions in elasticsearch server
func (ei *elasticProcessor) SaveTransactions(
	body *block.Body,
	header data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	selfShardID uint32,
	mbsInDb map[string]bool,
) error {
	if _, ok := ei.enabledIndexes[txIndex]; !ok {
		return nil
	}

	txs, alteredAccounts := ei.prepareTransactionsForDatabase(body, header, txPool, selfShardID)
	buffSlice := serializeTransactions(txs, selfShardID, ei.getExistingObjMap, mbsInDb)

	for idx := range buffSlice {
		err := ei.elasticClient.DoBulkRequest(&buffSlice[idx], txIndex)
		if err != nil {
			log.Warn("indexer indexing bulk of transactions",
				"error", err.Error())
			return err
		}
	}

	ei.indexAlteredAccounts(alteredAccounts)

	return nil
}

// SaveShardStatistics will prepare and save information about a shard statistics in elasticsearch server
func (ei *elasticProcessor) SaveShardStatistics(tpsBenchmark statistics.TPSBenchmark) error {
	if _, ok := ei.enabledIndexes[tpsIndex]; !ok {
		return nil
	}

	buff := prepareGeneralInfo(tpsBenchmark)

	for _, shardInfo := range tpsBenchmark.ShardStatistics() {
		serializedShardInfo, serializedMetaInfo := serializeShardInfo(shardInfo)
		if serializedShardInfo == nil {
			continue
		}

		buff.Grow(len(serializedMetaInfo) + len(serializedShardInfo))
		_, err := buff.Write(serializedMetaInfo)
		if err != nil {
			log.Warn("elastic search: update TPS write meta", "error", err.Error())
		}
		_, err = buff.Write(serializedShardInfo)
		if err != nil {
			log.Warn("elastic search: update TPS write serialized data", "error", err.Error())
		}
	}

	return ei.elasticClient.DoBulkRequest(&buff, tpsIndex)
}

// SaveValidatorsRating will save validators rating
func (ei *elasticProcessor) SaveValidatorsRating(index string, validatorsRatingInfo []workItems.ValidatorRatingInfo) error {
	if _, ok := ei.enabledIndexes[ratingIndex]; !ok {
		return nil
	}

	var buff bytes.Buffer

	infosRating := ValidatorsRatingInfo{ValidatorsInfos: validatorsRatingInfo}

	marshalizedInfoRating, err := json.Marshal(&infosRating)
	if err != nil {
		log.Debug("indexer: marshal", "error", "could not marshal validators rating")
		return err
	}

	buff.Grow(len(marshalizedInfoRating))
	_, err = buff.Write(marshalizedInfoRating)
	if err != nil {
		log.Warn("elastic search: save validators rating, write", "error", err.Error())
	}

	req := &esapi.IndexRequest{
		Index:      ratingIndex,
		DocumentID: index,
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	return ei.elasticClient.DoRequest(req)
}

// SaveShardValidatorsPubKeys will prepare and save information about a shard validators public keys in elasticsearch server
func (ei *elasticProcessor) SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error {
	if _, ok := ei.enabledIndexes[validatorsIndex]; !ok {
		return nil
	}

	var buff bytes.Buffer

	shardValPubKeys := ValidatorsPublicKeys{
		PublicKeys: make([]string, 0, len(shardValidatorsPubKeys)),
	}
	for _, validatorPk := range shardValidatorsPubKeys {
		strValidatorPk := ei.validatorPubkeyConverter.Encode(validatorPk)
		shardValPubKeys.PublicKeys = append(shardValPubKeys.PublicKeys, strValidatorPk)
	}

	marshalizedValidatorPubKeys, err := json.Marshal(shardValPubKeys)
	if err != nil {
		log.Debug("indexer: marshal", "error", "could not marshal validators public keys")
		return err
	}

	buff.Grow(len(marshalizedValidatorPubKeys))
	_, err = buff.Write(marshalizedValidatorPubKeys)
	if err != nil {
		log.Warn("elastic search: save shard validators pub keys, write", "error", err.Error())
	}

	req := &esapi.IndexRequest{
		Index:      validatorsIndex,
		DocumentID: fmt.Sprintf("%d_%d", shardID, epoch),
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	return ei.elasticClient.DoRequest(req)
}

// SaveRoundsInfo will prepare and save information about a slice of rounds in elasticsearch server
func (ei *elasticProcessor) SaveRoundsInfo(infos []workItems.RoundInfo) error {
	if _, ok := ei.enabledIndexes[roundIndex]; !ok {
		return nil
	}

	var buff bytes.Buffer

	for _, info := range infos {
		serializedRoundInfo, meta := serializeRoundInfo(info)

		buff.Grow(len(meta) + len(serializedRoundInfo))
		_, err := buff.Write(meta)
		if err != nil {
			log.Warn("indexer: cannot write meta", "error", err.Error())
		}

		_, err = buff.Write(serializedRoundInfo)
		if err != nil {
			log.Warn("indexer: cannot write serialized round info", "error", err.Error())
		}
	}

	return ei.elasticClient.DoBulkRequest(&buff, roundIndex)
}

func (ei *elasticProcessor) indexAlteredAccounts(accounts map[string]struct{}) {
	for address := range accounts {
		addressBytes, err := ei.addressPubkeyConverter.Decode(address)
		if err != nil {
			log.Warn("cannot decode address", "address", address, "error", err)
			continue
		}

		account, err := ei.accountsDB.LoadAccount(addressBytes)
		if err != nil {
			log.Warn("cannot load account", "address bytes", addressBytes, "error", err)
			continue
		}

		userAccount, ok := account.(state.UserAccountHandler)
		if !ok {
			log.Warn("cannot cast AccountHandler to type UserAccountHandler")
			continue
		}

		err = ei.SaveAccount(userAccount)
		if err != nil {
			log.Warn("cannot index account",
				"address", ei.addressPubkeyConverter.Encode(userAccount.AddressBytes()),
				"error", err)
		}
	}
}

// SaveAccount will prepare and save information about an account in elasticsearch server
func (ei *elasticProcessor) SaveAccount(account state.UserAccountHandler) error {
	if _, ok := ei.enabledIndexes[accountsIndex]; !ok {
		return nil
	}

	if check.IfNil(account) {
		return nil
	}

	var buff bytes.Buffer

	balanceAsFloat := computeBalanceAsFloat(account.GetBalance(), ei.denomination, numDecimalsInFloatBalance)
	acc := AccountInfo{
		Address:    ei.addressPubkeyConverter.Encode(account.AddressBytes()),
		Nonce:      account.GetNonce(),
		Balance:    account.GetBalance().String(),
		BalanceNum: balanceAsFloat,
	}

	accBytes, err := json.Marshal(&acc)
	if err != nil {
		return err
	}

	buff.Grow(len(accBytes))
	_, err = buff.Write(accBytes)
	if err != nil {
		log.Warn("elastic search: save account, write", "error", err.Error())
	}

	req := &esapi.IndexRequest{
		Index:      accountsIndex,
		DocumentID: ei.addressPubkeyConverter.Encode(account.AddressBytes()),
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	return ei.elasticClient.DoRequest(req)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ei *elasticProcessor) IsInterfaceNil() bool {
	return ei == nil
}
