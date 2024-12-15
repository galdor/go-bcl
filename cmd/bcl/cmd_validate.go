package main

import (
	"go.n16f.net/bcl"
	"go.n16f.net/program"
)

func cmdValidate(p *program.Program) {
	source, data := readFileOrStdin(p.OptionalArgumentValue("path"))

	if _, err := bcl.Parse(data, source); err != nil {
		p.Fatal("cannot parse document: %v", err)
	}
}
