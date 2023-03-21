package utils

import (
	"fmt"
	"strings"
)

const Indentation = `  `

// Examples normalizes a command's examples to follow the conventions.
func Examples(s string) string {
	if len(s) == 0 {
		return s
	}
	return normalizer{s}.TrimTabs().IndentTabs(1).String()
}

// Parameter normalizes a command's examples to follow the conventions.
func HTMLParameter(s string) string {
	if len(s) == 0 {
		return s
	}
	return normalizer{s}.Trim().Definition().String()
}

func Parameter(s string) string {
	if len(s) == 0 {
		return s
	}
	return normalizer{s}.Trim().IndentTabs(2).String()
}

type Normalizer interface {
	String() string
	Print()
	Trim() Normalizer
	TrimTabs() Normalizer
	TrimAllTabs() Normalizer
	Definition() Normalizer
	Indent(int) Normalizer
	IndentTabs(int) Normalizer
	Snippet(string) Normalizer
}

type normalizer struct {
	string
}

func NewNormalizer(s string, a ...any) Normalizer {
	return normalizer{fmt.Sprintf(s, a...)}
}

func (s normalizer) String() string {
	return s.string
}

func (s normalizer) Print() {
	fmt.Println(s.string)
}

func (s normalizer) Trim() Normalizer {
	s.string = strings.TrimSpace(s.string)
	return s
}

func (s normalizer) TrimTabs() Normalizer {
	s.string = strings.TrimLeft(s.string, "\t")
	return s
}

func (s normalizer) TrimAllTabs() Normalizer {
	s.Trim()
	indentedLines := []string{}
	for _, line := range strings.Split(s.string, "\n") {
		trimmed := strings.TrimSpace(line)
		indentedLines = append(indentedLines, trimmed)
	}
	s.string = strings.Join(indentedLines, "\n") // + "\n"
	return s
}

func (s normalizer) Snippet(lang string) Normalizer {
	trimmedLines := []string{}
	s.string = "```" + lang + "\n" + s.string
	for _, line := range strings.Split(s.string, "\n") {
		trimmed := strings.TrimLeft(line, "\t")
		trimmedLines = append(trimmedLines, trimmed)
	}
	s.string = strings.Join(trimmedLines, "\n") + "\n```\n"
	return s
}

func (s normalizer) Definition() Normalizer {
	trimmed := ": "
	for _, line := range strings.Split(s.string, "\n") {
		trimmed += strings.TrimSpace(line) + " "
	}
	s.string = trimmed + "\n"
	return s
}

func (s normalizer) Indent(by int) Normalizer {
	indentedLines := []string{}
	for _, line := range strings.Split(s.string, "\n") {
		trimmed := strings.TrimSpace(line)
		indented := strings.Repeat(Indentation, by) + trimmed
		indentedLines = append(indentedLines, indented)
	}
	s.string = strings.Join(indentedLines, "\n") // + "\n"
	return s
}

func (s normalizer) IndentTabs(by int) Normalizer {
	indentedLines := []string{}
	for _, line := range strings.Split(s.string, "\n") {
		trimmed := strings.TrimLeft(line, "\t")
		indented := strings.Repeat(Indentation, by) + trimmed
		indentedLines = append(indentedLines, indented)
	}
	s.string = strings.Join(indentedLines, "\n") // + "\n"
	return s
}
