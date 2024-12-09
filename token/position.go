package token

import (
	"bufio"
	"io"
)

// Pos is the file offset to the beginning of a token, starting from 0
type Pos int

type Position struct {
	Filename string
	Offset   Pos
	Line     int
	Column   int
}

// TODO: offset to position could be the method of pos which takes a ref to a
// file and returns a proper position struct.

func OffsetToPosition(reader io.Reader, pos *Position) {
	// Wrap into a buffered reader to use utility functions e.g. ReadString()
	bufReader := bufio.NewReader(reader)
	offset := pos.Offset
	line := 1
	column := 1
	cumulativeOffset := 0

	for {
		l, err := bufReader.ReadString('\n')
		lineLength := len(l)
		// err is most likely io.EOF or error where no data was read,
		// so we want to break.
		// If there is an other error, e.g. \n delimeter is not found, we
		// might have some remaining data in l left to be processed. In
		// this scenario we don't want to break.
		if err != nil && lineLength == 0 {
			break
		}

		// The offset that we are looking for is in the current line.
		if cumulativeOffset+lineLength > int(offset) {
			// By subtracting the bytes that we accumulated so far from the
			// offset that we are looking for, we get the bytes from the
			// start of the current line up to the thing that we are looking for.
			// This is the column number. Since the human readable representation
			// of the column in the text editor starts from 1, we add 1 to get
			// the 1 based indexing.
			column = int(offset) - cumulativeOffset + 1
			// Target reached.
			break
		}

		cumulativeOffset += lineLength
		line++
	}

	pos.Line = line
	pos.Column = column
}

type File struct {
	name  string // file name e.g. "foo.sol"
	src   string // file content; source code passed to the parser
	lines []int  // offsets of the first character of each line
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Src() string {
	return f.src
}

func NewFile(name, src string) *File {
	return &File{
		name: name,
		src:  src,
	}
}
