package common

import "github.com/urfave/cli/v2"

type Generator interface {
	GenFile(ctx *cli.Context) (data []byte, targetDir string, targetFile string)
}
