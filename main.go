/*
Copyright © 2023 Matt Simons
*/
package main

import (
	"os"

	"github.com/testernetes/bdk/cmd"
	_ "github.com/testernetes/bdk/steps"
)

func main() {
	rootCmd := cmd.NewRootCommand()
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
