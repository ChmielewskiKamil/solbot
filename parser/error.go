package parser

import "solparsor/token"

type Error struct {
	Pos token.Position
	Msg string
}

type ErrorList []Error

func (el *ErrorList) Add(pos token.Position, msg string) {
	*el = append(*el, Error{pos, msg})
}
