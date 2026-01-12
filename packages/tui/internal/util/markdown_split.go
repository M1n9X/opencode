package util

import (
	"regexp"
	"strings"
)

type SegmentType int

const (
	SegmentText SegmentType = iota
	SegmentCode
)

type Segment struct {
	Type     SegmentType
	Content  string
	Language string
}

var codeBlockRegex = regexp.MustCompile("(?ms)^```([a-zA-Z0-9_+\\-]*)\n(.*?)\n```$")

func SplitMarkdown(text string) []Segment {
	var segments []Segment

	// Find all matches including submatches
	matches := codeBlockRegex.FindAllStringSubmatchIndex(text, -1)

	lastIndex := 0
	for _, match := range matches {
		// match indices: [start, end, langStart, langEnd, contentStart, contentEnd]
		start, end := match[0], match[1]
		langStart, langEnd := match[2], match[3]
		contentStart, contentEnd := match[4], match[5]

		// Add preceeding text
		if start > lastIndex {
			segments = append(segments, Segment{
				Type:    SegmentText,
				Content: text[lastIndex:start],
			})
		}

		language := ""
		if langStart != -1 {
			language = text[langStart:langEnd]
		}

		content := ""
		if contentStart != -1 {
			content = text[contentStart:contentEnd]
		}

		segments = append(segments, Segment{
			Type:     SegmentCode,
			Language: strings.TrimSpace(language),
			Content:  content,
		})

		lastIndex = end
	}

	// Add remaining text
	if lastIndex < len(text) {
		segments = append(segments, Segment{
			Type:    SegmentText,
			Content: text[lastIndex:],
		})
	}

	return segments
}
