package token

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

type SourceFile struct {
	name                    string // file name e.g. "foo.sol"
	filePathFromProjectRoot string // path to the file from project root e.g. where foundry.toml is defined
	content                 string // file content; source code passed to the parser
	lines                   []int  // offsets of the first character of each line
}

func (f *SourceFile) Name() string {
	return f.name
}

func (f *SourceFile) Content() string {
	return f.content
}

func NewSourceFile(fileNameOrPath, src string) (*SourceFile, error) {
	// Passing input string is useful for testing.
	if src != "" {
		return &SourceFile{
			name:    fileNameOrPath,
			content: src,
		}, nil
	}

	// In production use cases it is handy to read straight from file/path.
	content, err := os.ReadFile(fileNameOrPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from path %s: %w", fileNameOrPath, err)
	}

	// Compute relative path from project root
	relativePath, err := getRelativePath(fileNameOrPath)
	if err != nil {
		return nil, fmt.Errorf("failed to determine relative path for %s: %w", fileNameOrPath, err)
	}

	return &SourceFile{
		name:                    filepath.Base(fileNameOrPath),
		filePathFromProjectRoot: relativePath,
		content:                 string(content),
	}, nil
}

// projectMarkers defines files or directories that indicate the project root.
var projectMarkers = []string{
	"hardhat.config.js",
	"hardhat.config.ts",
	"foundry.toml",
	"remappings.txt",
	"truffle.js",
	"truffle-config.js",
	"ape-config.yaml",
	".git",
	"package.json",
	"node_modules",
}

// findProjectRoot traverses up the directory tree to locate the project root.
func findProjectRoot(startPath string) (string, error) {
	// startPath is the dir where .sol file is.
	currPath := startPath
	for {
		for _, marker := range projectMarkers {
			// Look for markers at each dir.
			if _, err := os.Stat(filepath.Join(currPath, marker)); err == nil {
				return currPath, nil
			}
		}
		parentDir := filepath.Dir(currPath)
		if parentDir == currPath {
			return "", fmt.Errorf("Couldn't find foundry/hardhat/ape project root. You can init a git repo (or add foundry.toml etc.) to create one.")
		}
		currPath = parentDir
	}
}

// getRelativePath computes the relative path to the project root.
func getRelativePath(filePath string) (string, error) {
	projectRoot, err := findProjectRoot(filepath.Dir(filePath))
	if err != nil {
		return "", err
	}
	return filepath.Rel(projectRoot, filePath)
}
