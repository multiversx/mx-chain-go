package flags

import "github.com/urfave/cli"

var (
	// GenesisFile defines a flag for the path of the bootstrapping file.
	GenesisFile = cli.StringFlag{
		Name:  "genesis-file",
		Usage: "The node will extract bootstrapping info from the genesis.json",
		Value: "genesis.json",
	}
	WithUI = cli.BoolTFlag{
		Name:  "with-ui",
		Usage: "If true, the application will be accompanied by a UI. The node will have to be manually started from the UI",
	}
	Port = cli.IntFlag{
		Name:  "port",
		Usage: "Port number on which the application will start",
		Value: 4000,
	}
	MaxAllowedPeers = cli.IntFlag{
		Name:  "max-allowed-peers",
		Usage: "Maximum connections the user is willing to accept",
		Value: 4,
	}
)
