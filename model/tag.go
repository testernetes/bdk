package model

import (
	"strings"

	messages "github.com/cucumber/messages/go/v21"
)

type Tag struct {
	string
}

type Filter struct {
	string
	bool
}

func (f Filter) filters(t Tag) bool {
	if t.string == f.string {
		return true == f.bool
	}
	return false == f.bool
}

func isFiltered(tags []Tag, filter []Filter) bool {
	for _, f := range filter {
		if f.bool && len(tags) == 0 {
			return true
		}
		for _, t := range tags {
			if !f.filters(t) {
				return true
			}
		}
	}
	return false
}

func NewTags(in []*messages.Tag) (out []Tag) {
	for _, t := range in {
		out = append(out, Tag{t.Name[1:]})
	}
	return
}

func NewFilter(in string) (out []Filter) {
	for _, andTags := range strings.Split(in, "&&") {
		for _, tag := range strings.Split(andTags, ",") {
			tag = strings.TrimSpace(tag)
			tag = strings.ReplaceAll(tag, "@", "")

			if len(tag) == 0 {
				continue
			}

			if tag[0] == '~' {
				tag = tag[1:]
				out = append(out, Filter{tag, false})
				continue
			}
			out = append(out, Filter{tag, true})
		}
	}
	return
}
