package main

import (
	"os"
	"github.com/falcosecurity/falco-operator/cmd"
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
		os.Exit(1)
	}
}
