package util

import (
	"reflect"
	"testing"
)

func TestSplitMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Segment
	}{
		{
			name:  "plain text",
			input: "Hello world",
			expected: []Segment{
				{Type: SegmentText, Content: "Hello world"},
			},
		},
		{
			name:  "single code block",
			input: "```go\nfunc main() {}\n```",
			expected: []Segment{
				{Type: SegmentCode, Content: "func main() {}", Language: "go"},
			},
		},
		{
			name:  "text and code block",
			input: "Here is code:\n```go\nfunc main() {}\n```",
			expected: []Segment{
				{Type: SegmentText, Content: "Here is code:\n"},
				{Type: SegmentCode, Content: "func main() {}", Language: "go"},
			},
		},
		{
			name:  "code block and text",
			input: "```go\nfunc main() {}\n```\nEnd",
			expected: []Segment{
				{Type: SegmentCode, Content: "func main() {}", Language: "go"},
				{Type: SegmentText, Content: "\nEnd"},
			},
		},
		{
			name:  "mixed content",
			input: "Start\n```python\nprint('hi')\n```\nMiddle\n```js\nconsole.log('hi')\n```\nEnd",
			expected: []Segment{
				{Type: SegmentText, Content: "Start\n"},
				{Type: SegmentCode, Content: "print('hi')", Language: "python"},
				{Type: SegmentText, Content: "\nMiddle\n"},
				{Type: SegmentCode, Content: "console.log('hi')", Language: "js"},
				{Type: SegmentText, Content: "\nEnd"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitMarkdown(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("SplitMarkdown() = %v, want %v", got, tt.expected)
			}
		})
	}
}
