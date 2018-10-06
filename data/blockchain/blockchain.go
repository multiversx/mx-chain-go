package blockchain

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/synctime"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

type BlockChain struct {
	Blocks   []block.Block
	SyncTime *synctime.SyncTime
	DoLog    bool
}

func New(blocks []block.Block, syncTime *synctime.SyncTime, doLog bool) BlockChain {
	blockChain := BlockChain{blocks, syncTime, doLog}
	return blockChain
}

func (bc *BlockChain) CheckIfBlockIsValid(receivedBlock *block.Block) bool {

	currentBlock := bc.GetCurrentBlock()

	if currentBlock == nil {
		if receivedBlock.GetNonce() == 0 {
			if receivedBlock.PrevHash != "" {
				bc.Log(fmt.Sprintf("Hash not match: local block hash is %s and node received block with previous hash %s", currentBlock.GetHash(), receivedBlock.GetHash()))
				return false
			}
		} else if receivedBlock.GetNonce() > 0 { // to resolve the situation when a node comes later in the network and it have not implemented the bootstrap mechanism (he will accept the first block received)
			bc.Log(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d", -1, receivedBlock.GetNonce()))
			bc.Log(fmt.Sprintf("\n"+bc.SyncTime.GetFormatedCurrentTime()+">>>>>>>>>>>>>>>>>>>> ACCEPTED BLOCK WITH NONCE %d BECAUSE BOOSTRAP IS NOT IMPLEMENTED YET <<<<<<<<<<<<<<<<<<<<\n", receivedBlock.GetNonce()))
		}

		return true
	}

	if receivedBlock.GetNonce() < currentBlock.GetNonce()+1 {
		bc.Log(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d", currentBlock.GetNonce(), receivedBlock.GetNonce()))
		return false

	} else if receivedBlock.GetNonce() == currentBlock.GetNonce()+1 {
		if receivedBlock.GetPrevHash() != currentBlock.GetHash() {
			bc.Log(fmt.Sprintf("Hash not match: local block hash is %s and node received block with previous hash %s", currentBlock.GetHash(), receivedBlock.GetHash()))
			return false
		}
	} else if receivedBlock.GetNonce() > currentBlock.GetNonce()+1 { // to resolve the situation when a node misses some Blocks and it have not implemented the bootstrap mechanism (he will accept the next block received)
		bc.Log(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d", currentBlock.GetNonce(), receivedBlock.GetNonce()))
		bc.Log(fmt.Sprintf("\n"+bc.SyncTime.GetFormatedCurrentTime()+">>>>>>>>>>>>>>>>>>>> ACCEPTED BLOCK WITH NONCE %d BECAUSE BOOSTRAP IS NOT IMPLEMENTED YET <<<<<<<<<<<<<<<<<<<<\n", receivedBlock.GetNonce()))
	}

	return true
}

func (bc *BlockChain) AddBlock(block block.Block) {
	bc.Blocks = append(bc.Blocks, block)
}

func (bc *BlockChain) GetCurrentBlock() *block.Block {
	if len(bc.Blocks) == 0 {
		return nil
	}

	return &bc.Blocks[len(bc.Blocks)-1]
}

func (bc *BlockChain) Log(message string) {
	if bc.DoLog {
		fmt.Printf(message + "\n")
	}
}
