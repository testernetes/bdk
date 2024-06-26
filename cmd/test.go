/*
Copyright © 2023 Matt Simons
*/
package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"plugin"
	"strings"
	"sync"
	"time"

	gherkin "github.com/cucumber/gherkin/go/v26"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/testernetes/bdk/model"
	"github.com/testernetes/bdk/printers"
	"github.com/testernetes/bdk/stepdef"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var format string
var tags string
var fastFail bool
var debug bool

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
				step, ok := v.(*stepdef.StepDefinition)
				if !ok {
					return errors.New(fmt.Sprintf("expected Step in %s to be a scheme.StepDefinition however it was %T", p, v))
				}
				model.StepFunctions.Register(*step)
			}

			gomega.RegisterFailHandler(func(message string, _ ...int) {
				panic(message)
			})

			filter := model.NewFilter(tags)

			printer, err := printers.NewPrinter(format)
			if err != nil {
				return fmt.Errorf("error creating formatter: %s\n", err)
			}

			features := []*model.Feature{}

			for _, gdp := range args {
				err = filepath.Walk(gdp, func(path string, info fs.FileInfo, err error) error {
					if info.IsDir() {
						return nil
					}

					if !strings.HasSuffix(info.Name(), ".feature") {
						return nil
					}

					gdf, err := os.Open(path)
					if err != nil {
						fmt.Println(err.Error())
						return nil
					}
					gd, err := gherkin.ParseGherkinDocument(bufio.NewReader(gdf), (&messages.Incrementing{}).NewId)
					if err != nil {
						return fmt.Errorf("error while loading document: %s\n", err)
					}
					defer gdf.Close()

					if gd.Feature == nil {
						return nil
					}

					feature, err := model.NewFeature(path, gd.Feature, filter)
					if err != nil {
						return fmt.Errorf("error creating feature from doc: %s\n", err)
					}

					if feature != nil {
						features = append(features, feature)
					}
					return nil
				})
				if err != nil {
					fmt.Printf("%s\n", err)
				}
			}

			ctx := context.Background()
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)

			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)

			go func() {
				var exitnow bool
				for {
					select {
					case <-c:
						fmt.Printf("\nUser Interrupted, jumping to cleanup now. Press ^C again to skip cleanup.\n\n")
						cancel()
						if exitnow {
							panic("exited before cleanup could finish, there may be some leftover resources")
						}
						exitnow = true
					case <-ctx.Done():
						return
					}
				}
			}()

			log.SetLogger(zap.New(zap.UseDevMode(debug), zap.WriteTo(os.Stderr)))
			ctx = log.IntoContext(ctx, log.Log.WithCallDepth(1))

			events := make(model.Events)
			go printer.Print(events)

			exitCode := 0
			var wg sync.WaitGroup
			for i := range features {
				wg.Add(1)
				go func(feature *model.Feature) {
					defer wg.Done()
					err := feature.Run(ctx, &events)
					if err != nil {
						fmt.Println(err.Error()) // should be stderr
						exitCode = 1
						if fastFail {
							cancel()
						}
					}
				}(features[i])
			}

			wg.Wait()

			events.Close()
			signal.Stop(c)
			cancel()

			time.Sleep(time.Second)
			os.Exit(exitCode)
			return nil
		},
	}
	cmd.Flags().StringVarP(&format, "format", "f", "simple", "the format printer")
	cmd.Flags().StringVarP(&tags, "tags", "t", "", "tags to filter")
	cmd.Flags().BoolVarP(&fastFail, "fast-fail", "", false, "stop testing on first failure")
	cmd.Flags().BoolVarP(&debug, "debug", "D", false, "show debug logs")

	cmd.Flags().String("format-configmap-name", "results", "name of configmap to write results to")
	viper.BindPFlag("format-configmap-name", cmd.Flags().Lookup("format-configmap-name"))
	cmd.Flags().String("format-configmap-namespace", "default", "namespace of configmap to write results to")
	viper.BindPFlag("format-configmap-namespace", cmd.Flags().Lookup("format-configmap-namespace"))
	return cmd
}
