package main

import (
	"os"

	"github.com/eljakani/ward/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
