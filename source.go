package bcl

import "fmt"

type Point struct {
	Offset int
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
