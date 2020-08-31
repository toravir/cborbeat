package main

import (
	"os"

	"github.com/toravir/cborbeat/cmd"

	_ "github.com/toravir/cborbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
