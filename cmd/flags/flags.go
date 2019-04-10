package flags

import "github.com/urfave/cli"

var (
	// GenesisFile defines a flag for the path of the bootstrapping file.
	GenesisFile = cli.StringFlag{
		Name:  "genesis-file",
		Usage: "The node will extract bootstrapping info from the genesis.json",
		Value: "genesis.json",
	}
	// PrivateKey defines a flag for the path of the private key used when starting the node
	PrivateKey = cli.StringFlag{
		Name:  "private-key",
		Usage: "Private key that the node will load on startup and will sign transactions - temporary until we have a wallet that can do that",
		Value: "",
	}
	// WithUI defines a flag for choosing the option of starting with/without UI. If false, the node will start automatically
	WithUI = cli.BoolTFlag{
		Name:  "with-ui",
		Usage: "If true, the application will be accompanied by a UI. The node will have to be manually started from the UI",
	}
	// Port defines a flag for setting the port on which the node will listen for connections
	Port = cli.IntFlag{
		Name:  "port",
		Usage: "Port number on which the application will start",
		Value: 32000,
	}
	// ProfileMode defines a flag for profiling the binary
	ProfileMode = cli.StringFlag{
		Name:  "profile-mode",
		Usage: "Profiling mode. Available options: cpu, mem, mutex, block",
		Value: "",
	}
	// P2PSeed defines a flag to be used as a seed when generating P2P credentials. Useful for seed nodes.
	P2PSeed = cli.StringFlag{
		Name:  "p2p-seed",
		Usage: "P2P seed will be used when generating credentials for p2p component. Can be any string.",
		Value: "seed",
	}
)
