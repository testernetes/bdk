package model

import (
	"context"
	"errors"
	"strings"

	messages "github.com/cucumber/messages/go/v21"
)

type Feature struct {
	*messages.Feature
	Path string `json:"string"`

	//Rules []*Rule
	Scenarios []*Scenario `json:"scenarios"` // or Scenario Outline
}

func NewFeature(path string, featureDoc *messages.Feature, filters []Filter) (*Feature, error) {
	f := &Feature{
		Feature: featureDoc,
		Path:    path,
	}
	var rules []*messages.Rule

	var backgroundDoc *messages.Background
	for _, fc := range featureDoc.Children {
		if fc.Background != nil {
			if backgroundDoc != nil {
				return f, errors.New("a feature can only have one background")
			}
			backgroundDoc = fc.Background
		}
	}

	for _, fc := range featureDoc.Children {
		if fc.Rule != nil {
			rules = append(rules, fc.Rule)
		}
		if fc.Scenario != nil {
			scenarioTags := NewTags(append(featureDoc.Tags, fc.Scenario.Tags...))
			if isFiltered(scenarioTags, filters) {
				continue
			}

			if len(fc.Scenario.Examples) == 0 {
				s, err := NewScenario(backgroundDoc, fc.Scenario)
				if err != nil {
					return f, err
				}
				f.Scenarios = append(f.Scenarios, s)
			}

			for _, example := range fc.Scenario.Examples {
				for _, r := range example.TableBody {
					replacer := map[string]string{}
					for i, v := range r.Cells {
						key := "<" + example.TableHeader.Cells[i].Value + ">"
						replacer[key] = v.Value
					}
					scn := deepCopyScenarioDoc(fc.Scenario)
					for k, v := range replacer {
						scn.Name = strings.ReplaceAll(scn.Name, k, v)
						scn.Description = strings.ReplaceAll(scn.Description, k, v)
						for _, s := range scn.Steps {
							s.Text = strings.ReplaceAll(s.Text, k, v)
							if s.DocString != nil {
								s.DocString.Content = strings.ReplaceAll(s.DocString.Content, k, v)
							}
							if s.DataTable != nil {
								for _, row := range s.DataTable.Rows {
									for _, cell := range row.Cells {
										cell.Value = strings.ReplaceAll(cell.Value, k, v)
									}
								}

							}
						}
					}
					s, err := NewScenario(backgroundDoc, scn)
					if err != nil {
						return f, err
					}
					f.Scenarios = append(f.Scenarios, s)
				}
			}
		}
	}

	//for _, RuleDoc := range rules {
	//	s, err := NewRule(ruleDoc, scheme)
	//	if err != nil {
	//		return f, err
	//	}
	//	f.Rules = append(f.Rules, s)
	//}

	if len(f.Scenarios) == 0 {
		return nil, nil
	}
	return f, nil
}

func (f *Feature) Run(ctx context.Context, events Events) error {
	events.StartFeature(f)
	for _, scenario := range f.Scenarios {
		err := scenario.Run(ctx, events)
		if err != nil {
			return err
		}
	}
	events.FinishFeature(f)
	return nil
}

func deepCopyScenarioDoc(in *messages.Scenario) *messages.Scenario {
	if in == nil {
		return nil
	}
	out := &messages.Scenario{}

	out.Tags = in.Tags
	out.Keyword = in.Keyword
	out.Name = in.Name
	out.Description = in.Description
	for _, s := range in.Steps {
		out.Steps = append(out.Steps, deepCopyStepDoc(s))
	}
	return out
}

func deepCopyStepDoc(in *messages.Step) *messages.Step {
	if in == nil {
		return nil
	}
	out := &messages.Step{}
	out.Keyword = in.Keyword
	out.KeywordType = in.KeywordType
	out.Text = in.Text
	out.DocString = deepCopyDocString(in.DocString)
	out.DataTable = deepCopyDataTable(in.DataTable)
	return out
}

func deepCopyDocString(in *messages.DocString) *messages.DocString {
	if in == nil {
		return nil
	}
	out := &messages.DocString{
		Content:   in.Content,
		MediaType: in.MediaType,
		Delimiter: in.Delimiter,
	}
	return out
}

func deepCopyDataTable(in *messages.DataTable) *messages.DataTable {
	if in == nil {
		return nil
	}
	out := &messages.DataTable{
		Rows: []*messages.TableRow{},
	}

	for _, r := range in.Rows {
		row := &messages.TableRow{Cells: []*messages.TableCell{}}
		for _, c := range r.Cells {
			cell := &messages.TableCell{Value: c.Value}
			row.Cells = append(row.Cells, cell)
		}
		out.Rows = append(out.Rows, row)
	}
	return out
}
