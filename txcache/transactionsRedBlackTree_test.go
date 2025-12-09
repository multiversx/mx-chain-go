package txcache

import (
	"container/list"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func Test_TransactionsComparator(t *testing.T) {
	t.Parallel()

	t.Run("should return 1 if tx1.Nonce > tx2.Nonce", func(t *testing.T) {
		t.Parallel()

		tx1 := &list.Element{
			Value: &WrappedTransaction{
				Tx: &transaction.Transaction{
					Nonce: 10,
				},
			},
		}

		tx2 := &list.Element{
			Value: &WrappedTransaction{
				Tx: &transaction.Transaction{
					Nonce: 9,
				},
			},
		}

		require.Equal(t, 1, TransactionsComparator(tx1, tx2))
	})

	t.Run("should return 1 if tx1.GasPrice < tx2.GasPrice", func(t *testing.T) {
		t.Parallel()

		tx1 := &list.Element{
			Value: &WrappedTransaction{
				Tx: &transaction.Transaction{
					Nonce:    10,
					GasPrice: oneMilion,
				},
			},
		}

		tx2 := &list.Element{
			Value: &WrappedTransaction{
				Tx: &transaction.Transaction{
					Nonce:    10,
					GasPrice: oneBillion,
				},
			},
		}

		require.Equal(t, 1, TransactionsComparator(tx1, tx2))
	})

	t.Run("should return 1 if tx1.TxHash > tx2.TxHash", func(t *testing.T) {
		t.Parallel()

		tx1 := &list.Element{
			Value: &WrappedTransaction{
				Tx: &transaction.Transaction{
					Nonce:    10,
					GasPrice: oneBillion,
				},
				TxHash: []byte("txHash2"),
			},
		}

		tx2 := &list.Element{
			Value: &WrappedTransaction{
				Tx: &transaction.Transaction{
					Nonce:    10,
					GasPrice: oneBillion,
				},
				TxHash: []byte("txHash1"),
			},
		}

		require.Equal(t, 1, TransactionsComparator(tx1, tx2))
	})

	t.Run("should return -1 if tx1.Nonce < tx2.Nonce", func(t *testing.T) {
		t.Parallel()

		tx1 := &list.Element{
			Value: &WrappedTransaction{
				Tx: &transaction.Transaction{
					Nonce: 10,
				},
			},
		}

		tx2 := &list.Element{
			Value: &WrappedTransaction{
				Tx: &transaction.Transaction{
					Nonce: 11,
				},
			},
		}

		require.Equal(t, -1, TransactionsComparator(tx1, tx2))
	})

	t.Run("should return -1 if tx1.GasPrice > tx2.GasPrice", func(t *testing.T) {
		t.Parallel()

		tx1 := &list.Element{
			Value: &WrappedTransaction{
				Tx: &transaction.Transaction{
					Nonce:    10,
					GasPrice: oneBillion,
				},
			},
		}

		tx2 := &list.Element{
			Value: &WrappedTransaction{
				Tx: &transaction.Transaction{
					Nonce:    10,
					GasPrice: oneMilion,
				},
			},
		}

		require.Equal(t, -1, TransactionsComparator(tx1, tx2))
	})

	t.Run("should return -1 if tx1.TxHash < tx2.TxHash", func(t *testing.T) {
		t.Parallel()

		tx1 := &list.Element{
			Value: &WrappedTransaction{
				Tx: &transaction.Transaction{
					Nonce:    10,
					GasPrice: oneBillion,
				},
				TxHash: []byte("txHash1"),
			},
		}

		tx2 := &list.Element{
			Value: &WrappedTransaction{
				Tx: &transaction.Transaction{
					Nonce:    10,
					GasPrice: oneBillion,
				},
				TxHash: []byte("txHash2"),
			},
		}

		require.Equal(t, -1, TransactionsComparator(tx1, tx2))
	})

	t.Run("should return 0 when tx1 == tx2", func(t *testing.T) {
		t.Parallel()

		tx1 := &list.Element{
			Value: &WrappedTransaction{
				Tx: &transaction.Transaction{
					Nonce:    10,
					GasPrice: oneBillion,
				},
				TxHash: []byte("txHash1"),
			},
		}

		tx2 := &list.Element{
			Value: &WrappedTransaction{
				Tx: &transaction.Transaction{
					Nonce:    10,
					GasPrice: oneBillion,
				},
				TxHash: []byte("txHash1"),
			},
		}

		require.Equal(t, 0, TransactionsComparator(tx1, tx2))
	})
}

func TestTxRedBlackTree_Operations(t *testing.T) {
	t.Parallel()

	trb := NewTransactionsRedBlackTree()
	require.NotNil(t, trb)

	tx1 := &list.Element{
		Value: &WrappedTransaction{
			Tx: &transaction.Transaction{
				Nonce:    10,
				GasPrice: oneBillion,
			},
			TxHash: []byte("txHash1"),
		},
	}

	trb.Insert(tx1)
	err := trb.contains(tx1)
	require.Equal(t, errItemAlreadyInCache, err)
	require.Equal(t, uint64(1), trb.Len())

	trb.Remove(tx1)
	err = trb.contains(tx1)
	require.NoError(t, err)
	require.Equal(t, uint64(0), trb.Len())

	trb.Insert(tx1)
	err = trb.contains(tx1)
	require.Equal(t, errItemAlreadyInCache, err)
	require.Equal(t, uint64(1), trb.Len())

	tx2 := &list.Element{
		Value: &WrappedTransaction{
			Tx: &transaction.Transaction{
				Nonce:    11,
				GasPrice: oneBillion,
			},
			TxHash: []byte("txHash2"),
		},
	}

	// tx2 should be placed after tx1, because tx2.Nonce > tx1.Nonce
	parentElement, err := trb.FindInsertionPlace(tx2)
	require.Nil(t, err)
	require.Equal(t, tx1, parentElement)

	trb.Insert(tx2)
	err = trb.contains(tx2)
	require.Equal(t, errItemAlreadyInCache, err)
	require.Equal(t, uint64(2), trb.Len())

	tx3 := &list.Element{
		Value: &WrappedTransaction{
			Tx: &transaction.Transaction{
				Nonce:    0,
				GasPrice: oneBillion,
			},
			TxHash: []byte("txHash3"),
		},
	}

	// tx3 should be placed before tx1, because its nonce is the smallest one
	parentElement, err = trb.FindInsertionPlace(tx3)
	require.Nil(t, err)
	require.Nil(t, parentElement)

	trb.Insert(tx3)
	err = trb.contains(tx3)
	require.Equal(t, errItemAlreadyInCache, err)

	// in case of equal nonce, the tx with higher gas price should be parent of the other tx
	tx4 := &list.Element{
		Value: &WrappedTransaction{
			Tx: &transaction.Transaction{
				Nonce:    11,
				GasPrice: oneMilion,
			},
			TxHash: []byte("txHash4"),
		},
	}

	parentElement, err = trb.FindInsertionPlace(tx4)
	require.Nil(t, err)
	require.Equal(t, tx2, parentElement)

	trb.Insert(tx4)
	err = trb.contains(tx4)
	require.Equal(t, errItemAlreadyInCache, err)

	// in case of equal nonce and equal gas price, the tx with lower txHash should be parent of the other tx
	tx5 := &list.Element{
		Value: &WrappedTransaction{
			Tx: &transaction.Transaction{
				Nonce:    11,
				GasPrice: oneMilion,
			},
			TxHash: []byte("txHash5"),
		},
	}

	parentElement, err = trb.FindInsertionPlace(tx5)
	require.Nil(t, err)
	require.Equal(t, tx4, parentElement)
}
