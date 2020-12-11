package types

import (
	"encoding/json"
	"time"
)

const (
	// MessageSaveBlock -
	MessageSaveBlock = 0
	// MessageRevertBlock -
	MessageRevertBlock = 1
	// MessageSaveRounds -
	MessageSaveRounds = 2
	// MessageTps -
	MessageTps = 3
	// MessageValidatorsPubKeys -
	MessageValidatorsPubKeys = 4
	// MessageValidatorsRating -
	MessageValidatorsRating = 5
	// MessageAccounts -
	MessageAccounts = 6
)

// DataToSend -
type DataToSend struct {
	MessageType int             `json:"messageType"`
	Info        json.RawMessage `json:"info"`
}

// Block -
type Block struct {
	Nonce                 uint64        `json:"nonce"`
	Round                 uint64        `json:"round"`
	Epoch                 uint32        `json:"epoch"`
	Hash                  string        `json:"hash"`
	MiniBlocksHashes      []string      `json:"miniBlocksHashes"`
	NotarizedBlocksHashes []string      `json:"notarizedBlocksHashes"`
	Proposer              uint64        `json:"proposer"`
	Validators            []uint64      `json:"validators"`
	PubKeyBitmap          string        `json:"pubKeyBitmap"`
	Size                  int64         `json:"size"`
	SizeTxs               int64         `json:"sizeTxs"`
	Timestamp             time.Duration `json:"timestamp"`
	StateRootHash         string        `json:"stateRootHash"`
	PrevHash              string        `json:"prevHash"`
	ShardID               uint32        `json:"shardId"`
	TxCount               uint32        `json:"txCount"`
}

// ShardValidatorsKeys -
type ShardValidatorsKeys struct {
	Epoch      uint32   `json:"epoch"`
	ShardID    uint32   `json:"shardID"`
	PublicKeys []string `json:"publicKeys"`
}
