package blockchain

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/davecgh/go-spew/spew"
)

type BlockChain struct {
	DoLog    bool
	Blocks   []block.Block
	SyncTime *ntp.SyncTime
}

func New(blocks []block.Block, syncTime *ntp.SyncTime, doLog bool) BlockChain {
	blockChain := BlockChain{DoLog: doLog, Blocks: blocks, SyncTime: syncTime}
	return blockChain
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

func (bc *BlockChain) CheckIfBlockIsValid(receivedBlock *block.Block) bool {

	currentBlock := bc.GetCurrentBlock()

	if currentBlock == nil {
		if receivedBlock.Nonce == 0 {
			if receivedBlock.PrevHash != "" {
				bc.Log(fmt.Sprintf("Hash not match: local block hash is %s and node received block with previous hash %s", currentBlock.Hash, receivedBlock.Hash))
				return false
			}
		} else if receivedBlock.Nonce > 0 { // to resolve the situation when a node comes later in the network and it have not implemented the bootstrap mechanism (he will accept the first block received)
			bc.Log(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d", -1, receivedBlock.Nonce))
			bc.Log(fmt.Sprintf("\n"+bc.SyncTime.GetFormatedCurrentTime()+">>>>>>>>>>>>>>>>>>>> ACCEPTED BLOCK WITH NONCE %d BECAUSE BOOSTRAP IS NOT IMPLEMENTED YET <<<<<<<<<<<<<<<<<<<<\n", receivedBlock.Nonce))
		}

		return true
	}

	if receivedBlock.Nonce < currentBlock.Nonce+1 {
		bc.Log(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d", currentBlock.Nonce, receivedBlock.Nonce))
		return false

	} else if receivedBlock.Nonce == currentBlock.Nonce+1 {
		if receivedBlock.PrevHash != currentBlock.Hash {
			bc.Log(fmt.Sprintf("Hash not match: local block hash is %s and node received block with previous hash %s", currentBlock.Hash, receivedBlock.Hash))
			return false
		}
	} else if receivedBlock.Nonce > currentBlock.Nonce+1 { // to resolve the situation when a node misses some Blocks and it have not implemented the bootstrap mechanism (he will accept the next block received)
		bc.Log(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d", currentBlock.Nonce, receivedBlock.Nonce))
		bc.Log(fmt.Sprintf("\n"+bc.SyncTime.GetFormatedCurrentTime()+">>>>>>>>>>>>>>>>>>>> ACCEPTED BLOCK WITH NONCE %d BECAUSE BOOSTRAP IS NOT IMPLEMENTED YET <<<<<<<<<<<<<<<<<<<<\n", receivedBlock.Nonce))
	}

	return true
}

func (bc *BlockChain) Log(message string) {
	if bc.DoLog {
		fmt.Printf(message + "\n")
	}
}

func (bc *BlockChain) Print() {
	spew.Dump(bc)
}
