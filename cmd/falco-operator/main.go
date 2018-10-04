package main

import (
	"os"
	"github.com/mumoshu/falco-operator/cmd"
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
		os.Exit(1)
	}
}
