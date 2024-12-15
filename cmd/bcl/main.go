package main

import (
	"go.n16f.net/program"
)

var p *program.Program

func main() {
	var c *program.Command

	p = program.NewProgram("bcl", "utilities for the BCL language")

	c = p.AddCommand("format", "parse a BCL file and print it", cmdFormat)
	c.AddOptionalArgument("path", "the path of the file")

	c = p.AddCommand("validate", "parse a BCL file", cmdValidate)
	c.AddOptionalArgument("path", "the path of the file")

	p.ParseCommandLine()
	p.Run()
}
