package steps

import (
	"context"

	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/bdk/store"
)

var ISetVar = stepdef.StepDefinition{
	Name:     "i-set-var",
	Text:     "^I set {reference} to {text}$",
	Function: SetVarFunc,
	StepArg:  stepdef.NoStepArg,
	Help:     `Assigns a value to a variable`,
	Examples: `Given a resource called cm:
	  """
	  apiVersion: v1
	  kind: ConfigMap
	  metadata:
	    name: example
	    namespace: default
	  data:
	    foo: bar
	  """`,
}

var SetVarFunc = func(ctx context.Context, key, value string) (err error) {
	key = "scn-var-" + key
	store.Save(ctx, key, value)
	return nil
}
