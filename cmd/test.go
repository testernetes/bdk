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
	"github.com/spf13/viper"
	"github.com/testernetes/bdk/formatters"
	"github.com/testernetes/bdk/model"
	"github.com/testernetes/bdk/scheme"
)

var format string
var tags string

// testCmd represents running a test suite
func NewTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test features...",
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
				step, ok := v.(*scheme.StepDefinition)
				if !ok {
					return errors.New(fmt.Sprintf("expected Step in %s to be a scheme.StepDefinition however it was %T", p, v))
				}
				scheme.Default.AddToScheme(*step)
			}

			gomega.RegisterFailHandler(func(message string, _ ...int) {
				panic(message)
			})

			filter := model.NewFilter(tags)

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

				feature, err := model.NewFeature(gd.Feature, scheme.Default, filter)
				if err != nil {
					return fmt.Errorf("error creating feature from doc: %s\n", err)
				}

				if feature == nil {
					continue
				}

				ctx := context.TODO()
				feature.Run(ctx)

				f, err := formatters.NewFormatter(format)
				if err != nil {
					return fmt.Errorf("error creating formatter: %s\n", err)
				}
				f.Print(feature)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&format, "format", "f", "simple", "the format printer")
	cmd.Flags().String("format-configmap-name", "results", "name of configmap to write results to")
	viper.BindPFlag("format-configmap-name", cmd.Flags().Lookup("format-configmap-name"))
	cmd.Flags().String("format-configmap-namespace", "default", "namespace of configmap to write results to")
	viper.BindPFlag("format-configmap-namespace", cmd.Flags().Lookup("format-configmap-namespace"))
	cmd.Flags().StringVarP(&tags, "tags", "t", "", "tags to filter")
	return cmd
}
