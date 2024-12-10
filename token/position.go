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
	name                        string // file name e.g. "foo.sol"
	relativePathFromProjectRoot string // path to the file from project root e.g. where foundry.toml is defined
	content                     string // file content; source code passed to the parser
	lines                       []int  // offsets of the first character of each line
}

func (sf *SourceFile) Name() string {
	return sf.name
}

func (sf *SourceFile) Content() string {
	return sf.content
}

func (sf *SourceFile) RelativePathFromProjectRoot() string {
	return sf.relativePathFromProjectRoot
}

func (sf *SourceFile) LineOffsets() []int {
	return sf.lines
}

func (sf *SourceFile) ComputeLineOffsets() {
	// Explicitly add 0th line offset so that the lines length include it.
	// Since we append pointing to next line, on first iter, 0th one would
	// be skipped.
	lines := []int{0}
	for offset, char := range sf.content {
		if char == '\n' {
			// +1 to skip \n and point to start of next line
			lines = append(lines, offset+1)
		}
	}
	sf.lines = lines
}

// GetLineAndColumn returns the (line, column) position in a source file based on the
// provided offset. If the provided offset is invalid, the function returns (-1, -1).
// If the line offsets were not computed yet for this SourceFile, GetLineAndColumn
// will call ComputeLineOffsets function.
func (sf *SourceFile) GetLineAndColumn(offset Pos) (int, int) {
	// Check if the offset is valid
	if offset < 0 || int(offset) >= len(sf.content) {
		return -1, -1
	}

	if len(sf.lines) == 0 {
		sf.ComputeLineOffsets()
	}

	// Initialize the search range for binary search.
	// `low` is the start of the range, and `high` is the end of the range.
	// The `-1` for `high` ensures we stay within the valid index range of the array,
	// since arrays are zero-indexed.
	low, high := 0, len(sf.lines)-1

	// Perform binary search to find the line that contains the offset.
	for low <= high {
		// Calculate the middle index of the current search range.
		// This is where we will check if the offset falls before, at, or after this line.
		mid := (low + high) / 2

		// If the starting position of the current line (`sf.lines[mid]`)
		// is less than or equal to the target `offset`, then the offset
		// might be within this line or one of the lines after it.
		// In this case, we adjust `low` to narrow the search range to the right half.
		if sf.lines[mid] <= int(offset) {
			low = mid + 1
		} else {
			// If the starting position of the current line (`sf.lines[mid]`)
			// is greater than the target `offset`, then the offset must be in one
			// of the lines before this line. We adjust `high` to narrow the search
			// range to the left half.
			high = mid - 1
		}
	}

	// After the loop ends:
	// - `high` will point to the largest index where `sf.lines[high]` is still less than
	//   or equal to the `offset`. This means it identifies the start of the line that
	//   contains the `offset`.
	// - `low` will be the index of the next line, which is strictly greater than the `offset`.
	// We stop the loop because we've narrowed down the line where the `offset` resides.

	// +1 because line numbers in text editors start at 1 instead of 0.
	line := high + 1
	// The same +1 based indexing applies to columns.
	column := int(offset) - sf.lines[high] + 1

	return line, column
}

// GetOffset returns the offset in the source file based on the provided line and column.
// If the provided line or column is invalid, it returns -1.
// If the line offsets were not computed yet for this SourceFile, GetOffset
// will call ComputeLineOffsets function.
func (sf *SourceFile) GetOffset(line, column int) Pos {
	// Validate the line number.
	if line <= 0 || line > len(sf.lines) {
		return -1 // Invalid line number.
	}

	// Ensure the line offsets are computed.
	if len(sf.lines) == 0 {
		sf.ComputeLineOffsets()
	}

	// Get the starting offset for the specified line.
	lineStart := sf.lines[line-1] // Line numbers are 1-based, so subtract 1 for the index.

	// Compute the offset by adding the column offset.
	offset := lineStart + column - 1 // Columns are also 1-based, so subtract 1.

	// Validate the computed offset against the file length.
	if offset < 0 || offset >= len(sf.content) {
		return -1 // Invalid column for the given line.
	}

	return Pos(offset)
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
		name:                        filepath.Base(fileNameOrPath),
		relativePathFromProjectRoot: relativePath,
		content:                     string(content),
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
