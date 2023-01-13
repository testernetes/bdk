/*
Copyright © 2023 Matt Simons
*/
package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"plugin"

	gherkin "github.com/cucumber/gherkin/go/v26"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"github.com/testernetes/bdk/formatters"
	"github.com/testernetes/bdk/models"
)

// testCmd represents running a test suite
var testCmd = &cobra.Command{
	Use:   "test [-f format] features...",
	Short: "Run a test suite of feature files",
	Long:  "",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		for _, p := range plugins {
			plugin, err := plugin.Open(p)
			if err != nil {
				return err
			}
			v, err := plugin.Lookup("Step")
			if err != nil {
				return errors.New(fmt.Sprintf("could not find a variable called Step in %s", p))
			}
			step, ok := v.(*models.StepDefinition)
			if !ok {
				return errors.New(fmt.Sprintf("expected Step in %s to be a models.StepDefinition however it was %T", p, v))
			}
			models.Scheme.Register(*step)
		}

		gomega.RegisterFailHandler(func(message string, _ ...int) {
			panic(message)
		})

		for _, gdp := range args {
			gdf, err := os.Open(gdp)

			gd, err := gherkin.ParseGherkinDocument(bufio.NewReader(gdf), (&messages.Incrementing{}).NewId)
			if err != nil {
				return fmt.Errorf("error while loading document: %s\n", err)
			}
			defer gdf.Close()

			if gd.Feature == nil {
				continue
			}

			feature, err := models.NewFeature(gd.Feature, models.Scheme)
			if err != nil {
				return fmt.Errorf("error creating feature from doc: %s\n", err)
			}

			ctx := context.TODO()
			feature.Run(ctx)

			formatters.Print(formatter, feature)
		}
		return nil
	},
}
