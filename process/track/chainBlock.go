package track

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/data"
)

type headerInfo struct {
	hash   string
	header data.HeaderHandler
}

type ChainBlock struct {
	headerInfo headerInfo
	next       map[string]*ChainBlock // key is block header hash
}

func NewChainBlock(header data.HeaderHandler, hash string) *ChainBlock {
	return &ChainBlock{
		headerInfo: headerInfo{
			hash:   hash,
			header: header,
		},
		next: make(map[string]*ChainBlock),
	}
}

func (cb *ChainBlock) Add(header data.HeaderHandler, hash string) bool {
	nextChainBlock := &ChainBlock{
		headerInfo: headerInfo{
			hash:   hash,
			header: header,
		},
		next: make(map[string]*ChainBlock),
	}

	if bytes.Equal(header.GetPrevHash(), []byte(cb.headerInfo.hash)) {
		cb.next[hash] = nextChainBlock
		return true
	}

	for _, nextBlock := range cb.next {
		if nextBlock.Add(header, hash) {
			return true
		}
	}

	return false
}

func (cb *ChainBlock) LongestChain(currentChain []data.HeaderHandler) ([]data.HeaderHandler, int) {
	longestHeight := len(currentChain)
	var longestChain = currentChain

	for _, nextBlock := range cb.next {
		newChain := append(currentChain, nextBlock.headerInfo.header)
		chain, length := nextBlock.LongestChain(newChain)
		if length > longestHeight {
			longestHeight = length
			longestChain = chain
		}
	}

	return longestChain, longestHeight
}
