package parser

import "solbot/token"

type Error struct {
	Pos token.Pos
	Msg string
}

type ErrorList []Error

func (el *ErrorList) Add(pos token.Pos, msg string) {
	*el = append(*el, Error{pos, msg})
}
