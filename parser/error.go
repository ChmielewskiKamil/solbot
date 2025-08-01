package parser

import (
	"fmt"
	"solbot/token"
	"strings"
)

type Error struct {
	Filename string
	Line     int
	Column   int
	Msg      string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.Filename, e.Line, e.Column, e.Msg)
}

type ErrorList []Error

func (e ErrorList) Error() string {
	if len(e) == 0 {
		return "no errors"
	}

	var b strings.Builder
	for i, err := range e {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(err.Error())
	}
	return b.String()
}

func (p *parser) addError(pos token.Pos, msg string) {
	line, col := p.file.GetLineAndColumn(pos)
	err := Error{
		Filename: p.file.Name(),
		Line:     line,
		Column:   col,
		Msg:      msg,
	}
	p.errors = append(p.errors, err)
}
