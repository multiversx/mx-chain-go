package block

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/display"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

var log = logger.DefaultLogger()

type baseProcessor struct {
	accounts     state.AccountsAdapter
	dataPool     dataRetriever.PoolsHolder
	forkDetector process.ForkDetector
	hasher       hashing.Hasher
	marshalizer  marshal.Marshalizer
	store        dataRetriever.StorageService
}

// RevertAccountState reverts the account state for cleanup failed process
func (bp *baseProcessor) RevertAccountState() {
	err := bp.accounts.RevertToSnapshot(0)
	if err != nil {
		log.Error(err.Error())
	}
}

// CheckBlockValidity method checks if the given block is valid
func (bp *baseProcessor) CheckBlockValidity(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) bool {
	if header == nil {
		log.Info(process.ErrNilBlockHeader.Error())
		return false
	}

	if blockChain == nil {
		log.Info(process.ErrNilBlockChain.Error())
		return false
	}

	if blockChain.GetCurrentBlockHeader() == nil {
		if header.GetNonce() == 1 { // first block after genesis
			if bytes.Equal(header.GetPrevHash(), blockChain.GetGenesisHeaderHash()) {
				// TODO add genesis block verification
				return true
			}

			log.Info(fmt.Sprintf("hash not match: local block hash is empty and node received block with previous hash %s\n",
				toB64(header.GetPrevHash())))

			return false
		}

		log.Info(fmt.Sprintf("nonce not match: local block nonce is 0 and node received block with nonce %d\n",
			header.GetNonce()))

		return false
	}

	if header.GetNonce() != blockChain.GetCurrentBlockHeader().GetNonce()+1 {
		log.Info(fmt.Sprintf("nonce not match: local block nonce is %d and node received block with nonce %d\n",
			blockChain.GetCurrentBlockHeader().GetNonce(), header.GetNonce()))

		return false
	}

	blockHeader, ok := blockChain.GetCurrentBlockHeader().(*block.Header)
	if !ok {
		log.Error(process.ErrWrongTypeAssertion.Error())
		return false
	}

	prevHeaderHash, err := bp.computeHeaderHash(blockHeader)
	if err != nil {
		log.Info(err.Error())
		return false
	}

	if !bytes.Equal(header.GetPrevHash(), prevHeaderHash) {
		log.Info(fmt.Sprintf("hash not match: local block hash is %s and node received block with previous hash %s\n",
			toB64(prevHeaderHash), toB64(header.GetPrevHash())))

		return false
	}

	if body != nil {
		// TODO add body verification here
	}

	return true
}

// verifyStateRoot verifies the state root hash given as parameter against the
// Merkle trie root hash stored for accounts and returns if equal or not
func (bp *baseProcessor) verifyStateRoot(rootHash []byte) bool {
	return bytes.Equal(bp.accounts.RootHash(), rootHash)
}

// getRootHash returns the accounts merkle tree root hash
func (bp *baseProcessor) getRootHash() []byte {
	return bp.accounts.RootHash()
}

func (bp *baseProcessor) validateHeader(chainHandler data.ChainHandler, headerHandler data.HeaderHandler) error {
	// basic validation was already done on interceptor
	if chainHandler.GetCurrentBlockHeader() == nil {
		if !bp.isFirstBlockInEpoch(headerHandler) {
			return process.ErrWrongNonceInBlock
		}
	} else {
		if bp.isCorrectNonce(chainHandler.GetCurrentBlockHeader().GetNonce(), headerHandler.GetNonce()) {
			return process.ErrWrongNonceInBlock
		}

		if !bytes.Equal(headerHandler.GetPrevHash(), chainHandler.GetCurrentBlockHeaderHash()) {

			log.Info(fmt.Sprintf(
				"header.Nonce = %d has header.PrevHash = %s and blockChain.CurrentBlockHeader.Nonce = %d has blockChain.CurrentBlockHeaderHash = %s\n",
				headerHandler.GetNonce(),
				toB64(headerHandler.GetPrevHash()),
				chainHandler.GetCurrentBlockHeader().GetNonce(),
				toB64(chainHandler.GetCurrentBlockHeaderHash())))

			return process.ErrInvalidBlockHash
		}
	}

	return nil
}

func (bp *baseProcessor) isCorrectNonce(currentBlockNonce, receivedBlockNonce uint64) bool {
	return currentBlockNonce+1 != receivedBlockNonce
}

func (bp *baseProcessor) isFirstBlockInEpoch(headerHandler data.HeaderHandler) bool {
	return headerHandler.GetRound() == 0
}

func (bp *baseProcessor) computeHeaderHash(header data.HeaderHandler) ([]byte, error) {
	headerMarsh, err := bp.marshalizer.Marshal(header)
	if err != nil {
		return nil, err
	}

	headerHash := bp.hasher.Compute(string(headerMarsh))

	return headerHash, nil
}

func displayHeader(headerHandler data.HeaderHandler) []*display.LineData {
	lines := make([]*display.LineData, 0)

	lines = append(lines, display.NewLineData(false, []string{
		"Header",
		"Nonce",
		fmt.Sprintf("%d", headerHandler.GetNonce())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Epoch",
		fmt.Sprintf("%d", headerHandler.GetEpoch())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Round",
		fmt.Sprintf("%d", headerHandler.GetRound())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"TimeStamp",
		fmt.Sprintf("%d", headerHandler.GetTimeStamp())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Prev hash",
		toB64(headerHandler.GetPrevHash())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Prev rand seed",
		toB64(headerHandler.GetPrevRandSeed())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Rand seed",
		toB64(headerHandler.GetRandSeed())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Pub keys bitmap",
		toHex(headerHandler.GetPubKeysBitmap())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Signature",
		toB64(headerHandler.GetSignature())}))
	lines = append(lines, display.NewLineData(true, []string{
		"",
		"Root hash",
		toB64(headerHandler.GetRootHash())}))
	return lines
}

func toHex(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return "0x" + hex.EncodeToString(buff)
}

func toB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return base64.StdEncoding.EncodeToString(buff)
}

func checkForNils(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
	if blockChain == nil {
		return process.ErrNilBlockChain
	}
	if header == nil {
		return process.ErrNilBlockHeader
	}
	if body == nil {
		return process.ErrNilMiniBlocks
	}
	return nil
}

// checkProcessorNilParameters will check the imput parameters for nil values
func checkProcessorNilParameters(
	accounts state.AccountsAdapter,
	dataPool dataRetriever.PoolsHolder,
	forkDetector process.ForkDetector,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	store dataRetriever.StorageService,
) error {
	if accounts == nil {
		return process.ErrNilAccountsAdapter
	}
	if dataPool == nil {
		return process.ErrNilDataPoolHolder
	}
	if forkDetector == nil {
		return process.ErrNilForkDetector
	}
	if hasher == nil {
		return process.ErrNilHasher
	}
	if marshalizer == nil {
		return process.ErrNilMarshalizer
	}
	if store == nil {
		return process.ErrNilTxStorage
	}

	return nil
}
