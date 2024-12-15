package main

import (
	"io"
	"os"
)

func readFileOrStdin(filePath *string) (string, []byte) {
	var source string
	var data []byte
	var err error

	if filePath == nil || *filePath == "-" {
		source = "<stdin>"

		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			p.Fatal("cannot read stdin: %v", err)
		}
	} else {
		source = *filePath

		data, err = os.ReadFile(*filePath)
		if err != nil {
			p.Fatal("cannot read %q: %v", *filePath, err)
		}
	}

	return source, data
}
