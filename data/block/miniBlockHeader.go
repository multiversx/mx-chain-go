package block

import "github.com/ElrondNetwork/elrond-go/data"

// GetTypeInt32 gets the miniBlock type
func (m *MiniBlockHeader) GetTypeInt32() int32 {
	if m == nil {
		return -1
	}

	return int32(m.Type)
}

// SetHash sets the miniBlock hash
func (m *MiniBlockHeader) SetHash(hash []byte) error {
	if m == nil {
		return data.ErrNilPointerReceiver
	}

	m.Hash = hash

	return nil
}

// SetSenderShardID sets the miniBlock sender shardID
func (m *MiniBlockHeader) SetSenderShardID(shardID uint32) error {
	if m == nil {
		return data.ErrNilPointerReceiver
	}

	m.SenderShardID = shardID

	return nil
}

// SetReceiverShardID sets the miniBlock receiver ShardID
func (m *MiniBlockHeader) SetReceiverShardID(shardID uint32) error {
	if m == nil {
		return data.ErrNilPointerReceiver
	}

	m.ReceiverShardID = shardID

	return nil
}

// SetTxCount sets the miniBlock txs count
func (m *MiniBlockHeader) SetTxCount(count uint32) error {
	if m == nil {
		return data.ErrNilPointerReceiver
	}

	m.TxCount = count

	return nil
}

// SetTypeInt32 sets the miniBlock type
func (m *MiniBlockHeader) SetTypeInt32(t int32) error {
	if m == nil {
		return data.ErrNilPointerReceiver
	}

	m.Type = Type(t)

	return nil
}

// SetReserved sets the miniBlock reserved field
func (m *MiniBlockHeader) SetReserved(reserved []byte) error {
	if m == nil {
		return data.ErrNilPointerReceiver
	}

	m.Reserved = reserved

	return nil
}

// ShallowClone returns the miniBlockHeader swallow clone
func (m *MiniBlockHeader) ShallowClone() data.MiniBlockHeaderHandler {
	if m == nil {
		return nil
	}

	mbhCopy := *m

	return &mbhCopy
}
