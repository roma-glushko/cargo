package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

func Run(cCtx *cli.Context) error {
	fmt.Println("added task: ", cCtx.Args().First())

	return nil
}
