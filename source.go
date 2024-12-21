package bcl

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"
)

type ParseErrors []error

func (errs ParseErrors) Error() string {
	var buf bytes.Buffer

	for _, err := range errs {
		fmt.Fprintln(&buf, err)

		var syntaxErr *SyntaxError
		if errors.As(err, &syntaxErr) {
			syntaxErr.Location.PrintSource(&buf, syntaxErr.Lines, 2, "  ")
		}
	}

	return strings.TrimRight(buf.String(), "\n")
}

type SyntaxError struct {
	Source      string
	Lines       []string
	Location    Span
	Description string
}

func (err *SyntaxError) Error() string {
	msg := err.Location.String() + ": " + err.Description
	if err.Source != "" {
		msg = err.Source + ":" + msg
	}
	return msg
}

type Point struct {
	Offset int // counted in characters (runes), not in bytes
	Line   int
	Column int
}

func (p Point) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

func (p Point) Equal(p2 Point) bool {
	return p.Offset == p2.Offset
}

type Span struct {
	Start Point
	End   Point
}

func NewSpanAt(start Point, len int) Span {
	end := Point{
		Offset: start.Offset + len - 1,
		Line:   start.Line,
		Column: start.Column + len - 1,
	}

	return Span{
		Start: start,
		End:   end,
	}
}

func (s Span) Len() int {
	return s.End.Offset - s.Start.Offset + 1
}

func (s Span) String() string {
	if p, ok := s.Point(); ok {
		return p.String()
	} else {
		return fmt.Sprintf("%v-%v", s.Start, s.End)
	}
}

func (s Span) Point() (Point, bool) {
	if s.Start.Equal(s.End) {
		return s.Start, true
	}

	return Point{}, false
}

func (s Span) PrintSource(w io.Writer, lines []string, context int, indent string) {
	nbLineDigits := int(math.Floor(math.Log10(float64(len(lines)))) + 1)
	offset := nbLineDigits + 3

	printLine := func(l int) {
		fmt.Fprintf(w, "%s%*d │ ", indent, nbLineDigits, l+1)
		fmt.Fprintln(w, lines[l])
	}

	lstart := s.Start.Line - 1
	lend := s.End.Line - 1

	for l := max(lstart-context, 0); l < lstart; l++ {
		printLine(l)
	}

	for l := lstart; l <= lend; l++ {
		printLine(l)

		line := lines[l]

		cstart := 0
		if l == lstart {
			cstart = s.Start.Column - 1
		}

		cend := len(line)
		if l == lend {
			cend = s.End.Column
		}

		fmt.Fprint(w, indent)
		for c := 0; c < offset; c++ {
			fmt.Fprint(w, " ")
		}

		for c := 0; c < len(line); c++ {
			char := ' '
			if c >= cstart && c < cend {
				char = '^'
			}

			fmt.Fprint(w, string(char))
		}

		// The final point can appear just after the end of the line
		if cend >= len(line) {
			fmt.Fprint(w, string('^'))
		}

		fmt.Fprintln(w)
	}

	for l := lend + 1; l < min(lend+context+1, len(lines)); l++ {
		printLine(l)
	}
}

var lineRE = regexp.MustCompile("\r?\n")

func splitLines(data []byte) []string {
	lines := lineRE.Split(string(data), -1)
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
