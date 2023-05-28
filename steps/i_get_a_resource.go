package steps

import (
	"context"

	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/contextutils"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/bdk/scheme"
)

func init() {
	scheme.Default.MustAddToScheme(IGetAResource)
}

var IGetAResource = scheme.StepDefinition{
	Name: "i-get",
	Text: "I get <reference>",
	Help: `Gets the referenced resource. Step will fail if the reference was not defined in a previous step.`,
	Examples: `
	Given a cm from file blah.yaml
	And I get cm
	Then cm jsonpath '{.metadata.uid}' should not be empty`,
	Parameters: []parameters.Parameter{parameters.Reference},
	Function: func(ctx context.Context, ref string) error {
		o := contextutils.LoadObject(ctx, ref)
		Expect(o).ShouldNot(BeNil(), ErrNoResource, ref)

		c := contextutils.MustGetClientFrom(ctx)
		Eventually(c.Get).WithContext(ctx).WithArguments(o).Should(Succeed(), "Failed to create resource")

		contextutils.SaveObject(ctx, ref, o)

		return nil
	},
}
