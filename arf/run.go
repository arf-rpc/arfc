package arf

import (
	"fmt"
	"github.com/arf-rpc/arfc/arf/common"
	"github.com/arf-rpc/arfc/arf/golang"
	"github.com/arf-rpc/arfc/arf/ruby"
	"github.com/arf-rpc/arfc/output"
	"github.com/arf-rpc/idl"
	"github.com/arf-rpc/idl/ast"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"strings"
)

type outputFile struct {
	data       []byte
	outputDir  string
	outputFile string
}

func Run(c *cli.Context) error {
	inputArg := c.String("input")

	errored := false
	fs, err := idl.ParseFile(inputArg, func(err error) {
		output.Errorf("%s", err.Error())
		errored = true
	})
	if errored {
		return fmt.Errorf("errors found parsing input")
	}
	if err != nil {
		return fmt.Errorf("Error parsing input: %w\n", err)
	}

	rawLang := c.String("lang")
	lang := strings.ToLower(rawLang)
	var makeGen func(tree *ast.PackageTree) common.Generator

	switch lang {
	case "ruby":
		if c.IsSet("golang-package") {
			output.Warnf("Providing golang-package with lang Ruby has no effect")
		}
		if c.IsSet("golang-module") {
			output.Warnf("Providing golang-module with lang Ruby has no effect")
		}
		makeGen = ruby.NewGenerator
	case "go", "golang":
		if c.IsSet("ruby-module") {
			output.Warnf("Providing ruby-module with lang Goland has no effect")
		}
		if c.IsSet("ruby-flat") {
			output.Warnf("Providing ruby-flat with lang Goland has no effect")
		}
		makeGen = golang.NewGenerator
	default:
		output.Errorf("unknown output language `%s': only 'go'/'golang', and 'ruby' are supported.", lang)
	}
	var outputs []*outputFile

	for _, t := range fs.Packages {
		gen := makeGen(t)
		data, targetDir, targetFile := gen.GenFile(c)
		outputs = append(outputs, &outputFile{
			data:       data,
			outputDir:  targetDir,
			outputFile: targetFile,
		})
	}

	for _, v := range outputs {
		if err = os.MkdirAll(v.outputDir, os.ModePerm); err != nil {
			output.Errorf("Failed creating output directory `%s`: %s", v.outputDir, err)
		}
	}

	for _, v := range outputs {
		if err = os.WriteFile(filepath.Join(v.outputDir, v.outputFile), v.data, os.ModePerm); err != nil {
			output.Errorf("Failed writing output file `%s`: %s", v.outputFile, err)
		}
	}

	return nil
}
