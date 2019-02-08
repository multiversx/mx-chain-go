package blockchain

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// Metachain holds the block information for the metachain shard
type Metachain struct {
	ChainStorer
}

// NewMetachain will initialize a new metachain instance
func NewMetachain(metaBlockUnit storage.Storer) (*Metachain, error) {
	if metaBlockUnit == nil {
		return nil, ErrMetaBlockUnitNil
	}
	return &Metachain{
		ChainStorer: ChainStorer{
			chain: map[UnitType]storage.Storer{
				MetaBlockUnit: metaBlockUnit,
			},
		},
	}, nil
}