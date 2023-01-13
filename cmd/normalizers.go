package cmd

import (
	"strings"
)

const Indentation = `  `

// Examples normalizes a command's examples to follow the conventions.
func Examples(s string) string {
	if len(s) == 0 {
		return s
	}
	return normalizer{s}.trimTabs().indentTabs(1).string
}

// Parameter normalizes a command's examples to follow the conventions.
func Parameter(s string) string {
	if len(s) == 0 {
		return s
	}
	return normalizer{s}.trim().indent(2).string
}

type normalizer struct {
	string
}

func (s normalizer) trim() normalizer {
	s.string = strings.TrimSpace(s.string)
	return s
}

func (s normalizer) trimTabs() normalizer {
	s.string = strings.TrimLeft(s.string, "\t")
	return s
}

func (s normalizer) indent(by int) normalizer {
	indentedLines := []string{}
	for _, line := range strings.Split(s.string, "\n") {
		trimmed := strings.TrimSpace(line)
		indented := strings.Repeat(Indentation, by) + trimmed
		indentedLines = append(indentedLines, indented)
	}
	s.string = strings.Join(indentedLines, "\n") + "\n"
	return s
}

func (s normalizer) indentTabs(by int) normalizer {
	indentedLines := []string{}
	for _, line := range strings.Split(s.string, "\n") {
		trimmed := strings.TrimLeft(line, "\t")
		indented := strings.Repeat(Indentation, by) + trimmed
		indentedLines = append(indentedLines, indented)
	}
	s.string = strings.Join(indentedLines, "\n") + "\n"
	return s
}
