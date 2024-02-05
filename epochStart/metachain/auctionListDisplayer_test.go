package metachain

import (
	"math"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

func createDisplayerArgs() ArgsAuctionListDisplayer {
	return ArgsAuctionListDisplayer{
		TableDisplayHandler:      NewTableDisplayer(),
		ValidatorPubKeyConverter: &testscommon.PubkeyConverterMock{},
		AddressPubKeyConverter:   &testscommon.PubkeyConverterMock{},
		AuctionConfig:            createSoftAuctionConfig(),
		Denomination:             0,
	}
}

func TestNewAuctionListDisplayer(t *testing.T) {
	t.Parallel()

	t.Run("invalid auction config", func(t *testing.T) {
		args := createDisplayerArgs()
		args.AuctionConfig.MaxNumberOfIterations = 0
		ald, err := NewAuctionListDisplayer(args)
		require.Nil(t, ald)
		requireInvalidValueError(t, err, "for max number of iterations")
	})

	t.Run("should work", func(t *testing.T) {
		args := createDisplayerArgs()
		ald, err := NewAuctionListDisplayer(args)
		require.Nil(t, err)
		require.False(t, ald.IsInterfaceNil())
	})
}

func TestAuctionListDisplayer_DisplayOwnersData(t *testing.T) {
	t.Parallel()

	_ = logger.SetLogLevel("*:DEBUG")
	defer func() {
		_ = logger.SetLogLevel("*:INFO")
	}()

	owner := []byte("owner")
	validator := &state.ValidatorInfo{PublicKey: []byte("pubKey")}
	wasDisplayCalled := false

	args := createDisplayerArgs()
	args.AddressPubKeyConverter = &testscommon.PubkeyConverterStub{
		SilentEncodeCalled: func(pkBytes []byte, log core.Logger) string {
			require.Equal(t, owner, pkBytes)
			return "ownerEncoded"
		},
	}
	args.ValidatorPubKeyConverter = &testscommon.PubkeyConverterStub{
		SilentEncodeCalled: func(pkBytes []byte, log core.Logger) string {
			require.Equal(t, validator.PublicKey, pkBytes)
			return "pubKeyEncoded"
		},
	}
	args.TableDisplayHandler = &testscommon.TableDisplayerMock{
		DisplayTableCalled: func(tableHeader []string, lines []*display.LineData, message string) {
			require.Equal(t, []string{
				"Owner",
				"Num staked nodes",
				"Num active nodes",
				"Num auction nodes",
				"Total top up",
				"Top up per node",
				"Auction list nodes",
			}, tableHeader)
			require.Equal(t, "Initial nodes config in auction list", message)
			require.Equal(t, []*display.LineData{
				{
					Values:              []string{"ownerEncoded", "4", "4", "1", "100.0", "25.0", "pubKeyEncoded"},
					HorizontalRuleAfter: false,
				},
			}, lines)

			wasDisplayCalled = true
		},
	}
	ald, _ := NewAuctionListDisplayer(args)

	ownersData := map[string]*OwnerAuctionData{
		"owner": {
			numStakedNodes:           4,
			numActiveNodes:           4,
			numAuctionNodes:          1,
			numQualifiedAuctionNodes: 4,
			totalTopUp:               big.NewInt(100),
			topUpPerNode:             big.NewInt(25),
			qualifiedTopUpPerNode:    big.NewInt(15),
			auctionList:              []state.ValidatorInfoHandler{&state.ValidatorInfo{PublicKey: []byte("pubKey")}},
		},
	}

	ald.DisplayOwnersData(ownersData)
	require.True(t, wasDisplayCalled)
}

func TestAuctionListDisplayer_DisplayOwnersSelectedNodes(t *testing.T) {
	t.Parallel()

	_ = logger.SetLogLevel("*:DEBUG")
	defer func() {
		_ = logger.SetLogLevel("*:INFO")
	}()

	owner := []byte("owner")
	validator := &state.ValidatorInfo{PublicKey: []byte("pubKey")}
	wasDisplayCalled := false

	args := createDisplayerArgs()
	args.AddressPubKeyConverter = &testscommon.PubkeyConverterStub{
		SilentEncodeCalled: func(pkBytes []byte, log core.Logger) string {
			require.Equal(t, owner, pkBytes)
			return "ownerEncoded"
		},
	}
	args.ValidatorPubKeyConverter = &testscommon.PubkeyConverterStub{
		SilentEncodeCalled: func(pkBytes []byte, log core.Logger) string {
			require.Equal(t, validator.PublicKey, pkBytes)
			return "pubKeyEncoded"
		},
	}
	args.TableDisplayHandler = &testscommon.TableDisplayerMock{
		DisplayTableCalled: func(tableHeader []string, lines []*display.LineData, message string) {
			require.Equal(t, []string{
				"Owner",
				"Num staked nodes",
				"TopUp per node",
				"Total top up",
				"Num auction nodes",
				"Num qualified auction nodes",
				"Num active nodes",
				"Qualified top up per node",
				"Selected auction list nodes",
			}, tableHeader)
			require.Equal(t, "Selected nodes config from auction list", message)
			require.Equal(t, []*display.LineData{
				{
					Values:              []string{"ownerEncoded", "4", "25.0", "100.0", "1", "1", "4", "15.0", "pubKeyEncoded"},
					HorizontalRuleAfter: false,
				},
			}, lines)

			wasDisplayCalled = true
		},
	}
	ald, _ := NewAuctionListDisplayer(args)

	ownersData := map[string]*OwnerAuctionData{
		"owner": {
			numStakedNodes:           4,
			numActiveNodes:           4,
			numAuctionNodes:          1,
			numQualifiedAuctionNodes: 1,
			totalTopUp:               big.NewInt(100),
			topUpPerNode:             big.NewInt(25),
			qualifiedTopUpPerNode:    big.NewInt(15),
			auctionList:              []state.ValidatorInfoHandler{&state.ValidatorInfo{PublicKey: []byte("pubKey")}},
		},
	}

	ald.DisplayOwnersSelectedNodes(ownersData)
	require.True(t, wasDisplayCalled)
}

func TestAuctionListDisplayer_DisplayAuctionList(t *testing.T) {
	t.Parallel()

	_ = logger.SetLogLevel("*:DEBUG")
	defer func() {
		_ = logger.SetLogLevel("*:INFO")
	}()

	owner := []byte("owner")
	validator := &state.ValidatorInfo{PublicKey: []byte("pubKey")}
	wasDisplayCalled := false

	args := createDisplayerArgs()
	args.AddressPubKeyConverter = &testscommon.PubkeyConverterStub{
		SilentEncodeCalled: func(pkBytes []byte, log core.Logger) string {
			require.Equal(t, owner, pkBytes)
			return "ownerEncoded"
		},
	}
	args.ValidatorPubKeyConverter = &testscommon.PubkeyConverterStub{
		SilentEncodeCalled: func(pkBytes []byte, log core.Logger) string {
			require.Equal(t, validator.PublicKey, pkBytes)
			return "pubKeyEncoded"
		},
	}
	args.TableDisplayHandler = &testscommon.TableDisplayerMock{
		DisplayTableCalled: func(tableHeader []string, lines []*display.LineData, message string) {
			require.Equal(t, []string{
				"Owner",
				"Registered key",
				"Qualified TopUp per node",
			}, tableHeader)
			require.Equal(t, "Final selected nodes from auction list", message)
			require.Equal(t, []*display.LineData{
				{
					Values:              []string{"ownerEncoded", "pubKeyEncoded", "15.0"},
					HorizontalRuleAfter: true,
				},
			}, lines)

			wasDisplayCalled = true
		},
	}
	ald, _ := NewAuctionListDisplayer(args)

	auctionList := []state.ValidatorInfoHandler{&state.ValidatorInfo{PublicKey: []byte("pubKey")}}
	ownersData := map[string]*OwnerAuctionData{
		"owner": {
			numStakedNodes:           4,
			numActiveNodes:           4,
			numAuctionNodes:          1,
			numQualifiedAuctionNodes: 1,
			totalTopUp:               big.NewInt(100),
			topUpPerNode:             big.NewInt(25),
			qualifiedTopUpPerNode:    big.NewInt(15),
			auctionList:              auctionList,
		},
	}

	ald.DisplayAuctionList(auctionList, ownersData, 1)
	require.True(t, wasDisplayCalled)
}

func TestGetPrettyValue(t *testing.T) {
	t.Parallel()

	require.Equal(t, "1234.0", getPrettyValue(big.NewInt(1234), big.NewInt(1)))
	require.Equal(t, "123.4", getPrettyValue(big.NewInt(1234), big.NewInt(10)))
	require.Equal(t, "12.34", getPrettyValue(big.NewInt(1234), big.NewInt(100)))
	require.Equal(t, "1.234", getPrettyValue(big.NewInt(1234), big.NewInt(1000)))
	require.Equal(t, "0.1234", getPrettyValue(big.NewInt(1234), big.NewInt(10000)))
	require.Equal(t, "0.01234", getPrettyValue(big.NewInt(1234), big.NewInt(100000)))
	require.Equal(t, "0.00123", getPrettyValue(big.NewInt(1234), big.NewInt(1000000)))
	require.Equal(t, "0.00012", getPrettyValue(big.NewInt(1234), big.NewInt(10000000)))
	require.Equal(t, "0.00001", getPrettyValue(big.NewInt(1234), big.NewInt(100000000)))
	require.Equal(t, "0.00000", getPrettyValue(big.NewInt(1234), big.NewInt(1000000000)))
	require.Equal(t, "0.00000", getPrettyValue(big.NewInt(1234), big.NewInt(10000000000)))

	require.Equal(t, "1.0", getPrettyValue(big.NewInt(1), big.NewInt(1)))
	require.Equal(t, "0.1", getPrettyValue(big.NewInt(1), big.NewInt(10)))
	require.Equal(t, "0.01", getPrettyValue(big.NewInt(1), big.NewInt(100)))
	require.Equal(t, "0.001", getPrettyValue(big.NewInt(1), big.NewInt(1000)))
	require.Equal(t, "0.0001", getPrettyValue(big.NewInt(1), big.NewInt(10000)))
	require.Equal(t, "0.00001", getPrettyValue(big.NewInt(1), big.NewInt(100000)))
	require.Equal(t, "0.00000", getPrettyValue(big.NewInt(1), big.NewInt(1000000)))
	require.Equal(t, "0.00000", getPrettyValue(big.NewInt(1), big.NewInt(10000000)))

	oneEGLD := big.NewInt(1000000000000000000)
	denominationEGLD := big.NewInt(int64(math.Pow10(18)))

	require.Equal(t, "0.00000", getPrettyValue(big.NewInt(0), denominationEGLD))
	require.Equal(t, "1.00000", getPrettyValue(oneEGLD, denominationEGLD))
	require.Equal(t, "1.10000", getPrettyValue(big.NewInt(1100000000000000000), denominationEGLD))
	require.Equal(t, "1.10000", getPrettyValue(big.NewInt(1100000000000000001), denominationEGLD))
	require.Equal(t, "1.11000", getPrettyValue(big.NewInt(1110000000000000001), denominationEGLD))
	require.Equal(t, "0.11100", getPrettyValue(big.NewInt(111000000000000001), denominationEGLD))
	require.Equal(t, "0.01110", getPrettyValue(big.NewInt(11100000000000001), denominationEGLD))
	require.Equal(t, "0.00111", getPrettyValue(big.NewInt(1110000000000001), denominationEGLD))
	require.Equal(t, "0.00011", getPrettyValue(big.NewInt(111000000000001), denominationEGLD))
	require.Equal(t, "0.00001", getPrettyValue(big.NewInt(11100000000001), denominationEGLD))
	require.Equal(t, "0.00000", getPrettyValue(big.NewInt(1110000000001), denominationEGLD))
	require.Equal(t, "0.00000", getPrettyValue(big.NewInt(111000000001), denominationEGLD))

	require.Equal(t, "2.00000", getPrettyValue(big.NewInt(0).Mul(oneEGLD, big.NewInt(2)), denominationEGLD))
	require.Equal(t, "20.00000", getPrettyValue(big.NewInt(0).Mul(oneEGLD, big.NewInt(20)), denominationEGLD))
	require.Equal(t, "2000000.00000", getPrettyValue(big.NewInt(0).Mul(oneEGLD, big.NewInt(2000000)), denominationEGLD))

	require.Equal(t, "3.22220", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(2222200000000000000)), denominationEGLD))
	require.Equal(t, "1.22222", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(222220000000000000)), denominationEGLD))
	require.Equal(t, "1.02222", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(22222000000000000)), denominationEGLD))
	require.Equal(t, "1.00222", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(2222200000000000)), denominationEGLD))
	require.Equal(t, "1.00022", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(222220000000000)), denominationEGLD))
	require.Equal(t, "1.00002", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(22222000000000)), denominationEGLD))
	require.Equal(t, "1.00000", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(2222200000000)), denominationEGLD))
	require.Equal(t, "1.00000", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(222220000000)), denominationEGLD))
}
