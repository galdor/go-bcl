package main

import (
	"go.n16f.net/bcl"
	"go.n16f.net/pp"
	"go.n16f.net/program"
)

func cmdFormat(p *program.Program) {
	source, data := readFileOrStdin(p.OptionalArgumentValue("path"))

	doc, err := bcl.Parse(data, source)
	if err != nil {
		p.Fatal("cannot parse document:\n%v", err)
	}

	// TODO
	pp.Print(doc, "document")
}
