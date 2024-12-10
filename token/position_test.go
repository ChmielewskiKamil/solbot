package token

import "testing"

func Test_GetPosition_LinesAndColumnsBasedOnOffset(t *testing.T) {
	fileContent := `hello
world
greetings
uint256
function`

	sf := &SourceFile{
		name:    "",
		content: fileContent,
	}

	sf.ComputeLineOffsets()

	if len(sf.Content()) == 0 {
		t.Fatalf("Length of source file content is zero.")
	}

	if len(sf.LineOffsets()) != 5 {
		t.Fatalf("File has different length than expected.")
	}

	tests := []struct {
		name     string // Name of the test case for clarity.
		offset   int    // The input offset to test.
		expected [2]int // The expected line and column numbers.
	}{
		{
			name:     "Start of file",
			offset:   0,
			expected: [2]int{1, 1}, // Line 1, Column 1.
		},
		{
			name:     "End of first line",
			offset:   5,
			expected: [2]int{1, 6}, // Line 1, Column 6.
		},
		{
			name:     "Start of second line",
			offset:   6,
			expected: [2]int{2, 1}, // Line 2, Column 1.
		},
		{
			name:     "Middle of third line",
			offset:   16,
			expected: [2]int{3, 5}, // Line 3, Column 5.
		},
		{
			name:     "Start of fifth line",
			offset:   30,
			expected: [2]int{5, 1}, // Line 5, Column 1.
		},
		{
			name:     "Out of bounds negative",
			offset:   -1,
			expected: [2]int{-1, -1}, // Invalid offset.
		},
		{
			name:     "Out of bounds too large",
			offset:   len(fileContent) + 1,
			expected: [2]int{-1, -1}, // Invalid offset.
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, column := sf.GetLineAndColumn(Pos(tt.offset))
			if line != tt.expected[0] || column != tt.expected[1] {
				t.Errorf("offset %d: expected (line, column) = %v, got (%d, %d)",
					tt.offset, tt.expected, line, column)
			}
		})
	}
}

func Fuzz_LineAndColumnOffsetConsistency(f *testing.F) {
	// Seed with sample inputs.
	f.Add("hello\nworld\ngreetings\nuint256\nfunction", 0)

	f.Fuzz(func(t *testing.T, content string, initialOffset int) {
		if len(content) == 0 {
			return // Skip empty content.
		}

		sf := &SourceFile{
			name:    "test",
			content: content,
		}
		sf.ComputeLineOffsets()

		// Clamp initialOffset to valid range.
		if initialOffset < 0 {
			initialOffset = 0
		} else if initialOffset >= len(content) {
			initialOffset = len(content) - 1
		}

		// Get line and column for the given offset.
		line, column := sf.GetLineAndColumn(Pos(initialOffset))
		if line == -1 || column == -1 {
			t.Fatalf("Invalid position for offset %d: line=%d, column=%d", initialOffset, line, column)
		}

		// Get offset back from line and column.
		recalculatedOffset := sf.GetOffset(line, column)
		if recalculatedOffset != Pos(initialOffset) {
			t.Errorf("Mismatch for initial offset %d: recalculated offset = %d", initialOffset, recalculatedOffset)
		}

	})
}
