package analyzer

type Error struct {
	Loc string
	Msg string
}

type ErrorList []Error

func (el *ErrorList) Add(loc, msg string) {
	*el = append(*el, Error{loc, msg})
}
