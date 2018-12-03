package main

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/cmd/flags"
	"github.com/ElrondNetwork/elrond-go-sandbox/cmd/templates"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	log := logger.NewDefaultLogger()
	cli.AppHelpTemplate = templates.BootNodeHelp
	app.Name = "BootNode CLI App"
	app.Usage = "This is the entrypoint for starting a new bootstrap node - the app will start after the genessis timestamp"
	app.Flags = []cli.Flag{
		flags.GenesisFile,
	}
	app.Action = func(c *cli.Context) error {
		err := startNode(c, log)
		if err != nil {
			log.Panic("Could not start node", err)
			return err
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		panic("woot")
	}
}

func startNode(c *cli.Context, log *logger.Logger) error {
	(*log).Info("Starting node...")
	return nil
}