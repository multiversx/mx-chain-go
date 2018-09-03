package ex03ONetTesting

import "os"
import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
)

func Main() {
	app := cli.NewApp()

	cmdPrint := func(ctx *cli.Context) {
		fmt.Printf("nothing %v", *ctx)
	}

	app.Commands = []cli.Command{
		{
			Name:   "time",
			Action: cmdPrint,
		},
		{
			Name:   "counter",
			Action: cmdPrint,
		},
	}

	app.Run(os.Args)
}
